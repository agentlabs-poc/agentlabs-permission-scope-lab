package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/contracttest"
	"context"
	"database/sql"
	"errors"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func TestProviderContract(t *testing.T) {
	contracttest.Run(t, contracttest.Factory{
		Open:   Open,
		Create: CreateFixture,
		OpenWithLimit: func(ctx context.Context, path string, limit int) (storage.Provider, error) {
			return OpenWithOptions(ctx, path, Options{MaxSnapshotRecords: limit})
		},
	})
}

func TestOpenRejectsUnknownExistingDatabaseWithoutMigration(t *testing.T) {
	path := t.TempDir() + "/unknown.db"
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`CREATE TABLE foreign_data(value TEXT); INSERT INTO foreign_data VALUES ('keep')`); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = Open(t.Context(), path); !errors.Is(err, storage.ErrNotABVDatabase) {
		t.Fatalf("want non-ABV error, got %v", err)
	}
	db, _ = sql.Open("sqlite", path)
	defer db.Close()
	var value string
	if err := db.QueryRow(`SELECT value FROM foreign_data`).Scan(&value); err != nil || value != "keep" {
		t.Fatalf("foreign database changed: %q %v", value, err)
	}
}

func TestFixtureRefusesExistingPath(t *testing.T) {
	path := t.TempDir() + "/existing.db"
	if err := os.WriteFile(path, []byte("do not overwrite"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateFixture(t.Context(), path, nil); err == nil {
		t.Fatal("fixture overwrote existing path")
	}
	got, _ := os.ReadFile(path)
	if string(got) != "do not overwrite" {
		t.Fatal("existing file changed")
	}
}

func TestEveryAcquiredConnectionHasForeignKeysAndFileUsesWAL(t *testing.T) {
	path := t.TempDir() + "/authority.db"
	p, err := Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	actual := p.(*provider)
	actual.db.SetMaxOpenConns(4)
	conns := make([]*sql.Conn, 0, 4)
	for range 4 {
		c, err := actual.connection(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, c)
		var fk int
		if err := c.QueryRowContext(t.Context(), `PRAGMA foreign_keys`).Scan(&fk); err != nil || fk != 1 {
			t.Fatalf("foreign_keys=%d err=%v", fk, err)
		}
		if _, err := c.ExecContext(t.Context(), `INSERT INTO installations(tenant_id,application_id) VALUES ('escape','missing-application')`); err == nil {
			t.Fatal("connection accepted a foreign-key violation")
		}
	}
	for _, c := range conns {
		c.Close()
	}
	var mode string
	if err := actual.db.QueryRowContext(t.Context(), `PRAGMA journal_mode`).Scan(&mode); err != nil || mode != "wal" {
		t.Fatalf("journal=%q err=%v", mode, err)
	}
}

func TestFixtureRejectsConflictingCopiesOfSharedCatalog(t *testing.T) {
	first := contractFixture(t)
	area, err := domain.NewArea("other", first.Area.ApplicationID())
	if err != nil {
		t.Fatal(err)
	}
	second := minimalFixture(area)
	second.Catalog.CompatibilityEnabled = !first.Catalog.CompatibilityEnabled
	if _, err := CreateFixture(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{first, second}); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("want malformed conflicting catalog, got %v", err)
	}
}

func TestReadRejectsCanonicalPayloadAndIndexDisagreement(t *testing.T) {
	path := t.TempDir() + "/authority.db"
	base := contractFixture(t)
	opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	p := opened.(*provider)
	defer p.Close()
	bad := []byte(`{"version":"1","grant_id":"different","revision":1,"permissions":["read"],"scope":{}}`)
	if _, err := p.db.ExecContext(t.Context(), `UPDATE grant_contents SET canonical_json=? WHERE tenant_id=? AND application_id=? AND grant_id='G1'`, bad, base.Area.TenantID(), base.Area.ApplicationID()); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	err = p.Read(t.Context(), base.Area, func(storage.Snapshot) error { calls.Add(1); return nil })
	if !errors.Is(err, domain.ErrMalformed) || calls.Load() != 0 {
		t.Fatalf("corrupt evidence reached callback: calls=%d err=%v", calls.Load(), err)
	}
}

func TestFailureAfterFirstInsertRollsBackWholeWriteSetAndPersistsAfterReopen(t *testing.T) {
	path := t.TempDir() + "/authority.db"
	base := contractFixture(t)
	opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	p := opened.(*provider)
	if _, err := p.db.ExecContext(t.Context(), `
		CREATE TRIGGER reject_second_assignment
		BEFORE INSERT ON assignments
		WHEN NEW.assignment_id = 'blocked-second'
		BEGIN
			SELECT RAISE(ABORT, 'forced second-row rejection');
		END`); err != nil {
		t.Fatal(err)
	}
	first := domain.Assignment{Version: "1", ID: "inserted-first", GrantID: "G1", GrantRevision: 1, Recipient: domain.Recipient{Type: "user", ID: "first-user"}, Status: "enabled"}
	second := domain.Assignment{Version: "1", ID: "blocked-second", GrantID: "G1", GrantRevision: 1, Recipient: domain.Recipient{Type: "user", ID: "second-user"}, Status: "enabled"}
	err = p.Update(t.Context(), base.Area, func(storage.Snapshot) (storage.WriteSet, error) {
		return storage.WriteSet{NewAssignments: []domain.Assignment{first, second}}, nil
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want forced insert conflict, got %v", err)
	}
	assertSQLiteAssignmentsAbsent(t, p, base.Area, first.ID, second.ID)
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	assertSQLiteAssignmentsAbsent(t, reopened, base.Area, first.ID, second.ID)
}

func TestTwoProvidersNeverReplayCompetingCallback(t *testing.T) {
	path := t.TempDir() + "/authority.db"
	base := contractFixture(t)
	p1, err := CreateFixture(t.Context(), path, []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	defer p1.Close()
	p2, err := Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer p2.Close()
	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- p1.Update(context.Background(), base.Area, func(storage.Snapshot) (storage.WriteSet, error) {
			close(entered)
			<-release
			return storage.WriteSet{}, nil
		})
	}()
	<-entered
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	var calls atomic.Int32
	err = p2.Update(ctx, base.Area, func(storage.Snapshot) (storage.WriteSet, error) { calls.Add(1); return storage.WriteSet{}, nil })
	if err == nil || (!errors.Is(err, domain.ErrConflict) && !errors.Is(err, context.DeadlineExceeded)) {
		t.Fatalf("want explicit conflict/deadline, got %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("callback ran or replayed %d times before lock acquisition", calls.Load())
	}
	close(release)
	if first := <-done; first != nil {
		t.Fatalf("first writer: %v", first)
	}
}

// Helpers deliberately stay SQLite-local; provider-neutral cases live in contracttest.
func contractFixture(t *testing.T) storage.Snapshot {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	return minimalFixture(area)
}

func minimalFixture(area domain.Area) storage.Snapshot {
	content := domain.GrantContent{Version: "1", GrantID: "G1", Revision: 1, Permissions: []string{"read"}, Scope: map[string]string{}}
	return storage.Snapshot{Area: area, Catalog: domain.Catalog{ApplicationID: area.ApplicationID(), Permissions: map[string]domain.PermissionDefinition{"read": {ID: "read", Active: true}}, Scopes: map[string]domain.ScopeDefinition{}, SupportedKeys: map[string][]string{}}, Controls: map[string]domain.GrantControl{}, Contents: map[domain.GrantKey]domain.GrantContent{{ID: "G1", Revision: 1}: content}, Assignments: map[string]domain.Assignment{}, Roles: map[domain.RoleKey]domain.RoleContent{}, Teams: map[string]domain.Team{}, Memberships: []domain.Membership{}, TrustedRoots: map[string]bool{}}
}

func assertSQLiteAssignmentsAbsent(t *testing.T, p storage.Provider, area domain.Area, ids ...string) {
	t.Helper()
	if err := p.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
		for _, id := range ids {
			if _, exists := snapshot.Assignments[id]; exists {
				t.Fatalf("assignment %q survived rollback", id)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
