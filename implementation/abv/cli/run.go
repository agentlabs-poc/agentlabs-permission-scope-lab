// Package cli provides a reusable command runner without process exits or SQL.
package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

func Run(ctx context.Context, args []string, in io.Reader, out, diag io.Writer,
	connect application.Connect, scenarios application.ScenarioRunner) int {
	if len(args) == 1 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
		if _, err := fmt.Fprintln(out, "ABV local testing CLI\nCommands: inspect, check, assign, grant, scenario\nEvery data operation requires --tenant ID --app ID. No default context."); err != nil {
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
	command := args[0]
	switch command {
	case "inspect", "check", "assign", "grant", "scenario":
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
		case "--tenant", "--app", "--db", "--file", "--fixture-context", "--case":
		default:
			return fail(2, "unknown flag")
		}
		if _, exists := flags[name]; exists {
			return fail(2, "duplicate flag")
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
		if len(positional) != 2 || (positional[0] != "enable" && positional[0] != "disable") || empty(positional[1]) || !only(flags, "--tenant", "--app", "--db", "--fixture-context") || flags["--db"] == "" || flags["--fixture-context"] == "" {
			return fail(2, "grant requires enable/disable and ID")
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

func inspectKind(kind string) bool {
	switch kind {
	case "permission", "scope", "role", "grant", "grant-control", "assignment", "team", "membership":
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
