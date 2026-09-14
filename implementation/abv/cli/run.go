// Package cli provides a reusable command runner without process exits or SQL.
package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func Run(ctx context.Context, args []string, in io.Reader, out, diag io.Writer,
	connect application.Connect, scenarios application.ScenarioRunner, catalogConnect ...application.CatalogConnect) int {
	if len(args) == 1 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
		if _, err := fmt.Fprintln(out, "ABV local testing CLI\nCommands: inspect, check, assign, grant, assignment, role, catalog, scenario\nTenant operations require --tenant ID --app ID; catalog operations require --app ID. No default context."); err != nil {
			return 4
		}
		return 0
	}
	fail := func(code int, msg string) int { fmt.Fprintln(diag, msg); return code }
	if ctx.Err() != nil {
		return fail(4, "command cancelled")
	}
	if len(args) == 0 {
		return fail(2, "command required; use help")
	}
	if len(catalogConnect) > 1 {
		return fail(2, "multiple catalog connectors are unsupported")
	}
	command := args[0]
	switch command {
	case "inspect", "check", "assign", "grant", "assignment", "role", "team", "scenario", "catalog":
	default:
		return fail(2, "unknown command")
	}
	flags := map[string]string{}
	var positional []string
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			continue
		}
		name, value, inline := strings.Cut(arg, "=")
		switch name {
		case "--tenant", "--app", "--db", "--file", "--fixture-context", "--case", "--revision", "--permissions", "--support-assignment",
			"--prefix", "--offset", "--limit", "--active", "--active-only", "--name", "--latest", "--id", "--managed", "--application", "--namespace", "--parent", "--roots", "--human":
		default:
			return fail(2, "unknown flag")
		}
		if _, exists := flags[name]; exists {
			return fail(2, "duplicate flag")
		}
		// --active-only is a presence flag: it carries no value and must not
		// consume the next argument.
		if name == "--active-only" || name == "--latest" || name == "--application" || name == "--roots" {
			if inline {
				return fail(2, "flag takes no value")
			}
			flags[name] = "present"
			continue
		}
		if !inline {
			i++
			if i >= len(args) || strings.HasPrefix(args[i], "--") {
				return fail(2, "missing flag value")
			}
			value = args[i]
		}
		if strings.TrimSpace(value) == "" {
			return fail(2, "empty flag value")
		}
		flags[name] = value
	}
	if command == "catalog" {
		if !catalogVerb(positional) || flags["--app"] == "" || flags["--db"] == "" || flags["--fixture-context"] == "" {
			return fail(2, "catalog requires a supported verb, its argument, application, database and fixture context")
		}
		allowed := []string{"--app", "--db", "--fixture-context"}
		if positional[0] == "register-platform-permission" {
			allowed = append(allowed, "--namespace")
			if empty(flags["--namespace"]) {
				return fail(2, "register-platform-permission requires --namespace")
			}
		}
		switch positional[0] {
		case "list-permissions":
			allowed = append(allowed, "--prefix", "--active-only", "--offset", "--limit")
		case "list-scopes":
			allowed = append(allowed, "--offset", "--limit")
		case "set-permission-status":
			allowed = append(allowed, "--active")
			if flags["--active"] == "" {
				return fail(2, "set-permission-status requires --active")
			}
		}
		if !only(flags, allowed...) {
			return fail(2, "unsupported catalog flag")
		}
		app, err := domain.NewApplication(flags["--app"])
		if err != nil {
			return fail(2, "explicit application required; wildcards are unsupported")
		}
		if len(catalogConnect) == 0 || catalogConnect[0] == nil {
			return fail(5, "catalog support is unavailable")
		}
		api, closeConnection, err := catalogConnect[0](ctx, app, flags["--db"])
		if err != nil {
			return report(diag, err)
		}
		if closeConnection == nil || nilCapability(api) {
			if closeConnection != nil {
				_ = closeConnection()
			}
			return fail(4, "database connection unavailable")
		}
		code := catalog(ctx, api, app, positional, flags, out, diag)
		if err := closeConnection(); err != nil && code == 0 {
			return fail(4, "database close failed")
		}
		return code
	}
	area, err := domain.NewArea(flags["--tenant"], flags["--app"])
	if err != nil {
		return fail(2, "explicit tenant and application required; wildcards are unsupported")
	}
	switch command {
	case "inspect":
		if len(positional) != 2 || !inspectKind(positional[0]) || empty(positional[1]) || !only(flags, "--tenant", "--app", "--db") || flags["--db"] == "" {
			return fail(2, "inspect requires kind and ID")
		}
	case "check":
		if len(positional) != 1 || positional[0] != "assignment" || !only(flags, "--tenant", "--app", "--db", "--file") || flags["--db"] == "" || flags["--file"] == "" {
			return fail(2, "check requires assignment")
		}
	case "assign":
		if len(positional) != 0 || !only(flags, "--tenant", "--app", "--db", "--file", "--fixture-context") || flags["--db"] == "" || flags["--file"] == "" || flags["--fixture-context"] == "" {
			return fail(2, "assign accepts flags only")
		}
	case "grant":
		publish := len(positional) == 1 && positional[0] == "publish" && only(flags, "--tenant", "--app", "--db", "--file", "--support-assignment", "--fixture-context") && flags["--db"] != "" && flags["--file"] != "" && flags["--support-assignment"] != "" && flags["--fixture-context"] != ""
		status := len(positional) == 2 && (positional[0] == "enable" || positional[0] == "disable") && !empty(positional[1]) && only(flags, "--tenant", "--app", "--db", "--fixture-context") && flags["--db"] != "" && flags["--fixture-context"] != ""
		if !publish && !status {
			return fail(2, "grant requires publish flags or enable/disable and ID")
		}
	case "assignment":
		if len(positional) != 2 || (positional[0] != "enable" && positional[0] != "disable") || empty(positional[1]) || !only(flags, "--tenant", "--app", "--db", "--fixture-context") || flags["--db"] == "" || flags["--fixture-context"] == "" {
			return fail(2, command+" requires enable/disable and ID")
		}
	case "role":
		if len(positional) == 0 || flags["--db"] == "" || flags["--fixture-context"] == "" {
			return fail(2, "role requires a verb, database and fixture context")
		}
		switch positional[0] {
		case "publish":
			// No positional id: an id is issued, not chosen. --id names an
			// existing role, which makes the publication a new revision of it.
			// --application ships the role to every tenant; without it the role
			// belongs to the tenant publishing it.
			if len(positional) != 1 || !only(flags, "--tenant", "--app", "--db", "--fixture-context", "--revision", "--permissions", "--name", "--id", "--application") || !has(flags, "--revision") || !has(flags, "--permissions") || empty(flags["--name"]) {
				return fail(2, "role publish requires name, revision, permissions, database and fixture context; --id only to add a revision")
			}
			revision, parseErr := strconv.ParseInt(flags["--revision"], 10, 64)
			permissions, listErr := catalogList(flags["--permissions"], true)
			if parseErr != nil || revision <= 0 || listErr != nil || len(permissions) == 0 {
				return fail(2, "malformed role publication")
			}
			flags["--revision"] = strconv.FormatInt(revision, 10)
		case "get":
			if len(positional) != 2 || empty(positional[1]) || !only(flags, "--tenant", "--app", "--db", "--fixture-context", "--revision") || !has(flags, "--revision") {
				return fail(2, "role get requires ID, revision, database and fixture context")
			}
			revision, parseErr := strconv.ParseInt(flags["--revision"], 10, 64)
			if parseErr != nil || revision <= 0 {
				return fail(2, "malformed role revision")
			}
			flags["--revision"] = strconv.FormatInt(revision, 10)
		case "list":
			if len(positional) != 1 || !only(flags, "--tenant", "--app", "--db", "--fixture-context", "--id", "--name", "--latest", "--offset", "--limit", "--managed") {
				return fail(2, "role list accepts id, name, latest, managed, offset and limit only")
			}
			if has(flags, "--managed") && flags["--managed"] != "tenant" && flags["--managed"] != "application" {
				return fail(2, "--managed is tenant or application")
			}
		default:
			return fail(2, "role requires publish, get or list")
		}
	case "team":
		if len(positional) == 0 || flags["--db"] == "" || flags["--fixture-context"] == "" {
			return fail(2, "team requires a verb, database and fixture context")
		}
		switch positional[0] {
		case "get":
			if len(positional) != 2 || empty(positional[1]) || !only(flags, "--tenant", "--app", "--db", "--fixture-context") {
				return fail(2, "team get requires ID, database and fixture context")
			}
		case "list":
			if len(positional) != 1 || !only(flags, "--tenant", "--app", "--db", "--fixture-context", "--parent", "--roots", "--name", "--offset", "--limit") {
				return fail(2, "team list accepts parent, roots, name, offset and limit only")
			}
		case "members":
			if len(positional) != 1 || !only(flags, "--tenant", "--app", "--db", "--fixture-context", "--id", "--human", "--offset", "--limit") {
				return fail(2, "team members accepts id, human, offset and limit only")
			}
			if empty(flags["--id"]) == empty(flags["--human"]) {
				return fail(2, "team members requires exactly one of --id and --human")
			}
		case "create":
			if len(positional) != 1 || !only(flags, "--tenant", "--app", "--db", "--fixture-context", "--name", "--parent") || empty(flags["--name"]) {
				return fail(2, "team create requires --name, and --parent for a subteam")
			}
		case "reparent":
			if len(positional) != 2 || empty(positional[1]) || !only(flags, "--tenant", "--app", "--db", "--fixture-context", "--parent", "--roots") {
				return fail(2, "team reparent requires ID and either --parent or --roots")
			}
			if empty(flags["--parent"]) == !has(flags, "--roots") {
				return fail(2, "team reparent requires exactly one of --parent and --roots")
			}
		case "delete":
			if len(positional) != 2 || empty(positional[1]) || !only(flags, "--tenant", "--app", "--db", "--fixture-context") {
				return fail(2, "team delete requires ID")
			}
		case "add-member", "remove-member":
			if len(positional) != 1 || !only(flags, "--tenant", "--app", "--db", "--fixture-context", "--id", "--human") || empty(flags["--id"]) || empty(flags["--human"]) {
				return fail(2, "team "+positional[0]+" requires --id and --human")
			}
		default:
			return fail(2, "team requires get, list, members, create, reparent, delete, add-member or remove-member")
		}
	case "scenario":
		if len(positional) != 2 || empty(positional[1]) || (positional[0] != "seed" && positional[0] != "run") || flags["--db"] == "" {
			return fail(2, "scenario requires seed/run and scenario name")
		}
		if positional[0] == "seed" && (!only(flags, "--tenant", "--app", "--db") || flags["--case"] != "") {
			return fail(2, "scenario seed accepts database and context flags only")
		}
		if positional[0] == "run" && (!only(flags, "--tenant", "--app", "--db", "--case") || flags["--case"] == "") {
			return fail(2, "scenario run requires case")
		}
	}
	if command == "scenario" {
		if scenarios == nil {
			return fail(5, "scenario support is unavailable")
		}
		var err error
		if positional[0] == "seed" {
			err = scenarios.Seed(ctx, area, positional[1], flags["--db"])
		} else {
			err = scenarios.Run(ctx, area, positional[1], flags["--case"], flags["--db"])
		}
		if err != nil {
			return report(diag, err)
		}
		if positional[0] == "run" {
			return output(out, diag, "observed expected rejection; assignment was not created\n")
		}
		if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated administration"); err != nil {
			return 4
		}
		return output(out, diag, "scenario seeded\n")
	}
	if connect == nil {
		return fail(5, "database connector is unavailable")
	}
	api, closeConnection, err := connect(ctx, area, flags["--db"])
	if err != nil {
		return report(diag, err)
	}
	if closeConnection == nil {
		return fail(4, "database connection unavailable")
	}
	if api == nil {
		_ = closeConnection()
		return fail(4, "database connection unavailable")
	}
	code := dispatch(ctx, command, positional, flags, in, out, diag, api, area)
	if err := closeConnection(); err != nil && code == 0 {
		return fail(4, "database close failed")
	}
	return code
}

func empty(value string) bool { return strings.TrimSpace(value) == "" }

// catalogVerb reports whether the positional arguments name a supported catalog
// operation. list-permissions takes no argument; the others take exactly one.
func catalogVerb(positional []string) bool {
	if len(positional) == 0 {
		return false
	}
	switch positional[0] {
	case "list-permissions", "list-scopes":
		return len(positional) == 1
	case "register-permission", "register-scope", "get-permission", "get-scope", "set-permission-status",
		"register-platform-permission":
		return len(positional) == 2 && !empty(positional[1])
	}
	return false
}

func inspectKind(kind string) bool {
	switch kind {
	// A permission is read through catalog get-permission, which is
	// application-scoped. Inspect requires a tenant a permission does not have.
	case "scope", "role", "grant", "grant-control", "assignment", "team", "membership":
		return true
	default:
		return false
	}
}

func output(out, diag io.Writer, value string) int {
	if _, err := io.WriteString(out, value); err != nil {
		_, _ = fmt.Fprintln(diag, "output failed")
		return 4
	}
	return 0
}

func report(diag io.Writer, err error) int {
	code, message := 4, "operation unavailable"
	switch {
	case errors.Is(err, domain.ErrMalformed):
		code, message = 2, "malformed input"
	case errors.Is(err, domain.ErrRejected), errors.Is(err, domain.ErrNotFound):
		code, message = 3, "operation rejected or record not found"
	case errors.Is(err, domain.ErrUnsupported):
		code, message = 5, "unsupported operation"
	case errors.Is(err, domain.ErrConflict):
		message = "operation conflict"
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		message = "command cancelled"
	}
	_, _ = fmt.Fprintln(diag, message)
	return code
}

func only(flags map[string]string, allowed ...string) bool {
	for flag := range flags {
		found := false
		for _, candidate := range allowed {
			if flag == candidate {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
