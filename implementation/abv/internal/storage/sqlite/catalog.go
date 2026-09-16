package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"database/sql"
	"encoding/json"
	"strings"
)

func (p *provider) ReadCatalog(ctx context.Context, app domain.Application, callback func(domain.Catalog) error) (err error) {
	if err = app.Validate(); err != nil {
		return err
	}
	if callback == nil {
		return domain.ErrMalformed
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	conn, err := p.connection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	return p.transaction(ctx, conn, "BEGIN", func() error {
		catalog, err := p.readCatalog(ctx, conn, app.ID())
		if err != nil {
			return err
		}
		return callback(catalog)
	})
}

// UpdatePlatformCatalog writes at the platform boundary, in one transaction,
// without requiring an applications row for the namespace — a platform namespace
// is reserved precisely so it can never be an application.
func (p *provider) UpdatePlatformCatalog(ctx context.Context, namespace string, callback func() (storage.CatalogWriteSet, error)) (err error) {
	if strings.TrimSpace(namespace) == "" || callback == nil {
		return domain.ErrMalformed
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	conn, err := p.connection(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil && err == nil {
			err = classify(closeErr)
		}
	}()
	return p.transaction(ctx, conn, "BEGIN IMMEDIATE", func() error {
		writes, err := callback()
		if err != nil {
			return err
		}
		if writes.PlatformPermission == nil {
			return domain.ErrMalformed
		}
		if err := insertPermissionAt(ctx, conn, domain.PlatformBoundary, namespace, *writes.PlatformPermission); err != nil {
			return err
		}
		// A platform permission is in every application's catalog, so a write to
		// it invalidates every cached view. Bumping all of them keeps the
		// read-page-reread guarantee true rather than nearly true.
		return bumpEveryCatalogGeneration(ctx, conn)
	})
}

func (p *provider) UpdateCatalog(ctx context.Context, app domain.Application, callback func(domain.Catalog) (storage.CatalogWriteSet, error)) (err error) {
	if err = app.Validate(); err != nil {
		return err
	}
	if callback == nil {
		return domain.ErrMalformed
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	conn, err := p.connection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	return p.transaction(ctx, conn, "BEGIN IMMEDIATE", func() error {
		catalog, err := p.readCatalog(ctx, conn, app.ID())
		if err != nil {
			return err
		}
		writes, err := callback(catalog)
		if err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		// Exactly one write kind per call.
		kinds := 0
		for _, set := range []bool{writes.Permission != nil, writes.Scope != nil, writes.PermissionStatus != nil, writes.ApplicationRole != nil, writes.PlatformPermission != nil} {
			if set {
				kinds++
			}
		}
		if kinds != 1 {
			return domain.ErrMalformed
		}
		authoritative, err := p.readCatalog(ctx, conn, app.ID())
		if err != nil {
			return err
		}
		if err := bumpCatalogGeneration(ctx, conn, app.ID()); err != nil {
			return err
		}
		if writes.Permission != nil {
			if err := validation.CheckPermissionRegistration(authoritative, *writes.Permission); err != nil {
				return err
			}
			return insertPermissionRecord(ctx, conn, app.ID(), *writes.Permission)
		}
		if writes.PlatformPermission != nil {
			if err := validation.CheckPermissionRegistrationAt(domain.PlatformBoundary, authoritative, *writes.PlatformPermission); err != nil {
				return err
			}
			return insertPermissionAt(ctx, conn, domain.PlatformBoundary, writes.PlatformNamespace, *writes.PlatformPermission)
		}
		if writes.ApplicationRole != nil {
			role := *writes.ApplicationRole
			if err := validation.CheckApplicationRolePublication(authoritative, role); err != nil {
				return err
			}
			// The same id rule the tenant path holds: an issued id must be new,
			// and a supplied one must already name an application role, which
			// makes the publication a new revision of it.
			existing, err := applicationRoleRevisions(ctx, conn, app.ID(), role.ID)
			if err != nil {
				return err
			}
			if writes.ApplicationRoleIssued && len(existing) > 0 {
				return domain.ErrConflict
			}
			if !writes.ApplicationRoleIssued && len(existing) == 0 {
				return domain.ErrNotFound
			}
			slot, err := codec.RenderRevision(role.Revision)
			if err != nil {
				return err
			}
			for _, held := range existing {
				if held == slot {
					return domain.ErrConflict
				}
			}
			return insertRole(ctx, conn, app.ID(), "", role)
		}
		if writes.PermissionStatus != nil {
			if err := validation.CheckPermissionStatus(authoritative, writes.PermissionStatus.ID, writes.PermissionStatus.Active); err != nil {
				return err
			}
			return p.updatePermissionStatus(ctx, conn, app.ID(), *writes.PermissionStatus)
		}
		if err := validation.CheckScopeRegistration(authoritative, *writes.Scope); err != nil {
			return err
		}
		return insertScopeRecord(ctx, conn, app.ID(), *writes.Scope)
	})
}

func (p *provider) readCatalog(ctx context.Context, conn *sql.Conn, applicationID string) (domain.Catalog, error) {
	// Does this application exist? The application registry domain owns that
	// fact, and now owns it alone — there is no local table to fall back to, so
	// a provider opened without a registry can answer nothing. Open refuses that
	// rather than letting it fail one read at a time.
	exists, err := p.registry.ApplicationExists(ctx, applicationID)
	if err != nil {
		return domain.Catalog{}, err
	}
	if !exists {
		return domain.Catalog{}, domain.ErrNotFound
	}
	r := snapshotReader{conn: conn, ctx: ctx, limit: p.maxSnapshotRecords}
	var catalog domain.Catalog
	if err := r.catalog(applicationID, &catalog); err != nil {
		return domain.Catalog{}, err
	}
	return catalog, nil
}

// updatePermissionStatus flips an existing definition's active flag. It never
// inserts: the identifier must already be registered, which the caller has
// validated against the authoritative catalog read in the same transaction.
func (p *provider) updatePermissionStatus(ctx context.Context, conn *sql.Conn, applicationID string, definition domain.PermissionDefinition) error {
	key, err := codec.ParsePermission(definition.ID)
	if err != nil {
		return err
	}
	slots := key.Slots()
	payload, err := permissionPayload(definition.Active)
	if err != nil {
		return err
	}
	result, err := conn.ExecContext(ctx, `
		UPDATE abv_l1_records SET value=?
		 WHERE boundary='application' AND tenant_id='' AND key1='abv' AND key2='permission' AND key3=?
		   AND key4=? AND key5=? AND key6=? AND key7=? AND key8=? AND key9=? AND key10=?`,
		payload, applicationID,
		slots[0], slots[1], slots[2], slots[3], slots[4], slots[5], slots[6])
	if err != nil {
		return classify(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return classify(err)
	}
	if affected != 1 {
		return domain.ErrRejected
	}
	return nil
}

// insertPermissionRecord writes a permission as an L1 record. The identifier is
// decomposed by the shared codec — storage never splits the string itself, so
// the two layers cannot disagree about what a row means.
func insertPermissionRecord(ctx context.Context, conn *sql.Conn, applicationID string, definition domain.PermissionDefinition) error {
	return insertPermissionAt(ctx, conn, domain.ApplicationBoundary, applicationID, definition)
}

// insertPermissionAt writes a permission at a named boundary. key3 holds the
// namespace: the application at the application boundary, whatever the platform
// defines at the platform boundary.
func insertPermissionAt(ctx context.Context, conn *sql.Conn, boundary domain.Boundary, namespace string, definition domain.PermissionDefinition) error {
	key, err := codec.ParsePermission(definition.ID)
	if err != nil {
		return err
	}
	slots := key.Slots()
	payload, err := permissionPayload(definition.Active)
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, `
		INSERT INTO abv_l1_records
		  (boundary, tenant_id, key1, key2, key3,
		   key4, key5, key6, key7, key8, key9, key10, value)
		VALUES (?, '', 'abv', 'permission', ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		string(boundary), namespace,
		slots[0], slots[1], slots[2], slots[3], slots[4], slots[5], slots[6],
		payload)
	if err != nil {
		return classify(err)
	}
	return nil
}

func permissionPayload(active bool) ([]byte, error) {
	raw, err := json.Marshal(struct {
		Active bool `json:"active"`
	}{Active: active})
	if err != nil {
		return nil, domain.ErrMalformed
	}
	return raw, nil
}

// insertScopeRecord writes a scope as an L1 record. key3 is the application,
// key4 the scope key: a scope key is a flat token, so it needs no decomposition.
func insertScopeRecord(ctx context.Context, conn *sql.Conn, applicationID string, definition domain.ScopeDefinition) error {
	// The payload is empty: a scope record's presence is the fact. $self is a
	// reserved token the evaluator knows, not something a key declares.
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO abv_l1_records
		  (boundary, tenant_id, key1, key2, key3, key4, value)
		VALUES ('application', '', 'abv', 'scope', ?, ?, '{}')`,
		applicationID, definition.Key); err != nil {
		return classify(err)
	}
	return nil
}

// applicationRoleRevisions lists the revision slots an application role already
// holds. An empty result means the id names no application role.
func applicationRoleRevisions(ctx context.Context, conn *sql.Conn, applicationID, id string) ([]string, error) {
	rows, err := conn.QueryContext(ctx, `
		SELECT key5 FROM abv_l1_records
		 WHERE boundary='application' AND tenant_id='' AND key1='abv' AND key2='role' AND key3=? AND key4=?
		 ORDER BY key5`, applicationID, id)
	if err != nil {
		return nil, classify(err)
	}
	defer rows.Close()
	var slots []string
	for rows.Next() {
		var slot string
		if err := rows.Scan(&slot); err != nil {
			return nil, classify(err)
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, classify(err)
	}
	return slots, nil
}
