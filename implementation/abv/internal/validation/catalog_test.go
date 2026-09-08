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
		keys []string
		want error
	}{
		{"valid", valid, []string{"dept"}, nil},
		{"malformed ID", domain.PermissionDefinition{ID: "bad*", Active: true}, nil, domain.ErrMalformed},
		{"inactive", domain.PermissionDefinition{ID: valid.ID}, nil, domain.ErrRejected},
		{"duplicate", domain.PermissionDefinition{ID: read, Active: true}, nil, domain.ErrConflict},
		{"unknown key", valid, []string{"missing"}, domain.ErrRejected},
		{"duplicate key", valid, []string{"dept", "dept"}, domain.ErrRejected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := CheckPermissionRegistration(c, tc.def, tc.keys); !errors.Is(err, tc.want) {
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
		{"valid", domain.ScopeDefinition{Key: "region", AllowedTokens: []string{"$self"}}, nil},
		{"duplicate", domain.ScopeDefinition{Key: "dept", AllowedTokens: []string{}}, domain.ErrConflict},
		{"malformed key", domain.ScopeDefinition{Key: "*", AllowedTokens: []string{}}, domain.ErrMalformed},
		{"nil tokens", domain.ScopeDefinition{Key: "region"}, domain.ErrMalformed},
		{"unsupported token", domain.ScopeDefinition{Key: "region", AllowedTokens: []string{"$owner"}}, domain.ErrRejected},
		{"duplicate token", domain.ScopeDefinition{Key: "region", AllowedTokens: []string{"$self", "$self"}}, domain.ErrRejected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := CheckScopeRegistration(c, tc.def); !errors.Is(err, tc.want) {
				t.Fatalf("got %v want %v", err, tc.want)
			}
		})
	}
}
