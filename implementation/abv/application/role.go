package application

import (
	"agentlabs.local/abv/domain"
	"context"
)

type RoleAPI interface {
	PublishRole(context.Context, domain.Area, domain.FixtureContext, domain.RoleContent) (domain.RoleContent, error)
	GetRole(context.Context, domain.Area, domain.FixtureContext, string, int64) (domain.RoleContent, error)
	ListRoles(context.Context, domain.Area, domain.FixtureContext, domain.RoleFilter) (domain.RolePage, error)
	PublishApplicationRole(context.Context, domain.Application, domain.FixtureContext, domain.RoleContent) (domain.RoleContent, error)
}
