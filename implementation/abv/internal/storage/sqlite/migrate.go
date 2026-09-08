package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
)

//go:embed migrations/001_initial.sql
var initialMigration string

func migrate(ctx context.Context, conn *sql.Conn) error {
	return migrateWithConnection(ctx, conn)
}

func migrateWithConnection(ctx context.Context, conn transactionConnection) error {
	return runTransaction(ctx, conn, "BEGIN IMMEDIATE", func() error {
		_, err := conn.ExecContext(ctx, initialMigration)
		return err
	})
}

func hasMarker(ctx context.Context, conn *sql.Conn) (bool, error) {
	var exists int
	err := conn.QueryRowContext(ctx, `SELECT 1 FROM sqlite_schema WHERE type='table' AND name='abv_metadata'`).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var version int
	err = conn.QueryRowContext(ctx, `SELECT schema_version FROM abv_metadata WHERE marker = 'agentlabs-abv'`).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return version == 1, nil
}
