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

// CreateTeam creates a team, optionally inside another. The handbook's *team
// create* covers "creating teams and subteams", which is why one operation does
// both rather than two.
//
// The id is issued, never supplied — the same rule roles hold. The caller names
// the team; Auth-AL names the record.
func (s *Service) CreateTeam(ctx context.Context, area domain.Area, identity domain.Identity, name, parentID string) (domain.Team, error) {
	fail := func(err error) (domain.Team, error) { return domain.Team{}, err }
	if err := s.teamWriteAuthority(ctx, area, identity); err != nil {
		return fail(err)
	}
	proposed := domain.Team{ID: s.ids.Next(), Name: name, ParentID: parentID}
	err := s.provider.UpdateAdministered(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area {
			return storage.WriteSet{}, domain.ErrRejected
		}
		// The bounding team is the parent, because creating a subteam is the act
		// `auth:group::create` covers "creating teams and subteams" with. A
		// top-level team has no parent and so no bounding team, which only an
		// unscoped route satisfies — the tenant administrator's.
		if err := s.authorizeUnder(ctx, snapshot, identity, groupCreate, parentID, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		if err := validation.CheckTeamCreation(snapshot.Teams, proposed); err != nil {
			return storage.WriteSet{}, err
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		created := proposed
		return storage.WriteSet{NewTeam: &created}, nil
	})
	if err != nil {
		return fail(err)
	}
	return proposed, nil
}

// SetTeamParent moves a team, and refuses a move that would put it inside its
// own subtree. The handbook's *team write* covers this: it "includes human
// membership management", which makes it broader than membership, and a
// re-parent is a change to the team.
func (s *Service) SetTeamParent(ctx context.Context, area domain.Area, identity domain.Identity, id, parentID string) (domain.Team, error) {
	fail := func(err error) (domain.Team, error) { return domain.Team{}, err }
	if err := s.teamWriteAuthority(ctx, area, identity); err != nil {
		return fail(err)
	}
	var result domain.Team
	err := s.provider.UpdateAdministered(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area {
			return storage.WriteSet{}, domain.ErrRejected
		}
		// Both ends, and deliberately two calls rather than one. A move changes
		// where a team hangs, so it is a write to the team *and* a change to what
		// the new parent contains; requiring authority over each is the
		// conservative reading, and conservative is the direction that can be
		// loosened later.
		//
		// They cannot be one call: a single route can never carry team=id and
		// team=parentID at once, so asking for both in one material map would ask
		// for a route that cannot exist. In practice this means only an unscoped
		// route — the tenant administrator's — moves a team between two parents,
		// which is the honest consequence rather than a rule invented here.
		if err := s.authorizeTeam(ctx, snapshot, identity, groupWrite, id, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		if parentID != snapshot.Teams[id].ParentID {
			// The other end. A move *up* to top level has no new parent to name,
			// and the end state is the one CreateTeam reserves for an unscoped
			// route — so it takes the same unscoped route. Checking only the moved
			// team here let a holder scoped to one team produce, by moving, the
			// state it could not produce by creating.
			if err := s.authorizeUnder(ctx, snapshot, identity, groupWrite, parentID, s.clock.Now()); err != nil {
				return storage.WriteSet{}, err
			}
		}
		if err := validation.CheckTeamReparent(snapshot.Teams, id, parentID); err != nil {
			return storage.WriteSet{}, err
		}
		// B13: "Change/remove team parent under an affected enabled binding |
		// Reject; equivalent authority elsewhere is not an exemption."
		//
		// CheckTeamReparent above is a pure function over the teams map — ids,
		// existence, cycles — so it cannot see what the move re-anchors. A team's
		// position is how its bindings reach their parent support, and moving it
		// changes the inherited scope of every binding at or beneath it without
		// touching one grant, one assignment or one record anyone would think to
		// review. DeleteTeam already refuses while an assignment names the team;
		// the asymmetry between removing a team and moving it was the whole bug.
		// Re-asserting the parent a team already has changes nothing and
		// re-anchors nothing, so B13 — which governs *changing* or removing a
		// parent — does not reach it. Refusing here broke idempotent retries: a
		// caller whose request timed out after succeeding got a conflict on the
		// repeat. UpgradeAssignment carries the same shortcut for its own no-op.
		if snapshot.Teams[id].ParentID != parentID {
			if err := refuseAffectedBindings(snapshot, administrativeChain(snapshot), id); err != nil {
				return storage.WriteSet{}, err
			}
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		result = snapshot.Teams[id]
		result.ParentID = parentID
		changed := result
		return storage.WriteSet{TeamParent: &changed}, nil
	})
	if err != nil {
		return fail(err)
	}
	return result, nil
}

// DeleteTeam removes a team, and refuses while anything depends on it: a child
// team, a membership, or an assignment naming it.
//
// Nothing cascades. A delete that quietly removed memberships would make one
// administrative act perform another, and the handbook keeps team
// administration, membership administration and assignment authority distinct.
// The caller empties the team first, deliberately.
func (s *Service) DeleteTeam(ctx context.Context, area domain.Area, identity domain.Identity, id string) error {
	if err := s.teamWriteAuthority(ctx, area, identity); err != nil {
		return err
	}
	return s.provider.UpdateAdministered(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area {
			return storage.WriteSet{}, domain.ErrRejected
		}
		if err := s.authorizeTeam(ctx, snapshot, identity, groupDelete, id, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		if err := validation.CheckTeamDeletion(snapshot.Teams, snapshot.Memberships, snapshot.Ownerships, snapshot.Assignments, id); err != nil {
			return storage.WriteSet{}, err
		}
		// And the administrative chain, which is in another area and so not in the
		// assignments above. An administrative assignment names a team exactly as a
		// business one does.
		if err := refuseAdministrativeDependants(administrativeChain(snapshot), id); err != nil {
			return storage.WriteSet{}, err
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		return storage.WriteSet{RemovedTeam: id}, nil
	})
}

// AddMember puts one human in one team. Identity is the pair, so adding someone
// already in the team is ErrConflict rather than a silent second row.
//
// The administrator need not hold the team's permissions personally: the
// handbook is explicit that "the approved rule does not invent an additional
// requirement that this membership administrator personally possess each of the
// team's business permissions".
func (s *Service) AddMember(ctx context.Context, area domain.Area, identity domain.Identity, teamID, humanID string) error {
	return s.membership(ctx, area, identity, teamID, humanID, true)
}

// RemoveMember takes one human out of one team.
func (s *Service) RemoveMember(ctx context.Context, area domain.Area, identity domain.Identity, teamID, humanID string) error {
	return s.membership(ctx, area, identity, teamID, humanID, false)
}

func (s *Service) membership(ctx context.Context, area domain.Area, identity domain.Identity, teamID, humanID string, add bool) error {
	if err := s.teamWriteAuthority(ctx, area, identity); err != nil {
		return err
	}
	m := domain.Membership{TeamID: teamID, HumanID: humanID}
	return s.provider.UpdateAdministered(ctx, area, func(snapshot storage.Snapshot) (storage.WriteSet, error) {
		if snapshot.Area != area {
			return storage.WriteSet{}, domain.ErrRejected
		}
		// `auth:group::write` "includes human membership" — Q-092 — so moving a
		// human in or out is the same authority as changing the team, bounded by
		// the team being changed.
		if err := s.authorizeTeam(ctx, snapshot, identity, groupWrite, teamID, s.clock.Now()); err != nil {
			return storage.WriteSet{}, err
		}
		if err := validation.CheckMembership(snapshot.Teams, m); err != nil {
			return storage.WriteSet{}, err
		}
		present := false
		for _, held := range snapshot.Memberships {
			if held == m {
				present = true
				break
			}
		}
		if add && present {
			return storage.WriteSet{}, domain.ErrConflict
		}
		if !add && !present {
			return storage.WriteSet{}, domain.ErrNotFound
		}
		if err := ctx.Err(); err != nil {
			return storage.WriteSet{}, err
		}
		proposed := m
		if add {
			return storage.WriteSet{AddedMembership: &proposed}, nil
		}
		return storage.WriteSet{RemovedMembership: &proposed}, nil
	})
}

// refuseAffectedBindings rejects while an enabled binding sits at or beneath
// this team.
//
// Beneath, not just at: a subteam's authority is resolved through the chain
// above it, so moving an ancestor moves the subteam's inherited scope too. A
// team records its parent and not its children, so the walk builds the index it
// needs first.
//
// Disabled bindings are left out. They hold nothing to re-anchor, and enabling
// one revalidates against the structure as it then is.
func refuseAffectedBindings(snapshot, chain storage.Snapshot, teamID string) error {
	// A children index, built once, then walked downward. Growing the set by
	// repeated passes over every team was correct and cost a pass per level —
	// and team depth is whatever a tenant makes it, with no cap at creation.
	children := make(map[string][]string, len(snapshot.Teams))
	for id, team := range snapshot.Teams {
		if team.ParentID != "" {
			children[team.ParentID] = append(children[team.ParentID], id)
		}
	}
	// Sorted, so the walk visits a branching tree in the same order every run.
	// Nothing about the answer depends on order — the set is complete either way
	// — but a walk whose order is a map's is one whose bugs appear and vanish
	// between runs, and a test cannot pin what it cannot reproduce.
	for parent := range children {
		sort.Strings(children[parent])
	}
	subtree := map[string]bool{teamID: true}
	for frontier := []string{teamID}; len(frontier) > 0; {
		id := frontier[len(frontier)-1]
		frontier = frontier[:len(frontier)-1]
		for _, child := range children[id] {
			// Guarded against a cycle already in the store: the set only grows,
			// so a team already in it is never queued twice.
			if !subtree[child] {
				subtree[child] = true
				frontier = append(frontier, child)
			}
		}
	}
	// Both areas' assignments, over the one subtree. The teams are tenant-wide, so
	// a move re-anchors an administrative binding exactly as it re-anchors a
	// business one — and B13 is about what the move affects, not about which area
	// happens to record it. When there is no separate administrative chain the two
	// maps are the same map, and scanning it twice costs a pass and answers the
	// same.
	for _, assignments := range []map[string]domain.Assignment{snapshot.Assignments, chain.Assignments} {
		for _, assignment := range assignments {
			if assignment.Status != "enabled" || assignment.Recipient.Type != "group" {
				continue
			}
			if subtree[assignment.Recipient.ID] {
				return domain.ErrConflict
			}
		}
	}
	return nil
}
