package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"errors"
	"reflect"
	"slices"
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
			f.Snapshot.Catalog.Permissions[payslipExport] = domain.PermissionDefinition{ID: payslipExport, Active: true}
			f.Snapshot.Catalog.Permissions["hrms:payroll:payslip::inactive"] = domain.PermissionDefinition{ID: "hrms:payroll:payslip::inactive"}
			before := cloneRootSnapshot(f.Snapshot)

			root, err := lineage.ResolveParentTeam(f.Snapshot, f.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}], "Team1", now)
			if err != nil {
				t.Fatal(err)
			}
			want := []string{lab.PayslipDelete, payslipExport, lab.PayslipRead, lab.PayslipWrite}
			if !reflect.DeepEqual(root.Permissions, want) || !slices.Contains(root.Permissions, payslipExport) {
				t.Fatalf("root permissions = %v, want %v", root.Permissions, want)
			}
			ordinary, err := lineage.ResolveParentTeam(f.Snapshot, f.Child, "Team2", now)
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
	root, err := lineage.ResolveParentTeam(crm.Snapshot, crm.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}], "Team1", now)
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
	f.Snapshot.Catalog.Permissions[payslipExport] = domain.PermissionDefinition{ID: payslipExport, Active: true}
	g0 := f.Snapshot.Contents[domain.GrantKey{ID: "G0", Revision: 1}]
	g0.Scope = map[string]string{"dept": "FIN"}
	g0.Validity = &domain.Validity{ExpiresAt: &expires}
	f.Snapshot.Contents[domain.GrantKey{ID: "G0", Revision: 1}] = g0
	f.Snapshot.Memberships = append(f.Snapshot.Memberships, domain.Membership{TeamID: "RootTeam", HumanID: "root-user"})
	f.Snapshot.Catalog.SupportedKeys[payslipExport] = []string{"dept"}

	root, err := lineage.ResolveParentTeam(f.Snapshot, f.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}], "Team1", now)
	if err != nil || !reflect.DeepEqual(root.Predicates, []domain.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "G0"}}) || len(root.Validities) != 1 || root.Validities[0].ExpiresAt == nil || !root.Validities[0].ExpiresAt.Equal(expires) {
		t.Fatalf("root shape not preserved: %#v, %v", root, err)
	}
	identity := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "root-user"}, HumanID: "root-user"}
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
			f.Snapshot.Catalog.Permissions[payslipExport] = domain.PermissionDefinition{ID: "other", Active: true}
		}},
		{"untrusted", func(f *lab.TeamFINC17Case) { delete(f.Snapshot.TrustedRoots, "G0") }},
		{"disabled grant", func(f *lab.TeamFINC17Case) {
			c := f.Snapshot.Controls["G0"]
			c.Status = "disabled"
			f.Snapshot.Controls["G0"] = c
		}},
		{"disabled assignment", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["A0"]
			a.Status = "disabled"
			f.Snapshot.Assignments["A0"] = a
		}},
		{"expired", func(f *lab.TeamFINC17Case) {
			expiry := now
			g := f.Snapshot.Contents[domain.GrantKey{ID: "G0", Revision: 1}]
			g.Validity = &domain.Validity{ExpiresAt: &expiry}
			f.Snapshot.Contents[domain.GrantKey{ID: "G0", Revision: 1}] = g
		}},
		{"missing assignment", func(f *lab.TeamFINC17Case) { delete(f.Snapshot.Assignments, "A0") }},
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
				f.Snapshot.Catalog.Permissions[payslipExport] = domain.PermissionDefinition{ID: payslipExport, Active: true}
			}
			tc.edit(&f)
			if _, err := lineage.ResolveParentTeam(f.Snapshot, f.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}], "Team1", now); err == nil {
				t.Fatal("invalid root was accepted")
			}
		})
	}
}

func TestRootCatalogCompatibilityFailsClosed(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	f.Snapshot.Catalog.CompatibilityEnabled = true
	f.Snapshot.Catalog.Permissions[payslipExport] = domain.PermissionDefinition{ID: payslipExport, Active: true}
	g0 := f.Snapshot.Contents[domain.GrantKey{ID: "G0", Revision: 1}]
	g0.Scope = map[string]string{"dept": "FIN"}
	f.Snapshot.Contents[domain.GrantKey{ID: "G0", Revision: 1}] = g0
	if _, err := lineage.ResolveParentTeam(f.Snapshot, f.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}], "Team1", time.Time{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("incompatible computed permission accepted: %v", err)
	}
}

func cloneRootSnapshot(source storage.Snapshot) storage.Snapshot {
	result := source
	result.Catalog.Permissions = cloneRootMap(source.Catalog.Permissions)
	result.Catalog.Scopes = cloneRootMap(source.Catalog.Scopes)
	result.Catalog.SupportedKeys = make(map[string][]string, len(source.Catalog.SupportedKeys))
	for key, value := range source.Catalog.SupportedKeys {
		result.Catalog.SupportedKeys[key] = slices.Clone(value)
	}
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
