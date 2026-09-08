package lab

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
)

const (
	PayslipRead   = "hrms:payroll:payslip::read"
	PayslipWrite  = "hrms:payroll:payslip::write"
	PayslipDelete = "hrms:payroll:payslip::delete"
)

// AdministrationPremise is trusted test-only fixture context. It deliberately
// does not model an Auth catalog lookup inside the HRMS business snapshot.
type AdministrationPremise struct {
	HumanID         string
	RecipientTeamID string
}

type TeamFINC17Case struct {
	Snapshot       storage.Snapshot
	Child          domain.GrantContent
	Proposed       domain.Assignment
	Issuer         domain.Identity
	Administration AdministrationPremise
}

// TeamFINC17 returns the worked FIN/C17 business records and the absent,
// proposed A2 assignment. Callers may mutate the returned fixture freely.
func TeamFINC17(area domain.Area) TeamFINC17Case {
	g0 := domain.GrantContent{Version: "1", GrantID: "G0", Revision: 1, Permissions: []string{PayslipRead, PayslipWrite, PayslipDelete}, Scope: map[string]string{}}
	g1 := domain.GrantContent{Version: "1", GrantID: "G1", Revision: 1, ParentGrantID: "G0", Permissions: []string{PayslipRead, PayslipWrite}, Scope: map[string]string{"dept": "FIN"}}
	g2 := domain.GrantContent{Version: "1", GrantID: "G2", Revision: 1, ParentGrantID: "G1", Permissions: []string{PayslipRead}, Scope: map[string]string{"cert": "C17"}}
	return TeamFINC17Case{
		Snapshot: storage.Snapshot{
			Area: area,
			Catalog: domain.Catalog{
				ApplicationID: area.ApplicationID(),
				Permissions: map[string]domain.PermissionDefinition{
					PayslipRead: {ID: PayslipRead, Active: true}, PayslipWrite: {ID: PayslipWrite, Active: true}, PayslipDelete: {ID: PayslipDelete, Active: true},
				},
				Scopes: map[string]domain.ScopeDefinition{
					"dept": {Key: "dept", AllowedTokens: []string{}}, "cert": {Key: "cert", AllowedTokens: []string{}}, "user": {Key: "user", AllowedTokens: []string{"$self"}},
				},
				SupportedKeys: map[string][]string{
					PayslipRead: {"dept", "cert", "user"}, PayslipWrite: {"dept", "user"}, PayslipDelete: {"dept", "user"},
				},
			},
			Controls: map[string]domain.GrantControl{
				"G0": {Version: "1", ID: "G0", Status: "enabled"}, "G1": {Version: "1", ID: "G1", Status: "enabled"}, "G2": {Version: "1", ID: "G2", Status: "enabled"},
			},
			Contents: map[domain.GrantKey]domain.GrantContent{
				{ID: "G0", Revision: 1}: g0, {ID: "G1", Revision: 1}: g1, {ID: "G2", Revision: 1}: g2,
			},
			Assignments: map[string]domain.Assignment{
				"A0": {Version: "1", ID: "A0", GrantID: "G0", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "RootTeam"}, Status: "enabled"},
				"A1": {Version: "1", ID: "A1", GrantID: "G1", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "Team1"}, Status: "enabled"},
			},
			Roles: map[domain.RoleKey]domain.RoleContent{
				{ID: "payslip-reader", Revision: 1}: {ID: "payslip-reader", Revision: 1, Permissions: []string{PayslipRead}},
			},
			Teams: map[string]domain.Team{
				"RootTeam": {ID: "RootTeam"}, "Team1": {ID: "Team1", ParentID: "RootTeam"}, "Team2": {ID: "Team2", ParentID: "Team1"}, "AssignmentAdmins": {ID: "AssignmentAdmins"},
			},
			Memberships:  []domain.Membership{{TeamID: "Team1", HumanID: "maya"}, {TeamID: "Team2", HumanID: "nutan"}, {TeamID: "AssignmentAdmins", HumanID: "maya"}},
			TrustedRoots: map[string]bool{"G0": true},
		},
		Child:          domain.GrantContent{Version: "1", GrantID: "G2", Revision: 1, ParentGrantID: "G1", Permissions: []string{PayslipRead}, Scope: map[string]string{"cert": "C17"}},
		Proposed:       domain.Assignment{Version: "1", ID: "A2", GrantID: "G2", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "Team2"}, Status: "enabled"},
		Issuer:         domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "maya"}, HumanID: "maya"},
		Administration: AdministrationPremise{HumanID: "maya", RecipientTeamID: "Team2"},
	}
}
