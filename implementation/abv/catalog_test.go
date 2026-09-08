package abv_test

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"errors"
	"testing"
	"time"
)

type catalogAdministration struct{ allow bool }

func (catalogAdministration) CheckAssignment(context.Context, abv.Evidence, domain.Identity, domain.Assignment, time.Time) error {
	return domain.ErrUnsupported
}
func (a catalogAdministration) CheckPermissionRegistration(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.PermissionDefinition, []string, time.Time) error {
	if !a.allow {
		return domain.ErrRejected
	}
	return nil
}
func (a catalogAdministration) CheckScopeRegistration(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.ScopeDefinition, time.Time) error {
	if !a.allow {
		return domain.ErrRejected
	}
	return nil
}

func TestFacadeForwardsProtectedCatalogRegistration(t *testing.T) {
	app, _ := domain.NewApplication("hrms")
	id := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "publisher"}, HumanID: "publisher"}
	provider := &catalogMemoryProvider{catalog: domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{}, Scopes: map[string]domain.ScopeDefinition{}, SupportedKeys: map[string][]string{}}}
	facade, err := abv.New(provider, catalogAdministration{allow: true}, clock{})
	if err != nil {
		t.Fatal(err)
	}
	definition := domain.ScopeDefinition{Key: "dept", AllowedTokens: []string{"$self"}}
	if got, err := facade.RegisterScope(t.Context(), app, id, definition); err != nil || got.Key != definition.Key {
		t.Fatalf("got=%#v err=%v", got, err)
	}
	denied, _ := abv.New(provider, catalogAdministration{}, clock{})
	got, err := denied.RegisterPermission(t.Context(), app, id, domain.PermissionDefinition{ID: "hrms:payroll:payslip::export", Active: true}, []string{"dept"})
	if !errors.Is(err, domain.ErrRejected) || got != (domain.PermissionDefinition{}) {
		t.Fatalf("denied got=%#v err=%v", got, err)
	}
}

type catalogMemoryProvider struct{ catalog domain.Catalog }

func (*catalogMemoryProvider) Read(context.Context, domain.Area, func(storage.Snapshot) error) error {
	return nil
}
func (*catalogMemoryProvider) Update(context.Context, domain.Area, func(storage.Snapshot) (storage.WriteSet, error)) error {
	return nil
}
func (*catalogMemoryProvider) Close() error { return nil }
func (p *catalogMemoryProvider) ReadCatalog(_ context.Context, _ domain.Application, cb func(domain.Catalog) error) error {
	return cb(p.catalog)
}
func (p *catalogMemoryProvider) UpdateCatalog(_ context.Context, _ domain.Application, cb func(domain.Catalog) (storage.CatalogWriteSet, error)) error {
	w, err := cb(p.catalog)
	if err != nil {
		return err
	}
	if w.Scope != nil {
		p.catalog.Scopes[w.Scope.Key] = *w.Scope
	}
	if w.Permission != nil {
		p.catalog.Permissions[w.Permission.ID] = *w.Permission
	}
	return nil
}
