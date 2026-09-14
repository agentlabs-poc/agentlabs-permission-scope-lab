package sqlite

import (
	"agentlabs.local/abv/domain"
	"errors"
	"context"
	"database/sql"
	"encoding/json"
)

// teamPayload is a team record's value column. The parent is state about the
// team rather than part of its identity — Team2 moving from Team1 to RootTeam is
// the same team — so it lives here and not in a key slot. A slot would make a
// re-parent a delete and an insert.
//
// It holds an id rather than a name: a name is editable, and a hierarchy built on
// names would break the moment a team were renamed.
type teamPayload struct {
	ParentID string `json:"parent_id"`
}

// insertTeam writes a team as a tenant-scoped L1 record:
//
//	key1=abv  key2=team  key3=<team id>  key4=<name>
//
// No application appears, because a team has no application dimension: it is an
// Auth-owned collection of this tenant's humans, and an application's own
// groupings are a separate thing the application keeps itself.
func insertTeam(ctx context.Context, conn *sql.Conn, tenantID string, team domain.Team) error {
	raw, err := json.Marshal(teamPayload{ParentID: team.ParentID})
	if err != nil {
		return domain.ErrMalformed
	}
	_, err = conn.ExecContext(ctx, `
		INSERT INTO abv_l1_records
		  (boundary, tenant_id, key1, key2, key3, key4, value)
		VALUES ('tenant', ?, 'abv', 'team', ?, ?, ?)`,
		tenantID, team.ID, team.Name, string(raw))
	return classify(err)
}

// insertMembership writes one human's place in one team:
//
//	key1=abv  key2=membership  key3=<team id>  key4=<human id>
//
// The value is empty: presence is the fact. Identity is the pair, so writing the
// same membership twice is the same row rather than a duplicate.
func insertMembership(ctx context.Context, conn *sql.Conn, tenantID string, m domain.Membership) error {
	_, err := conn.ExecContext(ctx, `
		INSERT INTO abv_l1_records
		  (boundary, tenant_id, key1, key2, key3, key4, value)
		VALUES ('tenant', ?, 'abv', 'membership', ?, ?, '{}')`,
		tenantID, m.TeamID, m.HumanID)
	return classify(err)
}

// seedTeam and seedMembership are the fixture path, and differ from the write
// path in one way: a team already present for this tenant is left alone rather
// than treated as a conflict.
//
// A tenant's teams are the same teams whichever application's snapshot names
// them — that is what "a team belongs to the tenant" means — so a fixture
// seeding two applications for one tenant states the same fact twice. The write
// path keeps its conflict, because a caller creating a team that exists is a
// different situation from a fixture restating one.
func seedTeam(ctx context.Context, conn *sql.Conn, tenantID string, team domain.Team) error {
	present, err := teamExists(ctx, conn, tenantID, team.ID)
	if err != nil || present {
		return err
	}
	return insertTeam(ctx, conn, tenantID, team)
}

func seedMembership(ctx context.Context, conn *sql.Conn, tenantID string, m domain.Membership) error {
	var exists int
	err := conn.QueryRowContext(ctx, `
		SELECT 1 FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='membership'
		   AND key3=? AND key4=?`, tenantID, m.TeamID, m.HumanID).Scan(&exists)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return classify(err)
	}
	return insertMembership(ctx, conn, tenantID, m)
}

// teamExists reports whether this tenant already holds the team.
func teamExists(ctx context.Context, conn *sql.Conn, tenantID, id string) (bool, error) {
	var exists int
	err := conn.QueryRowContext(ctx, `
		SELECT 1 FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='team' AND key3=?`,
		tenantID, id).Scan(&exists)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, classify(err)
}

// updateTeamParent re-parents a team. The parent is state in the value, so this
// is one update of one record rather than a delete and an insert — which is the
// reason the parent is not in a key slot.
func updateTeamParent(ctx context.Context, conn *sql.Conn, tenantID string, team domain.Team) error {
	raw, err := json.Marshal(teamPayload{ParentID: team.ParentID})
	if err != nil {
		return domain.ErrMalformed
	}
	result, err := conn.ExecContext(ctx, `
		UPDATE abv_l1_records SET value=?
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='team' AND key3=?`,
		string(raw), tenantID, team.ID)
	if err != nil {
		return classify(err)
	}
	return exactlyOne(result)
}

// deleteTeam removes a team row. Validation has already refused the delete if
// anything depends on it, so this removes one row and nothing else: no cascade.
func deleteTeam(ctx context.Context, conn *sql.Conn, tenantID, id string) error {
	result, err := conn.ExecContext(ctx, `
		DELETE FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='team' AND key3=?`,
		tenantID, id)
	if err != nil {
		return classify(err)
	}
	return exactlyOne(result)
}

// deleteMembership removes one human from one team.
func deleteMembership(ctx context.Context, conn *sql.Conn, tenantID string, m domain.Membership) error {
	result, err := conn.ExecContext(ctx, `
		DELETE FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='membership'
		   AND key3=? AND key4=?`,
		tenantID, m.TeamID, m.HumanID)
	if err != nil {
		return classify(err)
	}
	return exactlyOne(result)
}

// exactlyOne turns "the row was not there" into ErrNotFound rather than a silent
// success, so a caller learns its write did nothing.
func exactlyOne(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return classify(err)
	}
	if affected != 1 {
		return domain.ErrNotFound
	}
	return nil
}
