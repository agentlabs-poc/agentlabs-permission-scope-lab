package abv

import (
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

type RoleAdministration interface {
	CheckRolePublication(context.Context, Evidence, domain.Identity, domain.RoleContent, time.Time) error
}

// RoleReadAdministration gates the two role reads, separately from publication,
// so an adapter can supply one without the other.
type RoleReadAdministration interface {
	CheckRoleRead(context.Context, domain.Area, domain.Identity, time.Time) error
}

func (f *Facade) PublishRole(ctx context.Context, area domain.Area, identity domain.Identity, proposed domain.RoleContent) (domain.RoleContent, error) {
	return f.service.PublishRole(ctx, area, identity, proposed)
}

func (f *Facade) GetRole(ctx context.Context, area domain.Area, identity domain.Identity, id string, revision int64) (domain.RoleContent, error) {
	return f.service.GetRole(ctx, area, identity, id, revision)
}

func (f *Facade) ListRoles(ctx context.Context, area domain.Area, identity domain.Identity, filter domain.RoleFilter) (domain.RolePage, error) {
	return f.service.ListRoles(ctx, area, identity, filter)
}
