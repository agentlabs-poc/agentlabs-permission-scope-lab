package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/contracttest"
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type beginThenCancelConnection struct {
	conn      *sql.Conn
	beginSeen bool
	rollbacks int
}

func (c *beginThenCancelConnection) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if strings.HasPrefix(query, "BEGIN") && !c.beginSeen {
		c.beginSeen = true
		if _, err := c.conn.ExecContext(context.Background(), query, args...); err != nil {
			return nil, err
		}
		return nil, context.Canceled
	}
	if query == "ROLLBACK" {
		c.rollbacks++
	}
	return c.conn.ExecContext(ctx, query, args...)
}

func (c *beginThenCancelConnection) Raw(callback func(any) error) error {
	return c.conn.Raw(callback)
}

func TestProviderContract(t *testing.T) {
	contracttest.Run(t, contracttest.Factory{
		Open:   Open,
		Create: CreateFixture,
		OpenWithLimit: func(ctx context.Context, path string, limit int) (storage.Provider, error) {
			return OpenWithOptions(ctx, path, Options{MaxSnapshotRecords: limit})
		},
	})
}

func TestCancelledBeginThatExecutedIsRolledBackBeforeConnectionReuse(t *testing.T) {
	path := t.TempDir() + "/authority.db"
	opened, err := Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	p := opened.(*provider)
	defer p.Close()
	conn, err := p.connection(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	uncertain := &beginThenCancelConnection{conn: conn}
	bodyCalled := false
	err = p.transaction(t.Context(), uncertain, "BEGIN IMMEDIATE", func() error {
		bodyCalled = true
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation sentinel lost: %v", err)
	}
	if bodyCalled {
		t.Fatal("transaction body ran after failed BEGIN result")
	}
	if uncertain.rollbacks != 1 {
		t.Fatalf("uncertain BEGIN cleanup count = %d, want 1", uncertain.rollbacks)
	}
	if _, err := conn.ExecContext(t.Context(), "BEGIN IMMEDIATE"); err != nil {
		t.Fatalf("connection retained dirty transaction: %v", err)
	}
	if _, err := conn.ExecContext(t.Context(), "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationCancelledBeginThatExecutedLeavesConnectionAndSchemaClean(t *testing.T) {
	db, err := sql.Open("sqlite", t.TempDir()+"/migration.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	conn, err := db.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	uncertain := &beginThenCancelConnection{conn: conn}
	err = migrateWithConnection(t.Context(), uncertain)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation sentinel lost: %v", err)
	}
	if uncertain.rollbacks != 1 {
		t.Fatalf("uncertain migration BEGIN cleanup count = %d, want 1", uncertain.rollbacks)
	}
	var markerTables int
	if err := conn.QueryRowContext(t.Context(), `SELECT count(*) FROM sqlite_schema WHERE type='table' AND name='abv_metadata'`).Scan(&markerTables); err != nil {
		t.Fatal(err)
	}
	if markerTables != 0 {
		t.Fatal("cancelled migration left a partial ownership schema")
	}
	if err := migrate(t.Context(), conn); err != nil {
		t.Fatalf("connection was not reusable for clean migration: %v", err)
	}
}

func TestOpenPreservesMigrationBeginConflictClassification(t *testing.T) {
	path := t.TempDir() + "/migration-conflict.db"
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	locker, err := db.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer locker.Close()
	if _, err := locker.ExecContext(t.Context(), "BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	defer locker.ExecContext(context.Background(), "ROLLBACK")

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	_, err = open(ctx, path, Options{}, true)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("migration lock lost conflict classification: %v", err)
	}
	if errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("migration lock was reclassified as unavailable: %v", err)
	}
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

func TestMarkerCancellationRemainsCancellationNotUnsupported(t *testing.T) {
	path := t.TempDir() + "/authority.db"
	opened, err := Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	p := opened.(*provider)
	defer p.Close()
	conn, err := p.connection(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	found, err := hasMarker(ctx, conn)
	if found || !errors.Is(err, context.Canceled) || errors.Is(err, storage.ErrNotABVDatabase) {
		t.Fatalf("marker result found=%v err=%v", found, err)
	}
}

func TestLockedMarkerReadRemainsConflictNotUnsupported(t *testing.T) {
	path := t.TempDir() + "/unknown.db"
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(t.Context(), `CREATE TABLE foreign_data(value TEXT)`); err != nil {
		t.Fatal(err)
	}
	locker, err := db.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer locker.Close()
	if _, err := locker.ExecContext(t.Context(), "BEGIN EXCLUSIVE"); err != nil {
		t.Fatal(err)
	}
	defer locker.ExecContext(context.Background(), "ROLLBACK")
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	_, err = Open(ctx, path)
	if !errors.Is(err, domain.ErrConflict) || errors.Is(err, storage.ErrNotABVDatabase) {
		t.Fatalf("locked marker classified as %v", err)
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

func TestCorruptCatalogProjectionsDoNotReachCallback(t *testing.T) {
	cases := map[string]func(*testing.T, *provider, domain.Area){
		"compatibility boolean": func(t *testing.T, p *provider, area domain.Area) {
			p.db.SetMaxOpenConns(1)
			conn, err := p.connection(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			if _, err := conn.ExecContext(t.Context(), `PRAGMA ignore_check_constraints=ON`); err != nil {
				conn.Close()
				t.Fatal(err)
			}
			if _, err := conn.ExecContext(t.Context(), `UPDATE applications SET compatibility_enabled=2 WHERE application_id=?`, area.ApplicationID()); err != nil {
				conn.Close()
				t.Fatal(err)
			}
			if err := conn.Close(); err != nil {
				t.Fatal(err)
			}
		},
		"duplicate token": func(t *testing.T, p *provider, area domain.Area) {
			if _, err := p.db.ExecContext(t.Context(), `UPDATE scope_definitions SET allowed_tokens_json='["$self","$self"]' WHERE application_id=?`, area.ApplicationID()); err != nil {
				t.Fatal(err)
			}
		},
		"unsupported token": func(t *testing.T, p *provider, area domain.Area) {
			if _, err := p.db.ExecContext(t.Context(), `UPDATE scope_definitions SET allowed_tokens_json='["other"]' WHERE application_id=?`, area.ApplicationID()); err != nil {
				t.Fatal(err)
			}
		},
		"null token": func(t *testing.T, p *provider, area domain.Area) {
			if _, err := p.db.ExecContext(t.Context(), `UPDATE scope_definitions SET allowed_tokens_json='[null]' WHERE application_id=?`, area.ApplicationID()); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			base := contractFixture(t)
			base.Catalog.Scopes["dept"] = domain.ScopeDefinition{Key: "dept", AllowedTokens: []string{"$self"}}
			path := t.TempDir() + "/authority.db"
			opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{base})
			if err != nil {
				t.Fatal(err)
			}
			p := opened.(*provider)
			defer p.Close()
			corrupt(t, p, base.Area)
			var calls atomic.Int32
			err = p.Read(t.Context(), base.Area, func(storage.Snapshot) error { calls.Add(1); return nil })
			if !errors.Is(err, domain.ErrMalformed) || calls.Load() != 0 {
				t.Fatalf("corrupt catalog reached callback: calls=%d err=%v", calls.Load(), err)
			}
		})
	}
}

func TestFixtureRejectsInvalidCatalogTokens(t *testing.T) {
	for name, tokens := range map[string][]string{
		"duplicate":   {"$self", "$self"},
		"unsupported": {"other"},
	} {
		t.Run(name, func(t *testing.T) {
			base := contractFixture(t)
			base.Catalog.Scopes["dept"] = domain.ScopeDefinition{Key: "dept", AllowedTokens: tokens}
			if _, err := CreateFixture(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{base}); !errors.Is(err, domain.ErrMalformed) {
				t.Fatalf("invalid fixture tokens accepted: %v", err)
			}
		})
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
	if errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("definite writer conflict was polluted by rollback cleanup: %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("callback ran or replayed %d times before lock acquisition", calls.Load())
	}
	close(release)
	if first := <-done; first != nil {
		t.Fatalf("first writer: %v", first)
	}
}

func TestReadTransactionPinsOneVersionAcrossCatalogAndAssignmentQueries(t *testing.T) {
	path := t.TempDir() + "/authority.db"
	base := contractFixture(t)
	opened, err := CreateFixture(t.Context(), path, []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	reader := opened.(*provider)
	defer reader.Close()
	writerOpened, err := Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	writer := writerOpened.(*provider)
	defer writer.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	afterCatalog := make(chan struct{})
	continueRead := make(chan struct{})
	var release sync.Once
	defer release.Do(func() { close(continueRead) })
	reader.afterCatalog = func(ctx context.Context) error {
		close(afterCatalog)
		select {
		case <-continueRead:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	readDone := make(chan error, 1)
	callbackSawNew := make(chan bool, 1)
	go func() {
		readDone <- reader.Read(ctx, base.Area, func(snapshot storage.Snapshot) error {
			_, exists := snapshot.Assignments["between-queries"]
			callbackSawNew <- exists
			return nil
		})
	}()
	select {
	case <-afterCatalog:
	case <-ctx.Done():
		t.Fatalf("reader did not reach catalog boundary: %v", ctx.Err())
	}
	created := domain.Assignment{Version: "1", ID: "between-queries", GrantID: "G1", GrantRevision: 1, Recipient: domain.Recipient{Type: "user", ID: "between-user"}, Status: "enabled"}
	if err := writer.Update(ctx, base.Area, func(storage.Snapshot) (storage.WriteSet, error) {
		return storage.WriteSet{NewAssignments: []domain.Assignment{created}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	release.Do(func() { close(continueRead) })
	select {
	case err := <-readDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatalf("reader did not finish: %v", ctx.Err())
	}
	select {
	case sawNew := <-callbackSawNew:
		if sawNew {
			t.Fatal("one snapshot combined pre-commit catalog with post-commit assignment")
		}
	case <-ctx.Done():
		t.Fatalf("callback result missing: %v", ctx.Err())
	}
	if err := writer.Read(ctx, base.Area, func(snapshot storage.Snapshot) error {
		if _, exists := snapshot.Assignments[created.ID]; !exists {
			t.Fatal("later snapshot missed committed assignment")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
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
