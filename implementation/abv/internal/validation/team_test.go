package validation

import (
	"agentlabs.local/abv/domain"
	"errors"
	"testing"
)

func hierarchy() map[string]domain.Team {
	return map[string]domain.Team{
		"fibggi2jur5s": {ID: "fibggi2jur5s", Name: "RootTeam"},
		"fibggi2juubk": {ID: "fibggi2juubk", Name: "Team1", ParentID: "fibggi2jur5s"},
		"fibggi2juxhc": {ID: "fibggi2juxhc", Name: "Team2", ParentID: "fibggi2juubk"},
	}
}

// A team with a child cannot be deleted. Nor can one with a member, or one an
// assignment names: the rule is that anything depending on it blocks the delete,
// and nothing cascades.
func TestDeleteRefusesWhileAnythingDependsOnTheTeam(t *testing.T) {
	teams := hierarchy()
	none := map[string]domain.Assignment{}

	// Team1 has a child.
	if err := CheckTeamDeletion(teams, nil, none, "fibggi2juubk"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a team with a child was deletable: %v", err)
	}
	// Team2 is a leaf, so with nothing else depending on it the delete stands.
	if err := CheckTeamDeletion(teams, nil, none, "fibggi2juxhc"); err != nil {
		t.Fatalf("an empty leaf was not deletable: %v", err)
	}
	// A member blocks it.
	members := []domain.Membership{{TeamID: "fibggi2juxhc", HumanID: "fi7io4lvjqio"}}
	if err := CheckTeamDeletion(teams, members, none, "fibggi2juxhc"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a team with a member was deletable: %v", err)
	}
	// So does an assignment naming it.
	held := map[string]domain.Assignment{"A1": {ID: "A1", Recipient: domain.Recipient{Type: "group", ID: "fibggi2juxhc"}}}
	if err := CheckTeamDeletion(teams, nil, held, "fibggi2juxhc"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a team an assignment names was deletable: %v", err)
	}
	// An assignment to a human, not this group, does not block it.
	other := map[string]domain.Assignment{"A1": {ID: "A1", Recipient: domain.Recipient{Type: "user", ID: "fibggi2juxhc"}}}
	if err := CheckTeamDeletion(teams, nil, other, "fibggi2juxhc"); err != nil {
		t.Fatalf("a user assignment blocked a group's delete: %v", err)
	}
	if err := CheckTeamDeletion(teams, nil, none, "fy6x28qcdreo"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleting an absent team gave %v", err)
	}
}

// A move that would put a team inside its own subtree is refused at the write,
// not left for resolution to catch. Accepting one would leave the store holding
// a state that can never resolve.
func TestReparentRefusesACycleAtTheWrite(t *testing.T) {
	teams := hierarchy()
	// RootTeam under its own grandchild.
	if err := CheckTeamReparent(teams, "fibggi2jur5s", "fibggi2juxhc"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a cycle was accepted: %v", err)
	}
	// A team under itself.
	if err := CheckTeamReparent(teams, "fibggi2juubk", "fibggi2juubk"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a team became its own parent: %v", err)
	}
	// Downward moves are fine.
	if err := CheckTeamReparent(teams, "fibggi2juxhc", "fibggi2jur5s"); err != nil {
		t.Fatalf("a legal move was refused: %v", err)
	}
	// So is becoming a root.
	if err := CheckTeamReparent(teams, "fibggi2juxhc", ""); err != nil {
		t.Fatalf("promoting a team to a root was refused: %v", err)
	}
	if err := CheckTeamReparent(teams, "fibggi2juxhc", "fy6x28qcdreo"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("an absent parent gave %v", err)
	}
}

// The foreign keys the fold removed are now validation rules, and they still
// refuse the same things.
func TestTheLostForeignKeysAreNowRules(t *testing.T) {
	teams := hierarchy()
	// memberships -> teams: a membership cannot name a team that does not exist.
	absent := domain.Membership{TeamID: "fy6x28qcdreo", HumanID: "fi7io4lvjqio"}
	if err := CheckMembership(teams, absent); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a membership named an absent team: %v", err)
	}
	// A team cannot be created under a parent that does not exist.
	orphan := domain.Team{ID: "fy6x28qcdreo", Name: "Finance", ParentID: "fy6x28qcdrxx"}
	if err := CheckTeamCreation(teams, orphan); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a team was created under an absent parent: %v", err)
	}
	// Both halves of a membership must be ids, not names.
	if err := CheckMembership(teams, domain.Membership{TeamID: "Team1", HumanID: "fi7io4lvjqio"}); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("a team name passed where an id is required: %v", err)
	}
	if err := CheckMembership(teams, domain.Membership{TeamID: "fibggi2juubk", HumanID: "Maya"}); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("a human name passed where an id is required: %v", err)
	}
}
