package lab

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const a2JSON = `{"version":"1","id":"A2","grant_id":"G2","grant_revision":1,"recipient":{"type":"group","id":"Team2"},"status":"enabled"}`

func TestConnectDoesNotCreateMissingDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.db")
	area, _ := domain.NewArea("acme", "hrms")
	if _, _, err := Connect(t.Context(), area, path); err == nil {
		t.Fatal("missing database connected")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("ordinary connect created database: %v", err)
	}
}

func TestGenericABVDatabaseCannotUseFixtureIdentity(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	path := filepath.Join(t.TempDir(), "generic.db")
	fixture := TeamFINC17(area)
	provider, err := CreateSQLite(t.Context(), path, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if err = provider.Close(); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeConnection()
	if _, err = api.Assign(t.Context(), area, domain.FixtureContext{Name: "maya-team1"}, []byte(a2JSON)); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("unmarked database assignment error = %v", err)
	}
}

func TestMarkerBindsExactAreaAndFixtureContext(t *testing.T) {
	seedArea, _ := domain.NewArea("acme", "hrms")
	wrongArea, _ := domain.NewArea("other", "hrms")
	path := filepath.Join(t.TempDir(), "lab.db")
	if err := (Scenarios{}).Seed(t.Context(), seedArea, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		area    domain.Area
		fixture string
	}{
		{seedArea, "maya"}, {wrongArea, "maya-team1"},
	} {
		api, closeConnection, err := Connect(t.Context(), tc.area, path)
		if err != nil {
			t.Fatal(err)
		}
		_, assignErr := api.Assign(t.Context(), tc.area, domain.FixtureContext{Name: tc.fixture}, []byte(a2JSON))
		if closeErr := closeConnection(); closeErr != nil {
			t.Fatal(closeErr)
		}
		if !errors.Is(assignErr, domain.ErrRejected) {
			t.Fatalf("%q assignment error = %v", tc.fixture, assignErr)
		}
	}
}

func TestUnsupportedMarkerFormatCannotUseFixtureIdentity(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	path := filepath.Join(t.TempDir(), "lab.db")
	if err := (Scenarios{}).Seed(t.Context(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`UPDATE abv_lab_metadata SET format_version=2`); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	_, assignErr := api.Assign(t.Context(), area, domain.FixtureContext{Name: "maya-team1"}, []byte(a2JSON))
	if err = closeConnection(); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(assignErr, domain.ErrUnsupported) {
		t.Fatalf("assignment error = %v", assignErr)
	}
}
