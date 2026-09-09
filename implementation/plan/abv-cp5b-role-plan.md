# CP5-B Immutable Role Publication Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. User override: no review passes. Keep focused tests and final verification.

**Goal:** Publish immutable permission-bundle revisions through ABV, SQLite and CLI without changing any grant adoption.

**Architecture:** Reuse tenant/application `Area`, `RoleContent`, the roles table,
existing Update transaction and protected mutation/facade pattern. Publishing a
role is not assigning it: administrative approval plus registered-definition
validation protects publication; grant adoption remains separately authorized
and boundary-validated by its eventual operation.

**Tech Stack:** Go 1.25-compatible; existing SQLite and CLI, no dependencies or migration.

**Spec:** [Q-089-B](../../docs/role-revisions.md) and
[Q-118](../../docs/role-grant-contract.md). The following execution choices are
internal prototype details, not new canonical schemas or permission spellings.

## Global Constraints

- User requests working implementation first, without independent review passes.
- Roles keep the existing tenant/application storage boundary; not application-only catalog administration.
- Existing published role revision is immutable; duplicate key conflicts, never overwrite.
- Publication cannot modify grant contents, adoption, assignments, controls or memberships.
- Permissions are explicit, nonempty, unique, supported grammar and active registrations in the same application.
- Role has no recipient, scope, parent grant or independent access authority.
- No grant adoption/publication, role deletion, retirement, mode changes, PostgreSQL or Auth integration in this slice.
- No new canonical JSON/YAML: use existing internal RoleContent and labeled CLI output.
- Reuse existing helpers; no review agents; Sol-medium coding, apply_patch edits, scoped commits, no worker push.
- Whole slice reassessment after 40 minutes. Each task gets one 15-minute attempt
  and at most one 8-minute correction; at cap report concrete gap, narrow/replan.

## Publication semantics and rationale

The caller supplies a positive int64 revision and role ID. SQLite's composite
key preserves immutability and isolates identical IDs in different Areas. This
prototype does not allocate numbers or define a canonical consecutive-number
scheme: unused positive revisions can be supplied, and lookup still resolves an
explicit adopted revision. No implicit latest selection or auto-adoption is added.
No role grant is fabricated to run validation. The administrative adapter receives
the entire proposal and must authorize its role ID and permission bundle; possession
of business permissions is not automatically publishing authority.

```text
CLI: tenant + app + lab role-publisher context + ID/revision/permissions
  -> facade PublishRole
  -> existing SQLite Update transaction
     -> role-publication administrative check (isolated snapshot/proposal)
     -> registered role-definition validation
     -> insert one immutable role revision
  -> committed result

Existing grant: role_id R, role_revision 1 -> remains revision 1
New role revision R/2 -> available definition, not adopted authority
```

### Task 1: Protected immutable role publication

**Files:** modify `implementation/abv/internal/storage/provider.go` and
`internal/storage/sqlite/update.go` for a single optional `NewRoleRevision` write
category. Create `internal/storage/sqlite/role.go`, `role_test.go`,
`internal/validation/role.go`, `role_test.go`, `internal/mutation/role.go`,
`role_test.go`, and public `role.go`, `role_test.go`, all under `implementation/abv`.
Read existing catalog publication, transaction cleanup, role snapshot loading,
codec.PermissionList, cloneSnapshot and validateSupportedIdentity first.

**Interfaces:**

```go
// Added field on existing storage.WriteSet:
NewRoleRevision *domain.RoleContent
// New pure validator:
func CheckRolePublication(domain.Area, domain.Catalog, domain.RoleContent) error
// Optional capability in mutation and mirrored in public abv:
type RoleAdministration interface {
    CheckRolePublication(context.Context, storage.Snapshot, domain.Identity,
        domain.RoleContent, time.Time) error
}
// Public signature uses abv.Evidence alias in place of storage.Snapshot.
// Methods on *mutation.Service and *abv.Facade:
PublishRole(context.Context, domain.Area, domain.Identity, domain.RoleContent) (domain.RoleContent, error)
```

- [x] **1. RED tests:** use existing real SQLite fixtures; insert R/1, publish
  R/2 and verify both literal permission lists after close/reopen. Seed a grant
  referring to R/1 and prove its selected permissions and stored content/assignment
  stay unchanged after R/2 publication. Example desired assertion:

```go
got, err := svc.PublishRole(t.Context(), area, identity,
    domain.RoleContent{ID: "payslip-reader", Revision: 2,
        Permissions: []string{lab.PayslipRead, lab.PayslipWrite}})
if err != nil { t.Fatal(err) }
if got.ID != "payslip-reader" || got.Revision != 2 { t.Fatalf("bad result: %#v", got) }
// Fresh provider read must show literal rev1 read; rev2 read/write;
// a rev1-referencing grant still selects read only, not rev2's write.
```

  Cover nil/empty/duplicate/unknown/inactive permissions; blank/wildcard/invalid
  UTF8 role ID and nonpositive revision; invalid Area/identity; missing role admin;
  rejected publisher; wrong-area evidence; mixed writes; callback error/cancellation;
  duplicate revision including concurrent insertion; different tenant/app same key;
  hostile callback/catalog mutation; administration mutates permission slice;
  zero output and no DB change on failure. Check shared snapshot limits continue
  to reject rather than truncate. Existing role data must never be overwritten.
- [x] **2. Run RED:** `go test ./internal/validation ./internal/storage/sqlite ./internal/mutation . -run 'RolePublication|PublishRole' -count=1`.
- [x] **3. Implement:** validate Area, identity and role shape; optional admin
  must exist (including typed-nil protection). In Update callback require exact
  Area/catalog match; call admin with cloned snapshot and separately cloned role
  permissions; validate original proposal against unmodified catalog; reject
  duplicate RoleKey and check cancellation. Return one NewRoleRevision write.
  Provider includes role in mutually-exclusive category checks, validates against
  authoritative catalog loaded inside the same transaction (not callback-mutated
  maps), encodes permissions and performs INSERT only. Composite SQL key gives
  duplicate conflict. Reuse runTransaction; no new transaction wrapper.

```go
// Return only from successfully committed Update:
return proposed, nil
// Any error returns domain.RoleContent{}, err.
// Never update grants when inserting a role revision.
```

- [x] **4. GREEN:** focused command; `go test ./...`;
  `go test -race ./internal/storage/sqlite ./internal/mutation ./internal/validation .`;
  `go vet ./...`; `git diff --check`.
- [x] **5. Commit exact files:** `feat(abv): publish immutable role revisions`.
  Report RED/GREEN, SHA and API details. No reviewer or push. Controller dispatches Task2 only after this passes.

### Task 2: Bounded lab role publisher and CLI

**Files:** create `implementation/abv/application/role.go`, `cli/role.go`,
`cli/role_test.go`, `internal/lab/role.go`, `internal/lab/role_test.go`.
Modify `cli/run.go`, `cli/scenario.go` (existing dispatch), `internal/lab/application.go`,
`cmd/abv/main_test.go` and `docs/local-testing.md`. Read current dispatch location
before coding; if it differs, change only actual dispatch owner and report it.
Optional operation checks reuse nilCapability and report/output helpers.

**Interfaces consumed:** Task1 public facade PublishRole and RoleAdministration.
**Produced:**

```go
// application/role.go
type RoleAPI interface {
    PublishRole(context.Context, domain.Area, domain.FixtureContext,
        domain.RoleContent) (domain.RoleContent, error)
}
```

**Exact CLI:**

```sh
abv role publish payslip-reader --revision 2 --permissions hrms:payroll:payslip::read,hrms:payroll:payslip::write --tenant acme --app hrms --db lab.db --fixture-context maya-role-publisher
abv inspect role payslip-reader --tenant acme --app hrms --db lab.db
```

Require positive base10 int64 revision and nonempty comma list, no silent dropping
of entries. Reject duplicate/unknown flags, missing tenant/app/db/context and excess
positionals. No file or JSON input. Labeled internal role projection output with
ID, revision and permissions, emitted only after successful publication.

**Lab premise:** add a wrapper embedding existing AssignmentStatusAdministration
so other operations keep their behavior. Its explicit role-publication premise
authorizes only Maya's version1 direct-user identity, exact connected Area, role
`payslip-reader`, and permission selection within read/write. Require current
`AssignmentAdmins` membership as the lab fixture's existing revocable admin link.
This wrapper explicitly adds role administration; membership alone does not imply
publishing authority in the model. `maya-role-publisher` selects this lab operation;
`maya-team1` and `application-publisher` must be rejected here. Reuse strict marker
verification. No role scope or parent source is invented; publication does not
assign business access. No production permission name or new canonical admin grant.

- [x] **1. RED:** real seeded-DB CLI publication of revision2, reopen inspect both
  revisions, preserve G0/G1/G2 and assignments. Include wrong fixture/role, permission
  outside lab publish ceiling (registered delete), missing membership, unregistered
  permission, malformed revision/list, duplicate revision, nil/typed-nil capability,
  missing/unmarked DB, cancellation, output and close failures. Failure has no
  success output or partial role record. Existing commands must still pass.
- [x] **2. Run RED:** `go test ./cli ./internal/lab ./cmd/abv -run 'Role|Publish' -count=1`.
- [x] **3. Implement:** parser/dispatch -> optional RoleAPI -> lab marker/context
  -> facade PublishRole -> administration + validation + insertion. No direct SQL
  mutation in CLI/lab. Parse revision via strconv.ParseInt; use existing list
  parsing helper if applicable. Preserve existing connectors and signatures.

```go
// No new persistence path in the adapter:
return a.facade.PublishRole(ctx, area, TeamFINC17(area).Issuer, proposed)
```

- [x] **4. GREEN:** focused command, `go test ./...`,
  `go test -race ./cli ./internal/lab ./cmd/abv`, `go vet ./...`, `go build ./...`,
  `git diff --check`. Run compiled CLI example above on a disposable DB and inspect
  after reopening. Update local-testing with exact commands and lab-only caveat.
- [x] **5. Commit explicit files:** `feat(abv): expose role publication CLI`.
  Report commands, output, SHA and limitations; no review or push.

## Final controller exit

Run full Go/full race/vet/build/module verification and existing site checks.
Record completion with rationale, preserve all prior decisions, commit/push scoped
checkpoint. Leave grant publication/adoption and remaining definition lifecycle
pending. No whole-handbook or whole-ABV completion claim.
