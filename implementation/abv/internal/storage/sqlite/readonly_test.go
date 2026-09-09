package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestOpenReadOnlyRejectsInvalidDatabaseWithoutChangingIt(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing.db")
		if _, err := OpenReadOnly(t.Context(), path); err == nil {
			t.Fatal("opened missing database")
		}
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("missing database was created: %v", err)
		}
	})

	t.Run("blank", func(t *testing.T) {
		if _, err := OpenReadOnly(t.Context(), " \t"); !errors.Is(err, domain.ErrMalformed) {
			t.Fatalf("want malformed path, got %v", err)
		}
	})

	t.Run("nonregular", func(t *testing.T) {
		if _, err := OpenReadOnly(t.Context(), t.TempDir()); err == nil {
			t.Fatal("opened directory")
		}
	})

	for _, tc := range []struct {
		name   string
		schema string
	}{
		{name: "empty", schema: ""},
		{name: "foreign", schema: `CREATE TABLE foreign_data(value TEXT); INSERT INTO foreign_data VALUES ('keep')`},
		{name: "unsupported marker", schema: `CREATE TABLE abv_metadata(marker TEXT PRIMARY KEY, schema_version INTEGER NOT NULL); INSERT INTO abv_metadata VALUES ('agentlabs-abv', 2)`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "database.db")
			db, err := sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			if tc.schema == "" {
				err = os.WriteFile(path, nil, 0600)
			} else {
				_, err = db.Exec(tc.schema)
			}
			if closeErr := db.Close(); err == nil {
				err = closeErr
			}
			if err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = OpenReadOnly(t.Context(), path); !errors.Is(err, storage.ErrNotABVDatabase) {
				t.Fatalf("want non-ABV error, got %v", err)
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(after, before) {
				t.Fatal("failed read-only open changed database bytes")
			}
		})
	}
}

func TestOpenReadOnlyReadsOnlyTheRequestedArea(t *testing.T) {
	first := contractFixture(t)
	secondArea, err := domain.NewArea("other-tenant", first.Area.ApplicationID())
	if err != nil {
		t.Fatal(err)
	}
	second := minimalFixture(secondArea)
	path := filepath.Join(t.TempDir(), "authority.db")
	fixture, err := CreateFixture(t.Context(), path, []storage.Snapshot{first, second})
	if err != nil {
		t.Fatal(err)
	}
	if err = fixture.Close(); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReadOnly(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var got storage.Snapshot
	if err = reader.Read(t.Context(), first.Area, func(snapshot storage.Snapshot) error { got = snapshot; return nil }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, first) {
		t.Fatalf("snapshot mismatch\n got: %#v\nwant: %#v", got, first)
	}
	wrong, err := domain.NewArea(first.Area.TenantID(), "other-application")
	if err != nil {
		t.Fatal(err)
	}
	if err = reader.Read(t.Context(), wrong, func(storage.Snapshot) error { return nil }); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want wrong-area not found, got %v", err)
	}
}

func TestReaderConnectionIsReadOnlyAndSeesLaterCommits(t *testing.T) {
	base := contractFixture(t)
	base.Controls["G1"] = domain.GrantControl{Version: "1", ID: "G1", Status: "enabled"}
	path := filepath.Join(t.TempDir(), "authority.db")
	writer, err := CreateFixture(t.Context(), path, []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()

	reader, err := OpenReadOnly(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	conn, err := reader.provider.connection(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.ExecContext(t.Context(), `UPDATE grant_controls SET status='disabled'`); err == nil {
		t.Fatal("reader connection accepted a write")
	}
	if err = conn.Close(); err != nil {
		t.Fatal(err)
	}

	for range 2 {
		if err = reader.Read(t.Context(), base.Area, func(snapshot storage.Snapshot) error {
			if !reflect.DeepEqual(snapshot, base) {
				t.Fatal("read changed stored records")
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	var mode string
	if err = reader.provider.db.QueryRowContext(t.Context(), `PRAGMA journal_mode`).Scan(&mode); err != nil || mode != "wal" {
		t.Fatalf("journal mode=%q err=%v", mode, err)
	}

	before := base.Controls["G1"]
	after := before
	after.Status = "disabled"
	if err = writer.Update(t.Context(), base.Area, func(storage.Snapshot) (storage.WriteSet, error) {
		return storage.WriteSet{GrantStatusChange: &storage.GrantStatusChange{Before: before, After: after}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err = reader.Read(t.Context(), base.Area, func(snapshot storage.Snapshot) error {
		if got := snapshot.Controls["G1"]; got != after {
			t.Fatalf("reader saw stale control: %#v", got)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestOpenReadOnlyCancellationAndMalformedRowFail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "authority.db")
	base := contractFixture(t)
	fixture, err := CreateFixture(t.Context(), path, []storage.Snapshot{base})
	if err != nil {
		t.Fatal(err)
	}
	p := fixture.(*provider)
	bad := []byte(`{"version":"1","grant_id":"different","revision":1,"permissions":["read"],"scope":{}}`)
	if _, err = p.db.ExecContext(t.Context(), `UPDATE grant_contents SET canonical_json=? WHERE tenant_id=? AND application_id=?`, bad, base.Area.TenantID(), base.Area.ApplicationID()); err != nil {
		t.Fatal(err)
	}
	if err = fixture.Close(); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err = OpenReadOnly(ctx, path); !errors.Is(err, context.Canceled) {
		t.Fatalf("want cancellation, got %v", err)
	}
	reader, err := OpenReadOnly(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	called := false
	if err = reader.Read(t.Context(), base.Area, func(storage.Snapshot) error { called = true; return nil }); !errors.Is(err, domain.ErrMalformed) || called {
		t.Fatalf("malformed row reached callback: called=%v err=%v", called, err)
	}
}
