package lab

import (
	"agentlabs.local/abv/domain"
	"context"
)

// FixedRegistry answers the two registry questions for one area and refuses
// every other, which is what a lab fixture needs now that Auth-AL keeps no copy
// of those facts.
//
// It is deliberately not the registry domain. The lab demonstrates Auth-AL, and
// composing the real registry here would make every scenario depend on a second
// store to say one thing the fixture already knows. `implementation/wiring` is
// where the two domains meet for real; this is the smallest thing that satisfies
// the port.
//
// It says yes only for its own area, so a scenario cannot accidentally resolve
// authority in a tenant or application it never seeded.
type FixedRegistry struct{ area domain.Area }

func NewFixedRegistry(area domain.Area) (*FixedRegistry, error) {
	if err := area.Validate(); err != nil {
		return nil, err
	}
	return &FixedRegistry{area: area}, nil
}

func (r *FixedRegistry) ApplicationExists(ctx context.Context, applicationID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return applicationID == r.area.ApplicationID(), nil
}

func (r *FixedRegistry) Installed(ctx context.Context, tenantID, applicationID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return tenantID == r.area.TenantID() && applicationID == r.area.ApplicationID(), nil
}
