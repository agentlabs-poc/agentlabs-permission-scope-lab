package authmiddleware

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	maxRoutes          = 10000
	maxGrantsPerRoute  = 256
	maxTotalPredicates = 10000
	maxMaterialEntries = 10000
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

	authority, err := e.source.Load(ctx, AuthorityQuery{Context: request.Context, Permission: request.Permission})
	if err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if len(authority.Routes) > maxRoutes {
		return Result{}, errors.New("authority exceeds route limit")
	}

	now := e.clock.Now()
	var selected []string
	predicateCount := 0
	for i := range authority.Routes {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		route := &authority.Routes[i]
		predicateCount += len(route.Predicates)
		if predicateCount > maxTotalPredicates {
			return Result{}, errors.New("authority exceeds predicate limit")
		}
		if err := validateRoute(ctx, *route, request); err != nil {
			return Result{}, fmt.Errorf("invalid authority route %d: %w", i, err)
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

func validateRoute(ctx context.Context, route Route, request Request) error {
	if route.ValidFrom != nil && route.ValidUntil != nil && !route.ValidFrom.Before(*route.ValidUntil) {
		return errors.New("invalid route validity interval")
	}
	if err := validateArea(route.Area); err != nil || route.Area != request.Context.Area {
		return errors.New("route area does not match query")
	}
	if invalidID(route.HumanID) || route.HumanID != request.Context.Identity.HumanID {
		return errors.New("route human does not match query")
	}
	if !validPermission(route.Permission) || route.Permission != request.Permission {
		return errors.New("route permission does not match query")
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
