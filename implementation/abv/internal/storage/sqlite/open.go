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

type Options struct {
	MaxSnapshotRecords int
	// Registry answers the two questions about applications and installations.
	// It is required: those facts belong to the application registry domain and
	// Auth-AL no longer keeps a copy, so a provider without one can resolve
	// nothing. Open refuses rather than letting every read fail separately.
	//
	// Required means *a* registry, not *this* registry — the legacy tables
	// behind an adapter satisfy it, which is the point of a port.
	Registry Registry
}

type provider struct {
	db                 *sql.DB
	// registry answers whether an application exists and whether a tenant holds
	// it. Both facts belong to the application registry domain, not here. When
	// it is absent the provider falls back to its own tables, which is what a
	// lab fixture and the existing tests use.
	registry Registry
	maxSnapshotRecords int
	// afterCatalog is an internal deterministic test seam for proving that one
	// read transaction pins all cross-query evidence to the same DB version.
	afterCatalog func(context.Context) error
}

// Open is OpenWithOptions with the defaults. The registry is still required —
// it is a fact Auth-AL does not hold, not an option.
func Open(ctx context.Context, path string, registry Registry) (storage.Provider, error) {
	return OpenWithOptions(ctx, path, Options{Registry: registry})
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
	if options.Registry == nil {
		return nil, domain.ErrUnsupported
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
	p := &provider{db: db, maxSnapshotRecords: limit, registry: options.Registry}
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
	} else {
		found, markerErr := hasMarker(ctx, conn)
		if markerErr != nil {
			return fail(classify(markerErr))
		}
		if !found {
			return fail(storage.ErrNotABVDatabase)
		}
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
	for _, category := range []error{
		domain.ErrMalformed,
		domain.ErrRejected,
		domain.ErrUnsupported,
		domain.ErrNotFound,
		domain.ErrConflict,
		domain.ErrUnavailable,
	} {
		if errors.Is(err, category) {
			return err
		}
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
