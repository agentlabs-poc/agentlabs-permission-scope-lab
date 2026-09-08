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
	assignment, ok := s.Assignments[sourceID]
	if !ok || assignment.ID != sourceID {
		return domain.ErrRejected
	}
	if assignment.Recipient.Type == "user" {
		return domain.ErrUnsupported
	}
	if assignment.Recipient.Type != "group" || !validAssignment(assignment) || assignment.GrantID != parent.GrantID {
		return domain.ErrRejected
	}
	unique, err := uniqueAssignment(s, parent.GrantID, assignment.Recipient.ID)
	if err != nil {
		return err
	}
	if unique.ID != sourceID {
		return domain.ErrRejected
	}
	assignment = unique
	if err := validateTeamChain(s, assignment.Recipient.ID); err != nil {
		return err
	}
	resolver := routeResolver{s: s, now: now, active: map[string]bool{}, remaining: maxChainSteps}
	anchored, err := resolver.resolve(assignment, assignment.Recipient.ID)
	if err != nil {
		return err
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
