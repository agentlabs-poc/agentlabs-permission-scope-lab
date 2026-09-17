package abv_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"context"
	"errors"
	"path/filepath"
	"testing"
)

// grantWriter is the grant half of the lab application. Same shape as
// teamWriter: the tests run against a real SQLite store seeded with the worked
// fixture, not against a validation function in isolation.
type grantWriter interface {
	CreateGrant(context.Context, domain.Area, domain.FixtureContext, string, domain.GrantContent) (domain.Grant, domain.GrantContent, error)
	DeleteGrant(context.Context, domain.Area, domain.FixtureContext, string) error
	GetGrant(context.Context, domain.Area, domain.FixtureContext, string, int64) (domain.Grant, domain.GrantContent, error)
	ListGrants(context.Context, domain.Area, domain.FixtureContext, domain.GrantFilter) (domain.GrantPage, error)
	ListGrantRevisions(context.Context, domain.Area, domain.FixtureContext, string, int, int) (domain.GrantRevisionPage, error)
	DeleteAssignment(context.Context, domain.Area, domain.FixtureContext, string) error
}

func openGrantLab(t *testing.T) (grantWriter, domain.Area) {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "grants.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := lab.Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeConnection() })
	writer, ok := api.(grantWriter)
	if !ok {
		t.Fatalf("lab application does not expose the grant operations: %T", api)
	}
	return writer, area
}

// The fixture's fk3x9r2m5iv8 holds read and write over dept=FIN, under the root fk3x9r2m0dq3.
const (
	payslipRead  = "hrms:payroll:payslip::read"
	payslipWrite = "hrms:payroll:payslip::write"
)

// A created grant is issued an id, gets its first revision in the same act, and
// reads back. Revision 1 cannot arrive any other way: publication amends a grant
// and requires a predecessor.
func TestCreateGrantIssuesAnIDAndWritesRevisionOne(t *testing.T) {
	api, area := openGrantLab(t)
	grant, content, err := api.CreateGrant(t.Context(), area, teamFixture, "fk3x9r2m5iv8", domain.GrantContent{
		Permissions: []string{payslipRead},
		Scope:       map[string]string{"cert": "C17"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if grant.ID == "" || grant.Status != "enabled" || grant.TrustedRoot {
		t.Fatalf("head=%#v, want an issued id, enabled, and not a root", grant)
	}
	if content.Revision != 1 || content.GrantID != grant.ID || content.ParentGrantID != "fk3x9r2m5iv8" {
		t.Fatalf("content=%#v, want revision 1 under fk3x9r2m5iv8", content)
	}

	// The caller cannot name the grant: whatever id it puts in the content is
	// replaced by the issued one.
	other, _, err := api.CreateGrant(t.Context(), area, teamFixture, "fk3x9r2m5iv8", domain.GrantContent{
		GrantID: "chosen-by-the-caller", Revision: 99,
		Permissions: []string{payslipRead}, Scope: map[string]string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if other.ID == "chosen-by-the-caller" || other.ID == grant.ID {
		t.Fatalf("id %q was accepted from the caller or reissued", other.ID)
	}

	readHead, readContent, err := api.GetGrant(t.Context(), area, teamFixture, grant.ID, 1)
	if err != nil || readHead.ID != grant.ID || readContent.Scope["cert"] != "C17" {
		t.Fatalf("head=%#v content=%#v err=%v", readHead, readContent, err)
	}
	// revision 0 is the head alone — checking whether a grant is enabled should
	// not require naming a revision.
	if _, bare, err := api.GetGrant(t.Context(), area, teamFixture, grant.ID, 0); err != nil || bare.GrantID != "" {
		t.Fatalf("revision 0 returned content %#v err=%v", bare, err)
	}
}

// Creation refuses what it cannot support, and every refusal is a different
// answer rather than one blanket rejection.
func TestCreateGrantRefusesParentlessUnregisteredAndUnknownParent(t *testing.T) {
	api, area := openGrantLab(t)
	good := domain.GrantContent{Permissions: []string{payslipRead}, Scope: map[string]string{}}

	// A parentless create is not a root. Establishing one needs trust evidence,
	// and no grant operation writes that.
	if _, _, err := api.CreateGrant(t.Context(), area, teamFixture, "", good); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a parentless create gave %v, want ErrRejected", err)
	}
	if _, _, err := api.CreateGrant(t.Context(), area, teamFixture, "absent", good); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("an unknown parent gave %v, want ErrRejected", err)
	}
	// The permission has to be in this application's catalog.
	unregistered := domain.GrantContent{Permissions: []string{"hrms:payroll:payslip::export"}, Scope: map[string]string{}}
	if _, _, err := api.CreateGrant(t.Context(), area, teamFixture, "fk3x9r2m5iv8", unregistered); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("an unregistered permission gave %v, want ErrRejected", err)
	}
	// So does the scope key.
	unknownScope := domain.GrantContent{Permissions: []string{payslipRead}, Scope: map[string]string{"branch": "B1"}}
	if _, _, err := api.CreateGrant(t.Context(), area, teamFixture, "fk3x9r2m5iv8", unknownScope); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("an unregistered scope key gave %v, want ErrRejected", err)
	}
	// A child must state its permissions. Naming no source selects nothing,
	// which is not the same as selecting everything the parent has.
	sourceless := domain.GrantContent{Scope: map[string]string{}}
	if _, _, err := api.CreateGrant(t.Context(), area, teamFixture, "fk3x9r2m5iv8", sourceless); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("a sourceless child gave %v, want ErrMalformed", err)
	}
	// Permissions and a role are a mixture Q-118 refuses.
	mixed := domain.GrantContent{Permissions: []string{payslipRead}, RoleID: "fi9jvxobqsxs", RoleRevision: 1, Scope: map[string]string{}}
	if _, _, err := api.CreateGrant(t.Context(), area, teamFixture, "fk3x9r2m5iv8", mixed); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("a mixed source gave %v, want ErrMalformed", err)
	}
	// An absent scope is not an empty one. The handbook forbids the default by
	// name — "omitting scope or supplying null remains invalid; no missing-scope
	// default to {} is permitted" — because a caller that failed to populate it
	// would otherwise be handed the widest child its parent allows. CreateGrant
	// substituted {} here, one line before a validator that already refuses nil,
	// which is how PublishGrantRevision answers the same input.
	absentScope := domain.GrantContent{Permissions: []string{payslipRead}}
	if _, _, err := api.CreateGrant(t.Context(), area, teamFixture, "fk3x9r2m5iv8", absentScope); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("an absent scope gave %v, want ErrMalformed", err)
	}
	// And the empty scope it was being turned into is still perfectly valid when
	// a caller means it, so this is a refusal to guess rather than a new limit.
	if _, _, err := api.CreateGrant(t.Context(), area, teamFixture, "fk3x9r2m5iv8", good); err != nil {
		t.Fatalf("an explicit empty scope was refused: %v", err)
	}
}

// Q-132: a grant with a child grant can be neither disabled nor deleted, and a
// root is not exempt. The refusal is the dependency rule doing its ordinary work
// — it is what refuses any parent — rather than a rule of the root's own.
//
// This replaced two special cases: a refusal to disable a root, which
// contradicted "an ordinary grant subject to status", and a refusal to delete
// one, which had no rule behind it at all.
func TestARootWithAChildIsRefusedLikeAnyOtherParent(t *testing.T) {
	api, area := openGrantLab(t)

	// ErrConflict, not ErrUnsupported: the answer a parent gets, not the answer
	// a root used to get.
	if err := api.DeleteGrant(t.Context(), area, teamFixture, "fk3x9r2m0dq3"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleting a root with a child gave %v, want ErrConflict", err)
	}
	// And the same answer for the child in the middle, which is the point: one
	// rule, no special subject.
	if err := api.DeleteGrant(t.Context(), area, teamFixture, "fk3x9r2m5iv8"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleting an ordinary parent gave %v, want ErrConflict", err)
	}
	grant, _, err := api.GetGrant(t.Context(), area, teamFixture, "fk3x9r2m0dq3", 1)
	if err != nil || !grant.TrustedRoot {
		t.Fatalf("root after the refusal = %#v err=%v", grant, err)
	}
	// The root's own assignment is refused too, by the rule that a deletion must
	// not leave a dependent binding naming what it rests on.
	if err := api.DeleteAssignment(t.Context(), area, teamFixture, "fm5b7t4p0dq3"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleting the root's assignment gave %v, want ErrConflict", err)
	}
}

// Listing answers both questions a reader has: which grants are here, and which
// of them is the root.
func TestListGrantsFiltersByStatusAndRoot(t *testing.T) {
	api, area := openGrantLab(t)
	all, err := api.ListGrants(t.Context(), area, teamFixture, domain.GrantFilter{})
	if err != nil || all.Total != 3 {
		t.Fatalf("unfiltered gave total=%d err=%v, want the fixture's three", all.Total, err)
	}
	// Ordered by id, so paging is stable across calls.
	if len(all.Grants) != 3 || all.Grants[0].ID != "fk3x9r2m0dq3" || all.Grants[2].ID != "fk3x9r2man0d" {
		t.Fatalf("order=%#v, want fk3x9r2m0dq3..fk3x9r2man0d", all.Grants)
	}
	yes, no := true, false
	roots, err := api.ListGrants(t.Context(), area, teamFixture, domain.GrantFilter{Root: &yes})
	if err != nil || roots.Total != 1 || roots.Grants[0].ID != "fk3x9r2m0dq3" {
		t.Fatalf("roots=%#v total=%d err=%v, want only fk3x9r2m0dq3", roots.Grants, roots.Total, err)
	}
	children, err := api.ListGrants(t.Context(), area, teamFixture, domain.GrantFilter{Root: &no})
	if err != nil || children.Total != 2 {
		t.Fatalf("non-roots total=%d err=%v, want 2", children.Total, err)
	}
	// The filter narrows the total too — a total counting everything would walk
	// a filtered pager off the end.
	page, err := api.ListGrants(t.Context(), area, teamFixture, domain.GrantFilter{Offset: 1, Limit: 1})
	if err != nil || len(page.Grants) != 1 || page.Grants[0].ID != "fk3x9r2m5iv8" || page.Total != 3 {
		t.Fatalf("offset 1 limit 1 gave %#v total=%d err=%v", page.Grants, page.Total, err)
	}
	past, err := api.ListGrants(t.Context(), area, teamFixture, domain.GrantFilter{Offset: 99})
	if err != nil || len(past.Grants) != 0 {
		t.Fatalf("offset past the end gave %#v err=%v, want an empty page", past.Grants, err)
	}
	for _, bad := range []domain.GrantFilter{{Offset: -1}, {Limit: -1}, {Limit: 9999}, {Status: "retired"}} {
		if _, err := api.ListGrants(t.Context(), area, teamFixture, bad); !errors.Is(err, domain.ErrMalformed) {
			t.Fatalf("filter %#v gave %v, want ErrMalformed", bad, err)
		}
	}
}

// Revisions come back newest first, because the latest is what a new assignment
// would adopt.
func TestListGrantRevisionsIsNewestFirst(t *testing.T) {
	api, area := openGrantLab(t)
	page, err := api.ListGrantRevisions(t.Context(), area, teamFixture, "fk3x9r2m5iv8", 0, 0)
	if err != nil || page.Total != 1 || page.Revisions[0].Revision != 1 {
		t.Fatalf("fk3x9r2m5iv8 revisions=%#v total=%d err=%v", page.Revisions, page.Total, err)
	}
	if _, err := api.ListGrantRevisions(t.Context(), area, teamFixture, "absent", 0, 0); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("an unknown grant gave %v, want ErrNotFound", err)
	}
}

// Delete is permanent and destroys the grant whole, but refuses while anything
// still depends on it — the rule teams settled.
func TestDeleteGrantRefusesWhileAnythingDependsOnIt(t *testing.T) {
	api, area := openGrantLab(t)
	// fk3x9r2m5iv8 has a child (fk3x9r2man0d) and an assignment. Both are reasons to refuse.
	if err := api.DeleteGrant(t.Context(), area, teamFixture, "fk3x9r2m5iv8"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleting a grant with a child gave %v, want ErrConflict", err)
	}
	if err := api.DeleteGrant(t.Context(), area, teamFixture, "absent"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleting an absent grant gave %v, want ErrNotFound", err)
	}

	// A fresh leaf can go, and takes its revision with it: a head with no
	// content, or content with no head, is never left behind.
	grant, _, err := api.CreateGrant(t.Context(), area, teamFixture, "fk3x9r2m5iv8", domain.GrantContent{
		Permissions: []string{payslipWrite}, Scope: map[string]string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := api.DeleteGrant(t.Context(), area, teamFixture, grant.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := api.GetGrant(t.Context(), area, teamFixture, grant.ID, 0); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the head survived deletion: %v", err)
	}
	if _, err := api.ListGrantRevisions(t.Context(), area, teamFixture, grant.ID, 0, 0); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("revisions survived deletion: %v", err)
	}
}
