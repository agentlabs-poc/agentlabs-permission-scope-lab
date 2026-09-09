package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"slices"
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
