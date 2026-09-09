package application

import (
	"agentlabs.local/abv/domain"
	"context"
)

type RoleAPI interface {
	PublishRole(context.Context, domain.Area, domain.FixtureContext, domain.RoleContent) (domain.RoleContent, error)
}
