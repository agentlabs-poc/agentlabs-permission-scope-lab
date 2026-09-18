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

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 15 * time.Second
	// The drain has to outlast both, or a request still arriving when the signal
	// lands turns an orderly stop into a reported failure.
	shutdownGrace = readTimeout + writeTimeout + 5*time.Second
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
		PlatformNamespace:      lab.PlatformNamespace,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 4
	}
	defer service.Close()

	if err := install(service.Applications(), area, operatorIdentity()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 4
	}
	handler, err := service.Handler(labAgents{}, printQuestion(os.Stdout))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 4
	}
	fmt.Printf("auth-service listening on %s for %s/%s\n", listen, area.TenantID(), area.ApplicationID())
	return serve(listen, handler)
}

// install registers the application and installs it into the tenant, so the
// demonstration has an area to answer about. In a deployment this is a tenant
// administrator's own act, through the registry's own surface.
//
// A conflict means it is already there, which a second run of the lab produces
// and nothing else. Any other failure is fatal: a service that starts without
// its installation answers every question ErrNotFound, which is a service that
// denies everything rather than an error an operator can see.
// applications is the part of the registry startup uses. It is an interface so
// each branch below can be tested for what it tolerates: through the real
// facade, a caller the registry refuses is refused at registration and the
// installation is never reached, which left the second branch unexercised.
type applications interface {
	RegisterApplication(context.Context, regdomain.Identity, string, string) (regdomain.Application, error)
	Install(context.Context, regdomain.Identity, string, string) error
}

func install(registry applications, area domain.Area, as regdomain.Identity) error {
	ctx := context.Background()
	_, err := registry.RegisterApplication(ctx, as, area.ApplicationID(), area.ApplicationID())
	if err != nil && !errors.Is(err, regdomain.ErrConflict) {
		return fmt.Errorf("register %s: %w", area.ApplicationID(), err)
	}
	err = registry.Install(ctx, as, area.TenantID(), area.ApplicationID())
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
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
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
		// Longer than the read and write timeouts above, because a connection
		// that is mid-body when the signal arrives is entitled to live until
		// one of those fires. A grace shorter than they are made every shutdown
		// with one stalled caller report failure — and any caller at all, even
		// an unauthenticated one, could hold a connection open.
		ctx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			// Not a failure exit: the service stopped accepting work and drained
			// what it could. A supervisor reads a non-zero status as a crash,
			// and a slow caller is not one.
			fmt.Fprintln(os.Stderr, "stopped with requests still in flight:", err)
		}
	}
	return 0
}

// printQuestion records one question the service accepted and answered, so a
// demonstration can show what crossed the boundary rather than asserting it. The
// body is the claim worth checking: whether the application asked what a human
// holds, or asked for a decision.
//
// It is an observer rather than a wrapper around the handler. A wrapper has to
// read the body to see the question, which puts that read in front of the
// service's authentication — an unauthenticated caller could then make the
// service buffer every request, and write lines of its own choosing into the
// record this log is.
func printQuestion(to io.Writer) wiring.Observer {
	return func(method, path string, question []byte) {
		// Compacted rather than echoed: a body is caller-supplied text, and a
		// line break inside one would otherwise forge a line of its own. The
		// path is caller-supplied too, and carries no structure to preserve, so
		// anything that could break the line is simply removed.
		var compact bytes.Buffer
		if json.Compact(&compact, question) != nil {
			return
		}
		fmt.Fprintf(to, "  auth  <- %s %s\n        %s\n", method, oneLine(method + path)[len(method):], compact.Bytes())
	}
}

// oneLine strips anything that could end a line or hide what follows it.
func oneLine(text string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, text)
}
