package mutation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/lab"
	"context"
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
	// Which branch holds the binding. A walk that followed one path would be
	// caught by one of these and not the other, whichever path it chose — with a
	// single fixture it was only caught by half the possible bugs.
	build := func(t *testing.T, bindBeneath bool, status, branch string) *mutation.Service {
		t.Helper()
		fixture := lab.TeamFINC17(area)
		snapshot := fixture.Snapshot
		// A subteam of Team2, and a grant hung below Team2's own.
		// Two levels below the moved team, not one. The guard grows the subtree
		// to a fixpoint, and a walk that stopped at direct children would have
		// passed every test until this one: the claim the commit makes loudest
		// is that "affected" reaches all the way down.
		//
		// It also branches. B18: "inspect all affected bindings and descendants,
		// not one path" — so the bound subtree hangs off the *second* child by
		// id, and a walk that followed one branch would miss it.
		snapshot.Teams["fibggi2ja000"] = domain.Team{ID: "fibggi2ja000", Name: "fp8h2w6ya000", ParentID: "fibggi2juxhc"}
		snapshot.Teams["fibggi2ja001"] = domain.Team{ID: "fibggi2ja001", Name: "fp8h2w6ya001", ParentID: "fibggi2ja000"}
		snapshot.Teams["fibggi2jv4hu"] = domain.Team{ID: "fibggi2jv4hu", Name: "fp8h2w6yv4hu", ParentID: "fibggi2juxhc"}
		snapshot.Teams["fibggi2jv5k0"] = domain.Team{ID: "fibggi2jv5k0", Name: "fp8h2w6yv5k0", ParentID: "fibggi2jv4hu"}
		// The bound team sits at the foot of whichever branch the case names.
		bound := map[string]string{"first": "fibggi2ja001", "last": "fibggi2jv5k0"}[branch]
		snapshot.Memberships = append(snapshot.Memberships, domain.Membership{TeamID: "fibggi2jv5k0", HumanID: "fi7io4lvk35s"})
		snapshot.Memberships = append(snapshot.Memberships, domain.Membership{TeamID: "fibggi2ja001", HumanID: "fi7io4lvk35s"})
		snapshot.Controls["fk3x9r2mv5k0"] = domain.GrantControl{Version: "1", ID: "fk3x9r2mv5k0", Status: "enabled"}
		snapshot.Contents[domain.GrantKey{ID: "fk3x9r2mv5k0", Revision: 1}] = domain.GrantContent{
			Version: "1", GrantID: "fk3x9r2mv5k0", Revision: 1, ParentGrantID: "fk3x9r2man0d",
			Permissions: []string{lab.PayslipRead}, Scope: map[string]string{},
		}
		if bindBeneath {
			snapshot.Assignments["fm5b7t4pv5k0"] = domain.Assignment{
				Version: "1", ID: "fm5b7t4pv5k0", GrantID: "fk3x9r2mv5k0", GrantRevision: 1,
				Recipient: domain.Recipient{Type: "group", ID: bound}, Status: status,
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
		if _, err := build(t, false, "enabled", "last").SetTeamParent(t.Context(), area, publisher, "fibggi2juxhc", "fibggi2jur5s"); err != nil {
			t.Fatalf("moving a team with nothing beneath it was refused: %v", err)
		}
	})
	for _, branch := range []string{"first", "last"} {
		t.Run("an enabled binding beneath its "+branch+" branch", func(t *testing.T) {
			if _, err := build(t, true, "enabled", branch).SetTeamParent(t.Context(), area, publisher, "fibggi2juxhc", "fibggi2jur5s"); !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("moving a team above an enabled binding gave %v, want ErrConflict", err)
			}
		})
	}
	// A disabled binding holds nothing, so there is nothing of its to
	// re-anchor, and it must not stop the move. Without this the guard could
	// quietly become "any binding at all", which would make a team unmovable
	// forever once anything below it had ever been bound — disabled records are
	// retained, so that state never clears.
	t.Run("a disabled binding beneath it", func(t *testing.T) {
		if _, err := build(t, true, "disabled", "last").SetTeamParent(t.Context(), area, publisher, "fibggi2juxhc", "fibggi2jur5s"); err != nil {
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
	//
	// The parent has to be one Team1 could legitimately move to. The first
	// version used Team2, Team1's own descendant, so CheckTeamReparent rejected
	// it as a cycle and the control passed with the guard removed entirely.
	if _, err := service.SetTeamParent(t.Context(), area, publisher, "fibggi2juubk", "fibggi2jv0n4"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a real move of a bound team gave %v, want ErrConflict", err)
	}
	moved, err := service.SetTeamParent(t.Context(), area, publisher, "fibggi2juubk", "fibggi2jur5s")
	if err != nil {
		t.Fatalf("re-asserting the parent a team already has was refused: %v", err)
	}
	if moved.ParentID != "fibggi2jur5s" {
		t.Fatalf("the no-op changed the parent: %#v", moved)
	}
}

// refuseAffectedBindings walks a graph it does not own. CheckTeamReparent runs
// first and refuses a cycle, so the walk's own guard against one is never
// reached through SetTeamParent — which means the ordering is load-bearing and
// nothing asserted it, and the walk's behaviour on a graph that is already
// broken was a matter of reading rather than of test.
//
// Both matter at migration: a store written by anything other than this code can
// hold a cycle or a parent id that names no team, and the answer must be a
// refusal or a bounded walk, never a hang.
func TestMovingATeamInAGraphThatIsAlreadyBroken(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	service := func(t *testing.T, bend func(*storage.Snapshot)) *mutation.Service {
		t.Helper()
		fixture := lab.TeamFINC17(area)
		snapshot := fixture.Snapshot
		bend(&snapshot)
		provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/broken.db", []storage.Snapshot{snapshot})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = provider.Close() })
		statusAdmin, err := lab.NewAssignmentStatusAdministration(area, fixture.Administration)
		if err != nil {
			t.Fatal(err)
		}
		built, err := mutation.New(provider,
			&lab.RoleAdministration{AssignmentStatusAdministration: statusAdmin},
			&fixedClock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)})
		if err != nil {
			t.Fatal(err)
		}
		return built
	}
	publisher := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}

	// A cycle among teams the move does not touch. The walk must terminate and
	// the answer must be the guard's, not a hang.
	t.Run("a cycle elsewhere in the graph", func(t *testing.T) {
		moved := service(t, func(s *storage.Snapshot) {
			s.Teams["fibggi2jc001"] = domain.Team{ID: "fibggi2jc001", Name: "fp8h2w6yc001", ParentID: "fibggi2jc002"}
			s.Teams["fibggi2jc002"] = domain.Team{ID: "fibggi2jc002", Name: "fp8h2w6yc002", ParentID: "fibggi2jc001"}
		})
		// Team1 holds an enabled binding, so the guard is what answers.
		if _, err := moved.SetTeamParent(t.Context(), area, publisher, "fibggi2juubk", "fibggi2jv0n4"); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("gave %v, want ErrConflict — the guard, not the cycle", err)
		}
	})

	// A cycle the walk actually enters. Every member of a cycle has its parent
	// inside the cycle, so descending from a team outside one can never reach it
	// — the only way in is for the moved team itself to be a member. Then the
	// walk descends into it, comes back to where it started, and the set having
	// only grown is what stops it.
	t.Run("the moved team is itself in a cycle", func(t *testing.T) {
		moved := service(t, func(s *storage.Snapshot) {
			// Team1 and a team below it, each the other's parent.
			s.Teams["fibggi2jc001"] = domain.Team{ID: "fibggi2jc001", Name: "fp8h2w6yc001", ParentID: "fibggi2juubk"}
			team1 := s.Teams["fibggi2juubk"]
			team1.ParentID = "fibggi2jc001"
			s.Teams["fibggi2juubk"] = team1
		})
		// On its own deadline, because the failure this guards against is a walk
		// that never returns — and a test that hangs blocks a suite for ten
		// minutes and then reports a timeout on whatever ran last. Removing the
		// guard should fail here, not everywhere.
		answered := make(chan error, 1)
		go func() {
			_, err := moved.SetTeamParent(context.Background(), area, publisher, "fibggi2juubk", "fibggi2jv0n4")
			answered <- err
		}()
		select {
		case err := <-answered:
			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("gave %v, want ErrConflict", err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("the walk did not settle on a graph that cycles through the team being moved")
		}
	})

	// A parent id naming no team. It is a leaf as far as the walk is concerned,
	// and it must not disturb a move that has nothing to do with it.
	t.Run("a parent id naming no team", func(t *testing.T) {
		moved := service(t, func(s *storage.Snapshot) {
			s.Teams["fibggi2jd001"] = domain.Team{ID: "fibggi2jd001", Name: "fp8h2w6yd001", ParentID: "fibggi2jnope"}
		})
		// Team2 holds nothing, so an unrelated dangling team must not refuse it.
		if _, err := moved.SetTeamParent(t.Context(), area, publisher, "fibggi2juxhc", "fibggi2jur5s"); err != nil {
			t.Fatalf("a dangling team elsewhere refused an unrelated move: %v", err)
		}
	})

	// And the ordering itself: a cycle the move *would* create is refused as a
	// cycle, before the binding guard is reached. Nothing asserted which of the
	// two answers a caller gets.
	t.Run("the move would create a cycle", func(t *testing.T) {
		moved := service(t, func(*storage.Snapshot) {})
		if _, err := moved.SetTeamParent(t.Context(), area, publisher, "fibggi2jur5s", "fibggi2juxhc"); !errors.Is(err, domain.ErrRejected) {
			t.Fatalf("gave %v, want ErrRejected — the cycle check answers first", err)
		}
	})
}
