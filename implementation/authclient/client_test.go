package authclient

import (
	"agentlabs.local/authmiddleware"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func init() { AllowCleartext() }

func ask() authmiddleware.AuthorityQuery {
	return authmiddleware.AuthorityQuery{
		Context: authmiddleware.RequestContext{
			Area:     authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"},
			Identity: authmiddleware.Identity{Version: "1", HumanID: "fi7io4lvjqio"},
		},
	}
}

// answering starts a server returning one body, and a client pointed at it.
func answering(t *testing.T, status int, body string) (*Source, *http.Request) {
	t.Helper()
	var seen *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Clone(r.Context())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	source, err := New(server.URL, Credential{Type: "service_account", ID: "agent_hrms", Bearer: "secret"}, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	return source, seen
}

func codeOf(t *testing.T, err error) string {
	t.Helper()
	var failure *Error
	if !errors.As(err, &failure) {
		t.Fatalf("err = %v, want a client Error", err)
	}
	// Every failure must also be an evaluation error, or the application renders
	// an outage as the caller's own bad request.
	var evaluation *authmiddleware.EvaluationError
	if !errors.As(err, &evaluation) {
		t.Fatalf("err = %v, want an EvaluationError the gate can recognise", err)
	}
	return failure.Code
}

const sound = `{"version":"1","tenant_id":"acme","application_id":"hrms","human_id":"fi7io4lvjqio",` +
	`"resolved_grants":[{"version":"1","grant_id":"g","revision":1,"permissions":["hrms:payroll:payslip::read"],"scope":{"dept":"FIN"}}]}`

// The gate allows on whatever this returns, so everything the answer claims is
// checked against the question. Each of these corroborations could be deleted
// with the suite green until now.
func TestTheAnswerIsCorroboratedAgainstTheQuestion(t *testing.T) {
	for name, tc := range map[string]struct{ body, want string }{
		"another tenant":      {strings.Replace(sound, `"tenant_id":"acme"`, `"tenant_id":"other"`, 1), "WRONG_AREA"},
		"another application": {strings.Replace(sound, `"application_id":"hrms"`, `"application_id":"crm"`, 1), "WRONG_AREA"},
		"another human":       {strings.Replace(sound, `"human_id":"fi7io4lvjqio"`, `"human_id":"fn2q6v8sbo1e"`, 1), "WRONG_SUBJECT"},
		"another version":     {strings.Replace(sound, `"version":"1"`, `"version":"2"`, 1), "UNSUPPORTED_VERSION"},
	} {
		t.Run(name, func(t *testing.T) {
			source, _ := answering(t, http.StatusOK, tc.body)
			got, err := source.Load(t.Context(), ask())
			if code := codeOf(t, err); code != tc.want {
				t.Fatalf("code = %q, want %q", code, tc.want)
			}
			if len(got.Routes) != 0 {
				t.Fatalf("a refused answer still produced %d routes", len(got.Routes))
			}
		})
	}

	source, _ := answering(t, http.StatusOK, sound)
	got, err := source.Load(t.Context(), ask())
	if err != nil || len(got.Routes) != 1 {
		t.Fatalf("a sound answer gave %#v, %v", got.Routes, err)
	}
	if !slices.Equal(got.Routes[0].Permissions, []string{"hrms:payroll:payslip::read"}) || len(got.Routes[0].GrantIDs) == 0 {
		t.Fatalf("route = %#v", got.Routes[0])
	}
}

// The permission is no longer one of the corroborations, because the answer is
// no longer filtered by one: the question names the human and the area, so a
// grant for some other permission is an ordinary part of the answer. What must
// still hold is that the grant's own permissions are carried across rather than
// the query's stamped on — that stamping is how a grant for reading the
// directory once came back approved for reading payroll, and the gate would
// have matched it.
func TestAGrantCarriesItsOwnPermissionsAndNotTheQuestions(t *testing.T) {
	body := strings.Replace(sound, `["hrms:payroll:payslip::read"]`, `["hrms:payroll:payslip::write","hrms:directory:employee::read"]`, 1)
	source, _ := answering(t, http.StatusOK, body)
	got, err := source.Load(t.Context(), ask())
	if err != nil || len(got.Routes) != 1 {
		t.Fatalf("Load() = %#v, %v", got.Routes, err)
	}
	if !slices.Equal(got.Routes[0].Permissions, []string{"hrms:payroll:payslip::write", "hrms:directory:employee::read"}) {
		t.Fatalf("route = %#v", got.Routes[0])
	}
}

// And the question itself carries no permission filter. Asking for one would
// narrow the answer to the request in hand, and an answer that describes one
// request cannot be cached against the human it is about.
func TestTheQuestionDoesNotNarrowTheAnswerToOnePermission(t *testing.T) {
	var body []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sound))
	}))
	t.Cleanup(server.Close)
	source, err := New(server.URL, Credential{Type: "service_account", ID: "agent_hrms", Bearer: "secret"}, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := source.Load(t.Context(), ask()); err != nil {
		t.Fatal(err)
	}
	var sent struct {
		Options struct {
			Permissions []string `json:"permissions"`
		} `json:"options"`
	}
	if err := json.Unmarshal(body, &sent); err != nil {
		t.Fatalf("request body %q: %v", body, err)
	}
	if len(sent.Options.Permissions) != 0 {
		t.Fatalf("the question narrowed the answer to %q", sent.Options.Permissions)
	}
}

// An authorization answer must come from the host that was dialled. Following a
// Location header handed the gate an attacker-authored answer, with the bearer
// credential forwarded to them — demonstrated before it was fixed.
func TestARedirectIsNotAnAuthorityService(t *testing.T) {
	var reached bool
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		_, _ = w.Write([]byte(sound))
	}))
	t.Cleanup(attacker.Close)
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, attacker.URL, http.StatusTemporaryRedirect)
	}))
	t.Cleanup(auth.Close)

	source, err := New(auth.URL, Credential{Type: "service_account", ID: "agent_hrms", Bearer: "secret"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := source.Load(t.Context(), ask())
	if err == nil {
		t.Fatalf("followed a redirect and accepted %d routes", len(got.Routes))
	}
	if reached {
		t.Fatal("the redirect target was contacted, and would have received the credential")
	}
}

func TestEveryFailureIsAnEvaluationFailure(t *testing.T) {
	for name, tc := range map[string]struct {
		status int
		body   string
		want   string
	}{
		"a refusal":        {http.StatusForbidden, `{"error_code":"NOT_ENTITLED_TO_ASK"}`, "AUTH_REFUSED"},
		"an outage":        {http.StatusServiceUnavailable, `{}`, "AUTH_REFUSED"},
		"not the contract": {http.StatusOK, `{"version":"1","surprise":true}`, "AUTH_MALFORMED"},
		"not JSON":         {http.StatusOK, `<html>`, "AUTH_MALFORMED"},
		"oversized":        {http.StatusOK, `{"version":"1","tenant_id":"` + strings.Repeat("x", 1<<20) + `"}`, "AUTH_OVERSIZED"},
	} {
		t.Run(name, func(t *testing.T) {
			source, _ := answering(t, tc.status, tc.body)
			if code := codeOf(t, mustFail(t, source)); code != tc.want {
				t.Fatalf("code = %q, want %q", code, tc.want)
			}
		})
	}
}

func mustFail(t *testing.T, source *Source) error {
	t.Helper()
	got, err := source.Load(t.Context(), ask())
	if err == nil {
		t.Fatalf("expected a failure, got %d routes", len(got.Routes))
	}
	return err
}

// The route is the contract, and a coordinated edit on both sides used to move
// it with the suite green. This is the only thing that pins the path a deployed
// client will dial.
func TestTheRouteIsThePinnedPath(t *testing.T) {
	source, _ := answering(t, http.StatusOK, sound)
	if _, err := source.Load(t.Context(), ask()); err != nil {
		t.Fatal(err)
	}
	const want = "/api/v1/acme/abv/applications/hrms/authority.resolve"
	if got := source.route(ask()); !strings.HasSuffix(got, want) {
		t.Fatalf("route = %q, want a path ending %q", got, want)
	}
	// An area id is not guaranteed to be one safe path segment.
	escaping := ask()
	escaping.Context.Area.TenantID = "acme/../evil"
	if got := source.route(escaping); strings.Contains(got, "/evil/") {
		t.Fatalf("a crafted tenant escaped its segment: %q", got)
	}
}

func TestNewRefusesAnUnusableConfiguration(t *testing.T) {
	sound := Credential{Type: "service_account", ID: "agent_hrms"}
	for name, tc := range map[string]struct {
		url        string
		credential Credential
	}{
		"no url":           {"", sound},
		"not a url":        {"://nonsense", sound},
		"no host":          {"https://", sound},
		"no credential":    {"https://auth.example", Credential{}},
		"no credential id": {"https://auth.example", Credential{Type: "service_account"}},
	} {
		t.Run(name, func(t *testing.T) {
			if source, err := New(tc.url, tc.credential, nil); err == nil || source != nil {
				t.Fatalf("accepted %#v", tc)
			}
		})
	}
}
