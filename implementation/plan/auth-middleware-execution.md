# AUTH-MW-02 — bounded execution ledger

User requested autonomous continuation after pinning API-side middleware/evaluator
and Auth-service-side core ABV placement. Current activity begins with M0 contract
design from auth-middleware-implementation-plan.md, not guessed wire schemas.

## Active unit

M0: authmw_contract, Sol-medium, dispatched 9 September 2026 09:23 UTC.
First attempt stops by 09:43 UTC. Exclusive output: auth-middleware-contract.md.
Controller separately prepares auth-middleware-acceptance.md and records/publishes
documentation already requested. No worker commit/push or shared-file editing.

M0 report must establish exact internal types, trust/evidence boundaries, freshness
limits, permitted first-slice behavior and unknown production integration choices.
Existing policy/result JSON is reused. A missing canonical policy is not settled
by coding it. Local internal harness conventions are labeled as such.

After an adequate handoff, controller prepares exact bounded implementation briefs;
no worker is dispatched against undefined interface names. M1–M4 remain pending.
No review passes; tests still required. Sol-medium first; once per failed bounded
unit, a narrowed Astra-medium attempt may run for at most 15 minutes. Repeated
failure causes regrouping, not silent looping.

Prior three 123 design workers completed their drafts; do not redispatch them.
Their output is proposed design, not proven DDL, approved permissions or runtime.
Storage migration remains separate from this middleware work.

AUTH-MW-04: during M0 the user authorized a direct SQLite adapter for local
validator/evaluator testing. Worker notified to replace fixture-only authority
with existing-schema read-only SQLite authority adapter planning. No HTTP service
needed. M0 deadline remains unchanged; code package seam must keep SQL and ABV
mutations outside middleware evaluator logic. ABV already has its SQLite provider.
