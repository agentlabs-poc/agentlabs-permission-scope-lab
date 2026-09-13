package abv

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/mutation"
	"context"
	"time"
)

type CatalogAdministration interface {
	CheckPermissionRegistration(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.PermissionDefinition, time.Time) error
	CheckScopeRegistration(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.ScopeDefinition, time.Time) error
}

func (f *Facade) RegisterPermission(ctx context.Context, app domain.Application, identity domain.Identity, definition domain.PermissionDefinition) (domain.PermissionDefinition, error) {
	return f.service.RegisterPermission(ctx, app, identity, definition)
}

func (f *Facade) RegisterScope(ctx context.Context, app domain.Application, identity domain.Identity, definition domain.ScopeDefinition) (domain.ScopeDefinition, error) {
	return f.service.RegisterScope(ctx, app, identity, definition)
}

// PermissionAdministration gates the read and status operations on an
// application's permission catalog. An adapter that does not implement it makes
// those operations unsupported rather than unprotected.
type PermissionAdministration = mutation.PermissionAdministration

// GetPermission returns one registered definition by exact identifier. A retired
// definition is returned with Active false rather than hidden.
func (f *Facade) GetPermission(ctx context.Context, app domain.Application, identity domain.Identity, id string) (domain.PermissionDefinition, error) {
	return f.service.GetPermission(ctx, app, identity, id)
}

// ListPermissions returns one bounded page of an application's catalog, ordered
// by identifier. The prefix filter is an administrative convenience for grouping
// identifiers; it confers no authority, and evaluation never matches on a prefix.
func (f *Facade) ListPermissions(ctx context.Context, app domain.Application, identity domain.Identity, filter domain.PermissionFilter) (domain.PermissionPage, error) {
	return f.service.ListPermissions(ctx, app, identity, filter)
}

// SetPermissionStatus retires or restores a registered permission. Retirement is
// reversible, so this is live state rather than a one-way transition. Nothing
// cascades: grants and assignments referencing the identifier are never
// rewritten, deleted or disabled.
func (f *Facade) SetPermissionStatus(ctx context.Context, app domain.Application, identity domain.Identity, id string, active bool) (domain.PermissionDefinition, error) {
	return f.service.SetPermissionStatus(ctx, app, identity, id, active)
}

// ScopeAdministration gates reads of an application's scope catalog.
type ScopeAdministration = mutation.ScopeAdministration

// GetScope returns one registered scope definition by exact key.
func (f *Facade) GetScope(ctx context.Context, app domain.Application, identity domain.Identity, key string) (domain.ScopeDefinition, error) {
	return f.service.GetScope(ctx, app, identity, key)
}

// ListScopes returns one bounded page of an application's scope catalog, ordered
// by key. There is no prefix filter: a scope key is a flat token, so a prefix
// would be a string match inside one slot rather than a structural one.
func (f *Facade) ListScopes(ctx context.Context, app domain.Application, identity domain.Identity, filter domain.ScopeFilter) (domain.ScopePage, error) {
	return f.service.ListScopes(ctx, app, identity, filter)
}

// RegisterPlatformPermission registers a permission in a namespace the platform
// defines. See internal/mutation for why the leading-noun rule does not apply.
func (f *Facade) RegisterPlatformPermission(ctx context.Context, namespace string, identity domain.Identity, definition domain.PermissionDefinition) (domain.PermissionDefinition, error) {
	return f.service.RegisterPlatformPermission(ctx, namespace, identity, definition)
}
