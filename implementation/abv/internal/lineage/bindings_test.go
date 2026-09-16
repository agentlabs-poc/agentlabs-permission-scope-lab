package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/lab"
	"errors"
	"reflect"
	"testing"
	"time"
)

// class is what a caller branches on, and therefore what parity has to mean.
// ErrInactive and ErrIneligible both wrap domain.ErrRejected, so errors.Is
// between two of them is true in both directions — a comparison built on it
// accepts exactly the substitutions it claims to forbid, which is what the first
// version of this test did.
func class(err error) string {
	switch {
	case err == nil:
		return "allowed"
	case errors.Is(err, lineage.ErrIneligible):
		return "ineligible"
	case errors.Is(err, lineage.ErrInactive):
		return "inactive"
	case errors.Is(err, domain.ErrRejected):
		return "rejected"
	case errors.Is(err, domain.ErrMalformed):
		return "malformed"
	default:
		return "other: " + err.Error()
	}
}

// An index is only ever an optimisation, so the property is that it cannot
// change an answer — and the answer includes which refusal.
//
// Every bend below is aimed at the root binding, because that is the only lookup
// the index serves: the first step of a chain is resolved by a scan on both
// paths, so a bend anywhere else is answered before the index is reached and
// proves nothing about it. Six of the eight bends in the first version of this
// test were that shape.
func TestAnIndexedResolveAnswersExactlyAsAnUnindexedOne(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	const rootBinding, rootGrant, rootTeam = "fm5b7t4p0dq3", "fk3x9r2m0dq3", "fibggi2jur5s"

	copyOf := func(f *lab.TeamFINC17Case, id string) domain.Assignment {
		duplicate := f.Snapshot.Assignments[rootBinding]
		duplicate.ID = id
		return duplicate
	}

	for name, bend := range map[string]func(*lab.TeamFINC17Case){
		"the worked fixture": func(*lab.TeamFINC17Case) {},

		// Two rows binding the same pair. A scan counts both and refuses.
		"the root binding duplicated": func(f *lab.TeamFINC17Case) {
			f.Snapshot.Assignments["fm5b7t4p0dq3two"] = copyOf(f, "fm5b7t4p0dq3two")
		},
		// The divergence that used to be an allow: the second row binding the
		// pair is invalid, so the index skipped it and handed back the good one
		// while a scan refused the route outright.
		"the root binding duplicated by an invalid row": func(f *lab.TeamFINC17Case) {
			broken := copyOf(f, "fm5b7t4p0dq3bad")
			broken.Status = "revoked"
			f.Snapshot.Assignments[broken.ID] = broken
		},
		"the root binding is itself invalid": func(f *lab.TeamFINC17Case) {
			broken := f.Snapshot.Assignments[rootBinding]
			broken.Status = "revoked"
			f.Snapshot.Assignments[rootBinding] = broken
		},
		// A pair this chain never touches, made unusable *without* a duplicate.
		// The refusal-too-wide divergence had two causes and only the duplicate
		// one was defended: poisoning the area on an invalid row survived every
		// test here, because no bend put an invalid row anywhere but on the pair
		// under resolution.
		"an unrelated pair has an invalid row": func(f *lab.TeamFINC17Case) {
			broken := copyOf(f, "fm5b7t4pelsebad")
			broken.GrantID = "fk3x9r2mstray"
			broken.Recipient = domain.Recipient{Type: "group", ID: "fibggi2jv0n4"}
			broken.Status = "revoked"
			f.Snapshot.Assignments[broken.ID] = broken
		},
		// The divergence that used to be a refusal: a duplicate on a pair this
		// chain never touches. A scan refuses only the route that owns it.
		"an unrelated pair duplicated": func(f *lab.TeamFINC17Case) {
			for _, id := range []string{"fm5b7t4pelse1", "fm5b7t4pelse2"} {
				stray := copyOf(f, id)
				stray.GrantID = "fk3x9r2mstray"
				stray.Recipient = domain.Recipient{Type: "group", ID: "fibggi2jv0n4"}
				f.Snapshot.Assignments[id] = stray
			}
		},
		// A human-held row binding the same grant as the root's team-held one.
		// A scan ignores it because it filters on the recipient type; an index
		// that did not filter would record it and answer with it.
		"a human holds the root grant too": func(f *lab.TeamFINC17Case) {
			direct := copyOf(f, "fm5b7t4p0dq3usr")
			direct.Recipient = domain.Recipient{Type: "user", ID: rootTeam}
			f.Snapshot.Assignments[direct.ID] = direct
		},
		"the root binding removed": func(f *lab.TeamFINC17Case) {
			delete(f.Snapshot.Assignments, rootBinding)
		},
		"the root binding disabled": func(f *lab.TeamFINC17Case) {
			disabled := f.Snapshot.Assignments[rootBinding]
			disabled.Status = "disabled"
			f.Snapshot.Assignments[rootBinding] = disabled
		},
		// Mis-keyed, and binding a pair nothing else binds, so only the key
		// check can catch it. A scan reads every row, so it refuses whatever it
		// was asked about.
		"an unrelated row keyed by something other than its id": func(f *lab.TeamFINC17Case) {
			stray := copyOf(f, rootBinding)
			stray.GrantID = "fk3x9r2mstray"
			stray.Recipient = domain.Recipient{Type: "group", ID: "fibggi2jv0n4"}
			f.Snapshot.Assignments["fm5b7t4pnotitsid"] = stray
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := lab.TeamFINC17(area)
			bend(&f)
			child := f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1}]

			plain, plainErr := lineage.ResolveParentTeam(f.Snapshot, child, "fibggi2juxhc", now)
			indexed, indexedErr := lineage.ResolveParentTeamIndexed(
				f.Snapshot, lineage.IndexBindings(f.Snapshot), child, "fibggi2juxhc", now)

			if class(plainErr) != class(indexedErr) {
				t.Fatalf("scan answered %s (%v), index answered %s (%v)",
					class(plainErr), plainErr, class(indexedErr), indexedErr)
			}
			// Whole routes, not their lengths: a route that agreed on how many
			// predicates it had and disagreed on what they said would have
			// passed the first version of this.
			if !reflect.DeepEqual(plain, indexed) {
				t.Fatalf("scan gave %#v, index gave %#v", plain, indexed)
			}
		})
	}
}

// Every bend above must actually be a bend — one that changes nothing proves the
// two paths agree about the fixture and nothing more. These are the answers, so
// a bend that stops mattering shows up here rather than passing quietly.
func TestTheParityBendsReachTheIndexAndChangeTheAnswer(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	f := lab.TeamFINC17(area)
	child := f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1}]
	if _, err := lineage.ResolveParentTeamIndexed(f.Snapshot, lineage.IndexBindings(f.Snapshot), child, "fibggi2juxhc", now); err != nil {
		t.Fatalf("the worked fixture does not resolve, so every agreement above is two refusals agreeing: %v", err)
	}

	// The index is consulted, and this is what says so: an index built from a
	// snapshot the root binding has been deleted from must refuse, against the
	// snapshot where it is still present. If binding() ignored its index — or
	// fell back to a scan on a miss, as it used to — this would answer allow.
	stale := lab.TeamFINC17(area)
	delete(stale.Snapshot.Assignments, "fm5b7t4p0dq3")
	if _, err := lineage.ResolveParentTeamIndexed(f.Snapshot, lineage.IndexBindings(stale.Snapshot), child, "fibggi2juxhc", now); err == nil {
		t.Fatal("an index that does not describe the snapshot was not consulted at all")
	}

	// And the shape a caller actually risks: two snapshots differing in one
	// binding's revision, with the index built over the wrong one. That is the
	// pairing the adoption guard maintains by hand, and it is pinned here
	// because the guard itself cannot show it — a dependent deep enough to read
	// that entry through the index cannot break without its parent breaking
	// first, and the parent's first chain step is a scan.
	moved := lab.TeamFINC17(area)
	rebound := moved.Snapshot.Assignments["fm5b7t4p0dq3"]
	rebound.GrantRevision = 2
	moved.Snapshot.Assignments["fm5b7t4p0dq3"] = rebound
	withRight, rightErr := lineage.ResolveParentTeamIndexed(moved.Snapshot, lineage.IndexBindings(moved.Snapshot), child, "fibggi2juxhc", now)
	withWrong, wrongErr := lineage.ResolveParentTeamIndexed(moved.Snapshot, lineage.IndexBindings(f.Snapshot), child, "fibggi2juxhc", now)
	if class(rightErr) == class(wrongErr) && reflect.DeepEqual(withRight, withWrong) {
		t.Fatalf("an index built over the other snapshot answered identically (%v), so the pairing is unobservable and nothing protects it", rightErr)
	}

	// The index's own refusal on a mis-keyed row, which no resolve can reach:
	// the first chain step scans, and a scan refuses such a row area-wide before
	// the index is consulted. Building the index from a snapshot that has one
	// and resolving against a snapshot that does not is the only way to ask the
	// index the question directly — and it must answer the same way the scan
	// would have, or the check is decoration.
	misKeyed := lab.TeamFINC17(area)
	stray := misKeyed.Snapshot.Assignments["fm5b7t4p0dq3"]
	stray.GrantID = "fk3x9r2mstray"
	stray.Recipient = domain.Recipient{Type: "group", ID: "fibggi2jv0n4"}
	misKeyed.Snapshot.Assignments["fm5b7t4pnotitsid"] = stray
	// class, not errors.Is. This line used errors.Is(err, domain.ErrRejected),
	// which ErrInactive satisfies — so the one assertion covering the one branch
	// no resolve can reach accepted exactly the substitution this file exists to
	// forbid. A review found it by swapping the branch to ErrInactive and
	// watching the suite stay green.
	if _, err := lineage.ResolveParentTeamIndexed(f.Snapshot, lineage.IndexBindings(misKeyed.Snapshot), child, "fibggi2juxhc", now); class(err) != "rejected" {
		t.Fatalf("an index built over a mis-keyed row answered %s (%v), want a rejection", class(err), err)
	}
}
