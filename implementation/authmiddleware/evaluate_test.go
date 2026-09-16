package authmiddleware

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func route(grants []string, predicates ...Predicate) Route {
	r := validRequest()
	return Route{Area: r.Context.Area, HumanID: "maya", Permissions: []string{r.Permission}, GrantIDs: grants, Predicates: predicates}
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
	if source.query != (AuthorityQuery{Context: request.Context}) {
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
		"wrong boundary":             {Routes: []Route{route([]string{"fk3x9r2m5iv8"}, Predicate{Key: "department", Value: "ENG", SourceGrantID: "fk3x9r2m5iv8"})}},
		"all does not match literal": {Routes: []Route{route([]string{"fk3x9r2m5iv8"}, Predicate{Key: "department", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"})}},
		"parent predicates use AND": {Routes: []Route{route([]string{"fk3x9r2m5iv8", "fk3x9r2man0d"},
			Predicate{Key: "department", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"},
			Predicate{Key: "employee", Value: "someone-else", SourceGrantID: "fk3x9r2man0d"})}},
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

func TestEvaluateAddingPredicateOnlyNarrowsRoute(t *testing.T) {
	broad := route([]string{"fk3x9r2m5iv8"})
	narrow := route([]string{"fk3x9r2m5iv8"}, Predicate{Key: "department", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"})
	for name, test := range map[string]struct {
		selection Selection
		want      Decision
	}{
		"matching exact":  {Selection{Kind: SelectionExact, Value: "FIN"}, Allow},
		"other exact":     {Selection{Kind: SelectionExact, Value: "ENG"}, Deny},
		"all departments": {Selection{Kind: SelectionAll}, Deny},
	} {
		t.Run(name, func(t *testing.T) {
			request := validRequest()
			request.Material["department"] = test.selection
			if got, err := evaluate(t, request, Authority{Routes: []Route{broad}}, time.Time{}); err != nil || got.Decision != Allow {
				t.Fatalf("broad Evaluate() = %#v, %v", got, err)
			}
			got, err := evaluate(t, request, Authority{Routes: []Route{narrow}}, time.Time{})
			if err != nil || got.Decision != test.want {
				t.Fatalf("narrow Evaluate() = %#v, %v; want %q", got, err, test.want)
			}
		})
	}
}

// A malformed route costs that route and nothing else.
//
// It used to fail the whole evaluation, which was defensible while the question
// named a permission — every route in the answer was then about this request.
// Unfiltered, the answer describes everything the human holds here, so one
// unusable grant took away every other grant they had. authority-lineage.md
// forbids that in terms, and Q-051's own table says a deny is not established by
// one failed grant "if another route could authorize it".
//
// Two things must hold, and the second is the one that matters: the other routes
// still decide, and the malformed route itself never authorizes.
func TestAMalformedRouteCostsThatRouteAndNothingElse(t *testing.T) {
	request := validRequest()
	valid := route([]string{"fk3x9r2m5iv8"}, Predicate{Key: "department", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"})
	cases := map[string]Route{
		"no permission":     func() Route { r := valid; r.Permissions = nil; return r }(),
		"duplicate grant":   func() Route { r := valid; r.GrantIDs = []string{"fk3x9r2m5iv8", "fk3x9r2m5iv8"}; return r }(),
		"missing grant":     func() Route { r := valid; r.GrantIDs = nil; return r }(),
		"too many grants":   func() Route { r := valid; r.GrantIDs = manyGrants(257); return r }(),
		"empty key":         route([]string{"fk3x9r2m5iv8"}, Predicate{Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"}),
		"empty value":       route([]string{"fk3x9r2m5iv8"}, Predicate{Key: "department", SourceGrantID: "fk3x9r2m5iv8"}),
		"foreign source":    route([]string{"fk3x9r2m5iv8"}, Predicate{Key: "department", Value: "FIN", SourceGrantID: "fk3x9r2man0d"}),
		"unsupported token": route([]string{"fk3x9r2m5iv8"}, Predicate{Key: "department", Value: "$manager", SourceGrantID: "fk3x9r2m5iv8"}),
		"inverted validity": func() Route {
			r := valid
			from, until := time.Date(2026, 9, 9, 13, 0, 0, 0, time.UTC), time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
			r.ValidFrom, r.ValidUntil = &from, &until
			return r
		}(),
	}
	for name, malformed := range cases {
		t.Run(name, func(t *testing.T) {
			// Beside a sound route, the sound one still decides.
			got, err := evaluate(t, request, Authority{Routes: []Route{valid, malformed}}, time.Time{})
			if err != nil || got.Decision != Allow || !reflect.DeepEqual(got.GrantIDs, []string{"fk3x9r2m5iv8"}) {
				t.Fatalf("one malformed route took away a sound one: %#v, %v", got, err)
			}
			// Alone, it authorizes nothing — and the refusal is an evaluation
			// failure, not a denial. Q-051 lets a failed grant stand aside for an
			// allow another route earns; it does not let one establish a deny,
			// because a route this gate could not read is exactly a route that
			// might have authorized.
			alone, err := evaluate(t, request, Authority{Routes: []Route{malformed}}, time.Time{})
			if err == nil || !reflect.DeepEqual(alone, Result{}) {
				t.Fatalf("a malformed route produced a decision: %#v, %v", alone, err)
			}
			var evaluation *EvaluationError
			if !errors.As(err, &evaluation) || evaluation.Code != "AUTHORITY_MALFORMED" {
				t.Fatalf("err = %v, want an evaluation failure", err)
			}
		})
	}
}

// The two that are not about one route. An answer describing another area or
// another human is the source answering about somebody else, and nothing in it
// can be trusted — so these still fail the evaluation outright.
func TestAnAnswerAboutSomebodyElseFailsTheEvaluation(t *testing.T) {
	request := validRequest()
	valid := route([]string{"fk3x9r2m5iv8"}, Predicate{Key: "department", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"})
	for name, wrong := range map[string]Route{
		"wrong tenant":      func() Route { r := valid; r.Area.TenantID = "other"; return r }(),
		"wrong application": func() Route { r := valid; r.Area.ApplicationID = "other"; return r }(),
		"wrong human":       func() Route { r := valid; r.HumanID = "other"; return r }(),
		"no human":          func() Route { r := valid; r.HumanID = ""; return r }(),
	} {
		t.Run(name, func(t *testing.T) {
			got, err := evaluate(t, request, Authority{Routes: []Route{valid, wrong}}, time.Time{})
			if err == nil || !reflect.DeepEqual(got, Result{}) {
				t.Fatalf("an answer about somebody else was used: %#v, %v", got, err)
			}
		})
	}
}

// A root grant's permissions are the application's whole active catalog,
// recomputed at resolve time — the widest legitimate route in this system is
// produced by the system itself. A per-route ceiling of 256 answered "we could
// not check your access" to every request a root holder made, including ones an
// entirely different grant authorized.
func TestARootWideRouteIsAnOrdinaryAnswer(t *testing.T) {
	wide := route([]string{"fk3x9r2m5iv8"}, Predicate{Key: "department", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"})
	wide.Permissions = make([]string, 0, 603)
	for i := 0; i < 602; i++ {
		wide.Permissions = append(wide.Permissions, fmt.Sprintf("hrms:payroll:item%d::read", i))
	}
	wide.Permissions = append(wide.Permissions, validRequest().Permission)
	got, err := evaluate(t, validRequest(), Authority{Routes: []Route{wide}}, time.Time{})
	if err != nil || got.Decision != Allow {
		t.Fatalf("a catalog-wide root route was refused: %#v, %v", got, err)
	}
}

func manyGrants(n int) []string {
	grants := make([]string, n)
	for i := range grants {
		grants[i] = string(rune(0x1000 + i))
	}
	return grants
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

func TestEvaluateUsesDirectHumanForGroupSelf(t *testing.T) {
	self := route([]string{"group-grant"}, Predicate{Key: "employee", Value: "$self", SourceGrantID: "group-grant"})
	for name, selection := range map[string]struct {
		selection Selection
		want      Decision
	}{
		"human exact": {Selection{Kind: SelectionExact, Value: "maya"}, Allow},
		"other exact": {Selection{Kind: SelectionExact, Value: "agent-17"}, Deny},
		"all humans":  {Selection{Kind: SelectionAll}, Deny},
	} {
		t.Run(name, func(t *testing.T) {
			request := validRequest()
			request.Material["employee"] = selection.selection
			got, err := evaluate(t, request, Authority{Routes: []Route{self}}, time.Time{})
			if err != nil || got.Decision != selection.want {
				t.Fatalf("Evaluate() = %#v, %v; want %q", got, err, selection.want)
			}
		})
	}
}

func TestEvaluateAcceptsExactWorkLimits(t *testing.T) {
	grants := make([]string, 256)
	for i := range grants {
		grants[i] = string(rune(0x1000 + i))
	}
	predicates := make([]Predicate, 10_000)
	for i := range predicates {
		predicates[i] = Predicate{Key: "department", Value: "FIN", SourceGrantID: grants[0]}
	}
	routes := make([]Route, 10_000)
	for i := range routes {
		routes[i] = route([]string{"G"})
	}
	material := make(Material, 10_000)
	for i := 0; i < 10_000; i++ {
		material["k"+string(rune(0x1000+i))+"x"] = Selection{Kind: SelectionAll}
	}
	for name, test := range map[string]struct {
		request   Request
		authority Authority
	}{
		"material":   {Request{Context: validRequest().Context, Permission: validRequest().Permission, Material: material}, Authority{Routes: []Route{route([]string{"G"})}}},
		"routes":     {validRequest(), Authority{Routes: routes}},
		"grants":     {validRequest(), Authority{Routes: []Route{route(grants)}}},
		"predicates": {validRequest(), Authority{Routes: []Route{route(grants[:1], predicates...)}}},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := evaluate(t, test.request, test.authority, time.Time{})
			if err != nil || got.Decision != Allow {
				t.Fatalf("Evaluate() = %#v, %v", got, err)
			}
		})
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

	// Permissions are bounded across the answer, not per route, because a root
	// route legitimately carries the whole catalog. Untested, this ceiling was
	// how a root holder got a 503 on every request.
	wide := make([]Route, 4097)
	for i := range wide {
		r := route([]string{"G"})
		r.Permissions = make([]string, 256)
		for j := range r.Permissions {
			r.Permissions[j] = fmt.Sprintf("hrms:payroll:i%d:%d::read", i, j)
		}
		wide[i] = r
	}
	if got, err := evaluate(t, validRequest(), Authority{Routes: wide}, time.Time{}); err == nil || !reflect.DeepEqual(got, Result{}) {
		t.Fatalf("accepted excess permissions: %#v, %v", got, err)
	}

	predicates := make([]Predicate, 10001)
	for i := range predicates {
		predicates[i] = Predicate{Key: "department", Value: "FIN", SourceGrantID: "G"}
	}
	if got, err := evaluate(t, validRequest(), Authority{Routes: []Route{route([]string{"G"}, predicates...)}}, time.Time{}); err == nil || !reflect.DeepEqual(got, Result{}) {
		t.Fatalf("accepted excess predicates: %#v, %v", got, err)
	}
}

// A route carrying some other permission is not a malformed answer. The gate
// asks what this human holds, so the answer names every grant they hold here —
// including the ones for permissions this request is not about. Those are
// ordinary, and they simply do not match. Before the question stopped carrying
// a permission, one of them meant the source had answered the wrong question
// and the request failed with 503 instead of being denied.
func TestARouteForAnotherPermissionIsANonMatchAndNotAnError(t *testing.T) {
	other := route([]string{"fk3x9r2m5iv8"}, Predicate{Key: "department", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"})
	other.Permissions = []string{"certificate::write"}
	got, err := evaluate(t, validRequest(), Authority{Routes: []Route{other}}, time.Time{})
	if err != nil {
		t.Fatalf("unrelated route reported as an unusable answer: %v", err)
	}
	if got.Decision != Deny {
		t.Fatalf("Evaluate() = %#v", got)
	}
}

// And the request's permission is picked out of a route that carries several,
// which is the ordinary shape now: one grant, every permission it names.
func TestARouteAllowsOnAnyPermissionItCarries(t *testing.T) {
	many := route([]string{"fk3x9r2m5iv8"}, Predicate{Key: "department", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"})
	many.Permissions = []string{"certificate::write", validRequest().Permission, "certificate::delete"}
	got, err := evaluate(t, validRequest(), Authority{Routes: []Route{many}}, time.Time{})
	if err != nil || got.Decision != Allow {
		t.Fatalf("Evaluate() = %#v, %v", got, err)
	}
}
