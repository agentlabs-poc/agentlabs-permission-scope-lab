package sqlite

import (
	"agentlabs.local/abv/domain"
	"context"
	"database/sql"
	"encoding/json"
)

func (p *provider) insertRole(ctx context.Context, conn *sql.Conn, area domain.Area, role domain.RoleContent) error {
	raw, err := json.Marshal(role.Permissions)
	if err != nil {
		return domain.ErrMalformed
	}
	_, err = conn.ExecContext(ctx, `INSERT INTO roles(tenant_id,application_id,role_id,revision,permissions_json) VALUES(?,?,?,?,?)`, area.TenantID(), area.ApplicationID(), role.ID, role.Revision, raw)
	return classify(err)
}
