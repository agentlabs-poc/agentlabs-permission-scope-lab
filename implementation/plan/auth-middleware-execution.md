# AUTH-MW-02 — bounded execution ledger

User requested autonomous continuation after pinning API-side middleware/evaluator
and Auth-service-side core ABV placement. Current activity begins with M0 contract
design from auth-middleware-implementation-plan.md, not guessed wire schemas.

## M0 handoff

M0: authmw_contract, Sol-medium, dispatched 9 September 2026 09:23 UTC.
First attempt stops by 09:43 UTC. Exclusive output: auth-middleware-contract.md.
Controller separately prepares auth-middleware-acceptance.md and records/publishes
documentation already requested. No worker commit/push or shared-file editing.

Handoff received 09:34 UTC, within the original bound. The contract draft now
defines the in-process authority-source/evaluator seam, existing policy/result
JSON, completeness and error behavior, handler responsibility, and the requested
SQLite test adapter. No runtime code is delivered by this checkpoint.

Before adapter coding, settle its package wiring under the ABV module so existing
internal read/lineage code is reusable without an evaluator dependency on SQLite.
The current issuance resolver rejects `$self`: literal SQLite integration and
resolved-fixture `$self` consumer tests are separate acceptance claims. Missing,
malformed or unsupported authority must never become an empty successful read.

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

## Implementation started after user continuation

- M1 policy/result codecs: authmw_m1, Sol-medium, 20-minute attempt; owns only
  new implementation/authmiddleware module. Request extraction remains M3.
- SQLite read-only opener: authmw_sqlite_reader, Sol-medium, dispatched 10:12 UTC,
  20-minute attempt; owns only readonly.go/readonly_test.go in existing storage.
- Auth-owned human route query: authmw_authority_read, Sol-medium, dispatched
  10:13 UTC, 20-minute attempt; owns human.go/human_test.go and narrow resolver
  inactivity diagnostics. No new authority model or issuance semantics.

These are independent file sets; user explicitly requests useful parallelism.
Controller owns contract wiring, acceptance and publication. Exact task briefs and
RED/GREEN reports are in the plan's ignored SDD directory. No worker commits,
pushes or reviews. Completion requires fresh test evidence, not worker dispatch.

Next checkpoint is regrouped as M1/M2 + SQLite adapter + in-process runnable test
command, before M3 HTTP plumbing. Rationale: the user specifically wants direct
SQLite testing of the validator; this proves the core sooner without pretending
the API wrapper is finished. M3 and constrained application-handler demonstrations
remain required for the original complete middleware slice. No policy is deferred
or changed by this sequencing. Reassess after sixty minutes of coding.

### Verified checkpoint at 10:23 UTC

M1 policy/result codecs complete: fresh controller `go test ./... -count=1` and
`go vet ./...` pass; worker also records race and RED/GREEN evidence. Read-only
SQLite opener complete: fresh focused `TestOpenReadOnly|TestReader` tests and
SQLite vet pass; worker reports full ABV regression passing. No HTTP behavior or
complete SQLite-to-evaluator integration is claimed yet.

M2 dispatched to authmw_m2 (Sol-medium) at 10:22 UTC; twenty-minute attempt.
The independent authority-query task continues within its original bound.
