package storage

import (
	"agentlabs.local/abv/domain"
	"context"
)

type CatalogWriteSet struct {
	Permission    *domain.PermissionDefinition
	SupportedKeys []string
	Scope         *domain.ScopeDefinition
	// PermissionStatus updates an existing definition's active flag. It is
	// distinct from Permission, which inserts a new one: a status change must
	// never create an identifier, and a registration must never overwrite one.
	PermissionStatus *domain.PermissionDefinition
}

type CatalogProvider interface {
	ReadCatalog(context.Context, domain.Application, func(domain.Catalog) error) error
	UpdateCatalog(context.Context, domain.Application, func(domain.Catalog) (CatalogWriteSet, error)) error
}
