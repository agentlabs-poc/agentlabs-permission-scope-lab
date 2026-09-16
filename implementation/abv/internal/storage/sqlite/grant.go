package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// A grant is two record types, because a grant is two things that change at
// different rates.
//
//	key1=abv  key2=grant           key3=<application>  key4=<id>
//	key1=abv  key2=grant_revision  key3=<application>  key4=<id>  key5=<revision>
//
// The head holds what changes — the status, and whether trusted establishment
// recorded this grant as a root. The revision holds what never does. Publishing
// revision 3 does not touch the head, and disabling the grant does not touch a
// revision; putting the status on a revision would let two revisions of one
// grant disagree about whether it is enabled.
//
// The revision is a key slot for the reason a role's is: record identity is the
// whole key path, so two revisions have to differ within it. It is zero-padded
// because a slot is TEXT and "10" sorts before "2".
type grantHeadPayload struct {
	Status string `json:"status"`
	// Trust evidence, not contract. Q-119 refuses a root flag in the *content*,
	// because content is what a caller submits and a self-asserted flag would be
	// a parentless escape. This is what Auth recorded about the grant, which is
	// where evidence belongs, and it never appears in the canonical JSON.
	TrustedRoot bool `json:"trusted_root"`
}

// grantRevisionPayload is everything the key path does not already carry.
//
// grant_id and revision are absent because they are key slots, and version is
// absent because it is always "1" — ValidateContent rejects anything else, so
// storing it in every row would persist a constant. It is rebuilt on the way
// out, the way a permission identifier is rebuilt from its slots.
type grantRevisionPayload struct {
	ParentGrantID string            `json:"parent_grant_id,omitempty"`
	Permissions   []string          `json:"permissions,omitempty"`
	RoleID        string            `json:"role_id,omitempty"`
	RoleRevision  int64             `json:"role_revision,omitempty"`
	Scope         map[string]string `json:"scope"`
	Validity      *domain.Validity  `json:"validity,omitempty"`
}

func encodeRevision(content domain.GrantContent) ([]byte, error) {
	raw, err := json.Marshal(grantRevisionPayload{
		ParentGrantID: content.ParentGrantID,
		Permissions:   content.Permissions,
		RoleID:        content.RoleID,
		RoleRevision:  content.RoleRevision,
		Scope:         content.Scope,
		Validity:      content.Validity,
	})
	if err != nil {
		return nil, domain.ErrMalformed
	}
	return raw, nil
}

// decodeRevision rebuilds the complete content from the row: the identity from
// the key slots, the format version from the contract, the rest from the value.
func decodeRevision(id string, revision int64, raw []byte) (domain.GrantContent, error) {
	var payload grantRevisionPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return domain.GrantContent{}, domain.ErrMalformed
	}
	content := domain.GrantContent{
		Version: "1", GrantID: id, Revision: revision,
		ParentGrantID: payload.ParentGrantID,
		Permissions:   payload.Permissions,
		RoleID:        payload.RoleID,
		RoleRevision:  payload.RoleRevision,
		Scope:         payload.Scope,
		Validity:      payload.Validity,
	}
	// No missing-scope default here either. SCOPE-007 is path-agnostic — "never
	// drop invalid restrictions or default invalid/missing scope to {}" — so it
	// holds on the way back out as firmly as on the way in: a row with no scope is a
	// malformed record, and reporting it is the only way anyone finds out. This
	// substitution would have turned every such row into the widest child its
	// parent allows, silently, at read time.
	if err := codec.ValidateContent(content); err != nil {
		return domain.GrantContent{}, err
	}
	return content, nil
}

// insertGrantHead requires a base-36 Snowflake, as every identifier Auth-AL
// issues is. This was deliberately loose while the corpus still carried the
// handbook's illustrative G0/G1/G2; the sweep converted them, so acceptance now
// matches issuance.
func insertGrantHead(ctx context.Context, conn *sql.Conn, area domain.Area, grant domain.Grant) error {
	if !codec.ValidRoleID(grant.ID) ||
		(grant.Status != "enabled" && grant.Status != "disabled") {
		return domain.ErrMalformed
	}
	raw, err := json.Marshal(grantHeadPayload{Status: grant.Status, TrustedRoot: grant.TrustedRoot})
	if err != nil {
		return domain.ErrMalformed
	}
	_, err = conn.ExecContext(ctx, `
		INSERT INTO abv_l1_records
		  (boundary, tenant_id, key1, key2, key3, key4, value)
		VALUES ('tenant', ?, 'abv', 'grant', ?, ?, ?)`,
		area.TenantID(), area.ApplicationID(), grant.ID, string(raw))
	return classify(err)
}

func insertGrantRevisionRow(ctx context.Context, conn *sql.Conn, area domain.Area, content domain.GrantContent) error {
	slot, err := codec.RenderRevision(content.Revision)
	if err != nil {
		return err
	}
	raw, err := encodeRevision(content)
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, `
		INSERT INTO abv_l1_records
		  (boundary, tenant_id, key1, key2, key3, key4, key5, value)
		VALUES ('tenant', ?, 'abv', 'grant_revision', ?, ?, ?, ?)`,
		area.TenantID(), area.ApplicationID(), content.GrantID, slot, string(raw))
	return classify(err)
}

func readGrantHead(ctx context.Context, conn *sql.Conn, area domain.Area, id string) (domain.Grant, error) {
	var raw string
	err := conn.QueryRowContext(ctx, `
		SELECT value FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='grant'
		   AND key3=? AND key4=?`, area.TenantID(), area.ApplicationID(), id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Grant{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Grant{}, classify(err)
	}
	var payload grantHeadPayload
	if json.Unmarshal([]byte(raw), &payload) != nil ||
		(payload.Status != "enabled" && payload.Status != "disabled") {
		return domain.Grant{}, domain.ErrMalformed
	}
	return domain.Grant{ID: id, Status: payload.Status, TrustedRoot: payload.TrustedRoot}, nil
}

// latestGrantRevision returns the highest revision of one grant. key5 is
// zero-padded, so ORDER BY key5 DESC is the numeric order.
func latestGrantRevision(ctx context.Context, conn *sql.Conn, area domain.Area, id string) (domain.GrantContent, error) {
	var slot, raw string
	err := conn.QueryRowContext(ctx, `
		SELECT key5,value FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='grant_revision'
		   AND key3=? AND key4=?
		 ORDER BY key5 DESC LIMIT 1`, area.TenantID(), area.ApplicationID(), id).Scan(&slot, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.GrantContent{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.GrantContent{}, classify(err)
	}
	revision, err := codec.ParseRevision(slot)
	if err != nil {
		return domain.GrantContent{}, err
	}
	return decodeRevision(id, revision, []byte(raw))
}

// deleteGrant removes a head and every revision it owns. It is add-only's
// mirror: a grant is created whole and destroyed whole, never left as a head
// with no content or content with no head.
func deleteGrant(ctx context.Context, conn *sql.Conn, area domain.Area, id string) error {
	result, err := conn.ExecContext(ctx, `
		DELETE FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv'
		   AND key2 IN ('grant','grant_revision') AND key3=? AND key4=?`,
		area.TenantID(), area.ApplicationID(), id)
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

// insertGrant writes a grant whole: the head and revision 1, in the caller's
// transaction. The two are never written apart — a head with no content is a
// grant that can be enabled and supplies nothing, and content with no head is
// authority with no live switch.
//
// The content is validated against the catalog here rather than by the caller,
// because the catalog has to be read inside the same transaction that writes:
// a permission retired between the check and the write would otherwise be
// admitted.
func (p *provider) insertGrant(ctx context.Context, conn *sql.Conn, area domain.Area, proposed storage.NewGrant) error {
	if proposed.Content.GrantID != proposed.Grant.ID || proposed.Content.Revision != 1 {
		return domain.ErrMalformed
	}
	if err := codec.ValidateContent(proposed.Content); err != nil {
		return err
	}
	// A parent is required. Establishment writes the parentless kind, and it is
	// deliberately not this operation: a root needs trust evidence, and no grant
	// operation may write that.
	if proposed.Content.ParentGrantID == "" {
		return domain.ErrRejected
	}
	if _, err := readGrantHead(ctx, conn, area, proposed.Content.ParentGrantID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrRejected
		}
		return err
	}
	// Both reads are inside the caller's transaction, which is the point: a
	// permission retired or a role revision published between a check outside
	// and the write would otherwise be admitted.
	authoritative := storage.Snapshot{Roles: map[domain.RoleKey]domain.RoleContent{}}
	r := snapshotReader{conn: conn, ctx: ctx, area: area, limit: p.maxSnapshotRecords}
	if err := r.catalog(area.ApplicationID(), &authoritative.Catalog); err != nil {
		return err
	}
	if err := r.roles(&authoritative); err != nil {
		return err
	}
	if err := validation.CheckContent(area, authoritative.Catalog, proposed.Content, authoritative.Roles); err != nil {
		return err
	}
	if err := insertGrantHead(ctx, conn, area, proposed.Grant); err != nil {
		return err
	}
	return insertGrantRevisionRow(ctx, conn, area, proposed.Content)
}
