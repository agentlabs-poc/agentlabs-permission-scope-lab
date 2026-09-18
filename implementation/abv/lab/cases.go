package lab

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
)

const (
	PayslipRead      = "hrms:payroll:payslip::read"
	PayslipWrite     = "hrms:payroll:payslip::write"
	PayslipDelete    = "hrms:payroll:payslip::delete"
	AssignmentCreate = PlatformNamespace + ":assignment::create"
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
	// The root names no permissions. Its coverage is the registered catalog,
	// computed at resolution (Q-122), and a stored list on a root is never read
	// — so one could only go stale and mislead whoever read the row. The fixture
	// carried the three payslip permissions here until root establishment gave
	// the shape a writer to compare against.
	g0 := domain.GrantContent{Version: "1", GrantID: "fk3x9r2m0dq3", Revision: 1, Scope: map[string]string{}}
	g1 := domain.GrantContent{Version: "1", GrantID: "fk3x9r2m5iv8", Revision: 1, ParentGrantID: "fk3x9r2m0dq3", Permissions: []string{PayslipRead, PayslipWrite}, Scope: map[string]string{"dept": "FIN"}}
	g2 := domain.GrantContent{Version: "1", GrantID: "fk3x9r2man0d", Revision: 1, ParentGrantID: "fk3x9r2m5iv8", Permissions: []string{PayslipRead}, Scope: map[string]string{"cert": "C17"}}
	return TeamFINC17Case{
		Snapshot: storage.Snapshot{
			Area: area,
			Catalog: domain.Catalog{
				ApplicationID: area.ApplicationID(),
				Permissions: map[string]domain.PermissionDefinition{
					PayslipRead:   {ID: PayslipRead, Active: true, Boundary: domain.ApplicationBoundary, Namespace: area.ApplicationID()},
					PayslipWrite:  {ID: PayslipWrite, Active: true, Boundary: domain.ApplicationBoundary, Namespace: area.ApplicationID()},
					PayslipDelete: {ID: PayslipDelete, Active: true, Boundary: domain.ApplicationBoundary, Namespace: area.ApplicationID()},
				},
				Scopes: map[string]domain.ScopeDefinition{
					"dept": {Key: "dept", Boundary: domain.ApplicationBoundary},
					"cert": {Key: "cert", Boundary: domain.ApplicationBoundary},
					"user": {Key: "user", Boundary: domain.ApplicationBoundary},
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

// TenantGenesis is the fixture at the moment before any authority exists: the
// tenant's teams and memberships are there, the application's catalog is there,
// and not one grant is.
//
// It is the only state in which a root can be established, which is why the
// worked fixture cannot stand in — TeamFINC17 already has a root, and "exactly
// one root per area" refuses before any other rule is reached.
func TenantGenesis(area domain.Area) storage.Snapshot {
	snapshot := TeamFINC17(area).Snapshot
	snapshot.Controls = map[string]domain.GrantControl{}
	snapshot.Contents = map[domain.GrantKey]domain.GrantContent{}
	snapshot.Assignments = map[string]domain.Assignment{}
	snapshot.TrustedRoots = map[string]bool{}
	return snapshot
}

// The three group permissions Q-092 approved, spelled in the platform's own
// namespace. Create covers teams and subteams, write includes human membership,
// delete is its own authority.
const (
	GroupCreate = PlatformNamespace + ":group::create"
	GroupWrite  = PlatformNamespace + ":group::write"
	GroupDelete = PlatformNamespace + ":group::delete"
)

// The administrative chain's ids. They are spelled out rather than issued because
// a fixture is a record of a state, and a state's identifiers are facts about it.
const (
	// TeamAdminsTeam holds the narrow administrative grant. It is the only team the
	// administrative fixture adds to the tenant's own.
	TeamAdminsTeam = "fibggi2jv1n8"
	AuthRootGrant  = "fk3x9r2mau01"
	// TenantAdminGrant is the unscoped administrative authority: every group verb,
	// no `team` predicate, so it satisfies any team. It is what "the tenant
	// administrator" means, and Q-155 notes it is the only kind a root hands out
	// with no predicates.
	TenantAdminGrant = "fk3x9r2mau02"
	// TeamAdminGrant is administrative authority over exactly one team. It carries
	// `auth:group::write` scoped to the C17 team, which is what lets its holder
	// add and remove that team's members and nobody else's.
	TeamAdminGrant = "fk3x9r2mau03"
)

// AuthAdministration returns the tenant's administrative chain — Q-155 /
// ADMIN-007's two-chain model, as records.
//
// Its area is the platform's namespace rather than an application, which is the
// whole point: Q-151 slices a root's ceiling by namespace, so `auth:group::*`
// can only ever be carried by a chain rooted here. An application's own root
// cannot reach it, however wide that root is.
//
// It shares the tenant's teams and memberships with the business fixture,
// because a tenant's teams exist once rather than once per application. The two
// snapshots name the same teams, and the second naming is the same team.
//
//	fk3x9r2mau01  auth root, scope {}          held by fibggi2jur5s
//	    └── fk3x9r2mau02  create/delete/write, scope {}
//	                                           held by fibggi2juubk  (maya)
//	        └── fk3x9r2mau03  write, {team: fibggi2juxhc}
//	                                           held by fibggi2jv1n8  (priya)
//
// It mirrors the business fixture's shape, and that is the point: same walk, same
// narrowing, same containment, one namespace over. Maya holds every group verb over
// every team; priya holds membership writes over exactly one.
//
// The holder of the narrow grant is a team of its own — TeamAdmins, a subteam of
// FIN — rather than the C17 team the grant is scoped to. Who administers a team and
// who is in it are different facts, and hanging the grant on its own subject would
// have made them look like one. It also keeps C17 a team nothing binds, which the
// re-parent and delete guards need in order to be tested in both directions.
//
// Because a route travels the team hierarchy, bending that hierarchy severs the
// administrative chain along with the business one. That is correct rather than
// inconvenient — see Reanchored for the fixtures that deliberately bend it.
func AuthAdministration(area domain.Area) (storage.Snapshot, error) {
	authArea, err := domain.NewArea(area.TenantID(), PlatformNamespace)
	if err != nil {
		return storage.Snapshot{}, err
	}
	business := TeamFINC17(area).Snapshot
	// The administrators' own team, and the one team this fixture adds to the
	// tenant. It hangs under FIN because a chain's first hop requires the holder's
	// direct parent to hold the grant being narrowed, and FIN holds the tenant
	// administrator's.
	teams := map[string]domain.Team{TeamAdminsTeam: {ID: TeamAdminsTeam, Name: "fp8h2w6yv1n8", ParentID: "fibggi2juubk"}}
	for id, team := range business.Teams {
		teams[id] = team
	}
	memberships := append(append([]domain.Membership{}, business.Memberships...),
		domain.Membership{TeamID: TeamAdminsTeam, HumanID: "fi7io4lvjwu8"})
	root := domain.GrantContent{Version: "1", GrantID: AuthRootGrant, Revision: 1, Scope: map[string]string{}}
	tenantAdmin := domain.GrantContent{
		Version: "1", GrantID: TenantAdminGrant, Revision: 1, ParentGrantID: AuthRootGrant,
		Permissions: []string{GroupCreate, GroupDelete, GroupWrite}, Scope: map[string]string{},
	}
	teamAdmin := domain.GrantContent{
		Version: "1", GrantID: TeamAdminGrant, Revision: 1, ParentGrantID: TenantAdminGrant,
		Permissions: []string{GroupWrite}, Scope: map[string]string{"team": "fibggi2juxhc"},
	}
	return storage.Snapshot{
		Area: authArea,
		Catalog: domain.Catalog{
			ApplicationID: PlatformNamespace,
			Permissions: map[string]domain.PermissionDefinition{
				GroupCreate: {ID: GroupCreate, Active: true, Boundary: domain.PlatformBoundary, Namespace: PlatformNamespace},
				GroupWrite:  {ID: GroupWrite, Active: true, Boundary: domain.PlatformBoundary, Namespace: PlatformNamespace},
				GroupDelete: {ID: GroupDelete, Active: true, Boundary: domain.PlatformBoundary, Namespace: PlatformNamespace},
			},
			// Q-156's key, and the only scope key Auth owns. Its values name Auth's
			// own records, so they are resolved on the way in rather than treated as
			// opaque the way an application's are.
			Scopes: map[string]domain.ScopeDefinition{
				"team": {Key: "team", Boundary: domain.PlatformBoundary},
			},
		},
		Controls: map[string]domain.GrantControl{
			AuthRootGrant:    {Version: "1", ID: AuthRootGrant, Status: "enabled"},
			TenantAdminGrant: {Version: "1", ID: TenantAdminGrant, Status: "enabled"},
			TeamAdminGrant:   {Version: "1", ID: TeamAdminGrant, Status: "enabled"},
		},
		Contents: map[domain.GrantKey]domain.GrantContent{
			{ID: AuthRootGrant, Revision: 1}:    root,
			{ID: TenantAdminGrant, Revision: 1}: tenantAdmin,
			{ID: TeamAdminGrant, Revision: 1}:   teamAdmin,
		},
		Assignments: map[string]domain.Assignment{
			"fm5b7t4pau01": {Version: "1", ID: "fm5b7t4pau01", GrantID: AuthRootGrant, GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "fibggi2jur5s"}, Status: "enabled"},
			"fm5b7t4pau02": {Version: "1", ID: "fm5b7t4pau02", GrantID: TenantAdminGrant, GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "fibggi2juubk"}, Status: "enabled"},
			"fm5b7t4pau03": {Version: "1", ID: "fm5b7t4pau03", GrantID: TeamAdminGrant, GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: TeamAdminsTeam}, Status: "enabled"},
		},
		Roles:        map[domain.RoleKey]domain.RoleContent{},
		Teams:        teams,
		Memberships:  memberships,
		Ownerships:   []domain.Ownership{},
		TrustedRoots: map[string]bool{AuthRootGrant: true},
	}, nil
}

// Administered pairs a business snapshot with the administrative chain that
// authorizes changing the tenant's teams. A store holding only the first can
// resolve business authority and no administrative authority at all, which after
// Q-155 means it can answer a request and refuse every team write.
func Administered(area domain.Area, business storage.Snapshot) ([]storage.Snapshot, error) {
	administrative, err := AuthAdministration(area)
	if err != nil {
		return nil, err
	}
	return []storage.Snapshot{business, administrative}, nil
}

// ReanchoredHolder is the team Reanchored hangs the tenant administrator's grant
// on. It is a subteam of fibggi2jv0n4, which is parentless — so the two-team
// ladder a chain needs sits entirely outside the FIN/C17 hierarchy.
const ReanchoredHolder = "fibggi2jv0n5"

// Reanchored moves an administrative chain off the FIN/C17 hierarchy and onto a
// ladder of its own, which is what a fixture that bends that hierarchy needs.
//
// A route's first hop is fixed: the parent grant must be held by the recipient's
// *direct parent team*. So a chain is as deep as the team ladder under it, and a
// cycle or a dangling parent anywhere along that ladder takes the authority away
// — correctly, and fail-closed. A test about whether the re-parent walk
// terminates would then be a test about authorization instead, answering before
// the walk it was written for is reached.
//
// The ladder is fibggi2jv0n4 → ReanchoredHolder, and the named human joins the
// second. The narrow grant is dropped rather than re-anchored: its whole content
// is a team predicate, and there is no holder for it that does not travel the
// hierarchy this exists to step off.
func Reanchored(administrative storage.Snapshot, memberID string) storage.Snapshot {
	result := administrative
	result.Teams = map[string]domain.Team{}
	for id, team := range administrative.Teams {
		result.Teams[id] = team
	}
	result.Teams[ReanchoredHolder] = domain.Team{ID: ReanchoredHolder, Name: "fp8h2w6yv0n5", ParentID: "fibggi2jv0n4"}
	result.Memberships = append(append([]domain.Membership{}, administrative.Memberships...),
		domain.Membership{TeamID: ReanchoredHolder, HumanID: memberID})
	result.Assignments = map[string]domain.Assignment{
		"fm5b7t4pau01": {Version: "1", ID: "fm5b7t4pau01", GrantID: AuthRootGrant, GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "fibggi2jv0n4"}, Status: "enabled"},
		"fm5b7t4pau02": {Version: "1", ID: "fm5b7t4pau02", GrantID: TenantAdminGrant, GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: ReanchoredHolder}, Status: "enabled"},
	}
	result.Controls = map[string]domain.GrantControl{
		AuthRootGrant:    {Version: "1", ID: AuthRootGrant, Status: "enabled"},
		TenantAdminGrant: {Version: "1", ID: TenantAdminGrant, Status: "enabled"},
	}
	result.Contents = map[domain.GrantKey]domain.GrantContent{}
	for key, content := range administrative.Contents {
		if key.ID != TeamAdminGrant {
			result.Contents[key] = content
		}
	}
	return result
}
