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
	return storage.Snapshot{Area: area, Catalog: domain.Catalog{ApplicationID: area.ApplicationID(), Permissions: map[string]domain.PermissionDefinition{"read": {ID: "read", Active: true}, "write": {ID: "write", Active: true}}, Scopes: map[string]domain.ScopeDefinition{}, SupportedKeys: map[string][]string{}}, Controls: map[string]domain.GrantControl{}, Contents: map[domain.GrantKey]domain.GrantContent{{ID: "grant", Revision: 1}: {Version: "1", GrantID: "grant", Revision: 1, RoleID: "reader", RoleRevision: 1, Scope: map[string]string{}}}, Assignments: map[string]domain.Assignment{"assignment": {Version: "1", ID: "assignment", GrantID: "grant", GrantRevision: 1, Recipient: domain.Recipient{Type: "user", ID: "u"}, Status: "enabled"}}, Roles: map[domain.RoleKey]domain.RoleContent{{ID: "reader", Revision: 1}: {ID: "reader", Revision: 1, Permissions: []string{"read"}}}, Teams: map[string]domain.Team{}, Memberships: []domain.Membership{}, TrustedRoots: map[string]bool{}}
}

func TestRolePublicationInsertIsImmutableAndAreaBound(t *testing.T) {
	a1, _ := domain.NewArea("a", "hrms")
	a2, _ := domain.NewArea("b", "hrms")
	path := t.TempDir() + "/a.db"
	p, err := CreateFixture(t.Context(), path, []storage.Snapshot{roleSnapshot(a1), roleSnapshot(a2)})
	if err != nil {
		t.Fatal(err)
	}
	role := domain.RoleContent{ID: "reader", Revision: 2, Permissions: []string{"read", "write"}}
	if err = p.Update(t.Context(), a1, func(s storage.Snapshot) (storage.WriteSet, error) {
		return storage.WriteSet{NewRoleRevision: &role}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err = p.Update(t.Context(), a2, func(s storage.Snapshot) (storage.WriteSet, error) {
		copy := role
		copy.Permissions = []string{"write"}
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
	p, err = Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	for _, tc := range []struct {
		area domain.Area
		want []string
	}{{a1, []string{"read", "write"}}, {a2, []string{"write"}}} {
		if err := p.Read(t.Context(), tc.area, func(s storage.Snapshot) error {
			if !reflect.DeepEqual(s.Roles[domain.RoleKey{ID: "reader", Revision: 1}].Permissions, []string{"read"}) || !reflect.DeepEqual(s.Roles[domain.RoleKey{ID: "reader", Revision: 2}].Permissions, tc.want) {
				t.Fatalf("roles=%#v", s.Roles)
			}
			permissions, err := validation.SelectedPermissions(s.Contents[domain.GrantKey{ID: "grant", Revision: 1}], s.Roles)
			if err != nil || !reflect.DeepEqual(permissions, []string{"read"}) {
				t.Fatalf("adopted revision changed: permissions=%v err=%v", permissions, err)
			}
			if got := s.Assignments["assignment"]; got.GrantID != "grant" || got.GrantRevision != 1 || got.Status != "enabled" {
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
	role := domain.RoleContent{ID: "concurrent", Revision: 1, Permissions: []string{"read"}}
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
	role := domain.RoleContent{ID: "new", Revision: 1, Permissions: []string{"read"}}
	for _, cb := range []func(storage.Snapshot) (storage.WriteSet, error){
		func(s storage.Snapshot) (storage.WriteSet, error) {
			s.Catalog.Permissions["forged"] = domain.PermissionDefinition{ID: "forged", Active: true}
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
