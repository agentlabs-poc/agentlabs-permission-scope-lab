package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/lineage"
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestResolveParentTeamBaseline(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	fixture := lab.TeamFINC17(area)
	if _, exists := fixture.Snapshot.Assignments["fm5b7t4pan0d"]; exists {
		t.Fatal("proposed fm5b7t4pan0d was seeded as established evidence")
	}

	got, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, fixture.Proposed.Recipient.ID, time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	wantPredicates := []domain.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"}}
	if got.Area != area || got.GrantID != "fk3x9r2m5iv8" ||
		!reflect.DeepEqual(got.Permissions, []string{lab.PayslipRead, lab.PayslipWrite}) ||
		!reflect.DeepEqual(got.Predicates, wantPredicates) ||
		!reflect.DeepEqual(got.AssignmentIDs, []string{"fm5b7t4p0dq3", "fm5b7t4p5iv8"}) {
		t.Fatalf("wrong parent support route: %#v", got)
	}
	fixture.Snapshot.Memberships = nil // resolution must never depend on Maya
	if _, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", time.Time{}); err != nil {
		t.Fatalf("issuer membership became a permanent lineage dependency: %v", err)
	}
}

func TestResolveParentTeamRejectsIneligibleOrInferredSupport(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	for _, tc := range []struct {
		name string
		edit func(*lab.TeamFINC17Case)
		want error
	}{
		{"the grant is held only by an unrelated team", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Teams["fp8h2w6y4hu7"] = domain.Team{Name: "team", ID: "fp8h2w6y4hu7", ParentID: "fibggi2jur5s"}
			a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			a.Recipient.ID = "fp8h2w6y4hu7"
			f.Snapshot.Assignments["fm5b7t4p5iv8"] = a
		}, domain.ErrRejected},
		{"fk3x9r2m5iv8 control disabled", func(f *lab.TeamFINC17Case) {
			c := f.Snapshot.Controls["fk3x9r2m5iv8"]
			c.Status = "disabled"
			f.Snapshot.Controls["fk3x9r2m5iv8"] = c
		}, domain.ErrRejected},
		{"fp8h2w6y5iv8 assignment disabled", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			a.Status = "disabled"
			f.Snapshot.Assignments["fm5b7t4p5iv8"] = a
		}, domain.ErrRejected},
		{"fp8h2w6yan0d parent missing", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Teams["fibggi2juxhc"] = domain.Team{ID: "fibggi2juxhc", Name: "fp8h2w6yan0d"}
		}, domain.ErrRejected},
		{"cyclic team graph with disabled edge", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Teams["fibggi2juubk"] = domain.Team{ID: "fibggi2juubk", Name: "fp8h2w6y5iv8", ParentID: "fibggi2juxhc"}
			a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			a.Status = "disabled"
			f.Snapshot.Assignments["fm5b7t4p5iv8"] = a
		}, domain.ErrRejected},
		{"duplicate source binding", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Assignments["fm5b7t4p5iv8-copy"] = domain.Assignment{Version: "1", ID: "fm5b7t4p5iv8-copy", GrantID: "fk3x9r2m5iv8", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "fibggi2juubk"}, Status: "disabled"}
		}, domain.ErrRejected},
		{"assignment map disagrees with record", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Assignments["wrong-key"] = f.Snapshot.Assignments["fm5b7t4p5iv8"]
		}, domain.ErrRejected},
		{"selected child differs from stored revision", func(f *lab.TeamFINC17Case) {
			f.Child.Scope["cert"] = "fi7io4lvkfsw"
		}, domain.ErrRejected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			tc.edit(&fixture)
			got, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", time.Time{})
			if !errors.Is(err, tc.want) || !reflect.DeepEqual(got, domain.Route{}) {
				t.Fatalf("got %#v, %v; want empty route, %v", got, err, tc.want)
			}
		})
	}
}

func TestResolveParentTeamUsesOnlyActualTeam1Revision(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Teams["fp8h2w6y4hu7"] = domain.Team{Name: "team", ID: "fp8h2w6y4hu7", ParentID: "fibggi2jur5s"}
	broader := fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}]
	broader.Revision = 2
	broader.Permissions = []string{lab.PayslipRead, lab.PayslipWrite, lab.PayslipDelete}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 2}] = broader
	fixture.Snapshot.Assignments["fm5b7t4p4hu7"] = domain.Assignment{Version: "1", ID: "fm5b7t4p4hu7", GrantID: "fk3x9r2m5iv8", GrantRevision: 2, Recipient: domain.Recipient{Type: "group", ID: "fp8h2w6y4hu7"}, Status: "enabled"}

	got, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", time.Time{})
	if err != nil || !reflect.DeepEqual(got.Permissions, []string{lab.PayslipRead, lab.PayslipWrite}) || !reflect.DeepEqual(got.AssignmentIDs, []string{"fm5b7t4p0dq3", "fm5b7t4p5iv8"}) {
		t.Fatalf("unrelated broader holding changed route: %#v, %v", got, err)
	}
}

func TestResolveParentTeamValidityBoundaryAndCopies(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	start := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	expiry := start.Add(time.Hour)
	fixture := lab.TeamFINC17(area)
	g0 := fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m0dq3", Revision: 1}]
	g0.Validity = &domain.Validity{NotBefore: &start}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m0dq3", Revision: 1}] = g0
	g1 := fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}]
	g1.Validity = &domain.Validity{ExpiresAt: &expiry}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}] = g1

	got, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", start)
	if err != nil || len(got.Validities) != 2 || got.Validities[0].NotBefore == nil || got.Validities[1].ExpiresAt == nil {
		t.Fatalf("inclusive start or inherited validity failed: %#v, %v", got, err)
	}
	*got.Validities[0].NotBefore = time.Time{}
	*got.Validities[1].ExpiresAt = time.Time{}
	if start.IsZero() || expiry.IsZero() || g0.Validity.NotBefore.IsZero() || g1.Validity.ExpiresAt.IsZero() {
		t.Fatal("route aliases inherited validity")
	}
	if _, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", expiry); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("exact expiry remained eligible: %v", err)
	}
}

func TestResolveParentTeamPreservesExactRootRoleRevision(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	g0 := fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m0dq3", Revision: 1}]
	g0.Permissions = nil
	g0.RoleID, g0.RoleRevision = "root-role", 1
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m0dq3", Revision: 1}] = g0
	fixture.Snapshot.Roles[domain.RoleKey{ID: "root-role", Revision: 1}] = domain.RoleContent{Name: "payslip-reader", ID: "root-role", Revision: 1, Permissions: []string{lab.PayslipRead, lab.PayslipWrite, lab.PayslipDelete}}
	fixture.Snapshot.Roles[domain.RoleKey{ID: "root-role", Revision: 2}] = domain.RoleContent{Name: "payslip-reader", ID: "root-role", Revision: 2, Permissions: []string{lab.PayslipRead}}

	got, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", time.Time{})
	if err != nil || !reflect.DeepEqual(got.Permissions, []string{lab.PayslipRead, lab.PayslipWrite}) {
		t.Fatalf("lost exact root role permissions: %#v, %v", got, err)
	}
}

func TestResolveParentTeamDoesNotUseTrustedRootToSkipTeamCeiling(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Teams["fidfcosw0iyo"] = domain.Team{Name: "team", ID: "fidfcosw0iyo", ParentID: "fibggi2juxhc"}
	rootAssignment := fixture.Snapshot.Assignments["fm5b7t4p0dq3"]
	rootAssignment.Recipient.ID = "fibggi2juxhc"
	fixture.Snapshot.Assignments["fm5b7t4p0dq3"] = rootAssignment
	child := domain.GrantContent{Version: "1", GrantID: "fk3x9r2mfs5i", Revision: 1, ParentGrantID: "fk3x9r2m0dq3", Permissions: []string{lab.PayslipRead}, Scope: map[string]string{}}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2mfs5i", Revision: 1}] = child
	fixture.Snapshot.Controls["fk3x9r2mfs5i"] = domain.GrantControl{Version: "1", ID: "fk3x9r2mfs5i", Status: "enabled"}

	if _, err := lineage.ResolveParentTeam(fixture.Snapshot, child, "fidfcosw0iyo", time.Time{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("trusted root on non-root team skipped the team ceiling: %v", err)
	}
}

func TestResolveParentTeamBoundsTraversalWithoutReturningPartialProof(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	parent := "fibggi2juxhc"
	for i := 0; i < 256; i++ {
		id := "deep-" + strconv.Itoa(i)
		fixture.Snapshot.Teams[parent] = domain.Team{Name: "team", ID: parent, ParentID: id}
		fixture.Snapshot.Teams[id] = domain.Team{Name: "team", ID: id}
		parent = id
	}
	got, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", time.Time{})
	if !errors.Is(err, domain.ErrUnavailable) || !reflect.DeepEqual(got, domain.Route{}) {
		t.Fatalf("unbounded or partial traversal: %#v, %v", got, err)
	}
}
