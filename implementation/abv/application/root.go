package application

import (
	"agentlabs.local/abv/domain"
	"context"
)

// RootAPI is root establishment's own surface, and its separateness from
// GrantAPI is the contract rather than an organisational preference. Q-113
// requires that grant administration can never confer root authority; an adapter
// may implement GrantAPI and not this, and then no caller holding a grant
// vocabulary can start a lineage.
type RootAPI interface {
	// EstablishAuthRoot creates a tenant's authority over Auth itself. Its actor
	// is Auth platform administration, necessarily: a tenant has no
	// administrator until this runs.
	EstablishAuthRoot(context.Context, domain.Area, domain.FixtureContext, string) (domain.Grant, domain.GrantContent, error)
	// EstablishRoot creates a tenant's ceiling inside one application. Its actor
	// is the tenant administrator, holding Auth-boundary authority — the
	// authority to create an application's ceiling comes from outside that
	// application entirely.
	EstablishRoot(context.Context, domain.Area, domain.FixtureContext, string) (domain.Grant, domain.GrantContent, error)
}
