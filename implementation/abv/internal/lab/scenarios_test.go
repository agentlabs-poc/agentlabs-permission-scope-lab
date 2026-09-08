package lab

import (
	"agentlabs.local/abv/domain"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSeedIsNewOnlyAndLeavesExistingFileUntouched(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	path := filepath.Join(t.TempDir(), "existing.db")
	if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := (Scenarios{}).Seed(t.Context(), area, "team-fin-c17", path); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("existing seed error = %v", err)
	}
	if raw, _ := os.ReadFile(path); string(raw) != "keep" {
		t.Fatalf("existing file changed: %q", raw)
	}
}

func TestSeedRejectsUnknownScenarioWithoutCreatingFile(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	path := filepath.Join(t.TempDir(), "unknown.db")
	if err := (Scenarios{}).Seed(t.Context(), area, "unknown", path); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("error = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("unknown scenario created file: %v", err)
	}
}

func TestUnsupportedPermissionScenarioObservesRejectionAndNoWrite(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	path := filepath.Join(t.TempDir(), "negative.db")
	if err := (Scenarios{}).Run(t.Context(), area, "team-fin-c17", "unsupported-permission", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	defer closeConnection()
	if _, err = api.Inspect(t.Context(), area, "assignment", "A2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("A2 inspection = %v", err)
	}
}

func TestSeedSupportsLiteralFilesystemPath(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	path := filepath.Join(t.TempDir(), "scenario #1.db")
	if err := (Scenarios{}).Seed(t.Context(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	if err = closeConnection(); err != nil {
		t.Fatal(err)
	}
	if api == nil {
		t.Fatal("nil API")
	}
}
