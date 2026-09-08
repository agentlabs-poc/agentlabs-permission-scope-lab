package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"strings"
)

func (s *Service) SetAssignmentStatus(ctx context.Context, area domain.Area, identity domain.Identity, assignmentID, status string) (domain.Assignment, error) {
	fail := func(err error) (domain.Assignment, error) { return domain.Assignment{}, err }
	if ctx == nil || strings.TrimSpace(assignmentID) == "" || strings.Contains(assignmentID, "*") || (status != "enabled" && status != "disabled") {
		return fail(domain.ErrMalformed)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if err := area.Validate(); err != nil {
		return fail(err)
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return fail(err)
	}
	admin, ok := s.administration.(AssignmentStatusAdministration)
	if !ok || nilInterface(admin) {
		return fail(domain.ErrUnsupported)
	}
	var after domain.Assignment
	err := s.provider.Update(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area || snapshot.Catalog.ApplicationID != area.ApplicationID() {
			return storage.WriteSet{}, domain.ErrRejected
		}
		before, ok := snapshot.Assignments[assignmentID]
		if !ok {
			return storage.WriteSet{}, domain.ErrNotFound
		}
		if before.ID != assignmentID || !validStoredAssignment(before) {
			return storage.WriteSet{}, domain.ErrRejected
		}
		if before.Recipient.Type != "group" || snapshot.TrustedRoots[before.GrantID] {
			return storage.WriteSet{}, domain.ErrUnsupported
		}
		after = before
		after.Status = status
		now := s.clock.Now()
		if err := admin.CheckAssignmentStatus(ctx, cloneSnapshot(snapshot), identity, after, now); err != nil {
			return storage.WriteSet{}, err
		}
		dependents, err := lineage.DependentTeamAssignments(ctx, snapshot, assignmentID)
		if err != nil {
			return storage.WriteSet{}, err
		}
		if status == "disabled" {
			for _, dependent := range dependents {
				if dependent.Status == "enabled" {
					return storage.WriteSet{}, domain.ErrRejected
				}
			}
		} else {
			staged := cloneSnapshot(snapshot)
			staged.Assignments[assignmentID] = after
			content, ok := staged.Contents[domain.GrantKey{ID: after.GrantID, Revision: after.GrantRevision}]
			if !ok || content.GrantID != after.GrantID || content.Revision != after.GrantRevision {
				return storage.WriteSet{}, domain.ErrRejected
			}
			parent, err := lineage.ResolveParentTeam(staged, content, after.Recipient.ID, now)
			if err != nil {
				return storage.WriteSet{}, err
			}
			route, err := validation.Narrow(area, parent, content, staged.Roles)
			if err != nil {
				return storage.WriteSet{}, err
			}
			if err = eligibleRoute(route, s.clock.Now()); err != nil {
				return storage.WriteSet{}, err
			}
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		return storage.WriteSet{AssignmentStatusChange: &storage.AssignmentStatusChange{Before: before, After: after}}, nil
	})
	if err != nil {
		return fail(err)
	}
	return after, nil
}
