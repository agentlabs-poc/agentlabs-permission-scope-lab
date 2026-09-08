# ABV implementation progress

## Current delivery — CP1–CP3 complete and independently reviewed

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
| CP4: lifecycle | Not started | Full affected-binding and enablement tests required. |
| CP5: registration/Auth integration | Not started | Trusted real administration and reviewed definition operations required. |
| CP6: PostgreSQL | Not started | Backend equivalence and transfer rehearsal required. |

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
├── CP4 lifecycle                                       PENDING
├── CP5 registration + real Auth integration             PENDING
└── CP6 PostgreSQL                                      PENDING
```

Three of six checkpoints are delivered for the local prototype. Checkpoints
differ in size; this is not a percentage-of-effort estimate. CP4–CP6 still need exact operation plans before
implementation, as stated in the approved execution plan.

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
