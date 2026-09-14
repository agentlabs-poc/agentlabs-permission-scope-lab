package domain

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// slugPattern is the shape an application id must have, matching what
// agentlabs-auth's registry already enforces on its own slugs.
//
// The reserved-name list — system, platform, auth, admin — is deliberately NOT
// here. That list is the auth service's policy about its own namespaces, and
// holding a copy would mean this domain drifting out of step with it.
var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,63}$`)

// Application statuses. The real registry has five; the other three describe a
// publishing lifecycle this domain does not yet have, and inventing them here
// would be designing ahead of need.
const (
	StatusActive    = "active"
	StatusSuspended = "suspended"
)

// Application is a registered capability surface: the thing that owns a
// permission catalog somewhere else.
//
// Slug is the id, and it is a slug rather than a generated identifier for one
// reason: an application id appears in the canonical path of every record that
// application owns, in every domain. A Snowflake there would make every path
// unreadable, and unlike a role there is no separate name to carry the meaning.
type Application struct {
	Slug   string
	Name   string
	Status string
}

// Installation is one tenant holding one application. Its identity is the pair
// and its presence is the fact — the shape a membership holds, for the same
// reason: a relationship has no id of its own.
type Installation struct {
	TenantID string
	Slug     string
}

type ApplicationFilter struct {
	Status string
	Offset int
	Limit  int
}

type ApplicationPage struct {
	Applications []Application
	Total        int
}

// InstallationFilter answers in both directions: a tenant's applications, or an
// application's tenants. Exactly one of TenantID and Slug is required — an
// unfiltered listing of every installation is unbounded in the dimension that
// grows fastest, and is not a question anyone asks.
type InstallationFilter struct {
	TenantID string
	Slug     string
	Offset   int
	Limit    int
}

type InstallationPage struct {
	Installations []Installation
	Total         int
}

// Identity is the acting human. The registry validates its shape and nothing
// more: who may act is the injected administration's decision.
type Identity struct {
	Version string
	HumanID string
}

func ValidSlug(slug string) bool { return slugPattern.MatchString(slug) }

func ValidName(name string) bool {
	return strings.TrimSpace(name) != "" && utf8.ValidString(name) && len(name) <= 255
}

func ValidStatus(status string) bool {
	return status == StatusActive || status == StatusSuspended
}

func ValidTenant(tenantID string) bool {
	return strings.TrimSpace(tenantID) != "" && utf8.ValidString(tenantID) && !strings.Contains(tenantID, "*")
}

func (i Identity) Valid() bool {
	return i.Version == "1" && strings.TrimSpace(i.HumanID) != "" && utf8.ValidString(i.HumanID)
}
