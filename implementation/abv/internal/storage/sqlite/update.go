package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

type transactionConnection interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	Raw(func(any) error) error
}

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
		categories := 0
		if len(writes.NewAssignments) != 0 {
			categories++
		}
		if writes.GrantStatusChange != nil {
			categories++
		}
		if writes.AssignmentStatusChange != nil {
			categories++
		}
		if writes.NewRoleRevision != nil {
			categories++
		}
		if writes.NewGrantRevision != nil {
			categories++
		}
		if writes.NewGrant != nil {
			categories++
		}
		if writes.RemovedGrant != "" {
			categories++
		}
		for _, set := range []bool{writes.NewTeam != nil, writes.TeamParent != nil, writes.RemovedTeam != "",
			writes.AddedMembership != nil, writes.RemovedMembership != nil} {
			if set {
				categories++
			}
		}
		if categories > 1 {
			return domain.ErrMalformed
		}
		// Team and membership writes. Validation already ran against the
		// authoritative snapshot inside this transaction, so each of these is one
		// row and nothing else: no cascade, and no second record touched.
		if writes.NewTeam != nil {
			return insertTeam(ctx, conn, area.TenantID(), *writes.NewTeam)
		}
		if writes.TeamParent != nil {
			return updateTeamParent(ctx, conn, area.TenantID(), *writes.TeamParent)
		}
		if writes.RemovedTeam != "" {
			return deleteTeam(ctx, conn, area.TenantID(), writes.RemovedTeam)
		}
		if writes.AddedMembership != nil {
			return insertMembership(ctx, conn, area.TenantID(), *writes.AddedMembership)
		}
		if writes.RemovedMembership != nil {
			return deleteMembership(ctx, conn, area.TenantID(), *writes.RemovedMembership)
		}
		if writes.GrantStatusChange != nil {
			return p.writeGrantStatus(ctx, conn, area, s, *writes.GrantStatusChange)
		}
		if writes.AssignmentStatusChange != nil {
			return p.writeAssignmentStatus(ctx, conn, area, *writes.AssignmentStatusChange)
		}
		if writes.NewRoleRevision != nil {
			authoritative, err := p.readCatalog(ctx, conn, area.ApplicationID())
			if err != nil {
				return err
			}
			if err := validation.CheckRolePublication(area, authoritative, *writes.NewRoleRevision); err != nil {
				return err
			}
			if err := p.insertRole(ctx, conn, area, *writes.NewRoleRevision); err != nil {
				return err
			}
			// In the same transaction as the write, so a reader that sees one
			// generation before an offset walk and the same after knows nothing
			// moved between its pages. Without this a role listing would report a
			// generation that never changes, and the guarantee would be a lie.
			//
			// The counter is per application while a role is per tenant, so one
			// tenant's publication invalidates every tenant's cached view of that
			// application. Over-invalidation is wrong in the cheap direction —
			// a retry — where the alternative is a missed row. Narrowing it waits
			// on whether a role is tenant- or application-scoped.
			return bumpGeneration(ctx, conn, area.ApplicationID())
		}
		if writes.NewGrantRevision != nil {
			return p.insertGrantRevision(ctx, conn, area, *writes.NewGrantRevision)
		}
		if writes.NewGrant != nil {
			return p.insertGrant(ctx, conn, area, *writes.NewGrant)
		}
		if writes.RemovedGrant != "" {
			return deleteGrant(ctx, conn, area, writes.RemovedGrant)
		}
		return p.writeAssignments(ctx, conn, area, writes.NewAssignments)
	})
}

func (p *provider) writeAssignmentStatus(ctx context.Context, conn *sql.Conn, area domain.Area, change storage.AssignmentStatusChange) error {
	// Both sides must be records this store could have written. encodeAssignment
	// is the same check the insert path runs, so a status change cannot smuggle
	// in a shape creation would have refused.
	for _, side := range []domain.Assignment{change.Before, change.After} {
		if _, err := encodeAssignment(side); err != nil {
			return err
		}
	}
	before, after := change.Before, change.After
	before.Status = after.Status
	if before != after {
		return domain.ErrMalformed
	}
	// The caller names an assignment by id, which the key path does not carry —
	// so this is the scan the layout trades for Q-104 being structural. It reads
	// the whole record back and requires it to equal what the caller saw, so a
	// concurrent change to any field, not only the status, loses.
	current, err := readAssignmentByID(ctx, conn, area, change.Before.ID)
	if err != nil {
		return err
	}
	if current != change.Before {
		return domain.ErrConflict
	}
	return updateAssignmentValue(ctx, conn, area, change.After)
}

func (p *provider) writeGrantStatus(ctx context.Context, conn *sql.Conn, area domain.Area, snapshot storage.Snapshot, change storage.GrantStatusChange) error {
	before, after := change.Before, change.After
	for _, control := range []domain.GrantControl{before, after} {
		if control.Version == "" || !utf8.ValidString(control.Version) || control.ID == "" || strings.TrimSpace(control.ID) == "" || strings.Contains(control.ID, "*") || !utf8.ValidString(control.ID) {
			return domain.ErrMalformed
		}
		if control.Version != "1" {
			return domain.ErrUnsupported
		}
		if control.Status != "enabled" && control.Status != "disabled" {
			return domain.ErrMalformed
		}
	}
	if before.ID != after.ID {
		return domain.ErrMalformed
	}
	current, ok := snapshot.Controls[before.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if current != before {
		return domain.ErrConflict
	}
	// The compare-and-set is on the stored status, which now lives inside the
	// head's value rather than in a column of its own. json_extract keeps the
	// conditional update one statement, so a concurrent change still loses.
	head, err := readGrantHead(ctx, conn, area, before.ID)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(grantHeadPayload{Status: after.Status, TrustedRoot: head.TrustedRoot})
	if err != nil {
		return domain.ErrMalformed
	}
	result, err := conn.ExecContext(ctx, `
		UPDATE abv_l1_records SET value=?
		 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='grant'
		   AND key3=? AND key4=? AND json_extract(value,'$.status')=?`,
		string(raw), area.TenantID(), area.ApplicationID(), before.ID, before.Status)
	if err != nil {
		return classify(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return classify(err)
	}
	if rows != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (p *provider) transaction(ctx context.Context, conn transactionConnection, begin string, body func() error) (err error) {
	return runTransaction(ctx, conn, begin, body)
}

func runTransaction(ctx context.Context, conn transactionConnection, begin string, body func() error) (err error) {
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
	if _, err = conn.ExecContext(ctx, begin); err != nil {
		classified := classify(err)
		// BEGIN's result is uncertain, so always attempt rollback. cleanup marks
		// the connection bad when rollback cannot establish cleanliness; the
		// operation still reports the original cancellation/conflict category.
		_ = cleanup()
		return classified
	}
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
		// The adopted revision used to be a foreign key. One table holding every
		// record type cannot carry that constraint, so the check is explicit
		// here — the same move the registry fold made for installations.
		slot, slotErr := codec.RenderRevision(a.GrantRevision)
		if slotErr != nil {
			return slotErr
		}
		err = conn.QueryRowContext(ctx, `
			SELECT 1 FROM abv_l1_records
			 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='grant_revision'
			   AND key3=? AND key4=? AND key5=?`,
			area.TenantID(), area.ApplicationID(), a.GrantID, slot).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrRejected
		}
		if err != nil {
			return classify(err)
		}
		if a.Recipient.Type == "group" {
			// A team is tenant-scoped and carries no application, so this asks
			// the record store rather than a table, and does not name one.
			err = conn.QueryRowContext(ctx, `
				SELECT 1 FROM abv_l1_records
				 WHERE boundary='tenant' AND tenant_id=? AND key1='abv' AND key2='team' AND key3=?`,
				area.TenantID(), a.Recipient.ID).Scan(&exists)
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrRejected
			}
			if err != nil {
				return classify(err)
			}
		}
		// Two separate questions now, because the key path answers only one of
		// them. Q-104's duplicate binding is a primary-key collision the insert
		// would refuse anyway; checking it here turns that into ErrConflict
		// before anything is written, and keeps the message honest.
		if _, err := readAssignmentByBinding(ctx, conn, area, a.GrantID, a.Recipient); err == nil {
			return domain.ErrConflict
		} else if !errors.Is(err, domain.ErrNotFound) {
			return err
		}
		// A repeated id is not a Q-104 duplicate — it is two different bindings
		// claiming one handle, which the key path cannot forbid because the id
		// lives in the value. It is refused here instead.
		if _, err := readAssignmentByID(ctx, conn, area, a.ID); err == nil {
			return domain.ErrConflict
		} else if !errors.Is(err, domain.ErrNotFound) {
			return err
		}
		ready = append(ready, prepared{assignment: a, canonical: raw})
	}
	for _, item := range ready {
		a := item.assignment
		if err := insertAssignment(ctx, conn, area, a); err != nil {
			return err
		}
	}
	return nil
}

var _ storage.Provider = (*provider)(nil)
