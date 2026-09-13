// Package mutation coordinates protected authority writes.
package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/storage"
	"context"
	"reflect"
	"time"
)

// Administration is the independent administrative-authority gate. The
// snapshot supplied to it is isolated from the evidence used by ABV.
type Administration interface {
	CheckAssignment(context.Context, storage.Snapshot, domain.Identity, domain.Assignment, time.Time) error
}

type GrantStatusAdministration interface {
	CheckGrantStatus(context.Context, storage.Snapshot, domain.Identity, domain.GrantControl, time.Time) error
}

type AssignmentStatusAdministration interface {
	CheckAssignmentStatus(context.Context, storage.Snapshot, domain.Identity, domain.Assignment, time.Time) error
}

type RoleAdministration interface {
	CheckRolePublication(context.Context, storage.Snapshot, domain.Identity, domain.RoleContent, time.Time) error
}

// RoleReadAdministration gates reads of a tenant's role catalog, the way
// PermissionAdministration does for permissions. It is separate from
// RoleAdministration so an adapter can supply publication without reads; a
// provider that does not implement it makes the reads ErrUnsupported rather than
// unprotected.
// ApplicationRoleAdministration gates publication of a role the application
// ships. It is separate from RoleAdministration because the authority differs:
// shipping a role is the platform acting, composing one is a tenant acting.
type ApplicationRoleAdministration interface {
	CheckApplicationRolePublication(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.RoleContent, time.Time) error
}

// PlatformAdministration gates registration in a platform namespace. It is
// separate from CatalogAdministration because the authority differs: defining
// platform vocabulary is the platform acting, and no application administrator
// may reach it.
type PlatformAdministration interface {
	CheckPlatformPermissionRegistration(context.Context, string, domain.Identity, domain.PermissionDefinition, time.Time) error
}

type RoleReadAdministration interface {
	CheckRoleRead(context.Context, domain.Area, domain.Identity, time.Time) error
}

// PermissionAdministration gates the read and status operations on an
// application's permission catalog. It is separate from CatalogAdministration so
// an adapter can supply registration without the rest; a provider that does not
// implement it makes those operations ErrUnsupported rather than unprotected.
type PermissionAdministration interface {
	CheckPermissionRead(context.Context, domain.Application, domain.Identity, time.Time) error
	CheckPermissionStatus(context.Context, domain.Application, domain.Catalog, domain.Identity, string, bool, time.Time) error
}

// ScopeAdministration gates reads of an application's scope catalog, the way
// PermissionAdministration does for permissions.
type ScopeAdministration interface {
	CheckScopeRead(context.Context, domain.Application, domain.Identity, time.Time) error
}

type CatalogAdministration interface {
	CheckPermissionRegistration(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.PermissionDefinition, time.Time) error
	CheckScopeRegistration(context.Context, domain.Application, domain.Catalog, domain.Identity, domain.ScopeDefinition, time.Time) error
}

type Clock interface{ Now() time.Time }

type Service struct {
	provider       storage.Provider
	administration Administration
	clock          Clock
	ids            IDs
}

// IDs issues record identifiers. It is a seam for the same reason Clock is: a
// generated id is not the caller's to choose, and a test needs it predictable.
type IDs interface {
	Next() string
}

func New(provider storage.Provider, administration Administration, clock Clock) (*Service, error) {
	ids, err := codec.NewSnowflakes(0, nil)
	if err != nil {
		return nil, err
	}
	return NewWithIDs(provider, administration, clock, ids)
}

// NewWithIDs builds a service with an explicit id source. Production supplies a
// generator whose node id comes from its own reserved block; New defaults to
// node 0, which is correct for a single-process lab and wrong for a fleet.
func NewWithIDs(provider storage.Provider, administration Administration, clock Clock, ids IDs) (*Service, error) {
	if nilInterface(provider) || nilInterface(administration) || nilInterface(clock) || nilInterface(ids) {
		return nil, domain.ErrMalformed
	}
	return &Service{provider: provider, administration: administration, clock: clock, ids: ids}, nil
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	kind := reflect.ValueOf(value).Kind()
	return (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface || kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice) && reflect.ValueOf(value).IsNil()
}
