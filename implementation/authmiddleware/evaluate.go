package authmiddleware

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

const (
	maxRoutes         = 10000
	maxGrantsPerRoute = 256
	// Permissions are bounded across the whole answer rather than per route,
	// because the widest legitimate route in this system is produced by the
	// system itself: a root grant's permissions are the application's entire
	// active catalog, recomputed at resolve time. A per-route ceiling of 256
	// locked every root holder out of the whole application with a 503 — the
	// lab's own demonstrations run to 603 permissions in one application — and
	// it did so for requests an entirely different grant authorized.
	maxTotalPermissions = 1 << 20
	maxTotalPredicates  = 10000
	maxMaterialEntries  = 10000
)

type Evaluator struct {
	source AuthoritySource
	clock  Clock
}

func New(source AuthoritySource, clock Clock) (*Evaluator, error) {
	if nilInterface(source) || nilInterface(clock) {
		return nil, errors.New("authority source and clock are required")
	}
	return &Evaluator{source: source, clock: clock}, nil
}

func (e *Evaluator) Evaluate(ctx context.Context, request Request) (Result, error) {
	if ctx == nil {
		return Result{}, errors.New("context is required")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if err := validateRequest(ctx, request); err != nil {
		return Result{}, err
	}

	authority, err := e.source.Load(ctx, AuthorityQuery{Context: request.Context})
	if err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if len(authority.Routes) > maxRoutes {
		return Result{}, unusableAuthority(errors.New("authority exceeds route limit"))
	}

	now := e.clock.Now()
	var selected []string
	predicateCount, permissionCount := 0, 0
	unusable := false
	for i := range authority.Routes {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		route := &authority.Routes[i]
		predicateCount += len(route.Predicates)
		if predicateCount > maxTotalPredicates {
			return Result{}, unusableAuthority(errors.New("authority exceeds predicate limit"))
		}
		permissionCount += len(route.Permissions)
		if permissionCount > maxTotalPermissions {
			return Result{}, unusableAuthority(errors.New("authority exceeds permission limit"))
		}
		// An answer that describes the wrong person, or the wrong area, is not
		// an answer about this request at all, and no part of it can be trusted.
		if err := validateAnswer(*route, request); err != nil {
			return Result{}, unusableAuthority(fmt.Errorf("invalid authority route %d: %w", i, err))
		}
		// Everything else is about one route, and costs one route — but it is
		// remembered, because it changes what a deny would mean.
		//
		// This used to fail the whole evaluation outright. That was defensible
		// while the question named a permission, since every route in the answer
		// was then about this request. It is wrong now that the answer describes
		// everything the human holds here: one unusable grant took away every
		// other grant they had, which authority-lineage.md forbids in terms —
		// "missing support stops the affected authority route, not necessarily
		// all authority of that user or group" — and which this codebase had
		// already fixed one layer up, in lineage's routeScoped skip.
		if err := validateRoute(ctx, *route, request); err != nil {
			if ctx.Err() != nil {
				return Result{}, ctx.Err()
			}
			unusable = true
			continue
		}
		if route.ValidFrom != nil && now.Before(*route.ValidFrom) || route.ValidUntil != nil && !now.Before(*route.ValidUntil) {
			continue
		}
		matches, err := routeMatches(ctx, *route, request)
		if err != nil {
			return Result{}, err
		}
		if matches && (selected == nil || tupleLess(route.GrantIDs, selected)) {
			selected = route.GrantIDs
		}
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	// An allow stands on its own route: Q-051 / DECISION-003 requires only that
	// some complete valid route authorizes, and a different route being unusable
	// says nothing about this one.
	//
	// A deny does not. The same table qualifies it — "one failed grant alone does
	// not establish this if another route could authorize it" — and a route this
	// gate could not read is exactly a route that might have. So a deny reached
	// with an unusable route in the answer is not a completed decision; it is a
	// failure to establish, and the person is told we could not check rather than
	// that they have no access.
	if selected == nil && unusable {
		return Result{}, unusableAuthority(errors.New("every authorizing route was unusable"))
	}
	if selected == nil {
		return Result{
			Version:            "1",
			Decision:           Deny,
			ErrorCode:          "NO_AUTHORIZING_GRANT",
			ErrorMessage:       "You do not have access to this resource.",
			ErrorMessageReason: "No grant authorizes this operation within the requested boundary.",
		}, nil
	}
	return Result{Version: "1", Decision: Allow, GrantIDs: append([]string(nil), selected...)}, nil
}

func validateRequest(ctx context.Context, request Request) error {
	if err := validateArea(request.Context.Area); err != nil {
		return err
	}
	identity := request.Context.Identity
	if identity.Version != "1" || identity.Actor.Type != "user" || invalidID(identity.Actor.ID) || invalidID(identity.HumanID) || identity.Actor.ID != identity.HumanID {
		return errors.New("invalid direct-human identity")
	}
	if !validPermission(request.Permission) {
		return errors.New("invalid request permission")
	}
	if len(request.Material) > maxMaterialEntries {
		return errors.New("request material exceeds entry limit")
	}
	for key, selection := range request.Material {
		if err := ctx.Err(); err != nil {
			return err
		}
		if invalidText(key) || strings.Contains(key, "*") {
			return errors.New("invalid material key")
		}
		switch selection.Kind {
		case SelectionExact:
			if invalidText(selection.Value) || strings.Contains(selection.Value, "*") {
				return errors.New("invalid exact selection")
			}
		case SelectionAll:
			if selection.Value != "" {
				return errors.New("all selection cannot have a value")
			}
		default:
			return errors.New("unsupported selection kind")
		}
	}
	return nil
}

func validateArea(area Area) error {
	if invalidID(area.TenantID) || invalidID(area.ApplicationID) {
		return errors.New("invalid area")
	}
	return nil
}

// unusableAuthority names an answer this gate cannot work with.
//
// It is an evaluation failure and never the caller's fault, which is the whole
// distinction Q-051 / DECISION-003 draws: "report inability to complete
// evaluation as a separate evaluation error, not a third authorization
// decision". Returned as a plain error it reached the application
// as "your request was bad": a person would retry a request that was never the
// problem, and an operator would never learn that the authority service is
// answering nonsense. The routes here came from the source, not from the
// request — nothing the caller sent can produce one.
func unusableAuthority(cause error) *EvaluationError {
	return &EvaluationError{
		Version:       "1",
		Code:          "AUTHORITY_MALFORMED",
		Message:       "We could not check your access.",
		MessageReason: "the authority service returned an answer this gate cannot use",
		Cause:         cause,
	}
}

// validateAnswer holds a route to the question that was asked. A route about
// another area or another human means the source answered about somebody else,
// and that is a statement about the whole answer rather than about one route —
// so it is the one validation that still fails the evaluation outright.
func validateAnswer(route Route, request Request) error {
	if err := validateArea(route.Area); err != nil || route.Area != request.Context.Area {
		return errors.New("route area does not match query")
	}
	if invalidID(route.HumanID) || route.HumanID != request.Context.Identity.HumanID {
		return errors.New("route human does not match query")
	}
	return nil
}

// validateRoute checks what one route needs to be usable. A failure costs that
// route and nothing else — see the call site.
//
// Nothing here validates the *spelling* of a permission. An answer about a
// person names everything they hold, and the request's own permission is
// already validated, so a permission string that is not a permission simply
// cannot equal it. Rejecting the route for carrying one bought nothing and cost
// every other grant the person held, because the two sides do not agree on the
// grammar: Auth-AL's codec accepts a permission with a space in it and this
// package does not.
func validateRoute(ctx context.Context, route Route, request Request) error {
	if route.ValidFrom != nil && route.ValidUntil != nil && !route.ValidFrom.Before(*route.ValidUntil) {
		return errors.New("invalid route validity interval")
	}
	if len(route.Permissions) == 0 {
		return errors.New("route carries no permission")
	}
	if len(route.GrantIDs) == 0 || len(route.GrantIDs) > maxGrantsPerRoute {
		return errors.New("invalid contributing grant count")
	}
	grants := make(map[string]struct{}, len(route.GrantIDs))
	for _, id := range route.GrantIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if invalidID(id) {
			return errors.New("invalid contributing grant ID")
		}
		if _, duplicate := grants[id]; duplicate {
			return errors.New("duplicate contributing grant ID")
		}
		grants[id] = struct{}{}
	}
	for _, predicate := range route.Predicates {
		if err := ctx.Err(); err != nil {
			return err
		}
		if invalidText(predicate.Key) || strings.Contains(predicate.Key, "*") || invalidText(predicate.Value) || invalidID(predicate.SourceGrantID) {
			return errors.New("invalid predicate")
		}
		if strings.HasPrefix(predicate.Value, "$") && predicate.Value != "$self" || strings.Contains(predicate.Value, "*") {
			return errors.New("unsupported predicate value")
		}
		if _, ok := grants[predicate.SourceGrantID]; !ok {
			return errors.New("predicate source is not a contributing grant")
		}
	}
	return nil
}

func routeMatches(ctx context.Context, route Route, request Request) (bool, error) {
	// The permission first, because a route that does not carry it cannot
	// authorize this request whatever its predicates say. This used to be part
	// of validating the route, when the answer was filtered to one permission
	// and anything else meant the source had answered the wrong question.
	if !slices.Contains(route.Permissions, request.Permission) {
		return false, nil
	}
	for _, predicate := range route.Predicates {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		selection, ok := request.Material[predicate.Key]
		if !ok || selection.Kind != SelectionExact {
			return false, nil
		}
		value := predicate.Value
		if value == "$self" {
			value = request.Context.Identity.HumanID
		}
		if selection.Value != value {
			return false, nil
		}
	}
	return true, nil
}

func tupleLess(left, right []string) bool {
	for i := 0; i < len(left) && i < len(right); i++ {
		if left[i] != right[i] {
			return left[i] < right[i]
		}
	}
	return len(left) < len(right)
}

func invalidID(value string) bool {
	return invalidText(value) || strings.Contains(value, "*")
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	kind := reflect.ValueOf(value).Kind()
	return (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface || kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice) && reflect.ValueOf(value).IsNil()
}
