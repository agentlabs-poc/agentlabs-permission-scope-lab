package lineage

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"sort"
	"strings"
)

type bindingKey struct{ grantID, teamID string }
type recipientBindingKey struct {
	grantID   string
	recipient domain.Recipient
}

type dependentNode struct {
	assignment domain.Assignment
	content    domain.GrantContent
}

// DependentTeamAssignments returns every structurally downstream team binding.
// Status and validity do not remove records from this integrity inventory.
func DependentTeamAssignments(ctx context.Context, s storage.Snapshot, assignmentID string) ([]domain.Assignment, error) {
	fail := func(err error) ([]domain.Assignment, error) { return nil, err }
	if ctx == nil || strings.TrimSpace(assignmentID) == "" {
		return fail(domain.ErrMalformed)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if err := s.Area.Validate(); err != nil {
		return fail(err)
	}
	if s.Catalog.ApplicationID != s.Area.ApplicationID() {
		return fail(domain.ErrRejected)
	}

	holdings := make(map[bindingKey]dependentNode, len(s.Assignments))
	children := make(map[bindingKey][]dependentNode, len(s.Assignments))
	nonGroupParents := make(map[string]bool)
	bindings := make(map[recipientBindingKey]bool, len(s.Assignments))
	for id, assignment := range s.Assignments {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		if id != assignment.ID || !validDependentAssignment(assignment) {
			return fail(domain.ErrRejected)
		}
		content, err := assignmentContent(s, assignment)
		if err != nil {
			return fail(err)
		}
		if err = validation.CheckContent(s.Area, s.Catalog, content, s.Roles); err != nil {
			return fail(err)
		}
		binding := recipientBindingKey{assignment.GrantID, assignment.Recipient}
		if bindings[binding] {
			return fail(domain.ErrRejected)
		}
		bindings[binding] = true
		if assignment.Recipient.Type != "group" {
			if assignment.Status == "enabled" && content.ParentGrantID != "" {
				nonGroupParents[content.ParentGrantID] = true
			}
			continue
		}
		team, ok := s.Teams[assignment.Recipient.ID]
		if !ok || team.ID != assignment.Recipient.ID {
			return fail(domain.ErrRejected)
		}
		if err = validateTeamChain(s, team.ID); err != nil {
			return fail(err)
		}
		node := dependentNode{assignment: assignment, content: content}
		key := bindingKey{assignment.GrantID, team.ID}
		holdings[key] = node
		if content.ParentGrantID != "" && team.ParentID != "" {
			parent := bindingKey{content.ParentGrantID, team.ParentID}
			children[parent] = append(children[parent], node)
		}
	}

	selected, ok := holdings[bindingKey{s.Assignments[assignmentID].GrantID, s.Assignments[assignmentID].Recipient.ID}]
	if !ok || selected.assignment.ID != assignmentID {
		return fail(domain.ErrRejected)
	}
	visitedGrants := map[string]bool{}
	activeAssignments, activeGrants, activeTeams := map[string]bool{}, map[string]bool{}, map[string]bool{}
	node := selected
	for steps := 0; ; steps++ {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		if steps == maxChainSteps {
			return fail(domain.ErrUnavailable)
		}
		id, grantID, teamID := node.assignment.ID, node.assignment.GrantID, node.assignment.Recipient.ID
		if activeAssignments[id] || activeGrants[grantID] || activeTeams[teamID] {
			return fail(domain.ErrRejected)
		}
		activeAssignments[id], activeGrants[grantID], activeTeams[teamID] = true, true, true
		visitedGrants[grantID] = true
		team := s.Teams[teamID]
		if node.content.ParentGrantID == "" || team.ParentID == "" {
			break
		}
		parent, exists := holdings[bindingKey{node.content.ParentGrantID, team.ParentID}]
		if !exists {
			break
		}
		node = parent
	}

	result := make([]domain.Assignment, 0)
	visited := map[string]bool{selected.assignment.ID: true}
	activeAssignments, activeGrants, activeTeams = map[string]bool{selected.assignment.ID: true}, map[string]bool{selected.assignment.GrantID: true}, map[string]bool{selected.assignment.Recipient.ID: true}
	var walk func(dependentNode, int) error
	walk = func(parent dependentNode, depth int) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if nonGroupParents[parent.assignment.GrantID] {
			return domain.ErrUnsupported
		}
		for _, child := range children[bindingKey{parent.assignment.GrantID, parent.assignment.Recipient.ID}] {
			if depth == maxChainSteps {
				return domain.ErrUnavailable
			}
			a, grantID, teamID := child.assignment, child.assignment.GrantID, child.assignment.Recipient.ID
			if activeAssignments[a.ID] || activeGrants[grantID] || activeTeams[teamID] {
				return domain.ErrRejected
			}
			if visited[a.ID] {
				continue
			}
			visited[a.ID], visitedGrants[grantID] = true, true
			activeAssignments[a.ID], activeGrants[grantID], activeTeams[teamID] = true, true, true
			result = append(result, a)
			if err := walk(child, depth+1); err != nil {
				return err
			}
			delete(activeAssignments, a.ID)
			delete(activeGrants, grantID)
			delete(activeTeams, teamID)
		}
		return nil
	}
	if err := walk(selected, 0); err != nil {
		return fail(err)
	}
	for grantID := range visitedGrants {
		if nonGroupParents[grantID] {
			return fail(domain.ErrUnsupported)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func validDependentAssignment(a domain.Assignment) bool {
	return a.Version == "1" && strings.TrimSpace(a.ID) != "" && strings.TrimSpace(a.GrantID) != "" && a.GrantRevision > 0 &&
		(a.Recipient.Type == "group" || a.Recipient.Type == "user") && strings.TrimSpace(a.Recipient.ID) != "" &&
		(a.Status == "enabled" || a.Status == "disabled")
}
