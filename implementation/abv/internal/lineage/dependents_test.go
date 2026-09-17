package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/lab"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"testing"
)

func TestDependentTeamAssignmentsFindsDirectChild(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments["fm5b7t4pan0d"] = fixture.Proposed
	got, err := lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "fm5b7t4p5iv8")
	if err != nil || len(got) != 1 || got[0] != fixture.Proposed {
		t.Fatalf("dependents = %#v, %v", got, err)
	}
}

func TestDependentTeamAssignmentsUsesUnambiguousBindingKeys(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Teams["fp8h2w6y0dq3"] = domain.Team{Name: "team", ID: "fp8h2w6y0dq3"}
	fixture.Snapshot.Teams["X\x00T"] = domain.Team{Name: "team", ID: "X\x00T"}
	addBinding(&fixture.Snapshot, "fm5b7t4p4hu7", "G\x00X", "", "fp8h2w6y0dq3", "enabled")
	addBinding(&fixture.Snapshot, "fm5b7t4p9mzc", "G", "", "X\x00T", "enabled")
	before := snapshotEvidence(fixture.Snapshot)
	got, err := lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "fm5b7t4p5iv8")
	if err != nil || len(got) != 0 {
		t.Fatalf("dependents = %#v, %v", got, err)
	}
	assertSnapshotUnchanged(t, fixture.Snapshot, before)
}

func TestDependentTeamAssignmentsStructuralMatrix(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	tests := []struct {
		name string
		edit func(*storage.Snapshot)
		want []string
		bad  bool
	}{
		{"three levels and sorted branches", func(s *storage.Snapshot) {
			addBinding(s, "fm5b7t4pan0d", "fk3x9r2man0d", "fk3x9r2m5iv8", "fibggi2juxhc", "enabled")
			s.Teams["fidfcosw0iyo"] = domain.Team{Name: "team", ID: "fidfcosw0iyo", ParentID: "fibggi2juxhc"}
			addBinding(s, "fm5b7t4pkxan", "fk3x9r2mkxan", "fk3x9r2man0d", "fidfcosw0iyo", "enabled")
			s.Teams["fp8h2w6yzcp2"] = domain.Team{Name: "team", ID: "fp8h2w6yzcp2", ParentID: "fibggi2juubk"}
			addBinding(s, "fm5b7t4pfs5i", "fk3x9r2mfs5i", "fk3x9r2m5iv8", "fp8h2w6yzcp2", "enabled")
		}, []string{"fm5b7t4pan0d", "fm5b7t4pfs5i", "fm5b7t4pkxan"}, false},
		{"unrelated holder is excluded", func(s *storage.Snapshot) {
			s.Teams["fp8h2w6yer4h"] = domain.Team{Name: "team", ID: "fp8h2w6yer4h"}
			s.Teams["fp8h2w6yu7kx"] = domain.Team{Name: "team", ID: "fp8h2w6yu7kx", ParentID: "fp8h2w6yer4h"}
			s.Assignments["fm5b7t4p4hu7"] = domain.Assignment{Version: "1", ID: "fm5b7t4p4hu7", GrantID: "fk3x9r2m5iv8", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "fp8h2w6yer4h"}, Status: "enabled"}
			addBinding(s, "fm5b7t4pan0d", "fk3x9r2man0d", "fk3x9r2m5iv8", "fibggi2juxhc", "enabled")
		}, []string{"fm5b7t4pan0d"}, false},
		{"parent grant without parent team is excluded", func(s *storage.Snapshot) {
			addBinding(s, "fm5b7t4pan0d", "fk3x9r2man0d", "fk3x9r2m5iv8", "fibggi2juubk", "enabled")
		}, nil, false},
		{"parent team without parent grant is excluded", func(s *storage.Snapshot) {
			addBinding(s, "fm5b7t4pan0d", "fk3x9r2man0d", "fk3x9r2m0dq3", "fibggi2juxhc", "enabled")
		}, nil, false},
		{"unassigned definition is excluded", func(s *storage.Snapshot) {
			s.Contents[domain.GrantKey{ID: "fk3x9r2m4hu7", Revision: 1}] = content("fk3x9r2m4hu7", "fk3x9r2m5iv8")
		}, nil, false},
		{"disabled bridge stays structural", func(s *storage.Snapshot) {
			addBinding(s, "fm5b7t4pan0d", "fk3x9r2man0d", "fk3x9r2m5iv8", "fibggi2juxhc", "disabled")
			s.Teams["fidfcosw0iyo"] = domain.Team{Name: "team", ID: "fidfcosw0iyo", ParentID: "fibggi2juxhc"}
			addBinding(s, "fm5b7t4pfs5i", "fk3x9r2mfs5i", "fk3x9r2man0d", "fidfcosw0iyo", "enabled")
		}, []string{"fm5b7t4pan0d", "fm5b7t4pfs5i"}, false},
		{"missing upstream holding terminates ancestor proof", func(s *storage.Snapshot) {
			delete(s.Assignments, "fm5b7t4p0dq3")
			addBinding(s, "fm5b7t4pan0d", "fk3x9r2man0d", "fk3x9r2m5iv8", "fibggi2juxhc", "enabled")
		}, []string{"fm5b7t4pan0d"}, false},
		{"missing adopted content rejects", func(s *storage.Snapshot) {
			addBinding(s, "fm5b7t4pan0d", "fk3x9r2man0d", "fk3x9r2m5iv8", "fibggi2juxhc", "enabled")
			delete(s.Contents, domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1})
		}, nil, true},
		{"missing recipient team rejects", func(s *storage.Snapshot) {
			addBinding(s, "fm5b7t4pan0d", "fk3x9r2man0d", "fk3x9r2m5iv8", "fibggi2juxhc", "enabled")
			delete(s.Teams, "fibggi2juxhc")
		}, nil, true},
		{"duplicate binding rejects even when disabled", func(s *storage.Snapshot) {
			addBinding(s, "fm5b7t4pan0d", "fk3x9r2man0d", "fk3x9r2m5iv8", "fibggi2juxhc", "enabled")
			addBinding(s, "fm5b7t4pfs5i", "fk3x9r2man0d", "fk3x9r2m5iv8", "fibggi2juxhc", "disabled")
		}, nil, true},
		{"enabled direct-human dependency is unsupported", func(s *storage.Snapshot) {
			s.Contents[domain.GrantKey{ID: "fk3x9r2mzcp2", Revision: 1}] = content("fk3x9r2mzcp2", "fk3x9r2m5iv8")
			s.Assignments["fm5b7t4pzcp2"] = domain.Assignment{Version: "1", ID: "fm5b7t4pzcp2", GrantID: "fk3x9r2mzcp2", GrantRevision: 1, Recipient: domain.Recipient{Type: "user", ID: "fi7io4lvjqio"}, Status: "enabled"}
		}, nil, true},
		{"newer unadopted content is ignored", func(s *storage.Snapshot) {
			addBinding(s, "fm5b7t4pan0d", "fk3x9r2man0d", "fk3x9r2m5iv8", "fibggi2juxhc", "enabled")
			newer := content("fk3x9r2man0d", "wrong")
			newer.Revision = 2
			s.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 2}] = newer
		}, []string{"fm5b7t4pan0d"}, false},
		{"disabled team cycle rejects", func(s *storage.Snapshot) {
			addBinding(s, "fm5b7t4pan0d", "fk3x9r2man0d", "fk3x9r2m5iv8", "fibggi2juxhc", "disabled")
			s.Teams["fibggi2juubk"] = domain.Team{ID: "fibggi2juubk", Name: "fp8h2w6y5iv8", ParentID: "fibggi2juxhc"}
		}, nil, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			test.edit(&fixture.Snapshot)
			before := snapshotEvidence(fixture.Snapshot)
			got, err := lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "fm5b7t4p5iv8")
			if test.bad {
				if err == nil || len(got) != 0 {
					t.Fatalf("dependents = %#v, %v", got, err)
				}
			} else {
				ids := make([]string, len(got))
				for i := range got {
					ids[i] = got[i].ID
				}
				if err != nil || !slices.Equal(ids, test.want) {
					t.Fatalf("dependents = %v, %v; want %v", ids, err, test.want)
				}
			}
			assertSnapshotUnchanged(t, fixture.Snapshot, before)
		})
	}
}

func TestDependentTeamAssignmentsRejectsIncompleteOuterEvidence(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	tests := []struct {
		name string
		edit func(*storage.Snapshot)
		id   string
	}{
		{"missing requested assignment", func(*storage.Snapshot) {}, "missing"},
		{"assignment map ID mismatch", func(s *storage.Snapshot) {
			a := s.Assignments["fm5b7t4p5iv8"]
			a.ID = "wrong"
			s.Assignments["fm5b7t4p5iv8"] = a
		}, "fm5b7t4p5iv8"},
		{"catalog outside area", func(s *storage.Snapshot) { s.Catalog.ApplicationID = "fi7io4lvkfsw" }, "fm5b7t4p5iv8"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			test.edit(&fixture.Snapshot)
			before := snapshotEvidence(fixture.Snapshot)
			got, err := lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, test.id)
			if err == nil || len(got) != 0 {
				t.Fatalf("dependents = %#v, %v", got, err)
			}
			assertSnapshotUnchanged(t, fixture.Snapshot, before)
		})
	}
}

func TestDependentTeamAssignmentsRejectsCancellationAndOverflow(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	before := snapshotEvidence(fixture.Snapshot)
	got, err := lineage.DependentTeamAssignments(ctx, fixture.Snapshot, "fm5b7t4p5iv8")
	if !errors.Is(err, context.Canceled) || len(got) != 0 {
		t.Fatalf("cancelled = %#v, %v", got, err)
	}
	assertSnapshotUnchanged(t, fixture.Snapshot, before)

	parentGrant, parentTeam := "fk3x9r2m5iv8", "fibggi2juubk"
	for i := 2; i <= 255; i++ {
		grant, team, assignment := "G"+strconv.Itoa(i), "Team"+strconv.Itoa(i), "A"+strconv.Itoa(i)
		fixture.Snapshot.Teams[team] = domain.Team{Name: "team", ID: team, ParentID: parentTeam}
		addBinding(&fixture.Snapshot, assignment, grant, parentGrant, team, "enabled")
		parentGrant, parentTeam = grant, team
	}
	before = snapshotEvidence(fixture.Snapshot)
	got, err = lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "fm5b7t4p5iv8")
	if err != nil || len(got) != 254 {
		t.Fatalf("bounded lineage = %d records, %v", len(got), err)
	}
	assertSnapshotUnchanged(t, fixture.Snapshot, before)
	fixture.Snapshot.Teams["fp8h2w6yfs5i"] = domain.Team{Name: "team", ID: "fp8h2w6yfs5i", ParentID: parentTeam}
	addBinding(&fixture.Snapshot, "fm5b7t4pu7kx", "fk3x9r2mu7kx", parentGrant, "fp8h2w6yfs5i", "enabled")
	before = snapshotEvidence(fixture.Snapshot)
	got, err = lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "fm5b7t4p5iv8")
	if err == nil || len(got) != 0 {
		t.Fatalf("overflow = %d records, %v", len(got), err)
	}
	assertSnapshotUnchanged(t, fixture.Snapshot, before)
}

func addBinding(s *storage.Snapshot, assignmentID, grantID, parentGrantID, teamID, status string) {
	s.Contents[domain.GrantKey{ID: grantID, Revision: 1}] = content(grantID, parentGrantID)
	s.Assignments[assignmentID] = domain.Assignment{Version: "1", ID: assignmentID, GrantID: grantID, GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: teamID}, Status: status}
}

func content(id, parent string) domain.GrantContent {
	return domain.GrantContent{Version: "1", GrantID: id, Revision: 1, ParentGrantID: parent, Permissions: []string{lab.PayslipRead}, Scope: map[string]string{}}
}

func snapshotEvidence(s storage.Snapshot) string { return fmt.Sprintf("%#v", s) }

func assertSnapshotUnchanged(t *testing.T, got storage.Snapshot, before string) {
	t.Helper()
	if snapshotEvidence(got) != before {
		t.Fatal("snapshot mutated")
	}
}

// Retiring a permission must not wedge the whole area's status writes.
//
// The inventory validated every assignment's content before answering about
// one, and failed on the first that did not hold — so retiring a permission one
// unrelated grant selects made every assignment unstatusable, including the
// root's own, which selects no permission at all. Q-125 makes retirement an
// ordinary act that does not require editing references first, and Q-132 makes
// bottom-up dismantle the only remedy for a subtree. The administrator could not
// perform its first half.
func TestARetiredPermissionDoesNotWedgeTheInventory(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments["fm5b7t4pan0d"] = fixture.Proposed

	// The control, so the refusal below cannot be a broken fixture.
	if _, err := lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "fm5b7t4p0dq3"); err != nil {
		t.Fatalf("the fixture does not answer, so nothing below means anything: %v", err)
	}

	write := fixture.Snapshot.Catalog.Permissions[lab.PayslipWrite]
	write.Active = false
	fixture.Snapshot.Catalog.Permissions[lab.PayslipWrite] = write

	// The root's binding selects no permission whatever, so a retirement cannot
	// be about it.
	if _, err := lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "fm5b7t4p0dq3"); err != nil {
		t.Fatalf("one retired permission wedged the inventory: %v", err)
	}

	// And the binding whose own grant selects the retired permission is still
	// found as a dependent. Dropping it would be the widening direction: this
	// graph exists to refuse a disable while something depends on the binding.
	got, err := lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "fm5b7t4p5iv8")
	if err != nil {
		t.Fatalf("dependents after retirement: %v", err)
	}
	if len(got) != 1 || got[0] != fixture.Proposed {
		t.Fatalf("an ineligible binding stopped being a dependent: %#v", got)
	}

	// Corruption is still fatal: a structure this cannot read is not one it may
	// reason about.
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}] = domain.GrantContent{
		Version: "1", GrantID: "fk3x9r2m5iv8", Revision: 1, Permissions: []string{""}, Scope: map[string]string{},
	}
	if _, err := lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "fm5b7t4p0dq3"); err == nil {
		t.Fatal("a corrupt grant content was tolerated")
	}
}
