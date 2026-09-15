package abv_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"context"
	"encoding/json"
	"errors"
	"os"
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

// The lab's publication gate admits exactly one route — Maya publishing a fk3x9r2man0d
// revision through fm5b7t4p5iv8 — so the upgrade test works on fk3x9r2man0d, which is also the
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

// The fixture holds fm5b7t4p0dq3 (fk3x9r2m0dq3 to fp8h2w6ykxan) and fm5b7t4p5iv8 (fk3x9r2m5iv8 to fp8h2w6y5iv8).
func TestGetAndListAssignmentsAnswerBothDirections(t *testing.T) {
	api, area := openAssignmentLab(t)

	got, err := api.GetAssignment(t.Context(), area, teamFixture, "fm5b7t4p5iv8")
	if err != nil || got.GrantID != "fk3x9r2m5iv8" || got.Recipient.ID != "fibggi2juubk" || got.GrantRevision != 1 {
		t.Fatalf("fm5b7t4p5iv8 = %#v err=%v", got, err)
	}
	if _, err := api.GetAssignment(t.Context(), area, teamFixture, "absent"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("absent gave %v, want ErrNotFound", err)
	}

	// Which recipients hold this grant.
	byGrant, err := api.ListAssignments(t.Context(), area, teamFixture, domain.AssignmentFilter{GrantID: "fk3x9r2m5iv8"})
	if err != nil || byGrant.Total != 1 || byGrant.Assignments[0].ID != "fm5b7t4p5iv8" {
		t.Fatalf("by grant = %#v total=%d err=%v", byGrant.Assignments, byGrant.Total, err)
	}
	// Which grants this recipient holds.
	team1 := domain.Recipient{Type: "group", ID: "fibggi2juubk"}
	byRecipient, err := api.ListAssignments(t.Context(), area, teamFixture, domain.AssignmentFilter{Recipient: &team1})
	if err != nil || byRecipient.Total != 1 || byRecipient.Assignments[0].GrantID != "fk3x9r2m5iv8" {
		t.Fatalf("by recipient = %#v total=%d err=%v", byRecipient.Assignments, byRecipient.Total, err)
	}

	// Neither direction, or both, is not a question anyone asks.
	if _, err := api.ListAssignments(t.Context(), area, teamFixture, domain.AssignmentFilter{}); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("unfiltered gave %v, want ErrMalformed", err)
	}
	if _, err := api.ListAssignments(t.Context(), area, teamFixture, domain.AssignmentFilter{GrantID: "fk3x9r2m5iv8", Recipient: &team1}); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("both filters gave %v, want ErrMalformed", err)
	}
}

// Q-105: an upgrade selects the LATEST revision, never an intermediate one, and
// when the latest cannot be supported it rejects and changes nothing. Falling
// back to a revision that would pass is the behaviour the handbook forbids.
func TestUpgradeAssignmentTakesTheLatestAndNeverAnIntermediate(t *testing.T) {
	api, area := openAssignmentLab(t)

	// Assign fk3x9r2man0d revision 1 to fp8h2w6yan0d first, so the assignment exists before the
	// revisions it will later be upgraded past. That is the real sequence.
	if _, err := api.Assign(t.Context(), area, domain.FixtureContext{Name: "maya-team1"},
		[]byte(`{"version":"1","id":"fm5b7t4pan0d","grant_id":"fk3x9r2man0d","grant_revision":1,"recipient":{"type":"group","id":"fibggi2juxhc"},"status":"enabled"}`)); err != nil {
		t.Fatal(err)
	}

	for _, revision := range []int64{2, 3} {
		content, err := json.Marshal(domain.GrantContent{
			Version: "1", GrantID: "fk3x9r2man0d", Revision: revision, ParentGrantID: "fk3x9r2m5iv8",
			Permissions: []string{payslipRead},
			Scope:       map[string]string{"cert": "C17"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := api.PublishGrantRevision(t.Context(), area, grantPublisher, "fm5b7t4p5iv8", content); err != nil {
			t.Fatalf("publishing revision %d: %v", revision, err)
		}
	}

	// Publication alone does not move the assignment — Q-102.
	before, err := api.GetAssignment(t.Context(), area, teamFixture, "fm5b7t4pan0d")
	if err != nil || before.GrantRevision != 1 {
		t.Fatalf("publication moved the adoption: %#v err=%v", before, err)
	}

	after, err := api.UpgradeAssignment(t.Context(), area, teamFixture, "fm5b7t4pan0d")
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
	reread, err := api.GetAssignment(t.Context(), area, teamFixture, "fm5b7t4pan0d")
	if err != nil || reread.GrantRevision != 3 {
		t.Fatalf("the upgrade did not persist: %#v err=%v", reread, err)
	}

	// Upgrading again is a no-op, not a conflict: the caller asked for the
	// latest and the latest is what it has.
	again, err := api.UpgradeAssignment(t.Context(), area, teamFixture, "fm5b7t4pan0d")
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

	// The vehicle used to be fm5b7t4p0dq3, the root's own assignment, which is
	// now refused categorically — a root and its holder cannot be deleted at
	// all. That refusal would have masked this rule, so the dependency is built
	// one level down instead, where it is the only thing doing the refusing.
	//
	// fm5b7t4pan0d carries fk3x9r2man0d, whose parent is fk3x9r2m5iv8. So once
	// it exists, fm5b7t4p5iv8 is the support underneath it and removing that
	// would cut the route from below.
	if _, err := api.Assign(t.Context(), area, domain.FixtureContext{Name: "maya-team1"}, a2Proposal(t)); err != nil {
		t.Fatal(err)
	}
	if err := api.DeleteAssignment(t.Context(), area, teamFixture, "fm5b7t4p5iv8"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleting the supporting assignment gave %v, want ErrConflict", err)
	}
	if err := api.DeleteAssignment(t.Context(), area, teamFixture, "fm5b7t4p0dq3"); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("deleting the root's assignment gave %v, want ErrUnsupported", err)
	}
	// Clear the dependent, so the support can go too a few lines below — which
	// is what makes the refusal above a dependency and not a prohibition.
	if err := api.DeleteAssignment(t.Context(), area, teamFixture, "fm5b7t4pan0d"); err != nil {
		t.Fatal(err)
	}
	if err := api.DeleteAssignment(t.Context(), area, teamFixture, "absent"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleting an absent assignment gave %v, want ErrNotFound", err)
	}

	// fm5b7t4p5iv8 has nothing resting on it, so it can go — and it goes completely.
	if err := api.DeleteAssignment(t.Context(), area, teamFixture, "fm5b7t4p5iv8"); err != nil {
		t.Fatal(err)
	}
	if _, err := api.GetAssignment(t.Context(), area, teamFixture, "fm5b7t4p5iv8"); !errors.Is(err, domain.ErrNotFound) {
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

// a2Proposal is Team2's binding as the fixture files it, so the dependency this
// test needs is the one the rest of the corpus uses rather than a shape invented
// here.
func a2Proposal(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/a2.json")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
