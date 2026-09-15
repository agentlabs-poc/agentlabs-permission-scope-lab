package abv_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"context"
	"errors"
	"path/filepath"
	"testing"
)

// openTeamLab returns a facade over a real SQLite store seeded with the worked
// fixture, so these exercise the whole path rather than a validation function.
func openTeamLab(t *testing.T) (teamWriter, domain.Area) {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "teams.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := lab.Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeConnection() })
	writer, ok := api.(teamWriter)
	if !ok {
		t.Fatalf("lab application does not expose the team writes: %T", api)
	}
	return writer, area
}

type teamWriter interface {
	CreateTeam(context.Context, domain.Area, domain.FixtureContext, string, string) (domain.Team, error)
	SetTeamParent(context.Context, domain.Area, domain.FixtureContext, string, string) (domain.Team, error)
	DeleteTeam(context.Context, domain.Area, domain.FixtureContext, string) error
	AddMember(context.Context, domain.Area, domain.FixtureContext, string, string) error
	RemoveMember(context.Context, domain.Area, domain.FixtureContext, string, string) error
	GetTeam(context.Context, domain.Area, domain.FixtureContext, string) (domain.Team, error)
	ListMembers(context.Context, domain.Area, domain.FixtureContext, domain.MemberFilter) (domain.MemberPage, error)
}

var teamFixture = domain.FixtureContext{Name: "maya-role-publisher"}

// A created team is issued an id, persists, and reads back with the name the
// caller gave and the parent it was created under.
func TestCreateTeamIssuesAnIDAndPersists(t *testing.T) {
	api, area := openTeamLab(t)
	created, err := api.CreateTeam(t.Context(), area, teamFixture, "Finance", "")
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Name != "Finance" || created.ParentID != "" {
		t.Fatalf("created = %#v", created)
	}
	read, err := api.GetTeam(t.Context(), area, teamFixture, created.ID)
	if err != nil || read != created {
		t.Fatalf("read back %#v err=%v, want %#v", read, err, created)
	}
	// A subteam: create covers both, which is why there is no second operation.
	child, err := api.CreateTeam(t.Context(), area, teamFixture, "Payroll", created.ID)
	if err != nil || child.ParentID != created.ID {
		t.Fatalf("subteam = %#v err=%v", child, err)
	}
	if child.ID == created.ID {
		t.Fatal("two teams were issued the same id")
	}
	// A parent that does not exist is refused: the foreign key is now a rule.
	if _, err := api.CreateTeam(t.Context(), area, teamFixture, "Orphan", "fy6x28qcdrxx"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("created under an absent parent: %v", err)
	}
}

// Delete refuses while anything depends on the team, and nothing cascades.
func TestDeleteTeamRefusesWhileAnythingDependsOnIt(t *testing.T) {
	api, area := openTeamLab(t)
	const team1, team2 = "fibggi2juubk", "fibggi2juxhc"

	// fp8h2w6y5iv8 has a child.
	if err := api.DeleteTeam(t.Context(), area, teamFixture, team1); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleted a team with a child: %v", err)
	}
	// fp8h2w6yan0d has a member.
	if err := api.DeleteTeam(t.Context(), area, teamFixture, team2); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleted a team with a member: %v", err)
	}
	// Both survive: a refused delete removes nothing.
	for _, id := range []string{team1, team2} {
		if _, err := api.GetTeam(t.Context(), area, teamFixture, id); err != nil {
			t.Fatalf("a refused delete removed %s: %v", id, err)
		}
	}
	// An empty leaf goes.
	leaf, err := api.CreateTeam(t.Context(), area, teamFixture, "Temp", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := api.DeleteTeam(t.Context(), area, teamFixture, leaf.ID); err != nil {
		t.Fatalf("an empty leaf was not deletable: %v", err)
	}
	if _, err := api.GetTeam(t.Context(), area, teamFixture, leaf.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the deleted team survived: %v", err)
	}
	// Deleting it again is ErrNotFound, not a silent success.
	if err := api.DeleteTeam(t.Context(), area, teamFixture, leaf.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleting an absent team gave %v", err)
	}
}

// Re-parenting persists, and a cycle is refused at the write.
func TestSetTeamParentPersistsAndRefusesACycle(t *testing.T) {
	api, area := openTeamLab(t)
	const root, team1, team2 = "fibggi2jur5s", "fibggi2juubk", "fibggi2juxhc"

	moved, err := api.SetTeamParent(t.Context(), area, teamFixture, team2, "")
	if err != nil || moved.ParentID != "" {
		t.Fatalf("promoting to a root: %#v err=%v", moved, err)
	}
	read, err := api.GetTeam(t.Context(), area, teamFixture, team2)
	if err != nil || read.ParentID != "" {
		t.Fatalf("the move did not persist: %#v err=%v", read, err)
	}
	// The name survives a re-parent: only the parent changes.
	if read.Name != "fp8h2w6yan0d" {
		t.Fatalf("a re-parent changed the name: %#v", read)
	}
	// Put it back, then try to make its ancestor its descendant.
	if _, err := api.SetTeamParent(t.Context(), area, teamFixture, team2, team1); err != nil {
		t.Fatal(err)
	}
	if _, err := api.SetTeamParent(t.Context(), area, teamFixture, root, team2); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a cycle was accepted: %v", err)
	}
	// And the refused move changed nothing.
	if again, err := api.GetTeam(t.Context(), area, teamFixture, root); err != nil || again.ParentID != "" {
		t.Fatalf("a refused re-parent changed the row: %#v err=%v", again, err)
	}
}

// A membership's identity is the pair, so both operations say whether they did
// anything rather than succeeding silently.
func TestMembershipWritesAreExplicitAboutWhetherTheyDidAnything(t *testing.T) {
	api, area := openTeamLab(t)
	const team2, nutan = "fibggi2juxhc", "fi7io4lvjwu8"
	const maya = "fi7io4lvjqio"

	if err := api.AddMember(t.Context(), area, teamFixture, team2, maya); err != nil {
		t.Fatal(err)
	}
	if err := api.AddMember(t.Context(), area, teamFixture, team2, maya); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("adding the same member twice gave %v, want ErrConflict", err)
	}
	page, err := api.ListMembers(t.Context(), area, teamFixture, domain.MemberFilter{TeamID: team2})
	if err != nil || page.Total != 2 {
		t.Fatalf("roster = %d err=%v, want maya and nutan", page.Total, err)
	}
	if err := api.RemoveMember(t.Context(), area, teamFixture, team2, maya); err != nil {
		t.Fatal(err)
	}
	if err := api.RemoveMember(t.Context(), area, teamFixture, team2, maya); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("removing an absent member gave %v, want ErrNotFound", err)
	}
	// The other member is untouched: a membership write moves one row.
	page, err = api.ListMembers(t.Context(), area, teamFixture, domain.MemberFilter{TeamID: team2})
	if err != nil || page.Total != 1 || page.Members[0].HumanID != nutan {
		t.Fatalf("roster after removal = %#v err=%v", page.Members, err)
	}
	// A team that does not exist is refused — the lost foreign key, as a rule.
	if err := api.AddMember(t.Context(), area, teamFixture, "fy6x28qcdrxx", maya); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("added a member to an absent team: %v", err)
	}
}

// Every write is gated. A wrong fixture context reaches none of them.
func TestTeamWritesAreProtected(t *testing.T) {
	api, area := openTeamLab(t)
	wrong := domain.FixtureContext{Name: "someone-else"}
	const team2 = "fibggi2juxhc"

	if _, err := api.CreateTeam(t.Context(), area, wrong, "Finance", ""); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("create past the gate: %v", err)
	}
	if _, err := api.SetTeamParent(t.Context(), area, wrong, team2, ""); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("reparent past the gate: %v", err)
	}
	if err := api.DeleteTeam(t.Context(), area, wrong, team2); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("delete past the gate: %v", err)
	}
	if err := api.AddMember(t.Context(), area, wrong, team2, "fi7io4lvjqio"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("add-member past the gate: %v", err)
	}
	if err := api.RemoveMember(t.Context(), area, wrong, team2, "fi7io4lvjwu8"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("remove-member past the gate: %v", err)
	}
}
