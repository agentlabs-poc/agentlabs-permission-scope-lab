package application

import (
	"agentlabs.local/abv/domain"
	"context"
)

type CatalogAPI interface {
	RegisterPermission(context.Context, domain.Application, domain.FixtureContext, domain.PermissionDefinition, []string) (domain.PermissionDefinition, error)
	RegisterScope(context.Context, domain.Application, domain.FixtureContext, domain.ScopeDefinition) (domain.ScopeDefinition, error)
}

type CatalogConnect func(context.Context, domain.Application, string) (CatalogAPI, func() error, error)
