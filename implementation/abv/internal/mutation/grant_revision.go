package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"bytes"
	"context"
	"encoding/json"
	"time"
)

type GrantRevisionAdministration interface {
	CheckGrantRevisionPublication(context.Context, storage.Snapshot, domain.Identity, domain.GrantContent, time.Time) error
}

func (s *Service) PublishGrantRevision(ctx context.Context, area domain.Area, identity domain.Identity, sourceAssignmentID string, proposed domain.GrantContent) (domain.GrantContent, error) {
	fail := func(err error) (domain.GrantContent, error) { return domain.GrantContent{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if err := area.Validate(); err != nil {
		return fail(err)
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return fail(err)
	}
	if invalidIdentityPart(sourceAssignmentID) {
		return fail(domain.ErrMalformed)
	}
	if err := validateGrantRevisionProposal(proposed); err != nil {
		return fail(err)
	}
	admin, ok := s.administration.(GrantRevisionAdministration)
	if !ok || nilInterface(admin) {
		return fail(domain.ErrUnsupported)
	}
	proposed = cloneContent(proposed)
	err := s.provider.Update(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area || snapshot.Catalog.ApplicationID != area.ApplicationID() {
			return storage.WriteSet{}, domain.ErrRejected
		}
		if err := existingRevisableGrant(snapshot, proposed); err != nil {
			return storage.WriteSet{}, err
		}
		now := s.clock.Now()
		if err := admin.CheckGrantRevisionPublication(ctx, cloneSnapshot(snapshot), identity, cloneContent(proposed), now); err != nil {
			return storage.WriteSet{}, err
		}
		if err := validation.CheckContent(area, snapshot.Catalog, proposed, snapshot.Roles); err != nil {
			return storage.WriteSet{}, err
		}
		for _, value := range proposed.Scope {
			if value == "$self" {
				return storage.WriteSet{}, domain.ErrUnsupported
			}
		}
		parent, err := lineage.ResolveTeamAssignment(snapshot, sourceAssignmentID, now)
		if err != nil {
			return storage.WriteSet{}, err
		}
		if parent.GrantID != proposed.ParentGrantID || routeContainsGrant(snapshot, parent, proposed.GrantID) {
			return storage.WriteSet{}, domain.ErrRejected
		}
		if err = lineage.HasSource(snapshot, identity, parent, now); err != nil {
			return storage.WriteSet{}, err
		}
		if _, err = validation.Narrow(area, parent, proposed, snapshot.Roles); err != nil {
			return storage.WriteSet{}, err
		}
		if err = ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		if err = eligibleRoute(parent, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		write := cloneContent(proposed)
		return storage.WriteSet{NewGrantRevision: &write}, nil
	})
	if err != nil {
		return fail(err)
	}
	return proposed, nil
}

func validateGrantRevisionProposal(proposed domain.GrantContent) error {
	raw, err := json.Marshal(proposed)
	if err != nil {
		return domain.ErrMalformed
	}
	decoded, err := codec.DecodeContent(raw)
	if err != nil {
		return err
	}
	canonical, err := json.Marshal(decoded)
	if err != nil || !bytes.Equal(canonical, raw) {
		return domain.ErrMalformed
	}
	return nil
}

func existingRevisableGrant(snapshot storage.Snapshot, proposed domain.GrantContent) error {
	control, ok := snapshot.Controls[proposed.GrantID]
	if !ok || proposed.ParentGrantID == "" || control.Version != "1" || control.ID != proposed.GrantID || (control.Status != "enabled" && control.Status != "disabled") || snapshot.TrustedRoots[proposed.GrantID] {
		return domain.ErrRejected
	}
	var latest domain.GrantContent
	for key, content := range snapshot.Contents {
		if key.ID != content.GrantID || key.Revision != content.Revision {
			return domain.ErrRejected
		}
		if content.GrantID == proposed.GrantID && content.Revision > latest.Revision {
			latest = content
		}
	}
	if latest.GrantID == "" {
		return domain.ErrRejected
	}
	if proposed.Revision <= latest.Revision {
		return domain.ErrConflict
	}
	if proposed.ParentGrantID != latest.ParentGrantID {
		return domain.ErrRejected
	}
	return nil
}

func routeContainsGrant(snapshot storage.Snapshot, route domain.Route, grantID string) bool {
	if route.GrantID == grantID {
		return true
	}
	for _, assignmentID := range route.AssignmentIDs {
		if snapshot.Assignments[assignmentID].GrantID == grantID {
			return true
		}
	}
	return false
}
