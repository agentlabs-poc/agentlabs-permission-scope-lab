package lab_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"errors"
	"testing"
	"time"
)

func gate(t *testing.T, area domain.Area) *lab.RoleAdministration {
	t.Helper()
	status, err := lab.NewAssignmentStatusAdministration(area, lab.TeamFINC17(area).Administration)
	if err != nil {
		t.Fatal(err)
	}
	return &lab.RoleAdministration{AssignmentStatusAdministration: status}
}

// The binding is the whole claim: a workload credential is bound to one tenant
// application, and the gate is a comparison against that binding rather than a
// policy about the subject.
//
// It had no test. Deleting the comparison left the entire suite green, because
// the only other caller is a human, and the fixture gate checks the area for a
// human anyway. A credential could have resolved in any area.
func TestAServiceCredentialIsBoundToOneArea(t *testing.T) {
	bound, _ := domain.NewArea("acme", "hrms")
	elsewhere, _ := domain.NewArea("acme", "crm")
	admin := gate(t, bound)
	credential := domain.Identity{
		Version: "1",
		Actor:   domain.Actor{Type: "service_account", ID: lab.WorkloadClient},
		HumanID: "fi7io4lvjqio",
	}

	if err := admin.CheckAuthorityRead(t.Context(), bound, credential, time.Time{}); err != nil {
		t.Fatalf("the credential was refused in the area it is bound to: %v", err)
	}
	if err := admin.CheckAuthorityRead(t.Context(), elsewhere, credential, time.Time{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("the credential resolved outside its binding: %v", err)
	}

	// And the name is not the binding. A credential this deployment never issued
	// is refused even in the right area.
	unknown := credential
	unknown.Actor.ID = "agent_crm"
	if err := admin.CheckAuthorityRead(t.Context(), bound, unknown, time.Time{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("an unissued credential was admitted: %v", err)
	}
}

// Each actor type gets a different answer, and the differences are deliberate:
// a human may ask about themselves, a credential about anyone in its area, an
// agent about nobody because delegation is not implemented, and an unknown type
// is not a policy question at all.
func TestTheGateAnswersEachActorTypeDifferently(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	admin := gate(t, area)
	maya := lab.TeamFINC17(area).Issuer

	if err := admin.CheckAuthorityRead(t.Context(), area, maya, time.Time{}); err != nil {
		t.Fatalf("a human could not ask about their own authority: %v", err)
	}
	other := maya
	other.Actor.ID, other.HumanID = "fn2q6v8sbo1e", "fn2q6v8sbo1e"
	if err := admin.CheckAuthorityRead(t.Context(), area, other, time.Time{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a different human was admitted by the fixture gate: %v", err)
	}
	delegated := maya
	delegated.Actor = domain.Actor{Type: "agent", ID: "a-17"}
	if err := admin.CheckAuthorityRead(t.Context(), area, delegated, time.Time{}); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("an agent gave %v, want ErrUnsupported — delegation is not implemented, which is not a refusal", err)
	}
	unknown := maya
	unknown.Actor = domain.Actor{Type: "robot", ID: "r2"}
	if err := admin.CheckAuthorityRead(t.Context(), area, unknown, time.Time{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("an unknown actor type gave %v", err)
	}
}
