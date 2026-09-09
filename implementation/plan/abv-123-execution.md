# ABV-123-08 — bounded parallel contract design

User authorizes parallel autonomous subagents and requests setup followed by a
separate discussion of middleware and the core validator. This dispatch is design
only: no runtime replacement, migration, production Auth integration or new policy
approval. Existing source decisions govern all proposals.

## Work division

| Unit | Exclusive output | Exit condition | Bound |
|---|---|---|---|
| D1 operation mapping | abv-123-operation-matrix.md | Existing operations, approved/proposed permissions, required evidence, boundaries and actual implementation coverage traced to sources. | 20 minutes |
| D2 wire contracts | abv-123-wire-contracts.md | Reused record JSON, proposed versioned wire examples and failure/retry semantics; unresolved choices explicit. Does not choose D1 permission mappings. | 20 minutes |
| D3 storage contract | abv-123-storage-contract.md | Proposed one-store boundary keys/DDL, integrity/index/query map and bounded performance proof. Does not change public JSON or D1/D2 contracts. | 20 minutes |

All outputs live under implementation/plan/. Agents use Sol-medium with no
children, reviews, commits or pushes. Disjoint draft files permit parallel work;
integration happens only after handoff. Each records source/rationale, decisions
still needed, and exact uncompleted work if its cap expires. No silent extensions.

If a bounded Sol attempt cannot deliver, controller assesses the blocker and may
dispatch Astra-medium on the narrowed remainder once, capped at 15 minutes. A
policy ambiguity is surfaced, not decided by using a larger model. Stop after
that attempt and regroup rather than repeating the same loop.

Controller owns this ledger, synthesis and source reconciliation. The validator
is core; middleware is an integration responsibility. Naming or merging the
Auth Evaluator and authority-boundary validator is not authorized implicitly.
The upcoming user discussion may refine that distinction; workers must not preempt it.

Dispatched 9 September 2026 at08:38UTC:
D1 abv123_operations, D2 abv123_wire, D3 abv123_storage, all Sol-medium.
All first attempts stop by08:58UTC. No result is yet verified or approved.
Keep prior CP4-C implementation and evidence intact. Adoption coding stays paused.

All three workers subsequently returned their exclusive design drafts before
the cap, with worker-reported diff checks. No integrated contract approval or
runtime proof is claimed. D3 explicitly identifies folded support-list reference
integrity as requiring proof; a sixth table is a fallback proposal, not approved.
Controller middleware discussion uses consolidated handbook chapters5–7 and
canonical responsibility definitions. Current recommendation maps API-side
Auth Middleware to embedded evaluator integration and core Auth Validator to
the existing ABV responsibility, not a merged request/issuance decision engine.
