package authmiddleware

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type httpIdentitySource struct {
	requestContext RequestContext
	err            error
	called         int
	body           string
}

func (s *httpIdentitySource) Establish(_ context.Context, request *http.Request) (RequestContext, error) {
	s.called++
	raw, _ := io.ReadAll(request.Body)
	s.body = string(raw)
	return s.requestContext, s.err
}

type httpAuthoritySource struct {
	authority Authority
	err       error
	query     AuthorityQuery
	called    int
}

func (s *httpAuthoritySource) Load(_ context.Context, query AuthorityQuery) (Authority, error) {
	s.called++
	s.query = query
	return s.authority, s.err
}

type httpClock struct{}

func (httpClock) Now() time.Time { return time.Time{} }

func httpContext() RequestContext {
	return RequestContext{Area: Area{TenantID: "acme", ApplicationID: "hrms"}, Identity: Identity{Version: "1", Actor: Actor{Type: "user", ID: "maya"}, HumanID: "maya"}}
}

func httpPolicy(method string) Policy {
	return Policy{Version: "1", Method: method, Path: "/api/{tenant}/{application}/certificates/{cert}", Permission: "certificate::write", Inputs: map[string]Input{
		"cert": {Source: SourcePath, Name: "cert"},
		"dept": {Source: SourceBody, Name: "department_id"},
	}}
}

func httpEvaluator(t *testing.T, source *httpAuthoritySource) *Evaluator {
	t.Helper()
	e, err := New(source, httpClock{})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func allowAuthority() Authority {
	return Authority{Routes: []Route{{Area: httpContext().Area, HumanID: "maya", Permission: "certificate::write", GrantIDs: []string{"G1"}, Predicates: []Predicate{{Key: "cert", Value: "C17", SourceGrantID: "G1"}, {Key: "dept", Value: "FIN", SourceGrantID: "G1"}}}}}
}

func TestWrapAllowsExactlyOneBoundEffectWithExactInputs(t *testing.T) {
	identity := &httpIdentitySource{requestContext: httpContext()}
	authority := &httpAuthoritySource{authority: allowAuthority()}
	var gotValues InputValues
	var gotBody map[string]json.RawMessage
	executed := 0
	binder := func(_ context.Context, got RequestContext, values InputValues, body map[string]json.RawMessage) (BoundOperation, error) {
		if got != httpContext() {
			t.Fatalf("context = %#v", got)
		}
		gotValues, gotBody = values, body
		return BoundOperation{Material: Material{"cert": {Kind: SelectionExact, Value: "C17"}, "dept": {Kind: SelectionExact, Value: "FIN"}}, Execute: func(_ context.Context, w http.ResponseWriter) {
			executed++
			w.WriteHeader(http.StatusNoContent)
		}}, nil
	}
	failures := 0
	h, err := Wrap(httpPolicy(http.MethodPut), identity, httpEvaluator(t, authority), binder, func(http.ResponseWriter, *http.Request, Result, error) { failures++ })
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPut, "/api/acme/hrms/certificates/C17?department_id=WRONG", strings.NewReader(`{"department_id":"FIN","note":null}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent || executed != 1 || failures != 0 || identity.called != 1 || authority.called != 1 {
		t.Fatalf("code=%d executed=%d failures=%d identity=%d evaluate=%d", w.Code, executed, failures, identity.called, authority.called)
	}
	if string(gotValues["cert"]) != `"C17"` || string(gotValues["dept"]) != `"FIN"` || string(gotBody["note"]) != "null" {
		t.Fatalf("values=%q body=%q", gotValues, gotBody)
	}
	if identity.body != "" {
		t.Fatalf("identity source received business body %q", identity.body)
	}
	if authority.query.Permission != "certificate::write" || authority.query.Context != httpContext() {
		t.Fatalf("query = %#v", authority.query)
	}
}

func TestWrapStopsBeforeProtectedEffect(t *testing.T) {
	identityErr := errors.New("identity failed")
	evaluateErr := errors.New("authority failed")
	binderErr := errors.New("binding failed")
	tests := []struct {
		name      string
		method    string
		target    string
		body      string
		identity  *httpIdentitySource
		authority *httpAuthoritySource
		cancel    bool
		binderErr error
		wantErr   error
		wantDeny  bool
	}{
		{"identity error", http.MethodPut, "/api/acme/hrms/certificates/C17", `{}`, &httpIdentitySource{err: identityErr}, &httpAuthoritySource{authority: allowAuthority()}, false, nil, identityErr, false},
		{"invalid identity", http.MethodPut, "/api/acme/hrms/certificates/C17", `{}`, &httpIdentitySource{requestContext: RequestContext{Area: httpContext().Area}}, &httpAuthoritySource{authority: allowAuthority()}, false, nil, nil, false},
		{"tenant mismatch", http.MethodPut, "/api/other/hrms/certificates/C17", `{}`, &httpIdentitySource{requestContext: httpContext()}, &httpAuthoritySource{authority: allowAuthority()}, false, nil, nil, false},
		{"application mismatch", http.MethodPut, "/api/acme/other/certificates/C17", `{}`, &httpIdentitySource{requestContext: httpContext()}, &httpAuthoritySource{authority: allowAuthority()}, false, nil, nil, false},
		{"missing selected body does not use query", http.MethodPut, "/api/acme/hrms/certificates/C17?department_id=FIN", `{}`, &httpIdentitySource{requestContext: httpContext()}, &httpAuthoritySource{authority: allowAuthority()}, false, nil, nil, false},
		{"invalid body", http.MethodPut, "/api/acme/hrms/certificates/C17", `{"department_id":"FIN","department_id":"ENG"}`, &httpIdentitySource{requestContext: httpContext()}, &httpAuthoritySource{authority: allowAuthority()}, false, nil, nil, false},
		{"binder error", http.MethodPut, "/api/acme/hrms/certificates/C17", `{"department_id":"FIN"}`, &httpIdentitySource{requestContext: httpContext()}, &httpAuthoritySource{authority: allowAuthority()}, false, binderErr, binderErr, false},
		{"deny", http.MethodPut, "/api/acme/hrms/certificates/C17", `{"department_id":"FIN"}`, &httpIdentitySource{requestContext: httpContext()}, &httpAuthoritySource{}, false, nil, nil, true},
		{"evaluation error", http.MethodPut, "/api/acme/hrms/certificates/C17", `{"department_id":"FIN"}`, &httpIdentitySource{requestContext: httpContext()}, &httpAuthoritySource{err: evaluateErr}, false, nil, evaluateErr, false},
		{"cancelled", http.MethodPut, "/api/acme/hrms/certificates/C17", `{"department_id":"FIN"}`, &httpIdentitySource{requestContext: httpContext()}, &httpAuthoritySource{authority: allowAuthority()}, true, nil, context.Canceled, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			executed, bound := 0, 0
			var gotResult Result
			var gotErr error
			binder := func(ctx context.Context, _ RequestContext, _ InputValues, _ map[string]json.RawMessage) (BoundOperation, error) {
				bound++
				if tc.cancel {
					if cancel, ok := ctx.Value(cancelKey{}).(context.CancelFunc); ok {
						cancel()
					}
				}
				if tc.binderErr != nil {
					return BoundOperation{}, tc.binderErr
				}
				return BoundOperation{Material: Material{"cert": {Kind: SelectionExact, Value: "C17"}, "dept": {Kind: SelectionExact, Value: "FIN"}}, Execute: func(context.Context, http.ResponseWriter) { executed++ }}, nil
			}
			h, err := Wrap(httpPolicy(http.MethodPut), tc.identity, httpEvaluator(t, tc.authority), binder, func(_ http.ResponseWriter, _ *http.Request, result Result, err error) {
				gotResult, gotErr = result, err
			})
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.body))
			if tc.cancel {
				ctx, cancel := context.WithCancel(r.Context())
				r = r.WithContext(context.WithValue(ctx, cancelKey{}, context.CancelFunc(cancel)))
			}
			h.ServeHTTP(httptest.NewRecorder(), r)
			if executed != 0 {
				t.Fatal("protected effect executed")
			}
			if tc.wantDeny {
				if gotResult.Decision != Deny || gotResult.ErrorMessage == "" || gotResult.ErrorMessageReason == "" || gotErr != nil {
					t.Fatalf("failure = %#v, %v", gotResult, gotErr)
				}
			} else if len(gotResult.GrantIDs) != 0 || gotResult.Version != "" || gotErr == nil {
				t.Fatalf("failure = %#v, %v", gotResult, gotErr)
			}
			if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("error %v does not preserve %v", gotErr, tc.wantErr)
			}
			if (strings.Contains(tc.name, "mismatch") || strings.Contains(tc.name, "missing") || strings.Contains(tc.name, "invalid body") || strings.Contains(tc.name, "identity")) && tc.authority.called != 0 {
				t.Fatal("evaluated before request was trusted and bound")
			}
			_ = bound
		})
	}
}

type cancelKey struct{}

func TestWrapConstructionRoutingAndPolicyCopy(t *testing.T) {
	identity := &httpIdentitySource{requestContext: httpContext()}
	authority := &httpAuthoritySource{authority: allowAuthority()}
	evaluator := httpEvaluator(t, authority)
	binder := Binder(func(context.Context, RequestContext, InputValues, map[string]json.RawMessage) (BoundOperation, error) {
		return BoundOperation{Material: Material{"cert": {Kind: SelectionExact, Value: "C17"}, "dept": {Kind: SelectionExact, Value: "FIN"}}, Execute: func(context.Context, http.ResponseWriter) {}}, nil
	})
	failure := FailureHandler(func(http.ResponseWriter, *http.Request, Result, error) {})
	valid := httpPolicy(http.MethodPut)
	var nilIdentity *httpIdentitySource
	var nilBinder Binder
	var nilFailure FailureHandler
	for name, args := range map[string]struct {
		p Policy
		i IdentitySource
		e *Evaluator
		b Binder
		f FailureHandler
	}{
		"invalid policy":          {Policy{}, identity, evaluator, binder, failure},
		"bad pattern":             {Policy{Version: "1", Method: "GET", Path: "/{x}/{y...}/z", Permission: "x::read", Inputs: map[string]Input{}}, identity, evaluator, binder, failure},
		"nil identity":            {valid, nilIdentity, evaluator, binder, failure},
		"nil evaluator":           {valid, identity, nil, binder, failure},
		"uninitialized evaluator": {valid, identity, &Evaluator{}, binder, failure},
		"nil binder":              {valid, identity, evaluator, nilBinder, failure},
		"nil failure":             {valid, identity, evaluator, binder, nilFailure},
	} {
		t.Run(name, func(t *testing.T) {
			if h, err := Wrap(args.p, args.i, args.e, args.b, args.f); err == nil || h != nil {
				t.Fatalf("Wrap() = %#v, %v", h, err)
			}
		})
	}

	p := httpPolicy(http.MethodPut)
	h, err := Wrap(p, identity, evaluator, binder, failure)
	if err != nil {
		t.Fatal(err)
	}
	p.Permission = "forged::write"
	p.Inputs["dept"] = Input{Source: SourceBody, Name: "forged"}
	r := httptest.NewRequest(http.MethodPut, "/api/acme/hrms/certificates/C17", strings.NewReader(`{"department_id":"FIN"}`))
	h.ServeHTTP(httptest.NewRecorder(), r)
	if authority.query.Permission != "certificate::write" {
		t.Fatalf("mutated permission used: %#v", authority.query)
	}

	before := identity.called
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/not-this-route", nil))
	if w.Code != http.StatusNotFound || identity.called != before {
		t.Fatalf("unmatched route: code=%d identity=%d", w.Code, identity.called)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/acme/hrms/certificates/C17", nil))
	if w.Code != http.StatusMethodNotAllowed || identity.called != before {
		t.Fatalf("unknown action: code=%d identity=%d", w.Code, identity.called)
	}
}

func TestWrapRejectsHeadAndOversizedBody(t *testing.T) {
	identity := &httpIdentitySource{requestContext: httpContext()}
	authority := &httpAuthoritySource{authority: allowAuthority()}
	failures := 0
	binder := Binder(func(context.Context, RequestContext, InputValues, map[string]json.RawMessage) (BoundOperation, error) {
		t.Fatal("bound invalid request")
		return BoundOperation{}, nil
	})
	h, err := Wrap(httpPolicy(http.MethodGet), identity, httpEvaluator(t, authority), binder, func(http.ResponseWriter, *http.Request, Result, error) { failures++ })
	if err != nil {
		t.Fatal(err)
	}
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodHead, "/api/acme/hrms/certificates/C17", nil))
	if failures != 1 || authority.called != 0 {
		t.Fatalf("HEAD failures=%d evaluate=%d", failures, authority.called)
	}

	h, err = Wrap(httpPolicy(http.MethodPut), identity, httpEvaluator(t, authority), binder, func(http.ResponseWriter, *http.Request, Result, error) { failures++ })
	if err != nil {
		t.Fatal(err)
	}
	body := `{"department_id":"` + strings.Repeat("x", maxJSONBytes) + `"}`
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPut, "/api/acme/hrms/certificates/C17", strings.NewReader(body)))
	if failures != 2 || authority.called != 0 {
		t.Fatalf("large body failures=%d evaluate=%d", failures, authority.called)
	}
}

func TestWrapRejectsNilExecuteBeforeEvaluation(t *testing.T) {
	identity := &httpIdentitySource{requestContext: httpContext()}
	authority := &httpAuthoritySource{authority: allowAuthority()}
	var gotErr error
	h, err := Wrap(httpPolicy(http.MethodPut), identity, httpEvaluator(t, authority), func(_ context.Context, _ RequestContext, values InputValues, _ map[string]json.RawMessage) (BoundOperation, error) {
		if string(values["dept"]) != "null" {
			t.Fatalf("selected null = %q", values["dept"])
		}
		return BoundOperation{}, nil
	}, func(_ http.ResponseWriter, _ *http.Request, _ Result, err error) { gotErr = err })
	if err != nil {
		t.Fatal(err)
	}
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPut, "/api/acme/hrms/certificates/C17", strings.NewReader(`{"department_id":null}`)))
	if gotErr == nil || authority.called != 0 {
		t.Fatalf("error=%v evaluated=%d", gotErr, authority.called)
	}
}

func TestWrapGETUsesRoutedPathWithoutARequestBody(t *testing.T) {
	policy := Policy{Version: "1", Method: http.MethodGet, Path: "/api/{tenant}/{application}/certificates/{cert}", Permission: "certificate::write", Inputs: map[string]Input{"cert": {Source: SourcePath, Name: "cert"}}}
	identity := &httpIdentitySource{requestContext: httpContext()}
	authority := allowAuthority()
	authority.Routes[0].Predicates = authority.Routes[0].Predicates[:1]
	executed := 0
	h, err := Wrap(policy, identity, httpEvaluator(t, &httpAuthoritySource{authority: authority}), func(_ context.Context, _ RequestContext, values InputValues, body map[string]json.RawMessage) (BoundOperation, error) {
		if string(values["cert"]) != `"C17"` || body != nil {
			t.Fatalf("values=%q body=%q", values, body)
		}
		return BoundOperation{Material: Material{"cert": {Kind: SelectionExact, Value: "C17"}}, Execute: func(context.Context, http.ResponseWriter) { executed++ }}, nil
	}, func(http.ResponseWriter, *http.Request, Result, error) { t.Fatal("GET failed") })
	if err != nil {
		t.Fatal(err)
	}
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/acme/hrms/certificates/C17", nil))
	if executed != 1 {
		t.Fatalf("executed = %d", executed)
	}
}

func TestAllowedSynchronousEffectCompletesAfterAuthorityWithdrawalAndNextRequestDenies(t *testing.T) {
	identity := &httpIdentitySource{requestContext: httpContext()}
	authority := &httpAuthoritySource{authority: allowAuthority()}
	var effects []string
	denials := 0
	h, err := Wrap(httpPolicy(http.MethodPut), identity, httpEvaluator(t, authority), func(_ context.Context, _ RequestContext, values InputValues, _ map[string]json.RawMessage) (BoundOperation, error) {
		var cert, dept string
		if err := json.Unmarshal(values["cert"], &cert); err != nil {
			return BoundOperation{}, err
		}
		if err := json.Unmarshal(values["dept"], &dept); err != nil {
			return BoundOperation{}, err
		}
		return BoundOperation{
			Material: Material{"cert": {Kind: SelectionExact, Value: cert}, "dept": {Kind: SelectionExact, Value: dept}},
			Execute: func(context.Context, http.ResponseWriter) {
				authority.authority = Authority{}
				effects = append(effects, cert+":"+dept)
			},
		}, nil
	}, func(_ http.ResponseWriter, _ *http.Request, result Result, err error) {
		if err != nil || result.Decision != Deny {
			t.Fatalf("failure = %#v, %v", result, err)
		}
		denials++
	})
	if err != nil {
		t.Fatal(err)
	}
	request := func() *http.Request {
		return httptest.NewRequest(http.MethodPut, "/api/acme/hrms/certificates/C17", strings.NewReader(`{"department_id":"FIN"}`))
	}
	h.ServeHTTP(httptest.NewRecorder(), request())
	h.ServeHTTP(httptest.NewRecorder(), request())
	if len(effects) != 1 || effects[0] != "C17:FIN" || denials != 1 || authority.called != 2 {
		t.Fatalf("effects=%v denials=%d evaluations=%d", effects, denials, authority.called)
	}
}

func TestBusinessBodyRejectsNonObjectTrailingInvalidUnicodeAndDepth(t *testing.T) {
	deep := strings.Repeat(`{"x":`, maxJSONDepth+1) + `0` + strings.Repeat(`}`, maxJSONDepth+1)
	for name, raw := range map[string]string{
		"array":           `[]`,
		"scalar":          `true`,
		"root null":       `null`,
		"trailing":        `{} {}`,
		"invalid unicode": `{"x":"\ud800"}`,
		"too deep":        deep,
	} {
		t.Run(name, func(t *testing.T) {
			if body, err := decodeBusinessBody(strings.NewReader(raw)); err == nil || body != nil {
				t.Fatalf("decodeBusinessBody() = %q, %v", body, err)
			}
		})
	}
}
