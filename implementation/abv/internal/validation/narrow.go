package validation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"slices"
	"sort"
)

// Narrow constructs restrictions from a supported parent and checked child.
// It resolves the complete direct or exactly adopted role permission source.
// It does not discover lineage or bind recipient-relative tokens like $self.
func Narrow(area domain.Area, parent domain.Route, child domain.GrantContent, roles map[domain.RoleKey]domain.RoleContent) (domain.Route, error) {
	fail := func(err error) (domain.Route, error) { return domain.Route{}, err }
	if err := area.Validate(); err != nil {
		return fail(err)
	}
	if err := parent.Area.Validate(); err != nil {
		return fail(err)
	}
	if parent.Area != area || parent.GrantID == "" || child.ParentGrantID != parent.GrantID {
		return fail(domain.ErrRejected)
	}
	if err := codec.ValidateContent(child); err != nil {
		return fail(err)
	}
	permissions, err := selectedPermissions(child, roles)
	if err != nil {
		return fail(err)
	}
	for _, p := range permissions {
		if !slices.Contains(parent.Permissions, p) {
			return fail(domain.ErrRejected)
		}
	}
	result := domain.Route{
		Area: area, GrantID: child.GrantID,
		Permissions: slices.Clone(permissions), Predicates: slices.Clone(parent.Predicates),
		AssignmentIDs: slices.Clone(parent.AssignmentIDs),
	}
	for _, v := range parent.Validities {
		result.Validities = append(result.Validities, cloneValidity(v))
	}
	if child.Validity != nil {
		result.Validities = append(result.Validities, cloneValidity(*child.Validity))
	}
	keys := make([]string, 0, len(child.Scope))
	for k := range child.Scope {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		result.Predicates = append(result.Predicates, domain.Predicate{Key: k, Value: child.Scope[k], SourceGrantID: child.GrantID})
	}
	return result, nil
}
func cloneValidity(v domain.Validity) domain.Validity {
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
