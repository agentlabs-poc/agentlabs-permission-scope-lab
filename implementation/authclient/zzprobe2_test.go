package authclient_test

import (
	"agentlabs.local/authclient"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Does the client follow a redirect from the Auth service to another host,
// and does it carry the Authorization header there?
func TestProbeRedirect(t *testing.T) {
	var got http.Header
	var gotPath string
	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, gotPath = r.Header.Clone(), r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		// the attacker answers with a blanket grant, echoing what was asked for
		fmt.Fprint(w, `{"version":"1","tenant_id":"acme","application_id":"hrms","human_id":"`+human+`","resolved_grants":[{"version":"1","grant_id":"attackergrant","revision":1,"permissions":["`+perm+`"],"scope":{}}]}`)
	}))
	defer evil.Close()

	real := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, evil.URL+"/api/v1/acme/abv/applications/hrms/authority.resolve", http.StatusTemporaryRedirect)
	}))
	defer real.Close()

	src, err := authclient.New(real.URL, authclient.Credential{Type: "service_account", ID: "agent_hrms", Bearer: "SUPERSECRET"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	auth, err := src.Load(context.Background(), query())
	t.Logf("redirect: err=%v routes=%+v", err, auth.Routes)
	t.Logf("attacker received path=%q authorization=%q", gotPath, got.Get("Authorization"))
}

// Same, but the redirect target keeps the method? 301/302 on POST downgrades to GET.
func TestProbeRedirect302(t *testing.T) {
	var method string
	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"version":"1","tenant_id":"acme","application_id":"hrms","human_id":"`+human+`","resolved_grants":[{"version":"1","grant_id":"attackergrant","revision":1,"permissions":["`+perm+`"],"scope":{}}]}`)
	}))
	defer evil.Close()
	real := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, evil.URL+"/x", http.StatusFound)
	}))
	defer real.Close()
	src, _ := authclient.New(real.URL, authclient.Credential{Type: "service_account", ID: "a", Bearer: "SUPERSECRET"}, nil)
	auth, err := src.Load(context.Background(), query())
	t.Logf("302: method seen by attacker=%q err=%v routes=%d", method, err, len(auth.Routes))
}

// Base URL with a path/query/userinfo — what does JoinPath do?
func TestProbeBaseURL(t *testing.T) {
	for _, base := range []string{
		"http://auth.example/sub/path",
		"http://auth.example?x=1",
		"http://auth.example/#frag",
		"http://user:pass@auth.example",
		"file:///etc/passwd",
		"HTTP://AUTH.EXAMPLE",
		"http://auth.example/..",
	} {
		s, err := authclient.New(base, authclient.Credential{Type: "t", ID: "i"}, nil)
		t.Logf("base %-32q -> err=%v src=%v", base, err, s != nil)
	}
}

// Path traversal through tenant / application ids in the query.
func TestProbeRouteInjection(t *testing.T) {
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.Path+"?"+r.URL.RawQuery)
		w.WriteHeader(500)
	}))
	defer srv.Close()
	for _, pair := range [][2]string{
		{"../../..", "hrms"},
		{"acme", "hrms/../../../../admin"},
		{"acme/../evil", "hrms"},
		{"acme?x=", "hrms"},
		{"acme#", "hrms"},
	} {
		src, _ := authclient.New(srv.URL, authclient.Credential{Type: "t", ID: "i"}, srv.Client())
		q := query()
		q.Context.Area.TenantID, q.Context.Area.ApplicationID = pair[0], pair[1]
		_, _ = src.Load(context.Background(), q)
	}
	for _, p := range seen {
		t.Logf("server saw: %s", p)
	}
}
