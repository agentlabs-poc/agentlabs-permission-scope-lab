package abv

import (
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

// GrantAdministration gates bringing a grant into existence and ending one.
// They are separate authorities from publishing a revision, which has its own
// check: amending a grant a tenant already holds is not the same act as creating
// a new one, or destroying one.
//
// Establishing a root is deliberately not here. Q-113 requires that ordinary
// grant administration can never confer root authority, and the guarantee is
// that no grant operation writes trust evidence — not a check inside one.
type GrantAdministration interface {
	CheckGrantCreate(context.Context, domain.Area, domain.Identity, domain.GrantContent, time.Time) error
	CheckGrantDelete(context.Context, domain.Area, domain.Identity, string, time.Time) error
}

// GrantReadAdministration gates reads of an area's grants.
type GrantReadAdministration interface {
	CheckGrantRead(context.Context, domain.Area, domain.Identity, time.Time) error
}

// CreateGrant writes a grant's head and its first revision together. The
// identifier is issued by Auth-AL; the caller supplies the parent and content.
func (f *Facade) CreateGrant(ctx context.Context, area domain.Area, identity domain.Identity, parentGrantID string, proposed domain.GrantContent) (domain.Grant, domain.GrantContent, error) {
	return f.service.CreateGrant(ctx, area, identity, parentGrantID, proposed)
}

// DeleteGrant removes a grant and every revision it owns, refusing while
// anything depends on it.
func (f *Facade) DeleteGrant(ctx context.Context, area domain.Area, identity domain.Identity, id string) error {
	return f.service.DeleteGrant(ctx, area, identity, id)
}

// GetGrant returns the head, plus one revision when revision is non-zero.
func (f *Facade) GetGrant(ctx context.Context, area domain.Area, identity domain.Identity, id string, revision int64) (domain.Grant, domain.GrantContent, error) {
	return f.service.GetGrant(ctx, area, identity, id, revision)
}

func (f *Facade) ListGrants(ctx context.Context, area domain.Area, identity domain.Identity, filter domain.GrantFilter) (domain.GrantPage, error) {
	return f.service.ListGrants(ctx, area, identity, filter)
}

func (f *Facade) ListGrantRevisions(ctx context.Context, area domain.Area, identity domain.Identity, id string, offset, limit int) (domain.GrantRevisionPage, error) {
	return f.service.ListGrantRevisions(ctx, area, identity, id, offset, limit)
}
