package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"slices"
	"sort"
)

func (s *Service) PublishRole(ctx context.Context, area domain.Area, identity domain.Identity, proposed domain.RoleContent) (domain.RoleContent, error) {
	fail := func(err error) (domain.RoleContent, error) { return domain.RoleContent{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if err := area.Validate(); err != nil {
		return fail(err)
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return fail(err)
	}
	admin, ok := s.administration.(RoleAdministration)
	if !ok || nilInterface(admin) {
		return fail(domain.ErrUnsupported)
	}
	proposed.Permissions = slices.Clone(proposed.Permissions)
	err := s.provider.Update(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area || snapshot.Catalog.ApplicationID != area.ApplicationID() {
			return storage.WriteSet{}, domain.ErrRejected
		}
		key := domain.RoleKey{ID: proposed.ID, Revision: proposed.Revision}
		for storedKey, role := range snapshot.Roles {
			if storedKey.ID != role.ID || storedKey.Revision != role.Revision {
				return storage.WriteSet{}, domain.ErrRejected
			}
			if storedKey == key {
				return storage.WriteSet{}, domain.ErrConflict
			}
		}
		evidence := proposed
		evidence.Permissions = slices.Clone(proposed.Permissions)
		if err := admin.CheckRolePublication(ctx, cloneSnapshot(snapshot), identity, evidence, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		if err := validation.CheckRolePublication(area, snapshot.Catalog, proposed); err != nil {
			return storage.WriteSet{}, err
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		write := proposed
		write.Permissions = slices.Clone(proposed.Permissions)
		return storage.WriteSet{NewRoleRevision: &write}, nil
	})
	if err != nil {
		return fail(err)
	}
	return proposed, nil
}

// defaultRolePage and maxRolePage bound a role listing, as the catalog listings
// are bounded. There is no unbounded page and no query interface.
const (
	defaultRolePage = 100
	maxRolePage     = 500
)

// GetRole returns one published revision by its exact id and revision.
//
// A revision is always required. A caller wanting the newest bundle uses
// ListRoles with Revisions: LatestRevision, which is administrative and
// explicit — keeping the typed single read exact is what stops an implicit
// latest-role fallback, which Q-118 forbids, from appearing on the path that
// adoption uses.
//
// Permissions come back as published, not re-filtered against the catalog: a
// permission retired after publication still appears here. The record is what it
// is; whether it still resolves is CheckContent's question at evaluation.
func (s *Service) GetRole(ctx context.Context, area domain.Area, identity domain.Identity, id string, revision int64) (domain.RoleContent, error) {
	fail := func(err error) (domain.RoleContent, error) { return domain.RoleContent{}, err }
	admin, err := s.roleCatalog(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if !codec.ValidRoleID(id) {
		return fail(domain.ErrMalformed)
	}
	if _, err = codec.RenderRevision(revision); err != nil {
		return fail(err)
	}
	var result domain.RoleContent
	err = s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area {
			return domain.ErrRejected
		}
		if err := admin.CheckRoleRead(ctx, area, identity, s.clock.Now()); err != nil {
			return err
		}
		content, ok := snapshot.Roles[domain.RoleKey{ID: id, Revision: revision}]
		if !ok || content.ID != id || content.Revision != revision {
			return domain.ErrNotFound
		}
		result = content
		result.Permissions = slices.Clone(content.Permissions)
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return result, nil
}

// ListRoles returns one bounded page of a tenant's role catalog, ordered by id
// then revision.
//
// A role is a family of revisions, so AllRevisions is the default: the reverse
// would hide history behind a flag nobody sets. LatestRevision returns the
// highest revision of each selected role, computed here and never stored, so
// nothing can go stale.
//
// ID and Name are exact matches, never prefixes — both are flat tokens, so a
// prefix would be a string match inside one key slot rather than a structural
// one. Name is not unique, so filtering by it may select more than one role.
func (s *Service) ListRoles(ctx context.Context, area domain.Area, identity domain.Identity, filter domain.RoleFilter) (domain.RolePage, error) {
	fail := func(err error) (domain.RolePage, error) { return domain.RolePage{}, err }
	admin, err := s.roleCatalog(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if filter.Offset < 0 || filter.Limit < 0 || filter.Limit > maxRolePage {
		return fail(domain.ErrMalformed)
	}
	if filter.Revisions != domain.AllRevisions && filter.Revisions != domain.LatestRevision {
		return fail(domain.ErrMalformed)
	}
	if filter.ID != "" && !codec.ValidRoleID(filter.ID) {
		return fail(domain.ErrMalformed)
	}
	limit := filter.Limit
	if limit == 0 {
		limit = defaultRolePage
	}

	var page domain.RolePage
	err = s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area {
			return domain.ErrRejected
		}
		if err := admin.CheckRoleRead(ctx, area, identity, s.clock.Now()); err != nil {
			return err
		}
		matched := make([]domain.RoleContent, 0, len(snapshot.Roles))
		for key, content := range snapshot.Roles {
			if err := ctx.Err(); err != nil {
				return err
			}
			if content.ID != key.ID || content.Revision != key.Revision {
				return domain.ErrRejected
			}
			if filter.ID != "" && content.ID != filter.ID {
				continue
			}
			if filter.Name != "" && content.Name != filter.Name {
				continue
			}
			clone := content
			clone.Permissions = slices.Clone(content.Permissions)
			matched = append(matched, clone)
		}
		// Ordering by id then revision is what makes an offset mean the same
		// thing on every call, and it is the order an administrator reads a
		// revision history in.
		sort.Slice(matched, func(i, j int) bool {
			if matched[i].ID != matched[j].ID {
				return matched[i].ID < matched[j].ID
			}
			return matched[i].Revision < matched[j].Revision
		})
		if filter.Revisions == domain.LatestRevision {
			matched = latestPerRole(matched)
		}

		// Total follows the selector: rows under AllRevisions, roles under
		// LatestRevision, so the page count is right for what was asked.
		page.Total = len(matched)
		page.Generation = snapshot.Catalog.Generation
		if filter.Offset >= len(matched) {
			page.Roles = []domain.RoleContent{}
			return nil
		}
		end := filter.Offset + limit
		if end > len(matched) {
			end = len(matched)
		}
		page.Roles = matched[filter.Offset:end]
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return page, nil
}

// latestPerRole keeps the highest revision of each id. Its input is already
// ordered by id then revision ascending, so the last entry of each run wins.
func latestPerRole(ordered []domain.RoleContent) []domain.RoleContent {
	result := make([]domain.RoleContent, 0, len(ordered))
	for i, content := range ordered {
		if i+1 == len(ordered) || ordered[i+1].ID != content.ID {
			result = append(result, content)
		}
	}
	return result
}

// roleCatalog resolves the administration seam the two role reads share. An
// adapter that does not support them makes the operation unsupported rather than
// unprotected.
func (s *Service) roleCatalog(ctx context.Context, area domain.Area, identity domain.Identity) (RoleReadAdministration, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := area.Validate(); err != nil {
		return nil, err
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return nil, err
	}
	admin, ok := s.administration.(RoleReadAdministration)
	if !ok || nilInterface(admin) {
		return nil, domain.ErrUnsupported
	}
	return admin, nil
}
