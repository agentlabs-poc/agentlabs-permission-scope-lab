package abv_test

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"context"
	"errors"
	"testing"
	"time"
)

// seeded builds a catalog facade holding the given identifiers, all active
// unless the identifier appears in retired.
func seeded(t *testing.T, allow bool, active []string, retired []string) (*abv.Facade, domain.Application, domain.Identity) {
	t.Helper()
	permissions := map[string]domain.PermissionDefinition{}
	for _, id := range active {
		permissions[id] = domain.PermissionDefinition{ID: id, Active: true}
	}
	for _, id := range retired {
		permissions[id] = domain.PermissionDefinition{ID: id, Active: false}
	}
	provider := &catalogMemoryProvider{catalog: domain.Catalog{
		ApplicationID: "hrms",
		Permissions:   permissions,
		Scopes:        map[string]domain.ScopeDefinition{},
	}}
	facade, err := abv.New(provider, catalogAdministration{allow: allow}, clock{})
	if err != nil {
		t.Fatal(err)
	}
	app, _ := domain.NewApplication("hrms")
	id := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "admin"}, HumanID: "admin"}
	return facade, app, id
}

func TestGetPermissionReturnsRetiredRatherThanHiding(t *testing.T) {
	f, app, id := seeded(t, true,
		[]string{"hrms:employee:certificate::read"},
		[]string{"hrms:employee:certificate::write"})

	got, err := f.GetPermission(t.Context(), app, id, "hrms:employee:certificate::read")
	if err != nil || !got.Active || got.ID != "hrms:employee:certificate::read" {
		t.Fatalf("active got=%#v err=%v", got, err)
	}

	// A retired identifier is reported, not hidden: Q-126 permanence is only
	// observable if a caller naming an identifier learns it exists.
	got, err = f.GetPermission(t.Context(), app, id, "hrms:employee:certificate::write")
	if err != nil || got.Active {
		t.Fatalf("retired got=%#v err=%v", got, err)
	}

	if _, err = f.GetPermission(t.Context(), app, id, "hrms:employee:certificate::delete"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unregistered err=%v, want ErrNotFound", err)
	}
	if _, err = f.GetPermission(t.Context(), app, id, "  "); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("blank id err=%v, want ErrMalformed", err)
	}
	if _, err = f.GetPermission(t.Context(), app, id, "hrms:*::read"); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("wildcard err=%v, want ErrMalformed", err)
	}
}

func TestPermissionReadsAreProtected(t *testing.T) {
	f, app, id := seeded(t, false, []string{"hrms:employee:certificate::read"}, nil)

	if _, err := f.GetPermission(t.Context(), app, id, "hrms:employee:certificate::read"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("get err=%v, want ErrRejected", err)
	}
	if _, err := f.ListPermissions(t.Context(), app, id, domain.PermissionFilter{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("list err=%v, want ErrRejected", err)
	}

	// A proxy identity is not supported by any catalog operation.
	f, app, _ = seeded(t, true, []string{"hrms:employee:certificate::read"}, nil)
	proxy := domain.Identity{Version: "1", Actor: domain.Actor{Type: "agent", ID: "A-17"}, HumanID: "U-17"}
	if _, err := f.GetPermission(t.Context(), app, proxy, "hrms:employee:certificate::read"); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("proxy err=%v, want ErrUnsupported", err)
	}
}

func TestListPermissionsFiltersPagesAndOrders(t *testing.T) {
	f, app, id := seeded(t, true, []string{
		"hrms:employee:certificate::read",
		"hrms:employee:certificate::write",
		"hrms:payroll:payslip::read",
	}, []string{"hrms:employee:profile::read"})

	all, err := f.ListPermissions(t.Context(), app, id, domain.PermissionFilter{})
	if err != nil || len(all.Permissions) != 4 || all.Total != 4 {
		t.Fatalf("all got=%d total=%d err=%v", len(all.Permissions), all.Total, err)
	}
	// Ordering is by identifier, always — it is what makes an offset mean the
	// same thing on every call.
	for i := 1; i < len(all.Permissions); i++ {
		if all.Permissions[i-1].ID >= all.Permissions[i].ID {
			t.Fatalf("not ordered: %q then %q", all.Permissions[i-1].ID, all.Permissions[i].ID)
		}
	}

	byPrefix, err := f.ListPermissions(t.Context(), app, id, domain.PermissionFilter{Prefix: "hrms:employee:certificate:"})
	if err != nil || len(byPrefix.Permissions) != 2 || byPrefix.Total != 2 {
		t.Fatalf("prefix got=%d total=%d err=%v", len(byPrefix.Permissions), byPrefix.Total, err)
	}

	activeOnly, err := f.ListPermissions(t.Context(), app, id, domain.PermissionFilter{ActiveOnly: true})
	if err != nil || len(activeOnly.Permissions) != 3 || activeOnly.Total != 3 {
		t.Fatalf("active got=%d total=%d err=%v", len(activeOnly.Permissions), activeOnly.Total, err)
	}

	// Offset paging walks the catalog once, in order, with no repeats — and
	// Total lets a reader compute the page count up front rather than
	// discovering the end by walking into it.
	seen := map[string]bool{}
	for offset := 0; offset < all.Total; offset += 2 {
		page, err := f.ListPermissions(t.Context(), app, id, domain.PermissionFilter{Offset: offset, Limit: 2})
		if err != nil {
			t.Fatal(err)
		}
		if page.Total != all.Total {
			t.Fatalf("total moved between pages: %d then %d", all.Total, page.Total)
		}
		if page.Generation != all.Generation {
			t.Fatalf("generation moved during a quiet walk: %d then %d", all.Generation, page.Generation)
		}
		for _, definition := range page.Permissions {
			if seen[definition.ID] {
				t.Fatalf("identifier repeated across pages: %q", definition.ID)
			}
			seen[definition.ID] = true
		}
	}
	if len(seen) != 4 {
		t.Fatalf("paged over %d identifiers, want 4", len(seen))
	}

	// An offset past the end is an empty page, not an error — a reader that
	// jumps beyond the last page sees nothing rather than failing.
	beyond, err := f.ListPermissions(t.Context(), app, id, domain.PermissionFilter{Offset: 999, Limit: 2})
	if err != nil || len(beyond.Permissions) != 0 || beyond.Total != 4 {
		t.Fatalf("beyond got=%#v err=%v", beyond, err)
	}

	// A filter matching nothing is an empty page, never a fallback to everything.
	empty, err := f.ListPermissions(t.Context(), app, id, domain.PermissionFilter{Prefix: "codehost:"})
	if err != nil || len(empty.Permissions) != 0 || empty.Total != 0 {
		t.Fatalf("empty got=%#v err=%v", empty, err)
	}
}

// TestGenerationDetectsAWriteBetweenPages is why an offset walk is safe to
// cache: a reader compares the generation across its pages and retries when it
// moved, rather than silently missing a record an insert shifted past it.
func TestGenerationDetectsAWriteBetweenPages(t *testing.T) {
	f, app, id := seeded(t, true, []string{"hrms:a:b::read", "hrms:a:b::write"}, nil)

	before, err := f.ListPermissions(t.Context(), app, id, domain.PermissionFilter{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.RegisterPermission(t.Context(), app, id,
		domain.PermissionDefinition{ID: "hrms:a:b::approve", Active: true}); err != nil {
		t.Fatal(err)
	}
	after, err := f.ListPermissions(t.Context(), app, id, domain.PermissionFilter{Offset: 1, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if after.Generation == before.Generation {
		t.Fatal("generation did not move across a write; an offset walk could silently miss a record")
	}
	if after.Total == before.Total {
		t.Fatalf("total did not move: %d then %d", before.Total, after.Total)
	}
}

func TestListPermissionsRejectsUnboundedOrWildcardFilters(t *testing.T) {
	f, app, id := seeded(t, true, []string{"hrms:employee:certificate::read"}, nil)

	for name, filter := range map[string]domain.PermissionFilter{
		"negative limit":  {Limit: -1},
		"negative offset": {Offset: -1},
		"limit over cap":  {Limit: 501},
		"wildcard prefix": {Prefix: "hrms:*"},
		"partial segment": {Prefix: "hrms:emp"},
	} {
		if _, err := f.ListPermissions(t.Context(), app, id, filter); !errors.Is(err, domain.ErrMalformed) {
			t.Fatalf("%s: err=%v, want ErrMalformed", name, err)
		}
	}
}

func TestSetPermissionStatusIsReversibleAndIdempotent(t *testing.T) {
	const target = "hrms:employee:certificate::read"
	f, app, id := seeded(t, true, []string{target}, nil)

	retired, err := f.SetPermissionStatus(t.Context(), app, id, target, false)
	if err != nil || retired.Active {
		t.Fatalf("retire got=%#v err=%v", retired, err)
	}
	if got, _ := f.GetPermission(t.Context(), app, id, target); got.Active {
		t.Fatal("retirement did not persist")
	}

	// Idempotent: setting the status it already holds succeeds.
	if _, err = f.SetPermissionStatus(t.Context(), app, id, target, false); err != nil {
		t.Fatalf("repeat retire err=%v", err)
	}

	// Reversible: restoring is the same operation, not a separate one.
	restored, err := f.SetPermissionStatus(t.Context(), app, id, target, true)
	if err != nil || !restored.Active {
		t.Fatalf("restore got=%#v err=%v", restored, err)
	}
	if got, _ := f.GetPermission(t.Context(), app, id, target); !got.Active {
		t.Fatal("restoration did not persist")
	}
}

func TestSetPermissionStatusNeverCreatesAndIsProtected(t *testing.T) {
	f, app, id := seeded(t, true, []string{"hrms:employee:certificate::read"}, nil)

	// A status change is not a back door to registration: Q-126 makes an
	// identifier's meaning permanent, so one must never appear this way.
	if _, err := f.SetPermissionStatus(t.Context(), app, id, "hrms:employee:profile::read", true); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unregistered err=%v, want ErrNotFound", err)
	}
	if _, err := f.SetPermissionStatus(t.Context(), app, id, "hrms:*::read", false); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("wildcard err=%v, want ErrMalformed", err)
	}

	denied, appDenied, idDenied := seeded(t, false, []string{"hrms:employee:certificate::read"}, nil)
	if _, err := denied.SetPermissionStatus(t.Context(), appDenied, idDenied, "hrms:employee:certificate::read", false); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("denied err=%v, want ErrRejected", err)
	}
	// A denied status change writes nothing.
	if got, _ := f.GetPermission(t.Context(), app, id, "hrms:employee:certificate::read"); !got.Active {
		t.Fatal("denied change altered the catalog")
	}
}

func TestPermissionOperationsRequireCatalogSupport(t *testing.T) {
	// A provider and adapter without catalog support make these operations
	// unsupported rather than unprotected.
	facade, err := abv.New(plainProvider{}, plainAdministration{}, clock{})
	if err != nil {
		t.Fatal(err)
	}
	app, _ := domain.NewApplication("hrms")
	id := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "admin"}, HumanID: "admin"}

	if _, err := facade.GetPermission(t.Context(), app, id, "hrms:a::read"); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("get err=%v, want ErrUnsupported", err)
	}
	if _, err := facade.ListPermissions(t.Context(), app, id, domain.PermissionFilter{}); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("list err=%v, want ErrUnsupported", err)
	}
	if _, err := facade.SetPermissionStatus(t.Context(), app, id, "hrms:a::read", false); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("status err=%v, want ErrUnsupported", err)
	}
}

// plainProvider implements storage.Provider and deliberately not
// storage.CatalogProvider, so the catalog operations have no seam to use.
type plainProvider struct{}

func (plainProvider) Read(context.Context, domain.Area, func(storage.Snapshot) error) error {
	return nil
}
func (plainProvider) Update(context.Context, domain.Area, func(storage.Snapshot) (storage.WriteSet, error)) error {
	return nil
}
func (plainProvider) Close() error { return nil }

// plainAdministration implements only the mandatory assignment check.
type plainAdministration struct{}

func (plainAdministration) CheckAssignment(context.Context, abv.Evidence, domain.Identity, domain.Assignment, time.Time) error {
	return domain.ErrUnsupported
}
