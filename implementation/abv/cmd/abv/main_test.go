package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompiledBinaryGrantPublicationSeedInspectCheckAssignAndReopen(t *testing.T) {
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
	roleOutput, roleWarning := run(0, "role", "publish", "--id", "fi9jvxobqsxs", "--name", "payslip-reader", "--revision", "2", "--permissions", "hrms:payroll:payslip::read,hrms:payroll:payslip::write", "--tenant", "acme", "--app", "hrms", "--db", database, "--fixture-context", "maya-role-publisher")
	if !strings.Contains(roleOutput, "revision  2") || !strings.Contains(roleWarning, "does not grant business access") {
		t.Fatalf("role stdout=%q stderr=%q", roleOutput, roleWarning)
	}
	reopenedRoles, _ := run(0, "inspect", "role", "fi9jvxobqsxs", "--tenant", "acme", "--app", "hrms", "--db", database)
	if !strings.Contains(reopenedRoles, "1         hrms:payroll:payslip::read") || !strings.Contains(reopenedRoles, "2         hrms:payroll:payslip::read,hrms:payroll:payslip::write") {
		t.Fatalf("roles=%q", reopenedRoles)
	}
	g0Before, _ := run(0, "inspect", "grant", "fk3x9r2m0dq3", "--db", database, "--tenant", "acme", "--app", "hrms")
	grant, _ := run(0, "inspect", "grant", "fk3x9r2m5iv8", "--db", database, "--tenant", "acme", "--app", "hrms")
	if grant != `{"version":"1","grant_id":"fk3x9r2m5iv8","revision":1,"parent_grant_id":"fk3x9r2m0dq3","permissions":["hrms:payroll:payslip::read","hrms:payroll:payslip::write"],"scope":{"dept":"FIN"}}`+"\n" {
		t.Fatalf("grant output = %q", grant)
	}
	scopeOutput, warning := run(0, "catalog", "register-scope", "region", "--app", "hrms", "--db", database, "--fixture-context", "application-publisher")
	if !strings.Contains(scopeOutput, "internal projection: scope") || !strings.Contains(scopeOutput, "key  region") || !strings.Contains(warning, "LAB ONLY") {
		t.Fatalf("scope stdout=%q stderr=%q", scopeOutput, warning)
	}
	run(0, "catalog", "register-scope", "owner", "--app", "hrms", "--db", database, "--fixture-context", "application-publisher")
	permissionOutput, _ := run(0, "catalog", "register-permission", "hrms:payroll:payslip::export", "--app", "hrms", "--db", database, "--fixture-context", "application-publisher")
	if !strings.Contains(permissionOutput, "internal projection: permission") || !strings.Contains(permissionOutput, "active  true") {
		t.Fatalf("permission stdout=%q", permissionOutput)
	}
	persistedScope, _ := run(0, "inspect", "scope", "owner", "--db", database, "--tenant", "acme", "--app", "hrms")
	// The full permission contract, end to end against the real SQLite store.
	fetched, _ := run(0, "catalog", "get-permission", "hrms:payroll:payslip::export", "--app", "hrms", "--db", database, "--fixture-context", "application-publisher")
	if !strings.Contains(fetched, "id  hrms:payroll:payslip::export") || !strings.Contains(fetched, "active  true") {
		t.Fatalf("get-permission=%q", fetched)
	}
	run(3, "catalog", "get-permission", "hrms:payroll:payslip::nothing", "--app", "hrms", "--db", database, "--fixture-context", "application-publisher")

	listed, _ := run(0, "catalog", "list-permissions", "--app", "hrms", "--db", database, "--fixture-context", "application-publisher")
	if !strings.Contains(listed, "hrms:payroll:payslip::export  active=true") {
		t.Fatalf("list-permissions=%q", listed)
	}
	// A prefix that matches nothing is an empty page, never a fallback to all.
	narrowed, _ := run(0, "catalog", "list-permissions", "--prefix", "codehost:", "--app", "hrms", "--db", database, "--fixture-context", "application-publisher")
	if !strings.Contains(narrowed, "count  0") {
		t.Fatalf("empty prefix page=%q", narrowed)
	}

	retired, _ := run(0, "catalog", "set-permission-status", "hrms:payroll:payslip::export", "--active", "false", "--app", "hrms", "--db", database, "--fixture-context", "application-publisher")
	if !strings.Contains(retired, "active  false") {
		t.Fatalf("retire=%q", retired)
	}
	// Retirement survives a fresh process and is visible to active-only listing.
	afterRetire, _ := run(0, "catalog", "get-permission", "hrms:payroll:payslip::export", "--app", "hrms", "--db", database, "--fixture-context", "application-publisher")
	if !strings.Contains(afterRetire, "active  false") {
		t.Fatalf("retirement did not persist across processes: %q", afterRetire)
	}
	activeOnly, _ := run(0, "catalog", "list-permissions", "--active-only", "--app", "hrms", "--db", database, "--fixture-context", "application-publisher")
	if strings.Contains(activeOnly, "hrms:payroll:payslip::export") {
		t.Fatalf("retired permission still listed as active: %q", activeOnly)
	}
	// Reversible.
	restored, _ := run(0, "catalog", "set-permission-status", "hrms:payroll:payslip::export", "--active", "true", "--app", "hrms", "--db", database, "--fixture-context", "application-publisher")
	if !strings.Contains(restored, "active  true") {
		t.Fatalf("restore=%q", restored)
	}
	// A status change never creates an identifier.
	run(3, "catalog", "set-permission-status", "hrms:payroll:payslip::invented", "--active", "true", "--app", "hrms", "--db", database, "--fixture-context", "application-publisher")

	persistedPermission := fetched
	g0After, _ := run(0, "inspect", "grant", "fk3x9r2m0dq3", "--db", database, "--tenant", "acme", "--app", "hrms")
	if !strings.Contains(persistedScope, "owner") || !strings.Contains(persistedPermission, "true") || g0After != g0Before {
		t.Fatalf("reopen scope=%q permission=%q fk3x9r2m0dq3 before=%q after=%q", persistedScope, persistedPermission, g0Before, g0After)
	}
	run(0, "inspect", "assignment", "fm5b7t4p5iv8", "--db", database, "--tenant", "acme", "--app", "hrms")
	diagnosis, _ := run(0, "check", "assignment", "--file", a2, "--db", database, "--tenant", "acme", "--app", "hrms")
	if !strings.Contains(diagnosis, "does not authorize a later write") {
		t.Fatalf("diagnosis = %q", diagnosis)
	}
	run(0, "assign", "--file", a2, "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	after, _ := run(0, "inspect", "assignment", "fm5b7t4pan0d", "--db", database, "--tenant", "acme", "--app", "hrms")
	want, _ := os.ReadFile(a2)
	if after != strings.TrimSpace(string(want))+"\n" {
		t.Fatalf("reopened fm5b7t4pan0d = %q", after)
	}
	run(3, "assignment", "disable", "fm5b7t4p5iv8", "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	run(0, "assignment", "disable", "fm5b7t4pan0d", "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	run(0, "assignment", "disable", "fm5b7t4p5iv8", "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	run(3, "assignment", "enable", "fm5b7t4pan0d", "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	run(0, "assignment", "enable", "fm5b7t4p5iv8", "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	a2Disabled, _ := run(0, "inspect", "assignment", "fm5b7t4pan0d", "--db", database, "--tenant", "acme", "--app", "hrms")
	if a2Disabled != `{"version":"1","id":"fm5b7t4pan0d","grant_id":"fk3x9r2man0d","grant_revision":1,"recipient":{"type":"group","id":"fibggi2juxhc"},"status":"disabled"}`+"\n" {
		t.Fatalf("fm5b7t4pan0d was changed by fm5b7t4p5iv8 enable: %q", a2Disabled)
	}
	run(0, "assignment", "enable", "fm5b7t4pan0d", "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	a1Final, _ := run(0, "inspect", "assignment", "fm5b7t4p5iv8", "--db", database, "--tenant", "acme", "--app", "hrms")
	a2Final, _ := run(0, "inspect", "assignment", "fm5b7t4pan0d", "--db", database, "--tenant", "acme", "--app", "hrms")
	if a1Final != `{"version":"1","id":"fm5b7t4p5iv8","grant_id":"fk3x9r2m5iv8","grant_revision":1,"recipient":{"type":"group","id":"fibggi2juubk"},"status":"enabled"}`+"\n" || a2Final != after {
		t.Fatalf("assignments not restored: fm5b7t4p5iv8=%q fm5b7t4pan0d=%q", a1Final, a2Final)
	}
	g1Final, _ := run(0, "inspect", "grant", "fk3x9r2m5iv8", "--db", database, "--tenant", "acme", "--app", "hrms")
	g2, _ := run(0, "inspect", "grant", "fk3x9r2man0d", "--db", database, "--tenant", "acme", "--app", "hrms")
	g1Control, _ := run(0, "inspect", "grant-control", "fk3x9r2m5iv8", "--db", database, "--tenant", "acme", "--app", "hrms")
	g2Control, _ := run(0, "inspect", "grant-control", "fk3x9r2man0d", "--db", database, "--tenant", "acme", "--app", "hrms")
	if g1Final != grant || g2 != `{"version":"1","grant_id":"fk3x9r2man0d","revision":1,"parent_grant_id":"fk3x9r2m5iv8","permissions":["hrms:payroll:payslip::read"],"scope":{"cert":"C17"}}`+"\n" || g1Control != `{"version":"1","id":"fk3x9r2m5iv8","status":"enabled"}`+"\n" || g2Control != `{"version":"1","id":"fk3x9r2man0d","status":"enabled"}`+"\n" {
		t.Fatalf("grant records changed: fk3x9r2m5iv8=%q fk3x9r2man0d=%q G1-control=%q G2-control=%q", g1Final, g2, g1Control, g2Control)
	}
	run(0, "grant", "disable", "fk3x9r2man0d", "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	control, _ := run(0, "inspect", "grant-control", "fk3x9r2man0d", "--db", database, "--tenant", "acme", "--app", "hrms")
	if control != `{"version":"1","id":"fk3x9r2man0d","status":"disabled"}`+"\n" {
		t.Fatalf("disabled control = %q", control)
	}
	run(3, "check", "assignment", "--file", a2, "--db", database, "--tenant", "acme", "--app", "hrms")
	run(0, "grant", "enable", "fk3x9r2man0d", "--fixture-context", "maya-team1", "--db", database, "--tenant", "acme", "--app", "hrms")
	control, _ = run(0, "inspect", "grant-control", "fk3x9r2man0d", "--db", database, "--tenant", "acme", "--app", "hrms")
	if control != `{"version":"1","id":"fk3x9r2man0d","status":"enabled"}`+"\n" {
		t.Fatalf("enabled control = %q", control)
	}
	run(0, "check", "assignment", "--file", a2, "--db", database, "--tenant", "acme", "--app", "hrms")
	reopened, _ := run(0, "inspect", "assignment", "fm5b7t4pan0d", "--db", database, "--tenant", "acme", "--app", "hrms")
	if reopened != after {
		t.Fatalf("grant status changed assignment: before=%q after=%q", after, reopened)
	}
	g2v2 := filepath.Join(t.TempDir(), "g2-v2.json")
	g2v2JSON := `{"version":"1","grant_id":"fk3x9r2man0d","revision":2,"parent_grant_id":"fk3x9r2m5iv8","permissions":["hrms:payroll:payslip::read","hrms:payroll:payslip::write"],"scope":{"cert":"C17"}}`
	if err := os.WriteFile(g2v2, []byte(g2v2JSON), 0600); err != nil {
		t.Fatal(err)
	}
	published, publicationWarning := run(0, "grant", "publish", "--file", g2v2, "--support-assignment", "fm5b7t4p5iv8", "--tenant", "acme", "--app", "hrms", "--db", database, "--fixture-context", "maya-grant-publisher")
	if published != g2v2JSON+"\n" || !strings.Contains(publicationWarning, "transient publication evidence") {
		t.Fatalf("publication stdout=%q stderr=%q", published, publicationWarning)
	}
	reopenedGrant, _ := run(0, "inspect", "grant", "fk3x9r2man0d", "--db", database, "--tenant", "acme", "--app", "hrms")
	reopenedAssignment, _ := run(0, "inspect", "assignment", "fm5b7t4pan0d", "--db", database, "--tenant", "acme", "--app", "hrms")
	if reopenedGrant != published || reopenedAssignment != after {
		t.Fatalf("reopen grant=%q assignment before=%q after=%q", reopenedGrant, after, reopenedAssignment)
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
	run(3, "inspect", "assignment", "fm5b7t4pan0d", "--db", database, "--tenant", "acme", "--app", "hrms")
	run(3, "inspect", "grant", "fk3x9r2m5iv8", "--db", database, "--tenant", "acme", "--app", "fi7io4lvkfsw")
	run(3, "inspect", "grant", "fk3x9r2m5iv8", "--db", database, "--tenant", "fi7io4lvkfsw", "--app", "hrms")
	run(4, "scenario", "seed", "team-fin-c17", "--db", database, "--tenant", "acme", "--app", "hrms")
	duplicate := filepath.Join(dir, "duplicate.json")
	if err := os.WriteFile(duplicate, []byte(`{"version":"1","id":"fm5b7t4pan0d","id":"fi7io4lvkfsw","grant_id":"fk3x9r2man0d","grant_revision":1,"recipient":{"type":"group","id":"fibggi2juxhc"},"status":"enabled"}`), 0600); err != nil {
		t.Fatal(err)
	}
	run(2, "check", "assignment", "--file", duplicate, "--db", database, "--tenant", "acme", "--app", "hrms")
	run(3, "assign", "--file", filepath.Join(root, "testdata", "a2.json"), "--fixture-context", "unknown", "--db", database, "--tenant", "acme", "--app", "hrms")
	run(3, "inspect", "assignment", "fm5b7t4pan0d", "--db", database, "--tenant", "acme", "--app", "hrms")
	revision := filepath.Join(dir, "g2-v2.json")
	if err := os.WriteFile(revision, []byte(`{"version":"1","grant_id":"fk3x9r2man0d","revision":2,"parent_grant_id":"fk3x9r2m5iv8","permissions":["hrms:payroll:payslip::read"],"scope":{"cert":"C17"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	missingDB := filepath.Join(dir, "missing.db")
	run(4, "grant", "publish", "--file", revision, "--support-assignment", "fm5b7t4p5iv8", "--fixture-context", "maya-grant-publisher", "--db", missingDB, "--tenant", "acme", "--app", "hrms")
	if _, err := os.Stat(missingDB); !os.IsNotExist(err) {
		t.Fatalf("missing publication database was created: %v", err)
	}
}
