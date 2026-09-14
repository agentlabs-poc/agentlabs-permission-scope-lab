package registry

import (
	"agentlabs.local/registry/domain"
	"context"
)

// Port adapts this domain to the two-question interface Auth-AL declares.
//
// It lives here rather than in Auth-AL deliberately: Go interfaces are
// structural, so Auth-AL declares the shape it needs and this type satisfies it
// without either module importing the other. Whoever composes the two — a
// command, a test — imports both; neither domain imports its neighbour.
//
// The identity is the port's own, not the caller's: Auth-AL's identity means a
// human acting in Auth-AL, and passing it through would make the registry
// enforce Auth-AL's authorities. The port acts as a reader with no authority of
// its own, which is all these two questions need.
type Port struct {
	facade   *Facade
	identity domain.Identity
}

func NewPort(f *Facade, identity domain.Identity) (*Port, error) {
	if f == nil || !identity.Valid() {
		return nil, domain.ErrMalformed
	}
	return &Port{facade: f, identity: identity}, nil
}

// ApplicationExists reports whether the application is registered AND active.
//
// A suspended application is reported as absent, which is the mapping this
// adapter chooses: Auth-AL's question is whether it may accept catalog writes
// here, and it may not for an application the platform has suspended. That
// policy belongs here, at the seam, rather than in either domain.
func (p *Port) ApplicationExists(ctx context.Context, applicationID string) (bool, error) {
	app, err := p.facade.GetApplication(ctx, p.identity, applicationID)
	if errorIs(err, domain.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return app.Status == domain.StatusActive, nil
}

// Installed reports whether the tenant holds the application.
//
// One bit, from a domain that may one day have several installation states. The
// narrowing is deliberate and it is this adapter's decision: Auth-AL asks only
// whether the tenant may hold authority here.
func (p *Port) Installed(ctx context.Context, tenantID, applicationID string) (bool, error) {
	return p.facade.IsInstalled(ctx, p.identity, tenantID, applicationID)
}

func errorIs(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
