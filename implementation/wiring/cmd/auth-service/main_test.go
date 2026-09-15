package main

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	regdomain "agentlabs.local/registry/domain"
	"agentlabs.local/wiring"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// Arguments that do not name a complete configuration are refused, rather than
// half-applied. A flag whose value was silently dropped would start a service
// listening somewhere the operator did not ask for.
func TestIncompleteArgumentsAreRefused(t *testing.T) {
	for name, args := range map[string][]string{
		"nothing at all":   {},
		"an odd number":    {"--authority", "a.db", "--registry"},
		"no tenant":        {"--authority", "a.db", "--registry", "r.db"},
		"no application":   {"--authority", "a.db", "--registry", "r.db", "--tenant", "acme"},
		"a stray word":     {"authority", "a.db", "--registry", "r.db"},
		"an empty tenant":  {"--authority", "a.db", "--registry", "r.db", "--tenant", "", "--app", "hrms"},
		"a malformed area": {"--authority", "a.db", "--registry", "r.db", "--tenant", " ", "--app", "hrms"},
	} {
		t.Run(name, func(t *testing.T) {
			if code := run(args); code != 2 {
				t.Fatalf("exit = %d, want 2 — the service accepted %v", code, args)
			}
		})
	}
}

// The lab installs its own application at startup so a demonstration has an
// area to answer about, and a second run finds it already there. That conflict
// is the only failure this may swallow: a service that starts without its
// installation answers every question "not found", which is a service that
// denies everything rather than an error an operator can see.
func TestInstallationToleratesOnlyAConflict(t *testing.T) {
	dir := t.TempDir()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	authorityPath := filepath.Join(dir, "authority.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", authorityPath); err != nil {
		t.Fatal(err)
	}
	status, err := lab.NewAssignmentStatusAdministration(area, lab.TeamFINC17(area).Administration)
	if err != nil {
		t.Fatal(err)
	}
	service, err := wiring.Open(t.Context(), wiring.Config{
		AuthorityPath: authorityPath, RegistryPath: filepath.Join(dir, "registry.db"), CreateRegistry: true,
		Administration:         &lab.RoleAdministration{AssignmentStatusAdministration: status},
		RegistryAdministration: labRegistryAdmin{}, Operator: operatorIdentity(), Clock: clock{},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()

	if err := install(service, area, operatorIdentity()); err != nil {
		t.Fatalf("the first installation failed: %v", err)
	}
	// The second run of the lab, which is where the conflict comes from.
	if err := install(service, area, operatorIdentity()); err != nil {
		t.Fatalf("a second start refused to reuse its own installation: %v", err)
	}
	// And a failure that is not a conflict is fatal. The registry gate admits
	// only the lab's operator, so asking as anyone else is refused — which is
	// the shape every real installation failure arrives in.
	if err := install(service, area, regdomain.Identity{Version: "1", HumanID: "fi7io4lvk35s"}); err == nil {
		t.Fatal("an installation that was refused outright was treated as success")
	}
}

// The credential gate admits the fixture's token and nothing else. It is one
// comparison, and it is the line a deployment replaces — so the cases it must
// refuse are worth naming.
func TestTheCredentialGateAdmitsOnlyTheIssuedToken(t *testing.T) {
	for name, header := range map[string]string{
		"nothing":             "",
		"the wrong token":     "Bearer wrong",
		"the client id":       "Bearer " + lab.WorkloadClient,
		"no scheme":           lab.WorkloadToken,
		"the wrong scheme":    "Basic " + lab.WorkloadToken,
		"a prefix of it":      "Bearer " + lab.WorkloadToken[:len(lab.WorkloadToken)-1],
		"trailing whitespace": "Bearer " + lab.WorkloadToken + " ",
	} {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/acme/abv/applications/hrms/authority.resolve", nil)
			if header != "" {
				request.Header.Set("Authorization", header)
			}
			if _, err := (labAgents{}).Establish(request); err == nil {
				t.Fatalf("%q was admitted", header)
			}
		})
	}
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Authorization", "Bearer "+lab.WorkloadToken)
	identity, err := (labAgents{}).Establish(request)
	if err != nil {
		t.Fatalf("the issued token was refused: %v", err)
	}
	// The actor is the application's credential, never anything the body claims.
	if identity.Actor.Type != "service_account" || identity.Actor.ID != lab.WorkloadClient {
		t.Fatalf("actor = %#v", identity.Actor)
	}
}
