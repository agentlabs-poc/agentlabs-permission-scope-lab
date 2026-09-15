package mutation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"errors"
	"testing"
	"time"
)

// Q-105: an explicit upgrade "must also select the latest, and pass current
// checks", and "if its permitted authority cannot support revision 3, reject
// the upgrade and leave the assignment unchanged."
//
// The administrative gate answers who may act. These answer whether the
// authority actually holds — and they are the same checks CreateAssignment runs,
// because adopting a revision is the same act whether at creation or later.
// Without them an upgrade would adopt a revision the recipient has no route to.
func TestUpgradeAssignmentRunsTheSameSupportChecksAsCreation(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)

	// The child grant gets a second revision, so there is something to upgrade
	// to; then the parent team's supporting assignment is taken away.
	fixture := lab.TeamFINC17(area)
	child := fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1}]
	newer := child
	newer.Revision = 2
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 2}] = newer
	fixture.Snapshot.Assignments["fm5b7t4pan0d"] = domain.Assignment{
		Version: "1", ID: "fm5b7t4pan0d", GrantID: "fk3x9r2man0d", GrantRevision: 1,
		Recipient: domain.Recipient{Type: "group", ID: "fibggi2juxhc"}, Status: "enabled",
	}
	delete(fixture.Snapshot.Assignments, "fm5b7t4p5iv8")

	provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	// The record operations are gated by RoleAdministration, which is what
	// lab.Connect composes. lab.Administration gates assignment *creation* only,
	// so using it here would return ErrUnsupported before reaching the checks.
	statusAdmin, err := lab.NewAssignmentStatusAdministration(area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	admin := &lab.GrantRevisionAdministration{RoleAdministration: &lab.RoleAdministration{AssignmentStatusAdministration: statusAdmin}}
	service, err := mutation.New(provider, admin, &fixedClock{now: now})
	if err != nil {
		t.Fatal(err)
	}

	_, upErr := service.UpgradeAssignment(t.Context(), area, fixture.Issuer, "fm5b7t4pan0d")
	if upErr == nil {
		t.Fatal("the upgrade was accepted with no supporting route")
	}
	// It must fail on the missing support, not because the gate is absent —
	// ErrUnsupported would mean the operation never reached these checks.
	if errors.Is(upErr, domain.ErrUnsupported) {
		t.Fatalf("the upgrade never reached its support checks: %v", upErr)
	}

	// And it left the assignment exactly as it was — not partially applied.
	if err := provider.Read(t.Context(), area, func(s storage.Snapshot) error {
		if got := s.Assignments["fm5b7t4pan0d"]; got.GrantRevision != 1 {
			t.Fatalf("after a refused upgrade the assignment adopts revision %d, want 1", got.GrantRevision)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
