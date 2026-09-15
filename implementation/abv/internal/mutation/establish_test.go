package mutation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"errors"
	"testing"
	"time"
)

// The holder rules can only be reached in an area that has no root yet — in the
// worked fixture, "one root per area" refuses first, which is the right order.
func establishFixture(t *testing.T) (*mutation.Service, domain.Area) {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	fixture := lab.TeamFINC17(area)
	// Strip the seeded root and everything resting on it, so establishment has
	// somewhere to land.
	fixture.Snapshot.TrustedRoots = map[string]bool{}
	fixture.Snapshot.Contents = map[domain.GrantKey]domain.GrantContent{}
	fixture.Snapshot.Controls = map[string]domain.GrantControl{}
	fixture.Snapshot.Assignments = map[string]domain.Assignment{}

	provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = provider.Close() })
	statusAdmin, err := lab.NewAssignmentStatusAdministration(area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	admin := &lab.GrantRevisionAdministration{RoleAdministration: &lab.RoleAdministration{AssignmentStatusAdministration: statusAdmin}}
	service, err := mutation.New(provider, admin, &fixedClock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	return service, area
}

// A root must be held by a top-level team. rootRoute refuses a parented holder
// at resolution; establishment refuses it at the write, so the record is never
// stored in a shape resolution would reject.
func TestEstablishRootRequiresATopLevelHolder(t *testing.T) {
	service, area := establishFixture(t)
	issuer := lab.TeamFINC17(area).Issuer

	// Team1 has a parent, so it cannot hold a root.
	if _, _, err := service.EstablishRoot(t.Context(), area, issuer, "fibggi2juubk"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a parented holder gave %v, want ErrRejected", err)
	}
	if _, _, err := service.EstablishRoot(t.Context(), area, issuer, "fp8h2w6yzzzz"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("an absent holder gave %v, want ErrNotFound", err)
	}
	if _, _, err := service.EstablishRoot(t.Context(), area, issuer, "not-a-snowflake"); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("a malformed holder gave %v, want ErrMalformed", err)
	}

	// RootTeam has no parent, so it can — and the four writes land together.
	grant, content, err := service.EstablishRoot(t.Context(), area, issuer, "fibggi2jur5s")
	if err != nil {
		t.Fatal(err)
	}
	if !grant.TrustedRoot || grant.Status != "enabled" {
		t.Fatalf("established grant = %#v, want a trusted root", grant)
	}
	if content.Revision != 1 || content.ParentGrantID != "" || content.Permissions != nil || len(content.Scope) != 0 {
		t.Fatalf("root content = %#v, want revision 1 with no parent, no permissions and an empty scope", content)
	}
	// And the holder's assignment exists, which is what makes the root reachable
	// — three records would be an incomplete setup supplying no authority.
	page, err := service.ListAssignments(t.Context(), area, issuer, domain.AssignmentFilter{GrantID: grant.ID})
	if err != nil || page.Total != 1 || page.Assignments[0].Recipient.ID != "fibggi2jur5s" {
		t.Fatalf("holder assignment = %#v total=%d err=%v", page.Assignments, page.Total, err)
	}

	// Exactly one root per area.
	if _, _, err := service.EstablishRoot(t.Context(), area, issuer, "fibggi2jur5s"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a second root gave %v, want ErrConflict", err)
	}
}
