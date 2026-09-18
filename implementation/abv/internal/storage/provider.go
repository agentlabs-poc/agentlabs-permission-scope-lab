// Package storage defines the SQL-free persistence boundary used by ABV.
package storage

import (
	"agentlabs.local/abv/domain"
	"context"
)

type Snapshot struct {
	Area         domain.Area
	Catalog      domain.Catalog
	Controls     map[string]domain.GrantControl
	Contents     map[domain.GrantKey]domain.GrantContent
	Assignments  map[string]domain.Assignment
	Roles        map[domain.RoleKey]domain.RoleContent
	Teams        map[string]domain.Team
	Memberships  []domain.Membership
	Ownerships   []domain.Ownership
	TrustedRoots map[string]bool
	// Administrative carries the tenant's Auth-namespace authority, read in the
	// same transaction as everything above it.
	//
	// Administrative authority is an ordinary grant on the Auth root chain
	// (Q-155 / ADMIN-007), and that chain lives in the platform namespace — a
	// different area from the operation being gated. Q-151's namespace slice is
	// what makes it a different area rather than a convention: an application
	// root's ceiling is its own namespace, so no chain in an application's area
	// can ever carry an `auth:` permission.
	//
	// Nil means the snapshot is *already* the administrative one, which is the
	// case when the operation is performed in the platform namespace itself. It
	// is populated only by UpdateAdministered, because reading a second area for
	// every business resolve would double the cost of the hot path for a chain
	// that only administrative writes consult.
	Administrative *Snapshot
}

type GrantStatusChange struct {
	Before domain.GrantControl
	After  domain.GrantControl
}

type AssignmentStatusChange struct {
	Before domain.Assignment
	After  domain.Assignment
}

type WriteSet struct {
	NewAssignments  []domain.Assignment
	NewRoleRevision *domain.RoleContent
	// NewTeam creates a team; TeamParent re-parents one; RemovedTeam deletes
	// one. AddedMembership and RemovedMembership move one human in or out.
	// Exactly one is set per write.
	NewTeam           *domain.Team
	TeamParent        *domain.Team
	RemovedTeam       string
	AddedMembership   *domain.Membership
	RemovedMembership *domain.Membership
	// AddedOwnership and RemovedOwnership move one human in or out of a team's
	// owners. Separate from membership because Q-099 makes them separate facts.
	AddedOwnership   *domain.Ownership
	RemovedOwnership *domain.Ownership
	// NewGrant creates a grant whole: its head and its first revision, in one
	// transaction. Revision 1 cannot go through NewGrantRevision, which amends
	// an existing grant and requires a predecessor. RemovedGrant destroys one
	// whole, head and every revision — add-only's mirror.
	// RemovedAssignment deletes one by id. AssignmentRevisionChange is Q-105's
	// explicit adoption: the same binding, a different adopted revision.
	RemovedAssignment        string
	AssignmentRevisionChange *AssignmentStatusChange
	// NewRoot is establishment: a head, its first revision and the holder's
	// assignment, written together or not at all.
	NewRoot                *NewRoot
	NewGrant               *NewGrant
	RemovedGrant           string
	NewGrantRevision       *domain.GrantContent
	GrantStatusChange      *GrantStatusChange
	AssignmentStatusChange *AssignmentStatusChange
}

// NewRoot is everything a root is: the head carrying its trust evidence, the
// content computed coverage will be read against, and the assignment that makes
// it reachable. Q-117 requires that incomplete setup supply no authority, so
// these are one write or none.
type NewRoot struct {
	Grant        domain.Grant
	Content      domain.GrantContent
	Assignment   domain.Assignment
	HolderTeamID string
}

// NewGrant is a grant's two records, written together or not at all.
type NewGrant struct {
	Grant   domain.Grant
	Content domain.GrantContent
}

type Provider interface {
	Read(context.Context, domain.Area, func(Snapshot) error) error
	Update(context.Context, domain.Area, func(Snapshot) (WriteSet, error)) error
	// UpdateAdministered is Update with Snapshot.Administrative populated. An
	// administrative operation resolves the acting human's authority against the
	// tenant's Auth chain, which is a different area; this reads both inside one
	// transaction so the authority cannot change between the check and the write.
	UpdateAdministered(context.Context, domain.Area, func(Snapshot) (WriteSet, error)) error
	Close() error
}
