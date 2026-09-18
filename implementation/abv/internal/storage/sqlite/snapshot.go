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
	"strings"
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

// attachAdministrative reads the tenant's Auth-namespace authority into the
// snapshot already in hand, on the same connection and inside the same
// transaction — so the authority a gate resolves cannot change between the check
// and the write it authorizes.
//
// It deliberately does not go through p.snapshot. That path asks the registry
// whether the tenant holds the application, and the platform namespace is not an
// application: Auth is never registered and never installed, so the question has
// no true answer to give. Everything else is the same read.
//
// A provider that was not told where the Auth chain lives cannot answer an
// administrative question at all, and says so rather than falling back to the
// business snapshot — which would resolve `auth:` permissions against an
// application's own chain, exactly the leak Q-151 exists to prevent.
func (p *provider) attachAdministrative(ctx context.Context, conn *sql.Conn, s *storage.Snapshot) error {
	if p.platformNamespace == "" {
		return domain.ErrUnsupported
	}
	if s.Area.ApplicationID() == p.platformNamespace {
		// Already the administrative area. Nil is the answer, not an omission:
		// the chain the gate must walk is the one the caller is holding.
		//
		// Checked, not assumed. A namespace that names an *application* would make
		// this branch hand the application's own chain back as the administrative
		// one, and the caller would then resolve `<app>:group::write` against the
		// application's root — an application's root holder silently becoming the
		// tenant's team administrator, which is precisely the leak Q-151 exists to
		// prevent. A misconfigured namespace has to fail closed, not sideways.
		return platformCatalog(s.Catalog, p.platformNamespace)
	}
	area, err := domain.NewArea(s.Area.TenantID(), p.platformNamespace)
	if err != nil {
		return err
	}
	administrative := storage.Snapshot{
		Area: area, Controls: map[string]domain.GrantControl{}, Contents: map[domain.GrantKey]domain.GrantContent{},
		Assignments: map[string]domain.Assignment{}, Roles: map[domain.RoleKey]domain.RoleContent{},
		Teams: map[string]domain.Team{}, Memberships: []domain.Membership{}, Ownerships: []domain.Ownership{},
		TrustedRoots: map[string]bool{},
	}
	r := snapshotReader{conn: conn, ctx: ctx, area: area, limit: p.maxSnapshotRecords}
	if err := r.catalog(p.platformNamespace, &administrative.Catalog); err != nil {
		return err
	}
	for _, read := range []func(*storage.Snapshot) error{r.controls, r.contents, r.assignments, r.roles, r.teams, r.memberships, r.ownerships, r.roots} {
		if err := read(&administrative); err != nil {
			return err
		}
	}
	if err := platformCatalog(administrative.Catalog, p.platformNamespace); err != nil {
		return err
	}
	s.Administrative = &administrative
	return nil
}

// platformCatalog refuses a namespace that owns no platform permission.
//
// It is the one observable difference between "the platform's namespace" and "some
// application's name": Auth's own vocabulary is registered at the platform boundary
// under it. A namespace with none is either misconfigured or not yet bootstrapped,
// and both answers are the same — this store cannot resolve administrative
// authority, so it says so rather than resolving something else.
func platformCatalog(catalog domain.Catalog, namespace string) error {
	for _, definition := range catalog.Permissions {
		if definition.Boundary == domain.PlatformBoundary && definition.Namespace == namespace {
			return nil
		}
	}
	return domain.ErrUnsupported
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
			// Named from the slots, because there is no identifier yet — this is
			// the failure to build one, and it fires before the two permission
			// refusals below.
			return rowf(keyErr, "permission in namespace %q with slots %q", namespace, strings.Join(slots[:], "/"))
		}
		var value struct {
			Active *bool `json:"active"`
		}
		if json.Unmarshal(payload, &value) != nil || value.Active == nil {
			rows.Close()
			return rowf(domain.ErrMalformed, "permission %q", id)
		}
		// The boundary is kept rather than discarded: a root's ceiling is sliced
		// by it, even though evaluation uses the whole union.
		where := domain.Boundary(boundary)
		if !where.Valid() {
			rows.Close()
			return rowf(domain.ErrMalformed, "permission %q boundary %q", id, boundary)
		}
		catalog.Permissions[id] = domain.PermissionDefinition{ID: id, Active: *value.Active, Boundary: where, Namespace: namespace}
	}
	if err = finishRows(rows); err != nil {
		return err
	}
	// Scopes are L1 records: key3 the application or the platform's namespace,
	// key4 the key. The union mirrors the permission read above and for the same
	// reason — a platform key is vocabulary every application inherits.
	rows, err = r.conn.QueryContext(r.ctx, `
		SELECT boundary, key3, key4, value FROM abv_l1_records
		 WHERE tenant_id='' AND key1='abv' AND key2='scope'
		   AND ((boundary='application' AND key3=?) OR boundary='platform')
		 ORDER BY key3, key4`, applicationID)
	if err != nil {
		return classify(err)
	}
	for rows.Next() {
		var boundary, namespace, key string
		var payload []byte
		if err = rows.Scan(&boundary, &namespace, &key, &payload); err != nil {
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
			return rowf(domain.ErrMalformed, "scope %q of application %q", key, applicationID)
		}
		where := domain.Boundary(boundary)
		if !where.Valid() {
			rows.Close()
			return rowf(domain.ErrMalformed, "scope %q boundary %q", key, boundary)
		}
		// A scope key is a bare word, so unlike a permission id it carries no
		// namespace to tell two declarations apart — one catalog, one entry per
		// key. A key claimed by both an application and the platform is therefore
		// genuinely ambiguous, and which declaration survived used to depend on how
		// the application's id sorted against the platform's namespace: an
		// application whose id sorts first had its own opaque values validated as
		// Auth team ids, and one whose id sorts later did not.
		//
		// Refused rather than resolved, because there is no correct winner.
		// CheckScopeRegistration already refuses a new registration that collides,
		// so this state is only reachable for a key an application registered
		// before Auth owned it — a migration to notice loudly, not to guess at.
		if existing, seen := catalog.Scopes[key]; seen {
			rows.Close()
			return rowf(domain.ErrMalformed, "scope %q is claimed at both the %s and %s boundaries, in namespace %q",
				key, existing.Boundary, where, clip(namespace))
		}
		catalog.Scopes[key] = domain.ScopeDefinition{Key: key, Boundary: where}
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
			return rowf(domain.ErrMalformed, "grant %q", id)
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
			return rowf(e, "grant %q revision slot %q", id, slot)
		}
		content, e := decodeRevision(id, revision, []byte(raw))
		if e != nil {
			rows.Close()
			return rowf(e, "grant %q revision %d", id, revision)
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
			return rowf(e, "assignment of grant %q to %s %q", grantID, recipientType, recipientID)
		}
		// The snapshot is keyed by assignment id because that is how callers ask
		// for one. Two bindings carrying the same id would make that map lossy,
		// and the key path cannot forbid it — the id is in the value.
		if _, clash := s.Assignments[a.ID]; clash {
			rows.Close()
			return rowf(domain.ErrMalformed, "assignment %q of grant %q is a duplicate id", a.ID, grantID)
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
			return rowf(domain.ErrMalformed, "role %q revision slot %q", id, slot)
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
			return rowf(domain.ErrMalformed, "team %q", id)
		}
		// A root's parent is empty; any other parent must be a real id.
		if content.ParentID != "" && !codec.ValidRoleID(content.ParentID) {
			rows.Close()
			return rowf(domain.ErrMalformed, "team %q names parent %s", id, clip(content.ParentID))
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
			return rowf(domain.ErrMalformed, "ownership of team %q by human %q", team, human)
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
			return rowf(domain.ErrMalformed, "membership of team %q by human %q", team, human)
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

// rowf names the row a snapshot load stopped on.
//
// One unreadable row stops the whole area — that is the decision, and it is the
// only safe direction: skipping the row means answering authorization questions
// from a store we have admitted we cannot fully read, and if the row was
// somebody's restriction, skipping it widens their access silently.
//
// The cost of that rule is diagnosis, not safety. The load held the grant id and
// the revision at the point of failure and threw both away, so an operator saw a
// bare "malformed input" and had to bisect a database. Naming the row turns that
// into a lookup. The identity stays on Auth's side of the wire — the application
// sees a 503 and nothing else — so nothing about the contract changes.
//
// The wrapped error keeps its class: a row that is unsupported rather than
// malformed still reads as unsupported.
func rowf(err error, format string, args ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// clip bounds a value that came out of a row's payload rather than its key
// columns.
//
// Every other thing rowf names is a key-column value, which the schema bounds.
// A team's parent is the exception: it is read from the record's JSON, and it is
// printed precisely in the branch where it failed to be an identifier — so it is
// whatever the row happens to hold, at whatever length.
func clip(value string) string {
	const limit = 64
	if len(value) <= limit {
		return fmt.Sprintf("%q", value)
	}
	return fmt.Sprintf("%q (truncated from %d bytes)", value[:limit], len(value))
}
