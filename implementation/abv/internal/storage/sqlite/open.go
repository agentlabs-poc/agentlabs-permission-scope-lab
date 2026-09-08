// Package sqlite implements the ABV storage provider with SQLite transactions.
package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
	modernsqlite "modernc.org/sqlite"
)

const defaultMaxSnapshotRecords = 10_000

type Options struct{ MaxSnapshotRecords int }

type provider struct {
	db                 *sql.DB
	maxSnapshotRecords int
}

func Open(ctx context.Context, path string) (storage.Provider, error) {
	return OpenWithOptions(ctx, path, Options{})
}

func OpenWithOptions(ctx context.Context, path string, options Options) (storage.Provider, error) {
	return open(ctx, path, options, false)
}

func open(ctx context.Context, path string, options Options, fixtureCreated bool) (storage.Provider, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(path) == "" {
		return nil, domain.ErrMalformed
	}
	limit := options.MaxSnapshotRecords
	if limit == 0 {
		limit = defaultMaxSnapshotRecords
	}
	if limit < 0 {
		return nil, domain.ErrMalformed
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve sqlite path: %w", domain.ErrUnavailable)
	}
	_, statErr := os.Stat(abs)
	newFile := errors.Is(statErr, os.ErrNotExist)
	if statErr != nil && !newFile {
		return nil, fmt.Errorf("stat sqlite path: %w", domain.ErrUnavailable)
	}
	if fixtureCreated {
		newFile = true
	} else if newFile {
		file, createErr := os.OpenFile(abs, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if createErr != nil {
			return nil, fmt.Errorf("create sqlite path: %w", domain.ErrConflict)
		}
		if closeErr := file.Close(); closeErr != nil {
			return nil, fmt.Errorf("close new sqlite path: %w", domain.ErrUnavailable)
		}
	}
	u := url.URL{Scheme: "file", Path: abs}
	q := u.Query()
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "busy_timeout(0)")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, classify(err)
	}
	p := &provider{db: db, maxSnapshotRecords: limit}
	fail := func(err error) (storage.Provider, error) { _ = db.Close(); return nil, err }
	conn, err := p.connection(ctx)
	if err != nil {
		return fail(err)
	}
	defer conn.Close()
	if newFile {
		if err := migrate(ctx, conn); err != nil {
			return fail(classify(err))
		}
	} else if !hasMarker(ctx, conn) {
		return fail(storage.ErrNotABVDatabase)
	}
	var mode string
	if err := conn.QueryRowContext(ctx, `PRAGMA journal_mode=WAL`).Scan(&mode); err != nil {
		return fail(classify(err))
	}
	if strings.ToLower(mode) != "wal" {
		return fail(fmt.Errorf("sqlite WAL unavailable: %w", domain.ErrUnavailable))
	}
	return p, nil
}

func (p *provider) connection(ctx context.Context) (*sql.Conn, error) {
	conn, err := p.db.Conn(ctx)
	if err != nil {
		return nil, classify(err)
	}
	if _, err = conn.ExecContext(ctx, `PRAGMA foreign_keys=ON`); err != nil {
		conn.Close()
		return nil, classify(err)
	}
	var enabled int
	if err = conn.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&enabled); err != nil || enabled != 1 {
		conn.Close()
		if err != nil {
			return nil, classify(err)
		}
		return nil, fmt.Errorf("sqlite foreign keys disabled: %w", domain.ErrUnavailable)
	}
	return conn, nil
}

func (p *provider) Close() error { return classify(p.db.Close()) }

func classify(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var se *modernsqlite.Error
	if errors.As(err, &se) {
		switch se.Code() & 0xff {
		case 5, 6:
			return fmt.Errorf("sqlite write conflict: %w", domain.ErrConflict)
		case 19:
			return fmt.Errorf("sqlite constraint: %w", domain.ErrConflict)
		}
	}
	return fmt.Errorf("sqlite storage failure: %w", domain.ErrUnavailable)
}
