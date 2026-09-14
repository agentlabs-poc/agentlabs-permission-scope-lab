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

// GrantAPI is the grant record's own surface: creating a grant, reading one,
// listing them, and ending one. Publishing a revision and changing a status are
// older operations with their own entry points.
//
// Establishing a root is not here, and its absence is the contract: Q-113
// requires that grant administration cannot confer root authority, and nothing
// reachable through this interface writes trust evidence.
type GrantAPI interface {
	CreateGrant(context.Context, domain.Area, domain.FixtureContext, string, domain.GrantContent) (domain.Grant, domain.GrantContent, error)
	DeleteGrant(context.Context, domain.Area, domain.FixtureContext, string) error
	GetGrant(context.Context, domain.Area, domain.FixtureContext, string, int64) (domain.Grant, domain.GrantContent, error)
	ListGrants(context.Context, domain.Area, domain.FixtureContext, domain.GrantFilter) (domain.GrantPage, error)
	ListGrantRevisions(context.Context, domain.Area, domain.FixtureContext, string, int, int) (domain.GrantRevisionPage, error)
}

// AssignmentAPI is the assignment record's own surface. Creating one and
// changing its status are older operations with their own entry points.
type AssignmentAPI interface {
	GetAssignment(context.Context, domain.Area, domain.FixtureContext, string) (domain.Assignment, error)
	ListAssignments(context.Context, domain.Area, domain.FixtureContext, domain.AssignmentFilter) (domain.AssignmentPage, error)
	DeleteAssignment(context.Context, domain.Area, domain.FixtureContext, string) error
	UpgradeAssignment(context.Context, domain.Area, domain.FixtureContext, string) (domain.Assignment, error)
}

// TeamAPI reads a tenant's teams and memberships.
type TeamAPI interface {
	GetTeam(context.Context, domain.Area, domain.FixtureContext, string) (domain.Team, error)
	ListTeams(context.Context, domain.Area, domain.FixtureContext, domain.TeamFilter) (domain.TeamPage, error)
	ListMembers(context.Context, domain.Area, domain.FixtureContext, domain.MemberFilter) (domain.MemberPage, error)
	CreateTeam(context.Context, domain.Area, domain.FixtureContext, string, string) (domain.Team, error)
	SetTeamParent(context.Context, domain.Area, domain.FixtureContext, string, string) (domain.Team, error)
	DeleteTeam(context.Context, domain.Area, domain.FixtureContext, string) error
	AddMember(context.Context, domain.Area, domain.FixtureContext, string, string) error
	RemoveMember(context.Context, domain.Area, domain.FixtureContext, string, string) error
}
