# ABV implementation progress

## Current delivery — CP1 and CP2 complete and independently reviewed

Started after the user approved the six-checkpoint plan. Baseline handbook,
presentation and plans were committed and pushed as `64b0831`. Isolated worktree
setup is `bac62de`; coding work runs on `implementation/abv-cp1` before integration.
The user subsequently requested sol-medium coding subagents.

| Checkpoint | Current state | Exit evidence |
|---|---|---|
| CP1: records, decoder, pure checks | Complete; role-source fix independently approved at `6a192f4` | [CP1 evidence](../abv/docs/acceptance.md) |
| CP2: SQLite provider | Complete; all review fixes approved at `8ac7593` | [Provider evidence](../abv/docs/sqlite-provider.md): isolation, rollback, persistence, cancellation and cross-query consistency verified. |
| CP3: resolver and working CLI | In progress: Task 4 lineage first | Actual team support, both gates, atomic assignment, restart demo required. |
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
├── CP3 working in-process ABV + CLI                     IN PROGRESS
│   ├── Task 4: actual parent/team lineage resolution    ACTIVE
│   ├── Task 5: administration + ABV, atomic assignment
│   ├── Task 6: inspect/check/assign commands + restart demo
│   └── Task 7: adversarial and end-to-end acceptance
├── CP4 lifecycle                                       PENDING
├── CP5 registration + real Auth integration             PENDING
└── CP6 PostgreSQL                                      PENDING
```

Two of six checkpoints are delivered. Checkpoints differ in size; this is not
a percentage-of-effort estimate. CP2 does not provide runtime authorization or
working CLI mutation commands. CP4–CP6 still need exact operation plans before
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

Next: CP3 / Task 4. Reuse the existing plan and preserved execution evidence;
do not redispatch CP1/CP2 or claim either supplies the two authorization gates.
