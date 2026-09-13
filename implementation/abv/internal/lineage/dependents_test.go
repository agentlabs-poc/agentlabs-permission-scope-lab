package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
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
	fixture.Snapshot.Assignments["A2"] = fixture.Proposed
	got, err := lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "A1")
	if err != nil || len(got) != 1 || got[0] != fixture.Proposed {
		t.Fatalf("dependents = %#v, %v", got, err)
	}
}

func TestDependentTeamAssignmentsUsesUnambiguousBindingKeys(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Teams["T"] = domain.Team{Name: "team",ID: "T"}
	fixture.Snapshot.Teams["X\x00T"] = domain.Team{Name: "team",ID: "X\x00T"}
	addBinding(&fixture.Snapshot, "AX", "G\x00X", "", "T", "enabled")
	addBinding(&fixture.Snapshot, "AY", "G", "", "X\x00T", "enabled")
	before := snapshotEvidence(fixture.Snapshot)
	got, err := lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "A1")
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
			addBinding(s, "A2", "G2", "G1", "fibggi2juxhc", "enabled")
			s.Teams["fidfcosw0iyo"] = domain.Team{Name: "team",ID: "fidfcosw0iyo", ParentID: "fibggi2juxhc"}
			addBinding(s, "A4", "G4", "G2", "fidfcosw0iyo", "enabled")
			s.Teams["Branch"] = domain.Team{Name: "team",ID: "Branch", ParentID: "fibggi2juubk"}
			addBinding(s, "A3", "G3", "G1", "Branch", "enabled")
		}, []string{"A2", "A3", "A4"}, false},
		{"unrelated holder is excluded", func(s *storage.Snapshot) {
			s.Teams["OtherRoot"] = domain.Team{Name: "team",ID: "OtherRoot"}
			s.Teams["OtherChild"] = domain.Team{Name: "team",ID: "OtherChild", ParentID: "OtherRoot"}
			s.Assignments["AX"] = domain.Assignment{Version: "1", ID: "AX", GrantID: "G1", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "OtherRoot"}, Status: "enabled"}
			addBinding(s, "A2", "G2", "G1", "fibggi2juxhc", "enabled")
		}, []string{"A2"}, false},
		{"parent grant without parent team is excluded", func(s *storage.Snapshot) {
			addBinding(s, "A2", "G2", "G1", "fibggi2juubk", "enabled")
		}, nil, false},
		{"parent team without parent grant is excluded", func(s *storage.Snapshot) {
			addBinding(s, "A2", "G2", "G0", "fibggi2juxhc", "enabled")
		}, nil, false},
		{"unassigned definition is excluded", func(s *storage.Snapshot) {
			s.Contents[domain.GrantKey{ID: "GX", Revision: 1}] = content("GX", "G1")
		}, nil, false},
		{"disabled bridge stays structural", func(s *storage.Snapshot) {
			addBinding(s, "A2", "G2", "G1", "fibggi2juxhc", "disabled")
			s.Teams["fidfcosw0iyo"] = domain.Team{Name: "team",ID: "fidfcosw0iyo", ParentID: "fibggi2juxhc"}
			addBinding(s, "A3", "G3", "G2", "fidfcosw0iyo", "enabled")
		}, []string{"A2", "A3"}, false},
		{"missing upstream holding terminates ancestor proof", func(s *storage.Snapshot) {
			delete(s.Assignments, "A0")
			addBinding(s, "A2", "G2", "G1", "fibggi2juxhc", "enabled")
		}, []string{"A2"}, false},
		{"missing adopted content rejects", func(s *storage.Snapshot) {
			addBinding(s, "A2", "G2", "G1", "fibggi2juxhc", "enabled")
			delete(s.Contents, domain.GrantKey{ID: "G2", Revision: 1})
		}, nil, true},
		{"missing recipient team rejects", func(s *storage.Snapshot) {
			addBinding(s, "A2", "G2", "G1", "fibggi2juxhc", "enabled")
			delete(s.Teams, "fibggi2juxhc")
		}, nil, true},
		{"duplicate binding rejects even when disabled", func(s *storage.Snapshot) {
			addBinding(s, "A2", "G2", "G1", "fibggi2juxhc", "enabled")
			addBinding(s, "A3", "G2", "G1", "fibggi2juxhc", "disabled")
		}, nil, true},
		{"enabled direct-human dependency is unsupported", func(s *storage.Snapshot) {
			s.Contents[domain.GrantKey{ID: "GU", Revision: 1}] = content("GU", "G1")
			s.Assignments["AU"] = domain.Assignment{Version: "1", ID: "AU", GrantID: "GU", GrantRevision: 1, Recipient: domain.Recipient{Type: "user", ID: "fi7io4lvjqio"}, Status: "enabled"}
		}, nil, true},
		{"newer unadopted content is ignored", func(s *storage.Snapshot) {
			addBinding(s, "A2", "G2", "G1", "fibggi2juxhc", "enabled")
			newer := content("G2", "wrong")
			newer.Revision = 2
			s.Contents[domain.GrantKey{ID: "G2", Revision: 2}] = newer
		}, []string{"A2"}, false},
		{"disabled team cycle rejects", func(s *storage.Snapshot) {
			addBinding(s, "A2", "G2", "G1", "fibggi2juxhc", "disabled")
			s.Teams["fibggi2juubk"] = domain.Team{ID: "fibggi2juubk", Name: "Team1", ParentID: "fibggi2juxhc"}
		}, nil, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			test.edit(&fixture.Snapshot)
			before := snapshotEvidence(fixture.Snapshot)
			got, err := lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "A1")
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
		{"assignment map ID mismatch", func(s *storage.Snapshot) { a := s.Assignments["A1"]; a.ID = "wrong"; s.Assignments["A1"] = a }, "A1"},
		{"catalog outside area", func(s *storage.Snapshot) { s.Catalog.ApplicationID = "fi7io4lvkfsw" }, "A1"},
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
	got, err := lineage.DependentTeamAssignments(ctx, fixture.Snapshot, "A1")
	if !errors.Is(err, context.Canceled) || len(got) != 0 {
		t.Fatalf("cancelled = %#v, %v", got, err)
	}
	assertSnapshotUnchanged(t, fixture.Snapshot, before)

	parentGrant, parentTeam := "G1", "fibggi2juubk"
	for i := 2; i <= 255; i++ {
		grant, team, assignment := "G"+strconv.Itoa(i), "Team"+strconv.Itoa(i), "A"+strconv.Itoa(i)
		fixture.Snapshot.Teams[team] = domain.Team{Name: "team",ID: team, ParentID: parentTeam}
		addBinding(&fixture.Snapshot, assignment, grant, parentGrant, team, "enabled")
		parentGrant, parentTeam = grant, team
	}
	before = snapshotEvidence(fixture.Snapshot)
	got, err = lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "A1")
	if err != nil || len(got) != 254 {
		t.Fatalf("bounded lineage = %d records, %v", len(got), err)
	}
	assertSnapshotUnchanged(t, fixture.Snapshot, before)
	fixture.Snapshot.Teams["Team256"] = domain.Team{Name: "team",ID: "Team256", ParentID: parentTeam}
	addBinding(&fixture.Snapshot, "A256", "G256", parentGrant, "Team256", "enabled")
	before = snapshotEvidence(fixture.Snapshot)
	got, err = lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "A1")
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
