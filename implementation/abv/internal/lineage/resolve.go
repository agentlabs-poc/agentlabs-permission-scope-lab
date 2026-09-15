// Package lineage resolves established, area-bound grant support without I/O.
package lineage

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// $self travels the chain like any other scope value.
//
// It used to stop it. Two checks — one on the proposed child, one at every step
// of the climb — refused any content whose scope carried the token, as
// ErrUnsupported and without a comment. The effect was worse than refusing the
// grant that used it: a $self grant anywhere in a chain killed every grant below
// it too.
//
// SELF-001 settles the opposite, and GROUP-004 makes it the preferred practice:
// "An Employees group may receive one self-scoped payslip-read grant. Vinay's
// group membership makes that grant applicable to Vinay, and self resolves
// through his trusted user-to-employee relationship. Maya's membership makes the
// same grant applicable to Maya, with her own employee relationship." A single
// self-service grant instead of one per employee is the recommended shape, and
// it is exactly what those two checks refused.
//
// Nothing here resolves the token, and that is the division: Auth carries it to
// the evaluator, which binds it to the authorizing human at match time, and what
// the key *means* — the user-to-employee relationship — is the application's
// under CHARTER-001.

// maxChainSteps is a defensive traversal bound, not a canonical lineage limit.
const maxChainSteps = 256

// ErrInactive identifies established but currently ineffective lineage while
// remaining rejection-compatible for existing issuance callers.
var ErrInactive = fmt.Errorf("inactive authority: %w", domain.ErrRejected)

func ResolveParentTeam(s storage.Snapshot, child domain.GrantContent, recipientTeamID string, now time.Time) (domain.Route, error) {
	fail := func(err error) (domain.Route, error) { return domain.Route{}, err }
	if err := s.Area.Validate(); err != nil {
		return fail(err)
	}
	if s.Catalog.ApplicationID != s.Area.ApplicationID() {
		return fail(domain.ErrRejected)
	}
	if strings.TrimSpace(recipientTeamID) == "" {
		return fail(domain.ErrMalformed)
	}
	if err := validateSelectedContent(s, child, now); err != nil {
		return fail(err)
	}
	team, ok := s.Teams[recipientTeamID]
	if !ok || team.ID != recipientTeamID {
		return fail(domain.ErrRejected)
	}
	if err := validateTeamChain(s, recipientTeamID); err != nil {
		return fail(err)
	}
	if team.ParentID == "" || child.ParentGrantID == "" {
		return fail(domain.ErrRejected)
	}
	assignment, err := uniqueAssignment(s, child.ParentGrantID, team.ParentID)
	if err != nil {
		return fail(err)
	}
	resolver := routeResolver{s: s, now: now, active: map[string]bool{child.GrantID: true}, remaining: maxChainSteps}
	return resolver.resolve(assignment, team.ParentID)
}

type routeResolver struct {
	s         storage.Snapshot
	now       time.Time
	active    map[string]bool
	remaining int
}

func (r *routeResolver) resolve(assignment domain.Assignment, holderTeamID string) (domain.Route, error) {
	fail := func(err error) (domain.Route, error) { return domain.Route{}, err }
	if r.remaining == 0 {
		return fail(domain.ErrUnavailable)
	}
	r.remaining--
	if r.active[assignment.GrantID] {
		return fail(domain.ErrRejected)
	}
	r.active[assignment.GrantID] = true
	defer delete(r.active, assignment.GrantID)
	if assignment.Status != "enabled" {
		return fail(ErrInactive)
	}
	content, err := assignmentContent(r.s, assignment)
	if err != nil {
		return fail(err)
	}
	if err = validateSelectedContent(r.s, content, r.now); err != nil {
		return fail(err)
	}
	team, ok := r.s.Teams[holderTeamID]
	if !ok || team.ID != holderTeamID {
		return fail(domain.ErrRejected)
	}
	if content.ParentGrantID == "" {
		if !r.s.TrustedRoots[content.GrantID] || team.ParentID != "" {
			return fail(domain.ErrRejected)
		}
		return rootRoute(r.s, content, assignment.ID)
	}
	if team.ParentID == "" {
		return fail(domain.ErrRejected)
	}
	parentAssignment, err := uniqueAssignment(r.s, content.ParentGrantID, team.ParentID)
	if err != nil {
		return fail(err)
	}
	parent, err := r.resolve(parentAssignment, team.ParentID)
	if err != nil {
		return fail(err)
	}
	result, err := validation.Narrow(r.s.Area, parent, content, r.s.Roles)
	if err != nil {
		return fail(err)
	}
	result.AssignmentIDs = append(result.AssignmentIDs, assignment.ID)
	return result, nil
}

func rootRoute(s storage.Snapshot, content domain.GrantContent, assignmentID string) (domain.Route, error) {
	// A root's ceiling is sliced by boundary, and the catalog is not.
	//
	// An application's catalog is its own permissions union every platform one,
	// which is right for evaluation: a request inside an application may
	// legitimately require a platform permission. It is wrong for a ceiling.
	// Without this, an application root would carry every auth:* permission —
	// including whichever one authorises establishing an application root, so
	// the thing created by the authority could create more of that authority.
	//
	// The slice follows the namespace in key3, which is what the root is a root
	// *of*: the platform's own namespace takes the platform permissions, an
	// application takes its own.
	// One rule covers both roots: a root takes the permissions registered under
	// its OWN namespace. For an application root that is the application's own;
	// for the Auth root, whose namespace is the platform's, it is the platform
	// permissions. No discriminator is needed and none is invented.
	namespace := s.Catalog.ApplicationID
	permissions := make([]string, 0, len(s.Catalog.Permissions))
	for id, definition := range s.Catalog.Permissions {
		if definition.ID != id {
			return domain.Route{}, domain.ErrRejected
		}
		if definition.Active && definition.Namespace == namespace {
			permissions = append(permissions, id)
		}
	}
	sort.Strings(permissions)
	computed := content
	computed.Permissions = permissions
	computed.RoleID, computed.RoleRevision = "", 0
	if err := validation.CheckContent(s.Area, s.Catalog, computed, s.Roles); err != nil {
		return domain.Route{}, err
	}
	result := domain.Route{Area: s.Area, GrantID: content.GrantID, Permissions: permissions, AssignmentIDs: []string{assignmentID}}
	if content.Validity != nil {
		result.Validities = append(result.Validities, copyValidity(*content.Validity))
	}
	keys := make([]string, 0, len(content.Scope))
	for key := range content.Scope {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result.Predicates = append(result.Predicates, domain.Predicate{Key: key, Value: content.Scope[key], SourceGrantID: content.GrantID})
	}
	return result, nil
}

func validateSelectedContent(s storage.Snapshot, content domain.GrantContent, now time.Time) error {
	stored, ok := s.Contents[domain.GrantKey{ID: content.GrantID, Revision: content.Revision}]
	if !ok || !reflect.DeepEqual(stored, content) {
		return domain.ErrRejected
	}
	control, ok := s.Controls[content.GrantID]
	if !ok || control.ID != content.GrantID || control.Version != "1" || (control.Status != "enabled" && control.Status != "disabled") {
		return domain.ErrRejected
	}
	if control.Status == "disabled" {
		return ErrInactive
	}
	if err := validation.CheckContent(s.Area, s.Catalog, content, s.Roles); err != nil {
		return err
	}
	if content.Validity != nil && !eligible(*content.Validity, now) {
		return ErrInactive
	}
	return nil
}

func assignmentContent(s storage.Snapshot, assignment domain.Assignment) (domain.GrantContent, error) {
	content, ok := s.Contents[domain.GrantKey{ID: assignment.GrantID, Revision: assignment.GrantRevision}]
	if !ok || content.GrantID != assignment.GrantID || content.Revision != assignment.GrantRevision {
		return domain.GrantContent{}, domain.ErrRejected
	}
	return content, nil
}

func uniqueAssignment(s storage.Snapshot, grantID, teamID string) (domain.Assignment, error) {
	var found domain.Assignment
	count := 0
	for key, assignment := range s.Assignments {
		if key != assignment.ID {
			return domain.Assignment{}, domain.ErrRejected
		}
		if assignment.GrantID == grantID && assignment.Recipient.Type == "group" && assignment.Recipient.ID == teamID {
			if !validAssignment(assignment) {
				return domain.Assignment{}, domain.ErrRejected
			}
			found, count = assignment, count+1
		}
	}
	if count == 0 {
		return domain.Assignment{}, ErrInactive
	}
	if count != 1 {
		return domain.Assignment{}, domain.ErrRejected
	}
	return found, nil
}

func validAssignment(a domain.Assignment) bool {
	return a.Version == "1" && strings.TrimSpace(a.ID) != "" && strings.TrimSpace(a.GrantID) != "" && a.GrantRevision > 0 &&
		a.Recipient.Type == "group" && strings.TrimSpace(a.Recipient.ID) != "" && (a.Status == "enabled" || a.Status == "disabled")
}

func validateTeamChain(s storage.Snapshot, start string) error {
	seen := map[string]bool{}
	for steps, id := 0, start; id != ""; steps++ {
		if steps == maxChainSteps {
			return domain.ErrUnavailable
		}
		if seen[id] {
			return domain.ErrRejected
		}
		seen[id] = true
		team, ok := s.Teams[id]
		if !ok || team.ID != id {
			return domain.ErrRejected
		}
		id = team.ParentID
	}
	return nil
}

func eligible(validity domain.Validity, now time.Time) bool {
	return (validity.NotBefore == nil || !now.Before(*validity.NotBefore)) && (validity.ExpiresAt == nil || now.Before(*validity.ExpiresAt))
}

func copyValidity(v domain.Validity) domain.Validity {
	result := v
	if v.NotBefore != nil {
		value := *v.NotBefore
		result.NotBefore = &value
	}
	if v.ExpiresAt != nil {
		value := *v.ExpiresAt
		result.ExpiresAt = &value
	}
	return result
}
