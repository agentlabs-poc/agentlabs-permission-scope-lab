package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/sqlite"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type catalogAdmin struct {
	permission func(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.PermissionDefinition, []string, time.Time) error
	scope      func(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.ScopeDefinition, time.Time) error
}

func (catalogAdmin) CheckAssignment(context.Context, storage.Snapshot, domain.Identity, domain.Assignment, time.Time) error {
	return domain.ErrUnsupported
}
func (a catalogAdmin) CheckPermissionRegistration(ctx context.Context, app domain.Application, c domain.Catalog, id domain.Identity, d domain.PermissionDefinition, keys []string, now time.Time) error {
	return a.permission(ctx, app, c, id, d, keys, now)
}
func (a catalogAdmin) CheckScopeRegistration(ctx context.Context, app domain.Application, c domain.Catalog, id domain.Identity, d domain.ScopeDefinition, now time.Time) error {
	return a.scope(ctx, app, c, id, d, now)
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestRegisterCatalogDefinitionsPersistsHostileAdminCannotForgeEvidence(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	app, _ := domain.NewApplication("hrms")
	identity := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "publisher"}, HumanID: "publisher"}
	snapshot := storage.Snapshot{Area: area, Catalog: domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{"read": {ID: "read", Active: true}}, Scopes: map[string]domain.ScopeDefinition{}, SupportedKeys: map[string][]string{}}, Controls: map[string]domain.GrantControl{}, Contents: map[domain.GrantKey]domain.GrantContent{}, Assignments: map[string]domain.Assignment{}, Roles: map[domain.RoleKey]domain.RoleContent{}, Teams: map[string]domain.Team{}, Memberships: []domain.Membership{}, TrustedRoots: map[string]bool{}}
	path := t.TempDir() + "/authority.db"
	provider, err := sqlite.CreateFixture(t.Context(), path, []storage.Snapshot{snapshot})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(123, 0)
	admin := catalogAdmin{
		scope: func(_ context.Context, gotApp domain.Application, c domain.Catalog, gotID domain.Identity, d domain.ScopeDefinition, gotNow time.Time) error {
			if gotApp != app || gotID != identity || d.Key != "dept" || !reflect.DeepEqual(d.AllowedTokens, []string{"$self"}) || gotNow != now {
				return domain.ErrRejected
			}
			c.Scopes["forged"] = domain.ScopeDefinition{Key: "forged", AllowedTokens: []string{}}
			d.AllowedTokens[0] = "forged"
			return nil
		},
		permission: func(_ context.Context, gotApp domain.Application, c domain.Catalog, gotID domain.Identity, d domain.PermissionDefinition, keys []string, gotNow time.Time) error {
			if gotApp != app || gotID != identity || d.ID != "hrms:payroll:payslip::export" || !reflect.DeepEqual(keys, []string{"dept"}) || gotNow != now {
				return domain.ErrRejected
			}
			c.Scopes["dept"] = domain.ScopeDefinition{Key: "dept", AllowedTokens: []string{"forged"}}
			keys[0] = "forged"
			return nil
		},
	}
	service, _ := New(provider, admin, fixedClock{now})
	tokens := []string{"$self"}
	if got, err := service.RegisterScope(t.Context(), app, identity, domain.ScopeDefinition{Key: "dept", AllowedTokens: tokens}); err != nil || got.Key != "dept" {
		t.Fatalf("scope=%#v err=%v", got, err)
	}
	keys := []string{"dept"}
	permission := domain.PermissionDefinition{ID: "hrms:payroll:payslip::export", Active: true}
	if got, err := service.RegisterPermission(t.Context(), app, identity, permission, keys); err != nil || got != permission {
		t.Fatalf("permission=%#v err=%v", got, err)
	}
	if tokens[0] != "$self" || keys[0] != "dept" {
		t.Fatal("caller input mutated")
	}
	if err := provider.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := sqlite.Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	err = reopened.(storage.CatalogProvider).ReadCatalog(t.Context(), app, func(c domain.Catalog) error {
		if !reflect.DeepEqual(c.Scopes["dept"].AllowedTokens, []string{"$self"}) || !reflect.DeepEqual(c.SupportedKeys[permission.ID], []string{"dept"}) || c.Permissions[permission.ID] != permission {
			t.Fatalf("persisted catalog=%#v", c)
		}
		if _, ok := c.Scopes["forged"]; ok {
			t.Fatal("admin forged catalog")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegisterScopePreservesExplicitEmptyAllowedTokens(t *testing.T) {
	app, _ := domain.NewApplication("hrms")
	identity := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "publisher"}, HumanID: "publisher"}
	provider := &catalogFake{catalog: domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{}, Scopes: map[string]domain.ScopeDefinition{}, SupportedKeys: map[string][]string{}}}
	admin := catalogAdmin{
		scope: func(_ context.Context, _ domain.Application, _ domain.Catalog, _ domain.Identity, d domain.ScopeDefinition, _ time.Time) error {
			if d.AllowedTokens == nil {
				t.Fatal("administration received nil tokens")
			}
			return nil
		},
		permission: func(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.PermissionDefinition, []string, time.Time) error {
			return nil
		},
	}
	service, _ := New(provider, admin, fixedClock{})
	input := domain.ScopeDefinition{Key: "region", AllowedTokens: []string{}}
	got, err := service.RegisterScope(t.Context(), app, identity, input)
	if err != nil || got.AllowedTokens == nil || input.AllowedTokens == nil {
		t.Fatalf("scope=%#v input=%#v err=%v", got, input, err)
	}
}

type catalogFake struct {
	catalog  domain.Catalog
	wrongApp bool
	writeErr error
}

func (p *catalogFake) Read(context.Context, domain.Area, func(storage.Snapshot) error) error {
	return nil
}
func (p *catalogFake) Update(context.Context, domain.Area, func(storage.Snapshot) (storage.WriteSet, error)) error {
	return nil
}
func (p *catalogFake) Close() error { return nil }
func (p *catalogFake) ReadCatalog(_ context.Context, _ domain.Application, cb func(domain.Catalog) error) error {
	return cb(cloneCatalog(p.catalog))
}
func (p *catalogFake) UpdateCatalog(_ context.Context, _ domain.Application, cb func(domain.Catalog) (storage.CatalogWriteSet, error)) error {
	c := cloneCatalog(p.catalog)
	if p.wrongApp {
		c.ApplicationID = "other"
	}
	w, err := cb(c)
	if err != nil {
		return err
	}
	if p.writeErr != nil {
		return p.writeErr
	}
	if w.Permission != nil {
		p.catalog.Permissions[w.Permission.ID] = *w.Permission
		p.catalog.SupportedKeys[w.Permission.ID] = append([]string(nil), w.SupportedKeys...)
	}
	if w.Scope != nil {
		p.catalog.Scopes[w.Scope.Key] = *w.Scope
	}
	return nil
}

type oldAdmin struct{}

func (oldAdmin) CheckAssignment(context.Context, storage.Snapshot, domain.Identity, domain.Assignment, time.Time) error {
	return nil
}

type noCatalogProvider struct{}

func (*noCatalogProvider) Read(context.Context, domain.Area, func(storage.Snapshot) error) error {
	return nil
}
func (*noCatalogProvider) Update(context.Context, domain.Area, func(storage.Snapshot) (storage.WriteSet, error)) error {
	return nil
}
func (*noCatalogProvider) Close() error { return nil }

func TestRegisterPermissionRejectsFailuresWithoutResultOrWrite(t *testing.T) {
	app, _ := domain.NewApplication("hrms")
	identity := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "publisher"}, HumanID: "publisher"}
	definition := domain.PermissionDefinition{ID: "hrms:payroll:payslip::export", Active: true}
	base := domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{"read": {ID: "read", Active: true}}, Scopes: map[string]domain.ScopeDefinition{"dept": {Key: "dept", AllowedTokens: []string{"$self"}}}, SupportedKeys: map[string][]string{}}
	approve := catalogAdmin{permission: func(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.PermissionDefinition, []string, time.Time) error {
		return nil
	}}
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()
	duringAdmin, cancelDuringAdmin := context.WithCancel(t.Context())
	cancellingAdmin := catalogAdmin{permission: func(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.PermissionDefinition, []string, time.Time) error {
		cancelDuringAdmin()
		return nil
	}}
	cases := []struct {
		name     string
		provider storage.Provider
		admin    Administration
		ctx      context.Context
		app      domain.Application
		identity domain.Identity
		def      domain.PermissionDefinition
		keys     []string
		want     error
	}{
		{"no catalog administration", &catalogFake{catalog: cloneCatalog(base)}, oldAdmin{}, t.Context(), app, identity, definition, []string{"dept"}, domain.ErrUnsupported},
		{"no catalog provider", &noCatalogProvider{}, approve, t.Context(), app, identity, definition, []string{"dept"}, domain.ErrUnsupported},
		{"invalid application", &catalogFake{catalog: cloneCatalog(base)}, approve, t.Context(), domain.Application{}, identity, definition, []string{"dept"}, domain.ErrMalformed},
		{"malformed identity", &catalogFake{catalog: cloneCatalog(base)}, approve, t.Context(), app, domain.Identity{}, definition, []string{"dept"}, domain.ErrMalformed},
		{"unsupported identity", &catalogFake{catalog: cloneCatalog(base)}, approve, t.Context(), app, domain.Identity{Version: "1", Actor: domain.Actor{Type: "service", ID: "publisher"}, HumanID: "publisher"}, definition, []string{"dept"}, domain.ErrUnsupported},
		{"wrong application evidence", &catalogFake{catalog: cloneCatalog(base), wrongApp: true}, approve, t.Context(), app, identity, definition, []string{"dept"}, domain.ErrRejected},
		{"wrong publisher", &catalogFake{catalog: cloneCatalog(base)}, catalogAdmin{permission: func(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.PermissionDefinition, []string, time.Time) error {
			return domain.ErrRejected
		}}, t.Context(), app, identity, definition, []string{"dept"}, domain.ErrRejected},
		{"invalid registered key", &catalogFake{catalog: cloneCatalog(base)}, approve, t.Context(), app, identity, definition, []string{"missing"}, domain.ErrRejected},
		{"duplicate definition", &catalogFake{catalog: cloneCatalog(base)}, approve, t.Context(), app, identity, domain.PermissionDefinition{ID: "read", Active: true}, nil, domain.ErrConflict},
		{"pre-cancelled", &catalogFake{catalog: cloneCatalog(base)}, approve, cancelled, app, identity, definition, []string{"dept"}, context.Canceled},
		{"cancelled inside administration", &catalogFake{catalog: cloneCatalog(base)}, cancellingAdmin, duringAdmin, app, identity, definition, []string{"dept"}, context.Canceled},
		{"provider write error", &catalogFake{catalog: cloneCatalog(base), writeErr: domain.ErrUnavailable}, approve, t.Context(), app, identity, definition, []string{"dept"}, domain.ErrUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := New(tc.provider, tc.admin, fixedClock{})
			if err != nil {
				t.Fatal(err)
			}
			got, err := s.RegisterPermission(tc.ctx, tc.app, tc.identity, tc.def, tc.keys)
			if !errors.Is(err, tc.want) || got != (domain.PermissionDefinition{}) {
				t.Fatalf("got=%#v err=%v want=%v", got, err, tc.want)
			}
			if p, ok := tc.provider.(*catalogFake); ok {
				if _, exists := p.catalog.Permissions[definition.ID]; exists {
					t.Fatal("failed operation wrote")
				}
			}
		})
	}
}
