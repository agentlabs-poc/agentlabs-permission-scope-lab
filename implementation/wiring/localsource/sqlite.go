// Package localsource answers the gate's authority question from Auth-AL's own
// store, in the same process.
//
// It sits on the Auth service's side of the boundary, with the composition root,
// because that is where its dependencies are: it holds an *abv.Facade and opens
// Auth's database. It lived inside abv and was mistaken for the client — an
// application can never use it, whatever directory it is in. What an application
// imports is the HTTP source, which needs none of this.
package localsource

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"agentlabs.local/authmiddleware"
	"context"
	"errors"
	"reflect"
	"slices"
	"time"
)

type SQLiteAuthoritySource struct {
	authority  *abv.Facade
	credential domain.Actor
}

// Open takes a registry because the store no longer answers whether a tenant
// holds an application — that fact moved to the application registry domain.
//
// It takes an administration for a newer reason: asking what a human is
// entitled to is a gated read, and the gate belongs to whoever deploys this
// rather than to the adapter. Until Facade.ResolveAuthority existed there was
// nothing to gate, because this package walked the lineage itself through
// Auth-AL's internal packages — an enforcement path with no gate at all.
//
// It takes a credential because an agent is not the person it asks about. The
// request carries the human; this names the application asking, and the two are
// different parties — which is what the identity block has always been shaped to
// say and what this path could not express until resolution admitted an actor
// that is not its subject.
//
// The store is opened read-only. An agent resolving authority on every request
// has no business holding a writable handle on the tenant's authority.
func Open(ctx context.Context, path string, credential domain.Actor, administration abv.Administration, clock abv.Clock, registry abv.Registry) (*SQLiteAuthoritySource, error) {
	if ctx == nil || nilInterface(clock) || nilInterface(administration) {
		return nil, errors.New("context, administration and clock are required")
	}
	if credential.Type == "" || credential.ID == "" {
		return nil, errors.New("a credential is required: an agent asks as itself, not as the human")
	}
	authority, err := abv.OpenSQLiteReadOnly(ctx, path, administration, clock, registry)
	if err != nil {
		return nil, err
	}
	return &SQLiteAuthoritySource{authority: authority, credential: credential}, nil
}

// Load answers the evaluator's one question through Auth-AL's one public read.
//
// Everything this used to do between the store and the answer — mapping each
// contributing assignment to its grant through the raw snapshot, folding the
// chain's validity windows to the narrowest — is what ResolveAuthority already
// does. Keeping a second copy here meant two foldings of the same rules, and
// only one of them was the one Auth-AL tests.
func (s *SQLiteAuthoritySource) Load(ctx context.Context, query authmiddleware.AuthorityQuery) (authmiddleware.Authority, error) {
	if ctx == nil {
		return authmiddleware.Authority{}, errors.New("context is required")
	}
	area, err := domain.NewArea(query.Context.Area.TenantID, query.Context.Area.ApplicationID)
	if err != nil {
		return authmiddleware.Authority{}, err
	}
	// The agent asks as itself about the request's human. The actor on the
	// request is the caller of the *application*, which is the application's
	// business and not Auth's — what Auth is asked is "what does this human
	// hold", by a credential entitled to ask it.
	identity := domain.Identity{
		Version: "1",
		Actor:   s.credential,
		HumanID: query.Context.Identity.HumanID,
	}
	// No permission filter. The gate asks what this human holds here, not
	// whether they hold one thing, so the answer describes the person and could
	// be cached against them; narrowing it to the permission in hand would have
	// made every answer single-use.
	//
	// The source is *not* omitted, and that corrects an assumption. "A decision
	// never reads the explanation" is true of the decision and false of the
	// result: the approved allow block requires grant_ids, and grant_ids is the
	// contributing chain, which only the lineage carries. So the one caller the
	// OmitSource option was written for is the one that cannot use it.
	resolved, err := s.authority.ResolveAuthority(ctx, area, identity, domain.ResolveOptions{})
	if err != nil {
		// Reported as an evaluation failure, because that is what an
		// AuthoritySource owes its caller. A bare error reaches the application
		// as an unrecognised one and gets rendered as the caller's own bad
		// request — no availability signal, and the blame in the wrong place.
		return authmiddleware.Authority{}, &authmiddleware.EvaluationError{
			Version: "1", Code: "AUTHORITY_UNAVAILABLE",
			Message:       "We could not check your access.",
			MessageReason: "the authority store could not answer",
			Cause:         err,
		}
	}
	return routesFor(resolved, query)
}

// routesFor turns one answer into the evaluator's vocabulary, checking what it
// is allowed to assume on the way.
func routesFor(resolved domain.ResolvedAuthority, query authmiddleware.AuthorityQuery) (authmiddleware.Authority, error) {
	// The same four things the HTTP source checks, for the same reasons. Both
	// implementations answer one contract, and an invariant enforced on one side
	// and trusted on the other is how the two drift — this side was the trusting
	// one. The permission corroboration that used to sit beside these is gone
	// from both sides: the answer is not filtered by permission any more, so a
	// grant that does not carry the one in hand is ordinary.
	//
	// Nothing in-process can produce these today: ResolveAuthority answers about
	// the area and human it was asked about. That is what makes it defence in
	// depth rather than a fix, and also what makes it worth having — the day this source is given a different implementation, or
	// the facade grows a cache, the assumptions are written down instead of
	// remembered.
	if resolved.Version != "1" {
		return failed("UNSUPPORTED_VERSION", "the answer states contract version "+resolved.Version)
	}
	if resolved.TenantID != query.Context.Area.TenantID || resolved.ApplicationID != query.Context.Area.ApplicationID {
		return failed("WRONG_AREA", "the answer describes a different area")
	}
	if resolved.HumanID != query.Context.Identity.HumanID {
		return failed("WRONG_SUBJECT", "the answer describes a different human")
	}
	routes := make([]authmiddleware.Route, len(resolved.ResolvedGrants))
	for i, grant := range resolved.ResolvedGrants {
		// A grant states its own contract version, and it is that version which
		// says how to read the grant's scope and validity — the two fields that
		// decide a boundary.
		if grant.Version != "1" {
			return failed("UNSUPPORTED_VERSION", "a resolved grant states contract version "+grant.Version)
		}
		routes[i] = convertGrant(resolved, grant)
	}
	return authmiddleware.Authority{Routes: routes}, nil
}

// failed reports an answer this source cannot use. Every one of these is an
// evaluation failure and never a denial: the routes came from the store, not
// from the request, so nothing the caller sent can produce one.
func failed(code, reason string) (authmiddleware.Authority, error) {
	return authmiddleware.Authority{}, &authmiddleware.EvaluationError{
		Version: "1", Code: code,
		Message:       "We could not check your access.",
		MessageReason: reason,
	}
}

func (s *SQLiteAuthoritySource) Close() error { return s.authority.Close() }

// convertGrant renames a resolved grant into the evaluator's vocabulary. It is
// a translation and nothing more: every value here was decided on Auth-AL's side
// of the boundary, which is what keeps the lineage rules there.
func convertGrant(resolved domain.ResolvedAuthority, grant domain.ResolvedGrant) authmiddleware.Route {
	route := authmiddleware.Route{
		// The answer's boundaries, not the query's. They cannot disagree — the
		// area asked for is the area read — but taking them from the answer is
		// why the envelope echoes them, and it keeps a route describing where it
		// actually came from rather than where it was asked for.
		Area:    authmiddleware.Area{TenantID: resolved.TenantID, ApplicationID: resolved.ApplicationID},
		HumanID: resolved.HumanID,
		// The grant's own permissions, not the one that was asked about. Stamping
		// the query's permission on was how a grant that did not carry it became
		// a route that claimed it.
		Permissions: slices.Clone(grant.Permissions),
		GrantIDs:    grantChain(grant),
		Predicates:  make([]authmiddleware.Predicate, 0, len(grant.Scope)),
	}
	for key, value := range grant.Scope {
		route.Predicates = append(route.Predicates, authmiddleware.Predicate{
			Key: key, Value: value, SourceGrantID: grant.GrantID,
		})
	}
	if grant.Validity != nil {
		route.ValidFrom = copyTime(grant.Validity.NotBefore)
		route.ValidUntil = copyTime(grant.Validity.ExpiresAt)
	}
	return route
}

// grantChain is the evidence the result contract requires on an allow. Without
// the explanation there is one grant to name — the one that reaches the human —
// and naming it is what the approved deny/allow blocks call grant_ids.
func grantChain(grant domain.ResolvedGrant) []string {
	// An empty lineage is the same as no explanation at all: the grant that
	// reaches the human is the one to name. Without this an allow could carry
	// zero grant_ids, which Q-066 forbids — validateRoute then refuses the route
	// and a human with real authority is answered "we could not check". The HTTP
	// source guards both conditions; this one guarded only the first.
	if grant.Source == nil || len(grant.Source.Lineage) == 0 {
		return []string{grant.GrantID}
	}
	chain := make([]string, 0, len(grant.Source.Lineage))
	for _, step := range grant.Source.Lineage {
		chain = append(chain, step.GrantID)
	}
	return chain
}

func copyTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
