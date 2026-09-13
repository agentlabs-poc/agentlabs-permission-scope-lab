package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"sort"
	"unicode/utf8"
	"strings"
)

func (s *Service) RegisterPermission(ctx context.Context, app domain.Application, identity domain.Identity, definition domain.PermissionDefinition) (domain.PermissionDefinition, error) {
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
	err := provider.UpdateCatalog(ctx, app, func(catalog domain.Catalog) (storage.CatalogWriteSet, error) {
		if catalog.ApplicationID != app.ID() {
			return storage.CatalogWriteSet{}, domain.ErrRejected
		}
		if err := admin.CheckPermissionRegistration(ctx, app, cloneCatalog(catalog), identity, definition, s.clock.Now()); err != nil {
			return storage.CatalogWriteSet{}, err
		}
		if err := validation.CheckPermissionRegistration(catalog, definition); err != nil {
			return storage.CatalogWriteSet{}, err
		}
		if err := ctx.Err(); err != nil {
			return storage.CatalogWriteSet{}, err
		}
		result := definition
		return storage.CatalogWriteSet{Permission: &result}, nil
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
	err := provider.UpdateCatalog(ctx, app, func(catalog domain.Catalog) (storage.CatalogWriteSet, error) {
		if catalog.ApplicationID != app.ID() {
			return storage.CatalogWriteSet{}, domain.ErrRejected
		}
		evidence := definition
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
	result.Scopes = cloneMap(source.Scopes)
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
		prefix, err := codec.NounPrefix(filter.Prefix)
		if err != nil {
			return err
		}
		matched := make([]domain.PermissionDefinition, 0, len(catalog.Permissions))
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
			if !matchesPrefix(id, prefix) {
				continue
			}
			matched = append(matched, definition)
		}
		// Ordering by identifier is what makes an offset mean the same thing on
		// every call. Without a total order the same offset could name different
		// records between pages.
		sort.Slice(matched, func(i, j int) bool { return matched[i].ID < matched[j].ID })

		page.Total = len(matched)
		page.Generation = catalog.Generation
		if filter.Offset >= len(matched) {
			page.Permissions = []domain.PermissionDefinition{}
			return nil
		}
		end := filter.Offset + limit
		if end > len(matched) {
			end = len(matched)
		}
		selected := matched[filter.Offset:end]
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
	if filter.Offset < 0 {
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


// matchesPrefix reports whether an identifier's noun path starts with the given
// whole segments. Comparison is segment by segment, never a string prefix: only
// whole segments have a structural form in the key slots.
func matchesPrefix(id string, prefix []string) bool {
	if len(prefix) == 0 {
		return true
	}
	key, err := codec.ParsePermission(id)
	if err != nil || len(key.Nouns) < len(prefix) {
		return false
	}
	for i, segment := range prefix {
		if key.Nouns[i] != segment {
			return false
		}
	}
	return true
}

const (
	defaultScopePage = 100
	maxScopePage     = 500
)

// GetScope returns one registered scope definition by exact key.
func (s *Service) GetScope(ctx context.Context, app domain.Application, identity domain.Identity, key string) (domain.ScopeDefinition, error) {
	fail := func(err error) (domain.ScopeDefinition, error) { return domain.ScopeDefinition{}, err }
	provider, admin, err := s.scopeCatalog(ctx, app, identity)
	if err != nil {
		return fail(err)
	}
	if invalidScopeKey(key) {
		return fail(domain.ErrMalformed)
	}
	var result domain.ScopeDefinition
	err = provider.ReadCatalog(ctx, app, func(catalog domain.Catalog) error {
		if catalog.ApplicationID != app.ID() {
			return domain.ErrRejected
		}
		if err := admin.CheckScopeRead(ctx, app, identity, s.clock.Now()); err != nil {
			return err
		}
		definition, ok := catalog.Scopes[key]
		if !ok || definition.Key != key {
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

// ListScopes returns one bounded page of an application's scope catalog, ordered
// by key. There is no prefix filter: a scope key is flat, so a prefix would be a
// string match inside one slot rather than a structural one.
func (s *Service) ListScopes(ctx context.Context, app domain.Application, identity domain.Identity, filter domain.ScopeFilter) (domain.ScopePage, error) {
	fail := func(err error) (domain.ScopePage, error) { return domain.ScopePage{}, err }
	provider, admin, err := s.scopeCatalog(ctx, app, identity)
	if err != nil {
		return fail(err)
	}
	if filter.Offset < 0 || filter.Limit < 0 || filter.Limit > maxScopePage {
		return fail(domain.ErrMalformed)
	}
	limit := filter.Limit
	if limit == 0 {
		limit = defaultScopePage
	}

	var page domain.ScopePage
	err = provider.ReadCatalog(ctx, app, func(catalog domain.Catalog) error {
		if catalog.ApplicationID != app.ID() {
			return domain.ErrRejected
		}
		if err := admin.CheckScopeRead(ctx, app, identity, s.clock.Now()); err != nil {
			return err
		}
		matched := make([]domain.ScopeDefinition, 0, len(catalog.Scopes))
		for key, definition := range catalog.Scopes {
			if err := ctx.Err(); err != nil {
				return err
			}
			if definition.Key != key {
				return domain.ErrRejected
			}
			matched = append(matched, definition)
		}
		sort.Slice(matched, func(i, j int) bool { return matched[i].Key < matched[j].Key })

		page.Total = len(matched)
		page.Generation = catalog.Generation
		if filter.Offset >= len(matched) {
			page.Scopes = []domain.ScopeDefinition{}
			return nil
		}
		end := filter.Offset + limit
		if end > len(matched) {
			end = len(matched)
		}
		page.Scopes = matched[filter.Offset:end]
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return page, nil
}

func (s *Service) scopeCatalog(ctx context.Context, app domain.Application, identity domain.Identity) (storage.CatalogProvider, ScopeAdministration, error) {
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
	admin, ok := s.administration.(ScopeAdministration)
	if !ok || nilInterface(admin) {
		return nil, nil, domain.ErrUnsupported
	}
	return provider, admin, nil
}

func invalidScopeKey(key string) bool {
	return strings.TrimSpace(key) == "" || !utf8.ValidString(key) || strings.Contains(key, "*")
}
