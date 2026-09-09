package main

import (
	"agentlabs.local/abv/localadapter"
	"agentlabs.local/authmiddleware"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

type clock struct{}

func (clock) Now() time.Time { return time.Now() }

type values []string

func (v *values) String() string         { return strings.Join(*v, ",") }
func (v *values) Set(value string) error { *v = append(*v, value); return nil }

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("auth-evaluate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var db, tenant, application, human, permission string
	var boundaries, alls values
	flags.StringVar(&db, "db", "", "existing ABV SQLite database")
	flags.StringVar(&tenant, "tenant", "", "trusted test tenant")
	flags.StringVar(&application, "application", "", "trusted test application")
	flags.StringVar(&human, "human", "", "trusted test human")
	flags.StringVar(&permission, "permission", "", "permission")
	flags.Var(&boundaries, "boundary", "exact key=value boundary (repeatable)")
	flags.Var(&alls, "all", "all-values key boundary (repeatable)")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		if err == nil {
			fmt.Fprintln(stderr, "unexpected positional argument")
		}
		return 2
	}
	if empty(db) || empty(tenant) || empty(application) || empty(human) || empty(permission) {
		fmt.Fprintln(stderr, "db, tenant, application, human and permission are required")
		return 2
	}
	material := authmiddleware.Material{}
	for _, raw := range boundaries {
		key, value, ok := strings.Cut(raw, "=")
		if !ok || malformed(key) || malformed(value) || strings.Contains(value, "=") {
			fmt.Fprintln(stderr, "malformed boundary; expected key=value")
			return 2
		}
		if _, duplicate := material[key]; duplicate {
			fmt.Fprintln(stderr, "duplicate material key")
			return 2
		}
		material[key] = authmiddleware.Selection{Kind: authmiddleware.SelectionExact, Value: value}
	}
	for _, key := range alls {
		if malformed(key) || strings.Contains(key, "=") {
			fmt.Fprintln(stderr, "malformed all boundary; expected key")
			return 2
		}
		if _, duplicate := material[key]; duplicate {
			fmt.Fprintln(stderr, "duplicate material key")
			return 2
		}
		material[key] = authmiddleware.Selection{Kind: authmiddleware.SelectionAll}
	}
	now := clock{}
	source, err := localadapter.Open(context.Background(), db, now)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 4
	}
	evaluator, err := authmiddleware.New(source, now)
	var result authmiddleware.Result
	if err == nil {
		result, err = evaluator.Evaluate(context.Background(), authmiddleware.Request{
			Context: authmiddleware.RequestContext{
				Area:     authmiddleware.Area{TenantID: tenant, ApplicationID: application},
				Identity: authmiddleware.Identity{Version: "1", Actor: authmiddleware.Actor{Type: "user", ID: human}, HumanID: human},
			},
			Permission: permission,
			Material:   material,
		})
		if err == nil {
			err = json.NewEncoder(stdout).Encode(result)
		}
	}
	if closeErr := source.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 4
	}
	if result.Decision == authmiddleware.Deny {
		return 3
	}
	return 0
}

func empty(value string) bool { return strings.TrimSpace(value) == "" }

func malformed(value string) bool {
	return empty(value) || value != strings.TrimSpace(value) || !utf8.ValidString(value) || strings.Contains(value, "*")
}
