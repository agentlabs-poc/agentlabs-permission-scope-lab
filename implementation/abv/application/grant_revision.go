package application

import (
	"agentlabs.local/abv/domain"
	"context"
)

type GrantRevisionAPI interface {
	PublishGrantRevision(context.Context, domain.Area, domain.FixtureContext, string, []byte) (domain.GrantContent, error)
}
