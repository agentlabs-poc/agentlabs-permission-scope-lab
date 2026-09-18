package abv_test

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"errors"
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

// UpdateAdministered delegates: a fake's snapshot is the whole world it has, so
// it is its own administrative chain, and a nil Snapshot.Administrative says
// exactly that.
func (p *publicRoleProvider) UpdateAdministered(ctx context.Context, area domain.Area, callback func(storage.Snapshot) (storage.WriteSet, error)) error {
	return p.Update(ctx, area, callback)
}
func TestFacadePublishRole(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	p := &publicRoleProvider{storage.Snapshot{Area: area, Catalog: domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{"hrms:payroll:payslip::read": {ID: "hrms:payroll:payslip::read", Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"}}}, Roles: map[domain.RoleKey]domain.RoleContent{}}}
	f, err := abv.New(p, publicRoleAdmin{}, clock{})
	if err != nil {
		t.Fatal(err)
	}
	// No id: a new role is issued one, and the issued id comes back in the
	// returned record. A caller cannot choose an identifier.
	role := domain.RoleContent{Name: "payslip-reader", Revision: 1, Permissions: []string{"hrms:payroll:payslip::read"}}
	identity := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvl534"}, HumanID: "fi7io4lvl534"}
	got, err := f.PublishRole(t.Context(), area, identity, role)
	if err != nil || got.ID == "" || got.Name != "payslip-reader" {
		t.Fatalf("got=%#v err=%v", got, err)
	}
	// An id the caller invented names no role, so it is refused rather than
	// quietly creating one.
	invented := domain.RoleContent{Name: "x", ID: "zzzzzzzz", Revision: 1, Permissions: []string{"hrms:payroll:payslip::read"}}
	if _, err := f.PublishRole(t.Context(), area, identity, invented); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("an invented id gave %v, want ErrNotFound", err)
	}
}
