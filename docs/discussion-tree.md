# Authorization handbook — discussion tree

This is the navigation and coverage map for the eleven-stage
[roadmap and decision log](handbook-roadmap.md). The log is authoritative for
individual decisions; this tree shows where each discussion belongs and what
still needs a conclusion. Examples are in [grant examples](grant-examples.md)
and [endpoint-completion cases](endpoint-completion-cases.md).
Start with the [working handbook](handbook.md) to read the detailed chapters.
The [working grant chapter](grant-model.md) consolidates definitions, rationale,
examples, and unresolved boundaries for the active branch.

PLAN-001 establishes the eleven-stage roadmap. PROCESS-001 governs reference
IDs and proposal-versus-agreement status; PROCESS-002 governs tree traversal;
PROCESS-003 requires detailed, reconstructable notes and working chapters.
PROCESS-004 requires meaningful checkpoint commits/pushes and periodic
reconciliation across the handbook, log, tree, examples, and original lab prose.
PROCESS-005 requires justification and reuse checks before adopting any new field.
PROCESS-006 resumes recording and preserves earlier designs with deprecation labels.

## Current position

- **Current: stage 5, grant administration — ADMIN-004/005, Q-043.** Q-042
  is agreed as REGISTRATION-004. The user directed ending the registration
  detour; remaining lifecycle details are parked, not finalized or removed
  from v1. Return to the saved administration branch.

```text
Scope and registration → agreed foundations; remaining details parked
Grant administration  → CURRENT: canonical model, then assignable bounds
Requests/resolution   → NEXT after remaining grant questions
```

### Historical positions before the return to grant administration

- **Current: Q-041 approved REGISTRATION-003.** Relationship validation is an
  explicit application-level registration choice; enabled applies to every
  grant, including existing and role-based grants. Q-042 is active: enabling
  validation when existing grants are incompatible. Earlier positions below
  are retained history, not live questions.

- **Current: Q-040 approved REGISTRATION-002 with optional relationship
  declarations in registration.** Q-041 is active: what Auth validates when
  that metadata is absent. Individual permission/key registration and runtime
  scope requirements remain required. The Q-039 position below is history.

- **Current: Q-039 is approved as REGISTRATION-001.** Applications register
  scope and permission contracts; Auth validates grant acceptance while remaining
  domain-agnostic. Q-040 is next: explicit permission-scope compatibility.
  Return afterwards to remaining scope governance and ADMIN-004/005.

### Preserved position history through Q-038

- Current as of Q-038: application-owned scope meanings, supported target
  relationships, and endpoint-specific trusted fact bindings are agreed.
  Q-028's conceptual question is answered; SCOPE-004's detailed validation
  mechanics remain open. Q-039 is now active: sharing definition contracts
  across grant validation and request evaluation. Earlier active-position notes
  below are retained history, not additional unanswered copies of Q-038.

- Latest: Q-037 / ARCH-005 is agreed, closing the embedded-agent sidebar below.
  Auth and application supply their respective material; shared evaluation and
  endpoint enforcement remain one gate. Return to stage 6: Q-038 now asks about
  explicitly defined scope-key meanings and supported target relationships.
  Earlier "clarify" and "no next question" notes below are historical positions.

- Q-036 / ARCH-004 is agreed: canonical Layer 1 and application-specific
  Layer 2 jointly establish authorization within the single endpoint-owned gate.
  Clarify the embedded auth agent's integration across these sources, then return
  to scope-key meanings and ownership. See [system responsibilities](system-overview.md).

- Active branch: **6. Scope and target → scope-key definitions and governance**.
  Sidebar concluded: CONTRACT-007 / Q-035 is agreed. The endpoint declares its
  permission, material, and sources. [System block diagram](system-overview.md)
  includes the requested SVG; return to scope-key definitions now.
  SCOPE-006/008 and SCOPE-007's canonical v1 format are agreed; Q-034 is answered.
  Next: scope-key definitions/governance, before returning to ADMIN-004/005.
  SCOPE-004 remains open. SCOPE-002's earlier typed-format direction is historical;
  its definition-governance questions remain open under the canonical key-value model.
  Q-025 approved this detour; return to ADMIN-004/005. Q-024's fields stay withdrawn.
  The detailed chapter is [scope and target](scope-model.md).
- Just concluded: SCOPE-007 / Q-034 — flat string-value scope, explicit {} for
  tenant-wide reach, no missing/null default, and strict validation.
  CONTRACT-006's one endpoint-owned gate remains current with INPUT-001 and
  ENFORCEMENT-002; the older two-mode design remains deprecated.
- Returned from the group/self sidebar to bounds on grant administration and
  remaining lifecycle questions in stage 5;
  then declared request and resolved
  request/grants. Return to the remaining siblings in the tree before v1.
- A parent branch stays open until all its required children are settled or
  explicitly excluded from v1. Settled examples do not close the whole branch.
- Current grant layouts have been reconciled in [current grant formats](grant-format.md).
  The six earlier layouts remain deprecated examples; GRANT-EX-007 uses canonical
  scope. Complete lifecycle and resolved-grant schemas are not declared finished.
- Reconciliation checkpoint: [register](reconciliation.md) identifies current
  versus historical material without deletion. [Cross-domain use cases](use-case-examples.md)
  adds 16 scenario groups; this does not close stage 10 or interrupt the scope-key branch.

## Historical mind map at a glance — before the return to stage 5

The position markers in this snapshot are preserved history. The current
position above and detailed tree below now place us in grant administration.

This overview is the user's requested whole-handbook mind map. It shows progress
by topic, not by a guessed completion percentage. The detailed tree below keeps
the individual reference IDs and open siblings. No entire stage is claimed
complete simply because its central concept is settled.

```text
Authorization Handbook — working v1
├── 1. Purpose/authority       Core boundary agreed; governance open
├── 2. Principles             Core invariants agreed; remaining rules open
├── 3. Vocabulary             Several terms agreed; identity/fact terms open
├── 4. Permissions            Operation vs reach agreed; catalog/grammar open
├── 5. Grants                 Bindings/roles/dependencies agreed; RETURN HERE
│   └── Administration        Separation/audit agreed; assignable bounds open
├── 6. Scope                  Definition + v1 format agreed; WE ARE HERE
│   └── Next                  Define keys, their meanings, and ownership
├── 7. Requests/resolution    One endpoint-owned gate agreed; data shapes open
├── 8. Decision semantics     AND/alternative grants agreed; remaining cases open
├── 9. Enforcement/lifecycle  Safety rule agreed; freshness/audit/concurrency open
├── 10. Challenge/verify      16 cross-domain cases; comprehensive validation open
└── 11. Publish v1            Final reconciliation and implementation roadmap open
```

The immediate path is:

```text
6. Scope-key definitions and remaining scope semantics
   → 5. Administrative grants and grant lifecycle
   → 7. Authorization request and resolved request/grant contracts
   → Remaining open siblings across all stages, then validation/publication
```

Before moving between branches, update this position and preserve the return
point. Deprecations are visible history, not unfinished work we must revive.
Current grant-format reconciliation does not close administrative bounds or
declare that the original application handbook/UI implements the new design.

## Traversal rules

PROCESS-002: keep the full tree, stable references, conclusions, and a return
point whenever discussion branches. User requested this explicitly.

1. Record each new question beneath its parent branch with its reference ID.
2. On agreement, update the decision log, examples affected, and this tree.
3. On a detour, keep the unfinished parent and siblings visible. Resume the
   parent after the detour, unless the user steers to another branch.
4. Keep rejected/superseded paths as history; do not reopen them without a new
   user request or concrete contradiction requiring discussion.
5. Before closing a stage, review every open child and cross-stage dependency.
6. Before publication, verify coverage of every stage and reconcile the glossary,
   examples, requirements, and implementation gap register.

## Tree

Legend: **settled** means an agreed concept, **active** means the current question,
**open** means unfinished, and **working** means a proposal needs consolidation.
No complete stage is claimed below.

```text
Authorization handbook
├── 1. Purpose and authority [open]
│   ├── Shared rules across AgentLabs [settled: CHARTER-001]
│   ├── Auth versus application responsibility [settled: CHARTER-002]
│   ├── Reusable evaluator/application integration [working: ARCH-001]
│   ├── Canonical/application responsibility layers [settled: ARCH-004 / Q-036]
│   ├── Embedded agent integrates Auth/application material [settled: ARCH-005 / Q-037]
│   └── Audience, exceptions, ownership of handbook changes [open]
├── 2. Principles [open]
│   ├── Establish authority or prevent protected execution [settled: PRINCIPLE-001]
│   ├── Auth-first design [settled: PRINCIPLE-002]
│   ├── Implicit enclosing tenant boundary [settled: TENANT-001]
│   ├── Resolution cannot amplify authority [settled: RESOLUTION-001]
│   └── Least privilege, trusted input rules, exceptions [open]
├── 3. Core vocabulary [open]
│   ├── Group and team are synonyms [settled: TERM-001]
│   ├── Groups/memberships in Auth; optional app sync [working: GROUP-001-A]
│   ├── Human-only group membership [working: GROUP-003]
│   ├── Principal, actor, subject, user identity, membership [open]
│   ├── Resource type versus instance; target; context [open]
│   └── Domain fact as explanatory vocabulary [working: TERM-002]
├── 4. Permissions and operations [open]
│   ├── Capability separate from scope [settled: PERMISSION-001]
│   ├── Naming grammar, catalog ownership, registration [open]
│   │   └── Registration principle now settled [REGISTRATION-001 / Q-039; mechanics open]
│   ├── Hierarchies, wildcards, aliases, evolution [open]
│   └── Mapping operations; multiple required permissions [open]
├── 5. Grants and lifecycle [active; returned from scope model after Q-042]
│   ├── Capability/scope/conditions stay bound [settled: GRANT-001]
│   ├── Multiple permissions under shared constraints [settled: GRANT-002]
│   ├── Direct and group-derived grants [settled: GRANT-003]
│   ├── Group-based grants preferred; direct grants supported [settled: GROUP-004]
│   ├── Reusable role versus group versus grant [settled: ROLE-001]
│   ├── Role changes and existing grants [settled: ROLE-002]
│   ├── Common expanded permission-set grant form [settled: RESOLUTION-003]
│   ├── Grant versus assignment [settled: TERM-004]
│   ├── Group-derived applicability [settled: RESOLUTION-004]
│   ├── Membership and role revision mechanics [open]
│   ├── Authority to administer grants [settled: ADMIN-001]
│   ├── Administration does not confer business access [settled: ADMIN-002]
│   ├── Explicit authorized, audited self-assignment [settled: ADMIN-003]
│   ├── Grantor bounds, recipients, scope validation [return point: ADMIN-004 / Q-023]
│   ├── Same grant model for administrative authority [working: ADMIN-005 / Q-024]
│   │   └── Ordinary model with grant-as-target [active: Q-043]
│   ├── Status, validity, provenance, membership changes [open]
│   └── Human-dependent service/agent authority
│       ├── All services/agents depend on human authority [settled: AUTHORITY-002]
│       ├── Loss of required upstream authority removes derived access [settled: DELEGATION-002]
│       ├── Who may delegate what; identity attribution [working: DELEGATION-001]
│       └── Delegation encoding, ceilings, expiry, growth, chains, reactivation [open]
├── 6. Scope and target [formerly active; remaining detail parked after Q-042]
│   ├── Explicit selections and attribute/relationship selectors [settled: SCOPE-001]
│   ├── Earlier typed format; definition-governance questions [partly deprecated: SCOPE-002 / Q-026]
│   ├── Scope ownership; shared and app-defined meanings [settled: SCOPE-003 / Q-027]
│   ├── Explicit resource compatibility [formerly working: SCOPE-004 / Q-028; meaning agreed Q-038, mechanics open]
│   ├── Selector/query enforcement and fixed endpoint modes [deprecated: SCOPE-005 / Q-029]
│   ├── Canonical boundary-selector definition [settled: SCOPE-006]
│   ├── Canonical key-value representation; empty/missing scope [settled: SCOPE-007]
│   ├── Scope keys: meaning, ownership, accepted references [active; no next question issued]
│   │   ├── Explicit key meanings and target mappings [formerly active; settled: Q-038; refines SCOPE-004]
│   │   ├── Shared definition contract for validation/evaluation [formerly active: Q-039; registration refinement agreed]
│   │   ├── Permission-scope compatibility declaration [formerly active: Q-040; optional feature agreed REGISTRATION-002]
│   │   ├── Omitted optional relationship metadata [formerly active: Q-041; superseded by explicit choice]
│   │   ├── Upfront application choice; all-grants validation when enabled [settled: REGISTRATION-003 / Q-041]
│   │   └── Enabling validation with incompatible existing grants [formerly active; settled: REGISTRATION-004 / Q-042]
│   ├── AND within scope; alternatives through complete grants [settled: SCOPE-008]
│   ├── Self resolves per human even in group-derived grants [settled: SELF-001]
│   ├── Scope type, descriptor, referenced resource, resolved scope [open]
│   ├── Exact/subtree and application boundary meanings [open]
│   └── Targets for read/list/create/update/move; relationship timing [open]
├── 7. Requests and resolution [open]
│   ├── Two endpoint-declared modes [deprecated: CONTRACT-002]
│   ├── One declaration per HTTP method/route [retained: CONTRACT-004; mode part deprecated]
│   ├── Earlier split declarations and mode validation [deprecated: CONTRACT-001, CONTRACT-003]
│   ├── One endpoint-owned gate and shared evaluator [settled: CONTRACT-006]
│   ├── Method/route action mapping and identified path/body inputs [settled: INPUT-001]
│   ├── Explicit required material/source declaration [settled: CONTRACT-007 / Q-035]
│   ├── Logical system block diagram [available: system-overview.md]
│   ├── Requested IDs versus established resource facts [working: FACT-001]
│   ├── Raw request → authorization request → resolved request [open; prepared deprecated]
│   ├── Stored grants → applicable grants → expanded/resolved grants [open]
│   └── Typed inputs/outputs, fact provenance, failures, context binding [open]
├── 8. Decision semantics [open]
│   ├── Grant binding and subset invariants [settled: GRANT-001, RESOLUTION-001]
│   ├── Same-permission positive grant alternatives [settled: DECISION-001]
│   ├── Positive-only grants; explicit deny grants excluded from v1 [settled: DECISION-002]
│   ├── Other union/intersection rules [open]
│   ├── Conditions, dependencies, conflicting and missing information [open]
│   └── Decision reasons, contributing grants, audit explanation [open]
├── 9. Enforcement and change over time [open]
│   ├── Earlier prepared/middleware-allow enforcement [deprecated: ENFORCEMENT-001]
│   ├── Safe fact gathering and actual-use enforcement [settled: ENFORCEMENT-002]
│   ├── Earlier completion-mode selection [deprecated: CONTRACT-005; distinction retained in CONTRACT-006]
│   ├── Query/row/field restrictions, bulk/partial results, mutations [open]
│   ├── Membership synchronization guarantees [open: SYNC-001]
│   ├── Revocation freshness, caches, concurrent change, resource moves [open]
│   └── Audit storage, correlation, versions, information disclosure [open]
├── 10. Challenge the model [open]
│   ├── Git hosting [available: UC-GIT-001 through UC-GIT-004]
│   ├── Ticketing [available: UC-TICKET-001 through UC-TICKET-004]
│   ├── HRMS [available: UC-HRMS-001 through UC-HRMS-004]
│   ├── Accounting [available: UC-ACCOUNT-001 through UC-ACCOUNT-004]
│   ├── Current grant layouts [available: grant-format.md, GRANT-EX-007]
│   ├── Earlier grant layouts [deprecated syntax: GRANT-EX-001 through GRANT-EX-006]
│   ├── Beyond-self completion cases [available: EC-001 through EC-007]
│   ├── Complete HRMS and repository scenarios with expected outcomes [open]
│   ├── Adversarial/missing/stale/conflicting input cases [open]
│   └── Verify implementation and external references; gap register [open]
└── 11. Publish the foundation [open]
    ├── Canonical glossary and consistent contract examples [open]
    ├── Requirements versus guidance versus implementation evidence [open]
    └── Resolve/defer all branches, version v1, implementation roadmap [open]
```

## Conclusions of detours and return points

| Detour | Conclusion retained | Still to revisit | Return branch |
|---|---|---|---|
| Teams/groups | Team means group; many grants can apply through membership. | Consolidate ownership/human-only proposals; synchronization and nesting. | 3 and 5 |
| Services and agents | All authority is human-dependent and remains a subset; no independent-service path. | Delegation encoding, grantor powers, ceilings, change propagation. | 5 and 9 |
| Middleware versus endpoint | CONTRACT-006: one endpoint-owned gate. Earlier two-mode and prepared design is deprecated, preserved in authorization-flow.md. | Concrete declaration/input/result schemas, fact provenance, enforcement interfaces. | 7 |
| Cases beyond self | Ownership, containment, relationships, and resource conditions can require application facts. | Test actual operations; do not assume coverage percentages or adopt every illustrative policy. | 9 and 10 |
| Grant examples | Tenant is enclosing context; each capability stays attached to its scope and conditions. | Final schema, roles, resolved forms, combination semantics. | 5, 7 and 8 |
| Positive-grant combination | Valid complete grants are alternative routes; explicit deny grants are outside v1. | Other constraints and multi-operation combination rules. | Returned to 5 for administration, then 8 |

## Historical branches

Question return index: Q-001 → groups/sync; Q-002 → team/group vocabulary;
Q-003 → human group membership; Q-004 → delegation dependency;
Q-005 and Q-006 → rejected independent-service branch;
Q-007 → retained subset authority; Q-008 → permission versus scope;
Q-009 → scope selectors; Q-010 and Q-011 → endpoint-mode contract;
Q-012 → grant binding; Q-013 → role definition;
Q-014 → role changes; Q-015 → role expansion;
Q-016 → grant/assignment terminology;
Q-017 → group-derived applicability;
Q-018 → positive-grant combination;
Q-019 → positive-only v1 decision;
Q-020 → grant-administration authority;
Q-021 → group-based access preferred, not exclusive (returned to administration bounds).
Q-022 → separate administration and explicit audited self-assignment (ADMIN-002/003 agreed).
Q-023 → whole-grant bounds; user challenges separate format (ADMIN-004 still proposed).
Q-024 → new scope fields challenged; syntax withdrawn (PROCESS-005; ADMIN-005 open).
Q-025 → approved stage 6 scope-model detour; return to ADMIN-004/005.
Q-026 → scope-type/ownership challenge; SCOPE-002 not yet agreed.
Q-027 → scope-owned target selection and grant binding (SCOPE-003 agreed).
Q-028 → explicit scope/resource compatibility (SCOPE-004 proposed).
Q-029 → route/JSON two-mode query walkthrough (SCOPE-005 deprecated, never agreed).
Q-030 → scope as boundary selector (SCOPE-006 agreed).
Q-031 → canonical scope definition and proposed key-value shape (SCOPE-006/007).
Q-032 → AND within scope, alternatives through grants (SCOPE-008 agreed).
Q-033 → endpoint-owned gate and recording resumed (CONTRACT-006, PROCESS-006).
Q-034 → canonical v1 scope and explicit empty-object semantics (SCOPE-007 agreed).
Q-035 → endpoint permission/material-source declaration (CONTRACT-007 agreed; sidebar concluded).
Q-036 → canonical Layer 1/application Layer 2 responsibilities (ARCH-004 agreed; runtime integration clarification, then return to scope-key meanings).
Q-037 → embedded auth agent across both sources (ARCH-005 agreed; sidebar concluded).
Q-038 → scope-key meanings and explicit supported target relationships (open; return to stage 6).
Update: Q-038 is now agreed and answers Q-028's conceptual compatibility question; detailed validation mechanisms remain open.
Q-039 → application-owned definition contract shared across grant validation and request evaluation (open).
Update: Q-039 is now approved as refined in REGISTRATION-001: applications register permission/scope contracts; Auth validates without domain interpretation.
Q-040 → explicit permission-scope compatibility declarations (open).
Update: Q-040 is approved with optionality qualification; REGISTRATION-002 binds optional declarations through registration.
Q-041 → behavior when optional relationship metadata is absent (proposed, open).
Update: Q-041 approved as refined in REGISTRATION-003: explicit upfront application-level choice, mandatory checks for every grant when enabled; omission-based mode inference superseded.
Q-042 → enabling relationship validation when existing grants are incompatible (open).
Update: Q-042 agreed as REGISTRATION-004; further registration lifecycle detail parked by user direction. Return to stage 5.
Q-043 → canonical grant administration with grant-as-target (open; ADMIN-005 before ADMIN-004 bounds).

The log retains these for traceability; they are not current options:

- GROUP-002: mixed human/service group membership — not adopted.
- SERVICE-001, SERVICE-002, AUTHORITY-001, AGENT-001, TERM-003:
  earlier independent-service distinction or narrower agent-only proposals —
  superseded/not adopted under AUTHORITY-002.
- RESOLUTION-002 and ARCH-003: universal dynamic three-way middleware outcomes
  or universal handler preparation — superseded by CONTRACT-002.
- ARCH-002/Q-010: intermediate discussion, now deprecated. CONTRACT-002 first
  settled two modes; CONTRACT-006 now deprecates that design in favor of one
  endpoint-owned gate. Required application facts remain endpoint-owned.

Scope selector syntax, example JSON, and application-specific policies remain
illustrative until their respective branches close.
