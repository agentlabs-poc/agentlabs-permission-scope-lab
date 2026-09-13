package validation

import (
	"agentlabs.local/abv/domain"
	"errors"
	"testing"
)

func TestCheckRolePublication(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	catalog := domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{
		"hrms:payroll:payslip::read": {ID: "hrms:payroll:payslip::read", Active: true}, "hrms:payroll:payslip::write": {ID: "hrms:payroll:payslip::write", Active: true}, "hrms:payroll:payslip::old": {ID: "hrms:payroll:payslip::old"},
	}}
	valid := domain.RoleContent{Name: "payslip-reader", ID: "reader", Revision: 2, Permissions: []string{"hrms:payroll:payslip::read", "hrms:payroll:payslip::write"}}
	if err := CheckRolePublication(area, catalog, valid); err != nil {
		t.Fatal(err)
	}
	badUTF8 := string([]byte{0xff})
	for _, tc := range []struct {
		name    string
		area    domain.Area
		catalog domain.Catalog
		role    domain.RoleContent
		want    error
	}{
		{"area", domain.Area{}, catalog, valid, domain.ErrMalformed},
		{"catalog", area, domain.Catalog{ApplicationID: "fi7io4lvkfsw", Permissions: catalog.Permissions}, valid, domain.ErrRejected},
		{"blank id", area, catalog, domain.RoleContent{Name: "payslip-reader", ID: " ", Revision: 1, Permissions: []string{"hrms:payroll:payslip::read"}}, domain.ErrMalformed},
		{"wildcard id", area, catalog, domain.RoleContent{Name: "payslip-reader", ID: "r*", Revision: 1, Permissions: []string{"hrms:payroll:payslip::read"}}, domain.ErrMalformed},
		{"utf8 id", area, catalog, domain.RoleContent{Name: "payslip-reader", ID: badUTF8, Revision: 1, Permissions: []string{"hrms:payroll:payslip::read"}}, domain.ErrMalformed},
		{"revision", area, catalog, domain.RoleContent{Name: "payslip-reader", ID: "r", Permissions: []string{"hrms:payroll:payslip::read"}}, domain.ErrMalformed},
		{"nil permissions", area, catalog, domain.RoleContent{ID: "r", Revision: 1}, domain.ErrMalformed},
		{"duplicate", area, catalog, domain.RoleContent{Name: "payslip-reader", ID: "r", Revision: 1, Permissions: []string{"read", "read"}}, domain.ErrMalformed},
		{"unknown", area, catalog, domain.RoleContent{Name: "payslip-reader", ID: "r", Revision: 1, Permissions: []string{"missing"}}, domain.ErrRejected},
		{"inactive", area, catalog, domain.RoleContent{Name: "payslip-reader", ID: "r", Revision: 1, Permissions: []string{"old"}}, domain.ErrRejected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := CheckRolePublication(tc.area, tc.catalog, tc.role); !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want=%v", err, tc.want)
			}
		})
	}
}
