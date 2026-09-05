# Handbook completion audit — MEASURE-001

Publication follow-up: the user has now authorized commit/push and continued
discussion. The frozen-gate statements in this audit describe its original
snapshot. Publishing the audit does not change its 34/69 score or imply approval
of every proposed checkpoint grouping.

Audit date: 2026-09-05. Coverage: approved discussion through Q-050-F, with
SVG-001 and the local reconciliation included. Base commit:
`8b44880c1a63a71402dc335d8ed23db6f2205719`, plus the uncommitted working tree.

This is a measured **checkpoint-closure score under an explicit audit rubric**,
not estimated effort, a production-readiness rating, or an approved change to
v1 scope. This initial checkpoint grouping is proposed as the tracking baseline;
the arithmetic is reproducible, but choosing the granularity is a judgment.
The user requested a measured score rather than the previous approximate 60%.

## Result

- Defined checkpoints: **69**.
- Completed checkpoints: **34**.
- Open checkpoints: **35**.
- Closure score: **34 / 69 × 100 = 49.3%**.
- Remaining-checkpoint score: **35 / 69 × 100 = 50.7%**.
- Approved scope exclusions applied by this audit: **0**.

The previous 60% was an impression and must not be represented as measured.
It is superseded for this audit by the explicit count above. This does **not**
mean work went backwards; the old number had no fixed denominator.

## Method and completion rule

Use the eleven existing [roadmap](handbook-roadmap.md) stages. Break their
current deliverables and still-open bundles into independently closable
checkpoints, listed below. Do not count every question, every repeated mention
in a chapter, every example, or deprecated alternatives as separate progress.

Each row is worth one checkpoint, without subjective effort weights:

- **DONE (1):** the precise criterion is satisfied by an approved rule and its
  documented rationale/examples, or by the specifically identified documentary
  deliverable. The evidence column identifies the basis.
- **OPEN (0):** missing, proposed, partially specified, parked, or not yet verified
  to the required scope. Existing discussion earns no fractional credit.
- A completed principle does not complete its wire schema, lifecycle, or
  operation-specific contract. Those are separate promised outputs.
- A diagram, passing reader tests, or archived simulator does not demonstrate
  authorization implementation or complete conformance.
- Parked work is not excluded work. Only an explicit approved v1 deferral can
  move a row out of the denominator; record that separately, not as completion.
- No new authorization behavior is selected here. Where a feature is open,
  deciding and documenting its exclusion can resolve the question; the audit
  does not mandate implementing it.

The compact tree has uneven detail: several open bullets bundle schemas,
operations, and lifecycle mechanisms together. Counting its lines directly would
give those bundles too little weight. This audit makes those closure points
visible. Conversely, it does not award multiple points for every individual
question under already-grouped criteria.

The counts measure **closure of this checklist**, not the percentage of hours,
pages, risk, or engineering effort remaining. A small-looking open concurrency
contract may cost more work than several completed conceptual checkpoints.

## Score by roadmap stage

| Stage | Done | Total | Remaining |
|---|---:|---:|---:|
| 1. Purpose and architecture | 2 | 4 | 2 |
| 2. Principles | 4 | 5 | 1 |
| 3. Vocabulary and identity | 3 | 5 | 2 |
| 4. Permissions | 2 | 4 | 2 |
| 5. Grants and authority | 7 | 13 | 6 |
| 6. Scope and registration | 5 | 7 | 2 |
| 7. Requests and resolution | 6 | 10 | 4 |
| 8. Decision semantics | 1 | 4 | 3 |
| 9. Enforcement and time | 2 | 9 | 7 |
| 10. Challenge and verify | 2 | 4 | 2 |
| 11. Publish the foundation | 0 | 4 | 4 |
| **Total** | **34** | **69** | **35** |

## Detailed closure register

`HC-xx-yy` identifiers identify measurement rows, not canonical permission,
policy, or decision identifiers. “DONE” covers only the wording of that row,
not the whole surrounding topic.

### 1. Purpose and architecture

| Checkpoint | Status | Completion criterion | Evidence or outstanding gap |
|---|---|---|---|
| HC-01-01 | DONE | Define canonical Layer 1 versus application Layer 2 ownership. | ARCH-004 / Q-036; [system overview](system-overview.md). |
| HC-01-02 | DONE | Define how the embedded Auth Agent integrates both layers. | ARCH-005 / Q-037; [system overview](system-overview.md). |
| HC-01-03 | OPEN | Finalize handbook audience, products governed, and mandatory applicability. | Stage 1 [roadmap](handbook-roadmap.md) and [current tree](discussion-tree.md) leave audience/governance open. |
| HC-01-04 | OPEN | Define ownership of rule changes, exception approval, and governance. | Stage 1 [roadmap](handbook-roadmap.md); no completed governance contract. |

### 2. Principles

| Checkpoint | Status | Completion criterion | Evidence or outstanding gap |
|---|---|---|---|
| HC-02-01 | DONE | Establish the trusted implicit tenant boundary and its non-bypass rule. | TENANT-001; [grant model](grant-model.md) and [scope model](scope-model.md). |
| HC-02-02 | DONE | Establish that resolution and delegated authority cannot amplify access. | RESOLUTION-001 / AUTHORITY-002 / DELEGATION-002; [grant model](grant-model.md). |
| HC-02-03 | DONE | Require prevention of protected execution when necessary authority or enforcement fails. | PRINCIPLE-001 / ENFORCEMENT-002; [endpoint authorization](endpoint-authorization.md). |
| HC-02-04 | DONE | Require auth-first endpoint design and a server-owned declaration. | PRINCIPLE-002 / CONTRACT-004/007; [endpoint authorization](endpoint-authorization.md). |
| HC-02-05 | OPEN | Complete the numbered principle catalog with required-versus-guidance wording and compliance/counterexamples. | Stage 2 finished result in [roadmap](handbook-roadmap.md); individual principles exist, final consolidation remains open. |

### 3. Vocabulary and identity

| Checkpoint | Status | Completion criterion | Evidence or outstanding gap |
|---|---|---|---|
| HC-03-01 | DONE | Settle team/group synonymy, Auth ownership, and human-only authorization membership. | TERM-001 / Q-045; [groups and membership](groups-and-membership.md). |
| HC-03-02 | DONE | Distinguish grant/assignment, role, and group without independent expanded assignments. | TERM-004 / ROLE-001 / RESOLUTION-003; [grant model](grant-model.md). |
| HC-03-03 | DONE | Settle permission/boundary/request-material vocabulary without a canonical target wrapper. | TERM-005 / Q-043; [vocabulary](authorization-vocabulary.md). |
| HC-03-04 | OPEN | Complete the actor/principal/subject/user/membership identity glossary. | Stage 3 [tree](discussion-tree.md); current vocabulary chapter intentionally covers only part of the glossary. |
| HC-03-05 | OPEN | Specify trusted identity/tenant mappings and human/proxy attribution rules. | Identity and provenance gaps in [tree](discussion-tree.md) and [endpoint authorization](endpoint-authorization.md); does not require an application-specific database schema. |

### 4. Permissions

| Checkpoint | Status | Completion criterion | Evidence or outstanding gap |
|---|---|---|---|
| HC-04-01 | DONE | Separate the operation identified by a permission from scope reach. | PERMISSION-001; [grant model](grant-model.md). |
| HC-04-02 | DONE | Mandate exactly one required permission covering each protected method/route. | CONTRACT-008 / Q-049; [endpoint authorization](endpoint-authorization.md). |
| HC-04-03 | OPEN | Finalize permission naming/catalog governance and evolution rules. | Stage 4 [tree](discussion-tree.md); historical lab grammar is not an adopted full catalog contract. |
| HC-04-04 | OPEN | Decide v1 treatment of permission hierarchies, wildcards, and aliases. | Stage 4 [roadmap](handbook-roadmap.md); explicit exclusion would also close this item, not assumed support. |

### 5. Grants and authority

| Checkpoint | Status | Completion criterion | Evidence or outstanding gap |
|---|---|---|---|
| HC-05-01 | DONE | Keep the complete recipient/permissions/scope/conditions binding intact, including multiple permissions and direct/group routes. | GRANT-001/002/003; [grant model](grant-model.md). |
| HC-05-02 | DONE | Make group-based human grants preferred while permitting direct grants. | GROUP-004 / Q-021; [groups and membership](groups-and-membership.md). |
| HC-05-03 | DONE | Define live role bundles and their effect on referencing grants. | ROLE-001/002; [grant model](grant-model.md); concrete revision mechanics are separately open. |
| HC-05-04 | DONE | Separate grant administration from business access and require explicit authorized audited self-assignment. | ADMIN-001/002/003; [grant model](grant-model.md). |
| HC-05-05 | DONE | Settle the five ordinary-model administrative authorization rules. | ADMIN-004/005 / Q-044; [grant model](grant-model.md). |
| HC-05-06 | DONE | Settle ordinary grant survival after the issuer loses issuance authority. | ADMIN-006 / Q-046; [grant model](grant-model.md). |
| HC-05-07 | DONE | Require all service/agent access to depend on a human and stay within that human's authority. | AUTHORITY-002 / DELEGATION-002; [grant model](grant-model.md). |
| HC-05-08 | OPEN | Encode administrative recipient/permission/scope bounds and define containment validation. | Q-044 explicitly leaves encoding/containment open; [grant model](grant-model.md). |
| HC-05-09 | OPEN | Define bootstrap/seed authority and its governed creation procedure. | Bootstrap remains a gap in [grant model](grant-model.md) and [use cases](use-case-examples.md). |
| HC-05-10 | OPEN | Specify delegation ceilings, chains, expiry, growth, and reactivation semantics. | Human dependency is agreed; these mechanisms remain open in [grant model](grant-model.md). |
| HC-05-11 | OPEN | Complete ordinary grant status/validity and create/change/revoke lifecycle rules. | [Grant model](grant-model.md) still labels lifecycle details open; issuer-loss behavior alone does not close them. |
| HC-05-12 | OPEN | Define group membership change, nesting, and optional-sync lifecycle behavior. | [Groups and membership](groups-and-membership.md) and [grant model](grant-model.md); freshness timing is counted under stage 9. |
| HC-05-13 | OPEN | Define role revision/change validation and evidence for existing referencing grants. | Live role semantics are agreed, revision and compatibility-change mechanics remain open in [grant model](grant-model.md). |

### 6. Scope and registration

| Checkpoint | Status | Completion criterion | Evidence or outstanding gap |
|---|---|---|---|
| HC-06-01 | DONE | Settle flat string-value scope syntax, validation, and explicit empty versus missing/null scope. | SCOPE-007 / Q-034; [scope model](scope-model.md). |
| HC-06-02 | DONE | Settle AND within a scope and representation of alternatives through separate grants. | SCOPE-008; [scope model](scope-model.md); complete-grant evaluation is counted separately under decisions. |
| HC-06-03 | DONE | Anchor self to the authorizing human and require application-defined boundary meanings. | SELF-001 / Q-038; [scope model](scope-model.md). |
| HC-06-04 | DONE | Require application registration of permissions/scope contracts with abstract Auth validation. | REGISTRATION-001 / Q-039; [application registration](application-registration.md). |
| HC-06-05 | DONE | Settle optional upfront support-validation mode, all-grant checks, and safe activation. | REGISTRATION-002/003/004 / Q-040/041/042; [application registration](application-registration.md). |
| HC-06-06 | OPEN | Complete registration format, definition-management authority, distribution, and change/removal/versioning rules. | Remaining registration mechanics in [application registration](application-registration.md); parked does not mean closed. |
| HC-06-07 | OPEN | Finish reference/existence handling and treatment of application-specific exact/subtree boundary behavior. | Scope/reference gaps in [tree](discussion-tree.md) and [application registration](application-registration.md); explicit delegation or exclusion can close them, not a requirement to add built-in types. |

### 7. Requests and resolution

| Checkpoint | Status | Completion criterion | Evidence or outstanding gap |
|---|---|---|---|
| HC-07-01 | DONE | Establish one endpoint-owned gate and its logical request flow without a prepared handoff. | CONTRACT-006 / SVG-001; [endpoint authorization](endpoint-authorization.md) and [system overview](system-overview.md). |
| HC-07-02 | DONE | Distinguish request inputs from a resolved evaluation view and from an authorization decision. | RESOLUTION-005 / Q-047/047-A; [endpoint authorization](endpoint-authorization.md). |
| HC-07-03 | DONE | Define dependent resolved grants and membership/direct/group retrieval with preserved bindings. | RESOLUTION-006 / Q-048; [grant model](grant-model.md). |
| HC-07-04 | DONE | Settle the top-level string version convention and unsupported/malformed-version rejection. | CONTRACT-009/010 / Q-050-A; [contract publication](contract-publication.md). |
| HC-07-05 | DONE | Settle the partial endpoint-policy shape and GET/PUT source/name bindings without a relationship block. | CONTRACT-011/012 / Q-050-B/C; [endpoint policy](endpoint-policy-format.md). |
| HC-07-06 | DONE | Settle required input presence/source binding and application-owned value validation. | INPUT-002/003 / Q-050-E/F; [endpoint policy](endpoint-policy-format.md). |
| HC-07-07 | OPEN | Complete remaining structural endpoint-policy validation, nested input syntax, and supported-source decisions. | Q-050 remains open in [endpoint policy](endpoint-policy-format.md); no reopening of agreed fields or validation ownership. |
| HC-07-08 | OPEN | Complete grant and role wire schemas, including the agreed lifecycle representation. | [Grant formats](grant-format.md) are working illustrations, not complete published contracts. |
| HC-07-09 | OPEN | Complete authorization-request, resolved-request, and resolved-grant transport contracts. | Meaning is agreed; concrete contracts remain open in [endpoint authorization](endpoint-authorization.md) and [grant model](grant-model.md). |
| HC-07-10 | OPEN | Specify the handler/embedded-agent integration contract and failures at its boundaries. | Exact integration APIs remain open in [system overview](system-overview.md); no requirement to implement an SDK for handbook completion. |

### 8. Decision semantics

| Checkpoint | Status | Completion criterion | Evidence or outstanding gap |
|---|---|---|---|
| HC-08-01 | DONE | Settle alternative complete positive grants, no cross-grant field mixing, and no explicit deny grants in v1. | DECISION-001/002 / GRANT-001; [grant model](grant-model.md). |
| HC-08-02 | OPEN | Finalize decision outcomes, reasons, and missing/unknown/failure behavior. | Decision-result branch is open in [tree](discussion-tree.md); fail-closed principle is not a full result contract. |
| HC-08-03 | OPEN | Specify conditions and their evaluation when evidence is missing, invalid, or unsupported. | Stage 8 [roadmap](handbook-roadmap.md) and open decision branches in [grant model](grant-model.md). |
| HC-08-04 | OPEN | Define decision-result restrictions and contributing-grant/dependency provenance fields. | Decision result/audit explanation is open in [tree](discussion-tree.md); event storage is separately counted under stage 9. |

### 9. Enforcement and time

| Checkpoint | Status | Completion criterion | Evidence or outstanding gap |
|---|---|---|---|
| HC-09-01 | DONE | Require actual output/effects to stay within authorized boundaries and request bindings. | CONTRACT-012 / Q-050-C; [endpoint policy](endpoint-policy-format.md). |
| HC-09-02 | DONE | Require review of effective constraints rather than mere input usage, with counterexamples. | ENFORCEMENT-003 / Q-050-D; [endpoint policy](endpoint-policy-format.md). |
| HC-09-03 | OPEN | Define collection, count, export, and row/field restriction contracts. | Stage 9 [roadmap](handbook-roadmap.md) and [endpoint authorization](endpoint-authorization.md). |
| HC-09-04 | OPEN | Define create/update/move authorization for existing and proposed state. | PUT binding is not a move policy; [endpoint policy](endpoint-policy-format.md) leaves these rules open. |
| HC-09-05 | OPEN | Define bulk authorization, partial success, and operation transaction semantics. | Bulk branch remains open in [tree](discussion-tree.md). |
| HC-09-06 | OPEN | Define freshness, caching, revocation propagation, and stale membership/dependency behavior. | Stage 9 [tree](discussion-tree.md); conceptual dependency invalidation does not establish propagation timing. |
| HC-09-07 | OPEN | Define concurrent-change/check-to-use consistency guarantees. | [System overview](system-overview.md) leaves concurrency/atomicity mechanisms open. |
| HC-09-08 | OPEN | Complete audit event, correlation, retention/storage, versioning, and disclosure rules. | Stage 9 [roadmap](handbook-roadmap.md); explicit audited assignment alone is not an audit contract. |
| HC-09-09 | OPEN | Define non-HTTP/background-operation integration requirements or explicitly defer them. | [Endpoint authorization](endpoint-authorization.md) lists non-HTTP adapters as open; no new independent service authority. |

### 10. Challenge and verify

| Checkpoint | Status | Completion criterion | Evidence or outstanding gap |
|---|---|---|---|
| HC-10-01 | DONE | Document cross-domain worked cases with grants, material, and expected outcomes. | Sixteen scenario groups in [use-case examples](use-case-examples.md); counted as examples, not complete conformance. |
| HC-10-02 | DONE | Distinguish design claims, historical implementations, and unverified gaps in an evidence register. | [Reconciliation register](reconciliation.md) explicitly records limitations and archived prototypes; this is not credit for implementing authorization. |
| HC-10-03 | OPEN | Complete an adversarial expected-outcome suite mapped to the finalized rules. | Current examples deliberately omit open operation/freshness/administration cases; [use-case examples](use-case-examples.md). |
| HC-10-04 | OPEN | Close end-to-end HRMS/repository scenario review against the final contracts. | Stage 10 [tree](discussion-tree.md) remains open; no credit from reader-rendering tests. |

### 11. Publish the foundation

| Checkpoint | Status | Completion criterion | Evidence or outstanding gap |
|---|---|---|---|
| HC-11-01 | OPEN | Finish whole-handbook terminology and cross-chapter consistency review. | Stage 11 [roadmap](handbook-roadmap.md); local reconciliations are checkpoints, not final editorial acceptance. |
| HC-11-02 | OPEN | Package the complete versioned contract definitions and consistent examples for v1. | [Contract publication](contract-publication.md) and [tree](discussion-tree.md); separate from authoring individual schemas in stage 7. |
| HC-11-03 | OPEN | Obtain v1 acceptance with every gap settled or explicitly approved for exclusion/deferment. | [Roadmap](handbook-roadmap.md) closure criterion; parked/unanswered items are not presumed deferred. |
| HC-11-04 | OPEN | Publish a separate implementation roadmap tied to remaining verified/unverified gaps. | Stage 11 [roadmap](handbook-roadmap.md); this does not require implementing production authorization. |

## Overlap and scope controls

- Stages 1–6 settle ownership, concepts, and lifecycle meaning. Stage 7 settles
  representations and integration contracts. Stage 11 packages and reviews the
  complete handbook for release. Publishing a partial example does not satisfy
  all three.
- Group/role change semantics belong to stage 5; the cross-system freshness and
  concurrent-change guarantees belong to stage 9.
- Stage 8 defines the evaluator's decision/result evidence. Stage 9 defines
  the audit event lifecycle and storage/disclosure obligations.
- Stage 10's case authoring and final scenario review are different from building
  a production authorization engine. This score does not require completing
  an engine to finish the handbook, but implementation claims need evidence
  or explicit unverified status.
- The remaining permission hierarchy decision does not reopen SCOPE-007's ban
  on arbitrary wildcard/array/expression scope syntax.
- The scope-reference row does not impose universal department/project types,
  revive a canonical target wrapper, or require a relationship block.
- Retired prepared modes, independent service authority, and rejected policy
  fields are historical alternatives, not outstanding deliverables.

## How to update without returning to estimates

1. Keep the row IDs and criteria stable once this baseline is accepted.
2. Change OPEN to DONE only with linked evidence that satisfies the whole row.
3. If a row is only partly handled, leave it OPEN and record what remains.
4. If the scope or granularity must change, record the change and recompute the
   denominator explicitly. Do not silently compare different baselines.
5. Report completed/open counts with the percentage. Record approved exclusions
   separately so a deferment cannot masquerade as finished work.
6. Recheck final consistency and release acceptance; a high subtotal cannot
   override an unresolved security contract or publish the handbook by itself.

This report records analysis only. It does not approve the rubric as canonical,
settle open questions, alter prior agreements, commit, or push. The user's
commit/push gate remains frozen.
