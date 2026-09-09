package mutation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type revisionAdmin struct {
	check  func(storage.Snapshot, domain.Identity, domain.GrantContent) error
	source *string
}

func (a revisionAdmin) CheckAssignment(context.Context, storage.Snapshot, domain.Identity, domain.Assignment, time.Time) error {
	return nil
}
func (a revisionAdmin) CheckGrantRevisionPublication(_ context.Context, s storage.Snapshot, i domain.Identity, source string, g domain.GrantContent, _ time.Time) error {
	if a.source != nil {
		*a.source = source
	}
	if a.check != nil {
		return a.check(s, i, g)
	}
	return nil
}

type revisionProvider struct {
	snapshot storage.Snapshot
	err      error
	writes   int
}

func (p *revisionProvider) Read(_ context.Context, _ domain.Area, cb func(storage.Snapshot) error) error {
	return cb(p.snapshot)
}
func (p *revisionProvider) Close() error { return nil }
func (p *revisionProvider) Update(_ context.Context, _ domain.Area, cb func(storage.Snapshot) (storage.WriteSet, error)) error {
	w, err := cb(p.snapshot)
	if err != nil {
		return err
	}
	if p.err != nil {
		return p.err
	}
	if w.NewGrantRevision != nil {
		p.writes++
	}
	return nil
}

func revisionCandidate() domain.GrantContent {
	return domain.GrantContent{Version: "1", GrantID: "G2", Revision: 2, ParentGrantID: "G1", Permissions: []string{"read"}, Scope: map[string]string{"cert": "C17"}}
}

func revisionFixture(area domain.Area) (storage.Snapshot, domain.Identity) {
	g0 := domain.GrantContent{Version: "1", GrantID: "G0", Revision: 1, Permissions: []string{"read", "delete"}, Scope: map[string]string{}}
	g1 := domain.GrantContent{Version: "1", GrantID: "G1", Revision: 1, ParentGrantID: "G0", Permissions: []string{"read"}, Scope: map[string]string{"dept": "FIN"}}
	g2 := domain.GrantContent{Version: "1", GrantID: "G2", Revision: 1, ParentGrantID: "G1", Permissions: []string{"read"}, Scope: map[string]string{"cert": "C17"}}
	return storage.Snapshot{
		Area:        area,
		Catalog:     domain.Catalog{ApplicationID: area.ApplicationID(), Permissions: map[string]domain.PermissionDefinition{"read": {ID: "read", Active: true}, "delete": {ID: "delete", Active: true}}, Scopes: map[string]domain.ScopeDefinition{"dept": {Key: "dept"}, "cert": {Key: "cert"}, "user": {Key: "user", AllowedTokens: []string{"$self"}}}},
		Controls:    map[string]domain.GrantControl{"G0": {Version: "1", ID: "G0", Status: "enabled"}, "G1": {Version: "1", ID: "G1", Status: "enabled"}, "G2": {Version: "1", ID: "G2", Status: "enabled"}},
		Contents:    map[domain.GrantKey]domain.GrantContent{{ID: "G0", Revision: 1}: g0, {ID: "G1", Revision: 1}: g1, {ID: "G2", Revision: 1}: g2},
		Assignments: map[string]domain.Assignment{"A0": {Version: "1", ID: "A0", GrantID: "G0", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "Root"}, Status: "enabled"}, "A1": {Version: "1", ID: "A1", GrantID: "G1", GrantRevision: 1, Recipient: domain.Recipient{Type: "group", ID: "Team1"}, Status: "enabled"}},
		Roles:       map[domain.RoleKey]domain.RoleContent{{ID: "reader", Revision: 1}: {ID: "reader", Revision: 1, Permissions: []string{"read"}}},
		Teams:       map[string]domain.Team{"Root": {ID: "Root"}, "Team1": {ID: "Team1", ParentID: "Root"}},
		Memberships: []domain.Membership{{TeamID: "Team1", HumanID: "maya"}}, TrustedRoots: map[string]bool{"G0": true},
	}, domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "maya"}, HumanID: "maya"}
}

func TestPublishGrantRevisionUsesSeparateAdministrationAndActualSource(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	snapshot, issuer := revisionFixture(area)
	candidate := revisionCandidate()
	p := &revisionProvider{snapshot: snapshot}
	var adminSource string
	admin := revisionAdmin{check: func(s storage.Snapshot, gotID domain.Identity, got domain.GrantContent) error {
		if gotID != issuer || !reflect.DeepEqual(got, candidate) {
			return domain.ErrRejected
		}
		s.Catalog.Permissions["read"] = domain.PermissionDefinition{}
		s.Contents[domain.GrantKey{ID: "G1", Revision: 1}] = domain.GrantContent{}
		got.Permissions[0] = "forged"
		got.Scope["cert"] = "forged"
		return nil
	}, source: &adminSource}
	s, _ := New(p, admin, fixedClock{now: time.Now()})
	got, err := s.PublishGrantRevision(t.Context(), area, issuer, "A1", candidate)
	if err != nil || adminSource != "A1" || !reflect.DeepEqual(got, candidate) || p.writes != 1 || candidate.Permissions[0] != "read" || candidate.Scope["cert"] != "C17" {
		t.Fatalf("got=%#v writes=%d input=%#v err=%v", got, p.writes, candidate, err)
	}
}

func TestPublishGrantRevisionRejectsInvalidAuthorityAndContentWithoutWrite(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	wrongArea, _ := domain.NewArea("other", "hrms")
	valid := revisionCandidate()
	for _, tc := range []struct {
		name     string
		mutate   func(*storage.Snapshot, *domain.GrantContent)
		area     domain.Area
		identity domain.Identity
		source   string
		admin    Administration
		perr     error
		want     error
	}{
		{name: "missing admin", admin: oldAdmin{}, want: domain.ErrUnsupported},
		{name: "wrong area", area: wrongArea, admin: revisionAdmin{}, want: domain.ErrRejected},
		{name: "bad identity", identity: domain.Identity{}, admin: revisionAdmin{}, want: domain.ErrMalformed},
		{name: "bad source", source: " ", admin: revisionAdmin{}, want: domain.ErrMalformed},
		{name: "missing grant", admin: revisionAdmin{}, mutate: func(_ *storage.Snapshot, g *domain.GrantContent) { g.GrantID = "missing" }, want: domain.ErrRejected},
		{name: "root", admin: revisionAdmin{}, mutate: func(_ *storage.Snapshot, g *domain.GrantContent) { g.GrantID, g.ParentGrantID = "G0", "" }, want: domain.ErrRejected},
		{name: "changed parent", admin: revisionAdmin{}, mutate: func(_ *storage.Snapshot, g *domain.GrantContent) { g.ParentGrantID = "G0" }, want: domain.ErrRejected},
		{name: "duplicate revision", admin: revisionAdmin{}, mutate: func(_ *storage.Snapshot, g *domain.GrantContent) { g.Revision = 1 }, want: domain.ErrConflict},
		{name: "lower revision", admin: revisionAdmin{}, mutate: func(s *storage.Snapshot, g *domain.GrantContent) {
			x := s.Contents[domain.GrantKey{ID: "G2", Revision: 1}]
			x.Revision = 3
			s.Contents[domain.GrantKey{ID: "G2", Revision: 3}] = x
			g.Revision = 2
		}, want: domain.ErrConflict},
		{name: "unrelated source", source: "A0", admin: revisionAdmin{}, want: domain.ErrRejected},
		{name: "lost membership", admin: revisionAdmin{}, mutate: func(s *storage.Snapshot, _ *domain.GrantContent) { s.Memberships = nil }, want: domain.ErrRejected},
		{name: "disabled source", admin: revisionAdmin{}, mutate: func(s *storage.Snapshot, _ *domain.GrantContent) {
			a := s.Assignments["A1"]
			a.Status = "disabled"
			s.Assignments["A1"] = a
		}, want: domain.ErrRejected},
		{name: "expired source", admin: revisionAdmin{}, mutate: func(s *storage.Snapshot, _ *domain.GrantContent) {
			past := time.Now().Add(-time.Hour)
			g := s.Contents[domain.GrantKey{ID: "G1", Revision: 1}]
			g.Validity = &domain.Validity{ExpiresAt: &past}
			s.Contents[domain.GrantKey{ID: "G1", Revision: 1}] = g
		}, want: domain.ErrRejected},
		{name: "unknown permission", admin: revisionAdmin{}, mutate: func(_ *storage.Snapshot, g *domain.GrantContent) { g.Permissions = []string{"unknown"} }, want: domain.ErrRejected},
		{name: "unknown scope", admin: revisionAdmin{}, mutate: func(_ *storage.Snapshot, g *domain.GrantContent) { g.Scope = map[string]string{"unknown": "x"} }, want: domain.ErrRejected},
		{name: "amplified permissions", admin: revisionAdmin{}, mutate: func(_ *storage.Snapshot, g *domain.GrantContent) { g.Permissions = []string{"delete"} }, want: domain.ErrRejected},
		{name: "self scope", admin: revisionAdmin{}, mutate: func(_ *storage.Snapshot, g *domain.GrantContent) { g.Scope = map[string]string{"user": "$self"} }, want: domain.ErrUnsupported},
		{name: "unknown role", admin: revisionAdmin{}, mutate: func(_ *storage.Snapshot, g *domain.GrantContent) {
			g.Permissions, g.RoleID, g.RoleRevision = nil, "unknown", 1
		}, want: domain.ErrRejected},
		{name: "unknown role revision", admin: revisionAdmin{}, mutate: func(_ *storage.Snapshot, g *domain.GrantContent) {
			g.Permissions, g.RoleID, g.RoleRevision = nil, "reader", 2
		}, want: domain.ErrRejected},
		{name: "provider", admin: revisionAdmin{}, perr: domain.ErrUnavailable, want: domain.ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, issuer := revisionFixture(area)
			candidate := cloneContent(valid)
			if tc.mutate != nil {
				tc.mutate(&snapshot, &candidate)
			}
			callArea := tc.area
			if callArea == (domain.Area{}) {
				callArea = area
			}
			identity := tc.identity
			if identity == (domain.Identity{}) && tc.name != "bad identity" {
				identity = issuer
			}
			source := tc.source
			if source == "" {
				source = "A1"
			}
			p := &revisionProvider{snapshot: snapshot, err: tc.perr}
			s, _ := New(p, tc.admin, fixedClock{now: time.Now()})
			got, err := s.PublishGrantRevision(t.Context(), callArea, identity, source, candidate)
			if !errors.Is(err, tc.want) || !reflect.DeepEqual(got, domain.GrantContent{}) || p.writes != 0 {
				t.Fatalf("got=%#v writes=%d err=%v want=%v", got, p.writes, err, tc.want)
			}
		})
	}
}

func TestPublishGrantRevisionCancellationAndCandidateFutureValidity(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	snapshot, issuer := revisionFixture(area)
	future := time.Now().Add(time.Hour)
	candidate := revisionCandidate()
	candidate.Validity = &domain.Validity{NotBefore: &future}
	p := &revisionProvider{snapshot: snapshot}
	s, _ := New(p, revisionAdmin{}, fixedClock{now: time.Now()})
	if got, err := s.PublishGrantRevision(t.Context(), area, issuer, "A1", candidate); err != nil || !reflect.DeepEqual(got, candidate) {
		t.Fatalf("future candidate got=%#v err=%v", got, err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	p = &revisionProvider{snapshot: snapshot}
	s, _ = New(p, revisionAdmin{check: func(storage.Snapshot, domain.Identity, domain.GrantContent) error { cancel(); return nil }}, fixedClock{now: time.Now()})
	if got, err := s.PublishGrantRevision(ctx, area, issuer, "A1", revisionCandidate()); !errors.Is(err, context.Canceled) || !reflect.DeepEqual(got, domain.GrantContent{}) || p.writes != 0 {
		t.Fatalf("cancel got=%#v writes=%d err=%v", got, p.writes, err)
	}
}
