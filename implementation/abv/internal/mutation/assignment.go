package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"
)

func (s *Service) CreateAssignment(ctx context.Context, area domain.Area, identity domain.Identity, proposed domain.Assignment) (domain.Receipt, error) {
	fail := func(err error) (domain.Receipt, error) { return domain.Receipt{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if err := area.Validate(); err != nil {
		return fail(err)
	}
	if err := validateProposal(proposed); err != nil {
		return fail(err)
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return fail(err)
	}
	err := s.provider.Update(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area || snapshot.Catalog.ApplicationID != area.ApplicationID() {
			return storage.WriteSet{}, domain.ErrRejected
		}
		now := s.clock.Now()
		if err := s.administration.CheckAssignment(ctx, cloneSnapshot(snapshot), identity, proposed, now); err != nil {
			return storage.WriteSet{}, err
		}
		content, err := exactLatestContent(snapshot, proposed)
		if err != nil {
			return storage.WriteSet{}, err
		}
		if err = validation.CheckContent(area, snapshot.Catalog, content, snapshot.Roles); err != nil {
			return storage.WriteSet{}, err
		}
		team, ok := snapshot.Teams[proposed.Recipient.ID]
		if !ok || team.ID != proposed.Recipient.ID || team.ParentID == "" {
			return storage.WriteSet{}, domain.ErrRejected
		}
		parent, err := lineage.ResolveParentTeam(snapshot, content, proposed.Recipient.ID, now)
		if err != nil {
			return storage.WriteSet{}, err
		}
		if err = lineage.HasSource(snapshot, identity, parent, now); err != nil {
			return storage.WriteSet{}, err
		}
		child, err := validation.Narrow(area, parent, content, snapshot.Roles)
		if err != nil {
			return storage.WriteSet{}, err
		}
		if err = rejectDuplicate(snapshot, proposed); err != nil {
			return storage.WriteSet{}, err
		}
		if err = ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		if err = eligibleRoute(child, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		return storage.WriteSet{NewAssignments: []domain.Assignment{proposed}}, nil
	})
	if err != nil {
		return fail(err)
	}
	return domain.Receipt{AssignmentID: proposed.ID}, nil
}

// CheckAssignment diagnoses proposal shape and lineage from a fresh read. It
// deliberately does not establish administrative or source authority.
func (s *Service) CheckAssignment(ctx context.Context, area domain.Area, raw []byte) (domain.Diagnostic, error) {
	if err := ctx.Err(); err != nil {
		return domain.Diagnostic{}, err
	}
	if err := area.Validate(); err != nil {
		return domain.Diagnostic{}, err
	}
	proposed, err := codec.DecodeAssignment(raw)
	if err != nil {
		return domain.Diagnostic{}, err
	}
	if err = validateProposal(proposed); err != nil {
		return domain.Diagnostic{}, err
	}
	var route domain.Route
	err = s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area || snapshot.Catalog.ApplicationID != area.ApplicationID() {
			return domain.ErrRejected
		}
		content, err := exactLatestContent(snapshot, proposed)
		if err != nil {
			return err
		}
		if err = validation.CheckContent(area, snapshot.Catalog, content, snapshot.Roles); err != nil {
			return err
		}
		team, ok := snapshot.Teams[proposed.Recipient.ID]
		if !ok || team.ID != proposed.Recipient.ID || team.ParentID == "" {
			return domain.ErrRejected
		}
		parent, err := lineage.ResolveParentTeam(snapshot, content, proposed.Recipient.ID, s.clock.Now())
		if err != nil {
			return err
		}
		route, err = validation.Narrow(area, parent, content, snapshot.Roles)
		return err
	})
	if err != nil {
		return domain.Diagnostic{}, err
	}
	return domain.Diagnostic{Summary: "proposal is structurally and lineally valid; administrative and source authority are not established", Route: &route}, nil
}

func validateProposal(proposed domain.Assignment) error {
	raw, err := json.Marshal(proposed)
	if err != nil {
		return domain.ErrMalformed
	}
	decoded, err := codec.DecodeAssignment(raw)
	if err != nil {
		return err
	}
	if decoded != proposed {
		return domain.ErrMalformed
	}
	if proposed.Recipient.Type != "group" {
		return domain.ErrUnsupported
	}
	if proposed.Status != "enabled" {
		return domain.ErrUnsupported
	}
	return nil
}

func validateSupportedIdentity(identity domain.Identity) error {
	if invalidIdentityPart(identity.Version) || invalidIdentityPart(identity.Actor.Type) || invalidIdentityPart(identity.Actor.ID) || invalidIdentityPart(identity.HumanID) {
		return domain.ErrMalformed
	}
	if identity.Version != "1" || identity.Actor.Type != "user" || identity.Actor.ID != identity.HumanID {
		return domain.ErrUnsupported
	}
	return nil
}

func invalidIdentityPart(value string) bool {
	return strings.TrimSpace(value) == "" || value == "*" || !utf8.ValidString(value)
}

func exactLatestContent(snapshot storage.Snapshot, proposed domain.Assignment) (domain.GrantContent, error) {
	key := domain.GrantKey{ID: proposed.GrantID, Revision: proposed.GrantRevision}
	content, ok := snapshot.Contents[key]
	if !ok || content.GrantID != key.ID || content.Revision != key.Revision {
		return domain.GrantContent{}, domain.ErrRejected
	}
	for storedKey, stored := range snapshot.Contents {
		if storedKey.ID != stored.GrantID || storedKey.Revision != stored.Revision {
			return domain.GrantContent{}, domain.ErrRejected
		}
		if stored.GrantID == proposed.GrantID && stored.Revision > proposed.GrantRevision {
			return domain.GrantContent{}, domain.ErrRejected
		}
	}
	return content, nil
}

func rejectDuplicate(snapshot storage.Snapshot, proposed domain.Assignment) error {
	for id, assignment := range snapshot.Assignments {
		if id != assignment.ID {
			return domain.ErrRejected
		}
		if assignment.ID == proposed.ID || (assignment.GrantID == proposed.GrantID && assignment.Recipient == proposed.Recipient) {
			return domain.ErrConflict
		}
	}
	return nil
}

func eligibleRoute(route domain.Route, now time.Time) error {
	for _, validity := range route.Validities {
		if (validity.NotBefore != nil && now.Before(*validity.NotBefore)) || (validity.ExpiresAt != nil && !now.Before(*validity.ExpiresAt)) {
			return domain.ErrRejected
		}
	}
	return nil
}

func cloneSnapshot(source storage.Snapshot) storage.Snapshot {
	result := source
	result.Catalog = cloneCatalog(source.Catalog)
	result.Controls = cloneMap(source.Controls)
	result.Contents = make(map[domain.GrantKey]domain.GrantContent, len(source.Contents))
	for key, value := range source.Contents {
		result.Contents[key] = cloneContent(value)
	}
	result.Assignments = cloneMap(source.Assignments)
	result.Roles = make(map[domain.RoleKey]domain.RoleContent, len(source.Roles))
	for key, value := range source.Roles {
		value.Permissions = append([]string(nil), value.Permissions...)
		result.Roles[key] = value
	}
	result.Teams = cloneMap(source.Teams)
	result.Memberships = append([]domain.Membership(nil), source.Memberships...)
	result.TrustedRoots = cloneMap(source.TrustedRoots)
	return result
}

func cloneMap[K comparable, V any](source map[K]V) map[K]V {
	result := make(map[K]V, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func cloneContent(content domain.GrantContent) domain.GrantContent {
	content.Permissions = append([]string(nil), content.Permissions...)
	content.Scope = cloneMap(content.Scope)
	if content.Validity != nil {
		validity := *content.Validity
		if validity.NotBefore != nil {
			value := *validity.NotBefore
			validity.NotBefore = &value
		}
		if validity.ExpiresAt != nil {
			value := *validity.ExpiresAt
			validity.ExpiresAt = &value
		}
		content.Validity = &validity
	}
	return content
}
