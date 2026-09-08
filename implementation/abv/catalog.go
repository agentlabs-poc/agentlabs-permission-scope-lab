package abv

import (
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

type CatalogAdministration interface {
	CheckPermissionRegistration(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.PermissionDefinition, []string, time.Time) error
	CheckScopeRegistration(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.ScopeDefinition, time.Time) error
}

func (f *Facade) RegisterPermission(ctx context.Context, app domain.Application, identity domain.Identity, definition domain.PermissionDefinition, supportedKeys []string) (domain.PermissionDefinition, error) {
	return f.service.RegisterPermission(ctx, app, identity, definition, supportedKeys)
}

func (f *Facade) RegisterScope(ctx context.Context, app domain.Application, identity domain.Identity, definition domain.ScopeDefinition) (domain.ScopeDefinition, error) {
	return f.service.RegisterScope(ctx, app, identity, definition)
}
