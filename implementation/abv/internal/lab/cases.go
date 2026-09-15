package lab

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
)

const (
	PayslipRead      = "hrms:payroll:payslip::read"
	PayslipWrite     = "hrms:payroll:payslip::write"
	PayslipDelete    = "hrms:payroll:payslip::delete"
	AssignmentCreate = "auth:assignment::create"
)

// AdministrationPremise is trusted test-only fixture context. It deliberately
// does not model an Auth catalog lookup inside the HRMS business snapshot.
type AdministrationPremise struct {
	HumanID         string
	PermissionID    string
	RecipientTeamID string
}

type TeamFINC17Case struct {
	Snapshot       storage.Snapshot
	Child          domain.GrantContent
	Proposed       domain.Assignment
	Issuer         domain.Identity
	Administration AdministrationPremise
}

// payslip builds the fixture's three identifiers for one application. A
// permission's leading noun is its application, so a fixture cannot use the
// hrms spelling inside some other application and still be coherent — for the
// hrms area these are exactly the PayslipRead/Write/Delete constants.
func payslip(applicationID string) (read, write, remove string) {
	base := applicationID + ":payroll:payslip::"
	return base + "read", base + "write", base + "delete"
}

// TeamFINC17 returns the worked FIN/C17 business records and the absent,
// proposed fm5b7t4pan0d assignment. Callers may mutate the returned fixture freely.
func TeamFINC17(area domain.Area) TeamFINC17Case {
	PayslipRead, PayslipWrite, PayslipDelete := payslip(area.ApplicationID())
	g0 := domain.GrantContent{Version: "1", GrantID: "fk3x9r2m0dq3", Revision: 1, Permissions: []string{PayslipRead, PayslipWrite, PayslipDelete}, Scope: map[string]string{}}
	g1 := domain.GrantContent{Version: "1", GrantID: "fk3x9r2m5iv8", Revision: 1, ParentGrantID: "fk3x9r2m0dq3", Permissions: []string{PayslipRead, PayslipWrite}, Scope: map[string]string{"dept": "FIN"}}
	g2 := domain.GrantContent{Version: "1", GrantID: "fk3x9r2man0d", Revision: 1, ParentGrantID: "fk3x9r2m5iv8", Permissions: []string{PayslipRead}, Scope: map[string]string{"cert": "C17"}}
	return TeamFINC17Case{
		Snapshot: storage.Snapshot{
			Area: area,
			Catalog: domain.Catalog{
				ApplicationID: area.ApplicationID(),
				Permissions: map[string]domain.PermissionDefinition{
					PayslipRead: {ID: PayslipRead, Active: true}, PayslipWrite: {ID: PayslipWrite, Active: true}, PayslipDelete: {ID: PayslipDelete, Active: true},
				},
				Scopes: map[string]domain.ScopeDefinition{
					"dept": {Key: "dept"}, "cert": {Key: "cert"}, "user": {Key: "user"},
				},
			},
			Controls: map[string]domain.GrantControl{
				"fk3x9r2m0dq3": {Version: "1", ID: "fk3x9r2m0dq3", Status: "enabled"}, "fk3x9r2m5iv8": {Version: "1", ID: "fk3x9r2m5iv8", Status: "enabled"}, "fk3x9r2man0d": {Version: "1", ID: "fk3x9r2man0d", Status: "enabled"},
			},
			Contents: map[domain.GrantKey]domain.GrantContent{
				{ID: "fk3x9r2m0dq3", Revision: 1}: g0, {ID: "fk3x9r2m5iv8", Revision: 1}: g1, {ID: "fk3x9r2man0d", Revision: 1}: g2,
			},
			Assignments: map[string]domain.Assignment{
				"fm5b7t4p0dq3": {Version: "1", ID: "fm5b7t4p0dq3", GrantID: "fk3x9r2m0dq3", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "fibggi2jur5s"}, Status: "enabled"},
				"fm5b7t4p5iv8": {Version: "1", ID: "fm5b7t4p5iv8", GrantID: "fk3x9r2m5iv8", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "fibggi2juubk"}, Status: "enabled"},
			},
			Roles: map[domain.RoleKey]domain.RoleContent{
				{ID: "fi9jvxobqsxs", Revision: 1}: {ID: "fi9jvxobqsxs", Name: "payslip-reader", Revision: 1, Permissions: []string{PayslipRead}},
			},
			Teams: map[string]domain.Team{
				"fibggi2jur5s": {ID: "fibggi2jur5s", Name: "fp8h2w6ykxan"}, "fibggi2juubk": {ID: "fibggi2juubk", Name: "fp8h2w6y5iv8", ParentID: "fibggi2jur5s"}, "fibggi2juxhc": {ID: "fibggi2juxhc", Name: "fp8h2w6yan0d", ParentID: "fibggi2juubk"}, "fibggi2jv0n4": {ID: "fibggi2jv0n4", Name: "AssignmentAdmins"},
			},
			Memberships:  []domain.Membership{{TeamID: "fibggi2juubk", HumanID: "fi7io4lvjqio"}, {TeamID: "fibggi2juxhc", HumanID: "fi7io4lvjwu8"}, {TeamID: "fibggi2jv0n4", HumanID: "fi7io4lvjqio"}},
			TrustedRoots: map[string]bool{"fk3x9r2m0dq3": true},
		},
		Child:          domain.GrantContent{Version: "1", GrantID: "fk3x9r2man0d", Revision: 1, ParentGrantID: "fk3x9r2m5iv8", Permissions: []string{PayslipRead}, Scope: map[string]string{"cert": "C17"}},
		Proposed:       domain.Assignment{Version: "1", ID: "fm5b7t4pan0d", GrantID: "fk3x9r2man0d", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "fibggi2juxhc"}, Status: "enabled"},
		Issuer:         domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"},
		Administration: AdministrationPremise{HumanID: "fi7io4lvjqio", PermissionID: AssignmentCreate, RecipientTeamID: "fibggi2juxhc"},
	}
}
