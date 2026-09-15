package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// EstablishRoot is the four writes that bring a root into existence, in the
// caller's transaction:
//
//	abv.grant            the head, trusted_root=true
//	abv.grant_revision   revision 1 — no parent, no permissions, scope {}
//	   trust evidence    the head's trusted_root field carries it
//	abv.assignment       to the holder team, adopting revision 1
//
// Atomicity is the mechanism rather than a workflow. Q-117 requires that
// incomplete setup provide no authority, and one transaction gives that for
// free: there is no moment where a root exists without its assignment, because
// a half-established root cannot be committed.
//
// No grant operation can reach this. That is the whole of Q-113's requirement,
// expressed as vocabulary instead of as a check.
func (p *provider) establishRoot(ctx context.Context, conn *sql.Conn, area domain.Area, e storage.NewRoot) error {
	// A root is the ceiling for its whole area, so exactly one may exist.
	existing, err := rootOf(ctx, conn, area)
	if err != nil {
		return err
	}
	if existing != "" {
		return domain.ErrConflict
	}
	// Q-114: registration precedes acceptance. An empty catalog computes an
	// empty ceiling, and a ceiling of nothing is not a ceiling.
	state, err := catalogHasPermissions(ctx, conn, area.ApplicationID())
	if err != nil {
		return err
	}
	if !state {
		return domain.ErrRejected
	}
	// A root under a ceiling is not a root. rootRoute enforces this at
	// resolution; refusing here means the record is never written in a shape
	// resolution would reject.
	team, err := readTeam(ctx, conn, area.TenantID(), e.HolderTeamID)
	if err != nil {
		return err
	}
	if team.ParentID != "" {
		return domain.ErrRejected
	}
	if err := insertGrantHead(ctx, conn, area, e.Grant); err != nil {
		return err
	}
	if err := insertGrantRevisionRow(ctx, conn, area, e.Content); err != nil {
		return err
	}
	return insertAssignment(ctx, conn, area, e.Assignment)
}

// rootOf returns the area's root grant id, or "" when it has none.
func rootOf(ctx context.Context, conn *sql.Conn, area domain.Area) (string, error) {
	var id string
	err := conn.QueryRowContext(ctx, `
		SELECT key4 FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='grant' AND key3=?
		   AND json_extract(value,'$.trusted_root')=1
		 LIMIT 1`, area.TenantID(), area.ApplicationID()).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", classify(err)
	}
	return id, nil
}

func catalogHasPermissions(ctx context.Context, conn *sql.Conn, namespace string) (bool, error) {
	var found int
	err := conn.QueryRowContext(ctx, `
		SELECT 1 FROM abv_l1_records
		 WHERE tenant_id='' AND key1='abv' AND key2='permission' AND key3=?
		 LIMIT 1`, namespace).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, classify(err)
	}
	return true, nil
}

// readTeam decodes the value through teamPayload rather than reaching into the
// JSON by hand. The first version of this extracted "$.parent" — the field is
// parent_id — so every holder read as top-level and a parented team could hold a
// root. Naming the shape once is what stops that.
func readTeam(ctx context.Context, conn *sql.Conn, tenantID, id string) (domain.Team, error) {
	var name, raw string
	err := conn.QueryRowContext(ctx, `
		SELECT key4, value FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='team' AND key3=?`,
		tenantID, id).Scan(&name, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Team{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Team{}, classify(err)
	}
	var payload teamPayload
	if json.Unmarshal([]byte(raw), &payload) != nil {
		return domain.Team{}, domain.ErrMalformed
	}
	return domain.Team{ID: id, Name: name, ParentID: payload.ParentID}, nil
}
