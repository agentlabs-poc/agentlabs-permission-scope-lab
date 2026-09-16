package hrms_test

import (
	"agentlabs.local/apps/hrms"
	"agentlabs.local/authmiddleware"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

type clock struct{}

func (clock) Now() time.Time { return time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC) }

// stubAuthority is what an application's own tests need instead of Auth: a fixed
// answer to "what does this human hold". The point of this module is that it can
// be tested without a database, a schema, or the authority domain at all — the
// integration test that drives a real store lives on the Auth side, where the
// store is.
type stubAuthority struct{ routes []authmiddleware.Route }

func (s stubAuthority) Load(context.Context, authmiddleware.AuthorityQuery) (authmiddleware.Authority, error) {
	return authmiddleware.Authority{Routes: s.routes}, nil
}

func finRead(human string) authmiddleware.Route {
	return authmiddleware.Route{
		Area:        authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"},
		HumanID:     human,
		Permissions: []string{"hrms:payroll:payslip::read"},
		GrantIDs:    []string{"fk3x9r2m0dq3", "fk3x9r2m5iv8"},
		Predicates:  []authmiddleware.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"}},
	}
}

// The application is testable on its own, against a stubbed authority. That is
// the claim this module exists to make: an application needs the gate and an
// answer, not the record store.
func TestTheApplicationDecidesWithoutAnAuthorityDatabase(t *testing.T) {
	const maya = "fi7io4lvjqio"
	store := hrms.NewStore(hrms.DefaultRecords())
	evaluator, err := authmiddleware.New(stubAuthority{routes: []authmiddleware.Route{finRead(maya)}}, clock{})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := hrms.NewHandler(store, evaluator, hrms.TrustedIdentity("acme", "hrms", maya))
	if err != nil {
		t.Fatal(err)
	}

	for name, tc := range map[string]struct {
		path string
		want int
	}{
		"inside the boundary":  {"/api/v1/acme/FIN/C17", http.StatusOK},
		"outside the boundary": {"/api/v1/acme/ENG/C18", http.StatusForbidden},
	} {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Code != tc.want {
				t.Fatalf("status = %d, want %d — body %s", recorder.Code, tc.want, recorder.Body.String())
			}
			if tc.want != http.StatusOK {
				var result authmiddleware.Result
				if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
					t.Fatalf("a denial must carry the canonical result block: %v", err)
				}
				if result.Decision != authmiddleware.Deny {
					t.Fatalf("result = %#v", result)
				}
			}
		})
	}
}

// failingAuthority stands in for an Auth service that cannot answer.
type failingAuthority struct{ err error }

func (f failingAuthority) Load(context.Context, authmiddleware.AuthorityQuery) (authmiddleware.Authority, error) {
	return authmiddleware.Authority{}, f.err
}

// What the application answers when authority cannot be established, asserted in
// the application's own module.
//
// Its suite passed while every evaluation failure rendered as a denial: the
// branches were reached only by a test two modules away, so `go test ./...` here
// was a false green on the distinction Q-128 exists to protect — an outage must
// never be reported as "you may not".
func TestAnOutageIsNotADenial(t *testing.T) {
	for name, tc := range map[string]struct {
		err    error
		status int
	}{
		"the canonical evaluation block": {
			&authmiddleware.EvaluationError{
				Version: "1", Code: "AUTH_UNREACHABLE",
				Message: "We could not check your access.", MessageReason: "the authority service did not answer",
			},
			http.StatusServiceUnavailable,
		},
		"an outage wrapped by any source": {
			&authmiddleware.EvaluationError{
				Version: "1", Code: "AUTHORITY_UNAVAILABLE",
				Message: "We could not check your access.", MessageReason: "the authority store could not answer",
				Cause: errors.New("dial tcp: connection refused"),
			},
			http.StatusServiceUnavailable,
		},
	} {
		t.Run(name, func(t *testing.T) {
			evaluator, err := authmiddleware.New(failingAuthority{err: tc.err}, clock{})
			if err != nil {
				t.Fatal(err)
			}
			handler, err := hrms.NewHandler(hrms.NewStore(hrms.DefaultRecords()), evaluator,
				hrms.TrustedIdentity("acme", "hrms", "fi7io4lvjqio"))
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/acme/FIN/C17", nil))
			if recorder.Code == http.StatusForbidden {
				t.Fatalf("an outage was rendered as a denial: %s", recorder.Body.String())
			}
			if recorder.Code != tc.status {
				t.Fatalf("status = %d, want %d — %s", recorder.Code, tc.status, recorder.Body.String())
			}
		})
	}
}

// And a denial is a denial, with the canonical block the client contract
// requires rather than a bare status.
func TestADenialCarriesTheCanonicalBlock(t *testing.T) {
	evaluator, err := authmiddleware.New(stubAuthority{}, clock{})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := hrms.NewHandler(hrms.NewStore(hrms.DefaultRecords()), evaluator,
		hrms.TrustedIdentity("acme", "hrms", "fi7io4lvjqio"))
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/acme/FIN/C17", nil))
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", recorder.Code)
	}
	var result authmiddleware.Result
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil || result.Decision != authmiddleware.Deny {
		t.Fatalf("body = %s (%v)", recorder.Body.String(), err)
	}
	if result.ErrorCode == "" || result.ErrorMessage == "" {
		t.Fatalf("a denial must carry both messages: %#v", result)
	}
}

// movingAuthority answers the gate's question and, as a side effect, moves the
// record the request named. Load runs after the binder has gathered material and
// before the bound effect executes, so this is exactly the window Q-074 is about.
type movingAuthority struct {
	inner stubAuthority
	store *hrms.Store
	moved bool
}

func (m *movingAuthority) Load(ctx context.Context, q authmiddleware.AuthorityQuery) (authmiddleware.Authority, error) {
	if !m.moved {
		m.moved = m.store.Move("acme", "FIN", "ENG", "C17")
	}
	return m.inner.Load(ctx, q)
}

// A record that has left the authorized boundary between the decision and the
// effect must not be served under the decision that was made about it.
//
// > "Step 3 must not update C-17 under the earlier Finance-bound allow. The
// > protected data operation must preserve the evaluated tenant, department, and
// > requested-record binding." — concurrent-enforcement.md:84-87
//
// The effect re-reads under the evaluated boundary rather than closing over what
// the binder found, which is the whole guarantee. Nothing tested it: replacing
// the re-read with the bind-time record left every suite green.
func TestARecordThatLeavesTheBoundaryIsNotServedUnderTheOldAllow(t *testing.T) {
	const maya = "fi7io4lvjqio"
	store := hrms.NewStore(hrms.DefaultRecords())
	source := &movingAuthority{inner: stubAuthority{routes: []authmiddleware.Route{finRead(maya)}}, store: store}
	evaluator, err := authmiddleware.New(source, clock{})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := hrms.NewHandler(store, evaluator, hrms.TrustedIdentity("acme", "hrms", maya))
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/acme/FIN/C17", nil))
	if !source.moved {
		t.Fatal("the record never moved, so this test says nothing")
	}
	if recorder.Code == http.StatusOK {
		t.Fatalf("a record outside the evaluated boundary was served: %s", recorder.Body.String())
	}
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 — it is not there under the boundary that was authorized: %s", recorder.Code, recorder.Body.String())
	}
	if body := recorder.Body.String(); strings.Contains(body, "ENG") || strings.Contains(body, "FIN annual") {
		t.Fatalf("the moved record was disclosed anyway: %s", body)
	}
	// It did move, and it is readable where it now lives — by somebody with
	// authority there. The refusal above is about the boundary, not the record.
	if record, ok := store.Get("acme", "ENG", "C17"); !ok || record.Title != "FIN annual" {
		t.Fatalf("the record is not where the move put it: %#v ok=%v", record, ok)
	}
}

// > "too much intelligence in endpoint about auth. it shold just deny." …
// > "return **deny**, not a successful result narrowed to the rows the caller
// > could access." — collection-enforcement.md:7-11
//
// A self-scoped route is the case where narrowing is most tempting: the caller
// demonstrably may see *some* rows, and an endpoint that filtered to those would
// look helpful and be wrong. Q-071 says the ask itself is refused.
//
// It works because the collection binders supply no `user` material at all, so
// the predicate cannot match — which is a property of the bindings rather than
// of any rule written down in the gate, and nothing tested it. The failure it
// guards against is somebody "improving" bindDepartment to add `user: $self`, or
// filtering the returned rows; neither would have failed a test.
func TestASelfScopedRouteDeniesCollectionsRatherThanNarrowingThem(t *testing.T) {
	const maya = "fi7io4lvjqio"
	// Scoped to the caller and nothing else. finRead also narrows to dept=FIN,
	// and against the all-certificates endpoint that predicate refuses on its own
	// — so with it the self predicate was never what decided, and teaching the
	// all-certificates binder about $self left the suite green.
	selfRead := finRead(maya)
	selfRead.Predicates = []authmiddleware.Predicate{
		{Key: "user", Value: "$self", SourceGrantID: "fk3x9r2m5iv8"},
	}

	store := hrms.NewStore(hrms.DefaultRecords())
	evaluator, err := authmiddleware.New(stubAuthority{routes: []authmiddleware.Route{selfRead}}, clock{})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := hrms.NewHandler(store, evaluator, hrms.TrustedIdentity("acme", "hrms", maya))
	if err != nil {
		t.Fatal(err)
	}

	// The single-record read still works, because that binder does supply the
	// record's employee — so the refusals below are about the ask, not about the
	// route being unusable. It reaches an ENG record too, which is the proof that
	// nothing but the self predicate is doing the refusing below.
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/acme/FIN/C17", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("a self-scoped route could not read the caller's own record: %d %s", recorder.Code, recorder.Body.String())
	}
	elsewhere := httptest.NewRecorder()
	handler.ServeHTTP(elsewhere, httptest.NewRequest(http.MethodGet, "/api/v1/acme/ENG/C18", nil))
	if elsewhere.Code != http.StatusForbidden {
		t.Fatalf("the route is narrowed by something other than the caller: %d", elsewhere.Code)
	}

	for name, path := range map[string]string{
		"a department listing": "/api/v1/acme/departments/FIN/certificates",
		"every certificate":    "/api/v1/acme/certificates",
	} {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
			if recorder.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403 — %s", recorder.Code, recorder.Body.String())
			}
			// And not one row leaked on the way to refusing.
			for _, certificate := range []string{"C17", "C18", "C19"} {
				if strings.Contains(recorder.Body.String(), certificate) {
					t.Fatalf("the refusal carried rows: %s", recorder.Body.String())
				}
			}
		})
	}
}

// The write path records the grants that authorized it. This is the whole point
// of handing the effect its Result: the endpoint can name why it was permitted
// to make the change, and it can only name what the gate gave it.
func TestAWriteRecordsTheGrantsThatAuthorizedIt(t *testing.T) {
	const maya = "fi7io4lvjqio"
	route := finRead(maya)
	route.Permissions = []string{"hrms:payroll:payslip::read", "hrms:payroll:payslip::write"}
	store := hrms.NewStore(hrms.DefaultRecords())
	evaluator, err := authmiddleware.New(stubAuthority{routes: []authmiddleware.Route{route}}, clock{})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := hrms.NewHandler(store, evaluator, hrms.TrustedIdentity("acme", "hrms", maya))
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/v1/acme/certificates/C17",
		strings.NewReader(`{"department_id":"FIN","title":"revised"}`)))
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body)
	}
	writes := store.Writes()
	if len(writes) != 1 {
		t.Fatalf("writes=%#v", writes)
	}
	if writes[0].CertificateID != "C17" || !reflect.DeepEqual(writes[0].GrantIDs, []string{"fk3x9r2m0dq3", "fk3x9r2m5iv8"}) {
		t.Fatalf("evidence=%#v", writes[0])
	}

	// A refused write records nothing: the effect never ran, so there is no
	// change to account for and no evidence to invent.
	refused := httptest.NewRecorder()
	handler.ServeHTTP(refused, httptest.NewRequest(http.MethodPut, "/api/v1/acme/certificates/C18",
		strings.NewReader(`{"department_id":"ENG","title":"stolen"}`)))
	if refused.Code == http.StatusOK {
		t.Fatalf("a write outside the boundary succeeded: %s", refused.Body)
	}
	if len(store.Writes()) != 1 {
		t.Fatalf("a refused write left evidence: %#v", store.Writes())
	}
}
