package abv_test

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/lab"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type clock struct{ now time.Time }

func (c clock) Now() time.Time { return c.now }

type externalAdministration struct{ area domain.Area }

func (a externalAdministration) CheckAssignment(_ context.Context, evidence abv.Evidence, identity domain.Identity, proposed domain.Assignment, _ time.Time) error {
	if evidence.Area == a.area && identity.Actor.Type == "user" && identity.Actor.ID == "fi7io4lvjqio" && identity.HumanID == "fi7io4lvjqio" && proposed.Recipient == (domain.Recipient{Type: "group", ID: "fibggi2juxhc"}) {
		return nil
	}
	return domain.ErrRejected
}

func TestFacadeAdministrativePortIsImplementableOutsideInternalPackages(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	provider := &memoryProvider{snapshot: fixture.Snapshot}
	facade, err := abv.New(provider, externalAdministration{area: area}, clock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = facade.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed); err != nil {
		t.Fatal(err)
	}
}

func TestOldAdministrationCannotAuthorizeGrantStatus(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	p := &memoryProvider{snapshot: fixture.Snapshot}
	facade, err := abv.New(p, externalAdministration{area: area}, clock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	proposed := domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "disabled"}
	got, err := facade.SetGrantStatus(t.Context(), area, fixture.Issuer, proposed)
	if !errors.Is(err, domain.ErrUnsupported) || got != (domain.GrantControl{}) {
		t.Fatalf("old adapter acquired authority: %#v, %v", got, err)
	}
	if p.snapshot.Controls["fk3x9r2man0d"].Status != "enabled" {
		t.Fatal("unauthorized write")
	}
}

func TestOldAdministrationCannotAuthorizeAssignmentStatus(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	p := &memoryProvider{snapshot: fixture.Snapshot}
	facade, err := abv.New(p, externalAdministration{area: area}, clock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	got, err := facade.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "fm5b7t4p5iv8", "disabled")
	if !errors.Is(err, domain.ErrUnsupported) || got != (domain.Assignment{}) {
		t.Fatalf("old adapter gained assignment control: %#v, %v", got, err)
	}
}

func TestFacadeInspectsCanonicalRecordsAndDiagnosesWithoutWriting(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	admin, _ := lab.NewAdministration(area, fixture.Administration)
	facade, err := abv.New(provider, admin, clock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	defer facade.Close()

	record, err := facade.Inspect(t.Context(), area, "assignment", "fm5b7t4p5iv8")
	if err != nil {
		t.Fatal(err)
	}
	wantJSON, _ := json.Marshal(fixture.Snapshot.Assignments["fm5b7t4p5iv8"])
	if record.Area != area || record.Kind != "assignment" || record.ID != "fm5b7t4p5iv8" || !bytes.Equal(record.CanonicalJSON, wantJSON) {
		t.Fatalf("record = %#v, want exact area-bound canonical fm5b7t4p5iv8", record)
	}

	raw, _ := json.Marshal(fixture.Proposed)
	diagnostic, err := facade.CheckAssignment(t.Context(), area, raw)
	if err != nil {
		t.Fatal(err)
	}
	if diagnostic.Route == nil || diagnostic.Route.GrantID != "fk3x9r2man0d" || diagnostic.Summary != "proposal is structurally and lineally valid; administrative and source authority are not established" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	assertCount(t, provider, area, 2)
}

func TestInspectCarriesAreaForEverySupportedKind(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	provider := &memoryProvider{snapshot: fixture.Snapshot}
	admin, _ := lab.NewAdministration(area, fixture.Administration)
	facade, _ := abv.New(provider, admin, clock{now: time.Now()})
	for _, test := range []struct {
		kind, id  string
		canonical bool
	}{
		{kind: "scope", id: "dept"},
		{kind: "role", id: "fi9jvxobqsxs"},
		{kind: "grant", id: "fk3x9r2m5iv8", canonical: true},
		{kind: "assignment", id: "fm5b7t4p5iv8", canonical: true},
		{kind: "team", id: "fibggi2juubk"},
		{kind: "membership", id: "fi7io4lvjqio"},
	} {
		t.Run(test.kind, func(t *testing.T) {
			record, err := facade.Inspect(t.Context(), area, test.kind, test.id)
			if err != nil {
				t.Fatal(err)
			}
			if record.Area != area || record.Kind != test.kind || record.ID != test.id {
				t.Fatalf("record boundary = %#v", record)
			}
			if test.canonical != (len(record.CanonicalJSON) > 0) {
				t.Fatalf("canonical presence = %v", len(record.CanonicalJSON) > 0)
			}
			if !test.canonical && len(record.Rows) == 0 {
				t.Fatalf("unlabeled/empty table = %#v", record.Rows)
			}
		})
	}
}

func TestSQLiteFacadeReopensCommittedAssignment(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	path := t.TempDir() + "/authority.db"
	provider, err := lab.CreateSQLite(t.Context(), path, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	admin, _ := lab.NewAdministration(area, fixture.Administration)
	facade, _ := abv.New(provider, admin, clock{now: time.Now()})
	if _, err = facade.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed); err != nil {
		t.Fatal(err)
	}
	if err = facade.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := abv.OpenSQLite(t.Context(), path, admin, clock{now: time.Now()}, mustFixedRegistry(t, area))
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	record, err := reopened.Inspect(t.Context(), area, "assignment", "fm5b7t4pan0d")
	if err != nil || record.ID != "fm5b7t4pan0d" || len(record.CanonicalJSON) == 0 {
		t.Fatalf("reopened record=%#v err=%v", record, err)
	}
}

func TestFacadeDiagnosticIsNotATicketForAssignment(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	fixture := lab.TeamFINC17(area)
	provider := &memoryProvider{snapshot: fixture.Snapshot}
	admin, _ := lab.NewAdministration(area, fixture.Administration)
	facade, err := abv.New(provider, admin, clock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(fixture.Proposed)
	if _, err := facade.CheckAssignment(t.Context(), area, raw); err != nil {
		t.Fatal(err)
	}
	control := provider.snapshot.Controls["fk3x9r2m5iv8"]
	control.Status = "disabled"
	provider.snapshot.Controls["fk3x9r2m5iv8"] = control
	if receipt, err := facade.CreateAssignment(t.Context(), area, fixture.Issuer, fixture.Proposed); !errors.Is(err, domain.ErrRejected) || receipt != (domain.Receipt{}) {
		t.Fatalf("stale diagnosis yielded receipt=%#v err=%v", receipt, err)
	}
}

type memoryProvider struct{ snapshot storage.Snapshot }

func (p *memoryProvider) Read(_ context.Context, _ domain.Area, callback func(storage.Snapshot) error) error {
	return callback(p.snapshot)
}
func (p *memoryProvider) Update(_ context.Context, _ domain.Area, callback func(storage.Snapshot) (storage.WriteSet, error)) error {
	writes, err := callback(p.snapshot)
	if err != nil {
		return err
	}
	for _, assignment := range writes.NewAssignments {
		p.snapshot.Assignments[assignment.ID] = assignment
	}
	if writes.GrantStatusChange != nil {
		p.snapshot.Controls[writes.GrantStatusChange.After.ID] = writes.GrantStatusChange.After
	}
	return nil
}

// UpdateAdministered delegates: a fake's snapshot is the whole world it has, so
// it is its own administrative chain, and a nil Snapshot.Administrative says
// exactly that.
func (p *memoryProvider) UpdateAdministered(ctx context.Context, area domain.Area, callback func(storage.Snapshot) (storage.WriteSet, error)) error {
	return p.Update(ctx, area, callback)
}
func (p *memoryProvider) Close() error { return nil }

func assertCount(t *testing.T, provider storage.Provider, area domain.Area, want int) {
	t.Helper()
	if err := provider.Read(t.Context(), area, func(snapshot storage.Snapshot) error {
		if len(snapshot.Assignments) != want {
			t.Fatalf("count = %d, want %d", len(snapshot.Assignments), want)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// mustFixedRegistry composes the smallest thing that satisfies the port, now
// that Auth-AL keeps no copy of the registry's facts and cannot be opened
// without one.
func mustFixedRegistry(t *testing.T, area domain.Area) *lab.FixedRegistry {
	t.Helper()
	registry, err := lab.NewFixedRegistry(area)
	if err != nil {
		t.Fatal(err)
	}
	return registry
}
