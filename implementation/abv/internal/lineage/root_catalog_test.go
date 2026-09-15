package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

const payslipExport = "hrms:payroll:payslip::export"

func TestRootCatalogComputesActiveApplicationPermissions(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	now := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	for _, tenant := range []string{"acme", "globex"} {
		t.Run(tenant, func(t *testing.T) {
			tenantArea, _ := domain.NewArea(tenant, "hrms")
			f := lab.TeamFINC17(tenantArea)
			f.Snapshot.Catalog.Permissions[payslipExport] = domain.PermissionDefinition{ID: payslipExport, Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"}
			f.Snapshot.Catalog.Permissions["hrms:payroll:payslip::inactive"] = domain.PermissionDefinition{ID: "hrms:payroll:payslip::inactive", Boundary: domain.ApplicationBoundary, Namespace: "hrms"}
			before := cloneRootSnapshot(f.Snapshot)

			root, err := lineage.ResolveParentTeam(f.Snapshot, f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}], "fibggi2juubk", now)
			if err != nil {
				t.Fatal(err)
			}
			want := []string{lab.PayslipDelete, payslipExport, lab.PayslipRead, lab.PayslipWrite}
			if !reflect.DeepEqual(root.Permissions, want) || !slices.Contains(root.Permissions, payslipExport) {
				t.Fatalf("root permissions = %v, want %v", root.Permissions, want)
			}
			ordinary, err := lineage.ResolveParentTeam(f.Snapshot, f.Child, "fibggi2juxhc", now)
			if err != nil || !reflect.DeepEqual(ordinary.Permissions, []string{lab.PayslipRead, lab.PayslipWrite}) {
				t.Fatalf("ordinary route = %#v, %v", ordinary, err)
			}
			if !reflect.DeepEqual(f.Snapshot, before) {
				t.Fatal("resolution mutated stored snapshot")
			}
		})
	}

	crmArea, _ := domain.NewArea("acme", "crm")
	crm := lab.TeamFINC17(crmArea)
	root, err := lineage.ResolveParentTeam(crm.Snapshot, crm.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}], "fibggi2juubk", now)
	if err != nil || slices.Contains(root.Permissions, payslipExport) {
		t.Fatalf("separate application inherited export: %#v, %v", root, err)
	}
	_ = area
}

func TestRootCatalogPreservesScopeValidityAndRevalidatesSource(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	now := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	expires := now.Add(time.Hour)
	f := lab.TeamFINC17(area)
	f.Snapshot.Catalog.Permissions[payslipExport] = domain.PermissionDefinition{ID: payslipExport, Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"}
	g0 := f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m0dq3", Revision: 1}]
	g0.Scope = map[string]string{"dept": "FIN"}
	g0.Validity = &domain.Validity{ExpiresAt: &expires}
	f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m0dq3", Revision: 1}] = g0
	f.Snapshot.Memberships = append(f.Snapshot.Memberships, domain.Membership{TeamID: "fibggi2jur5s", HumanID: "fn2q6v8sbo1e"})

	root, err := lineage.ResolveParentTeam(f.Snapshot, f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}], "fibggi2juubk", now)
	if err != nil || !reflect.DeepEqual(root.Predicates, []domain.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "fk3x9r2m0dq3"}}) || len(root.Validities) != 1 || root.Validities[0].ExpiresAt == nil || !root.Validities[0].ExpiresAt.Equal(expires) {
		t.Fatalf("root shape not preserved: %#v, %v", root, err)
	}
	identity := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fn2q6v8sbo1e"}, HumanID: "fn2q6v8sbo1e"}
	if err := lineage.HasSource(f.Snapshot, identity, root, now); err != nil {
		t.Fatalf("computed root source did not revalidate: %v", err)
	}
}

func TestRootCatalogRejectsInvalidOrIneligibleRoot(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	now := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name string
		edit func(*lab.TeamFINC17Case)
	}{
		{"catalog key mismatch", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Catalog.Permissions[payslipExport] = domain.PermissionDefinition{ID: "hrms:payroll:payslip::other", Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"}
		}},
		{"untrusted", func(f *lab.TeamFINC17Case) { delete(f.Snapshot.TrustedRoots, "fk3x9r2m0dq3") }},
		{"disabled grant", func(f *lab.TeamFINC17Case) {
			c := f.Snapshot.Controls["fk3x9r2m0dq3"]
			c.Status = "disabled"
			f.Snapshot.Controls["fk3x9r2m0dq3"] = c
		}},
		{"disabled assignment", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["fm5b7t4p0dq3"]
			a.Status = "disabled"
			f.Snapshot.Assignments["fm5b7t4p0dq3"] = a
		}},
		{"expired", func(f *lab.TeamFINC17Case) {
			expiry := now
			g := f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m0dq3", Revision: 1}]
			g.Validity = &domain.Validity{ExpiresAt: &expiry}
			f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m0dq3", Revision: 1}] = g
		}},
		{"missing assignment", func(f *lab.TeamFINC17Case) { delete(f.Snapshot.Assignments, "fm5b7t4p0dq3") }},
		{"empty effective catalog", func(f *lab.TeamFINC17Case) {
			for id, d := range f.Snapshot.Catalog.Permissions {
				d.Active = false
				f.Snapshot.Catalog.Permissions[id] = d
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := lab.TeamFINC17(area)
			if tc.name != "catalog key mismatch" && tc.name != "empty effective catalog" {
				f.Snapshot.Catalog.Permissions[payslipExport] = domain.PermissionDefinition{ID: payslipExport, Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"}
			}
			tc.edit(&f)
			if _, err := lineage.ResolveParentTeam(f.Snapshot, f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}], "fibggi2juubk", now); err == nil {
				t.Fatal("invalid root was accepted")
			}
		})
	}
}

func cloneRootSnapshot(source storage.Snapshot) storage.Snapshot {
	result := source
	result.Catalog.Permissions = cloneRootMap(source.Catalog.Permissions)
	result.Catalog.Scopes = cloneRootMap(source.Catalog.Scopes)
	result.Controls = cloneRootMap(source.Controls)
	result.Contents = make(map[domain.GrantKey]domain.GrantContent, len(source.Contents))
	for key, value := range source.Contents {
		value.Permissions, value.Scope = slices.Clone(value.Permissions), cloneRootMap(value.Scope)
		if value.Validity != nil {
			validity := *value.Validity
			if validity.NotBefore != nil {
				v := *validity.NotBefore
				validity.NotBefore = &v
			}
			if validity.ExpiresAt != nil {
				v := *validity.ExpiresAt
				validity.ExpiresAt = &v
			}
			value.Validity = &validity
		}
		result.Contents[key] = value
	}
	result.Assignments, result.Teams, result.TrustedRoots = cloneRootMap(source.Assignments), cloneRootMap(source.Teams), cloneRootMap(source.TrustedRoots)
	result.Roles = make(map[domain.RoleKey]domain.RoleContent, len(source.Roles))
	for key, value := range source.Roles {
		value.Permissions = slices.Clone(value.Permissions)
		result.Roles[key] = value
	}
	result.Memberships = slices.Clone(source.Memberships)
	return result
}

func cloneRootMap[K comparable, V any](source map[K]V) map[K]V {
	result := make(map[K]V, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

// A root's ceiling is the permissions registered under its OWN namespace, and
// this is the reason that matters: an application's catalog is its own
// permissions union every platform one, which is right for evaluation and wrong
// for a ceiling.
//
// Without the slice an application root would carry every auth:* permission —
// including whichever one authorises establishing an application root. The thing
// created by the authority could then create more of that authority.
func TestAnApplicationRootDoesNotCarryPlatformPermissions(t *testing.T) {
	now := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)

	// A platform permission is in every application's catalog. This one is the
	// dangerous shape on purpose: the authority to administer applications.
	const platformAdmin = "auth:tenant:application::admin"
	f.Snapshot.Catalog.Permissions[platformAdmin] = domain.PermissionDefinition{
		ID: platformAdmin, Active: true, Boundary: domain.PlatformBoundary, Namespace: "auth",
	}

	root, err := lineage.ResolveParentTeam(f.Snapshot, f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}], "fibggi2juubk", now)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(root.Permissions, platformAdmin) {
		t.Fatalf("the HRMS root carries %q — holding this root would confer the authority that creates roots", platformAdmin)
	}
	for _, p := range root.Permissions {
		if !strings.HasPrefix(p, "hrms:") {
			t.Fatalf("the HRMS root carries %q, which is not HRMS's", p)
		}
	}

	// And the same catalog, read as the platform's own namespace, yields the
	// other half — the two ceilings are disjoint, which is what keeps the two
	// lineages from becoming one.
	// The Auth root's area names the platform's namespace, not an application —
	// its catalog and its area move together, as they do for any area.
	authArea, err := domain.NewArea("acme", "auth")
	if err != nil {
		t.Fatal(err)
	}
	platform := f.Snapshot
	platform.Area = authArea
	platform.Catalog.ApplicationID = "auth"
	authRoot, err := lineage.ResolveParentTeam(platform, platform.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}], "fibggi2juubk", now)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(authRoot.Permissions, platformAdmin) {
		t.Fatalf("the Auth root does not carry %q; its ceiling is the platform catalog", platformAdmin)
	}
	for _, p := range authRoot.Permissions {
		if slices.Contains(root.Permissions, p) {
			t.Fatalf("%q is in both ceilings; they must be disjoint", p)
		}
	}
}
