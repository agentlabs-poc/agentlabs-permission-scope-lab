package lab

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

const grantRevisionFixtureContext = "maya-grant-publisher"

type GrantRevisionAdministration struct {
	*RoleAdministration
}

func (a *GrantRevisionAdministration) CheckGrantRevisionPublication(ctx context.Context, snapshot abv.Evidence, identity domain.Identity, sourceAssignmentID string, proposed domain.GrantContent, _ time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if snapshot.Area != a.area || identity != (domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}) || sourceAssignmentID != "fm5b7t4p5iv8" || proposed.GrantID != "fk3x9r2man0d" {
		return domain.ErrRejected
	}
	for _, membership := range snapshot.Memberships {
		if membership == (domain.Membership{TeamID: "fibggi2jv0n4", HumanID: "fi7io4lvjqio"}) {
			return nil
		}
	}
	return domain.ErrRejected
}
