package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
)

func (s *Service) RegisterPermission(ctx context.Context, app domain.Application, identity domain.Identity, definition domain.PermissionDefinition, supportedKeys []string) (domain.PermissionDefinition, error) {
	fail := func(err error) (domain.PermissionDefinition, error) { return domain.PermissionDefinition{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if err := app.Validate(); err != nil {
		return fail(err)
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return fail(err)
	}
	provider, ok := s.provider.(storage.CatalogProvider)
	if !ok || nilInterface(provider) {
		return fail(domain.ErrUnsupported)
	}
	admin, ok := s.administration.(CatalogAdministration)
	if !ok || nilInterface(admin) {
		return fail(domain.ErrUnsupported)
	}
	keys := append([]string(nil), supportedKeys...)
	err := provider.UpdateCatalog(ctx, app, func(catalog domain.Catalog) (storage.CatalogWriteSet, error) {
		if catalog.ApplicationID != app.ID() {
			return storage.CatalogWriteSet{}, domain.ErrRejected
		}
		if err := admin.CheckPermissionRegistration(ctx, app, cloneCatalog(catalog), identity, definition, append([]string(nil), keys...), s.clock.Now()); err != nil {
			return storage.CatalogWriteSet{}, err
		}
		if err := validation.CheckPermissionRegistration(catalog, definition, keys); err != nil {
			return storage.CatalogWriteSet{}, err
		}
		if err := ctx.Err(); err != nil {
			return storage.CatalogWriteSet{}, err
		}
		result := definition
		return storage.CatalogWriteSet{Permission: &result, SupportedKeys: append([]string(nil), keys...)}, nil
	})
	if err != nil {
		return fail(err)
	}
	return definition, nil
}

func (s *Service) RegisterScope(ctx context.Context, app domain.Application, identity domain.Identity, definition domain.ScopeDefinition) (domain.ScopeDefinition, error) {
	fail := func(err error) (domain.ScopeDefinition, error) { return domain.ScopeDefinition{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if err := app.Validate(); err != nil {
		return fail(err)
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return fail(err)
	}
	provider, ok := s.provider.(storage.CatalogProvider)
	if !ok || nilInterface(provider) {
		return fail(domain.ErrUnsupported)
	}
	admin, ok := s.administration.(CatalogAdministration)
	if !ok || nilInterface(admin) {
		return fail(domain.ErrUnsupported)
	}
	definition.AllowedTokens = append([]string(nil), definition.AllowedTokens...)
	err := provider.UpdateCatalog(ctx, app, func(catalog domain.Catalog) (storage.CatalogWriteSet, error) {
		if catalog.ApplicationID != app.ID() {
			return storage.CatalogWriteSet{}, domain.ErrRejected
		}
		evidence := definition
		evidence.AllowedTokens = append([]string(nil), definition.AllowedTokens...)
		if err := admin.CheckScopeRegistration(ctx, app, cloneCatalog(catalog), identity, evidence, s.clock.Now()); err != nil {
			return storage.CatalogWriteSet{}, err
		}
		if err := validation.CheckScopeRegistration(catalog, definition); err != nil {
			return storage.CatalogWriteSet{}, err
		}
		if err := ctx.Err(); err != nil {
			return storage.CatalogWriteSet{}, err
		}
		result := definition
		result.AllowedTokens = append([]string(nil), definition.AllowedTokens...)
		return storage.CatalogWriteSet{Scope: &result}, nil
	})
	if err != nil {
		return fail(err)
	}
	return definition, nil
}

func cloneCatalog(source domain.Catalog) domain.Catalog {
	result := source
	result.Permissions = cloneMap(source.Permissions)
	result.Scopes = make(map[string]domain.ScopeDefinition, len(source.Scopes))
	for key, value := range source.Scopes {
		value.AllowedTokens = append([]string(nil), value.AllowedTokens...)
		result.Scopes[key] = value
	}
	result.SupportedKeys = make(map[string][]string, len(source.SupportedKeys))
	for key, value := range source.SupportedKeys {
		result.SupportedKeys[key] = append([]string(nil), value...)
	}
	return result
}
