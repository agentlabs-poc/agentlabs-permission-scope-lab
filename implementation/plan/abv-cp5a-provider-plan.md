# CP5-A Catalog Provider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. User override: no reviewer dispatches; get working first, retain test-first verification.

**Goal:** Add the application-only persistence seam needed by protected registration.

**Architecture:** Preserve tenant `Provider` unchanged. Add an optional SQL-free
catalog capability, implemented by the same SQLite provider and transaction
helper. The protected coordinator and CLI are subsequent units, not part of this
internal persistence deliverable.

**Tech Stack:** Go 1.25-compatible, existing modernc SQLite; no new dependencies.

**Spec:** [CP5-A design](abv-cp5a-design.md), approved by “keep moving”.

## Global Constraints

- Tenant operations retain mandatory tenant/application `Area`.
- Catalog operations require explicit application-only context; no fake tenant.
- Existing application only; add-only; no overwrite or reactivation.
- No production Auth integration, PostgreSQL, deletion, reparenting or new wire schema.
- This internal provider is not an administrative authorization gate.
- No review dispatches per latest user instruction. Tests remain mandatory.
- One coding attempt: 15 minutes; one focused correction: 8 minutes. At the cap,
  stop and report the concrete gap so the controller can split or replan.

## Delivery ordering and regrouping rationale

The original catalog-write unit combines persistence and protected coordination.
Split it into two because each has a distinct testable contract: transactional
storage does not prove authorization. This plan covers persistence only. Next:
protected registration coordinator/facade, computed roots, CLI acceptance. No
unit may be reported as whole CP5-A completion. Persisted provider tests supply
the fixtures for later protected-operation tests.

### Task 1: Application-only catalog persistence

**Files:**
- Create `implementation/abv/domain/application.go` and `application_test.go`.
- Create `implementation/abv/internal/storage/catalog.go`.
- Create `implementation/abv/internal/validation/catalog.go` and `catalog_test.go`.
- Create `implementation/abv/internal/storage/sqlite/catalog.go` and `catalog_test.go`.
- Modify `implementation/abv/internal/storage/sqlite/snapshot.go` only to reuse its
  existing catalog loader with an explicit application ID and shared record count.
- Read `internal/storage/sqlite/update.go`, `fixture.go`, `open.go`, and
  `internal/codec/content.go` before coding. Do not duplicate transaction cleanup.

**Interfaces produced (internal projections, not canonical wire records):**

```go
// domain/application.go
type Application struct { id string }
func NewApplication(id string) (Application, error)
func (a Application) ID() string
func (a Application) Validate() error

// internal/storage/catalog.go
type CatalogWriteSet struct {
    Permission *domain.PermissionDefinition
    SupportedKeys []string
    Scope *domain.ScopeDefinition
}
type CatalogProvider interface {
    ReadCatalog(context.Context, domain.Application, func(domain.Catalog) error) error
    UpdateCatalog(context.Context, domain.Application, func(domain.Catalog) (CatalogWriteSet, error)) error
}

// internal/validation/catalog.go
func CheckPermissionRegistration(domain.Catalog, domain.PermissionDefinition, []string) error
func CheckScopeRegistration(domain.Catalog, domain.ScopeDefinition) error
```

**Input/error rules:** Application rejects blank, wildcard-containing or invalid
UTF-8 strings without normalization. Permission uses existing `codec.PermissionList`
syntax; Active must be true. Scope key rejects blank, invalid UTF-8 and wildcard
strings, using current content scope-key restrictions. AllowedTokens uses existing
`selectedTokens`; SupportedKeys uses `selectedKeys` regardless of compatibility
mode. Existing identifier returns ErrConflict even if inactive. Invalid shape
returns ErrMalformed; unsupported token/key/active state returns ErrRejected.
Missing application returns ErrNotFound. Duplicate SQL keys return ErrConflict.
Keep existing storage cancellation/unavailability classifications.

- [ ] **1. Write failing tests before implementation.** Start with a real
  temporary fixture, create via `CreateFixture`, and assert the returned provider
  implements `storage.CatalogProvider`. Before the new interface exists, use a
  local interface with the planned signatures after introducing only type stubs;
  the intended RED is missing capability/behavior, not an unrelated build error.
  Use existing `contracttest` snapshots or a minimal application catalog fixture.
  The primary behavioral assertion after insertion is:

```go
err = catalogs.ReadCatalog(t.Context(), app, func(c domain.Catalog) error {
    got, ok := c.Permissions["hrms:payroll:payslip::export"]
    if !ok || !got.Active { t.Fatal("registered permission missing") }
    if c.ApplicationID != "hrms" { t.Fatal("application boundary changed") }
    return nil
})
if err != nil { t.Fatal(err) }
```

  Cover valid scope/permission with support keys, close/reopen persistence, two
  tenants sharing the catalog, unrelated application isolation, duplicate/inactive
  duplicate refusal, bad token/key/ID, nil callback, invalid context, unknown app,
  callback error, canceled context before and during callback, and snapshot limit.
  Assert database unchanged on failure using a fresh read, not callback evidence.
  Cover two concurrent duplicate inserts with at most one success; other result
  must be ErrConflict. Never busy-loop/retry in the provider.

- [ ] **2. Run RED.** From `implementation/abv`, run
  `go test ./domain ./internal/validation ./internal/storage/sqlite -run 'Application|Catalog' -count=1`.
  Record the expected failure and cause in the worker report.

- [ ] **3. Implement the minimal provider.** Reuse `p.connection` and
  `p.transaction`: BEGIN for reads, BEGIN IMMEDIATE for updates. Validate context
  and callback before DB work, verify existing application, load bounded catalog,
  invoke callback once, check cancellation, validate one write category and insert.
  Reject mixed/empty writes and SupportedKeys supplied without Permission.
  Permission and support rows commit together. No catalog-mode update.
  Validate against persisted catalog evidence, not a callback-mutated snapshot:
  reload inside the same write transaction or preserve a deep copy before callback.
  Refactor `snapshotReader.catalog` to accept explicit application ID while retaining
  its row counting and strict decode checks for tenant snapshots. Do not create
  an Area with an invented tenant. Use existing SQL tables and constraints.

```go
// Core transaction ordering; p.loadCatalog represents the refactored loader.
// Keep loading private; its exact helper name is implementer-owned.
// Validate request against authoritative evidence AFTER callback returns;
// never trust changes made to its maps as registered dependencies.
// INSERT permission first, then its supported keys; transaction rolls both back.
```

- [ ] **4. Verify GREEN and preservation.** Run the focused command above,
  `go test ./...`, and
  `go test -race ./internal/storage/sqlite ./internal/validation ./domain`.
  Assert existing tenant snapshots retain their prior records and compatibility
  mode. No source change outside the listed ownership except a narrowly necessary
  shared-helper change reported to the controller.
- [ ] **5. Commit scoped code/tests.** Run `git diff --check`; stage explicit
  task files; commit `feat(abv): add application catalog persistence`.
  Do not commit local reports or push. Report SHA, RED/GREEN evidence, changed
  files, remaining limitations and any cap reached. Controller owns publication.
