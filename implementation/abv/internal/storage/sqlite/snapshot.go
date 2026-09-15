package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
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
	// Does this tenant hold this application? That is the registry's fact, and
	// now its only home — the installations table is gone, so there is nothing
	// local to disagree with it. It gates every read of a tenant's authority,
	// which is why it runs first.
	held, err := p.registry.Installed(ctx, area.TenantID(), area.ApplicationID())
	if err != nil {
		return storage.Snapshot{}, err
	}
	if !held {
		return storage.Snapshot{}, domain.ErrNotFound
	}
	s := storage.Snapshot{Area: area, Controls: map[string]domain.GrantControl{}, Contents: map[domain.GrantKey]domain.GrantContent{}, Assignments: map[string]domain.Assignment{}, Roles: map[domain.RoleKey]domain.RoleContent{}, Teams: map[string]domain.Team{}, Memberships: []domain.Membership{}, Ownerships: []domain.Ownership{}, TrustedRoots: map[string]bool{}}
	if err := r.catalog(area.ApplicationID(), &s.Catalog); err != nil {
		return storage.Snapshot{}, err
	}
	if p.afterCatalog != nil {
		if err := p.afterCatalog(ctx); err != nil {
			return storage.Snapshot{}, err
		}
	}
	if err := r.controls(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err := r.contents(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err := r.assignments(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err := r.roles(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err := r.teams(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err := r.memberships(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err := r.ownerships(&s); err != nil {
		return storage.Snapshot{}, err
	}
	if err := r.roots(&s); err != nil {
		return storage.Snapshot{}, err
	}
	return s, nil
}

func (r *snapshotReader) catalog(applicationID string, catalog *domain.Catalog) error {
	state, err := readCatalogState(r.ctx, r.conn, applicationID)
	if err != nil {
		return err
	}
	if err := r.add(); err != nil {
		return err
	}
	*catalog = domain.Catalog{
		ApplicationID: applicationID, Generation: state.Generation,
		Permissions: map[string]domain.PermissionDefinition{}, Scopes: map[string]domain.ScopeDefinition{},
		CompatibilityEnabled: state.CompatibilityEnabled,
	}
	// Permissions are L1 records. The identifier is rebuilt from its slots by
	// the shared codec; storage never assembles the string itself.
	rows, err := r.conn.QueryContext(r.ctx, `
		SELECT boundary,key3,key4,key5,key6,key7,key8,key9,key10,value
		  FROM abv_l1_records
		 WHERE tenant_id='' AND key1='abv' AND key2='permission'
		   AND ((boundary='application' AND key3=?) OR boundary='platform')
		 ORDER BY key3,key4,key5,key6,key7,key8,key9,key10`, applicationID)
	if err != nil {
		return classify(err)
	}
	for rows.Next() {
		var slots codec.PermissionSlots
		var boundary, namespace string
		var payload []byte
		if err = rows.Scan(&boundary, &namespace, &slots[0], &slots[1], &slots[2], &slots[3], &slots[4], &slots[5], &slots[6], &payload); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		// The identifier is rebuilt from the namespace the row carries, which is
		// the application for an application permission and the platform's own
		// namespace for a platform one. An application's catalog holds both:
		// platform permissions are vocabulary every application inherits, the
		// same way application roles are vocabulary every tenant inherits.
		id, keyErr := codec.PermissionFromSlots(namespace, slots)
		if keyErr != nil {
			rows.Close()
			return keyErr
		}
		var value struct {
			Active *bool `json:"active"`
		}
		if json.Unmarshal(payload, &value) != nil || value.Active == nil {
			rows.Close()
			return domain.ErrMalformed
		}
		// The boundary is kept rather than discarded: a root's ceiling is sliced
		// by it, even though evaluation uses the whole union.
		where := domain.Boundary(boundary)
		if !where.Valid() {
			rows.Close()
			return domain.ErrMalformed
		}
		catalog.Permissions[id] = domain.PermissionDefinition{ID: id, Active: *value.Active, Boundary: where, Namespace: namespace}
	}
	if err = finishRows(rows); err != nil {
		return err
	}
	// Scopes are L1 records: key3 the application, key4 the key.
	rows, err = r.conn.QueryContext(r.ctx, `
		SELECT key4, value FROM abv_l1_records
		 WHERE boundary='application' AND tenant_id='' AND key1='abv' AND key2='scope' AND key3=? ORDER BY key4`, applicationID)
	if err != nil {
		return classify(err)
	}
	for rows.Next() {
		var key string
		var payload []byte
		if err = rows.Scan(&key, &payload); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		// A scope record's payload is empty: its presence is the fact.
		if key == "" || len(payload) == 0 {
			rows.Close()
			return domain.ErrMalformed
		}
		catalog.Scopes[key] = domain.ScopeDefinition{Key: key}
	}
	if err = finishRows(rows); err != nil {
		return err
	}
	return nil
}

// controls reads the grant heads. One row now carries both the live status and
// the trust evidence that used to need its own three-column table, so roots()
// below is fed from here rather than from a second read.
func (r *snapshotReader) controls(s *storage.Snapshot) error {
	rows, err := r.areaRows(`
		SELECT key4,value FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='grant' AND key3=?
		 ORDER BY key4`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id, raw string
		if err = rows.Scan(&id, &raw); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		var payload grantHeadPayload
		if id == "" || json.Unmarshal([]byte(raw), &payload) != nil ||
			(payload.Status != "enabled" && payload.Status != "disabled") {
			rows.Close()
			return domain.ErrMalformed
		}
		s.Controls[id] = domain.GrantControl{Version: "1", ID: id, Status: payload.Status}
		if payload.TrustedRoot {
			s.TrustedRoots[id] = true
		}
	}
	return finishRows(rows)
}

func (r *snapshotReader) contents(s *storage.Snapshot) error {
	rows, err := r.areaRows(`
		SELECT key4,key5,value FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='grant_revision' AND key3=?
		 ORDER BY key4,key5`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id, slot, raw string
		if err = rows.Scan(&id, &slot, &raw); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		revision, e := codec.ParseRevision(slot)
		if e != nil {
			rows.Close()
			return e
		}
		content, e := decodeRevision(id, revision, []byte(raw))
		if e != nil {
			rows.Close()
			return e
		}
		s.Contents[domain.GrantKey{ID: id, Revision: revision}] = content
	}
	return finishRows(rows)
}
func (r *snapshotReader) assignments(s *storage.Snapshot) error {
	rows, err := r.areaRows(`
		SELECT key4,key5,key6,value FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='assignment' AND key3=?
		 ORDER BY key4,key5,key6`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var grantID, recipientType, recipientID, raw string
		if err = rows.Scan(&grantID, &recipientType, &recipientID, &raw); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		a, e := decodeAssignment(grantID, recipientType, recipientID, []byte(raw))
		if e != nil {
			rows.Close()
			return e
		}
		// The snapshot is keyed by assignment id because that is how callers ask
		// for one. Two bindings carrying the same id would make that map lossy,
		// and the key path cannot forbid it — the id is in the value.
		if _, clash := s.Assignments[a.ID]; clash {
			rows.Close()
			return domain.ErrMalformed
		}
		s.Assignments[a.ID] = a
	}
	return finishRows(rows)
}

// roles reads the tenant's role revisions from the L1 record store. The
// identity fields come back out of their key slots and the bundle out of the
// value, each by the same codec that wrote them.
func (r *snapshotReader) roles(s *storage.Snapshot) error {
	// Both kinds, in one read: the roles this tenant composed and the roles the
	// application ships to every tenant. They are the same record told apart by
	// the boundary, and a tenant administrator reads its catalog as one list.
	rows, err := r.conn.QueryContext(r.ctx, `
		SELECT key4, key5, key6, tenant_id, value FROM abv_l1_records
		 WHERE key1='abv' AND key2='role' AND key3=?
		   AND ((boundary='tenant' AND tenant_id=?) OR boundary='application')
		 ORDER BY key4, key5`, r.area.ApplicationID(), r.area.TenantID())
	if err != nil {
		err = classify(err)
	}
	if err != nil {
		return err
	}
	for rows.Next() {
		var id, slot, name, tenant, payload string
		if err = rows.Scan(&id, &slot, &name, &tenant, &payload); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		revision, revErr := codec.ParseRevision(slot)
		var content rolePayload
		if !codec.ValidRoleID(id) || revErr != nil || name == "" ||
			json.Unmarshal([]byte(payload), &content) != nil ||
			codec.PermissionList(content.Permissions) != nil {
			rows.Close()
			return domain.ErrMalformed
		}
		// An empty tenant is the whole distinction: a role the application ships
		// carries none; a role a tenant composed carries its own.
		managed := domain.TenantManaged
		if tenant == "" {
			managed = domain.ApplicationManaged
		}
		s.Roles[domain.RoleKey{ID: id, Revision: revision}] = domain.RoleContent{
			ID: id, Name: name, Revision: revision, Permissions: content.Permissions, Managed: managed,
		}
	}
	return finishRows(rows)
}

func (r *snapshotReader) teams(s *storage.Snapshot) error {
	// Teams are tenant-scoped L1 records and carry no application: the tenant is
	// the whole of the scoping, so this read is not area-bound on an application.
	rows, err := r.conn.QueryContext(r.ctx, `
		SELECT key3, key4, value FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='team'
		 ORDER BY key3`, r.area.TenantID())
	if err != nil {
		return classify(err)
	}
	for rows.Next() {
		var id, name, payload string
		if err = rows.Scan(&id, &name, &payload); err != nil {
			rows.Close()
			return classify(err)
		}
		if err = r.add(); err != nil {
			rows.Close()
			return err
		}
		var content teamPayload
		if !codec.ValidRoleID(id) || name == "" || json.Unmarshal([]byte(payload), &content) != nil {
			rows.Close()
			return domain.ErrMalformed
		}
		// A root's parent is empty; any other parent must be a real id.
		if content.ParentID != "" && !codec.ValidRoleID(content.ParentID) {
			rows.Close()
			return domain.ErrMalformed
		}
		s.Teams[id] = domain.Team{ID: id, Name: name, ParentID: content.ParentID}
	}
	return finishRows(rows)
}
// ownerships reads who may administer each team. Q-099 keeps this separate from
// membership: identical shape, different relationship, and neither implies the
// other.
func (r *snapshotReader) ownerships(s *storage.Snapshot) error {
	rows, err := r.conn.QueryContext(r.ctx, `
		SELECT key3, key4 FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='ownership'
		 ORDER BY key3, key4`, r.area.TenantID())
	if err != nil {
		return classify(err)
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
		s.Ownerships = append(s.Ownerships, domain.Ownership{TeamID: team, HumanID: human})
	}
	return finishRows(rows)
}

func (r *snapshotReader) memberships(s *storage.Snapshot) error {
	rows, err := r.conn.QueryContext(r.ctx, `
		SELECT key3, key4 FROM abv_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='membership'
		 ORDER BY key3, key4`, r.area.TenantID())
	if err != nil {
		return classify(err)
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
		// Both halves are ids: the team's is issued here, the human's by the auth
		// service, and both render base 36 so one spelling serves the system.
		if !codec.ValidRoleID(team) || !codec.ValidHumanID(human) {
			rows.Close()
			return domain.ErrMalformed
		}
		s.Memberships = append(s.Memberships, domain.Membership{TeamID: team, HumanID: human})
	}
	return finishRows(rows)
}
// roots is satisfied by controls: the trusted-root marker is a field on the
// grant head, not a separate table. It stays as a named step so the snapshot's
// reading order still says what it loads.
func (r *snapshotReader) roots(_ *storage.Snapshot) error { return nil }
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
