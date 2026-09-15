package appdemo_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	storageSQLite "agentlabs.local/abv/internal/storage/sqlite"
	"agentlabs.local/abv/lab"
	"agentlabs.local/apps/hrms"
	"agentlabs.local/authmiddleware"
	"agentlabs.local/wiring/localsource"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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
	status, err := lab.NewAssignmentStatusAdministration(fixture.Snapshot.Area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	// The agent asks as itself about whichever human the request carries.
	source, err := localsource.Open(t.Context(), dbPath,
		domain.Actor{Type: "service_account", ID: lab.WorkloadClient},
		&lab.RoleAdministration{AssignmentStatusAdministration: status}, &fixedClock{now: time.Now()}, labRegistry{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = source.Close() })
	evaluator, err := authmiddleware.New(source, &fixedClock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	store := hrms.NewStore(hrms.DefaultRecords())
	handler, err := hrms.NewHandler(store, evaluator, hrms.TrustedIdentity("acme", "hrms", "fi7io4lvjqio"))
	if err != nil {
		t.Fatal(err)
	}

	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusOK, `"certificate_id":"C17"`, `"title":"FIN annual"`)
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/FIN/C18", "", http.StatusNotFound, `!"title":"ENG confidential"`)
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/departments/FIN/certificates", "", http.StatusOK, `"certificate_id":"C17"`, `"certificate_id":"C19"`)
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/departments/FIN/certificates", "", http.StatusOK, `!"department_id":"ENG"`, `!"title":"ENG confidential"`)
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/certificates", "", http.StatusForbidden, `"decision":"deny"`)
	nutan, err := hrms.NewHandler(store, evaluator, hrms.TrustedIdentity("acme", "hrms", "fi7io4lvjwu8"))
	if err != nil {
		t.Fatal(err)
	}
	assertResponse(t, nutan, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusOK, `"certificate_id":"C17"`)
	assertResponse(t, nutan, http.MethodGet, "/api/v1/acme/departments/FIN/certificates", "", http.StatusForbidden, `"decision":"deny"`, `!"certificate_id":"C19"`)
	outsider, err := hrms.NewHandler(store, evaluator, hrms.TrustedIdentity("acme", "hrms", "outsider"))
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

	before := fixture.Snapshot.Controls["fk3x9r2m5iv8"]
	after := before
	after.Status = "disabled"
	if err := provider.Update(t.Context(), area, func(storage.Snapshot) (storage.WriteSet, error) {
		return storage.WriteSet{GrantStatusChange: &storage.GrantStatusChange{Before: before, After: after}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusForbidden, `"decision":"deny"`)
}

func TestSQLiteHTTPDemoTracksProtectedDescendantAssignmentAndGrantControls(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	dbPath := filepath.Join(t.TempDir(), "authority.db")
	if err := (lab.Scenarios{}).Seed(t.Context(), area, "team-fin-c17", dbPath); err != nil {
		t.Fatal(err)
	}
	api, closeAPI, err := lab.Connect(t.Context(), area, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeAPI() })
	raw, err := json.Marshal(fixture.Proposed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := api.Assign(t.Context(), area, domain.FixtureContext{Name: "maya-team1"}, raw); err != nil {
		t.Fatal(err)
	}
	assignmentStatus := api.(interface {
		SetAssignmentStatus(context.Context, domain.Area, domain.FixtureContext, string, string) (domain.Assignment, error)
	})
	grantStatus := api.(interface {
		SetGrantStatus(context.Context, domain.Area, domain.FixtureContext, domain.GrantControl) (domain.GrantControl, error)
	})
	fixtureContext := domain.FixtureContext{Name: "maya-team1"}
	status, err := lab.NewAssignmentStatusAdministration(fixture.Snapshot.Area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	// The agent asks as itself about whichever human the request carries.
	source, err := localsource.Open(t.Context(), dbPath,
		domain.Actor{Type: "service_account", ID: lab.WorkloadClient},
		&lab.RoleAdministration{AssignmentStatusAdministration: status}, &fixedClock{now: time.Now()}, labRegistry{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = source.Close() })
	handler, err := hrms.NewHandler(hrms.NewStore(hrms.DefaultRecords()), evaluatorFor(t, source), hrms.TrustedIdentity("acme", "hrms", "fi7io4lvjwu8"))
	if err != nil {
		t.Fatal(err)
	}
	assertAccess := func(status int) {
		t.Helper()
		fragments := []string{`"certificate_id":"C17"`}
		if status != http.StatusOK {
			fragments = []string{`"decision":"deny"`, `!"title":"FIN annual"`}
		}
		assertResponse(t, handler, http.MethodGet, "/api/v1/acme/FIN/C17", "", status, fragments...)
	}

	assertAccess(http.StatusOK)
	if _, err := assignmentStatus.SetAssignmentStatus(t.Context(), area, fixtureContext, "fm5b7t4pan0d", "disabled"); err != nil {
		t.Fatal(err)
	}
	assertAccess(http.StatusForbidden)
	if _, err := assignmentStatus.SetAssignmentStatus(t.Context(), area, fixtureContext, "fm5b7t4pan0d", "enabled"); err != nil {
		t.Fatal(err)
	}
	assertAccess(http.StatusOK)
	if _, err := grantStatus.SetGrantStatus(t.Context(), area, fixtureContext, domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "disabled"}); err != nil {
		t.Fatal(err)
	}
	assertAccess(http.StatusForbidden)
	if _, err := grantStatus.SetGrantStatus(t.Context(), area, fixtureContext, domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "enabled"}); err != nil {
		t.Fatal(err)
	}
	assertAccess(http.StatusOK)
}

func TestHTTPDemoRejectsBoundaryIdentityAndBodyClaims(t *testing.T) {
	store := hrms.NewStore(hrms.DefaultRecords())
	evaluator := evaluatorFor(t, staticSource{routes: []authmiddleware.Route{{
		Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, HumanID: "fi7io4lvjqio", Permission: lab.PayslipWrite,
		GrantIDs: []string{"fk3x9r2m5iv8"}, Predicates: []authmiddleware.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"}},
	}}})
	handler, err := hrms.NewHandler(store, evaluator, hrms.TrustedIdentity("acme", "hrms", "fi7io4lvjqio"))
	if err != nil {
		t.Fatal(err)
	}
	assertResponse(t, handler, http.MethodPut, "/api/v1/acme/certificates/C18", `{"department_id":"FIN","title":"stolen"}`, http.StatusNotFound, `!stolen`)
	assertResponse(t, handler, http.MethodPut, "/api/v1/acme/certificates/C17", `{"department_id":"FIN","title":9}`, http.StatusBadRequest, `!FIN annual`)
	assertResponse(t, handler, http.MethodPut, "/api/v1/acme/certificates/C17?department_id=FIN", `{"title":"query fallback"}`, http.StatusBadRequest, `!query fallback`)
	assertResponse(t, handler, http.MethodPut, "/api/v1/acme/certificates/C17", `{"department_id":"FIN","title":"forged identity","human_id":"fi7io4lvjqio"}`, http.StatusBadRequest, `!forged identity`)

	wrong, err := hrms.NewHandler(store, evaluator, hrms.TrustedIdentity("fi7io4lvkfsw", "hrms", "fi7io4lvjqio"))
	if err != nil {
		t.Fatal(err)
	}
	assertResponse(t, wrong, http.MethodPut, "/api/v1/acme/certificates/C17", `{"department_id":"FIN","title":"wrong tenant"}`, http.StatusBadRequest, `!wrong tenant`)
	if got, _ := store.Get("acme", "FIN", "C17"); got.Title != "FIN annual" {
		t.Fatalf("record changed after rejected requests: %#v", got)
	}
}

func TestHTTPDemoSelfAndTimeoutFixturesNeverDiscloseOrExecute(t *testing.T) {
	store := hrms.NewStore(hrms.DefaultRecords())
	self := staticSource{routes: []authmiddleware.Route{{
		Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, HumanID: "fi7io4lvjqio", Permission: lab.PayslipRead,
		GrantIDs: []string{"self"}, Predicates: []authmiddleware.Predicate{{Key: "user", Value: "$self", SourceGrantID: "self"}},
	}}}
	handler, err := hrms.NewHandler(store, evaluatorFor(t, self), hrms.TrustedIdentity("acme", "hrms", "fi7io4lvjqio"))
	if err != nil {
		t.Fatal(err)
	}
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusOK, `"employee_id":"fi7io4lvjqio"`)
	assertResponse(t, handler, http.MethodGet, "/api/v1/acme/FIN/C19", "", http.StatusForbidden, `!"employee_id":"fi7io4lvjwu8"`)

	timedOut, err := hrms.NewHandler(store, evaluatorFor(t, staticSource{err: context.DeadlineExceeded}), hrms.TrustedIdentity("acme", "hrms", "fi7io4lvjqio"))
	if err != nil {
		t.Fatal(err)
	}
	before, _ := store.Get("acme", "FIN", "C17")
	// The canonical evaluation-error block, the same four fields its sibling
	// branch carries. This asserted {"error":"request failed"} until a review
	// pointed out that it was pinning the one evaluation outcome that carried
	// no error code and neither message — so the two failure paths disagreed
	// with each other on the wire, and this test was what held them apart.
	// `!deadline` stays: the operator's reason may say the question did not
	// finish, and must not leak the Go error's own text.
	assertResponse(t, timedOut, http.MethodPut, "/api/v1/acme/certificates/C17", `{"department_id":"FIN","title":"must not run"}`, http.StatusServiceUnavailable,
		`"version":"1"`, `"error_code":"AUTHORITY_TIMEOUT"`, `"error_message":"We could not check your access."`, `"error_message_reason":`, `!deadline`, `!"decision"`)
	after, _ := store.Get("acme", "FIN", "C17")
	if after != before {
		t.Fatalf("timeout executed update: before=%#v after=%#v", before, after)
	}

	evaluationFailure, err := hrms.NewHandler(store, evaluatorFor(t, staticSource{err: &authmiddleware.EvaluationError{
		Version: "1", Code: "AUTHORITY_UNAVAILABLE", Message: "Authorization is unavailable.", MessageReason: "Authority could not be loaded.",
	}}), hrms.TrustedIdentity("acme", "hrms", "fi7io4lvjqio"))
	if err != nil {
		t.Fatal(err)
	}
	assertResponse(t, evaluationFailure, http.MethodGet, "/api/v1/acme/FIN/C17", "", http.StatusServiceUnavailable,
		`"error_message":"Authorization is unavailable."`, `"error_message_reason":"Authority could not be loaded."`, `!"decision"`)
}

func TestHTTPDemoAllDepartmentGrantReturnsWholeTenantCollection(t *testing.T) {
	store := hrms.NewStore(hrms.DefaultRecords())
	handler, err := hrms.NewHandler(store, evaluatorFor(t, staticSource{routes: []authmiddleware.Route{{
		Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, HumanID: "fi7io4lvjqio", Permission: lab.PayslipRead, GrantIDs: []string{"all"},
	}}}), hrms.TrustedIdentity("acme", "hrms", "fi7io4lvjqio"))
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
	evaluator, err := authmiddleware.New(source, &fixedClock{now: time.Now()})
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
