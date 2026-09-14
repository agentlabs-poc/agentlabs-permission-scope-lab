//go:build !race

// Timing assertions are meaningless under the race detector, which is roughly
// fifteen times slower. This file measures; -race proves correctness elsewhere.

package abv_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

// DeleteGrant must refuse while a child depends on the grant, which means it has
// to find the children. The parent is in the value rather than a key slot, and
// the envelope's index does not cover it — so the question is what that costs.
//
// It costs nothing extra, and the reason is structural: every write already
// loads the whole area snapshot inside its transaction before the callback runs.
// The child search is a pass over a map that is already in memory, so it is
// bounded by the same limit that bounds every other write, not by an index.
//
// This measures the whole operation at a realistic depth to keep that claim
// honest, and to catch the day it stops being true.
func TestDeleteGrantStaysBoundedAsChildrenGrow(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "scale.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := lab.Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeConnection() })
	writer := api.(grantWriter)

	const children = 400
	leaves := make([]string, 0, children)
	start := time.Now()
	for i := 0; i < children; i++ {
		grant, _, err := writer.CreateGrant(t.Context(), area, teamFixture, "G1", domain.GrantContent{
			Permissions: []string{payslipRead},
			Scope:       map[string]string{"cert": fmt.Sprintf("C%04d", i)},
		})
		if err != nil {
			t.Fatalf("child %d: %v", i, err)
		}
		leaves = append(leaves, grant.ID)
	}
	t.Logf("created %d children in %s", children, time.Since(start).Round(time.Millisecond))

	// The refusal is the expensive direction: it has to look at every revision
	// in the area before it can say no.
	refuse := time.Now()
	if err := writer.DeleteGrant(t.Context(), area, teamFixture, "G1"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("delete with %d children gave %v, want ErrConflict", children, err)
	}
	refused := time.Since(refuse)

	// And the accepting direction, on a leaf, with the same population present.
	accept := time.Now()
	if err := writer.DeleteGrant(t.Context(), area, teamFixture, leaves[len(leaves)/2]); err != nil {
		t.Fatalf("deleting a leaf: %v", err)
	}
	accepted := time.Since(accept)

	t.Logf("refuse %s · accept %s · at %d grants in the area", refused.Round(time.Microsecond), accepted.Round(time.Microsecond), children+3)

	// A generous ceiling: this is not a benchmark, it is a tripwire for the day
	// the child search stops being a pass over an already-loaded snapshot and
	// becomes a query per child.
	const budget = 400 * time.Millisecond
	if refused > budget || accepted > budget {
		t.Fatalf("refuse=%s accept=%s, both should stay under %s", refused, accepted, budget)
	}
}
