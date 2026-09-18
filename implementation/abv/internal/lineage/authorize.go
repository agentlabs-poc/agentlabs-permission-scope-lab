package lineage

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"slices"
	"time"
)

// Authorize answers one administrative question: does this identity hold a route
// carrying this permission whose predicates the named material satisfies?
//
// It is Q-155 / ADMIN-007's whole mechanism. Administrative authority is an
// ordinary grant on the Auth root chain, so establishing it is the walk the
// business side already uses — ResolveHuman — followed by the match the client
// side already uses. Nothing here is a second authority model; the twenty-eight
// administrative gates that compared an identifier to a constant collapse into
// this one call, and what an operation must state is only which permission it
// requires and which material bounds it.
//
// The snapshot is the administrative one. Administrative permissions live in the
// platform namespace, and Q-151's namespace slice means no application root's
// ceiling can carry them — so the chain being walked is always the tenant's Auth
// chain, never the application's.
//
// Failure to establish authority is a refusal; failure to *evaluate* is an error
// and is returned as one, which is Q-051 / DECISION-003. A caller must not read
// ErrUnavailable from a torn snapshot as "not authorized".
func Authorize(ctx context.Context, s storage.Snapshot, identity domain.Identity, permission string, material map[string]string, now time.Time) error {
	if permission == "" {
		return domain.ErrMalformed
	}
	for key, value := range material {
		if key == "" || value == "" {
			return domain.ErrMalformed
		}
	}
	routes, err := ResolveHuman(ctx, s, identity, permission, now)
	if err != nil {
		return err
	}
	for _, route := range routes {
		if routeAuthorizes(route, identity, permission, material) {
			return nil
		}
	}
	return domain.ErrRejected
}

// routeAuthorizes mirrors the client's routeMatches, and deliberately: a rule
// that answered differently at Auth than at an application would be a second
// enforcement rule, and the handbook has one.
//
// A request carries one value per key, so a route that accumulated two values of
// the same key is unsatisfiable — which is the whole of why administrative
// sideways escalation is impossible rather than guarded. Nothing here tests for
// that case; it falls out of matching every predicate against one material map.
func routeAuthorizes(route domain.Route, identity domain.Identity, permission string, material map[string]string) bool {
	if !slices.Contains(route.Permissions, permission) {
		return false
	}
	for _, predicate := range route.Predicates {
		value := predicate.Value
		if value == domain.SelfToken {
			// $self binds to the authorizing human, the same way the client binds
			// it. An administrative predicate rarely uses it; when it does, the
			// material still has to name the key.
			value = identity.HumanID
		}
		selected, ok := material[predicate.Key]
		if !ok || selected != value {
			return false
		}
	}
	return true
}
