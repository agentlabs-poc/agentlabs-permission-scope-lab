package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"strings"
	"testing"
)

type rootAPI struct {
	apiSpy
	verb   string
	holder string
	calls  int
}

func (a *rootAPI) EstablishAuthRoot(_ context.Context, area domain.Area, fixture domain.FixtureContext, holderTeamID string) (domain.Grant, domain.GrantContent, error) {
	a.area, a.fixture, a.holder, a.verb, a.calls = area, fixture, holderTeamID, "auth", a.calls+1
	return a.established()
}

func (a *rootAPI) EstablishRoot(_ context.Context, area domain.Area, fixture domain.FixtureContext, holderTeamID string) (domain.Grant, domain.GrantContent, error) {
	a.area, a.fixture, a.holder, a.verb, a.calls = area, fixture, holderTeamID, "application", a.calls+1
	return a.established()
}

func (a *rootAPI) established() (domain.Grant, domain.GrantContent, error) {
	grant := domain.Grant{ID: "fyb94x7tldds", Status: "enabled", TrustedRoot: true}
	return grant, domain.GrantContent{Version: "1", GrantID: grant.ID, Revision: 1, Scope: map[string]string{}}, nil
}

func runRoot(t *testing.T, api *rootAPI, args ...string) (int, string, string) {
	t.Helper()
	connector := &connectorSpy{api: api}
	var out, diag bytes.Buffer
	code := Run(t.Context(), append(args, grantArea...), strings.NewReader(""), &out, &diag, connector.connect, nil)
	return code, out.String(), diag.String()
}

// Each verb reaches its own method. They are separate all the way down because
// their actors differ: Auth platform administration for the Auth root, the
// tenant administrator for an application's.
func TestRootVerbsReachTheirOwnOperation(t *testing.T) {
	api := &rootAPI{}
	code, out, diag := runRoot(t, api, "root", "establish", "--team", "fibggi2jur5s")
	if code != 0 || api.verb != "application" || api.holder != "fibggi2jur5s" {
		t.Fatalf("exit=%d verb=%q holder=%q stderr=%q", code, api.verb, api.holder, diag)
	}
	// The rendering has to say it is a root and that its coverage is computed,
	// or a reader takes the empty permission line for an empty ceiling.
	if !strings.Contains(out, "kind  root") || !strings.Contains(out, "computed from the catalog") {
		t.Fatalf("stdout=%q", out)
	}
	if code, _, _ := runRoot(t, api, "root", "establish-auth", "--team", "fibggi2jur5s"); code != 0 || api.verb != "auth" {
		t.Fatalf("exit=%d verb=%q", code, api.verb)
	}
}

// Establishment is not in the grant vocabulary, and the command layout says so
// as plainly as the interfaces do (Q-113).
func TestRootRefusesUnsupportedFormsAndUnknownVerbs(t *testing.T) {
	api := &rootAPI{}
	for _, args := range [][]string{
		{"root"},
		{"root", "establish"},
		{"root", "establish", "--team", " "},
		{"root", "rotate", "--team", "fibggi2jur5s"},
		{"root", "establish", "fyb94x7tldds", "--team", "fibggi2jur5s"},
		{"root", "establish", "--team", "fibggi2jur5s", "--parent", "fyb94x7tldds"},
		{"grants", "establish", "--team", "fibggi2jur5s"},
	} {
		code, _, diag := runRoot(t, api, args...)
		if code != 2 {
			t.Fatalf("%v gave exit=%d stderr=%q", args, code, diag)
		}
	}
	if api.calls != 0 {
		t.Fatalf("a refused form still reached the operation: %d calls", api.calls)
	}
}

// An adapter that implements every grant operation and not RootAPI cannot start
// a lineage — the guarantee is reachability, not a check inside a grant verb.
func TestRootIsUnsupportedWithoutTheCapability(t *testing.T) {
	connector := &connectorSpy{api: &apiSpy{}}
	var out, diag bytes.Buffer
	code := Run(t.Context(), append([]string{"root", "establish", "--team", "fibggi2jur5s"}, grantArea...),
		strings.NewReader(""), &out, &diag, connector.connect, nil)
	if code != 5 || !strings.Contains(diag.String(), "unsupported") {
		t.Fatalf("exit=%d stderr=%q", code, diag.String())
	}
}
