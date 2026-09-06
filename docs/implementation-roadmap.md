# Implementation roadmap — separate from handbook acceptance

This satisfies HC-11-04's documentary deliverable: an implementation roadmap
tied to verified and unverified gaps. It does **not** authorize implementation,
approve unfinished contracts, require runtime completion for handbook acceptance,
or certify a secure deployment. Phase labels are work references, not new policy.

## Evidence baseline

| Area | What is established | What is not established |
|---|---|---|
| Handbook decisions | Approved rules and core JSON are recorded with references. | Full contract closure and final v1 acceptance; see [closure checklist](v1-closure.md). |
| Handbook reader | Existing build and reader test suite are available for fresh verification; prior checks passed. | Authorization-engine correctness. Reader tests check page/asset behavior, not grant evaluation. |
| Illustrations | [Worked cases](use-case-examples.md) and dependency/Auth-service SVGs explain the design. | Executable conformance or exhaustive security coverage. |
| Historical prototypes | [Preserved source history](history/scratchpad-import/README.md) and retired explorers remain available. | Current-model compliance. No prototype is promoted into a production reference by this roadmap. |
| Deployment | No runtime/deployment verification is performed by this documentation pass. | Legacy proxy enforcement coverage, freshness guarantees, or end-to-end tenant isolation in a deployed system. |

## Dependency order

```text
I-01 accepted contract baseline
  ├── I-02 trusted identity + registration + initial setup
  └── I-03 record lifecycle + supporting-lineage resolution
          ↓ (requires I-02)
      I-04 Auth administrative gate + boundary-safe writes
          ↓
      I-05 embedded agent + application endpoint integration
          ↓
      I-06 freshness + concurrent-use + background integration
          ↓
      I-07 conformance, migration review, and deployment acceptance
```

This is logical dependency ordering, not a requirement for separate services,
repositories, or database tables. Design freshness guarantees before relying on
cache use; I-06 is their integration/verification stage, not permission to ignore
them during earlier design.

## Work packages and exit evidence

| Phase | Work and dependency | Governing sources / unresolved inputs | Required evidence before completion |
|---|---|---|---|
| I-01 | Freeze the approved v1 contract package and distinguish historical examples. | HC-07-07/08/09/10, HC-08-02/03/04, HC-11-02/03; [contract inventory](grant-contract-closure.md). | Versioned schemas/excerpts accurately labeled; unknown field/default/variant choices approved; final acceptance register names every exclusion/deferral. |
| I-02 | Establish trusted actor/human/tenant context, application registration, and governed initial authority. Depends on accepted relevant I-01 contracts. | Q-039–Q-042, Q-085–Q-087, Q-113–Q-116; full identity/registry/root trust and setup contracts remain open. | Tenant/identity mismatch cases, unsupported registration cases, explicit initial group membership, no ordinary root manufacture, no completed-bootstrap replay mutation. Legacy-token adapters require separate approved mapping and tests. |
| I-03 | Implement grant controls, immutable revisions, assignments, membership/team links, delegation limits, and top-down actual-support resolution. | Q-090–Q-112A; full support-discovery, hierarchy/delegation records, lifecycle validation remain open. | Direct/group/proxy route fixtures; no cross-grant mixing; parent/team ceilings; duplicate-assignment rejection; orphan/cycle handling; global versus route disablement; explicit latest-only adoption without changing old assignments. |
| I-04 | Guard Auth APIs with both administrative evaluation and proposed-authority boundary validation; persist exactly checked changes. Depends on I-02/03. | Q-093 / Q-100 / Q-110; concrete APIs and conflict/consistency contracts remain open. | Tests where administrative permission passes but source bounds fail, and vice versa; parent-change guard cases; concurrent support change cannot commit stale approved authority. |
| I-05 | Integrate the shared agent at one endpoint-owned gate with one permission, declared material, and constrained application execution. | Q-033 / Q-049 / Q-050-B–F; complete policy/request/result/adapter contracts remain open. | GET and PUT boundary mismatch cases; missing/source-invalid inputs; no prepared fallback; explicit deny/error handling; actual output/effects constrained by the authorized request. |
| I-06 | Establish and test freshness, mutation-to-consumer visibility, check-to-use consistency, and execution-time background attribution. | Q-069 / Q-074 / Q-075 / Q-110; cache protocol, membership propagation, in-flight behavior, and background adapter contracts remain open. | Confirmed deletion cannot authorize a later-started check from stale cache; races tested against the agreed ordering; timeout is not absence proof; queued work reauthorizes at execution with current human support. No unsupported TTL grace or deployment claim. |
| I-07 | Reconcile migration/compatibility and run final cross-domain and adversarial deployment review. Depends on the relevant preceding packages. | HC-10-03/04, Q-087-A, [reconciliation](reconciliation.md), [use cases](use-case-examples.md). | HRMS and repository end-to-end evidence, proxy coverage of every intended protected legacy path, preserved/adopted revision behavior, failure injection, reviewed outstanding-risk register, explicit deployment acceptance. |

## Coverage and ownership boundaries

Auth owns authorization registrations, grants, assignments, authorization
memberships, and its own protected write path. Application integration establishes
trusted domain material and enforces actual data boundaries. The agent crosses
these responsibilities without moving business-workflow design into Auth.

The roadmap does not select an implementation language, cache, database, identity
provider, or cross-service transaction mechanism. Those choices must satisfy the
accepted contracts. External audit-system design remains excluded under Q-076;
authorization-result evidence is still in scope.

## Verification handoff

For each implementation package, retain the governing approved references,
contract versions, executable fixtures, actual command results, and unresolved
limitations. A passing documentation build is never substituted for those tests.
Do not mark a package implemented from this roadmap alone. Security-critical
gaps in the [closure checklist](v1-closure.md) must be settled or explicitly
scoped before the corresponding implementation can claim conformance.
