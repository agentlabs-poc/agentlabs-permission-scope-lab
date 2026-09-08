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

func RunGrantStatus(t *testing.T, factory Factory) {
	t.Run("control status persists without changing authority content", func(t *testing.T) {
		seeded := fixtures(t)[0]
		path := t.TempDir() + "/authority.db"
		provider, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
		if err != nil {
			t.Fatal(err)
		}
		err = provider.Update(t.Context(), seeded.Area, func(s storage.Snapshot) (storage.WriteSet, error) {
			before := s.Controls["G1"]
			after := before
			after.Status = "disabled"
			return storage.WriteSet{GrantStatusChange: &storage.GrantStatusChange{Before: before, After: after}}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if err = provider.Close(); err != nil {
			t.Fatal(err)
		}
		provider, err = factory.Open(t.Context(), path)
		if err != nil {
			t.Fatal(err)
		}
		defer provider.Close()
		control := seeded.Controls["G1"]
		control.Status = "disabled"
		seeded.Controls["G1"] = control
		assertSnapshot(t, provider, seeded)
	})

	t.Run("same grant ID updates only the selected tenant and application", func(t *testing.T) {
		seeded := fixtures(t)
		path := t.TempDir() + "/authority.db"
		provider, err := factory.Create(t.Context(), path, seeded)
		if err != nil {
			t.Fatal(err)
		}
		defer provider.Close()
		for i := range seeded {
			before := seeded[i].Controls["G1"]
			after := before
			after.Status = "disabled"
			if err := provider.Update(t.Context(), seeded[i].Area, func(storage.Snapshot) (storage.WriteSet, error) {
				return storage.WriteSet{GrantStatusChange: &storage.GrantStatusChange{Before: before, After: after}}, nil
			}); err != nil {
				t.Fatalf("area %d: %v", i, err)
			}
			seeded[i].Controls["G1"] = after
			for j := range seeded {
				assertSnapshot(t, provider, seeded[j])
			}
		}
	})

	t.Run("rejects missing stale malformed and mixed changes without effects after reopen", func(t *testing.T) {
		cases := map[string]struct {
			change storage.GrantStatusChange
			mixed  bool
			want   error
		}{
			"missing":          {change: statusChange("missing", "enabled", "disabled"), want: domain.ErrNotFound},
			"wrong before":     {change: statusChange("G1", "disabled", "enabled"), want: domain.ErrConflict},
			"empty version":    {change: statusChange("G1", "enabled", "disabled"), want: domain.ErrMalformed},
			"future version":   {change: statusChange("G1", "enabled", "disabled"), want: domain.ErrUnsupported},
			"invalid status":   {change: statusChange("G1", "enabled", "paused"), want: domain.ErrMalformed},
			"changed ID":       {change: statusChange("G1", "enabled", "disabled"), want: domain.ErrMalformed},
			"wildcard ID":      {change: statusChange("G*", "enabled", "disabled"), want: domain.ErrMalformed},
			"invalid UTF-8 ID": {change: statusChange(string([]byte{0xff}), "enabled", "disabled"), want: domain.ErrMalformed},
			"mixed":            {change: statusChange("G1", "enabled", "disabled"), mixed: true, want: domain.ErrMalformed},
		}
		invalid := cases["empty version"]
		invalid.change.Before.Version = ""
		cases["empty version"] = invalid
		future := cases["future version"]
		future.change.After.Version = "2"
		cases["future version"] = future
		changed := cases["changed ID"]
		changed.change.After.ID = "G2"
		cases["changed ID"] = changed
		for name, tc := range cases {
			t.Run(name, func(t *testing.T) {
				seeded := fixtures(t)[0]
				path := t.TempDir() + "/authority.db"
				provider, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
				if err != nil {
					t.Fatal(err)
				}
				err = provider.Update(t.Context(), seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) {
					writes := storage.WriteSet{GrantStatusChange: &tc.change}
					if tc.mixed {
						writes.NewAssignments = []domain.Assignment{newAssignment("must-not-persist")}
					}
					return writes, nil
				})
				if !errors.Is(err, tc.want) {
					t.Fatalf("want %v, got %v", tc.want, err)
				}
				if err := provider.Close(); err != nil {
					t.Fatal(err)
				}
				provider, err = factory.Open(t.Context(), path)
				if err != nil {
					t.Fatal(err)
				}
				defer provider.Close()
				assertSnapshot(t, provider, seeded)
			})
		}
	})

	t.Run("callback errors and cancellation do not change control", func(t *testing.T) {
		for _, name := range []string{"error", "cancel"} {
			t.Run(name, func(t *testing.T) {
				seeded := fixtures(t)[0]
				path := t.TempDir() + "/authority.db"
				provider, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithCancel(t.Context())
				err = provider.Update(ctx, seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) {
					change := statusChange("G1", "enabled", "disabled")
					if name == "cancel" {
						cancel()
						return storage.WriteSet{GrantStatusChange: &change}, nil
					}
					return storage.WriteSet{}, errors.New("stop")
				})
				if err == nil || name == "cancel" && !errors.Is(err, context.Canceled) {
					t.Fatalf("unexpected error: %v", err)
				}
				if err := provider.Close(); err != nil {
					t.Fatal(err)
				}
				provider, err = factory.Open(t.Context(), path)
				if err != nil {
					t.Fatal(err)
				}
				defer provider.Close()
				assertSnapshot(t, provider, seeded)
			})
		}
	})

	t.Run("competing writer cannot run callback", func(t *testing.T) {
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
			change := statusChange("G1", "enabled", "disabled")
			return storage.WriteSet{GrantStatusChange: &change}, nil
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

func statusChange(id, beforeStatus, afterStatus string) storage.GrantStatusChange {
	return storage.GrantStatusChange{
		Before: domain.GrantControl{Version: "1", ID: id, Status: beforeStatus},
		After:  domain.GrantControl{Version: "1", ID: id, Status: afterStatus},
	}
}
