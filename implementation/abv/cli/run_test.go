package cli

import (
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type apiSpy struct {
	area     domain.Area
	kind, id string
	raw      []byte
	fixture  domain.FixtureContext
}

func (s *apiSpy) Inspect(_ context.Context, area domain.Area, kind, id string) (domain.Record, error) {
	s.area, s.kind, s.id = area, kind, id
	return domain.Record{Kind: kind, ID: id, Rows: [][]string{{"field", "value"}, {"id", id}}}, nil
}
func (s *apiSpy) CheckAssignment(_ context.Context, area domain.Area, raw []byte) (domain.Diagnostic, error) {
	s.area, s.raw = area, append([]byte(nil), raw...)
	return domain.Diagnostic{Summary: "valid proposal"}, nil
}
func (s *apiSpy) Assign(_ context.Context, area domain.Area, fixture domain.FixtureContext, raw []byte) (domain.Receipt, error) {
	s.area, s.fixture, s.raw = area, fixture, append([]byte(nil), raw...)
	return domain.Receipt{AssignmentID: "A2"}, nil
}

type connectorSpy struct {
	path          string
	area          domain.Area
	calls, closes int
	api           application.API
	err, closeErr error
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("pipe closed") }

func (s *connectorSpy) connect(_ context.Context, area domain.Area, path string) (application.API, func() error, error) {
	s.calls++
	s.area, s.path = area, path
	if s.err != nil {
		return nil, nil, s.err
	}
	return s.api, func() error { s.closes++; return s.closeErr }, nil
}

// Parsing must fail before any adapter call. Nil adapters turn an accidental dispatch into a failure.
func TestInvalidCommandNeverDispatches(t *testing.T) {
	cases := [][]string{
		{}, {"unknown"}, {"inspect", "grant", "G1", "--app", "hrms"},
		{"inspect", "grant", "G1", "--tenant", "acme"},
		{"inspect", "grant", "G1", "--tenant"},
		{"inspect", "grant", "G1", "--tenant", "*", "--app", "hrms"},
		{"inspect", "grant", "G1", "--tenant", "acme", "--app", "hrms", "--skip-abv"},
		{"inspect", "grant", "G1", "--tenant", "acme", "--tenant", "other", "--app", "hrms"},
		{"inspect", "unknown", "G1", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"inspect", "grant", "", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"inspect", "grant", "G1", "--db=", "--tenant", "acme", "--app", "hrms"},
		{"check", "assignment", "--file=", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"assign", "--file", "a.json", "--db", "", "--fixture-context", "maya", "--tenant", "acme", "--app", "hrms"},
		{"inspect", "grant", "G1", "--file", "a.json", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"check", "assignment", "--file", "a.json", "--db", "x", "--case", "bad", "--tenant", "acme", "--app", "hrms"},
		{"assign", "--file", "a.json", "--db", "x", "--fixture-context", "maya", "--case", "bad", "--tenant", "acme", "--app", "hrms"},
		{"scenario", "seed", "team-fin-c17", "--file", "a.json", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"scenario", "run", "team-fin-c17", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"scenario", "seed", "team-fin-c17", "--case", "bad", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"inspect", "grant", "G1", "--db", "x", "--db", "y", "--tenant", "acme", "--app", "hrms"},
	}
	for _, args := range cases {
		var out, diag bytes.Buffer
		if got := Run(context.Background(), args, strings.NewReader(""), &out, &diag, nil, nil); got != 2 {
			t.Fatalf("%q: exit %d, want malformed; %s", args, got, diag.String())
		}
		if out.Len() != 0 || diag.Len() == 0 {
			t.Fatalf("wrong output streams for %q", args)
		}
	}
}

func TestMissingContextMakesZeroConnectorAndAPICalls(t *testing.T) {
	api := &apiSpy{}
	connector := &connectorSpy{api: api}
	var out, diag bytes.Buffer
	got := Run(context.Background(), []string{"inspect", "grant", "G1", "--db", "x", "--app", "hrms"}, strings.NewReader(""), &out, &diag, connector.connect, nil)
	if got != 2 || connector.calls != 0 || api.kind != "" {
		t.Fatalf("exit=%d connector=%d api=%+v", got, connector.calls, api)
	}
}
func TestInspectForwardsExactAreaPathAndCloses(t *testing.T) {
	for _, kind := range []string{"permission", "scope", "role", "grant", "assignment", "team", "membership"} {
		api := &apiSpy{}
		connector := &connectorSpy{api: api}
		var out, diag bytes.Buffer
		args := []string{"inspect", kind, "G1", "--db", "relative.db", "--tenant", "acme", "--app", "hrms"}
		if got := Run(context.Background(), args, strings.NewReader(""), &out, &diag, connector.connect, nil); got != 0 {
			t.Fatalf("%s: exit %d: %s", kind, got, diag.String())
		}
		if connector.calls != 1 || connector.closes != 1 || connector.path != "relative.db" || connector.area.TenantID() != "acme" || connector.area.ApplicationID() != "hrms" {
			t.Fatalf("%s: wrong connector forwarding: %+v", kind, connector)
		}
		if api.kind != kind || api.id != "G1" || !strings.Contains(out.String(), "internal projection") {
			t.Fatalf("%s: wrong dispatch/output: %q", kind, out.String())
		}
	}
}

func TestConnectorAndCloseFailuresAreUnavailable(t *testing.T) {
	for _, connector := range []*connectorSpy{
		{err: domain.ErrUnavailable},
		{api: &apiSpy{}, closeErr: errors.New("close failed")},
	} {
		var out, diag bytes.Buffer
		got := Run(context.Background(), []string{"inspect", "team", "Team1", "--db", "x", "--tenant", "acme", "--app", "hrms"}, strings.NewReader(""), &out, &diag, connector.connect, nil)
		if got != 4 || diag.Len() == 0 {
			t.Fatalf("exit %d, stderr %q", got, diag.String())
		}
	}
}

func TestNilAPIFromSuccessfulConnectorStillClosesExactlyOnce(t *testing.T) {
	connector := &connectorSpy{}
	var out, diag bytes.Buffer
	got := Run(context.Background(), []string{"inspect", "team", "Team1", "--db", "x", "--tenant", "acme", "--app", "hrms"}, strings.NewReader(""), &out, &diag, connector.connect, nil)
	if got != 4 || connector.calls != 1 || connector.closes != 1 {
		t.Fatalf("exit=%d calls=%d closes=%d", got, connector.calls, connector.closes)
	}
}

func TestRecordOutputFailureIsUnavailableAndStillCloses(t *testing.T) {
	connector := &connectorSpy{api: &apiSpy{}}
	var diag bytes.Buffer
	got := Run(context.Background(), []string{"inspect", "team", "Team1", "--db", "x", "--tenant", "acme", "--app", "hrms"}, strings.NewReader(""), failingWriter{}, &diag, connector.connect, nil)
	if got != 4 || connector.closes != 1 || diag.Len() == 0 {
		t.Fatalf("exit=%d closes=%d stderr=%q", got, connector.closes, diag.String())
	}
}

func TestCheckAndAssignReadInjectedInputAndForwardArea(t *testing.T) {
	for _, args := range [][]string{
		{"check", "assignment", "--file", "-", "--db", "x", "--tenant", "acme", "--app", "hrms"},
		{"assign", "--file", "-", "--fixture-context", "maya-team1", "--db", "x", "--tenant", "acme", "--app", "hrms"},
	} {
		api := &apiSpy{}
		connector := &connectorSpy{api: api}
		var out, diag bytes.Buffer
		if got := Run(context.Background(), args, strings.NewReader("proposal"), &out, &diag, connector.connect, nil); got != 0 {
			t.Fatalf("%q exit %d: %s", args, got, diag.String())
		}
		if string(api.raw) != "proposal" || api.area.TenantID() != "acme" || api.area.ApplicationID() != "hrms" || connector.closes != 1 {
			t.Fatalf("%q wrong forwarding", args)
		}
	}
}
func TestHelpIsReusableWithoutAuthorityContext(t *testing.T) {
	for _, arg := range []string{"help", "--help", "-h"} {
		var out, diag bytes.Buffer
		if got := Run(context.Background(), []string{arg}, strings.NewReader(""), &out, &diag, nil, nil); got != 0 {
			t.Fatal("help failed", got)
		}
		if !strings.Contains(out.String(), "--tenant") || !strings.Contains(out.String(), "--app") || diag.Len() != 0 {
			t.Fatal("help must explain both boundaries")
		}
	}
}
func TestCancelledCommandDoesNotDispatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out, diag bytes.Buffer
	if got := Run(ctx, []string{"assign", "--tenant", "acme", "--app", "hrms"}, strings.NewReader(""), &out, &diag, nil, nil); got != 4 {
		t.Fatal("cancelled command did not report evaluation failure", got)
	}
}

func TestExpiredDeadlineDoesNotDispatch(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	var out, diag bytes.Buffer
	if got := Run(ctx, []string{"assign", "--tenant", "acme", "--app", "hrms"}, strings.NewReader(""), &out, &diag, nil, nil); got != 4 {
		t.Fatal("expired deadline did not report evaluation failure", got)
	}
	if out.Len() != 0 || diag.Len() == 0 {
		t.Fatal("deadline result used wrong output stream")
	}
}
