package abv_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"context"
	"errors"
	"path/filepath"
	"testing"
)

type establisher interface {
	EstablishRoot(context.Context, domain.Area, domain.FixtureContext, string) (domain.Grant, domain.GrantContent, error)
	GetGrant(context.Context, domain.Area, domain.FixtureContext, string, int64) (domain.Grant, domain.GrantContent, error)
	ListGrants(context.Context, domain.Area, domain.FixtureContext, domain.GrantFilter) (domain.GrantPage, error)
	ListAssignments(context.Context, domain.Area, domain.FixtureContext, domain.AssignmentFilter) (domain.AssignmentPage, error)
	CreateTeam(context.Context, domain.Area, domain.FixtureContext, string, string) (domain.Team, error)
}

func openEstablishLab(t *testing.T) (establisher, domain.Area) {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "establish.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := lab.Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeConnection() })
	e, ok := api.(establisher)
	if !ok {
		t.Fatalf("lab application does not expose establishment: %T", api)
	}
	return e, area
}

// Until now no operation could create a trusted root: the only writer of
// TrustedRoot was a lab fixture, so every grant traced back to seeded data and
// the creation lifecycle had no beginning. This is that gap closing.
func TestEstablishRootCreatesARootNothingElseCould(t *testing.T) {
	api, area := openEstablishLab(t)

	// The fixture already has a root, and a root is the ceiling for its whole
	// area — exactly one may exist.
	if _, _, err := api.EstablishRoot(t.Context(), area, teamFixture, "fibggi2jur5s"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("establishing a second root gave %v, want ErrConflict", err)
	}
}
