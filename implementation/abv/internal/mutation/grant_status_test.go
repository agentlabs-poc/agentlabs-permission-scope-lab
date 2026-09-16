package mutation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/sqlite"
	"agentlabs.local/abv/lab"
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
	fixture.Snapshot.Assignments["fm5b7t4pan0d"] = fixture.Proposed
	newer := fixture.Child
	newer.Revision = 2
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 2}] = newer
	beforeAssignments := len(fixture.Snapshot.Assignments)
	path := t.TempDir() + "/authority.db"
	provider, err := lab.CreateSQLite(t.Context(), path, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	service, _ := mutation.New(provider, grantAdministration{}, &fixedClock{now: time.Now()})
	disabled := domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "disabled"}
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
	reopened, err := sqlite.Open(t.Context(), path, allowAllRegistry{})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err = reopened.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
		if snapshot.Controls["fk3x9r2man0d"] != enabled || len(snapshot.Assignments) != beforeAssignments || snapshot.Assignments["fm5b7t4pan0d"].GrantRevision != 1 {
			t.Fatalf("unexpected persisted state: control=%#v assignments=%#v", snapshot.Controls["fk3x9r2man0d"], snapshot.Assignments)
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
			fixture.Snapshot.Teams["ficfwlfpxibk"] = domain.Team{Name: "team", ID: "ficfwlfpxibk", ParentID: "fibggi2jur5s"}
			control := fixture.Snapshot.Controls["fk3x9r2man0d"]
			control.Status = "disabled"
			fixture.Snapshot.Controls["fk3x9r2man0d"] = control
			if test.brokenMode != "" && test.brokenMode != "upstream" {
				fixture.Snapshot.Assignments["fm5b7t4pan0d"] = fixture.Proposed
				broken := fixture.Proposed
				broken.ID, broken.Recipient.ID, broken.Status = "fm5b7t4pfs5i", "ficfwlfpxibk", test.brokenMode
				if test.brokenMode == "user" {
					broken.Recipient = domain.Recipient{Type: "user", ID: "fi7io4lvjqio"}
					broken.Status = "enabled"
				}
				fixture.Snapshot.Assignments[broken.ID] = broken
			}
			if test.brokenMode == "upstream" {
				fixture.Snapshot.Assignments["fm5b7t4pan0d"] = fixture.Proposed
				upstream := fixture.Snapshot.Assignments["fm5b7t4p5iv8"]
				upstream.Status = "disabled"
				fixture.Snapshot.Assignments["fm5b7t4p5iv8"] = upstream
			}
			provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
			if err != nil {
				t.Fatal(err)
			}
			defer provider.Close()
			service, _ := mutation.New(provider, grantAdministration{}, &fixedClock{now: time.Now()})
			proposed := domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "enabled"}
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
				if snapshot.Controls["fk3x9r2man0d"].Status != wantStatus {
					t.Fatalf("status = %q, want %q", snapshot.Controls["fk3x9r2man0d"].Status, wantStatus)
				}
				if test.brokenMode == "disabled" && snapshot.Assignments["fm5b7t4pfs5i"].Status != "disabled" {
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
	fixture.Snapshot.Assignments["fm5b7t4pan0d"] = fixture.Proposed
	delete(fixture.Snapshot.Assignments, "fm5b7t4p5iv8")
	admin := grantAdministration{check: func(snapshot storage.Snapshot, _ domain.Identity, _ domain.GrantControl) error {
		delete(snapshot.Controls, "fk3x9r2man0d")
		return nil
	}}
	provider, _ := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	defer provider.Close()
	service, _ := mutation.New(provider, admin, &fixedClock{now: time.Now()})
	proposed := domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "disabled"}
	if got, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, proposed); err != nil || got != proposed {
		t.Fatalf("withdrawal = %#v, %v", got, err)
	}
}

func TestSetGrantStatusRejectsExpiryCrossingAndSameStateBypass(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments["fm5b7t4pan0d"] = fixture.Proposed
	start := time.Now()
	expires := start.Add(time.Second)
	key := domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1}
	content := fixture.Snapshot.Contents[key]
	content.Validity = &domain.Validity{ExpiresAt: &expires}
	fixture.Snapshot.Contents[key] = content
	control := fixture.Snapshot.Controls["fk3x9r2man0d"]
	control.Status = "disabled"
	fixture.Snapshot.Controls["fk3x9r2man0d"] = control
	provider, _ := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	defer provider.Close()
	service, _ := mutation.New(provider, grantAdministration{}, &sequenceClock{times: []time.Time{start, expires}})
	proposed := domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "enabled"}
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
	fixture.Snapshot.Assignments["fm5b7t4pan0d"] = fixture.Proposed
	start, expires := time.Now(), time.Now().Add(time.Second)
	key := domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}
	content := fixture.Snapshot.Contents[key]
	content.Validity = &domain.Validity{ExpiresAt: &expires}
	fixture.Snapshot.Contents[key] = content
	control := fixture.Snapshot.Controls["fk3x9r2man0d"]
	control.Status = "disabled"
	fixture.Snapshot.Controls["fk3x9r2man0d"] = control
	provider, _ := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	defer provider.Close()
	service, _ := mutation.New(provider, grantAdministration{}, &sequenceClock{times: []time.Time{start, expires}})
	got, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "enabled"})
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
	proposed := domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "disabled"}
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
	parentTeam, parentGrant := "fibggi2jur5s", "fk3x9r2m0dq3"
	for i := 1; i <= 258; i++ {
		team, grant, assignment := fmt.Sprintf("DeepTeam%d", i), fmt.Sprintf("DeepGrant%d", i), fmt.Sprintf("DeepAssignment%d", i)
		fixture.Snapshot.Teams[team] = domain.Team{Name: "team", ID: team, ParentID: parentTeam}
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
	other, _ := domain.NewArea("fi7io4lvkfsw", "hrms")
	fixture := lab.TeamFINC17(area)
	provider := &failingCommitProvider{snapshot: fixture.Snapshot}
	service, _ := mutation.New(provider, grantAdministration{}, &fixedClock{now: time.Now()})
	for name, request := range map[string]struct {
		area    domain.Area
		control domain.GrantControl
		want    error
	}{
		"wrong area": {area: other, control: domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "disabled"}, want: domain.ErrRejected},
		"wildcard":   {area: area, control: domain.GrantControl{Version: "1", ID: "G*", Status: "disabled"}, want: domain.ErrMalformed},
		"bad status": {area: area, control: domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "paused"}, want: domain.ErrMalformed},
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
	limited, err := sqlite.OpenWithOptions(t.Context(), path, sqlite.Options{MaxSnapshotRecords: 2, Registry: allowAllRegistry{}})
	if err != nil {
		t.Fatal(err)
	}
	service, _ := mutation.New(limited, grantAdministration{}, &fixedClock{now: time.Now()})
	proposed := domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "disabled"}
	got, err := service.SetGrantStatus(t.Context(), area, fixture.Issuer, proposed)
	if !errors.Is(err, storage.ErrSnapshotLimit) || !errors.Is(err, domain.ErrUnavailable) || got != (domain.GrantControl{}) {
		t.Fatalf("limited snapshot = %#v, %v", got, err)
	}
	if err = limited.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := sqlite.Open(t.Context(), path, allowAllRegistry{})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err = reopened.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
		if snapshot.Controls["fk3x9r2man0d"].Status != "enabled" {
			t.Fatalf("status after limited attempt = %q", snapshot.Controls["fk3x9r2man0d"].Status)
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
			fixture.Snapshot.Assignments["fm5b7t4pan0d"] = fixture.Proposed
			ancestor := fixture.Snapshot.Controls["fk3x9r2m5iv8"]
			ancestor.Status = "disabled"
			fixture.Snapshot.Controls["fk3x9r2m5iv8"] = ancestor
			descendant := fixture.Snapshot.Controls["fk3x9r2man0d"]
			descendant.Status = descendantStatus
			fixture.Snapshot.Controls["fk3x9r2man0d"] = descendant
			if _, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", time.Now()); !errors.Is(err, domain.ErrRejected) {
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
				if snapshot.Controls["fk3x9r2man0d"].Status != descendantStatus || snapshot.Assignments["fm5b7t4pan0d"].Status != "enabled" {
					t.Fatal("descendant state changed")
				}
				_, err := lineage.ResolveParentTeam(snapshot, fixture.Child, "fibggi2juxhc", time.Now())
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

// allowAllRegistry satisfies the port for tests whose subject is storage rather
// than the installation gate.
type allowAllRegistry struct{}

func (allowAllRegistry) ApplicationExists(context.Context, string) (bool, error) { return true, nil }
func (allowAllRegistry) Installed(context.Context, string, string) (bool, error) { return true, nil }

// Q-132: a grant with a child grant can be neither disabled nor deleted.
//
// Both halves need a subject whose *only* dependent is a child grant. Delete
// already refused a grant an assignment names — an older rule, against dangling
// references — so a parent that is also held is refused either way and says
// nothing about this rule. The first version of this test used exactly such a
// parent, and removing the child check left the whole suite green.
//
// An assignment is deliberately not a child. Holding a grant *is* an assignment,
// so counting one would make disable unavailable for every grant anybody holds,
// and Q-079's operational pause would name an operation nobody could perform.
func TestQ132RefusesDisableAndDeleteWhileAChildGrantExists(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)

	// A parent and a child, neither of them held by anyone.
	for _, id := range []string{"fk3x9r2mpppp", "fk3x9r2mcccc"} {
		fixture.Snapshot.Controls[id] = domain.GrantControl{Version: "1", ID: id, Status: "enabled"}
	}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2mpppp", Revision: 1}] = domain.GrantContent{
		Version: "1", GrantID: "fk3x9r2mpppp", Revision: 1, ParentGrantID: "fk3x9r2man0d",
		Permissions: []string{lab.PayslipRead}, Scope: map[string]string{},
	}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2mcccc", Revision: 1}] = domain.GrantContent{
		Version: "1", GrantID: "fk3x9r2mcccc", Revision: 1, ParentGrantID: "fk3x9r2mpppp",
		Permissions: []string{lab.PayslipRead}, Scope: map[string]string{},
	}
	provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/q132.db", []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	// One provider, two administrations, because each admits the operation the
	// other refuses: the stub here admits a status change and not a deletion,
	// and the lab's own gate the reverse for these grants. Both gates answer
	// before the dependency rule by design, so isolating that rule means passing
	// whichever gate the half under test needs.
	status, err := mutation.New(provider, grantAdministration{}, &fixedClock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	statusAdmin, err := lab.NewAssignmentStatusAdministration(area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	service, err := mutation.New(provider,
		&lab.RoleAdministration{AssignmentStatusAdministration: statusAdmin},
		&fixedClock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	disable := func(id string) error {
		_, err := status.SetGrantStatus(t.Context(), area, fixture.Issuer,
			domain.GrantControl{Version: "1", ID: id, Status: "disabled"})
		return err
	}

	// The parent holds nothing and is held by nobody. Only its child refuses it.
	if err := disable("fk3x9r2mpppp"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("disabling a parent gave %v, want ErrConflict", err)
	}
	if err := service.DeleteGrant(t.Context(), area, fixture.Issuer, "fk3x9r2mpppp"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("deleting a parent gave %v, want ErrConflict", err)
	}

	// The leaf is neither, which is what makes the refusals above about the
	// child rather than about these two grants.
	if err := disable("fk3x9r2mcccc"); err != nil {
		t.Fatalf("disabling a leaf: %v", err)
	}
	// Enabling is not constrained by anything beneath — B11's rule is that an
	// explicit disable is not undone elsewhere, not that enable is guarded.
	if _, err := status.SetGrantStatus(t.Context(), area, fixture.Issuer,
		domain.GrantControl{Version: "1", ID: "fk3x9r2mcccc", Status: "enabled"}); err != nil {
		t.Fatalf("enabling a leaf: %v", err)
	}
	if err := service.DeleteGrant(t.Context(), area, fixture.Issuer, "fk3x9r2mcccc"); err != nil {
		t.Fatalf("deleting a leaf: %v", err)
	}
	// Bottom-up: with the child gone, the parent goes too.
	if err := disable("fk3x9r2mpppp"); err != nil {
		t.Fatalf("disabling the parent once its child is gone: %v", err)
	}
	if err := service.DeleteGrant(t.Context(), area, fixture.Issuer, "fk3x9r2mpppp"); err != nil {
		t.Fatalf("deleting the parent once its child is gone: %v", err)
	}
}
