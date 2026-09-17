package authmiddleware

import (
	"context"
	"time"
)

type Area struct {
	TenantID      string
	ApplicationID string
}

type Actor struct {
	Type string
	ID   string
}

type Identity struct {
	Version string
	Actor   Actor
	HumanID string
}

type RequestContext struct {
	Area     Area
	Identity Identity
}

type SelectionKind uint8

const (
	SelectionExact SelectionKind = iota + 1
	SelectionAll
)

type Selection struct {
	Kind  SelectionKind
	Value string
}

type Material map[string]Selection

type Request struct {
	Context    RequestContext
	Permission string
	Material   Material
}

// AuthorityQuery names the human and the area, and nothing else. The gate asks
// what this person holds here, not whether they hold one thing: the answer then
// describes the person rather than the request, which is what makes it worth
// caching. Filtering it by permission would have made every answer single-use.
type AuthorityQuery struct {
	Context RequestContext
}

type Predicate struct {
	Key           string
	Value         string
	SourceGrantID string
}

type Route struct {
	Area    Area
	HumanID string
	// Permissions is what this route carries, not what was asked about. The
	// gate asks what a human holds rather than whether they hold one thing, so
	// an answer describes the person and can be reused; a route naming a single
	// permission could only ever have described one request.
	Permissions []string
	GrantIDs    []string
	Predicates  []Predicate
	ValidFrom   *time.Time
	ValidUntil  *time.Time
}

type Authority struct {
	Routes []Route
}

type AuthoritySource interface {
	Load(context.Context, AuthorityQuery) (Authority, error)
}

type Clock interface {
	Now() time.Time
}
