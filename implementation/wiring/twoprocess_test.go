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
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

const (
	maya  = "fi7io4lvjqio" // Team1 — payslip read and write within dept=FIN
	nutan = "fi7io4lvjwu8" // Team2 — read only, and narrowed again to cert=C17
)

// authService starts the Auth side against a real seeded store, with Team2's
// assignment made so both women hold something. A test that forgot this would
// be reading nutan's refusals as a permission rule when they were really "no
// grant at all" — the same hazard a review caught in demonstration 21.
func authService(t *testing.T) *httptest.Server {
	t.Helper()
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
		AuthorityPath: authorityPath, RegistryPath: filepath.Join(dir, "registry.db"),
		CreateRegistry:         true,
		Administration:         &lab.RoleAdministration{AssignmentStatusAdministration: status},
		RegistryAdministration: regAdmin{}, Operator: operator, Clock: clock{},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = service.Close() })
	if _, err := service.Applications().RegisterApplication(t.Context(), operator, "hrms", "HRMS"); err != nil {
		t.Fatal(err)
	}
	if err := service.Applications().Install(t.Context(), operator, "acme", "hrms"); err != nil {
		t.Fatal(err)
	}
	// Team2's assignment is only proposed in the seeded fixture, so until this
	// runs nutan holds nothing at all — and every refusal of hers would be
	// "no grant" wearing the costume of a permission rule.
	fixture := lab.TeamFINC17(area)
	if _, err := service.Authority().CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed); err != nil {
		t.Fatal(err)
	}
	handler, err := service.Handler(agents{})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

// application is the client side: the gate, an HTTP authority source, and a
// handler. It links no authority records, which is the whole point.
func application(t *testing.T, authURL string, doer authclient.Doer, human string) http.Handler {
	t.Helper()
	source, err := authclient.New(authURL,
		authclient.Credential{Type: "service_account", ID: lab.WorkloadClient, Bearer: lab.WorkloadToken}, doer)
	if err != nil {
		t.Fatal(err)
	}
	evaluator, err := authmiddleware.New(source, clock{})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := hrms.NewHandler(hrms.NewStore(hrms.DefaultRecords()), evaluator,
		hrms.TrustedIdentity("acme", "hrms", human))
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func call(t *testing.T, handler http.Handler, method, path, body string) (int, string) {
	t.Helper()
	var request *http.Request
	if body == "" {
		request = httptest.NewRequest(method, path, nil)
	} else {
		request = httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder.Code, recorder.Body.String()
}

// Every endpoint the application declares, for both women, across a real HTTP
// boundary. Demonstration 21 captures these; this is the same matrix as an
// assertion, so a change that breaks one of them fails here rather than in a
// screenshot nobody reruns.
func TestEveryEndpointForBothHumans(t *testing.T) {
	auth := authService(t)
	const body = `{"department_id":"FIN","title":"revised"}`
	for name, tc := range map[string]struct {
		human, method, path, body string
		want                      int
	}{
		// dept=FIN is inside Team1's narrowing, dept=ENG is outside it.
		"maya reads inside her boundary":   {maya, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusOK},
		"maya reads outside it":            {maya, http.MethodGet, "/api/v1/acme/ENG/C18", "", http.StatusForbidden},
		"maya reads a second FIN record":   {maya, http.MethodGet, "/api/v1/acme/FIN/C19", "", http.StatusOK},
		"maya lists her department":        {maya, http.MethodGet, "/api/v1/acme/departments/FIN/certificates", "", http.StatusOK},
		"maya lists another department":    {maya, http.MethodGet, "/api/v1/acme/departments/ENG/certificates", "", http.StatusForbidden},
		"maya writes":                      {maya, http.MethodPut, "/api/v1/acme/certificates/C17", body, http.StatusOK},
		"nutan reads the certificate":      {nutan, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusOK},
		"nutan reads the other FIN record": {nutan, http.MethodGet, "/api/v1/acme/FIN/C19", "", http.StatusForbidden},
		// The row the whole architecture is about: one endpoint, one policy, two
		// people. Permissions are selected by each child grant; scope is
		// inherited and narrowed.
		"nutan writes what she may read": {nutan, http.MethodPut, "/api/v1/acme/certificates/C17", body, http.StatusForbidden},
		// Q-071: an all-values ask is refused, never narrowed to the subset the
		// asker could have seen.
		"maya asks for every department":  {maya, http.MethodGet, "/api/v1/acme/certificates", "", http.StatusForbidden},
		"nutan asks for every department": {nutan, http.MethodGet, "/api/v1/acme/certificates", "", http.StatusForbidden},
	} {
		t.Run(name, func(t *testing.T) {
			status, answer := call(t, application(t, auth.URL, auth.Client(), tc.human), tc.method, tc.path, tc.body)
			if status != tc.want {
				t.Fatalf("status = %d, want %d — %s", status, tc.want, answer)
			}
		})
	}
}

// A request the application itself cannot make sense of is the caller's error,
// however healthy Auth is. An earlier version of the failure renderer turned
// everything that was not a denial into 503, and an integration test caught it:
// a body with a numeric title is a bad request and saying "try later" about it
// is a lie.
func TestTheApplicationsOwnFaultsAreNotAuthorityFailures(t *testing.T) {
	auth := authService(t)
	app := application(t, auth.URL, auth.Client(), maya)
	for name, tc := range map[string]struct {
		method, path, body string
		want               int
	}{
		"a numeric title":      {http.MethodPut, "/api/v1/acme/certificates/C17", `{"department_id":"FIN","title":7}`, http.StatusBadRequest},
		"not JSON at all":      {http.MethodPut, "/api/v1/acme/certificates/C17", `<html>`, http.StatusBadRequest},
		"an empty body":        {http.MethodPut, "/api/v1/acme/certificates/C17", ` `, http.StatusBadRequest},
		"a missing field":      {http.MethodPut, "/api/v1/acme/certificates/C17", `{"title":"revised"}`, http.StatusBadRequest},
		"a surplus field":      {http.MethodPut, "/api/v1/acme/certificates/C17", `{"department_id":"FIN","title":"t","extra":1}`, http.StatusBadRequest},
		"an oversized title":   {http.MethodPut, "/api/v1/acme/certificates/C17", `{"department_id":"FIN","title":"` + strings.Repeat("x", 500) + `"}`, http.StatusBadRequest},
		"an unknown endpoint":  {http.MethodGet, "/api/v1/acme/nothing/here/at/all", "", http.StatusNotFound},
		"a method not mounted": {http.MethodDelete, "/api/v1/acme/certificates/C17", "", http.StatusMethodNotAllowed},
	} {
		t.Run(name, func(t *testing.T) {
			status, answer := call(t, app, tc.method, tc.path, tc.body)
			if status != tc.want {
				t.Fatalf("status = %d, want %d — %s", status, tc.want, answer)
			}
		})
	}
}

// counting is a Doer that records how many questions actually left the
// application, so a test can assert that a request was refused *before* any
// authority was loaded rather than after.
type counting struct {
	inner authclient.Doer
	asked atomic.Int64
}

func (c *counting) Do(request *http.Request) (*http.Response, error) {
	c.asked.Add(1)
	return c.inner.Do(request)
}

// The tenant in the path is not the tenant the application trusts. Nothing in
// the request may move the area the question is asked about, or the path is an
// authorization parameter.
//
// It is refused as a bad request and not as a denial, which is right: no
// authority was consulted and none could be — there is nothing to decide about
// an area this handler does not serve. The assertion that matters is the second
// one: the refusal happens before the question leaves the process.
func TestThePathCannotChooseTheTenant(t *testing.T) {
	auth := authService(t)
	doer := &counting{inner: auth.Client()}
	status, answer := call(t, application(t, auth.URL, doer, maya), http.MethodGet, "/api/v1/globex/FIN/C17", "")
	if status == http.StatusOK {
		t.Fatalf("another tenant's path was answered 200 — %s", answer)
	}
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 — %s", status, answer)
	}
	if asked := doer.asked.Load(); asked != 0 {
		t.Fatalf("%d questions left the application for an area it does not serve", asked)
	}
}

// One request, one question. A gate that asked twice would double every
// application's load on the service every other application depends on, and a
// gate that cached without saying so would answer from a stale authority.
func TestOneRequestAsksExactlyOneQuestion(t *testing.T) {
	auth := authService(t)
	doer := &counting{inner: auth.Client()}
	app := application(t, auth.URL, doer, maya)
	if status, answer := call(t, app, http.MethodGet, "/api/v1/acme/FIN/C17", ""); status != http.StatusOK {
		t.Fatalf("status = %d — %s", status, answer)
	}
	if asked := doer.asked.Load(); asked != 1 {
		t.Fatalf("one request asked %d questions, want 1", asked)
	}
	// And the next request asks again rather than reusing the first answer:
	// nothing caches, so nothing can be stale, and that is a property worth
	// holding until an epoch exists to invalidate against.
	if status, _ := call(t, app, http.MethodGet, "/api/v1/acme/FIN/C17", ""); status != http.StatusOK {
		t.Fatalf("second request = %d", status)
	}
	if asked := doer.asked.Load(); asked != 2 {
		t.Fatalf("two requests asked %d questions, want 2 — an answer was reused", asked)
	}
}

// The whole stack under concurrent load, with the race detector watching. The
// store is shared, the evaluator is shared, and one authority source serves
// every goroutine.
func TestTheStackIsCorrectUnderConcurrency(t *testing.T) {
	auth := authService(t)
	for _, human := range []string{maya, nutan} {
		app := application(t, auth.URL, auth.Client(), human)
		want := map[string]int{
			"/api/v1/acme/FIN/C17":      http.StatusOK,
			"/api/v1/acme/ENG/C18":      http.StatusForbidden,
			"/api/v1/acme/FIN/C19":      http.StatusOK,
			"/api/v1/acme/certificates": http.StatusForbidden,
		}
		if human == nutan {
			want["/api/v1/acme/FIN/C19"] = http.StatusForbidden
		}
		var group sync.WaitGroup
		for range 8 {
			for path, expected := range want {
				group.Add(1)
				go func() {
					defer group.Done()
					status, answer := call(t, app, http.MethodGet, path, "")
					if status != expected {
						t.Errorf("%s %s = %d, want %d — %s", human, path, status, expected, answer)
					}
				}()
			}
		}
		group.Wait()
	}
}

// answer builds what a well-behaved Auth would return for one human: an
// unconstrained route carrying one permission. Tests bend one field at a time.
func answer(human, permission string) string {
	return `{"version":"1","tenant_id":"acme","application_id":"hrms","human_id":"` + human + `",` +
		`"resolved_grants":[{"version":"1","grant_id":"fk3x9r2m5iv8","revision":1,` +
		`"permissions":["` + permission + `"],"scope":{}}]}`
}

// An Auth that answers something other than the truth must not be able to open
// the gate, and must not be able to close it either: every one of these is an
// evaluation failure, which is 503 — never 200, and never the 403 that would
// tell a person they lack access when the truth is that nobody knows.
func TestAnAuthThatMisbehavesCannotDecideAnything(t *testing.T) {
	const read = "hrms:payroll:payslip::read"
	for name, reply := range map[string]http.HandlerFunc{
		"an internal error": func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) },
		// Auth's own 403 means this application may not ask. It is not the
		// person's denial and must never be rendered as one.
		"a refusal to answer": func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusForbidden) },
		"a demand to log in":  func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusUnauthorized) },
		"not JSON": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`<html>not an authority service</html>`))
		},
		"an empty body": func(w http.ResponseWriter, _ *http.Request) {},
		"another contract version": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read), `"version":"1"`, `"version":"9"`, 1)))
		},
		"an answer about another human": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(answer(nutan, read)))
		},
		"an answer about another area": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read), `"tenant_id":"acme"`, `"tenant_id":"globex"`, 1)))
		},
		// The permission is the one dimension that decides what may be done. It
		// used to be stamped on from the question while the grant's own
		// permissions were decoded and never read, so a grant for reading
		// authorized a write.
		"a grant for a different permission": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(answer(maya, "hrms:payroll:payslip::write")))
		},
		"a field neither side agreed on": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read), `"scope":{}`, `"scope":{},"surprise":true`, 1)))
		},
	} {
		t.Run(name, func(t *testing.T) {
			hostile := httptest.NewServer(reply)
			defer hostile.Close()
			status, body := call(t, application(t, hostile.URL, hostile.Client(), maya),
				http.MethodGet, "/api/v1/acme/FIN/C17", "")
			if status == http.StatusOK {
				t.Fatalf("a misbehaving authority opened the gate — %s", body)
			}
			if status != http.StatusServiceUnavailable {
				t.Fatalf("status = %d, want 503 — %s", status, body)
			}
		})
	}
}

// A redirect is not an authority service. Following one would hand the
// credential to whoever controls the Location header and then believe whatever
// came back — a complete authorization bypass with no code change anywhere.
//
// The attacker here is willing: it answers with an unconstrained accepted route.
// The assertions are that the gate does not open, and that the attacker never
// hears from the application at all.
func TestARedirectNeverReachesTheAttacker(t *testing.T) {
	var reached atomic.Int64
	var credential atomic.Value
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached.Add(1)
		credential.Store(r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(answer(maya, "hrms:payroll:payslip::read")))
	}))
	defer attacker.Close()
	redirecting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, attacker.URL+r.URL.Path, http.StatusTemporaryRedirect)
	}))
	defer redirecting.Close()

	// The application's own client, not httptest's: the redirect policy under
	// test belongs to authclient, and passing a Doer would substitute someone
	// else's.
	status, body := call(t, application(t, redirecting.URL, nil, maya), http.MethodGet, "/api/v1/acme/FIN/C17", "")
	if status == http.StatusOK {
		t.Fatalf("a redirect authorized the request — %s", body)
	}
	if status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503 — %s", status, body)
	}
	if reached.Load() != 0 {
		t.Fatalf("the application called the redirect target %d times, sending %v", reached.Load(), credential.Load())
	}
}
