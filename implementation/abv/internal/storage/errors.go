package storage

import (
	"agentlabs.local/abv/domain"
	"fmt"
)

var (
	ErrSnapshotLimit  = fmt.Errorf("storage snapshot record limit exceeded: %w", domain.ErrUnavailable)
	ErrNotABVDatabase = fmt.Errorf("not an ABV database: %w", domain.ErrUnsupported)
)
