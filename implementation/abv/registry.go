package abv

import "context"

// Registry answers the two questions Auth-AL cannot answer itself: whether an
// application exists, and whether a tenant holds it.
//
// It is a port, not a table. Auth-AL does not know or care what is behind it —
// the application registry domain, the legacy auth service, or both during a
// migration — and does not change when that changes. Nothing in this package
// imports the registry.
//
// Installed returns one bit because that is Auth-AL's question: may this tenant
// hold authority here. A richer registry has several installation states;
// mapping them to one bit is the adapter's decision and belongs there, where the
// policy it encodes is visible.
type Registry interface {
	ApplicationExists(ctx context.Context, applicationID string) (bool, error)
	Installed(ctx context.Context, tenantID, applicationID string) (bool, error)
}
