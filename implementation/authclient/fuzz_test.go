package authclient

import (
	"agentlabs.local/authmiddleware"
	"encoding/json"
	"testing"
)

func query() authmiddleware.AuthorityQuery {
	return authmiddleware.AuthorityQuery{
		Context: authmiddleware.RequestContext{
			Area:     authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"},
			Identity: authmiddleware.Identity{Version: "1", HumanID: "fi7io4lvjqio"},
		},
		Permission: "hrms:payroll:payslip::read",
	}
}

// The gate allows on whatever this decoder returns, and what it decodes arrives
// from another process. authmiddleware fuzzes its own two wire contracts for
// exactly this reason; this is the third, and it had none.
//
// The invariant is not "never fail" — it is that a failure returns NOTHING. A
// decoder that returned a partial Authority beside an error would hand the gate
// routes nobody granted.
func FuzzDecodeNeverReturnsRoutesWithAnError(f *testing.F) {
	for _, seed := range []string{
		`{"version":"1","tenant_id":"acme","application_id":"hrms","human_id":"fi7io4lvjqio","resolved_grants":[]}`,
		`{"version":"1","tenant_id":"acme","application_id":"hrms","human_id":"fi7io4lvjqio","resolved_grants":[{"version":"1","grant_id":"g","revision":1,"permissions":["p"],"scope":{"dept":"FIN"}}]}`,
		`{"version":"2","tenant_id":"acme","application_id":"hrms","human_id":"fi7io4lvjqio","resolved_grants":[]}`,
		`{"version":"1","tenant_id":"other","application_id":"hrms","human_id":"fi7io4lvjqio","resolved_grants":[]}`,
		`{"version":"1","tenant_id":"acme","application_id":"hrms","human_id":"someone-else","resolved_grants":[]}`,
		`{}`, `null`, `[]`, `{"version":"1"}`,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		var answer resolveResponse
		if json.Unmarshal(raw, &answer) != nil {
			return
		}
		got, err := answer.decode(query())
		if err != nil && len(got.Routes) != 0 {
			t.Fatalf("a failed decode returned %d routes", len(got.Routes))
		}
		if err != nil {
			return
		}
		// A successful decode must have been answered about what was asked, and
		// every route must carry the area and human it was asked about — the
		// gate trusts these without rechecking them.
		q := query()
		for _, route := range got.Routes {
			if route.Area.TenantID != q.Context.Area.TenantID || route.Area.ApplicationID != q.Context.Area.ApplicationID {
				t.Fatalf("accepted a route for another area: %#v", route.Area)
			}
			if route.HumanID != q.Context.Identity.HumanID {
				t.Fatalf("accepted a route for another human: %q", route.HumanID)
			}
			if route.Permission != q.Permission {
				t.Fatalf("accepted a route for another permission: %q", route.Permission)
			}
			// grant_ids is the evidence an allow must carry.
			if len(route.GrantIDs) == 0 {
				t.Fatalf("accepted a route with no authorizing grant: %#v", route)
			}
		}
	})
}
