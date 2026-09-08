package mutation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/sqlite"
	"context"
	"errors"
	"fmt"
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

func TestSetGrantStatusRejectsParentExpiryCrossing(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments["A2"] = fixture.Proposed
	start, expires := time.Now(), time.Now().Add(time.Second)
	key := domain.GrantKey{ID: "G1", Revision: 1}
	content := fixture.Snapshot.Contents[key]
	content.Validity = &domain.Validity{ExpiresAt: &expires}
	fixture.Snapshot.Contents[key] = content
	control := fixture.Snapshot.Controls["G2"]
	control.Status = "disabled"
	fixture.Snapshot.Controls["G2"] = control
	provider, _ := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	defer provider.Close()
	service, _ := mutation.New(provider, grantAdministration{}, &sequenceClock{times: []time.Time{start, expires}})
	got, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, domain.GrantControl{Version: "1", ID: "G2", Status: "enabled"})
	if !errors.Is(err, domain.ErrRejected) || got != (domain.GrantControl{}) {
		t.Fatalf("parent expiry crossing = %#v, %v", got, err)
	}
}

func TestSetGrantStatusCancellationAndCompetingWriterReturnNoControl(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	ctx, cancel := context.WithCancel(t.Context())
	cancelling := grantAdministration{check: func(storage.Snapshot, domain.Identity, domain.GrantControl) error { cancel(); return nil }}
	service, _ := mutation.New(&failingCommitProvider{snapshot: fixture.Snapshot}, cancelling, &fixedClock{now: time.Now()})
	proposed := domain.GrantControl{Version: "1", ID: "G2", Status: "disabled"}
	if got, err := service.SetGrantStatus(ctx, area, fixture.Issuer, proposed); !errors.Is(err, context.Canceled) || got != (domain.GrantControl{}) {
		t.Fatalf("cancelled = %#v, %v", got, err)
	}
	competing := &failingCommitProvider{snapshot: fixture.Snapshot}
	service, _ = mutation.New(competing, grantAdministration{}, &fixedClock{now: time.Now()})
	if got, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, proposed); !errors.Is(err, domain.ErrConflict) || got != (domain.GrantControl{}) || competing.returned.GrantStatusChange == nil {
		t.Fatalf("competing writer = %#v, writes=%#v, err=%v", got, competing.returned, err)
	}
}

func TestSetGrantStatusRejectsDepthOverflow(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	parentTeam, parentGrant := "RootTeam", "G0"
	for i := 1; i <= 258; i++ {
		team, grant, assignment := fmt.Sprintf("DeepTeam%d", i), fmt.Sprintf("DeepGrant%d", i), fmt.Sprintf("DeepAssignment%d", i)
		fixture.Snapshot.Teams[team] = domain.Team{ID: team, ParentID: parentTeam}
		fixture.Snapshot.Controls[grant] = domain.GrantControl{Version: "1", ID: grant, Status: "enabled"}
		fixture.Snapshot.Contents[domain.GrantKey{ID: grant, Revision: 1}] = domain.GrantContent{Version: "1", GrantID: grant, Revision: 1, ParentGrantID: parentGrant, Permissions: []string{lab.PayslipRead}, Scope: map[string]string{}}
		fixture.Snapshot.Assignments[assignment] = domain.Assignment{Version: "1", ID: assignment, GrantID: grant, GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: team}, Status: "enabled"}
		parentTeam, parentGrant = team, grant
	}
	target := fixture.Snapshot.Controls[parentGrant]
	target.Status = "disabled"
	fixture.Snapshot.Controls[parentGrant] = target
	service, _ := mutation.New(&failingCommitProvider{snapshot: fixture.Snapshot}, grantAdministration{}, &fixedClock{now: time.Now()})
	target.Status = "enabled"
	if got, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, target); !errors.Is(err, domain.ErrUnavailable) || got != (domain.GrantControl{}) {
		t.Fatalf("depth overflow = %#v, %v", got, err)
	}
}

func TestSetGrantStatusBoundaryValidationAndTypedNil(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	other, _ := domain.NewArea("other", "hrms")
	fixture := lab.TeamFINC17(area)
	provider := &failingCommitProvider{snapshot: fixture.Snapshot}
	service, _ := mutation.New(provider, grantAdministration{}, &fixedClock{now: time.Now()})
	for name, request := range map[string]struct {
		area    domain.Area
		control domain.GrantControl
		want    error
	}{
		"wrong area": {area: other, control: domain.GrantControl{Version: "1", ID: "G2", Status: "disabled"}, want: domain.ErrRejected},
		"wildcard":   {area: area, control: domain.GrantControl{Version: "1", ID: "G*", Status: "disabled"}, want: domain.ErrMalformed},
		"bad status": {area: area, control: domain.GrantControl{Version: "1", ID: "G2", Status: "paused"}, want: domain.ErrMalformed},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := service.SetGrantStatus(t.Context(), request.area, fixture.Issuer, request.control)
			if !errors.Is(err, request.want) || got != (domain.GrantControl{}) {
				t.Fatalf("got %#v, %v; want %v", got, err, request.want)
			}
		})
	}
	var nilAdmin *grantAdministration
	if _, err := mutation.New(provider, nilAdmin, &fixedClock{now: time.Now()}); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("typed nil error = %v", err)
	}
}

func TestSetGrantStatusSnapshotLimitDoesNotWrite(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	path := t.TempDir() + "/authority.db"
	seeded, err := lab.CreateSQLite(t.Context(), path, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if err = seeded.Close(); err != nil {
		t.Fatal(err)
	}
	limited, err := sqlite.OpenWithOptions(t.Context(), path, sqlite.Options{MaxSnapshotRecords: 2})
	if err != nil {
		t.Fatal(err)
	}
	service, _ := mutation.New(limited, grantAdministration{}, &fixedClock{now: time.Now()})
	proposed := domain.GrantControl{Version: "1", ID: "G2", Status: "disabled"}
	got, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, proposed)
	if !errors.Is(err, storage.ErrSnapshotLimit) || !errors.Is(err, domain.ErrUnavailable) || got != (domain.GrantControl{}) {
		t.Fatalf("limited snapshot = %#v, %v", got, err)
	}
	if err = limited.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := sqlite.Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err = reopened.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
		if snapshot.Controls["G2"].Status != "enabled" {
			t.Fatalf("status after limited attempt = %q", snapshot.Controls["G2"].Status)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSetGrantStatusPreservesDescendantStateAndEffectiveness(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	for _, descendantStatus := range []string{"disabled", "enabled"} {
		t.Run(descendantStatus, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			fixture.Snapshot.Assignments["A2"] = fixture.Proposed
			ancestor := fixture.Snapshot.Controls["G1"]
			ancestor.Status = "disabled"
			fixture.Snapshot.Controls["G1"] = ancestor
			descendant := fixture.Snapshot.Controls["G2"]
			descendant.Status = descendantStatus
			fixture.Snapshot.Controls["G2"] = descendant
			if _, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "Team2", time.Now()); !errors.Is(err, domain.ErrRejected) {
				t.Fatalf("descendant effective before restore: %v", err)
			}
			provider, _ := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
			defer provider.Close()
			service, _ := mutation.New(provider, grantAdministration{}, &fixedClock{now: time.Now()})
			ancestor.Status = "enabled"
			if _, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, ancestor); err != nil {
				t.Fatal(err)
			}
			if err := provider.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
				if snapshot.Controls["G2"].Status != descendantStatus || snapshot.Assignments["A2"].Status != "enabled" {
					t.Fatal("descendant state changed")
				}
				_, err := lineage.ResolveParentTeam(snapshot, fixture.Child, "Team2", time.Now())
				if descendantStatus == "enabled" && err != nil || descendantStatus == "disabled" && !errors.Is(err, domain.ErrRejected) {
					t.Fatalf("descendant effectiveness error = %v", err)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
