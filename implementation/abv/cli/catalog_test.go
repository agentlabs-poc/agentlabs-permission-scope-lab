package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

type catalogSpy struct {
	app        domain.Application
	fixture    domain.FixtureContext
	permission domain.PermissionDefinition
	scope      domain.ScopeDefinition
	filter      domain.PermissionFilter
	scopeFilter domain.ScopeFilter
	page       domain.PermissionPage
	id         string
	namespace  string
	active     bool
	err        error
}

func (s *catalogSpy) GetScope(_ context.Context, app domain.Application, fixture domain.FixtureContext, key string) (domain.ScopeDefinition, error) {
	if s.err != nil {
		return domain.ScopeDefinition{}, s.err
	}
	s.app, s.fixture, s.id = app, fixture, key
	return domain.ScopeDefinition{Key: key}, nil
}

func (s *catalogSpy) ListScopes(_ context.Context, app domain.Application, fixture domain.FixtureContext, filter domain.ScopeFilter) (domain.ScopePage, error) {
	if s.err != nil {
		return domain.ScopePage{}, s.err
	}
	s.app, s.fixture, s.scopeFilter = app, fixture, filter
	return domain.ScopePage{}, nil
}

func (s *catalogSpy) GetPermission(_ context.Context, app domain.Application, fixture domain.FixtureContext, id string) (domain.PermissionDefinition, error) {
	if s.err != nil {
		return domain.PermissionDefinition{}, s.err
	}
	s.app, s.fixture, s.id = app, fixture, id
	return domain.PermissionDefinition{ID: id, Active: true}, nil
}

func (s *catalogSpy) ListPermissions(_ context.Context, app domain.Application, fixture domain.FixtureContext, filter domain.PermissionFilter) (domain.PermissionPage, error) {
	if s.err != nil {
		return domain.PermissionPage{}, s.err
	}
	s.app, s.fixture, s.filter = app, fixture, filter
	return s.page, nil
}

func (s *catalogSpy) SetPermissionStatus(_ context.Context, app domain.Application, fixture domain.FixtureContext, id string, active bool) (domain.PermissionDefinition, error) {
	if s.err != nil {
		return domain.PermissionDefinition{}, s.err
	}
	s.app, s.fixture, s.id, s.active = app, fixture, id, active
	return domain.PermissionDefinition{ID: id, Active: active}, nil
}

func (s *catalogSpy) RegisterPermission(_ context.Context, app domain.Application, fixture domain.FixtureContext, definition domain.PermissionDefinition) (domain.PermissionDefinition, error) {
	if s.err != nil {
		return domain.PermissionDefinition{}, s.err
	}
	s.app, s.fixture, s.permission = app, fixture, definition
	return definition, nil
}
func (s *catalogSpy) RegisterPlatformPermission(_ context.Context, namespace string, fixture domain.FixtureContext, definition domain.PermissionDefinition) (domain.PermissionDefinition, error) {
	s.fixture, s.namespace, s.permission = fixture, namespace, definition
	if s.err != nil {
		return domain.PermissionDefinition{}, s.err
	}
	return definition, nil
}

func (s *catalogSpy) RegisterScope(_ context.Context, app domain.Application, fixture domain.FixtureContext, definition domain.ScopeDefinition) (domain.ScopeDefinition, error) {
	if s.err != nil {
		return domain.ScopeDefinition{}, s.err
	}
	s.app, s.fixture, s.scope = app, fixture, definition
	return definition, nil
}

type catalogConnectorSpy struct {
	api           application.CatalogAPI
	closeErr      error
	calls, closes int
}

func (s *catalogConnectorSpy) connect(_ context.Context, _ domain.Application, _ string) (application.CatalogAPI, func() error, error) {
	s.calls++
	return s.api, func() error { s.closes++; return s.closeErr }, nil
}

func TestCatalogCommandsForwardValidatedDefinitions(t *testing.T) {
	spy := &catalogSpy{}
	connect := func(_ context.Context, app domain.Application, path string) (application.CatalogAPI, func() error, error) {
		if app.ID() != "hrms" || path != "lab.db" {
			t.Fatalf("connector app/path = %q/%q", app.ID(), path)
		}
		return spy, func() error { return nil }, nil
	}
	for _, args := range [][]string{
		{"catalog", "register-scope", "owner", "--app", "hrms", "--db", "lab.db", "--fixture-context", "application-publisher"},
		{"catalog", "register-permission", "hrms:payroll:payslip::export", "--app", "hrms", "--db", "lab.db", "--fixture-context", "application-publisher"},
	} {
		var out, diag bytes.Buffer
		if got := Run(t.Context(), args, strings.NewReader(""), &out, &diag, nil, nil, connect); got != 0 {
			t.Fatalf("%q exit=%d stderr=%q", args, got, diag.String())
		}
		if !strings.Contains(out.String(), "internal projection:") || !strings.Contains(diag.String(), "LAB ONLY") {
			t.Fatalf("stdout=%q stderr=%q", out.String(), diag.String())
		}
	}
	if spy.scope.Key != "owner" {
		t.Fatalf("scope=%+v", spy.scope)
	}
	if spy.permission.ID != "hrms:payroll:payslip::export" || !spy.permission.Active || spy.fixture.Name != "application-publisher" {
		t.Fatalf("permission=%+v fixture=%+v", spy.permission, spy.fixture)
	}
}

func TestCatalogCommandsRejectMalformedListsAndTenant(t *testing.T) {
	for _, args := range [][]string{
		{"catalog", "register-scope", "region", "--tenant", "acme", "--app", "hrms", "--db", "x", "--fixture-context", "application-publisher"},
	} {
		var out, diag bytes.Buffer
		if got := Run(t.Context(), args, strings.NewReader(""), &out, &diag, nil, nil); got != 2 || out.Len() != 0 {
			t.Fatalf("%q exit=%d stdout=%q stderr=%q", args, got, out.String(), diag.String())
		}
	}
}

func TestCatalogCapabilityAndFailures(t *testing.T) {
	args := []string{"catalog", "register-scope", "region", "--app", "hrms", "--db", "x", "--fixture-context", "application-publisher"}
	var out, diag bytes.Buffer
	if got := Run(t.Context(), args, strings.NewReader(""), &out, &diag, nil, nil); got != 5 {
		t.Fatalf("missing connector exit=%d", got)
	}

	typedNil := &catalogConnectorSpy{api: (*catalogSpy)(nil)}
	out.Reset()
	diag.Reset()
	if got := Run(t.Context(), args, strings.NewReader(""), &out, &diag, nil, nil, typedNil.connect); got != 4 || typedNil.closes != 1 {
		t.Fatalf("typed nil exit=%d closes=%d", got, typedNil.closes)
	}

	failed := &catalogConnectorSpy{api: &catalogSpy{err: domain.ErrRejected}}
	out.Reset()
	diag.Reset()
	if got := Run(t.Context(), args, strings.NewReader(""), &out, &diag, nil, nil, failed.connect); got != 3 || out.Len() != 0 || failed.closes != 1 {
		t.Fatalf("mutation failure exit=%d stdout=%q closes=%d", got, out.String(), failed.closes)
	}

	for _, tc := range []struct {
		out      io.Writer
		closeErr error
	}{{failingWriter{}, nil}, {&bytes.Buffer{}, errors.New("close")}} {
		connector := &catalogConnectorSpy{api: &catalogSpy{}, closeErr: tc.closeErr}
		diag.Reset()
		if got := Run(t.Context(), args, strings.NewReader(""), tc.out, &diag, nil, nil, connector.connect); got != 4 || connector.closes != 1 {
			t.Fatalf("I/O failure exit=%d closes=%d", got, connector.closes)
		}
	}

	connector := &catalogConnectorSpy{api: &catalogSpy{}}
	if got := Run(t.Context(), args, strings.NewReader(""), &out, &diag, nil, nil, connector.connect, connector.connect); got != 2 || connector.calls != 0 {
		t.Fatalf("multiple connectors exit=%d calls=%d", got, connector.calls)
	}
}
