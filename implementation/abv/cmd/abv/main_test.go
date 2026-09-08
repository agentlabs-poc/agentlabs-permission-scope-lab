package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompiledBinarySeedInspectCheckAssignAndReopen(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "abv")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, "./cmd/abv")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	database := filepath.Join(t.TempDir(), "workflow.db")
	a2 := filepath.Join(root, "testdata", "a2.json")
	run := func(want int, args ...string) (string, string) {
		t.Helper()
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
	run(0, "scenario", "seed", "team-fin-c17", "--db", database, "--tenant", "acme", "--app", "hrms")
	grant, _ := run(0, "inspect", "grant", "G1", "--db", database, "--tenant", "acme", "--app", "hrms")
	if grant != `{"version":"1","grant_id":"G1","revision":1,"parent_grant_id":"G0","permissions":["hrms:payroll:payslip::read","hrms:payroll:payslip::write"],"scope":{"dept":"FIN"}}`+"\n" {
		t.Fatalf("grant output = %q", grant)
	}
	run(0, "inspect", "assignment", "A1", "--db", database, "--tenant", "acme", "--app", "hrms")
	diagnosis, _ := run(0, "check", "assignment", "--file", a2, "--db", database, "--tenant", "acme", "--app", "hrms")
	if !strings.Contains(diagnosis, "does not authorize a later write") {
		t.Fatalf("diagnosis = %q", diagnosis)
	}
	run(0, "assign", "--file", a2, "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	after, _ := run(0, "inspect", "assignment", "A2", "--db", database, "--tenant", "acme", "--app", "hrms")
	want, _ := os.ReadFile(a2)
	if after != strings.TrimSpace(string(want))+"\n" {
		t.Fatalf("reopened A2 = %q", after)
	}
	run(0, "grant", "disable", "G2", "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	control, _ := run(0, "inspect", "grant-control", "G2", "--db", database, "--tenant", "acme", "--app", "hrms")
	if control != `{"version":"1","id":"G2","status":"disabled"}`+"\n" {
		t.Fatalf("disabled control = %q", control)
	}
	run(3, "check", "assignment", "--file", a2, "--db", database, "--tenant", "acme", "--app", "hrms")
	run(0, "grant", "enable", "G2", "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	control, _ = run(0, "inspect", "grant-control", "G2", "--db", database, "--tenant", "acme", "--app", "hrms")
	if control != `{"version":"1","id":"G2","status":"enabled"}`+"\n" {
		t.Fatalf("enabled control = %q", control)
	}
	run(0, "check", "assignment", "--file", a2, "--db", database, "--tenant", "acme", "--app", "hrms")
	reopened, _ := run(0, "inspect", "assignment", "A2", "--db", database, "--tenant", "acme", "--app", "hrms")
	if reopened != after {
		t.Fatalf("grant status changed assignment: before=%q after=%q", after, reopened)
	}
	testCompiledBinaryNegativeCases(t, binary, root)
}

func testCompiledBinaryNegativeCases(t *testing.T, binary, root string) {
	run := func(want int, args ...string) {
		t.Helper()
		command := exec.CommandContext(t.Context(), binary, args...)
		if output, err := command.CombinedOutput(); (err == nil && want != 0) || (err != nil && err.(*exec.ExitError).ExitCode() != want) {
			t.Fatalf("%q: want %d: %v\n%s", args, want, err, output)
		}
	}
	dir := t.TempDir()
	database := filepath.Join(dir, "negative.db")
	run(5, "scenario", "seed", "unknown", "--db", database, "--tenant", "acme", "--app", "hrms")
	run(0, "scenario", "run", "team-fin-c17", "--case", "unsupported-permission", "--db", database, "--tenant", "acme", "--app", "hrms")
	run(4, "assign", "--file", filepath.Join(dir, "missing.json"), "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	run(3, "inspect", "assignment", "A2", "--db", database, "--tenant", "acme", "--app", "hrms")
	run(3, "inspect", "grant", "G1", "--db", database, "--tenant", "acme", "--app", "other")
	run(3, "inspect", "grant", "G1", "--db", database, "--tenant", "other", "--app", "hrms")
	run(4, "scenario", "seed", "team-fin-c17", "--db", database, "--tenant", "acme", "--app", "hrms")
	duplicate := filepath.Join(dir, "duplicate.json")
	if err := os.WriteFile(duplicate, []byte(`{"version":"1","id":"A2","id":"other","grant_id":"G2","grant_revision":1,"recipient":{"type":"group","id":"Team2"},"status":"enabled"}`), 0600); err != nil {
		t.Fatal(err)
	}
	run(2, "check", "assignment", "--file", duplicate, "--db", database, "--tenant", "acme", "--app", "hrms")
	run(3, "assign", "--file", filepath.Join(root, "testdata", "a2.json"), "--fixture-context", "unknown", "--db", database, "--tenant", "acme", "--app", "hrms")
	run(3, "inspect", "assignment", "A2", "--db", database, "--tenant", "acme", "--app", "hrms")
}
