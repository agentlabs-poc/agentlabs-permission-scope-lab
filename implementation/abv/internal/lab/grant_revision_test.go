package lab

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func proposedGrantRevision() domain.GrantContent {
	return domain.GrantContent{Version: "1", GrantID: "G2", Revision: 2, ParentGrantID: "G1", Permissions: []string{PayslipRead, PayslipWrite}, Scope: map[string]string{"cert": "C17"}}
}

func proposedGrantRevisionJSON(t *testing.T, grant domain.GrantContent) []byte {
	t.Helper()
	raw, err := json.Marshal(grant)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestGrantRevisionAdministrationRequiresExactBoundedPremiseAndMembership(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	base, err := NewAssignmentStatusAdministration(area, TeamFINC17(area).Administration)
	if err != nil {
		t.Fatal(err)
	}
	admin := &GrantRevisionAdministration{RoleAdministration: &RoleAdministration{AssignmentStatusAdministration: base}}
	identity, snapshot, proposed := TeamFINC17(area).Issuer, TeamFINC17(area).Snapshot, proposedGrantRevision()
	if err := admin.CheckGrantRevisionPublication(t.Context(), snapshot, identity, "A1", proposed, time.Now()); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*abv.Evidence, *domain.Identity, *string, *domain.GrantContent){
		func(s *abv.Evidence, _ *domain.Identity, _ *string, _ *domain.GrantContent) {
			other, _ := domain.NewArea("other", "hrms")
			s.Area = other
		},
		func(_ *abv.Evidence, i *domain.Identity, _ *string, _ *domain.GrantContent) { i.Version = "2" },
		func(_ *abv.Evidence, _ *domain.Identity, source *string, _ *domain.GrantContent) { *source = "A2" },
		func(_ *abv.Evidence, _ *domain.Identity, _ *string, g *domain.GrantContent) { g.GrantID = "G1" },
		func(s *abv.Evidence, _ *domain.Identity, _ *string, _ *domain.GrantContent) { s.Memberships = nil },
	} {
		gotSnapshot, gotIdentity, source, gotGrant := snapshot, identity, "A1", proposed
		mutate(&gotSnapshot, &gotIdentity, &source, &gotGrant)
		if err := admin.CheckGrantRevisionPublication(context.Background(), gotSnapshot, gotIdentity, source, gotGrant, time.Now()); !errors.Is(err, domain.ErrRejected) {
			t.Fatalf("error=%v", err)
		}
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := admin.CheckGrantRevisionPublication(cancelled, snapshot, identity, "A1", proposed, time.Now()); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error=%v", err)
	}
}

func TestGrantPublicationUsesMarkedFixtureAndPreservesAssignment(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	path := filepath.Join(t.TempDir(), "lab.db")
	if err := (Scenarios{}).Seed(t.Context(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	assignment := TeamFINC17(area).Proposed
	raw := []byte(`{"version":"1","id":"A2","grant_id":"G2","grant_revision":1,"recipient":{"type":"group","id":"Team2"},"status":"enabled"}`)
	if _, err = api.Assign(t.Context(), area, domain.FixtureContext{Name: labFixtureContext}, raw); err != nil {
		t.Fatal(err)
	}
	publication := api.(application.GrantRevisionAPI)
	got, err := publication.PublishGrantRevision(t.Context(), area, domain.FixtureContext{Name: grantRevisionFixtureContext}, "A1", proposedGrantRevisionJSON(t, proposedGrantRevision()))
	if err != nil || got.Revision != 2 {
		t.Fatalf("publish=%+v err=%v", got, err)
	}
	if err = closeConnection(); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err = Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeConnection()
	publication = api.(application.GrantRevisionAPI)
	grant, err := api.Inspect(t.Context(), area, "grant", "G2")
	if err != nil || string(grant.CanonicalJSON) != `{"version":"1","grant_id":"G2","revision":2,"parent_grant_id":"G1","permissions":["hrms:payroll:payslip::read","hrms:payroll:payslip::write"],"scope":{"cert":"C17"}}` {
		t.Fatalf("grant=%q err=%v", grant.CanonicalJSON, err)
	}
	reopened, err := api.Inspect(t.Context(), area, "assignment", "A2")
	if err != nil || string(reopened.CanonicalJSON) != string(raw) || assignment.ID != "A2" {
		t.Fatalf("assignment=%q err=%v", reopened.CanonicalJSON, err)
	}
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var count int
	if err = db.QueryRow(`SELECT count(*) FROM grant_contents WHERE grant_id='G2'`).Scan(&count); err != nil || count != 2 {
		t.Fatalf("revisions=%d err=%v", count, err)
	}
	for _, tc := range []struct {
		fixture, source string
		grant           domain.GrantContent
	}{
		{"maya-team1", "A1", proposedGrantRevision()}, {"application-publisher", "A1", proposedGrantRevision()}, {"maya-role-publisher", "A1", proposedGrantRevision()},
		{grantRevisionFixtureContext, "A2", func() domain.GrantContent { g := proposedGrantRevision(); g.Revision = 3; return g }()},
		{grantRevisionFixtureContext, "A1", func() domain.GrantContent {
			g := proposedGrantRevision()
			g.Revision = 3
			g.Permissions = []string{PayslipDelete}
			return g
		}()},
		{grantRevisionFixtureContext, "A1", proposedGrantRevision()},
	} {
		if _, err := publication.PublishGrantRevision(t.Context(), area, domain.FixtureContext{Name: tc.fixture}, tc.source, proposedGrantRevisionJSON(t, tc.grant)); err == nil {
			t.Fatalf("unexpected publish: %+v", tc)
		}
	}
	if _, err := publication.PublishGrantRevision(t.Context(), area, domain.FixtureContext{Name: grantRevisionFixtureContext}, "A1", []byte(`{"version":"1","grant_id":"G2","grant_id":"other","revision":3,"parent_grant_id":"G1","permissions":["hrms:payroll:payslip::read"],"scope":{"cert":"C17"}}`)); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("malformed publication error=%v", err)
	}
}

func TestGrantPublicationRejectsMissingMembershipAndUnmarkedDatabase(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	for _, marked := range []bool{true, false} {
		path := filepath.Join(t.TempDir(), "lab.db")
		if marked {
			if err := (Scenarios{}).Seed(t.Context(), area, "team-fin-c17", path); err != nil {
				t.Fatal(err)
			}
			db, _ := sql.Open("sqlite", "file:"+path)
			_, err := db.Exec(`DELETE FROM memberships WHERE team_id='AssignmentAdmins' AND human_id='maya'`)
			_ = db.Close()
			if err != nil {
				t.Fatal(err)
			}
		} else {
			provider, err := CreateSQLite(t.Context(), path, []storage.Snapshot{TeamFINC17(area).Snapshot})
			if err != nil {
				t.Fatal(err)
			}
			if err = provider.Close(); err != nil {
				t.Fatal(err)
			}
		}
		api, closeConnection, err := Connect(t.Context(), area, path)
		if err != nil {
			t.Fatal(err)
		}
		_, err = api.(application.GrantRevisionAPI).PublishGrantRevision(t.Context(), area, domain.FixtureContext{Name: grantRevisionFixtureContext}, "A1", proposedGrantRevisionJSON(t, proposedGrantRevision()))
		_ = closeConnection()
		if !errors.Is(err, domain.ErrRejected) {
			t.Fatalf("marked=%v err=%v", marked, err)
		}
	}
}
