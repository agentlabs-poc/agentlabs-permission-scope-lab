package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// An assignment is keyed by what it binds, not by its own identifier:
//
//	key1=abv  key2=assignment  key3=<application>
//	key4=<grant id>  key5=<recipient type>  key6=<recipient id>
//
// Q-104 says a grant has at most one current assignment to a given recipient
// type and identity, and that retained disabled ones count. In the old table
// that was a UNIQUE of its own. `abv_l1_records` is shared by every record type
// and cannot carry a per-type constraint, so the pair either occupies the key
// path — where the envelope's existing primary key enforces it for nothing — or
// the rule stops being enforced by storage at all.
//
// This is why the identifier stays in the value. The primary key is all ten
// slots, so adding the id to the path would widen what is unique to
// (grant, recipient, id) and admit exactly the duplicate Q-104 forbids.
type assignmentPayload struct {
	ID            string `json:"id"`
	GrantRevision int64  `json:"grant_revision"`
	Status        string `json:"status"`
}

func encodeAssignment(a domain.Assignment) ([]byte, error) {
	if err := validAssignmentShape(a); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(assignmentPayload{ID: a.ID, GrantRevision: a.GrantRevision, Status: a.Status})
	if err != nil {
		return nil, domain.ErrMalformed
	}
	return raw, nil
}

// decodeAssignment rebuilds the whole record: the binding from the key slots,
// the format version from the contract, the rest from the value.
func decodeAssignment(grantID, recipientType, recipientID string, raw []byte) (domain.Assignment, error) {
	var payload assignmentPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return domain.Assignment{}, domain.ErrMalformed
	}
	a := domain.Assignment{
		Version: "1", ID: payload.ID,
		GrantID: grantID, GrantRevision: payload.GrantRevision,
		Recipient: domain.Recipient{Type: recipientType, ID: recipientID},
		Status:    payload.Status,
	}
	if err := validAssignmentShape(a); err != nil {
		return domain.Assignment{}, err
	}
	return a, nil
}

func validAssignmentShape(a domain.Assignment) error {
	if a.Version != "1" {
		if a.Version == "" {
			return domain.ErrMalformed
		}
		return domain.ErrUnsupported
	}
	if a.GrantRevision <= 0 {
		return domain.ErrMalformed
	}
	// Every identifier here is a base-36 Snowflake: the assignment's own, the
	// grant it binds, and the recipient — a team id or a human id, both already
	// settled. Accepting anything else would let an illustrative id back in.
	for _, id := range []string{a.ID, a.GrantID, a.Recipient.ID} {
		if !codec.ValidRoleID(id) {
			return domain.ErrMalformed
		}
	}
	if a.Recipient.Type != "user" && a.Recipient.Type != "group" {
		return domain.ErrMalformed
	}
	if a.Status != "enabled" && a.Status != "disabled" {
		return domain.ErrMalformed
	}
	return nil
}

func insertAssignment(ctx context.Context, conn *sql.Conn, area domain.Area, a domain.Assignment) error {
	raw, err := encodeAssignment(a)
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, `
		INSERT INTO abv_l1_records
		  (boundary, tenant_id, key1, key2, key3, key4, key5, key6, value)
		VALUES ('tenant', ?, 'abv', 'assignment', ?, ?, ?, ?, ?)`,
		area.TenantID(), area.ApplicationID(), a.GrantID, a.Recipient.Type, a.Recipient.ID, string(raw))
	return classify(err)
}

// readAssignmentByBinding is the lookup the key path makes free, and it is the
// one resolution performs per level of every route it walks.
func readAssignmentByBinding(ctx context.Context, conn *sql.Conn, area domain.Area, grantID string, recipient domain.Recipient) (domain.Assignment, error) {
	var raw string
	err := conn.QueryRowContext(ctx, `
		SELECT value FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='assignment'
		   AND key3=? AND key4=? AND key5=? AND key6=?`,
		area.TenantID(), area.ApplicationID(), grantID, recipient.Type, recipient.ID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Assignment{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Assignment{}, classify(err)
	}
	return decodeAssignment(grantID, recipient.Type, recipient.ID, []byte(raw))
}

// readAssignmentByID is the lookup the key path does not make free. It scans the
// area's assignments, which is affordable because every write already loads the
// whole area snapshot anyway — and keeping it means the operations' signatures
// do not change, so the key layout stays inside this file.
func readAssignmentByID(ctx context.Context, conn *sql.Conn, area domain.Area, id string) (domain.Assignment, error) {
	rows, err := conn.QueryContext(ctx, `
		SELECT key4,key5,key6,value FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='assignment' AND key3=?
		   AND json_extract(value,'$.id')=?
		 ORDER BY key4,key5,key6 LIMIT 1`,
		area.TenantID(), area.ApplicationID(), id)
	if err != nil {
		return domain.Assignment{}, classify(err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return domain.Assignment{}, classify(err)
		}
		return domain.Assignment{}, domain.ErrNotFound
	}
	var grantID, recipientType, recipientID, raw string
	if err := rows.Scan(&grantID, &recipientType, &recipientID, &raw); err != nil {
		return domain.Assignment{}, classify(err)
	}
	return decodeAssignment(grantID, recipientType, recipientID, []byte(raw))
}

func deleteAssignmentByID(ctx context.Context, conn *sql.Conn, area domain.Area, id string) error {
	found, err := readAssignmentByID(ctx, conn, area, id)
	if err != nil {
		return err
	}
	result, err := conn.ExecContext(ctx, `
		DELETE FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='assignment'
		   AND key3=? AND key4=? AND key5=? AND key6=?`,
		area.TenantID(), area.ApplicationID(), found.GrantID, found.Recipient.Type, found.Recipient.ID)
	if err != nil {
		return classify(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return classify(err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// updateAssignmentValue rewrites the payload of one binding. The compare-and-set
// is on the stored id, so a concurrent delete-and-recreate of the same binding
// under a different id loses rather than being silently overwritten.
func updateAssignmentValue(ctx context.Context, conn *sql.Conn, area domain.Area, a domain.Assignment) error {
	raw, err := encodeAssignment(a)
	if err != nil {
		return err
	}
	result, err := conn.ExecContext(ctx, `
		UPDATE abv_l1_records SET value=?
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='assignment'
		   AND key3=? AND key4=? AND key5=? AND key6=? AND json_extract(value,'$.id')=?`,
		string(raw), area.TenantID(), area.ApplicationID(),
		a.GrantID, a.Recipient.Type, a.Recipient.ID, a.ID)
	if err != nil {
		return classify(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return classify(err)
	}
	if affected != 1 {
		return domain.ErrConflict
	}
	return nil
}
