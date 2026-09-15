package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

type authorityAPI struct {
	apiSpy
	identity domain.Identity
	opts     domain.ResolveOptions
	calls    int
}

func (a *authorityAPI) ResolveAuthority(_ context.Context, area domain.Area, fixture domain.FixtureContext, identity domain.Identity, opts domain.ResolveOptions) (domain.ResolvedAuthority, error) {
	a.area, a.fixture, a.identity, a.opts, a.calls = area, fixture, identity, opts, a.calls+1
	return domain.ResolvedAuthority{
		Version: "1", TenantID: area.TenantID(), ApplicationID: area.ApplicationID(), HumanID: identity.HumanID,
		ResolvedGrants: []domain.ResolvedGrant{{
			Version: "1", GrantID: "fk3x9r2m5iv8", Revision: 1,
			Permissions: []string{"hrms:payroll:payslip::read"},
			Scope:       map[string]string{"dept": "FIN"},
			Source: &domain.Source{
				AssignmentID: "fm5b7t4p5iv8", TeamID: "fibggi2juubk", Via: "membership",
				Lineage: []domain.LineageStep{{GrantID: "fk3x9r2m0dq3", Revision: 1, AssignmentID: "fm5b7t4p0dq3", TeamID: "fibggi2jur5s", Root: true}},
			},
		}},
	}, nil
}

func runResolve(t *testing.T, api *authorityAPI, args ...string) (int, string, string) {
	t.Helper()
	connector := &connectorSpy{api: api}
	var out, diag bytes.Buffer
	code := Run(t.Context(), append(args, grantArea...), strings.NewReader(""), &out, &diag, connector.connect, nil)
	return code, out.String(), diag.String()
}

// What prints is the contract, not a projection — so it must parse as the
// document an application receives over the wire.
func TestResolvePrintsTheCanonicalDocument(t *testing.T) {
	api := &authorityAPI{}
	code, out, diag := runResolve(t, api, "resolve", "--human", "fi7io4lvjqio")
	if code != 0 || api.calls != 1 {
		t.Fatalf("exit=%d calls=%d stderr=%q", code, api.calls, diag)
	}
	// Caller and subject are one identity today; the verb must not invent a
	// different actor.
	if api.identity.HumanID != "fi7io4lvjqio" || api.identity.Actor.ID != "fi7io4lvjqio" || api.identity.Actor.Type != "user" {
		t.Fatalf("identity = %#v", api.identity)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("output is not the document: %v\n%s", err, out)
	}
	for _, key := range []string{"version", "tenant_id", "application_id", "human_id", "resolved_grants"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("document is missing %q: %s", key, out)
		}
	}
	// The three boundaries are echoed, so a client can tell "nothing here" from
	// "answered about someone else".
	if decoded["human_id"] != "fi7io4lvjqio" || decoded["tenant_id"] != "acme" {
		t.Fatalf("boundaries not echoed: %s", out)
	}
	if !strings.Contains(out, `"lineage"`) || !strings.Contains(out, `"root": true`) {
		t.Fatalf("explanation missing: %s", out)
	}
}

func TestResolveForwardsFilterAndOmission(t *testing.T) {
	api := &authorityAPI{}
	if code, _, diag := runResolve(t, api, "resolve", "--human", "fi7io4lvjqio",
		"--permissions", "hrms:payroll:payslip::read,hrms:payroll:payslip::write", "--no-source"); code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, diag)
	}
	if len(api.opts.Permissions) != 2 || api.opts.Permissions[0] != "hrms:payroll:payslip::read" {
		t.Fatalf("filter = %#v", api.opts.Permissions)
	}
	if !api.opts.OmitSource {
		t.Fatal("--no-source did not reach the read")
	}
}

// --client is the difference between "a human asking about themselves" and "an
// application asking about somebody". It changes the actor on an authorization
// call, so the verb must build exactly the identity it claims to and not quietly
// keep the subject as the actor.
func TestResolveClientMakesTheCallerAServiceCredential(t *testing.T) {
	api := &authorityAPI{}
	if code, _, diag := runResolve(t, api, "resolve", "--human", "fi7io4lvjwu8", "--client", "agent_hrms"); code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, diag)
	}
	if api.identity.Actor.Type != "service_account" || api.identity.Actor.ID != "agent_hrms" {
		t.Fatalf("actor = %#v, want the credential", api.identity.Actor)
	}
	// The subject is untouched: the credential asks, the human is asked about.
	if api.identity.HumanID != "fi7io4lvjwu8" {
		t.Fatalf("subject = %q, want the human", api.identity.HumanID)
	}

	// Without it the caller is the subject, and the two must not blur.
	if code, _, _ := runResolve(t, api, "resolve", "--human", "fi7io4lvjwu8"); code != 0 {
		t.Fatalf("exit=%d", code)
	}
	if api.identity.Actor.Type != "user" || api.identity.Actor.ID != api.identity.HumanID {
		t.Fatalf("actor = %#v, want the subject acting as itself", api.identity.Actor)
	}
}

func TestResolveRefusesUnsupportedForms(t *testing.T) {
	api := &authorityAPI{}
	for _, args := range [][]string{
		{"resolve"},
		{"resolve", "--human", " "},
		{"resolve", "fi7io4lvjqio"},
		{"resolve", "--human", "fi7io4lvjqio", "--team", "fibggi2juubk"},
		{"resolve", "--human", "fi7io4lvjqio", "--client", " "},
	} {
		if code, _, diag := runResolve(t, api, args...); code != 2 {
			t.Fatalf("%v gave exit=%d stderr=%q", args, code, diag)
		}
	}
	if api.calls != 0 {
		t.Fatalf("a refused form still reached the read: %d", api.calls)
	}
	// An adapter without the capability cannot be talked into one.
	connector := &connectorSpy{api: &apiSpy{}}
	var out, diag bytes.Buffer
	if code := Run(t.Context(), append([]string{"resolve", "--human", "fi7io4lvjqio"}, grantArea...),
		strings.NewReader(""), &out, &diag, connector.connect, nil); code != 5 {
		t.Fatalf("exit=%d stderr=%q", code, diag.String())
	}
}
