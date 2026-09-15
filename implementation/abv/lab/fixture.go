// Package lab contains controlled disposable fixtures for local ABV scenarios.
package lab

import (
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/sqlite"
	"context"
)

// CreateSQLite creates a new SQLite fixture and refuses every existing path.
func CreateSQLite(ctx context.Context, path string, snapshots []storage.Snapshot) (storage.Provider, error) {
	return sqlite.CreateFixture(ctx, path, snapshots)
}
