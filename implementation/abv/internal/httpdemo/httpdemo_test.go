package httpdemo

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/storage"
	storageSQLite "agentlabs.local/abv/internal/storage/sqlite"
	"agentlabs.local/abv/localadapter"
	"agentlabs.local/authmiddleware"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Date(2026, 9, 9, 10, 0, 0, 0, time.UTC) }

func TestSQLiteHTTPDemoConstrainsRecordsAndObservesDisablement(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments[fixture.Proposed.ID] = fixture.Proposed
	dbPath := filepath.Join(t.TempDir(), "authority.db")
	provider, err := storageSQLite.CreateFixture(t.Context(), dbPath, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = provider.Close() })
	source, err := localadapter.Open(t.Context(), dbPath, fixedClock{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = source.Close() })
	evaluator, err := authmiddleware.New(source, fixedClock{})
	if err != nil {
		t.Fatal(err)
	}
	store := NewStore(DefaultRecords())
	handler, err := NewHandler(store, evaluator, TrustedIdentity("acme", "hrms", "maya"))
	if err != nil {
		t.Fatal(err)
	}

	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusOK, `"certificate_id":"C17"`, `"title":"FIN annual"`)
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/FIN/C18", "", http.StatusNotFound, `!"title":"ENG confidential"`)
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/departments/FIN/certificates", "", http.StatusOK, `"certificate_id":"C17"`, `"certificate_id":"C19"`)
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/departments/FIN/certificates", "", http.StatusOK, `!"department_id":"ENG"`, `!"title":"ENG confidential"`)
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/certificates", "", http.StatusForbidden, `"decision":"deny"`)
	nutan, err := NewHandler(store, evaluator, TrustedIdentity("acme", "hrms", "nutan"))
	if err != nil {
		t.Fatal(err)
	}
	assertResponse(t, nutan, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusOK, `"certificate_id":"C17"`)
	assertResponse(t, nutan, http.MethodGet, "/api/v1/acme/departments/FIN/certificates", "", http.StatusForbidden, `"decision":"deny"`, `!"certificate_id":"C19"`)
	outsider, err := NewHandler(store, evaluator, TrustedIdentity("acme", "hrms", "outsider"))
	if err != nil {
		t.Fatal(err)
	}
	assertResponse(t, outsider, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusForbidden, `"decision":"deny"`, `!FIN annual`)

	assertResponse(t, handler, http.MethodPut, "/api/v1/acme/certificates/C17", `{"department_id":"FIN","title":"FIN revised"}`, http.StatusOK, `"title":"FIN revised"`)
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusOK, `"title":"FIN revised"`)
	assertResponse(t, handler, http.MethodPut, "/api/v1/acme/certificates/C18", `{"department_id":"FIN","title":"stolen"}`, http.StatusNotFound, `!ENG confidential`, `!stolen`)
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/ENG/C18", "", http.StatusForbidden, `!ENG confidential`, `!stolen`)
	if got, ok := store.Get("acme", "ENG", "C18"); !ok || got.Title != "ENG confidential" {
		t.Fatalf("ENG record changed: %#v, ok=%v", got, ok)
	}

	before := fixture.Snapshot.Controls["G1"]
	after := before
	after.Status = "disabled"
	if err := provider.Update(t.Context(), area, func(storage.Snapshot) (storage.WriteSet, error) {
		return storage.WriteSet{GrantStatusChange: &storage.GrantStatusChange{Before: before, After: after}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusForbidden, `"decision":"deny"`)
}

func TestHTTPDemoRejectsBoundaryIdentityAndBodyClaims(t *testing.T) {
	store := NewStore(DefaultRecords())
	evaluator := evaluatorFor(t, staticSource{routes: []authmiddleware.Route{{
		Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, HumanID: "maya", Permission: lab.PayslipWrite,
		GrantIDs: []string{"G1"}, Predicates: []authmiddleware.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "G1"}},
	}}})
	handler, err := NewHandler(store, evaluator, TrustedIdentity("acme", "hrms", "maya"))
	if err != nil {
		t.Fatal(err)
	}
	assertResponse(t, handler, http.MethodPut, "/api/v1/acme/certificates/C18", `{"department_id":"FIN","title":"stolen"}`, http.StatusNotFound, `!stolen`)
	assertResponse(t, handler, http.MethodPut, "/api/v1/acme/certificates/C17", `{"department_id":"FIN","title":9}`, http.StatusBadRequest, `!FIN annual`)
	assertResponse(t, handler, http.MethodPut, "/api/v1/acme/certificates/C17?department_id=FIN", `{"title":"query fallback"}`, http.StatusBadRequest, `!query fallback`)
	assertResponse(t, handler, http.MethodPut, "/api/v1/acme/certificates/C17", `{"department_id":"FIN","title":"forged identity","human_id":"maya"}`, http.StatusBadRequest, `!forged identity`)

	wrong, err := NewHandler(store, evaluator, TrustedIdentity("other", "hrms", "maya"))
	if err != nil {
		t.Fatal(err)
	}
	assertResponse(t, wrong, http.MethodPut, "/api/v1/acme/certificates/C17", `{"department_id":"FIN","title":"wrong tenant"}`, http.StatusBadRequest, `!wrong tenant`)
	if got, _ := store.Get("acme", "FIN", "C17"); got.Title != "FIN annual" {
		t.Fatalf("record changed after rejected requests: %#v", got)
	}
}

func TestHTTPDemoSelfAndTimeoutFixturesNeverDiscloseOrExecute(t *testing.T) {
	store := NewStore(DefaultRecords())
	self := staticSource{routes: []authmiddleware.Route{{
		Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, HumanID: "maya", Permission: lab.PayslipRead,
		GrantIDs: []string{"self"}, Predicates: []authmiddleware.Predicate{{Key: "user", Value: "$self", SourceGrantID: "self"}},
	}}}
	handler, err := NewHandler(store, evaluatorFor(t, self), TrustedIdentity("acme", "hrms", "maya"))
	if err != nil {
		t.Fatal(err)
	}
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusOK, `"employee_id":"maya"`)
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/FIN/C19", "", http.StatusForbidden, `!"employee_id":"nutan"`)

	timedOut, err := NewHandler(store, evaluatorFor(t, staticSource{err: context.DeadlineExceeded}), TrustedIdentity("acme", "hrms", "maya"))
	if err != nil {
		t.Fatal(err)
	}
	before, _ := store.Get("acme", "FIN", "C17")
	assertResponse(t, timedOut, http.MethodPut, "/api/v1/acme/certificates/C17", `{"department_id":"FIN","title":"must not run"}`, http.StatusServiceUnavailable, `"error":"request failed"`, `!deadline`)
	after, _ := store.Get("acme", "FIN", "C17")
	if after != before {
		t.Fatalf("timeout executed update: before=%#v after=%#v", before, after)
	}

	evaluationFailure, err := NewHandler(store, evaluatorFor(t, staticSource{err: &authmiddleware.EvaluationError{
		Version: "1", Code: "AUTHORITY_UNAVAILABLE", Message: "Authorization is unavailable.", MessageReason: "Authority could not be loaded.",
	}}), TrustedIdentity("acme", "hrms", "maya"))
	if err != nil {
		t.Fatal(err)
	}
	assertResponse(t, evaluationFailure, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusServiceUnavailable,
		`"error_message":"Authorization is unavailable."`, `"error_message_reason":"Authority could not be loaded."`, `!"decision"`)
}

func TestHTTPDemoAllDepartmentGrantReturnsWholeTenantCollection(t *testing.T) {
	store := NewStore(DefaultRecords())
	handler, err := NewHandler(store, evaluatorFor(t, staticSource{routes: []authmiddleware.Route{{
		Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, HumanID: "maya", Permission: lab.PayslipRead, GrantIDs: []string{"all"},
	}}}), TrustedIdentity("acme", "hrms", "maya"))
	if err != nil {
		t.Fatal(err)
	}
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/certificates", "", http.StatusOK,
		`"certificate_id":"C17"`, `"certificate_id":"C18"`, `"certificate_id":"C19"`)
}

type staticSource struct {
	routes []authmiddleware.Route
	err    error
}

func (s staticSource) Load(context.Context, authmiddleware.AuthorityQuery) (authmiddleware.Authority, error) {
	return authmiddleware.Authority{Routes: s.routes}, s.err
}

func evaluatorFor(t *testing.T, source authmiddleware.AuthoritySource) *authmiddleware.Evaluator {
	t.Helper()
	evaluator, err := authmiddleware.New(source, fixedClock{})
	if err != nil {
		t.Fatal(err)
	}
	return evaluator
}

func assertResponse(t *testing.T, handler http.Handler, method, target, body string, status int, fragments ...string) {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != status {
		t.Fatalf("%s %s: status=%d body=%s", method, target, response.Code, response.Body.String())
	}
	for _, fragment := range fragments {
		absent := strings.HasPrefix(fragment, "!")
		fragment = strings.TrimPrefix(fragment, "!")
		if strings.Contains(response.Body.String(), fragment) == absent {
			t.Fatalf("%s %s: body=%q fragment=%q absent=%v", method, target, response.Body.String(), fragment, absent)
		}
	}
	if strings.Contains(response.Body.String(), "ENG confidential") && status != http.StatusOK {
		t.Fatalf("secret title disclosed: %s", response.Body.String())
	}
	if status == http.StatusForbidden {
		var result authmiddleware.Result
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || result.Decision != authmiddleware.Deny {
			t.Fatalf("invalid deny body: %q, err=%v", response.Body.String(), err)
		}
	}
}
