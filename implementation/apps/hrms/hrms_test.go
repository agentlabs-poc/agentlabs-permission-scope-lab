package hrms_test

import (
	"agentlabs.local/apps/hrms"
	"agentlabs.local/authmiddleware"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
		Area:       authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"},
		HumanID:    human,
		Permission: "hrms:payroll:payslip::read",
		GrantIDs:   []string{"fk3x9r2m0dq3", "fk3x9r2m5iv8"},
		Predicates: []authmiddleware.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"}},
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
