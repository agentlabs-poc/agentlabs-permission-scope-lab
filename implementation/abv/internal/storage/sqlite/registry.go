package sqlite

import "context"

// Registry is the storage layer's view of the application registry port. It
// mirrors abv.Registry rather than importing it: storage sits below the facade
// and must not depend on it.
//
// Whether an application exists and whether a tenant holds it are the
// application registry domain's facts, not Auth-AL's. Auth-AL asks; it does not
// store the answer.
type Registry interface {
	ApplicationExists(ctx context.Context, applicationID string) (bool, error)
	Installed(ctx context.Context, tenantID, applicationID string) (bool, error)
}
