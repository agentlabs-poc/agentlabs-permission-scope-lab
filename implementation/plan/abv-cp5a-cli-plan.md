# CP5-A Catalog CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. User requests no reviewers; keep test verification.

**Goal:** Demonstrate protected catalog registration through the in-process CLI.

**Architecture:** Add a separate application-only connector for catalog commands,
using the existing facade registration methods and an explicitly bounded lab
publisher adapter. Keep the tenant connector and its callers backward compatible.

**Tech Stack:** Go 1.25-compatible, existing SQLite and CLI; no new dependency.

**Spec:** [CP5-A](abv-cp5a-design.md).

## Global Constraints

- Requires verified provider, protected registration and computed root units.
- No fake tenant and no tenant-admin fallback for catalog publication.
- Fixed lab context is not authenticated identity. Print the lab-only warning.
- Only existing, marked lab databases; never create a missing database on command.
- No new canonical JSON/YAML; labeled output for internal definitions.
- No Auth integration, roles, PostgreSQL, deletion, reparenting or retirement here.
- No review dispatches. One 15-minute attempt, one 8-minute correction maximum,
  then report/reassess. Full-slice verification is controller-owned.

### Task 1: Catalog connector, commands and working example

**Files:** create `implementation/abv/application/catalog.go`,
`implementation/abv/cli/catalog.go`, `catalog_test.go`,
`implementation/abv/internal/lab/catalog.go`, `catalog_test.go`.
Modify `cli/run.go`, `internal/lab/application.go` only for shared marker loading,
`cmd/abv/main.go`, `cmd/abv/main_test.go`, and `docs/local-testing.md`.
Read existing parser, `nilCapability`, `report`, lab connector/marker and command
tests before editing. No service/provider/root changes unless a concrete failure
requires controller coordination.

**Interfaces produced:**

```go
// application/catalog.go
type CatalogAPI interface {
    RegisterPermission(context.Context, domain.Application, domain.FixtureContext,
        domain.PermissionDefinition, []string) (domain.PermissionDefinition, error)
    RegisterScope(context.Context, domain.Application, domain.FixtureContext,
        domain.ScopeDefinition) (domain.ScopeDefinition, error)
}
type CatalogConnect func(context.Context, domain.Application, string) (CatalogAPI, func() error, error)
// lab/catalog.go
func ConnectCatalog(context.Context, domain.Application, string) (application.CatalogAPI, func() error, error)
```

Preserve existing `cli.Run` call sites with a trailing variadic
`catalogConnect ...application.CatalogConnect`. Accept zero (capability unavailable)
or one connector; refuse more than one. Main passes `lab.ConnectCatalog` as the
new last argument. Catalog commands dispatch before constructing `domain.Area`.
All prior commands keep their mandatory tenant/app context and flag validation.

**Exact command surface (prototype CLI, not a canonical schema):**

```sh
abv catalog register-scope region --app hrms --db lab.db --fixture-context application-publisher
abv catalog register-scope owner --allowed-tokens '$self' --app hrms --db lab.db --fixture-context application-publisher
abv catalog register-permission hrms:payroll:payslip::export --supported-keys dept,region --app hrms --db lab.db --fixture-context application-publisher
```

Each command requires app/db/fixture-context, rejects tenant/file/case flags,
unknown/duplicate flags and excess positional arguments. Lists use comma-separated
items; if the optional flag is present reject empty entries, never silently drop
one or trim an invalid input into success. Pure validators handle duplicate or
unsupported values. Permission is always registered Active=true; no active toggle.

**Lab authority:** fixed identity version 1, actor user `application-publisher`,
human_id `application-publisher`. Its adapter captures one validated Application
and checks exact matching application, catalog application and identity on each
operation. It rejects assignment administration; no membership inference. The
fixture name is exactly `application-publisher`. Existing `maya-team1` cannot
publish. Verify existing lab marker/version/scenario and requested application
before protected operation; do not require a fabricated tenant. Extract shared
marker decoding from existing verifyMarker so both connectors preserve its strict
checks. Existing tenant verification still compares the actual tenant and app.

- [ ] **1. RED:** add real CLI/lab end-to-end tests that seed the current scenario,
  register scope then permission, close/reopen and inspect through existing tenant
  `inspect permission`/`inspect scope`. Assert literal persisted values, unchanged
  G0/G1/G2 content and assignments, success output and lab warning.
  Cover missing connector, nil/typed-nil API, unknown app, missing file (must stay
  absent), unmarked DB, wrong fixture `maya-team1`, rejected tenant flag, malformed
  lists, duplicate definition, unsupported key/token, canceled operation, output
  writer error and close error. On failed mutation stdout must not contain success.
  Test tenant commands still reject missing tenant. For full root behavior, the
  root unit's Go tests remain the acceptance proof; don't invent another scenario.
- [ ] **2. Run RED:** `go test ./cli ./internal/lab ./cmd/abv -run Catalog -count=1`.
- [ ] **3. Implement:** reuse parser/helper patterns; one dedicated catalog
  dispatch handler, optional connector and lab adapter. Registration calls the
  protected facade, never internal SQL insert. Use existing error codes and
  buffered labeled output such as `internal projection: permission` plus ID and
  active flag (scope output includes key and tokens). No JSON serialization.

```go
// Main's only new wiring:
// cli.Run(ctx, args, in, out, diag, lab.Connect, lab.Scenarios{}, lab.ConnectCatalog)
// Catalog branch validates Application before connecting; no NewArea call.
// Successful definition is rendered only after facade registration returns nil.
```

- [ ] **4. GREEN:** focused command, `go test ./...`,
  `go test -race ./cli ./internal/lab ./cmd/abv`, `go vet ./...`, `go build ./...`,
  `git diff --check`. Document the three working commands, lab-only authority,
  add-only limits and ordinary child non-expansion in local-testing.md.
- [ ] **5. Commit explicit files:** `feat(abv): expose protected catalog CLI`.
  No push or reviewer. Report RED/GREEN, SHA, tests and remaining limits.

## Exit

This is additive CP5-A, not full CP5. Role management, complete retirement and
definition lifecycle coverage, and CP4 publication/adoption remain pending.
