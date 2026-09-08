package domain

import (
	"errors"
	"testing"
)

func TestApplicationRequiresExactValidIdentifier(t *testing.T) {
	for _, id := range []string{"", " ", "hr*ms", string([]byte{0xff})} {
		if _, err := NewApplication(id); !errors.Is(err, ErrMalformed) {
			t.Fatalf("accepted invalid application %q: %v", id, err)
		}
	}
	for _, id := range []string{"hrms", " HRMS ", "人事"} {
		a, err := NewApplication(id)
		if err != nil || a.ID() != id || a.Validate() != nil {
			t.Fatalf("application normalized or rejected: %#v %v", a, err)
		}
	}
	if err := (Application{}).Validate(); !errors.Is(err, ErrMalformed) {
		t.Fatal("zero application accepted")
	}
}
