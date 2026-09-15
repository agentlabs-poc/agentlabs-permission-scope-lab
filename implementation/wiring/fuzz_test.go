package wiring_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"agentlabs.local/wiring"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

var (
	once      sync.Once
	fuzzMux   http.Handler
	fuzzSetup error
)

func fuzzHandler(t *testing.T) http.Handler {
	t.Helper()
	once.Do(func() {
		// Not t.TempDir(): that belongs to whichever subtest happens to run
		// first, and Go removes it when that subtest ends — taking the shared
		// store with it and leaving every later seed to re-seed a half-deleted
		// database. The directory here outlives the fuzz target, as it must.
		dir, err := os.MkdirTemp("", "wiring-fuzz-")
		if err != nil {
			fuzzSetup = err
			return
		}
		area, _ := domain.NewArea("acme", "hrms")
		path := filepath.Join(dir, "authority.db")
		if fuzzSetup = (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", path); fuzzSetup != nil {
			return
		}
		status, err := lab.NewAssignmentStatusAdministration(area, lab.TeamFINC17(area).Administration)
		if err != nil {
			fuzzSetup = err
			return
		}
		service, err := wiring.Open(context.Background(), wiring.Config{
			AuthorityPath: path, RegistryPath: filepath.Join(dir, "registry.db"), CreateRegistry: true,
			Administration:         &lab.RoleAdministration{AssignmentStatusAdministration: status},
			RegistryAdministration: regAdmin{}, Operator: operator, Clock: clock{},
		})
		if err != nil {
			fuzzSetup = err
			return
		}
		if _, err := service.Applications().RegisterApplication(context.Background(), operator, "hrms", "HRMS"); err != nil {
			fuzzSetup = err
			return
		}
		if err := service.Applications().Install(context.Background(), operator, "acme", "hrms"); err != nil {
			fuzzSetup = err
			return
		}
		fuzzMux, fuzzSetup = service.Handler(agents{})
	})
	if fuzzSetup != nil {
		t.Fatal(fuzzSetup)
	}
	return fuzzMux
}

// The handler reads whatever any caller sends. authmiddleware fuzzes its two
// wire contracts because they are wire contracts; this is the third, and the
// only one whose input arrives from outside the process entirely.
//
// The invariant is not "never fail" — it is that a body which does not establish
// a question never produces authority. An answer carrying routes must be an
// answer to the area in the path.
func FuzzResolveNeverAnswersAMalformedQuestion(f *testing.F) {
	for _, seed := range []string{
		`{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":"fi7io4lvjqio"},"options":{}}`,
		`{"version":"1","identity":{"version":"1","human_id":"fi7io4lvjqio"},"options":{"permissions":["hrms:payroll:payslip::read"]}}`,
		`{"version":"2","identity":{"version":"1","human_id":"x"},"options":{}}`,
		`{"version":"1","identity":{"version":"1","human_id":"../../etc/passwd"},"options":{}}`,
		`{"version":"1","identity":{"version":"1","human_id":"*"},"options":{}}`,
		`{}`, `null`, `[]`, `{"version":"1"}`, `{"version":"1","unknown":true}`,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, body []byte) {
		handler := fuzzHandler(t)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost,
			"/api/v1/acme/abv/applications/hrms/authority.resolve", bytes.NewReader(body))
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			// Every refusal must still be the contract's shape, so a client can
			// tell a refusal from a decision rather than guessing.
			var failure map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &failure); err != nil || failure["error_code"] == "" {
				t.Fatalf("status %d carried no error code: %s", recorder.Code, recorder.Body.String())
			}
			return
		}
		var answered domain.ResolvedAuthority
		if err := json.Unmarshal(recorder.Body.Bytes(), &answered); err != nil {
			t.Fatalf("a 200 that is not the contract: %v — %s", err, recorder.Body.String())
		}
		// A successful answer is about the area in the path. The path selects
		// the area; nothing in the body may move it.
		if answered.TenantID != "acme" || answered.ApplicationID != "hrms" {
			t.Fatalf("the body moved the area: %#v", answered)
		}
		if answered.Version != "1" {
			t.Fatalf("unversioned answer: %#v", answered)
		}
	})
}
