// Package contracttest defines provider-neutral storage conformance tests.
package contracttest

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

type Factory struct {
	Open          func(context.Context, string) (storage.Provider, error)
	Create        func(context.Context, string, []storage.Snapshot) (storage.Provider, error)
	OpenWithLimit func(context.Context, string, int) (storage.Provider, error)
}

func Run(t *testing.T, factory Factory) {
	t.Helper()
	RunGrantStatus(t, factory)
	t.Run("round trips every family and persists", func(t *testing.T) {
		path := t.TempDir() + "/authority.db"
		want := fixtures(t)
		provider, err := factory.Create(t.Context(), path, want)
		if err != nil {
			t.Fatal(err)
		}
		assertSnapshot(t, provider, want[0])
		if err := provider.Close(); err != nil {
			t.Fatal(err)
		}
		provider, err = factory.Open(t.Context(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer provider.Close()
		assertSnapshot(t, provider, want[0])
	})

	t.Run("requires an installed nonzero area without callback", func(t *testing.T) {
		path := t.TempDir() + "/authority.db"
		provider, err := factory.Create(t.Context(), path, fixtures(t)[:1])
		if err != nil {
			t.Fatal(err)
		}
		defer provider.Close()
		missing, _ := domain.NewArea("missing", "hrms")
		for name, area := range map[string]domain.Area{"zero": {}, "missing": missing} {
			t.Run(name, func(t *testing.T) {
				var calls atomic.Int32
				err := provider.Read(t.Context(), area, func(storage.Snapshot) error { calls.Add(1); return nil })
				if err == nil || calls.Load() != 0 {
					t.Fatalf("read: calls=%d err=%v", calls.Load(), err)
				}
				err = provider.Update(t.Context(), area, func(storage.Snapshot) (storage.WriteSet, error) { calls.Add(1); return storage.WriteSet{}, nil })
				if err == nil || calls.Load() != 0 {
					t.Fatalf("update: calls=%d err=%v", calls.Load(), err)
				}
			})
		}
	})

	t.Run("isolates tenant and application reads and writes", func(t *testing.T) {
		path := t.TempDir() + "/authority.db"
		want := fixtures(t)
		provider, err := factory.Create(t.Context(), path, want)
		if err != nil {
			t.Fatal(err)
		}
		defer provider.Close()
		for _, seeded := range want {
			assertSnapshot(t, provider, seeded)
		}
		for i, seeded := range want {
			a := domain.Assignment{Version: "1", ID: "same-new-id", GrantID: "G1", GrantRevision: 2,
				Recipient: domain.Recipient{Type: "user", ID: []string{"tenant-user", "other-tenant-user", "other-app-user"}[i]}, Status: "enabled"}
			if err := provider.Update(t.Context(), seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) {
				return storage.WriteSet{NewAssignments: []domain.Assignment{a}}, nil
			}); err != nil {
				t.Fatalf("area %d: %v", i, err)
			}
		}
		for i, seeded := range want {
			err := provider.Read(t.Context(), seeded.Area, func(got storage.Snapshot) error {
				if len(got.Assignments) != 2 {
					t.Fatalf("area %d leaked assignments: %#v", i, got.Assignments)
				}
				if got.Assignments["same-new-id"].Recipient.ID != []string{"tenant-user", "other-tenant-user", "other-app-user"}[i] {
					t.Fatalf("area %d saw another area's write: %#v", i, got.Assignments["same-new-id"])
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("callback errors cancellations and panics roll back without replay", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			invoke func(storage.Provider, domain.Area, *atomic.Int32) error
		}{
			{"error", func(p storage.Provider, a domain.Area, calls *atomic.Int32) error {
				return p.Update(t.Context(), a, func(storage.Snapshot) (storage.WriteSet, error) {
					calls.Add(1)
					return storage.WriteSet{}, errors.New("stop")
				})
			}},
			{"cancel", func(p storage.Provider, a domain.Area, calls *atomic.Int32) error {
				ctx, cancel := context.WithCancel(t.Context())
				return p.Update(ctx, a, func(storage.Snapshot) (storage.WriteSet, error) {
					calls.Add(1)
					cancel()
					return storage.WriteSet{NewAssignments: []domain.Assignment{newAssignment("cancelled")}}, nil
				})
			}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				path := t.TempDir() + "/authority.db"
				seeded := fixtures(t)[0]
				p, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				var calls atomic.Int32
				if err := tc.invoke(p, seeded.Area, &calls); err == nil || calls.Load() != 1 {
					t.Fatalf("calls=%d err=%v", calls.Load(), err)
				} else if tc.name == "cancel" && !errors.Is(err, context.Canceled) {
					t.Fatalf("cancellation sentinel lost: %v", err)
				}
				assertAssignmentAbsent(t, p, seeded.Area, "cancelled")
			})
		}
		t.Run("panic", func(t *testing.T) {
			path := t.TempDir() + "/authority.db"
			seeded := fixtures(t)[0]
			p, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
			if err != nil {
				t.Fatal(err)
			}
			defer p.Close()
			var calls atomic.Int32
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("panic was swallowed")
					}
				}()
				_ = p.Update(t.Context(), seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) { calls.Add(1); panic("boom") })
			}()
			if calls.Load() != 1 {
				t.Fatalf("callback replayed: %d", calls.Load())
			}
			if err := p.Update(t.Context(), seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) { return storage.WriteSet{}, nil }); err != nil {
				t.Fatalf("dirty connection after panic: %v", err)
			}
		})
	})

	t.Run("snapshot mutation is inert and explicit write is committed", func(t *testing.T) {
		path := t.TempDir() + "/authority.db"
		seeded := fixtures(t)[0]
		p, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		if err := p.Update(t.Context(), seeded.Area, func(s storage.Snapshot) (storage.WriteSet, error) {
			delete(s.Assignments, "A1")
			s.Controls["G1"] = domain.GrantControl{Version: "1", ID: "G1", Status: "disabled"}
			return storage.WriteSet{NewAssignments: []domain.Assignment{newAssignment("A2")}}, nil
		}); err != nil {
			t.Fatal(err)
		}
		err = p.Read(t.Context(), seeded.Area, func(s storage.Snapshot) error {
			if s.Assignments["A1"].Status != "disabled" || s.Controls["G1"].Status != "enabled" {
				t.Fatalf("snapshot mutation persisted: %#v", s)
			}
			if _, ok := s.Assignments["A2"]; !ok {
				t.Fatal("explicit write missing")
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("multi assignment uniqueness failure is atomic after reopen", func(t *testing.T) {
		path := t.TempDir() + "/authority.db"
		seeded := fixtures(t)[0]
		p, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
		if err != nil {
			t.Fatal(err)
		}
		first := newAssignment("first")
		duplicate := seeded.Assignments["A1"]
		duplicate.ID = "duplicate-id"
		duplicate.Status = "enabled"
		err = p.Update(t.Context(), seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) {
			return storage.WriteSet{NewAssignments: []domain.Assignment{first, duplicate}}, nil
		})
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("want conflict, got %v", err)
		}
		assertAssignmentAbsent(t, p, seeded.Area, first.ID)
		if err := p.Close(); err != nil {
			t.Fatal(err)
		}
		p, err = factory.Open(t.Context(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		assertAssignmentAbsent(t, p, seeded.Area, first.ID)
	})

	t.Run("grant recipient uniqueness spans revisions and disabled rows", func(t *testing.T) {
		path := t.TempDir() + "/authority.db"
		seeded := fixtures(t)[0]
		older := seeded.Contents[domain.GrantKey{ID: "G1", Revision: 2}]
		older.Revision = 1
		seeded.Contents[domain.GrantKey{ID: "G1", Revision: 1}] = older
		p, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		duplicate := seeded.Assignments["A1"]
		duplicate.ID = "new-id"
		duplicate.GrantRevision = 1
		duplicate.Status = "enabled"
		err = p.Update(t.Context(), seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) {
			return storage.WriteSet{NewAssignments: []domain.Assignment{duplicate}}, nil
		})
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("accepted duplicate grant/recipient at another revision: %v", err)
		}
	})

	t.Run("rejects malformed and cross area proposed assignments from stored facts", func(t *testing.T) {
		path := t.TempDir() + "/authority.db"
		seeded := fixtures(t)[0]
		p, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		cases := map[string]domain.Assignment{
			"malformed":       {Version: "1", ID: "bad", GrantID: "G1", GrantRevision: 2, Recipient: domain.Recipient{Type: "service", ID: "x"}, Status: "enabled"},
			"missing content": {Version: "1", ID: "missing", GrantID: "elsewhere", GrantRevision: 2, Recipient: domain.Recipient{Type: "user", ID: "x"}, Status: "enabled"},
			"missing group":   {Version: "1", ID: "group", GrantID: "G1", GrantRevision: 2, Recipient: domain.Recipient{Type: "group", ID: "elsewhere"}, Status: "enabled"},
		}
		for name, proposed := range cases {
			t.Run(name, func(t *testing.T) {
				err := p.Update(t.Context(), seeded.Area, func(s storage.Snapshot) (storage.WriteSet, error) {
					s.Contents[domain.GrantKey{ID: proposed.GrantID, Revision: proposed.GrantRevision}] = seeded.Contents[domain.GrantKey{ID: "G1", Revision: 2}]
					s.Teams[proposed.Recipient.ID] = domain.Team{ID: proposed.Recipient.ID}
					return storage.WriteSet{NewAssignments: []domain.Assignment{proposed}}, nil
				})
				if err == nil {
					t.Fatal("accepted invalid proposal from callback-mutated snapshot")
				}
			})
		}
	})

	t.Run("snapshot bound fails instead of truncating", func(t *testing.T) {
		path := t.TempDir() + "/authority.db"
		seeded := fixtures(t)[0]
		p, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
		if err != nil {
			t.Fatal(err)
		}
		p.Close()
		p, err = factory.OpenWithLimit(t.Context(), path, 2)
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		var calls atomic.Int32
		err = p.Read(t.Context(), seeded.Area, func(storage.Snapshot) error { calls.Add(1); return nil })
		if !errors.Is(err, storage.ErrSnapshotLimit) || calls.Load() != 0 {
			t.Fatalf("calls=%d err=%v", calls.Load(), err)
		}
	})

	t.Run("read snapshot stays consistent while another provider commits", func(t *testing.T) {
		path := t.TempDir() + "/authority.db"
		seeded := fixtures(t)[0]
		reader, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
		writer, err := factory.Open(t.Context(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer writer.Close()
		loaded := make(chan struct{})
		committed := make(chan struct{})
		readDone := make(chan error, 1)
		go func() {
			readDone <- reader.Read(context.Background(), seeded.Area, func(snapshot storage.Snapshot) error {
				close(loaded)
				<-committed
				if _, exists := snapshot.Assignments["concurrent"]; exists {
					return errors.New("partially refreshed snapshot")
				}
				return nil
			})
		}()
		<-loaded
		err = writer.Update(t.Context(), seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) {
			return storage.WriteSet{NewAssignments: []domain.Assignment{newAssignment("concurrent")}}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		close(committed)
		if err := <-readDone; err != nil {
			t.Fatal(err)
		}
		if err := writer.Read(t.Context(), seeded.Area, func(snapshot storage.Snapshot) error {
			if _, exists := snapshot.Assignments["concurrent"]; !exists {
				t.Fatal("later snapshot missed committed assignment")
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})
}

func fixtures(t *testing.T) []storage.Snapshot {
	t.Helper()
	hrms := domain.Catalog{ApplicationID: "hrms", Permissions: map[string]domain.PermissionDefinition{"許可🚀": {ID: "許可🚀", Active: true}}, Scopes: map[string]domain.ScopeDefinition{"部門": {Key: "部門", AllowedTokens: []string{"$self"}}}, CompatibilityEnabled: true, SupportedKeys: map[string][]string{"許可🚀": {"部門"}}}
	crm := domain.Catalog{ApplicationID: "crm", Permissions: map[string]domain.PermissionDefinition{"許可🚀": {ID: "許可🚀", Active: false}}, Scopes: map[string]domain.ScopeDefinition{}, SupportedKeys: map[string][]string{}}
	a1, _ := domain.NewArea("acme", "hrms")
	a2, _ := domain.NewArea("beta", "hrms")
	a3, _ := domain.NewArea("acme", "crm")
	nb := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	ex := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	makeSnapshot := func(area domain.Area, catalog domain.Catalog, human string) storage.Snapshot {
		content := domain.GrantContent{Version: "1", GrantID: "G1", Revision: 2, Permissions: []string{"許可🚀"}, Scope: map[string]string{"部門": "財務"}, Validity: &domain.Validity{NotBefore: &nb, ExpiresAt: &ex}}
		assignment := domain.Assignment{Version: "1", ID: "A1", GrantID: "G1", GrantRevision: 2, Recipient: domain.Recipient{Type: "group", ID: "Team1"}, Status: "disabled"}
		return storage.Snapshot{Area: area, Catalog: catalog,
			Controls: map[string]domain.GrantControl{"G1": {Version: "1", ID: "G1", Status: "enabled"}},
			Contents: map[domain.GrantKey]domain.GrantContent{{ID: "G1", Revision: 2}: content}, Assignments: map[string]domain.Assignment{"A1": assignment},
			Roles: map[domain.RoleKey]domain.RoleContent{{ID: "reader", Revision: 3}: {ID: "reader", Revision: 3, Permissions: []string{"許可🚀"}}},
			Teams: map[string]domain.Team{"Team1": {ID: "Team1", ParentID: "preserved-orphan-parent"}}, Memberships: []domain.Membership{{TeamID: "Team1", HumanID: human}}, TrustedRoots: map[string]bool{"G1": true}}
	}
	return []storage.Snapshot{makeSnapshot(a1, hrms, "人間🚀"), makeSnapshot(a2, hrms, "other"), makeSnapshot(a3, crm, "third")}
}

func newAssignment(id string) domain.Assignment {
	return domain.Assignment{Version: "1", ID: id, GrantID: "G1", GrantRevision: 2, Recipient: domain.Recipient{Type: "user", ID: id + "-user"}, Status: "enabled"}
}
func assertSnapshot(t *testing.T, p storage.Provider, want storage.Snapshot) {
	t.Helper()
	err := p.Read(t.Context(), want.Area, func(got storage.Snapshot) error {
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("snapshot mismatch\n got: %#v\nwant: %#v", got, want)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func assertAssignmentAbsent(t *testing.T, p storage.Provider, area domain.Area, id string) {
	t.Helper()
	err := p.Read(t.Context(), area, func(s storage.Snapshot) error {
		if _, ok := s.Assignments[id]; ok {
			t.Fatalf("assignment %q unexpectedly persisted", id)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
