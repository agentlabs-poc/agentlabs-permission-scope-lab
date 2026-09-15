package domain

import "time"

type Identity struct {
	Version string `json:"version"`
	Actor   Actor  `json:"actor"`
	HumanID string `json:"human_id"`
}
type Actor struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
type Recipient struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Grant is a grant's head as it is stored: its live status, plus whether
// trusted establishment recorded it as a root.
//
// GrantControl below is the *wire* form Q-107 approves — version, id, status,
// and nothing else. TrustedRoot is deliberately not in it: Q-119 refuses a root
// flag in submitted content, and this is what Auth recorded rather than what a
// caller said. The snapshot reader builds the wire form where it is needed, so
// there is no conversion method here for callers that do not exist.
type Grant struct {
	ID          string
	Status      string
	TrustedRoot bool
}

type GrantControl struct {
	Version string `json:"version"`
	ID      string `json:"id"`
	Status  string `json:"status"`
}
type Validity struct {
	NotBefore *time.Time `json:"not_before,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
type GrantContent struct {
	Version       string            `json:"version"`
	GrantID       string            `json:"grant_id"`
	Revision      int64             `json:"revision"`
	ParentGrantID string            `json:"parent_grant_id,omitempty"`
	Permissions   []string          `json:"permissions,omitempty"`
	RoleID        string            `json:"role_id,omitempty"`
	RoleRevision  int64             `json:"role_revision,omitempty"`
	Scope         map[string]string `json:"scope"`
	Validity      *Validity         `json:"validity,omitempty"`
}
type Assignment struct {
	Version       string    `json:"version"`
	ID            string    `json:"id"`
	GrantID       string    `json:"grant_id"`
	GrantRevision int64     `json:"grant_revision"`
	Recipient     Recipient `json:"recipient"`
	Status        string    `json:"status"`
}

// The following types are internal projections, NOT canonical JSON contracts.
// RoleContent is one published role revision: a reusable permission bundle,
// immutable once published. Q-118 approves version, id, revision and permissions
// as the record's shape; Name is ours, because a generated id is unreadable.
//
// ID is a base-36 Snowflake. Name is a human label and is deliberately NOT
// unique: the record's identity is the id, and a lookup by name may match more
// than one role.
type RoleContent struct {
	ID          string
	Name        string
	Revision    int64
	Permissions []string
	// Managed says who owns this role. It is derived from the record's
	// boundary and never supplied by a caller: which operation published it
	// decides. A reader needs it to know what it may change — a tenant
	// administrator may revise its own roles and not the application's.
	Managed RoleManagement
}

// Boundary says who owns a record, and it is the one fact the key path cannot
// carry. Two of the three were removed as drift precisely because the key path
// did hold them — a tenant is present or it is not — but nothing distinguishes
// an application record from a platform one: both have no tenant, and both put
// a namespace in key3. That distinction is genuinely new information.
//
// Deriving it from key3's value would mean Auth-AL holding the platform's
// reserved-namespace list, which belongs to the auth service and not to us.
type Boundary string

const (
	// PlatformBoundary is authority the platform itself defines, in a namespace
	// no application can claim. key3 holds that namespace and is NOT checked
	// against any application.
	PlatformBoundary Boundary = "platform"
	// ApplicationBoundary is authority an application defines. key3 holds the
	// application, and a permission's first noun must equal it.
	ApplicationBoundary Boundary = "application"
	// TenantBoundary is authority a tenant composed inside an application. key3
	// still holds the application — a tenant role belongs to one — but nothing
	// about an identifier is checked against it.
	TenantBoundary Boundary = "tenant"
)

func (b Boundary) Valid() bool {
	return b == PlatformBoundary || b == ApplicationBoundary || b == TenantBoundary
}

// RoleManagement distinguishes a role the application ships from one a tenant
// composed.
//
// Both are the same record type in the same store, told apart by whether the
// record carries a tenant. They coexist and never shadow each other: each has
// its own issued id, and a grant adopts an exact id and revision, so there is
// nothing to resolve between them. A tenant that outgrows a shipped role
// composes its own; the shipped one does not disappear.
type RoleManagement int

const (
	// TenantManaged is a role a tenant composed from the application's
	// vocabulary. It is the common case, and the zero value.
	TenantManaged RoleManagement = iota
	// ApplicationManaged is a role the application ships to every tenant, the
	// way it ships permissions and scope keys.
	ApplicationManaged
)

func (m RoleManagement) String() string {
	if m == ApplicationManaged {
		return "application"
	}
	return "tenant"
}

// Revisions selects which revisions of a role a listing returns.
//
// A role is a family of revisions, so AllRevisions is the default: the reverse
// would hide history behind a flag nobody sets. LatestRevision is computed at
// read time, never stored, so nothing can go stale.
type Revisions int

const (
	AllRevisions Revisions = iota
	LatestRevision
)

// RoleFilter bounds a role listing. ID and Name are exact matches, never
// prefixes: both are flat tokens, so a prefix would be a string match inside one
// key slot rather than a structural one.
type RoleFilter struct {
	ID        string
	Name      string
	Revisions Revisions
	// Managed narrows to one kind. Nil returns both, which is what a tenant
	// administrator reading its catalog wants: the roles the application ships
	// alongside the ones the tenant composed.
	Managed *RoleManagement
	Offset  int
	Limit   int
}

// RolePage is one page of a role listing, ordered by id then revision.
//
// Total follows the selector: rows under AllRevisions, distinct roles under
// LatestRevision, so a reader can compute the page count for what it asked for.
type RolePage struct {
	Roles      []RoleContent
	Total      int
	Generation int64
}

// Team is an Auth-owned collection of explicit human members, optionally inside
// another team. It belongs to a tenant and to no application: the handbook is
// explicit that "a different application may have no department concept at all",
// and that applications keep their own business groupings separately.
//
// ID is a base-36 Snowflake. Name is a human label and is deliberately NOT
// unique. ParentID is an id rather than a name, because a name is editable and a
// hierarchy built on one would break when a team is renamed. A root team's
// ParentID is "" — the real value, not an omission.
//
// This is an internal projection, not a canonical JSON contract: P-05 lists the
// complete team and membership records as pending, and the handbook forbids
// encoding the team-parent relationship by inventing a field.
type Team struct {
	ID       string
	Name     string
	ParentID string
}

// Membership is one human's place in one team. Its identity is the pair, which
// is what makes it add-only and idempotent by construction.
//
// HumanID is issued by the auth service rather than here, and is required to be
// a base-36 Snowflake so one spelling serves every identifier in the system.
type Membership struct {
	TeamID  string
	HumanID string
}

// TeamFilter bounds a team listing. ParentID is a pointer because "" is a real
// value — the parent a root team holds — so it cannot double as "unset": nil
// lists every team, and a pointer to "" lists roots only.
// GrantFilter narrows a grant listing. Root is a tri-state: nil means every
// grant, so "which grant is this area's root" is one call rather than a scan.
type GrantFilter struct {
	Status string
	Root   *bool
	Offset int
	Limit  int
}

type GrantPage struct {
	Grants []Grant
	Total  int
}

type GrantRevisionPage struct {
	Revisions []GrantContent
	Total     int
}

// AssignmentFilter answers in both directions: a grant's recipients, or a
// recipient's grants. Exactly one is required, the rule ListMembers and
// ListInstallations already hold — an unfiltered listing is unbounded in the
// dimension that grows fastest.
type AssignmentFilter struct {
	GrantID   string
	Recipient *Recipient
	Status    string
	Offset    int
	Limit     int
}

type AssignmentPage struct {
	Assignments []Assignment
	Total       int
}

// Ownership is authority to administer a team, held by a human. Q-099: it is
// separate from the team's business authority, and being an owner supplies
// neither that authority nor permission to assign grants.
//
// It is the same shape as a Membership and deliberately a different type —
// they are two different relationships between the same two things, and Q-099
// exists to say so.
type Ownership struct {
	TeamID  string
	HumanID string
}

// OwnerFilter answers in both directions: a team's owners, or a human's teams.
// Exactly one is required, the rule ListMembers already holds.
type OwnerFilter struct {
	TeamID  string
	HumanID string
	Offset  int
	Limit   int
}

type OwnerPage struct {
	Owners []Ownership
	Total  int
}

type TeamFilter struct {
	ParentID *string
	Name     string
	Offset   int
	Limit    int
}

type TeamPage struct {
	Teams      []Team
	Total      int
	Generation int64
}

// MemberFilter bounds a membership listing, and answers in both directions: a
// team's roster, or one human's teams. Exactly one of TeamID and HumanID is
// required — an unfiltered listing of every membership is not a question anyone
// asks, and would be unbounded in the dimension that grows fastest.
type MemberFilter struct {
	TeamID  string
	HumanID string
	Offset  int
	Limit   int
}

type MemberPage struct {
	Members    []Membership
	Total      int
	Generation int64
}
type PermissionDefinition struct {
	ID     string
	Active bool
	// Boundary is where the permission was registered: `application` for one an
	// application declared, `platform` for Auth's own vocabulary.
	//
	// An application's catalog holds both, because a platform permission is
	// vocabulary every application inherits — and that union is right for
	// *evaluation*, where a request inside an application may legitimately
	// require a platform permission.
	//
	// It is wrong for a root's *ceiling*, which is why these two fields exist. A
	// root computes its coverage from the catalog, so without them an
	// application root would carry every auth:* permission — including whichever
	// one authorises establishing an application root. The thing created by the
	// authority would be able to create more of that authority.
	Boundary Boundary
	// Namespace is key3: the application for an application permission, the
	// platform's own namespace for a platform one. It is what a root's ceiling
	// is sliced by, and one rule covers both roots — a root takes the
	// permissions registered under its own namespace, which for the Auth root is
	// the platform's and for an application root is that application's.
	//
	// It cannot be recovered from the identifier. An application permission's
	// leading noun is its application and could stand in, but a platform
	// permission's leading noun is whatever the platform chose.
	Namespace string
}

// PermissionFilter bounds a catalog listing. Prefix is an administrative
// convenience for grouping identifiers; it confers no authority, and evaluation
// never matches on a prefix.
//
// Paging is by offset rather than a cursor, deliberately. A catalog is browsed
// in a UI where a reader jumps between pages rather than walking one, and a
// cursor cannot answer "page 20" without walking to it. Ordering is by the key
// slots, which is what makes an offset mean the same thing on every call.
type PermissionFilter struct {
	Prefix     string
	ActiveOnly bool
	Offset     int
	Limit      int
}

// PermissionPage is one page of a catalog listing, ordered by the key slots.
//
// Total is every record matching the filter, not the page, so a reader can
// compute how many pages exist.
//
// Generation is the application catalog's version at the moment of the read. It
// is unchanged across pages exactly when nothing was written between them, which
// is what makes an offset walk safe to cache: read a generation, page through,
// read it again, and retry if it moved. Without it an insert between two pages
// can shift a row across the boundary and the walk silently misses it.
type PermissionPage struct {
	Permissions []PermissionDefinition
	Total       int
	Generation  int64
}

// The two implicit boundaries. Neither is ever a registered scope key: an empty
// scope is already a complete scope, and $self is resolved by the evaluator
// rather than declared by a definition.
const (
	// SelfToken binds a boundary to the authorizing human. It is not the group
	// that received an assignment, and not the agent sending the request.
	SelfToken = "$self"
	// ReservedTokenPrefix marks a value as a reserved token rather than an
	// application value. $self is the only one.
	ReservedTokenPrefix = "$"
)

// ScopeDefinition is a registered boundary key. It carries nothing else: $self
// is a reserved token the evaluator knows, not something a key declares, and a
// per-key list of permitted tokens appears nowhere in the handbook.
type ScopeDefinition struct {
	Key string
}

// ScopeFilter bounds a scope listing. A scope key is a single flat token rather
// than a path, so there is no prefix filter: a prefix here would be a string
// match inside one slot, which is what the key layout exists to avoid, and a
// scope catalog is a handful of keys rather than hundreds.
type ScopeFilter struct {
	Offset int
	Limit  int
}

// ScopePage is one page of a scope listing, ordered by key. Total and
// Generation carry the same meaning as on a permission page.
type ScopePage struct {
	Scopes     []ScopeDefinition
	Total      int
	Generation int64
}
type Catalog struct {
	ApplicationID string
	Generation    int64
	Permissions   map[string]PermissionDefinition
	Scopes        map[string]ScopeDefinition
	// CompatibilityEnabled is the application's declared choice under Q-041.
	// Nothing enforces it yet: permission/scope relationship validation has no
	// approved representation (P-11), so the declaration is stored and the check
	// is unimplemented rather than implemented against an invented shape.
	CompatibilityEnabled bool
}
type GrantKey struct {
	ID       string
	Revision int64
}
type RoleKey struct {
	ID       string
	Revision int64
}
type Predicate struct{ Key, Value, SourceGrantID string }
type Route struct {
	Area          Area
	GrantID       string
	Permissions   []string
	Predicates    []Predicate
	AssignmentIDs []string
	Validities    []Validity
}
type Diagnostic struct {
	Summary string
	Route   *Route
}
type Record struct {
	Area          Area
	Kind, ID      string
	CanonicalJSON []byte
	Rows          [][]string
}
type Receipt struct{ AssignmentID string }

// FixtureContext selects a lab fixture only; it is never authenticated identity.
type FixtureContext struct{ Name string }

// ResolvedAuthority is what one human is entitled to inside one area: every
// grant that reaches them, already folded.
//
// It is the only read the enforcement side needs, and the shape is deliberately
// the canonical grant block rather than a parallel vocabulary — `scope` is
// `scope`, `permissions` is `permissions`. Two things differ from a stored
// grant, and the name says which way: these are *resolved*.
//
//   - Permissions are expanded. A grant that adopted a role arrives carrying the
//     adopted revision's permissions, because a client must not have to fetch a
//     role or pin a revision.
//   - Scope and validity are effective — folded down the whole chain — because a
//     client must never fold a chain itself. That is what keeps the root's
//     namespace slice, selected-versus-inherited permissions and
//     inherited-and-ANDed scope on this side of the boundary.
type ResolvedAuthority struct {
	Version string `json:"version"`
	// TenantID, ApplicationID and HumanID echo the three boundaries the caller
	// named, so a client can confirm it was answered about what it asked about.
	TenantID       string          `json:"tenant_id"`
	ApplicationID  string          `json:"application_id"`
	HumanID        string          `json:"human_id"`
	ResolvedGrants []ResolvedGrant `json:"resolved_grants"`
}

// ResolvedGrant is one grant as it reaches a human, with the evidence of how.
type ResolvedGrant struct {
	Version       string `json:"version"`
	GrantID       string `json:"grant_id"`
	Revision      int64  `json:"revision"`
	ParentGrantID string `json:"parent_grant_id,omitempty"`
	// Permissions and Scope are effective, not stored. See ResolvedAuthority.
	Permissions []string          `json:"permissions"`
	Scope       map[string]string `json:"scope"`
	// Validity is the narrowest window across the chain, or nil when unbounded.
	Validity *Validity `json:"validity,omitempty"`
	Source   *Source   `json:"source,omitempty"`
}

// Source is why a human holds a grant. It is annotation and never a decision
// input: a gate matches Permissions, Scope and Validity and nothing else. An
// application that reasons over a lineage takes on every rule the lineage
// follows, which is the whole reason resolution stays on this side.
type Source struct {
	AssignmentID string `json:"assignment_id"`
	TeamID       string `json:"team_id"`
	// Via is how the human reaches the holding team. Groups-only is deliberate:
	// a direct human assignment is refused at both write and read.
	Via string `json:"via"`
	// AdoptedRole is set only when the grant adopted one. Explanation, not
	// authority — Permissions above is already the resolved list.
	AdoptedRole *AdoptedRole `json:"adopted_role,omitempty"`
	// Lineage is ordered root-first, one step per contributing assignment.
	Lineage []LineageStep `json:"lineage"`
}

type AdoptedRole struct {
	RoleID   string `json:"role_id"`
	Revision int64  `json:"revision"`
}

// LineageStep is one grant in the chain, and the assignment and team that
// carried it to the next.
type LineageStep struct {
	GrantID      string `json:"grant_id"`
	Revision     int64  `json:"revision"`
	AssignmentID string `json:"assignment_id"`
	TeamID       string `json:"team_id"`
	Root         bool   `json:"root,omitempty"`
}

// ResolveOptions narrow what a resolve returns. Every zero value asks for the
// complete answer, because the complete answer is the cacheable one.
type ResolveOptions struct {
	// Permissions filters to grants carrying at least one of these. Empty means
	// everything the human holds — a gate deciding one request wants a filter, a
	// menu wants all of them, and one call serves both.
	Permissions []string
	// OmitSource drops the explanation. A gate ignores it; an audit path asks
	// for it; a bearer token must not carry it.
	OmitSource bool
}
