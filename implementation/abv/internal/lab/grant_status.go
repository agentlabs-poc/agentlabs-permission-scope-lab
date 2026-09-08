package lab

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

type GrantStatusAdministration struct{ *Administration }

func NewGrantStatusAdministration(area domain.Area, premise AdministrationPremise) (*GrantStatusAdministration, error) {
	administration, err := NewAdministration(area, premise)
	if err != nil {
		return nil, err
	}
	return &GrantStatusAdministration{Administration: administration}, nil
}

func (a *GrantStatusAdministration) CheckGrantStatus(ctx context.Context, snapshot abv.Evidence, identity domain.Identity, proposed domain.GrantControl, _ time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if snapshot.Area != a.area || identity.HumanID != "maya" || identity.Actor != (domain.Actor{Type: "user", ID: "maya"}) || proposed.ID != "G2" || proposed.Version != "1" || (proposed.Status != "enabled" && proposed.Status != "disabled") {
		return domain.ErrRejected
	}
	for _, membership := range snapshot.Memberships {
		if membership == (domain.Membership{TeamID: "AssignmentAdmins", HumanID: "maya"}) {
			return nil
		}
	}
	return domain.ErrRejected
}
