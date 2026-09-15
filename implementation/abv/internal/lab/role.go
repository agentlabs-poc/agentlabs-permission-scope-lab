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

// The grant operations. As with teams, the lab admits the same fixture
// administrator for all of them; a real deployment holds these as separate
// authorities, which is why they are separate methods.
//
// There is no root establishment here, and that is the point rather than an
// omission: Q-113 requires that ordinary grant administration cannot confer root
// authority, and no method on this interface writes trust evidence.
func (a *RoleAdministration) CheckGrantCreate(ctx context.Context, area domain.Area, identity domain.Identity, proposed domain.GrantContent, _ time.Time) error {
	if err := a.teamGate(ctx, area, identity); err != nil {
		return err
	}
	if proposed.ParentGrantID == "" {
		return domain.ErrRejected
	}
	return nil
}

func (a *RoleAdministration) CheckGrantDelete(ctx context.Context, area domain.Area, identity domain.Identity, _ string, _ time.Time) error {
	return a.teamGate(ctx, area, identity)
}

func (a *RoleAdministration) CheckGrantRead(ctx context.Context, area domain.Area, identity domain.Identity, _ time.Time) error {
	return a.teamGate(ctx, area, identity)
}

// The assignment record operations. As elsewhere, the lab admits the same
// fixture administrator for all of them.
func (a *RoleAdministration) CheckAssignmentRead(ctx context.Context, area domain.Area, identity domain.Identity, _ time.Time) error {
	return a.teamGate(ctx, area, identity)
}

func (a *RoleAdministration) CheckAssignmentDelete(ctx context.Context, area domain.Area, identity domain.Identity, _ string, _ time.Time) error {
	return a.teamGate(ctx, area, identity)
}

// Adoption is gated separately from deletion because Q-105 makes an upgrade a
// fresh selection that must pass current checks, not a lifecycle toggle.
func (a *RoleAdministration) CheckAssignmentAdoption(ctx context.Context, area domain.Area, identity domain.Identity, _ domain.Assignment, _ time.Time) error {
	return a.teamGate(ctx, area, identity)
}

// The ownership operations. The lab admits the same fixture administrator, and
// gates them separately from the team writes because the handbook has not
// decided whether team-write authority carries ownership transfer.
func (a *RoleAdministration) CheckOwnershipWrite(ctx context.Context, area domain.Area, identity domain.Identity, _ domain.Ownership, _ time.Time) error {
	return a.teamGate(ctx, area, identity)
}

func (a *RoleAdministration) CheckOwnershipRead(ctx context.Context, area domain.Area, identity domain.Identity, _ time.Time) error {
	return a.teamGate(ctx, area, identity)
}

// Root establishment. The lab admits the same fixture administrator for both,
// and they are separate methods because their actors are different in a real
// deployment: Auth platform administration for the Auth root, the tenant
// administrator for an application's.
//
// Neither names a permission, because the handbook has not chosen one.
func (a *RoleAdministration) CheckAuthRootEstablishment(ctx context.Context, area domain.Area, identity domain.Identity, _ string, _ time.Time) error {
	return a.teamGate(ctx, area, identity)
}

func (a *RoleAdministration) CheckRootEstablishment(ctx context.Context, area domain.Area, identity domain.Identity, _ string, _ time.Time) error {
	return a.teamGate(ctx, area, identity)
}

// CheckAuthorityRead gates asking what a human is entitled to. The lab admits
// the same fixture administrator as every other read; in a deployment the caller
// is the application enforcing its own endpoints, which is a different actor
// from a tenant administrator browsing records.
func (a *RoleAdministration) CheckAuthorityRead(ctx context.Context, area domain.Area, identity domain.Identity, _ time.Time) error {
	return a.teamGate(ctx, area, identity)
}
