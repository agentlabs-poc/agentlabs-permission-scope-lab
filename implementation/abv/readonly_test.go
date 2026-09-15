package abv_test

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/lab"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// A read-only store answers and never changes. It exists because an agent
// resolving authority on every request has no business holding a writable handle
// on a tenant's authority — and because, until it existed, the enforcement
// adapter reached past this facade into the internal packages to avoid one.
func TestReadOnlyStoreAnswersButRefusesEveryWrite(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "authority.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	registry, err := lab.NewFixedRegistry(area)
	if err != nil {
		t.Fatal(err)
	}
	status, err := lab.NewAssignmentStatusAdministration(area, lab.TeamFINC17(area).Administration)
	if err != nil {
		t.Fatal(err)
	}
	admin := &lab.RoleAdministration{AssignmentStatusAdministration: status}

	facade, err := abv.OpenSQLiteReadOnly(t.Context(), path, admin, clock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)}, registry)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = facade.Close() })
	maya := lab.TeamFINC17(area).Issuer

	// It answers.
	resolved, err := facade.ResolveAuthority(t.Context(), area, maya, domain.ResolveOptions{})
	if err != nil {
		t.Fatalf("a read-only store refused a read: %v", err)
	}
	if len(resolved.ResolvedGrants) == 0 {
		t.Fatal("resolved nothing from a seeded store")
	}
	if _, err := facade.GetTeam(t.Context(), area, maya, "fibggi2juubk"); err != nil {
		t.Fatalf("a read-only store refused a record read: %v", err)
	}

	// And refuses every write, through whichever door. Unsupported rather than
	// rejected: this store does not do that, which is not a question of
	// authority.
	for name, write := range map[string]func() error{
		"create a team": func() error {
			_, err := facade.CreateTeam(t.Context(), area, maya, "Auditors", "")
			return err
		},
		"create a grant": func() error {
			_, _, err := facade.CreateGrant(t.Context(), area, maya, "fk3x9r2m5iv8",
				domain.GrantContent{Permissions: []string{lab.PayslipRead}, Scope: map[string]string{}})
			return err
		},
		"establish a root": func() error {
			_, _, err := facade.EstablishRoot(t.Context(), area, maya, "fibggi2jur5s")
			return err
		},
		"add a member": func() error {
			return facade.AddMember(t.Context(), area, maya, "fibggi2juubk", "fn2q6v8sbo1e")
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := write(); !errors.Is(err, domain.ErrUnsupported) {
				t.Fatalf("a read-only store allowed a write: %v", err)
			}
		})
	}

	// Nothing was written by the attempts.
	after, err := facade.ResolveAuthority(t.Context(), area, maya, domain.ResolveOptions{})
	if err != nil || len(after.ResolvedGrants) != len(resolved.ResolvedGrants) {
		t.Fatalf("the store changed under refused writes: %d then %d, %v",
			len(resolved.ResolvedGrants), len(after.ResolvedGrants), err)
	}
}

// The gate is still consulted. Read-only narrows what can be done to the
// records; it decides nothing about who may ask.
func TestReadOnlyStoreStillRunsTheGate(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	path := filepath.Join(t.TempDir(), "authority.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	registry, _ := lab.NewFixedRegistry(area)
	status, err := lab.NewAssignmentStatusAdministration(area, lab.TeamFINC17(area).Administration)
	if err != nil {
		t.Fatal(err)
	}
	facade, err := abv.OpenSQLiteReadOnly(t.Context(), path,
		&lab.RoleAdministration{AssignmentStatusAdministration: status}, clock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)}, registry)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = facade.Close() })

	unissued := domain.Identity{
		Version: "1", Actor: domain.Actor{Type: "service_account", ID: "agent_crm"}, HumanID: "fi7io4lvjqio",
	}
	if _, err := facade.ResolveAuthority(t.Context(), area, unissued, domain.ResolveOptions{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a read-only store skipped the gate: %v", err)
	}
	if _, err := abv.OpenSQLiteReadOnly(t.Context(), path, nil, clock{now: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)}, nil); err == nil {
		t.Fatal("opened without a registry")
	}
}
