package localsource

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"agentlabs.local/authmiddleware"
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC) }

// The route describes where it came from, not where it was asked for. They
// cannot disagree — the area asked for is the area read — but the conversion
// takes the answer's echo rather than copying the query back at itself.
func TestConvertGrantUsesTheResolvedBoundaries(t *testing.T) {
	resolved := domain.ResolvedAuthority{
		Version: "1", TenantID: "resolved-tenant", ApplicationID: "resolved-app", HumanID: "fi7io4lvjqio",
	}
	grant := domain.ResolvedGrant{GrantID: "fk3x9r2m0dq3", Scope: map[string]string{"dept": "FIN"}}
	query := authmiddleware.AuthorityQuery{
		Context:    authmiddleware.RequestContext{Area: authmiddleware.Area{TenantID: "query-tenant", ApplicationID: "query-app"}},
		Permission: "hrms:payroll:payslip::read",
	}
	got := convertGrant(resolved, grant, query)
	if got.Area != (authmiddleware.Area{TenantID: "resolved-tenant", ApplicationID: "resolved-app"}) {
		t.Fatalf("area=%+v", got.Area)
	}
	if got.HumanID != "fi7io4lvjqio" || got.Permission != query.Permission {
		t.Fatalf("route=%+v", got)
	}
	if len(got.Predicates) != 1 || got.Predicates[0] != (authmiddleware.Predicate{Key: "dept", Value: "FIN", SourceGrantID: "fk3x9r2m0dq3"}) {
		t.Fatalf("predicates=%+v", got.Predicates)
	}
}

// Without the explanation there is one grant to name: the one that reaches the
// human. The mapping from contributing assignments to their grants used to live
// here and is Auth-AL's now, which is why the old test for it is gone rather
// than rewritten.
func TestGrantChainNamesTheReachingGrantWithoutSource(t *testing.T) {
	bare := domain.ResolvedGrant{GrantID: "fk3x9r2man0d"}
	if got := grantChain(bare); len(got) != 1 || got[0] != "fk3x9r2man0d" {
		t.Fatalf("chain=%v", got)
	}
	explained := domain.ResolvedGrant{GrantID: "fk3x9r2man0d", Source: &domain.Source{
		Lineage: []domain.LineageStep{{GrantID: "fk3x9r2m0dq3"}, {GrantID: "fk3x9r2man0d"}},
	}}
	if got := grantChain(explained); !reflect.DeepEqual(got, []string{"fk3x9r2m0dq3", "fk3x9r2man0d"}) {
		t.Fatalf("chain=%v", got)
	}
}

// An AuthoritySource owes its caller an evaluation failure, not a bare error.
//
// The application tells an outage from a denial by type: an EvaluationError
// answers 503, anything else falls through to the caller's-fault branch. A
// source that returns the store's own error is therefore reported as the
// caller's malformed request while the store is down — no availability signal,
// and the blame in the wrong place.
//
// This had no test: returning `err` unwrapped left every suite green.
func TestAFailureIsReportedAsAnEvaluationFailure(t *testing.T) {
	// A store that opens and then cannot answer. Every read asks the registry
	// first, so a registry that fails makes the read fail where it matters —
	// after Open, inside Load. Pointing at a missing file instead would fail at
	// Open and never reach the line under test, which is how the first version
	// of this test passed while proving nothing.
	path := filepath.Join(t.TempDir(), "authority.db")
	area, _ := domain.NewArea("acme", "hrms")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	source, err := Open(context.Background(), path,
		domain.Actor{Type: "service_account", ID: "agent_hrms"}, refusingAdmin{}, fixedClock{}, brokenRegistry{})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()

	_, loadErr := source.Load(context.Background(), authmiddleware.AuthorityQuery{
		Context: authmiddleware.RequestContext{
			Area:     authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"},
			Identity: authmiddleware.Identity{Version: "1", HumanID: "fi7io4lvjqio"},
		},
		Permission: "hrms:payroll:payslip::read",
	})
	if loadErr == nil {
		t.Fatal("a missing store answered")
	}
	var evaluation *authmiddleware.EvaluationError
	if !errors.As(loadErr, &evaluation) {
		t.Fatalf("err = %v (%T), want an EvaluationError the application can recognise", loadErr, loadErr)
	}
	if evaluation.Message == "" || evaluation.MessageReason == "" {
		t.Fatalf("an evaluation failure must carry both messages: %#v", evaluation)
	}
}

type refusingAdmin struct{}

func (refusingAdmin) CheckAssignment(context.Context, abv.Evidence, domain.Identity, domain.Assignment, time.Time) error {
	return domain.ErrRejected
}

// brokenRegistry is the store's own dependency failing, which is what an outage
// looks like from inside a read.
type brokenRegistry struct{}

func (brokenRegistry) ApplicationExists(context.Context, string) (bool, error) { return true, nil }
func (brokenRegistry) Installed(context.Context, string, string) (bool, error) {
	return false, errors.New("registry unavailable")
}

// The two AuthoritySource implementations answer the same contract, and one of
// them was enforcing an invariant the other trusted.
//
// convertGrant stamps the asked-for permission onto every route, so a grant that
// does not carry it becomes a route claiming it and the evaluator never knows —
// which is the bug authclient's own decode already carries a comment about: "a
// grant for reading the directory came back approved for reading payroll."
// ResolveAuthority does filter, so nothing can produce this through the real
// path today. That is the point: the trusting side is the one nothing tested,
// and drift starts where one side assumes.
func TestAGrantWithoutTheAskedPermissionDoesNotBecomeARoute(t *testing.T) {
	const read, write = "hrms:payroll:payslip::read", "hrms:payroll:payslip::write"
	query := authmiddleware.AuthorityQuery{
		Context:    authmiddleware.RequestContext{Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}},
		Permission: read,
	}
	resolved := func(permissions ...string) domain.ResolvedAuthority {
		return domain.ResolvedAuthority{
			Version: "1", TenantID: "acme", ApplicationID: "hrms", HumanID: "fi7io4lvjqio",
			ResolvedGrants: []domain.ResolvedGrant{{
				GrantID: "fk3x9r2m5iv8", Permissions: permissions, Scope: map[string]string{"dept": "FIN"},
			}},
		}
	}

	// The control: a grant that does carry it becomes exactly one route.
	authority, err := routesFor(resolved(read, write), query)
	if err != nil || len(authority.Routes) != 1 || authority.Routes[0].Permission != read {
		t.Fatalf("a grant carrying the permission gave %#v, %v", authority, err)
	}

	for name, permissions := range map[string][]string{
		"a different permission": {write},
		"none at all":            nil,
		"a near miss":            {read + "x"},
	} {
		t.Run(name, func(t *testing.T) {
			authority, err := routesFor(resolved(permissions...), query)
			if err == nil {
				t.Fatalf("a grant carrying %v became %#v", permissions, authority.Routes)
			}
			if len(authority.Routes) != 0 {
				t.Fatalf("a refused answer still carried routes: %#v", authority.Routes)
			}
			// And it is an evaluation failure, so the application answers 503
			// rather than telling the person they lack access.
			var evaluation *authmiddleware.EvaluationError
			if !errors.As(err, &evaluation) {
				t.Fatalf("err = %T %v, want an EvaluationError", err, err)
			}
		})
	}
}
