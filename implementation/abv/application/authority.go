package application

import (
	"agentlabs.local/abv/domain"
	"context"
)

// AuthorityAPI is the enforcement side's read, and the only surface on this
// boundary that is not administration. Everything else here changes authority or
// reads a record; this one answers what a human is entitled to.
type AuthorityAPI interface {
	ResolveAuthority(context.Context, domain.Area, domain.FixtureContext, domain.Identity, domain.ResolveOptions) (domain.ResolvedAuthority, error)
}
