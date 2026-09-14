package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The CLI had no Go test at all — only a shell demonstration, which CI does not
// run. That is how `registry --db PATH` shipped able to panic on an empty verb.
// This exercises the compiled binary the way abv's own CLI test does, and asserts
// the exit codes, because a demonstration reads them and so does a script.
func TestCompiledBinaryRegisterInstallStatusAndExitCodes(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "registry")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, "./cmd/registry")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	database := filepath.Join(t.TempDir(), "registry.db")
	run := func(want int, args ...string) (string, string) {
		t.Helper()
		args = append(args, "--db", database)
		command := exec.CommandContext(context.Background(), binary, args...)
		var stdout, stderr strings.Builder
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		got := 0
		if exit, ok := err.(*exec.ExitError); ok {
			got = exit.ExitCode()
		} else if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%q exit %d, want %d; stdout=%q stderr=%q", args, got, want, stdout.String(), stderr.String())
		}
		return stdout.String(), stderr.String()
	}

	// A verb is required, and its absence must not panic or create the database.
	if _, diag := run(2); !strings.Contains(diag, "usage: registry") {
		t.Fatalf("no verb gave %q, want the usage line", diag)
	}

	out, _ := run(0, "register", "hrms", "--name", "HRMS")
	if !strings.Contains(out, "slug  hrms") || !strings.Contains(out, "status  active") {
		t.Fatalf("register printed %q", out)
	}
	// Add-only: the slug is the id and is never reassigned.
	if _, diag := run(4, "register", "hrms", "--name", "Again"); !strings.Contains(diag, "conflict") {
		t.Fatalf("re-register printed %q, want a conflict", diag)
	}
	if _, diag := run(2, "register", "HRMS", "--name", "Uppercase"); !strings.Contains(diag, "malformed") {
		t.Fatalf("uppercase slug printed %q, want malformed", diag)
	}

	// Install requires the application: the foreign key, as a rule across domains.
	if _, diag := run(3, "install", "absent", "--tenant", "acme"); !strings.Contains(diag, "not found") {
		t.Fatalf("installing an unregistered application printed %q", diag)
	}
	run(0, "install", "hrms", "--tenant", "acme")
	run(4, "install", "hrms", "--tenant", "acme")

	// Disable is reversible and is not uninstall — the record survives it.
	run(0, "disable", "hrms", "--tenant", "acme")
	out, _ = run(0, "installations", "--tenant", "acme")
	if !strings.Contains(out, "total  1") || !strings.Contains(out, "disabled") {
		t.Fatalf("after disable the listing printed %q, want the record still there", out)
	}
	run(0, "enable", "hrms", "--tenant", "acme")

	// An unfiltered installation listing is not a question anyone asks.
	if _, diag := run(2, "installations"); !strings.Contains(diag, "malformed") {
		t.Fatalf("unfiltered installations printed %q, want malformed", diag)
	}

	// Suspend is a status change, never a delete, and never a create.
	out, _ = run(0, "status", "hrms", "--set", "suspended")
	if !strings.Contains(out, "status  suspended") {
		t.Fatalf("suspend printed %q", out)
	}
	if _, diag := run(3, "status", "absent", "--set", "suspended"); !strings.Contains(diag, "not found") {
		t.Fatalf("suspending an absent application printed %q", diag)
	}
	out, _ = run(0, "list", "--status", "suspended")
	if !strings.Contains(out, "total  1") || !strings.Contains(out, "hrms") {
		t.Fatalf("filtered list printed %q", out)
	}
	run(0, "status", "hrms", "--set", "active")

	// Uninstall says whether it did anything.
	run(0, "uninstall", "hrms", "--tenant", "acme")
	if _, diag := run(3, "uninstall", "hrms", "--tenant", "acme"); !strings.Contains(diag, "not found") {
		t.Fatalf("uninstalling twice printed %q", diag)
	}
}
