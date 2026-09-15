package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"sort"
	"time"
)

// OwnershipAdministration gates the ownership writes and reads.
//
// Which permission that gate should demand is deliberately not decided here.
// `team-administration.md`: "It does not decide whether auth:group::write
// authorizes ownership transfer; that operation's exact permission remains
// open." So this is a separate method from the team gates — an adapter may
// answer it the same way or differently, and the handbook can settle it later
// without changing this signature.
type OwnershipAdministration interface {
	CheckOwnershipWrite(context.Context, domain.Area, domain.Identity, domain.Ownership, time.Time) error
	CheckOwnershipRead(context.Context, domain.Area, domain.Identity, time.Time) error
}

func (s *Service) ownershipAdministration(ctx context.Context, area domain.Area, identity domain.Identity) (OwnershipAdministration, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := area.Validate(); err != nil {
		return nil, err
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return nil, err
	}
	admin, ok := s.administration.(OwnershipAdministration)
	if !ok || nilInterface(admin) {
		return nil, domain.ErrUnsupported
	}
	return admin, nil
}

// AddOwner is add-only: adding an ownership that exists is a conflict rather
// than a quiet no-op, so a caller learns it was already so.
func (s *Service) AddOwner(ctx context.Context, area domain.Area, identity domain.Identity, teamID, humanID string) error {
	admin, err := s.ownershipAdministration(ctx, area, identity)
	if err != nil {
		return err
	}
	proposed := domain.Ownership{TeamID: teamID, HumanID: humanID}
	return s.provider.Update(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area {
			return storage.WriteSet{}, domain.ErrRejected
		}
		if err := admin.CheckOwnershipWrite(ctx, area, identity, proposed, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		if err := validation.CheckOwnership(snapshot.Teams, proposed); err != nil {
			return storage.WriteSet{}, err
		}
		for _, existing := range snapshot.Ownerships {
			if existing == proposed {
				return storage.WriteSet{}, domain.ErrConflict
			}
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		added := proposed
		return storage.WriteSet{AddedOwnership: &added}, nil
	})
}

// RemoveOwner says whether it did anything: removing an ownership that is not
// there is ErrNotFound, not success.
//
// It permits removing the last owner. Team administration is tenant-wide
// (scope: {}), so a team with no owners is still administrable by whoever holds
// that — there is always a way back, and refusing would mean a team could never
// be handed over in one step.
func (s *Service) RemoveOwner(ctx context.Context, area domain.Area, identity domain.Identity, teamID, humanID string) error {
	admin, err := s.ownershipAdministration(ctx, area, identity)
	if err != nil {
		return err
	}
	proposed := domain.Ownership{TeamID: teamID, HumanID: humanID}
	return s.provider.Update(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area {
			return storage.WriteSet{}, domain.ErrRejected
		}
		if err := admin.CheckOwnershipWrite(ctx, area, identity, proposed, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		if err := validation.CheckOwnership(snapshot.Teams, proposed); err != nil {
			return storage.WriteSet{}, err
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		removed := proposed
		return storage.WriteSet{RemovedOwnership: &removed}, nil
	})
}

// ListOwners answers in both directions — a team's owners, or a human's teams —
// and requires exactly one, the rule ListMembers already holds.
func (s *Service) ListOwners(ctx context.Context, area domain.Area, identity domain.Identity, filter domain.OwnerFilter) (domain.OwnerPage, error) {
	fail := func(err error) (domain.OwnerPage, error) { return domain.OwnerPage{}, err }
	admin, err := s.ownershipAdministration(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if (filter.TeamID == "") == (filter.HumanID == "") {
		return fail(domain.ErrMalformed)
	}
	if filter.Offset < 0 || filter.Limit < 0 || filter.Limit > maxGrantPage {
		return fail(domain.ErrMalformed)
	}
	var page domain.OwnerPage
	err = s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area {
			return domain.ErrRejected
		}
		if err := admin.CheckOwnershipRead(ctx, area, identity, s.clock.Now()); err != nil {
			return err
		}
		matched := make([]domain.Ownership, 0, len(snapshot.Ownerships))
		for _, o := range snapshot.Ownerships {
			if err := ctx.Err(); err != nil {
				return err
			}
			if filter.TeamID != "" && o.TeamID != filter.TeamID {
				continue
			}
			if filter.HumanID != "" && o.HumanID != filter.HumanID {
				continue
			}
			matched = append(matched, o)
		}
		sort.Slice(matched, func(i, j int) bool {
			if matched[i].TeamID != matched[j].TeamID {
				return matched[i].TeamID < matched[j].TeamID
			}
			return matched[i].HumanID < matched[j].HumanID
		})
		page = domain.OwnerPage{Owners: grantPage(matched, filter.Offset, filter.Limit), Total: len(matched)}
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return page, nil
}
