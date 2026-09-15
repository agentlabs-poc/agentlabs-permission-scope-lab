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
	check   func(storage.Snapshot, domain.Identity, domain.RoleContent) error
	readErr error
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
	id := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvl534"}, HumanID: "fi7io4lvl534"}
	role := domain.RoleContent{Name: "payslip-reader", Revision: 2, Permissions: []string{"hrms:payroll:payslip::read"}}
	snap := storage.Snapshot{Area: area, Catalog: domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{"hrms:payroll:payslip::read": {ID: "hrms:payroll:payslip::read", Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"}}}, Roles: map[domain.RoleKey]domain.RoleContent{}}
	p := &roleProvider{snapshot: snap}
	admin := roleAdmin{check: func(s storage.Snapshot, _ domain.Identity, r domain.RoleContent) error {
		s.Area = domain.Area{}
		s.Catalog.Permissions["hrms:payroll:payslip::read"] = domain.PermissionDefinition{}
		r.Permissions[0] = "forged"
		return nil
	}}
	s, _ := New(p, admin, fixedClock{})
	got, err := s.PublishRole(t.Context(), area, id, role)
	// The id is issued, so the returned record carries one the caller did not
	// supply; everything else must round-trip unchanged, and the caller's slice
	// must be untouched by the administration callback.
	want := role
	want.ID = got.ID
	if err != nil || got.ID == "" || !reflect.DeepEqual(got, want) || p.writes != 1 || role.Permissions[0] != "hrms:payroll:payslip::read" {
		t.Fatalf("got=%#v writes=%d input=%#v err=%v", got, p.writes, role, err)
	}
}

func TestPublishRoleFailuresReturnZeroAndDoNotWrite(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	id := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvl534"}, HumanID: "fi7io4lvl534"}
	role := domain.RoleContent{Name: "payslip-reader", Revision: 2, Permissions: []string{"hrms:payroll:payslip::read"}}
	// A supplied id is only legal when it names a role that already exists.
	existing := domain.RoleContent{Name: "payslip-reader", ID: "reader", Revision: 2, Permissions: []string{"hrms:payroll:payslip::read"}}
	base := storage.Snapshot{Area: area, Catalog: domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{"hrms:payroll:payslip::read": {ID: "hrms:payroll:payslip::read", Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"}}}, Roles: map[domain.RoleKey]domain.RoleContent{}}
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
			x.Roles = map[domain.RoleKey]domain.RoleContent{{ID: "reader", Revision: 2}: existing}
			return x
		}(), roleAdmin{}, id, existing, nil, domain.ErrConflict}, {"invented id", base, roleAdmin{}, id, existing, nil, domain.ErrNotFound}, {"wrong area", func() storage.Snapshot { x := base; x.Area, _ = domain.NewArea("fi7io4lvkfsw", "hrms"); return x }(), roleAdmin{}, id, role, nil, domain.ErrRejected}, {"commit", base, roleAdmin{}, id, role, domain.ErrUnavailable, domain.ErrUnavailable},
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
	id := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvl534"}, HumanID: "fi7io4lvl534"}
	p := &roleProvider{snapshot: storage.Snapshot{Area: area, Catalog: domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{"hrms:payroll:payslip::read": {ID: "hrms:payroll:payslip::read", Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"}}}, Roles: map[domain.RoleKey]domain.RoleContent{}}}
	ctx, cancel := context.WithCancel(t.Context())
	admin := roleAdmin{check: func(storage.Snapshot, domain.Identity, domain.RoleContent) error { cancel(); return nil }}
	s, _ := New(p, admin, fixedClock{})
	got, err := s.PublishRole(ctx, area, id, domain.RoleContent{Name: "payslip-reader", Revision: 1, Permissions: []string{"hrms:payroll:payslip::read"}})
	if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(got, domain.RoleContent{}) || p.writes != 0 {
		t.Fatalf("got=%#v writes=%d err=%v", got, p.writes, err)
	}
}

// roleReader serves a fixed snapshot to the two read operations.
type roleReader struct {
	snapshot storage.Snapshot
	err      error
}

func (p *roleReader) Read(_ context.Context, _ domain.Area, cb func(storage.Snapshot) error) error {
	if p.err != nil {
		return p.err
	}
	return cb(p.snapshot)
}
func (p *roleReader) Close() error { return nil }
func (p *roleReader) Update(context.Context, domain.Area, func(storage.Snapshot) (storage.WriteSet, error)) error {
	return domain.ErrUnsupported
}

func (a roleAdmin) CheckRoleRead(context.Context, domain.Area, domain.Identity, time.Time) error {
	if a.readErr != nil {
		return a.readErr
	}
	return nil
}

func roleReadFixture(t *testing.T, readErr error) (*Service, domain.Area, domain.Identity) {
	t.Helper()
	area, _ := domain.NewArea("acme", "hrms")
	read := "hrms:payroll:payslip::read"
	write := "hrms:payroll:payslip::write"
	snap := storage.Snapshot{
		Area:    area,
		Catalog: domain.Catalog{ApplicationID: "hrms", Generation: 9},
		Roles: map[domain.RoleKey]domain.RoleContent{
			{ID: "aaa", Revision: 1}:  {ID: "aaa", Name: "reader", Revision: 1, Permissions: []string{read}},
			{ID: "aaa", Revision: 2}:  {ID: "aaa", Name: "reader", Revision: 2, Permissions: []string{read, write}},
			{ID: "aaa", Revision: 10}: {ID: "aaa", Name: "reader", Revision: 10, Permissions: []string{write}},
			{ID: "bbb", Revision: 1}:  {ID: "bbb", Name: "reader", Revision: 1, Permissions: []string{read}},
		},
	}
	service, err := New(&roleReader{snapshot: snap}, roleAdmin{readErr: readErr}, fixedClock{})
	if err != nil {
		t.Fatal(err)
	}
	return service, area, domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvl534"}, HumanID: "fi7io4lvl534"}
}

// A role is a family of revisions, and the default listing says so.
func TestListRolesReturnsEveryRevisionByDefault(t *testing.T) {
	service, area, identity := roleReadFixture(t, nil)
	page, err := service.ListRoles(t.Context(), area, identity, domain.RoleFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 4 || len(page.Roles) != 4 || page.Generation != 9 {
		t.Fatalf("total=%d len=%d generation=%d", page.Total, len(page.Roles), page.Generation)
	}
	// Ordered by id, then revision — numerically, not as text, which is what
	// the padded slot buys: revision 10 comes after 2, never between 1 and 2.
	want := []int64{1, 2, 10, 1}
	for i, role := range page.Roles {
		if role.Revision != want[i] {
			t.Fatalf("position %d is revision %d, want %d", i, role.Revision, want[i])
		}
	}
}

// LatestRevision is computed at read time, and Total counts roles rather than
// rows so the page count is right for what was asked.
func TestListRolesLatestCollapsesToOneRowPerRole(t *testing.T) {
	service, area, identity := roleReadFixture(t, nil)
	page, err := service.ListRoles(t.Context(), area, identity, domain.RoleFilter{Revisions: domain.LatestRevision})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 2 || len(page.Roles) != 2 {
		t.Fatalf("total=%d len=%d, want 2 roles", page.Total, len(page.Roles))
	}
	if page.Roles[0].ID != "aaa" || page.Roles[0].Revision != 10 {
		t.Fatalf("latest of aaa is %d, want 10", page.Roles[0].Revision)
	}
	if page.Roles[1].ID != "bbb" || page.Roles[1].Revision != 1 {
		t.Fatalf("latest of bbb is %d, want 1", page.Roles[1].Revision)
	}
}

func TestListRolesFiltersExactlyAndPages(t *testing.T) {
	service, area, identity := roleReadFixture(t, nil)
	byID, err := service.ListRoles(t.Context(), area, identity, domain.RoleFilter{ID: "aaa"})
	if err != nil || byID.Total != 3 {
		t.Fatalf("one role's history is %d entries err=%v", byID.Total, err)
	}
	// The name is not unique, so filtering by it may select several roles.
	byName, err := service.ListRoles(t.Context(), area, identity, domain.RoleFilter{Name: "reader"})
	if err != nil || byName.Total != 4 {
		t.Fatalf("name matched %d, want every row err=%v", byName.Total, err)
	}
	// Past the end is an empty page with a true total, never an error.
	past, err := service.ListRoles(t.Context(), area, identity, domain.RoleFilter{Offset: 99})
	if err != nil || past.Total != 4 || len(past.Roles) != 0 || past.Roles == nil {
		t.Fatalf("total=%d roles=%#v err=%v", past.Total, past.Roles, err)
	}
	for _, bad := range []domain.RoleFilter{{Limit: maxRolePage + 1}, {Offset: -1}, {Limit: -1}, {ID: "r*"}, {Revisions: 7}} {
		if _, err := service.ListRoles(t.Context(), area, identity, bad); !errors.Is(err, domain.ErrMalformed) {
			t.Fatalf("filter %#v gave %v, want ErrMalformed", bad, err)
		}
	}
}

// GetRole always requires a revision: keeping the typed single read exact is
// what stops an implicit latest-role fallback appearing where adoption reads.
func TestGetRoleRequiresAnExactRevision(t *testing.T) {
	service, area, identity := roleReadFixture(t, nil)
	role, err := service.GetRole(t.Context(), area, identity, "aaa", 2)
	if err != nil || role.Revision != 2 || len(role.Permissions) != 2 || role.Name != "reader" {
		t.Fatalf("role=%#v err=%v", role, err)
	}
	if _, err := service.GetRole(t.Context(), area, identity, "aaa", 3); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("an unpublished revision gave %v, want ErrNotFound", err)
	}
	if _, err := service.GetRole(t.Context(), area, identity, "aaa", 0); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("a missing revision gave %v, want ErrMalformed — never the latest", err)
	}
	if _, err := service.GetRole(t.Context(), area, identity, "payslip-reader", 1); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("a non-base-36 id gave %v, want ErrMalformed", err)
	}
}

// Unsupported, never unprotected: a denied read yields nothing.
func TestRoleReadsAreProtected(t *testing.T) {
	service, area, identity := roleReadFixture(t, domain.ErrRejected)
	if _, err := service.GetRole(t.Context(), area, identity, "aaa", 1); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("get past a denying gate: %v", err)
	}
	if _, err := service.ListRoles(t.Context(), area, identity, domain.RoleFilter{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("list past a denying gate: %v", err)
	}
}

// A mutated returned slice must not reach the stored snapshot.
func TestRoleReadsReturnCopies(t *testing.T) {
	service, area, identity := roleReadFixture(t, nil)
	role, err := service.GetRole(t.Context(), area, identity, "aaa", 1)
	if err != nil {
		t.Fatal(err)
	}
	role.Permissions[0] = "forged"
	again, err := service.GetRole(t.Context(), area, identity, "aaa", 1)
	if err != nil || again.Permissions[0] == "forged" {
		t.Fatalf("a caller reached into stored content: %#v err=%v", again, err)
	}
}

// A tenant may not revise a role the application ships: the shipped role is the
// application's property, and a tenant wanting a different bundle composes its
// own rather than editing someone else's.
func TestATenantCannotReviseAnApplicationRole(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	read := "hrms:payroll:payslip::read"
	snap := storage.Snapshot{
		Area:    area,
		Catalog: domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{read: {ID: read, Active: true, Boundary: domain.ApplicationBoundary}}},
		Roles: map[domain.RoleKey]domain.RoleContent{
			{ID: "aaaaaaaaaaaa", Revision: 1}: {ID: "aaaaaaaaaaaa", Name: "Viewer", Revision: 1, Permissions: []string{read}, Managed: domain.ApplicationManaged},
			{ID: "bbbbbbbbbbbb", Revision: 1}: {ID: "bbbbbbbbbbbb", Name: "own", Revision: 1, Permissions: []string{read}, Managed: domain.TenantManaged},
		},
	}
	p := &roleProvider{snapshot: snap}
	s, _ := New(p, roleAdmin{}, fixedClock{})
	identity := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvl534"}, HumanID: "fi7io4lvl534"}

	shipped := domain.RoleContent{ID: "aaaaaaaaaaaa", Name: "Viewer", Revision: 2, Permissions: []string{read}}
	if _, err := s.PublishRole(t.Context(), area, identity, shipped); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a tenant revised an application role: %v", err)
	}
	if p.writes != 0 {
		t.Fatalf("a refused revision wrote %d times", p.writes)
	}
	// Its own role it may revise.
	own := domain.RoleContent{ID: "bbbbbbbbbbbb", Name: "own", Revision: 2, Permissions: []string{read}}
	if _, err := s.PublishRole(t.Context(), area, identity, own); err != nil {
		t.Fatalf("a tenant could not revise its own role: %v", err)
	}
}

// A listing returns both kinds by default, and Managed narrows to one.
func TestListRolesReturnsBothKindsAndNarrows(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	read := "hrms:payroll:payslip::read"
	snap := storage.Snapshot{
		Area:    area,
		Catalog: domain.Catalog{ApplicationID: "hrms", Generation: 3},
		Roles: map[domain.RoleKey]domain.RoleContent{
			{ID: "aaaaaaaaaaaa", Revision: 1}: {ID: "aaaaaaaaaaaa", Name: "Viewer", Revision: 1, Permissions: []string{read}, Managed: domain.ApplicationManaged},
			{ID: "bbbbbbbbbbbb", Revision: 1}: {ID: "bbbbbbbbbbbb", Name: "own", Revision: 1, Permissions: []string{read}, Managed: domain.TenantManaged},
		},
	}
	s, _ := New(&roleReader{snapshot: snap}, roleAdmin{}, fixedClock{})
	identity := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvl534"}, HumanID: "fi7io4lvl534"}

	both, err := s.ListRoles(t.Context(), area, identity, domain.RoleFilter{})
	if err != nil || both.Total != 2 {
		t.Fatalf("a tenant sees %d roles, want both kinds err=%v", both.Total, err)
	}
	shipped := domain.ApplicationManaged
	only, err := s.ListRoles(t.Context(), area, identity, domain.RoleFilter{Managed: &shipped})
	if err != nil || only.Total != 1 || only.Roles[0].Name != "Viewer" {
		t.Fatalf("narrowing to shipped gave %#v err=%v", only.Roles, err)
	}
	composed := domain.TenantManaged
	mine, err := s.ListRoles(t.Context(), area, identity, domain.RoleFilter{Managed: &composed})
	if err != nil || mine.Total != 1 || mine.Roles[0].Name != "own" {
		t.Fatalf("narrowing to composed gave %#v err=%v", mine.Roles, err)
	}
}
