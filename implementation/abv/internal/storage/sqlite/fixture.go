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
	"os"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
)

// CreateFixture initializes and loads only a brand-new disposable database.
// It is an internal test/lab boundary, not an ordinary provider mutation API.
func CreateFixture(ctx context.Context, path string, snapshots []storage.Snapshot) (storage.Provider, error) {
	if strings.TrimSpace(path) == "" {
		return nil, domain.ErrMalformed
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, errors.Join(domain.ErrConflict, err)
	}
	if err = file.Close(); err != nil {
		return nil, errors.Join(domain.ErrUnavailable, err)
	}
	// A fixture provider answers the registry's questions for exactly the areas
	// it seeded, and nothing else. It has to answer them somehow: the facts live
	// in the registry domain now, and a fixture is not going to compose one.
	opened, err := open(ctx, path, Options{Registry: seededRegistry(snapshots)}, true)
	if err != nil {
		return nil, err
	}
	p := opened.(*provider)
	conn, err := p.connection(ctx)
	if err != nil {
		p.Close()
		return nil, err
	}
	err = p.transaction(ctx, conn, "BEGIN IMMEDIATE", func() error { return seedSnapshots(ctx, conn, snapshots) })
	_ = conn.Close()
	if err != nil {
		p.Close()
		return nil, err
	}
	return p, nil
}

func seedSnapshots(ctx context.Context, conn *sql.Conn, snapshots []storage.Snapshot) error {
	catalogs := map[string]domain.Catalog{}
	for _, s := range snapshots {
		if err := s.Area.Validate(); err != nil {
			return err
		}
		if s.Catalog.ApplicationID != s.Area.ApplicationID() {
			return domain.ErrMalformed
		}
		if old, ok := catalogs[s.Catalog.ApplicationID]; ok && !reflect.DeepEqual(old, s.Catalog) {
			return malformedf("conflicting shared catalog fixture")
		}
		catalogs[s.Catalog.ApplicationID] = s.Catalog
	}
	apps := make([]string, 0, len(catalogs))
	for app := range catalogs {
		apps = append(apps, app)
	}
	sort.Strings(apps)
	for _, app := range apps {
		if err := seedCatalog(ctx, conn, catalogs[app]); err != nil {
			return err
		}
	}
	for _, s := range snapshots {
		if err := seedArea(ctx, conn, s); err != nil {
			return err
		}
	}
	return nil
}

func seedCatalog(ctx context.Context, conn *sql.Conn, c domain.Catalog) error {
	if invalid(c.ApplicationID) || c.Permissions == nil || c.Scopes == nil {
		return domain.ErrMalformed
	}
	// The catalog's own state is a record now. Generation starts at zero: a
	// seeded catalog has been written once, as a whole, and nothing has read it
	// to compare against.
	if err := writeCatalogState(ctx, conn, c.ApplicationID, catalogPayload{CompatibilityEnabled: c.CompatibilityEnabled, Generation: c.Generation}); err != nil {
		return err
	}
	permissionIDs := sortedKeys(c.Permissions)
	for _, id := range permissionIDs {
		d := c.Permissions[id]
		if invalid(id) || d.ID != id {
			return domain.ErrMalformed
		}
		if err := insertPermissionRecord(ctx, conn, c.ApplicationID, d); err != nil {
			return err
		}
	}
	scopeKeys := sortedKeys(c.Scopes)
	for _, key := range scopeKeys {
		d := c.Scopes[key]
		if invalid(key) || d.Key != key {
			return domain.ErrMalformed
		}
		if err := insertScopeRecord(ctx, conn, c.ApplicationID, d); err != nil {
			return err
		}
	}
	return nil
}

func seedArea(ctx context.Context, conn *sql.Conn, s storage.Snapshot) error {
	// Installation is the registry's fact and is no longer seeded here. A
	// fixture's caller composes a registry that agrees the area exists — see
	// lab.FixedRegistry.
	tenant := s.Area.TenantID()
	for _, key := range sortedRoleKeys(s.Roles) {
		r := s.Roles[key]
		if !codec.ValidRoleID(r.ID) || r.ID != key.ID || r.Revision != key.Revision ||
			invalid(r.Name) || codec.PermissionList(r.Permissions) != nil {
			return domain.ErrMalformed
		}
		// Through the same writer the contract uses: a fixture that wrote rows
		// its own way could seed a shape the contract cannot produce.
		tenant := s.Area.TenantID()
		if r.Managed == domain.ApplicationManaged {
			tenant = ""
		}
		if err := insertRole(ctx, conn, s.Area.ApplicationID(), tenant, r); err != nil {
			return err
		}
	}
	// The trusted-root marker is a field on the grant head rather than its own
	// table, so the set has to be known before any head is written.
	roots := map[string]bool{}
	for _, id := range sortedKeys(s.TrustedRoots) {
		if !s.TrustedRoots[id] || invalid(id) {
			return domain.ErrMalformed
		}
		found := false
		for key := range s.Contents {
			if key.ID == id {
				found = true
				break
			}
		}
		if !found {
			return domain.ErrMalformed
		}
		roots[id] = true
	}
	for _, id := range sortedKeys(s.Controls) {
		c := s.Controls[id]
		if invalid(id) || c.ID != id || c.Version != "1" || (c.Status != "enabled" && c.Status != "disabled") {
			return domain.ErrMalformed
		}
		if err := insertGrantHead(ctx, conn, s.Area, domain.Grant{ID: id, Status: c.Status, TrustedRoot: roots[id]}); err != nil {
			return err
		}
	}
	for _, key := range sortedGrantKeys(s.Contents) {
		g := s.Contents[key]
		if err := codec.ValidateContent(g); err != nil {
			return err
		}
		if g.GrantID != key.ID || g.Revision != key.Revision {
			return domain.ErrMalformed
		}
		raw, _ := json.Marshal(g)
		decoded, err := codec.DecodeContent(raw)
		canonical, marshalErr := json.Marshal(decoded)
		if err != nil || marshalErr != nil || !bytes.Equal(canonical, raw) {
			return domain.ErrMalformed
		}
		if err = insertGrantRevisionRow(ctx, conn, s.Area, decoded); err != nil {
			return err
		}
	}
	for _, id := range sortedKeys(s.Teams) {
		team := s.Teams[id]
		if invalid(id) || team.ID != id || (team.ParentID != "" && invalid(team.ParentID)) {
			return domain.ErrMalformed
		}
		// A tenant's teams exist once, not once per application. A fixture that
		// seeds two snapshots for the same tenant — one per application — names
		// the same teams twice, and the second naming is the same team rather
		// than a conflict.
		if err := seedTeam(ctx, conn, tenant, team); err != nil {
			return err
		}
	}
	for _, m := range s.Memberships {
		if invalid(m.TeamID) || invalid(m.HumanID) {
			return domain.ErrMalformed
		}
		if err := seedMembership(ctx, conn, tenant, m); err != nil {
			return err
		}
	}
	for _, id := range sortedKeys(s.Assignments) {
		a := s.Assignments[id]
		if a.ID != id {
			return domain.ErrMalformed
		}
		raw, err := json.Marshal(a)
		if err != nil {
			return domain.ErrMalformed
		}
		decoded, err := codec.DecodeAssignment(raw)
		if err != nil || decoded != a {
			return domain.ErrMalformed
		}
		if a.Recipient.Type == "group" {
			if _, ok := s.Teams[a.Recipient.ID]; !ok {
				return domain.ErrMalformed
			}
		}
		if _, ok := s.Contents[domain.GrantKey{ID: a.GrantID, Revision: a.GrantRevision}]; !ok {
			return domain.ErrMalformed
		}
		if err = insertAssignment(ctx, conn, s.Area, a); err != nil {
			return err
		}
	}
	return nil
}

func invalid(value string) bool { return strings.TrimSpace(value) == "" || !utf8.ValidString(value) }
func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
func sortedGrantKeys(values map[domain.GrantKey]domain.GrantContent) []domain.GrantKey {
	keys := make([]domain.GrantKey, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].ID == keys[j].ID {
			return keys[i].Revision < keys[j].Revision
		}
		return keys[i].ID < keys[j].ID
	})
	return keys
}
func sortedRoleKeys(values map[domain.RoleKey]domain.RoleContent) []domain.RoleKey {
	keys := make([]domain.RoleKey, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].ID == keys[j].ID {
			return keys[i].Revision < keys[j].Revision
		}
		return keys[i].ID < keys[j].ID
	})
	return keys
}

// seededRegistry answers for the areas a fixture was given. A fixture is a
// closed world — it seeds what it seeds — so agreeing with itself is the honest
// answer, and disagreeing about anything else is the honest refusal.
type fixtureRegistry struct {
	applications map[string]bool
	installed    map[[2]string]bool
}

func seededRegistry(snapshots []storage.Snapshot) *fixtureRegistry {
	r := &fixtureRegistry{applications: map[string]bool{}, installed: map[[2]string]bool{}}
	for _, s := range snapshots {
		r.applications[s.Area.ApplicationID()] = true
		r.installed[[2]string{s.Area.TenantID(), s.Area.ApplicationID()}] = true
	}
	return r
}

func (r *fixtureRegistry) ApplicationExists(ctx context.Context, applicationID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return r.applications[applicationID], nil
}

func (r *fixtureRegistry) Installed(ctx context.Context, tenantID, applicationID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return r.installed[[2]string{tenantID, applicationID}], nil
}
