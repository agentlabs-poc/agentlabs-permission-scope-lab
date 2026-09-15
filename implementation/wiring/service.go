package wiring

import (
	"agentlabs.local/abv"
	"agentlabs.local/registry"
	regdomain "agentlabs.local/registry/domain"
	"context"
	"errors"
	"time"
)

// Clock is the one seam both domains take, named once here so a deployment
// supplies it once.
type Clock interface{ Now() time.Time }

// Config is what a deployment must decide. Everything else about the assembly —
// which port satisfies which interface, and in which order the stores open — is
// this package's business and not a caller's.
type Config struct {
	// AuthorityPath and RegistryPath are the two stores. They are separate
	// files because they are separate domains; nothing joins them in SQL.
	AuthorityPath string
	RegistryPath  string
	// CreateRegistry allows the registry store to be created when absent. The
	// authority store is never created here: a store that appears by accident is
	// an empty tenant, and an empty tenant answers every question with "no".
	CreateRegistry bool
	// Administration answers who may do what, in each domain. Auth-AL's gates
	// are discovered on this value, so one type may satisfy several of them.
	Administration         abv.Administration
	RegistryAdministration registry.Administration
	// Operator is the identity Auth-AL's questions reach the registry as. It is
	// not a tenant's identity: the question "does this tenant hold this
	// application" is asked by the service, not by whoever prompted it.
	Operator regdomain.Identity
	Clock    Clock
}

// Service is the assembled Auth service: both L1 domains, and the port between
// them.
//
// It exists because the assembly had no home. The rule was documented here and
// the steps were carried out twice — once in a test, once in a command — where
// they could drift apart without anything noticing. Now there is one of them.
type Service struct {
	authority    *abv.Facade
	applications *registry.Facade
}

// Open assembles the service. The order is not arbitrary: the registry opens
// first because Auth-AL cannot open without a port into it, which is the
// dependency direction made concrete.
func Open(ctx context.Context, cfg Config) (*Service, error) {
	if ctx == nil {
		return nil, errors.New("context is required")
	}
	if cfg.AuthorityPath == "" || cfg.RegistryPath == "" {
		return nil, errors.New("both store paths are required")
	}
	if cfg.Clock == nil || cfg.Administration == nil || cfg.RegistryAdministration == nil {
		return nil, errors.New("administration for both domains and a clock are required")
	}
	applications, err := registry.Open(ctx, cfg.RegistryPath, cfg.RegistryAdministration, cfg.Clock, cfg.CreateRegistry)
	if err != nil {
		return nil, err
	}
	// The port is the whole mechanism. Auth-AL declares the interface, the
	// registry provides a type with those methods, and this is the only package
	// that knows both exist — neither imports the other.
	port, err := registry.NewPort(applications, cfg.Operator)
	if err != nil {
		_ = applications.Close()
		return nil, err
	}
	var _ abv.Registry = port
	authority, err := abv.OpenSQLite(ctx, cfg.AuthorityPath, cfg.Administration, cfg.Clock, port)
	if err != nil {
		_ = applications.Close()
		return nil, err
	}
	return &Service{authority: authority, applications: applications}, nil
}

// Authority is Auth-AL: the records, the operations that change them, and the
// one read that says what a human is entitled to.
func (s *Service) Authority() *abv.Facade { return s.authority }

// Applications is the registry: which applications exist, and which tenants
// hold them.
func (s *Service) Applications() *registry.Facade { return s.applications }

// Close closes both stores, and reports the first failure rather than the last.
// Closing continues past an error: leaving the second store open because the
// first failed would leak a handle for no gain.
func (s *Service) Close() error {
	authorityErr := s.authority.Close()
	registryErr := s.applications.Close()
	if authorityErr != nil {
		return authorityErr
	}
	return registryErr
}
