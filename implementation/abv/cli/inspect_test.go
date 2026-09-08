package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"testing"
)

type canonicalAPI struct{ apiSpy }

func (*canonicalAPI) Inspect(_ context.Context, area domain.Area, kind, id string) (domain.Record, error) {
	return domain.Record{Area: area, Kind: kind, ID: id, CanonicalJSON: []byte(`{"version":"1","id":"A1"}`)}, nil
}

func TestInspectCanonicalJSONIsAloneOnStdout(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	var out bytes.Buffer
	if err := inspect(context.Background(), &canonicalAPI{}, area, "assignment", "A1", &out); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "{\"version\":\"1\",\"id\":\"A1\"}\n" {
		t.Fatalf("output = %q", got)
	}
}
