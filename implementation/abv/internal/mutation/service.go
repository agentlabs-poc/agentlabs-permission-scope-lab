// Package mutation coordinates protected authority writes.
package mutation

import (
	"agentlabs.local/abv/domain"
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

type Clock interface{ Now() time.Time }

type Service struct {
	provider       storage.Provider
	administration Administration
	clock          Clock
}

func New(provider storage.Provider, administration Administration, clock Clock) (*Service, error) {
	if nilInterface(provider) || nilInterface(administration) || nilInterface(clock) {
		return nil, domain.ErrMalformed
	}
	return &Service{provider: provider, administration: administration, clock: clock}, nil
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	kind := reflect.ValueOf(value).Kind()
	return (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface || kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice) && reflect.ValueOf(value).IsNil()
}
