# ABV implementation progress

## CP1 — foundations: implemented, independent review pending

Started after the user approved the six-checkpoint plan. Baseline handbook,
presentation and plans were committed and pushed as `64b0831`. Isolated worktree
setup is `bac62de`; coding work runs on `implementation/abv-cp1` before integration.
The user subsequently requested sol-medium coding subagents.

| Checkpoint | Current state | Exit evidence |
|---|---|---|
| CP1: records, decoder, pure checks | Implementation and hardening verified; independent review pending | [CP1 evidence](../abv/docs/acceptance.md) |
| CP2: SQLite provider | Not started | Isolation, rollback, persistence and concurrency conformance required. |
| CP3: resolver and working CLI | Not started | Actual team support, both gates, atomic assignment, restart demo required. |
| CP4: lifecycle | Not started | Full affected-binding and enablement tests required. |
| CP5: registration/Auth integration | Not started | Trusted real administration and reviewed definition operations required. |
| CP6: PostgreSQL | Not started | Backend equivalence and transfer rehearsal required. |

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

No whole-handbook criterion or production Auth integration is closed by CP1.
