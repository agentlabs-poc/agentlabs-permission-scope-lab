package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestAssignDisplaysFixtureIdentityLimit(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	var out, diag bytes.Buffer
	if err := assign(context.Background(), &apiSpy{}, area, "maya-team1", []byte("proposal"), &out, &diag); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(diag.String(), "not authenticated identity") || out.String() != "assignment A2 created\n" {
		t.Fatalf("stdout=%q stderr=%q", out.String(), diag.String())
	}
}
