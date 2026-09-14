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
	// The gate cannot pin the id. An id is issued by the service, so it does not
	// exist when a new role is proposed — a policy that keyed on one could only
	// ever admit roles that already exist. It gates the name, the bundle and the
	// publisher's membership, which are the things a proposal actually carries.
	if snapshot.Area != a.area || identity != (domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}) {
		return domain.ErrRejected
	}
	if proposed.Name == "" {
		return domain.ErrRejected
	}
	for _, permission := range proposed.Permissions {
		if permission != PayslipRead && permission != PayslipWrite {
			return domain.ErrRejected
		}
	}
	for _, membership := range snapshot.Memberships {
		if membership == (domain.Membership{TeamID: "fibggi2jv0n4", HumanID: "fi7io4lvjqio"}) {
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
	if area != a.area || identity != (domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}) {
		return domain.ErrRejected
	}
	return nil
}

// CheckApplicationRolePublication gates a role the application ships. It is the
// catalog publisher acting, not a tenant administrator — the same identity that
// registers permissions and scope keys, because shipping a role is the same kind
// of act.
func (a *RoleAdministration) CheckApplicationRolePublication(ctx context.Context, app domain.Application, catalog domain.Catalog, identity domain.Identity, proposed domain.RoleContent, _ time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if app.ID() != a.area.ApplicationID() || catalog.ApplicationID != app.ID() {
		return domain.ErrRejected
	}
	if identity != (domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}) {
		return domain.ErrRejected
	}
	if proposed.Name == "" {
		return domain.ErrRejected
	}
	return nil
}

// CheckTeamRead gates the team and membership reads. Reading who is in a team is
// weaker than publishing into one, so it admits the same fixture publisher.
func (a *RoleAdministration) CheckTeamRead(ctx context.Context, area domain.Area, identity domain.Identity, _ time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if area != a.area || identity != (domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}) {
		return domain.ErrRejected
	}
	return nil
}

// The three team operations. The lab admits the same fixture administrator for
// all three; a real deployment would hold three separate authorities, which is
// why they are three methods rather than one.
func (a *RoleAdministration) CheckTeamCreate(ctx context.Context, area domain.Area, identity domain.Identity, proposed domain.Team, _ time.Time) error {
	if err := a.teamGate(ctx, area, identity); err != nil {
		return err
	}
	if proposed.Name == "" {
		return domain.ErrRejected
	}
	return nil
}

func (a *RoleAdministration) CheckTeamWrite(ctx context.Context, area domain.Area, identity domain.Identity, _ string, _ time.Time) error {
	return a.teamGate(ctx, area, identity)
}

func (a *RoleAdministration) CheckTeamDelete(ctx context.Context, area domain.Area, identity domain.Identity, _ string, _ time.Time) error {
	return a.teamGate(ctx, area, identity)
}

func (a *RoleAdministration) teamGate(ctx context.Context, area domain.Area, identity domain.Identity) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if area != a.area || identity != (domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}) {
		return domain.ErrRejected
	}
	return nil
}
