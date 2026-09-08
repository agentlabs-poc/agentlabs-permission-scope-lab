package cli

import (
	"agentlabs.local/abv/domain"
	"bytes"
	"context"
	"strings"
	"testing"
)

type scenarioSpy struct {
	area                     domain.Area
	scenario, caseName, path string
}

func (*scenarioSpy) Seed(context.Context, domain.Area, string, string) error {
	return domain.ErrUnsupported
}
func (s *scenarioSpy) Run(_ context.Context, area domain.Area, scenario, caseName, path string) error {
	s.area, s.scenario, s.caseName, s.path = area, scenario, caseName, path
	return nil
}

func TestScenarioRunReportsObservedRejectionNotMutationSuccess(t *testing.T) {
	spy := &scenarioSpy{}
	var out, diag bytes.Buffer
	args := []string{"scenario", "run", "team-fin-c17", "--case", "unsupported-permission", "--db", "new.db", "--tenant", "acme", "--app", "hrms"}
	if got := Run(context.Background(), args, strings.NewReader(""), &out, &diag, nil, spy); got != 0 {
		t.Fatalf("exit %d: %s", got, diag.String())
	}
	if spy.path != "new.db" || spy.area.TenantID() != "acme" || spy.area.ApplicationID() != "hrms" || !strings.Contains(out.String(), "observed expected rejection") || !strings.Contains(out.String(), "was not created") {
		t.Fatalf("spy=%+v stdout=%q", spy, out.String())
	}
}
