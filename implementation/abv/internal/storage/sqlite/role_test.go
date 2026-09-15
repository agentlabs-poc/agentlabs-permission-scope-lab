package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/validation"
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
)

func roleSnapshot(area domain.Area) storage.Snapshot {
	return storage.Snapshot{Area: area, Catalog: domain.Catalog{ApplicationID: area.ApplicationID(), Permissions: map[string]domain.PermissionDefinition{"hrms:payroll:payslip::read": {ID: "hrms:payroll:payslip::read", Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"}, "hrms:payroll:payslip::write": {ID: "hrms:payroll:payslip::write", Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"}}, Scopes: map[string]domain.ScopeDefinition{}}, Controls: map[string]domain.GrantControl{}, Contents: map[domain.GrantKey]domain.GrantContent{{ID: "fk3x9r2mbo1e", Revision: 1}: {Version: "1", GrantID: "fk3x9r2mbo1e", Revision: 1, RoleID: "fr4j5x7z1bo1", RoleRevision: 1, Scope: map[string]string{}}}, Assignments: map[string]domain.Assignment{"fm5b7t4pbo1e": {Version: "1", ID: "fm5b7t4pbo1e", GrantID: "fk3x9r2mbo1e", GrantRevision: 1, Recipient: domain.Recipient{Type: "user", ID: "fn2q6v8s5iv8"}, Status: "enabled"}}, Roles: map[domain.RoleKey]domain.RoleContent{{ID: "fr4j5x7z1bo1", Revision: 1}: {ID: "fr4j5x7z1bo1", Name: "payslip-reader", Revision: 1, Permissions: []string{"hrms:payroll:payslip::read"}}}, Teams: map[string]domain.Team{}, Memberships: []domain.Membership{}, Ownerships: []domain.Ownership{}, TrustedRoots: map[string]bool{}}
}

func TestRolePublicationInsertIsImmutableAndAreaBound(t *testing.T) {
	a1, _ := domain.NewArea("a", "hrms")
	a2, _ := domain.NewArea("b", "hrms")
	path := t.TempDir() + "/a.db"
	p, err := CreateFixture(t.Context(), path, []storage.Snapshot{roleSnapshot(a1), roleSnapshot(a2)})
	if err != nil {
		t.Fatal(err)
	}
	role := domain.RoleContent{Name: "payslip-reader", ID: "fr4j5x7z1bo1", Revision: 2, Permissions: []string{"hrms:payroll:payslip::read", "hrms:payroll:payslip::write"}}
	if err = p.Update(t.Context(), a1, func(s storage.Snapshot) (storage.WriteSet, error) {
		return storage.WriteSet{NewRoleRevision: &role}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err = p.Update(t.Context(), a2, func(s storage.Snapshot) (storage.WriteSet, error) {
		copy := role
		copy.Permissions = []string{"hrms:payroll:payslip::write"}
		return storage.WriteSet{NewRoleRevision: &copy}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err = p.Update(t.Context(), a1, func(s storage.Snapshot) (storage.WriteSet, error) {
		return storage.WriteSet{NewRoleRevision: &role}, nil
	}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate err=%v", err)
	}
	if err = p.Close(); err != nil {
		t.Fatal(err)
	}
	p, err = Open(t.Context(), path, allowAllRegistry{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	for _, tc := range []struct {
		area domain.Area
		want []string
	}{{a1, []string{"hrms:payroll:payslip::read", "hrms:payroll:payslip::write"}}, {a2, []string{"hrms:payroll:payslip::write"}}} {
		if err := p.Read(t.Context(), tc.area, func(s storage.Snapshot) error {
			if !reflect.DeepEqual(s.Roles[domain.RoleKey{ID: "fr4j5x7z1bo1", Revision: 1}].Permissions, []string{"hrms:payroll:payslip::read"}) || !reflect.DeepEqual(s.Roles[domain.RoleKey{ID: "fr4j5x7z1bo1", Revision: 2}].Permissions, tc.want) {
				t.Fatalf("roles=%#v", s.Roles)
			}
			permissions, err := validation.SelectedPermissions(s.Contents[domain.GrantKey{ID: "fk3x9r2mbo1e", Revision: 1}], s.Roles)
			if err != nil || !reflect.DeepEqual(permissions, []string{"hrms:payroll:payslip::read"}) {
				t.Fatalf("adopted revision changed: permissions=%v err=%v", permissions, err)
			}
			if got := s.Assignments["fm5b7t4pbo1e"]; got.GrantID != "fk3x9r2mbo1e" || got.GrantRevision != 1 || got.Status != "enabled" {
				t.Fatalf("assignment changed: %#v", got)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRolePublicationConcurrentDuplicateHasOneWinner(t *testing.T) {
	area, _ := domain.NewArea("a", "hrms")
	p, err := CreateFixture(t.Context(), t.TempDir()+"/a.db", []storage.Snapshot{roleSnapshot(area)})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	role := domain.RoleContent{Name: "payslip-reader", ID: "concurrent", Revision: 1, Permissions: []string{"hrms:payroll:payslip::read"}}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- p.Update(context.Background(), area, func(storage.Snapshot) (storage.WriteSet, error) { return storage.WriteSet{NewRoleRevision: &role}, nil })
		}()
	}
	wg.Wait()
	close(errs)
	var success, conflict int
	for err := range errs {
		if err == nil {
			success++
		} else if errors.Is(err, domain.ErrConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
}

func TestRolePublicationProviderRejectsMixedAndHostileWrites(t *testing.T) {
	area, _ := domain.NewArea("a", "hrms")
	p, err := CreateFixture(t.Context(), t.TempDir()+"/a.db", []storage.Snapshot{roleSnapshot(area)})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	role := domain.RoleContent{Name: "payslip-reader", ID: "new", Revision: 1, Permissions: []string{"hrms:payroll:payslip::read"}}
	for _, cb := range []func(storage.Snapshot) (storage.WriteSet, error){
		func(s storage.Snapshot) (storage.WriteSet, error) {
			s.Catalog.Permissions["hrms:payroll:payslip::forged"] = domain.PermissionDefinition{ID: "hrms:payroll:payslip::forged", Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"}
			x := role
			x.Permissions = []string{"forged"}
			return storage.WriteSet{NewRoleRevision: &x}, nil
		},
		func(s storage.Snapshot) (storage.WriteSet, error) {
			return storage.WriteSet{NewRoleRevision: &role, NewAssignments: []domain.Assignment{{ID: "x"}}}, nil
		},
	} {
		if err := p.Update(t.Context(), area, cb); err == nil {
			t.Fatal("hostile/mixed write accepted")
		}
	}
	sentinel := errors.New("callback failed")
	if err := p.Update(t.Context(), area, func(storage.Snapshot) (storage.WriteSet, error) {
		return storage.WriteSet{NewRoleRevision: &role}, sentinel
	}); !errors.Is(err, sentinel) {
		t.Fatalf("callback error=%v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	if err := p.Update(ctx, area, func(storage.Snapshot) (storage.WriteSet, error) {
		cancel()
		return storage.WriteSet{NewRoleRevision: &role}, nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error=%v", err)
	}
	if err := p.Read(t.Context(), area, func(s storage.Snapshot) error {
		if _, ok := s.Roles[domain.RoleKey{ID: "new", Revision: 1}]; ok {
			t.Fatal("failed operation wrote role")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
