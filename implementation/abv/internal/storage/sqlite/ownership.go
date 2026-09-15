package sqlite

import (
	"agentlabs.local/abv/domain"
	"context"
	"database/sql"
)

// Ownership is authority to administer a team, and it is the membership shape
// for the same reason: a relationship has no id of its own.
//
//	key1=abv  key2=ownership  key3=<team id>  key4=<human id>
//
// key3 is the team rather than the application, which is where this record
// departs from the rest — a team is a tenant's, not an application's, and
// ownership follows the thing it owns.
//
// The value is empty. Presence is the fact; there is nothing about an ownership
// except that it exists.
//
// Membership and ownership are identical paths but for key2, and Q-099 exists to
// say they are different relationships: an owner is not thereby a member, and an
// owner has none of the team's business authority.
func insertOwnership(ctx context.Context, conn *sql.Conn, tenantID string, o domain.Ownership) error {
	_, err := conn.ExecContext(ctx, `
		INSERT INTO abv_l1_records
		  (boundary, tenant_id, key1, key2, key3, key4, value)
		VALUES ('tenant', ?, 'abv', 'ownership', ?, ?, '{}')`,
		tenantID, o.TeamID, o.HumanID)
	return classify(err)
}

func deleteOwnership(ctx context.Context, conn *sql.Conn, tenantID string, o domain.Ownership) error {
	result, err := conn.ExecContext(ctx, `
		DELETE FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='ownership'
		   AND key3=? AND key4=?`, tenantID, o.TeamID, o.HumanID)
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
