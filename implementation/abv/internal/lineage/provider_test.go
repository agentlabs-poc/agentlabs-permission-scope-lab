package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestTeamFINC17RoundTripsAndTenantCannotSupplyMissingSupport(t *testing.T) {
	acme, _ := domain.NewArea("acme", "hrms")
	globex, _ := domain.NewArea("globex", "hrms")
	accounting, _ := domain.NewArea("acme", "accounting")
	missing := lab.TeamFINC17(acme)
	delete(missing.Snapshot.Assignments, "A1")
	complete := lab.TeamFINC17(globex)
	otherApplication := lab.TeamFINC17(accounting)
	provider, err := lab.CreateSQLite(context.Background(), filepath.Join(t.TempDir(), "lineage.db"), []storage.Snapshot{missing.Snapshot, complete.Snapshot, otherApplication.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = provider.Close() })

	err = provider.Read(context.Background(), acme, func(snapshot storage.Snapshot) error {
		_, resolveErr := lineage.ResolveParentTeam(snapshot, missing.Child, "Team2", time.Time{})
		return resolveErr
	})
	if !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("other tenant supplied absent A1: %v", err)
	}
	err = provider.Read(context.Background(), globex, func(snapshot storage.Snapshot) error {
		got, resolveErr := lineage.ResolveParentTeam(snapshot, complete.Child, "Team2", time.Time{})
		if resolveErr == nil && len(got.AssignmentIDs) != 2 {
			t.Fatalf("fixture route did not roundtrip: %#v", got)
		}
		return resolveErr
	})
	if err != nil {
		t.Fatalf("complete tenant fixture did not resolve: %v", err)
	}
	err = provider.Read(context.Background(), accounting, func(snapshot storage.Snapshot) error {
		_, resolveErr := lineage.ResolveParentTeam(snapshot, otherApplication.Child, "Team2", time.Time{})
		return resolveErr
	})
	if err != nil {
		t.Fatalf("complete other-application fixture did not resolve independently: %v", err)
	}
}
