package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/lab"
	"testing"
	"time"
)

// A chain may state one key twice with different values, because Narrow appends
// each child's predicates to its parent's and nothing forbids restating a key.
// Every predicate must hold against one material value, so such a route
// authorizes nothing.
//
// Folding it into a scope object without noticing would turn an unmatchable
// route into a matchable one — silently, and in the direction that grants access.
// So the route is dropped.
func TestContradictoryNarrowingIsDroppedRatherThanFolded(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	now := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	f := lab.TeamFINC17(area)
	maya := f.Issuer

	before, err := lineage.ResolveAuthority(t.Context(), f.Snapshot, maya, domain.ResolveOptions{}, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.ResolvedGrants) == 0 {
		t.Fatal("fixture resolves nothing to start with")
	}
	held := before.ResolvedGrants[0]
	if held.Scope["dept"] != "FIN" {
		t.Fatalf("expected the FIN route first: %#v", held.Scope)
	}

	// Narrow the same key to a different value further down the chain. Team1's
	// grant says dept=FIN; make its own parent say dept=ENG.
	key := domain.GrantKey{ID: "fk3x9r2m0dq3", Revision: 1}
	root := f.Snapshot.Contents[key]
	root.Permissions = []string{lab.PayslipRead, lab.PayslipWrite, lab.PayslipDelete}
	root.Scope = map[string]string{"dept": "ENG"}
	f.Snapshot.Contents[key] = root

	after, err := lineage.ResolveAuthority(t.Context(), f.Snapshot, maya, domain.ResolveOptions{}, now)
	if err != nil {
		t.Fatalf("a contradictory chain failed instead of resolving to nothing: %v", err)
	}
	for _, grant := range after.ResolvedGrants {
		if grant.GrantID == held.GrantID {
			t.Fatalf("an unmatchable route was shipped with scope %#v", grant.Scope)
		}
	}
}

// Agreement is not contradiction: a child restating its parent's value narrows
// to the same place, and the route survives with one entry for the key.
func TestRestatingTheSameValueIsNotAContradiction(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	now := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	f := lab.TeamFINC17(area)

	key := domain.GrantKey{ID: "fk3x9r2m0dq3", Revision: 1}
	root := f.Snapshot.Contents[key]
	root.Permissions = []string{lab.PayslipRead, lab.PayslipWrite, lab.PayslipDelete}
	root.Scope = map[string]string{"dept": "FIN"}
	f.Snapshot.Contents[key] = root

	resolved, err := lineage.ResolveAuthority(t.Context(), f.Snapshot, f.Issuer, domain.ResolveOptions{}, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved.ResolvedGrants) == 0 {
		t.Fatal("restating the same value dropped the route")
	}
	if got := resolved.ResolvedGrants[0].Scope["dept"]; got != "FIN" {
		t.Fatalf("scope = %q, want FIN stated twice folding to once", got)
	}
}
