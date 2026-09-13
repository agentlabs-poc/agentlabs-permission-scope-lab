package abv

import (
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

// TeamReadAdministration gates reads of a tenant's teams and memberships.
type TeamReadAdministration interface {
	CheckTeamRead(context.Context, domain.Area, domain.Identity, time.Time) error
}

func (f *Facade) GetTeam(ctx context.Context, area domain.Area, identity domain.Identity, id string) (domain.Team, error) {
	return f.service.GetTeam(ctx, area, identity, id)
}

func (f *Facade) ListTeams(ctx context.Context, area domain.Area, identity domain.Identity, filter domain.TeamFilter) (domain.TeamPage, error) {
	return f.service.ListTeams(ctx, area, identity, filter)
}

func (f *Facade) ListMembers(ctx context.Context, area domain.Area, identity domain.Identity, filter domain.MemberFilter) (domain.MemberPage, error) {
	return f.service.ListMembers(ctx, area, identity, filter)
}
