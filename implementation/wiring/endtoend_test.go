package wiring_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"agentlabs.local/apps/hrms"
	"agentlabs.local/authclient"
	"agentlabs.local/authmiddleware"
	"agentlabs.local/wiring"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// agents establishes the calling application from its credentials. The lab has
// no issuer, so it answers with the workload credential the fixture gate admits;
// a deployment verifies a real token here.
type agents struct{}

func (agents) Establish(*http.Request) (domain.Identity, error) {
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
		authclient.Credential{Type: "service_account", ID: lab.WorkloadClient}, auth.Client())
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
		authclient.Credential{Type: "service_account", ID: lab.WorkloadClient}, down.Client())
	if err != nil {
		t.Fatal(err)
	}
	_, err = source.Load(t.Context(), authmiddleware.AuthorityQuery{
		Context: authmiddleware.RequestContext{
			Area:     authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"},
			Identity: authmiddleware.Identity{Version: "1", HumanID: "fi7io4lvjqio"},
		},
		Permission: "hrms:payroll:payslip::read",
	})
	if err == nil {
		t.Fatal("a dead authority service answered")
	}
	var failure *authclient.Error
	if !asClientError(err, &failure) || failure.Code != "AUTH_UNREACHABLE" {
		t.Fatalf("err = %v, want an AUTH_UNREACHABLE evaluation failure", err)
	}
}

func asClientError(err error, target **authclient.Error) bool {
	for err != nil {
		if typed, ok := err.(*authclient.Error); ok {
			*target = typed
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
