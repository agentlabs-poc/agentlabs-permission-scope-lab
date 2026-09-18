package wiring_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"agentlabs.local/apps/hrms"
	"agentlabs.local/authclient"
	"agentlabs.local/authmiddleware"
	"agentlabs.local/wiring"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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
	server, _, _ := authServiceWithStore(t)
	return server
}

// authServiceWithStore also hands back the service behind it, for the tests that
// change authority while an application is running against it.
func authServiceWithStore(t *testing.T) (*httptest.Server, *wiring.Service, domain.Area) {
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
		PlatformNamespace: lab.PlatformNamespace,
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
	return server, service, area
}

// application is the client side: the gate, an HTTP authority source, and a
// handler. It links no authority records, which is the whole point.
func application(t *testing.T, authURL string, doer authclient.Doer, human string) http.Handler {
	return applicationHolding(t, authURL, doer, human, hrms.DefaultRecords())
}

func applicationHolding(t *testing.T, authURL string, doer authclient.Doer, human string, records []hrms.Record) http.Handler {
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
	handler, err := hrms.NewHandler(hrms.NewStore(records), evaluator,
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

// The control for every case below. `answer` is a fixture, and eleven
// assertions of the form "this must not open the gate" say nothing at all if
// the fixture could not open it anyway. A review made exactly that point by
// misspelling a field in it: all eleven still passed.
func TestTheFixtureAnswerOpensTheGate(t *testing.T) {
	willing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(answer(maya, "hrms:payroll:payslip::read")))
	}))
	defer willing.Close()
	status, body := call(t, application(t, willing.URL, willing.Client(), maya), http.MethodGet, "/api/v1/acme/FIN/C17", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 — the fixture cannot open the gate, so every refusal below proves nothing: %s", status, body)
	}
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
		"a field neither side agreed on": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read), `"scope":{}`, `"scope":{},"surprise":true`, 1)))
		},
		// These are the answers authclient accepts and the gate then refuses.
		// They used to reach the caller as 400 "your request failed", which is
		// a lie told to the one party who did nothing wrong — and a review
		// found every case above happened to be caught a step earlier, so the
		// invariant this test states was never actually tested here.
		"a scope value that is empty": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read), `"scope":{}`, `"scope":{"dept":""}`, 1)))
		},
		"a wildcard in a scope value": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read), `"scope":{}`, `"scope":{"dept":"*"}`, 1)))
		},
		"a token that is not $self": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read), `"scope":{}`, `"scope":{"user":"$other"}`, 1)))
		},
		"a wildcard in a scope key": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read), `"scope":{}`, `"scope":{"de*pt":"FIN"}`, 1)))
		},
		// The answer direction used to be strict about unknown fields and lax
		// about saying two things at once, which made the strictness
		// decorative. The service refuses both of these in a question.
		"a subject named twice": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read),
				`"human_id":"`+maya+`"`, `"human_id":"`+nutan+`","human_id":"`+maya+`"`, 1)))
		},
		"a grant's permissions named twice": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read),
				`"permissions":["`+read+`"]`, `"permissions":["x"],"permissions":["`+read+`"]`, 1)))
		},
		"a second document after the answer": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(answer(maya, read) + `{"version":"1","tenant_id":"acme"}`))
		},
		// Each grant states its own contract version, and it decides how that
		// grant's scope and validity are to be read.
		"a grant from another contract version": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read),
				`"grant_id":"fk3x9r2m5iv8"`, `"grant_id":"fk3x9r2m5iv8","version":"9"`, 1)))
		},
		"a grant that states no version": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read),
				`{"version":"1","grant_id":"fk3x9r2m5iv8"`, `{"grant_id":"fk3x9r2m5iv8"`, 1)))
		},
		"a grant with no id": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(strings.Replace(answer(maya, read), `"grant_id":"fk3x9r2m5iv8"`, `"grant_id":""`, 1)))
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

// Authority is live, not a snapshot the application took at startup. Withdrawing
// a grant on the Auth side changes the very next answer the application gives —
// which is the whole reason nothing caches yet, and the property an epoch would
// have to preserve if anything ever did.
func TestWithdrawingAnAssignmentChangesTheNextAnswer(t *testing.T) {
	auth, service, area := authServiceWithStore(t)
	app := application(t, auth.URL, auth.Client(), nutan)
	if status, body := call(t, app, http.MethodGet, "/api/v1/acme/FIN/C17", ""); status != http.StatusOK {
		t.Fatalf("status = %d before withdrawal, want 200 — %s", status, body)
	}
	// The assignment is what binds Team2 to the grant. Disabling it takes
	// nutan's route away at the source, and nothing tells the application.
	if _, err := service.Authority().SetAssignmentStatus(t.Context(), area, lab.TeamFINC17(area).Issuer,
		lab.TeamFINC17(area).Proposed.ID, "disabled"); err != nil {
		t.Fatal(err)
	}
	status, body := call(t, app, http.MethodGet, "/api/v1/acme/FIN/C17", "")
	if status == http.StatusOK {
		t.Fatalf("a withdrawn assignment still authorized the request — %s", body)
	}
	if status != http.StatusForbidden {
		t.Fatalf("status = %d after withdrawal, want 403 — %s", status, body)
	}
}

// $self resolves per human, across the wire, in a grant held by a group.
//
// SELF-001 settles it and GROUP-004 makes it the preferred practice: one
// self-scoped grant to a team instead of one grant per employee. It reaches the
// application as ordinary material — the record's own employee — and the answer
// is different for each person who asks through the same grant.
func TestSelfResolvesPerHumanOverTheWire(t *testing.T) {
	auth, service, area := authServiceWithStore(t)
	issuer := lab.TeamFINC17(area).Issuer

	// C20 is the discriminating record: same department as C19, same
	// certificate rule, and a different employee. Without it the only record
	// nutan could not reach was in another department, so `dept=FIN` refused it
	// on its own and the `user` predicate was never consulted — the test passed
	// with $self matching anybody, which a review demonstrated.
	records := append(hrms.DefaultRecords(), hrms.Record{
		TenantID: "acme", DepartmentID: "FIN", CertificateID: "C20",
		EmployeeID: "fi7io4lvk35s", OwnerID: "fi7io4lvk35s", Title: "FIN someone else's",
	})
	nutansApp := applicationHolding(t, auth.URL, auth.Client(), nutan, records)

	// Her route is narrowed to cert=C17, so neither FIN record is hers to read
	// yet. That refusal is what the self grant changes — for one of them.
	for _, path := range []string{"/api/v1/acme/FIN/C19", "/api/v1/acme/FIN/C20"} {
		if status, body := call(t, nutansApp, http.MethodGet, path, ""); status != http.StatusForbidden {
			t.Fatalf("%s = %d before the self grant, want 403 — %s", path, status, body)
		}
	}
	grant, content, err := service.Authority().CreateGrant(t.Context(), area, issuer, "fk3x9r2m5iv8",
		domain.GrantContent{Version: "1", Permissions: []string{lab.PayslipRead}, Scope: map[string]string{"user": "$self"}})
	if err != nil {
		t.Fatal(err)
	}
	// To the group, not to the person: the grant says "yourself", and who that
	// is depends on who is asking.
	if _, err := service.Authority().CreateAssignment(t.Context(), area, issuer, domain.Assignment{
		Version: "1", ID: "fm5b7t4pslf1", GrantID: grant.ID, GrantRevision: content.Revision,
		Recipient: domain.Recipient{Type: "group", ID: "fibggi2juxhc"}, Status: "enabled",
	}); err != nil {
		t.Fatal(err)
	}

	// C19's employee is nutan, so the grant that never names her now reaches it.
	if status, body := call(t, nutansApp, http.MethodGet, "/api/v1/acme/FIN/C19", ""); status != http.StatusOK {
		t.Fatalf("status = %d after the self grant, want 200 — %s", status, body)
	}
	// C20's is not. Same department, same permission, same grant: the only
	// thing that differs is whose record it is.
	if status, body := call(t, nutansApp, http.MethodGet, "/api/v1/acme/FIN/C20", ""); status != http.StatusForbidden {
		t.Fatalf("the self grant reached another person's record in the same department: %d — %s", status, body)
	}
	// And it did not widen for the person it does not belong to either: maya
	// holds all of FIN through Team1, so she is the wrong control — the third
	// human holds nothing, and the group grant must not reach her.
	stranger := applicationHolding(t, auth.URL, auth.Client(), "fi7io4lvk35s", records)
	if status, body := call(t, stranger, http.MethodGet, "/api/v1/acme/FIN/C20", ""); status != http.StatusForbidden {
		t.Fatalf("a self grant assigned to Team2 reached a human outside it: %d — %s", status, body)
	}
}

// A human the deployment knows nothing about is answered, not errored: an empty
// authority is a denial, and a denial is a decision. An evaluation error here
// would tell an operator something is broken when nothing is.
func TestAHumanWithNoAuthorityIsDeniedRatherThanFailed(t *testing.T) {
	auth := authService(t)
	status, body := call(t, application(t, auth.URL, auth.Client(), "fi7io4lvk35s"),
		http.MethodGet, "/api/v1/acme/FIN/C17", "")
	if status != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 — a human with nothing is denied, not an outage: %s", status, body)
	}
}

// asked records one question exactly as it left the application.
type asked struct {
	authorization string
	body          map[string]any
}

// inspecting stands in for Auth and keeps the question, then answers it.
func inspecting(t *testing.T, questions *[]asked, permission string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		question := asked{authorization: r.Header.Get("Authorization")}
		if err := json.Unmarshal(raw, &question.body); err != nil {
			t.Error(err)
			return
		}
		*questions = append(*questions, question)
		_, _ = w.Write([]byte(answer(maya, permission)))
	}))
}

// The application asks as *itself* about a human it is not.
//
// This is the architecture's central claim and nothing tested it: a review
// changed the client to ask as the human instead of as the service account, and
// every test in every module still passed. The actor is the application's own
// credential; the subject is whoever the request carries; they differ, and a
// deployment that collapsed them would have every application impersonating
// every user it served.
func TestTheApplicationAsksAsItselfAboutAHuman(t *testing.T) {
	const read = "hrms:payroll:payslip::read"
	var questions []asked
	auth := inspecting(t, &questions, read)
	defer auth.Close()

	if status, body := call(t, application(t, auth.URL, auth.Client(), maya), http.MethodGet, "/api/v1/acme/FIN/C17", ""); status != http.StatusOK {
		t.Fatalf("status = %d — %s", status, body)
	}
	if len(questions) != 1 {
		t.Fatalf("%d questions asked, want 1", len(questions))
	}
	question := questions[0]

	// The credential authenticates the application, and it is the token, never
	// the id the application is known by.
	if question.authorization != "Bearer "+lab.WorkloadToken {
		t.Fatalf("authorization = %q", question.authorization)
	}
	if strings.Contains(question.authorization, lab.WorkloadClient) {
		t.Fatal("the credential id was sent as the secret")
	}

	identity, ok := question.body["identity"].(map[string]any)
	if !ok {
		t.Fatalf("the question carries no identity block: %#v", question.body)
	}
	actor, ok := identity["actor"].(map[string]any)
	if !ok {
		t.Fatalf("the question names no actor: %#v", identity)
	}
	if actor["type"] != "service_account" || actor["id"] != lab.WorkloadClient {
		t.Fatalf("actor = %#v, want the application's own credential", actor)
	}
	if identity["human_id"] != maya {
		t.Fatalf("human_id = %v, want the subject of the request", identity["human_id"])
	}
	if actor["id"] == identity["human_id"] {
		t.Fatal("the application asked as the human — the actor and the subject collapsed")
	}

	// And the question is about a person, not about a request. Nothing in it
	// names the endpoint, the method, the resource, or asks for a verdict.
	encoded, err := json.Marshal(question.body)
	if err != nil {
		t.Fatal(err)
	}
	for _, leaked := range []string{"GET", "/api/v1", "C17", "FIN", "decision", "allow", "deny", "certificate"} {
		if strings.Contains(string(encoded), leaked) {
			t.Fatalf("the question carries %q — it is about a request, not a person: %s", leaked, encoded)
		}
	}
	// Nor does it narrow the reply to the permission in hand. The question is
	// "what does this person hold here", and the complete answer is the one
	// worth caching against them; a reply filtered to one permission could only
	// ever have served the request that asked for it.
	options, ok := question.body["options"].(map[string]any)
	if !ok {
		t.Fatalf("the question carries no options: %#v", question.body)
	}
	if permissions, narrowed := options["permissions"]; narrowed {
		t.Fatalf("options.permissions = %#v, want no permission filter at all", permissions)
	}
}

// A grant for some other permission is not a misbehaving authority. The answer
// describes the person, so it names every grant they hold here — including the
// ones this request is not about. Those are ordinary, and the gate denies on
// them rather than reporting that Auth answered the wrong question.
//
// What must still hold is that they cannot authorize: the permission is read
// from the grant, and it used to be stamped on from the question while the
// grant's own permissions were decoded and never read, so a grant for reading
// authorized a write.
func TestAGrantForAnotherPermissionDeniesRatherThanFailing(t *testing.T) {
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(answer(maya, "hrms:payroll:payslip::write")))
	}))
	defer other.Close()
	status, body := call(t, application(t, other.URL, other.Client(), maya),
		http.MethodGet, "/api/v1/acme/FIN/C17", "")
	if status != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 — %s", status, body)
	}
	if !strings.Contains(body, "NO_AUTHORIZING_GRANT") {
		t.Fatalf("body = %s", body)
	}
}

// The credential is enforced through the whole stack, not only in its own unit
// test. An application whose token Auth does not recognise cannot decide
// anything — and what it must not do is call that a denial, because the person
// making the request is not the one who is unauthenticated.
func TestAnApplicationAuthDoesNotRecogniseDecidesNothing(t *testing.T) {
	auth := authService(t)
	source, err := authclient.New(auth.URL,
		authclient.Credential{Type: "service_account", ID: lab.WorkloadClient, Bearer: "not-the-issued-token"}, auth.Client())
	if err != nil {
		t.Fatal(err)
	}
	evaluator, err := authmiddleware.New(source, clock{})
	if err != nil {
		t.Fatal(err)
	}
	app, err := hrms.NewHandler(hrms.NewStore(hrms.DefaultRecords()), evaluator, hrms.TrustedIdentity("acme", "hrms", maya))
	if err != nil {
		t.Fatal(err)
	}
	status, body := call(t, app, http.MethodGet, "/api/v1/acme/FIN/C17", "")
	if status == http.StatusOK {
		t.Fatalf("an unrecognised credential still opened the gate — %s", body)
	}
	if status == http.StatusForbidden {
		t.Fatalf("the application's own credential problem was rendered as the person's denial — %s", body)
	}
	if status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503 — %s", status, body)
	}
}

// validity travels on the wire and is honoured. Expiry was unit-tested on a
// hand-built Route and never once sent as JSON, so the client could have
// dropped the block entirely — grants that never expire — with every test green.
func TestAValidityWindowIsHonouredAcrossTheWire(t *testing.T) {
	const read = "hrms:payroll:payslip::read"
	// clock{} is frozen, so these windows are fixed relative to it rather than
	// to the wall clock, and the test cannot rot.
	now := clock{}.Now()
	for name, tc := range map[string]struct {
		validity string
		want     int
	}{
		"a window that has closed":     {`{"expires_at":"` + now.Add(-time.Hour).Format(time.RFC3339Nano) + `"}`, http.StatusForbidden},
		"a window not yet open":        {`{"not_before":"` + now.Add(time.Hour).Format(time.RFC3339Nano) + `"}`, http.StatusForbidden},
		"a window that is open now":    {`{"not_before":"` + now.Add(-time.Hour).Format(time.RFC3339Nano) + `","expires_at":"` + now.Add(time.Hour).Format(time.RFC3339Nano) + `"}`, http.StatusOK},
		"a window closing this moment": {`{"expires_at":"` + now.Format(time.RFC3339Nano) + `"}`, http.StatusForbidden},
	} {
		t.Run(name, func(t *testing.T) {
			timed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(strings.Replace(answer(maya, read), `"scope":{}`, `"scope":{},"validity":`+tc.validity, 1)))
			}))
			defer timed.Close()
			status, body := call(t, application(t, timed.URL, timed.Client(), maya), http.MethodGet, "/api/v1/acme/FIN/C17", "")
			if status != tc.want {
				t.Fatalf("status = %d, want %d — %s", status, tc.want, body)
			}
		})
	}
}

// The policy's method is part of the policy. Go's ServeMux routes HEAD to the
// GET handler, so the gate sees a method its policy does not name — and without
// its own check it would authorize it.
func TestAMethodThePolicyDoesNotNameIsRefused(t *testing.T) {
	auth := authService(t)
	status, body := call(t, application(t, auth.URL, auth.Client(), maya), http.MethodHead, "/api/v1/acme/FIN/C17", "")
	if status == http.StatusOK {
		t.Fatalf("HEAD reached a handler whose policy names GET — %s", body)
	}
}

// The PUT's department comes from the request body, and it is the only material
// in the system that does. A gate that bound a constant instead would authorize
// every write against whichever department the constant named — which a review
// demonstrated, undetected by any test.
func TestTheWriteIsBoundToTheDepartmentTheBodyNames(t *testing.T) {
	auth := authService(t)
	app := application(t, auth.URL, auth.Client(), maya)
	// maya holds write within dept=FIN and nowhere else.
	if status, body := call(t, app, http.MethodPut, "/api/v1/acme/certificates/C17",
		`{"department_id":"FIN","title":"revised"}`); status != http.StatusOK {
		t.Fatalf("a write inside her boundary = %d — %s", status, body)
	}
	if status, body := call(t, app, http.MethodPut, "/api/v1/acme/certificates/C18",
		`{"department_id":"ENG","title":"revised"}`); status != http.StatusForbidden {
		t.Fatalf("a write naming another department = %d, want 403 — %s", status, body)
	}
}

// Membership removal is the handbook's own worked example of a withdrawal, and
// it is a different code path from disabling an assignment: the route is built
// from the teams a human is in, not from the assignment loop. The existing test
// covers the assignment; this covers the one most access actually travels.
func TestRemovingAMembershipChangesTheNextAnswer(t *testing.T) {
	auth, service, area := authServiceWithStore(t)
	app := application(t, auth.URL, auth.Client(), maya)
	if status, body := call(t, app, http.MethodGet, "/api/v1/acme/FIN/C17", ""); status != http.StatusOK {
		t.Fatalf("status = %d before removal, want 200 — %s", status, body)
	}
	// Team1 is how maya reaches the FIN grant. Out of the team, out of the route.
	if err := service.Authority().RemoveMember(t.Context(), area, lab.TeamFINC17(area).Issuer, "fibggi2juubk", maya); err != nil {
		t.Fatal(err)
	}
	status, body := call(t, app, http.MethodGet, "/api/v1/acme/FIN/C17", "")
	if status == http.StatusOK {
		t.Fatalf("a removed member still reached the record — %s", body)
	}
	if status != http.StatusForbidden {
		t.Fatalf("status = %d after removal, want 403 — %s", status, body)
	}
	// nutan is in Team2 and untouched by the removal, so this is a withdrawal of
	// one membership rather than of the grant behind it.
	if status, body := call(t, application(t, auth.URL, auth.Client(), nutan), http.MethodGet, "/api/v1/acme/FIN/C17", ""); status != http.StatusOK {
		t.Fatalf("another team's member lost access too: %d — %s", status, body)
	}
}
