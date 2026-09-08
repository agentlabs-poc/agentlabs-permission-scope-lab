package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
)

func catalogList(value string, present bool) ([]string, error) {
	if !present {
		return []string{}, nil
	}
	items := strings.Split(value, ",")
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item == "" || item != strings.TrimSpace(item) {
			return nil, domain.ErrMalformed
		}
		if _, ok := seen[item]; ok {
			return nil, domain.ErrRejected
		}
		seen[item] = struct{}{}
	}
	return items, nil
}

func catalog(ctx context.Context, api application.CatalogAPI, app domain.Application, positional []string, flags map[string]string, out, diag io.Writer) int {
	fixture := domain.FixtureContext{Name: flags["--fixture-context"]}
	var rendered bytes.Buffer
	switch positional[0] {
	case "register-permission":
		keys, err := catalogList(flags["--supported-keys"], has(flags, "--supported-keys"))
		if err != nil {
			return report(diag, err)
		}
		definition, err := api.RegisterPermission(ctx, app, fixture, domain.PermissionDefinition{ID: positional[1], Active: true}, keys)
		if err != nil {
			return report(diag, err)
		}
		fmt.Fprintf(&rendered, "internal projection: permission\nid  %s\nactive  %t\n", definition.ID, definition.Active)
	case "register-scope":
		tokens, err := catalogList(flags["--allowed-tokens"], has(flags, "--allowed-tokens"))
		if err != nil {
			return report(diag, err)
		}
		definition, err := api.RegisterScope(ctx, app, fixture, domain.ScopeDefinition{Key: positional[1], AllowedTokens: tokens})
		if err != nil {
			return report(diag, err)
		}
		fmt.Fprintf(&rendered, "internal projection: scope\nkey  %s\nallowed tokens  %s\n", definition.Key, strings.Join(definition.AllowedTokens, ","))
	}
	if _, err := fmt.Fprintln(diag, "LAB ONLY: fixed fixture identity; not authenticated application administration"); err != nil {
		return 4
	}
	return output(out, diag, rendered.String())
}

func has(flags map[string]string, name string) bool { _, ok := flags[name]; return ok }
