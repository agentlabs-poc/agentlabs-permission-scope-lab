package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
		if (writes.Permission == nil) == (writes.Scope == nil) || writes.Permission == nil && len(writes.SupportedKeys) != 0 {
			return domain.ErrMalformed
		}
		authoritative, err := p.readCatalog(ctx, conn, app.ID())
		if err != nil {
			return err
		}
		if writes.Permission != nil {
			if err := validation.CheckPermissionRegistration(authoritative, *writes.Permission, writes.SupportedKeys); err != nil {
				return err
			}
			return p.insertPermission(ctx, conn, app.ID(), *writes.Permission, writes.SupportedKeys)
		}
		if len(writes.SupportedKeys) != 0 {
			return domain.ErrMalformed
		}
		if err := validation.CheckScopeRegistration(authoritative, *writes.Scope); err != nil {
			return err
		}
		return p.insertScope(ctx, conn, app.ID(), *writes.Scope)
	})
}

func (p *provider) readCatalog(ctx context.Context, conn *sql.Conn, applicationID string) (domain.Catalog, error) {
	var exists int
	if err := conn.QueryRowContext(ctx, `SELECT 1 FROM applications WHERE application_id=?`, applicationID).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return domain.Catalog{}, domain.ErrNotFound
	} else if err != nil {
		return domain.Catalog{}, classify(err)
	}
	r := snapshotReader{conn: conn, ctx: ctx, limit: p.maxSnapshotRecords}
	var catalog domain.Catalog
	if err := r.catalog(applicationID, &catalog); err != nil {
		return domain.Catalog{}, err
	}
	return catalog, nil
}

func (p *provider) insertPermission(ctx context.Context, conn *sql.Conn, applicationID string, definition domain.PermissionDefinition, keys []string) error {
	if _, err := conn.ExecContext(ctx, `INSERT INTO permissions(application_id,permission_id,active) VALUES(?,?,1)`, applicationID, definition.ID); err != nil {
		return classify(err)
	}
	for ordinal, key := range keys {
		if _, err := conn.ExecContext(ctx, `INSERT INTO supported_scope_keys(application_id,permission_id,scope_key,ordinal) VALUES(?,?,?,?)`, applicationID, definition.ID, key, ordinal); err != nil {
			return classify(err)
		}
	}
	return nil
}

func (p *provider) insertScope(ctx context.Context, conn *sql.Conn, applicationID string, definition domain.ScopeDefinition) error {
	raw, err := json.Marshal(definition.AllowedTokens)
	if err != nil {
		return domain.ErrMalformed
	}
	_, err = conn.ExecContext(ctx, `INSERT INTO scope_definitions(application_id,scope_key,allowed_tokens_json) VALUES(?,?,?)`, applicationID, definition.Key, raw)
	return classify(err)
}
