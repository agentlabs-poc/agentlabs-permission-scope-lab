package abv

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/sqlite"
	"context"
)

// OpenSQLiteReadOnly opens a store that can answer but never change.
//
// Enforcement is the reason it exists. An agent resolving authority on every
// request has no business holding a writable handle on the tenant's authority,
// and until now it had no way not to: OpenSQLite was the only door, so the
// enforcement adapter reached past the facade into the internal packages instead
// — which is how it ended up with its own copy of work the facade now does.
//
// The administration is still required and still consulted. A read-only store
// narrows what can be done to the records; it decides nothing about who may ask,
// which remains CheckAuthorityRead's question.
func OpenSQLiteReadOnly(ctx context.Context, path string, administration Administration, clock Clock, registry Registry) (*Facade, error) {
	if registry == nil {
		return nil, domain.ErrMalformed
	}
	reader, err := sqlite.OpenReadOnly(ctx, path, registry)
	if err != nil {
		return nil, err
	}
	facade, err := New(readOnlyProvider{reader}, administration, clock)
	if err != nil {
		_ = reader.Close()
		return nil, err
	}
	return facade, nil
}

// readOnlyProvider is a Provider whose Update refuses.
//
// It refuses rather than being absent because the facade takes one Provider for
// every operation, and a caller that reaches a write through a read-only store
// should be told so — ErrUnsupported is "this store does not do that", which is
// exactly the case. The alternative, a second facade type with only the reads on
// it, would duplicate forty method signatures to say the same thing.
type readOnlyProvider struct{ reader *sqlite.Reader }

func (p readOnlyProvider) Read(ctx context.Context, area domain.Area, fn func(storage.Snapshot) error) error {
	return p.reader.Read(ctx, area, fn)
}

func (p readOnlyProvider) Update(context.Context, domain.Area, func(storage.Snapshot) (storage.WriteSet, error)) error {
	return domain.ErrUnsupported
}

func (p readOnlyProvider) UpdateAdministered(context.Context, domain.Area, func(storage.Snapshot) (storage.WriteSet, error)) error {
	return domain.ErrUnsupported
}

func (p readOnlyProvider) Close() error { return p.reader.Close() }
