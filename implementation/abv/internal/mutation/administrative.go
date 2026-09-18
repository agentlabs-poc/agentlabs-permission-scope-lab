package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"time"
)

// The three group verbs Q-092 approved. Create covers subteams; write covers
// human membership; delete is its own authority.
//
// They are verbs rather than whole identifiers because a platform permission's
// leading noun is whatever the platform chose for its namespace, and this
// package does not get to hold that opinion — see groupPermission.
const (
	groupCreate = "create"
	groupWrite  = "write"
	groupDelete = "delete"
)

// administrativeChain returns the snapshot an administrative question is
// answered against — Q-155 / ADMIN-007.
//
// Administrative authority is an ordinary grant whose chain begins at the Auth
// root, and that root is in the platform's namespace. A nil Administrative means
// the operation is already being performed there, so the snapshot in hand is the
// chain. Anything else, and the provider read the chain alongside this area's
// records inside the same transaction.
func administrativeChain(s storage.Snapshot) storage.Snapshot {
	if s.Administrative != nil {
		return *s.Administrative
	}
	return s
}

// groupPermission spells one group verb in the platform's own namespace. The
// administrative chain's area *is* that namespace, which is why the namespace
// never has to be configured twice.
func groupPermission(chain storage.Snapshot, verb string) string {
	return chain.Area.ApplicationID() + ":group::" + verb
}

// authorizeTeam is the one check that replaced three fixture comparisons.
//
// The material is a single `team` predicate, which is Q-156's platform scope key.
// An unscoped administrative route — the tenant administrator's, the only kind a
// root can hand out with no predicates — satisfies any team. A route scoped to one
// team satisfies that team and, because a request carries one value per key, no
// other.
//
// The id's shape is checked here rather than trusted from the operation's own
// validation, which runs afterwards. An authorization step whose safety depends on
// what the *next* check does is one reordering away from authorizing a string
// nobody validated, and an empty id used to be spelled the same way as "this
// operation has no bounding team" — see authorizeUnbounded, which is now the only
// way to say that.
//
// No test can tell this check from the validation behind it: every caller's
// validation refuses the same inputs with the same error, so removing this line
// changes no observable answer and a mutation check will report it surviving. It
// stays anyway, and this paragraph is the reason — a guard at an authorization
// boundary that is redundant today is not the same as one that is unnecessary.
func (s *Service) authorizeTeam(ctx context.Context, snapshot storage.Snapshot, identity domain.Identity, verb, teamID string, now time.Time) error {
	if !codec.ValidRoleID(teamID) {
		return domain.ErrMalformed
	}
	chain := administrativeChain(snapshot)
	return lineage.Authorize(ctx, chain, identity, groupPermission(chain, verb),
		map[string]string{validation.TeamScopeKey: teamID}, now)
}

// authorizeUnbounded is the case where no team bounds the operation: creating a
// top-level team, and moving one up to become top-level. Empty material is
// satisfied only by a route carrying no predicates at all, so "only the tenant
// administrator may" is what it computes rather than what it asserts.
func (s *Service) authorizeUnbounded(ctx context.Context, snapshot storage.Snapshot, identity domain.Identity, verb string, now time.Time) error {
	chain := administrativeChain(snapshot)
	return lineage.Authorize(ctx, chain, identity, groupPermission(chain, verb), map[string]string{}, now)
}

// teamWriteAuthority runs the checks every administrative team write shares
// before a transaction is opened. It replaced the interface lookup the three
// gates used to need: a team write no longer asks a deployment whether the
// caller may act, it resolves the caller's own authority.
func (s *Service) teamWriteAuthority(ctx context.Context, area domain.Area, identity domain.Identity) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := area.Validate(); err != nil {
		return err
	}
	return validateSupportedIdentity(identity)
}

// authorizeUnder authorizes an operation bounded by a parent that may be absent:
// creating a team under one, and the far end of a move. An empty parent is not a
// wildcard — it is the absence of a bounding team, which only an unscoped route
// satisfies.
func (s *Service) authorizeUnder(ctx context.Context, snapshot storage.Snapshot, identity domain.Identity, verb, parentID string, now time.Time) error {
	if parentID == "" {
		return s.authorizeUnbounded(ctx, snapshot, identity, verb, now)
	}
	return s.authorizeTeam(ctx, snapshot, identity, verb, parentID, now)
}

// refuseAdministrativeDependants refuses while the administrative chain names a
// team, which the operation's own area cannot see.
//
// Q-155 made the administrative chain a chain of grants and assignments like any
// other, and it binds the *same tenant-wide team records* — so every rule about
// what depends on a team has to read both. Without this, "deleting a team is
// refused while an assignment names it" stopped being true the moment
// administration became a grant: the team holding a tenant's Auth root would be
// deletable whenever it happened to hold no business assignment, which takes away
// all administrative authority for that tenant with no record of why.
func refuseAdministrativeDependants(chain storage.Snapshot, teamID string) error {
	for _, assignment := range chain.Assignments {
		if assignment.Recipient.Type == "group" && assignment.Recipient.ID == teamID {
			return domain.ErrConflict
		}
	}
	return nil
}
