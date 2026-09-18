package wiring

import (
	"agentlabs.local/abv"
	"agentlabs.local/registry"
	regdomain "agentlabs.local/registry/domain"
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
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
	// CreateRegistry and CreateAuthority allow each store to be created when
	// absent, and both default to refusing.
	//
	// The refusal is the point. abv.OpenSQLite creates a missing store
	// unconditionally, so a typo'd path used to succeed silently and leave an
	// empty database behind — and an empty tenant answers every question with
	// "no", which is a deny-everything service rather than an error an operator
	// can see. Administrative writes would land in the accidental store too.
	CreateRegistry  bool
	CreateAuthority bool
	// Administration answers who may do what, in each domain. Auth-AL's gates
	// are discovered on this value, so one type may satisfy several of them.
	Administration         abv.Administration
	RegistryAdministration registry.Administration
	// Operator is the identity Auth-AL's questions reach the registry as. It is
	// not a tenant's identity: the question "does this tenant hold this
	// application" is asked by the service, not by whoever prompted it.
	Operator regdomain.Identity
	Clock    Clock
	// PlatformNamespace names the namespace Auth's own permissions live in, and
	// with it the area a tenant's administrative chain is in — Q-155 / ADMIN-007.
	// Left empty, the administrative operations refuse rather than resolving
	// `auth:` authority against an application's own chain.
	PlatformNamespace string
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
	port         abv.Registry
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
	// A typed nil in an interface is not nil, and both domains hold their own
	// reflective check for exactly that reason. Catching it here means assembly
	// fails rather than the first request.
	if nilInterface(cfg.Clock) || nilInterface(cfg.Administration) || nilInterface(cfg.RegistryAdministration) {
		return nil, errors.New("administration for both domains and a clock are required")
	}
	if !cfg.CreateAuthority {
		if _, err := os.Stat(cfg.AuthorityPath); err != nil {
			return nil, fmt.Errorf("authority store %q: %w", cfg.AuthorityPath, err)
		}
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
	authority, err := abv.OpenSQLiteWithOptions(ctx, cfg.AuthorityPath, cfg.Administration, cfg.Clock, abv.Options{
		Registry: port, PlatformNamespace: cfg.PlatformNamespace,
	})
	if err != nil {
		_ = applications.Close()
		return nil, err
	}
	return &Service{authority: authority, applications: applications, port: port}, nil
}

// Registry is the port Auth-AL was opened against. It is exposed so the
// composed binary can demonstrate the seam it exists to demonstrate — asking the
// port is not the same as asking the registry facade, and the port owns the
// narrowing from an application record to one bit.
func (s *Service) Registry() abv.Registry { return s.port }

// nilInterface reports a typed nil held in an interface, which == nil does not.
func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
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
