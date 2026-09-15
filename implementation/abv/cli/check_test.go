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
		GrantID: "fk3x9r2man0d", Permissions: []string{"read"},
		Predicates:    []domain.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"}},
		AssignmentIDs: []string{"fm5b7t4p0dq3", "fm5b7t4p5iv8"},
	}}, nil
}

func TestCheckPrintsLabeledReadOnlyRouteWithoutSerializingIt(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	var out bytes.Buffer
	if err := check(context.Background(), &checkAPI{}, area, []byte("proposal"), &out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, value := range []string{"READ-ONLY DIAGNOSIS", "does not authorize a later write", "internal projection: proposed route", "fk3x9r2man0d", "dept", "FIN", "fk3x9r2m5iv8", "fm5b7t4p0dq3,fm5b7t4p5iv8"} {
		if !strings.Contains(text, value) {
			t.Fatalf("missing %q in %q", value, text)
		}
	}
	if strings.HasPrefix(strings.TrimSpace(text), "{") {
		t.Fatalf("route serialized as JSON: %q", text)
	}
}
