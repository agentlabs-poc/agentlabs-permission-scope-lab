// Command registry is the lab CLI for the application registry domain.
//
// It exists for the same reason the abv CLI does: so the domain is exercised
// through its contract rather than by touching its tables.
package main

import (
	"agentlabs.local/registry"
	"agentlabs.local/registry/domain"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type clock struct{}

func (clock) Now() time.Time { return time.Now().UTC() }

// labAdmin is the fixture gate. A real deployment holds three separate
// authorities here; the lab admits one operator so the contract can be run.
type labAdmin struct{}

func (labAdmin) CheckApplicationWrite(_ context.Context, i domain.Identity, _ string, _ time.Time) error {
	return allow(i)
}
func (labAdmin) CheckApplicationRead(_ context.Context, i domain.Identity, _ time.Time) error {
	return allow(i)
}
func (labAdmin) CheckInstallationWrite(_ context.Context, i domain.Identity, _, _ string, _ time.Time) error {
	return allow(i)
}
func allow(i domain.Identity) error {
	if i.HumanID != operator.HumanID {
		return domain.ErrRejected
	}
	return nil
}

var operator = domain.Identity{Version: "1", HumanID: "fi7io4lvjqio"}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, diag *os.File) int {
	if len(args) < 2 {
		fmt.Fprintln(diag, "usage: registry <verb> [args] --db PATH")
		return 2
	}
	flags := map[string]string{}
	var positional []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				flags[args[i]] = args[i+1]
				i++
				continue
			}
			flags[args[i]] = "true"
			continue
		}
		positional = append(positional, args[i])
	}
	// Both checks precede Open: a run with no verb used to index positional[0]
	// and panic, after having already created the database file.
	if len(positional) == 0 {
		fmt.Fprintln(diag, "usage: registry <verb> [args] --db PATH")
		return 2
	}
	path := flags["--db"]
	if path == "" {
		fmt.Fprintln(diag, "--db is required")
		return 2
	}
	ctx := context.Background()
	f, err := registry.Open(ctx, path, labAdmin{}, clock{}, true)
	if err != nil {
		return report(diag, err)
	}
	defer f.Close()

	switch positional[0] {
	case "register":
		if len(positional) != 2 || flags["--name"] == "" {
			fmt.Fprintln(diag, "register requires a slug and --name")
			return 2
		}
		app, err := f.RegisterApplication(ctx, operator, positional[1], flags["--name"])
		if err != nil {
			return report(diag, err)
		}
		fmt.Fprintf(out, "application\nslug  %s\nname  %s\nstatus  %s\n", app.Slug, app.Name, app.Status)
	case "status":
		if len(positional) != 2 || flags["--set"] == "" {
			fmt.Fprintln(diag, "status requires a slug and --set")
			return 2
		}
		app, err := f.SetApplicationStatus(ctx, operator, positional[1], flags["--set"])
		if err != nil {
			return report(diag, err)
		}
		fmt.Fprintf(out, "application\nslug  %s\nname  %s\nstatus  %s\n", app.Slug, app.Name, app.Status)
	case "get":
		if len(positional) != 2 {
			fmt.Fprintln(diag, "get requires a slug")
			return 2
		}
		app, err := f.GetApplication(ctx, operator, positional[1])
		if err != nil {
			return report(diag, err)
		}
		fmt.Fprintf(out, "application\nslug  %s\nname  %s\nstatus  %s\n", app.Slug, app.Name, app.Status)
	case "list":
		page, err := f.ListApplications(ctx, operator, domain.ApplicationFilter{Status: flags["--status"]})
		if err != nil {
			return report(diag, err)
		}
		fmt.Fprintf(out, "applications\ncount  %d\ntotal  %d\n", len(page.Applications), page.Total)
		for _, a := range page.Applications {
			fmt.Fprintf(out, "%-12s %-20s %s\n", a.Slug, a.Name, a.Status)
		}
	case "install", "uninstall":
		if len(positional) != 2 || flags["--tenant"] == "" {
			fmt.Fprintf(diag, "%s requires a slug and --tenant\n", positional[0])
			return 2
		}
		op := f.Install
		if positional[0] == "uninstall" {
			op = f.Uninstall
		}
		if err := op(ctx, operator, flags["--tenant"], positional[1]); err != nil {
			return report(diag, err)
		}
		fmt.Fprintf(out, "installation\n%s  tenant=%s  slug=%s\n", positional[0]+"ed", flags["--tenant"], positional[1])
	case "enable", "disable":
		if len(positional) != 2 || flags["--tenant"] == "" {
			fmt.Fprintf(diag, "%s requires a slug and --tenant\n", positional[0])
			return 2
		}
		status := domain.StatusEnabled
		if positional[0] == "disable" {
			status = domain.StatusDisabled
		}
		i, err := f.SetInstallationStatus(ctx, operator, flags["--tenant"], positional[1], status)
		if err != nil {
			return report(diag, err)
		}
		fmt.Fprintf(out, "installation\ntenant  %s\nslug  %s\nstatus  %s\n", i.TenantID, i.Slug, i.Status)
	case "installed":
		if len(positional) != 2 || flags["--tenant"] == "" {
			fmt.Fprintln(diag, "installed requires a slug and --tenant")
			return 2
		}
		held, err := f.IsInstalled(ctx, operator, flags["--tenant"], positional[1])
		if err != nil {
			return report(diag, err)
		}
		fmt.Fprintf(out, "installed  %t\n", held)
	case "installations":
		page, err := f.ListInstallations(ctx, operator, domain.InstallationFilter{TenantID: flags["--tenant"], Slug: flags["--slug"]})
		if err != nil {
			return report(diag, err)
		}
		fmt.Fprintf(out, "installations\ncount  %d\ntotal  %d\n", len(page.Installations), page.Total)
		for _, i := range page.Installations {
			fmt.Fprintf(out, "tenant=%-10s slug=%-10s %s\n", i.TenantID, i.Slug, i.Status)
		}
	default:
		fmt.Fprintln(diag, "unknown verb")
		return 2
	}
	return 0
}

// report maps this domain's errors to exit codes, the same scheme the abv CLI
// uses so a demonstration reads the same way.
//
// It matches on the sentinel, never on the rendered message. Reading an exit
// code out of error text makes the wording load-bearing: rephrasing a message
// silently changes what the shell sees, and any wrapped error that happens to
// contain "not found" is miscoded. The same mistake, made in the composition
// demonstration, reported Auth-AL's own missing record as the registry refusing.
func report(diag *os.File, err error) int {
	// The message is the error's own, so "rejected" and "not found" stay
	// distinguishable on screen even though they share an exit code.
	fmt.Fprintln(diag, err)
	switch {
	case errors.Is(err, domain.ErrMalformed):
		return 2
	case errors.Is(err, domain.ErrRejected), errors.Is(err, domain.ErrNotFound):
		return 3
	case errors.Is(err, domain.ErrConflict):
		return 4
	case errors.Is(err, domain.ErrUnsupported):
		return 5
	}
	return 4
}
