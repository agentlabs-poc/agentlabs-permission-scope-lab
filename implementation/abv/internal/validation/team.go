package validation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"strings"
	"unicode/utf8"
)

// CheckTeamCreation validates a proposed team against the tenant's existing
// ones. The id is issued rather than supplied, so what is checked here is the
// name and the parent.
func CheckTeamCreation(teams map[string]domain.Team, proposed domain.Team) error {
	if !codec.ValidRoleID(proposed.ID) {
		return domain.ErrMalformed
	}
	if strings.TrimSpace(proposed.Name) == "" || !utf8.ValidString(proposed.Name) {
		return domain.ErrMalformed
	}
	if _, exists := teams[proposed.ID]; exists {
		return domain.ErrConflict
	}
	// An empty parent is a root. Any other parent must be a team that exists:
	// the foreign key that used to enforce this went with the fold, so the check
	// lives here now.
	if proposed.ParentID == "" {
		return nil
	}
	if !codec.ValidRoleID(proposed.ParentID) {
		return domain.ErrMalformed
	}
	if parent, ok := teams[proposed.ParentID]; !ok || parent.ID != proposed.ParentID {
		return domain.ErrNotFound
	}
	return nil
}

// CheckTeamReparent validates a change of parent.
//
// It refuses a cycle at the write rather than letting resolution catch it. A
// cycle would make a team its own ancestor, and accepting one would leave the
// store holding a state that can never resolve — the same defect a cascading
// delete would produce, one step later. The rule throughout is: validate at the
// write, refuse rather than defer.
func CheckTeamReparent(teams map[string]domain.Team, id, parentID string) error {
	if !codec.ValidRoleID(id) {
		return domain.ErrMalformed
	}
	team, ok := teams[id]
	if !ok || team.ID != id {
		return domain.ErrNotFound
	}
	if parentID == "" {
		return nil
	}
	if !codec.ValidRoleID(parentID) {
		return domain.ErrMalformed
	}
	if parent, ok := teams[parentID]; !ok || parent.ID != parentID {
		return domain.ErrNotFound
	}
	if parentID == id {
		return domain.ErrRejected
	}
	// Walk up from the proposed parent: meeting this team means the move would
	// put it inside its own subtree.
	seen := map[string]bool{id: true}
	for steps, walk := 0, parentID; walk != ""; steps++ {
		if steps == maxTeamDepth {
			return domain.ErrUnavailable
		}
		if seen[walk] {
			return domain.ErrRejected
		}
		seen[walk] = true
		next, ok := teams[walk]
		if !ok || next.ID != walk {
			return domain.ErrRejected
		}
		walk = next.ParentID
	}
	return nil
}

// maxTeamDepth bounds the walk, so a hierarchy corrupted by some other route
// cannot make a write loop forever.
const maxTeamDepth = 64

// CheckTeamDeletion refuses while anything depends on the team.
//
// A child team, a membership, or an assignment naming it each block the delete.
// Nothing cascades and nothing is soft-deleted: a delete that quietly removed
// memberships would make one administrative act perform another, and the
// handbook keeps team administration, membership administration and assignment
// authority distinct.
func CheckTeamDeletion(teams map[string]domain.Team, memberships []domain.Membership, ownerships []domain.Ownership, assignments map[string]domain.Assignment, id string) error {
	if !codec.ValidRoleID(id) {
		return domain.ErrMalformed
	}
	team, ok := teams[id]
	if !ok || team.ID != id {
		return domain.ErrNotFound
	}
	for _, other := range teams {
		if other.ParentID == id {
			return domain.ErrConflict
		}
	}
	for _, m := range memberships {
		if m.TeamID == id {
			return domain.ErrConflict
		}
	}
	// An ownership depends on its team exactly as a membership does: deleting
	// the team without it would leave a row naming a team that no longer exists.
	// Ownership grants no authority, but an orphan record is still a record
	// nothing can resolve.
	for _, o := range ownerships {
		if o.TeamID == id {
			return domain.ErrConflict
		}
	}
	for _, a := range assignments {
		if a.Recipient.Type == "group" && a.Recipient.ID == id {
			return domain.ErrConflict
		}
	}
	return nil
}

// CheckMembership validates adding or removing one human's place in one team.
// The team must exist — the foreign key that enforced that went with the fold.
func CheckMembership(teams map[string]domain.Team, m domain.Membership) error {
	if !codec.ValidRoleID(m.TeamID) || !codec.ValidHumanID(m.HumanID) {
		return domain.ErrMalformed
	}
	if team, ok := teams[m.TeamID]; !ok || team.ID != m.TeamID {
		return domain.ErrNotFound
	}
	return nil
}

// CheckOwnership validates an ownership write. It is CheckMembership's shape
// because ownership is membership's shape — and it checks nothing more, which is
// the point rather than an omission.
//
// Ownership grants no authority (Q-099), so there is no ceiling to stay within
// and no lineage to walk. A membership distributes the team's authority to a
// human; an ownership lets a human administer the team and gives them none of
// its business authority. The lighter check follows from the weaker fact.
func CheckOwnership(teams map[string]domain.Team, o domain.Ownership) error {
	if !codec.ValidRoleID(o.TeamID) || !codec.ValidHumanID(o.HumanID) {
		return domain.ErrMalformed
	}
	if team, ok := teams[o.TeamID]; !ok || team.ID != o.TeamID {
		return domain.ErrNotFound
	}
	return nil
}
