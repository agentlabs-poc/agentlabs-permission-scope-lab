// Command auth-service serves the standard endpoint set over HTTP.
//
// It is the Auth side of the boundary: it holds the records and answers one
// question about them. The application on the other side holds none.
package main

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	regdomain "agentlabs.local/registry/domain"
	"agentlabs.local/wiring"
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type clock struct{}

func (clock) Now() time.Time { return time.Now().UTC() }

// labAgents admits the fixture's workload credential and nothing else.
//
// A deployment verifies a real token here and derives the calling application
// from it. The lab has no issuer — that belongs to the Auth service this
// repository migrates into rather than becomes — so the credential is modelled
// and the shape it arrives in is what matters.
type labAgents struct{}

func (labAgents) Establish(r *http.Request) (domain.Identity, error) {
	// Constant time, because this is the line a deployment replaces with a real
	// verification and the shape it copies should be the right one.
	offered := []byte(r.Header.Get("Authorization"))
	expected := []byte("Bearer " + lab.WorkloadToken)
	if subtle.ConstantTimeCompare(offered, expected) != 1 {
		return domain.Identity{}, fmt.Errorf("unrecognised credential")
	}
	return domain.Identity{Version: "1", Actor: domain.Actor{Type: "service_account", ID: lab.WorkloadClient}}, nil
}

func main() { os.Exit(run(os.Args[1:])) }

const usage = "usage: auth-service --authority PATH --registry PATH --tenant ID --app ID [--listen ADDR]"

func run(args []string) int {
	// Every flag here takes a value, so the arguments are pairs. Anything else
	// is a mistake worth naming rather than a value quietly dropped.
	if len(args)%2 != 0 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	flags := map[string]string{}
	for i := 0; i < len(args); i += 2 {
		if !strings.HasPrefix(args[i], "--") {
			fmt.Fprintln(os.Stderr, usage)
			return 2
		}
		flags[args[i]] = args[i+1]
	}
	authority, registry, listen := flags["--authority"], flags["--registry"], flags["--listen"]
	if authority == "" || registry == "" {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	if listen == "" {
		listen = "127.0.0.1:8080"
	}
	area, err := domain.NewArea(flags["--tenant"], flags["--app"])
	if err != nil {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	status, err := lab.NewAssignmentStatusAdministration(area, lab.TeamFINC17(area).Administration)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 4
	}
	service, err := wiring.Open(context.Background(), wiring.Config{
		AuthorityPath: authority, RegistryPath: registry, CreateRegistry: true,
		Administration:         &lab.RoleAdministration{AssignmentStatusAdministration: status},
		RegistryAdministration: labRegistryAdmin{},
		Operator:               operatorIdentity(),
		Clock:                  clock{},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 4
	}
	defer service.Close()

	if err := install(service, area, operatorIdentity()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 4
	}
	handler, err := service.Handler(labAgents{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 4
	}
	fmt.Printf("auth-service listening on %s for %s/%s\n", listen, area.TenantID(), area.ApplicationID())
	return serve(listen, logged(handler, os.Stdout))
}

// install registers the application and installs it into the tenant, so the
// demonstration has an area to answer about. In a deployment this is a tenant
// administrator's own act, through the registry's own surface.
//
// A conflict means it is already there, which a second run of the lab produces
// and nothing else. Any other failure is fatal: a service that starts without
// its installation answers every question ErrNotFound, which is a service that
// denies everything rather than an error an operator can see.
func install(service *wiring.Service, area domain.Area, as regdomain.Identity) error {
	ctx := context.Background()
	_, err := service.Applications().RegisterApplication(ctx, as, area.ApplicationID(), area.ApplicationID())
	if err != nil && !errors.Is(err, regdomain.ErrConflict) {
		return fmt.Errorf("register %s: %w", area.ApplicationID(), err)
	}
	err = service.Applications().Install(ctx, as, area.TenantID(), area.ApplicationID())
	if err != nil && !errors.Is(err, regdomain.ErrConflict) {
		return fmt.Errorf("install %s into %s: %w", area.ApplicationID(), area.TenantID(), err)
	}
	return nil
}

// serve runs until a signal, then drains. The store is SQLite and closing it is
// the caller's deferred business — which a process that only ever dies inside
// ListenAndServe would never reach.
func serve(listen string, handler http.Handler) int {
	server := &http.Server{
		Addr:              listen,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		// A caller that completes its headers and then stalls its body holds a
		// goroutine and a connection for as long as it likes. This service is
		// the one every application's authorization depends on.
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	stopping := make(chan os.Signal, 1)
	signal.Notify(stopping, os.Interrupt, syscall.SIGTERM)
	failed := make(chan error, 1)
	go func() { failed <- server.ListenAndServe() }()
	select {
	case err := <-failed:
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintln(os.Stderr, err)
			return 4
		}
	case <-stopping:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 4
		}
	}
	return 0
}

// status remembers what the service answered, so logged can refuse to write a
// line for a question the service never accepted.
type status struct {
	http.ResponseWriter
	code int
}

func (s *status) WriteHeader(code int) { s.code = code; s.ResponseWriter.WriteHeader(code) }

// logged prints the route and the body of every answered question, so a
// demonstration can show what actually crossed the boundary rather than
// asserting it. The body is the claim worth checking: whether the application
// asked what a human holds, or asked for a decision.
//
// Two rules make the log evidence rather than decoration. It reads to the
// service's own limit, so a question this wrapper can log is exactly a question
// the service can answer. And it writes only after the service answered 200 —
// an unauthenticated caller must not be able to put lines into the record a
// demonstration reads.
func logged(next http.Handler, to io.Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, wiring.MaxRequestBytes+1))
		// Whatever was read is handed back whole, error or not: the service
		// must see the request this wrapper saw.
		r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(body), r.Body))
		answered := &status{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(answered, r)
		if err != nil || answered.code != http.StatusOK {
			return
		}
		// Compacted rather than printed raw: a body is caller-supplied text, and
		// a newline inside one would otherwise forge a line of its own.
		var compact bytes.Buffer
		if json.Compact(&compact, body) != nil {
			return
		}
		fmt.Fprintf(to, "  auth  <- %s %s\n        %s\n", r.Method, r.URL.Path, compact.Bytes())
	})
}
