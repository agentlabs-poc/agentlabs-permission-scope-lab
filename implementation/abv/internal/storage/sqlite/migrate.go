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
	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, initialMigration); err != nil {
		_, rollbackErr := conn.ExecContext(context.Background(), "ROLLBACK")
		return errors.Join(err, rollbackErr)
	}
	_, err := conn.ExecContext(ctx, "COMMIT")
	return err
}

func hasMarker(ctx context.Context, conn *sql.Conn) bool {
	var version int
	err := conn.QueryRowContext(ctx, `SELECT schema_version FROM abv_metadata WHERE marker = 'agentlabs-abv'`).Scan(&version)
	return err == nil && version == 1
}
