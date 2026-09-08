package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"strings"
	"testing"
)

type checkAPI struct{ apiSpy }

func (a *checkAPI) CheckAssignment(_ context.Context, _ domain.Area, _ []byte) (domain.Diagnostic, error) {
	return domain.Diagnostic{Summary: "valid proposal", Route: &domain.Route{
		GrantID: "G2", Permissions: []string{"read"},
		Predicates:    []domain.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "G1"}},
		AssignmentIDs: []string{"A0", "A1"},
	}}, nil
}

func TestCheckPrintsLabeledReadOnlyRouteWithoutSerializingIt(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	var out bytes.Buffer
	if err := check(context.Background(), &checkAPI{}, area, []byte("proposal"), &out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, value := range []string{"READ-ONLY DIAGNOSIS", "does not authorize a later write", "internal projection: proposed route", "G2", "dept", "FIN", "G1", "A0,A1"} {
		if !strings.Contains(text, value) {
			t.Fatalf("missing %q in %q", value, text)
		}
	}
	if strings.HasPrefix(strings.TrimSpace(text), "{") {
		t.Fatalf("route serialized as JSON: %q", text)
	}
}
