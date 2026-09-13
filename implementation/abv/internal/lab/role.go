package lab

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

const roleFixtureContext = "maya-role-publisher"

type RoleAdministration struct {
	*AssignmentStatusAdministration
}

func (a *RoleAdministration) CheckRolePublication(ctx context.Context, snapshot abv.Evidence, identity domain.Identity, proposed domain.RoleContent, _ time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if snapshot.Area != a.area || identity != (domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "maya"}, HumanID: "maya"}) || proposed.ID != "fi9jvxobqsxs" {
		return domain.ErrRejected
	}
	for _, permission := range proposed.Permissions {
		if permission != PayslipRead && permission != PayslipWrite {
			return domain.ErrRejected
		}
	}
	for _, membership := range snapshot.Memberships {
		if membership == (domain.Membership{TeamID: "AssignmentAdmins", HumanID: "maya"}) {
			return nil
		}
	}
	return domain.ErrRejected
}

// CheckRoleRead gates the role reads. Reading a tenant's role catalog is a
// weaker act than publishing into it, so it admits the same fixture publisher
// without requiring the proposal checks publication makes.
func (a *RoleAdministration) CheckRoleRead(ctx context.Context, area domain.Area, identity domain.Identity, _ time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if area != a.area || identity != (domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "maya"}, HumanID: "maya"}) {
		return domain.ErrRejected
	}
	return nil
}
