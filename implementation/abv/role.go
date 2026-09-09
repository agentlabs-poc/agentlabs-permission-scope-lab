package abv

import (
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

type RoleAdministration interface {
	CheckRolePublication(context.Context, Evidence, domain.Identity, domain.RoleContent, time.Time) error
}

func (f *Facade) PublishRole(ctx context.Context, area domain.Area, identity domain.Identity, proposed domain.RoleContent) (domain.RoleContent, error) {
	return f.service.PublishRole(ctx, area, identity, proposed)
}
