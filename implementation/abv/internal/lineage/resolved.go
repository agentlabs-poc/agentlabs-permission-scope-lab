package lineage

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"slices"
	"time"
)

// ResolveAuthority answers what one human is entitled to inside one area: every
// grant that reaches them, folded, with the evidence of how.
//
// It is the enforcement side's only read. Everything it returns is effective
// rather than stored, because a client must never fold a chain itself — that is
// what keeps the root's namespace slice, selected-versus-inherited permissions
// and inherited-and-ANDed scope on this side of the boundary.
func ResolveAuthority(ctx context.Context, s storage.Snapshot, identity domain.Identity, opts domain.ResolveOptions, now time.Time) (domain.ResolvedAuthority, error) {
	fail := func(err error) (domain.ResolvedAuthority, error) {
		return domain.ResolvedAuthority{}, err
	}
	held, err := collectHumanRoutes(ctx, s, identity, opts.Permissions, now)
	if err != nil {
		return fail(err)
	}
	result := domain.ResolvedAuthority{Area: s.Area, HumanID: identity.HumanID, Grants: make([]domain.ResolvedGrant, 0, len(held))}
	for _, entry := range held {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		scope, satisfiable := foldScope(entry.route.Predicates)
		// An unsatisfiable route is dropped rather than shipped. Narrow appends
		// each child's predicates to its parent's, so a chain may state one key
		// twice with different values — dept=FIN above, dept=ENG below. Matching
		// requires every predicate to hold against one material value, so such a
		// route can never authorize anything.
		//
		// Folding it into a scope object without noticing would turn an
		// unmatchable route into a matchable one, which is why the fold reports
		// rather than overwrites.
		if !satisfiable {
			continue
		}
		leaf, ok := s.Assignments[entry.assignmentID]
		if !ok || leaf.ID != entry.assignmentID {
			return fail(domain.ErrRejected)
		}
		content, ok := s.Contents[domain.GrantKey{ID: leaf.GrantID, Revision: leaf.GrantRevision}]
		if !ok || content.GrantID != leaf.GrantID {
			return fail(domain.ErrRejected)
		}
		grant := domain.ResolvedGrant{
			Version: "1", GrantID: leaf.GrantID, Revision: leaf.GrantRevision,
			ParentGrantID: content.ParentGrantID,
			Permissions:   slices.Clone(entry.route.Permissions),
			Scope:         scope,
			Validity:      foldValidity(entry.route.Validities),
		}
		if !opts.OmitSource {
			source, err := buildSource(s, entry, content)
			if err != nil {
				return fail(err)
			}
			grant.Source = source
		}
		result.Grants = append(result.Grants, grant)
	}
	return result, nil
}

// foldScope turns the route's accumulated predicates into the effective scope,
// and reports whether they can hold at once. Two predicates on one key agree
// only when their values match; otherwise no material satisfies both.
func foldScope(predicates []domain.Predicate) (map[string]string, bool) {
	scope := make(map[string]string, len(predicates))
	for _, predicate := range predicates {
		if existing, seen := scope[predicate.Key]; seen && existing != predicate.Value {
			return nil, false
		}
		scope[predicate.Key] = predicate.Value
	}
	return scope, true
}

// foldValidity returns the narrowest window across the chain: the latest start
// and the earliest end. A route is in force only where every contributing grant
// is, so the intersection is the answer and nil means unbounded.
func foldValidity(validities []domain.Validity) *domain.Validity {
	var folded domain.Validity
	bounded := false
	for _, validity := range validities {
		if validity.NotBefore != nil && (folded.NotBefore == nil || folded.NotBefore.Before(*validity.NotBefore)) {
			value := *validity.NotBefore
			folded.NotBefore, bounded = &value, true
		}
		if validity.ExpiresAt != nil && (folded.ExpiresAt == nil || validity.ExpiresAt.Before(*folded.ExpiresAt)) {
			value := *validity.ExpiresAt
			folded.ExpiresAt, bounded = &value, true
		}
	}
	if !bounded {
		return nil
	}
	return &folded
}

// buildSource is the explanation: the assignment and team the human reaches the
// grant through, the role it adopted if any, and the chain back to the root.
//
// It is annotation and never a decision input. A gate matches permissions, scope
// and validity; an application that reasons over a lineage takes on every rule
// the lineage follows.
func buildSource(s storage.Snapshot, entry heldRoute, content domain.GrantContent) (domain.Source, error) {
	leaf := s.Assignments[entry.assignmentID]
	source := domain.Source{
		AssignmentID: entry.assignmentID,
		TeamID:       leaf.Recipient.ID,
		// Groups-only is deliberate, and enforced at both write and read: a
		// direct human assignment is refused rather than resolved.
		Via:     "membership",
		Lineage: make([]domain.LineageStep, 0, len(entry.route.AssignmentIDs)),
	}
	if content.RoleID != "" {
		source.AdoptedRole = &domain.AdoptedRole{RoleID: content.RoleID, Revision: content.RoleRevision}
	}
	// AssignmentIDs is ordered root-first: rootRoute seeds it and each Narrow
	// appends. The chain is therefore already the explanation, and until now it
	// was resolved and discarded.
	for _, id := range entry.route.AssignmentIDs {
		assignment, ok := s.Assignments[id]
		if !ok || assignment.ID != id {
			return domain.Source{}, domain.ErrRejected
		}
		source.Lineage = append(source.Lineage, domain.LineageStep{
			GrantID: assignment.GrantID, Revision: assignment.GrantRevision,
			AssignmentID: id, TeamID: assignment.Recipient.ID,
			Root: s.TrustedRoots[assignment.GrantID],
		})
	}
	return source, nil
}
