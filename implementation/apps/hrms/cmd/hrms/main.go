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
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type clock struct{}

func (clock) Now() time.Time { return time.Now().UTC() }

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, diag io.Writer) int {
	// A flag takes the next argument unless that argument is itself a flag, so a
	// switch with no value is present rather than silently swallowing what
	// follows it — and a flag in final position is not dropped.
	valued := map[string]bool{"--auth": true, "--tenant": true, "--app": true, "--human": true, "--client": true, "--listen": true}
	flags := map[string]string{}
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "--") {
			continue
		}
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			flags[args[i]] = args[i+1]
			i++
			continue
		}
		// A switch, or a flag whose value is missing. The two are told apart by
		// whether the flag is one that takes a value: --listen with nothing after
		// it used to fall back to the default port in silence, against a caller
		// who had said exactly where to listen.
		if valued[args[i]] {
			fmt.Fprintf(diag, "%s needs a value\n", args[i])
			return 2
		}
		flags[args[i]] = ""
	}
	auth, tenant, application := flags["--auth"], flags["--tenant"], flags["--app"]
	human, listen, credential := flags["--human"], flags["--listen"], flags["--client"]
	// The secret is not a flag. A token on a command line is a token in ps and
	// in shell history, and this is the shape the migration copies.
	token := os.Getenv("HRMS_AUTH_TOKEN")
	if auth == "" || tenant == "" || application == "" || human == "" || credential == "" || token == "" {
		fmt.Fprintln(diag, "usage: HRMS_AUTH_TOKEN=... hrms --auth URL --tenant ID --app ID --human ID --client ID [--listen ADDR] [--allow-cleartext]")
		return 2
	}
	if listen == "" {
		listen = "127.0.0.1:8081"
	}
	// The application asks as itself about whichever human the request carries.
	// --human is the lab standing in for authentication, which is the
	// application's own business and not Auth's.
	// Cleartext is asked for, never inferred. Switching the guard off because
	// the URL says http:// would mean a deployment could never hit the guard at
	// all: a typo would silently downgrade, and the bearer and the authority
	// answers would cross the network in the open with nothing said.
	if _, asked := flags["--allow-cleartext"]; asked {
		authclient.AllowCleartext()
	}
	source, err := authclient.New(auth,
		authclient.Credential{Type: "service_account", ID: credential, Bearer: token}, nil)
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
	server := &http.Server{
		Addr: listen, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(diag, err)
		return 4
	}
	return 0
}
