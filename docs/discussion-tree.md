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

## Current position

- Active branch: **5. Grants → administrative bounds** (ADMIN-002/003 agreed;
  Q-022 answered; exact bounds next).
- Just concluded: GROUP-004 / Q-021 — membership-based access is preferred;
  direct human grants remain supported. Documentation detail reviewed and
  indexed in handbook.md, including the new authorization-flow.md chapter.
- Returned from the group/self sidebar to bounds on grant administration and
  remaining lifecycle questions in stage 5;
  then declared request, prepared context, and resolved
  request/grants. Return to the remaining siblings in the tree before v1.
- A parent branch stays open until all its required children are settled or
  explicitly excluded from v1. Settled examples do not close the whole branch.

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
│   ├── Hierarchies, wildcards, aliases, evolution [open]
│   └── Mapping operations; multiple required permissions [open]
├── 5. Grants and lifecycle [active]
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
│   ├── Grantor bounds, recipients, scope validation [active: next question pending]
│   ├── Status, validity, provenance, membership changes [open]
│   └── Human-dependent service/agent authority
│       ├── All services/agents depend on human authority [settled: AUTHORITY-002]
│       ├── Loss of required upstream authority removes derived access [settled: DELEGATION-002]
│       ├── Who may delegate what; identity attribution [working: DELEGATION-001]
│       └── Delegation encoding, ceilings, expiry, growth, chains, reactivation [open]
├── 6. Scope and target [open]
│   ├── Explicit selections and attribute/relationship selectors [settled: SCOPE-001]
│   ├── Self resolves per human even in group-derived grants [settled: SELF-001]
│   ├── Scope type, descriptor, referenced resource, resolved scope [open]
│   ├── Exact/subtree, multi-dimensional scope, empty/missing scope [open]
│   └── Targets for read/list/create/update/move; relationship timing [open]
├── 7. Requests and resolution [open]
│   ├── Two endpoint-declared modes [settled: CONTRACT-002]
│   ├── One declaration per HTTP method/route [settled: CONTRACT-004]
│   ├── Declaration details and mode validation [working: CONTRACT-001, CONTRACT-003]
│   ├── Requested IDs versus established resource facts [working: FACT-001]
│   ├── Raw request → authorization request → prepared/resolved request [open]
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
│   ├── Prepared cannot authorize protected output [settled: ENFORCEMENT-001]
│   ├── Additional resolution versus applying restrictions [settled: CONTRACT-005]
│   ├── Query/row/field restrictions, bulk/partial results, mutations [open]
│   ├── Membership synchronization guarantees [open: SYNC-001]
│   ├── Revocation freshness, caches, concurrent change, resource moves [open]
│   └── Audit storage, correlation, versions, information disclosure [open]
├── 10. Challenge the model [open]
│   ├── Grant illustrations [available: GRANT-EX-001 through GRANT-EX-006]
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
| Middleware versus endpoint | Each endpoint has one mode; middleware-complete uses allow/deny, endpoint-completion uses deny/prepared. | Complete declaration schema, input provenance, mode validation, examples. | 7 |
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

The log retains these for traceability; they are not current options:

- GROUP-002: mixed human/service group membership — not adopted.
- SERVICE-001, SERVICE-002, AUTHORITY-001, AGENT-001, TERM-003:
  earlier independent-service distinction or narrower agent-only proposals —
  superseded/not adopted under AUTHORITY-002.
- RESOLUTION-002 and ARCH-003: universal dynamic three-way middleware outcomes
  or universal handler preparation — superseded by CONTRACT-002.
- ARCH-002/Q-010: intermediate discussion; its unresolved mode question is now
  resolved by CONTRACT-002/CONTRACT-004. Preserve its no-app-DB middleware boundary.

Scope selector syntax, example JSON, and application-specific policies remain
illustrative until their respective branches close.
