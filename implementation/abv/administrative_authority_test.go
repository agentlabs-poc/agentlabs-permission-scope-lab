package abv_test

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/lab"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// priya holds lab.TeamAdminGrant: `auth:group::write` scoped to one team, and
// nothing else. maya — teamFixture — holds every group verb unscoped.
var priya = domain.FixtureContext{Name: lab.TeamAdminFixtureContext}

// The C17 team priya's grant names, and the FIN team it does not.
const (
	hersC17    = "fibggi2juxhc"
	notHersFIN = "fibggi2juubk"
)

// A holder scoped to one team administers that team and no other — Q-155 /
// ADMIN-007, and the containment property is the scope predicate rather than a
// rule written anywhere.
//
// Both halves matter and the second is the one worth the test: an implementation
// that resolved the permission and ignored the predicate passes the first.
func TestAdministrativeScopeBoundsTheTeamNotThePermission(t *testing.T) {
	api, area := openTeamLab(t)
	if err := api.AddMember(t.Context(), area, priya, hersC17, "fi7io4lvk35s"); err != nil {
		t.Fatalf("the holder of write over C17 could not add to C17: %v", err)
	}
	if err := api.RemoveMember(t.Context(), area, priya, hersC17, "fi7io4lvk35s"); err != nil {
		t.Fatalf("the holder of write over C17 could not remove from C17: %v", err)
	}
	// Same permission, same holder, a different team. The route carries
	// team=fibggi2juxhc, the request carries team=fibggi2juubk, and one value per
	// key is the whole of the refusal.
	if err := api.AddMember(t.Context(), area, priya, notHersFIN, "fi7io4lvk35s"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("adding to a team outside the grant's scope gave %v, want ErrRejected", err)
	}
	// And the tenant administrator, whose route carries no predicate at all,
	// satisfies both — because a route with no predicates has nothing to fail.
	if err := api.AddMember(t.Context(), area, teamFixture, notHersFIN, "fi7io4lvk35s"); err != nil {
		t.Fatalf("the tenant administrator could not add to FIN: %v", err)
	}
}

// Q-092 made create, write and delete three permissions rather than one
// authority, and the resolution honours that: priya's grant names write alone.
//
// It is the same call with a different permission, which is the point — the
// twenty-eight gates collapsed into one evaluation, and what each operation
// contributes is which permission it requires.
func TestAdministrativeVerbsAreSeparateAuthorities(t *testing.T) {
	api, area := openTeamLab(t)
	if _, err := api.CreateTeam(t.Context(), area, priya, "Subteam", hersC17); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a write-only holder created a subteam: %v", err)
	}
	if err := api.DeleteTeam(t.Context(), area, priya, hersC17); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a write-only holder deleted a team: %v", err)
	}
	// Whereas the holder of every verb does both. The delete is of a team it
	// creates here, because C17 has dependants.
	created, err := api.CreateTeam(t.Context(), area, teamFixture, "Spare", hersC17)
	if err != nil {
		t.Fatal(err)
	}
	if err := api.DeleteTeam(t.Context(), area, teamFixture, created.ID); err != nil {
		t.Fatal(err)
	}
}

// Creating a top-level team has no bounding team, so only an unscoped route
// satisfies it. A scoped holder can create inside its own team and cannot create
// a peer of the tenant's root.
func TestOnlyAnUnscopedRouteCreatesATopLevelTeam(t *testing.T) {
	api, area := openTeamLab(t)
	if _, err := api.CreateTeam(t.Context(), area, priya, "Rootish", ""); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a scoped holder created a top-level team: %v", err)
	}
	if _, err := api.CreateTeam(t.Context(), area, teamFixture, "Rootish", ""); err != nil {
		t.Fatalf("the tenant administrator could not create a top-level team: %v", err)
	}
}

// Administrative authority is an ordinary grant, so a disabled one supplies no
// authority — no administrative-specific lifecycle, and nothing a caller has to
// remember to revoke separately.
//
// The grant is disabled in the fixture rather than through an operation because
// the grant-status gate is still one of the fixtures this change deliberately did
// not touch, and widening it to admit a second id would be a test arranging its
// own permission.
func TestADisabledAdministrativeGrantSuppliesNoAuthority(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	fixture := lab.TeamFINC17(area)
	administrative, err := lab.AuthAdministration(area)
	if err != nil {
		t.Fatal(err)
	}
	control := administrative.Controls[lab.TeamAdminGrant]
	control.Status = "disabled"
	administrative.Controls[lab.TeamAdminGrant] = control

	facade := administeredFacade(t, area, fixture, administrative)
	if err := facade.AddMember(t.Context(), area, priyaIdentity, hersC17, "fi7io4lvk35s"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a disabled administrative grant still authorized a membership write: %v", err)
	}
	// And the tenant administrator is untouched: one grant is disabled, not the
	// authority model.
	if err := facade.AddMember(t.Context(), area, fixture.Issuer, hersC17, "fi7io4lvk35s"); err != nil {
		t.Fatalf("disabling one administrative grant took away another: %v", err)
	}
}

// priyaIdentity is the same human lab.TeamAdminFixtureContext acts as, named
// directly where a test drives the facade rather than the lab application.
var priyaIdentity = domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjwu8"}, HumanID: "fi7io4lvjwu8"}

// administeredFacade opens a facade over a store holding both chains.
func administeredFacade(t *testing.T, area domain.Area, fixture lab.TeamFINC17Case, administrative storage.Snapshot) *abv.Facade {
	t.Helper()
	provider, err := lab.CreateSQLite(t.Context(), filepath.Join(t.TempDir(), "administered.db"),
		[]storage.Snapshot{fixture.Snapshot, administrative})
	if err != nil {
		t.Fatal(err)
	}
	admin, err := lab.NewAdministration(area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	facade, err := abv.New(provider, admin, clock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = facade.Close() })
	return facade
}

// A store that does not know where the Auth chain lives refuses every team
// write, rather than resolving `auth:group::*` against the application's own
// chain. Fail-closed, and the answer says the store cannot do it rather than
// that the caller may not.
func TestAStoreWithNoAdministrativeChainRefusesEveryTeamWrite(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	fixture := lab.TeamFINC17(area)
	// The business records alone. Nothing declares a platform-boundary record, so
	// nothing names the platform namespace, so there is no administrative chain to
	// resolve against.
	provider, err := lab.CreateSQLite(t.Context(), filepath.Join(t.TempDir(), "business-only.db"), []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	admin, err := lab.NewAdministration(area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	facade, err := abv.New(provider, admin, clock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	defer facade.Close()
	if _, err := facade.CreateTeam(t.Context(), area, fixture.Issuer, "Finance", ""); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("a store with no administrative chain gave %v, want ErrUnsupported", err)
	}
}

// A platform scope key's value names one of Auth's own records, so it is resolved
// on the way in rather than carried as an opaque string — Q-156 / SCOPE-010,
// refining Q-148 for this one case.
//
// An application's scope values stay opaque, and must: Auth has no way to know
// what "dept=FIN" denotes. A `team` value it can check, and a grant naming a team
// that does not exist is a grant that will never authorize anything — better
// refused at the write than stored as a route nobody can explain.
func TestAPlatformScopeValueIsResolvedRatherThanTrusted(t *testing.T) {
	facade, authArea := administrativeFacade(t)
	maya := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}
	propose := func(value string) error {
		_, _, err := facade.CreateGrant(t.Context(), authArea, maya, lab.TenantAdminGrant, domain.GrantContent{
			Permissions: []string{lab.GroupWrite}, Scope: map[string]string{"team": value},
		})
		return err
	}
	// A well-formed id that names no team. Opaque would have accepted it.
	if err := propose("fibggi2jzzzz"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a team id naming no team gave %v, want ErrRejected", err)
	}
	// Not an id at all, which is a malformed value rather than a missing record.
	if err := propose("Finance"); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("a value that cannot be a team id gave %v, want ErrMalformed", err)
	}
	// $self binds to a human, and a team is not a human. The token is legal in an
	// application's scope and meaningless here.
	if err := propose(domain.SelfToken); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("$self as a team gave %v, want ErrRejected", err)
	}
	if err := propose(hersC17); err != nil {
		t.Fatalf("a value naming a real team was refused: %v", err)
	}
}

// The wart Q-155 records rather than hides: narrowing *accepts* a contradictory
// re-scope, and the safety is that the route authorizes nothing rather than that
// the write is refused.
//
// This is what makes sideways escalation impossible without anything enforcing
// it. Predicates accumulate conjunctively and a request carries one value per
// key, so a grant scoped to team B beneath a parent scoped to team A produces a
// route demanding both — which no request can satisfy.
func TestASidewaysAdministrativeGrantIsWrittenAndAuthorizesNothing(t *testing.T) {
	facade, authArea := administrativeFacade(t)
	maya := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}
	// A subteam of the administrators' team to hold the sideways grant. Its parent
	// holds the grant being narrowed, which is what a chain's first hop requires.
	holder, err := facade.CreateTeam(t.Context(), authArea, maya, "Sideways", lab.TeamAdminsTeam)
	if err != nil {
		t.Fatal(err)
	}
	// Beneath a grant scoped {team: C17}, scoped to FIN instead. Accepted.
	sideways, content, err := facade.CreateGrant(t.Context(), authArea, maya, lab.TeamAdminGrant, domain.GrantContent{
		Permissions: []string{lab.GroupWrite}, Scope: map[string]string{"team": notHersFIN},
	})
	if err != nil {
		t.Fatalf("the contradictory re-scope was refused, which the handbook records as accepted: %v", err)
	}
	raw, err := json.Marshal(domain.Assignment{
		Version: "1", ID: "fm5b7t4psid0", GrantID: sideways.ID, GrantRevision: content.Revision,
		Recipient: domain.Recipient{Type: "group", ID: holder.ID}, Status: "enabled",
	})
	if err != nil {
		t.Fatal(err)
	}
	// And it resolves — to a route carrying both values of one key.
	diagnostic, err := facade.CheckAssignment(t.Context(), authArea, raw)
	if err != nil {
		t.Fatalf("the sideways grant did not resolve at all: %v", err)
	}
	values := map[string]bool{}
	for _, predicate := range diagnostic.Route.Predicates {
		if predicate.Key == "team" {
			values[predicate.Value] = true
		}
	}
	if !values[hersC17] || !values[notHersFIN] || len(values) != 2 {
		t.Fatalf("route predicates = %#v, want team pinned to both %s and %s", diagnostic.Route.Predicates, hersC17, notHersFIN)
	}
	// Which is unsatisfiable, because a request names one value per key. Nothing
	// enforces that; it is what matching every predicate against one material map
	// computes.
}

// administrativeFacade opens a facade bound to the platform namespace, where the
// administrative chain is. Team writes work through it because a tenant's teams
// are the tenant's rather than an application's, and the chain in hand is its own
// administrative chain.
func administrativeFacade(t *testing.T) (*abv.Facade, domain.Area) {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	authArea, err := domain.NewArea("acme", lab.PlatformNamespace)
	if err != nil {
		t.Fatal(err)
	}
	fixture := lab.TeamFINC17(area)
	administrative, err := lab.AuthAdministration(area)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := lab.CreateSQLite(t.Context(), filepath.Join(t.TempDir(), "administrative.db"),
		[]storage.Snapshot{fixture.Snapshot, administrative})
	if err != nil {
		t.Fatal(err)
	}
	// RoleAdministration rather than the narrower Administration, because these
	// tests create grants — and grant administration is still one of the fixtures.
	status, err := lab.NewAssignmentStatusAdministration(authArea, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	facade, err := abv.New(provider, &lab.RoleAdministration{AssignmentStatusAdministration: status},
		clock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = facade.Close() })
	return facade, authArea
}

// The other half of failing closed, and the half a mutation check found nothing
// guarding: a platform namespace that names an *application*.
//
// `attachAdministrative` treats "the area I am in is the platform namespace" as
// "the snapshot in hand is the administrative chain" — which is right when the
// namespace is the platform's and catastrophic when it is an application's, because
// the caller then resolves `<app>:group::write` against the application's own root.
// An application's root holder would silently become the tenant's team
// administrator, which is the leak Q-151's namespace slice exists to prevent.
//
// So the branch checks rather than assumes: a namespace owning no platform-boundary
// permission is not a platform namespace, and this store cannot answer an
// administrative question at all.
func TestANamespaceThatNamesAnApplicationIsNotAnAdministrativeChain(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	fixture := lab.TeamFINC17(area)
	snapshot := fixture.Snapshot
	// The application registered a permission spelled like Auth's, which is its
	// right — the namespace is the application's own. And the acting human is in the
	// team holding the application's root, so an application-chain resolve would
	// succeed if one were attempted.
	snapshot.Catalog.Permissions["hrms:group::write"] = domain.PermissionDefinition{
		ID: "hrms:group::write", Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms",
	}
	snapshot.Memberships = append(snapshot.Memberships, domain.Membership{TeamID: "fibggi2jur5s", HumanID: "fi7io4lvjqio"})
	path := filepath.Join(t.TempDir(), "misconfigured.db")
	provider, err := lab.CreateSQLite(t.Context(), path, []storage.Snapshot{snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.Close(); err != nil {
		t.Fatal(err)
	}
	registry, err := lab.NewFixedRegistry(area)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := lab.NewAdministration(area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	// Reopened with the namespace pointed at the application, which is the
	// misconfiguration. A deployment names this value; nothing stops it naming this.
	facade, err := abv.OpenSQLiteWithOptions(t.Context(), path, admin,
		clock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)},
		abv.Options{Registry: registry, PlatformNamespace: "hrms"})
	if err != nil {
		t.Fatal(err)
	}
	defer facade.Close()
	if err := facade.AddMember(t.Context(), area, fixture.Issuer, hersC17, "fi7io4lvk35s"); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("an application treated as the platform namespace gave %v, want ErrUnsupported", err)
	}

	// And the same misconfiguration one area over: a namespace naming some *other*
	// application rather than this one. The chain is then read from that area, and
	// the answer must still be "this store cannot say" rather than a denial —
	// ErrRejected here would tell an operator their administrator lacks authority,
	// when what they actually have is a typo in a namespace.
	elsewhere, err := abv.OpenSQLiteWithOptions(t.Context(), path, admin,
		clock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)},
		abv.Options{Registry: registry, PlatformNamespace: "crm"})
	if err != nil {
		t.Fatal(err)
	}
	defer elsewhere.Close()
	if err := elsewhere.AddMember(t.Context(), area, fixture.Issuer, hersC17, "fi7io4lvk35s"); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("a namespace naming another application gave %v, want ErrUnsupported", err)
	}
}
