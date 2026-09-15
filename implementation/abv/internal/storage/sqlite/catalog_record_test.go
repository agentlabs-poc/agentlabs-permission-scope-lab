package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"testing"
)

// The catalog's own state is a record, and these two facts are the last thing
// that kept Auth-AL holding a table beside the envelope.
func TestCatalogStateIsARecordAndTheLastTwoTablesAreGone(t *testing.T) {
	base := contractFixture(t)
	path := t.TempDir() + "/authority.db"
	opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	p := opened.(*provider)
	defer p.Close()

	for _, gone := range []string{"applications", "installations"} {
		if _, err := p.db.Exec(`SELECT 1 FROM ` + gone + ` LIMIT 1`); err == nil {
			t.Fatalf("the %s table still exists; it should be folded away", gone)
		}
	}

	// Two tables, and that is what the 123 architecture says a domain is: the
	// ownership marker, and one canonical key/value record store.
	rows, err := p.db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(tables) != 2 || tables[0] != "abv_l1_records" || tables[1] != "abv_metadata" {
		t.Fatalf("tables = %v, want exactly abv_l1_records and abv_metadata", tables)
	}

	// The record itself: application boundary, no tenant, the application in
	// key3 and nothing after it.
	var boundary, tenant, k1, k4, value string
	if err := p.db.QueryRow(`
		SELECT boundary,tenant_id,key1,key4,value FROM abv_l1_records
		 WHERE key2='catalog' AND key3=?`, base.Area.ApplicationID()).Scan(&boundary, &tenant, &k1, &k4, &value); err != nil {
		t.Fatalf("no catalog record for the seeded application: %v", err)
	}
	if boundary != "application" || tenant != "" || k1 != "abv" || k4 != "" {
		t.Fatalf("canonical path wrong: boundary=%q tenant=%q key1=%q key4=%q", boundary, tenant, k1, k4)
	}

	// And the generation bump moves it, in the value rather than a column.
	conn, err := p.connection(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	before, err := readCatalogState(t.Context(), conn, base.Area.ApplicationID())
	if err != nil {
		t.Fatal(err)
	}
	if err := bumpCatalogGeneration(t.Context(), conn, base.Area.ApplicationID()); err != nil {
		t.Fatal(err)
	}
	after, err := readCatalogState(t.Context(), conn, base.Area.ApplicationID())
	if err != nil {
		t.Fatal(err)
	}
	if after.Generation != before.Generation+1 {
		t.Fatalf("generation %d -> %d, want one step", before.Generation, after.Generation)
	}
	// The bump must not disturb the declaration sharing the record.
	if after.CompatibilityEnabled != before.CompatibilityEnabled {
		t.Fatal("bumping the generation changed the compatibility declaration")
	}
}

// A platform permission is in every application's catalog, so a write to it
// invalidates every cached view — which is why the bump is not per application.
// The column increment did this in one statement; so does the record update.
func TestPlatformPermissionBumpsEveryCatalogGeneration(t *testing.T) {
	base := contractFixture(t)
	second := base
	otherArea, err := domain.NewArea(base.Area.TenantID(), "payroll")
	if err != nil {
		t.Fatal(err)
	}
	second.Area = otherArea
	second.Catalog = domain.Catalog{
		ApplicationID: "payroll",
		Permissions:   map[string]domain.PermissionDefinition{},
		Scopes:        map[string]domain.ScopeDefinition{},
	}
	path := t.TempDir() + "/authority.db"
	opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{base, second})
	if err != nil {
		t.Fatal(err)
	}
	p := opened.(*provider)
	defer p.Close()
	conn, err := p.connection(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	before := map[string]int64{}
	for _, app := range []string{base.Area.ApplicationID(), "payroll"} {
		state, err := readCatalogState(t.Context(), conn, app)
		if err != nil {
			t.Fatal(err)
		}
		before[app] = state.Generation
	}
	if err := bumpEveryCatalogGeneration(t.Context(), conn); err != nil {
		t.Fatal(err)
	}
	for _, app := range []string{base.Area.ApplicationID(), "payroll"} {
		state, err := readCatalogState(t.Context(), conn, app)
		if err != nil {
			t.Fatal(err)
		}
		if state.Generation != before[app]+1 {
			t.Fatalf("%s generation %d -> %d, want one step", app, before[app], state.Generation)
		}
	}
}
