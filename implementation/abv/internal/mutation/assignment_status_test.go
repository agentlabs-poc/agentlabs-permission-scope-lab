package mutation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/sqlite"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type assignmentStatusAdministration struct {
	check func(storage.Snapshot, domain.Identity, domain.Assignment) error
}

func (a assignmentStatusAdministration) CheckAssignment(context.Context, storage.Snapshot, domain.Identity, domain.Assignment, time.Time) error {
	return domain.ErrUnsupported
}

func (a assignmentStatusAdministration) CheckAssignmentStatus(_ context.Context, snapshot storage.Snapshot, identity domain.Identity, after domain.Assignment, _ time.Time) error {
	if a.check != nil {
		return a.check(snapshot, identity, after)
	}
	return nil
}

func TestSetAssignmentStatusSQLiteBottomUpAndPreservesRecords(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments["fm5b7t4pan0d"] = fixture.Proposed
	control := fixture.Snapshot.Controls["fk3x9r2man0d"]
	control.Status = "disabled"
	fixture.Snapshot.Controls["fk3x9r2man0d"] = control
	expired := time.Now().Add(-time.Hour)
	content := fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1}]
	content.Validity = &domain.Validity{ExpiresAt: &expired}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1}] = content
	path := t.TempDir() + "/authority.db"
	provider, err := lab.CreateSQLite(t.Context(), path, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	service, _ := mutation.New(provider, assignmentStatusAdministration{}, &fixedClock{now: time.Now()})
	if got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4p5iv8", "disabled"); !errors.Is(err, domain.ErrRejected) || got != (domain.Assignment{}) {
		t.Fatalf("parent disable = %#v, %v", got, err)
	}
	if err = provider.Close(); err != nil {
		t.Fatal(err)
	}
	provider, err = sqlite.Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	if err = provider.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
		if snapshot.Assignments["fm5b7t4p5iv8"].Status != "enabled" {
			t.Fatal("failed parent disable wrote")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	service, _ = mutation.New(provider, assignmentStatusAdministration{}, &fixedClock{now: time.Now()})
	if got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4pan0d", "disabled"); err != nil || got.Status != "disabled" {
		t.Fatalf("leaf disable = %#v, %v", got, err)
	}
	if got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4p5iv8", "disabled"); err != nil || got.Status != "disabled" {
		t.Fatalf("parent after leaf = %#v, %v", got, err)
	}
	if err = provider.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := sqlite.Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err = reopened.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
		if snapshot.Assignments["fm5b7t4p5iv8"].Status != "disabled" || snapshot.Assignments["fm5b7t4pan0d"].Status != "disabled" || snapshot.Controls["fk3x9r2man0d"] != control || len(snapshot.Assignments) != 3 {
			t.Fatalf("unexpected persisted state: %#v", snapshot)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSetAssignmentStatusSQLiteInspectsEveryForkAndDisabledBridge(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	add := func(id, grant, parentGrant, team, parentTeam, status string) {
		fixture.Snapshot.Teams[team] = domain.Team{Name: "team",ID: team, ParentID: parentTeam}
		fixture.Snapshot.Controls[grant] = domain.GrantControl{Version: "1", ID: grant, Status: "enabled"}
		fixture.Snapshot.Contents[domain.GrantKey{ID: grant, Revision: 1}] = domain.GrantContent{Version: "1", GrantID: grant, Revision: 1, ParentGrantID: parentGrant, Permissions: []string{lab.PayslipRead}, Scope: map[string]string{}}
		fixture.Snapshot.Assignments[id] = domain.Assignment{Version: "1", ID: id, GrantID: grant, GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: team}, Status: status}
	}
	add("fm5b7t4pan0d", "fk3x9r2man0d", "fk3x9r2m5iv8", "fibggi2juxhc", "fibggi2juubk", "disabled")
	add("fm5b7t4pfs5i", "fk3x9r2mfs5i", "fk3x9r2man0d", "fidfcosw0iyo", "fibggi2juxhc", "enabled")
	add("fm5b7t4pkxan", "fk3x9r2mkxan", "fk3x9r2m5iv8", "fidfcosw0m4g", "fibggi2juubk", "enabled")
	// Same grant, but no parent-team holding relationship to fm5b7t4p5iv8.
	add("fm5b7t4pp2fs", "fk3x9r2man0d", "fk3x9r2m5iv8", "ficfwlfpxlhc", "fibggi2jur5s", "enabled")
	provider, _ := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	defer provider.Close()
	service, _ := mutation.New(provider, assignmentStatusAdministration{}, &fixedClock{now: time.Now()})
	for _, id := range []string{"fm5b7t4p5iv8", "fm5b7t4pfs5i", "fm5b7t4p5iv8", "fm5b7t4pkxan"} {
		got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, id, "disabled")
		if id == "fm5b7t4p5iv8" {
			if !errors.Is(err, domain.ErrRejected) || got != (domain.Assignment{}) {
				t.Fatalf("fm5b7t4p5iv8 bypassed descendant at step %s: %#v, %v", id, got, err)
			}
		} else if err != nil {
			t.Fatalf("disable %s: %v", id, err)
		}
	}
	if got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4p5iv8", "disabled"); err != nil || got.Status != "disabled" {
		t.Fatalf("unrelated reuse blocked fm5b7t4p5iv8: %#v, %v", got, err)
	}
	if err := provider.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
		if snapshot.Assignments["fm5b7t4pan0d"].Status != "disabled" || snapshot.Assignments["fm5b7t4pp2fs"].Status != "enabled" {
			t.Fatal("status cascade")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSetAssignmentStatusRestoreUsesAdoptedRevisionAndCurrentHolding(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	for _, test := range []struct {
		name   string
		mutate func(*lab.TeamFINC17Case)
		want   error
	}{
		{name: "valid new reality"},
		{name: "disabled grant", mutate: func(f *lab.TeamFINC17Case) {
			c := f.Snapshot.Controls["fk3x9r2man0d"]
			c.Status = "disabled"
			f.Snapshot.Controls["fk3x9r2man0d"] = c
		}, want: domain.ErrRejected},
		{name: "disabled support", mutate: func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			a.Status = "disabled"
			f.Snapshot.Assignments["fm5b7t4p5iv8"] = a
		}, want: domain.ErrRejected},
		{name: "changed parent without holding", mutate: func(f *lab.TeamFINC17Case) {
			f.Snapshot.Teams["fibggi2juxhc"] = domain.Team{ID: "fibggi2juxhc", Name: "fp8h2w6yan0d", ParentID: "fibggi2jur5s"}
		}, want: domain.ErrRejected},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			assignment := fixture.Proposed
			assignment.Status = "disabled"
			fixture.Snapshot.Assignments["fm5b7t4pan0d"] = assignment
			newer := fixture.Child
			newer.Revision = 2
			fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 2}] = newer
			if test.mutate != nil {
				test.mutate(&fixture)
			}
			wantA1 := fixture.Snapshot.Assignments["fm5b7t4p5iv8"].Status
			provider, _ := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
			defer provider.Close()
			service, _ := mutation.New(provider, assignmentStatusAdministration{}, &fixedClock{now: time.Now()})
			got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4pan0d", "enabled")
			if !errors.Is(err, test.want) || test.want != nil && got != (domain.Assignment{}) {
				t.Fatalf("restore = %#v, %v; want %v", got, err, test.want)
			}
			if err := provider.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
				want := "enabled"
				if test.want != nil {
					want = "disabled"
				}
				if snapshot.Assignments["fm5b7t4pan0d"].Status != want || snapshot.Assignments["fm5b7t4pan0d"].GrantRevision != 1 || snapshot.Assignments["fm5b7t4p5iv8"].Status != wantA1 {
					t.Fatalf("state = %#v", snapshot.Assignments)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSetAssignmentStatusGateCancellationConflictAndBoundaries(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	other, _ := domain.NewArea("fi7io4lvkfsw", "hrms")
	fixture := lab.TeamFINC17(area)
	provider := &failingCommitProvider{snapshot: fixture.Snapshot}
	rejecting, _ := mutation.New(provider, assignmentStatusAdministration{check: func(storage.Snapshot, domain.Identity, domain.Assignment) error { return domain.ErrRejected }}, &fixedClock{now: time.Now()})
	if got, err := rejecting.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4p5iv8", "enabled"); !errors.Is(err, domain.ErrRejected) || got != (domain.Assignment{}) {
		t.Fatalf("no-op gate = %#v, %v", got, err)
	}
	service, _ := mutation.New(provider, assignmentStatusAdministration{}, &fixedClock{now: time.Now()})
	for _, request := range []struct {
		area       domain.Area
		id, status string
	}{{other, "fm5b7t4p5iv8", "disabled"}, {area, "A*", "disabled"}, {area, "fm5b7t4p5iv8", "paused"}} {
		if got, err := service.SetAssignmentStatus(t.Context(), request.area, fixture.Issuer, request.id, request.status); err == nil || got != (domain.Assignment{}) {
			t.Fatalf("boundary = %#v, %v", got, err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if got, err := service.SetAssignmentStatus(ctx, area, fixture.Issuer, "fm5b7t4p5iv8", "disabled"); !errors.Is(err, context.Canceled) || got != (domain.Assignment{}) {
		t.Fatalf("cancel = %#v, %v", got, err)
	}
	if got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4p5iv8", "disabled"); !errors.Is(err, domain.ErrConflict) || got != (domain.Assignment{}) || provider.returned.AssignmentStatusChange == nil {
		t.Fatalf("conflict = %#v, %#v, %v", got, provider.returned, err)
	}
	if got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4p0dq3", "disabled"); !errors.Is(err, domain.ErrUnsupported) || got != (domain.Assignment{}) {
		t.Fatalf("trusted root = %#v, %v", got, err)
	}
	var nilAdmin *assignmentStatusAdministration
	if _, err := mutation.New(provider, nilAdmin, &fixedClock{now: time.Now()}); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("typed nil = %v", err)
	}
	userFixture := lab.TeamFINC17(area)
	user := userFixture.Snapshot.Assignments["fm5b7t4p5iv8"]
	user.Recipient = domain.Recipient{Type: "user", ID: "fi7io4lvjqio"}
	userFixture.Snapshot.Assignments["fm5b7t4p5iv8"] = user
	service, _ = mutation.New(&failingCommitProvider{snapshot: userFixture.Snapshot}, assignmentStatusAdministration{}, &fixedClock{now: time.Now()})
	if got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4p5iv8", "disabled"); !errors.Is(err, domain.ErrUnsupported) || got != (domain.Assignment{}) {
		t.Fatalf("direct user = %#v, %v", got, err)
	}
}

func TestSetAssignmentStatusRejectsInvalidUTF8IDBeforeUpdate(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	provider := &failingCommitProvider{snapshot: fixture.Snapshot}
	service, _ := mutation.New(provider, assignmentStatusAdministration{}, &fixedClock{now: time.Now()})
	got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, string([]byte{'A', 0xff}), "disabled")
	if !errors.Is(err, domain.ErrMalformed) || got != (domain.Assignment{}) || provider.callbacks != 0 {
		t.Fatalf("invalid UTF-8 reached update: got=%#v callbacks=%d err=%v", got, provider.callbacks, err)
	}
}

func TestSetAssignmentStatusIsolatesAdministrativeEvidence(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	admin := assignmentStatusAdministration{check: func(snapshot storage.Snapshot, _ domain.Identity, _ domain.Assignment) error {
		delete(snapshot.Assignments, "fm5b7t4p0dq3")
		delete(snapshot.Controls, "fk3x9r2m5iv8")
		return nil
	}}
	provider, _ := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	defer provider.Close()
	service, _ := mutation.New(provider, admin, &fixedClock{now: time.Now()})
	if _, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4p5iv8", "disabled"); err != nil {
		t.Fatal(err)
	}
	if err := provider.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
		if _, ok := snapshot.Assignments["fm5b7t4p0dq3"]; !ok {
			t.Fatal("administrator mutated evidence")
		}
		if _, ok := snapshot.Controls["fk3x9r2m5iv8"]; !ok {
			t.Fatal("administrator mutated controls")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSetAssignmentStatusRestoreRechecksParentAndChildExpiry(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	for _, grantID := range []string{"fk3x9r2m5iv8", "fk3x9r2man0d"} {
		t.Run(grantID, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			a := fixture.Proposed
			a.Status = "disabled"
			fixture.Snapshot.Assignments[a.ID] = a
			start, expiry := time.Now(), time.Now().Add(time.Second)
			key := domain.GrantKey{ID: grantID, Revision: 1}
			content := fixture.Snapshot.Contents[key]
			content.Validity = &domain.Validity{ExpiresAt: &expiry}
			fixture.Snapshot.Contents[key] = content
			provider, _ := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
			defer provider.Close()
			service, _ := mutation.New(provider, assignmentStatusAdministration{}, &sequenceClock{times: []time.Time{start, expiry}})
			if got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4pan0d", "enabled"); !errors.Is(err, domain.ErrRejected) || got != (domain.Assignment{}) {
				t.Fatalf("expiry crossing = %#v, %v", got, err)
			}
			if err := provider.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
				if snapshot.Assignments["fm5b7t4pan0d"].Status != "disabled" {
					t.Fatal("expiry crossing wrote")
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSetAssignmentStatusRejectsDepthAndSnapshotOverflowWithoutWrite(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	parentTeam, parentGrant := "fibggi2juubk", "fk3x9r2m5iv8"
	for i := 2; i <= 258; i++ {
		team, grant, assignment := fmt.Sprintf("Team%d", i), fmt.Sprintf("G%d", i), fmt.Sprintf("A%d", i)
		fixture.Snapshot.Teams[team] = domain.Team{Name: "team",ID: team, ParentID: parentTeam}
		fixture.Snapshot.Controls[grant] = domain.GrantControl{Version: "1", ID: grant, Status: "enabled"}
		fixture.Snapshot.Contents[domain.GrantKey{ID: grant, Revision: 1}] = domain.GrantContent{Version: "1", GrantID: grant, Revision: 1, ParentGrantID: parentGrant, Permissions: []string{lab.PayslipRead}, Scope: map[string]string{}}
		fixture.Snapshot.Assignments[assignment] = domain.Assignment{Version: "1", ID: assignment, GrantID: grant, GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: team}, Status: "disabled"}
		parentTeam, parentGrant = team, grant
	}
	service, _ := mutation.New(&failingCommitProvider{snapshot: fixture.Snapshot}, assignmentStatusAdministration{}, &fixedClock{now: time.Now()})
	if got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4p5iv8", "disabled"); !errors.Is(err, domain.ErrUnavailable) || got != (domain.Assignment{}) {
		t.Fatalf("depth = %#v, %v", got, err)
	}

	path := t.TempDir() + "/authority.db"
	seeded, _ := lab.CreateSQLite(t.Context(), path, []storage.Snapshot{lab.TeamFINC17(area).Snapshot})
	_ = seeded.Close()
	limited, _ := sqlite.OpenWithOptions(t.Context(), path, sqlite.Options{MaxSnapshotRecords: 2})
	service, _ = mutation.New(limited, assignmentStatusAdministration{}, &fixedClock{now: time.Now()})
	got, err := service.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4p5iv8", "disabled")
	if !errors.Is(err, storage.ErrSnapshotLimit) || got != (domain.Assignment{}) {
		t.Fatalf("snapshot = %#v, %v", got, err)
	}
	_ = limited.Close()
	reopened, _ := sqlite.Open(t.Context(), path)
	defer reopened.Close()
	if err := reopened.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
		if snapshot.Assignments["fm5b7t4p5iv8"].Status != "enabled" {
			t.Fatal("overflow wrote")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
