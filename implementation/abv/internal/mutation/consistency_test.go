package mutation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestParentDisablementBeforeMutationAcquisitionConflictsThenIsSeen(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	path := t.TempDir() + "/authority.db"
	provider, err := lab.CreateSQLite(t.Context(), path, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	baseAdmin, _ := lab.NewAdministration(area, fixture.Administration)
	admin := &countingAdministration{next: baseAdmin}
	service, _ := mutation.New(provider, admin, &fixedClock{now: time.Now()})

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ready := make(chan error, 1)
	release := make(chan struct{})
	done := make(chan error, 1)
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	go func() {
		conn, err := db.Conn(t.Context())
		if err != nil {
			ready <- err
			return
		}
		defer conn.Close()
		if _, err = conn.ExecContext(t.Context(), "BEGIN IMMEDIATE"); err == nil {
			control := fixture.Snapshot.Controls["G1"]
			control.Status = "disabled"
			raw, _ := json.Marshal(control)
			_, err = conn.ExecContext(t.Context(), `UPDATE grant_controls SET status=?, canonical_json=? WHERE tenant_id=? AND application_id=? AND grant_id=?`, control.Status, raw, area.TenantID(), area.ApplicationID(), control.ID)
		}
		ready <- err
		if err != nil {
			return
		}
		<-release
		_, err = conn.ExecContext(t.Context(), "COMMIT")
		if err != nil {
			_, _ = conn.ExecContext(context.Background(), "ROLLBACK")
		}
		done <- err
	}()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	select {
	case err = <-ready:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	receipt, err := service.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed)
	if !errors.Is(err, domain.ErrConflict) || receipt != (domain.Receipt{}) {
		t.Fatalf("locked attempt receipt=%#v err=%v", receipt, err)
	}
	if got := admin.calls.Load(); got != 0 {
		t.Fatalf("callback ran %d times before mutation acquisition", got)
	}
	close(release)
	select {
	case err = <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	receipt, err = service.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed)
	if !errors.Is(err, domain.ErrRejected) || receipt != (domain.Receipt{}) {
		t.Fatalf("fresh attempt missed committed disablement: receipt=%#v err=%v", receipt, err)
	}
	if got := admin.calls.Load(); got != 1 {
		t.Fatalf("fresh callback count=%d, want 1", got)
	}
	assertAssignmentCount(t, provider, area, 2)
}

func TestParentDisablementAfterAssignmentCommitIsNotRetroactive(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	path := t.TempDir() + "/authority.db"
	provider, err := lab.CreateSQLite(t.Context(), path, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	admin, _ := lab.NewAdministration(area, fixture.Administration)
	service, _ := mutation.New(provider, admin, &fixedClock{now: time.Now()})

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	committed := make(chan domain.Receipt, 1)
	disabled := make(chan error, 1)
	go func() {
		receipt, err := service.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed)
		if err != nil {
			disabled <- err
			return
		}
		committed <- receipt
		control := fixture.Snapshot.Controls["G1"]
		control.Status = "disabled"
		raw, _ := json.Marshal(control)
		_, err = db.ExecContext(t.Context(), `UPDATE grant_controls SET status=?, canonical_json=? WHERE tenant_id=? AND application_id=? AND grant_id=?`, control.Status, raw, area.TenantID(), area.ApplicationID(), control.ID)
		disabled <- err
	}()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	select {
	case receipt := <-committed:
		if receipt.AssignmentID != fixture.Proposed.ID {
			t.Fatalf("receipt=%#v", receipt)
		}
	case err := <-disabled:
		t.Fatalf("assignment did not commit before disablement: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	select {
	case err := <-disabled:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	assertAssignmentCount(t, provider, area, 3)
}

func TestCancellationDoesNotReplayMutationCallback(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	baseAdmin, _ := lab.NewAdministration(area, fixture.Administration)
	ctx, cancel := context.WithCancel(t.Context())
	admin := &countingAdministration{next: baseAdmin, cancel: cancel}
	service, _ := mutation.New(provider, admin, &fixedClock{now: time.Now()})

	receipt, err := service.CreateAssignment(ctx, area, fixture.Issuer, fixture.Proposed)
	if !errors.Is(err, context.Canceled) || receipt != (domain.Receipt{}) {
		t.Fatalf("cancelled attempt receipt=%#v err=%v", receipt, err)
	}
	if got := admin.calls.Load(); got != 1 {
		t.Fatalf("callback count=%d, want 1", got)
	}
	assertAssignmentCount(t, provider, area, 2)
}

type countingAdministration struct {
	next   mutation.Administration
	cancel context.CancelFunc
	calls  atomic.Int32
}

func (a *countingAdministration) CheckAssignment(ctx context.Context, snapshot storage.Snapshot, identity domain.Identity, proposed domain.Assignment, now time.Time) error {
	a.calls.Add(1)
	if a.cancel != nil {
		a.cancel()
	}
	return a.next.CheckAssignment(ctx, snapshot, identity, proposed, now)
}
