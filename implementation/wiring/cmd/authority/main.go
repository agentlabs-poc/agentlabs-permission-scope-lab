// Command authority is the composed demonstration: Auth-AL reading a tenant's
// authority, with the two questions about applications answered by the
// application registry rather than by its own tables.
//
// It is the only binary that links both domains, and it is deliberately small —
// the point is to show the seam working, not to be a product.
package main

import (
	"agentlabs.local/abv"
	abvdomain "agentlabs.local/abv/domain"
	"agentlabs.local/registry"
	regdomain "agentlabs.local/registry/domain"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type clock struct{}

func (clock) Now() time.Time { return time.Now().UTC() }

type regAdmin struct{}

func (regAdmin) CheckApplicationWrite(_ context.Context, i regdomain.Identity, _ string, _ time.Time) error {
	return regAllow(i)
}
func (regAdmin) CheckApplicationRead(_ context.Context, i regdomain.Identity, _ time.Time) error {
	return regAllow(i)
}
func (regAdmin) CheckInstallationWrite(_ context.Context, i regdomain.Identity, _, _ string, _ time.Time) error {
	return regAllow(i)
}
func regAllow(i regdomain.Identity) error {
	if i.HumanID != operator.HumanID {
		return regdomain.ErrRejected
	}
	return nil
}

// abvAdmin refuses everything. That is intentional: this demonstration is about
// the registry gate, which runs first, so Auth-AL's own gate is never the reason
// a call fails here.
type abvAdmin struct{}

func (abvAdmin) CheckAssignment(context.Context, abv.Evidence, abvdomain.Identity, abvdomain.Assignment, time.Time) error {
	return abvdomain.ErrRejected
}

var operator = regdomain.Identity{Version: "1", HumanID: "fi7io4lvjqio"}

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: authority read --tenant T --app A --registry PATH --authority PATH")
		return 2
	}
	flags := map[string]string{}
	for i := 0; i < len(args)-1; i++ {
		if strings.HasPrefix(args[i], "--") {
			flags[args[i]] = args[i+1]
		}
	}
	ctx := context.Background()

	// The registry domain, opened on its own store.
	reg, err := registry.Open(ctx, flags["--registry"], regAdmin{}, clock{}, true)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}
	defer reg.Close()

	// Its port. Auth-AL declares this shape; the registry satisfies it.
	port, err := registry.NewPort(reg, operator)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}

	// Auth-AL, told to ask the port instead of its own tables.
	facade, err := abv.OpenSQLiteWithRegistry(ctx, flags["--authority"], abvAdmin{}, clock{}, port)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}
	defer facade.Close()

	area, err := abvdomain.NewArea(flags["--tenant"], flags["--app"])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	// Ask the port directly first, so the demonstration reports which domain
	// actually answered rather than inferring it from an error string. An
	// earlier version matched on "not found" and misattributed Auth-AL's own
	// missing-record error to the registry.
	exists, err := port.ApplicationExists(ctx, area.ApplicationID())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}
	held, err := port.Installed(ctx, area.TenantID(), area.ApplicationID())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}
	fmt.Printf("registry says   application=%v  installed=%v\n", exists, held)

	// Then the actual read, which passes that same gate inside Auth-AL.
	_, err = facade.Inspect(ctx, area, "team", "fibggi2juubk")
	switch {
	case err == nil:
		fmt.Println("auth-al         read the tenant's authority")
		return 0
	case !exists || !held:
		fmt.Println("auth-al         refused at the installation gate")
		return 3
	default:
		// Past the gate. Whatever failed now is Auth-AL's own business.
		fmt.Printf("auth-al         past the gate, then its own: %v\n", err)
		return 4
	}
}
