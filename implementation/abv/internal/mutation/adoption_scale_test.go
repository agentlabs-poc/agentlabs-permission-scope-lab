//go:build !race

// Timing assertions are meaningless under the race detector, which is roughly
// fifteen times slower. This file measures; -race proves correctness elsewhere.

package mutation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/lab"
	"fmt"
	"testing"
	"time"
)

// Adoption revalidates every enabled binding beneath the one being upgraded, and
// each of those is a chain walk. So the cost is the fan-out times the depth, and
// it runs inside the write transaction — where a tenant's whole area is already
// held.
//
// A review measured this before the guard became differential and found it
// super-linear: at 1600 dependents it was already at the 400 ms budget that
// TestDeleteGrantStaysBoundedAsChildrenGrow sets for the comparable path, with
// nothing watching. This is the tripwire that was missing.
//
// The differential check does not change the healthy path — the second walk runs
// only for a dependent that fails the first — which is the property worth keeping
// and the reason the accepting direction is measured too. Accepting is the
// expensive direction here, because refusing stops at the first dependent it
// cannot support while accepting has to look at all of them.
//
// The walk used to rescan every assignment in the area at each chain step, and
// there is a chain per dependent, so doubling the fan-out more than doubled the
// work: 19.5 / 54.8 / 132 / 328 ms at 200 / 400 / 800 / 1600, about 2.4–2.8×
// per doubling. A Bindings index built once per call and shared across
// dependents makes it 16.7 / 44.2 / 86.5 / 185 ms — about 2.1×, which is as
// close to linear as this shape gets.
//
// Worth recording because the obvious version of that fix was slower than the
// problem: an index built per chain cost ~434 ms at 800, against ~150 ms for the
// scans it replaced. Chains are shallow and areas are wide, so the index only
// pays when it is shared.
func TestAdoptionStaysBoundedAsDependentsGrow(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	const dependents = 800

	build := func(t *testing.T, keepsRead bool) *mutation.Service {
		t.Helper()
		fixture := lab.TeamFINC17(area)
		snapshot := fixture.Snapshot
		snapshot.Assignments["fm5b7t4pan0d"] = domain.Assignment{
			Version: "1", ID: "fm5b7t4pan0d", GrantID: "fk3x9r2man0d", GrantRevision: 1,
			Recipient: domain.Recipient{Type: "group", ID: "fibggi2juxhc"}, Status: "enabled",
		}
		// Wide fan-out: one team and one grant per dependent, all hanging from
		// the binding being upgraded.
		for i := range dependents {
			team := fmt.Sprintf("fibggi2j%04x", i)
			grant := fmt.Sprintf("fk3x9r2m%04x", i)
			assignment := fmt.Sprintf("fm5b7t4p%04x", i)
			snapshot.Teams[team] = domain.Team{ID: team, Name: fmt.Sprintf("fp8h2w6y%04x", i), ParentID: "fibggi2juxhc"}
			snapshot.Memberships = append(snapshot.Memberships, domain.Membership{TeamID: team, HumanID: "fi7io4lvk35s"})
			snapshot.Controls[grant] = domain.GrantControl{Version: "1", ID: grant, Status: "enabled"}
			snapshot.Contents[domain.GrantKey{ID: grant, Revision: 1}] = domain.GrantContent{
				Version: "1", GrantID: grant, Revision: 1, ParentGrantID: "fk3x9r2man0d",
				Permissions: []string{lab.PayslipRead}, Scope: map[string]string{},
			}
			snapshot.Assignments[assignment] = domain.Assignment{
				Version: "1", ID: assignment, GrantID: grant, GrantRevision: 1,
				Recipient: domain.Recipient{Type: "group", ID: team}, Status: "enabled",
			}
		}
		adopted := []string{lab.PayslipRead}
		if !keepsRead {
			adopted = []string{lab.PayslipWrite}
		}
		snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 2}] = domain.GrantContent{
			Version: "1", GrantID: "fk3x9r2man0d", Revision: 2, ParentGrantID: "fk3x9r2m5iv8",
			Permissions: adopted, Scope: map[string]string{"cert": "C17"},
		}
		provider, err := lab.CreateSQLite(t.Context(), t.TempDir()+"/scale.db", []storage.Snapshot{snapshot})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = provider.Close() })
		statusAdmin, err := lab.NewAssignmentStatusAdministration(area, fixture.Administration)
		if err != nil {
			t.Fatal(err)
		}
		service, err := mutation.New(provider,
			&lab.GrantRevisionAdministration{RoleAdministration: &lab.RoleAdministration{AssignmentStatusAdministration: statusAdmin}},
			&fixedClock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)})
		if err != nil {
			t.Fatal(err)
		}
		return service
	}
	issuer := lab.TeamFINC17(area).Issuer

	// Built before the clock starts. Seeding 800 teams into SQLite costs more
	// than the operation does, and timing it would measure the fixture.
	accepting, refusing := build(t, true), build(t, false)

	accept := time.Now()
	if after, err := accepting.UpgradeAssignment(t.Context(), area, issuer, "fm5b7t4pan0d"); err != nil || after.GrantRevision != 2 {
		t.Fatalf("the accepting direction did not adopt: %#v err=%v", after, err)
	}
	accepted := time.Since(accept)

	// Refusing is the cheaper direction, not the dearer one: it stops at the
	// first dependent it cannot support, while accepting has to look at all of
	// them. It is measured because it pays the same snapshot load, which bounds
	// how much of the accepting number is fixture rather than work.
	refuse := time.Now()
	if _, err := refusing.UpgradeAssignment(t.Context(), area, issuer, "fm5b7t4pan0d"); err == nil {
		t.Fatal("the refusing direction adopted")
	}
	refused := time.Since(refuse)

	t.Logf("accept %s · refuse %s · at %d dependents",
		accepted.Round(time.Microsecond), refused.Round(time.Microsecond), dependents)

	// The same generous ceiling the comparable write path uses. Not a benchmark
	// — a tripwire for the day this stops being passes over a snapshot that is
	// already in memory.
	const budget = 400 * time.Millisecond
	if accepted > budget || refused > budget {
		t.Fatalf("accept=%s refuse=%s, both should stay under %s", accepted, refused, budget)
	}
}
