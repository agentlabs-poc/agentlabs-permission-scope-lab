// Package localadapter connects the middleware evaluator to the bounded ABV
// SQLite prototype for local, in-process testing.
package localadapter

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	storageSQLite "agentlabs.local/abv/internal/storage/sqlite"
	"agentlabs.local/authmiddleware"
	"context"
	"errors"
	"reflect"
	"time"
)

type SQLiteAuthoritySource struct {
	reader *storageSQLite.Reader
	clock  authmiddleware.Clock
}

func Open(ctx context.Context, path string, clock authmiddleware.Clock) (*SQLiteAuthoritySource, error) {
	if ctx == nil || nilInterface(clock) {
		return nil, errors.New("context and clock are required")
	}
	reader, err := storageSQLite.OpenReadOnly(ctx, path)
	if err != nil {
		return nil, err
	}
	return &SQLiteAuthoritySource{reader: reader, clock: clock}, nil
}

func (s *SQLiteAuthoritySource) Load(ctx context.Context, query authmiddleware.AuthorityQuery) (authmiddleware.Authority, error) {
	if ctx == nil {
		return authmiddleware.Authority{}, errors.New("context is required")
	}
	area, err := domain.NewArea(query.Context.Area.TenantID, query.Context.Area.ApplicationID)
	if err != nil {
		return authmiddleware.Authority{}, err
	}
	identity := domain.Identity{
		Version: query.Context.Identity.Version,
		Actor:   domain.Actor{Type: query.Context.Identity.Actor.Type, ID: query.Context.Identity.Actor.ID},
		HumanID: query.Context.Identity.HumanID,
	}
	var authority authmiddleware.Authority
	// ponytail: the bounded full snapshot is the prototype ceiling; replace it
	// with indexed authority queries when measured database size or latency needs it.
	err = s.reader.Read(ctx, area, func(snapshot storage.Snapshot) error {
		routes, err := lineage.ResolveHuman(ctx, snapshot, identity, query.Permission, s.clock.Now())
		if err != nil {
			return err
		}
		converted := make([]authmiddleware.Route, len(routes))
		for i, route := range routes {
			converted[i], err = convertRoute(route, snapshot.Assignments, query)
			if err != nil {
				return err
			}
		}
		authority.Routes = converted
		return nil
	})
	if err != nil {
		return authmiddleware.Authority{}, err
	}
	return authority, nil
}

func (s *SQLiteAuthoritySource) Close() error { return s.reader.Close() }

func convertRoute(route domain.Route, assignments map[string]domain.Assignment, query authmiddleware.AuthorityQuery) (authmiddleware.Route, error) {
	result := authmiddleware.Route{
		Area: authmiddleware.Area{TenantID: route.Area.TenantID(), ApplicationID: route.Area.ApplicationID()}, HumanID: query.Context.Identity.HumanID, Permission: query.Permission,
		GrantIDs: make([]string, len(route.AssignmentIDs)), Predicates: make([]authmiddleware.Predicate, len(route.Predicates)),
	}
	for i, id := range route.AssignmentIDs {
		assignment, ok := assignments[id]
		if !ok || assignment.ID != id {
			return authmiddleware.Route{}, errors.New("missing contributing assignment")
		}
		result.GrantIDs[i] = assignment.GrantID
	}
	for i, predicate := range route.Predicates {
		result.Predicates[i] = authmiddleware.Predicate{Key: predicate.Key, Value: predicate.Value, SourceGrantID: predicate.SourceGrantID}
	}
	for _, validity := range route.Validities {
		if validity.NotBefore != nil && (result.ValidFrom == nil || result.ValidFrom.Before(*validity.NotBefore)) {
			result.ValidFrom = copyTime(validity.NotBefore)
		}
		if validity.ExpiresAt != nil && (result.ValidUntil == nil || validity.ExpiresAt.Before(*result.ValidUntil)) {
			result.ValidUntil = copyTime(validity.ExpiresAt)
		}
	}
	return result, nil
}

func copyTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
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
