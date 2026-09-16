package mutation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/lab"
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

// Q-102 asks for "complete current boundary/dependency validation", and the
// checks above are the boundary half: they validate the route being changed and
// nothing that rests on it. A parent's new revision reaches every team beneath
// it — "unchanged child JSON is not proof of unchanged effective reach".
//
// Validating is not freezing. A revision that moves a child's inherited scope is
// a legitimate act by whoever holds the parent. What must not happen is adopting
// a revision a dependent cannot be supported under: left to resolution, that
// surfaces as a read failure for the whole human, long after the write.
func TestUpgradeAssignmentRefusesWhenADependentCannotBeSupported(t *testing.T) {
	// A grandchild of the grant being upgraded, held by a team below Team2, so
	// the adoption under test is the one the lab's gate admits.

	// The control. Without it a refusal below could mean the fixture never
	// supported an upgrade at all, and the rule under test would be untouched.
	t.Run("a revision the dependent can still be supported under", func(t *testing.T) {
		service, area, issuer := upgradeFixture(t, true, nil)
		after, err := service.UpgradeAssignment(t.Context(), area, issuer, "fm5b7t4pan0d")
		if err != nil || after.GrantRevision != 2 {
			t.Fatalf("a supportable adoption was refused: %#v err=%v", after, err)
		}
	})

	t.Run("a revision that drops what the dependent selects", func(t *testing.T) {
		service, area, issuer := upgradeFixture(t, false, nil)
		if _, err := service.UpgradeAssignment(t.Context(), area, issuer, "fm5b7t4pan0d"); !errors.Is(err, domain.ErrRejected) {
			t.Fatalf("adopting a revision a dependent cannot be supported under gave %v, want ErrRejected", err)
		}
		// Q-105: rejected, and left exactly as it was.
		page, err := service.ListAssignments(t.Context(), area, issuer, domain.AssignmentFilter{GrantID: "fk3x9r2man0d"})
		if err != nil || len(page.Assignments) != 1 || page.Assignments[0].GrantRevision != 1 {
			t.Fatalf("the refused upgrade did not leave the assignment unchanged: %#v err=%v", page.Assignments, err)
		}
	})
}

// The guard refuses only for damage this adoption does. Everything here is a
// dependent that is already not resolving, or not resolving for a reason the
// adoption cannot touch — and an adoption must not be refused for any of them.
//
// The first version asked only "does this dependent resolve afterwards?" and so
// refused whenever the answer was no, whatever the reason and whatever the
// adoption. A reviewer showed what that costs: a dependent whose own grant is
// disabled or expired blocked an adoption whose content was identical to what
// was already adopted, and revision content is immutable, so an expired
// dependent froze its ancestor's binding permanently.
func TestUpgradeAssignmentRefusesOnlyForDamageItCauses(t *testing.T) {
	const dependentGrant, dependentBinding = "fk3x9r2mv5k0", "fm5b7t4pv5k0"
	lapsed := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)

	for name, tc := range map[string]struct {
		keepsRead bool
		bend      func(*storage.Snapshot)
	}{
		// The dependent holds nothing, so nothing of its can stop working.
		"the dependent's grant is disabled": {true, func(s *storage.Snapshot) {
			s.Controls[dependentGrant] = domain.GrantControl{Version: "1", ID: dependentGrant, Status: "disabled"}
		}},
		"the dependent's grant has lapsed": {true, func(s *storage.Snapshot) {
			content := s.Contents[domain.GrantKey{ID: dependentGrant, Revision: 1}]
			content.Validity = &domain.Validity{ExpiresAt: &lapsed}
			s.Contents[domain.GrantKey{ID: dependentGrant, Revision: 1}] = content
		}},
		// This one needs the adoption to actually take away what the dependent
		// selects, or the binding's own status is never what decides it: a
		// disabled binding whose chain is otherwise intact resolves in both
		// snapshots, so the differential check alone would let it through.
		"the dependent's binding is disabled, and the adoption drops what it selects": {false, func(s *storage.Snapshot) {
			binding := s.Assignments[dependentBinding]
			binding.Status = "disabled"
			s.Assignments[dependentBinding] = binding
		}},
		"the dependent's binding is disabled": {true, func(s *storage.Snapshot) {
			binding := s.Assignments[dependentBinding]
			binding.Status = "disabled"
			s.Assignments[dependentBinding] = binding
		}},
		// The harshest case: the adoption takes away exactly what the dependent
		// selects, and the dependent was already broken before it. Nothing this
		// write does made it worse.
		"already broken, and the adoption drops what it selects": {false, func(s *storage.Snapshot) {
			s.Controls[dependentGrant] = domain.GrantControl{Version: "1", ID: dependentGrant, Status: "disabled"}
		}},
	} {
		t.Run(name, func(t *testing.T) {
			service, area, issuer := upgradeFixture(t, tc.keepsRead, tc.bend)
			after, err := service.UpgradeAssignment(t.Context(), area, issuer, "fm5b7t4pan0d")
			if err != nil || after.GrantRevision != 2 {
				t.Fatalf("an adoption was refused for a dependent it did not break: %#v err=%v", after, err)
			}
		})
	}

	// And an area broken somewhere this adoption cannot reach. The inventory
	// covers every assignment, so it fails on one in a disjoint branch — which
	// would otherwise mean a single deactivated permission anywhere in a tenant
	// froze every adoption in it.
	t.Run("a disjoint branch is broken", func(t *testing.T) {
		service, area, issuer := upgradeFixture(t, true, func(s *storage.Snapshot) {
			definition := s.Catalog.Permissions[lab.PayslipDelete]
			definition.Active = false
			s.Catalog.Permissions[lab.PayslipDelete] = definition
			s.Controls["fk3x9r2mzzzz"] = domain.GrantControl{Version: "1", ID: "fk3x9r2mzzzz", Status: "enabled"}
			s.Contents[domain.GrantKey{ID: "fk3x9r2mzzzz", Revision: 1}] = domain.GrantContent{
				Version: "1", GrantID: "fk3x9r2mzzzz", Revision: 1, ParentGrantID: "fk3x9r2m0dq3",
				Permissions: []string{lab.PayslipDelete}, Scope: map[string]string{},
			}
			s.Assignments["fm5b7t4pzzzz"] = domain.Assignment{
				Version: "1", ID: "fm5b7t4pzzzz", GrantID: "fk3x9r2mzzzz", GrantRevision: 1,
				Recipient: domain.Recipient{Type: "group", ID: "fibggi2jur5s"}, Status: "enabled",
			}
		})
		after, err := service.UpgradeAssignment(t.Context(), area, issuer, "fm5b7t4pan0d")
		if err != nil || after.GrantRevision != 2 {
			t.Fatalf("breakage in a disjoint branch refused this adoption: %#v err=%v", after, err)
		}
	})

	// The control for the whole set: a dependent that *was* resolving, broken by
	// this adoption and nothing else, is still refused.
	t.Run("broken by this adoption", func(t *testing.T) {
		service, area, issuer := upgradeFixture(t, false, nil)
		if _, err := service.UpgradeAssignment(t.Context(), area, issuer, "fm5b7t4pan0d"); !errors.Is(err, domain.ErrRejected) {
			t.Fatalf("the damage this adoption does gave %v, want ErrRejected", err)
		}
	})

}

func upgradeFixture(t *testing.T, parentKeepsRead bool, bend func(*storage.Snapshot)) (*mutation.Service, domain.Area, domain.Identity) {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	fixture := lab.TeamFINC17(area)
	snapshot := fixture.Snapshot

	// Team2 gains a subteam, and nutan's own team gains a member below her.
	snapshot.Teams["fibggi2jv5k0"] = domain.Team{ID: "fibggi2jv5k0", Name: "fp8h2w6yv5k0", ParentID: "fibggi2juxhc"}
	snapshot.Memberships = append(snapshot.Memberships, domain.Membership{TeamID: "fibggi2jv5k0", HumanID: "fi7io4lvk35s"})

	// Team2's own binding, which is the one that will adopt a new revision.
	snapshot.Assignments["fm5b7t4pan0d"] = domain.Assignment{
		Version: "1", ID: "fm5b7t4pan0d", GrantID: "fk3x9r2man0d", GrantRevision: 1,
		Recipient: domain.Recipient{Type: "group", ID: "fibggi2juxhc"}, Status: "enabled",
	}
	// The grandchild selects read from its parent, and its binding is enabled.
	snapshot.Controls["fk3x9r2mv5k0"] = domain.GrantControl{Version: "1", ID: "fk3x9r2mv5k0", Status: "enabled"}
	snapshot.Contents[domain.GrantKey{ID: "fk3x9r2mv5k0", Revision: 1}] = domain.GrantContent{
		Version: "1", GrantID: "fk3x9r2mv5k0", Revision: 1, ParentGrantID: "fk3x9r2man0d",
		Permissions: []string{lab.PayslipRead}, Scope: map[string]string{},
	}
	snapshot.Assignments["fm5b7t4pv5k0"] = domain.Assignment{
		Version: "1", ID: "fm5b7t4pv5k0", GrantID: "fk3x9r2mv5k0", GrantRevision: 1,
		Recipient: domain.Recipient{Type: "group", ID: "fibggi2jv5k0"}, Status: "enabled",
	}

	// The revision to be adopted. One keeps the permission the grandchild
	// selects; the other drops it, leaving the grandchild unsupportable.
	adopted := []string{lab.PayslipRead}
	if !parentKeepsRead {
		adopted = []string{lab.PayslipWrite}
	}
	snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 2}] = domain.GrantContent{
		Version: "1", GrantID: "fk3x9r2man0d", Revision: 2, ParentGrantID: "fk3x9r2m5iv8",
		Permissions: adopted, Scope: map[string]string{"cert": "C17"},
	}

	if bend != nil {
		bend(&snapshot)
	}
	provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/authority.db", []storage.Snapshot{snapshot})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = provider.Close() })
	statusAdmin, err := lab.NewAssignmentStatusAdministration(area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	admin := &lab.GrantRevisionAdministration{RoleAdministration: &lab.RoleAdministration{AssignmentStatusAdministration: statusAdmin}}
	service, err := mutation.New(provider, admin, &fixedClock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	return service, area, fixture.Issuer
}
