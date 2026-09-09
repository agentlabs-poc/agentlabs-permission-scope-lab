package lab

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/application"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestRoleAdministrationRequiresExactBoundedPremiseAndMembership(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	base, err := NewAssignmentStatusAdministration(area, TeamFINC17(area).Administration)
	if err != nil {
		t.Fatal(err)
	}
	admin := &RoleAdministration{AssignmentStatusAdministration: base}
	identity := TeamFINC17(area).Issuer
	role := domain.RoleContent{ID: "payslip-reader", Revision: 2, Permissions: []string{PayslipRead, PayslipWrite}}
	if err := admin.CheckRolePublication(t.Context(), TeamFINC17(area).Snapshot, identity, role, time.Now()); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*abv.Evidence, *domain.Identity, *domain.RoleContent){
		func(_ *abv.Evidence, _ *domain.Identity, r *domain.RoleContent) { r.ID = "other" },
		func(_ *abv.Evidence, _ *domain.Identity, r *domain.RoleContent) {
			r.Permissions = []string{PayslipDelete}
		},
		func(_ *abv.Evidence, i *domain.Identity, _ *domain.RoleContent) { i.Actor.ID = "other" },
		func(s *abv.Evidence, _ *domain.Identity, _ *domain.RoleContent) { s.Memberships = nil },
	} {
		snapshot, gotIdentity, gotRole := TeamFINC17(area).Snapshot, identity, role
		mutate(&snapshot, &gotIdentity, &gotRole)
		if err := admin.CheckRolePublication(context.Background(), snapshot, gotIdentity, gotRole, time.Now()); !errors.Is(err, domain.ErrRejected) {
			t.Fatalf("error=%v snapshot=%+v identity=%+v role=%+v", err, snapshot, gotIdentity, gotRole)
		}
	}
}

func TestRolePublicationUsesMarkedFixtureAndPreservesRevisions(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	path := filepath.Join(t.TempDir(), "lab.db")
	if err := (Scenarios{}).Seed(t.Context(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	roleAPI := api.(application.RoleAPI)
	role := domain.RoleContent{ID: "payslip-reader", Revision: 2, Permissions: []string{PayslipRead, PayslipWrite}}
	got, err := roleAPI.PublishRole(t.Context(), area, domain.FixtureContext{Name: roleFixtureContext}, role)
	if err != nil || got.ID != role.ID || got.Revision != 2 {
		t.Fatalf("publish=%+v, %v", got, err)
	}
	if err := closeConnection(); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err = Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeConnection()
	record, err := api.Inspect(t.Context(), area, "role", "payslip-reader")
	if err != nil || len(record.Rows) != 3 || record.Rows[1][0] != "1" || record.Rows[2][0] != "2" {
		t.Fatalf("reopened role=%+v, %v", record.Rows, err)
	}
	for _, proposed := range []domain.RoleContent{
		{ID: "other", Revision: 2, Permissions: []string{PayslipRead}},
		{ID: "payslip-reader", Revision: 3, Permissions: []string{PayslipDelete}},
		{ID: "payslip-reader", Revision: 3, Permissions: []string{"hrms:payroll:payslip::export"}},
		role,
	} {
		if _, err := api.(application.RoleAPI).PublishRole(t.Context(), area, domain.FixtureContext{Name: roleFixtureContext}, proposed); err == nil {
			t.Fatalf("proposal unexpectedly published: %+v", proposed)
		}
	}
	for _, fixture := range []string{"maya-team1", "application-publisher"} {
		if _, err := roleAPI.PublishRole(t.Context(), area, domain.FixtureContext{Name: fixture}, domain.RoleContent{ID: "payslip-reader", Revision: 4, Permissions: []string{PayslipRead}}); !errors.Is(err, domain.ErrRejected) {
			t.Fatalf("fixture %q error=%v", fixture, err)
		}
	}
	record, err = api.Inspect(t.Context(), area, "role", "payslip-reader")
	if err != nil || len(record.Rows) != 3 {
		t.Fatalf("failed publications changed roles: %+v, %v", record.Rows, err)
	}
}

func TestRolePublicationRequiresCurrentAdminMembership(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	path := filepath.Join(t.TempDir(), "lab.db")
	if err := (Scenarios{}).Seed(t.Context(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`DELETE FROM memberships WHERE team_id='AssignmentAdmins' AND human_id='maya'`); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeConnection()
	got, err := api.(application.RoleAPI).PublishRole(t.Context(), area, domain.FixtureContext{Name: roleFixtureContext}, domain.RoleContent{ID: "payslip-reader", Revision: 2, Permissions: []string{PayslipRead}})
	if !errors.Is(err, domain.ErrRejected) || got.ID != "" || got.Revision != 0 || got.Permissions != nil {
		t.Fatalf("publish=%+v, %v", got, err)
	}
}

func TestRolePublicationRejectsUnmarkedDatabase(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	path := filepath.Join(t.TempDir(), "generic.db")
	fixture := TeamFINC17(area)
	provider, err := CreateSQLite(t.Context(), path, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if err = provider.Close(); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeConnection()
	if _, err := api.(application.RoleAPI).PublishRole(t.Context(), area, domain.FixtureContext{Name: roleFixtureContext}, domain.RoleContent{ID: "payslip-reader", Revision: 2, Permissions: []string{PayslipRead}}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("unmarked publish error=%v", err)
	}
}
