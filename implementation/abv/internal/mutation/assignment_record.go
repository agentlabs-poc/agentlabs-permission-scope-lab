package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
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
		if err := admin.CheckAssignmentAdoption(ctx, area, identity, proposed, s.clock.Now()); err != nil {
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
