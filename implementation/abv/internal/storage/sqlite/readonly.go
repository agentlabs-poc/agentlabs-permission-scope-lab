package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Reader exposes only read access to an existing ABV SQLite database.
type Reader struct{ provider *provider }

func OpenReadOnly(ctx context.Context, path string) (*Reader, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(path) == "" {
		return nil, domain.ErrMalformed
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve sqlite path: %w", domain.ErrUnavailable)
	}
	info, err := os.Stat(abs)
	if err != nil || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("stat sqlite path: %w", domain.ErrUnavailable)
	}
	u := url.URL{Scheme: "file", Path: abs}
	q := u.Query()
	q.Add("mode", "ro")
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "busy_timeout(0)")
	q.Add("_pragma", "query_only(1)")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, classify(err)
	}
	p := &provider{db: db, maxSnapshotRecords: defaultMaxSnapshotRecords}
	fail := func(err error) (*Reader, error) { _ = db.Close(); return nil, err }
	conn, err := p.connection(ctx)
	if err != nil {
		return fail(err)
	}
	defer conn.Close()
	var queryOnly int
	if err = conn.QueryRowContext(ctx, `PRAGMA query_only`).Scan(&queryOnly); err != nil {
		return fail(classify(err))
	}
	if queryOnly != 1 {
		return fail(fmt.Errorf("sqlite query-only mode disabled: %w", domain.ErrUnavailable))
	}
	found, err := hasMarker(ctx, conn)
	if err != nil {
		return fail(classify(err))
	}
	if !found {
		return fail(storage.ErrNotABVDatabase)
	}
	return &Reader{provider: p}, nil
}

func (r *Reader) Read(ctx context.Context, area domain.Area, callback func(storage.Snapshot) error) error {
	return r.provider.Read(ctx, area, callback)
}

func (r *Reader) Close() error { return r.provider.Close() }
