package abv

import (
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

// AssignmentRecordAdministration gates reading an area's assignments, removing
// one, and adopting a newer grant revision into one.
//
// Adoption is its own authority rather than a status change: Q-105 makes an
// upgrade a fresh selection that must pass current checks, which is closer to
// creating an assignment than to enabling one.
type AssignmentRecordAdministration interface {
	CheckAssignmentRead(context.Context, domain.Area, domain.Identity, time.Time) error
	CheckAssignmentDelete(context.Context, domain.Area, domain.Identity, string, time.Time) error
	CheckAssignmentAdoption(context.Context, domain.Area, domain.Identity, domain.Assignment, time.Time) error
}

func (f *Facade) GetAssignment(ctx context.Context, area domain.Area, identity domain.Identity, id string) (domain.Assignment, error) {
	return f.service.GetAssignment(ctx, area, identity, id)
}

// ListAssignments answers a grant's recipients or a recipient's grants, and
// requires exactly one of the two.
func (f *Facade) ListAssignments(ctx context.Context, area domain.Area, identity domain.Identity, filter domain.AssignmentFilter) (domain.AssignmentPage, error) {
	return f.service.ListAssignments(ctx, area, identity, filter)
}

// DeleteAssignment removes a route permanently, refusing while a dependent
// route rests on it.
func (f *Facade) DeleteAssignment(ctx context.Context, area domain.Area, identity domain.Identity, id string) error {
	return f.service.DeleteAssignment(ctx, area, identity, id)
}

// UpgradeAssignment adopts the latest published revision of the assignment's
// grant — never an intermediate one, and never a fallback when the latest
// cannot be supported.
func (f *Facade) UpgradeAssignment(ctx context.Context, area domain.Area, identity domain.Identity, id string) (domain.Assignment, error) {
	return f.service.UpgradeAssignment(ctx, area, identity, id)
}
