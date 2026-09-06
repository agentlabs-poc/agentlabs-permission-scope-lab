# Authorization handbook — discussion tree

**Q-130 approved — D09 finite-batch composition resolved:** [different complete routes may cover different items](bulk-enforcement.md).
Full batch preflight and no cross-grant fragment mixing remain required. Move
source/destination composition and exact batch transports remain separate gaps.

**Q-129 approved — D08 already-allowed synchronous case resolved:** [the same bounded operation may finish](concurrent-enforcement.md).
Q-074 and Q-110 remain stronger obligations; queued/streamed/long-running and
not-yet-allowed cases remain distinct. Next impact-first branch: complete-route
coverage of multi-item operations (D09), without grant-fragment mixing.

**Q-128 approved — D08 reduction coverage resolved:** [all confirmed authority reductions govern new checks](authority-freshness.md).
No stale grace from membership, assignment/grant controls, delegation, or adopted
narrowing. Remaining: already-in-flight/allowed operations and exact confirmation,
ordering, and evidence contracts. No full consistency checkpoint is closed yet.

**Q-127 approved — D05 chain scope resolved:** [proxy-to-proxy delegation is not supported in v1](delegation-lifecycle.md).
Direct human-to-proxy delegation remains supported. Chain implementation is no
longer an open v1 requirement; exact direct-delegation lifecycle still needs closure.
Next impact-first branch: authority-change visibility (D08).

**Q-126 approved — D04 semantic reuse rule resolved:** [existing permission identifiers cannot be repurposed](permission-lifecycle.md).
Materially different operations require new names, including after retirement.
Scope-definition evolution and exact lifecycle/restoration contracts remain.
Continue impact-first coverage with delegation chains (D05).

**Q-125 approved — D04 retirement policy resolved:** [retirement may proceed despite grant references](permission-lifecycle.md).
Effective retirement removes the permission from root support without altering
stored grants/assignments. Identifier evolution, scope changes, and exact lifecycle
contracts remain; do not re-ask whether all references must be migrated first.

**Q-124 approved — D03 continuation rule resolved:** [same-intent, authorized, revalidated retry](bootstrap-initial-assignment.md).
No silent authority replacement or conflicting-attempt merge. Exact setup
evidence, concurrency coordination, and deliberate recovery remain contract gaps.
Continue horizontal coverage with registered-definition lifecycle (D04).

**Q-117 approved — D03 visibility rule resolved:** [bootstrap authority becomes usable only after complete validated establishment](bootstrap-initial-assignment.md).
Earlier parked/unapproved entries below are history. Partial/concurrent attempt
handling and governed recovery remain open; the entire bootstrap branch is not closed.

**Q-123 resolved by user correction:** [one shared catalog per application](root-permission-evolution.md),
not per-tenant release selection. This removes the proposed selective-upgrade
branch. D01/D04 still require source encoding, publication consistency, and
definition lifecycle work; do not count tenant version selection as pending.

**Q-122 approved — D02 mechanism selected:** [computed root permission coverage](root-permission-evolution.md),
not automatic root revision publication/adoption for catalog growth. Exact
encoding and applicable catalog/version binding remain open. The Q-121 next-step
notice below is historical; it no longer asks which mechanism to select.

**Q-121 approved — D01 partially resolved:** [application platform authority](application-platform-authority.md)
is responsible for capability publication; tenant business permission is not
required to define the capability. Existing authority is reused, not a new
publisher grant. D01's exact setup/binding contracts remain open; D02 is the
next root-growth mechanism discussion. No whole-checkpoint closure is claimed.

**Remaining agenda — [ASSESS-001](discussion-assessment.md):** 13 design topics
+ five contract reviews + final acceptance. Eight assistant-owned drafting,
scenario, reconciliation, and verification packages support that same agenda;
they are not eight additional discussions. Every currently open criterion is
mapped in the assessment. No new policy approval or checkpoint closure is implied.

## Current execution tree — closure, not new feature exploration

```text
V1 closure [CLOSURE-001: 38/68 criteria DONE; 30 OPEN]
├── Documentary consolidation
│   ├── Numbered principles + compliance examples [DONE: HC-02-05]
│   ├── Identity/authority glossary [DONE: HC-03-04]
│   └── Separate implementation roadmap [DONE: HC-11-04]
├── Authority contract package [OPEN]
│   ├── Source-support evidence and boundary validation
│   ├── Root/derived, direct/role, lifecycle record completion
│   └── Membership/hierarchy, delegation, identity/registration dependencies
├── Evaluation and enforcement contracts [OPEN]
│   ├── Request/resolved/result and handler-agent interfaces
│   └── Freshness, concurrent use, operation/background cases
├── Audience and change governance [OPEN: user choice required]
└── Final scenarios, reconciliation, package, acceptance [OPEN]
```

[All 33 prior open criteria and their disposition](v1-closure.md) ·
[grant contract inventory](grant-contract-closure.md) ·
[current measured progress](milestone-progress.md).
No approved rules are changed by this tree. Q-117 remains unapproved and parked;
the earlier active-proposal position below is historical for this closure pass.

**Q-118 agreed — bounded contract completion:** [role-based immutable revision](role-grant-contract.md)
and exclusive permission-source validation. Direct permissions exclude both role
fields; role-based content requires the pair. The [record reference](grant-record-reference.md)
reflects this approval; original proposed status remains in chapter history.
No additional full checkpoint is closed by settling this variant alone.

**Q-119 agreed:** [root grant-revision representation](root-grant-format.md)
omits the parent field, without a root flag or a changed initial authority ceiling.
Trusted establishment is still required; full evidence/procedure remains open.

**Active Q-120:** [new application permissions and root authority](root-permission-evolution.md).
**Q-120A direction accepted:** root coverage grows automatically for legitimate
application capability upgrades, without a separate manual root-expansion step.
Wildcard versus materialized-revision mechanism remains open; no `*` is approved.
Ordinary child selection/adoption and tenant/application boundaries remain.

## Current revision branch — decisions through Q-111

[Milestone remaining percentages](milestone-progress.md) use the fixed checklist;
recent partial advances are shown separately, without guessed fractional credit.

```text
Reusable-grant revisions
├── Per-assignment explicit adoption [Q-102 AGREED IN PRINCIPLE]
│   └── Advisory update list, never automatic adoption [Q-102A AGREED]
├── Resolve current adopted lineage top-to-bottom [Q-103 AGREED]
│   └── Not latest published elsewhere; child ceilings remain live
├── One current assignment per grant/recipient in tenant [Q-104 AGREED]
│   └── Disabled retained assignments count; no duplicate bypass
├── Latest published revision for creation and explicit upgrades [Q-104A / Q-105]
│   └── Existing assignments can stay; intermediate upgrade choice excluded
├── Published authority content immutable; enablement remains live [Q-106]
├── Core grant / revision / assignment JSON [Q-107 APPROVED]
├── Grant validity retained; assignment validity deferred in v1 [Q-108 APPROVED]
│   └── Grant window in immutable revision; explicit adoption [Q-109 APPROVED]
├── Preserve validated state through Auth write [Q-110 AGREED]
├── Hierarchy integrity: reject grant/team ancestor loops [Q-111 AGREED]
├── Direct assignment: parent defines maximum permissions/scope [Q-112 clarified]
│   └── Lineage-supported latest reaffirmed; extra revision field withdrawn [Q-112A]
└── OPEN: support-discovery/evidence contracts, complete schemas,
    timestamp/clock contracts, concurrency, and revision lifecycle mechanics
```

[Rules, rationale, and cases](grant-revisions.md) · [approved core JSON](grant-revision-format.md).
Recording and exploration are now lab-only; scratchpad sources are preserved
in [lab history](history/scratchpad-import/README.md), not used for new work.
The older open revision labels below are narrowed by these approvals, not full
schema completion or a newly measured percentage.

[Q-112/Q-112A clarification and remaining implementation contracts](direct-human-parent-context.md).
Current prioritization is deepest authority dependencies first, then safe changes,
resolution consistency, full contracts, and editorial publication—not easy scores.

**Q-113 agreed with clarification:** [legitimate root establishment](bootstrap-authority.md)
requires trusted setup; ordinary operations are already bounded. A cross-rule
review is recorded, not a new parent-omission policy or completed trust procedure.
**Q-114 agreed as corrected:** registration of relevant permission/scope contracts
precedes seed acceptance; trusted setup establishes maximum intended tenant
authority and explicit administrator assignment. Ordinary administration then
distributes equal/narrower authority through users, groups, and membership.
The FIN-read-only bootstrap framing is superseded. [Setup flow SVG](assets/bootstrap-registration-flow.svg).
Trust implementation, exact root/initial-recipient records, and recovery remain open.

**Q-115 agreed:** [initial user and administrators group](bootstrap-initial-assignment.md).
Trusted setup creates a selected legitimate human user and administrators group,
assigns root authority to the group, and explicitly adds the human as a member.
No special identity; full trust/user/membership contracts remain open.

**Q-116 agreed:** repeated completed bootstrap reports already initialized without
authority changes. Normal administration remains available; bootstrap cannot
reset grants, re-enable records, or restore/add administrator membership.

**Active Q-117, proposed:** no usable initial authority until the complete setup
arrangement is validated and durably established. Partial records may remain;
completion evidence, visibility, continuation, concurrency, and recovery remain open.

## Current position — Q-101 framing pinned through Q-101E-3

```text
Dependent authority: current binding branch
├── Grant lineage: parent_grant_id [Q-101 AGREED]
│   ├── Actual assignments + parent-team context remain required
│   └── No additional parent_assignment_id for the settled cases
├── Four-part binding: G1 / Team1 / G2 / TeamY
│   ├── Permission subset; effective scope AND; child-team ceiling
│   └── Repeats across descendants and affected shared assignments
├── Assignment eligibility ≠ grant enablement ≠ effective authority
│   ├── Disabled grant supplies no authority anywhere [Q-101A/B]
│   ├── Disabled assignment stops its route
│   └── Still-enabled descendants recover with valid support [Q-101C]
├── Missing required support: orphaned lineage ineffective [Q-094 / Q-101D]
│   └── Higher-layer warning/prevention is separate from resolution
├── Structural changes: bottom-up assignment handling [Q-101E-1/E-2]
│   ├── Relevant bindings removed OR disabled [Q-101E-3 CORRECTION]
│   └── Explicit enablement validates current relationships and boundaries
├── Covered: direct/shared routes, subgroup support, branches, parent changes
└── OPEN CONTRACTS: reusable-grant revisions, complete lifecycle/hierarchy
    schemas, detailed recovery/rebinding, concurrency/freshness and runtime tests
```

Read the [pinned chapter, diagrams, rationale, and case matrix](parent-grant-bindings.md).
The remaining handbook branches and historical scores below are preserved; this
does not close their deliverables or recalculate completion. Q-101 supersedes
older specific-assignment-ID and removal-only wording, not actual support checks.

## Earlier checkpoint — Q-100 makes Auth's internal authority gate explicit

```text
Post-0.0.1 canonical direction
├── Grant definition: permissions + scope; no recipient [Q-090 AGREED]
├── Assignment: grant reference + recipient + status [Q-090 AGREED]
├── Subteam / subgroup: explicit dependent authority [Q-091 AGREED]
│   ├── One supporting parent assignment per derived route
│   ├── Permissions subset; parent scope AND child constraints
│   └── Human memberships remain direct, not inherited
├── Team administration: create/write/delete; write includes membership [Q-092 AGREED]
│   └── Creating a subteam does not assign its authority; assignment is separate
├── Assignment: admin authority + assigner's valid supporting parent route [Q-093 AGREED]
├── Sub-: parent-dependent relationship, not new team/grant entity types [Q-095 AGREED]
│   └── Child-team total authority and each child grant remain within their parents
├── Missing required parent support: orphaned route cannot authorize [Q-094 AGREED]
├── Owner rotation does not rewrite team-held authority lineage [Q-099 AGREED]
│   ├── Acting administrator still needs current authorization [Q-093 refined]
│   ├── Selected team assignment remains continuing support
│   ├── Explicit personal dependencies remain; no automatic rebinding
│   └── Q-096: two-owner recommendation and ownership contracts remain OPEN
├── Auth Service's endpoint-owned authority-change gate [Q-100 AGREED architecture]
│   ├── Administrative evaluator: caller's operation + administrative scope
│   ├── Authority validator: complete proposal + supporting lineage + ceilings
│   ├── Both must pass before the validated change is persisted
│   └── Runtime dependency and application enforcement remain mandatory
└── OPEN: combined dependency/hierarchy format, definition revisions,
    assignment lifecycle, owner rules, and administrative boundaries
```

See [grant/assignment rationale and JSON](grant-assignments.md) and
[subgroup rationale and examples](subgroups.md). Recording these decisions is
authorized; this is not a commit/push instruction. The `0.0.1` tag preserves the
baseline. The prior 35/68 score below is a baseline checklist score, not a fresh
completion measurement for the expanded model; re-audit remains pending.

[Q-100's internal-flow SVG and rationale](auth-service-authority-gate.md) explain
the two Auth Service responsibilities without replacing the application diagram.
Validator API, operation coverage, and concurrency/source-binding contracts are
still open. This addition does not claim a complete security implementation.

[Q-092's initial team-admin grant](team-administration.md) is recorded with
its tenant-wide reach and membership-distribution consequences. [Q-093](assignment-authority.md)
records the supporting-source and separate administrative checks. [Q-095](authority-lineage.md)
clarifies team, grant/assignment, scope, and membership relationships. Exact link
fields remain open. Q-094's orphan definition is now agreed, with explicit rather
than automatic ownership transfer; full lifecycle remains open.
[Q-099](ownership-lineage.md) settles the distinction between team-held support
and the acting administrator, without importing Q-097/Q-098 from the scratchpad.
The two-owner recommendation and exact ownership contracts remain open.

<details>
<summary>Earlier Q-096 position — dependency distinction resolved by Q-099</summary>

Q-096 proposes
team-held authority and two owners, including a pending distinction between
team-source dependency and the original administrator's membership.

</details>

<details>
<summary>Previous position — Q-089-B and 0.0.1 baseline score, preserved</summary>

The user has asked to move horizontally to high-impact decisions. This is the
current map; the older traversal snapshots are collapsed below as history.
Read the [handbook](handbook.md) for chapters and the
[decision log](handbook-roadmap.md) for individual references and rationale.

**Commit/push gate: reopened by the user.** The verified reconciliation checkpoint
is authorized for commit/push; continue the handbook discussion. The user's
latest cadence is immediate local recording, with commit/push roughly every ten
questions, or on explicit request. After freezing recording during Q-089 discussion,
the user explicitly requested recording, commit, and push of Q-089-B. This reopens
recording for that decision and its reconciliation checkpoint, not unrelated future
discussion. The prior checkpoint was through Q-087-B;
Q-083 was the first approval since checkpoint `01d68a5`. Previous freeze
notices below are historical.

Counts are closed/total checkpoints from MEASURE-001, not effort estimates.

```text
Authorization Handbook — 35/68 in-scope closed (51.5%); 1 excluded, not completed
├── 1. Purpose and architecture [2/4]
│   └── Layers/agent agreed; audience and governance OPEN
├── 2. Principles [4/5]
│   └── Core boundaries/subset/auth-first agreed; consolidation OPEN
├── 3. Vocabulary and identity [3/5]
│   ├── Core terms and human groups agreed
│   ├── Trusted proxy/human attribution and canonical reuse agreed [Q-085]
│   ├── Shared actor/human identity shape agreed [Q-086]
│   ├── Human-subject compatibility with enforced proxy limits agreed [Q-087-A]
│   └── JWT identity mapping and deliberate sub/human_id duplication agreed [Q-087-B]
├── 4. Permissions [3/4]
│   ├── Naming and one permission per endpoint agreed
│   ├── No implicit inheritance, wildcards, or aliases [Q-057/058/059]
│   └── Catalog governance/evolution OPEN
├── 5. Grants and authority [7/13]
│   ├── Whole grants, human-dependent proxies, separate administration agreed
│   ├── Immutable role revisions + explicit validated grant adoption agreed [Q-089-B]
│   ├── Automatic delegation reactivation agreed [Q-070]
│   ├── Direct human membership only; nested groups unsupported [Q-077]
│   ├── Create, enable, disable, delete agreed; separate revoke superseded [Q-082]
│   ├── Enabled/disabled status agreed [Q-081 revised]
│   ├── Optional time validity independent of status agreed [Q-083]
│   ├── Minimal persistent bootstrap grants and bounded self-assignment agreed [Q-088]
│   ├── Original live-role expansion question superseded; undecided history retained [Q-089]
│   └── Admin encoding, full bootstrap procedure, lifecycle, chains, group/role changes OPEN
├── 6. Scope and registration [5/7]
│   └── Flat AND boundaries and registration agreed; detailed evolution PARKED
├── 7. Requests and resolution [6/10]
│   └── One gate and partial policy agreed; full schemas/integration OPEN
├── 8. Decision semantics [1/4]
│   ├── Allow/deny/error meanings and minimal shapes agreed [Q-051–067]
│   ├── Q-084 DISAPPROVED: no business-logic scope creep; revised question withdrawn
│   ├── Authorization-only condition semantics remain OPEN; Q-084 is not pending
│   └── Full validation, code catalog, provenance OPEN
├── 9. Enforcement and time [2/8 in scope; 1 excluded]
│   ├── Both-boundary moves agreed [Q-068]
│   ├── No stale use after confirmed grant deletion [Q-069, terminology Q-082]
│   ├── Deny instead of automatic collection filtering [Q-071]
│   ├── Explicit Finance request vs tenant-wide request agreed [Q-072]
│   ├── Complete-batch authorization before effects agreed [Q-073]
│   ├── Preserve application boundaries through concurrent changes agreed [Q-074]
│   ├── Execution-time authorization for queued work agreed [Q-075]
│   ├── Audit policy/system design OUTSIDE HANDBOOK SCOPE [Q-076]
│   └── Counts/export, transactions, freshness, concurrency, jobs OPEN
├── 10. Challenge and verify [2/4]
│   └── Examples available; adversarial and final scenario review OPEN
└── 11. Publish the foundation [0/4]
    └── Reconcile chapters, package contracts, accept v1, implementation roadmap OPEN
```

</details>

<details>
<summary>Historical opening tree — Q-050-F-era snapshot, superseded above</summary>

The original opening snapshot is retained unchanged below. Its open labels do
not override subsequent approvals, including Q-057–Q-072.

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

</details>

## Discussion progression and current selection

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
outside v1. **Q-059 / PERMISSION-005 is agreed:** no permission aliases in v1.
These three decisions close HC-04-04; catalog evolution stays open.
**Q-060 / DECISION-008 is agreed:** return supporting-grant references with allow
to the endpoint, without requiring every request to be persistently audited.
**Q-061 / DECISION-009 is not adopted:** no returned boundary fields required;
existing endpoint enforcement is unchanged. **Q-062 / DECISION-010 is agreed:**
minimal allow JSON with version, decision, and grant_ids.
**Q-063 / DECISION-011 is agreed:** minimal deny JSON with version, decision,
and error_code/error_message/error_message_reason.
**Q-064 / DECISION-012 is agreed:** evaluation-error JSON with version and the
three error fields, without decision.
**Q-065 / DECISION-013 is agreed:** reject mixed known-variant fields instead
of partially interpreting a result. **Q-066 / DECISION-014 is agreed:** require
a non-empty supporting-grant ID list for allow. **Q-067 / DECISION-015 is agreed:**
reject unknown result fields. Remaining field-level work is parked, not excluded;
full validation and result contracts remain unfinished.

### Current impact-first pass — PROCESS-007

The user requested horizontal coverage based on impact. The following is the
assistant's recommended next pass, not approval of the unanswered policy choices.
Settle governing behavior before returning to fine-grained formats; move across
branches while retaining the unfinished details in the audit.

| Priority | Branch | Why this is high impact | Audit anchors |
|---|---|---|---|
| 1 — governing rule agreed; details open | Boundary-changing writes | Can move data into or out of a caller's authority. | HC-09-04 |
| 2 — new-check revocation rule agreed; details open | Revocation, freshness, and in-flight changes | Determines when removed authority stops being usable. | HC-09-06/07 |
| 3 — restoration rule agreed; administration and details open | Administrative bounds and delegation lifecycle | Controls creation and persistence of authority. | HC-05-08/10 |
| 4 — governing list/bulk rules agreed; details open | Collections, exports, and bulk operations | A single operation can expose or change many records. | HC-09-03/05 |
| 5 — boundary-at-use rule agreed; details open | Concurrent application changes | A valid earlier check must not lead to an out-of-boundary effect. | HC-09-07 |
| 6 — execution-time rule agreed; details open | Queued/background work | Delayed execution must not accidentally preserve removed authority. | HC-09-09 |
| 7 — excluded by user | Authority-change audit | Another layer's responsibility, not handbook scope. | HC-09-08 EXCLUDED |
| 8 — nesting unsupported; lifecycle details open | Group membership composition | Determines which group grants become applicable and which dependencies resolution follows. | HC-05-12 |
| 9 — operations/status/time validity agreed | Ordinary grant lifecycle | Distinguishes creation, reversible suspension, permanent deletion, and effective authority; complete contracts remain open. | HC-05-11 / HC-07-08 |

**Q-068 / ENFORCEMENT-004 is agreed:** a move requires authority over both current
and proposed boundaries. See [operation-specific enforcement](operation-enforcement.md).
No new policy fields or extra gate; composition and concurrency remain open.
**Q-069 / FRESHNESS-001 is agreed:** no use of revoked grants by new checks after
Auth confirms revocation, even through stale cache. See
[authority freshness](authority-freshness.md) for the timing boundary and trade-off.
**Q-070 / DELEGATION-003 is agreed as corrected:** affected access is inactive
without human support and works again when support returns under a still-valid
delegation. Explicit renewal is not adopted. See
[delegation lifecycle](delegation-lifecycle.md).
**Q-071 / ENFORCEMENT-005 is agreed as corrected:** deny the partially authorized
collection request; do not derive and return an authorized subset. The user
rejected grant-derived filtering as too much authorization intelligence in the
endpoint. **Q-072 / ENFORCEMENT-006 is agreed:** Finance-only authority can cover
an explicitly Finance-bounded request, not an all-departments request. See
[collection enforcement](collection-enforcement.md).
**Q-073 / ENFORCEMENT-007 is agreed:** complete-batch authorization and required
boundary checks before effects, without automatic partial execution. See
[bulk enforcement](bulk-enforcement.md).
**Q-074 / ENFORCEMENT-008 is agreed:** preserve evaluated application boundaries
through use despite concurrent changes. See [concurrent enforcement](concurrent-enforcement.md).
**Q-075 / ENFORCEMENT-009 is agreed:** queued work needs execution-time
authorization, not reuse of submission's allow. See
[background authorization](background-authorization.md).
**Q-076 / AUDIT-001 is outside handbook scope:** the user places audit policy
and system design in another layer. The proposal is not adopted; see
[scope decision and retained history](authority-change-audit.md).
**Q-077 / GROUP-005 is agreed:** nested authorization groups are not supported;
use explicit human-to-group links. See [groups and membership](groups-and-membership.md).
**Q-082 / GRANT-008 is agreed:** create, enable, disable, and delete are the four
canonical operations. The separate revoke operation is superseded; permanent
withdrawal remains terminal through delete. Q-079's reversible suspension remains.
**Q-081 revised / GRANT-007 is agreed:** enabled/disabled status only, distinct
from effective validity. Its original three-state variant was never approved.
**Q-083 / GRANT-009 is agreed:** optional not-before/expiry bounds independent
of status; enable does not reset them.
See [grant lifecycle](grant-lifecycle.md).
**Q-084 is DISAPPROVED:** the user's intent is to prevent business-logic scope
creep into the authorization handbook. Layer 2 remains authorization integration.
The assistant's revised question is withdrawn, not pending. No condition engine
or removal of existing restrictions is approved by this rejection. See
[decision, rationale, and preserved proposals](grant-conditions.md).
**Q-085 is agreed:** retain the verified proxy identity and establish its human
authority anchor through trusted Auth-governed evidence; a caller-supplied human
identifier cannot establish that association. See [proxy attribution](proxy-attribution.md).
The user also requires canonical representation across JWT and other requester
contexts. **Q-086 is agreed:** a versioned `actor` plus `human_id` block, with
transport mappings still open. See [identity context](identity-context.md).
**Q-087-A is agreed:** keep human-subject compatibility while requiring evaluator
coverage for protected proxy-token operations, including legacy paths. Actual
deployment compatibility is not yet verified. **Q-087-B is agreed:** the exact JWT
identity excerpt preserves human `sub` and the unchanged canonical block. Retaining
`human_id` makes the block self-contained outside JWT; duplication inside JWT is
intentional and values must agree. No duplicate `agent_id` is adopted. See
[mapping, trade-off, and alternative](jwt-identity-mapping.md).
**Q-088 is agreed:** minimal bootstrap-created grants may remain;
the bootstrapped human can explicitly assign himself more grants when each
assignment fits his current administrative authority;
seed origin gives no special exemption from grant evaluation. See
[bootstrap example and rationale](bootstrap-authority.md).
**Q-089-B / ROLE-003 is agreed:** roles publish immutable revisions; each grant
explicitly adopts one revision, with authorization and boundary validation when
changing that selection. Publication does not update existing grants. Auth's own
APIs follow the same framework. The original Q-089 alternatives remain undecided
history, not selected policies; their live-update premise is superseded. See
[approved formats and rationale](role-revisions.md) and
[preserved original discussion](role-change-authority.md).
Administration bounds, restoration mechanics, collection details, transaction
guarantees, and full job integration remain open. Audit-system deliverables are
excluded, not remaining handbook work; older snapshots are superseded on that point.

Decision results, cross-boundary mutation rules, and freshness/dependency behavior
are high-impact candidates. Their precise order and answers are **not adopted**
by the earlier reconciliation. PROCESS-007 now sets the impact-first direction
and the table above is the current recommended pass. Do not reopen the settled endpoint-policy shape or
return to the parked registration-detail rabbit hole without a reason.

Earlier reconciliation note, now qualified: a completion percentage had no
explicit measurement baseline. [MEASURE-001](handbook-completion-audit.md)
now counts 35 completed, 33 open, and 1 excluded row: 35/68 in-scope closure
(51.5%). Q-076 removes audit-system design from the denominator, not adds a
completed checkpoint. Historical counts were 34/69 then 35/69 (50.7%) after
Q-059. This is a checkpoint rubric, not effort or release readiness.
Decision-result checkpoints remain open despite their partial agreements.

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
