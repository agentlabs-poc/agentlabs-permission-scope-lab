package main

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/httpdemo"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/localadapter"
	"agentlabs.local/authmiddleware"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

type clock struct{}

func (clock) Now() time.Time { return time.Now() }

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, diag io.Writer) int {
	flags := flag.NewFlagSet("auth-http-demo", flag.ContinueOnError)
	flags.SetOutput(diag)
	var db, tenant, application, human string
	flags.StringVar(&db, "db", "", "existing ABV SQLite database")
	flags.StringVar(&tenant, "tenant", "", "trusted test tenant")
	flags.StringVar(&application, "application", "", "trusted test application")
	flags.StringVar(&human, "human", "", "trusted test human")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		if err == nil {
			fmt.Fprintln(diag, "unexpected positional argument")
		}
		return 2
	}
	if strings.TrimSpace(db) == "" || strings.TrimSpace(tenant) == "" || strings.TrimSpace(application) == "" || strings.TrimSpace(human) == "" {
		fmt.Fprintln(diag, "db, tenant, application and human are required")
		return 2
	}
	// As in auth-evaluate: the demo trusts the area it was given, and a
	// deployment composes the real registry in its place.
	area, err := domain.NewArea(tenant, application)
	if err != nil {
		fmt.Fprintln(diag, err)
		return 2
	}
	registry, err := lab.NewFixedRegistry(area)
	if err != nil {
		fmt.Fprintln(diag, err)
		return 4
	}
	// The gate this path never had. Resolving what a human is entitled to is a
	// gated read, and until Facade.ResolveAuthority existed the adapter walked
	// the lineage itself with nothing deciding who may ask. The lab answers it
	// with the fixture's rules; a deployment answers it with its own.
	status, err := lab.NewAssignmentStatusAdministration(area, lab.TeamFINC17(area).Administration)
	if err != nil {
		fmt.Fprintln(diag, err)
		return 4
	}
	admin := &lab.RoleAdministration{AssignmentStatusAdministration: status}
	// The agent asks as itself. The human it asks about arrives on the request.
	credential := domain.Actor{Type: "service_account", ID: lab.WorkloadClient}
	source, err := localadapter.Open(context.Background(), db, credential, admin, clock{}, registry)
	if err != nil {
		fmt.Fprintln(diag, err)
		return 4
	}
	defer source.Close()
	evaluator, err := authmiddleware.New(source, clock{})
	if err != nil {
		fmt.Fprintln(diag, err)
		return 4
	}
	handler, err := httpdemo.NewHandler(httpdemo.NewStore(httpdemo.DefaultRecords()), evaluator, httpdemo.TrustedIdentity(tenant, application, human))
	if err != nil {
		fmt.Fprintln(diag, err)
		return 4
	}
	fmt.Fprintln(out, "LAB ONLY: trusted flag context; no network listener or JWT authentication")
	for _, example := range []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/v1/" + tenant + "/FIN/C17", ""},
		{http.MethodGet, "/api/v1/" + tenant + "/FIN/C18", ""},
		{http.MethodPut, "/api/v1/" + tenant + "/certificates/C17", `{"department_id":"FIN","title":"FIN walkthrough"}`},
		{http.MethodGet, "/api/v1/" + tenant + "/departments/FIN/certificates", ""},
		{http.MethodGet, "/api/v1/" + tenant + "/certificates", ""},
	} {
		request := httptest.NewRequest(example.method, example.path, bytes.NewBufferString(example.body))
		if example.body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		fmt.Fprintf(out, "%s %s %d %s\n", example.method, example.path, response.Code, strings.TrimSpace(response.Body.String()))
	}
	return 0
}
