// Package cli provides a reusable command runner without process exits or SQL.
// CP1 supports help and validates command/context syntax; all data operations
// are explicitly unsupported until their protected application adapters exist.
package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"context"
	"fmt"
	"io"
	"strings"
)

func Run(ctx context.Context, args []string, in io.Reader, out, diag io.Writer,
	api application.API, scenarios application.ScenarioRunner) int {
	if len(args) == 1 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h") {
		if _, err := fmt.Fprintln(out, "ABV local testing CLI (CP1: help only)\nCommands: inspect, check, assign, scenario\nEvery data operation requires --tenant ID --app ID. No default context."); err != nil {
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
	case "inspect", "check", "assign", "scenario":
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
	if _, err := domain.NewArea(flags["--tenant"], flags["--app"]); err != nil {
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
	return fail(5, "command is not implemented in CP1; no authority was read or written")
}

func empty(value string) bool { return strings.TrimSpace(value) == "" }

func inspectKind(kind string) bool {
	switch kind {
	case "grant", "assignment":
		return true
	default:
		return false
	}
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
