package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"
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
	definition.AllowedTokens = slices.Clone(definition.AllowedTokens)
	err := provider.UpdateCatalog(ctx, app, func(catalog domain.Catalog) (storage.CatalogWriteSet, error) {
		if catalog.ApplicationID != app.ID() {
			return storage.CatalogWriteSet{}, domain.ErrRejected
		}
		evidence := definition
		evidence.AllowedTokens = slices.Clone(definition.AllowedTokens)
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
		result.AllowedTokens = slices.Clone(definition.AllowedTokens)
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
		value.AllowedTokens = slices.Clone(value.AllowedTokens)
		result.Scopes[key] = value
	}
	result.SupportedKeys = make(map[string][]string, len(source.SupportedKeys))
	for key, value := range source.SupportedKeys {
		result.SupportedKeys[key] = append([]string(nil), value...)
	}
	return result
}

// defaultPermissionPage and maxPermissionPage bound a catalog listing. A listing
// is always bounded: there is no unbounded page and no query interface.
const (
	defaultPermissionPage = 100
	maxPermissionPage     = 500
)

// GetPermission returns one registered definition by exact identifier. A retired
// definition is returned with Active false rather than hidden: a caller naming a
// specific identifier is entitled to learn that it exists and is retired, which
// is how Q-126's permanence is observable.
func (s *Service) GetPermission(ctx context.Context, app domain.Application, identity domain.Identity, id string) (domain.PermissionDefinition, error) {
	fail := func(err error) (domain.PermissionDefinition, error) { return domain.PermissionDefinition{}, err }
	provider, admin, err := s.permissionCatalog(ctx, app, identity)
	if err != nil {
		return fail(err)
	}
	if err = codec.PermissionList([]string{id}); err != nil {
		return fail(err)
	}
	var result domain.PermissionDefinition
	err = provider.ReadCatalog(ctx, app, func(catalog domain.Catalog) error {
		if catalog.ApplicationID != app.ID() {
			return domain.ErrRejected
		}
		if err := admin.CheckPermissionRead(ctx, app, identity, s.clock.Now()); err != nil {
			return err
		}
		definition, ok := catalog.Permissions[id]
		if !ok || definition.ID != id {
			return domain.ErrNotFound
		}
		result = definition
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return result, nil
}

// ListPermissions returns one bounded page of an application's catalog, ordered
// by identifier. The filter supports only a prefix, an active-only flag and a
// cursor: no wildcard, no expression, no alternative ordering.
func (s *Service) ListPermissions(ctx context.Context, app domain.Application, identity domain.Identity, filter domain.PermissionFilter) (domain.PermissionPage, error) {
	fail := func(err error) (domain.PermissionPage, error) { return domain.PermissionPage{}, err }
	provider, admin, err := s.permissionCatalog(ctx, app, identity)
	if err != nil {
		return fail(err)
	}
	if err = validateFilter(filter); err != nil {
		return fail(err)
	}
	limit := filter.Limit
	if limit == 0 {
		limit = defaultPermissionPage
	}

	var page domain.PermissionPage
	err = provider.ReadCatalog(ctx, app, func(catalog domain.Catalog) error {
		if catalog.ApplicationID != app.ID() {
			return domain.ErrRejected
		}
		if err := admin.CheckPermissionRead(ctx, app, identity, s.clock.Now()); err != nil {
			return err
		}
		selected := make([]domain.PermissionDefinition, 0, len(catalog.Permissions))
		for id, definition := range catalog.Permissions {
			if err := ctx.Err(); err != nil {
				return err
			}
			if definition.ID != id {
				return domain.ErrRejected
			}
			if filter.ActiveOnly && !definition.Active {
				continue
			}
			if !strings.HasPrefix(id, filter.Prefix) {
				continue
			}
			if filter.After != "" && id <= filter.After {
				continue
			}
			selected = append(selected, definition)
		}
		sort.Slice(selected, func(i, j int) bool { return selected[i].ID < selected[j].ID })
		if len(selected) > limit {
			selected = selected[:limit]
			page.NextAfter = selected[len(selected)-1].ID
		}
		page.Permissions = selected
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return page, nil
}

// SetPermissionStatus retires or restores a registered permission. Retirement is
// reversible, so this is live state rather than a one-way transition.
//
// Nothing cascades in either direction: grants and assignments referencing the
// identifier are never rewritten, deleted or disabled. Resolution reads the
// active flag before examining any grant, so the flip alone stops — or resumes —
// every future evaluation across every tenant.
func (s *Service) SetPermissionStatus(ctx context.Context, app domain.Application, identity domain.Identity, id string, active bool) (domain.PermissionDefinition, error) {
	fail := func(err error) (domain.PermissionDefinition, error) { return domain.PermissionDefinition{}, err }
	provider, admin, err := s.permissionCatalog(ctx, app, identity)
	if err != nil {
		return fail(err)
	}
	if err = codec.PermissionList([]string{id}); err != nil {
		return fail(err)
	}
	result := domain.PermissionDefinition{ID: id, Active: active}
	err = provider.UpdateCatalog(ctx, app, func(catalog domain.Catalog) (storage.CatalogWriteSet, error) {
		if catalog.ApplicationID != app.ID() {
			return storage.CatalogWriteSet{}, domain.ErrRejected
		}
		if err := admin.CheckPermissionStatus(ctx, app, cloneCatalog(catalog), identity, id, active, s.clock.Now()); err != nil {
			return storage.CatalogWriteSet{}, err
		}
		if err := validation.CheckPermissionStatus(catalog, id, active); err != nil {
			return storage.CatalogWriteSet{}, err
		}
		if err := ctx.Err(); err != nil {
			return storage.CatalogWriteSet{}, err
		}
		changed := result
		return storage.CatalogWriteSet{PermissionStatus: &changed}, nil
	})
	if err != nil {
		return fail(err)
	}
	return result, nil
}

// permissionCatalog resolves the provider and administration seams shared by the
// three permission operations. A provider or adapter that does not support them
// makes the operation unsupported rather than unprotected.
func (s *Service) permissionCatalog(ctx context.Context, app domain.Application, identity domain.Identity) (storage.CatalogProvider, PermissionAdministration, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if err := app.Validate(); err != nil {
		return nil, nil, err
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return nil, nil, err
	}
	provider, ok := s.provider.(storage.CatalogProvider)
	if !ok || nilInterface(provider) {
		return nil, nil, domain.ErrUnsupported
	}
	admin, ok := s.administration.(PermissionAdministration)
	if !ok || nilInterface(admin) {
		return nil, nil, domain.ErrUnsupported
	}
	return provider, admin, nil
}

func validateFilter(filter domain.PermissionFilter) error {
	if filter.Limit < 0 || filter.Limit > maxPermissionPage {
		return domain.ErrMalformed
	}
	if !utf8.ValidString(filter.After) || strings.Contains(filter.After, "*") {
		return domain.ErrMalformed
	}
	// A prefix must end on a noun-segment boundary. Only whole segments have a
	// structural form — equality on leading key slots. A partial segment would
	// fall back to a string match inside one slot, which is the cost the slot
	// decomposition exists to avoid, so it is rejected rather than answered
	// slowly and silently.
	if _, err := codec.NounPrefix(filter.Prefix); err != nil {
		return err
	}
	return nil
}
