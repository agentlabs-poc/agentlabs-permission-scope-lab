# Authorization handbook — discussion tree

## Current position — reconciled through Q-050-F

The user has asked to move horizontally to high-impact decisions. This is the
current map; the older traversal snapshots are collapsed below as history.
Read the [handbook](handbook.md) for chapters and the
[decision log](handbook-roadmap.md) for individual references and rationale.

**Commit/push gate: reopened by the user.** The verified reconciliation checkpoint
is authorized for commit/push; continue the handbook discussion. Previous freeze
notices below are historical.

```text
Authorization Handbook — working edition, not published v1
├── 1. Purpose and architecture [core agreed; governance open]
│   ├── Canonical Layer 1 + application Layer 2 [ARCH-004 / Q-036]
│   ├── Embedded auth agent works across both [ARCH-005 / Q-037]
│   └── Audience, exceptions, change ownership [OPEN]
├── 2. Principles [core agreed; final consolidation open]
│   ├── Auth-first endpoints; trusted implicit tenant boundary
│   ├── Resolution and delegation cannot amplify authority
│   └── Necessary authority/enforcement failure prevents protected execution
├── 3. Vocabulary and identity [partly complete]
│   ├── Permission + scope boundary + request material [TERM-005 / Q-043]
│   ├── Team = group; Auth-owned human membership [Q-045]
│   └── Complete identity/actor/provenance glossary; sync mechanics [OPEN]
├── 4. Permissions [partly complete]
│   ├── Operation distinct from reach [PERMISSION-001]
│   ├── Exactly one required permission per endpoint [Q-049]
│   ├── Namespaced noun :: verb with variable depth [Q-056]
│   └── Detailed name validation/evolution, hierarchy and wildcards [OPEN]
├── 5. Grants and authority [core rules agreed; mechanics open]
│   ├── Whole permission/scope/condition binding; many grants per human
│   ├── Human group access preferred, direct access supported
│   ├── Live roles; resolved views never become independent assignments
│   ├── Services/agents always human-dependent subsets
│   ├── Admin authority separate from use; explicit audited self-assignment
│   ├── Ordinary grants survive issuer's later issuance-right loss [Q-046]
│   └── Admin scope encoding, containment, lifecycle and delegation details [OPEN]
├── 6. Scope and registration [core format agreed; mechanics parked]
│   ├── Required flat key-value boundary selector [SCOPE-007]
│   ├── AND within scope; separate grants for alternatives [SCOPE-008]
│   ├── {} is tenant-wide; missing/null invalid; self is per human
│   ├── App registers meanings/permissions; Auth validates [Q-039]
│   ├── Optional support validation explicitly enabled/disabled [Q-040/041]
│   ├── Reject activation with incompatible existing grants [Q-042]
│   └── Governance, version evolution, detailed compatibility mechanics [OPEN]
├── 7. Requests and resolution [core flow and partial policy agreed]
│   ├── One endpoint-owned gate; no prepared handoff [CONTRACT-006]
│   ├── Request vs resolved material; resolved does not mean allowed [Q-047]
│   ├── Memberships → direct/group grants → dependent views [Q-048]
│   ├── Policy: version, method, path, one permission, source/name inputs
│   │   ├── Top-level string version "1" [Q-050-A]
│   │   ├── GET/PUT partial structure [Q-050-B]
│   │   ├── No relationship block; endpoint enforces relationships [Q-050-C]
│   │   ├── Every declared input required at exact source [Q-050-E]
│   │   └── Application owns value validation; no duplicate schema [Q-050-F]
│   └── Full policy validation/publication and resolved-data schemas [OPEN]
├── 8. Decision semantics [HIGH-IMPACT OPEN BRANCH]
│   ├── Alternative complete positive grants; no explicit deny grants in v1
│   └── Decision-result contract, reasons, missing information, conditions [OPEN]
├── 9. Enforcement and time [HIGH-IMPACT OPEN BRANCH]
│   ├── Mandatory constraint of actual output/effects [Q-050-C]
│   ├── Review real constraints, not mere input usage [Q-050-D]
│   ├── Update/move, create, list/count/export/bulk contracts [OPEN]
│   └── Freshness, revocation, concurrency, audit and dependency change [OPEN]
├── 10. Challenge and verify [examples available; conformance open]
│   ├── 16 scenario groups: Git / ticketing / HRMS / accounting
│   ├── Request-flow SVG [SVG-001]; all simulations archived, handbook homepage
│   └── Complete adversarial scenario suite and runtime verification [OPEN]
└── 11. Publish the foundation [OPEN]
    ├── Final glossary, schemas, requirements/guidance distinction
    └── Close or explicitly defer gaps; implementation roadmap and governance
```

## Next discussion selection

Q-051 / DECISION-003 is **agreed**, including the Auth-service timeout example:
completed decisions are allow/deny; evaluation errors are separate and supply
no authorization. The [decision-results chapter](decision-results.md) retains
the original proposed status as history and records the rationale.

Q-052 / DECISION-004 is **agreed**: every completed deny must include an internal
machine-readable reason. Rationale, alternatives, examples, and consequences are
recorded in the decision-results chapter. Continue one question at a time;
the result schema, reason catalogue, and public error representation remain open.

Q-053 / DECISION-005 is **agreed as refined**: evaluator-provided `error_message`
and `error_message_reason` both reach the UI; the UI controls presentation.
Q-053-A's server-only second-message proposal is not adopted.
**Q-054 / DECISION-006 is agreed:** use the same message fields for evaluation
errors while keeping them distinct from completed denials.
**Q-055 / DECISION-007 is agreed:** add error_code for a stable machine-readable
cause alongside the readable messages. See the
[decision-results chapter](decision-results.md) for rationale and alternatives.

Permission-retention sidebar: the user requested the earlier detailed explanation
remain active; it is restored in [Permission](permission-model.md) and the reader.
**Q-056 / PERMISSION-002 is agreed with depth:** variable-length namespace before
the `::` verb separator; the original deeper examples are restored.
**Q-057 / PERMISSION-003 is agreed:** no automatic parent-to-child permission
inheritance. **Q-058 / PERMISSION-004 is agreed:** wildcard permission names are
outside v1. **Q-059 / PERMISSION-005 is proposed:** exclude permission aliases
from v1. Catalog evolution stays open. Then return
to decision-result contracts. Neither retention nor these narrow agreements
alone closes the audit's broader contract checkpoints.

Decision results, cross-boundary mutation rules, and freshness/dependency behavior
are high-impact candidates. Their precise order and answers are **not adopted**
by this reconciliation. Do not reopen the settled endpoint-policy shape or
return to the parked registration-detail rabbit hole without a reason.

Earlier reconciliation note, now qualified: a completion percentage had no
explicit measurement baseline. [MEASURE-001](handbook-completion-audit.md) now
counts 34 completed and 35 open checkpoints out of 69 (49.3% closure). This is a
proposed checkpoint rubric, not an effort estimate or release-readiness score.
Q-051's narrow agreement does not finish HC-08-02 or change that count.

## Return points and non-reopened decisions

| Branch/detour | Settled result | Still open |
|---|---|---|
| Scope format | Flat AND boundaries, explicit empty scope, human-relative self. | Governance and application-specific semantics beyond agreed examples. |
| Registration | Registered permission/scope validation; optional all-grant support checks. | Representation/evolution mechanics, parked by user direction. |
| Administration | Separate grant/use authority, bounded issuance, ordinary grant lifecycle. | Exact canonical admin scope encoding and containment validation. |
| Endpoint policy | One permission, explicit required sources, application-owned values, no relationship block. | Remaining structural validation, nested input syntax, publication. |
| Resolution | Dependent views with retained complete bindings and membership routes. | Transport/result contracts and freshness guarantees. |
| Enforcement | Actual output/effects must be constrained; input usage is insufficient. | Operation-specific and concurrent-change semantics. |

The [system diagram](system-overview.md) explains component cooperation; this
tree explains discussion coverage. Neither claims the historical runtime
implements the current design. PROCESS-003 still requires detailed rationale
in chapters, not only this map.

## Preserved traversal history

The full previous tree is also available as an
[unchanged archive](history/reconciliation-2026-09-05/docs/discussion-tree.md.txt).
Earlier active/open/next labels below describe their checkpoint only.
The current map above and later decisions govern present status.

<details>
<summary>Deprecated checkpoint navigation and original reference index</summary>

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

- **Current: Q-050-F settled INPUT-003.** Application-owned value validation;
  no type/nullability fields added to endpoint policy. The shape was already
  approved and is not reopened. Next branch: decision-result contracts. Remaining
  structural policy validation/publication details and update/move semantics stay
  open. Earlier value-validation-ownership-open labels below are history.

- **Current: Q-050-E settled INPUT-002.** Every declared input is required at
  its source; no silent omission/default/fallback, even with `{}` grants.
  The policy chapter records rationale, optional-input alternative, and PUT cases.
  Next: input value-validation responsibilities and remaining policy validation.
  Full publication and update/move contracts remain open. Earlier missing-input
  labels are history for presence/source behavior, not all validation details.

- **Current: Q-050-D settled ENFORCEMENT-003.** Review actual boundary/request
  enforcement on output and mutations, not merely input usage. The policy chapter
  records examples and why this does not guarantee detection of every breach.
  Next: remaining Q-050 input/policy validation before publication. Update/move
  and decision-result contracts remain open. Earlier proposed/open labels are history.

- **Current: Q-050-C settled CONTRACT-012 as revised.** No relationship block;
  endpoint implementation must keep output/effects within authorized boundaries.
  The policy chapter records rationale, grant/query example, and rejected resolver
  proposal. **Q-050-D is open:** review actual constraint enforcement, not merely
  input usage; no guarantee against every breach. Remaining policy validation,
  publication, and update/move contracts are still open. Earlier labels are history.

- **Current: Q-050-B settled CONTRACT-011's partial policy structure.** Version,
  method/path, one permission, and named source/name inputs are agreed. The
  endpoint-policy-format chapter contains GET and PUT examples and their rationale.
  **Next: Q-050-C relationship bindings.** Full policy validation/publication and
  update/move semantics remain open. Earlier not-discussed notes are history.

- **Current: Q-050-A settled CONTRACT-010.** Required top-level string
  `version: "1"`; reject missing/malformed/unsupported versions. Contract
  versions are distinct from document revisions and interpreted per contract type.
  **Q-050 remains open:** endpoint policy structure and binding syntax are next.
  Earlier version-syntax-open statements below are retained history.

- **Current: Q-049 settled CONTRACT-008:** exactly one required permission per
  protected endpoint, mandated by Auth. Multi-permission endpoint declarations
  are not a remaining v1 option; multi-permission grants remain supported.
  **CONTRACT-009:** every published JSON/YAML contract includes a version.
  **Next: Q-050, endpoint policy contract**, explicitly not yet discussed.
  Requirements are agreed, but version syntax and the policy schema are not.
  Return to decision results/enforcement afterwards. Earlier positions are history.

- **Current: Q-048 settled RESOLUTION-006.** Obtain the human's valid memberships,
  then direct and membership-group grants; resolved views retain each source and
  dependency. The grant chapter captures rationale, the Finance/self JSON example,
  counterexamples, and open retrieval/schema/freshness mechanics. Next branch:
  stage 8 decision semantics, including multiple required permissions. Earlier
  checkpoint labels below are history, not additional live questions.

- **Current: Q-047 / Q-047-A settled RESOLUTION-005 and clarified CONTRACT-007.**
  The endpoint predeclares permissions, inputs, sources, and how to establish
  any required relationship. Scope and other mandatory checks determine which
  trusted material is needed; resolved is not allowed. Rationale and `{}` cases
  are in the endpoint chapter. **Next: Q-048, resolved grants**, still in stage 7.
  No new wire schema or multi-permission combination policy is adopted.
  Earlier current-position labels below are preserved checkpoint history.

- **Current: Q-046 settled ADMIN-006.** Ordinary human/group grants do not
  depend merely on their issuing administrator retaining issuance authority.
  Grant validity, membership dependencies, and human-dependent automation remain
  enforced. Rationale and revocation consequences are in the grant chapter.
  **Next: Q-047, request versus resolved request**, returning to stages 7/8.
  Administrative encoding and detailed lifecycle mechanics remain open.
  Earlier current-position labels below are preserved checkpoint history.

- **Current: Q-045 settled GROUP-001-A and GROUP-003.** Auth owns authorization
  groups/membership; application sync is optional; members are humans. The
  [group chapter](groups-and-membership.md) preserves rationale and consequences.
  Q-046 next addresses issuer-authority changes for ordinary human/group grants.
  Human-dependent automation remains unchanged; sync mechanics stay parked.

- **Current: Q-044 settled ADMIN-004/005's governing rules.** Ordinary grants
  express administrative authority; complete assignments must fit associated
  bounds and normal validation still applies. Encoding/containment stay open.
  Q-045 next consolidates earlier group ownership and human-only membership
  directions. Registration implementation detail stays parked.

- **Current: Q-043 settled TERM-005's vocabulary correction.** Use permission,
  scope boundaries, requests, and trusted request material; no extra entity.
  [Detailed rationale](authorization-vocabulary.md) and [explanation audit](explanation-audit.md)
  capture examples and remaining gaps. Stay in stage 5, ADMIN-004/005; their
  representation and assignable bounds remain open. Registration detail stays
  parked. Earlier Q-043-open positions below are historical.

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

- Current as of Q-038: application-owned scope meanings, supported resource
  relationships, and endpoint-specific trusted fact bindings are agreed.
  Q-028's conceptual question is answered; SCOPE-004's detailed validation
  mechanics remain open. Q-039 is now active: sharing definition contracts
  across grant validation and request evaluation. Earlier active-position notes
  below are retained history, not additional unanswered copies of Q-038.

- Latest: Q-037 / ARCH-005 is agreed, closing the embedded-agent sidebar below.
  Auth and application supply their respective material; shared evaluation and
  endpoint enforcement remain one gate. Return to stage 6: Q-038 now asks about
  explicitly defined scope-key meanings and supported resource relationships.
  Earlier "clarify" and "no next question" notes below are historical positions.

- Q-036 / ARCH-004 is agreed: canonical Layer 1 and application-specific
  Layer 2 jointly establish authorization within the single endpoint-owned gate.
  Clarify the embedded auth agent's integration across these sources, then return
  to scope-key meanings and ownership. See [system responsibilities](system-overview.md).

- Active branch: **6. Scope boundaries → scope-key definitions and governance**.
  Sidebar concluded: CONTRACT-007 / Q-035 is agreed. The endpoint declares its
  permission, material, and sources. [System block diagram](system-overview.md)
  includes the requested SVG; return to scope-key definitions now.
  SCOPE-006/008 and SCOPE-007's canonical v1 format are agreed; Q-034 is answered.
  Next: scope-key definitions/governance, before returning to ADMIN-004/005.
  SCOPE-004 remains open. SCOPE-002's earlier typed-format direction is historical;
  its definition-governance questions remain open under the canonical key-value model.
  Q-025 approved this detour; return to ADMIN-004/005. Q-024's fields stay withdrawn.
  The detailed chapter is [scope boundaries](scope-model.md).
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
│   │   └── Ownership principle now agreed [Q-045; earlier working label is history]
│   ├── Human-only group membership [working: GROUP-003]
│   │   └── Member-type principle now agreed [Q-045; mechanics open]
│   │   └── Consolidate group ownership/membership directions [active: Q-045]
│   │       └── Consolidation concluded [Q-045 agreed]
│   ├── Principal, actor, subject, user identity, membership [open]
│   ├── Resource type versus instance; context [open]
│   ├── Permission/boundary/request-material vocabulary [settled: TERM-005 / Q-043]
│   └── Domain fact as explanatory vocabulary [working: TERM-002]
├── 4. Permissions and operations [open]
│   ├── Capability separate from scope [settled: PERMISSION-001]
│   ├── Naming grammar, catalog ownership, registration [open]
│   │   └── Registration principle now settled [REGISTRATION-001 / Q-039; mechanics open]
│   ├── Hierarchies, wildcards, aliases, evolution [open]
│   └── Mapping operations; multiple required permissions [open]
│       └── Exactly one required permission per endpoint [settled: CONTRACT-008 / Q-049; prior multiple-permission open label is history]
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
│   │   └── Five administration rules [settled: Q-044; exact encoding open]
│   ├── Same grant model for administrative authority [working: ADMIN-005 / Q-024]
│   │   └── Ordinary model now agreed [Q-044; prior working label is history]
│   │   └── Q-043 framing withdrawn; vocabulary settled, admin model still open
│   ├── Status, validity, provenance, membership changes [open]
│   │   └── Ordinary grants after issuer-authority loss [active: Q-046]
│   │       └── Now settled: ADMIN-006 / Q-046; prior active label is history
│   └── Human-dependent service/agent authority
│       ├── All services/agents depend on human authority [settled: AUTHORITY-002]
│       ├── Loss of required upstream authority removes derived access [settled: DELEGATION-002]
│       ├── Who may delegate what; identity attribution [working: DELEGATION-001]
│       └── Delegation encoding, ceilings, expiry, growth, chains, reactivation [open]
├── 6. Scope boundaries [formerly active; remaining detail parked after Q-042]
│   ├── Explicit selections and attribute/relationship selectors [settled: SCOPE-001]
│   ├── Earlier typed format; definition-governance questions [partly deprecated: SCOPE-002 / Q-026]
│   ├── Scope ownership; shared and app-defined meanings [settled: SCOPE-003 / Q-027]
│   ├── Explicit resource compatibility [formerly working: SCOPE-004 / Q-028; meaning agreed Q-038, mechanics open]
│   ├── Selector/query enforcement and fixed endpoint modes [deprecated: SCOPE-005 / Q-029]
│   ├── Canonical boundary-selector definition [settled: SCOPE-006]
│   ├── Canonical key-value representation; empty/missing scope [settled: SCOPE-007]
│   ├── Scope keys: meaning, ownership, accepted references [active; no next question issued]
│   │   ├── Explicit key meanings and fact bindings [formerly active; settled: Q-038; refines SCOPE-004]
│   │   ├── Shared definition contract for validation/evaluation [formerly active: Q-039; registration refinement agreed]
│   │   ├── Permission-scope compatibility declaration [formerly active: Q-040; optional feature agreed REGISTRATION-002]
│   │   ├── Omitted optional relationship metadata [formerly active: Q-041; superseded by explicit choice]
│   │   ├── Upfront application choice; all-grants validation when enabled [settled: REGISTRATION-003 / Q-041]
│   │   └── Enabling validation with incompatible existing grants [formerly active; settled: REGISTRATION-004 / Q-042]
│   ├── AND within scope; alternatives through complete grants [settled: SCOPE-008]
│   ├── Self resolves per human even in group-derived grants [settled: SELF-001]
│   ├── Scope type, descriptor, referenced resource, resolved scope [open]
│   ├── Exact/subtree and application boundary meanings [open]
│   └── Request material for read/list/create/update/move; relationship timing [open]
├── 7. Requests and resolution [open]
│   ├── Two endpoint-declared modes [deprecated: CONTRACT-002]
│   ├── One declaration per HTTP method/route [retained: CONTRACT-004; mode part deprecated]
│   ├── Earlier split declarations and mode validation [deprecated: CONTRACT-001, CONTRACT-003]
│   ├── One endpoint-owned gate and shared evaluator [settled: CONTRACT-006]
│   ├── Method/route action mapping and identified path/body inputs [settled: INPUT-001]
│   ├── Explicit required material/source declaration [settled: CONTRACT-007 / Q-035]
│   │   └── Endpoint policy JSON/YAML contract [active: Q-050; schema not yet discussed]
│   │       └── Shared version convention [settled: CONTRACT-010 / Q-050-A; full policy schema remains open]
│   │       ├── Method/path, one permission, source/name inputs [settled partial structure: CONTRACT-011 / Q-050-B; GET and PUT examples]
│   │       ├── Required declared inputs; no default/source fallback [settled: INPUT-002 / Q-050-E]
│   │       ├── Application-owned value validation; no duplicate policy schema [settled: INPUT-003 / Q-050-F]
│   │       └── Relationship bindings and trusted fact connections [active: Q-050-C]
│   │           └── Revised and settled: CONTRACT-012 / Q-050-C, no relationship block; endpoint enforcement mandatory
│   │   └── Permissions, inputs, sources, relationship bindings [clarified: Q-047-A]
│   ├── Logical system block diagram [available: system-overview.md]
│   ├── Requested IDs versus established resource facts [working: FACT-001]
│   ├── Raw request → authorization request → resolved request [open; prepared deprecated]
│   │   └── Meaning of request versus trusted evaluation view [active: Q-047]
│   │       └── Now settled: RESOLUTION-005 / Q-047 / Q-047-A; prior active label is history
│   ├── Stored grants → applicable grants → expanded/resolved grants [open]
│   │   └── Resolved-grant meaning and retained dependencies [active: Q-048]
│   │       └── Now settled: RESOLUTION-006 / Q-048, including membership-based retrieval; prior active label is history
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
│   ├── Review actual output/mutation constraints, not mere input usage [settled: ENFORCEMENT-003 / Q-050-D]
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
Q-027 → scope-owned boundary selection and grant binding (SCOPE-003 agreed).
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
Q-038 → scope-key meanings and explicit supported resource relationships (open; return to stage 6).
Update: Q-038 is now agreed and answers Q-028's conceptual compatibility question; detailed validation mechanisms remain open.
Q-039 → application-owned definition contract shared across grant validation and request evaluation (open).
Update: Q-039 is now approved as refined in REGISTRATION-001: applications register permission/scope contracts; Auth validates without domain interpretation.
Q-040 → explicit permission-scope compatibility declarations (open).
Update: Q-040 is approved with optionality qualification; REGISTRATION-002 binds optional declarations through registration.
Q-041 → behavior when optional relationship metadata is absent (proposed, open).
Update: Q-041 approved as refined in REGISTRATION-003: explicit upfront application-level choice, mandatory checks for every grant when enabled; omission-based mode inference superseded.
Q-042 → enabling relationship validation when existing grants are incompatible (open).
Update: Q-042 agreed as REGISTRATION-004; further registration lifecycle detail parked by user direction. Return to stage 5.
Q-043 → canonical grant administration with grant-administration request material (open; ADMIN-005 before ADMIN-004 bounds).
Update: Q-043 was reformulated and settled TERM-005 after a tenability check. Original framing is withdrawn, not an approved administration model; its exact wording is preserved in history/q043-vocabulary.md.
Q-044 → ADMIN-004/005 governing rules approved; concrete administrative scope contracts remain open.
Q-045 → group ownership and human-only membership consolidation (open).
Update: Q-045 approved GROUP-001-A and GROUP-003; rationale is captured in groups-and-membership.md.
Q-046 → ordinary human/group grant lifecycle after issuing-administrator authority changes (open).
Update: Q-046 approved ADMIN-006; prior open label is history. Rationale and explicit-revocation consequences are in grant-model.md.
Q-047 → return to requests/resolution: distinguish request inputs from trusted evaluation material without a prepared handoff or new authority (open).
Update: Q-047 / Q-047-A approved RESOLUTION-005 and clarified CONTRACT-007; prior open label is history. The endpoint chapter records rationale and selective material examples.
Q-048 → resolved grant as a dependent evaluation view, not a new assignment (open).
Update: Q-048 approved RESOLUTION-006, including Vinay's membership/direct/group retrieval flow. Prior open label is history; next is decision semantics, with detailed schemas still open.
Q-049 → CONTRACT-008 approved: exactly one required permission per protected endpoint. Earlier AND-across-permissions proposal not adopted.
Publication sidebar → CONTRACT-009 requires a version in every published JSON/YAML contract; syntax and compatibility remain open.
Q-050 → endpoint policy contract discussion, including version representation (open).
Q-050-A → CONTRACT-010 approved: version field/string, type-local meaning, revision distinction, and strict rejection rules. Q-050's remaining policy schema stays open.
Q-050-B → CONTRACT-011 approved partial structure; requested PUT body example added with rationale and current/proposed distinction.
Q-050-C → relationship bindings and trusted facts (open); full Q-050 policy schema remains unfinished.
Update: Q-050-C approved CONTRACT-012 instead: endpoint-enforced relationships without a policy block. Prior open label/proposal is history; other full-policy gaps remain.
Q-050-D → review criterion for actual constraint enforcement on output/mutations, not mere input usage (open).
Update: Q-050-D approved ENFORCEMENT-003, including the limits of this review. Prior open label is history; next is remaining input/policy validation.
Q-050-E → INPUT-002 approved: required input presence at declared source, independent of scope breadth; value validation remains open.
Q-050-F → INPUT-003 approved: application owns input type/nullability/domain validation, policy shape unchanged. Prior ownership-open labels are history; return to decision-result contracts.

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

</details>
