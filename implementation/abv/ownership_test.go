package abv_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"context"
	"errors"
	"path/filepath"
	"testing"
)

type ownerAPI interface {
	AddOwner(context.Context, domain.Area, domain.FixtureContext, string, string) error
	RemoveOwner(context.Context, domain.Area, domain.FixtureContext, string, string) error
	ListOwners(context.Context, domain.Area, domain.FixtureContext, domain.OwnerFilter) (domain.OwnerPage, error)
	ListMembers(context.Context, domain.Area, domain.FixtureContext, domain.MemberFilter) (domain.MemberPage, error)
}

func openOwnerLab(t *testing.T) (ownerAPI, domain.Area) {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "owners.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := lab.Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeConnection() })
	writer, ok := api.(ownerAPI)
	if !ok {
		t.Fatalf("lab application does not expose the ownership operations: %T", api)
	}
	return writer, area
}

// Add is add-only and remove says whether it did anything — the rule membership
// writes already hold, because a caller that cannot tell the difference cannot
// report it either.
func TestOwnershipWritesAreExplicitAboutWhetherTheyDidAnything(t *testing.T) {
	api, area := openOwnerLab(t)

	if err := api.AddOwner(t.Context(), area, teamFixture, team1, maya); err != nil {
		t.Fatal(err)
	}
	if err := api.AddOwner(t.Context(), area, teamFixture, team1, maya); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("adding an existing owner gave %v, want ErrConflict", err)
	}
	if err := api.RemoveOwner(t.Context(), area, teamFixture, team1, maya); err != nil {
		t.Fatal(err)
	}
	if err := api.RemoveOwner(t.Context(), area, teamFixture, team1, maya); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("removing an absent owner gave %v, want ErrNotFound", err)
	}

	// Both halves have to exist.
	if err := api.AddOwner(t.Context(), area, teamFixture, "fp8h2w6yzzzz", maya); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("an absent team gave %v, want ErrNotFound", err)
	}
	if err := api.AddOwner(t.Context(), area, teamFixture, team1, "not-a-snowflake"); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("a malformed maya id gave %v, want ErrMalformed", err)
	}
}

// A team may have several owners, and one maya may own several teams. Q-099
// talks about "Team1's owners", plural.
func TestOwnersArePluralAndListAnswersBothDirections(t *testing.T) {
	api, area := openOwnerLab(t)
	for _, o := range [][2]string{{team1, maya}, {team1, nutan}, {team2, maya}} {
		if err := api.AddOwner(t.Context(), area, teamFixture, o[0], o[1]); err != nil {
			t.Fatalf("%v: %v", o, err)
		}
	}

	byTeam, err := api.ListOwners(t.Context(), area, teamFixture, domain.OwnerFilter{TeamID: team1})
	if err != nil || byTeam.Total != 2 {
		t.Fatalf("team1 owners = %#v total=%d err=%v, want 2", byTeam.Owners, byTeam.Total, err)
	}
	byHuman, err := api.ListOwners(t.Context(), area, teamFixture, domain.OwnerFilter{HumanID: maya})
	if err != nil || byHuman.Total != 2 {
		t.Fatalf("this maya owns %#v total=%d err=%v, want 2", byHuman.Owners, byHuman.Total, err)
	}

	// Neither direction, or both, is not a question anyone asks.
	if _, err := api.ListOwners(t.Context(), area, teamFixture, domain.OwnerFilter{}); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("unfiltered gave %v, want ErrMalformed", err)
	}
	if _, err := api.ListOwners(t.Context(), area, teamFixture, domain.OwnerFilter{TeamID: team1, HumanID: maya}); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("both filters gave %v, want ErrMalformed", err)
	}

	// Removing the last owner is permitted: team administration is tenant-wide,
	// so a team with no owners is still administrable and there is a way back.
	for _, h := range []string{maya, nutan} {
		if err := api.RemoveOwner(t.Context(), area, teamFixture, team1, h); err != nil {
			t.Fatal(err)
		}
	}
	empty, err := api.ListOwners(t.Context(), area, teamFixture, domain.OwnerFilter{TeamID: team1})
	if err != nil || empty.Total != 0 {
		t.Fatalf("after removing every owner: %#v total=%d err=%v", empty.Owners, empty.Total, err)
	}
}

// Q-099's whole point: ownership is authority to administer a team and nothing
// else. It is not membership, and it carries none of the team's authority.
func TestOwningATeamIsNotMembershipAndCarriesNoAuthority(t *testing.T) {
	api, area := openOwnerLab(t)

	// The nutan is in no team in this fixture.
	before, err := api.ListMembers(t.Context(), area, teamFixture, domain.MemberFilter{HumanID: nutan})
	if err != nil {
		t.Fatal(err)
	}
	if err := api.AddOwner(t.Context(), area, teamFixture, team1, nutan); err != nil {
		t.Fatal(err)
	}
	after, err := api.ListMembers(t.Context(), area, teamFixture, domain.MemberFilter{HumanID: nutan})
	if err != nil {
		t.Fatal(err)
	}
	if after.Total != before.Total {
		t.Fatalf("owning a team changed this maya's memberships: %d -> %d", before.Total, after.Total)
	}

	// And the two relationships are independent in the other direction too: the
	// fixture's members of team1 did not become its owners.
	members, err := api.ListMembers(t.Context(), area, teamFixture, domain.MemberFilter{TeamID: team1})
	if err != nil {
		t.Fatal(err)
	}
	owners, err := api.ListOwners(t.Context(), area, teamFixture, domain.OwnerFilter{TeamID: team1})
	if err != nil {
		t.Fatal(err)
	}
	if members.Total == 0 {
		t.Fatal("the fixture has no members of team1; this test would prove nothing")
	}
	if owners.Total != 1 || owners.Owners[0].HumanID != nutan {
		t.Fatalf("owners = %#v, want only the one just added — members are not owners", owners.Owners)
	}
}
