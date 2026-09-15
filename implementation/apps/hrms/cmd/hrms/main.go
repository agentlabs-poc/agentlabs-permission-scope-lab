// Command hrms serves the example application against an Auth service it reaches
// over HTTP.
//
// It is given a URL where the in-process demonstration is given a database path,
// and that is the whole difference. This binary links no part of the authority
// domain: `go list -deps` finds authmiddleware and authclient and nothing else,
// which is the claim the split exists to make.
package main

import (
	"agentlabs.local/apps/hrms"
	"agentlabs.local/authclient"
	"agentlabs.local/authmiddleware"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type clock struct{}

func (clock) Now() time.Time { return time.Now().UTC() }

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, diag *os.File) int {
	flags := map[string]string{}
	for i := 0; i+1 < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			flags[args[i]] = args[i+1]
		}
	}
	auth, tenant, application := flags["--auth"], flags["--tenant"], flags["--app"]
	human, listen, credential := flags["--human"], flags["--listen"], flags["--client"]
	if auth == "" || tenant == "" || application == "" || human == "" || credential == "" {
		fmt.Fprintln(diag, "usage: hrms --auth URL --tenant ID --app ID --human ID --client ID [--listen ADDR]")
		return 2
	}
	if listen == "" {
		listen = "127.0.0.1:8081"
	}
	// The application asks as itself about whichever human the request carries.
	// --human is the lab standing in for authentication, which is the
	// application's own business and not Auth's.
	source, err := authclient.New(auth, authclient.Credential{Type: "service_account", ID: credential}, nil)
	if err != nil {
		fmt.Fprintln(diag, err)
		return 2
	}
	evaluator, err := authmiddleware.New(source, clock{})
	if err != nil {
		fmt.Fprintln(diag, err)
		return 4
	}
	handler, err := hrms.NewHandler(hrms.NewStore(hrms.DefaultRecords()), evaluator,
		hrms.TrustedIdentity(tenant, application, human))
	if err != nil {
		fmt.Fprintln(diag, err)
		return 4
	}
	fmt.Fprintf(out, "hrms listening on %s, asking %s as %s\n", listen, auth, credential)
	server := &http.Server{Addr: listen, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(diag, err)
		return 4
	}
	return 0
}
