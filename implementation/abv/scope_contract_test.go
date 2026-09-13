package abv_test

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"errors"
	"testing"
)

func seededScopes(t *testing.T, allow bool, keys ...string) (*abv.Facade, domain.Application, domain.Identity) {
	t.Helper()
	scopes := map[string]domain.ScopeDefinition{}
	for _, key := range keys {
		scopes[key] = domain.ScopeDefinition{Key: key}
	}
	provider := &catalogMemoryProvider{catalog: domain.Catalog{
		ApplicationID: "hrms",
		Permissions:   map[string]domain.PermissionDefinition{},
		Scopes:        scopes,
	}}
	facade, err := abv.New(provider, catalogAdministration{allow: allow}, clock{})
	if err != nil {
		t.Fatal(err)
	}
	app, _ := domain.NewApplication("hrms")
	id := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "admin"}, HumanID: "admin"}
	return facade, app, id
}

func TestGetScopeReturnsItsTokens(t *testing.T) {
	f, app, id := seededScopes(t, true, "dept", "user")

	plain, err := f.GetScope(t.Context(), app, id, "dept")
	if err != nil || plain.Key != "dept" {
		t.Fatalf("dept got=%#v err=%v", plain, err)
	}

	if _, err = f.GetScope(t.Context(), app, id, "region"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unregistered err=%v, want ErrNotFound", err)
	}
	if _, err = f.GetScope(t.Context(), app, id, "  "); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("blank err=%v, want ErrMalformed", err)
	}
	if _, err = f.GetScope(t.Context(), app, id, "de*t"); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("wildcard err=%v, want ErrMalformed", err)
	}
}

func TestListScopesPagesAndIsProtected(t *testing.T) {
	f, app, id := seededScopes(t, true, "cert", "dept", "region", "user")

	all, err := f.ListScopes(t.Context(), app, id, domain.ScopeFilter{})
	if err != nil || len(all.Scopes) != 4 || all.Total != 4 {
		t.Fatalf("all got=%d total=%d err=%v", len(all.Scopes), all.Total, err)
	}
	for i := 1; i < len(all.Scopes); i++ {
		if all.Scopes[i-1].Key >= all.Scopes[i].Key {
			t.Fatalf("not ordered: %q then %q", all.Scopes[i-1].Key, all.Scopes[i].Key)
		}
	}

	seen := map[string]bool{}
	for offset := 0; offset < all.Total; offset += 2 {
		page, err := f.ListScopes(t.Context(), app, id, domain.ScopeFilter{Offset: offset, Limit: 2})
		if err != nil {
			t.Fatal(err)
		}
		if page.Generation != all.Generation {
			t.Fatalf("generation moved during a quiet walk: %d then %d", all.Generation, page.Generation)
		}
		for _, definition := range page.Scopes {
			if seen[definition.Key] {
				t.Fatalf("key repeated across pages: %q", definition.Key)
			}
			seen[definition.Key] = true
		}
	}
	if len(seen) != 4 {
		t.Fatalf("paged over %d keys, want 4", len(seen))
	}

	beyond, err := f.ListScopes(t.Context(), app, id, domain.ScopeFilter{Offset: 99, Limit: 2})
	if err != nil || len(beyond.Scopes) != 0 || beyond.Total != 4 {
		t.Fatalf("beyond got=%#v err=%v", beyond, err)
	}
	for name, filter := range map[string]domain.ScopeFilter{
		"negative offset": {Offset: -1},
		"negative limit":  {Limit: -1},
		"limit over cap":  {Limit: 501},
	} {
		if _, err := f.ListScopes(t.Context(), app, id, filter); !errors.Is(err, domain.ErrMalformed) {
			t.Errorf("%s: err=%v, want ErrMalformed", name, err)
		}
	}

	// Reads are protected, like every other catalog operation.
	denied, appDenied, idDenied := seededScopes(t, false, "dept")
	if _, err := denied.GetScope(t.Context(), appDenied, idDenied, "dept"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("get err=%v, want ErrRejected", err)
	}
	if _, err := denied.ListScopes(t.Context(), appDenied, idDenied, domain.ScopeFilter{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("list err=%v, want ErrRejected", err)
	}
}

// TestScopeHasNoStatusOperation records a deliberate absence: a scope carries no
// active flag, and retiring one would invalidate published grants that reference
// the key rather than merely stopping future evaluation.
func TestScopeHasNoStatusOperation(t *testing.T) {
	f, _, _ := seededScopes(t, true, "dept")
	var facade any = f
	if _, exists := facade.(interface {
		SetScopeStatus(any, any, any, string, bool) (domain.ScopeDefinition, error)
	}); exists {
		t.Fatal("a scope status operation exists; retirement was deliberately left absent")
	}
}
