package localsource

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/authmiddleware"
	"reflect"
	"testing"
)

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
