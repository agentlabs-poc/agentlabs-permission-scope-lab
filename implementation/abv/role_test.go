package abv_test

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"testing"
	"time"
)

type publicRoleAdmin struct{}

func (publicRoleAdmin) CheckAssignment(context.Context, abv.Evidence, domain.Identity, domain.Assignment, time.Time) error {
	return nil
}
func (publicRoleAdmin) CheckRolePublication(context.Context, abv.Evidence, domain.Identity, domain.RoleContent, time.Time) error {
	return nil
}

type publicRoleProvider struct{ snapshot storage.Snapshot }

func (p *publicRoleProvider) Read(context.Context, domain.Area, func(storage.Snapshot) error) error {
	return nil
}
func (p *publicRoleProvider) Close() error { return nil }
func (p *publicRoleProvider) Update(_ context.Context, _ domain.Area, cb func(storage.Snapshot) (storage.WriteSet, error)) error {
	_, e := cb(p.snapshot)
	return e
}
func TestFacadePublishRole(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	p := &publicRoleProvider{storage.Snapshot{Area: area, Catalog: domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{"read": {ID: "read", Active: true}}}, Roles: map[domain.RoleKey]domain.RoleContent{}}}
	f, err := abv.New(p, publicRoleAdmin{}, clock{})
	if err != nil {
		t.Fatal(err)
	}
	role := domain.RoleContent{ID: "reader", Revision: 1, Permissions: []string{"read"}}
	if got, err := f.PublishRole(t.Context(), area, domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "p"}, HumanID: "p"}, role); err != nil || got.ID != "reader" {
		t.Fatalf("got=%#v err=%v", got, err)
	}
}
