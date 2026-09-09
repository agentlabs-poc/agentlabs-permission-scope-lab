package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

type roleAPI struct {
	apiSpy
	role domain.RoleContent
	err  error
}

func (a *roleAPI) PublishRole(_ context.Context, area domain.Area, fixture domain.FixtureContext, role domain.RoleContent) (domain.RoleContent, error) {
	a.area, a.fixture, a.role = area, fixture, role
	if a.err != nil {
		return domain.RoleContent{}, a.err
	}
	return role, nil
}

func TestRolePublishParsesAndForwardsProposal(t *testing.T) {
	api := &roleAPI{}
	connector := &connectorSpy{api: api}
	args := []string{"role", "publish", "payslip-reader", "--revision", "2", "--permissions", "hrms:payroll:payslip::read,hrms:payroll:payslip::write", "--tenant", "acme", "--app", "hrms", "--db", "lab.db", "--fixture-context", "maya-role-publisher"}
	var out, diag bytes.Buffer
	if got := Run(t.Context(), args, strings.NewReader(""), &out, &diag, connector.connect, nil); got != 0 {
		t.Fatalf("exit=%d stderr=%q", got, diag.String())
	}
	if api.role.ID != "payslip-reader" || api.role.Revision != 2 || strings.Join(api.role.Permissions, ",") != "hrms:payroll:payslip::read,hrms:payroll:payslip::write" || api.fixture.Name != "maya-role-publisher" {
		t.Fatalf("proposal=%+v fixture=%+v", api.role, api.fixture)
	}
	want := "internal projection: role\nid  payslip-reader\nrevision  2\npermissions  hrms:payroll:payslip::read,hrms:payroll:payslip::write\n"
	if out.String() != want || !strings.Contains(diag.String(), "LAB ONLY") {
		t.Fatalf("stdout=%q stderr=%q", out.String(), diag.String())
	}
}

func TestRolePublishRejectsMalformedInputBeforeConnecting(t *testing.T) {
	cases := [][]string{
		{"role", "publish", "payslip-reader", "--revision", "0", "--permissions", "read", "--tenant", "acme", "--app", "hrms", "--db", "x", "--fixture-context", "maya-role-publisher"},
		{"role", "publish", "payslip-reader", "--revision", "2x", "--permissions", "read", "--tenant", "acme", "--app", "hrms", "--db", "x", "--fixture-context", "maya-role-publisher"},
		{"role", "publish", "payslip-reader", "--revision", "2", "--permissions", "read,", "--tenant", "acme", "--app", "hrms", "--db", "x", "--fixture-context", "maya-role-publisher"},
		{"role", "publish", "payslip-reader", "extra", "--revision", "2", "--permissions", "read", "--tenant", "acme", "--app", "hrms", "--db", "x", "--fixture-context", "maya-role-publisher"},
		{"role", "publish", "payslip-reader", "--revision", "2", "--permissions", "read", "--permissions", "write", "--tenant", "acme", "--app", "hrms", "--db", "x", "--fixture-context", "maya-role-publisher"},
	}
	for _, args := range cases {
		connector := &connectorSpy{api: &roleAPI{}}
		var out, diag bytes.Buffer
		if got := Run(t.Context(), args, strings.NewReader(""), &out, &diag, connector.connect, nil); got != 2 || connector.calls != 0 || out.Len() != 0 {
			t.Fatalf("%q exit=%d calls=%d stdout=%q", args, got, connector.calls, out.String())
		}
	}
}

func TestRolePublishRequiresCapabilityAndSuppressesSuccessOutputOnFailure(t *testing.T) {
	args := []string{"role", "publish", "payslip-reader", "--revision", "2", "--permissions", "read", "--tenant", "acme", "--app", "hrms", "--db", "x", "--fixture-context", "maya-role-publisher"}
	for _, api := range []any{&apiSpy{}, (*roleAPI)(nil), &roleAPI{err: domain.ErrRejected}, &roleAPI{err: context.Canceled}} {
		connector := &connectorSpy{api: api.(interface {
			Inspect(context.Context, domain.Area, string, string) (domain.Record, error)
			CheckAssignment(context.Context, domain.Area, []byte) (domain.Diagnostic, error)
			Assign(context.Context, domain.Area, domain.FixtureContext, []byte) (domain.Receipt, error)
		})}
		var out, diag bytes.Buffer
		got := Run(t.Context(), args, strings.NewReader(""), &out, &diag, connector.connect, nil)
		if (got != 5 && got != 3 && got != 4) || out.Len() != 0 || connector.closes != 1 {
			t.Fatalf("%T exit=%d stdout=%q closes=%d", api, got, out.String(), connector.closes)
		}
	}
	connector := &connectorSpy{api: &roleAPI{}, closeErr: errors.New("close")}
	var out, diag bytes.Buffer
	if got := Run(t.Context(), args, strings.NewReader(""), &out, &diag, connector.connect, nil); got != 4 {
		t.Fatalf("close failure exit=%d", got)
	}
	connector = &connectorSpy{api: &roleAPI{}}
	diag.Reset()
	if got := Run(t.Context(), args, strings.NewReader(""), failingWriter{}, &diag, connector.connect, nil); got != 4 || connector.closes != 1 {
		t.Fatalf("output failure exit=%d closes=%d", got, connector.closes)
	}
}
