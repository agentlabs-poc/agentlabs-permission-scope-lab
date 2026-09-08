package abv

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"errors"
	"testing"
)

func TestInspectGrantControlRejectsMissingAndMalformedStoredControls(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	for _, tc := range []struct {
		name     string
		controls map[string]domain.GrantControl
		want     error
	}{
		{"missing", map[string]domain.GrantControl{}, domain.ErrNotFound},
		{"identity mismatch", map[string]domain.GrantControl{"G2": {Version: "1", ID: "other", Status: "enabled"}}, domain.ErrMalformed},
		{"version", map[string]domain.GrantControl{"G2": {Version: "2", ID: "G2", Status: "enabled"}}, domain.ErrMalformed},
		{"status", map[string]domain.GrantControl{"G2": {Version: "1", ID: "G2", Status: "paused"}}, domain.ErrMalformed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			record := domain.Record{Area: area, Kind: "grant-control", ID: "G2"}
			err := inspectSnapshot(storage.Snapshot{Area: area, Controls: tc.controls}, &record)
			if !errors.Is(err, tc.want) || len(record.CanonicalJSON) != 0 {
				t.Fatalf("error=%v json=%s", err, record.CanonicalJSON)
			}
		})
	}
}
