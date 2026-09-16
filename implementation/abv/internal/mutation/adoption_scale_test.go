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
// What this is *not* is a claim that the cost is linear. Measured on this
// machine: ~150 ms accepting at 800 dependents, and the review measured ~379 ms
// at 1600 — the walk rescans the assignment map per chain step per dependent, so
// doubling the fan-out more than doubles the work. The budget sits where it does
// to leave that headroom visible rather than to bless it. If the day comes that
// this fails, the fix is an index built once per call, not a bigger number here.
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

	// The refusing direction is the expensive one: every dependent fails the
	// first walk, so every one of them takes the second as well.
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
