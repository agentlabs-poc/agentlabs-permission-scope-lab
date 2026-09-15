package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/lab"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestResolveTeamAssignmentReturnsActualHoldingWithoutMutation(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	before := f.Snapshot.Assignments["fm5b7t4p5iv8"]
	route, err := lineage.ResolveTeamAssignment(f.Snapshot, "fm5b7t4p5iv8", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if route.GrantID != "fk3x9r2m5iv8" || !reflect.DeepEqual(route.Permissions, []string{lab.PayslipRead, lab.PayslipWrite}) || !reflect.DeepEqual(route.AssignmentIDs, []string{"fm5b7t4p0dq3", "fm5b7t4p5iv8"}) {
		t.Fatalf("wrong actual source: %#v", route)
	}
	if f.Snapshot.Assignments["fm5b7t4p5iv8"] != before {
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
		{"area mismatch", "fm5b7t4p5iv8", func(f *lab.TeamFINC17Case) { f.Snapshot.Catalog.ApplicationID = "fi7io4lvkfsw" }, domain.ErrRejected},
		{"malformed projection", "fm5b7t4p5iv8", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			a.ID = "fi7io4lvkfsw"
			f.Snapshot.Assignments["fm5b7t4p5iv8"] = a
		}, domain.ErrRejected},
		{"disabled assignment", "fm5b7t4p5iv8", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			a.Status = "disabled"
			f.Snapshot.Assignments["fm5b7t4p5iv8"] = a
		}, domain.ErrRejected},
		{"disabled control", "fm5b7t4p5iv8", func(f *lab.TeamFINC17Case) {
			c := f.Snapshot.Controls["fk3x9r2m5iv8"]
			c.Status = "disabled"
			f.Snapshot.Controls["fk3x9r2m5iv8"] = c
		}, domain.ErrRejected},
		{"cycle", "fm5b7t4p5iv8", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Teams["fibggi2jur5s"] = domain.Team{ID: "fibggi2jur5s", Name: "fp8h2w6ykxan", ParentID: "fibggi2juubk"}
		}, domain.ErrRejected},
		{"untrusted root", "fm5b7t4p5iv8", func(f *lab.TeamFINC17Case) { delete(f.Snapshot.TrustedRoots, "fk3x9r2m0dq3") }, domain.ErrRejected},
		{"user recipient", "fm5b7t4p5iv8", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			a.Recipient.Type = "user"
			f.Snapshot.Assignments["fm5b7t4p5iv8"] = a
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
	g := f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}]
	g.Validity = &domain.Validity{ExpiresAt: &expires}
	f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}] = g
	if _, err := lineage.ResolveTeamAssignment(f.Snapshot, "fm5b7t4p5iv8", expires); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("expired assignment resolved: %v", err)
	}
}
