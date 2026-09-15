package abv_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/storage"
	"errors"
	"testing"
)

const (
	rootTeam  = "fibggi2jur5s"
	team1     = "fibggi2juubk"
	team2     = "fibggi2juxhc"
	adminTeam = "fibggi2jv0n4"
	maya      = "fi7io4lvjqio"
	nutan     = "fi7io4lvjwu8"
)

func seededTeams(t *testing.T) (*lab.RoleAdministration, domain.Area, storage.Snapshot) {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	return nil, area, lab.TeamFINC17(area).Snapshot
}

// A team's identity is its id; the readable name is a label beside it, exactly
// as a role's is.
func TestTeamCarriesAnIDAndASeparateName(t *testing.T) {
	_, _, snapshot := seededTeams(t)
	team, ok := snapshot.Teams[team1]
	if !ok {
		t.Fatalf("fp8h2w6y5iv8 absent; have %v", snapshot.Teams)
	}
	if team.ID != team1 || team.Name != "fp8h2w6y5iv8" || team.ParentID != rootTeam {
		t.Fatalf("team = %#v", team)
	}
	// A root's parent is the empty string — the real value, not an omission.
	if root := snapshot.Teams[rootTeam]; root.ParentID != "" {
		t.Fatalf("a root team has a parent: %#v", root)
	}
}

// The parent is stored as an id, not a name. A name is editable, and a hierarchy
// built on names would break the moment a team were renamed.
func TestTeamParentIsAnIDNotAName(t *testing.T) {
	_, _, snapshot := seededTeams(t)
	for id, team := range snapshot.Teams {
		if team.ParentID == "" {
			continue
		}
		parent, ok := snapshot.Teams[team.ParentID]
		if !ok {
			t.Fatalf("team %s names a parent that is not a team id: %q", id, team.ParentID)
		}
		if parent.Name == team.ParentID {
			t.Fatalf("team %s appears to name its parent by name", id)
		}
	}
}

// One human in two teams is two memberships, not one record with a list. They
// are independent facts that change independently.
func TestMembershipsAreIndependentRecords(t *testing.T) {
	_, _, snapshot := seededTeams(t)
	count := 0
	for _, m := range snapshot.Memberships {
		if m.HumanID == maya {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("maya has %d memberships, want 2 (%v)", count, snapshot.Memberships)
	}
}

// Every id in both records is a base-36 Snowflake — the team's issued here, the
// human's by the auth service, one spelling for both.
func TestEveryTeamAndMemberIDIsBase36(t *testing.T) {
	_, _, snapshot := seededTeams(t)
	base36 := func(s string) bool {
		if s == "" || len(s) > 13 {
			return false
		}
		for _, r := range s {
			if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'z') {
				return false
			}
		}
		return true
	}
	for id, team := range snapshot.Teams {
		if !base36(id) {
			t.Fatalf("team id is not base-36: %q", id)
		}
		if team.ParentID != "" && !base36(team.ParentID) {
			t.Fatalf("parent id is not base-36: %q", team.ParentID)
		}
		// The name is deliberately NOT constrained — it is a label.
		if team.Name == "" {
			t.Fatalf("team %s has no name", id)
		}
	}
	for _, m := range snapshot.Memberships {
		if !base36(m.TeamID) || !base36(m.HumanID) {
			t.Fatalf("membership is not base-36 on both halves: %#v", m)
		}
	}
}

var _ = errors.Is
var _ = nutan
var _ = team2
var _ = adminTeam
