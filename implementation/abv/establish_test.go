package abv_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"testing"
)

type establisher interface {
	EstablishRoot(context.Context, domain.Area, domain.FixtureContext, string) (domain.Grant, domain.GrantContent, error)
	GetGrant(context.Context, domain.Area, domain.FixtureContext, string, int64) (domain.Grant, domain.GrantContent, error)
	ListGrants(context.Context, domain.Area, domain.FixtureContext, domain.GrantFilter) (domain.GrantPage, error)
	ListAssignments(context.Context, domain.Area, domain.FixtureContext, domain.AssignmentFilter) (domain.AssignmentPage, error)
	CreateTeam(context.Context, domain.Area, domain.FixtureContext, string, string) (domain.Team, error)
	EstablishAuthRoot(context.Context, domain.Area, domain.FixtureContext, string) (domain.Grant, domain.GrantContent, error)
	CreateGrant(context.Context, domain.Area, domain.FixtureContext, string, domain.GrantContent) (domain.Grant, domain.GrantContent, error)
	CheckAssignment(context.Context, domain.Area, []byte) (domain.Diagnostic, error)
	DeleteGrant(context.Context, domain.Area, domain.FixtureContext, string) error
	DeleteAssignment(context.Context, domain.Area, domain.FixtureContext, string) error
}

func openEstablishLab(t *testing.T, scenario string) (establisher, domain.Area) {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "establish.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, scenario, path); err != nil {
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
	api, area := openEstablishLab(t, "team-fin-c17")

	// The fixture already has a root, and a root is the ceiling for its whole
	// area — exactly one may exist.
	if _, _, err := api.EstablishRoot(t.Context(), area, teamFixture, "fibggi2jur5s"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("establishing a second root gave %v, want ErrConflict", err)
	}
}

// Establishing a root is only half the claim. The other half is that the root
// supports authority afterwards — that a child hung from it resolves, all the
// way back to the assignment written in the same transaction.
//
// This is the test that would have caught the hole. Root content names no
// permission source, and selectedPermissions looked the empty role key up in
// the role map, missed, and returned ErrRejected: every established root was
// written correctly and supported nothing at all.
func TestAnEstablishedRootSupportsAChild(t *testing.T) {
	api, area := openEstablishLab(t, "tenant-genesis")

	root, content, err := api.EstablishRoot(t.Context(), area, teamFixture, "fibggi2jur5s")
	if err != nil {
		t.Fatal(err)
	}
	if !root.TrustedRoot || content.Permissions != nil || content.ParentGrantID != "" {
		t.Fatalf("established root = %#v / %#v", root, content)
	}
	child, childContent, err := api.CreateGrant(t.Context(), area, teamFixture, root.ID, domain.GrantContent{
		Permissions: []string{lab.PayslipRead}, Scope: map[string]string{"dept": "FIN"},
	})
	if err != nil {
		t.Fatal(err)
	}
	proposal := domain.Assignment{
		Version: "1", ID: "fm5b7t4pan0d", GrantID: child.ID, GrantRevision: childContent.Revision,
		// Team1 is a child of the root's holder, which is what makes the chain
		// climb from the recipient to the root.
		Recipient: domain.Recipient{Type: "group", ID: "fibggi2juubk"}, Status: "enabled",
	}
	raw, err := json.Marshal(proposal)
	if err != nil {
		t.Fatal(err)
	}
	diagnostic, err := api.CheckAssignment(t.Context(), area, raw)
	if err != nil {
		t.Fatalf("a child of an established root did not resolve: %v", err)
	}
	// The ceiling came from the catalog, and the route reaches the root's own
	// assignment — the one establishment wrote beside the grant.
	if !slices.Contains(diagnostic.Route.Permissions, lab.PayslipRead) {
		t.Fatalf("route lost the selected permission: %#v", diagnostic.Route)
	}
	holder, err := api.ListAssignments(t.Context(), area, teamFixture, domain.AssignmentFilter{GrantID: root.ID})
	if err != nil || holder.Total != 1 {
		t.Fatalf("root assignment = %#v total=%d err=%v", holder.Assignments, holder.Total, err)
	}
	if !slices.Contains(diagnostic.Route.AssignmentIDs, holder.Assignments[0].ID) {
		t.Fatalf("route %v did not trace to the root assignment %s", diagnostic.Route.AssignmentIDs, holder.Assignments[0].ID)
	}

	// That the ceiling is the catalog *sliced by namespace* is the lineage
	// package's claim, and TestAnApplicationRootDoesNotCarryPlatformPermissions
	// makes it against a planted platform permission. Here the point is only
	// that an established root supports anything at all.
}

// The Auth root is the tenant's first authority, and no Go test reached it until
// now — only the shell demonstration did. Its area names the platform's
// namespace rather than an application, which is how one implementation serves
// both roots: the ceiling it computes is the platform catalog because that is
// the namespace it is in.
func TestEstablishAuthRootComputesFromThePlatformCatalog(t *testing.T) {
	tenantArea, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	authArea, err := domain.NewArea("acme", lab.PlatformNamespace)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "auth-root.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), tenantArea, "tenant-genesis", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := lab.Connect(t.Context(), authArea, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeConnection() })
	auth, ok := api.(establisher)
	if !ok {
		t.Fatalf("lab application does not expose establishment: %T", api)
	}

	// Q-114 first: Auth's own catalog is empty, and a ceiling of nothing is not
	// a ceiling.
	if _, _, err := auth.EstablishAuthRoot(t.Context(), authArea, teamFixture, "fibggi2jur5s"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("establishing against an empty platform catalog gave %v, want ErrRejected", err)
	}

	application, err := domain.NewApplication("hrms")
	if err != nil {
		t.Fatal(err)
	}
	catalog, closeCatalog, err := lab.ConnectCatalog(t.Context(), application, path)
	if err != nil {
		t.Fatal(err)
	}
	// The namespace is named separately from the application, because a platform
	// permission belongs to no application.
	if _, err := catalog.RegisterPlatformPermission(t.Context(), lab.PlatformNamespace, domain.FixtureContext{Name: "application-publisher"},
		domain.PermissionDefinition{ID: lab.AssignmentCreate, Active: true, Boundary: domain.PlatformBoundary}); err != nil {
		t.Fatal(err)
	}
	if err := closeCatalog(); err != nil {
		t.Fatal(err)
	}

	root, content, err := auth.EstablishAuthRoot(t.Context(), authArea, teamFixture, "fibggi2jur5s")
	if err != nil {
		t.Fatal(err)
	}
	if !root.TrustedRoot || content.ParentGrantID != "" || content.Permissions != nil || len(content.Scope) != 0 {
		t.Fatalf("auth root = %#v / %#v", root, content)
	}
	// And the ceiling it computes is the platform permission, reached through a
	// child that selects it — which no application root could do, because
	// no application root's namespace holds it.
	child, childContent, err := auth.CreateGrant(t.Context(), authArea, teamFixture, root.ID, domain.GrantContent{
		Permissions: []string{lab.AssignmentCreate}, Scope: map[string]string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(domain.Assignment{
		Version: "1", ID: "fm5b7t4pan0d", GrantID: child.ID, GrantRevision: childContent.Revision,
		Recipient: domain.Recipient{Type: "group", ID: "fibggi2juubk"}, Status: "enabled",
	})
	if err != nil {
		t.Fatal(err)
	}
	diagnostic, err := auth.CheckAssignment(t.Context(), authArea, raw)
	if err != nil {
		t.Fatalf("a child of the Auth root did not resolve: %v", err)
	}
	if !slices.Contains(diagnostic.Route.Permissions, lab.AssignmentCreate) {
		t.Fatalf("route lost the platform permission: %#v", diagnostic.Route)
	}
}

// Q-132, at the one moment it is most visible: a root with nothing beneath it.
//
// Establishment writes the root and its holding assignment together, so even
// then something depends on the grant and the grant cannot go first. Remove the
// holder, and the root is an ordinary leaf — deletable like any other.
//
// That is the rule working rather than a gap. The lab used to refuse both
// outright: disabling a root, which contradicted "an ordinary grant subject to
// status", and deleting one, which had no rule behind it. One dependency
// condition replaced both, and dismantling is bottom-up.
func TestAnEstablishedRootIsDismantledBottomUp(t *testing.T) {
	api, area := openEstablishLab(t, "tenant-genesis")

	root, _, err := api.EstablishRoot(t.Context(), area, teamFixture, "fibggi2jur5s")
	if err != nil {
		t.Fatal(err)
	}
	holder, err := api.ListAssignments(t.Context(), area, teamFixture, domain.AssignmentFilter{GrantID: root.ID})
	if err != nil || holder.Total != 1 {
		t.Fatalf("root assignment = %#v total=%d err=%v", holder.Assignments, holder.Total, err)
	}

	// The grant cannot go first: its own holder names it.
	if err := api.DeleteGrant(t.Context(), area, teamFixture, root.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleting a held root gave %v, want ErrConflict", err)
	}
	// The holder can, because nothing rests on it — nothing else exists yet.
	if err := api.DeleteAssignment(t.Context(), area, teamFixture, holder.Assignments[0].ID); err != nil {
		t.Fatalf("deleting the holder of an otherwise empty area: %v", err)
	}
	// And then so can the root.
	if err := api.DeleteGrant(t.Context(), area, teamFixture, root.ID); err != nil {
		t.Fatalf("deleting an unheld root with no child: %v", err)
	}
	if _, _, err := api.GetGrant(t.Context(), area, teamFixture, root.ID, 1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the root survived its deletion: %v", err)
	}

	// The area is back where it started, which is what makes this a dismantling
	// rather than damage: a root can be established again.
	if _, _, err := api.EstablishRoot(t.Context(), area, teamFixture, "fibggi2jur5s"); err != nil {
		t.Fatalf("a dismantled area could not be established again: %v", err)
	}
}
