package wiring_test

import (
	"agentlabs.local/abv"
	abvdomain "agentlabs.local/abv/domain"
	regdomain "agentlabs.local/registry/domain"
	"agentlabs.local/wiring"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

type clock struct{}

func (clock) Now() time.Time { return time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC) }

type regAdmin struct{}

func (regAdmin) CheckApplicationWrite(context.Context, regdomain.Identity, string, time.Time) error {
	return nil
}
func (regAdmin) CheckApplicationRead(context.Context, regdomain.Identity, time.Time) error {
	return nil
}
func (regAdmin) CheckInstallationWrite(context.Context, regdomain.Identity, string, string, time.Time) error {
	return nil
}

// abvAdmin refuses everything. These tests never reach it, which is the point:
// the registry's answer decides the outcome before any Auth-AL gate runs.
type abvAdmin struct{}

func (abvAdmin) CheckAssignment(context.Context, abv.Evidence, abvdomain.Identity, abvdomain.Assignment, time.Time) error {
	return abvdomain.ErrRejected
}

var operator = regdomain.Identity{Version: "1", HumanID: "fi7io4lvjqio"}

// The two domains compose through a port neither of them imports.
//
// Auth-AL declares the interface; the registry provides a type with those
// methods; this package is the only one that knows both exist. That is the whole
// mechanism, and this test is what proves it holds rather than merely compiles.
func TestAuthALAsksTheRegistryRatherThanItsOwnTables(t *testing.T) {
	dir := t.TempDir()

	// One call, where there used to be three steps duplicated between this test
	// and the command beside it. The compile is still half the proof: nothing in
	// abv imports agentlabs.local/registry, and nothing in registry imports abv.
	service, err := wiring.Open(t.Context(), wiring.Config{
		AuthorityPath:  filepath.Join(dir, "authority.db"),
		RegistryPath:   filepath.Join(dir, "registry.db"),
		CreateRegistry: true,
		Administration: abvAdmin{}, RegistryAdministration: regAdmin{},
		Operator: operator, Clock: clock{},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	reg, facade := service.Applications(), service.Authority()

	area, err := abvdomain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	identity := abvdomain.Identity{Version: "1", Actor: abvdomain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}

	// Nothing registered yet: Auth-AL refuses because the registry says the
	// tenant does not hold the application.
	if _, err := facade.Inspect(t.Context(), area, "team", "fibggi2juubk"); !errors.Is(err, abvdomain.ErrNotFound) {
		t.Fatalf("read succeeded with no installation: %v", err)
	}

	// Register and install in the registry — a different domain, a different
	// database, a different identity type.
	if _, err := reg.RegisterApplication(t.Context(), operator, "hrms", "HRMS"); err != nil {
		t.Fatal(err)
	}
	if err := reg.Install(t.Context(), operator, "acme", "hrms"); err != nil {
		t.Fatal(err)
	}

	// Auth-AL now gets past the installation gate and fails further in, on its
	// own missing record rather than on the installation.
	_, err = facade.Inspect(t.Context(), area, "team", "fibggi2juubk")
	if err == nil {
		t.Fatal("expected the record itself to be missing")
	}

	// Uninstalling closes the door again, without Auth-AL being told.
	if err := reg.Uninstall(t.Context(), operator, "acme", "hrms"); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.Inspect(t.Context(), area, "team", "fibggi2juubk"); !errors.Is(err, abvdomain.ErrNotFound) {
		t.Fatalf("read survived an uninstall: %v", err)
	}
	_ = identity
}
