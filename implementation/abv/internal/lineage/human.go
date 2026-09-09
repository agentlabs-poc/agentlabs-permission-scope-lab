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

// ResolveHuman returns eligible routes held by teams the human directly joins.
func ResolveHuman(ctx context.Context, s storage.Snapshot, identity domain.Identity, permission string, now time.Time) ([]domain.Route, error) {
	fail := func(err error) ([]domain.Route, error) { return nil, err }
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
	if err := codec.PermissionList([]string{permission}); err != nil {
		return fail(err)
	}
	definition, ok := s.Catalog.Permissions[permission]
	if !ok || definition.ID != permission || !definition.Active {
		return fail(domain.ErrRejected)
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

	type result struct {
		assignmentID string
		route        domain.Route
	}
	resolved := make([]result, 0)
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
		for _, candidate := range route.Permissions {
			if candidate == permission {
				resolved = append(resolved, result{assignment.ID, route})
				break
			}
		}
	}

	sort.Slice(resolved, func(i, j int) bool { return resolved[i].assignmentID < resolved[j].assignmentID })
	routes := make([]domain.Route, len(resolved))
	for i := range resolved {
		routes[i] = resolved[i].route
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	return routes, nil
}
