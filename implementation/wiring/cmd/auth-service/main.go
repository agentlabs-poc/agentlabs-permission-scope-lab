// Command auth-service serves the standard endpoint set over HTTP.
//
// It is the Auth side of the boundary: it holds the records and answers one
// question about them. The application on the other side holds none.
package main

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"agentlabs.local/wiring"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	if r.Header.Get("Authorization") != "Bearer "+lab.WorkloadClient {
		return domain.Identity{}, fmt.Errorf("unrecognised credential")
	}
	return domain.Identity{Version: "1", Actor: domain.Actor{Type: "service_account", ID: lab.WorkloadClient}}, nil
}

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	flags := map[string]string{}
	for i := 0; i+1 < len(args); i++ {
		if strings.HasPrefix(args[i], "--") && !strings.HasPrefix(args[i+1], "--") {
			flags[args[i]] = args[i+1]
		}
	}
	authority, registry, listen := flags["--authority"], flags["--registry"], flags["--listen"]
	if authority == "" || registry == "" {
		fmt.Fprintln(os.Stderr, "usage: auth-service --authority PATH --registry PATH [--listen ADDR]")
		return 2
	}
	if listen == "" {
		listen = "127.0.0.1:8080"
	}
	area, err := domain.NewArea(flags["--tenant"], flags["--app"])
	if err != nil {
		fmt.Fprintln(os.Stderr, "explicit --tenant and --app are required to build the lab's gate")
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

	// Registered and installed here so the demonstration has an area to answer
	// about. In a deployment this is a tenant administrator's own act, through
	// the registry's own surface.
	if _, err := service.Applications().RegisterApplication(context.Background(), operatorIdentity(), area.ApplicationID(), area.ApplicationID()); err != nil {
		fmt.Fprintln(os.Stderr, "register:", err)
	}
	if err := service.Applications().Install(context.Background(), operatorIdentity(), area.TenantID(), area.ApplicationID()); err != nil {
		fmt.Fprintln(os.Stderr, "install:", err)
	}

	handler, err := service.Handler(labAgents{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 4
	}
	fmt.Printf("auth-service listening on %s for %s/%s\n", listen, area.TenantID(), area.ApplicationID())
	server := &http.Server{Addr: listen, Handler: logged(handler), ReadHeaderTimeout: 5 * time.Second}
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 4
	}
	return 0
}

// logged prints the route and the body of every question, so a demonstration can
// show what actually crossed the boundary rather than asserting it. The body is
// the claim worth checking: whether the application asked what a human holds, or
// asked for a decision.
func logged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 8<<10))
		if err == nil {
			r.Body = io.NopCloser(bytes.NewReader(body))
			fmt.Printf("  auth  <- %s %s\n        %s\n", r.Method, r.URL.Path, bytes.TrimSpace(body))
		}
		next.ServeHTTP(w, r)
	})
}
