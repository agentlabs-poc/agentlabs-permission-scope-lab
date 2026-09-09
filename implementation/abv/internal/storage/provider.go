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
	NewGrantRevision       *domain.GrantContent
	GrantStatusChange      *GrantStatusChange
	AssignmentStatusChange *AssignmentStatusChange
}

type Provider interface {
	Read(context.Context, domain.Area, func(Snapshot) error) error
	Update(context.Context, domain.Area, func(Snapshot) (WriteSet, error)) error
	Close() error
}
