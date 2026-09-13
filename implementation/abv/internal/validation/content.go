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
		// Two boundaries are implicit and never registered keys:
		//
		//   {}       an empty scope is itself a complete scope — the whole
		//            application boundary, adding no local restriction
		//   $self    the authorizing human, a reserved token the evaluator
		//            resolves at match time
		//
		// A key does not declare either, and no other $-prefixed value exists.
		if strings.HasPrefix(value, domain.ReservedTokenPrefix) && value != domain.SelfToken {
			return domain.ErrRejected
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

// SelectedPermissions returns a copy of the content's complete, exact direct
// or adopted-role permission source after validating its typed shape.
func SelectedPermissions(g domain.GrantContent, roles map[domain.RoleKey]domain.RoleContent) ([]string, error) {
	if err := codec.ValidateContent(g); err != nil {
		return nil, err
	}
	permissions, err := selectedPermissions(g, roles)
	if err != nil {
		return nil, err
	}
	return slices.Clone(permissions), nil
}


