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

// The three team operations have no gate here any more, and their absence is the
// change Q-155 / ADMIN-007 made: creating, changing and deleting a team resolve
// the caller's own `auth:group::*` authority against the tenant's Auth chain,
// so there is nothing for a deployment to declare. What is left below is the
// gates that are still fixtures — every one of them is a rule adopted but not
// yet implemented, and the implementation is one family deep on purpose.

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

// CheckGrantDelete is grant-scoped, the way the status gate is.
//
// It discarded the id, which made it the same decision as the team gate the
// caller had already passed — so no in-tree administration could refuse the
// delete of one particular grant, and the ordering of the authorization gate
// against Q-132's dependency check was unobservable. Moving the dependency check
// above the gate left the whole suite green, and under any real administration
// that ordering is what stops a caller with no standing over a grant learning
// that it has a dependent.
func (a *RoleAdministration) CheckGrantDelete(ctx context.Context, area domain.Area, identity domain.Identity, id string, _ time.Time) error {
	if err := a.teamGate(ctx, area, identity); err != nil {
		return err
	}
	if id == unadministeredGrant {
		return domain.ErrRejected
	}
	return nil
}

// unadministeredGrant is a grant this fixture administrator may read and hold
// but never remove. A test names it to exercise a refusal that is about
// authorization rather than about the grant's dependents.
const unadministeredGrant = "fk3x9r2mzzzz"

// UnadministeredGrant is the id CheckGrantDelete refuses. Exported so a test can
// name it without restating a constant the gate owns.
func UnadministeredGrant() string { return unadministeredGrant }

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

// WorkloadClient is the lab's stand-in for an application's own credential.
//
// The Auth service already issues these: a workload client with an id and a
// secret, exchanged for a token the contract describes as *"bound to
// auth.registry.read and one tenant application"*. The lab models the shape and
// not the issuance — this repository never becomes that service, so a real
// credential belongs to the migration rather than here.
const WorkloadClient = "agent_hrms"

// WorkloadToken is the secret that credential authenticates with.
//
// It is deliberately not the id. An id names a credential and may be printed,
// logged and passed on a command line; a token authenticates it and may not.
// The lab issues neither, but it keeps them apart, because collapsing them is
// precisely the habit the migration would inherit.
const WorkloadToken = "shared_secret_fyd2k7x0q4nb"

// CheckAuthorityRead gates asking what a human is entitled to, and it is the one
// gate here that answers differently for different actors — because it is the
// only read whose caller need not be a person.
//
// For a service the question is a comparison rather than a policy: **is this
// credential bound to the area being asked about?** In the lab the binding is
// a.area, which is exactly what "bound to one tenant application" means. Nothing
// about the subject is checked, and that is the point — an application enforcing
// its own endpoints asks about many humans and is none of them.
//
// A human may still ask about their own authority, and only their own: the
// fixture gate holds a user actor to the fixture administrator.
func (a *RoleAdministration) CheckAuthorityRead(ctx context.Context, area domain.Area, identity domain.Identity, _ time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if area != a.area {
		return domain.ErrRejected
	}
	switch identity.Actor.Type {
	case "service_account":
		// The binding checked above is the area; this only refuses a credential
		// the lab has never heard of. The constant is named for the worked
		// fixture and is not itself the binding — a deployment's credential is
		// bound by what issued it, not by its spelling.
		if identity.Actor.ID != WorkloadClient {
			return domain.ErrRejected
		}
		return nil
	case "user":
		return a.teamGate(ctx, area, identity)
	case "agent":
		// Q-086 admits the type; this deployment has not implemented delegation
		// for it, which is "we do not do that" rather than "you may not".
		return domain.ErrUnsupported
	}
	return domain.ErrRejected
}
