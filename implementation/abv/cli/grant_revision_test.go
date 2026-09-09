package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

const revisionJSON = `{"version":"1","grant_id":"G2","revision":2,"parent_grant_id":"G1","permissions":["hrms:payroll:payslip::read","hrms:payroll:payslip::write"],"scope":{"cert":"C17"}}`

type grantRevisionAPI struct {
	apiSpy
	fixture domain.FixtureContext
	source  string
	raw     []byte
	err     error
}

func (a *grantRevisionAPI) PublishGrantRevision(_ context.Context, area domain.Area, fixture domain.FixtureContext, source string, raw []byte) (domain.GrantContent, error) {
	a.area, a.fixture, a.source, a.raw = area, fixture, source, append([]byte(nil), raw...)
	if a.err != nil {
		return domain.GrantContent{}, a.err
	}
	return domain.GrantContent{Version: "1", GrantID: "G2", Revision: 2, ParentGrantID: "G1", Permissions: []string{"hrms:payroll:payslip::read", "hrms:payroll:payslip::write"}, Scope: map[string]string{"cert": "C17"}}, nil
}

type nilGrantRevisionMap map[string]string

func (nilGrantRevisionMap) Inspect(context.Context, domain.Area, string, string) (domain.Record, error) {
	return domain.Record{}, nil
}
func (nilGrantRevisionMap) CheckAssignment(context.Context, domain.Area, []byte) (domain.Diagnostic, error) {
	return domain.Diagnostic{}, nil
}
func (nilGrantRevisionMap) Assign(context.Context, domain.Area, domain.FixtureContext, []byte) (domain.Receipt, error) {
	return domain.Receipt{}, nil
}
func (nilGrantRevisionMap) PublishGrantRevision(context.Context, domain.Area, domain.FixtureContext, string, []byte) (domain.GrantContent, error) {
	panic("typed nil capability called")
}

func TestGrantRevisionPublishParsesAndForwardsCanonicalContent(t *testing.T) {
	api := &grantRevisionAPI{}
	connector := &connectorSpy{api: api}
	args := []string{"grant", "publish", "--file", "-", "--support-assignment", "A1", "--tenant", "acme", "--app", "hrms", "--db", "lab.db", "--fixture-context", "maya-grant-publisher"}
	var out, diag bytes.Buffer
	if got := Run(t.Context(), args, strings.NewReader(revisionJSON), &out, &diag, connector.connect, nil); got != 0 {
		t.Fatalf("exit=%d stderr=%q", got, diag.String())
	}
	if api.source != "A1" || api.fixture.Name != "maya-grant-publisher" || string(api.raw) != revisionJSON || api.area.TenantID() != "acme" || connector.closes != 1 {
		t.Fatalf("api=%+v connector=%+v", api, connector)
	}
	if out.String() != revisionJSON+"\n" || !strings.Contains(diag.String(), "LAB ONLY") {
		t.Fatalf("stdout=%q stderr=%q", out.String(), diag.String())
	}
}

func TestGrantRevisionRejectsMalformedCommandAndContentBeforePublication(t *testing.T) {
	commands := [][]string{
		{"grant", "publish", "--file", "-", "--tenant", "acme", "--app", "hrms", "--db", "x", "--fixture-context", "maya-grant-publisher"},
		{"grant", "publish", "--file", "-", "--support-assignment", "A1", "--tenant", "acme", "--app", "hrms", "--fixture-context", "maya-grant-publisher"},
		{"grant", "publish", "--file", "-", "--support-assignment", "A1", "--app", "hrms", "--db", "x", "--fixture-context", "maya-grant-publisher"},
		{"grant", "publish", "G2", "--file", "-", "--support-assignment", "A1", "--tenant", "acme", "--app", "hrms", "--db", "x", "--fixture-context", "maya-grant-publisher"},
	}
	for _, args := range commands {
		connector := &connectorSpy{api: &grantRevisionAPI{}}
		var out, diag bytes.Buffer
		if got := Run(t.Context(), args, strings.NewReader(revisionJSON), &out, &diag, connector.connect, nil); got != 2 || connector.calls != 0 || out.Len() != 0 {
			t.Fatalf("%q exit=%d calls=%d stdout=%q", args, got, connector.calls, out.String())
		}
	}
}

func TestGrantRevisionRequiresCapabilityAndSuppressesSuccessOnFailure(t *testing.T) {
	args := []string{"grant", "publish", "--file", "-", "--support-assignment", "A1", "--tenant", "acme", "--app", "hrms", "--db", "x", "--fixture-context", "maya-grant-publisher"}
	for _, api := range []interface {
		Inspect(context.Context, domain.Area, string, string) (domain.Record, error)
		CheckAssignment(context.Context, domain.Area, []byte) (domain.Diagnostic, error)
		Assign(context.Context, domain.Area, domain.FixtureContext, []byte) (domain.Receipt, error)
	}{&apiSpy{}, (*grantRevisionAPI)(nil), nilGrantRevisionMap(nil), &grantRevisionAPI{err: domain.ErrRejected}, &grantRevisionAPI{err: context.Canceled}} {
		connector := &connectorSpy{api: api}
		var out, diag bytes.Buffer
		got := Run(t.Context(), args, strings.NewReader(revisionJSON), &out, &diag, connector.connect, nil)
		if (got != 5 && got != 3 && got != 4) || out.Len() != 0 || connector.closes != 1 {
			t.Fatalf("%T exit=%d stdout=%q closes=%d", api, got, out.String(), connector.closes)
		}
	}
	connector := &connectorSpy{api: &grantRevisionAPI{}, closeErr: errors.New("close")}
	var out, diag bytes.Buffer
	if got := Run(t.Context(), args, strings.NewReader(revisionJSON), &out, &diag, connector.connect, nil); got != 4 {
		t.Fatalf("close failure exit=%d", got)
	}
	connector = &connectorSpy{api: &grantRevisionAPI{}}
	if got := Run(t.Context(), args, strings.NewReader(revisionJSON), failingWriter{}, &diag, connector.connect, nil); got != 4 || connector.closes != 1 {
		t.Fatalf("output failure exit=%d closes=%d", got, connector.closes)
	}
}
