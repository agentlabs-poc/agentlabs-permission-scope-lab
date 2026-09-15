package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"strings"
	"testing"
)

type assignmentRecordAPI struct {
	apiSpy
	filter   domain.AssignmentFilter
	acted    string
	upgraded string
	calls    int
}

func (a *assignmentRecordAPI) GetAssignment(_ context.Context, area domain.Area, fixture domain.FixtureContext, id string) (domain.Assignment, error) {
	a.area, a.fixture, a.acted, a.calls = area, fixture, id, a.calls+1
	return domain.Assignment{Version: "1", ID: id, GrantID: "fk3x9r2m5iv8", GrantRevision: 2,
		Recipient: domain.Recipient{Type: "group", ID: "fibggi2juubk"}, Status: "enabled"}, nil
}

func (a *assignmentRecordAPI) ListAssignments(_ context.Context, area domain.Area, fixture domain.FixtureContext, filter domain.AssignmentFilter) (domain.AssignmentPage, error) {
	a.area, a.fixture, a.filter, a.calls = area, fixture, filter, a.calls+1
	return domain.AssignmentPage{Assignments: []domain.Assignment{{
		Version: "1", ID: "fm5b7t4p5iv8", GrantID: "fk3x9r2m5iv8", GrantRevision: 1,
		Recipient: domain.Recipient{Type: "group", ID: "fibggi2juubk"}, Status: "enabled"}}, Total: 1}, nil
}

func (a *assignmentRecordAPI) DeleteAssignment(_ context.Context, area domain.Area, fixture domain.FixtureContext, id string) error {
	a.area, a.fixture, a.acted, a.calls = area, fixture, id, a.calls+1
	return nil
}

func (a *assignmentRecordAPI) UpgradeAssignment(_ context.Context, area domain.Area, fixture domain.FixtureContext, id string) (domain.Assignment, error) {
	a.area, a.fixture, a.upgraded, a.calls = area, fixture, id, a.calls+1
	return domain.Assignment{Version: "1", ID: id, GrantID: "fk3x9r2m5iv8", GrantRevision: 3,
		Recipient: domain.Recipient{Type: "group", ID: "fibggi2juubk"}, Status: "enabled"}, nil
}

func runAssignment(t *testing.T, api *assignmentRecordAPI, args ...string) (int, string, string) {
	t.Helper()
	connector := &connectorSpy{api: api}
	var out, diag bytes.Buffer
	code := Run(t.Context(), append(args, grantArea...), strings.NewReader(""), &out, &diag, connector.connect, nil)
	return code, out.String(), diag.String()
}

// The binding prints before the id, because the binding is what identifies the
// record — printing the id first would suggest otherwise.
func TestAssignmentsGetRendersTheBindingBeforeTheHandle(t *testing.T) {
	api := &assignmentRecordAPI{}
	code, out, diag := runAssignment(t, api, "assignments", "get", "fm5b7t4p5iv8")
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, diag)
	}
	if strings.Index(out, "grant  fk3x9r2m5iv8") > strings.Index(out, "id  fm5b7t4p5iv8") {
		t.Fatalf("the id printed before the binding: %q", out)
	}
	for _, want := range []string{"grant  fk3x9r2m5iv8", "recipient  group fibggi2juubk", "adopted revision  2"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout=%q missing %q", out, want)
		}
	}
	if !strings.Contains(diag, "LAB ONLY") {
		t.Fatalf("stderr=%q", diag)
	}
}

func TestAssignmentsListForwardsEitherDirection(t *testing.T) {
	api := &assignmentRecordAPI{}
	if code, out, _ := runAssignment(t, api, "assignments", "list", "--grant", "fk3x9r2m5iv8"); code != 0 || !strings.Contains(out, "total  1") {
		t.Fatalf("exit=%d stdout=%q", code, out)
	}
	if api.filter.GrantID != "fk3x9r2m5iv8" || api.filter.Recipient != nil {
		t.Fatalf("by grant gave filter %#v", api.filter)
	}
	if code, _, _ := runAssignment(t, api, "assignments", "list", "--recipient", "fibggi2juubk", "--recipient-type", "group"); code != 0 {
		t.Fatal("by recipient refused")
	}
	if api.filter.Recipient == nil || api.filter.Recipient.ID != "fibggi2juubk" || api.filter.GrantID != "" {
		t.Fatalf("by recipient gave filter %#v", api.filter)
	}
}

func TestAssignmentsUpgradeReportsTheAdoptedRevision(t *testing.T) {
	api := &assignmentRecordAPI{}
	code, out, _ := runAssignment(t, api, "assignments", "upgrade", "fm5b7t4p5iv8")
	if code != 0 || api.upgraded != "fm5b7t4p5iv8" || !strings.Contains(out, "adopted revision  3") {
		t.Fatalf("exit=%d upgraded=%q stdout=%q", code, api.upgraded, out)
	}
}

// Malformed invocations are refused before the store is opened.
func TestAssignmentsRefusesMalformedInvocationsBeforeConnecting(t *testing.T) {
	cases := map[string][]string{
		"no verb":            {"assignments"},
		"unknown verb":       {"assignments", "explain", "fm5b7t4p5iv8"},
		"get without id":     {"assignments", "get"},
		"delete without id":  {"assignments", "delete"},
		"upgrade without id": {"assignments", "upgrade"},
		"list unfiltered":    {"assignments", "list"},
		"list both filters":  {"assignments", "list", "--grant", "fk3x9r2m5iv8", "--recipient", "x", "--recipient-type", "group"},
		"recipient no type":  {"assignments", "list", "--recipient", "fibggi2juubk"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			api := &assignmentRecordAPI{}
			code, out, _ := runAssignment(t, api, args...)
			if code != 2 || api.calls != 0 || out != "" {
				t.Fatalf("exit=%d calls=%d stdout=%q, want a refusal before connecting", code, api.calls, out)
			}
		})
	}
}

// An unfiltered listing is refused with the reason, not "missing flag".
func TestAssignmentsListExplainsWhyAFilterIsRequired(t *testing.T) {
	api := &assignmentRecordAPI{}
	code, _, diag := runAssignment(t, api, "assignments", "list")
	if code != 2 || !strings.Contains(diag, "unbounded") {
		t.Fatalf("exit=%d stderr=%q", code, diag)
	}
}
