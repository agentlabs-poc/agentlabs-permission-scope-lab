package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/lineage"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestResolveTeamAssignmentReturnsActualHoldingWithoutMutation(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	before := f.Snapshot.Assignments["A1"]
	route, err := lineage.ResolveTeamAssignment(f.Snapshot, "A1", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if route.GrantID != "G1" || !reflect.DeepEqual(route.Permissions, []string{lab.PayslipRead, lab.PayslipWrite}) || !reflect.DeepEqual(route.AssignmentIDs, []string{"A0", "A1"}) {
		t.Fatalf("wrong actual source: %#v", route)
	}
	if f.Snapshot.Assignments["A1"] != before {
		t.Fatal("resolver mutated input")
	}
}

func TestResolveTeamAssignmentRejectsInvalidEvidence(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	for _, tc := range []struct {
		name string
		id   string
		edit func(*lab.TeamFINC17Case)
		want error
	}{
		{"blank id", "", func(*lab.TeamFINC17Case) {}, domain.ErrMalformed},
		{"missing id", "missing", func(*lab.TeamFINC17Case) {}, domain.ErrRejected},
		{"area mismatch", "A1", func(f *lab.TeamFINC17Case) { f.Snapshot.Catalog.ApplicationID = "other" }, domain.ErrRejected},
		{"malformed projection", "A1", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["A1"]
			a.ID = "other"
			f.Snapshot.Assignments["A1"] = a
		}, domain.ErrRejected},
		{"disabled assignment", "A1", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["A1"]
			a.Status = "disabled"
			f.Snapshot.Assignments["A1"] = a
		}, domain.ErrRejected},
		{"disabled control", "A1", func(f *lab.TeamFINC17Case) {
			c := f.Snapshot.Controls["G1"]
			c.Status = "disabled"
			f.Snapshot.Controls["G1"] = c
		}, domain.ErrRejected},
		{"cycle", "A1", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Teams["RootTeam"] = domain.Team{ID: "RootTeam", ParentID: "Team1"}
		}, domain.ErrRejected},
		{"untrusted root", "A1", func(f *lab.TeamFINC17Case) { delete(f.Snapshot.TrustedRoots, "G0") }, domain.ErrRejected},
		{"user recipient", "A1", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["A1"]
			a.Recipient.Type = "user"
			f.Snapshot.Assignments["A1"] = a
		}, domain.ErrUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := lab.TeamFINC17(area)
			tc.edit(&f)
			got, err := lineage.ResolveTeamAssignment(f.Snapshot, tc.id, time.Time{})
			if !errors.Is(err, tc.want) || !reflect.DeepEqual(got, domain.Route{}) {
				t.Fatalf("got %#v, %v; want %v", got, err, tc.want)
			}
		})
	}
}

func TestResolveTeamAssignmentRejectsExactExpiry(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	expires := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	g := f.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}]
	g.Validity = &domain.Validity{ExpiresAt: &expires}
	f.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}] = g
	if _, err := lineage.ResolveTeamAssignment(f.Snapshot, "A1", expires); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("expired assignment resolved: %v", err)
	}
}
