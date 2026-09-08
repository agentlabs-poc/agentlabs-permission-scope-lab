package mutation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

type fixedClock struct{ now time.Time }

func (c *fixedClock) Now() time.Time { return c.now }

func TestCreateAssignmentPersistsExactProposalAfterBothChecks(t *testing.T) {
	area, err := domain.NewArea("tenant-fin", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	fixture := lab.TeamFINC17(area)
	provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	clock := &fixedClock{now: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)}
	admin, err := lab.NewAdministration(area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	service, err := mutation.New(provider, admin, clock)
	if err != nil {
		t.Fatal(err)
	}

	receipt, err := service.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.AssignmentID != "A2" {
		t.Fatalf("receipt = %#v", receipt)
	}
	if err := provider.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
		got, ok := snapshot.Assignments["A2"]
		if !ok || got != fixture.Proposed {
			t.Fatalf("persisted assignment = %#v, present=%v", got, ok)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

var _ mutation.Clock = (*fixedClock)(nil)

func TestCreateAssignmentFailuresDoNotWrite(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name      string
		change    func(*lab.TeamFINC17Case)
		want      error
		wantCount int
	}{
		{name: "administration absent while source present", change: func(c *lab.TeamFINC17Case) {
			c.Snapshot.Memberships = c.Snapshot.Memberships[:2]
		}, want: domain.ErrRejected, wantCount: 2},
		{name: "source absent while administration present", change: func(c *lab.TeamFINC17Case) {
			c.Snapshot.Memberships[0] = domain.Membership{TeamID: "Team1", HumanID: "someone-else"}
		}, want: domain.ErrRejected, wantCount: 2},
		{name: "malformed proposal", change: func(c *lab.TeamFINC17Case) {
			c.Proposed.ID = ""
		}, want: domain.ErrMalformed, wantCount: 2},
		{name: "unsupported user recipient", change: func(c *lab.TeamFINC17Case) {
			c.Proposed.Recipient.Type = "user"
		}, want: domain.ErrUnsupported, wantCount: 2},
		{name: "unsupported disabled creation", change: func(c *lab.TeamFINC17Case) {
			c.Proposed.Status = "disabled"
		}, want: domain.ErrUnsupported, wantCount: 2},
		{name: "selected content missing", change: func(c *lab.TeamFINC17Case) {
			delete(c.Snapshot.Contents, domain.GrantKey{ID: "G2", Revision: 1})
		}, want: domain.ErrRejected, wantCount: 2},
		{name: "parent support missing", change: func(c *lab.TeamFINC17Case) {
			delete(c.Snapshot.Assignments, "A1")
		}, want: domain.ErrRejected, wantCount: 1},
		{name: "child control disabled", change: func(c *lab.TeamFINC17Case) {
			control := c.Snapshot.Controls["G2"]
			control.Status = "disabled"
			c.Snapshot.Controls["G2"] = control
		}, want: domain.ErrRejected, wantCount: 2},
		{name: "child validity expired", change: func(c *lab.TeamFINC17Case) {
			expires := now
			key := domain.GrantKey{ID: "G2", Revision: 1}
			content := c.Snapshot.Contents[key]
			content.Validity = &domain.Validity{ExpiresAt: &expires}
			c.Snapshot.Contents[key], c.Child = content, content
		}, want: domain.ErrRejected, wantCount: 2},
		{name: "selected permissions exceed parent", change: func(c *lab.TeamFINC17Case) {
			key := domain.GrantKey{ID: "G2", Revision: 1}
			g := c.Snapshot.Contents[key]
			g.Permissions = []string{lab.PayslipDelete}
			c.Snapshot.Contents[key], c.Child = g, g
		}, want: domain.ErrRejected, wantCount: 2},
		{name: "wrong recipient", change: func(c *lab.TeamFINC17Case) {
			c.Proposed.Recipient.ID = "Team1"
		}, want: domain.ErrRejected, wantCount: 2},
		{name: "definition validation cannot establish permission", change: func(c *lab.TeamFINC17Case) {
			definition := c.Snapshot.Catalog.Permissions[lab.PayslipRead]
			definition.Active = false
			c.Snapshot.Catalog.Permissions[lab.PayslipRead] = definition
		}, want: domain.ErrRejected, wantCount: 2},
		{name: "older selected revision has no latest fallback", change: func(c *lab.TeamFINC17Case) {
			newer := c.Child
			newer.Revision = 2
			c.Snapshot.Contents[domain.GrantKey{ID: "G2", Revision: 2}] = newer
		}, want: domain.ErrRejected, wantCount: 2},
		{name: "disabled duplicate still occupies binding", change: func(c *lab.TeamFINC17Case) {
			c.Snapshot.Assignments["old-A2"] = domain.Assignment{Version: "1", ID: "old-A2", GrantID: "G2", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "Team2"}, Status: "disabled"}
		}, want: domain.ErrConflict, wantCount: 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			test.change(&fixture)
			provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
			if err != nil {
				t.Fatal(err)
			}
			defer provider.Close()
			admin, err := lab.NewAdministration(area, fixture.Administration)
			if err != nil {
				t.Fatal(err)
			}
			service, err := mutation.New(provider, admin, &fixedClock{now: now})
			if err != nil {
				t.Fatal(err)
			}
			receipt, err := service.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if receipt != (domain.Receipt{}) {
				t.Fatalf("failed operation returned receipt %#v", receipt)
			}
			assertAssignmentCount(t, provider, area, test.wantCount)
		})
	}
}

func TestCreateAssignmentRejectsWrongAreaWithoutWriting(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	other, _ := domain.NewArea("another-tenant", "hrms")
	fixture := lab.TeamFINC17(area)
	provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	admin, _ := lab.NewAdministration(area, fixture.Administration)
	service, _ := mutation.New(provider, admin, &fixedClock{now: time.Now()})
	receipt, err := service.CreateAssignment(t.Context(), other, fixture.Issuer, fixture.Proposed)
	if !errors.Is(err, domain.ErrNotFound) || receipt != (domain.Receipt{}) {
		t.Fatalf("cross-area attempt receipt=%#v err=%v", receipt, err)
	}
	assertAssignmentCount(t, provider, area, 2)
}

func TestLabAdministrationRejectsAnyChangedTrustedPremise(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	for name, change := range map[string]func(*lab.AdministrationPremise){
		"human":      func(p *lab.AdministrationPremise) { p.HumanID = "nutan" },
		"permission": func(p *lab.AdministrationPremise) { p.PermissionID = lab.PayslipRead },
		"recipient":  func(p *lab.AdministrationPremise) { p.RecipientTeamID = "Team1" },
	} {
		t.Run(name, func(t *testing.T) {
			premise := fixture.Administration
			change(&premise)
			if _, err := lab.NewAdministration(area, premise); !errors.Is(err, domain.ErrRejected) {
				t.Fatalf("changed premise error=%v", err)
			}
		})
	}
}

func TestAdministrativeSnapshotMutationCannotAlterBusinessEvidence(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	expires := time.Now().Add(time.Hour)
	key := domain.GrantKey{ID: "G1", Revision: 1}
	content := fixture.Snapshot.Contents[key]
	content.Validity = &domain.Validity{ExpiresAt: &expires}
	fixture.Snapshot.Contents[key] = content
	provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	var want, working storage.Snapshot
	if err := provider.Read(t.Context(), area, func(snapshot storage.Snapshot) error { want = snapshot; return nil }); err != nil {
		t.Fatal(err)
	}
	if err := provider.Read(t.Context(), area, func(snapshot storage.Snapshot) error { working = snapshot; return nil }); err != nil {
		t.Fatal(err)
	}
	provider.Close()
	admin := mutatingAdministration{check: func(snapshot storage.Snapshot) error {
		changed := snapshot.Controls["G1"]
		changed.Status = "disabled"
		snapshot.Controls["G1"] = changed
		snapshot.Memberships[0].HumanID = "forged"
		delete(snapshot.TrustedRoots, "G0")
		content := snapshot.Contents[key]
		content.Permissions[0] = lab.PayslipDelete
		*content.Validity.ExpiresAt = time.Time{}
		return nil
	}}
	memory := &writeProvider{snapshot: working}
	service, _ := mutation.New(memory, admin, &fixedClock{now: time.Now()})
	receipt, err := service.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed)
	if err != nil || receipt.AssignmentID != fixture.Proposed.ID {
		t.Fatalf("isolated adapter mutation receipt=%#v err=%v", receipt, err)
	}
	delete(memory.snapshot.Assignments, fixture.Proposed.ID)
	if !reflect.DeepEqual(memory.snapshot, want) {
		t.Fatalf("administrative adapter altered provider evidence\n got: %#v\nwant: %#v", memory.snapshot, want)
	}
}

func TestCreateAssignmentRechecksEligibilityImmediatelyBeforeWriteSet(t *testing.T) {
	for _, grantID := range []string{"G2", "G1"} {
		t.Run(grantID, func(t *testing.T) {
			area, _ := domain.NewArea("tenant-fin", "hrms")
			fixture := lab.TeamFINC17(area)
			start := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
			expires := start.Add(time.Second)
			key := domain.GrantKey{ID: grantID, Revision: 1}
			content := fixture.Snapshot.Contents[key]
			content.Validity = &domain.Validity{ExpiresAt: &expires}
			fixture.Snapshot.Contents[key] = content
			provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
			if err != nil {
				t.Fatal(err)
			}
			defer provider.Close()
			admin, _ := lab.NewAdministration(area, fixture.Administration)
			service, _ := mutation.New(provider, admin, &sequenceClock{times: []time.Time{start, expires}})
			receipt, err := service.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed)
			if !errors.Is(err, domain.ErrRejected) || receipt != (domain.Receipt{}) {
				t.Fatalf("%s expiry crossing yielded receipt=%#v err=%v", grantID, receipt, err)
			}
			assertAssignmentCount(t, provider, area, 2)
		})
	}
}

func TestProviderFailureReturnsNoReceiptAndCallbackIsNotReplayed(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	provider := &failingCommitProvider{snapshot: fixture.Snapshot}
	admin, _ := lab.NewAdministration(area, fixture.Administration)
	service, _ := mutation.New(provider, admin, &fixedClock{now: time.Now()})
	receipt, err := service.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed)
	if !errors.Is(err, domain.ErrConflict) || receipt != (domain.Receipt{}) {
		t.Fatalf("commit failure yielded receipt=%#v err=%v", receipt, err)
	}
	if provider.callbacks != 1 {
		t.Fatalf("callback count = %d, want 1", provider.callbacks)
	}
	if len(provider.returned.NewAssignments) != 1 || provider.returned.NewAssignments[0] != fixture.Proposed {
		t.Fatalf("coordinator write set = %#v", provider.returned)
	}
	ctx, cancel := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
	defer cancel()
	provider.callbacks = 0
	receipt, err = service.CreateAssignment(ctx, area, fixture.Issuer, fixture.Proposed)
	if !errors.Is(err, context.DeadlineExceeded) || receipt != (domain.Receipt{}) || provider.callbacks != 0 {
		t.Fatalf("expired deadline yielded receipt=%#v callbacks=%d err=%v", receipt, provider.callbacks, err)
	}
}

func TestNewRejectsTypedNilDependenciesAndIdentityWildcardsAreMalformed(t *testing.T) {
	clock := &fixedClock{now: time.Now()}
	provider := &failingCommitProvider{}
	admin := &mutatingAdministration{check: func(storage.Snapshot) error { return nil }}
	for name, call := range map[string]func() error{
		"provider": func() error {
			var value *failingCommitProvider
			_, err := mutation.New(value, admin, clock)
			return err
		},
		"administration": func() error {
			var value *mutatingAdministration
			_, err := mutation.New(provider, value, clock)
			return err
		},
		"clock": func() error { var value *fixedClock; _, err := mutation.New(provider, admin, value); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); !errors.Is(err, domain.ErrMalformed) {
				t.Fatalf("error = %v, want malformed", err)
			}
		})
	}

	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	provider.snapshot = fixture.Snapshot
	identity := fixture.Issuer
	identity.Actor.ID, identity.HumanID = "*", "*"
	service, _ := mutation.New(provider, admin, clock)
	if _, err := service.CreateAssignment(t.Context(), area, identity, fixture.Proposed); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("wildcard identity error = %v, want malformed", err)
	}
}

type mutatingAdministration struct{ check func(storage.Snapshot) error }

func (a mutatingAdministration) CheckAssignment(_ context.Context, snapshot storage.Snapshot, _ domain.Identity, _ domain.Assignment, _ time.Time) error {
	return a.check(snapshot)
}

type sequenceClock struct {
	mu    sync.Mutex
	times []time.Time
	index int
}

func (c *sequenceClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.index >= len(c.times) {
		return c.times[len(c.times)-1]
	}
	result := c.times[c.index]
	c.index++
	return result
}

type failingCommitProvider struct {
	snapshot  storage.Snapshot
	callbacks int
	returned  storage.WriteSet
}

func (p *failingCommitProvider) Read(context.Context, domain.Area, func(storage.Snapshot) error) error {
	return domain.ErrUnsupported
}
func (p *failingCommitProvider) Update(_ context.Context, _ domain.Area, callback func(storage.Snapshot) (storage.WriteSet, error)) error {
	p.callbacks++
	writes, err := callback(p.snapshot)
	if err != nil {
		return err
	}
	p.returned = writes
	return domain.ErrConflict
}
func (p *failingCommitProvider) Close() error { return nil }

type writeProvider struct{ snapshot storage.Snapshot }

func (p *writeProvider) Read(_ context.Context, _ domain.Area, callback func(storage.Snapshot) error) error {
	return callback(p.snapshot)
}
func (p *writeProvider) Update(_ context.Context, _ domain.Area, callback func(storage.Snapshot) (storage.WriteSet, error)) error {
	writes, err := callback(p.snapshot)
	if err != nil {
		return err
	}
	for _, assignment := range writes.NewAssignments {
		p.snapshot.Assignments[assignment.ID] = assignment
	}
	return nil
}
func (p *writeProvider) Close() error { return nil }

func assertAssignmentCount(t *testing.T, provider storage.Provider, area domain.Area, want int) {
	t.Helper()
	if err := provider.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
		if got := len(snapshot.Assignments); got != want {
			t.Fatalf("assignment count = %d, want %d", got, want)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
