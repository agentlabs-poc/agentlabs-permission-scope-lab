package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
)

func (p *provider) insertGrantRevision(ctx context.Context, conn *sql.Conn, area domain.Area, proposed domain.GrantContent) error {
	// The head must exist: publishing a revision amends a grant, it never
	// originates one. CreateGrant writes the head and revision 1 together.
	head, err := readGrantHead(ctx, conn, area, proposed.GrantID)
	if err != nil {
		return err
	}
	// A root's content is computed from the catalog at resolution, so there is
	// nothing to amend and a published revision would be ignored.
	if head.TrustedRoot {
		return domain.ErrRejected
	}
	latest, err := latestGrantRevision(ctx, conn, area, proposed.GrantID)
	if err != nil {
		return err
	}
	latestRevision := latest.Revision
	if proposed.Revision <= latestRevision {
		return domain.ErrConflict
	}
	if proposed.ParentGrantID != latest.ParentGrantID {
		return domain.ErrRejected
	}
	// The submitted content must be exactly canonical: decoding and re-encoding
	// it has to reproduce the bytes, so a caller cannot smuggle an unknown field
	// or a reordering past the record.
	raw, err := json.Marshal(proposed)
	if err != nil {
		return domain.ErrMalformed
	}
	decoded, err := codec.DecodeContent(raw)
	canonical, marshalErr := json.Marshal(decoded)
	if err != nil {
		return err
	}
	if marshalErr != nil || !bytes.Equal(raw, canonical) {
		return domain.ErrMalformed
	}
	authoritative := storage.Snapshot{Roles: map[domain.RoleKey]domain.RoleContent{}}
	r := snapshotReader{conn: conn, ctx: ctx, area: area, limit: p.maxSnapshotRecords}
	if err = r.catalog(area.ApplicationID(), &authoritative.Catalog); err != nil {
		return err
	}
	if err = r.roles(&authoritative); err != nil {
		return err
	}
	if err = validation.CheckContent(area, authoritative.Catalog, decoded, authoritative.Roles); err != nil {
		return err
	}
	return insertGrantRevisionRow(ctx, conn, area, decoded)
}
