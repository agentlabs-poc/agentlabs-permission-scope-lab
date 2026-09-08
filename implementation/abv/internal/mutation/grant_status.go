package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

func (s *Service) SetGrantStatus(ctx context.Context, area domain.Area, identity domain.Identity, proposed domain.GrantControl) (domain.GrantControl, error) {
	fail := func(err error) (domain.GrantControl, error) { return domain.GrantControl{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if err := area.Validate(); err != nil {
		return fail(err)
	}
	if err := validateGrantControl(proposed); err != nil {
		return fail(err)
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return fail(err)
	}
	admin, ok := s.administration.(GrantStatusAdministration)
	if !ok || nilInterface(admin) {
		return fail(domain.ErrUnsupported)
	}
	err := s.provider.Update(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area || snapshot.Catalog.ApplicationID != area.ApplicationID() {
			return storage.WriteSet{}, domain.ErrRejected
		}
		before, ok := snapshot.Controls[proposed.ID]
		if !ok {
			return storage.WriteSet{}, domain.ErrNotFound
		}
		if before.ID != proposed.ID || validateGrantControl(before) != nil {
			return storage.WriteSet{}, domain.ErrRejected
		}
		if snapshot.TrustedRoots[proposed.ID] {
			return storage.WriteSet{}, domain.ErrUnsupported
		}
		now := s.clock.Now()
		if err := admin.CheckGrantStatus(ctx, cloneSnapshot(snapshot), identity, proposed, now); err != nil {
			return storage.WriteSet{}, err
		}
		routes, err := stageGrantStatus(snapshot, proposed, now)
		if err != nil {
			return storage.WriteSet{}, err
		}
		if err = ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		finalNow := s.clock.Now()
		for _, route := range routes {
			if err = eligibleRoute(route, finalNow); err != nil {
				return storage.WriteSet{}, err
			}
		}
		if err = ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		return storage.WriteSet{GrantStatusChange: &storage.GrantStatusChange{Before: before, After: proposed}}, nil
	})
	if err != nil {
		return fail(err)
	}
	return proposed, nil
}

func validateGrantControl(control domain.GrantControl) error {
	if strings.TrimSpace(control.Version) == "" || strings.TrimSpace(control.ID) == "" || strings.Contains(control.ID, "*") || !utf8.ValidString(control.Version) || !utf8.ValidString(control.ID) {
		return domain.ErrMalformed
	}
	if control.Version != "1" {
		return domain.ErrUnsupported
	}
	if control.Status != "enabled" && control.Status != "disabled" {
		return domain.ErrMalformed
	}
	return nil
}

func stageGrantStatus(snapshot storage.Snapshot, proposed domain.GrantControl, now time.Time) ([]domain.Route, error) {
	type binding struct{ grantID, recipientType, recipientID string }
	seen := make(map[binding]bool, len(snapshot.Assignments))
	ids := make([]string, 0)
	for id, assignment := range snapshot.Assignments {
		if id != assignment.ID || !validStoredAssignment(assignment) {
			return nil, domain.ErrRejected
		}
		key := binding{assignment.GrantID, assignment.Recipient.Type, assignment.Recipient.ID}
		if seen[key] {
			return nil, domain.ErrConflict
		}
		seen[key] = true
		if proposed.Status == "enabled" && assignment.GrantID == proposed.ID && assignment.Status == "enabled" {
			ids = append(ids, id)
		}
	}
	if proposed.Status == "disabled" {
		return nil, nil
	}
	sort.Strings(ids)
	staged := cloneSnapshot(snapshot)
	staged.Controls[proposed.ID] = proposed
	routes := make([]domain.Route, 0, len(ids))
	for _, id := range ids {
		assignment := staged.Assignments[id]
		if assignment.Recipient.Type != "group" {
			return nil, domain.ErrUnsupported
		}
		content, ok := staged.Contents[domain.GrantKey{ID: assignment.GrantID, Revision: assignment.GrantRevision}]
		if !ok || content.GrantID != assignment.GrantID || content.Revision != assignment.GrantRevision {
			return nil, domain.ErrRejected
		}
		parent, err := lineage.ResolveParentTeam(staged, content, assignment.Recipient.ID, now)
		if err != nil {
			return nil, err
		}
		route, err := validation.Narrow(snapshot.Area, parent, content, staged.Roles)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	return routes, nil
}

func validStoredAssignment(a domain.Assignment) bool {
	raw, err := json.Marshal(a)
	if err != nil {
		return false
	}
	decoded, err := codec.DecodeAssignment(raw)
	return err == nil && decoded == a
}
