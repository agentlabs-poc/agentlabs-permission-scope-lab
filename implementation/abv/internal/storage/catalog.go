package storage

import (
	"agentlabs.local/abv/domain"
	"context"
)

type CatalogWriteSet struct {
	Permission    *domain.PermissionDefinition
	SupportedKeys []string
	Scope         *domain.ScopeDefinition
}

type CatalogProvider interface {
	ReadCatalog(context.Context, domain.Application, func(domain.Catalog) error) error
	UpdateCatalog(context.Context, domain.Application, func(domain.Catalog) (CatalogWriteSet, error)) error
}
