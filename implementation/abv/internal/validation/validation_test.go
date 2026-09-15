package validation

import (
	"agentlabs.local/abv/domain"
	"errors"
	"reflect"
	"testing"
	"time"
)

const read = "hrms:employee:certificate::read"
const write = "hrms:employee:certificate::write"

func area(t *testing.T) domain.Area {
	t.Helper()
	a, e := domain.NewArea("acme", "hrms")
	if e != nil {
		t.Fatal(e)
	}
	return a
}
func catalog() domain.Catalog {
	return domain.Catalog{ApplicationID: "hrms",
		Permissions:   map[string]domain.PermissionDefinition{read: {ID: read, Active: true, Boundary: domain.ApplicationBoundary}, write: {ID: write, Active: true, Boundary: domain.ApplicationBoundary}},
		Scopes:        map[string]domain.ScopeDefinition{"dept": {Key: "dept"}, "cert": {Key: "cert"}, "user": {Key: "user"}},
	}
}
func content() domain.GrantContent {
	return domain.GrantContent{Version: "1", GrantID: "fk3x9r2man0d", Revision: 1, ParentGrantID: "fk3x9r2m5iv8", Permissions: []string{read}, Scope: map[string]string{"cert": "C17"}}
}
func TestContentRequiresRegisteredDefinitionsAndContext(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*domain.Catalog, *domain.GrantContent)
		want error
	}{
		{"valid", func(c *domain.Catalog, g *domain.GrantContent) {}, nil},
		{"unknown permission", func(c *domain.Catalog, g *domain.GrantContent) {
			g.Permissions = []string{"hrms:employee:certificate::delete"}
		}, domain.ErrRejected},
		{"retired", func(c *domain.Catalog, g *domain.GrantContent) {
			c.Permissions[read] = domain.PermissionDefinition{ID: read, Active: false, Boundary: domain.ApplicationBoundary}
		}, domain.ErrRejected},
		{"unknown key", func(c *domain.Catalog, g *domain.GrantContent) { g.Scope["secret"] = "x" }, domain.ErrRejected},
		{"unknown token", func(c *domain.Catalog, g *domain.GrantContent) { g.Scope["user"] = "$owner" }, domain.ErrRejected},
		{"registered self", func(c *domain.Catalog, g *domain.GrantContent) { g.Scope["user"] = "$self" }, nil},
		{"different app catalog", func(c *domain.Catalog, g *domain.GrantContent) { c.ApplicationID = "accounting" }, domain.ErrRejected},
		{"a second registered permission", func(c *domain.Catalog, g *domain.GrantContent) { g.Permissions = []string{write} }, nil},
		{"nil scope", func(c *domain.Catalog, g *domain.GrantContent) { g.Scope = nil }, domain.ErrMalformed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, g := catalog(), content()
			tc.edit(&c, &g)
			err := CheckContent(area(t), c, g, nil)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v; want %v", err, tc.want)
			}
		})
	}
	if err := CheckContent(domain.Area{}, catalog(), content(), nil); !errors.Is(err, domain.ErrMalformed) {
		t.Fatal("zero outer boundary accepted", err)
	}
}
func TestRoleUsesExactAdoptedRevision(t *testing.T) {
	g := content()
	g.Permissions = nil
	g.RoleID = "reader"
	g.RoleRevision = 1
	roles := map[domain.RoleKey]domain.RoleContent{
		{ID: "reader", Revision: 1}: {ID: "reader", Name: "payslip-reader", Revision: 1, Permissions: []string{read}},
		{ID: "reader", Revision: 2}: {ID: "reader", Revision: 2, Permissions: []string{"hrms:employee:certificate::delete"}},
	}
	if err := CheckContent(area(t), catalog(), g, roles); err != nil {
		t.Fatal(err)
	}
	delete(roles, domain.RoleKey{ID: "reader", Revision: 1})
	if err := CheckContent(area(t), catalog(), g, roles); !errors.Is(err, domain.ErrRejected) {
		t.Fatal("silently adopted latest role", err)
	}
	for name, role := range map[string]domain.RoleContent{
		"mismatched ID":        {Name: "payslip-reader", ID: "fi7io4lvkfsw", Revision: 1, Permissions: []string{read}},
		"mismatched revision":  {ID: "reader", Revision: 2, Permissions: []string{read}},
		"duplicate permission": {ID: "reader", Revision: 1, Permissions: []string{read, read}},
	} {
		t.Run(name, func(t *testing.T) {
			bad := map[domain.RoleKey]domain.RoleContent{{ID: "reader", Revision: 1}: role}
			if err := CheckContent(area(t), catalog(), g, bad); err == nil {
				t.Fatal("accepted corrupt selected role")
			}
		})
	}
}

func TestContentChecksSelectedDefinitionIntegrityAndTokens(t *testing.T) {
	for name, edit := range map[string]func(*domain.Catalog, *domain.GrantContent){
		"permission identity mismatch": func(c *domain.Catalog, _ *domain.GrantContent) {
			c.Permissions[read] = domain.PermissionDefinition{ID: write, Active: true, Boundary: domain.ApplicationBoundary}
		},
		"scope identity mismatch": func(c *domain.Catalog, _ *domain.GrantContent) {
			c.Scopes["cert"] = domain.ScopeDefinition{Key: "fi7io4lvkfsw"}
		},
	} {
		t.Run(name, func(t *testing.T) {
			c, g := catalog(), content()
			edit(&c, &g)
			if err := CheckContent(area(t), c, g, nil); !errors.Is(err, domain.ErrRejected) {
				t.Fatalf("accepted corrupt selected definition: %v", err)
			}
		})
	}
}
func TestNarrowKeepsAllRestrictionsAndCopiesInputs(t *testing.T) {
	a := area(t)
	expires := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	parent := domain.Route{Area: a, GrantID: "fk3x9r2m5iv8", Permissions: []string{read, write}, Predicates: []domain.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"}}, AssignmentIDs: []string{"fm5b7t4p5iv8"}, Validities: []domain.Validity{{ExpiresAt: &expires}}}
	g := content()
	g.Scope = map[string]string{"dept": "ENG", "cert": "C17"}
	got, err := Narrow(a, parent, g, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"}, {Key: "cert", Value: "C17", SourceGrantID: "fk3x9r2man0d"}, {Key: "dept", Value: "ENG", SourceGrantID: "fk3x9r2man0d"}}
	if got.Area != a || got.GrantID != "fk3x9r2man0d" || !reflect.DeepEqual(got.Predicates, want) || !reflect.DeepEqual(got.Permissions, []string{read}) || len(got.Validities) != 1 {
		t.Fatalf("lost restrictions: %#v", got)
	}
	got.Predicates[0].Value = "BROKEN"
	got.Permissions[0] = "BROKEN"
	got.AssignmentIDs[0] = "BROKEN"
	*got.Validities[0].ExpiresAt = time.Time{}
	if parent.Predicates[0].Value != "FIN" || parent.Permissions[0] != read || parent.AssignmentIDs[0] != "fm5b7t4p5iv8" || g.Permissions[0] != read || expires.IsZero() {
		t.Fatal("mutated parent's authority through alias")
	}
	g.Scope = map[string]string{}
	got, err = Narrow(a, parent, g, nil)
	if err != nil || len(got.Predicates) != 1 || got.Predicates[0].Value != "FIN" {
		t.Fatal("{} widened inherited boundary", err)
	}
}

func TestNarrowAppendsChildValidityAndDeepCopiesEveryConstraint(t *testing.T) {
	a := area(t)
	parentStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	childEnd := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	parent := domain.Route{Area: a, GrantID: "fk3x9r2m5iv8", Permissions: []string{read}, Validities: []domain.Validity{{NotBefore: &parentStart}}}
	g := content()
	g.Validity = &domain.Validity{ExpiresAt: &childEnd}
	got, err := Narrow(a, parent, g, nil)
	if err != nil || len(got.Validities) != 2 || got.Validities[1].ExpiresAt == nil || !got.Validities[1].ExpiresAt.Equal(childEnd) {
		t.Fatalf("child validity was not appended: %#v, %v", got, err)
	}
	*got.Validities[0].NotBefore = time.Time{}
	*got.Validities[1].ExpiresAt = time.Time{}
	got.Permissions[0] = "changed"
	if parentStart.IsZero() || childEnd.IsZero() || g.Permissions[0] != read || g.Validity.ExpiresAt.IsZero() {
		t.Fatal("result aliases a parent, child, or selected-permission constraint")
	}
}
func TestNarrowRejectsPermissionAndOuterBoundaryEscape(t *testing.T) {
	a := area(t)
	g := content()
	parent := domain.Route{Area: a, GrantID: "fk3x9r2m5iv8", Permissions: []string{read}}
	g.Permissions = []string{write}
	if _, err := Narrow(a, parent, g, nil); !errors.Is(err, domain.ErrRejected) {
		t.Fatal("expanded permission", err)
	}
	g.Permissions = []string{read}
	if _, err := Narrow(domain.Area{}, parent, g, nil); !errors.Is(err, domain.ErrMalformed) {
		t.Fatal("zero context", err)
	}
	for _, ids := range [][2]string{{"fi7io4lvkfsw", "hrms"}, {"acme", "accounting"}} {
		other, e := domain.NewArea(ids[0], ids[1])
		if e != nil {
			t.Fatal(e)
		}
		if _, err := Narrow(other, parent, g, nil); !errors.Is(err, domain.ErrRejected) {
			t.Fatal("crossed boundary", err)
		}
	}
	parent.GrantID = "unrelated"
	if _, err := Narrow(a, parent, g, nil); !errors.Is(err, domain.ErrRejected) {
		t.Fatal("unrelated parent", err)
	}
}

func TestNarrowDerivesExactAdoptedRolePermissions(t *testing.T) {
	a := area(t)
	parent := domain.Route{Area: a, GrantID: "fk3x9r2m5iv8", Permissions: []string{read, write}}
	child := content()
	child.Permissions = nil
	child.RoleID = "reader"
	child.RoleRevision = 1
	roles := map[domain.RoleKey]domain.RoleContent{
		{ID: "reader", Revision: 1}: {ID: "reader", Name: "payslip-reader", Revision: 1, Permissions: []string{read}},
	}
	if err := CheckContent(a, catalog(), child, roles); err != nil {
		t.Fatal("valid role premise", err)
	}
	got, err := Narrow(a, parent, child, roles)
	if err != nil || !reflect.DeepEqual(got.Permissions, []string{read}) {
		t.Fatalf("did not derive exact adopted role: %#v, %v", got, err)
	}
	got.Permissions[0] = "changed"
	if roles[domain.RoleKey{ID: "reader", Revision: 1}].Permissions[0] != read {
		t.Fatal("result aliases adopted role permissions")
	}
}

func TestNarrowRejectsIncompleteOrCorruptPermissionSources(t *testing.T) {
	a := area(t)
	roleChild := content()
	roleChild.Permissions = nil
	roleChild.RoleID = "reader"
	roleChild.RoleRevision = 1
	for _, tc := range []struct {
		name              string
		parentPermissions []string
		child             domain.GrantContent
		roles             map[domain.RoleKey]domain.RoleContent
	}{
		{"missing role", []string{read}, roleChild, nil},
		{"only newer revision", []string{read}, roleChild, map[domain.RoleKey]domain.RoleContent{
			{ID: "reader", Revision: 2}: {ID: "reader", Name: "payslip-reader", Revision: 2, Permissions: []string{read}},
		}},
		{"mismatched role record", []string{read}, roleChild, map[domain.RoleKey]domain.RoleContent{
			{ID: "reader", Revision: 1}: {ID: "fi7io4lvkfsw", Revision: 1, Permissions: []string{read}},
		}},
		{"role cannot be partially trimmed", []string{read}, roleChild, map[domain.RoleKey]domain.RoleContent{
			{ID: "reader", Revision: 1}: {ID: "reader", Name: "payslip-reader", Revision: 1, Permissions: []string{read, write}},
		}},
		{"direct content cannot be partially trimmed", []string{read}, func() domain.GrantContent {
			g := content()
			g.Permissions = []string{read, write}
			return g
		}(), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parent := domain.Route{Area: a, GrantID: "fk3x9r2m5iv8", Permissions: tc.parentPermissions}
			if got, err := Narrow(a, parent, tc.child, tc.roles); !errors.Is(err, domain.ErrRejected) {
				t.Fatalf("accepted incomplete or corrupt permission source: %#v, %v", got, err)
			}
		})
	}
}
