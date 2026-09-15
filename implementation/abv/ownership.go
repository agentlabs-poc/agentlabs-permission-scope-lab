package abv

import (
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

// OwnershipAdministration gates ownership writes and reads.
//
// It is separate from TeamAdministration because the handbook has not decided
// whether team-write authority carries ownership transfer:
// "It does not decide whether auth:group::write authorizes ownership transfer;
// that operation's exact permission remains open." An adapter may answer it the
// same way as a team write or differently, and settling that later changes
// nothing here.
type OwnershipAdministration interface {
	CheckOwnershipWrite(context.Context, domain.Area, domain.Identity, domain.Ownership, time.Time) error
	CheckOwnershipRead(context.Context, domain.Area, domain.Identity, time.Time) error
}

// AddOwner gives a human authority to administer a team. Q-099: it gives them
// none of the team's business authority, does not make them a member, and does
// not let them assign grants.
func (f *Facade) AddOwner(ctx context.Context, area domain.Area, identity domain.Identity, teamID, humanID string) error {
	return f.service.AddOwner(ctx, area, identity, teamID, humanID)
}

// RemoveOwner withdraws it, and changes nothing else — not the team's grants,
// assignments, adopted revisions, parent links, or other memberships.
func (f *Facade) RemoveOwner(ctx context.Context, area domain.Area, identity domain.Identity, teamID, humanID string) error {
	return f.service.RemoveOwner(ctx, area, identity, teamID, humanID)
}

// ListOwners answers a team's owners or a human's teams, and requires exactly
// one of the two.
func (f *Facade) ListOwners(ctx context.Context, area domain.Area, identity domain.Identity, filter domain.OwnerFilter) (domain.OwnerPage, error) {
	return f.service.ListOwners(ctx, area, identity, filter)
}
