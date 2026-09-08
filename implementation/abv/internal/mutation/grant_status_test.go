package mutation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/sqlite"
	"context"
	"errors"
	"testing"
	"time"
)

type grantAdministration struct {
	check func(storage.Snapshot, domain.Identity, domain.GrantControl) error
}

func (a grantAdministration) CheckAssignment(context.Context, storage.Snapshot, domain.Identity, domain.Assignment, time.Time) error {
	return domain.ErrUnsupported
}

func (a grantAdministration) CheckGrantStatus(_ context.Context, snapshot storage.Snapshot, identity domain.Identity, proposed domain.GrantControl, _ time.Time) error {
	if a.check != nil {
		return a.check(snapshot, identity, proposed)
	}
	return nil
}

func TestSetGrantStatusPersistsOnlyControlAndUsesAdoptedRevision(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments["A2"] = fixture.Proposed
	newer := fixture.Child
	newer.Revision = 2
	fixture.Snapshot.Contents[domain.GrantKey{ID: "G2", Revision: 2}] = newer
	beforeAssignments := len(fixture.Snapshot.Assignments)
	path := t.TempDir() + "/authority.db"
	provider, err := lab.CreateSQLite(t.Context(), path, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	service, _ := mutation.New(provider, grantAdministration{}, &fixedClock{now: time.Now()})
	disabled := domain.GrantControl{Version: "1", ID: "G2", Status: "disabled"}
	if got, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, disabled); err != nil || got != disabled {
		t.Fatalf("disable = %#v, %v", got, err)
	}
	enabled := disabled
	enabled.Status = "enabled"
	if got, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, enabled); err != nil || got != enabled {
		t.Fatalf("enable = %#v, %v", got, err)
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
		if snapshot.Controls["G2"] != enabled || len(snapshot.Assignments) != beforeAssignments || snapshot.Assignments["A2"].GrantRevision != 1 {
			t.Fatalf("unexpected persisted state: control=%#v assignments=%#v", snapshot.Controls["G2"], snapshot.Assignments)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSetGrantStatusEnableRequiresEveryEnabledRouteButSkipsDisabledOnes(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	for _, test := range []struct {
		name       string
		brokenMode string
		want       error
	}{
		{name: "second enabled route broken", brokenMode: "enabled", want: domain.ErrRejected},
		{name: "same broken route disabled", brokenMode: "disabled"},
		{name: "no assignments"},
		{name: "direct human enabled", brokenMode: "user", want: domain.ErrUnsupported},
		{name: "required upstream disabled", brokenMode: "upstream", want: domain.ErrRejected},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			fixture.Snapshot.Teams["BrokenTeam"] = domain.Team{ID: "BrokenTeam", ParentID: "RootTeam"}
			control := fixture.Snapshot.Controls["G2"]
			control.Status = "disabled"
			fixture.Snapshot.Controls["G2"] = control
			if test.brokenMode != "" && test.brokenMode != "upstream" {
				fixture.Snapshot.Assignments["A2"] = fixture.Proposed
				broken := fixture.Proposed
				broken.ID, broken.Recipient.ID, broken.Status = "A3", "BrokenTeam", test.brokenMode
				if test.brokenMode == "user" {
					broken.Recipient = domain.Recipient{Type: "user", ID: "maya"}
					broken.Status = "enabled"
				}
				fixture.Snapshot.Assignments[broken.ID] = broken
			}
			if test.brokenMode == "upstream" {
				fixture.Snapshot.Assignments["A2"] = fixture.Proposed
				upstream := fixture.Snapshot.Assignments["A1"]
				upstream.Status = "disabled"
				fixture.Snapshot.Assignments["A1"] = upstream
			}
			provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
			if err != nil {
				t.Fatal(err)
			}
			defer provider.Close()
			service, _ := mutation.New(provider, grantAdministration{}, &fixedClock{now: time.Now()})
			proposed := domain.GrantControl{Version: "1", ID: "G2", Status: "enabled"}
			got, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, proposed)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if test.want != nil && got != (domain.GrantControl{}) {
				t.Fatalf("failed enable returned %#v", got)
			}
			if err := provider.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
				wantStatus := "enabled"
				if test.want != nil {
					wantStatus = "disabled"
				}
				if snapshot.Controls["G2"].Status != wantStatus {
					t.Fatalf("status = %q, want %q", snapshot.Controls["G2"].Status, wantStatus)
				}
				if test.brokenMode == "disabled" && snapshot.Assignments["A3"].Status != "disabled" {
					t.Fatal("disabled assignment changed")
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSetGrantStatusWithdrawalIgnoresBrokenSupportButStillUsesAdministrativeGate(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments["A2"] = fixture.Proposed
	delete(fixture.Snapshot.Assignments, "A1")
	admin := grantAdministration{check: func(snapshot storage.Snapshot, _ domain.Identity, _ domain.GrantControl) error {
		delete(snapshot.Controls, "G2")
		return nil
	}}
	provider, _ := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	defer provider.Close()
	service, _ := mutation.New(provider, admin, &fixedClock{now: time.Now()})
	proposed := domain.GrantControl{Version: "1", ID: "G2", Status: "disabled"}
	if got, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, proposed); err != nil || got != proposed {
		t.Fatalf("withdrawal = %#v, %v", got, err)
	}
}

func TestSetGrantStatusRejectsExpiryCrossingAndSameStateBypass(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments["A2"] = fixture.Proposed
	start := time.Now()
	expires := start.Add(time.Second)
	key := domain.GrantKey{ID: "G2", Revision: 1}
	content := fixture.Snapshot.Contents[key]
	content.Validity = &domain.Validity{ExpiresAt: &expires}
	fixture.Snapshot.Contents[key] = content
	control := fixture.Snapshot.Controls["G2"]
	control.Status = "disabled"
	fixture.Snapshot.Controls["G2"] = control
	provider, _ := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	defer provider.Close()
	service, _ := mutation.New(provider, grantAdministration{}, &sequenceClock{times: []time.Time{start, expires}})
	proposed := domain.GrantControl{Version: "1", ID: "G2", Status: "enabled"}
	if got, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, proposed); !errors.Is(err, domain.ErrRejected) || got != (domain.GrantControl{}) {
		t.Fatalf("expiry crossing = %#v, %v", got, err)
	}

	rejecting, _ := mutation.New(provider, grantAdministration{check: func(storage.Snapshot, domain.Identity, domain.GrantControl) error { return domain.ErrRejected }}, &fixedClock{now: start})
	if got, err := rejecting.SetGrantStatus(t.Context(), area, fixture.Issuer, control); !errors.Is(err, domain.ErrRejected) || got != (domain.GrantControl{}) {
		t.Fatalf("same-state bypass = %#v, %v", got, err)
	}
}
