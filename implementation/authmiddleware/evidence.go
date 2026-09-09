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

type AuthorityQuery struct {
	Context    RequestContext
	Permission string
}

type Predicate struct {
	Key           string
	Value         string
	SourceGrantID string
}

type Route struct {
	Area       Area
	HumanID    string
	Permission string
	GrantIDs   []string
	Predicates []Predicate
	ValidFrom  *time.Time
	ValidUntil *time.Time
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
