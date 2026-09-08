package lab

import (
	"agentlabs.local/abv/domain"
	"errors"
	"testing"
	"time"
)

func TestGrantStatusAdministrationRequiresCurrentDirectMembership(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := TeamFINC17(area)
	admin, err := NewGrantStatusAdministration(area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	control := domain.GrantControl{Version: "1", ID: "G2", Status: "disabled"}
	if err = admin.CheckGrantStatus(t.Context(), fixture.Snapshot, fixture.Issuer, control, time.Time{}); err != nil {
		t.Fatal(err)
	}
	fixture.Snapshot.Memberships = fixture.Snapshot.Memberships[:2]
	if err = admin.CheckGrantStatus(t.Context(), fixture.Snapshot, fixture.Issuer, control, time.Time{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("membership removal error = %v", err)
	}
}

func TestAssignmentStatusAdministrationIsSeparateAndExact(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := TeamFINC17(area)
	fixture.Snapshot.Assignments["A2"] = fixture.Proposed
	admin, err := NewAssignmentStatusAdministration(area, fixture.Administration)
	if err != nil {
		t.Fatal(err)
	}
	for _, assignment := range []domain.Assignment{
		fixture.Snapshot.Assignments["A1"],
		fixture.Proposed,
	} {
		assignment.Status = "disabled"
		if err = admin.CheckAssignmentStatus(t.Context(), fixture.Snapshot, fixture.Issuer, assignment, time.Time{}); err != nil {
			t.Fatalf("%s: %v", assignment.ID, err)
		}
	}
	wrong := fixture.Proposed
	wrong.GrantRevision = 2
	if err = admin.CheckAssignmentStatus(t.Context(), fixture.Snapshot, fixture.Issuer, wrong, time.Time{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("changed assignment error = %v", err)
	}
	fixture.Snapshot.Memberships = fixture.Snapshot.Memberships[:2]
	if err = admin.CheckAssignmentStatus(t.Context(), fixture.Snapshot, fixture.Issuer, fixture.Proposed, time.Time{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("membership removal error = %v", err)
	}
}
