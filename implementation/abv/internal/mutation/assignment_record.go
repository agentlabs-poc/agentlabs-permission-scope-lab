package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// AssignmentRecordAdministration gates reading an area's assignments, removing
// one, and adopting a newer grant revision into one.
//
// Adoption is its own authority rather than a status change: Q-105 makes an
// upgrade a fresh selection that must pass current checks, which is closer to
// creating an assignment than to enabling one.
type AssignmentRecordAdministration interface {
	CheckAssignmentRead(context.Context, domain.Area, domain.Identity, time.Time) error
	CheckAssignmentDelete(context.Context, domain.Area, domain.Identity, string, time.Time) error
	CheckAssignmentAdoption(context.Context, domain.Area, domain.Identity, domain.Assignment, time.Time) error
}

func (s *Service) assignmentRecordAdministration(ctx context.Context, area domain.Area, identity domain.Identity) (AssignmentRecordAdministration, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := area.Validate(); err != nil {
		return nil, err
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return nil, err
	}
	admin, ok := s.administration.(AssignmentRecordAdministration)
	if !ok || nilInterface(admin) {
		return nil, domain.ErrUnsupported
	}
	return admin, nil
}

func validAssignmentID(id string) bool {
	return strings.TrimSpace(id) != "" && !strings.Contains(id, "*") && utf8.ValidString(id)
}

func (s *Service) GetAssignment(ctx context.Context, area domain.Area, identity domain.Identity, id string) (domain.Assignment, error) {
	fail := func(err error) (domain.Assignment, error) { return domain.Assignment{}, err }
	admin, err := s.assignmentRecordAdministration(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if !validAssignmentID(id) {
		return fail(domain.ErrMalformed)
	}
	var result domain.Assignment
	err = s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area {
			return domain.ErrRejected
		}
		if err := admin.CheckAssignmentRead(ctx, area, identity, s.clock.Now()); err != nil {
			return err
		}
		found, ok := snapshot.Assignments[id]
		if !ok || found.ID != id {
			return domain.ErrNotFound
		}
		result = found
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return result, nil
}

// ListAssignments answers in both directions and refuses to answer neither.
func (s *Service) ListAssignments(ctx context.Context, area domain.Area, identity domain.Identity, filter domain.AssignmentFilter) (domain.AssignmentPage, error) {
	fail := func(err error) (domain.AssignmentPage, error) { return domain.AssignmentPage{}, err }
	admin, err := s.assignmentRecordAdministration(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if (filter.GrantID == "") == (filter.Recipient == nil) {
		return fail(domain.ErrMalformed)
	}
	if filter.GrantID != "" && !validAssignmentID(filter.GrantID) {
		return fail(domain.ErrMalformed)
	}
	if filter.Recipient != nil {
		r := *filter.Recipient
		if (r.Type != "user" && r.Type != "group") || !validAssignmentID(r.ID) {
			return fail(domain.ErrMalformed)
		}
	}
	if filter.Offset < 0 || filter.Limit < 0 || filter.Limit > maxGrantPage {
		return fail(domain.ErrMalformed)
	}
	if filter.Status != "" && filter.Status != "enabled" && filter.Status != "disabled" {
		return fail(domain.ErrMalformed)
	}
	var page domain.AssignmentPage
	err = s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area {
			return domain.ErrRejected
		}
		if err := admin.CheckAssignmentRead(ctx, area, identity, s.clock.Now()); err != nil {
			return err
		}
		matched := make([]domain.Assignment, 0, len(snapshot.Assignments))
		for id, a := range snapshot.Assignments {
			if err := ctx.Err(); err != nil {
				return err
			}
			if a.ID != id {
				return domain.ErrRejected
			}
			if filter.GrantID != "" && a.GrantID != filter.GrantID {
				continue
			}
			if filter.Recipient != nil && a.Recipient != *filter.Recipient {
				continue
			}
			if filter.Status != "" && a.Status != filter.Status {
				continue
			}
			matched = append(matched, a)
		}
		// Ordered by the binding, not the id, because that is the record's
		// identity and the order a reader of the rows would see.
		sort.Slice(matched, func(i, j int) bool {
			if matched[i].GrantID != matched[j].GrantID {
				return matched[i].GrantID < matched[j].GrantID
			}
			if matched[i].Recipient.Type != matched[j].Recipient.Type {
				return matched[i].Recipient.Type < matched[j].Recipient.Type
			}
			return matched[i].Recipient.ID < matched[j].Recipient.ID
		})
		page = domain.AssignmentPage{Assignments: grantPage(matched, filter.Offset, filter.Limit), Total: len(matched)}
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return page, nil
}

// DeleteAssignment removes a route permanently. Q-101 allows it — "relevant
// bindings may be removed OR disabled" — and Q-104 counts only current
// assignments, so a removed one stops blocking a fresh binding where a disabled
// one would not.
//
// It refuses while a child grant's route depends on this one: a dependent
// assignment's parent support is resolved through the parent team's assignment
// of the parent grant, so removing that is removing the support underneath it.
func (s *Service) DeleteAssignment(ctx context.Context, area domain.Area, identity domain.Identity, id string) error {
	admin, err := s.assignmentRecordAdministration(ctx, area, identity)
	if err != nil {
		return err
	}
	if !validAssignmentID(id) {
		return domain.ErrMalformed
	}
	return s.provider.Update(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area {
			return storage.WriteSet{}, domain.ErrRejected
		}
		if err := admin.CheckAssignmentDelete(ctx, area, identity, id, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		target, ok := snapshot.Assignments[id]
		if !ok || target.ID != id {
			return storage.WriteSet{}, domain.ErrNotFound
		}

		for _, other := range snapshot.Assignments {
			if other.ID == id {
				continue
			}
			content, ok := snapshot.Contents[domain.GrantKey{ID: other.GrantID, Revision: other.GrantRevision}]
			if ok && content.ParentGrantID == target.GrantID {
				return storage.WriteSet{}, domain.ErrConflict
			}
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		return storage.WriteSet{RemovedAssignment: id}, nil
	})
}

// UpgradeAssignment is Q-104's "authorized adoption operation", governed by
// Q-105: it selects the LATEST published revision, never an intermediate, and
// when the latest cannot be supported it rejects and leaves the assignment
// exactly as it was. Falling back to a revision that would pass is the specific
// behaviour the handbook forbids.
//
// It is not re-enablement, and re-enablement is not it: a disabled assignment
// keeps its adopted revision when enabled.
func (s *Service) UpgradeAssignment(ctx context.Context, area domain.Area, identity domain.Identity, id string) (domain.Assignment, error) {
	fail := func(err error) (domain.Assignment, error) { return domain.Assignment{}, err }
	admin, err := s.assignmentRecordAdministration(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if !validAssignmentID(id) {
		return fail(domain.ErrMalformed)
	}
	var after domain.Assignment
	err = s.provider.Update(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area || snapshot.Catalog.ApplicationID != area.ApplicationID() {
			return storage.WriteSet{}, domain.ErrRejected
		}
		before, ok := snapshot.Assignments[id]
		if !ok || before.ID != id {
			return storage.WriteSet{}, domain.ErrNotFound
		}
		latest := int64(0)
		for key, content := range snapshot.Contents {
			if key.ID != before.GrantID {
				continue
			}
			if key.Revision != content.Revision || content.GrantID != key.ID {
				return storage.WriteSet{}, domain.ErrRejected
			}
			if content.Revision > latest {
				latest = content.Revision
			}
		}
		if latest == 0 {
			return storage.WriteSet{}, domain.ErrRejected
		}
		// Already there. Not an error — an upgrade to what is already adopted is
		// a no-op, and reporting a conflict would make callers special-case it.
		if latest == before.GrantRevision {
			after = before
			return storage.WriteSet{}, nil
		}
		if latest < before.GrantRevision {
			return storage.WriteSet{}, domain.ErrConflict
		}
		proposed := before
		proposed.GrantRevision = latest
		now := s.clock.Now()
		if err := admin.CheckAssignmentAdoption(ctx, area, identity, proposed, now); err != nil {
			return storage.WriteSet{}, err
		}
		// Q-105: an upgrade "must also select the latest, and pass current
		// checks", and "if its permitted authority cannot support revision 3,
		// reject the upgrade and leave the assignment unchanged". The gate above
		// answers who may act; these answer whether the authority holds — and
		// they are the same checks CreateAssignment runs, because adopting a
		// revision is the same act as adopting it at creation.
		content, err := exactLatestContent(snapshot, proposed)
		if err != nil {
			return storage.WriteSet{}, err
		}
		if err := validation.CheckContent(area, snapshot.Catalog, content, snapshot.Roles); err != nil {
			return storage.WriteSet{}, err
		}
		if _, ok := snapshot.Teams[proposed.Recipient.ID]; !ok && proposed.Recipient.Type == "group" {
			return storage.WriteSet{}, domain.ErrRejected
		}
		parent, err := lineage.ResolveParentTeam(snapshot, content, proposed.Recipient.ID, now)
		if err != nil {
			return storage.WriteSet{}, err
		}
		if err := lineage.HasSource(snapshot, identity, parent, now); err != nil {
			return storage.WriteSet{}, err
		}
		// Q-103 is the rule that reaches downward, and it is the one to cite:
		// "validate the complete resulting authority **and affected bindings**;
		// unchanged child JSON is not proof of unchanged effective reach"
		// (grant-revisions.md:112-113). Q-102 asks for "boundary/dependency
		// validation" of an adoption and Q-105 says what to do when authority
		// cannot support a revision, but both speak about the assignment's own
		// upward boundary — which is all the checks above cover.
		//
		// Validating is not freezing. A revision that moves a child's inherited
		// scope is a legitimate act by someone with authority over the parent,
		// and B22 asks that the resulting authority be validated, not that it be
		// unchanged. What must not happen is adopting a revision a dependent
		// cannot be supported under — selecting a permission the new revision no
		// longer carries, say. Left to resolution that surfaces as a read
		// failure for the whole human, long after the write that caused it.
		//
		// Q-105 answers what to do about a revision that cannot be supported —
		// "reject the upgrade and leave the assignment unchanged" — of the
		// assignment's own authority. Applying the same answer downward is a
		// judgement, not a quotation: the Bxx rows are "analytical coverage of
		// the agreed rules, not executable conformance tests", and B12/B13
		// reject comparable parent changes outright. Refusing is the reading
		// that does not remove somebody else's access as a side effect.
		if err := dependentsStillResolve(ctx, snapshot, proposed, now); err != nil {
			return storage.WriteSet{}, err
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		after = proposed
		change := storage.AssignmentStatusChange{Before: before, After: proposed}
		return storage.WriteSet{AssignmentRevisionChange: &change}, nil
	})
	if err != nil {
		return fail(err)
	}
	return after, nil
}

// dependentsStillResolve checks every enabled binding beneath this one against
// the snapshot as it would be after the adoption.
//
// The staged snapshot is the point: each dependent is revalidated through the
// parent route it would actually have, not the one it has now, which is the only
// way "the complete resulting authority" can be read.
func dependentsStillResolve(ctx context.Context, snapshot storage.Snapshot, proposed domain.Assignment, now time.Time) error {
	dependents, err := lineage.DependentTeamAssignments(ctx, snapshot, proposed.ID)
	if err != nil {
		// The inventory covers the whole area, so it fails on any assignment
		// that does not validate — including one in a branch this adoption
		// cannot touch. Refusing on that would mean a single deactivated
		// permission anywhere in a tenant froze every adoption in it, which is
		// not a consequence of this write and not what Q-105 rejects an upgrade
		// for. An area in that state is already failing its reads; this
		// adoption is not what broke it and is not where it gets reported.
		return nil
	}
	if len(dependents) == 0 {
		return nil
	}
	staged := cloneSnapshot(snapshot)
	staged.Assignments[proposed.ID] = proposed
	// One index per snapshot, not per dependent. Each chain step otherwise
	// rescans every assignment in the area, and there is a chain per dependent.
	//
	// Each index must be built over the snapshot it is used with — they differ in
	// exactly one entry, the upgraded binding's revision. A review swapped them,
	// built both from one snapshot, and passed nil for both, and the module
	// stayed green every time, so the pairing is worth stating: a dependent only
	// reads that entry *through* the index if it sits two or more levels below
	// the binding being upgraded, and such a dependent cannot break alone.
	//
	// It cannot because narrowing is transitive. A grandchild selects from its
	// parent, which selects from the binding being upgraded, so an adoption that
	// takes away what the grandchild holds has already taken it from the child —
	// and the child's own first chain step is a scan on the right snapshot, so it
	// reports the breakage whatever the indexes say. The pairing is therefore
	// unobservable here by construction rather than by accident, and it is pinned
	// where it *is* observable: lineage's parity tests resolve a chain against an
	// index built over a different snapshot and require the answer to change.
	before, after := lineage.IndexBindings(snapshot), lineage.IndexBindings(staged)
	for _, dependent := range dependents {
		if err := ctx.Err(); err != nil {
			return err
		}
		// A disabled binding holds nothing, so nothing of its can stop working.
		// It is revalidated when it is enabled, which is where that check lives,
		// and the sibling guard on team moves carves out the same case for the
		// same reason — Q-101A/B: a disabled record supplies no authority
		// anywhere, so there is nothing left to validate.
		if dependent.Status != "enabled" {
			continue
		}
		if resolvesUnder(staged, after, dependent, now) {
			continue
		}
		// It does not resolve after. The question that decides whether this
		// adoption is to blame is whether it resolved before — and only then,
		// because the answer costs a second walk and almost every dependent
		// passes the first one.
		//
		// Without this the guard was absolute rather than differential: a
		// dependent whose own grant was disabled or expired refused an adoption
		// whose content was identical to what was already adopted. Revision
		// content is immutable, so an expired dependent would have frozen its
		// ancestor's binding for good.
		if !resolvesUnder(snapshot, before, dependent, now) {
			continue
		}
		return domain.ErrRejected
	}
	return nil
}

// resolvesUnder answers whether one binding has a complete route in the given
// snapshot: its content is there, its parent support resolves, and its own
// narrowing holds against that parent.
func resolvesUnder(snapshot storage.Snapshot, bindings *lineage.Bindings, dependent domain.Assignment, now time.Time) bool {
	content, ok := snapshot.Contents[domain.GrantKey{ID: dependent.GrantID, Revision: dependent.GrantRevision}]
	if !ok || content.GrantID != dependent.GrantID || content.Revision != dependent.GrantRevision {
		return false
	}
	parent, err := lineage.ResolveParentTeamIndexed(snapshot, bindings, content, dependent.Recipient.ID, now)
	if err != nil {
		return false
	}
	// The read rule, because this predicate answers a read question: does this
	// dependent resolve? Resolution narrows a grant to what the catalog still
	// supplies (Q-143), so testing it with the strict rule made the guard and
	// the evaluator disagree about exactly the grants Q-143 keeps alive — one
	// referencing a retired permission looked dead here and resolved there.
	//
	// The consequence was the harm Q-143 was raised to prevent. Because such a
	// dependent read as broken *before* the adoption as well as after, the
	// differential escape below waved the adoption through, and a person lost
	// access with nothing refusing the write.
	//
	// This was mislabelled as a write-path caller when the rules were split. It
	// is not: nothing here proposes a change.
	_, err = validation.NarrowSupplied(snapshot.Area, snapshot.Catalog, parent, content, snapshot.Roles)
	return err == nil
}
