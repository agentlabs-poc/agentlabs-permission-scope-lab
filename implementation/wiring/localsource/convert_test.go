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

// The two AuthoritySource implementations answer one contract, and this one was
// the trusting side: it read the answer's area, subject and versions without
// checking any of them, and stamped the asked-for permission onto every route.
//
// Nothing in-process can produce these — ResolveAuthority answers about the area
// and human it was asked about, and filters on the permissions — so this is
// defence in depth. That is also the reason it is worth having: the day this
// source is given a different implementation or the facade grows a cache, the
// assumptions are written down rather than remembered. The HTTP source carries
// each of these checks already, and its comment on the permission one records
// what it cost to learn: "a grant for reading the directory came back approved
// for reading payroll."
func TestTheAnswerIsCorroboratedAgainstTheQuestion(t *testing.T) {
	const read, write = "hrms:payroll:payslip::read", "hrms:payroll:payslip::write"
	const maya = "fi7io4lvjqio"
	query := authmiddleware.AuthorityQuery{
		Context: authmiddleware.RequestContext{
			Area:     authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"},
			Identity: authmiddleware.Identity{HumanID: maya},
		},
		Permission: read,
	}
	grant := func(id string, permissions ...string) domain.ResolvedGrant {
		return domain.ResolvedGrant{
			Version: "1", GrantID: id, Revision: 1,
			Permissions: permissions, Scope: map[string]string{"dept": "FIN"},
		}
	}
	sound := func() domain.ResolvedAuthority {
		return domain.ResolvedAuthority{
			Version: "1", TenantID: "acme", ApplicationID: "hrms", HumanID: maya,
			ResolvedGrants: []domain.ResolvedGrant{grant("fk3x9r2m5iv8", read, write)},
		}
	}

	// The control. Without it every refusal below could be a fixture that was
	// never usable, which is how a set of negative cases quietly proves nothing.
	authority, err := routesFor(sound(), query)
	if err != nil || len(authority.Routes) != 1 || authority.Routes[0].Permission != read {
		t.Fatalf("a sound answer gave %#v, %v", authority, err)
	}

	for name, bend := range map[string]func(*domain.ResolvedAuthority){
		"another contract version":     func(a *domain.ResolvedAuthority) { a.Version = "9" },
		"another tenant":               func(a *domain.ResolvedAuthority) { a.TenantID = "globex" },
		"another application":          func(a *domain.ResolvedAuthority) { a.ApplicationID = "crm" },
		"another human":                func(a *domain.ResolvedAuthority) { a.HumanID = "fi7io4lvjwu8" },
		"a grant from another version": func(a *domain.ResolvedAuthority) { a.ResolvedGrants[0].Version = "9" },
		"a grant stating no version":   func(a *domain.ResolvedAuthority) { a.ResolvedGrants[0].Version = "" },
		"a grant for another permission": func(a *domain.ResolvedAuthority) {
			a.ResolvedGrants[0].Permissions = []string{write}
		},
		"a grant carrying none at all": func(a *domain.ResolvedAuthority) { a.ResolvedGrants[0].Permissions = nil },
		// The second of several, which every earlier fixture here made
		// unreachable by carrying exactly one grant.
		"a second grant that does not carry it": func(a *domain.ResolvedAuthority) {
			a.ResolvedGrants = append(a.ResolvedGrants, grant("fk3x9r2man0d", write))
		},
		"a third grant from another version": func(a *domain.ResolvedAuthority) {
			a.ResolvedGrants = append(a.ResolvedGrants, grant("fk3x9r2man0d", read))
			a.ResolvedGrants = append(a.ResolvedGrants, grant("fk3x9r2mv5k0", read))
			a.ResolvedGrants[2].Version = "9"
		},
	} {
		t.Run(name, func(t *testing.T) {
			answer := sound()
			bend(&answer)
			authority, err := routesFor(answer, query)
			if err == nil {
				t.Fatalf("accepted, giving %#v", authority.Routes)
			}
			if len(authority.Routes) != 0 {
				t.Fatalf("a refused answer still carried routes: %#v", authority.Routes)
			}
			// An evaluation failure, so the application answers 503 rather than
			// telling the person they lack access for an answer they never saw.
			var evaluation *authmiddleware.EvaluationError
			if !errors.As(err, &evaluation) {
				t.Fatalf("err = %T %v, want an EvaluationError", err, err)
			}
			if evaluation.Message != "We could not check your access." {
				t.Fatalf("message = %q", evaluation.Message)
			}
		})
	}

	// And several sound grants all become routes, so the loop is not simply
	// refusing anything plural.
	many := sound()
	many.ResolvedGrants = append(many.ResolvedGrants, grant("fk3x9r2man0d", read), grant("fk3x9r2mv5k0", read, write))
	if authority, err := routesFor(many, query); err != nil || len(authority.Routes) != 3 {
		t.Fatalf("three sound grants gave %#v, %v", authority, err)
	}
}
