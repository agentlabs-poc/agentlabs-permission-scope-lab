// Package abv exposes area-bound inspection, diagnosis, and protected mutation.
package abv

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"agentlabs.local/abv/internal/mutation"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/sqlite"
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Facade struct {
	provider storage.Provider
	service  *mutation.Service
}

// Evidence is the immutable-by-contract snapshot supplied to an administrative
// adapter. The facade gives the adapter an isolated copy.
type Evidence = storage.Snapshot

type Administration interface {
	CheckAssignment(context.Context, Evidence, domain.Identity, domain.Assignment, time.Time) error
}

type GrantStatusAdministration interface {
	CheckGrantStatus(context.Context, Evidence, domain.Identity, domain.GrantControl, time.Time) error
}

type Clock interface{ Now() time.Time }

func New(provider storage.Provider, administration Administration, clock Clock) (*Facade, error) {
	service, err := mutation.New(provider, administration, clock)
	if err != nil {
		return nil, err
	}
	return &Facade{provider: provider, service: service}, nil
}

func OpenSQLite(ctx context.Context, path string, administration Administration, clock Clock) (*Facade, error) {
	provider, err := sqlite.Open(ctx, path)
	if err != nil {
		return nil, err
	}
	facade, err := New(provider, administration, clock)
	if err != nil {
		_ = provider.Close()
		return nil, err
	}
	return facade, nil
}

func (f *Facade) Close() error { return f.provider.Close() }

func (f *Facade) CreateAssignment(ctx context.Context, area domain.Area, identity domain.Identity, proposed domain.Assignment) (domain.Receipt, error) {
	return f.service.CreateAssignment(ctx, area, identity, proposed)
}

func (f *Facade) SetGrantStatus(ctx context.Context, area domain.Area, identity domain.Identity, proposed domain.GrantControl) (domain.GrantControl, error) {
	return f.service.SetGrantStatus(ctx, area, identity, proposed)
}

func (f *Facade) CheckAssignment(ctx context.Context, area domain.Area, raw []byte) (domain.Diagnostic, error) {
	return f.service.CheckAssignment(ctx, area, raw)
}

func (f *Facade) Inspect(ctx context.Context, area domain.Area, kind, id string) (domain.Record, error) {
	if err := area.Validate(); err != nil {
		return domain.Record{}, err
	}
	if strings.TrimSpace(id) == "" || id == "*" {
		return domain.Record{}, domain.ErrMalformed
	}
	record := domain.Record{Area: area, Kind: kind, ID: id}
	err := f.provider.Read(ctx, area, func(snapshot storage.Snapshot) error {
		if snapshot.Area != area || snapshot.Catalog.ApplicationID != area.ApplicationID() {
			return domain.ErrRejected
		}
		return inspectSnapshot(snapshot, &record)
	})
	if err != nil {
		return domain.Record{}, err
	}
	return record, nil
}

func inspectSnapshot(snapshot storage.Snapshot, record *domain.Record) error {
	switch record.Kind {
	case "assignment":
		value, ok := snapshot.Assignments[record.ID]
		if !ok || value.ID != record.ID {
			return domain.ErrNotFound
		}
		raw, err := json.Marshal(value)
		if err != nil {
			return domain.ErrMalformed
		}
		decoded, err := codec.DecodeAssignment(raw)
		if err != nil || decoded != value {
			return domain.ErrMalformed
		}
		record.CanonicalJSON = raw
	case "grant":
		var selected domain.GrantContent
		found := false
		for key, value := range snapshot.Contents {
			if key.ID != value.GrantID || key.Revision != value.Revision {
				return domain.ErrMalformed
			}
			if key.ID == record.ID && (!found || key.Revision > selected.Revision) {
				selected, found = value, true
			}
		}
		if !found {
			return domain.ErrNotFound
		}
		raw, err := json.Marshal(selected)
		if err != nil {
			return domain.ErrMalformed
		}
		decoded, err := codec.DecodeContent(raw)
		if err != nil || !reflect.DeepEqual(decoded, selected) {
			return domain.ErrMalformed
		}
		record.CanonicalJSON = raw
	case "grant-control":
		value, ok := snapshot.Controls[record.ID]
		if !ok {
			return domain.ErrNotFound
		}
		if value.ID != record.ID || value.Version != "1" || (value.Status != "enabled" && value.Status != "disabled") {
			return domain.ErrMalformed
		}
		raw, err := json.Marshal(value)
		if err != nil {
			return domain.ErrMalformed
		}
		record.CanonicalJSON = raw
	case "permission":
		value, ok := snapshot.Catalog.Permissions[record.ID]
		if !ok || value.ID != record.ID {
			return domain.ErrNotFound
		}
		record.Rows = [][]string{{"field", "value"}, {"active", strconv.FormatBool(value.Active)}}
	case "scope":
		value, ok := snapshot.Catalog.Scopes[record.ID]
		if !ok || value.Key != record.ID {
			return domain.ErrNotFound
		}
		record.Rows = [][]string{{"allowed_token"}}
		for _, token := range value.AllowedTokens {
			record.Rows = append(record.Rows, []string{token})
		}
	case "role":
		record.Rows = [][]string{{"revision", "permissions"}}
		for key, value := range snapshot.Roles {
			if key.ID != value.ID || key.Revision != value.Revision {
				return domain.ErrMalformed
			}
			if key.ID == record.ID {
				record.Rows = append(record.Rows, []string{strconv.FormatInt(key.Revision, 10), strings.Join(value.Permissions, ",")})
			}
		}
		sort.Slice(record.Rows[1:], func(i, j int) bool { return record.Rows[i+1][0] < record.Rows[j+1][0] })
		if len(record.Rows) == 1 {
			return domain.ErrNotFound
		}
	case "team":
		value, ok := snapshot.Teams[record.ID]
		if !ok || value.ID != record.ID {
			return domain.ErrNotFound
		}
		record.Rows = [][]string{{"field", "value"}, {"parent_id", value.ParentID}}
	case "membership":
		record.Rows = [][]string{{"team_id", "human_id"}}
		for _, value := range snapshot.Memberships {
			if value.TeamID == record.ID || value.HumanID == record.ID {
				record.Rows = append(record.Rows, []string{value.TeamID, value.HumanID})
			}
		}
		if len(record.Rows) == 1 {
			return domain.ErrNotFound
		}
	default:
		return domain.ErrUnsupported
	}
	return nil
}
