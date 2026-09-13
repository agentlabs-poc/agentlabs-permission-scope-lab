package validation

import (
	"agentlabs.local/abv/domain"
	"errors"
	"testing"
)

func TestCatalogPermissionRegistrationValidation(t *testing.T) {
	c := catalog()
	valid := domain.PermissionDefinition{ID: "hrms:payroll:payslip::export", Active: true}
	for _, tc := range []struct {
		name string
		def  domain.PermissionDefinition
		want error
	}{
		{"valid", valid, nil},
		{"malformed id", domain.PermissionDefinition{ID: "bad*", Active: true}, domain.ErrMalformed},
		{"inactive at registration", domain.PermissionDefinition{ID: valid.ID}, domain.ErrRejected},
		{"already registered", domain.PermissionDefinition{ID: read, Active: true}, domain.ErrConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := CheckPermissionRegistration(c, tc.def); !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
}

func TestCatalogScopeRegistrationValidation(t *testing.T) {
	c := catalog()
	for _, tc := range []struct {
		name string
		def  domain.ScopeDefinition
		want error
	}{
		{"valid", domain.ScopeDefinition{Key: "region"}, nil},
		{"duplicate", domain.ScopeDefinition{Key: "dept"}, domain.ErrConflict},
		{"malformed key", domain.ScopeDefinition{Key: "*"}, domain.ErrMalformed},
		{"blank key", domain.ScopeDefinition{Key: "  "}, domain.ErrMalformed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := CheckScopeRegistration(c, tc.def); !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
}

// TestScopeRegistrationRejectsAnyWildcard guards a defect a demo run caught: the
// check rejected a key that was exactly "*" but accepted one merely containing
// it, unlike the permission path.
func TestScopeRegistrationRejectsAnyWildcard(t *testing.T) {
	c := domain.Catalog{ApplicationID: "hrms", Scopes: map[string]domain.ScopeDefinition{}}
	for _, key := range []string{"*", "bad*", "*bad", "de*pt"} {
		definition := domain.ScopeDefinition{Key: key}
		if err := CheckScopeRegistration(c, definition); !errors.Is(err, domain.ErrMalformed) {
			t.Errorf("key %q: err=%v, want ErrMalformed", key, err)
		}
	}
	valid := domain.ScopeDefinition{Key: "dept"}
	if err := CheckScopeRegistration(c, valid); err != nil {
		t.Fatalf("valid key rejected: %v", err)
	}
}

// A permission's first noun IS the application. An identifier starting with
// anything else is canonically incorrect: key3 holds that one fact for every
// record type, and an identifier disagreeing with it would mean the row and the
// string it renders to say different things.
func TestPermissionMustStartWithItsApplication(t *testing.T) {
	catalog := domain.Catalog{
		ApplicationID: "hrms",
		Permissions:   map[string]domain.PermissionDefinition{},
		Scopes:        map[string]domain.ScopeDefinition{},
	}
	for _, id := range []string{"hrms::read", "hrms:payroll::read", "hrms:payroll:payslip::write"} {
		if err := CheckPermissionRegistration(catalog, domain.PermissionDefinition{ID: id, Active: true}); err != nil {
			t.Fatalf("rejected an identifier that starts with its application: %q -> %v", id, err)
		}
	}
	for _, id := range []string{"billing:invoice::read", "reporting:ledger:entry::export", "HRMS:payroll::read"} {
		err := CheckPermissionRegistration(catalog, domain.PermissionDefinition{ID: id, Active: true})
		if !errors.Is(err, domain.ErrRejected) {
			t.Fatalf("accepted an identifier that does not start with its application: %q -> %v", id, err)
		}
	}
}

// The leading-noun rule is the only thing the boundary changes. At the platform
// boundary key3 holds a namespace no application can claim, so there is nothing
// to compare against — and comparing would mean Auth-AL holding the platform's
// reserved list, which belongs to the auth service rather than here.
func TestTheLeadingNounRuleAppliesOnlyAtTheApplicationBoundary(t *testing.T) {
	catalog := domain.Catalog{
		ApplicationID: "hrms",
		Permissions:   map[string]domain.PermissionDefinition{},
		Scopes:        map[string]domain.ScopeDefinition{},
	}
	platform := domain.Catalog{
		ApplicationID: "system",
		Permissions:   map[string]domain.PermissionDefinition{},
		Scopes:        map[string]domain.ScopeDefinition{},
	}
	// system:user::read is a permission the auth service already ships. At the
	// application boundary it is unregistrable, because no application may be
	// called "system" — the registry reserves the name.
	const platformPermission = "system:user::read"
	if err := CheckPermissionRegistrationAt(domain.ApplicationBoundary, catalog, domain.PermissionDefinition{ID: platformPermission, Active: true}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("application boundary accepted a foreign namespace: %v", err)
	}
	if err := CheckPermissionRegistrationAt(domain.PlatformBoundary, platform, domain.PermissionDefinition{ID: platformPermission, Active: true}); err != nil {
		t.Fatalf("platform boundary rejected its own namespace: %v", err)
	}
	// The platform boundary does not check key3 at all, so any namespace passes.
	if err := CheckPermissionRegistrationAt(domain.PlatformBoundary, catalog, domain.PermissionDefinition{ID: "auth:client::read", Active: true}); err != nil {
		t.Fatalf("platform boundary compared against an application: %v", err)
	}
	// A boundary that is not one of the three is malformed, not silently
	// treated as the permissive case.
	if err := CheckPermissionRegistrationAt(domain.Boundary("elsewhere"), catalog, domain.PermissionDefinition{ID: "hrms:x::read", Active: true}); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("an unknown boundary gave %v, want ErrMalformed", err)
	}
}
