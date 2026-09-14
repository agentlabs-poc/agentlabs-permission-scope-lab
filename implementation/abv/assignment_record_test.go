package abv_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
)

type assignmentAPI interface {
	GetAssignment(context.Context, domain.Area, domain.FixtureContext, string) (domain.Assignment, error)
	ListAssignments(context.Context, domain.Area, domain.FixtureContext, domain.AssignmentFilter) (domain.AssignmentPage, error)
	DeleteAssignment(context.Context, domain.Area, domain.FixtureContext, string) error
	UpgradeAssignment(context.Context, domain.Area, domain.FixtureContext, string) (domain.Assignment, error)
	PublishGrantRevision(context.Context, domain.Area, domain.FixtureContext, string, []byte) (domain.GrantContent, error)
	Assign(context.Context, domain.Area, domain.FixtureContext, []byte) (domain.Receipt, error)
}

// The lab's publication gate admits exactly one route — Maya publishing a G2
// revision through A1 — so the upgrade test works on G2, which is also the
// realistic shape: the grant being amended is the one below the publisher's own.
var grantPublisher = domain.FixtureContext{Name: "maya-grant-publisher"}

func openAssignmentLab(t *testing.T) (assignmentAPI, domain.Area) {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "assignments.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := lab.Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeConnection() })
	writer, ok := api.(assignmentAPI)
	if !ok {
		t.Fatalf("lab application does not expose the assignment operations: %T", api)
	}
	return writer, area
}

// The fixture holds A0 (G0 to RootTeam) and A1 (G1 to Team1).
func TestGetAndListAssignmentsAnswerBothDirections(t *testing.T) {
	api, area := openAssignmentLab(t)

	got, err := api.GetAssignment(t.Context(), area, teamFixture, "A1")
	if err != nil || got.GrantID != "G1" || got.Recipient.ID != "fibggi2juubk" || got.GrantRevision != 1 {
		t.Fatalf("A1 = %#v err=%v", got, err)
	}
	if _, err := api.GetAssignment(t.Context(), area, teamFixture, "absent"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("absent gave %v, want ErrNotFound", err)
	}

	// Which recipients hold this grant.
	byGrant, err := api.ListAssignments(t.Context(), area, teamFixture, domain.AssignmentFilter{GrantID: "G1"})
	if err != nil || byGrant.Total != 1 || byGrant.Assignments[0].ID != "A1" {
		t.Fatalf("by grant = %#v total=%d err=%v", byGrant.Assignments, byGrant.Total, err)
	}
	// Which grants this recipient holds.
	team1 := domain.Recipient{Type: "group", ID: "fibggi2juubk"}
	byRecipient, err := api.ListAssignments(t.Context(), area, teamFixture, domain.AssignmentFilter{Recipient: &team1})
	if err != nil || byRecipient.Total != 1 || byRecipient.Assignments[0].GrantID != "G1" {
		t.Fatalf("by recipient = %#v total=%d err=%v", byRecipient.Assignments, byRecipient.Total, err)
	}

	// Neither direction, or both, is not a question anyone asks.
	if _, err := api.ListAssignments(t.Context(), area, teamFixture, domain.AssignmentFilter{}); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("unfiltered gave %v, want ErrMalformed", err)
	}
	if _, err := api.ListAssignments(t.Context(), area, teamFixture, domain.AssignmentFilter{GrantID: "G1", Recipient: &team1}); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("both filters gave %v, want ErrMalformed", err)
	}
}

// Q-105: an upgrade selects the LATEST revision, never an intermediate one, and
// when the latest cannot be supported it rejects and changes nothing. Falling
// back to a revision that would pass is the behaviour the handbook forbids.
func TestUpgradeAssignmentTakesTheLatestAndNeverAnIntermediate(t *testing.T) {
	api, area := openAssignmentLab(t)

	// Assign G2 revision 1 to Team2 first, so the assignment exists before the
	// revisions it will later be upgraded past. That is the real sequence.
	if _, err := api.Assign(t.Context(), area, domain.FixtureContext{Name: "maya-team1"},
		[]byte(`{"version":"1","id":"A2","grant_id":"G2","grant_revision":1,"recipient":{"type":"group","id":"fibggi2juxhc"},"status":"enabled"}`)); err != nil {
		t.Fatal(err)
	}

	for _, revision := range []int64{2, 3} {
		content, err := json.Marshal(domain.GrantContent{
			Version: "1", GrantID: "G2", Revision: revision, ParentGrantID: "G1",
			Permissions: []string{payslipRead},
			Scope:       map[string]string{"cert": "C17"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := api.PublishGrantRevision(t.Context(), area, grantPublisher, "A1", content); err != nil {
			t.Fatalf("publishing revision %d: %v", revision, err)
		}
	}

	// Publication alone does not move the assignment — Q-102.
	before, err := api.GetAssignment(t.Context(), area, teamFixture, "A2")
	if err != nil || before.GrantRevision != 1 {
		t.Fatalf("publication moved the adoption: %#v err=%v", before, err)
	}

	after, err := api.UpgradeAssignment(t.Context(), area, teamFixture, "A2")
	if err != nil {
		t.Fatal(err)
	}
	if after.GrantRevision != 3 {
		t.Fatalf("upgraded to revision %d, want 3 — never an intermediate", after.GrantRevision)
	}
	// Everything else about the binding is untouched.
	if after.ID != before.ID || after.GrantID != before.GrantID || after.Recipient != before.Recipient || after.Status != before.Status {
		t.Fatalf("the upgrade changed more than the revision: %#v -> %#v", before, after)
	}
	reread, err := api.GetAssignment(t.Context(), area, teamFixture, "A2")
	if err != nil || reread.GrantRevision != 3 {
		t.Fatalf("the upgrade did not persist: %#v err=%v", reread, err)
	}

	// Upgrading again is a no-op, not a conflict: the caller asked for the
	// latest and the latest is what it has.
	again, err := api.UpgradeAssignment(t.Context(), area, teamFixture, "A2")
	if err != nil || again.GrantRevision != 3 {
		t.Fatalf("re-upgrade gave %#v err=%v, want a no-op", again, err)
	}
	if _, err := api.UpgradeAssignment(t.Context(), area, teamFixture, "absent"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("upgrading an absent assignment gave %v, want ErrNotFound", err)
	}
}

// Deletion is permanent, and refuses while a dependent route rests on this one.
func TestDeleteAssignmentRefusesWhileADependentRouteRestsOnIt(t *testing.T) {
	api, area := openAssignmentLab(t)

	// A0 carries G0, the root. A1's grant G1 has G0 as its parent, so A0 is the
	// support underneath A1 — removing it would cut the route from below.
	if err := api.DeleteAssignment(t.Context(), area, teamFixture, "A0"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleting the supporting assignment gave %v, want ErrConflict", err)
	}
	if err := api.DeleteAssignment(t.Context(), area, teamFixture, "absent"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleting an absent assignment gave %v, want ErrNotFound", err)
	}

	// A1 has nothing resting on it, so it can go — and it goes completely.
	if err := api.DeleteAssignment(t.Context(), area, teamFixture, "A1"); err != nil {
		t.Fatal(err)
	}
	if _, err := api.GetAssignment(t.Context(), area, teamFixture, "A1"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the assignment survived deletion: %v", err)
	}
	// Q-104 counts only *current* assignments, so the binding is free again —
	// which is the difference between removing and disabling.
	team1 := domain.Recipient{Type: "group", ID: "fibggi2juubk"}
	page, err := api.ListAssignments(t.Context(), area, teamFixture, domain.AssignmentFilter{Recipient: &team1})
	if err != nil || page.Total != 0 {
		t.Fatalf("after deletion the recipient still holds %#v err=%v", page.Assignments, err)
	}
}
