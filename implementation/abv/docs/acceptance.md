# ABV first-slice acceptance evidence

## Current delivery

CP1, CP2 and CP3 are complete for the approved local prototype slice. Task 7
specification compliance, Task 7 quality and final CP3 integration received
independent approval at `5b81b84`, with no findings. Lifecycle operations, real
Auth/registration integration and PostgreSQL remain CP4–CP6; this is not full
ABV or production Auth completion.

CP4-A Task 3 provides the optional reusable `GrantStatusAPI`, bounded lab G2
enable/disable commands, and separate canonical `grant-control` inspection.
Binary acceptance assigns A2, disables G2, observes dependent diagnosis fail,
reenables G2, observes diagnosis recover, and proves A2 is byte-for-byte unchanged.
Missing optional adapters return status 5; marker, context, fixture and operation
refusals do not write. This evidence does not mark full CP4 complete.

- [CP2 provider evidence](sqlite-provider.md)
- [CP3 lineage evidence](lineage-resolution.md)
- [CP3 protected assignment evidence](assignment-creation.md)
- [Verified CLI walkthrough](local-testing.md)

## CP3 source-case map — independently reviewed

This maps every case in the [source contract, section 8](../../../docs/authority-boundary-validation.md#8-review-cases-and-rationale).
It is an evidence inventory, not a claim of 28 implemented operations. A
resolution check proves the effect of supplied state; it does not implement the
authorized lifecycle operation that creates that state. The laboratory admin
premise is not production authentication. Task 7 additions and this classification
have passed local verification and independent final review.

Test paths below are relative to `implementation/abv/`. The repeated failure
table is `TestCreateAssignmentFailuresDoNotWrite` in
[mutation/service_test.go](../internal/mutation/service_test.go); every row
checks the error, empty receipt and unchanged assignment count.

| Case | Evidence and supported result | Limit or remaining work |
|---|---|---|
| T01: FIN/C17 baseline | `TestCreateAssignmentPersistsExactProposalAfterBothChecks`; `TestSQLiteFacadeReopensCommittedAssignment` in `abv_test.go`. Exact A2 persists after both checks and survives reopen. | Lab administration only; real Auth adapter is CP5. |
| T02: source but no administration | Failure-table subtest `administration absent while source present`: no write. | Does not equate source membership with administration. |
| T03: administration but no source | Failure-table subtest `source absent while administration present`: no write. | Administration supplies no business authority. |
| T04: changed recipient | Failure-table subtest `wrong recipient`; `TestLabAdministrationRejectsAnyChangedTrustedPremise/recipient`. Both reject changed recipient premises. | Tests substitute Team1, not the source example's exact Team3 name; the bound recipient remains Team2. |
| T05: excessive permission | Failure-table subtest `selected permissions exceed parent`; compiled negative CLI scenario. Registered delete cannot be taken from the broader root through Team1's read/write ceiling. | No trimming of the submitted selection. |
| T06: empty child scope | `TestNarrowKeepsAllRestrictionsAndCopiesInputs` and `TestEmptyChildScopeRetainsParentPredicateAndPersists` in `acceptance_test.go`: FIN is retained through diagnosis, assignment and reopen. | Empty additional scope is not tenant-wide authority. |
| T07: FIN AND ENG | The same pure narrowing test preserves both predicates instead of overwriting FIN. | Contradictory-scope publication policy remains unresolved; no domain satisfiability claim. |
| T08: definition without holding | Failure-table subtest `parent support missing` leaves G1 content stored but rejects issuance. | Stored definitions are not authority. |
| T09: unrelated TeamX holding | `TestResolveParentTeamRejectsIneligibleOrInferredSupport/G1_only_at_TeamX` in `internal/lineage/resolve_test.go` returns an empty route and rejection. | TeamX cannot replace actual parent Team1. |
| T10: unrelated broader revision | `TestResolveParentTeamUsesOnlyActualTeam1Revision` preserves Team1's adopted revision and read/write ceiling. | No union with TeamX's broader revision. |
| T11: stale new assignment | Failure-table subtest `older selected revision has no latest fallback` rejects creation. | Publication and explicit upgrades are CP4 operations. |
| T12: disabled duplicate | Failure-table subtest `disabled duplicate still occupies binding`; provider conformance `grant recipient uniqueness spans revisions and disabled rows`. | Disablement does not free a current binding. |
| T13: disabled grant | Failure-table subtest `child control disabled`; resolver subtest `G1 control disabled`. Disabled content cannot support new issuance. | Actual grant enable/disable API is CP4; this is supplied-state evidence. |
| T14: disabled assignment | Resolver subtest `Team1 assignment disabled` and `TestHasSourceRevalidatesRouteAndIdentity/disabled_evidence_is_rechecked` reject the disabled support route. | Analogous parent-support proof, not an application access evaluator for A2; assignment lifecycle is CP4. |
| T15: ancestor restoration | No delivered lifecycle operation. | CP4 must prove enabled descendants regain effectiveness without a state rewrite. |
| T16: explicitly disabled descendant | No delivered lifecycle operation. | CP4 must prove restoring an ancestor does not enable an explicitly disabled child. |
| T17: enabled binding blocks reparenting | No delivered structural-write operation. | CP4: bottom-up affected-binding guard. |
| T18: reparent after binding removal/disablement | No delivered structural-write operation. | CP4: current-state validation and explicit re-enable, without automatic repair. |
| T19: another shared branch enabled | No delivered structural-write operation. | CP4 must check every affected branch, not only one. |
| T20: cycle with disabled bindings | Resolver subtest `cyclic team graph with disabled edge` rejects supplied cyclic team state. | Proposed grant-parent cycle rejection during structural writes remains CP4; the existing test does not prove that operation. |
| T21: cross-tenant support | `TestTeamFINC17RoundTripsAndTenantCannotSupplyMissingSupport` includes both another tenant and another app holding the otherwise missing A1; local lookup cannot borrow it. Provider conformance isolates every snapshot family and writes by both dimensions. | `TestInspectCarriesAreaForEverySupportedKind` checks facade forwarding; provider conformance proves isolation beneath that shared read seam. |
| T22: expired support | `TestResolveParentTeamValidityBoundaryAndCopies`, `TestHasSourceRechecksExactExpiry`, and `TestCreateAssignmentRechecksEligibilityImmediatelyBeforeWriteSet/G1` and `/G2` cover parent-only and child expiry crossing. | Current eligibility and pre-write recheck, not a production commit-time expiry guarantee. |
| T23: failed lookup | `TestCancellationDoesNotReplayMutationCallback`, `TestProviderFailureReturnsNoReceiptAndCallbackIsNotReplayed`, and provider cancellation/rollback tests preserve operational errors and no receipt. An expired deadline returns `context.DeadlineExceeded` before a callback. | This is local cancellation/deadline/conflict evidence, not a timeout injected midway through a remote Auth lookup; no network adapter exists. |
| T24: stale checked state | `TestFacadeDiagnosticIsNotATicketForAssignment`; `TestParentDisablementBeforeMutationAcquisitionConflictsThenIsSeen`; `TestParentDisablementAfterAssignmentCommitIsNotRetroactive`. | Withdrawal ordering is tested; a racing new-revision publication operation remains CP4. |
| T25: issuer later leaves | `TestResolveParentTeamBaseline` removes all membership evidence and still resolves otherwise valid team-held lineage; source-membership tests reject a fresh issuance attempt without Maya's source. | No permanent issuer field or dependency is invented; ownership lifecycle operations are not implemented. |
| T26: orphan support | Failure-table subtest `parent support missing` and missing-parent resolver cases reject stored but unsupported lineage without replacement. | CP4 owns authorized support removal and its affected-binding guards; no cascade cleanup is implemented. |
| T27: actual certificate belongs to ENG | Outside ABV: the application must enforce its actual-data boundary. | No HRMS database or application access evaluator is implemented here; this remains an integration responsibility. |
| T28: copied `$self` binding | `TestHasSourceLeavesDirectHumanAndSelfBindingExplicitlyUnsupported` and resolver subtest `child self changes recipient binding` return unsupported. | Correct fail-closed prototype limit, not a settled canonical binding/distribution policy. |

### Portability and concurrency evidence reused

[Provider conformance](../internal/storage/contracttest/suite.go) already checks
installation requirements, both isolation dimensions, complete record round
trips, rollback, reopen persistence, uniqueness including disabled bindings,
snapshot limits and two-provider read consistency. SQLite-specific tests also
prove cross-query snapshot consistency, failed second-insert rollback and
competing writers without callback replay. [Mutation consistency tests](../internal/mutation/consistency_test.go)
exercise the coordinator on file-backed SQLite with a separate writer.

Reusing these suites avoids a duplicate concurrency implementation. PostgreSQL
has not run this contract and remains CP6. Snapshot-state tests do not turn raw
test SQL into a supported lifecycle API.

### Task 7 verification and deferred-review evidence

The coordinator independently ran the following on the Task 7 candidate, after
the implementer reported its own passing run:

```text
go test ./... -count=1                           PASS
go test -race ./... -count=1                     PASS
go vet ./...                                    PASS
go build -o ./bin/abv ./cmd/abv                  PASS
go test . -run '^$' -bench '^BenchmarkBoundedAuthorityDiagnosis$' -benchtime=3x -count=1
                                                PASS (all three sizes)
git diff --check                                PASS
```

The import graph was inspected with `go list`: CLI imports the application seam
and domain, not SQL or validation; lineage/validation import no SQL or CLI;
storage imports no CLI; the facade imports no CLI or lab. The facade's reviewed
`OpenSQLite` constructor composes the SQLite provider without exposing raw writes.

`TestSnapshotLimitReturnsNoReceiptAndNoWrite` deliberately opens the scenario
with a 30-record limit, observes the explicit snapshot-limit error and empty
receipt, then reopens at the normal limit and proves A2 absent. Provider
conformance additionally proves no callback on overflow; bounded traversal
returns no partial route.

`TestAdministrativeSnapshotMutationCannotAlterBusinessEvidence` mutates the
adapter's control map, membership slice, trusted-root map, permission slice and
nested expiry pointer. A successful assignment and whole-original-snapshot
comparison demonstrate isolation without masking aliasing behind an already
disabled grant. Other copied field families are not independently mutated by
this test. Parent-only and child-only expiry crossings are separately covered.

Input-file close errors now propagate through `readInput` instead of being
discarded. Existing file-input tests pass. There is no deterministic test
injecting a read-only `os.File.Close` failure: no general filesystem abstraction
was introduced solely for that test. This limitation remains visible for review.

The final reviewer closed all three prior minor findings: unmasked snapshot
isolation evidence, parent-only expiry evidence and input-file Close handling.
The documented deterministic Close-failure injection limitation is retained;
it was not treated as an unimplemented error-handling branch. The complete CP3
changed surface was reviewed without rerunning already-recorded suites. Task 7
and final integration were assessed together in one bounded pass, with separate
specification, quality and merge-readiness verdicts. No review surface was waived.

### Bounded snapshot benchmark — 8 September 2026

Measured on Linux/amd64, Intel i7-1185G7, Go benchmark suffix `-8`, three timed
iterations per size. Exact coordinator-run output:

| Loader-counted records | ns/op | B/op | allocs/op |
|---:|---:|---:|---:|
| 100 | 716,603 | 225,048 | 5,304 |
| 1,000 | 6,045,637 | 2,612,701 | 56,177 |
| 10,000 | 58,614,506 | 25,278,218 | 564,799 |

The fixture is a shallow valid team/grant route plus unrelated grant/control
records, not a 10,000-deep chain. The executable count assertion includes one
application row, permission/scope/support rows and area-owned records; the
installation row is not counted by the loader. Each benchmark opens with its
exact record-count limit; 10,000 equals the provider's normal default ceiling.
Database and fixture creation are outside the timed loop.

These measurements cover SQLite snapshot loading and read-only boundary
diagnosis. They do not establish source/administrative authority and do not
measure mutation, administrative evidence copying, commit, real Auth calls or
application enforcement. Three local samples are not a latency SLA, capacity
forecast or PostgreSQL result. The measured allocations document the current
whole-snapshot cost; optimization is not part of this acceptance checkpoint.

## Historical CP1 acceptance evidence

**Current state: CP1 implemented, verified and independently reviewed.**
This page preserves CP1's evidence and limits at delivery, when no SQLite or
production ABV tests existed. SQLite persistence is now independently verified
in [CP2 evidence](sqlite-provider.md); actual lineage and protected assignment
are covered by the CP3 links above. Production Auth remains later work.
Statements below about missing persistence or resolution describe CP1.

## Scope

Task 1: mandatory outer context, core record representations, internal errors,
reusable application interface and no-write CLI shell.

Task 2: strict supported grant/assignment JSON, registered definitions and exact
role selection, permission subsets and accumulated scope/validity restrictions.

## Test-first evidence

- `go test ./domain ./cli -count=1` first failed because `Area`, core records,
  errors and `Run` were absent. After implementation both packages passed.
- `go test ./internal/codec ./internal/validation -count=1` first failed because
  decoders, `CheckContent` and `Narrow` were absent. Implemented those interfaces
  and checked their behavior with positive and negative fixtures.
- An escaped-key fixture initially double-escaped its key, making it a distinct
  literal rather than a duplicate. Corrected the test input after observing the
  exact decoded value; the real escaped duplicate is rejected.
- Initial full `go test ./... -count=1`, race suite and vet passed before the
  subagent hardening pass. The post-review evidence is recorded below.

## Hardening and verification — 8 September 2026

The sol-medium coding subagent added failing tests exposing unpaired Unicode
surrogate repair, present-empty parent ambiguity, invalid UTF-8 in typed content,
corrupt selected definition metadata and malformed CLI grammar. It then fixed
those behaviors. Valid Unicode remains exact; no new authority fields were added.
The human actor fixture was corrected to canonical `user`, not `human`.

The subagent reported a bounded three-second decoder fuzz run with two workers:
8,142 executions, no failure. This is bounded fuzz evidence, not exhaustive proof.

The coordinator independently ran these commands on the finished code:

```text
go test ./... -count=1          PASS: 22 named tests plus the fuzz seed corpus
go test -race ./... -count=1    PASS
go vet ./...                   PASS, no diagnostics
go build ./...                 PASS (packages, not a CLI executable yet)
git diff --check               PASS
```

Existing website regression suite: all 10 tests passed; `npm run build` passed.
The checkpoint has no external Go dependencies, SQLite driver, server or raw
write method. CLI currently admits grant/assignment inspection syntax; wider
inspection kinds from Task 6 will be implemented with their adapters at CP3.

## Independent review — correction before integration

The reviewer examined checkpoint `5751b62` and found an important role-path gap:
`CheckContent` could validate a child selecting a read-only role, while the old
`Narrow` interface accepted a separately supplied write permission from a broader
parent. A comment requiring correct expansion was not a sufficient API invariant.
The code has not been deployed or integrated with any persistence path.

The correction removes that expansion parameter. Both helpers derive direct or
exact-role-revision permissions through a shared implementation, and narrowing
must retain the complete selected set or reject it. It cannot substitute a
different parent permission or silently trim a role. Regression tests and a
scoped independent re-review were completed before CP1 completion.

The reviewer also requested an explicit cancellation-category choice. Standard
`context.Canceled` and `context.DeadlineExceeded` are retained for classification
with `errors.Is`, rather than inventing a canonical public error code.

**Re-review result at `6a192f4`:** both findings addressed; no new CP1 regression;
spec compliant, code quality approved and ready to merge. The coordinator also
reran full tests (25 named tests plus fuzz seeds), race detection, vet and package
build on the corrected code, all passing. The initial 22-test result above is
retained as evidence of the earlier checkpoint, not the final test count.

## Limits and source-case coverage

Source-case coverage is deliberately limited:

| Source case | CP1 evidence | Remaining proof |
|---|---|---|
| C01-T05: permission expansion | `TestNarrowRejectsPermissionAndOuterBoundaryEscape` rejects a permission outside the parent. | Actual source discovery and no-write transaction test in CP3. |
| C01-T06: child `{}` | `TestNarrowKeepsAllRestrictionsAndCopiesInputs` retains FIN. | End-to-end supported assignment in CP3. |
| C01-T07: FIN AND ENG | The same narrowing test preserves both predicates. | No new contradictory-scope publication policy is chosen. |
| C01-T21: cross-tenant support | Unit checks reject mismatched tenant or application on a supplied route. | Provider lookups, installation checks and actual lineage isolation in CP2/CP3. |
| C01-T22: validity | Time limits are decoded and deep-copied into narrowed routes. | Current-time eligibility at resolution and mutation. |
| C01-T28: cross-recipient self | Registration may recognize `$self`; this is not binding proof. | Still unresolved; no supported issuance claim. |

All other source cases remain later-checkpoint work; C01-T27 is application
actual-data enforcement, not an ABV success criterion. See the
[source matrix](../../../docs/authority-boundary-validation.md).

They do not establish actual parent-team support, source possession, authenticated
administration, provider isolation, latest revision selection, current time
eligibility, atomic save or application actual-data enforcement. Those belong to
later checkpoints. No whole C01 scenario matrix is claimed passing at CP1.
