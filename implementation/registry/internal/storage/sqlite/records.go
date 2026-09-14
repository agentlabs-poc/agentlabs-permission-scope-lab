package sqlite

import (
	"agentlabs.local/registry/domain"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// The canonical key layout for this domain.
//
//	application   key1=application_registry  key2=application   key3=<slug>
//	installation  key1=application_registry  key2=installation  key3=<slug>  + tenant
//
// key1 restates the table name and is carried for consistency with the other
// domains; whether it earns its slot is a question for all of them at once, not
// for this one alone.
const (
	domainKey       = "application_registry"
	applicationKind = "application"
	installKind     = "installation"
)

type applicationValue struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// installationValue carries the tenant-side status. The application's own status
// is platform-wide and lives on the application record; this one is this
// tenant's, and the two answer different questions.
type installationValue struct {
	Status string `json:"status"`
}

// transaction runs fn under an immediate write lock, so a read-validate-write
// sequence inside it is serialised without a version column.
func (s *Store) transaction(ctx context.Context, fn func(*sql.Tx) error) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return classify(err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = classify(tx.Commit())
	}()
	if _, err = tx.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		// SQLite's driver already began one; the explicit statement is a no-op
		// where the driver handles it, and its failure is not fatal here.
		err = nil
	}
	return fn(tx)
}

func (s *Store) InsertApplication(ctx context.Context, app domain.Application) error {
	raw, err := json.Marshal(applicationValue{Name: app.Name, Status: app.Status})
	if err != nil {
		return domain.ErrMalformed
	}
	return s.transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO application_registry_l1_records
			  (boundary, tenant_id, key1, key2, key3, value)
			VALUES ('application', '', ?, ?, ?, ?)`,
			domainKey, applicationKind, app.Slug, string(raw))
		return classify(err)
	})
}

func (s *Store) UpdateApplicationStatus(ctx context.Context, slug, status string) (domain.Application, error) {
	var result domain.Application
	err := s.transaction(ctx, func(tx *sql.Tx) error {
		var raw string
		err := tx.QueryRowContext(ctx, `
			SELECT value FROM application_registry_l1_records
			 WHERE boundary='application' AND tenant_id='' AND key1=? AND key2=? AND key3=?`,
			domainKey, applicationKind, slug).Scan(&raw)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return classify(err)
		}
		var value applicationValue
		if json.Unmarshal([]byte(raw), &value) != nil {
			return domain.ErrMalformed
		}
		value.Status = status
		updated, err := json.Marshal(value)
		if err != nil {
			return domain.ErrMalformed
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE application_registry_l1_records SET value=?
			 WHERE boundary='application' AND tenant_id='' AND key1=? AND key2=? AND key3=?`,
			string(updated), domainKey, applicationKind, slug); err != nil {
			return classify(err)
		}
		result = domain.Application{Slug: slug, Name: value.Name, Status: value.Status}
		return nil
	})
	return result, err
}

func (s *Store) Application(ctx context.Context, slug string) (domain.Application, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `
		SELECT value FROM application_registry_l1_records
		 WHERE boundary='application' AND tenant_id='' AND key1=? AND key2=? AND key3=?`,
		domainKey, applicationKind, slug).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Application{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Application{}, classify(err)
	}
	var value applicationValue
	if json.Unmarshal([]byte(raw), &value) != nil || !domain.ValidStatus(value.Status) {
		return domain.Application{}, domain.ErrMalformed
	}
	return domain.Application{Slug: slug, Name: value.Name, Status: value.Status}, nil
}

func (s *Store) Applications(ctx context.Context) ([]domain.Application, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT key3, value FROM application_registry_l1_records
		 WHERE boundary='application' AND tenant_id='' AND key1=? AND key2=?
		 ORDER BY key3`, domainKey, applicationKind)
	if err != nil {
		return nil, classify(err)
	}
	defer rows.Close()
	var result []domain.Application
	for rows.Next() {
		var slug, raw string
		if err := rows.Scan(&slug, &raw); err != nil {
			return nil, classify(err)
		}
		var value applicationValue
		if json.Unmarshal([]byte(raw), &value) != nil {
			return nil, domain.ErrMalformed
		}
		result = append(result, domain.Application{Slug: slug, Name: value.Name, Status: value.Status})
	}
	return result, classify(rows.Err())
}

func (s *Store) InsertInstallation(ctx context.Context, i domain.Installation) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO application_registry_l1_records
			  (boundary, tenant_id, key1, key2, key3, value)
			VALUES ('tenant', ?, ?, ?, ?, ?)`,
			i.TenantID, domainKey, installKind, i.Slug, `{"status":"`+i.Status+`"}`)
		return classify(err)
	})
}

func (s *Store) DeleteInstallation(ctx context.Context, i domain.Installation) error {
	return s.transaction(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `
			DELETE FROM application_registry_l1_records
			 WHERE boundary='tenant' AND tenant_id=? AND key1=? AND key2=? AND key3=?`,
			i.TenantID, domainKey, installKind, i.Slug)
		if err != nil {
			return classify(err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return classify(err)
		}
		if affected != 1 {
			return domain.ErrNotFound
		}
		return nil
	})
}

// Installation returns one tenant's hold on one application, with its status.
func (s *Store) Installation(ctx context.Context, tenantID, slug string) (domain.Installation, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `
		SELECT value FROM application_registry_l1_records
		 WHERE boundary='tenant' AND tenant_id=? AND key1=? AND key2=? AND key3=?`,
		tenantID, domainKey, installKind, slug).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Installation{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Installation{}, classify(err)
	}
	var value installationValue
	if json.Unmarshal([]byte(raw), &value) != nil || !domain.ValidInstallationStatus(value.Status) {
		return domain.Installation{}, domain.ErrMalformed
	}
	return domain.Installation{TenantID: tenantID, Slug: slug, Status: value.Status}, nil
}

// UpdateInstallationStatus enables or disables one tenant's installation. It
// never creates: a status change must not be a back door around installing.
func (s *Store) UpdateInstallationStatus(ctx context.Context, tenantID, slug, status string) (domain.Installation, error) {
	var result domain.Installation
	err := s.transaction(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			UPDATE application_registry_l1_records SET value=?
			 WHERE boundary='tenant' AND tenant_id=? AND key1=? AND key2=? AND key3=?`,
			`{"status":"`+status+`"}`, tenantID, domainKey, installKind, slug)
		if err != nil {
			return classify(err)
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return classify(err)
		}
		if affected != 1 {
			return domain.ErrNotFound
		}
		result = domain.Installation{TenantID: tenantID, Slug: slug, Status: status}
		return nil
	})
	return result, err
}

// Installations answers in either direction. Exactly one of tenantID and slug is
// non-empty, which the contract has already enforced.
func (s *Store) Installations(ctx context.Context, tenantID, slug string) ([]domain.Installation, error) {
	query := `
		SELECT tenant_id, key3, value FROM application_registry_l1_records
		 WHERE boundary='tenant' AND key1=? AND key2=? AND tenant_id=?
		 ORDER BY key3`
	arg := tenantID
	if tenantID == "" {
		query = `
		SELECT tenant_id, key3, value FROM application_registry_l1_records
		 WHERE boundary='tenant' AND key1=? AND key2=? AND key3=?
		 ORDER BY tenant_id`
		arg = slug
	}
	rows, err := s.db.QueryContext(ctx, query, domainKey, installKind, arg)
	if err != nil {
		return nil, classify(err)
	}
	defer rows.Close()
	var result []domain.Installation
	for rows.Next() {
		var tenant, s2, raw string
		if err := rows.Scan(&tenant, &s2, &raw); err != nil {
			return nil, classify(err)
		}
		var value installationValue
		if json.Unmarshal([]byte(raw), &value) != nil {
			return nil, domain.ErrMalformed
		}
		result = append(result, domain.Installation{TenantID: tenant, Slug: s2, Status: value.Status})
	}
	return result, classify(rows.Err())
}
