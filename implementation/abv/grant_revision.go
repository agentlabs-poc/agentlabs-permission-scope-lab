package abv

import (
	"agentlabs.local/abv/domain"
	"context"
	"time"
)

type GrantRevisionAdministration interface {
	CheckGrantRevisionPublication(context.Context, Evidence, domain.Identity, string, domain.GrantContent, time.Time) error
}

func (f *Facade) PublishGrantRevision(ctx context.Context, area domain.Area, identity domain.Identity, sourceAssignmentID string, proposed domain.GrantContent) (domain.GrantContent, error) {
	return f.service.PublishGrantRevision(ctx, area, identity, sourceAssignmentID, proposed)
}
