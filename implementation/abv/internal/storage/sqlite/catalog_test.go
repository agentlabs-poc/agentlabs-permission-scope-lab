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
	secondTenantArea, _ := domain.NewArea("fi7io4lvkfsw", "hrms")
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
		return storage.CatalogWriteSet{Scope: &domain.ScopeDefinition{Key: "dept"}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	permission := domain.PermissionDefinition{ID: "hrms:payroll:payslip::export", Active: true}
	if err = catalogs.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
		return storage.CatalogWriteSet{Permission: &permission}, nil
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
				permission := domain.PermissionDefinition{ID: "hrms:payroll:payslip::new", Active: true}
				scope := domain.ScopeDefinition{Key: "new"}
				return storage.CatalogWriteSet{Permission: &permission, Scope: &scope}, nil
			})
		},
		"keys alone": func() error {
			return p.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
				return storage.CatalogWriteSet{}, nil
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
	permission := domain.PermissionDefinition{ID: "hrms:payroll:payslip::new", Active: true}
	// A callback that mutates its catalog copy must not influence the store: the
	// write re-reads authoritatively, so the forged entry cannot persist.
	forgedProbe := domain.PermissionDefinition{ID: "hrms:payroll:payslip::probe", Active: true}
	if err = p.UpdateCatalog(t.Context(), app, func(c domain.Catalog) (storage.CatalogWriteSet, error) {
		c.Scopes["invented"] = domain.ScopeDefinition{Key: "invented"}
		return storage.CatalogWriteSet{Permission: &forgedProbe}, nil
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err = p.ReadCatalog(t.Context(), app, func(c domain.Catalog) error {
		if _, forged := c.Scopes["invented"]; forged {
			t.Fatal("a callback's mutation of its own copy reached the store")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
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
			permission := domain.PermissionDefinition{ID: "hrms:payroll:payslip::read", Active: active}
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
				definition := domain.PermissionDefinition{ID: "hrms:payroll:payslip::concurrent", Active: true}
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

// TestPermissionStatusUpdatePersistsAndNeverInserts proves the status write is a
// genuine update against real storage: it flips an existing row, survives a
// reopen, and cannot bring an identifier into existence.
func TestPermissionStatusUpdatePersistsAndNeverInserts(t *testing.T) {
	const id = "hrms:employee:certificate::read"
	path := t.TempDir() + "/authority.db"
	app, _ := domain.NewApplication("hrms")

	provider, err := CreateFixture(t.Context(), path, []storage.Snapshot{contractFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	catalogs, ok := provider.(storage.CatalogProvider)
	if !ok {
		t.Fatal("provider does not support catalogs")
	}

	register := func() error {
		return catalogs.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
			definition := domain.PermissionDefinition{ID: id, Active: true}
			return storage.CatalogWriteSet{Permission: &definition}, nil
		})
	}
	setStatus := func(target string, active bool) error {
		return catalogs.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
			definition := domain.PermissionDefinition{ID: target, Active: active}
			return storage.CatalogWriteSet{PermissionStatus: &definition}, nil
		})
	}
	active := func(p storage.CatalogProvider) bool {
		t.Helper()
		var got bool
		if err := p.ReadCatalog(t.Context(), app, func(c domain.Catalog) error {
			got = c.Permissions[id].Active
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		return got
	}

	if err = register(); err != nil {
		t.Fatalf("register: %v", err)
	}
	if !active(catalogs) {
		t.Fatal("registered permission is not active")
	}

	if err = setStatus(id, false); err != nil {
		t.Fatalf("retire: %v", err)
	}
	if active(catalogs) {
		t.Fatal("retirement did not persist")
	}

	// An identifier that was never registered cannot be created by a status
	// change, which would let a permanent identifier appear without its
	// administrative registration check.
	if err = setStatus("hrms:employee:profile::read", true); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unregistered status err=%v, want ErrNotFound", err)
	}

	// Retirement is reversible.
	if err = setStatus(id, true); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if err = provider.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if !active(reopened.(storage.CatalogProvider)) {
		t.Fatal("restored status did not survive a reopen")
	}
}

// TestPermissionsAreL1Records proves permissions live in the ABV-123 record
// store with their identifier decomposed across key slots, not in a table of
// their own with the identifier as one opaque string.
func TestPermissionsAreL1Records(t *testing.T) {
	const id = "hrms:employee:certificate::read"
	path := t.TempDir() + "/authority.db"
	app, _ := domain.NewApplication("hrms")

	opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{contractFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	catalogs := opened.(storage.CatalogProvider)
	db := opened.(*provider).db

	if err = catalogs.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
		definition := domain.PermissionDefinition{ID: id, Active: true}
		return storage.CatalogWriteSet{Permission: &definition}, nil
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	// The old dedicated table must be gone entirely.
	if err = db.QueryRow(`SELECT 1 FROM sqlite_schema WHERE type='table' AND name='permissions'`).Scan(new(int)); err == nil {
		t.Fatal("the permissions table still exists; permissions did not move")
	}

	// The row is in the record store, with the identifier in slots.
	var tenant, k1, k2, k3, k4, k5, k6, k7, k10, value string
	if err = db.QueryRow(`
		SELECT tenant_id, key1, key2, key3, key4, key5, key6, key7, key10, value
		  FROM abv_l1_records
		 WHERE tenant_id='' AND key1='abv' AND key2='permission' AND key3='hrms'
		   AND key4='employee' AND key5='certificate'`).
		Scan(&tenant, &k1, &k2, &k3, &k4, &k5, &k6, &k7, &k10, &value); err != nil {
		t.Fatalf("record not found in abv_l1_records: %v", err)
	}
	for name, got := range map[string]string{
		"tenant_id": tenant, "key1": k1, "key2": k2,
		"key3": k3, "key4": k4, "key5": k5, "key6": k6, "key7": k7, "key10": k10, "value": value,
	} {
		want := map[string]string{
			"tenant_id": "", // '' not NULL, so the identity key stays usable
			"key1": "abv", "key2": "permission",
			// key3 is the application AND the identifier's first segment — the
			// same fact, stored once. A permission whose first noun is not the
			// application is rejected at registration, which is what makes that
			// true rather than merely hoped for.
			"key3": "hrms",
			// Segments two onward. The first is not repeated here.
			"key4": "employee", "key5": "certificate", "key6": "",
			"key7": "",      // padding is contiguous
			"key10": "read", // the verb is pinned to the last slot, never floating
			"value": `{"active":true}`,
		}[name]
		if got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}

	// And it round-trips: the snapshot rebuilds the identifier from those slots.
	if err = catalogs.ReadCatalog(t.Context(), app, func(c domain.Catalog) error {
		definition, ok := c.Permissions[id]
		if !ok || definition.ID != id || !definition.Active {
			return domain.ErrNotFound
		}
		return nil
	}); err != nil {
		t.Fatalf("identifier did not rebuild from its slots: %v", err)
	}
}

// TestScopesAreL1Records proves scopes moved out of their own table and into the
// record store, with the application in key3 and the key in key4.
func TestScopesAreL1Records(t *testing.T) {
	path := t.TempDir() + "/authority.db"
	app, _ := domain.NewApplication("hrms")
	opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{contractFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	catalogs := opened.(storage.CatalogProvider)
	db := opened.(*provider).db

	if err = catalogs.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
		definition := domain.ScopeDefinition{Key: "region"}
		return storage.CatalogWriteSet{Scope: &definition}, nil
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	if err = db.QueryRow(`SELECT 1 FROM sqlite_schema WHERE type='table' AND name='scope_definitions'`).Scan(new(int)); err == nil {
		t.Fatal("the scope_definitions table still exists; scopes did not move")
	}

	var tenant, k1, k2, k3, k4, k5, value string
	if err = db.QueryRow(`
		SELECT tenant_id, key1, key2, key3, key4, key5, value
		  FROM abv_l1_records
		 WHERE tenant_id='' AND key1='abv' AND key2='scope' AND key3='hrms' AND key4='region'`).
		Scan(&tenant, &k1, &k2, &k3, &k4, &k5, &value); err != nil {
		t.Fatalf("record not found in abv_l1_records: %v", err)
	}
	for name, pair := range map[string][2]string{
		"tenant_id": {tenant, ""}, // an empty tenant is what application-wide means now
		"key1":      {k1, "abv"}, "key2": {k2, "scope"},
		"key3": {k3, "hrms"},   // the application, in every record type
		"key4": {k4, "region"}, // a scope key is flat: one slot, whole
		"key5": {k5, ""},       // nothing follows
		"value": {value, `{}`}, // a scope record's presence is the fact
	} {
		if pair[0] != pair[1] {
			t.Errorf("%s = %q, want %q", name, pair[0], pair[1])
		}
	}

	if err = catalogs.ReadCatalog(t.Context(), app, func(c domain.Catalog) error {
		definition, ok := c.Scopes["region"]
		if !ok || definition.Key != "region" {
			return domain.ErrNotFound
		}
		return nil
	}); err != nil {
		t.Fatalf("scope did not round-trip: %v", err)
	}
}
