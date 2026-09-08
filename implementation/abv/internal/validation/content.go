// Package validation implements pure definition and non-amplification checks.
// A passing check is NOT permission to save, proof of lineage, or an allow result.
package validation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"slices"
	"strings"
)

// CheckContent requires an established area and its application-owned catalog.
// Catalog loading and role lookup must already be partitioned by the provider.
func CheckContent(area domain.Area, catalog domain.Catalog, g domain.GrantContent, roles map[domain.RoleKey]domain.RoleContent) error {
	if err := area.Validate(); err != nil {
		return err
	}
	if catalog.ApplicationID != area.ApplicationID() {
		return domain.ErrRejected
	}
	if err := codec.ValidateContent(g); err != nil {
		return err
	}
	permissions, err := selectedPermissions(g, roles)
	if err != nil {
		return err
	}
	for _, permission := range permissions {
		registered, ok := catalog.Permissions[permission]
		if !ok || registered.ID != permission || !registered.Active {
			return domain.ErrRejected
		}
	}
	for key, value := range g.Scope {
		registered, ok := catalog.Scopes[key]
		if !ok || registered.Key != key {
			return domain.ErrRejected
		}
		if err := selectedTokens(registered.AllowedTokens); err != nil {
			return err
		}
		if strings.HasPrefix(value, "$") && (value != "$self" || !slices.Contains(registered.AllowedTokens, value)) {
			return domain.ErrRejected
		}
		if catalog.CompatibilityEnabled {
			for _, permission := range permissions {
				supported := catalog.SupportedKeys[permission]
				if err := selectedKeys(catalog, supported); err != nil || !slices.Contains(supported, key) {
					return domain.ErrRejected
				}
			}
		}
	}
	return nil
}

// selectedPermissions resolves the content's complete, exact permission source.
// Callers must validate the content shape before calling it.
func selectedPermissions(g domain.GrantContent, roles map[domain.RoleKey]domain.RoleContent) ([]string, error) {
	if g.Permissions != nil {
		return g.Permissions, nil
	}
	role, ok := roles[domain.RoleKey{ID: g.RoleID, Revision: g.RoleRevision}]
	if !ok || role.ID != g.RoleID || role.Revision != g.RoleRevision {
		return nil, domain.ErrRejected
	}
	if err := codec.PermissionList(role.Permissions); err != nil {
		return nil, err
	}
	return role.Permissions, nil
}

func selectedTokens(tokens []string) error {
	seen := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		if token != "$self" || seen[token] {
			return domain.ErrRejected
		}
		seen[token] = true
	}
	return nil
}

func selectedKeys(catalog domain.Catalog, keys []string) error {
	seen := make(map[string]bool, len(keys))
	for _, key := range keys {
		definition, ok := catalog.Scopes[key]
		if !ok || definition.Key != key || seen[key] {
			return domain.ErrRejected
		}
		seen[key] = true
	}
	return nil
}
