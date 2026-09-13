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

// TeamAPI reads a tenant's teams and memberships.
type TeamAPI interface {
	GetTeam(context.Context, domain.Area, domain.FixtureContext, string) (domain.Team, error)
	ListTeams(context.Context, domain.Area, domain.FixtureContext, domain.TeamFilter) (domain.TeamPage, error)
	ListMembers(context.Context, domain.Area, domain.FixtureContext, domain.MemberFilter) (domain.MemberPage, error)
}
