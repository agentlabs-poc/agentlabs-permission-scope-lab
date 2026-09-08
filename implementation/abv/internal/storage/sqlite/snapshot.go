package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type snapshotReader struct {
	conn         *sql.Conn
	ctx          context.Context
	area         domain.Area
	count, limit int
}

func (r *snapshotReader) add() error {
	r.count++
	if r.count > r.limit {
		return storage.ErrSnapshotLimit
	}
	return nil
}

func (p *provider) snapshot(ctx context.Context, conn *sql.Conn, area domain.Area) (storage.Snapshot, error) {
	r := snapshotReader{conn: conn, ctx: ctx, area: area, limit: p.maxSnapshotRecords}
	var installed int
	err := conn.QueryRowContext(ctx, `SELECT 1 FROM installations WHERE tenant_id=? AND application_id=?`, area.TenantID(), area.ApplicationID()).Scan(&installed)
	if errors.Is(err, sql.ErrNoRows) {
		return storage.Snapshot{}, domain.ErrNotFound
	}
	if err != nil {
		return storage.Snapshot{}, classify(err)
	}
	s := storage.Snapshot{Area: area, Controls: map[string]domain.GrantControl{}, Contents: map[domain.GrantKey]domain.GrantContent{}, Assignments: map[string]domain.Assignment{}, Roles: map[domain.RoleKey]domain.RoleContent{}, Teams: map[string]domain.Team{}, Memberships: []domain.Membership{}, TrustedRoots: map[string]bool{}}
	if err = r.catalog(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err = r.controls(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err = r.contents(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err = r.assignments(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err = r.roles(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err = r.teams(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err = r.memberships(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err = r.roots(&s); err != nil {
		return storage.Snapshot{}, err
	}
	return s, nil
}

func (r *snapshotReader) catalog(s *storage.Snapshot) error {
	var compat int
	if err := r.conn.QueryRowContext(r.ctx, `SELECT compatibility_enabled FROM applications WHERE application_id=?`, r.area.ApplicationID()).Scan(&compat); err != nil {
		return corruptOrDB(err)
	}
	if err := r.add(); err != nil {
		return err
	}
	s.Catalog = domain.Catalog{ApplicationID: r.area.ApplicationID(), Permissions: map[string]domain.PermissionDefinition{}, Scopes: map[string]domain.ScopeDefinition{}, CompatibilityEnabled: compat == 1, SupportedKeys: map[string][]string{}}
	rows, err := r.conn.QueryContext(r.ctx, `SELECT permission_id,active FROM permissions WHERE application_id=? ORDER BY permission_id`, r.area.ApplicationID())
	if err != nil {
		return classify(err)
	}
	for rows.Next() {
		var id string
		var active int
		if err = rows.Scan(&id, &active); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		if id == "" || (active != 0 && active != 1) {
			rows.Close()
			return domain.ErrMalformed
		}
		s.Catalog.Permissions[id] = domain.PermissionDefinition{ID: id, Active: active == 1}
	}
	if err = finishRows(rows); err != nil {
		return err
	}
	rows, err = r.conn.QueryContext(r.ctx, `SELECT scope_key,allowed_tokens_json FROM scope_definitions WHERE application_id=? ORDER BY scope_key`, r.area.ApplicationID())
	if err != nil {
		return classify(err)
	}
	for rows.Next() {
		var key string
		var raw []byte
		if err = rows.Scan(&key, &raw); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		var tokens []string
		if key == "" || json.Unmarshal(raw, &tokens) != nil || tokens == nil {
			rows.Close()
			return domain.ErrMalformed
		}
		s.Catalog.Scopes[key] = domain.ScopeDefinition{Key: key, AllowedTokens: tokens}
	}
	if err = finishRows(rows); err != nil {
		return err
	}
	rows, err = r.conn.QueryContext(r.ctx, `SELECT permission_id,scope_key FROM supported_scope_keys WHERE application_id=? ORDER BY permission_id,ordinal`, r.area.ApplicationID())
	if err != nil {
		return classify(err)
	}
	for rows.Next() {
		var permission, key string
		if err = rows.Scan(&permission, &key); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		if _, ok := s.Catalog.Permissions[permission]; !ok {
			rows.Close()
			return domain.ErrMalformed
		}
		if _, ok := s.Catalog.Scopes[key]; !ok {
			rows.Close()
			return domain.ErrMalformed
		}
		s.Catalog.SupportedKeys[permission] = append(s.Catalog.SupportedKeys[permission], key)
	}
	return finishRows(rows)
}

func (r *snapshotReader) controls(s *storage.Snapshot) error {
	rows, err := r.areaRows(`SELECT grant_id,version,status,canonical_json FROM grant_controls WHERE tenant_id=? AND application_id=? ORDER BY grant_id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id, version, status string
		var raw []byte
		if err = rows.Scan(&id, &version, &status, &raw); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		var control domain.GrantControl
		if id == "" || version != "1" || (status != "enabled" && status != "disabled") || json.Unmarshal(raw, &control) != nil {
			rows.Close()
			return domain.ErrMalformed
		}
		canonical, marshalErr := json.Marshal(control)
		if marshalErr != nil || !bytes.Equal(canonical, raw) || control.Version != version || control.ID != id || control.Status != status {
			rows.Close()
			return domain.ErrMalformed
		}
		s.Controls[id] = control
	}
	return finishRows(rows)
}
func (r *snapshotReader) contents(s *storage.Snapshot) error {
	rows, err := r.areaRows(`SELECT grant_id,revision,canonical_json FROM grant_contents WHERE tenant_id=? AND application_id=? ORDER BY grant_id,revision`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		var revision int64
		var raw []byte
		if err = rows.Scan(&id, &revision, &raw); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		content, e := codec.DecodeContent(raw)
		canonical, marshalErr := json.Marshal(content)
		if e != nil || marshalErr != nil || !bytes.Equal(canonical, raw) || content.GrantID != id || content.Revision != revision {
			rows.Close()
			if e != nil {
				return e
			}
			return domain.ErrMalformed
		}
		s.Contents[domain.GrantKey{ID: id, Revision: revision}] = content
	}
	return finishRows(rows)
}
func (r *snapshotReader) assignments(s *storage.Snapshot) error {
	rows, err := r.areaRows(`SELECT assignment_id,grant_id,grant_revision,recipient_type,recipient_id,status,canonical_json FROM assignments WHERE tenant_id=? AND application_id=? ORDER BY assignment_id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id, gid, rt, rid, status string
		var rev int64
		var raw []byte
		if err = rows.Scan(&id, &gid, &rev, &rt, &rid, &status, &raw); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		a, e := codec.DecodeAssignment(raw)
		canonical, marshalErr := json.Marshal(a)
		if e != nil || marshalErr != nil || !bytes.Equal(canonical, raw) || a.ID != id || a.GrantID != gid || a.GrantRevision != rev || a.Recipient.Type != rt || a.Recipient.ID != rid || a.Status != status {
			rows.Close()
			if e != nil {
				return e
			}
			return domain.ErrMalformed
		}
		s.Assignments[id] = a
	}
	return finishRows(rows)
}
func (r *snapshotReader) roles(s *storage.Snapshot) error {
	rows, err := r.areaRows(`SELECT role_id,revision,permissions_json FROM roles WHERE tenant_id=? AND application_id=? ORDER BY role_id,revision`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		var rev int64
		var raw []byte
		if err = rows.Scan(&id, &rev, &raw); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		var permissions []string
		if id == "" || rev <= 0 || json.Unmarshal(raw, &permissions) != nil || codec.PermissionList(permissions) != nil {
			rows.Close()
			return domain.ErrMalformed
		}
		s.Roles[domain.RoleKey{ID: id, Revision: rev}] = domain.RoleContent{ID: id, Revision: rev, Permissions: permissions}
	}
	return finishRows(rows)
}
func (r *snapshotReader) teams(s *storage.Snapshot) error {
	rows, err := r.areaRows(`SELECT team_id,parent_id FROM teams WHERE tenant_id=? AND application_id=? ORDER BY team_id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id, parent string
		if err = rows.Scan(&id, &parent); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		if id == "" {
			rows.Close()
			return domain.ErrMalformed
		}
		s.Teams[id] = domain.Team{ID: id, ParentID: parent}
	}
	return finishRows(rows)
}
func (r *snapshotReader) memberships(s *storage.Snapshot) error {
	rows, err := r.areaRows(`SELECT team_id,human_id FROM memberships WHERE tenant_id=? AND application_id=? ORDER BY team_id,human_id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var team, human string
		if err = rows.Scan(&team, &human); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		if team == "" || human == "" {
			rows.Close()
			return domain.ErrMalformed
		}
		s.Memberships = append(s.Memberships, domain.Membership{TeamID: team, HumanID: human})
	}
	return finishRows(rows)
}
func (r *snapshotReader) roots(s *storage.Snapshot) error {
	rows, err := r.areaRows(`SELECT grant_id FROM trusted_roots WHERE tenant_id=? AND application_id=? ORDER BY grant_id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		if id == "" {
			rows.Close()
			return domain.ErrMalformed
		}
		s.TrustedRoots[id] = true
	}
	return finishRows(rows)
}
func (r *snapshotReader) areaRows(query string) (*sql.Rows, error) {
	rows, err := r.conn.QueryContext(r.ctx, query, r.area.TenantID(), r.area.ApplicationID())
	if err != nil {
		return nil, classify(err)
	}
	return rows, nil
}
func finishRows(rows *sql.Rows) error {
	err := rows.Err()
	closeErr := rows.Close()
	if err != nil {
		return classify(err)
	}
	return classify(closeErr)
}
func corruptOrDB(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrMalformed
	}
	return classify(err)
}
func malformedf(message string) error { return fmt.Errorf("%s: %w", message, domain.ErrMalformed) }
