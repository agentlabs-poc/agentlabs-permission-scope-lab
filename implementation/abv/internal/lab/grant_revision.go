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
	if snapshot.Area != a.area || identity != (domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "maya"}, HumanID: "maya"}) || sourceAssignmentID != "A1" || proposed.GrantID != "G2" {
		return domain.ErrRejected
	}
	for _, membership := range snapshot.Memberships {
		if membership == (domain.Membership{TeamID: "AssignmentAdmins", HumanID: "maya"}) {
			return nil
		}
	}
	return domain.ErrRejected
}
