package domain

import (
	"errors"
	"testing"
)

// Removing either boundary check would accept an unqualified authority operation.
func TestAreaRequiresBothBoundaries(t *testing.T) {
	for _, ids := range [][2]string{{"", "hrms"}, {"acme", ""}, {"", ""}, {" ", "hrms"}, {"acme", "\t"}, {"*", "hrms"}, {"acme", "*"}} {
		if _, err := NewArea(ids[0], ids[1]); !errors.Is(err, ErrMalformed) {
			t.Fatalf("accepted missing/wildcard boundary %q: %v", ids, err)
		}
	}
	if err := (Area{}).Validate(); !errors.Is(err, ErrMalformed) {
		t.Fatal("accepted zero Area")
	}
}
func TestAreaPreservesExactIsolationKeys(t *testing.T) {
	keys := map[Area]bool{}
	for _, ids := range [][2]string{{"acme", "hrms"}, {"acme", "accounting"}, {"other", "hrms"}, {"Acme", "hrms"}, {" acme", "hrms"}, {"租户", "人事"}} {
		a, err := NewArea(ids[0], ids[1])
		if err != nil {
			t.Fatal(err)
		}
		if a.TenantID() != ids[0] || a.ApplicationID() != ids[1] {
			t.Fatal("normalized opaque context")
		}
		keys[a] = true
	}
	if len(keys) != 6 {
		t.Fatal("different outer boundaries collapsed")
	}
}
