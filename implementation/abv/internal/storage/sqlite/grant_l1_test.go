package sqlite

import (
	"agentlabs.local/abv/internal/storage"
	"testing"
)

// The fold is only finished when the old tables are gone and stay gone. Every
// previous fold carried this assertion — role_l1_test.go does the same for
// `roles` — because a table left standing beside the envelope is a second place
// the same fact can live, and the two drift silently.
func TestGrantTablesAreFoldedAwayAndTheRowsAreCanonical(t *testing.T) {
	base := contractFixture(t)
	area := base.Area
	path := t.TempDir() + "/authority.db"
	opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	provider := opened.(*provider)
	defer provider.Close()

	for _, gone := range []string{"grant_controls", "grant_contents", "trusted_roots", "assignments"} {
		if _, err := provider.db.Exec(`SELECT 1 FROM ` + gone + ` LIMIT 1`); err == nil {
			t.Fatalf("the %s table still exists; it should be folded away", gone)
		}
	}

	rows, err := provider.db.Query(`
		SELECT boundary,tenant_id,key1,key2,key3,key4,key5,key6
		  FROM abv_l1_records WHERE key2 IN ('grant','grant_revision')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := 0
	for rows.Next() {
		var boundary, tenant, k1, k2, k3, k4, k5, k6 string
		if err := rows.Scan(&boundary, &tenant, &k1, &k2, &k3, &k4, &k5, &k6); err != nil {
			t.Fatal(err)
		}
		seen++
		// A grant is a tenant record in one application, always. Every read is
		// partitioned this way, so a row that is not cannot be reached at all.
		if boundary != "tenant" || tenant != area.TenantID() || k1 != "abv" || k3 != area.ApplicationID() {
			t.Fatalf("canonical path wrong: boundary=%q tenant=%q key1=%q key3=%q", boundary, tenant, k1, k3)
		}
		if k4 == "" {
			t.Fatalf("the grant id slot is empty: key2=%q", k2)
		}
		switch k2 {
		case "grant":
			// The head occupies a prefix of the slots and stops. Its revisions
			// fill key5; it must not, or the two would collide.
			if k5 != "" {
				t.Fatalf("a head carries a revision slot: %q", k5)
			}
		case "grant_revision":
			if len(k5) != 10 {
				t.Fatalf("the revision slot is not padded: %q", k5)
			}
		}
		if k6 != "" {
			t.Fatalf("key6 is used by neither record type, but holds %q", k6)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if seen == 0 {
		t.Fatal("no grant rows at all; the fixture seeded nothing to assert on")
	}
}
