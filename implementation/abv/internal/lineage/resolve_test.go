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
	if _, exists := fixture.Snapshot.Assignments["A2"]; exists {
		t.Fatal("proposed A2 was seeded as established evidence")
	}

	got, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, fixture.Proposed.Recipient.ID, time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	wantPredicates := []domain.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "G1"}}
	if got.Area != area || got.GrantID != "G1" ||
		!reflect.DeepEqual(got.Permissions, []string{lab.PayslipRead, lab.PayslipWrite}) ||
		!reflect.DeepEqual(got.Predicates, wantPredicates) ||
		!reflect.DeepEqual(got.AssignmentIDs, []string{"A0", "A1"}) {
		t.Fatalf("wrong parent support route: %#v", got)
	}
	fixture.Snapshot.Memberships = nil // resolution must never depend on Maya
	if _, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "Team2", time.Time{}); err != nil {
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
		{"G1 only at TeamX", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Teams["TeamX"] = domain.Team{ID: "TeamX", ParentID: "RootTeam"}
			a := f.Snapshot.Assignments["A1"]
			a.Recipient.ID = "TeamX"
			f.Snapshot.Assignments["A1"] = a
		}, domain.ErrRejected},
		{"G1 control disabled", func(f *lab.TeamFINC17Case) {
			c := f.Snapshot.Controls["G1"]
			c.Status = "disabled"
			f.Snapshot.Controls["G1"] = c
		}, domain.ErrRejected},
		{"Team1 assignment disabled", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["A1"]
			a.Status = "disabled"
			f.Snapshot.Assignments["A1"] = a
		}, domain.ErrRejected},
		{"Team2 parent missing", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Teams["Team2"] = domain.Team{ID: "Team2"}
		}, domain.ErrRejected},
		{"cyclic team graph with disabled edge", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Teams["Team1"] = domain.Team{ID: "Team1", ParentID: "Team2"}
			a := f.Snapshot.Assignments["A1"]
			a.Status = "disabled"
			f.Snapshot.Assignments["A1"] = a
		}, domain.ErrRejected},
		{"duplicate source binding", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Assignments["A1-copy"] = domain.Assignment{Version: "1", ID: "A1-copy", GrantID: "G1", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "Team1"}, Status: "disabled"}
		}, domain.ErrRejected},
		{"assignment map disagrees with record", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Assignments["wrong-key"] = f.Snapshot.Assignments["A1"]
		}, domain.ErrRejected},
		{"selected child differs from stored revision", func(f *lab.TeamFINC17Case) {
			f.Child.Scope["cert"] = "other"
		}, domain.ErrRejected},
		{"child self changes recipient binding", func(f *lab.TeamFINC17Case) {
			f.Child.Scope = map[string]string{"user": "$self"}
			f.Snapshot.Contents[domain.GrantKey{ID: "G2", Revision: 1}] = f.Child
		}, domain.ErrUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			tc.edit(&fixture)
			got, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "Team2", time.Time{})
			if !errors.Is(err, tc.want) || !reflect.DeepEqual(got, domain.Route{}) {
				t.Fatalf("got %#v, %v; want empty route, %v", got, err, tc.want)
			}
		})
	}
}

func TestResolveParentTeamUsesOnlyActualTeam1Revision(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Teams["TeamX"] = domain.Team{ID: "TeamX", ParentID: "RootTeam"}
	broader := fixture.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}]
	broader.Revision = 2
	broader.Permissions = []string{lab.PayslipRead, lab.PayslipWrite, lab.PayslipDelete}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 2}] = broader
	fixture.Snapshot.Assignments["AX"] = domain.Assignment{Version: "1", ID: "AX", GrantID: "G1", GrantRevision: 2, Recipient: domain.Recipient{Type: "group", ID: "TeamX"}, Status: "enabled"}

	got, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "Team2", time.Time{})
	if err != nil || !reflect.DeepEqual(got.Permissions, []string{lab.PayslipRead, lab.PayslipWrite}) || !reflect.DeepEqual(got.AssignmentIDs, []string{"A0", "A1"}) {
		t.Fatalf("unrelated broader holding changed route: %#v, %v", got, err)
	}
}

func TestResolveParentTeamValidityBoundaryAndCopies(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	start := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	expiry := start.Add(time.Hour)
	fixture := lab.TeamFINC17(area)
	g0 := fixture.Snapshot.Contents[domain.GrantKey{ID: "G0", Revision: 1}]
	g0.Validity = &domain.Validity{NotBefore: &start}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "G0", Revision: 1}] = g0
	g1 := fixture.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}]
	g1.Validity = &domain.Validity{ExpiresAt: &expiry}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}] = g1

	got, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "Team2", start)
	if err != nil || len(got.Validities) != 2 || got.Validities[0].NotBefore == nil || got.Validities[1].ExpiresAt == nil {
		t.Fatalf("inclusive start or inherited validity failed: %#v, %v", got, err)
	}
	*got.Validities[0].NotBefore = time.Time{}
	*got.Validities[1].ExpiresAt = time.Time{}
	if start.IsZero() || expiry.IsZero() || g0.Validity.NotBefore.IsZero() || g1.Validity.ExpiresAt.IsZero() {
		t.Fatal("route aliases inherited validity")
	}
	if _, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "Team2", expiry); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("exact expiry remained eligible: %v", err)
	}
}

func TestResolveParentTeamPreservesExactRootRoleRevision(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	g0 := fixture.Snapshot.Contents[domain.GrantKey{ID: "G0", Revision: 1}]
	g0.Permissions = nil
	g0.RoleID, g0.RoleRevision = "root-role", 1
	fixture.Snapshot.Contents[domain.GrantKey{ID: "G0", Revision: 1}] = g0
	fixture.Snapshot.Roles[domain.RoleKey{ID: "root-role", Revision: 1}] = domain.RoleContent{ID: "root-role", Revision: 1, Permissions: []string{lab.PayslipRead, lab.PayslipWrite, lab.PayslipDelete}}
	fixture.Snapshot.Roles[domain.RoleKey{ID: "root-role", Revision: 2}] = domain.RoleContent{ID: "root-role", Revision: 2, Permissions: []string{lab.PayslipRead}}

	got, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "Team2", time.Time{})
	if err != nil || !reflect.DeepEqual(got.Permissions, []string{lab.PayslipRead, lab.PayslipWrite}) {
		t.Fatalf("lost exact root role permissions: %#v, %v", got, err)
	}
}

func TestResolveParentTeamDoesNotUseTrustedRootToSkipTeamCeiling(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Teams["Team3"] = domain.Team{ID: "Team3", ParentID: "Team2"}
	rootAssignment := fixture.Snapshot.Assignments["A0"]
	rootAssignment.Recipient.ID = "Team2"
	fixture.Snapshot.Assignments["A0"] = rootAssignment
	child := domain.GrantContent{Version: "1", GrantID: "G3", Revision: 1, ParentGrantID: "G0", Permissions: []string{lab.PayslipRead}, Scope: map[string]string{}}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "G3", Revision: 1}] = child
	fixture.Snapshot.Controls["G3"] = domain.GrantControl{Version: "1", ID: "G3", Status: "enabled"}

	if _, err := lineage.ResolveParentTeam(fixture.Snapshot, child, "Team3", time.Time{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("trusted root on non-root team skipped the team ceiling: %v", err)
	}
}

func TestResolveParentTeamBoundsTraversalWithoutReturningPartialProof(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	parent := "Team2"
	for i := 0; i < 256; i++ {
		id := "deep-" + strconv.Itoa(i)
		fixture.Snapshot.Teams[parent] = domain.Team{ID: parent, ParentID: id}
		fixture.Snapshot.Teams[id] = domain.Team{ID: id}
		parent = id
	}
	got, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "Team2", time.Time{})
	if !errors.Is(err, domain.ErrUnavailable) || !reflect.DeepEqual(got, domain.Route{}) {
		t.Fatalf("unbounded or partial traversal: %#v, %v", got, err)
	}
}
