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

func regressService(t *testing.T, area domain.Area, snapshots []storage.Snapshot) *mutation.Service {
	t.Helper()
	provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/regress.db", snapshots)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = provider.Close() })
	fixture := lab.TeamFINC17(area)
	status, err := lab.NewAssignmentStatusAdministration(area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	service, err := mutation.New(provider, &lab.RoleAdministration{AssignmentStatusAdministration: status},
		&fixedClock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

var maya = domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}
var priyaID = domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjwu8"}, HumanID: "fi7io4lvjwu8"}

// A1
func TestRegressDeleteTeamSeesAdministrativeAssignments(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	administrative, err := lab.AuthAdministration(area)
	if err != nil {
		t.Fatal(err)
	}
	s := regressService(t, area, []storage.Snapshot{lab.TeamFINC17(area).Snapshot, administrative})
	// Empty the administrators' team of everything the business area can see, so the
	// membership rule is not what answers.
	if err := s.RemoveMember(t.Context(), area, maya, lab.TeamAdminsTeam, "fi7io4lvjwu8"); err != nil {
		t.Fatal(err)
	}
	// fm5b7t4pau03 is an enabled administrative assignment naming it, and it is in
	// another area — so the business snapshot's assignments do not mention it.
	if err := s.DeleteTeam(t.Context(), area, maya, lab.TeamAdminsTeam); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleted a team an administrative assignment names: %v", err)
	}
}

// A2
func TestRegressReparentSeesAdministrativeBindings(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	administrative, err := lab.AuthAdministration(area)
	if err != nil {
		t.Fatal(err)
	}
	s := regressService(t, area, []storage.Snapshot{lab.TeamFINC17(area).Snapshot, administrative})
	if _, err := s.SetTeamParent(t.Context(), area, maya, lab.TeamAdminsTeam, "fibggi2jur5s"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("moved a team an enabled administrative binding sits at: %v", err)
	}
}

// B2
func TestRegressPromotingToTopLevelNeedsAnUnscopedRoute(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	administrative, err := lab.AuthAdministration(area)
	if err != nil {
		t.Fatal(err)
	}
	// A leaf nothing binds, and a scoped administrative grant over exactly it. The
	// holder is C17, whose parent holds the tenant administrator's grant — which is
	// what a chain's first hop requires.
	const leaf = "fibggi2ja777"
	administrative.Teams[leaf] = domain.Team{ID: leaf, Name: "fp8h2w6ya777", ParentID: "fibggi2juubk"}
	const scoped = "fk3x9r2mau06"
	administrative.Controls[scoped] = domain.GrantControl{Version: "1", ID: scoped, Status: "enabled"}
	administrative.Contents[domain.GrantKey{ID: scoped, Revision: 1}] = domain.GrantContent{
		Version: "1", GrantID: scoped, Revision: 1, ParentGrantID: lab.TenantAdminGrant,
		Permissions: []string{lab.GroupWrite}, Scope: map[string]string{"team": leaf},
	}
	administrative.Assignments["fm5b7t4pau06"] = domain.Assignment{
		Version: "1", ID: "fm5b7t4pau06", GrantID: scoped, GrantRevision: 1,
		Recipient: domain.Recipient{Type: "group", ID: "fibggi2juxhc"}, Status: "enabled",
	}
	s := regressService(t, area, []storage.Snapshot{lab.TeamFINC17(area).Snapshot, administrative})
	// She may write the team her grant names.
	if err := s.AddMember(t.Context(), area, priyaID, leaf, "fi7io4lvk35s"); err != nil {
		t.Fatalf("the scoped holder could not write the team her grant names: %v", err)
	}
	if err := s.RemoveMember(t.Context(), area, priyaID, leaf, "fi7io4lvk35s"); err != nil {
		t.Fatal(err)
	}
	// She may not make it top level: that is the end state CreateTeam reserves for
	// an unscoped route, and a move must not be a second door to it.
	if _, err := s.SetTeamParent(t.Context(), area, priyaID, leaf, ""); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a scoped holder promoted a team to top level: %v", err)
	}
	// And the tenant administrator may.
	if _, err := s.SetTeamParent(t.Context(), area, maya, leaf, ""); err != nil {
		t.Fatalf("the tenant administrator could not promote a team: %v", err)
	}
}

// B1
func TestRegressAnApplicationCannotBeItsOwnAdministrativeChain(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	snapshot := fixture.Snapshot
	// What a catalog read returns after the platform key exists: the key inside
	// the application's own catalog, at the platform boundary.
	snapshot.Catalog.Scopes["team"] = domain.ScopeDefinition{Key: "team", Boundary: domain.PlatformBoundary}
	snapshot.Catalog.Permissions["hrms:group::write"] = domain.PermissionDefinition{
		ID: "hrms:group::write", Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms",
	}
	// The root holder, so an application-chain resolve would succeed.
	snapshot.Memberships = append(snapshot.Memberships, domain.Membership{TeamID: "fibggi2jur5s", HumanID: "fi7io4lvjqio"})
	s := regressService(t, area, []storage.Snapshot{snapshot})
	if err := s.AddMember(t.Context(), area, maya, "fibggi2juxhc", "fi7io4lvk35s"); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("an application authorized a team write against its own chain: %v", err)
	}
}
