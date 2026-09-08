// Package application defines the reusable CLI's adapter boundary.
// Implementations must validate Area and establish trusted identity separately.
package application

import (
	"agentlabs.local/abv/domain"
	"context"
)

type API interface {
	Inspect(context.Context, domain.Area, string, string) (domain.Record, error)
	CheckAssignment(context.Context, domain.Area, []byte) (domain.Diagnostic, error)
	Assign(context.Context, domain.Area, domain.FixtureContext, []byte) (domain.Receipt, error)
}
type ScenarioRunner interface {
	Seed(context.Context, domain.Area, string, string) error
	Run(context.Context, domain.Area, string, string, string) error
}
