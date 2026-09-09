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
	if containsSelf(child) {
		return fail(domain.ErrUnsupported)
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
	if containsSelf(content) {
		return fail(domain.ErrUnsupported)
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
	permissions := make([]string, 0, len(s.Catalog.Permissions))
	for id, definition := range s.Catalog.Permissions {
		if definition.ID != id {
			return domain.Route{}, domain.ErrRejected
		}
		if definition.Active {
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

func containsSelf(content domain.GrantContent) bool {
	for _, value := range content.Scope {
		if value == "$self" {
			return true
		}
	}
	return false
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
