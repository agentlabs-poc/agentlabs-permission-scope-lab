package abv

import (
	"agentlabs.local/abv/domain"
	"context"
)

// ResolveAuthority answers what one human is entitled to inside one area: every
// grant that reaches them, folded, with the evidence of how.
//
// It is the only read on this facade that is not administration. Every other
// exported method changes authority or reads a record; this one is what an
// enforcing client consumes, and until it existed the resolution that powers it
// was reachable only from inside this module.
//
// What it returns is effective rather than stored — permissions expanded from
// any adopted role, scope and validity folded down the whole chain — because a
// client must never fold a chain itself. That is what keeps the root's namespace
// slice, selected-versus-inherited permissions and inherited-and-ANDed scope on
// this side of the boundary, free to change without an application changing.
func (f *Facade) ResolveAuthority(ctx context.Context, area domain.Area, identity domain.Identity, opts domain.ResolveOptions) (domain.ResolvedAuthority, error) {
	return f.service.ResolveAuthority(ctx, area, identity, opts)
}
