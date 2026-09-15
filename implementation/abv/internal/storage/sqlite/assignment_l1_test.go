package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"testing"
)

// The assignment's key path is the binding, not its own id, so that Q-104 — one
// current assignment per grant and recipient, disabled ones counting — is the
// envelope's primary key rather than a constraint this record type would have to
// add to a table it shares with every other one.
//
// These assert the layout and the rule together, because the rule is only true
// while the layout holds.
func TestAssignmentRowsAreKeyedByTheBindingAndQ104IsThePrimaryKey(t *testing.T) {
	// contractFixture carries a grant but no assignment, so one is seeded here
	// rather than reaching for the lab package, which imports this one.
	base := contractFixture(t)
	area := base.Area
	base.Assignments = map[string]domain.Assignment{
		"fm5b7t4p5iv8": {Version: "1", ID: "fm5b7t4p5iv8", GrantID: "fk3x9r2m5iv8", GrantRevision: 1,
			Recipient: domain.Recipient{Type: "group", ID: "fibggi2juubk"}, Status: "enabled"},
	}
	base.Teams = map[string]domain.Team{"fibggi2juubk": {ID: "fibggi2juubk", Name: "fp8h2w6y5iv8"}}
	path := t.TempDir() + "/authority.db"
	opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	p := opened.(*provider)
	defer p.Close()

	rows, err := p.db.Query(`
		SELECT boundary,tenant_id,key1,key3,key4,key5,key6,key7,value
		  FROM abv_l1_records WHERE key2='assignment'`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := 0
	for rows.Next() {
		var boundary, tenant, k1, k3, k4, k5, k6, k7, value string
		if err := rows.Scan(&boundary, &tenant, &k1, &k3, &k4, &k5, &k6, &k7, &value); err != nil {
			t.Fatal(err)
		}
		seen++
		if boundary != "tenant" || tenant != area.TenantID() || k1 != "abv" || k3 != area.ApplicationID() {
			t.Fatalf("canonical path wrong: boundary=%q tenant=%q key1=%q key3=%q", boundary, tenant, k1, k3)
		}
		if k4 == "" || k6 == "" {
			t.Fatalf("the binding is incomplete: grant=%q recipient=%q", k4, k6)
		}
		if k5 != "user" && k5 != "group" {
			t.Fatalf("recipient type slot holds %q", k5)
		}
		// The id must NOT be in the path. If it were, the primary key would be
		// (grant, recipient, id) and two assignments of one grant to one
		// recipient would both fit — exactly what Q-104 forbids.
		if k7 != "" {
			t.Fatalf("key7 is occupied by %q; the binding must end at key6", k7)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if seen == 0 {
		t.Fatal("no assignment rows; the fixture seeded nothing to assert on")
	}

	// Q-104, enforced by the envelope rather than by this record type: a second
	// assignment of the same grant to the same recipient has nowhere to go.
	conn, err := p.connection(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	existing, err := readAssignmentByBinding(t.Context(), conn, area, "fk3x9r2m5iv8", domain.Recipient{Type: "group", ID: "fibggi2juubk"})
	if err != nil {
		t.Fatal(err)
	}
	duplicate := existing
	duplicate.ID = "a-different-id"
	duplicate.Status = "disabled"
	if err := insertAssignment(t.Context(), conn, area, duplicate); err == nil {
		t.Fatal("a duplicate binding was accepted; Q-104 is not enforced by the key path")
	}

	// And the same binding in another tenant is a different record, not a clash.
	other, err := domain.NewArea("globex", area.ApplicationID())
	if err != nil {
		t.Fatal(err)
	}
	if err := insertAssignment(t.Context(), conn, other, existing); err != nil {
		t.Fatalf("the same binding in another tenant was refused: %v", err)
	}
}
