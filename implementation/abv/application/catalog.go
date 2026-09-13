package application

import (
	"agentlabs.local/abv/domain"
	"context"
)

type CatalogAPI interface {
	RegisterPermission(context.Context, domain.Application, domain.FixtureContext, domain.PermissionDefinition, []string) (domain.PermissionDefinition, error)
	RegisterScope(context.Context, domain.Application, domain.FixtureContext, domain.ScopeDefinition) (domain.ScopeDefinition, error)
	GetPermission(context.Context, domain.Application, domain.FixtureContext, string) (domain.PermissionDefinition, error)
	ListPermissions(context.Context, domain.Application, domain.FixtureContext, domain.PermissionFilter) (domain.PermissionPage, error)
	SetPermissionStatus(context.Context, domain.Application, domain.FixtureContext, string, bool) (domain.PermissionDefinition, error)
}

type CatalogConnect func(context.Context, domain.Application, string) (CatalogAPI, func() error, error)
