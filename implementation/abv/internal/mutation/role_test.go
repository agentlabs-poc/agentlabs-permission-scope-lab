package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type roleAdmin struct {
	check func(storage.Snapshot, domain.Identity, domain.RoleContent) error
}

func (a roleAdmin) CheckAssignment(context.Context, storage.Snapshot, domain.Identity, domain.Assignment, time.Time) error {
	return nil
}
func (a roleAdmin) CheckRolePublication(_ context.Context, s storage.Snapshot, i domain.Identity, r domain.RoleContent, _ time.Time) error {
	if a.check != nil {
		return a.check(s, i, r)
	}
	return nil
}

type roleProvider struct {
	snapshot storage.Snapshot
	writes   int
	err      error
}

func (p *roleProvider) Read(context.Context, domain.Area, func(storage.Snapshot) error) error {
	return nil
}
func (p *roleProvider) Close() error { return nil }
func (p *roleProvider) Update(_ context.Context, _ domain.Area, cb func(storage.Snapshot) (storage.WriteSet, error)) error {
	w, e := cb(p.snapshot)
	if e != nil {
		return e
	}
	if p.err != nil {
		return p.err
	}
	if w.NewRoleRevision != nil {
		p.writes++
	}
	return nil
}

func TestPublishRoleProtectsAndIsolatesProposal(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	id := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "p"}, HumanID: "p"}
	role := domain.RoleContent{ID: "reader", Revision: 2, Permissions: []string{"read"}}
	snap := storage.Snapshot{Area: area, Catalog: domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{"read": {ID: "read", Active: true}}}, Roles: map[domain.RoleKey]domain.RoleContent{}}
	p := &roleProvider{snapshot: snap}
	admin := roleAdmin{check: func(s storage.Snapshot, _ domain.Identity, r domain.RoleContent) error {
		s.Area = domain.Area{}
		s.Catalog.Permissions["read"] = domain.PermissionDefinition{}
		r.Permissions[0] = "forged"
		return nil
	}}
	s, _ := New(p, admin, fixedClock{})
	got, err := s.PublishRole(t.Context(), area, id, role)
	if err != nil || !reflect.DeepEqual(got, role) || p.writes != 1 || role.Permissions[0] != "read" {
		t.Fatalf("got=%#v writes=%d input=%#v err=%v", got, p.writes, role, err)
	}
}

func TestPublishRoleFailuresReturnZeroAndDoNotWrite(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	id := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "p"}, HumanID: "p"}
	role := domain.RoleContent{ID: "reader", Revision: 2, Permissions: []string{"read"}}
	base := storage.Snapshot{Area: area, Catalog: domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{"read": {ID: "read", Active: true}}}, Roles: map[domain.RoleKey]domain.RoleContent{}}
	for _, tc := range []struct {
		name  string
		snap  storage.Snapshot
		admin Administration
		id    domain.Identity
		role  domain.RoleContent
		perr  error
		want  error
	}{
		{"missing admin", base, oldAdmin{}, id, role, nil, domain.ErrUnsupported}, {"rejected", base, roleAdmin{check: func(storage.Snapshot, domain.Identity, domain.RoleContent) error { return domain.ErrRejected }}, id, role, nil, domain.ErrRejected}, {"identity", base, roleAdmin{}, domain.Identity{}, role, nil, domain.ErrMalformed}, {"duplicate", func() storage.Snapshot {
			x := base
			x.Roles = map[domain.RoleKey]domain.RoleContent{{ID: "reader", Revision: 2}: role}
			return x
		}(), roleAdmin{}, id, role, nil, domain.ErrConflict}, {"wrong area", func() storage.Snapshot { x := base; x.Area, _ = domain.NewArea("other", "hrms"); return x }(), roleAdmin{}, id, role, nil, domain.ErrRejected}, {"commit", base, roleAdmin{}, id, role, domain.ErrUnavailable, domain.ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &roleProvider{snapshot: tc.snap, err: tc.perr}
			s, _ := New(p, tc.admin, fixedClock{})
			got, err := s.PublishRole(t.Context(), area, tc.id, tc.role)
			if !errors.Is(err, tc.want) || !reflect.DeepEqual(got, domain.RoleContent{}) || p.writes != 0 {
				t.Fatalf("got=%#v writes=%d err=%v want=%v", got, p.writes, err, tc.want)
			}
		})
	}
}

func TestPublishRoleCancellationInsideAdministrationDoesNotWrite(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	id := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "p"}, HumanID: "p"}
	p := &roleProvider{snapshot: storage.Snapshot{Area: area, Catalog: domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{"read": {ID: "read", Active: true}}}, Roles: map[domain.RoleKey]domain.RoleContent{}}}
	ctx, cancel := context.WithCancel(t.Context())
	admin := roleAdmin{check: func(storage.Snapshot, domain.Identity, domain.RoleContent) error { cancel(); return nil }}
	s, _ := New(p, admin, fixedClock{})
	got, err := s.PublishRole(ctx, area, id, domain.RoleContent{ID: "reader", Revision: 1, Permissions: []string{"read"}})
	if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(got, domain.RoleContent{}) || p.writes != 0 {
		t.Fatalf("got=%#v writes=%d err=%v", got, p.writes, err)
	}
}
