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

func (f *Facade) CreateTeam(ctx context.Context, area domain.Area, identity domain.Identity, name, parentID string) (domain.Team, error) {
	return f.service.CreateTeam(ctx, area, identity, name, parentID)
}

func (f *Facade) SetTeamParent(ctx context.Context, area domain.Area, identity domain.Identity, id, parentID string) (domain.Team, error) {
	return f.service.SetTeamParent(ctx, area, identity, id, parentID)
}

func (f *Facade) DeleteTeam(ctx context.Context, area domain.Area, identity domain.Identity, id string) error {
	return f.service.DeleteTeam(ctx, area, identity, id)
}

func (f *Facade) AddMember(ctx context.Context, area domain.Area, identity domain.Identity, teamID, humanID string) error {
	return f.service.AddMember(ctx, area, identity, teamID, humanID)
}

func (f *Facade) RemoveMember(ctx context.Context, area domain.Area, identity domain.Identity, teamID, humanID string) error {
	return f.service.RemoveMember(ctx, area, identity, teamID, humanID)
}
