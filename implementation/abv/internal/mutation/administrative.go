package mutation

import (
	"agentlabs.local/abv/domain"
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
// The material is a single `team` predicate, which is Q-156's platform scope
// key. An unscoped administrative route — the tenant administrator's, the only
// kind a root can hand out with no predicates — satisfies any team. A route
// scoped to one team satisfies that team and, because a request carries one
// value per key, no other.
//
// An empty teamID is not a wildcard: it is the absence of a bounding team, which
// only an unscoped route can satisfy. Creating a top-level team is the case, and
// "only the tenant administrator may" is the intended reading.
func (s *Service) authorizeTeam(ctx context.Context, snapshot storage.Snapshot, identity domain.Identity, verb, teamID string, now time.Time) error {
	chain := administrativeChain(snapshot)
	material := map[string]string{}
	if teamID != "" {
		material[validation.TeamScopeKey] = teamID
	}
	return lineage.Authorize(ctx, chain, identity, groupPermission(chain, verb), material, now)
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
