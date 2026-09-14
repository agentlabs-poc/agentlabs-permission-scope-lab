package sqlite

import (
	"agentlabs.local/registry/domain"
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	modernsqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

//go:embed migrations/001_initial.sql
var schema string

const supportedSchemaVersion = 1

// Store is the registry's SQLite provider. It is this domain's only storage,
// and it names only this domain's table.
type Store struct{ db *sql.DB }

// Open opens an existing registry store, or creates one if create is set.
func Open(ctx context.Context, path string, create bool) (*Store, error) {
	resolved, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve registry path: %w", domain.ErrUnavailable)
	}
	if !create {
		if _, err := os.Stat(resolved); err != nil {
			return nil, fmt.Errorf("stat registry path: %w", domain.ErrUnavailable)
		}
	}
	db, err := sql.Open("sqlite", "file:"+resolved+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, classify(err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(ctx, create); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

// migrate applies the schema on creation and refuses any other version: a store
// this build does not understand is not opened rather than migrated.
func (s *Store) migrate(ctx context.Context, create bool) error {
	var version int
	err := s.db.QueryRowContext(ctx,
		`SELECT schema_version FROM registry_metadata WHERE marker='agentlabs-application-registry'`).Scan(&version)
	if err == nil {
		if version != supportedSchemaVersion {
			return fmt.Errorf("registry schema version %d: %w", version, domain.ErrUnsupported)
		}
		return nil
	}
	if !create {
		return fmt.Errorf("not an application registry store: %w", domain.ErrUnavailable)
	}
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return classify(err)
	}
	return nil
}

// classify turns a driver error into this domain's vocabulary. Errors never
// cross a domain boundary untranslated.
func classify(err error) error {
	if err == nil {
		return nil
	}
	for _, category := range []error{domain.ErrMalformed, domain.ErrRejected, domain.ErrConflict, domain.ErrNotFound, domain.ErrUnsupported, domain.ErrUnavailable} {
		if errors.Is(err, category) {
			return err
		}
	}
	var se *modernsqlite.Error
	if errors.As(err, &se) {
		switch se.Code() & 0xff {
		case sqlite3.SQLITE_BUSY, sqlite3.SQLITE_LOCKED, sqlite3.SQLITE_CONSTRAINT:
			return domain.ErrConflict
		}
	}
	return fmt.Errorf("registry storage failure: %w", domain.ErrUnavailable)
}
