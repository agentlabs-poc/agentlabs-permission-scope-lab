package validation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
)

// TeamScopeKey is the platform scope key that bounds administrative authority to
// the team it administers — Q-156.
const TeamScopeKey = "team"

// CheckPlatformScopeValues validates the scope values that name Auth's own
// records, which is the one place a scope value is not opaque.
//
// Q-148 settled that a scope value is opaque: nothing resolves `dept=FIN` to a
// record, because doing so needs application facts this model does not hold. A
// team is not an application fact — it is Auth's own record, in Auth's own store
// — so Q-156 validates it. A grant scoped to a team that does not exist is
// refused rather than stored as a boundary nothing can satisfy.
//
// This is deliberately separate from CheckContent. Content validation is about
// shape and the catalog; this is a referential check against the tenant's own
// relationship records, so it belongs where those records are in hand.
func CheckPlatformScopeValues(catalog domain.Catalog, teams map[string]domain.Team, g domain.GrantContent) error {
	for key, value := range g.Scope {
		registered, ok := catalog.Scopes[key]
		if !ok || registered.Boundary != domain.PlatformBoundary {
			continue
		}
		if key != TeamScopeKey {
			// A platform key this function does not know how to resolve is
			// refused rather than waved through. Adding a key is adding a rule
			// about what its value names, and silence would be the wrong default.
			return domain.ErrUnsupported
		}
		// $self is the evaluator's, not a team id, and it is already admitted by
		// content validation. It names no record here.
		if value == domain.SelfToken {
			return domain.ErrRejected
		}
		if !codec.ValidRoleID(value) {
			return domain.ErrMalformed
		}
		team, exists := teams[value]
		if !exists || team.ID != value {
			return domain.ErrRejected
		}
	}
	return nil
}
