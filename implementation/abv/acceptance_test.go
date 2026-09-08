package abv_test

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/sqlite"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"
)

func TestEmptyChildScopeRetainsParentPredicateAndPersists(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	key := domain.GrantKey{ID: "G2", Revision: 1}
	child := fixture.Snapshot.Contents[key]
	child.Scope = map[string]string{}
	fixture.Snapshot.Contents[key], fixture.Child = child, child
	path := t.TempDir() + "/authority.db"
	provider, err := lab.CreateSQLite(t.Context(), path, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	admin, _ := lab.NewAdministration(area, fixture.Administration)
	facade, err := abv.New(provider, admin, clock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(fixture.Proposed)
	diagnostic, err := facade.CheckAssignment(t.Context(), area, raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "G1"}}
	if diagnostic.Route == nil || !slices.Equal(diagnostic.Route.Predicates, want) {
		t.Fatalf("empty child scope route predicates = %#v, want %#v", diagnostic.Route, want)
	}
	if _, err = facade.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed); err != nil {
		t.Fatal(err)
	}
	if err = facade.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := abv.OpenSQLite(t.Context(), path, admin, clock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if record, err := reopened.Inspect(t.Context(), area, "assignment", fixture.Proposed.ID); err != nil || record.ID != fixture.Proposed.ID {
		t.Fatalf("reopened assignment=%#v err=%v", record, err)
	}
}

func TestSnapshotLimitReturnsNoReceiptAndNoWrite(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	path := t.TempDir() + "/authority.db"
	provider, err := lab.CreateSQLite(t.Context(), path, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	provider.Close()
	provider, err = sqlite.OpenWithOptions(t.Context(), path, sqlite.Options{MaxSnapshotRecords: 30})
	if err != nil {
		t.Fatal(err)
	}
	admin, _ := lab.NewAdministration(area, fixture.Administration)
	facade, _ := abv.New(provider, admin, clock{now: time.Now()})
	receipt, err := facade.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed)
	if !errors.Is(err, storage.ErrSnapshotLimit) || receipt != (domain.Receipt{}) {
		t.Fatalf("bounded snapshot receipt=%#v err=%v", receipt, err)
	}
	facade.Close()
	reopened, err := abv.OpenSQLite(t.Context(), path, admin, clock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if _, err := reopened.Inspect(t.Context(), area, "assignment", fixture.Proposed.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bounded snapshot failure wrote assignment: %v", err)
	}
}

func BenchmarkBoundedAuthorityDiagnosis(b *testing.B) {
	for _, records := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("records=%d", records), func(b *testing.B) {
			area, _ := domain.NewArea("tenant-fin", "hrms")
			fixture := benchmarkSnapshot(area, records)
			if got := snapshotRecordCount(fixture.Snapshot); got != records {
				b.Fatalf("loader record count = %d, want %d", got, records)
			}
			path := b.TempDir() + "/authority.db"
			provider, err := lab.CreateSQLite(context.Background(), path, []storage.Snapshot{fixture.Snapshot})
			if err != nil {
				b.Fatal(err)
			}
			provider.Close()
			provider, err = sqlite.OpenWithOptions(context.Background(), path, sqlite.Options{MaxSnapshotRecords: records})
			if err != nil {
				b.Fatal(err)
			}
			admin, _ := lab.NewAdministration(area, fixture.Administration)
			facade, _ := abv.New(provider, admin, clock{now: time.Now()})
			defer facade.Close()
			raw, _ := json.Marshal(fixture.Proposed)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := facade.CheckAssignment(context.Background(), area, raw); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func benchmarkSnapshot(area domain.Area, records int) lab.TeamFINC17Case {
	fixture := lab.TeamFINC17(area)
	// Loader count: application + catalog rows + every area-owned map/slice row.
	const baseRecords = 32
	fixture.Snapshot.Roles[domain.RoleKey{ID: "independent-role", Revision: 1}] = domain.RoleContent{ID: "independent-role", Revision: 1, Permissions: []string{lab.PayslipRead}}
	for i := 0; i < (records-baseRecords)/2; i++ {
		id := fmt.Sprintf("independent-%05d", i)
		fixture.Snapshot.Controls[id] = domain.GrantControl{Version: "1", ID: id, Status: "enabled"}
		fixture.Snapshot.Contents[domain.GrantKey{ID: id, Revision: 1}] = domain.GrantContent{Version: "1", GrantID: id, Revision: 1, Permissions: []string{lab.PayslipRead}, Scope: map[string]string{}}
	}
	return fixture
}

func snapshotRecordCount(snapshot storage.Snapshot) int {
	supported := 0
	for _, keys := range snapshot.Catalog.SupportedKeys {
		supported += len(keys)
	}
	return 1 + len(snapshot.Catalog.Permissions) + len(snapshot.Catalog.Scopes) + supported + len(snapshot.Controls) + len(snapshot.Contents) + len(snapshot.Assignments) + len(snapshot.Roles) + len(snapshot.Teams) + len(snapshot.Memberships) + len(snapshot.TrustedRoots)
}
