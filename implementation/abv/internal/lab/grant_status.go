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
	if snapshot.Area != a.area || identity.HumanID != "fi7io4lvjqio" || identity.Actor != (domain.Actor{Type: "user", ID: "fi7io4lvjqio"}) || proposed.ID != "fk3x9r2man0d" || proposed.Version != "1" || (proposed.Status != "enabled" && proposed.Status != "disabled") {
		return domain.ErrRejected
	}
	for _, membership := range snapshot.Memberships {
		if membership == (domain.Membership{TeamID: "fibggi2jv0n4", HumanID: "fi7io4lvjqio"}) {
			return nil
		}
	}
	return domain.ErrRejected
}
