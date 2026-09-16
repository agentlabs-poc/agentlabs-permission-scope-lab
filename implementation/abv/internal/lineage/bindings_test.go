package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/lab"
	"errors"
	"testing"
	"time"
)

// An index is only ever an optimisation, so the property that matters is that it
// cannot change an answer. Every chain in the worked fixture is resolved both
// ways and the two must agree — including where the answer is a refusal, because
// the index has its own view of what makes an area's assignments untrustworthy
// and a disagreement there would be a silent allow or a silent deny.
func TestAnIndexedResolveAnswersExactlyAsAnUnindexedOne(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)

	for name, bend := range map[string]func(*lab.TeamFINC17Case){
		"the worked fixture": func(*lab.TeamFINC17Case) {},
		// Deeper than the first step. The first is resolved by a scan in both
		// paths, so a duplicate there is caught identically either way and says
		// nothing about the index — this one is the root's own binding, which
		// only the indexed lookup serves.
		"a duplicate binding deeper in the chain": func(f *lab.TeamFINC17Case) {
			duplicate := f.Snapshot.Assignments["fm5b7t4p0dq3"]
			duplicate.ID = "fm5b7t4p0dq3copy"
			f.Snapshot.Assignments[duplicate.ID] = duplicate
		},
		// Mis-keyed *and* binding a pair nothing else binds, so the duplicate
		// check cannot catch it and the key check is the only thing that can. A
		// scan sees every row, so it refuses on a mis-keyed row anywhere in the
		// area; the index has to agree, or the two disagree on a store that is
		// already untrustworthy.
		"an unrelated row keyed by something other than its id": func(f *lab.TeamFINC17Case) {
			stray := f.Snapshot.Assignments["fm5b7t4p0dq3"]
			stray.GrantID = "fk3x9r2mstray"
			stray.Recipient = domain.Recipient{Type: "group", ID: "fibggi2jv0n4"}
			f.Snapshot.Assignments["fm5b7t4pnotitsid"] = stray
		},
		"a duplicate binding": func(f *lab.TeamFINC17Case) {
			duplicate := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			duplicate.ID = "fm5b7t4p5iv8copy"
			f.Snapshot.Assignments[duplicate.ID] = duplicate
		},
		"a row keyed by something other than its id": func(f *lab.TeamFINC17Case) {
			f.Snapshot.Assignments["fm5b7t4pwrong"] = f.Snapshot.Assignments["fm5b7t4p5iv8"]
		},
		"the supporting binding removed": func(f *lab.TeamFINC17Case) {
			delete(f.Snapshot.Assignments, "fm5b7t4p5iv8")
		},
		"the supporting binding disabled": func(f *lab.TeamFINC17Case) {
			disabled := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			disabled.Status = "disabled"
			f.Snapshot.Assignments["fm5b7t4p5iv8"] = disabled
		},
		"a binding held by a human rather than a team": func(f *lab.TeamFINC17Case) {
			direct := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			direct.ID = "fm5b7t4pdirect"
			direct.Recipient = domain.Recipient{Type: "user", ID: "fi7io4lvjqio"}
			f.Snapshot.Assignments[direct.ID] = direct
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := lab.TeamFINC17(area)
			bend(&f)
			child := f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1}]

			plain, plainErr := lineage.ResolveParentTeam(f.Snapshot, child, "fibggi2juxhc", now)
			indexed, indexedErr := lineage.ResolveParentTeamIndexed(
				f.Snapshot, lineage.IndexBindings(f.Snapshot), child, "fibggi2juxhc", now)

			switch {
			case plainErr == nil && indexedErr != nil:
				t.Fatalf("the index refused what the scan allowed: %v", indexedErr)
			case plainErr != nil && indexedErr == nil:
				t.Fatalf("the index allowed what the scan refused: %v", plainErr)
			case plainErr != nil:
				// Same kind, not merely both failing — the caller branches on it.
				if !errors.Is(indexedErr, plainErr) && !errors.Is(plainErr, indexedErr) {
					t.Fatalf("scan gave %v, index gave %v", plainErr, indexedErr)
				}
				return
			}
			if plain.GrantID != indexed.GrantID || len(plain.Permissions) != len(indexed.Permissions) ||
				len(plain.Predicates) != len(indexed.Predicates) || len(plain.AssignmentIDs) != len(indexed.AssignmentIDs) {
				t.Fatalf("scan gave %#v, index gave %#v", plain, indexed)
			}
			for i := range plain.AssignmentIDs {
				if plain.AssignmentIDs[i] != indexed.AssignmentIDs[i] {
					t.Fatalf("the chains differ: %v vs %v", plain.AssignmentIDs, indexed.AssignmentIDs)
				}
			}
		})
	}

	// And the fixture really does resolve, or every agreement above is two
	// refusals agreeing with each other.
	f := lab.TeamFINC17(area)
	child := f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1}]
	if route, err := lineage.ResolveParentTeamIndexed(f.Snapshot, lineage.IndexBindings(f.Snapshot), child, "fibggi2juxhc", now); err != nil || len(route.AssignmentIDs) == 0 {
		t.Fatalf("the worked fixture does not resolve: %#v, %v", route, err)
	}
}
