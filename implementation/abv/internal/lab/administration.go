package lab

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"time"
)

// Administration is a fixed, trusted lab premise. It is not an Auth resolver.
type Administration struct {
	area    domain.Area
	premise AdministrationPremise
}

func NewAdministration(area domain.Area, premise AdministrationPremise) (*Administration, error) {
	if err := area.Validate(); err != nil {
		return nil, err
	}
	if premise.HumanID != "maya" || premise.PermissionID != AssignmentCreate || premise.RecipientTeamID != "Team2" {
		return nil, domain.ErrRejected
	}
	return &Administration{area: area, premise: premise}, nil
}

func (a *Administration) CheckAssignment(ctx context.Context, snapshot storage.Snapshot, identity domain.Identity, proposed domain.Assignment, _ time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if snapshot.Area != a.area || identity.HumanID != a.premise.HumanID || identity.Actor.Type != "user" || identity.Actor.ID != identity.HumanID || proposed.Recipient.Type != "group" || proposed.Recipient.ID != a.premise.RecipientTeamID {
		return domain.ErrRejected
	}
	team, ok := snapshot.Teams["AssignmentAdmins"]
	if !ok || team.ID != "AssignmentAdmins" {
		return domain.ErrRejected
	}
	for _, membership := range snapshot.Memberships {
		if membership.TeamID == "AssignmentAdmins" && membership.HumanID == a.premise.HumanID {
			return nil
		}
	}
	return domain.ErrRejected
}
