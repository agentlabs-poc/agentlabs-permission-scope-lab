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
func ResolveHuman(ctx context.Context, s storage.Snapshot, identity domain.Identity, permission string, now time.Time) ([]domain.Route, error) {
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
	if err := validateIdentity(identity); err != nil {
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
		if assignment.Recipient.Type == "user" {
			if assignment.Recipient.ID == identity.HumanID {
				return fail(domain.ErrUnsupported)
			}
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
			if errors.Is(err, ErrInactive) {
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
