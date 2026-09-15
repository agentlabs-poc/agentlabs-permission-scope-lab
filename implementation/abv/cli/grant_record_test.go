package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"strings"
	"testing"
)

// grantRecordAPI records what the CLI forwarded, so these assert the parse
// rather than the store.
type grantRecordAPI struct {
	apiSpy
	parent  string
	content domain.GrantContent
	filter  domain.GrantFilter
	deleted string
	calls   int
}

func (a *grantRecordAPI) CreateGrant(_ context.Context, area domain.Area, fixture domain.FixtureContext, parent string, proposed domain.GrantContent) (domain.Grant, domain.GrantContent, error) {
	a.area, a.fixture, a.parent, a.content, a.calls = area, fixture, parent, proposed, a.calls+1
	grant := domain.Grant{ID: "fi8c8111kow0", Status: "enabled"}
	proposed.GrantID, proposed.Revision, proposed.ParentGrantID, proposed.Version = grant.ID, 1, parent, "1"
	return grant, proposed, nil
}

func (a *grantRecordAPI) DeleteGrant(_ context.Context, area domain.Area, fixture domain.FixtureContext, id string) error {
	a.area, a.fixture, a.deleted, a.calls = area, fixture, id, a.calls+1
	return nil
}

func (a *grantRecordAPI) GetGrant(_ context.Context, area domain.Area, fixture domain.FixtureContext, id string, revision int64) (domain.Grant, domain.GrantContent, error) {
	a.area, a.fixture, a.calls = area, fixture, a.calls+1
	grant := domain.Grant{ID: id, Status: "enabled", TrustedRoot: true}
	if revision == 0 {
		return grant, domain.GrantContent{}, nil
	}
	return grant, domain.GrantContent{Version: "1", GrantID: id, Revision: revision, Scope: map[string]string{}}, nil
}

func (a *grantRecordAPI) ListGrants(_ context.Context, area domain.Area, fixture domain.FixtureContext, filter domain.GrantFilter) (domain.GrantPage, error) {
	a.area, a.fixture, a.filter, a.calls = area, fixture, filter, a.calls+1
	return domain.GrantPage{Grants: []domain.Grant{{ID: "fk3x9r2m0dq3", Status: "enabled", TrustedRoot: true}}, Total: 1}, nil
}

func (a *grantRecordAPI) ListGrantRevisions(_ context.Context, area domain.Area, fixture domain.FixtureContext, id string, offset, limit int) (domain.GrantRevisionPage, error) {
	a.area, a.fixture, a.calls = area, fixture, a.calls+1
	return domain.GrantRevisionPage{Revisions: []domain.GrantContent{{GrantID: id, Revision: 2, Permissions: []string{"a", "b"}}}, Total: 1}, nil
}

var grantArea = []string{"--tenant", "acme", "--app", "hrms", "--db", "lab.db", "--fixture-context", "maya-role-publisher"}

func runGrant(t *testing.T, api *grantRecordAPI, args ...string) (int, string, string) {
	t.Helper()
	connector := &connectorSpy{api: api}
	var out, diag bytes.Buffer
	code := Run(t.Context(), append(args, grantArea...), strings.NewReader(""), &out, &diag, connector.connect, nil)
	return code, out.String(), diag.String()
}

func TestGrantsCreateParsesPermissionsAndScope(t *testing.T) {
	api := &grantRecordAPI{}
	code, out, diag := runGrant(t, api, "grants", "create", "--parent", "fk3x9r2m5iv8",
		"--permissions", "hrms:payroll:payslip::read,hrms:payroll:payslip::write", "--scope", "dept=FIN,cert=C17")
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, diag)
	}
	if api.parent != "fk3x9r2m5iv8" || len(api.content.Permissions) != 2 || api.content.Scope["dept"] != "FIN" || api.content.Scope["cert"] != "C17" {
		t.Fatalf("parent=%q content=%#v", api.parent, api.content)
	}
	// Scope renders as an AND, because that is what it means: the constraints
	// accumulate rather than replacing one another.
	if !strings.Contains(out, "scope  cert=C17 AND dept=FIN") || !strings.Contains(out, "kind  child") {
		t.Fatalf("stdout=%q", out)
	}
	if !strings.Contains(diag, "LAB ONLY") {
		t.Fatalf("stderr=%q", diag)
	}
}

// A root prints that its permissions are computed rather than printing an empty
// list, which would read as "none".
func TestGrantsGetRendersAComputedRootDistinctly(t *testing.T) {
	api := &grantRecordAPI{}
	code, out, _ := runGrant(t, api, "grants", "get", "fk3x9r2m0dq3", "--revision", "1")
	if code != 0 {
		t.Fatalf("exit=%d", code)
	}
	for _, want := range []string{"kind  root", "parent  (none — root)", "permissions  (computed from the catalog)", "scope  {} (adds no narrowing)"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout=%q missing %q", out, want)
		}
	}
}

func TestGrantsListAndRevisionsRender(t *testing.T) {
	api := &grantRecordAPI{}
	if code, out, _ := runGrant(t, api, "grants", "list", "--roots"); code != 0 || !strings.Contains(out, "fk3x9r2m0dq3   enabled   root") {
		t.Fatalf("exit=%d stdout=%q", code, out)
	}
	if api.filter.Root == nil || !*api.filter.Root {
		t.Fatalf("--roots gave filter %#v", api.filter)
	}
	if code, _, _ := runGrant(t, api, "grants", "list", "--children"); code != 0 {
		t.Fatal("children listing refused")
	}
	if api.filter.Root == nil || *api.filter.Root {
		t.Fatalf("--children gave filter %#v", api.filter)
	}
	if code, out, _ := runGrant(t, api, "grants", "revisions", "fk3x9r2m5iv8"); code != 0 || !strings.Contains(out, "revision 2    2 permissions") {
		t.Fatalf("exit=%d stdout=%q", code, out)
	}
}

// Malformed invocations are refused before the store is opened, so a bad
// command line never becomes a connection.
func TestGrantsRefusesMalformedInvocationsBeforeConnecting(t *testing.T) {
	cases := map[string][]string{
		"no verb":                 {"grants"},
		"unknown verb":            {"grants", "explain", "fk3x9r2m5iv8"},
		"create without parent":   {"grants", "create", "--permissions", "hrms:a::read"},
		"create with no source":   {"grants", "create", "--parent", "fk3x9r2m5iv8"},
		"create with both":        {"grants", "create", "--parent", "fk3x9r2m5iv8", "--permissions", "hrms:a::read", "--role", "r", "--role-revision", "1"},
		"create bad scope":        {"grants", "create", "--parent", "fk3x9r2m5iv8", "--permissions", "hrms:a::read", "--scope", "dept"},
		"list roots and children": {"grants", "list", "--roots", "--children"},
		"get without id":          {"grants", "get"},
		"delete without id":       {"grants", "delete"},
		"revisions without id":    {"grants", "revisions"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			api := &grantRecordAPI{}
			code, out, _ := runGrant(t, api, args...)
			if code != 2 || api.calls != 0 || out != "" {
				t.Fatalf("exit=%d calls=%d stdout=%q, want a refusal before connecting", code, api.calls, out)
			}
		})
	}
}

// A parentless create is refused by the command line itself, with a message
// that says why rather than "missing flag": establishing a root is not a grant
// operation, so there is no invocation of this verb that could do it.
func TestGrantsCreateExplainsWhyAParentIsRequired(t *testing.T) {
	api := &grantRecordAPI{}
	code, _, diag := runGrant(t, api, "grants", "create", "--permissions", "hrms:a::read")
	if code != 2 || !strings.Contains(diag, "not a root") {
		t.Fatalf("exit=%d stderr=%q", code, diag)
	}
}
