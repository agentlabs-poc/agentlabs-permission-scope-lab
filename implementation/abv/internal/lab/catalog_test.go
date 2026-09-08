package lab

import (
	"agentlabs.local/abv/domain"
	"errors"
	"path/filepath"
	"testing"
)

func TestCatalogPublisherRegistersAndTenantAdministratorCannot(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	app, _ := domain.NewApplication("hrms")
	path := filepath.Join(t.TempDir(), "lab.db")
	if err := (Scenarios{}).Seed(t.Context(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := ConnectCatalog(t.Context(), app, path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = api.RegisterScope(t.Context(), app, domain.FixtureContext{Name: "maya-team1"}, domain.ScopeDefinition{Key: "region", AllowedTokens: []string{}}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("maya error=%v", err)
	}
	if _, err = api.RegisterScope(t.Context(), app, domain.FixtureContext{Name: "application-publisher"}, domain.ScopeDefinition{Key: "region", AllowedTokens: []string{}}); err != nil {
		t.Fatal(err)
	}
	if err = closeConnection(); err != nil {
		t.Fatal(err)
	}
	ordinary, closeOrdinary, err := Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	record, err := ordinary.Inspect(t.Context(), area, "scope", "region")
	if err != nil || len(record.Rows) == 0 {
		t.Fatalf("scope=%+v err=%v", record, err)
	}
	if err = closeOrdinary(); err != nil {
		t.Fatal(err)
	}
}
