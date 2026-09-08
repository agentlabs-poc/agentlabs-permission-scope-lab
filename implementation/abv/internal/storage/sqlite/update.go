package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

func (p *provider) Read(ctx context.Context, area domain.Area, callback func(storage.Snapshot) error) (err error) {
	if err = area.Validate(); err != nil {
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
		s, err := p.snapshot(ctx, conn, area)
		if err != nil {
			return err
		}
		return callback(s)
	})
}

func (p *provider) Update(ctx context.Context, area domain.Area, callback func(storage.Snapshot) (storage.WriteSet, error)) (err error) {
	if err = area.Validate(); err != nil {
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
		s, err := p.snapshot(ctx, conn, area)
		if err != nil {
			return err
		}
		writes, err := callback(s)
		if err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		return p.writeAssignments(ctx, conn, area, writes.NewAssignments)
	})
}

func (p *provider) transaction(ctx context.Context, conn *sql.Conn, begin string, body func() error) (err error) {
	if _, err = conn.ExecContext(ctx, begin); err != nil {
		return classify(err)
	}
	active := true
	cleanup := func() error {
		if !active {
			return nil
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, rollbackErr := conn.ExecContext(cleanupCtx, "ROLLBACK")
		active = false
		if rollbackErr != nil {
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
			return classify(rollbackErr)
		}
		return nil
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = cleanup()
			panic(recovered)
		}
	}()
	if err = body(); err != nil {
		if rollbackErr := cleanup(); rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}
	if err = ctx.Err(); err != nil {
		if rollbackErr := cleanup(); rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}
	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		classified := classify(err)
		if rollbackErr := cleanup(); rollbackErr != nil {
			return errors.Join(classified, rollbackErr)
		}
		return classified
	}
	active = false
	return nil
}

func (p *provider) writeAssignments(ctx context.Context, conn *sql.Conn, area domain.Area, assignments []domain.Assignment) error {
	type prepared struct {
		assignment domain.Assignment
		canonical  []byte
	}
	type recipientKey struct{ grantID, recipientType, recipientID string }
	ready := make([]prepared, 0, len(assignments))
	ids := make(map[string]bool, len(assignments))
	recipients := make(map[recipientKey]bool, len(assignments))
	for _, a := range assignments {
		if err := ctx.Err(); err != nil {
			return err
		}
		raw, err := json.Marshal(a)
		if err != nil {
			return domain.ErrMalformed
		}
		decoded, err := codec.DecodeAssignment(raw)
		if err != nil {
			return err
		}
		if decoded != a {
			return domain.ErrMalformed
		}
		key := recipientKey{grantID: a.GrantID, recipientType: a.Recipient.Type, recipientID: a.Recipient.ID}
		if ids[a.ID] || recipients[key] {
			return domain.ErrConflict
		}
		ids[a.ID], recipients[key] = true, true
		var exists int
		err = conn.QueryRowContext(ctx, `SELECT 1 FROM grant_contents WHERE tenant_id=? AND application_id=? AND grant_id=? AND revision=?`, area.TenantID(), area.ApplicationID(), a.GrantID, a.GrantRevision).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrRejected
		}
		if err != nil {
			return classify(err)
		}
		if a.Recipient.Type == "group" {
			err = conn.QueryRowContext(ctx, `SELECT 1 FROM teams WHERE tenant_id=? AND application_id=? AND team_id=?`, area.TenantID(), area.ApplicationID(), a.Recipient.ID).Scan(&exists)
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrRejected
			}
			if err != nil {
				return classify(err)
			}
		}
		err = conn.QueryRowContext(ctx, `SELECT 1 FROM assignments WHERE tenant_id=? AND application_id=? AND (assignment_id=? OR (grant_id=? AND recipient_type=? AND recipient_id=?)) LIMIT 1`, area.TenantID(), area.ApplicationID(), a.ID, a.GrantID, a.Recipient.Type, a.Recipient.ID).Scan(&exists)
		if err == nil {
			return domain.ErrConflict
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return classify(err)
		}
		ready = append(ready, prepared{assignment: a, canonical: raw})
	}
	for _, item := range ready {
		a := item.assignment
		if _, err := conn.ExecContext(ctx, `INSERT INTO assignments(tenant_id,application_id,assignment_id,grant_id,grant_revision,recipient_type,recipient_id,status,canonical_json) VALUES(?,?,?,?,?,?,?,?,?)`, area.TenantID(), area.ApplicationID(), a.ID, a.GrantID, a.GrantRevision, a.Recipient.Type, a.Recipient.ID, a.Status, item.canonical); err != nil {
			return classify(err)
		}
	}
	return nil
}

var _ storage.Provider = (*provider)(nil)
