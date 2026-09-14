package registry_test

import (
	"agentlabs.local/registry"
	"agentlabs.local/registry/domain"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

type clock struct{}

func (clock) Now() time.Time { return time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC) }

// admin admits one identity and refuses everything else, so the gates are
// exercised rather than assumed.
type admin struct{ deny bool }

func (a admin) CheckApplicationWrite(_ context.Context, i domain.Identity, _ string, _ time.Time) error {
	return a.check(i)
}
func (a admin) CheckApplicationRead(_ context.Context, i domain.Identity, _ time.Time) error {
	return a.check(i)
}
func (a admin) CheckInstallationWrite(_ context.Context, i domain.Identity, _, _ string, _ time.Time) error {
	return a.check(i)
}
func (a admin) check(i domain.Identity) error {
	if a.deny || i.HumanID != "fi7io4lvjqio" {
		return domain.ErrRejected
	}
	return nil
}

var platform = domain.Identity{Version: "1", HumanID: "fi7io4lvjqio"}

func open(t *testing.T, deny bool) *registry.Facade {
	t.Helper()
	f, err := registry.Open(t.Context(), filepath.Join(t.TempDir(), "registry.db"), admin{deny: deny}, clock{}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}

// A slug is the id, and registering one is add-only: it appears in the canonical
// path of every record that application owns, so reusing it would re-point them.
func TestRegisterApplicationIsAddOnlyAndValidatesTheSlug(t *testing.T) {
	f := open(t, false)
	app, err := f.RegisterApplication(t.Context(), platform, "hrms", "HRMS")
	if err != nil || app.Slug != "hrms" || app.Status != domain.StatusActive {
		t.Fatalf("app=%#v err=%v", app, err)
	}
	if _, err := f.RegisterApplication(t.Context(), platform, "hrms", "Again"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("re-registering a slug gave %v, want ErrConflict", err)
	}
	// The shape agentlabs-auth enforces on its own slugs.
	for _, bad := range []string{"HRMS", "1hrms", "h", "hrms_payroll", "hrms.payroll", "", "hrms "} {
		if _, err := f.RegisterApplication(t.Context(), platform, bad, "X"); !errors.Is(err, domain.ErrMalformed) {
			t.Fatalf("accepted slug %q: %v", bad, err)
		}
	}
}

// Status is reversible and never creates.
func TestSetApplicationStatusIsReversibleAndNeverCreates(t *testing.T) {
	f := open(t, false)
	if _, err := f.RegisterApplication(t.Context(), platform, "hrms", "HRMS"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{domain.StatusSuspended, domain.StatusActive} {
		got, err := f.SetApplicationStatus(t.Context(), platform, "hrms", want)
		if err != nil || got.Status != want || got.Name != "HRMS" {
			t.Fatalf("status %q gave %#v err=%v", want, got, err)
		}
	}
	if _, err := f.SetApplicationStatus(t.Context(), platform, "absent", domain.StatusSuspended); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a status change created an application: %v", err)
	}
	if _, err := f.SetApplicationStatus(t.Context(), platform, "hrms", "retired"); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("an unknown status was accepted: %v", err)
	}
}

// Install requires the application to exist — the foreign key a single schema
// gave for free, once the domains are separate.
func TestInstallRequiresTheApplicationAndIsAddOnly(t *testing.T) {
	f := open(t, false)
	if _, err := f.RegisterApplication(t.Context(), platform, "hrms", "HRMS"); err != nil {
		t.Fatal(err)
	}
	if err := f.Install(t.Context(), platform, "acme", "absent"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("installed an application that does not exist: %v", err)
	}
	if err := f.Install(t.Context(), platform, "acme", "hrms"); err != nil {
		t.Fatal(err)
	}
	if err := f.Install(t.Context(), platform, "acme", "hrms"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("installing twice gave %v, want ErrConflict", err)
	}
	held, err := f.IsInstalled(t.Context(), platform, "acme", "hrms")
	if err != nil || !held {
		t.Fatalf("installed=%v err=%v", held, err)
	}
	if err := f.Uninstall(t.Context(), platform, "acme", "hrms"); err != nil {
		t.Fatal(err)
	}
	if err := f.Uninstall(t.Context(), platform, "acme", "hrms"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("uninstalling twice gave %v, want ErrNotFound", err)
	}
	held, err = f.IsInstalled(t.Context(), platform, "acme", "hrms")
	if err != nil || held {
		t.Fatalf("still installed after uninstall: %v err=%v", held, err)
	}
}

// The listing answers in both directions and refuses to answer neither.
func TestListInstallationsAnswersBothDirections(t *testing.T) {
	f := open(t, false)
	for _, slug := range []string{"hrms", "payroll"} {
		if _, err := f.RegisterApplication(t.Context(), platform, slug, slug); err != nil {
			t.Fatal(err)
		}
	}
	for _, pair := range [][2]string{{"acme", "hrms"}, {"acme", "payroll"}, {"globex", "hrms"}} {
		if err := f.Install(t.Context(), platform, pair[0], pair[1]); err != nil {
			t.Fatal(err)
		}
	}
	byTenant, err := f.ListInstallations(t.Context(), platform, domain.InstallationFilter{TenantID: "acme"})
	if err != nil || byTenant.Total != 2 {
		t.Fatalf("acme holds %d err=%v, want 2", byTenant.Total, err)
	}
	bySlug, err := f.ListInstallations(t.Context(), platform, domain.InstallationFilter{Slug: "hrms"})
	if err != nil || bySlug.Total != 2 {
		t.Fatalf("hrms is held by %d err=%v, want 2", bySlug.Total, err)
	}
	for _, bad := range []domain.InstallationFilter{{}, {TenantID: "acme", Slug: "hrms"}} {
		if _, err := f.ListInstallations(t.Context(), platform, bad); !errors.Is(err, domain.ErrMalformed) {
			t.Fatalf("filter %#v gave %v, want ErrMalformed", bad, err)
		}
	}
}

// Every operation is gated, and a denied gate writes nothing.
func TestEveryOperationIsGated(t *testing.T) {
	f := open(t, true)
	if _, err := f.RegisterApplication(t.Context(), platform, "hrms", "HRMS"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("register past the gate: %v", err)
	}
	if err := f.Install(t.Context(), platform, "acme", "hrms"); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("install past the gate: %v", err)
	}
	if _, err := f.ListApplications(t.Context(), platform, domain.ApplicationFilter{}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("list past the gate: %v", err)
	}
	// An identity that is not a direct human never reaches the gate at all.
	if _, err := f.GetApplication(t.Context(), domain.Identity{}, "hrms"); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("a malformed identity gave %v, want ErrUnsupported", err)
	}
}

// The port narrows this domain to the two questions Auth-AL asks, and a
// suspended application is reported absent — a policy that lives at the seam.
func TestPortAnswersTheTwoQuestionsAndMapsSuspended(t *testing.T) {
	f := open(t, false)
	if _, err := f.RegisterApplication(t.Context(), platform, "hrms", "HRMS"); err != nil {
		t.Fatal(err)
	}
	if err := f.Install(t.Context(), platform, "acme", "hrms"); err != nil {
		t.Fatal(err)
	}
	port, err := registry.NewPort(f, platform)
	if err != nil {
		t.Fatal(err)
	}
	exists, err := port.ApplicationExists(t.Context(), "hrms")
	if err != nil || !exists {
		t.Fatalf("exists=%v err=%v", exists, err)
	}
	// Absent is false and not an error: Auth-AL asks a question, not for a record.
	exists, err = port.ApplicationExists(t.Context(), "absent")
	if err != nil || exists {
		t.Fatalf("absent application: exists=%v err=%v", exists, err)
	}
	if _, err := f.SetApplicationStatus(t.Context(), platform, "hrms", domain.StatusSuspended); err != nil {
		t.Fatal(err)
	}
	exists, err = port.ApplicationExists(t.Context(), "hrms")
	if err != nil || exists {
		t.Fatalf("a suspended application still reported present: exists=%v err=%v", exists, err)
	}
	held, err := port.Installed(t.Context(), "acme", "hrms")
	if err != nil || !held {
		t.Fatalf("installed=%v err=%v", held, err)
	}
	held, err = port.Installed(t.Context(), "globex", "hrms")
	if err != nil || held {
		t.Fatalf("another tenant reported installed: %v err=%v", held, err)
	}
}

// Disabled and uninstalled are different things: disable is reversible and keeps
// the record, uninstall destroys it. That distinction is the reason tenant status
// exists at all.
func TestDisableIsReversibleAndNotUninstall(t *testing.T) {
	f := open(t, false)
	if _, err := f.RegisterApplication(t.Context(), platform, "hrms", "HRMS"); err != nil {
		t.Fatal(err)
	}
	if err := f.Install(t.Context(), platform, "acme", "hrms"); err != nil {
		t.Fatal(err)
	}
	// Created enabled.
	held, err := f.IsInstalled(t.Context(), platform, "acme", "hrms")
	if err != nil || !held {
		t.Fatalf("a new installation is not enabled: %v err=%v", held, err)
	}
	// Disabled: the tenant may not use it, but the record survives.
	if _, err := f.SetInstallationStatus(t.Context(), platform, "acme", "hrms", domain.StatusDisabled); err != nil {
		t.Fatal(err)
	}
	held, err = f.IsInstalled(t.Context(), platform, "acme", "hrms")
	if err != nil || held {
		t.Fatalf("a disabled installation still reports usable: %v err=%v", held, err)
	}
	got, err := f.GetInstallation(t.Context(), platform, "acme", "hrms")
	if err != nil || got.Status != domain.StatusDisabled {
		t.Fatalf("the record did not survive disabling: %#v err=%v", got, err)
	}
	// And it comes back.
	if _, err := f.SetInstallationStatus(t.Context(), platform, "acme", "hrms", domain.StatusEnabled); err != nil {
		t.Fatal(err)
	}
	if held, _ := f.IsInstalled(t.Context(), platform, "acme", "hrms"); !held {
		t.Fatal("re-enabling did not restore it")
	}
	// Install over an existing installation is a conflict, not a quiet
	// re-enable: install creates the relationship, status changes it.
	if _, err := f.SetInstallationStatus(t.Context(), platform, "acme", "hrms", domain.StatusDisabled); err != nil {
		t.Fatal(err)
	}
	if err := f.Install(t.Context(), platform, "acme", "hrms"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("install over a disabled installation gave %v, want ErrConflict", err)
	}
	// Status never creates.
	if _, err := f.SetInstallationStatus(t.Context(), platform, "globex", "hrms", domain.StatusEnabled); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a status change created an installation: %v", err)
	}
	if _, err := f.SetInstallationStatus(t.Context(), platform, "acme", "hrms", "paused"); !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("an unknown installation status was accepted: %v", err)
	}
}

// The port maps two facts to one bit: the application must be active AND the
// installation enabled. Auth-AL never learns there are two ways to answer no.
func TestPortRequiresBothActiveAndEnabled(t *testing.T) {
	f := open(t, false)
	if _, err := f.RegisterApplication(t.Context(), platform, "hrms", "HRMS"); err != nil {
		t.Fatal(err)
	}
	if err := f.Install(t.Context(), platform, "acme", "hrms"); err != nil {
		t.Fatal(err)
	}
	port, err := registry.NewPort(f, platform)
	if err != nil {
		t.Fatal(err)
	}
	check := func(wantExists, wantHeld bool, why string) {
		t.Helper()
		exists, err := port.ApplicationExists(t.Context(), "hrms")
		if err != nil || exists != wantExists {
			t.Fatalf("%s: exists=%v want %v err=%v", why, exists, wantExists, err)
		}
		held, err := port.Installed(t.Context(), "acme", "hrms")
		if err != nil || held != wantHeld {
			t.Fatalf("%s: installed=%v want %v err=%v", why, held, wantHeld, err)
		}
	}
	check(true, true, "active and enabled")

	// Platform-wide suspension: every tenant loses it.
	if _, err := f.SetApplicationStatus(t.Context(), platform, "hrms", domain.StatusSuspended); err != nil {
		t.Fatal(err)
	}
	check(false, true, "suspended application")

	// Tenant-side disable: only this tenant, and the application is fine.
	if _, err := f.SetApplicationStatus(t.Context(), platform, "hrms", domain.StatusActive); err != nil {
		t.Fatal(err)
	}
	if _, err := f.SetInstallationStatus(t.Context(), platform, "acme", "hrms", domain.StatusDisabled); err != nil {
		t.Fatal(err)
	}
	check(true, false, "disabled installation")
}
