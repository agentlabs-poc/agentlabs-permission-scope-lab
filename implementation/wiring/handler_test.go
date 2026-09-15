package wiring_test

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type refusingAgents struct{}

func (refusingAgents) Establish(*http.Request) (domain.Identity, error) {
	return domain.Identity{}, errors.New("no credential")
}

func post(t *testing.T, handler http.Handler, path, body string) (int, string) {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body))))
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
		"a blank subject":     {"/api/v1/acme/abv/applications/hrms/authority.resolve", `{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":" "},"options":{}}`, http.StatusBadRequest, "MALFORMED_REQUEST"},
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
