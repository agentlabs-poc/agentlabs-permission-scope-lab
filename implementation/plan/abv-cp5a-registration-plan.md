# CP5-A Protected Registration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. User override: no review dispatches; retain test-first verification.

**Goal:** Expose protected application-only permission and scope registration.

**Architecture:** Add registration methods to the existing mutation service and
facade. Detect optional catalog provider and registration-administration
capabilities, refusing missing capabilities. Authorization and validation run
inside the provider's catalog write transaction. Existing tenant methods do not
change their context or acquire publisher powers.

**Tech Stack:** Go 1.25-compatible, existing SQLite provider, no new dependencies.

**Spec:** [CP5-A](abv-cp5a-design.md).

## Global Constraints

- Prerequisite: [catalog provider task](abv-cp5a-provider-plan.md) passes tests.
- Application-only context for catalog operations; mandatory Area remains for tenant operations.
- No new canonical registration JSON, no fake tenant, no business interpretation.
- Catalog publishing authority is separate from tenant/team administration.
- Existing application only; add-only; no retirement, mode toggle or overwrite.
- No production Auth integration, PostgreSQL, deletion, reparenting or roles in this unit.
- User requests no review dispatches. Keep test-first verification and explicit failure evidence.
- One 15-minute coding attempt, at most one 8-minute focused correction; report and reassess at cap.

### Task 1: Protected service and facade

**Files:**
- Create `implementation/abv/internal/mutation/catalog.go` and `catalog_test.go`.
- Modify `implementation/abv/internal/mutation/assignment.go` only to extract its
  catalog deep-copy logic into `cloneCatalog`, reused by `cloneSnapshot`.
- Create `implementation/abv/catalog.go` and `catalog_test.go` for public forwarding
  methods and administration capability, keeping `abv.go` unchanged if possible.
- Read `mutation/service.go`, `assignment.go`, `grant_status.go`, `abv.go`,
  `storage/catalog.go`, `validation/catalog.go` and provider task report first.

**Interfaces consumed:** `domain.Application`; `storage.CatalogProvider` with
`ReadCatalog` and `UpdateCatalog`; `storage.CatalogWriteSet`; pure
`validation.CheckPermissionRegistration` and `CheckScopeRegistration`.

**Interfaces produced in mutation and mirrored in public abv:**

```go
type CatalogAdministration interface {
    CheckPermissionRegistration(context.Context, domain.Application, domain.Catalog,
        domain.Identity, domain.PermissionDefinition, []string, time.Time) error
    CheckScopeRegistration(context.Context, domain.Application, domain.Catalog,
        domain.Identity, domain.ScopeDefinition, time.Time) error
}
// Methods on *mutation.Service and *abv.Facade, with identical signatures:
RegisterPermission(context.Context, domain.Application, domain.Identity,
    domain.PermissionDefinition, []string) (domain.PermissionDefinition, error)
RegisterScope(context.Context, domain.Application, domain.Identity,
    domain.ScopeDefinition) (domain.ScopeDefinition, error)
```

These are Go interfaces over existing internal definition projections, not JSON
contracts. Existing constructor uses the provided administration adapter; the
adapter optionally implements CatalogAdministration. Absence must return
ErrUnsupported, not a fallback to CheckAssignment. A publisher adapter may reject
all assignment administration; this does not require tenant context for publishing.

- [x] **1. Write RED tests.** Use real SQLite fixtures with a focused test
  administration adapter; it checks exact application, identity and proposed
  definition. Example postcondition after a successful service call:

```go
got, err := service.RegisterPermission(t.Context(), app, publisher,
    domain.PermissionDefinition{ID: "hrms:payroll:payslip::export", Active: true}, []string{"dept"})
if err != nil { t.Fatal(err) }
if got.ID != "hrms:payroll:payslip::export" || !got.Active {
    t.Fatalf("unexpected result: %#v", got)
}
// Close/reopen and read through CatalogProvider; assert literal persisted
// permission and supported key, using the provider task's test conventions.
```

  Named cases: approved scope/permission persists; no catalog administration;
  no catalog provider; invalid application; malformed/unsupported identity;
  wrong-application evidence from a fake provider; wrong publisher; invalid
  registered key; duplicate definition; cancellation inside admin callback;
  provider write error; and admin mutates catalog maps/proposal slices to try
  forging approval evidence. Errors must return zero result and preserve DB.
  Use hostile adapter tests to prove mutations cannot change saved proposal or
  bypass validator. The adapter receives isolated copies of scope tokens and
  supported keys, not aliases used by persistence. Caller input remains unchanged.
  Facade test exercises real forwarding plus denied authority, not method presence.

- [x] **2. Run RED:** `go test ./internal/mutation . -run 'Register|Catalog' -count=1`.
  Record expected missing behavior, not accidental fixture/build breakage.
- [x] **3. Implement:** precheck cancellation, application and existing supported
  identity validation. Detect optional capabilities with existing `nilInterface`.
  Snapshot application must equal requested application. Clone caller slice inputs
  before entering the operation; isolate admin evidence and proposal copies.
  Call operation-specific administration once inside UpdateCatalog, then pure
  registration validator against untouched snapshot, then check cancellation and
  return one CatalogWriteSet. Only after successful commit return the definition.
  Do not derive authority from tenant grants or return success before commit.

```go
// Required operation ordering, using existing helpers where available:
// app.Validate -> validateSupportedIdentity -> optional capability checks
// -> UpdateCatalog callback -> app match -> admin(isolated copies)
// -> validation.Check... -> ctx.Err -> one CatalogWriteSet
// -> committed definition; on any error, zero definition.
```

- [x] **4. Run GREEN:** focused command, `go test ./...`,
  `go test -race ./internal/mutation .`, `go vet ./...`, `git diff --check`.
- [x] **5. Commit only listed source/tests**, message
  `feat(abv): protect application catalog registration`. No push and no review
  agents. Report RED/GREEN, SHA, exact files, concerns and remaining root/CLI work.

## Exit and next unit

This proves protected registration through Go calls, not CLI availability or
computed-root behavior. Those remain separate units under CP5-A. Roles and CP4
publication/adoption remain pending. Do not add them to satisfy this task.
