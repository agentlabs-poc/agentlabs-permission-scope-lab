//go:build !race

// Timing assertions are meaningless under the race detector, which is roughly
// fifteen times slower. This file measures; -race proves correctness elsewhere.

package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"fmt"
	"testing"
	"time"
)

// scaleCatalog registers n permissions across a realistic spread of domains and
// verbs, then reports how the catalog behaves at that size.
func scaleCatalog(t *testing.T, n int) (storage.CatalogProvider, domain.Application, int, func()) {
	t.Helper()
	path := t.TempDir() + "/authority.db"
	app, _ := domain.NewApplication("hrms")
	opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{contractFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	catalogs := opened.(storage.CatalogProvider)

	var baseline int
	if err := catalogs.ReadCatalog(t.Context(), app, func(c domain.Catalog) error {
		baseline = len(c.Permissions)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	domains := []string{"employee", "payroll", "leave", "recruitment", "benefits", "timesheet"}
	resources := []string{"certificate", "profile", "record", "request", "policy", "document", "entry", "summary", "report", "attachment"}
	verbs := []string{"read", "write", "approve", "delete", "export", "list", "submit", "reject", "archive", "restore"}

	start := time.Now()
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("hrms:%s:%s::%s",
			domains[i%len(domains)],
			resources[(i/len(domains))%len(resources)],
			verbs[(i/(len(domains)*len(resources)))%len(verbs)])
		// Registration requires Active true — a permission cannot be created
		// pre-retired — so a realistic mix is produced by registering and then
		// retiring, which is also the only path the contract offers.
		definition := domain.PermissionDefinition{ID: id, Active: true, Boundary: domain.ApplicationBoundary}
		if err := catalogs.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
			return storage.CatalogWriteSet{Permission: &definition}, nil
		}); err != nil {
			t.Fatalf("register %d (%s): %v", i, id, err)
		}
		if i%7 == 0 {
			retired := domain.PermissionDefinition{ID: id, Active: false, Boundary: domain.ApplicationBoundary}
			if err := catalogs.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
				return storage.CatalogWriteSet{PermissionStatus: &retired}, nil
			}); err != nil {
				t.Fatalf("retire %d (%s): %v", i, id, err)
			}
		}
	}
	t.Logf("registered %d permissions in %v  (%.2f ms each)",
		n, time.Since(start).Round(time.Millisecond),
		float64(time.Since(start).Microseconds())/float64(n)/1000)

	return catalogs, app, baseline, func() { opened.Close() }
}

func TestCatalogAtSixHundredPermissions(t *testing.T) {
	const n = 600
	catalogs, app, baseline, done := scaleCatalog(t, n)
	defer done()

	var loaded int
	start := time.Now()
	if err := catalogs.ReadCatalog(t.Context(), app, func(c domain.Catalog) error {
		loaded = len(c.Permissions)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	t.Logf("full catalog read: %d permissions in %v", loaded, elapsed.Round(time.Microsecond))
	if loaded != n+baseline {
		t.Fatalf("loaded %d, want %d — identifiers did not round-trip through their slots", loaded, n+baseline)
	}

	// Repeat reads, which is what every authority resolution does.
	start = time.Now()
	const reads = 50
	for i := 0; i < reads; i++ {
		if err := catalogs.ReadCatalog(t.Context(), app, func(domain.Catalog) error { return nil }); err != nil {
			t.Fatal(err)
		}
	}
	per := time.Since(start) / reads
	t.Logf("repeat catalog read: %v each over %d reads", per.Round(time.Microsecond), reads)
	if per > 50*time.Millisecond {
		t.Errorf("catalog read is %v at %d permissions — too slow for a per-resolution read", per, n)
	}
}

// TestGenerationIsOneWritePerChange guards the cost of the counter itself: it
// must not turn one catalog write into several.
func TestGenerationIsOneWritePerChange(t *testing.T) {
	catalogs, app, _, done := scaleCatalog(t, 20)
	defer done()

	var generation int64
	read := func() int64 {
		t.Helper()
		if err := catalogs.ReadCatalog(t.Context(), app, func(c domain.Catalog) error {
			generation = c.Generation
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		return generation
	}

	before := read()
	definition := domain.PermissionDefinition{ID: "hrms:employee:certificate::unique", Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"}
	if err := catalogs.UpdateCatalog(t.Context(), app, func(domain.Catalog) (storage.CatalogWriteSet, error) {
		return storage.CatalogWriteSet{Permission: &definition}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if after := read(); after != before+1 {
		t.Fatalf("generation moved %d -> %d; one write must advance it by exactly one", before, after)
	}
	// A read must not advance it, or every reader would invalidate every cache.
	if again := read(); again != before+1 {
		t.Fatalf("a read advanced the generation: %d -> %d", before+1, again)
	}
}
