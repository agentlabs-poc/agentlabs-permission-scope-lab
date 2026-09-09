package sqlite_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/sqlite"
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestGrantRevisionInsertIsImmutableAndAreaBound(t *testing.T) {
	a1, _ := domain.NewArea("a", "hrms")
	a2, _ := domain.NewArea("b", "hrms")
	f1, f2 := lab.TeamFINC17(a1), lab.TeamFINC17(a2)
	path := t.TempDir() + "/a.db"
	p, err := sqlite.CreateFixture(t.Context(), path, []storage.Snapshot{f1.Snapshot, f2.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	next := domain.GrantContent{Version: "1", GrantID: "G2", Revision: 2, ParentGrantID: "G1", RoleID: "payslip-reader", RoleRevision: 1, Scope: map[string]string{"cert": "C17"}}
	for _, area := range []domain.Area{a1, a2} {
		candidate := next
		if area == a2 {
			candidate.Permissions, candidate.RoleID, candidate.RoleRevision = []string{lab.PayslipRead}, "", 0
		}
		if err = p.Update(t.Context(), area, func(storage.Snapshot) (storage.WriteSet, error) {
			return storage.WriteSet{NewGrantRevision: &candidate}, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err = p.Close(); err != nil {
		t.Fatal(err)
	}
	p, err = sqlite.Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if err = p.Read(t.Context(), a1, func(s storage.Snapshot) error {
		if len(s.Contents) != 4 || s.Contents[domain.GrantKey{ID: "G2", Revision: 1}].Revision != 1 || !reflect.DeepEqual(s.Contents[domain.GrantKey{ID: "G2", Revision: 2}], next) || s.Controls["G2"].Status != "enabled" || len(s.Assignments) != 2 {
			t.Fatalf("unexpected snapshot: %#v", s)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestGrantRevisionRejectsInvalidOrForgedWritesAtomically(t *testing.T) {
	area, _ := domain.NewArea("a", "hrms")
	f := lab.TeamFINC17(area)
	p, err := sqlite.CreateFixture(t.Context(), t.TempDir()+"/a.db", []storage.Snapshot{f.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	valid := domain.GrantContent{Version: "1", GrantID: "G2", Revision: 2, ParentGrantID: "G1", Permissions: []string{lab.PayslipRead}, Scope: map[string]string{"cert": "C17"}}
	cases := []struct {
		name string
		make func(storage.Snapshot) storage.WriteSet
	}{
		{"duplicate", func(storage.Snapshot) storage.WriteSet {
			x := valid
			x.Revision = 1
			return storage.WriteSet{NewGrantRevision: &x}
		}},
		{"malformed", func(storage.Snapshot) storage.WriteSet {
			x := valid
			x.Revision = 0
			return storage.WriteSet{NewGrantRevision: &x}
		}},
		{"missing grant", func(storage.Snapshot) storage.WriteSet {
			x := valid
			x.GrantID = "missing"
			return storage.WriteSet{NewGrantRevision: &x}
		}},
		{"root", func(storage.Snapshot) storage.WriteSet {
			x := valid
			x.GrantID = "G0"
			x.ParentGrantID = ""
			return storage.WriteSet{NewGrantRevision: &x}
		}},
		{"changed parent", func(storage.Snapshot) storage.WriteSet {
			x := valid
			x.ParentGrantID = "G0"
			return storage.WriteSet{NewGrantRevision: &x}
		}},
		{"unknown permission", func(s storage.Snapshot) storage.WriteSet {
			x := valid
			x.Permissions = []string{"forged"}
			s.Catalog.Permissions["forged"] = domain.PermissionDefinition{ID: "forged", Active: true}
			return storage.WriteSet{NewGrantRevision: &x}
		}},
		{"unknown scope", func(storage.Snapshot) storage.WriteSet {
			x := valid
			x.Scope = map[string]string{"unknown": "x"}
			return storage.WriteSet{NewGrantRevision: &x}
		}},
		{"unknown role", func(s storage.Snapshot) storage.WriteSet {
			x := valid
			x.Permissions = nil
			x.RoleID = "forged"
			x.RoleRevision = 1
			s.Roles[domain.RoleKey{ID: "forged", Revision: 1}] = domain.RoleContent{ID: "forged", Revision: 1, Permissions: []string{lab.PayslipRead}}
			return storage.WriteSet{NewGrantRevision: &x}
		}},
		{"mixed", func(storage.Snapshot) storage.WriteSet {
			return storage.WriteSet{NewGrantRevision: &valid, NewAssignments: []domain.Assignment{{ID: "x"}}}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := p.Update(t.Context(), area, func(s storage.Snapshot) (storage.WriteSet, error) { return tc.make(s), nil }); err == nil {
				t.Fatal("accepted")
			}
		})
	}
	sentinel := errors.New("callback")
	if err := p.Update(t.Context(), area, func(storage.Snapshot) (storage.WriteSet, error) {
		return storage.WriteSet{NewGrantRevision: &valid}, sentinel
	}); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	if err := p.Update(ctx, area, func(storage.Snapshot) (storage.WriteSet, error) {
		cancel()
		return storage.WriteSet{NewGrantRevision: &valid}, nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := p.Read(t.Context(), area, func(s storage.Snapshot) error {
		if _, ok := s.Contents[domain.GrantKey{ID: "G2", Revision: 2}]; ok {
			t.Fatal("partial insert")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestGrantRevisionConcurrentDuplicateHasOneWinner(t *testing.T) {
	area, _ := domain.NewArea("a", "hrms")
	f := lab.TeamFINC17(area)
	p, err := sqlite.CreateFixture(t.Context(), t.TempDir()+"/a.db", []storage.Snapshot{f.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	next := domain.GrantContent{Version: "1", GrantID: "G2", Revision: 2, ParentGrantID: "G1", Permissions: []string{lab.PayslipRead}, Scope: map[string]string{"cert": "C17"}}
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- p.Update(context.Background(), area, func(storage.Snapshot) (storage.WriteSet, error) {
				return storage.WriteSet{NewGrantRevision: &next}, nil
			})
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
