package wiring_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type refusingAgents struct{}

func (refusingAgents) Establish(*http.Request) (domain.Identity, error) {
	return domain.Identity{}, errors.New("no credential")
}

func post(t *testing.T, handler http.Handler, path, body string) (int, string) {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
	// The fixture gate verifies this now, so these tests ask as an established
	// caller and the refusals they observe are the ones they name.
	request.Header.Set("Authorization", "Bearer "+lab.WorkloadToken)
	handler.ServeHTTP(recorder, request)
	var answered map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &answered); err != nil {
		t.Fatalf("status %d was not JSON: %s", recorder.Code, recorder.Body.String())
	}
	code, _ := answered["error_code"].(string)
	return recorder.Code, code
}

const goodBody = `{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":"fi7io4lvjqio"},"options":{}}`

// Every refusal the handler can give, none of which any test observed: statusFor
// and codeFor both measured 0% coverage, so every branch of the error mapping
// could be changed with the suite green.
func TestTheHandlerRefusesWithTheRightKind(t *testing.T) {
	handler := fuzzHandler(t)
	for name, tc := range map[string]struct {
		path, body string
		status     int
		code       string
	}{
		"a malformed area":    {"/api/v1/%20/abv/applications/hrms/authority.resolve", goodBody, http.StatusBadRequest, "MALFORMED_AREA"},
		"not JSON":            {"/api/v1/acme/abv/applications/hrms/authority.resolve", `<html>`, http.StatusBadRequest, "MALFORMED_REQUEST"},
		"an unknown field":    {"/api/v1/acme/abv/applications/hrms/authority.resolve", `{"version":"1","surprise":true}`, http.StatusBadRequest, "MALFORMED_REQUEST"},
		"a repeated field":    {"/api/v1/acme/abv/applications/hrms/authority.resolve", `{"version":"1","version":"1","identity":{"version":"1","human_id":"a"},"options":{}}`, http.StatusBadRequest, "MALFORMED_REQUEST"},
		"a trailing document": {"/api/v1/acme/abv/applications/hrms/authority.resolve", goodBody + `{"evil":1}`, http.StatusBadRequest, "MALFORMED_REQUEST"},
		"another version":     {"/api/v1/acme/abv/applications/hrms/authority.resolve", `{"version":"9","identity":{"version":"1","human_id":"a"},"options":{}}`, http.StatusNotImplemented, "UNSUPPORTED_VERSION"},
		// The identity block carries its own contract version so that contract can
		// evolve independently of the transport one. It was decoded and thrown
		// away, so a v2 block was read under v1 rules.
		"an identity from another version": {"/api/v1/acme/abv/applications/hrms/authority.resolve", `{"version":"1","identity":{"version":"9","actor":{"type":"service_account","id":"agent_hrms"},"human_id":"fi7io4lvjqio"},"options":{}}`, http.StatusNotImplemented, "UNSUPPORTED_VERSION"},
		"an identity stating no version":   {"/api/v1/acme/abv/applications/hrms/authority.resolve", `{"version":"1","identity":{"actor":{"type":"service_account","id":"agent_hrms"},"human_id":"fi7io4lvjqio"},"options":{}}`, http.StatusNotImplemented, "UNSUPPORTED_VERSION"},
		// A body cannot prove who is asking. Claiming an actor other than the one
		// the credential established is refused rather than ignored, so the day
		// somebody wires delegation evidence in, the claim is already not free.
		"a body claiming another actor": {"/api/v1/acme/abv/applications/hrms/authority.resolve", `{"version":"1","identity":{"version":"1","actor":{"type":"user","id":"fi7io4lvjqio"},"human_id":"fi7io4lvjqio"},"options":{}}`, http.StatusBadRequest, "MISMATCHED_ACTOR"},
		// Same credential, same application, only the id misspelt — which is what
		// a client configured with the wrong --client sends on every request. It
		// must say so, rather than read as an entitlement refusal that the
		// application then renders as a permanent outage.
		"a body claiming a near-miss of its own actor": {"/api/v1/acme/abv/applications/hrms/authority.resolve", `{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrm"},"human_id":"fi7io4lvjqio"},"options":{}}`, http.StatusBadRequest, "MISMATCHED_ACTOR"},
		"a blank subject": {"/api/v1/acme/abv/applications/hrms/authority.resolve", `{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":" "},"options":{}}`, http.StatusBadRequest, "MALFORMED_REQUEST"},
	} {
		t.Run(name, func(t *testing.T) {
			status, code := post(t, handler, tc.path, tc.body)
			if status != tc.status || code != tc.code {
				t.Fatalf("got %d/%s, want %d/%s", status, code, tc.status, tc.code)
			}
		})
	}
}

// One valid credential must not be able to map the deployment. Splitting "not
// installed" from "not yours" let a caller enumerate which tenants exist and
// which applications each holds — invisible in one process, a scanner on the
// wire.
func TestAreasTheCallerCannotAskAboutAreIndistinguishable(t *testing.T) {
	handler := fuzzHandler(t)
	answers := map[string][2]any{}
	for name, path := range map[string]string{
		"installed, not this caller's": "/api/v1/acme/abv/applications/crm/authority.resolve",
		"a tenant that does not exist": "/api/v1/nosuchtenant/abv/applications/hrms/authority.resolve",
		"an application not installed": "/api/v1/acme/abv/applications/nosuchapp/authority.resolve",
	} {
		status, code := post(t, handler, path, goodBody)
		answers[name] = [2]any{status, code}
	}
	var first [2]any
	for name, got := range answers {
		if first[0] == nil {
			first = got
			continue
		}
		if got != first {
			t.Fatalf("%q answered %v where another answered %v — the difference maps the deployment", name, got, first)
		}
	}
	if first[0] != http.StatusForbidden {
		t.Fatalf("status = %v, want 403 for every area the caller is not established in", first[0])
	}
}

// An unauthenticated caller is turned away before the body is parsed, so it gets
// no schema feedback and the service decodes nothing on its behalf.
func TestAnUnauthenticatedCallerIsRefusedBeforeParsing(t *testing.T) {
	service := fuzzService(t)
	handler, err := service.Handler(refusingAgents{})
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"a sound body":     goodBody,
		"an unsound body":  `<html>`,
		"an oversize body": string(make([]byte, 1<<17)),
	} {
		t.Run(name, func(t *testing.T) {
			status, code := post(t, handler, "/api/v1/acme/abv/applications/hrms/authority.resolve", body)
			if status != http.StatusUnauthorized || code != "UNAUTHENTICATED" {
				t.Fatalf("got %d/%s, want 401/UNAUTHENTICATED — the body was read first", status, code)
			}
		})
	}
}

func TestHandlerRequiresAnAgentIdentity(t *testing.T) {
	if handler, err := fuzzService(t).Handler(nil); err == nil || handler != nil {
		t.Fatal("mounted without a way to establish the caller")
	}
}

// Only a question the service accepted, decoded and answered is ever observed.
//
// The record a demonstration reads is evidence, so a caller that cannot
// authenticate must not be able to put anything into it — and must not be able
// to make the service buffer its body first either. An earlier version of this
// observed through a wrapper around the handler, which had to read the body to
// see the question and so read it before the authentication below: an
// unauthenticated caller could forge lines into the record, which a review
// demonstrated by writing one.
func TestOnlyAnAnsweredQuestionIsObserved(t *testing.T) {
	service := fuzzService(t)
	var observed []string
	record := func(method, path string, question []byte) {
		observed = append(observed, method+" "+path+" "+string(question))
	}
	refused, err := service.Handler(refusingAgents{}, record)
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"a sound body":        goodBody,
		"an unsound body":     "<html>\n  auth  <- POST /api/v1/victim/x/authority.resolve",
		"an oversize body":    string(make([]byte, 1<<17)),
		"a forged line break": `{"version":"1","identity":{"human_id":"a` + "\n" + `"}}`,
	} {
		t.Run("unauthenticated, "+name, func(t *testing.T) {
			if status, _ := post(t, refused, "/api/v1/acme/abv/applications/hrms/authority.resolve", body); status != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", status)
			}
			if len(observed) != 0 {
				t.Fatalf("an unauthenticated caller wrote into the record: %q", observed)
			}
		})
	}

	answering, err := service.Handler(agents{}, record)
	if err != nil {
		t.Fatal(err)
	}
	// Refused for its own reasons, after authentication: still not an answer, so
	// still not observed.
	for name, tc := range map[string]struct{ path, body string }{
		"malformed":       {"/api/v1/acme/abv/applications/hrms/authority.resolve", "<html>"},
		"another version": {"/api/v1/acme/abv/applications/hrms/authority.resolve", `{"version":"9","identity":{"version":"1","human_id":"a"},"options":{}}`},
		"another tenant":  {"/api/v1/globex/abv/applications/hrms/authority.resolve", goodBody},
	} {
		t.Run("refused, "+name, func(t *testing.T) {
			post(t, answering, tc.path, tc.body)
			if len(observed) != 0 {
				t.Fatalf("a refused question was recorded: %q", observed)
			}
		})
	}

	// And an answered one is, or nothing above would mean anything.
	if status, _ := post(t, answering, "/api/v1/acme/abv/applications/hrms/authority.resolve", goodBody); status != http.StatusOK {
		t.Fatalf("the sound question was not answered: %d", status)
	}
	if len(observed) != 1 {
		t.Fatalf("an answered question was recorded %d times, want 1: %q", len(observed), observed)
	}
	if !strings.Contains(observed[0], "fi7io4lvjqio") {
		t.Fatalf("the record does not carry the question: %q", observed[0])
	}
}
