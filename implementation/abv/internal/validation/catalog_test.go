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
