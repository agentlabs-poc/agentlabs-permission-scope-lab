// Package authclient answers the gate's authority question over HTTP.
//
// It is what an application imports beside the gate, and the two together are
// everything it needs: no authority records, no schema, no store. The other
// implementation of this port holds an Auth facade and opens Auth's database,
// which is why it lives on the Auth service's side and an application can never
// use it.
package authclient

import (
	"agentlabs.local/authmiddleware"
	"slices"
	"time"
)

// Version is the contract version this client speaks. A response carrying
// anything else is rejected rather than interpreted: CONTRACT-010 requires a
// consumer to reject an unsupported version and forbids guessing a default.
const Version = "1"

// identity is the Q-086 block, and it names both parties. Actor is this
// application's own credential; HumanID is the person it is asking about. They
// differ, and that is the point — an application asks about many humans and is
// none of them.
type identity struct {
	Version string `json:"version"`
	Actor   actor  `json:"actor"`
	HumanID string `json:"human_id"`
}

type actor struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type resolveRequest struct {
	Version  string   `json:"version"`
	Identity identity `json:"identity"`
	Options  options  `json:"options"`
}

// options mirror domain.ResolveOptions without importing it. The source is not
// omitted: the approved allow block requires grant_ids, and grant_ids is the
// contributing chain, which only the explanation carries.
type options struct {
	Permissions []string `json:"permissions,omitempty"`
	OmitSource  bool     `json:"omit_source,omitempty"`
}

type resolveResponse struct {
	Version        string          `json:"version"`
	TenantID       string          `json:"tenant_id"`
	ApplicationID  string          `json:"application_id"`
	HumanID        string          `json:"human_id"`
	ResolvedGrants []resolvedGrant `json:"resolved_grants"`
}

// resolvedGrant mirrors the contract completely, including the fields this
// client does not read. Decoding is strict — an unknown field means the two
// sides have drifted, and that should fail loudly rather than be ignored — so
// every field the answer may carry must be named here even when unused.
type resolvedGrant struct {
	Version       string            `json:"version"`
	GrantID       string            `json:"grant_id"`
	Revision      int64             `json:"revision"`
	ParentGrantID string            `json:"parent_grant_id,omitempty"`
	Permissions   []string          `json:"permissions"`
	Scope         map[string]string `json:"scope"`
	Validity      *validity         `json:"validity,omitempty"`
	Source        *source           `json:"source,omitempty"`
}

type validity struct {
	NotBefore *time.Time `json:"not_before,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type source struct {
	AssignmentID string        `json:"assignment_id"`
	TeamID       string        `json:"team_id"`
	Via          string        `json:"via"`
	AdoptedRole  *adoptedRole  `json:"adopted_role,omitempty"`
	Lineage      []lineageStep `json:"lineage"`
}

type adoptedRole struct {
	RoleID   string `json:"role_id"`
	Revision int64  `json:"revision"`
}

type lineageStep struct {
	GrantID      string `json:"grant_id"`
	Revision     int64  `json:"revision"`
	AssignmentID string `json:"assignment_id"`
	TeamID       string `json:"team_id"`
	Root         bool   `json:"root,omitempty"`
}

// decode turns one response into the gate's vocabulary. Every value here was
// decided on Auth's side — the permissions expanded, the scope and validity
// folded down the chain — which is what keeps the lineage rules there.
func (r resolveResponse) decode(query authmiddleware.AuthorityQuery) (authmiddleware.Authority, error) {
	if r.Version != Version {
		return authmiddleware.Authority{}, (&Error{Code: "UNSUPPORTED_VERSION", Message: "unsupported contract version " + r.Version}).evaluation()
	}
	// The three boundaries are echoed so a client can tell "nothing here" from
	// "answered about someone else". Checking them is the only reason to echo.
	if r.TenantID != query.Context.Area.TenantID || r.ApplicationID != query.Context.Area.ApplicationID {
		return authmiddleware.Authority{}, (&Error{Code: "WRONG_AREA", Message: "answered about a different area"}).evaluation()
	}
	if r.HumanID != query.Context.Identity.HumanID {
		return authmiddleware.Authority{}, (&Error{Code: "WRONG_SUBJECT", Message: "answered about a different human"}).evaluation()
	}
	routes := make([]authmiddleware.Route, 0, len(r.ResolvedGrants))
	for _, grant := range r.ResolvedGrants {
		// Each grant states its own contract version, and it was decoded and
		// never read — the same shape as the permission bug below, which was
		// also a field carried for correctness and consulted by nobody. The
		// envelope's version says how to read the envelope; a grant's says how
		// to read its scope and its validity, which are the fields that decide
		// a boundary.
		if grant.Version != Version {
			return authmiddleware.Authority{}, (&Error{
				Code:    "UNSUPPORTED_VERSION",
				Message: "a returned grant states contract version " + grant.Version,
			}).evaluation()
		}
		// The permission is checked, not assumed. Everything else the answer
		// claims is corroborated against the question — tenant, application,
		// human — and this was the exception: the one dimension that decides
		// what may be done was stamped on from the query while the grant's own
		// permissions were decoded and never read. A grant for reading the
		// directory came back approved for reading payroll.
		if !slices.Contains(grant.Permissions, query.Permission) {
			return authmiddleware.Authority{}, (&Error{
				Code:    "WRONG_PERMISSION",
				Message: "a returned grant does not carry the permission that was asked about",
			}).evaluation()
		}
		route := authmiddleware.Route{
			Area:       authmiddleware.Area{TenantID: r.TenantID, ApplicationID: r.ApplicationID},
			HumanID:    r.HumanID,
			Permission: query.Permission,
			GrantIDs:   chain(grant),
			Predicates: make([]authmiddleware.Predicate, 0, len(grant.Scope)),
		}
		for key, value := range grant.Scope {
			route.Predicates = append(route.Predicates, authmiddleware.Predicate{
				Key: key, Value: value, SourceGrantID: grant.GrantID,
			})
		}
		if grant.Validity != nil {
			route.ValidFrom, route.ValidUntil = grant.Validity.NotBefore, grant.Validity.ExpiresAt
		}
		routes = append(routes, route)
	}
	return authmiddleware.Authority{Routes: routes}, nil
}

// chain is the evidence an allow must carry. Without the explanation there is
// one grant to name — the one that reaches the human.
func chain(grant resolvedGrant) []string {
	if grant.Source == nil || len(grant.Source.Lineage) == 0 {
		return []string{grant.GrantID}
	}
	ids := make([]string, 0, len(grant.Source.Lineage))
	for _, step := range grant.Source.Lineage {
		ids = append(ids, step.GrantID)
	}
	return ids
}
