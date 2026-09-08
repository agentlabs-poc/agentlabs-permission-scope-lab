# CP4-A Grant Enable/Disable Implementation Plan

> **For agentic workers:** Use `superpowers:subagent-driven-development` with
> Sol-medium coding agents. Complete one bounded task at a time; independent
> review precedes publication. This plan does not deliver the rest of CP4.

**Goal:** Add protected grant-wide enable/disable to the reusable ABV and local CLI.

**Architecture:** Reuse the facade, administration port, ABV lineage checks and
provider transaction. Add one typed control update to the write set. Preserve
existing assignment APIs through optional, explicit grant-status capabilities.

**Tech Stack:** Existing Go 1.25-compatible module, standard library,
`database/sql`, pinned `modernc.org/sqlite` v1.58.0. No new dependency or migration.

**Spec:** [CP4-A design](abv-cp4a-design.md), especially the reconciled CP4-P03
rule; [canonical binding lifecycle](../../docs/parent-grant-bindings.md),
[grant revisions](../../docs/grant-revisions.md), and
[ABV operation coverage](../../docs/authority-boundary-validation.md#6-operation-coverage).

## Global constraints

- Tenant/application is mandatory explicit outer context on every operation.
- Grant enablement changes one global control; assignments remain separate.
- Inventory disabled assignments, but their missing support alone does not
  block grant enablement. They remain disabled and cannot supply authority.
- Every enabled assignment of the changed grant must pass its required current
  support checks. One failing required route rejects the whole enable operation.
- Preserve actual adopted grant/role revisions; enable is not latest-only creation.
- Preserve scope AND, permission subsets and actual parent-team holdings.
- Separate administrative permission from ABV transition/boundary validation.
- No implicit personal issuer dependency for existing team-held authority.
- No assignment, content, membership, parent, descendant-status or trust rewrite.
- No production Auth integration, public wire schema, root recovery or business DB.
- Unsupported recipient-relative/proxy/root-management operations fail explicitly;
  unsupported implementation coverage is not a new canonical prohibition.
- Use existing snapshot/traversal limits and transaction ordering; never truncate.

## Work limits and exit

Planning is complete when interfaces, outcomes, task boundaries and checks below
are coherent. Coding units have maximum budgets of 12, 15 and 12 minutes for
Tasks 1–3, with progress reports halfway through. Each review gets five minutes.
One focused fix unit per task, at most eight minutes, then reassess unresolved
findings rather than repeat. Reassess the whole slice at 60 minutes regardless
of task state. These are controller-chosen work limits, not service latency SLAs.
An overrun produces a tested partial result and exact remaining scope, never a
false completion claim or weakened gate. No extra feature enters a fix unit.

Exit: CLI disable/enable survives restart, preserves dependent states/adoptions,
passes the shared-binding positive/counterexample pair, two-gate/transaction
tests and independent review. Full CP4 remains incomplete afterward.

## Exact internal contracts

The operation consumes and returns the existing versioned `domain.GrantControl`:

```json
{"version":"1","id":"G2","status":"disabled"}
```

No separate revision, recipient, scope or optimistic-token field is added. Area
and identity are separate arguments. Only an existing control can be changed;
no upsert or trust establishment. Return a zero control on failure and the exact
committed control after provider success. Same-state requests still run the
applicable checks; an already-enabled flag is not permission to skip validation.

```go
// internal/storage/provider.go — internal persistence instruction, not JSON.
type GrantStatusChange struct {
    Before domain.GrantControl
    After  domain.GrantControl
}
type WriteSet struct {
    NewAssignments    []domain.Assignment // existing behavior unchanged
    GrantStatusChange *GrantStatusChange  // nil or one existing-control change
}

// internal/mutation/service.go — additive capability; old adapters still compile.
type GrantStatusAdministration interface {
    CheckGrantStatus(context.Context, storage.Snapshot, domain.Identity,
        domain.GrantControl, time.Time) error
}

// abv.go — public optional administrative port, Evidence is the existing alias.
type GrantStatusAdministration interface {
    CheckGrantStatus(context.Context, Evidence, domain.Identity,
        domain.GrantControl, time.Time) error
}

// Both mutation.Service and abv.Facade expose this method.
SetGrantStatus(context.Context, domain.Area, domain.Identity,
    domain.GrantControl) (domain.GrantControl, error)

// application/api.go — additive CLI capability, not a replacement for API.
type GrantStatusAPI interface {
    SetGrantStatus(context.Context, domain.Area, domain.FixtureContext,
        domain.GrantControl) (domain.GrantControl, error)
}
```

The coordinator type-asserts the optional administrative capability; missing or
typed-nil capability returns `ErrUnsupported`, never an implicit allow. CLI
does the equivalent for `GrantStatusAPI`. Existing constructors and three-method
application API remain source-compatible. No generic operation registry.

Prototype coverage: enablement of existing team-held grants is supported. An
enabled direct-human/proxy/recipient-relative route cannot be skipped because
the resolver cannot prove it: return unsupported without changing global status.
Trusted-root control operations remain outside this slice. An explicitly
disabled assignment does not require unsupported eligibility resolution merely
to remain disabled. If there are no enabled assignments, status alone establishes
no authority; no route is invented and subsequent assignment activation remains
guarded. Invalid stored shapes, duplicate bindings and incomplete snapshots are
still errors, not exclusions from inspection.

## Task 1 — conditional provider control update

**Files:** modify `internal/storage/provider.go`,
`internal/storage/sqlite/update.go`; create
`internal/storage/contracttest/grant_status.go`; invoke its
`RunGrantStatus(t *testing.T, factory Factory)` from existing `contracttest.Run`.
Paths in tasks are relative to `implementation/abv/`.

**Consumes:** current provider `Update`, snapshot and core `GrantControl`.
**Produces:** the write-set types above and atomic conditional control persistence.

- [ ] Add a failing conformance case using existing `fixtures(t)` and Factory:

```go
func RunGrantStatus(t *testing.T, factory Factory) {
    t.Run("control status persists without changing authority content", func(t *testing.T) {
        seeded := fixtures(t)[0]
        path := t.TempDir() + "/authority.db"
        provider, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
        if err != nil { t.Fatal(err) }
        err = provider.Update(t.Context(), seeded.Area, func(s storage.Snapshot) (storage.WriteSet, error) {
            before := s.Controls["G1"]
            after := before
            after.Status = "disabled"
            return storage.WriteSet{GrantStatusChange: &storage.GrantStatusChange{Before: before, After: after}}, nil
        })
        if err != nil { t.Fatal(err) }
        if err = provider.Close(); err != nil { t.Fatal(err) }
        provider, err = factory.Open(t.Context(), path)
        if err != nil { t.Fatal(err) }
        defer provider.Close()
        control := seeded.Controls["G1"]
        control.Status = "disabled"
        seeded.Controls["G1"] = control
        assertSnapshot(t, provider, seeded)
    })
}
```

- [ ] Run `go test ./internal/storage/... -count=1`; record the missing-type RED.
- [ ] Add the types. Reject mixed assignment/control write sets before effects.
  Validate before/after version `"1"`, identical nonblank/nonwildcard valid UTF-8
  IDs, and only `enabled`/`disabled` statuses. Verify the current stored control
  equals `Before`; missing control is not found, mismatch is conflict. Serialize
  the exact `After`, then perform the parameterized update inside existing Update:

```sql
UPDATE grant_controls SET status=?, canonical_json=?
WHERE tenant_id=? AND application_id=? AND grant_id=? AND status=?
```

  Bind After status/JSON, trusted Area, Before ID/status; require one affected
  row. Preserve existing rollback/cancellation/error classification and canonical
  JSON/index agreement. Do not expose this primitive through the public facade.
- [ ] Add conformance cases: same ID in each isolation dimension; unchanged
  adoption/content/assignments; missing control; wrong Before; invalid version,
  status and changed ID; mixed write set; cancellation/callback error; close and
  reopen after both success and failure. Use the existing two-provider fixture
  pattern to prove lock conflict causes zero competing callbacks.
- [ ] Run focused provider tests and race checks once on final source. Review
  this task's diff, then commit the verified provider increment. It is storage
  plumbing, not yet a public authorized grant-status feature.

## Task 2 — two-gate coordinator and facade

**Files:** add `internal/mutation/grant_status.go` and `grant_status_test.go`;
modify `internal/mutation/service.go`, `abv.go`, and `abv_test.go`.
**Consumes:** Task 1 write set, existing `cloneSnapshot`, identity validation,
`lineage.ResolveParentTeam`, `validation.Narrow`, `eligibleRoute`.
**Produces:** both SetGrantStatus methods and the optional administrative ports.

- [ ] Add a failing public compatibility/authority test. Existing
  `externalAdministration` lacks the new capability and must not acquire it:

```go
func TestOldAdministrationCannotAuthorizeGrantStatus(t *testing.T) {
    area, _ := domain.NewArea("tenant-fin", "hrms")
    fixture := lab.TeamFINC17(area)
    p := &memoryProvider{snapshot: fixture.Snapshot}
    facade, err := abv.New(p, externalAdministration{area: area}, clock{now: time.Now()})
    if err != nil { t.Fatal(err) }
    proposed := domain.GrantControl{Version: "1", ID: "G2", Status: "disabled"}
    got, err := facade.SetGrantStatus(t.Context(), area, fixture.Issuer, proposed)
    if !errors.Is(err, domain.ErrUnsupported) || got != (domain.GrantControl{}) {
        t.Fatalf("old adapter acquired authority: %#v, %v", got, err)
    }
    if p.snapshot.Controls["G2"].Status != "enabled" { t.Fatal("unauthorized write") }
}
```

- [ ] Run `go test . ./internal/mutation -count=1`; record missing-method RED.
- [ ] Implement the two-gate method. Validate input/context/identity before
  Update; match returned snapshot Area/catalog; read existing control; reject
  trusted-root management as unsupported. Run CheckGrantStatus on an isolated
  snapshot, with the exact desired control and clock value. The adapter must
  establish current administrative authority for this grant and operation,
  not borrow assignment-create permission or require the historical issuer.
- [ ] For disable, ABV proves only the existing control switch is withdrawn:
  no changed IDs/content/bindings/trust. Broken/expired support must not prevent
  authorized withdrawal; it cannot amplify authority. Still require complete
  provider evidence and the separate administrative check.
- [ ] For enable, stage the one status change in a cloned snapshot. Scan all
  assignments, validate map/record consistency and uniqueness (disabled bindings
  still count), and collect the enabled assignments referencing this grant in
  deterministic ID order. Do not treat disabled assignments as absent.
- [ ] For each collected assignment: require a supported group recipient; load
  exactly `Contents[GrantKey{ID: a.GrantID, Revision: a.GrantRevision}]`; call
  `ResolveParentTeam(staged, content, a.Recipient.ID, now)` then
  `validation.Narrow(area, parent, content, staged.Roles)`. Retain every complete
  route for the final time check. Never call latest-only `exactLatestContent`
  or `CreateAssignment`; neither fits re-enablement. A disabled upstream
  assignment required by one of these routes is still a failure.
- [ ] Validate all collected routes or reject the entire change. Do not require
  unrelated or explicitly disabled descendant routes to become effective;
  their own state/support still governs resolution. No recursive state writes.
  Recheck context and every retained validity window immediately before returning
  the single Before/After write set. Return the committed control only when Update
  succeeds. The administrative gate bounds the administrator; actual parent-team
  support supplies the existing grant's source ceiling, not permanent Maya membership.
- [ ] Test the following matrix with real SQLite and operation-specific test
  adapters. Create extra scenario rows only in the test fixture, not runtime SQL:

| Test | Exact assertion |
|---|---|
| Valid disable and enable | Only G2 control changes; persistence after reopen. |
| Missing administration / malicious evidence edits | Zero result and no write, or validated result using untouched ABV evidence; never forged expansion. |
| Two enabled group assignments, one broken | Enable rejected globally, G2 remains disabled. |
| Same broken assignment explicitly disabled | Enable may pass for the valid route; broken assignment remains disabled. |
| No enabled assignments | Control may be enabled after administration/shape checks; no assignment or authority is created. |
| Disabled required upstream support | Enabled child binding cannot use it; reject. |
| Published newer content | Current assignment revision unchanged; resolve the adopted content. |
| Parent-only and child-only expiry crossing | Reject before write, zero result, unchanged control. |
| Enabled direct-human/self route | Unsupported, not skipped; no partial global enable. |
| Explicitly disabled descendant | Stays disabled on ancestor restoration. |
| Enabled descendant | Resolver shows ineffective while ancestor disabled, eligible again only if all remaining conditions hold. |
| Competing withdrawal, cancellation, snapshot/depth overflow | Correct error; no control write or success result. |
| Wrong tenant/app, typed-nil capability, malformed/no-op request | No boundary fallback or bypass; same-state success requires its normal checks. |

- [ ] Run `go test . ./internal/mutation ./internal/lineage -count=1` and focused
  race checks, review and commit. Keep existing assignment creation tests intact.

## Task 3 — reusable CLI and bounded lab integration

**Files:** modify `application/api.go`, `cli/run.go`, `cli/scenario.go`,
`internal/lab/application.go`, `abv.go` inspection switch, relevant CLI/lab/binary
tests; add `cli/grant_status.go`, `internal/lab/grant_status.go` and their tests.
Update `implementation/abv/README.md`, `docs/local-testing.md`, `docs/acceptance.md`
and checkpoint progress with evidence, not new canonical rules.
**Consumes:** Task 2 facade/administration capability and existing lab marker.
**Produces:** commands below and optional application capability.

```sh
abv grant disable G2 --db PATH --tenant acme --app hrms --fixture-context maya-team1
abv grant enable G2 --db PATH --tenant acme --app hrms --fixture-context maya-team1
abv inspect grant-control G2 --db PATH --tenant acme --app hrms
```

- [ ] Write failing CLI-spy cases before parsing changes: missing context/ID,
  unknown verb, forbidden revision/recipient flags, absent optional API capability,
  exact Area/path/control forwarding, once-only close, output failure. Existing
  `inspect grant` continues to mean content, not silently replaced by control.
- [ ] Add the optional application interface and command dispatch. Construct
  `GrantControl{Version:"1", ID: positionalID, Status: mappedStatus}`; no new
  input-file/wire schema. A missing capability returns CLI status 5. Success
  prints existing canonical control JSON; typed errors retain existing statuses.
- [ ] Implement an explicit lab wrapper around the existing Administration.
  Its independent CheckGrantStatus premise is limited to direct human Maya,
  exact Area, G2, enable/disable and current membership in AssignmentAdmins.
  This same fixture group can hold two separately declared testing capabilities;
  the assignment method does not confer the new method's authority. Plain old
  adapters remain incapable, as tested in Task 2. No HRMS permission-registration
  rows or additional implicit owner powers are introduced.
- [ ] Wire that wrapper in lab Connect. Grant-status dispatch must verify the
  known fixture context, exact Area and existing scenario marker before calling
  the facade. The marker is still only accidental-misuse protection, not real
  authentication. Do not modify the base TeamFINC17 fixture counts or reset DBs.
- [ ] Add `inspect grant-control` using the existing control record. Verify
  identity/version/status and canonical JSON exactly; reject missing records and
  context mismatches. Never expose raw provider updates.
- [ ] Compile and test independent processes: seed disposable DB, assign A2,
  disable G2, inspect disabled control, verify G2-dependent diagnosis fails,
  enable G2, reopen control and verify diagnosis succeeds. Show the original
  assignment revision/state unchanged. Also test unmarked/wrong-context DBs,
  operation-specific admin refusal and unsupported adapters without writes.
- [ ] Run final `go test ./... -count=1`, `go test -race ./... -count=1`,
  `go vet ./...`, `go build -o ./bin/abv ./cmd/abv`, import-boundary checks,
  website build/tests and `git diff --check`. Independently review this task
  and final CP4-A integration in one bounded pass with separate verdicts.
  Commit/push verified checkpoints under standing authorization; preserve records.

## Plan self-review and deliberate limits

Task 1's write-set field is consumed by Task 2; its existing assignment member is
unchanged. Task 2's method and port signatures match the public facade and Task 3
adapter. Task 3's optional capability preserves old clients while returning
unsupported for a missing feature. Existing snapshot/transaction mechanisms are
reused; no dependency, migration, generic command bus or second authorization
engine is planned. CP4-P03 has both a positive case and its enabled/broken
counterexample. The model/source check replaces a redundant policy-approval loop.

This slice does not implement assignment activation after structural changes,
root management, direct-human/proxy source discovery, deletion, publication,
upgrades or parent changes. Their required scope remains in the main CP4 plan;
do not mark full CP4 complete when these three tasks pass.
