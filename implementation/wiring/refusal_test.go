package wiring_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"agentlabs.local/wiring"
	"context"
	"net/http"
	"path/filepath"
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
		status, code := post(t, handler, "/api/v1/acme/abv/applications/hrms/authority.resolve", goodBody)
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
