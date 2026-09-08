package lab

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

type AssignmentStatusAdministration struct{ *GrantStatusAdministration }

func NewAssignmentStatusAdministration(area domain.Area, premise AdministrationPremise) (*AssignmentStatusAdministration, error) {
	administration, err := NewGrantStatusAdministration(area, premise)
	if err != nil {
		return nil, err
	}
	return &AssignmentStatusAdministration{GrantStatusAdministration: administration}, nil
}

func (a *AssignmentStatusAdministration) CheckAssignmentStatus(ctx context.Context, snapshot abv.Evidence, identity domain.Identity, proposed domain.Assignment, _ time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if snapshot.Area != a.area || identity != (domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "maya"}, HumanID: "maya"}) || (proposed.Status != "enabled" && proposed.Status != "disabled") {
		return domain.ErrRejected
	}
	wanted := proposed.ID == "A1" && proposed.GrantID == "G1" && proposed.Recipient == (domain.Recipient{Type: "group", ID: "Team1"}) || proposed.ID == "A2" && proposed.GrantID == "G2" && proposed.Recipient == (domain.Recipient{Type: "group", ID: "Team2"})
	if !wanted {
		return domain.ErrRejected
	}
	current, ok := snapshot.Assignments[proposed.ID]
	if !ok {
		return domain.ErrRejected
	}
	current.Status = proposed.Status
	if current != proposed {
		return domain.ErrRejected
	}
	for _, membership := range snapshot.Memberships {
		if membership == (domain.Membership{TeamID: "AssignmentAdmins", HumanID: "maya"}) {
			return nil
		}
	}
	return domain.ErrRejected
}
