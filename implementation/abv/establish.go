package abv

import (
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

// RootEstablishment gates bringing a root into existence.
//
// It is deliberately not part of any grant administration interface. Q-113
// requires that ordinary grant operations can never confer root authority, and
// the guarantee is that no grant operation can reach these methods — not a check
// inside one. An administrator who may call every grant and assignment operation
// still cannot establish a root.
//
// Neither gate names a permission. Q-113 leaves the procedure open, so an
// adapter answers them and the handbook can settle the permission later without
// changing these signatures.
type RootEstablishment interface {
	CheckAuthRootEstablishment(context.Context, domain.Area, domain.Identity, string, time.Time) error
	CheckRootEstablishment(context.Context, domain.Area, domain.Identity, string, time.Time) error
}

// EstablishAuthRoot creates a tenant's authority over Auth itself: the first
// authority a tenant has, and what makes someone a tenant administrator.
//
// The actor is Auth platform administration, necessarily — a tenant has no
// administrator until this runs, so the authority cannot come from inside it.
func (f *Facade) EstablishAuthRoot(ctx context.Context, area domain.Area, identity domain.Identity, holderTeamID string) (domain.Grant, domain.GrantContent, error) {
	return f.service.EstablishAuthRoot(ctx, area, identity, holderTeamID)
}

// EstablishRoot creates a tenant's ceiling inside one application.
//
// The actor is the tenant administrator, holding Auth-boundary authority from
// the Auth root. That is what makes it non-circular: the authority to create an
// application's ceiling comes from outside that application entirely — and the
// namespace slice in rootRoute is what keeps it so, since without it the
// application's ceiling would contain the authority that made it.
func (f *Facade) EstablishRoot(ctx context.Context, area domain.Area, identity domain.Identity, holderTeamID string) (domain.Grant, domain.GrantContent, error) {
	return f.service.EstablishRoot(ctx, area, identity, holderTeamID)
}
