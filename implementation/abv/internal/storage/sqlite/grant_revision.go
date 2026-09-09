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
	"errors"
)

func (p *provider) insertGrantRevision(ctx context.Context, conn *sql.Conn, area domain.Area, proposed domain.GrantContent) error {
	var controlRaw []byte
	if err := conn.QueryRowContext(ctx, `SELECT canonical_json FROM grant_controls WHERE tenant_id=? AND application_id=? AND grant_id=?`, area.TenantID(), area.ApplicationID(), proposed.GrantID).Scan(&controlRaw); errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	} else if err != nil {
		return classify(err)
	}
	var control domain.GrantControl
	if err := json.Unmarshal(controlRaw, &control); err != nil || control.Version != "1" || control.ID != proposed.GrantID || (control.Status != "enabled" && control.Status != "disabled") {
		return domain.ErrMalformed
	}
	var trusted int
	if err := conn.QueryRowContext(ctx, `SELECT 1 FROM trusted_roots WHERE tenant_id=? AND application_id=? AND grant_id=?`, area.TenantID(), area.ApplicationID(), proposed.GrantID).Scan(&trusted); err == nil {
		return domain.ErrRejected
	} else if !errors.Is(err, sql.ErrNoRows) {
		return classify(err)
	}
	var latestRevision int64
	var latestRaw []byte
	if err := conn.QueryRowContext(ctx, `SELECT revision,canonical_json FROM grant_contents WHERE tenant_id=? AND application_id=? AND grant_id=? ORDER BY revision DESC LIMIT 1`, area.TenantID(), area.ApplicationID(), proposed.GrantID).Scan(&latestRevision, &latestRaw); errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	} else if err != nil {
		return classify(err)
	}
	latest, err := codec.DecodeContent(latestRaw)
	if err != nil || latest.GrantID != proposed.GrantID || latest.Revision != latestRevision {
		if err != nil {
			return err
		}
		return domain.ErrMalformed
	}
	if proposed.Revision <= latestRevision {
		return domain.ErrConflict
	}
	if proposed.ParentGrantID != latest.ParentGrantID {
		return domain.ErrRejected
	}
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
	_, err = conn.ExecContext(ctx, `INSERT INTO grant_contents(tenant_id,application_id,grant_id,revision,canonical_json) VALUES(?,?,?,?,?)`, area.TenantID(), area.ApplicationID(), decoded.GrantID, decoded.Revision, canonical)
	return classify(err)
}
