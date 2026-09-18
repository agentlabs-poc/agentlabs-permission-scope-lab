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
//
// Every permission the content selects must be registered and active. This is
// the write-path rule: a grant may not be authored, revised or assigned while it
// names a permission the catalog does not supply.
func CheckContent(area domain.Area, catalog domain.Catalog, g domain.GrantContent, roles map[domain.RoleKey]domain.RoleContent) error {
	_, err := checkContent(area, catalog, g, roles, true)
	return err
}

// SuppliedContent is the read-path counterpart, and it narrows where CheckContent
// refuses. It returns the subset of the content's selection that the catalog
// still supplies, and rejects only when that subset is empty.
//
// Q-125 retires a permission without rewriting the grants that reference it, and
// Q-143 settles what that leaves behind: the retirement withdraws that permission
// and nothing else. A grant selecting read and write, whose write was retired,
// still supplies read — so a route through it narrows rather than closing, and a
// route hanging beneath it that never selected write is untouched. It closes only
// when nothing it selects survives.
//
// The returned slice is the content's effective permission set. It is what a
// resolved route carries, and it can be narrower than the stored record. That
// divergence is the cost of Q-143 and is deliberate: what a grant gives cannot be
// read from the record alone, only from the record and the catalog together.
func SuppliedContent(area domain.Area, catalog domain.Catalog, g domain.GrantContent, roles map[domain.RoleKey]domain.RoleContent) ([]string, error) {
	return checkContent(area, catalog, g, roles, false)
}

// checkContent is one body for both rules. requireAll distinguishes them: the
// write path needs every selected permission supplied, the read path needs one.
func checkContent(area domain.Area, catalog domain.Catalog, g domain.GrantContent, roles map[domain.RoleKey]domain.RoleContent, requireAll bool) ([]string, error) {
	if err := area.Validate(); err != nil {
		return nil, err
	}
	if catalog.ApplicationID != area.ApplicationID() {
		return nil, domain.ErrRejected
	}
	if err := codec.ValidateContent(g); err != nil {
		return nil, err
	}
	selected, err := selectedPermissions(g, roles)
	if err != nil {
		return nil, err
	}
	supplied := make([]string, 0, len(selected))
	for _, permission := range selected {
		registered, ok := catalog.Permissions[permission]
		if !ok || registered.ID != permission || !registered.Active {
			if requireAll {
				return nil, domain.ErrRejected
			}
			continue
		}
		supplied = append(supplied, permission)
	}
	// A selection that is emptied by the catalog is a rejection: nothing the
	// grant names survives, so there is nothing left to supply.
	//
	// A selection that was empty to begin with is not. A trusted root stores no
	// permission list at all — Q-122 computes its coverage from the catalog at
	// resolve time rather than materializing it — so an empty selection here is
	// the root's ordinary shape, and rejecting it closed every route in the area.
	if len(selected) > 0 && len(supplied) == 0 {
		return nil, domain.ErrRejected
	}
	for key, value := range g.Scope {
		registered, ok := catalog.Scopes[key]
		if !ok || registered.Key != key {
			return nil, domain.ErrRejected
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
			return nil, domain.ErrRejected
		}
	}
	return supplied, nil
}

// selectedPermissions resolves the content's complete, exact permission source.
// Callers must validate the content shape before calling it.
func selectedPermissions(g domain.GrantContent, roles map[domain.RoleKey]domain.RoleContent) ([]string, error) {
	if g.Permissions != nil {
		return g.Permissions, nil
	}
	// Root content names neither source, and selects nothing here: its coverage
	// is the registered catalog, computed at resolution by rootRoute (Q-122).
	// Nothing was named, so there is nothing to check the catalog against — and
	// a root is the only shape that may say so, because ValidateContent admits a
	// sourceless content only with no parent and no local narrowing.
	//
	// Without this, an established root resolved to ErrRejected: the empty role
	// key missed, and the root that every lineage hangs from supported nothing.
	if g.RoleID == "" && g.RoleRevision == 0 {
		return nil, nil
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
