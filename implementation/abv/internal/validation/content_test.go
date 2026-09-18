package validation_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/validation"
	"errors"
	"slices"
	"testing"
)

// The write rule and the read rule differ in exactly one place, and it had no
// assertion: flipping requireAll left the whole module green while two write
// paths began admitting a grant that names a retired permission.
//
// A mixed selection is what distinguishes them. Every earlier test retired a
// grant's *whole* selection, where both rules agree.
func TestTheWriteRuleRefusesAMixedSelectionTheReadRuleNarrows(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	catalog := domain.Catalog{
		ApplicationID: "hrms",
		Permissions: map[string]domain.PermissionDefinition{
			"hrms:a::read":  {ID: "hrms:a::read", Active: true, Boundary: domain.ApplicationBoundary, Namespace: "hrms"},
			"hrms:b::write": {ID: "hrms:b::write", Active: false, Boundary: domain.ApplicationBoundary, Namespace: "hrms"},
		},
		Scopes: map[string]domain.ScopeDefinition{"dept": {Key: "dept"}},
	}
	mixed := domain.GrantContent{
		Version: "1", GrantID: "fk3x9r2m5iv8", Revision: 1, ParentGrantID: "fk3x9r2m0dq3",
		Permissions: []string{"hrms:a::read", "hrms:b::write"},
		Scope:       map[string]string{"dept": "FIN"},
	}
	// The write rule: nothing new may reference a retired permission.
	if err := validation.CheckContent(area, catalog, mixed, nil); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("the write rule accepted a retired permission: %v", err)
	}
	// The read rule: what is still supplied keeps working.
	supplied, err := validation.SuppliedContent(area, catalog, mixed, nil)
	if err != nil {
		t.Fatalf("the read rule closed a grant that still supplies read: %v", err)
	}
	if !slices.Equal(supplied, []string{"hrms:a::read"}) {
		t.Fatalf("supplied = %v, want only the active permission", supplied)
	}
}

// A grant whose whole selection is retired is rejected by BOTH rules, and that
// rejection had no test: deleting it left the module green and produced a phantom
// entitlement — a route with no permissions but a full scope and grant chain.
func TestBothRulesRejectASelectionTheCatalogHasEmptied(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	catalog := domain.Catalog{
		ApplicationID: "hrms",
		Permissions: map[string]domain.PermissionDefinition{
			"hrms:b::write": {ID: "hrms:b::write", Active: false, Boundary: domain.ApplicationBoundary, Namespace: "hrms"},
		},
		Scopes: map[string]domain.ScopeDefinition{"dept": {Key: "dept"}},
	}
	emptied := domain.GrantContent{
		Version: "1", GrantID: "fk3x9r2m5iv8", Revision: 1, ParentGrantID: "fk3x9r2m0dq3",
		Permissions: []string{"hrms:b::write"},
		Scope:       map[string]string{"dept": "FIN"},
	}
	if err := validation.CheckContent(area, catalog, emptied, nil); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("write rule: %v", err)
	}
	if got, err := validation.SuppliedContent(area, catalog, emptied, nil); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("read rule accepted an emptied selection: %v, %v", got, err)
	}
}
