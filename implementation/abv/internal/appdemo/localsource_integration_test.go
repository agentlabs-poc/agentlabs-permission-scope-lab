package appdemo_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	storageSQLite "agentlabs.local/abv/internal/storage/sqlite"
	"agentlabs.local/abv/lab"
	"agentlabs.local/authmiddleware"
	"agentlabs.local/wiring/localsource"
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
	newer := fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1}]
	newer.Revision = 2
	newer.Scope = map[string]string{"cert": "C18"}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 2}] = newer
	notBefore0, notBefore1 := now.Add(-2*time.Hour), now.Add(-time.Hour)
	expires0, expires1 := now.Add(2*time.Hour), now.Add(time.Hour)
	setValidity := func(id string, notBefore, expiresAt time.Time) {
		key := domain.GrantKey{ID: id, Revision: 1}
		content := fixture.Snapshot.Contents[key]
		content.Validity = &domain.Validity{NotBefore: &notBefore, ExpiresAt: &expiresAt}
		fixture.Snapshot.Contents[key] = content
	}
	setValidity("fk3x9r2m0dq3", notBefore0, expires0)
	setValidity("fk3x9r2m5iv8", notBefore1, expires1)
	dbPath := filepath.Join(t.TempDir(), "authority.db")
	provider, err := storageSQLite.CreateFixture(t.Context(), dbPath, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.Close(); err != nil {
		t.Fatal(err)
	}

	source, err := localsource.Open(t.Context(), dbPath, agentCredential, labGate(t, fixture.Snapshot.Area), &fixedClock{now}, labRegistry{})
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
	allowed, err := request("acme", "hrms", "fi7io4lvjwu8", exact("FIN", "C17"))
	if err != nil || allowed.Decision != authmiddleware.Allow || !reflect.DeepEqual(allowed.GrantIDs, []string{"fk3x9r2m0dq3", "fk3x9r2m5iv8", "fk3x9r2man0d"}) {
		t.Fatalf("allowed=%+v err=%v", allowed, err)
	}
	authority, err := source.Load(t.Context(), authmiddleware.AuthorityQuery{Context: authmiddleware.RequestContext{Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, Identity: authmiddleware.Identity{Version: "1", Actor: authmiddleware.Actor{Type: "user", ID: "fi7io4lvjwu8"}, HumanID: "fi7io4lvjwu8"}}})
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
		"wrong department":   {"fi7io4lvjwu8", exact("ENG", "C17")},
		"wrong certificate":  {"fi7io4lvjwu8", exact("FIN", "C18")},
		"all department":     {"fi7io4lvjwu8", authmiddleware.Material{"dept": {Kind: authmiddleware.SelectionAll}, "cert": {Kind: authmiddleware.SelectionExact, Value: "C17"}}},
		"missing membership": {"fi7io4lvkfsw", exact("FIN", "C17")},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := request("acme", "hrms", tc.human, tc.material)
			if err != nil || got.Decision != authmiddleware.Deny {
				t.Fatalf("got=%+v err=%v", got, err)
			}
		})
	}
	for _, boundary := range [][2]string{{"fi7io4lvkfsw", "hrms"}, {"acme", "fi7io4lvkfsw"}} {
		got, err := request(boundary[0], boundary[1], "fi7io4lvjwu8", exact("FIN", "C17"))
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
	source, err := localsource.Open(t.Context(), dbPath, agentCredential, labGate(t, fixture.Snapshot.Area), &fixedClock{now}, labRegistry{})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	evaluator, _ := authmiddleware.New(source, &fixedClock{now})
	request := authmiddleware.Request{
		Context:    authmiddleware.RequestContext{Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, Identity: authmiddleware.Identity{Version: "1", Actor: authmiddleware.Actor{Type: "user", ID: "fi7io4lvjwu8"}, HumanID: "fi7io4lvjwu8"}},
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
		if _, err := db.Exec(`UPDATE abv_l1_records SET value=? WHERE key2='assignment' AND json_extract(value,'$.id')='fm5b7t4p5iv8'`,
			`{"id":"fm5b7t4p5iv8","grant_revision":1,"status":"`+status+`"}`); err != nil {
			t.Fatal(err)
		}
	}
	setGrant := func(status string) {
		t.Helper()
		if _, err := db.Exec(`UPDATE abv_l1_records SET value=? WHERE key2='grant' AND key4='fk3x9r2m5iv8'`, `{"status":"`+status+`","trusted_root":false}`); err != nil {
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
	source, err := localsource.Open(t.Context(), dbPath, agentCredential, labGate(t, fixture.Snapshot.Area), &fixedClock{now}, labRegistry{})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	evaluator, _ := authmiddleware.New(source, &fixedClock{now})
	request := authmiddleware.Request{
		Context:    authmiddleware.RequestContext{Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, Identity: authmiddleware.Identity{Version: "1", Actor: authmiddleware.Actor{Type: "user", ID: "fi7io4lvjwu8"}, HumanID: "fi7io4lvjwu8"}},
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
	if _, err := assignmentStatus.SetAssignmentStatus(t.Context(), area, fixtureContext, "fm5b7t4pan0d", "disabled"); err != nil {
		t.Fatal(err)
	}
	assertDecision(authmiddleware.Deny)
	if _, err := assignmentStatus.SetAssignmentStatus(t.Context(), area, fixtureContext, "fm5b7t4pan0d", "enabled"); err != nil {
		t.Fatal(err)
	}
	assertDecision(authmiddleware.Allow)
	if _, err := grantStatus.SetGrantStatus(t.Context(), area, fixtureContext, domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "disabled"}); err != nil {
		t.Fatal(err)
	}
	assertDecision(authmiddleware.Deny)
	if _, err := grantStatus.SetGrantStatus(t.Context(), area, fixtureContext, domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "enabled"}); err != nil {
		t.Fatal(err)
	}
	assertDecision(authmiddleware.Allow)
}

func TestSQLiteAuthoritySourceReturnsZeroOnCorruptReadAndRejectsNilClock(t *testing.T) {
	now := time.Now()
	area, _ := domain.NewArea("acme", "hrms")
	// A typed nil in an interface is not nil, and both are refused: an
	// enforcement source with no clock cannot judge validity, and one with no
	// administration performs an ungated read — which is what this path used to
	// be.
	var nilClock *fixedClock
	if source, err := localsource.Open(t.Context(), "unused", agentCredential, labGate(t, area), nilClock, labRegistry{}); err == nil || source != nil {
		t.Fatalf("nil clock: source=%v err=%v", source, err)
	}
	var nilGate *lab.RoleAdministration
	if source, err := localsource.Open(t.Context(), "unused", agentCredential, nilGate, &fixedClock{now}, labRegistry{}); err == nil || source != nil {
		t.Fatalf("nil administration: source=%v err=%v", source, err)
	}
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
	source, err := localsource.Open(t.Context(), dbPath, agentCredential, labGate(t, fixture.Snapshot.Area), &fixedClock{now}, labRegistry{})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`UPDATE abv_l1_records SET value='{}' WHERE key2='assignment' AND json_extract(value,'$.id')='fm5b7t4pan0d'`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	query := authmiddleware.AuthorityQuery{Context: authmiddleware.RequestContext{Area: authmiddleware.Area{TenantID: "acme", ApplicationID: "hrms"}, Identity: authmiddleware.Identity{Version: "1", Actor: authmiddleware.Actor{Type: "user", ID: "fi7io4lvjwu8"}, HumanID: "fi7io4lvjwu8"}}}
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

// labRegistry answers for the fixture's area and nothing else. A stub that said
// yes to everything would hide the installation gate, which is exactly what the
// wrong-boundary cases here assert.
type labRegistry struct{}

func (labRegistry) ApplicationExists(_ context.Context, applicationID string) (bool, error) {
	return applicationID == "hrms", nil
}
func (labRegistry) Installed(_ context.Context, tenantID, applicationID string) (bool, error) {
	return tenantID == "acme" && applicationID == "hrms", nil
}

// agentCredential is what the application asks as. The lab's gate admits this
// one within its own area, standing in for a workload credential.
var agentCredential = domain.Actor{Type: "service_account", ID: lab.WorkloadClient}

// labGate is the administration this path never had. The adapter now performs a
// gated read, and the gate is the deployment's — here, the fixture's.
func labGate(t *testing.T, area domain.Area) *lab.RoleAdministration {
	t.Helper()
	status, err := lab.NewAssignmentStatusAdministration(area, lab.TeamFINC17(area).Administration)
	if err != nil {
		t.Fatal(err)
	}
	return &lab.RoleAdministration{AssignmentStatusAdministration: status}
}
