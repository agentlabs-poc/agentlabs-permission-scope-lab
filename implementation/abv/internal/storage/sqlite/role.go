package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"context"
	"database/sql"
	"encoding/json"
)

// rolePayload is a role record's value column. The permission list is the
// substance of the record, and it is never queried structurally — so it stays in
// the value as the list it already is, while the identity fields climb out into
// key slots where they can be.
type rolePayload struct {
	Permissions []string `json:"permissions"`
}

// insertRole writes a role revision as a tenant-scoped L1 record:
//
//	key1=abv  key2=role  key3=<application>  key4=<id>  key5=<revision>  key6=<name>
//
// The revision is a key slot rather than a column. Record identity is the whole
// key path, so two revisions of one role have to differ within it; put the
// revision in the value and revision 2 collides with revision 1. It is
// zero-padded because a slot is TEXT and "10" sorts before "2".
//
// Nothing is written to the envelope's revision column: it is drift from the
// canonical 123 shape and is to be removed.
func insertRole(ctx context.Context, conn *sql.Conn, applicationID, tenantID string, role domain.RoleContent) error {
	slot, err := codec.RenderRevision(role.Revision)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(rolePayload{Permissions: role.Permissions})
	if err != nil {
		return domain.ErrMalformed
	}
	// An empty tenant is the whole difference between an application role and a
	// tenant role: same key path, same payload, one carries a tenant and the
	// other does not. It is '' rather than NULL for the reason permissions and
	// scopes use '': a NULL is distinct in a unique index, so duplicates could
	// coexist.
	_, err = conn.ExecContext(ctx, `
		INSERT INTO abv_l1_records
		  (tenant_id, key1, key2, key3, key4, key5, key6, value)
		VALUES (?, 'abv', 'role', ?, ?, ?, ?, ?)`,
		tenantID, applicationID, role.ID, slot, role.Name, string(raw))
	return classify(err)
}

func (p *provider) insertRole(ctx context.Context, conn *sql.Conn, area domain.Area, role domain.RoleContent) error {
	tenant := area.TenantID()
	if role.Managed == domain.ApplicationManaged {
		tenant = ""
	}
	return insertRole(ctx, conn, area.ApplicationID(), tenant, role)
}
