package mutation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/lab"
	"errors"
	"testing"
	"time"
)

// The half of B13 the public fixture cannot show: a team that holds nothing
// itself, with a bound team beneath it.
//
// "Affected" is not "attached". A subteam resolves its authority through the
// chain above it, so moving an ancestor moves the subteam's inherited scope
// while the ancestor's own record shows no binding at all — which is exactly the
// move that looks safest to whoever makes it.
func TestSetTeamParentRefusesWhenTheBindingIsBeneathTheMovedTeam(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	build := func(t *testing.T, bindBeneath bool, status string) *mutation.Service {
		t.Helper()
		fixture := lab.TeamFINC17(area)
		snapshot := fixture.Snapshot
		// A subteam of Team2, and a grant hung below Team2's own.
		// Two levels below the moved team, not one. The guard grows the subtree
		// to a fixpoint, and a walk that stopped at direct children would have
		// passed every test until this one: the claim the commit makes loudest
		// is that "affected" reaches all the way down.
		snapshot.Teams["fibggi2jv4hu"] = domain.Team{ID: "fibggi2jv4hu", Name: "fp8h2w6yv4hu", ParentID: "fibggi2juxhc"}
		snapshot.Teams["fibggi2jv5k0"] = domain.Team{ID: "fibggi2jv5k0", Name: "fp8h2w6yv5k0", ParentID: "fibggi2jv4hu"}
		snapshot.Memberships = append(snapshot.Memberships, domain.Membership{TeamID: "fibggi2jv5k0", HumanID: "fi7io4lvk35s"})
		snapshot.Controls["fk3x9r2mv5k0"] = domain.GrantControl{Version: "1", ID: "fk3x9r2mv5k0", Status: "enabled"}
		snapshot.Contents[domain.GrantKey{ID: "fk3x9r2mv5k0", Revision: 1}] = domain.GrantContent{
			Version: "1", GrantID: "fk3x9r2mv5k0", Revision: 1, ParentGrantID: "fk3x9r2man0d",
			Permissions: []string{lab.PayslipRead}, Scope: map[string]string{},
		}
		if bindBeneath {
			snapshot.Assignments["fm5b7t4pv5k0"] = domain.Assignment{
				Version: "1", ID: "fm5b7t4pv5k0", GrantID: "fk3x9r2mv5k0", GrantRevision: 1,
				Recipient: domain.Recipient{Type: "group", ID: "fibggi2jv5k0"}, Status: status,
			}
		}
		provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{snapshot})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = provider.Close() })
		statusAdmin, err := lab.NewAssignmentStatusAdministration(area, fixture.Administration)
		if err != nil {
			t.Fatal(err)
		}
		service, err := mutation.New(provider,
			&lab.RoleAdministration{AssignmentStatusAdministration: statusAdmin},
			&fixedClock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)})
		if err != nil {
			t.Fatal(err)
		}
		return service
	}
	publisher := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}

	// Team2 itself holds nothing in either case, so the only difference is one
	// binding two levels down.
	t.Run("nothing beneath it", func(t *testing.T) {
		if _, err := build(t, false, "enabled").SetTeamParent(t.Context(), area, publisher, "fibggi2juxhc", "fibggi2jur5s"); err != nil {
			t.Fatalf("moving a team with nothing beneath it was refused: %v", err)
		}
	})
	t.Run("an enabled binding beneath it", func(t *testing.T) {
		if _, err := build(t, true, "enabled").SetTeamParent(t.Context(), area, publisher, "fibggi2juxhc", "fibggi2jur5s"); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("moving a team above an enabled binding gave %v, want ErrConflict", err)
		}
	})
	// A disabled binding holds nothing, so there is nothing of its to
	// re-anchor, and it must not stop the move. Without this the guard could
	// quietly become "any binding at all", which would make a team unmovable
	// forever once anything below it had ever been bound — disabled records are
	// retained, so that state never clears.
	t.Run("a disabled binding beneath it", func(t *testing.T) {
		if _, err := build(t, true, "disabled").SetTeamParent(t.Context(), area, publisher, "fibggi2juxhc", "fibggi2jur5s"); err != nil {
			t.Fatalf("a disabled binding blocked the move: %v", err)
		}
	})
}

// Re-asserting the parent a team already has is not a change, so B13 — which
// governs *changing* or removing a parent — does not reach it. Refusing it broke
// idempotent retries: a caller whose request timed out after succeeding got a
// conflict on the repeat, against a store already in the state it asked for.
func TestSetTeamParentIsANoOpAgainstTheParentItAlreadyHas(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	fixture := lab.TeamFINC17(area)
	provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	statusAdmin, err := lab.NewAssignmentStatusAdministration(area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	service, err := mutation.New(provider,
		&lab.RoleAdministration{AssignmentStatusAdministration: statusAdmin},
		&fixedClock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	publisher := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}

	// Team1 holds an enabled binding, so a real move is refused — which is what
	// makes this a test of the no-op and not of an unguarded team.
	if _, err := service.SetTeamParent(t.Context(), area, publisher, "fibggi2juubk", "fibggi2juxhc"); err == nil {
		t.Fatal("a real move of a bound team was allowed")
	}
	moved, err := service.SetTeamParent(t.Context(), area, publisher, "fibggi2juubk", "fibggi2jur5s")
	if err != nil {
		t.Fatalf("re-asserting the parent a team already has was refused: %v", err)
	}
	if moved.ParentID != "fibggi2jur5s" {
		t.Fatalf("the no-op changed the parent: %#v", moved)
	}
}
