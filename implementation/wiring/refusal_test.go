package wiring_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"agentlabs.local/wiring"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

// fixedAgents establishes whatever actor a test needs, so the refusals below are
// produced by the service rather than described.
type fixedAgents struct{ actor domain.Actor }

func (f fixedAgents) Establish(*http.Request) (domain.Identity, error) {
	return domain.Identity{Version: "1", Actor: f.actor}, nil
}

// Three arms of the error mapping had never been executed by any test — coverage
// measured zero on each. They are the answers a caller actually receives, and
// the last of them is the distinction this whole repository exists to protect.
func TestEveryRefusalTheServiceCanGiveReachesTheWire(t *testing.T) {
	// A credential bound to some other application. CheckAuthorityRead's
	// service-account arm is the only thing stopping one tenant's application
	// from reading another tenant's humans, and it could have been changed to
	// return nil with the suite green.
	t.Run("a credential not established here", func(t *testing.T) {
		service := fuzzService(t)
		handler, err := service.Handler(fixedAgents{domain.Actor{Type: "service_account", ID: "agent_other"}})
		if err != nil {
			t.Fatal(err)
		}
		// The body names the credential it actually arrived with, so the actor
		// check passes and the entitlement arm is what refuses. A body claiming
		// somebody else is a different refusal, covered in handler_test.
		body := `{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_other"},"human_id":"fi7io4lvjqio"},"options":{}}`
		status, code := post(t, handler, "/api/v1/acme/abv/applications/hrms/authority.resolve", body)
		if status != http.StatusForbidden || code != "NOT_ENTITLED_TO_ASK" {
			t.Fatalf("got %d/%s, want 403/NOT_ENTITLED_TO_ASK", status, code)
		}
		// And indistinguishable from an area that is not this caller's, so the
		// two cannot be told apart by anyone probing.
		entitled := fuzzHandler(t)
		otherStatus, otherCode := post(t, entitled, "/api/v1/acme/abv/applications/crm/authority.resolve", goodBody)
		if otherStatus != status || otherCode != code {
			t.Fatalf("an unestablished credential answered %d/%s where another application answered %d/%s", status, code, otherStatus, otherCode)
		}
	})

	// A human actor may ask about itself and nobody else. Over HTTP the actor
	// comes from the credential and the subject from the body, so this is a
	// different mechanism from the in-process rule and had no test.
	t.Run("a human asking about another human", func(t *testing.T) {
		service := fuzzService(t)
		handler, err := service.Handler(fixedAgents{domain.Actor{Type: "user", ID: "fi7io4lvjqio"}})
		if err != nil {
			t.Fatal(err)
		}
		status, code := post(t, handler, "/api/v1/acme/abv/applications/hrms/authority.resolve",
			`{"version":"1","identity":{"version":"1","actor":{"type":"user","id":"fi7io4lvjqio"},"human_id":"fi7io4lvjwu8"},"options":{}}`)
		if status != http.StatusNotImplemented || code != "UNSUPPORTED" {
			t.Fatalf("got %d/%s, want 501/UNSUPPORTED", status, code)
		}
		// Asking about itself is the control: the refusal is about impersonation,
		// not about being a person.
		if status, _ := post(t, handler, "/api/v1/acme/abv/applications/hrms/authority.resolve",
			`{"version":"1","identity":{"version":"1","actor":{"type":"user","id":"fi7io4lvjqio"},"human_id":"fi7io4lvjqio"},"options":{}}`); status != http.StatusOK {
			t.Fatalf("a human asking about itself got %d", status)
		}
	})

	// The rule the architecture rests on, from the Auth side: an inability to
	// establish authority is an evaluation failure and never a denial. The
	// client side of it is thoroughly tested; this arm had nothing, and one
	// reordering of that switch would start telling every user they lack access
	// whenever the database is unreachable.
	t.Run("the store cannot be read", func(t *testing.T) {
		dir := t.TempDir()
		area, err := domain.NewArea("acme", "hrms")
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "authority.db")
		if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", path); err != nil {
			t.Fatal(err)
		}
		status, err := lab.NewAssignmentStatusAdministration(area, lab.TeamFINC17(area).Administration)
		if err != nil {
			t.Fatal(err)
		}
		service, err := wiring.Open(t.Context(), wiring.Config{
			AuthorityPath: path, RegistryPath: filepath.Join(dir, "registry.db"), CreateRegistry: true,
			Administration:         &lab.RoleAdministration{AssignmentStatusAdministration: status},
			RegistryAdministration: regAdmin{}, Operator: operator, Clock: clock{},
		})
		if err != nil {
			t.Fatal(err)
		}
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
		// It answers while the store is open, so the refusal below is the outage
		// and not the fixture.
		if code, _ := post(t, handler, "/api/v1/acme/abv/applications/hrms/authority.resolve", goodBody); code != http.StatusOK {
			t.Fatalf("the fixture could not answer before the outage: %d", code)
		}
		if err := service.Close(); err != nil {
			t.Fatal(err)
		}
		code, errorCode := post(t, handler, "/api/v1/acme/abv/applications/hrms/authority.resolve", goodBody)
		if code == http.StatusForbidden {
			t.Fatal("an unreadable store was rendered as the person's denial")
		}
		if code != http.StatusServiceUnavailable || errorCode != "AUTHORITY_UNAVAILABLE" {
			t.Fatalf("got %d/%s, want 503/AUTHORITY_UNAVAILABLE", code, errorCode)
		}
	})
}

// The size limit is what stops an authenticated tenant application making the
// service every other application depends on decode arbitrary bytes per request.
// The only oversize test mounted a refusing identity source, so the 401 fired
// first and the limit was never reached.
//
// The first version of this test was the same defect wearing different clothes:
// both bodies carried an unknown field, so DisallowUnknownFields refused each of
// them for its content and the size decided nothing. It passed with the length
// check deleted outright. The body below is legal all the way through — a long
// options.permissions list — so refusing it can only be about its size.
func TestTheSizeLimitAppliesToACallerWhoIsEstablished(t *testing.T) {
	handler := fuzzHandler(t)
	legal := func(permissions int) string {
		// The same registered permission, repeated. Anything else is refused for
		// naming a permission the catalog does not hold, which would make the
		// content decide again.
		list := make([]string, permissions)
		for i := range list {
			list[i] = `"hrms:payroll:payslip::read"`
		}
		return `{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":"fi7io4lvjqio"},"options":{"permissions":[` +
			strings.Join(list, ",") + `]}}`
	}
	// Under the limit, and answered — so the body's shape is not what refuses it.
	small := legal(4)
	if len(small) > wiring.MaxRequestBytes {
		t.Fatal("the control body is not under the limit")
	}
	if status, code := post(t, handler, "/api/v1/acme/abv/applications/hrms/authority.resolve", small); status != http.StatusOK {
		t.Fatalf("a legal small body got %d/%s, so this test cannot attribute the refusal below", status, code)
	}
	// Over it, same shape, refused.
	big := legal(3000)
	if len(big) <= wiring.MaxRequestBytes {
		t.Fatalf("the oversize body is only %d bytes, under the %d limit", len(big), wiring.MaxRequestBytes)
	}
	if status, code := post(t, handler, "/api/v1/acme/abv/applications/hrms/authority.resolve", big); status != http.StatusBadRequest || code != "MALFORMED_REQUEST" {
		t.Fatalf("an oversize but otherwise legal body got %d/%s, want 400/MALFORMED_REQUEST", status, code)
	}
}

// An answer about a human who holds nothing is a completed answer, and so is one
// about a human this deployment has never heard of. Neither is a refusal, and
// the two must not be told apart — the echoed human_id is what a client
// corroborates against, and a 404 for one of them would let any caller ask
// whether a person exists.
func TestAHumanWithNothingAndAHumanWhoIsNotThereAnswerAlike(t *testing.T) {
	handler := fuzzHandler(t)
	body := func(human string) string {
		return `{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":"` + human + `"},"options":{}}`
	}
	answers := map[string]string{}
	for name, human := range map[string]string{
		// In the fixture and holding nothing, against never seen at all.
		"holds nothing": "fi7io4lvk35s",
		"is not there":  "fn2q6v8sbo1e",
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/acme/abv/applications/hrms/authority.resolve", strings.NewReader(body(human)))
		request.Header.Set("Authorization", "Bearer "+lab.WorkloadToken)
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s answered %d, want 200 — resolving nothing is an answer: %s", name, recorder.Code, recorder.Body.String())
		}
		answers[name] = strings.Replace(recorder.Body.String(), human, "<human>", 1)
	}
	if answers["holds nothing"] != answers["is not there"] {
		t.Fatalf("the two answers differ apart from the subject:\n%s\n%s", answers["holds nothing"], answers["is not there"])
	}
	if !strings.Contains(answers["holds nothing"], `"resolved_grants":[]`) {
		t.Fatalf("an empty answer is not an empty grant list: %s", answers["holds nothing"])
	}
}
