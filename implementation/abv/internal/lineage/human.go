package lineage

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"context"
	"errors"
	"sort"
	"time"
)

// maxHumanQueryRecords matches the prototype SQLite snapshot safety ceiling;
// it is not a canonical authority rule.
const maxHumanQueryRecords = 10_000

// ResolveHuman returns eligible routes held by teams the human directly joins,
// carrying one permission.
//
// It holds the actor to the subject, and that is not redundant with the subject
// check inside the walk. Its caller is localadapter.SQLiteAuthoritySource, which
// is opened with a path, a clock and a registry — and no administration, so it
// has no gate at all. This rule IS that path's gate.
//
// ResolveAuthority is the loosened entry, and it is loosened only because
// CheckAuthorityRead runs before it. Loosening the shared walk instead of the
// gated entry took the check away from the path that had nothing else: a forged
// or absent actor returned the subject's full route set, predicates and
// contributing grant ids included, stopped only by a rule in a different module
// that happened to run first.
func ResolveHuman(ctx context.Context, s storage.Snapshot, identity domain.Identity, permission string, now time.Time) ([]domain.Route, error) {
	if err := validateIdentity(identity); err != nil {
		return nil, err
	}
	held, err := collectHumanRoutes(ctx, s, identity, []string{permission}, now)
	if err != nil {
		return nil, err
	}
	routes := make([]domain.Route, len(held))
	for i := range held {
		routes[i] = held.route(i)
	}
	return routes, nil
}

// heldRoute is one resolved route and the assignment that holds it. The
// assignment is what orders the result and what the explanation hangs from.
type heldRoute struct {
	assignmentID string
	route        domain.Route
}

type heldRoutes []heldRoute

func (h heldRoutes) route(i int) domain.Route { return h[i].route }

// collectHumanRoutes is the walk both reads share: every eligible route held by
// a team the human directly joins, optionally narrowed to some permissions.
//
// An empty filter means everything the human holds. That is the complete answer,
// and the complete answer is the one worth caching — a gate deciding one request
// passes a filter, a menu passes none.
func collectHumanRoutes(ctx context.Context, s storage.Snapshot, identity domain.Identity, filter []string, now time.Time) (heldRoutes, error) {
	fail := func(err error) (heldRoutes, error) { return nil, err }
	if ctx == nil {
		return fail(domain.ErrMalformed)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if err := validateSubject(identity); err != nil {
		return fail(err)
	}
	if err := s.Area.Validate(); err != nil {
		return fail(err)
	}
	if s.Catalog.ApplicationID != s.Area.ApplicationID() {
		return fail(domain.ErrRejected)
	}
	// A filtered permission must be registered and active. Asking about one that
	// is not is a caller mistake, and answering "you hold nothing" would hide it.
	for _, permission := range filter {
		if err := codec.PermissionList([]string{permission}); err != nil {
			return fail(err)
		}
		definition, ok := s.Catalog.Permissions[permission]
		if !ok || definition.ID != permission || !definition.Active {
			return fail(domain.ErrRejected)
		}
	}
	if len(s.Memberships)+len(s.Assignments) > maxHumanQueryRecords {
		return fail(storage.ErrSnapshotLimit)
	}

	teams := make(map[string]bool)
	for _, membership := range s.Memberships {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		if invalidIdentityString(membership.TeamID) || invalidIdentityString(membership.HumanID) {
			return fail(domain.ErrMalformed)
		}
		team, exists := s.Teams[membership.TeamID]
		if !exists || team.ID != membership.TeamID {
			return fail(domain.ErrRejected)
		}
		if membership.HumanID == identity.HumanID {
			teams[membership.TeamID] = true
		}
	}

	resolved := make(heldRoutes, 0)
	for key, assignment := range s.Assignments {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		if key != assignment.ID {
			return fail(domain.ErrRejected)
		}
		// A direct human assignment is not an eligible route here — group-held
		// support is deliberate, and enforced at both write and read. It is one
		// route's ineligibility, so it is skipped like any other, for this human
		// exactly as for anybody else.
		//
		// It used to abort the whole answer when the recipient *was* the subject,
		// which made the difference observable: a caller could tell whether a
		// named human held a direct assignment by whether the question came back
		// 501 or 200. That is the enumeration this service merges its refusals to
		// prevent, reopened one error kind over — and it also took away every
		// group route the human legitimately held.
		if assignment.Recipient.Type == "user" {
			continue
		}
		if assignment.Recipient.Type != "group" || !teams[assignment.Recipient.ID] {
			continue
		}
		route, err := ResolveTeamAssignment(s, assignment.ID, now)
		if contextErr := ctx.Err(); contextErr != nil {
			return fail(contextErr)
		}
		if err != nil {
			// "Missing support stops the affected authority route, not
			// necessarily all authority of that user or group" —
			// authority-lineage.md:169. Only ErrInactive was skipped, so a
			// permission retired in one grant, a missing role revision, or a
			// child that no longer narrows its parent took away every *other*
			// grant the human held — and through localsource that reached the
			// application as an outage rather than a denial.
			if routeScoped(err) {
				continue
			}
			return fail(err)
		}
		if carries(route.Permissions, filter) {
			resolved = append(resolved, heldRoute{assignment.ID, route})
		}
	}

	sort.Slice(resolved, func(i, j int) bool { return resolved[i].assignmentID < resolved[j].assignmentID })
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	return resolved, nil
}

// carries reports whether a route answers the filter. An empty filter is
// satisfied by every route, which is what makes it "everything".
func carries(permissions, filter []string) bool {
	if len(filter) == 0 {
		return true
	}
	for _, candidate := range permissions {
		for _, wanted := range filter {
			if candidate == wanted {
				return true
			}
		}
	}
	return false
}

// routeScoped reports whether a failure describes one route rather than the
// answer as a whole.
//
// Two kinds qualify, and they are the two a chain walk says about the chain it
// walked: support that is disabled or lapsed, and authority that no longer holds
// — a retired permission, a missing role revision, a child outside its parent.
//
// Everything else stays fatal, deliberately. A cycle, a duplicate binding, a
// status the model does not define and a record the walk could not read are
// integrity failures of the area, and none of them licenses answering "this
// human holds nothing" when the truth is that nobody knows. Widening this to
// every rejection was the first attempt, and two existing tests caught it: they
// are named for failing closed on invalid evidence, which is exactly what it
// would have stopped doing.
func routeScoped(err error) bool {
	return errors.Is(err, ErrInactive) || errors.Is(err, ErrIneligible)
}
