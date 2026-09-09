package authmiddleware

import (
	"context"
	"reflect"
	"testing"
	"time"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type trustedSource struct {
	authority Authority
	err       error
	query     AuthorityQuery
}

func (s *trustedSource) Load(ctx context.Context, query AuthorityQuery) (Authority, error) {
	s.query = query
	return s.authority, s.err
}

type nilSource struct{}

func (*nilSource) Load(context.Context, AuthorityQuery) (Authority, error) { return Authority{}, nil }

type nilClock struct{}

func (*nilClock) Now() time.Time { return time.Time{} }

func validRequest() Request {
	return Request{
		Context: RequestContext{
			Area:     Area{TenantID: "tenant-fin", ApplicationID: "hrms"},
			Identity: Identity{Version: "1", Actor: Actor{Type: "user", ID: "maya"}, HumanID: "maya"},
		},
		Permission: "certificate::read",
		Material: Material{
			"department": {Kind: SelectionExact, Value: "FIN"},
			"employee":   {Kind: SelectionExact, Value: "maya"},
		},
	}
}

func TestNewRejectsNilDependencies(t *testing.T) {
	var typedNilSource *nilSource
	var typedNilClock *nilClock
	for name, tc := range map[string]struct {
		source AuthoritySource
		clock  Clock
	}{
		"nil source":       {nil, fixedClock{}},
		"typed nil source": {typedNilSource, fixedClock{}},
		"nil clock":        {&trustedSource{}, nil},
		"typed nil clock":  {&trustedSource{}, typedNilClock},
	} {
		t.Run(name, func(t *testing.T) {
			if evaluator, err := New(tc.source, tc.clock); err == nil || evaluator != nil {
				t.Fatalf("New() = %#v, %v", evaluator, err)
			}
		})
	}
}

func TestEvaluateRejectsMalformedRequestBeforeLoading(t *testing.T) {
	valid := validRequest()
	cases := map[string]func(*Request){
		"missing tenant":       func(r *Request) { r.Context.Area.TenantID = "" },
		"wildcard application": func(r *Request) { r.Context.Area.ApplicationID = "*" },
		"unsupported identity": func(r *Request) { r.Context.Identity.Version = "2" },
		"proxy actor":          func(r *Request) { r.Context.Identity.Actor.Type = "service" },
		"indirect human":       func(r *Request) { r.Context.Identity.Actor.ID = "proxy" },
		"bad permission":       func(r *Request) { r.Permission = "certificate::*" },
		"empty material key":   func(r *Request) { r.Material[""] = Selection{Kind: SelectionAll} },
		"empty exact":          func(r *Request) { r.Material["department"] = Selection{Kind: SelectionExact} },
		"valued all":           func(r *Request) { r.Material["department"] = Selection{Kind: SelectionAll, Value: "FIN"} },
		"unknown selection":    func(r *Request) { r.Material["department"] = Selection{} },
	}
	for name, edit := range cases {
		t.Run(name, func(t *testing.T) {
			source := &trustedSource{}
			evaluator, err := New(source, fixedClock{})
			if err != nil {
				t.Fatal(err)
			}
			request := valid
			request.Material = make(Material, len(valid.Material))
			for key, selection := range valid.Material {
				request.Material[key] = selection
			}
			edit(&request)
			if result, err := evaluator.Evaluate(t.Context(), request); err == nil || !reflect.DeepEqual(result, Result{}) {
				t.Fatalf("Evaluate() = %#v, %v", result, err)
			}
			if source.query != (AuthorityQuery{}) {
				t.Fatalf("source loaded malformed request: %#v", source.query)
			}
		})
	}
}
