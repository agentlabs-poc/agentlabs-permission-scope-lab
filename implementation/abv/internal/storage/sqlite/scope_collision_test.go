package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"errors"
	"strings"
	"testing"
)

// A scope key is a bare word, so one catalog holds one entry per key — and an
// application's catalog is its own keys union every platform key. A key claimed at
// both boundaries is therefore ambiguous, and the read refuses rather than picking.
//
// It used to pick, by whichever row the SQL ordering put last: `ORDER BY key3,
// key4` with key3 the namespace, so an application whose id sorted before the
// platform's namespace had its own key overwritten by the platform's and its own
// opaque values validated as Auth record ids — while an application whose id sorted
// later kept working. Two deployments, the same records, different answers,
// decided by a string comparison nobody wrote down.
func TestAScopeKeyClaimedAtTwoBoundariesIsRefusedRatherThanResolved(t *testing.T) {
	// Both orderings, because the defect was invisible in one of them.
	for _, applicationID := range []string{"api", "zapp"} {
		t.Run(applicationID, func(t *testing.T) {
			area, err := domain.NewArea("acme", applicationID)
			if err != nil {
				t.Fatal(err)
			}
			application := minimalFixture(area)
			application.Catalog.ApplicationID = applicationID
			application.Catalog.Permissions = map[string]domain.PermissionDefinition{
				applicationID + ":payroll:payslip::read": {
					ID: applicationID + ":payroll:payslip::read", Active: true,
					Boundary: domain.ApplicationBoundary, Namespace: applicationID,
				},
			}
			application.Contents = map[domain.GrantKey]domain.GrantContent{}
			// The application registered `team` before Auth owned it.
			application.Catalog.Scopes = map[string]domain.ScopeDefinition{
				"team": {Key: "team", Boundary: domain.ApplicationBoundary},
			}
			platformArea, err := domain.NewArea("acme", "auth")
			if err != nil {
				t.Fatal(err)
			}
			platform := minimalFixture(platformArea)
			platform.Catalog.ApplicationID = "auth"
			platform.Catalog.Permissions = map[string]domain.PermissionDefinition{
				"auth:group::write": {ID: "auth:group::write", Active: true, Boundary: domain.PlatformBoundary, Namespace: "auth"},
			}
			platform.Catalog.Scopes = map[string]domain.ScopeDefinition{
				"team": {Key: "team", Boundary: domain.PlatformBoundary},
			}
			platform.Contents = map[domain.GrantKey]domain.GrantContent{}

			opened, err := CreateFixture(t.Context(), t.TempDir()+"/collide.db", []storage.Snapshot{application, platform})
			if err != nil {
				t.Fatal(err)
			}
			defer opened.Close()
			err = opened.Read(t.Context(), area, func(storage.Snapshot) error { return nil })
			if !errors.Is(err, domain.ErrMalformed) {
				t.Fatalf("reading an ambiguous catalog gave %v, want ErrMalformed", err)
			}
			// And it names the key, because an operator has to know which one.
			if err == nil || !strings.Contains(err.Error(), "team") {
				t.Fatalf("the refusal did not name the colliding key: %v", err)
			}
		})
	}
}
