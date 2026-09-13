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
type RoleContent struct {
	ID          string
	Revision    int64
	Permissions []string
}
type Team struct{ ID, ParentID string }
type Membership struct{ TeamID, HumanID string }
type PermissionDefinition struct {
	ID     string
	Active bool
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
	ApplicationID        string
	Generation           int64
	Permissions          map[string]PermissionDefinition
	Scopes               map[string]ScopeDefinition
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
