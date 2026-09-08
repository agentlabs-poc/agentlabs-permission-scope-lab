package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCatalogProviderPersistsAndIsolatesApplicationCatalog(t *testing.T) {
	path := t.TempDir() + "/authority.db"
	hrms := contractFixture(t)
	hrms.Catalog.CompatibilityEnabled = true
	secondTenantArea, _ := domain.NewArea("other", "hrms")
	secondTenant := minimalFixture(secondTenantArea)
	secondTenant.Catalog = hrms.Catalog
	otherArea, _ := domain.NewArea("acme", "accounting")
	other := minimalFixture(otherArea)
	opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{hrms, secondTenant, other})
	if err != nil {
		t.Fatal(err)
	}
	catalogs, ok := opened.(storage.CatalogProvider)
	if !ok {
		t.Fatal("fixture provider lacks catalog capability")
	}
	app, _ := domain.NewApplication("hrms")
	if err = catalogs.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
		return storage.CatalogWriteSet{Scope: &domain.ScopeDefinition{Key: "dept", AllowedTokens: []string{"$self"}}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	permission := domain.PermissionDefinition{ID: "hrms:payroll:payslip::export", Active: true}
	if err = catalogs.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
		return storage.CatalogWriteSet{Permission: &permission, SupportedKeys: []string{"dept"}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	assertCatalogPermission(t, catalogs, app, permission.ID)
	if err := opened.Read(t.Context(), secondTenantArea, func(s storage.Snapshot) error {
		if _, ok := s.Catalog.Permissions[permission.ID]; !ok {
			t.Fatal("shared application catalog did not reach second tenant")
		}
		if !s.Catalog.CompatibilityEnabled {
			t.Fatal("compatibility mode changed")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := opened.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	assertCatalogPermission(t, reopened.(storage.CatalogProvider), app, permission.ID)
	otherApp, _ := domain.NewApplication("accounting")
	if err := reopened.(storage.CatalogProvider).ReadCatalog(t.Context(), otherApp, func(c domain.Catalog) error {
		if _, ok := c.Permissions[permission.ID]; ok {
			t.Fatal("permission crossed application boundary")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestCatalogProviderRejectsInvalidOperationsWithoutWrites(t *testing.T) {
	base := contractFixture(t)
	opened, err := CreateFixture(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	p := opened.(storage.CatalogProvider)
	app, _ := domain.NewApplication("hrms")
	unknown, _ := domain.NewApplication("missing")
	callbackErr := errors.New("stop")
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()
	readCallbackErr := errors.New("read stop")
	cases := map[string]func() error{
		"nil read callback":   func() error { return p.ReadCatalog(t.Context(), app, nil) },
		"nil update callback": func() error { return p.UpdateCatalog(t.Context(), app, nil) },
		"invalid app": func() error {
			return p.ReadCatalog(t.Context(), domain.Application{}, func(domain.Catalog) error { return nil })
		},
		"unknown app": func() error { return p.ReadCatalog(t.Context(), unknown, func(domain.Catalog) error { return nil }) },
		"cancelled":   func() error { return p.ReadCatalog(cancelled, app, func(domain.Catalog) error { return nil }) },
		"callback error": func() error {
			return p.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) { return storage.CatalogWriteSet{}, callbackErr })
		},
		"read callback error": func() error {
			return p.ReadCatalog(t.Context(), app, func(domain.Catalog) error { return readCallbackErr })
		},
		"cancel during read callback": func() error {
			ctx, cancel := context.WithCancel(t.Context())
			return p.ReadCatalog(ctx, app, func(domain.Catalog) error { cancel(); return nil })
		},
		"empty write": func() error {
			return p.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) { return storage.CatalogWriteSet{}, nil })
		},
		"mixed write": func() error {
			return p.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
				permission := domain.PermissionDefinition{ID: "new", Active: true}
				scope := domain.ScopeDefinition{Key: "new", AllowedTokens: []string{}}
				return storage.CatalogWriteSet{Permission: &permission, Scope: &scope}, nil
			})
		},
		"keys alone": func() error {
			return p.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
				return storage.CatalogWriteSet{SupportedKeys: []string{"x"}}, nil
			})
		},
	}
	for name, invoke := range cases {
		t.Run(name, func(t *testing.T) {
			if err := invoke(); err == nil {
				t.Fatal("operation accepted")
			}
		})
	}
	if err := p.ReadCatalog(t.Context(), app, func(c domain.Catalog) error {
		if len(c.Permissions) != 1 || len(c.Scopes) != 0 {
			t.Fatalf("failed operation changed catalog: %#v", c)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestCatalogProviderUsesPersistedEvidenceAndHandlesCancellation(t *testing.T) {
	base := contractFixture(t)
	opened, err := CreateFixture(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	p := opened.(storage.CatalogProvider)
	app, _ := domain.NewApplication("hrms")
	permission := domain.PermissionDefinition{ID: "new", Active: true}
	err = p.UpdateCatalog(t.Context(), app, func(c domain.Catalog) (storage.CatalogWriteSet, error) {
		c.Scopes["invented"] = domain.ScopeDefinition{Key: "invented", AllowedTokens: []string{}}
		return storage.CatalogWriteSet{Permission: &permission, SupportedKeys: []string{"invented"}}, nil
	})
	if !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("trusted callback mutation: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	err = p.UpdateCatalog(ctx, app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
		cancel()
		return storage.CatalogWriteSet{Permission: &permission}, nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("lost callback cancellation: %v", err)
	}
	if err := p.ReadCatalog(t.Context(), app, func(c domain.Catalog) error {
		if _, ok := c.Permissions[permission.ID]; ok {
			t.Fatal("failed write persisted")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestCatalogProviderDuplicateAndConcurrentInsertConflict(t *testing.T) {
	base := contractFixture(t)
	opened, err := CreateFixture(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	p := opened.(storage.CatalogProvider)
	app, _ := domain.NewApplication("hrms")
	for _, active := range []bool{true, false} {
		err := p.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
			permission := domain.PermissionDefinition{ID: "read", Active: active}
			return storage.CatalogWriteSet{Permission: &permission}, nil
		})
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("active=%v duplicate: %v", active, err)
		}
	}
	p1 := opened.(*provider)
	p2Opened, err := Open(t.Context(), p1Path(p1))
	if err != nil {
		t.Fatal(err)
	}
	defer p2Opened.Close()
	providers := []storage.CatalogProvider{p, p2Opened.(storage.CatalogProvider)}
	start := make(chan struct{})
	results := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for _, provider := range providers {
		go func(provider storage.CatalogProvider) {
			ready.Done()
			<-start
			results <- provider.UpdateCatalog(context.Background(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
				definition := domain.PermissionDefinition{ID: "concurrent", Active: true}
				return storage.CatalogWriteSet{Permission: &definition}, nil
			})
		}(provider)
	}
	ready.Wait()
	close(start)
	var successes atomic.Int32
	for range 2 {
		if err := <-results; err == nil {
			successes.Add(1)
		} else if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("unexpected result: %v", err)
		}
	}
	if successes.Load() != 1 {
		t.Fatalf("successes=%d", successes.Load())
	}
}

func TestCatalogProviderSnapshotLimit(t *testing.T) {
	base := contractFixture(t)
	path := t.TempDir() + "/authority.db"
	opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	opened.Close()
	limited, err := OpenWithOptions(t.Context(), path, Options{MaxSnapshotRecords: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer limited.Close()
	app, _ := domain.NewApplication("hrms")
	if err := limited.(storage.CatalogProvider).ReadCatalog(t.Context(), app, func(domain.Catalog) error { return nil }); !errors.Is(err, storage.ErrSnapshotLimit) {
		t.Fatalf("limit not enforced: %v", err)
	}
}

func assertCatalogPermission(t *testing.T, p storage.CatalogProvider, app domain.Application, id string) {
	t.Helper()
	if err := p.ReadCatalog(t.Context(), app, func(c domain.Catalog) error {
		got, ok := c.Permissions[id]
		if !ok || !got.Active {
			t.Fatal("registered permission missing")
		}
		if c.ApplicationID != "hrms" {
			t.Fatal("application boundary changed")
		}
		if len(c.SupportedKeys[id]) != 1 || c.SupportedKeys[id][0] != "dept" {
			t.Fatal("supported keys missing")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func p1Path(p *provider) string {
	var path string
	_ = p.db.QueryRow(`SELECT file FROM pragma_database_list WHERE name='main'`).Scan(&path)
	return path
}
