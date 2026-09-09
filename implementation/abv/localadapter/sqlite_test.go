package localadapter

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/storage"
	storageSQLite "agentlabs.local/abv/internal/storage/sqlite"
	"agentlabs.local/authmiddleware"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type fixedClock struct{ now time.Time }

func (c *fixedClock) Now() time.Time { return c.now }

func TestSQLiteAuthoritySourceEvaluatesRealSnapshot(t *testing.T) {
	now := time.Date(2026, 9, 9, 10, 0, 0, 0, time.UTC)
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments[fixture.Proposed.ID] = fixture.Proposed
	newer := fixture.Snapshot.Contents[domain.GrantKey{ID: "G2", Revision: 1}]
	newer.Revision = 2
	newer.Scope = map[string]string{"cert": "C18"}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "G2", Revision: 2}] = newer
	notBefore0, notBefore1 := now.Add(-2*time.Hour), now.Add(-time.Hour)
	expires0, expires1 := now.Add(2*time.Hour), now.Add(time.Hour)
	setValidity := func(id string, notBefore, expiresAt time.Time) {
		key := domain.GrantKey{ID: id, Revision: 1}
		content := fixture.Snapshot.Contents[key]
		content.Validity = &domain.Validity{NotBefore: &notBefore, ExpiresAt: &expiresAt}
		fixture.Snapshot.Contents[key] = content
	}
	setValidity("G0", notBefore0, expires0)
	setValidity("G1", notBefore1, expires1)
	dbPath := filepath.Join(t.TempDir(), "authority.db")
	provider, err := storageSQLite.CreateFixture(t.Context(), dbPath, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.Close(); err != nil {
		t.Fatal(err)
	}

	source, err := Open(t.Context(), dbPath, &fixedClock{now})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	evaluator, err := authmiddleware.New(source, &fixedClock{now})
	if err != nil {
		t.Fatal(err)
	}
	request := func(tenant, app, human string, material authmiddleware.Material) (authmiddleware.Result, error) {
		return evaluator.Evaluate(t.Context(), authmiddleware.Request{
			Context: authmiddleware.RequestContext{
				Area:     authmiddleware.Area{TenantID: tenant, ApplicationID: app},
				Identity: authmiddleware.Identity{Version: "1", Actor: authmiddleware.Actor{Type: "user", ID: human}, HumanID: human},
			},
			Permission: lab.PayslipRead,
			Material:   material,
		})
	}
	exact := func(dept, cert string) authmiddleware.Material {
		return authmiddleware.Material{
			"dept": {Kind: authmiddleware.SelectionExact, Value: dept},
			"cert": {Kind: authmiddleware.SelectionExact, Value: cert},
		}
	}
	allowed, err := request("acme", "hrms", "nutan", exact("FIN", "C17"))
	if err != nil || allowed.Decision != authmiddleware.Allow || !reflect.DeepEqual(allowed.GrantIDs, []string{"G0", "G1", "G2"}) {
		t.Fatalf("allowed=%+v err=%v", allowed, err)
	}
	authority, err := source.Load(t.Context(), authmiddleware.AuthorityQuery{Context: authmiddleware.RequestContext{Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, Identity: authmiddleware.Identity{Version: "1", Actor: authmiddleware.Actor{Type: "user", ID: "nutan"}, HumanID: "nutan"}}, Permission: lab.PayslipRead})
	if err != nil || len(authority.Routes) != 1 {
		t.Fatalf("authority=%+v err=%v", authority, err)
	}
	if got := authority.Routes[0]; got.ValidFrom == nil || !got.ValidFrom.Equal(notBefore1) || got.ValidUntil == nil || !got.ValidUntil.Equal(expires1) {
		t.Fatalf("validity=%+v", got)
	}
	for name, tc := range map[string]struct {
		human    string
		material authmiddleware.Material
	}{
		"wrong department":   {"nutan", exact("ENG", "C17")},
		"wrong certificate":  {"nutan", exact("FIN", "C18")},
		"all department":     {"nutan", authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionAll}, "cert": {Kind: authmiddleware.SelectionExact, Value: "C17"}}},
		"missing membership": {"other", exact("FIN", "C17")},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := request("acme", "hrms", tc.human, tc.material)
			if err != nil || got.Decision != authmiddleware.Deny {
				t.Fatalf("got=%+v err=%v", got, err)
			}
		})
	}
	for _, boundary := range [][2]string{{"other", "hrms"}, {"acme", "other"}} {
		got, err := request(boundary[0], boundary[1], "nutan", exact("FIN", "C17"))
		if err == nil || !reflect.DeepEqual(got, authmiddleware.Result{}) {
			t.Fatalf("boundary %v: got=%+v err=%v", boundary, got, err)
		}
	}
}

func TestSQLiteAuthoritySourceSeesCommittedStatusChanges(t *testing.T) {
	now := time.Now()
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments[fixture.Proposed.ID] = fixture.Proposed
	dbPath := filepath.Join(t.TempDir(), "status.db")
	provider, err := storageSQLite.CreateFixture(t.Context(), dbPath, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.Close(); err != nil {
		t.Fatal(err)
	}
	source, err := Open(t.Context(), dbPath, &fixedClock{now})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	evaluator, _ := authmiddleware.New(source, &fixedClock{now})
	request := authmiddleware.Request{
		Context:    authmiddleware.RequestContext{Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, Identity: authmiddleware.Identity{Version: "1", Actor: authmiddleware.Actor{Type: "user", ID: "nutan"}, HumanID: "nutan"}},
		Permission: lab.PayslipRead,
		Material:   authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionExact, Value: "FIN"}, "cert": {Kind: authmiddleware.SelectionExact, Value: "C17"}},
	}
	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	assertDecision := func(want authmiddleware.Decision) {
		t.Helper()
		got, err := evaluator.Evaluate(t.Context(), request)
		if err != nil || got.Decision != want {
			t.Fatalf("got=%+v err=%v want=%s", got, err, want)
		}
	}
	setAssignment := func(status string) {
		t.Helper()
		raw := `{"version":"1","id":"A1","grant_id":"G1","grant_revision":1,"recipient":{"type":"group","id":"Team1"},"status":"` + status + `"}`
		if _, err := db.Exec(`UPDATE assignments SET status=?,canonical_json=? WHERE assignment_id='A1'`, status, raw); err != nil {
			t.Fatal(err)
		}
	}
	setGrant := func(status string) {
		t.Helper()
		raw := `{"version":"1","id":"G1","status":"` + status + `"}`
		if _, err := db.Exec(`UPDATE grant_controls SET status=?,canonical_json=? WHERE grant_id='G1'`, status, raw); err != nil {
			t.Fatal(err)
		}
	}
	setAssignment("disabled")
	assertDecision(authmiddleware.Deny)
	setAssignment("enabled")
	assertDecision(authmiddleware.Allow)
	setGrant("disabled")
	assertDecision(authmiddleware.Deny)
	setGrant("enabled")
	assertDecision(authmiddleware.Allow)
}

func TestSQLiteAuthoritySourceSeesProtectedDescendantStatusChanges(t *testing.T) {
	now := time.Now()
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	dbPath := filepath.Join(t.TempDir(), "status.db")
	if err := (lab.Scenarios{}).Seed(t.Context(), area, "team-fin-c17", dbPath); err != nil {
		t.Fatal(err)
	}
	api, closeAPI, err := lab.Connect(t.Context(), area, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeAPI() })
	raw, err := json.Marshal(fixture.Proposed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := api.Assign(t.Context(), area, domain.FixtureContext{Name: "maya-team1"}, raw); err != nil {
		t.Fatal(err)
	}
	assignmentStatus := api.(interface {
		SetAssignmentStatus(context.Context, domain.Area, domain.FixtureContext, string, string) (domain.Assignment, error)
	})
	grantStatus := api.(interface {
		SetGrantStatus(context.Context, domain.Area, domain.FixtureContext, domain.GrantControl) (domain.GrantControl, error)
	})
	fixtureContext := domain.FixtureContext{Name: "maya-team1"}
	source, err := Open(t.Context(), dbPath, &fixedClock{now})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	evaluator, _ := authmiddleware.New(source, &fixedClock{now})
	request := authmiddleware.Request{
		Context:    authmiddleware.RequestContext{Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, Identity: authmiddleware.Identity{Version: "1", Actor: authmiddleware.Actor{Type: "user", ID: "nutan"}, HumanID: "nutan"}},
		Permission: lab.PayslipRead,
		Material:   authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionExact, Value: "FIN"}, "cert": {Kind: authmiddleware.SelectionExact, Value: "C17"}},
	}
	assertDecision := func(want authmiddleware.Decision) {
		t.Helper()
		got, err := evaluator.Evaluate(t.Context(), request)
		if err != nil || got.Decision != want {
			t.Fatalf("got=%+v err=%v want=%s", got, err, want)
		}
	}
	assertDecision(authmiddleware.Allow)
	if _, err := assignmentStatus.SetAssignmentStatus(t.Context(), area, fixtureContext, "A2", "disabled"); err != nil {
		t.Fatal(err)
	}
	assertDecision(authmiddleware.Deny)
	if _, err := assignmentStatus.SetAssignmentStatus(t.Context(), area, fixtureContext, "A2", "enabled"); err != nil {
		t.Fatal(err)
	}
	assertDecision(authmiddleware.Allow)
	if _, err := grantStatus.SetGrantStatus(t.Context(), area, fixtureContext, domain.GrantControl{Version: "1", ID: "G2", Status: "disabled"}); err != nil {
		t.Fatal(err)
	}
	assertDecision(authmiddleware.Deny)
	if _, err := grantStatus.SetGrantStatus(t.Context(), area, fixtureContext, domain.GrantControl{Version: "1", ID: "G2", Status: "enabled"}); err != nil {
		t.Fatal(err)
	}
	assertDecision(authmiddleware.Allow)
}

func TestSQLiteAuthoritySourceReturnsZeroOnCorruptReadAndRejectsNilClock(t *testing.T) {
	var nilClock *fixedClock
	if source, err := Open(t.Context(), "unused", nilClock); err == nil || source != nil {
		t.Fatalf("source=%v err=%v", source, err)
	}
	now := time.Now()
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments[fixture.Proposed.ID] = fixture.Proposed
	dbPath := filepath.Join(t.TempDir(), "corrupt.db")
	provider, err := storageSQLite.CreateFixture(t.Context(), dbPath, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.Close(); err != nil {
		t.Fatal(err)
	}
	source, err := Open(t.Context(), dbPath, &fixedClock{now})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`UPDATE assignments SET canonical_json='{}' WHERE assignment_id='A2'`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	query := authmiddleware.AuthorityQuery{Context: authmiddleware.RequestContext{Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, Identity: authmiddleware.Identity{Version: "1", Actor: authmiddleware.Actor{Type: "user", ID: "nutan"}, HumanID: "nutan"}}, Permission: lab.PayslipRead}
	got, err := source.Load(t.Context(), query)
	if err == nil || !reflect.DeepEqual(got, authmiddleware.Authority{}) {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	got, err = source.Load(cancelled, query)
	if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(got, authmiddleware.Authority{}) {
		t.Fatalf("cancel got=%+v err=%v", got, err)
	}
}

func TestConvertRouteRejectsMissingContributingAssignment(t *testing.T) {
	route := domain.Route{AssignmentIDs: []string{"missing"}}
	got, err := convertRoute(route, map[string]domain.Assignment{}, authmiddleware.AuthorityQuery{})
	if err == nil || !reflect.DeepEqual(got, authmiddleware.Route{}) {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestConvertRouteUsesResolvedArea(t *testing.T) {
	area, _ := domain.NewArea("resolved-tenant", "resolved-app")
	route := domain.Route{Area: area, AssignmentIDs: []string{"A0"}}
	query := authmiddleware.AuthorityQuery{Context: authmiddleware.RequestContext{Area: authmiddleware.Area{TenantID: "query-tenant", ApplicationID: "query-app"}}}
	got, err := convertRoute(route, map[string]domain.Assignment{"A0": {ID: "A0", GrantID: "G0"}}, query)
	if err != nil {
		t.Fatal(err)
	}
	if got.Area != (authmiddleware.Area{TenantID: "resolved-tenant", ApplicationID: "resolved-app"}) {
		t.Fatalf("area=%+v", got.Area)
	}
}
