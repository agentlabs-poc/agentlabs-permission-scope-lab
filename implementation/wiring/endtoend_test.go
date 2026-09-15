package wiring_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"agentlabs.local/apps/hrms"
	"agentlabs.local/authclient"
	"agentlabs.local/authmiddleware"
	"agentlabs.local/wiring"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// agents establishes the calling application from its credentials. The lab has
// no issuer, so it answers with the workload credential the fixture gate admits;
// a deployment verifies a real token here.
type agents struct{}

// It verifies the token, because a fixture that admits everyone leaves every
// test below running against an Auth service that authenticates nothing — and
// the credential gate then exists only in its own unit test, never in the loop.
func (agents) Establish(r *http.Request) (domain.Identity, error) {
	if r.Header.Get("Authorization") != "Bearer "+lab.WorkloadToken {
		return domain.Identity{}, errors.New("unrecognised credential")
	}
	return domain.Identity{
		Version: "1",
		Actor:   domain.Actor{Type: "service_account", ID: lab.WorkloadClient},
	}, nil
}

// The whole point, end to end and across a real HTTP boundary: an application
// that holds no authority records reaches a decision about a human it is not.
//
// Everything before this ran in one process against a database handle. This is
// the first time the gate's question leaves the application at all.
func init() { authclient.AllowCleartext() } // httptest speaks http; a deployment must not

func TestAnApplicationDecidesOverTheWire(t *testing.T) {
	const (
		maya  = "fi7io4lvjqio"
		nutan = "fi7io4lvjwu8"
	)
	dir := t.TempDir()
	area, _ := domain.NewArea("acme", "hrms")
	authorityPath := filepath.Join(dir, "authority.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", authorityPath); err != nil {
		t.Fatal(err)
	}
	status, err := lab.NewAssignmentStatusAdministration(area, lab.TeamFINC17(area).Administration)
	if err != nil {
		t.Fatal(err)
	}
	service, err := wiring.Open(t.Context(), wiring.Config{
		AuthorityPath: authorityPath, RegistryPath: filepath.Join(dir, "registry.db"),
		CreateRegistry:         true,
		Administration:         &lab.RoleAdministration{AssignmentStatusAdministration: status},
		RegistryAdministration: regAdmin{}, Operator: operator, Clock: clock{},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()

	// The registry is real here, not a stub that says yes. Auth-AL asks it
	// whether the tenant holds the application before reading a single record,
	// so the application has to exist and be installed — the first thing this
	// test hit was its own AREA_NOT_FOUND, which is the gate working.
	if _, err := service.Applications().RegisterApplication(t.Context(), operator, "hrms", "HRMS"); err != nil {
		t.Fatal(err)
	}
	if err := service.Applications().Install(t.Context(), operator, "acme", "hrms"); err != nil {
		t.Fatal(err)
	}

	handler, err := service.Handler(agents{})
	if err != nil {
		t.Fatal(err)
	}
	auth := httptest.NewServer(handler)
	defer auth.Close()

	// The application side. It is given a URL where it used to be given a
	// database path, and that is the whole difference it sees.
	source, err := authclient.New(auth.URL,
		authclient.Credential{Type: "service_account", ID: lab.WorkloadClient, Bearer: lab.WorkloadToken}, auth.Client())
	if err != nil {
		t.Fatal(err)
	}
	evaluator, err := authmiddleware.New(source, clock{})
	if err != nil {
		t.Fatal(err)
	}

	for name, tc := range map[string]struct {
		human string
		path  string
		want  int
	}{
		"inside her boundary":  {maya, "/api/v1/acme/FIN/C17", http.StatusOK},
		"outside her boundary": {maya, "/api/v1/acme/ENG/C18", http.StatusForbidden},
		// nutan holds nothing until Team2's assignment exists, so the same
		// request is refused for her — a different human, a different answer,
		// through one credential.
		"a second human": {nutan, "/api/v1/acme/FIN/C17", http.StatusForbidden},
	} {
		t.Run(name, func(t *testing.T) {
			app, err := hrms.NewHandler(hrms.NewStore(hrms.DefaultRecords()), evaluator,
				hrms.TrustedIdentity("acme", "hrms", tc.human))
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			app.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != tc.want {
				t.Fatalf("status = %d, want %d — %s", recorder.Code, tc.want, recorder.Body.String())
			}
		})
	}
}

// An unreachable Auth is an evaluation failure and never a denial. Q-128 is
// explicit, and a gate that answered "deny" here would fail closed in the
// reassuring direction while lying about why.
func TestAnUnreachableAuthIsNotADenial(t *testing.T) {
	down := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	down.Close()

	source, err := authclient.New(down.URL,
		authclient.Credential{Type: "service_account", ID: lab.WorkloadClient, Bearer: lab.WorkloadToken}, down.Client())
	if err != nil {
		t.Fatal(err)
	}
	// Asserted at the application's edge, which is where it matters and where an
	// earlier version of this test did not look. Calling Load directly certified
	// a property the assembled system did not have: authclient.Error was not an
	// EvaluationError, so the application's 503 branch went unreached and every
	// Auth outage was reported to the caller as its own bad request.
	evaluator, err := authmiddleware.New(source, clock{})
	if err != nil {
		t.Fatal(err)
	}
	app, err := hrms.NewHandler(hrms.NewStore(hrms.DefaultRecords()), evaluator,
		hrms.TrustedIdentity("acme", "hrms", "fi7io4lvjqio"))
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/acme/FIN/C17", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("an unreachable authority answered %d, want 503 — %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Code == http.StatusForbidden {
		t.Fatal("an outage was rendered as a denial")
	}
}
