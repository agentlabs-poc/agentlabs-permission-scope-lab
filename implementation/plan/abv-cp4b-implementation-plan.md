# CP4-B Assignment Enable/Disable Implementation Plan

> **For agentic workers:** After CP4-B01 review, use
> `superpowers:subagent-driven-development`, Sol-medium coding with Astra-medium
> fallback if blocked. Execute one bounded task at a time with independent review.

**Goal:** Protected team-assignment enable/disable, including bottom-up guards.
**Architecture:** Extend the existing typed write set and two-gate facade; add
one pure reverse-binding lookup, then connect the reusable CLI. No generic CRUD.
**Tech Stack:** Existing Go module, standard library and pinned SQLite provider.
**Spec:** [CP4-B01 design](abv-cp4b-design.md). Status: complete and independently approved through `2681550`.

All four tasks and the final integration review are complete. The 60-minute
reassessment restricted continuation to a bounded review/fix/publication stage;
no further feature work was opened. See [delivery evidence](progress.md).

## Global constraints

- Tenant/application is mandatory outer context; no global-ID lookup or defaults.
- One existing assignment changes status; all other fields and records stay intact.
- Administrative authority is operation-specific and separate from ABV.
- Disablement obeys bottom-up affected-binding guards, including ineffective but
  still-enabled bindings. Never cascade status writes automatically.
- Enablement preserves adopted grant/role revisions and validates current support.
- Membership and ownership do not substitute for parent-team holding relationships.
- Disabled bindings remain visible for integrity, uniqueness and cycle checks.
- Unsupported dependency discovery returns an error, not permission to skip a route.
- Reuse current snapshot limit and 256-step lineage bound; reject, never truncate.
- No new canonical JSON, schema migration, dependency, root operation or real Auth.

## Work bounds and shared contracts

Coding task caps: 12, 15, 15 and 12 minutes. Each review: five minutes. One focused
fix unit per task, at most eight minutes, then reassess. Reassess the entire slice
at 60 minutes even if tasks remain; no silent extension or false completion.
Report midway through each coding unit. These are work-control limits, not SLAs.

All file paths below are relative to `implementation/abv/`. Use a fresh per-plan
execution ledger; do not reopen completed CP1–CP4-A tasks.

```go
// internal/storage/provider.go — internal instruction, not a wire record.
type AssignmentStatusChange struct { Before, After domain.Assignment }
// Add to existing WriteSet, preserving its other fields:
AssignmentStatusChange *AssignmentStatusChange

// internal/lineage/dependents.go — pure, SQL-free structural discovery.
func DependentTeamAssignments(context.Context, storage.Snapshot,
    string /* assignmentID */) ([]domain.Assignment, error)
```

The facade, coordinator, administrative capability and application signatures
are defined exactly in the spec. Do not change existing constructors or APIs;
add optional capabilities. Do not treat assignment creation as enablement.

## Task 1 — conditional assignment-status persistence

**Files:** modify `internal/storage/provider.go`, `internal/storage/sqlite/update.go`,
`internal/storage/contracttest/suite.go`; add
`internal/storage/contracttest/assignment_status.go`.
**Consumes:** Provider.Update, Assignment, strict DecodeAssignment.
**Produces:** AssignmentStatusChange and provider-neutral conformance coverage.

- [x] Add and invoke `RunAssignmentStatus(t *testing.T, factory Factory)`.
  First failing case, using existing fixtures/Factory/assertSnapshot:

```go
seeded := fixtures(t)[0]
path := t.TempDir() + "/authority.db"
p, err := factory.Create(t.Context(), path, []storage.Snapshot{seeded})
if err != nil { t.Fatal(err) }
defer p.Close()
before := seeded.Assignments["A1"]
after := before
after.Status = "disabled"
err = p.Update(t.Context(), seeded.Area, func(storage.Snapshot) (storage.WriteSet, error) {
    return storage.WriteSet{AssignmentStatusChange: &storage.AssignmentStatusChange{Before: before, After: after}}, nil
})
if err != nil { t.Fatal(err) }
seeded.Assignments["A1"] = after
assertSnapshot(t, p, seeded)
```

- [x] Run `go test ./internal/storage/... -count=1`; capture missing-type RED,
  then real no-persistence RED before implementing the write.
- [x] Accept only one write category: assignment creation, grant-status change,
  or assignment-status change. Reject every mixed combination before effects.
- [x] Validate both records through codec round-trip; compare copies with Status
  equalized and reject any other difference. Obtain exact persisted Before from
  the database inside the transaction, not the callback's mutable snapshot maps.
  Missing row is ErrNotFound; mismatch is ErrConflict. No upsert or recipient move.
  Persist exact After JSON and indexed status using parameterized SQL:

```sql
UPDATE assignments SET status=?, canonical_json=?
WHERE tenant_id=? AND application_id=? AND assignment_id=?
  AND grant_id=? AND grant_revision=? AND recipient_type=? AND recipient_id=?
  AND status=?
```

  Require one row. Preserve existing cancellation, rollback and error categories.
- [x] Cover enable/disable/same-state, every immutable field alteration, malformed
  version/status/ID, stale Before, forged callback snapshot, missing row, each
  tenant/app isolation dimension, every mixed-write pair, rollback and reopen.
  Reuse two-provider lock-contention pattern; competing callback count stays zero.
- [x] Run focused provider tests, provider race checks and full Go suite once on
  final source. Self-review, commit, then independent task review.

## Task 2 — reverse team-binding discovery

**Files:** add `internal/lineage/dependents.go` and `dependents_test.go`.
**Consumes:** complete Snapshot, actual adopted assignment/content/team records;
existing private `validateTeamChain`, `assignmentContent`, `maxChainSteps`.
**Produces:** DependentTeamAssignments; no mutation or administrative decision.

- [x] Write a failing external `lineage_test` case using `lab.TeamFINC17(area)`
  with A2 inserted, proving A1 discovers A2:

```go
fixture := lab.TeamFINC17(area)
fixture.Snapshot.Assignments["A2"] = fixture.Proposed
got, err := lineage.DependentTeamAssignments(t.Context(), fixture.Snapshot, "A1")
if err != nil || len(got) != 1 || got[0] != fixture.Proposed {
    t.Fatalf("dependents = %#v, %v", got, err)
}
```

- [x] Run `go test ./internal/lineage -count=1` and capture RED.
- [x] Implement one bounded structural index/walk, not a general graph framework:

```text
validate context/Area/catalog and requested existing group assignment
validate map IDs, strict assignment shapes and unique grant/recipient pairs
load each assignment's exact adopted content, never newest published content
index group holdings by (grantID, recipientTeamID)
index group children by (content.parentGrantID, recipientTeam.parentID)
walk actual selected ancestor bindings ignoring status for structural cycles
walk every descendant group binding, including explicitly disabled nodes
  validate relevant team-parent chains using existing structural helper
  reject repeated assignment/grant/team on an active path; bound path length
  reject unprovable potentially connected enabled non-group dependencies
return all descendant assignment records sorted by assignment ID, excluding self
```

  An edge requires BOTH parent-grant and parent-team relationships. A reusable
  definition without an assignment is not a binding. Missing upstream holding is
  an orphan, not automatic failure of a withdrawal inventory; it terminates that
  ancestor walk without manufacturing support. Malformed/missing adopted content
  or recipient team is incomplete evidence and rejects the operation.
  Track cycles through disabled records; status and expiry never prune discovery.
  Use visited-node indexing for branches and active-path sets for cycles, with
  context checks during construction/traversal. No O(n²) scan per visited node.
  For enabled non-group assignments whose adopted parent grant matches a visited
  grant, fail unsupported rather than invent which team supplies their source.
- [x] Test: three levels; branching; one grant held at unrelated teams; mismatched
  grant/team links; unassigned definitions; disabled bridge with enabled lower
  binding; cycles including disabled edges; missing upstream holding; missing
  content/team; duplicate bindings; cancellation; exact depth-bound overflow;
  direct-human ambiguity; newer unadopted content not affecting the graph.
  Assert empty output on every error and unchanged input snapshot.
- [x] Run lineage tests/race once on final source; self-review, commit and review.

## Task 3 — protected coordinator and facade

**Files:** add `internal/mutation/assignment_status.go`, `assignment_status_test.go`;
modify `internal/mutation/service.go`, `abv.go`, `abv_test.go`.
**Consumes:** Tasks 1–2, cloneSnapshot, validStoredAssignment, identity validation,
ResolveParentTeam, validation.Narrow and eligibleRoute.
**Produces:** SetAssignmentStatus and optional AssignmentStatusAdministration.

- [x] Add a failing public compatibility test using existing externalAdministration:

```go
got, err := facade.SetAssignmentStatus(t.Context(), area, fixture.Issuer, "A1", "disabled")
if !errors.Is(err, domain.ErrUnsupported) || got != (domain.Assignment{}) {
    t.Fatalf("old adapter gained assignment control: %#v, %v", got, err)
}
```

- [x] Capture RED with `go test . ./internal/mutation -count=1`.
- [x] Implement the exact operation sequence:

```text
validate ctx, explicit Area, direct-human identity, exact ID and desired status
require non-nil optional assignment-status administration capability
Provider.Update: verify returned Area/catalog; load/validate existing assignment
reject non-group recipient and trusted-root assignment management as unsupported
After = Before with only Status replaced
CheckAssignmentStatus(ctx, cloned evidence, identity, After, now)
discover the complete relevant structural inventory using Task 2
if disabling: reject if ANY descendant assignment remains enabled
if enabling:
  stage After in an isolated snapshot; do not alter any grant control
  resolve its exact adopted content against actual current parent-team holdings
  apply validation.Narrow; preserve permission subset and scope AND
recheck cancellation and all retained route validity windows
return the single AssignmentStatusChange Before/After
return After only when Update commits; otherwise zero Assignment
```

  A disabled/expired/missing upstream route need not become eligible to withdraw
  a structurally unblocked assignment. Enablement must prove eligibility. No
  historical-issuer membership dependency or automatic descendant activation.
  A same-state call still passes its applicable checks. Other recipients of the
  same grant retain their independent status; do not call SetGrantStatus.
- [x] Direct SQLite-backed matrix: enabled child blocks parent disable even when
  its grant is disabled/expired; leaf-up disables succeed; every fork is inspected;
  disabled bridge cannot hide enabled lower binding; no unrelated holding blocked
  merely by grant reuse; restoration fails with disabled grant/support or changed
  team parent lacking support; valid new reality restores; adopted revision stays
  fixed despite newer publication; other assignments unchanged; no-op still gated;
  absent/malicious/typed-nil admin; malformed/wrong-Area request; parent and child
  expiry crossing; cancellation/competing writer; depth/snapshot overflow; zero
  result and no write on every failure, including after reopen.
- [x] Focused normal/race and full Go tests; self-review, commit and task review.

## Task 4 — CLI, bounded lab authority and acceptance

**Files:** add `cli/assignment_status.go`, `internal/lab/assignment_status.go` and
focused tests; modify `application/api.go`, `cli/run.go`, `cli/scenario.go`,
`internal/lab/application.go`, CLI/lab/binary tests, README and acceptance docs.
**Consumes:** Task 3 facade and optional administration port.
**Produces:** optional AssignmentStatusAPI and commands:

```sh
abv assignment disable A2 --db PATH --tenant acme --app hrms --fixture-context maya-team1
abv assignment enable A2 --db PATH --tenant acme --app hrms --fixture-context maya-team1
```

- [x] Write failing CLI tests for exact ID/status/context forwarding, missing
  inputs/unknown verb/forbidden revision-recipient flags before connection,
  absent and non-pointer typed-nil capability (status 5, zero method calls),
  output failure, close once and old command compatibility. Reuse CP4-A nil helper.
- [x] Add the optional seam and dispatch. Success prints the existing canonical
  version-1 Assignment JSON. `assign` remains creation; `inspect assignment` is
  reused unchanged. Do not read the assignment in CLI before the protected call.
- [x] Add an explicit lab wrapper preserving both existing administrative methods.
  Use `AssignmentStatusAdministration` embedding `*GrantStatusAdministration`,
  constructed by `NewAssignmentStatusAdministration(area domain.Area,
  premise AdministrationPremise) (*AssignmentStatusAdministration, error)`;
  construct its embedded adapter through `NewGrantStatusAdministration` and wire
  the new wrapper in `lab.Connect`.
  Its separate assignment-status premise permits direct Maya, exact Area, current
  AssignmentAdmins membership and only the existing A1/G1/Team1 or A2/G2/Team2
  identities, with unchanged adopted revision and enabled/disabled status. The
  stored current revision is validated by ABV, not pinned by the lab premise.
  This is a separately declared testing capability, not authority inherited from
  assignment-create or grant-status permission. Preserve base fixture counts.
- [x] Verify marker, exact Area and fixture context before facade dispatch. Test
  unmarked/wrong-context databases and administrative refusal with no write.
- [x] Extend the existing compiled-process scenario without adding/resetting seeds:

```text
seed → create A2 → disable A1 rejected
disable A2 → disable A1 → enable A2 rejected
enable A1 → inspect A2 still disabled → enable A2
reopen A1/A2 and assert exact original fields except explicit status changes
inspect G1/G2 controls and contents unchanged; diagnosis succeeds again
```

- [x] Run full Go suite, full race suite, vet, binary build, import-boundary checks,
  site build and `node --test tests/*.test.mjs`, and diff/link checks. Record
  rationale, exact test evidence, prototype limits and reviewed commit IDs.
  One combined independent Task 4 + full-slice review with separate verdicts;
  one bounded fix pass and scoped re-review if required. Publish verified
  checkpoints under standing authorization, preserving all history/evidence.

## Self-review and handoff

Provider Before/After is consumed only by the coordinator. Structural discovery
does not grant authority and does not use current eligibility to erase bindings.
Facade/admin/application signatures match the spec; output reuses Assignment.
The lab demonstrates bottom-up order with existing A1/A2, so no fixture migration
or unrelated source-selection subsystem is needed. The branch/failure matrix is
spread across pure discovery, protected operations and real CLI persistence.

CP4-B01's scope/interfaces are approved for execution. Approval does not authorize
changing canonical rules or claiming assignment lifecycle is already implemented.
