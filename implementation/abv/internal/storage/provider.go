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
	TrustedRoots map[string]bool
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
	NewAssignments         []domain.Assignment
	NewRoleRevision        *domain.RoleContent
	// NewTeam creates a team; TeamParent re-parents one; RemovedTeam deletes
	// one. AddedMembership and RemovedMembership move one human in or out.
	// Exactly one is set per write.
	NewTeam           *domain.Team
	TeamParent        *domain.Team
	RemovedTeam       string
	AddedMembership   *domain.Membership
	RemovedMembership *domain.Membership
	// NewGrant creates a grant whole: its head and its first revision, in one
	// transaction. Revision 1 cannot go through NewGrantRevision, which amends
	// an existing grant and requires a predecessor. RemovedGrant destroys one
	// whole, head and every revision — add-only's mirror.
	// RemovedAssignment deletes one by id. AssignmentRevisionChange is Q-105's
	// explicit adoption: the same binding, a different adopted revision.
	RemovedAssignment        string
	AssignmentRevisionChange *AssignmentStatusChange
	NewGrant                 *NewGrant
	RemovedGrant           string
	NewGrantRevision       *domain.GrantContent
	GrantStatusChange      *GrantStatusChange
	AssignmentStatusChange *AssignmentStatusChange
}

// NewGrant is a grant's two records, written together or not at all.
type NewGrant struct {
	Grant   domain.Grant
	Content domain.GrantContent
}

type Provider interface {
	Read(context.Context, domain.Area, func(Snapshot) error) error
	Update(context.Context, domain.Area, func(Snapshot) (WriteSet, error)) error
	Close() error
}
