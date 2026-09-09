package authmiddleware

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func route(grants []string, predicates ...Predicate) Route {
	r := validRequest()
	return Route{Area: r.Context.Area, HumanID: "maya", Permission: r.Permission, GrantIDs: grants, Predicates: predicates}
}

func evaluate(t *testing.T, request Request, authority Authority, now time.Time) (Result, error) {
	t.Helper()
	evaluator, err := New(&trustedSource{authority: authority}, fixedClock{now: now})
	if err != nil {
		t.Fatal(err)
	}
	return evaluator.Evaluate(t.Context(), request)
}

func TestEvaluateAllowsOneCompleteApplicableRoute(t *testing.T) {
	request := validRequest()
	source := &trustedSource{authority: Authority{Routes: []Route{route([]string{"G-parent", "G-child"},
		Predicate{Key: "department", Value: "FIN", SourceGrantID: "G-parent"},
		Predicate{Key: "employee", Value: "$self", SourceGrantID: "G-child"},
	)}}}
	evaluator, err := New(source, fixedClock{now: time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	got, err := evaluator.Evaluate(t.Context(), request)
	want := Result{Version: "1", Decision: Allow, GrantIDs: []string{"G-parent", "G-child"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Evaluate() = %#v, %v", got, err)
	}
	if source.query != (AuthorityQuery{Context: request.Context, Permission: request.Permission}) {
		t.Fatalf("query = %#v", source.query)
	}
	source.authority.Routes[0].GrantIDs[0] = "mutated"
	if got.GrantIDs[0] != "G-parent" {
		t.Fatal("result aliases source grant IDs")
	}
}

func TestEvaluateDeniesWithoutAnIndependentlyMatchingRoute(t *testing.T) {
	request := validRequest()
	tests := map[string]Authority{
		"empty complete evidence":    {},
		"wrong boundary":             {Routes: []Route{route([]string{"G1"}, Predicate{Key: "department", Value: "ENG", SourceGrantID: "G1"})}},
		"all does not match literal": {Routes: []Route{route([]string{"G1"}, Predicate{Key: "department", Value: "FIN", SourceGrantID: "G1"})}},
		"parent predicates use AND": {Routes: []Route{route([]string{"G1", "G2"},
			Predicate{Key: "department", Value: "FIN", SourceGrantID: "G1"},
			Predicate{Key: "employee", Value: "someone-else", SourceGrantID: "G2"})}},
		"routes cannot mix": {Routes: []Route{
			route([]string{"G-read"}, Predicate{Key: "department", Value: "ENG", SourceGrantID: "G-read"}),
			route([]string{"G-fin"}, Predicate{Key: "department", Value: "FIN", SourceGrantID: "G-fin"}, Predicate{Key: "employee", Value: "someone-else", SourceGrantID: "G-fin"}),
		}},
	}
	for name, authority := range tests {
		t.Run(name, func(t *testing.T) {
			r := request
			if name == "all does not match literal" {
				r.Material = Material{"department": {Kind: SelectionAll}}
			}
			got, err := evaluate(t, r, authority, time.Time{})
			if err != nil || got.Decision != Deny || got.ErrorCode != "NO_AUTHORIZING_GRANT" || got.ErrorMessage == "" || got.ErrorMessageReason == "" {
				t.Fatalf("Evaluate() = %#v, %v", got, err)
			}
		})
	}
}

func TestEvaluateValidatesEveryRouteBeforeAllowing(t *testing.T) {
	request := validRequest()
	valid := route([]string{"G1"}, Predicate{Key: "department", Value: "FIN", SourceGrantID: "G1"})
	cases := map[string]Route{
		"wrong area":        func() Route { r := valid; r.Area.TenantID = "other"; return r }(),
		"wrong human":       func() Route { r := valid; r.HumanID = "other"; return r }(),
		"wrong permission":  func() Route { r := valid; r.Permission = "certificate::write"; return r }(),
		"duplicate grant":   func() Route { r := valid; r.GrantIDs = []string{"G1", "G1"}; return r }(),
		"missing grant":     func() Route { r := valid; r.GrantIDs = nil; return r }(),
		"empty key":         route([]string{"G1"}, Predicate{Value: "FIN", SourceGrantID: "G1"}),
		"empty value":       route([]string{"G1"}, Predicate{Key: "department", SourceGrantID: "G1"}),
		"foreign source":    route([]string{"G1"}, Predicate{Key: "department", Value: "FIN", SourceGrantID: "G2"}),
		"unsupported token": route([]string{"G1"}, Predicate{Key: "department", Value: "$manager", SourceGrantID: "G1"}),
	}
	for name, malformed := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := evaluate(t, request, Authority{Routes: []Route{valid, malformed}}, time.Time{})
			if err == nil || !reflect.DeepEqual(got, Result{}) {
				t.Fatalf("malformed later route hidden by allow: %#v, %v", got, err)
			}
		})
	}
}

func TestEvaluateExpiredRouteDoesNotSuppressValidRouteAndOrderingIsStable(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	expiredAt := now
	expired := route([]string{"A"})
	expired.ValidUntil = &expiredAt
	futureAt := now.Add(time.Second)
	future := route([]string{"AA"})
	future.ValidFrom = &futureAt
	got, err := evaluate(t, validRequest(), Authority{Routes: []Route{route([]string{"Z"}), expired, future, route([]string{"B", "A"}), route([]string{"B"})}}, now)
	if err != nil || !reflect.DeepEqual(got.GrantIDs, []string{"B"}) {
		t.Fatalf("Evaluate() = %#v, %v", got, err)
	}
}

func TestEvaluateRejectsInvertedValidityInterval(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	from, until := now.Add(time.Hour), now
	malformed := route([]string{"G1"})
	malformed.ValidFrom, malformed.ValidUntil = &from, &until
	if got, err := evaluate(t, validRequest(), Authority{Routes: []Route{malformed}}, now); err == nil || !reflect.DeepEqual(got, Result{}) {
		t.Fatalf("accepted inverted validity: %#v, %v", got, err)
	}
}

func TestEvaluateUsesDirectHumanForGroupSelf(t *testing.T) {
	request := validRequest()
	got, err := evaluate(t, request, Authority{Routes: []Route{route([]string{"group-grant"}, Predicate{Key: "employee", Value: "$self", SourceGrantID: "group-grant"})}}, time.Time{})
	if err != nil || got.Decision != Allow {
		t.Fatalf("Evaluate() = %#v, %v", got, err)
	}
}

func TestEvaluatePreservesSourceAndCancellationErrors(t *testing.T) {
	sentinel := errors.New("incomplete read")
	evaluationErr := &EvaluationError{Version: "1", Code: "KNOWN", Message: "message", MessageReason: "reason", Cause: sentinel}
	for name, sourceErr := range map[string]error{"ordinary": sentinel, "evaluation": evaluationErr} {
		t.Run(name, func(t *testing.T) {
			evaluator, err := New(&trustedSource{err: sourceErr}, fixedClock{})
			if err != nil {
				t.Fatal(err)
			}
			got, err := evaluator.Evaluate(t.Context(), validRequest())
			if !reflect.DeepEqual(got, Result{}) || !errors.Is(err, sentinel) {
				t.Fatalf("Evaluate() = %#v, %v", got, err)
			}
			if name == "evaluation" {
				var actual *EvaluationError
				if !errors.As(err, &actual) || actual != evaluationErr {
					t.Fatalf("EvaluationError not preserved: %v", err)
				}
			}
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	evaluator, _ := New(&trustedSource{}, fixedClock{})
	got, err := evaluator.Evaluate(ctx, validRequest())
	if !reflect.DeepEqual(got, Result{}) || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Evaluate() = %#v, %v", got, err)
	}
}

func TestEvaluateRejectsSafetyCeilingOverflow(t *testing.T) {
	request := validRequest()
	material := make(Material, 10001)
	for i := 0; i < 10001; i++ {
		material[string(rune(0x1000+i))] = Selection{Kind: SelectionAll}
	}
	request.Material = material
	if got, err := evaluate(t, request, Authority{}, time.Time{}); err == nil || !reflect.DeepEqual(got, Result{}) {
		t.Fatalf("accepted excess material: %#v, %v", got, err)
	}

	routes := make([]Route, 10001)
	for i := range routes {
		routes[i] = route([]string{"G"})
	}
	if got, err := evaluate(t, validRequest(), Authority{Routes: routes}, time.Time{}); err == nil || !reflect.DeepEqual(got, Result{}) {
		t.Fatalf("accepted excess routes: %#v, %v", got, err)
	}

	grants := make([]string, 257)
	for i := range grants {
		grants[i] = string(rune(0x1000 + i))
	}
	if got, err := evaluate(t, validRequest(), Authority{Routes: []Route{route(grants)}}, time.Time{}); err == nil || !reflect.DeepEqual(got, Result{}) {
		t.Fatalf("accepted excess grants: %#v, %v", got, err)
	}

	predicates := make([]Predicate, 10001)
	for i := range predicates {
		predicates[i] = Predicate{Key: "department", Value: "FIN", SourceGrantID: "G"}
	}
	if got, err := evaluate(t, validRequest(), Authority{Routes: []Route{route([]string{"G"}, predicates...)}}, time.Time{}); err == nil || !reflect.DeepEqual(got, Result{}) {
		t.Fatalf("accepted excess predicates: %#v, %v", got, err)
	}
}
