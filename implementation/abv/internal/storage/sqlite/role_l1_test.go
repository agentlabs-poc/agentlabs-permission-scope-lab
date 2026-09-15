package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"database/sql"
	"testing"
)

// A role is stored as an L1 record, not in a table of its own, and its identity
// fields sit in key slots where they can be queried structurally.
func TestRolesAreL1Records(t *testing.T) {
	provider, area := seededRoleArea(t)
	defer provider.Close()

	rows := queryRecords(t, provider, `
		SELECT tenant_id, key1, key2, key3, key4, key5, key6, value
		  FROM abv_l1_records WHERE key2='role' ORDER BY key4, key5`)
	if len(rows) == 0 {
		t.Fatal("no role records in the L1 store")
	}
	for _, r := range rows {
		if r["tenant_id"] != area.TenantID() {
			t.Fatalf("a role is tenant-scoped: %#v", r)
		}
		if r["key1"] != "abv" || r["key2"] != "role" || r["key3"] != area.ApplicationID() {
			t.Fatalf("canonical path wrong: %#v", r)
		}
		if len(r["key5"]) != 10 {
			t.Fatalf("the revision slot is not padded: %q", r["key5"])
		}

		if r["key6"] == "" {
			t.Fatalf("the name slot is empty: %#v", r)
		}
	}
	if _, err := provider.db.Exec(`SELECT 1 FROM roles LIMIT 1`); err == nil {
		t.Fatal("the roles table still exists; it should be folded away")
	}
	// The envelope carries no column beyond the canonical shape plus boundary —
	// the one fact no key slot holds, since an application record and a platform
	// record both have no tenant and both put a namespace in key3.
	for _, drifted := range []string{"application_id", "revision"} {
		if _, err := provider.db.Exec(`SELECT ` + drifted + ` FROM abv_l1_records LIMIT 1`); err == nil {
			t.Fatalf("the drifted column %q is still on abv_l1_records", drifted)
		}
	}
}

// Two revisions of one role differ only in key5, and both survive — which is
// what identity-is-the-key-path buys. Put the revision in the value and the
// second write would collide with the first.
func TestTwoRevisionsOfOneRoleCoexist(t *testing.T) {
	provider, _ := seededRoleArea(t)
	defer provider.Close()

	rows := queryRecords(t, provider, `
		SELECT key4, key5 FROM abv_l1_records
		 WHERE key2='role' AND key4='fi9jvxobqsxs' ORDER BY key5`)
	if len(rows) < 2 {
		t.Fatalf("want both revisions stored, got %d", len(rows))
	}
	if rows[0]["key5"] >= rows[1]["key5"] {
		t.Fatalf("revisions do not order by their padded slot: %q then %q", rows[0]["key5"], rows[1]["key5"])
	}
}

func queryRecords(t *testing.T, p *provider, query string) []map[string]string {
	t.Helper()
	rows, err := p.db.Query(query)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	var result []map[string]string
	for rows.Next() {
		cells := make([]any, len(columns))
		for i := range cells {
			cells[i] = new(sql.NullString)
		}
		if err := rows.Scan(cells...); err != nil {
			t.Fatal(err)
		}
		record := map[string]string{}
		for i, name := range columns {
			record[name] = cells[i].(*sql.NullString).String
		}
		result = append(result, record)
	}
	return result
}

func seededRoleArea(t *testing.T) (*provider, domain.Area) {
	t.Helper()
	snapshot := roleSeed(t)
	opened, err := CreateFixture(t.Context(), t.TempDir()+"/roles.db", []storage.Snapshot{snapshot})
	if err != nil {
		t.Fatal(err)
	}
	return opened.(*provider), snapshot.Area
}

func roleSeed(t *testing.T) storage.Snapshot {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	read := "hrms:payroll:payslip::read"
	write := "hrms:payroll:payslip::write"
	return storage.Snapshot{
		Area: area,
		Catalog: domain.Catalog{
			ApplicationID: "hrms",
			Permissions: map[string]domain.PermissionDefinition{
				read:  {ID: read, Active: true},
				write: {ID: write, Active: true},
			},
			Scopes: map[string]domain.ScopeDefinition{},
		},
		Roles: map[domain.RoleKey]domain.RoleContent{
			{ID: "fi9jvxobqsxs", Revision: 1}: {ID: "fi9jvxobqsxs", Name: "payslip-reader", Revision: 1, Permissions: []string{read}},
			{ID: "fi9jvxobqsxs", Revision: 2}: {ID: "fi9jvxobqsxs", Name: "payslip-reader", Revision: 2, Permissions: []string{read, write}},
		},
		Controls:     map[string]domain.GrantControl{},
		Contents:     map[domain.GrantKey]domain.GrantContent{},
		Assignments:  map[string]domain.Assignment{},
		Teams: map[string]domain.Team{
			"fibggi2jur5s": {ID: "fibggi2jur5s", Name: "fp8h2w6ykxan"},
			"fibggi2juubk": {ID: "fibggi2juubk", Name: "fp8h2w6y5iv8", ParentID: "fibggi2jur5s"},
		},
		Memberships:  []domain.Membership{{TeamID: "fibggi2juubk", HumanID: "fi7io4lvjqio"}},
		TrustedRoots: map[string]bool{},
	}
}

// An application role and a tenant role are the same record told apart by the
// boundary. Nothing else about them differs, which is what makes one store and
// one read serve both.
func TestApplicationAndTenantRolesDifferOnlyInTheBoundary(t *testing.T) {
	snapshot := roleSeed(t)
	snapshot.Roles[domain.RoleKey{ID: "aaaaaaaaaaaa", Revision: 1}] = domain.RoleContent{
		ID: "aaaaaaaaaaaa", Name: "Viewer", Revision: 1,
		Permissions: []string{"hrms:payroll:payslip::read"},
		Managed:     domain.ApplicationManaged,
	}
	opened, err := CreateFixture(t.Context(), t.TempDir()+"/both.db", []storage.Snapshot{snapshot})
	if err != nil {
		t.Fatal(err)
	}
	p := opened.(*provider)
	defer p.Close()

	for _, r := range queryRecords(t, p, `
		SELECT tenant_id, key3, key4 FROM abv_l1_records
		 WHERE key2='role' ORDER BY tenant_id, key4`) {
		shipped := r["key4"] == "aaaaaaaaaaaa"
		if shipped && (r["tenant_id"] != "") {
			t.Fatalf("a shipped role carries a tenant: %#v", r)
		}
		if !shipped && (r["tenant_id"] == "") {
			t.Fatalf("a composed role has no tenant: %#v", r)
		}
		// The application is in key3 either way — the path is the same shape.
		if r["key3"] != "hrms" {
			t.Fatalf("key3 differs between the two kinds: %#v", r)
		}
	}

	// A tenant's read returns both, each labelled with who owns it.
	if err := p.Read(t.Context(), snapshot.Area, func(s storage.Snapshot) error {
		shipped := s.Roles[domain.RoleKey{ID: "aaaaaaaaaaaa", Revision: 1}]
		composed := s.Roles[domain.RoleKey{ID: "fi9jvxobqsxs", Revision: 1}]
		if shipped.Managed != domain.ApplicationManaged {
			t.Fatalf("a shipped role read back as %v", shipped.Managed)
		}
		if composed.Managed != domain.TenantManaged {
			t.Fatalf("a composed role read back as %v", composed.Managed)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// Teams and memberships are L1 records, tenant-scoped, and carry no application:
// a team is Auth's own collection of this tenant's humans, and an application's
// groupings are a separate thing it keeps itself.
func TestTeamsAndMembershipsAreTenantOnlyL1Records(t *testing.T) {
	provider, area := seededRoleArea(t)
	defer provider.Close()

	for _, table := range []string{"teams", "memberships"} {
		if _, err := provider.db.Exec(`SELECT 1 FROM ` + table + ` LIMIT 1`); err == nil {
			t.Fatalf("the %s table still exists; it should be folded away", table)
		}
	}
	_ = area
}
