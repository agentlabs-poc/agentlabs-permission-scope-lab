package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"context"
	"slices"
	"sort"
)

// defaultTeamPage and maxTeamPage bound a team or membership listing, as the
// catalog and role listings are bounded.
const (
	defaultTeamPage = 100
	maxTeamPage     = 500
)

// GetTeam returns one team by its id, with its name and its parent.
//
// A root team comes back with an empty ParentID, which is the real value rather
// than an omission: the top of the hierarchy is a team whose parent is empty.
func (s *Service) GetTeam(ctx context.Context, area domain.Area, identity domain.Identity, id string) (domain.Team, error) {
	fail := func(err error) (domain.Team, error) { return domain.Team{}, err }
	admin, err := s.teamCatalog(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if !codec.ValidRoleID(id) {
		return fail(domain.ErrMalformed)
	}
	var result domain.Team
	err = s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area {
			return domain.ErrRejected
		}
		if err := admin.CheckTeamRead(ctx, area, identity, s.clock.Now()); err != nil {
			return err
		}
		team, ok := snapshot.Teams[id]
		if !ok || team.ID != id {
			return domain.ErrNotFound
		}
		result = team
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return result, nil
}

// ListTeams returns one bounded page of the tenant's teams, ordered by id.
//
// ParentID is a pointer because "" is a real value — the parent a root team
// holds — so it cannot double as "unset". Nil lists every team; a pointer to ""
// lists roots only. Name is an exact match and is not unique, so it may select
// several teams.
func (s *Service) ListTeams(ctx context.Context, area domain.Area, identity domain.Identity, filter domain.TeamFilter) (domain.TeamPage, error) {
	fail := func(err error) (domain.TeamPage, error) { return domain.TeamPage{}, err }
	admin, err := s.teamCatalog(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if filter.Offset < 0 || filter.Limit < 0 || filter.Limit > maxTeamPage {
		return fail(domain.ErrMalformed)
	}
	if filter.ParentID != nil && *filter.ParentID != "" && !codec.ValidRoleID(*filter.ParentID) {
		return fail(domain.ErrMalformed)
	}
	limit := filter.Limit
	if limit == 0 {
		limit = defaultTeamPage
	}

	var page domain.TeamPage
	err = s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area {
			return domain.ErrRejected
		}
		if err := admin.CheckTeamRead(ctx, area, identity, s.clock.Now()); err != nil {
			return err
		}
		matched := make([]domain.Team, 0, len(snapshot.Teams))
		for id, team := range snapshot.Teams {
			if err := ctx.Err(); err != nil {
				return err
			}
			if team.ID != id {
				return domain.ErrRejected
			}
			if filter.ParentID != nil && team.ParentID != *filter.ParentID {
				continue
			}
			if filter.Name != "" && team.Name != filter.Name {
				continue
			}
			matched = append(matched, team)
		}
		sort.Slice(matched, func(i, j int) bool { return matched[i].ID < matched[j].ID })
		page.Total = len(matched)
		page.Generation = snapshot.Catalog.Generation
		if filter.Offset >= len(matched) {
			page.Teams = []domain.Team{}
			return nil
		}
		end := filter.Offset + limit
		if end > len(matched) {
			end = len(matched)
		}
		page.Teams = matched[filter.Offset:end]
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return page, nil
}

// ListMembers answers in both directions: a team's roster, or one human's teams.
//
// Exactly one of TeamID and HumanID is required. An unfiltered listing of every
// membership is not a question anyone asks, and it would be unbounded in the
// dimension that grows fastest — so it is refused rather than served slowly.
func (s *Service) ListMembers(ctx context.Context, area domain.Area, identity domain.Identity, filter domain.MemberFilter) (domain.MemberPage, error) {
	fail := func(err error) (domain.MemberPage, error) { return domain.MemberPage{}, err }
	admin, err := s.teamCatalog(ctx, area, identity)
	if err != nil {
		return fail(err)
	}
	if filter.Offset < 0 || filter.Limit < 0 || filter.Limit > maxTeamPage {
		return fail(domain.ErrMalformed)
	}
	if (filter.TeamID == "") == (filter.HumanID == "") {
		return fail(domain.ErrMalformed)
	}
	if filter.TeamID != "" && !codec.ValidRoleID(filter.TeamID) {
		return fail(domain.ErrMalformed)
	}
	if filter.HumanID != "" && !codec.ValidHumanID(filter.HumanID) {
		return fail(domain.ErrMalformed)
	}
	limit := filter.Limit
	if limit == 0 {
		limit = defaultTeamPage
	}

	var page domain.MemberPage
	err = s.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area {
			return domain.ErrRejected
		}
		if err := admin.CheckTeamRead(ctx, area, identity, s.clock.Now()); err != nil {
			return err
		}
		matched := make([]domain.Membership, 0, len(snapshot.Memberships))
		for _, m := range snapshot.Memberships {
			if err := ctx.Err(); err != nil {
				return err
			}
			if filter.TeamID != "" && m.TeamID != filter.TeamID {
				continue
			}
			if filter.HumanID != "" && m.HumanID != filter.HumanID {
				continue
			}
			matched = append(matched, m)
		}
		sort.Slice(matched, func(i, j int) bool {
			if matched[i].TeamID != matched[j].TeamID {
				return matched[i].TeamID < matched[j].TeamID
			}
			return matched[i].HumanID < matched[j].HumanID
		})
		page.Total = len(matched)
		page.Generation = snapshot.Catalog.Generation
		if filter.Offset >= len(matched) {
			page.Members = []domain.Membership{}
			return nil
		}
		end := filter.Offset + limit
		if end > len(matched) {
			end = len(matched)
		}
		page.Members = slices.Clone(matched[filter.Offset:end])
		return nil
	})
	if err != nil {
		return fail(err)
	}
	return page, nil
}

// teamCatalog resolves the administration seam the three team reads share.
func (s *Service) teamCatalog(ctx context.Context, area domain.Area, identity domain.Identity) (TeamReadAdministration, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := area.Validate(); err != nil {
		return nil, err
	}
	if err := validateSupportedIdentity(identity); err != nil {
		return nil, err
	}
	admin, ok := s.administration.(TeamReadAdministration)
	if !ok || nilInterface(admin) {
		return nil, domain.ErrUnsupported
	}
	return admin, nil
}
