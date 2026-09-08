package contracttest

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func RunAssignmentStatus(t *testing.T, factory Factory) {
	for _, tc := range []struct{ name, before, after string }{{"disable", "enabled", "disabled"}, {"enable", "disabled", "enabled"}, {"same state", "disabled", "disabled"}} {
		t.Run("assignment status "+tc.name+" persists after reopen", func(t *testing.T) {
			seeded := fixtures(t)[0]
			a := seeded.Assignments["A1"]
			a.Status = tc.before
			seeded.Assignments["A1"] = a
			path := t.TempDir() + "/authority.db"
			p, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
			if err != nil {
				t.Fatal(err)
			}
			before, after := a, a
			after.Status = tc.after
			err = p.Update(t.Context(), seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) {
				return storage.WriteSet{AssignmentStatusChange: &storage.AssignmentStatusChange{Before: before, After: after}}, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if err = p.Close(); err != nil {
				t.Fatal(err)
			}
			p, err = factory.Open(t.Context(), path)
			if err != nil {
				t.Fatal(err)
			}
			defer p.Close()
			seeded.Assignments["A1"] = after
			assertSnapshot(t, p, seeded)
		})
	}

	t.Run("rejects malformed stale missing immutable and mixed changes without effects", func(t *testing.T) {
		base := fixtures(t)[0].Assignments["A1"]
		valid := func() storage.AssignmentStatusChange {
			after := base
			after.Status = "enabled"
			return storage.AssignmentStatusChange{Before: base, After: after}
		}
		cases := map[string]struct {
			mutate                func(*storage.AssignmentStatusChange)
			newAssignments, grant bool
			want                  error
		}{
			"missing":                    {func(c *storage.AssignmentStatusChange) { c.Before.ID, c.After.ID = "missing", "missing" }, false, false, domain.ErrNotFound},
			"stale before":               {func(c *storage.AssignmentStatusChange) { c.Before.Status = "enabled" }, false, false, domain.ErrConflict},
			"empty version":              {func(c *storage.AssignmentStatusChange) { c.Before.Version = "" }, false, false, domain.ErrMalformed},
			"future version":             {func(c *storage.AssignmentStatusChange) { c.After.Version = "2" }, false, false, domain.ErrUnsupported},
			"invalid status":             {func(c *storage.AssignmentStatusChange) { c.After.Status = "paused" }, false, false, domain.ErrMalformed},
			"empty ID":                   {func(c *storage.AssignmentStatusChange) { c.Before.ID, c.After.ID = "", "" }, false, false, domain.ErrMalformed},
			"changed ID":                 {func(c *storage.AssignmentStatusChange) { c.After.ID = "A2" }, false, false, domain.ErrMalformed},
			"changed grant":              {func(c *storage.AssignmentStatusChange) { c.After.GrantID = "G2" }, false, false, domain.ErrMalformed},
			"changed revision":           {func(c *storage.AssignmentStatusChange) { c.After.GrantRevision++ }, false, false, domain.ErrMalformed},
			"changed recipient type":     {func(c *storage.AssignmentStatusChange) { c.After.Recipient.Type = "user" }, false, false, domain.ErrMalformed},
			"changed recipient ID":       {func(c *storage.AssignmentStatusChange) { c.After.Recipient.ID = "other" }, false, false, domain.ErrMalformed},
			"mixed assignment creation":  {func(*storage.AssignmentStatusChange) {}, true, false, domain.ErrMalformed},
			"mixed grant status":         {func(*storage.AssignmentStatusChange) {}, false, true, domain.ErrMalformed},
			"all three write categories": {func(*storage.AssignmentStatusChange) {}, true, true, domain.ErrMalformed},
		}
		for name, tc := range cases {
			t.Run(name, func(t *testing.T) {
				seeded := fixtures(t)[0]
				path := t.TempDir() + "/authority.db"
				p, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
				if err != nil {
					t.Fatal(err)
				}
				change := valid()
				tc.mutate(&change)
				err = p.Update(t.Context(), seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) {
					w := storage.WriteSet{AssignmentStatusChange: &change}
					if tc.newAssignments {
						w.NewAssignments = []domain.Assignment{newAssignment("must-not-persist")}
					}
					if tc.grant {
						g := statusChange("G1", "enabled", "disabled")
						w.GrantStatusChange = &g
					}
					return w, nil
				})
				if !errors.Is(err, tc.want) {
					t.Fatalf("want %v, got %v", tc.want, err)
				}
				if err = p.Close(); err != nil {
					t.Fatal(err)
				}
				p, err = factory.Open(t.Context(), path)
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				assertSnapshot(t, p, seeded)
			})
		}
	})

	t.Run("forged callback snapshot cannot replace persisted before", func(t *testing.T) {
		seeded := fixtures(t)[0]
		path := t.TempDir() + "/authority.db"
		p, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		err = p.Update(t.Context(), seeded.Area, func(s storage.Snapshot) (storage.WriteSet, error) {
			before := s.Assignments["A1"]
			before.Status = "enabled"
			s.Assignments["A1"] = before
			after := before
			after.Status = "disabled"
			return storage.WriteSet{AssignmentStatusChange: &storage.AssignmentStatusChange{Before: before, After: after}}, nil
		})
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("want conflict, got %v", err)
		}
		assertSnapshot(t, p, seeded)
	})

	t.Run("same assignment ID updates only selected tenant and application", func(t *testing.T) {
		seeded := fixtures(t)
		path := t.TempDir() + "/authority.db"
		p, err := factory.Create(t.Context(), path, seeded)
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		for i := range seeded {
			before := seeded[i].Assignments["A1"]
			after := before
			after.Status = "enabled"
			if err := p.Update(t.Context(), seeded[i].Area, func(storage.Snapshot) (storage.WriteSet, error) {
				return storage.WriteSet{AssignmentStatusChange: &storage.AssignmentStatusChange{Before: before, After: after}}, nil
			}); err != nil {
				t.Fatal(err)
			}
			seeded[i].Assignments["A1"] = after
			for j := range seeded {
				assertSnapshot(t, p, seeded[j])
			}
		}
	})

	t.Run("cancellation rolls back assignment status after reopen", func(t *testing.T) {
		seeded := fixtures(t)[0]
		path := t.TempDir() + "/authority.db"
		p, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(t.Context())
		err = p.Update(ctx, seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) {
			before := seeded.Assignments["A1"]
			after := before
			after.Status = "enabled"
			cancel()
			return storage.WriteSet{AssignmentStatusChange: &storage.AssignmentStatusChange{Before: before, After: after}}, nil
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("want cancellation, got %v", err)
		}
		if err = p.Close(); err != nil {
			t.Fatal(err)
		}
		p, err = factory.Open(t.Context(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		assertSnapshot(t, p, seeded)
	})

	t.Run("competing writer cannot run assignment callback", func(t *testing.T) {
		seeded := fixtures(t)[0]
		path := t.TempDir() + "/authority.db"
		first, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
		if err != nil {
			t.Fatal(err)
		}
		defer first.Close()
		second, err := factory.Open(t.Context(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer second.Close()
		entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
		go func() {
			done <- first.Update(context.Background(), seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) {
				close(entered)
				<-release
				return storage.WriteSet{}, nil
			})
		}()
		<-entered
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()
		var calls atomic.Int32
		err = second.Update(ctx, seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) {
			calls.Add(1)
			before := seeded.Assignments["A1"]
			after := before
			after.Status = "enabled"
			return storage.WriteSet{AssignmentStatusChange: &storage.AssignmentStatusChange{Before: before, After: after}}, nil
		})
		if err == nil || calls.Load() != 0 {
			t.Fatalf("calls=%d err=%v", calls.Load(), err)
		}
		close(release)
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	})
}
