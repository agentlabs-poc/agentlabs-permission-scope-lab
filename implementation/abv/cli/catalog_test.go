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
	keys       []string
	err        error
}

func (s *catalogSpy) RegisterPermission(_ context.Context, app domain.Application, fixture domain.FixtureContext, definition domain.PermissionDefinition, keys []string) (domain.PermissionDefinition, error) {
	if s.err != nil {
		return domain.PermissionDefinition{}, s.err
	}
	s.app, s.fixture, s.permission, s.keys = app, fixture, definition, append([]string(nil), keys...)
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
		{"catalog", "register-scope", "owner", "--allowed-tokens", "$self", "--app", "hrms", "--db", "lab.db", "--fixture-context", "application-publisher"},
		{"catalog", "register-permission", "hrms:payroll:payslip::export", "--supported-keys", "dept,region", "--app", "hrms", "--db", "lab.db", "--fixture-context", "application-publisher"},
	} {
		var out, diag bytes.Buffer
		if got := Run(t.Context(), args, strings.NewReader(""), &out, &diag, nil, nil, connect); got != 0 {
			t.Fatalf("%q exit=%d stderr=%q", args, got, diag.String())
		}
		if !strings.Contains(out.String(), "internal projection:") || !strings.Contains(diag.String(), "LAB ONLY") {
			t.Fatalf("stdout=%q stderr=%q", out.String(), diag.String())
		}
	}
	if spy.scope.Key != "owner" || len(spy.scope.AllowedTokens) != 1 || spy.scope.AllowedTokens[0] != "$self" {
		t.Fatalf("scope=%+v", spy.scope)
	}
	if spy.permission.ID != "hrms:payroll:payslip::export" || !spy.permission.Active || strings.Join(spy.keys, ",") != "dept,region" || spy.fixture.Name != "application-publisher" {
		t.Fatalf("permission=%+v keys=%v fixture=%+v", spy.permission, spy.keys, spy.fixture)
	}
}

func TestCatalogCommandsRejectMalformedListsAndTenant(t *testing.T) {
	for _, args := range [][]string{
		{"catalog", "register-scope", "owner", "--allowed-tokens", "$self,", "--app", "hrms", "--db", "x", "--fixture-context", "application-publisher"},
		{"catalog", "register-permission", "hrms:payroll:payslip::export", "--supported-keys", "dept, dept", "--app", "hrms", "--db", "x", "--fixture-context", "application-publisher"},
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
