package lineage

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

// HasSource verifies source access separately from recipient lineage. The route
// is treated only as an evidence reference and is reconstructed from snapshot.
func HasSource(s storage.Snapshot, identity domain.Identity, parent domain.Route, now time.Time) error {
	if err := validateIdentity(identity); err != nil {
		return err
	}
	if err := s.Area.Validate(); err != nil {
		return err
	}
	if s.Catalog.ApplicationID != s.Area.ApplicationID() || parent.Area != s.Area || parent.GrantID == "" || len(parent.AssignmentIDs) == 0 {
		return domain.ErrRejected
	}
	sourceID := parent.AssignmentIDs[len(parent.AssignmentIDs)-1]
	anchored, err := ResolveTeamAssignment(s, sourceID, now)
	if err != nil {
		return err
	}
	assignment := s.Assignments[sourceID]
	if assignment.GrantID != parent.GrantID {
		return domain.ErrRejected
	}
	if !reflect.DeepEqual(anchored, parent) {
		return domain.ErrRejected
	}
	for _, membership := range s.Memberships {
		if membership.TeamID == assignment.Recipient.ID && membership.HumanID == identity.HumanID {
			return nil
		}
	}
	return domain.ErrRejected
}

// ResolveTeamAssignment reconstructs an eligible established route from its
// exact team-held assignment. It performs no membership or administration check.
func ResolveTeamAssignment(s storage.Snapshot, assignmentID string, now time.Time) (domain.Route, error) {
	fail := func(err error) (domain.Route, error) { return domain.Route{}, err }
	if err := s.Area.Validate(); err != nil {
		return fail(err)
	}
	if s.Catalog.ApplicationID != s.Area.ApplicationID() {
		return fail(domain.ErrRejected)
	}
	if invalidIdentityString(assignmentID) {
		return fail(domain.ErrMalformed)
	}
	assignment, ok := s.Assignments[assignmentID]
	if !ok || assignment.ID != assignmentID {
		return fail(domain.ErrRejected)
	}
	if assignment.Recipient.Type == "user" {
		return fail(domain.ErrUnsupported)
	}
	if assignment.Recipient.Type != "group" || !validAssignment(assignment) {
		return fail(domain.ErrRejected)
	}
	unique, err := uniqueAssignment(s, assignment.GrantID, assignment.Recipient.ID)
	if err != nil || unique.ID != assignmentID {
		if err != nil {
			return fail(err)
		}
		return fail(domain.ErrRejected)
	}
	if err := validateTeamChain(s, assignment.Recipient.ID); err != nil {
		return fail(err)
	}
	resolver := routeResolver{s: s, now: now, active: map[string]bool{}, remaining: maxChainSteps}
	return resolver.resolve(unique, unique.Recipient.ID)
}

func validateIdentity(identity domain.Identity) error {
	if invalidIdentityString(identity.Version) || invalidIdentityString(identity.Actor.Type) || invalidIdentityString(identity.Actor.ID) || invalidIdentityString(identity.HumanID) {
		return domain.ErrMalformed
	}
	if identity.Version != "1" || identity.Actor.Type != "user" || identity.Actor.ID != identity.HumanID {
		return domain.ErrUnsupported
	}
	return nil
}

func invalidIdentityString(value string) bool {
	return strings.TrimSpace(value) == "" || value == "*" || !utf8.ValidString(value)
}
