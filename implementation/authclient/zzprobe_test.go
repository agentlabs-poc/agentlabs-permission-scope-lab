package authclient_test

import (
	"agentlabs.local/authclient"
	"agentlabs.local/authmiddleware"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type clk struct{ t time.Time }

func (c clk) Now() time.Time { return c.t }

const (
	human = "fi7io4lvjqio"
	perm  = "hrms:payroll:payslip::read"
)

func hostile(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
}

func query() authmiddleware.AuthorityQuery {
	return authmiddleware.AuthorityQuery{
		Context: authmiddleware.RequestContext{
			Area:     authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"},
			Identity: authmiddleware.Identity{Version: "1", Actor: authmiddleware.Actor{Type: "user", ID: human}, HumanID: human},
		},
		Permission: perm,
	}
}

func decide(t *testing.T, body string, material authmiddleware.Material) {
	t.Helper()
	srv := hostile(body)
	defer srv.Close()
	src, err := authclient.New(srv.URL, authclient.Credential{Type: "service_account", ID: "agent_hrms"}, srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	auth, err := src.Load(context.Background(), query())
	if err != nil {
		t.Logf("LOAD ERROR: %v", err)
		return
	}
	t.Logf("LOADED %d routes: %+v", len(auth.Routes), auth.Routes)
	ev, err := authmiddleware.New(src, clk{time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	q := query()
	res, err := ev.Evaluate(context.Background(), authmiddleware.Request{
		Context: q.Context, Permission: perm, Material: material,
	})
	if err != nil {
		t.Logf("EVALUATE ERROR: %v", err)
		return
	}
	t.Logf(">>> DECISION = %s  grants=%v code=%s", res.Decision, res.GrantIDs, res.ErrorCode)
}

func env(grants string) string {
	return `{"version":"1","tenant_id":"acme","application_id":"hrms","human_id":"` + human + `","resolved_grants":[` + grants + `]}`
}

// 1. server returns a grant for a DIFFERENT permission than the one asked about.
func TestProbeWrongPermission(t *testing.T) {
	decide(t, env(`{"version":"1","grant_id":"fk3x9r2m0dq3","revision":1,"permissions":["hrms:directory:profile::read"],"scope":{"dept":"FIN"}}`),
		authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionExact, Value: "FIN"}})
}

// 2. empty scope -> unconstrained route.
func TestProbeEmptyScope(t *testing.T) {
	decide(t, env(`{"version":"1","grant_id":"fk3x9r2m0dq3","revision":1,"permissions":["`+perm+`"],"scope":{}}`),
		authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionExact, Value: "ENG"}, "cert": {Kind: authmiddleware.SelectionExact, Value: "C18"}})
}

// 2b. scope omitted entirely (null).
func TestProbeNullScope(t *testing.T) {
	decide(t, env(`{"version":"1","grant_id":"fk3x9r2m0dq3","revision":1,"permissions":["`+perm+`"],"scope":null}`),
		authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionExact, Value: "ENG"}})
}

// 3. duplicate scope keys on the wire.
func TestProbeDuplicateScopeKeys(t *testing.T) {
	decide(t, env(`{"version":"1","grant_id":"fk3x9r2m0dq3","revision":1,"permissions":["`+perm+`"],"scope":{"dept":"FIN","dept":"ENG"}}`),
		authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionExact, Value: "ENG"}})
}

// 3b. duplicate top-level envelope fields.
func TestProbeDuplicateEnvelopeFields(t *testing.T) {
	body := `{"version":"1","tenant_id":"evil","tenant_id":"acme","application_id":"hrms","human_id":"` + human + `","resolved_grants":[{"version":"1","grant_id":"fk3x9r2m0dq3","revision":1,"permissions":["` + perm + `"],"scope":{}}]}`
	decide(t, body, authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionExact, Value: "ENG"}})
}

// 4. $self smuggled from the server.
func TestProbeSelf(t *testing.T) {
	decide(t, env(`{"version":"1","grant_id":"fk3x9r2m0dq3","revision":1,"permissions":["`+perm+`"],"scope":{"user":"$self"}}`),
		authmiddleware.Material{"user": {Kind: authmiddleware.SelectionExact, Value: human}})
}

// 5. absurd validity window.
func TestProbeValidity(t *testing.T) {
	decide(t, env(`{"version":"1","grant_id":"fk3x9r2m0dq3","revision":1,"permissions":["`+perm+`"],"scope":{},"validity":{"expires_at":"9999-12-31T23:59:59Z"}}`),
		authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionExact, Value: "ENG"}})
}

// 6. trailing garbage after the JSON document.
func TestProbeTrailingJSON(t *testing.T) {
	decide(t, env(`{"version":"1","grant_id":"fk3x9r2m0dq3","revision":1,"permissions":["`+perm+`"],"scope":{"dept":"FIN"}}`)+`{"version":"1","tenant_id":"x"}`,
		authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionExact, Value: "FIN"}})
}

// 7. deeply nested body.
func TestProbeDeepNesting(t *testing.T) {
	deep := ""
	for i := 0; i < 200000; i++ {
		deep += "["
	}
	srv := hostile(`{"version":"1","tenant_id":"acme","application_id":"hrms","human_id":"` + human + `","resolved_grants":` + deep)
	defer srv.Close()
	src, _ := authclient.New(srv.URL, authclient.Credential{Type: "service_account", ID: "a"}, srv.Client())
	_, err := src.Load(context.Background(), query())
	t.Logf("deep nesting err = %v", err)
}

// 8. what the client does with each non-200 status.
func TestProbeStatuses(t *testing.T) {
	for _, code := range []int{201, 204, 301, 400, 401, 403, 404, 500, 503} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(map[string]string{"version": "1", "error_code": "X"})
		}))
		src, _ := authclient.New(srv.URL, authclient.Credential{Type: "service_account", ID: "a"}, srv.Client())
		_, err := src.Load(context.Background(), query())
		t.Logf("status %d -> %v", code, err)
		srv.Close()
	}
}

// 9. server echoes a different human / area.
func TestProbeWrongEcho(t *testing.T) {
	decide(t, `{"version":"1","tenant_id":"acme","application_id":"hrms","human_id":"someone-else","resolved_grants":[]}`, nil)
	decide(t, `{"version":"1","tenant_id":"other","application_id":"hrms","human_id":"`+human+`","resolved_grants":[]}`, nil)
}

// 10. oversize + slow body.
func TestProbeOversize(t *testing.T) {
	big := make([]byte, 0)
	for len(big) < (1<<20)+100 {
		big = append(big, ' ')
	}
	srv := hostile(string(big))
	defer srv.Close()
	src, _ := authclient.New(srv.URL, authclient.Credential{Type: "service_account", ID: "a"}, srv.Client())
	_, err := src.Load(context.Background(), query())
	t.Logf("oversize err = %v", err)
}

// 11. no Doer, no ctx deadline: does the default 5s timeout actually bound it?
func TestProbeSlowLoris(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		time.Sleep(30 * time.Second)
	}))
	defer srv.Close()
	// caller-supplied Doer with no timeout, as wiring/endtoend_test.go does
	src, _ := authclient.New(srv.URL, authclient.Credential{Type: "service_account", ID: "a"}, srv.Client())
	start := time.Now()
	done := make(chan error, 1)
	go func() { _, err := src.Load(context.Background(), query()); done <- err }()
	select {
	case err := <-done:
		t.Logf("caller Doer: returned after %v: %v", time.Since(start), err)
	case <-time.After(8 * time.Second):
		t.Logf("caller Doer: STILL BLOCKED after 8s (no timeout, no deadline)")
	}
}
