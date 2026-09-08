# ABV implementation progress

## CP5-A started — registration persistence first

The user approved [CP5-A](abv-cp5a-design.md), then requested no review passes:
get it working first. Sol-medium implementation and mandatory tests continue;
do not describe new CP5 work as independently reviewed. The first bounded
[task](abv-cp5a-provider-plan.md) implements application-only catalog persistence.
Protected registration, computed roots and CLI acceptance remain subsequent work.
This grouping separates persistence proof from administrative authorization.
Existing completed CP1–CP4-A/B evidence below is unchanged.

## Current delivery — CP1–CP3 and CP4-A/B complete and independently reviewed

**Scope correction — 9 September 2026:** production Auth-service integration is
outside this ABV build, not remaining required work. CP5 previously mixed ABV
definition management with a production adapter; only the former remains here.
Historical integration-gap statements below are context, not completion blockers.

**Further user deferral:** CP6 PostgreSQL, deletion and parent changes are not
required for the current build. Their plans and canonical rules remain preserved
for later. Publication/adoption and CP5 definition management are not omitted.
Rationale: finish the bounded SQLite ABV/CLI build without backend migration or
destructive/structural editing operations.

Started after the user approved the six-checkpoint plan. Baseline handbook,
presentation and plans were committed and pushed as `64b0831`. Isolated worktree
setup is `bac62de`; CP1 originally ran on `implementation/abv-cp1`. The retained
worktree now uses `implementation/abv-cp3` for the completed first slice.
The user subsequently requested sol-medium coding subagents.

| Checkpoint | Current state | Exit evidence |
|---|---|---|
| CP1: records, decoder, pure checks | Complete; role-source fix independently approved at `6a192f4` | [CP1 evidence](../abv/docs/acceptance.md) |
| CP2: SQLite provider | Complete; all review fixes approved at `8ac7593` | [Provider evidence](../abv/docs/sqlite-provider.md): isolation, rollback, persistence, cancellation and cross-query consistency verified. |
| CP3: resolver and working CLI | Complete; Task 7 and final integration approved at `5b81b84` | [Acceptance](../abv/docs/acceptance.md): two gates, actual lineage, CLI/restart, no-write failures, source-case map and bounded benchmarks. |
| CP4: lifecycle | CP4-A and CP4-B complete; full assignment-control slice approved through `2681550` | [Acceptance](../abv/docs/acceptance.md#cp4-b--protected-assignment-enabledisable), [CP4-B01 design](abv-cp4b-design.md) and [execution plan](abv-cp4b-implementation-plan.md); other lifecycle operations remain. |
| CP5: ABV definition management | CP5-A catalog persistence in progress | Add-only permission/scope registration first; protected flow, computed roots, CLI and roles remain. Production Auth integration excluded. |
| CP6: PostgreSQL | Deferred by user; not required now | Provider boundary retained; backend implementation/transfer rehearsal are later work. |

### Execution tree

```text
ABV implementation
├── CP1 foundations                                      COMPLETE
│   ├── Tenant/application context + canonical records
│   ├── Strict JSON decoder
│   └── Pure definition/subset checks + reusable CLI seam
├── CP2 SQLite persistence                               COMPLETE
│   ├── SQL-free provider + schema + controlled fixtures
│   ├── Isolation, persistence and transaction conformance
│   ├── All independent-review findings resolved
│   └── Provider rationale, evidence and transaction SVG
├── CP3 working in-process ABV + CLI                     COMPLETE
│   ├── Task 4: actual parent/team lineage resolution    COMPLETE
│   ├── Task 5: administration + ABV, atomic assignment  COMPLETE
│   ├── Task 6: inspect/check/assign commands + restart demo COMPLETE
│   └── Task 7: adversarial and end-to-end acceptance    COMPLETE
├── CP4 lifecycle                                       TWO SLICES COMPLETE
│   ├── CP4-A grant enable/disable                       COMPLETE / REVIEWED
│   ├── CP4-P03 disabled binding clarification           RECONCILED
│   ├── Conditional provider control update             COMPLETE / REVIEWED
│   ├── Two-gate coordinator/facade                     COMPLETE / REVIEWED
│   ├── CLI and integration tests                       COMPLETE / REVIEWED
│   ├── CP4-B assignment enable/disable                  COMPLETE / REVIEWED
│   │   ├── Conditional assignment persistence          COMPLETE / REVIEWED
│   │   ├── Reverse team-binding discovery              COMPLETE / REVIEWED
│   │   ├── Protected coordinator/facade                COMPLETE / REVIEWED
│   │   └── CLI/lab/acceptance + final review            COMPLETE / REVIEWED
│   ├── Publication / explicit adoption                 PENDING
│   └── Deletion / parent changes                       DEFERRED BY USER
├── CP5 ABV definition management                       PENDING
└── CP6 PostgreSQL provider                             DEFERRED BY USER
```

Three numbered checkpoints are delivered for the local prototype. Checkpoints
differ in size; this is not a percentage-of-effort estimate. CP4-A additionally
delivers grant-wide enable/disable; CP4-B delivers protected team-assignment
enable/disable. Remaining current-build work is publication/adoption and CP5
definition management, each requiring a bounded operation plan. PostgreSQL,
deletion and parent changes are deferred, not completed. Real Auth-service
integration is excluded, not part of this build's completion denominator.

### CP4-B final delivery and rationale

`fb6e9a5` connects the optional CLI seam and independently bounded lab capability;
`2681550` closes final acceptance/documentation findings. Astra-medium approves
Task 4 specification, quality and the complete CP4-B integration with no Critical
or Important findings. All four coding units finished inside their caps, each
task used at most one focused correction, and final review used one fix wave.

At the 60-minute reassessment, implementation and tests were complete. Work was
restricted to a bounded closure stage (review, corrections and publication), not
another feature. Final full Go/race/vet/build, dependency/import checks and site
build/10 tests pass. Local file-link checks pass. Existing records, provider,
resolver and CLI patterns were reused; no migration or dependency was introduced.

The only deferred minor is the future validity-pointer graph-test fingerprint
limitation described below; the final reviewer accepted it as nonblocking.
Production Auth integration was removed from required scope after the user's
correction. Permission/scope/role management remains a distinct ABV concern.
No next lifecycle operation or external integration has been started.

### CP4-B Task 1 — conditional assignment persistence (historical checkpoint)

`3a5d30b` adds one conditional Before/After assignment-status write. Only status
may change; the provider reads the persisted Before inside the transaction, so
mutating the callback snapshot cannot forge a match. Exact tenant/application
scoping, strict records, mixed-write refusal and rollback preserve all other data.

Sol-medium implementation and independent spec/quality review completed within
their 12/five-minute caps, with no findings. Provider focused/race and full Go
tests passed; controller full Go/vet and site build/10 tests also passed.
This is internal persistence only: protected assignment controls still require
reverse binding discovery, both authority gates and CLI acceptance in Tasks 2–4.

### CP4-B Task 2 — reverse team-binding discovery (historical checkpoint)

`5e9b4c0` adds pure, indexed discovery using actual grant and team parents,
exact adopted revisions and complete disabled-binding inventory. Missing upstream
holdings remain distinct from malformed evidence; ambiguous enabled human routes
return unsupported. Scope/permission eligibility is not used to erase bindings.

Independent review found a delimiter-based uniqueness collision for legal IDs.
`6414d80` uses an exact comparable tuple, with a reproduced RED/GREEN regression.
The same bounded fix strengthens unchanged-input assertions. Scoped review approves
specification and quality; focused/race and controller full Go/vet tests pass.
A minor future-test limitation was accepted as deferred in final review: the snapshot fingerprint
does not inspect pointees of non-nil validity fields; current graph fixtures use
none. No runtime mutation or authorization bypass was found.

### CP4-B Task 3 — protected assignment status (historical checkpoint)

`c2d1d24` connects the operation-specific administrative gate, complete binding
inventory and boundary validation to the conditional write. Disablement is
bottom-up; enablement preserves exact adoption and validates current support.
Neither path changes another record or inherits administration from possession.

Review identified invalid-UTF8 ID handling and five ignored test-read errors.
`34319e4` corrects both, including a RED/GREEN regression proving malformed IDs
cannot enter Update. Focused normal/race and controller full Go/vet pass; scoped
review approves specification and quality with no remaining Task 3 findings.
SQLite operation tests cover forks, disabled bridges, unrelated shared holdings,
changed support, expiry, evidence isolation, bounded snapshots and no-write reopen.
The optional library operation is delivered; CLI/lab acceptance remains Task 4.

### CP4-A delivery and rationale

`f4a781c` adds conditional provider control updates; `b5a486b`, `2b9b4b2`
and `7ea2aba` add the two-gate coordinator and direct operation-specific matrix.
`ccf7eef` connects the reusable CLI and bounded lab adapter. The final review
found one incomplete typed-nil capability guard; `ed8f227` covers all nil-capable
kinds with a regression that first reproduced an unintended method call.
Scoped re-review approved the fix, Task 3 spec/quality and whole-slice integration.

All coding used Sol-medium; Astra-medium supplied the final integration review.
Each coding task finished within its 12/15/12-minute cap; one focused final fix
finished within its eight-minute cap. No coding escalation or policy reopening
was needed. Final full Go tests/race/vet/build and the site build/10 tests pass;
import boundaries remain intact. The existing provider, record and CLI paths
were reused without a dependency, migration or new canonical format.

This completion excludes assignment activation, deletion, parent changes,
publication/upgrades, root management, direct-human/proxy source discovery and
real Auth integration. It is a tested local prototype slice, not full ABV.

### Engineering choices and rationale

- Internal authority checks explicitly carry `Area`, and internal `Route` keeps
  that context. This makes cross-tenant/application route mixing rejectable.
  It changes internal call signatures, not the canonical JSON scope format.
- The CLI is a reusable command shell at CP1: help works; recognized data
  operations return unsupported, never successful reads or writes. Persistence
  and protected dispatch arrive together in CP2/CP3, avoiding a temporary bypass.
- Parser limits and narrow supported-input rules are prototype restrictions,
  not new public canonical schema decisions. Unsupported forms are not dropped.
- A successful pure subset check does not establish membership, active lineage,
  cross-recipient self binding, administration, or permission to save.
- Independent review found that the initial role-based narrowing API could
  accept a loose permission expansion not bound to the selected role. The fix
  derives the whole selected list from the exact role record, shared with
  definition validation. It cannot silently trim or substitute that list.
  This is an internal API correction, not a canonical model change.

No whole-handbook criterion or production Auth integration is closed by CP1.

CP1 delivery: `5751b62` implemented the foundation; `6a192f4` fixed the reviewed
role-source binding issue. Independent review then approved spec compliance,
code quality and integration. All 25 named tests plus fuzz seeds, race checks,
vet and package build pass. At that delivery, CP2 was the next checkpoint.

### CP2 delivery and rationale

Implementation `e43be35` added the SQL-free provider, SQLite schema, shared
application catalog, mandatory installation/context checks, controlled new-file
fixtures and provider-neutral conformance. `badb2fd` addressed four independent
review findings; `8ac7593` corrected the resulting migration error-classification
regression. Final independent verdict: spec compliant, quality approved, ready
to merge; no findings remain open for CP2.

The fixes ensure uncertain `BEGIN` outcomes are cleaned up, operational marker
query failures are not treated as missing support, malformed catalog values
cannot weaken checks, and evidence reads demonstrably use one snapshot across
a concurrent commit. Error classification preserves existing typed categories
and joins. These enforce approved guarantees, not new canonical policy.

Full Go tests, race tests, vet and build pass on final source, as do dependency
verification and staged whitespace checks. SQLite statement coverage is 72.6%,
reported as test coverage only. The existing documentation site builds and all
10 site tests pass. See the provider evidence for concrete tests and limitations.

### CP3 / Task 4 milestone

`0fa7163` implemented actual parent/team lineage, explicit source membership and
the reusable FIN/C17 scenario. `4550860` corrected final-source uniqueness
revalidation and added a genuine exact-expiry regression. Independent review:
spec and quality approved, ready for Task 5. Full tests/race/vet/build pass.

The resolver now takes exact selected child content rather than implicitly
choosing a revision from an ID; parent support follows actual adoptions. The lab
fixture explicitly establishes RootTeam/A0/G0 so a stored root-shaped definition
does not manufacture trust. Direct-human discovery, proxies and cross-recipient
self binding remain unsupported implementation cases, not new canonical bans.

At this Task 4 checkpoint, Task 5 was next; its delivery is recorded below.

### CP3 / Task 5 milestone

`651bb2a` adds the reusable facade, separately bounded lab administration and
transactional assignment creation. `d7fea85` corrects a scheduler-dependent
test failure by acknowledging the committed receipt before parent withdrawal.
Independent implementation review and scoped fix review close the Task 5 gate.
Full tests, race checks, vet and build pass on the corrected source.

The public evidence alias makes the administrative port implementable outside
internal packages without exposing raw provider writes. Administrative evidence
is deep-copied; ABV validates the original snapshot. Pre-write time eligibility
is not a production commit-time guarantee, and the administrative adapter remains
a controlled lab premise. Two nonblocking evidence improvements are retained for
Task 7: independently isolate mutated snapshot fields and cross parent-only expiry.

Task 6 is next. Its plan binds the parsed database path through an injected
connector and requires distinct scenario provenance for fixture-based mutation.
The optional lab metadata does not change canonical records or core schema v1.
Working CLI commands and first-slice acceptance are still required to close CP3.

### CP3 / Task 6 milestone

`fb836ba` implements the reusable CLI and controlled local scenarios. `5815abf`
closes a cleanup edge case and corrects the compiled missing-file regression
to exercise file opening against a valid seeded database. Independent review
and scoped fix review approve specification compliance and quality. Parent-run
full Go tests, race checks, vet and build pass on the corrected source.

The independent CLI demonstration saves A2 and reopens it in another process;
the out-of-bound permission scenario rejects and independently proves A2 absent.
The scenario marker guards accidental fixture use, not hostile database owners
or real authentication. Read-only diagnosis is never a save ticket. An input-file
close-error minor remains explicitly tracked in Task 7 alongside the source-case
matrix, isolation evidence and bounded-snapshot benchmarks.

Execution is now explicitly bounded: each work unit has scope, exit criteria,
time and retry/review limits. Overrun or stalled progress requires reassessment,
not another unchanged loop. This preserves security and verification gates.

### CP3 / Task 7 and final first-slice acceptance

`5b81b84` adds empty-child-scope persistence/reopen and snapshot-overflow no-write
proofs, bounded 100/1,000/10,000-record diagnosis benchmarks, stronger original
snapshot isolation evidence, parent-only expiry crossing and input-close error
propagation. All 28 source cases are classified against exact evidence or stated
limits; this does not mean all 28 operations are implemented.

Independent Task 7 specification, quality and complete CP3 integration verdicts
are approved with no findings. All three prior review minors are closed; the
absence of deterministic read-only file Close-failure injection remains documented.
The parent independently passed full Go tests, race tests, vet, executable build,
benchmarks, website build and all 10 website tests. No new dependency, canonical
decision, Auth-service change or duplicate concurrency suite was introduced.

The coding unit completed within its ten-minute budget. One bounded final review
covered both acceptance and integration with separate verdicts; no repeated
review loop was needed. Evidence remains available in the acceptance document.

Next: plan CP4's exact lifecycle operations and transaction effects, starting
with grant/assignment enable-disable and dependent effectiveness. Publication,
reparenting, affected shared branches and explicit upgrades remain in CP4's
required scope; they are not silently omitted or already implemented.
