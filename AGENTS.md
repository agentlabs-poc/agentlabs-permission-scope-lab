# Handbook working rules

## Q-130 approved — multiple complete routes across batch items

Different complete valid grant routes may authorize different items in one batch.
Every item requires the endpoint's same declared permission and complete scope/
lineage support; do not mix permission from one grant with scope from another.
One uncovered item denies the whole batch before effects; evaluation failure
remains error. No endpoint grant inspection or automatic subset filtering.
See docs/bulk-enforcement.md. Move-state composition and exact transports remain open.

## Q-129 approved — bounded completion after prior allow

An ordinary synchronous application operation already allowed before later
authority withdrawal may finish within its evaluated boundaries. No allow reuse
for new data/retries/operations. Preserve Q-074 changed-data-boundary protection
and Q-110's stronger Auth-write consistency. Queued/streamed/long-running and
not-yet-allowed evaluations are not covered. See docs/concurrent-enforcement.md.

## Q-128 approved — freshness for all confirmed authority reductions

New checks after Auth confirms a reduction cannot use withdrawn authority from
stale evidence. Covers membership, grant/assignment controls, delegation, and
adopted narrowing, not merely grant deletion. Other complete valid routes may
allow; inability to establish freshness is not proof of policy denial. No cache
protocol or per-request network call is prescribed. Already-in-flight work remains
open; Q-110 Auth writes and Q-074 application-boundary preservation stay unchanged.

## Q-127 approved — no proxy-to-proxy delegation chains in v1

Support direct human-to-agent/service-account delegation only. Multiple proxies
can have their own human-linked bounded delegations; collaboration transfers no
authority implicitly. Team/grant hierarchies remain supported. Do not flatten an
unsupported proxy chain into a claimed direct delegation. See docs/delegation-lifecycle.md.
Earlier chain-design gaps are not v1 requirements after this explicit decision.

## Q-126 approved — permission identifiers retain authorization meaning

Do not repurpose an existing permission identifier for a materially different
operation, even after retirement. Use a new identifier; cosmetic labels may
change without semantic change. Auth preserves/validates registration identity,
not business semantics. See docs/permission-lifecycle.md for rationale and limits.

## Q-125 approved — permission retirement may retain grant references

Authorized retirement need not wait for every referencing grant to be edited.
Once effective, the computed root no longer supplies that permission and stored
references cannot authorize it. Grants/assignments remain unchanged; no silent
rewrite/delete/disable. See docs/permission-lifecycle.md. Exact retirement API,
visibility, reuse, and mixed-grant consequences remain open, not assumed.

## Q-124 approved — bounded bootstrap continuation

Authorized retries may resume only the same intended interrupted setup after
revalidation. No silent administrator/authority replacement, conflicting-attempt
merge, or usable partial bootstrap authority. Completed replay remains Q-116;
incomplete visibility remains Q-117. Exact intent/evidence/recovery contracts
remain open. See docs/bootstrap-initial-assignment.md; do not re-ask retry policy.

## Q-117 approved — incomplete bootstrap grants no initial authority

The user has now approved the previously parked Q-117. New bootstrap authority
becomes usable only after the complete intended arrangement is validated and
durably established. Partial records may persist without supplying access.
Completed setup with a lost response follows Q-116's non-mutating replay rule.
No new status or transaction design is adopted; continuation/recovery remain open.
Earlier parked/unapproved notices below are historical, not the current gate.

## Q-123 corrected — one application version for everyone

The user explicitly rules out selective tenant application updates. Each
application has one shared current version/catalog; its existing roots compute
coverage from that catalog. Do not reintroduce per-tenant release pins or catalog
adoption. Tenant isolation and ordinary grant/role revision adoption remain.
See docs/root-permission-evolution.md; earlier version-binding proposals are
superseded, not requirements. Publication visibility and source encoding remain.

## Q-122 approved — computed root permission coverage

Root effective permission coverage is computed from applicable registered
application capabilities. Do not re-ask computed versus materialized revisions
or require automatic root assignment adoption for catalog additions. Scope,
normal enablement/validity, and ordinary explicit child/role adoption remain.
No root wire encoding/wildcard is approved. Exact source binding, applicable
catalog version, removal, and concurrency remain open; preserve earlier rationale.

## Q-121 approved — application platform authority

Use the existing application platform administrator for application permission/
scope publication, outside tenant business authority but within their actual
platform-management bounds. See docs/application-platform-authority.md. Do not
adopt the earlier invented publisher-grant example or require tenant payslip
access to define a payslip permission. Q-120A feeds applicable root growth;
mechanism, precise bindings, and platform-context contracts remain open.
Record approved decisions with rationale in the lab; the last commit/push was
the completed Q-120A/ASSESS-001 checkpoint, not approval of future proposals.

## Publication checkpoint — decisions through Q-120A and ASSESS-001

The user explicitly requested committing and pushing the current work. Publish
the verified handbook/reader changes and closure assessment on main, preserving
approved decisions, rationale, and clearly marked open proposals. This does not
approve Q-117, the root-growth mechanism, or any other pending contract choice.
Preserve both existing tags; `0.0.2` stays at `5193e88` and is not included in
this branch-only push. Earlier no-commit/no-push statements below describe the
preparation stage. No runtime work or scratchpad changes are authorized here.

## Current contract decision — Q-118 agreed

Q-118 approves role-based immutable grant content and exclusive permission-source
fields. Direct `permissions` excludes both `role_id` and `role_revision`; the
user explicitly notes role revision has no meaning with direct permissions.
Role-based content requires both role fields and no direct permissions. Preserve
the original proposal/rationale as history. Q-119 now approves root-content
parent omission, independently validated trusted establishment still required.
Q-120A accepts automatic root permission growth on legitimate application
capability upgrades, without a separate manual root update/adoption gate. The
original Q-120 deliberate-update recommendation is superseded in that respect.
Live wildcard versus automated materialized revisions remains open; Q-058's
stored-format rule is not superseded by implication. Ordinary child selection/
adoption stays fixed. Root trust, tenant/app binding, and revision mechanics remain open.
Local tag `0.0.2` remains at committed `5193e88`, excluding all uncommitted work.

## Current task — closure pass with decision preservation

The user approved consolidating approved decisions, writing documentary gaps,
asking only genuine blockers, and preparing v1 acceptance. They explicitly
reaffirmed: preserve agreed/decided content unless superseded. New consolidation
must cite its existing approvals; meaning changes require explicit user approval.
Preserve older decisions, rationale, and measurement snapshots with clear history
labels. Do not silently exclude criteria to meet today's hoped-for finish.

Current work tracker: docs/v1-closure.md. Q-117 is unapproved and parked, not
adopted or excluded. Grant contract gaps are in docs/grant-contract-closure.md;
do not reopen settled lineage or add withdrawn fields. This documentation pass
does not authorize runtime work, scratchpad mutation, commit, or push.

## Current publication authorization — Q-113 / Q-114 bootstrap checkpoint

The user explicitly requested “keep going. commit and push” after the
registration-first correction. Publish the verified Q-113/Q-114 documentation,
setup SVG, and reader reconciliation on main. Preserve prior history, scratchpad
originals, and tag `0.0.1`. No runtime changes or implicit approval of the next
bootstrap proposal is included. Continue lab-only, deepest-impact-first discussion
after the checkpoint; earlier publication instructions describe prior checkpoints.

## Current publication authorization — through Q-112A

The user requested “commit and push, and continue,” then asked for confirmation
that all scratchpad discussion was captured and correctly reconciled. Publish
this verified handbook checkpoint on main, including source-history capture and
the source-to-chapter reconciliation map. Preserve originals and tag `0.0.1`.
No runtime changes or promotion of historical/tentative contracts is authorized.
After publication, continue lab-only discussion, deepest dependencies first.
Earlier no-commit/no-push statements below describe preparation checkpoints.

## Latest recording instruction — lab only, including exploration

The user explicitly instructed: “stop using scratchpad, make everyting is capture
in labs.” All further discussion, uncertain exploration, proposals, rationale,
and approved decisions belong in this lab. Do not create or edit more scratchpad
files. The complete 26-file source snapshot is in docs/history/scratchpad-import;
originals remain unchanged. Historical proposals are not promoted by capture.
This supersedes every lab-first/scratch-for-uncertainty direction below. It does
not authorize commit/push, runtime changes, or new decisions.

Q-112A reaffirms Q-103's **lineage-supported latest**, resolved top-down from
current valid adopted support, not simply latest published elsewhere. The
assistant's independent `parent_grant_revision` proposal is withdrawn, not
canonical. Retain `parent_grant_id` as the explicit lineage link and separately
track support-discovery/validation/evidence contracts. Do not re-ask the governing
selection principle or claim every resolver detail is implemented.

## Current recording location — lab by default, scratchpad for high uncertainty

The user instructed: “you should start putting them in lab repo, stop using
scratch pad, we should use when uncertaininty is high.” Continue ordinary
handbook discussion, approved decisions, rationale, and clearly labeled bounded
proposals in this lab. Use the sibling scratchpad only for high-uncertainty
exploration; preserve its existing files as history. This supersedes the earlier
scratchpad-default and lab-recording freeze notices below for this work.

Q-102–Q-106 are recorded in docs/grant-revisions.md. Q-107's three core JSON
shapes and revision-selection fields were subsequently explicitly approved;
see docs/grant-revision-format.md. Earlier draft status is preserved as history.
Full schemas remain unfinished. Changing recording location or approving a
shape does not authorize commit/push, runtime migration, or unrelated promotion.

For every question, provide a reference, recommended answer, rationale, and a
check against the core philosophy. Name trade-offs and unresolved risks; do not
claim a recommendation is approved. Record decisions with sufficient explanation
and examples, preserve superseded reasoning, and ask one question at a time.

## Current prioritization — deepest dependencies first

Q-114 now agrees maximum intended initial authority inside the tenant, followed
by equal/narrower ordinary distribution. Registration of relevant permissions
and scope contracts, including Auth administration, is a mandatory prerequisite
to initial grant acceptance under Q-039–Q-042. Do not omit this dependency or
reintroduce the FIN-read-only seed recommendation. Full seed/root and initial
recipient contracts remain open. Q-115 now approves trusted setup creating a
selected legitimate human user and administrators group, explicit root assignment
to the group, and explicit human membership. No special username or evaluator
bypass. Earlier pending-arrangement wording is superseded. Q-116 now agrees
repeated completed bootstrap reports already initialized with no authority
mutation; ordinary authorized administration remains available. Q-117 incomplete
setup authority visibility is proposed only, not approved. No new bootstrap
status field, wire error code, or recovery contract is implied.

The user explicitly prioritizes important, deep-impact decisions before shallow
completion work. Work outward from authority sources and lineage, to safe
authority mutation, resolution consistency, full contracts/integration, then
editorial packaging. Do not optimize for easy checklist closures or re-ask settled
invariants. Q-112 reaffirms group assignment as preferred (not mandatory) and the
parent's permission/scope ceiling for direct assignments. Q-112A above reaffirms
lineage-supported latest; implementation contracts are not a new open vote on it.

For each substantive recommendation, check existing philosophy, rules, and bounds
before presenting it; name any tension and the remaining proof/contract gap.
Do not present a consequence of an already-approved boundary as a newly discovered
loophole or repeatedly seek approval for it. Q-113's user correction is the example:
ordinary grant operations are already bounded. Its trusted-root establishment
rule is agreed, while the actual trust procedure remains unfinished.

## Current publication authorization — Q-101 checkpoint

The user explicitly requested “commit and push, what next” after reviewing the
Q-101 lab pinning. Commit and push this verified documentation checkpoint to
the existing main branch. Preserve scratch originals and tag `0.0.1`; no runtime
migration, new policy approval, or unrelated publication is implied. Earlier
no-commit/no-push notices below describe the local preparation stage. Continued
discussion stays in the active scratchpad unless lab integration is requested.

## Current scoped authorization — pin Q-101 in the lab

The user requested pinning the approved scratchpad discussion in the lab with
diagrams and broad case coverage. Local documentation/reader reconciliation is
authorized for Q-101 through Q-101E-3. No commit, push, runtime migration, or
unrelated scratch-proposal promotion is authorized. Scratch originals and tag
`0.0.1` remain preserved; earlier freeze/publication notices below are history
for this scoped update.

Current source: docs/parent-grant-bindings.md. Parent-grant lineage uses
`parent_grant_id` plus actual assignments/team context, not an additional
`parent_assignment_id`. Grant enablement, assignment enablement, and effective
lineage validity remain distinct. Affected assignment changes proceed bottom-up.
Q-101E-3 permits relevant bindings to be removed OR disabled before parent
changes; explicit re-enablement validates current reality. Removal-only wording
is superseded. Preserve the older reasoning with clear history labels.

## Q-100 publication authorization

The user explicitly requested “lets commit and push” after the Q-100 integration.
Commit and push this verified checkpoint to main, including the earlier local
freeze-record commit `51f1a6c`. This does not reopen unrelated lab changes or
authorize future publication automatically. Scratchpad remains active; preserve
its files and tag `0.0.1`. Earlier no-commit/no-push notices below are history
for this checkpoint; the lab freeze otherwise remains in effect.

## Scoped exception — Q-100 integration

The user requested adding the Q-100 scratchpad write-up and SVG to the lab.
Local edits for this integration and its handbook/reader reconciliation are
authorized. The freeze remains for unrelated lab changes and further commits;
no push is authorized. Scratchpad remains active and its originals are preserved.

Q-100 records distinct administrative-operation evaluation and authority-boundary
validation inside one Auth endpoint-owned gate. Neither check substitutes for
the other or for runtime dependency/application enforcement. Exact contracts
remain open; Q-097/Q-098 are contextual scratch discussions, not independently
promoted decisions. See docs/auth-service-authority-gate.md.

## Latest gate — lab frozen; scratchpad active

The user instructed: “freeze agentlabs-permission-scope-lab and commit and
unfreeze scratchpad.” Record and commit this gate checkpoint, then make no
further lab file edits or commits until explicitly reopened. This instruction
does not authorize a push. Read-only inspection remains permitted.

The sibling `../auth-scratch pad` is unfrozen for continued discussion and
explicitly requested recording. Its prior freeze is lifted. Scratch proposals
must not be promoted into the frozen handbook automatically. No scratch grant,
membership, or ownership record changes are implied by reopening the workspace.

Rationale: hold the agreed lab baseline stable while exploring the next model
changes separately. Earlier gate statements below are preserved as history.

## Latest publication authorization

The user explicitly instructed “let commit and push” after the Q-099
reconciliation. Commit and push of the pending agreed handbook checkpoint to
main are authorized after fresh verification. Preserve tag `0.0.1`; exclude
external scratchpad files and do not promote Q-097/Q-098 or other open proposals.
Earlier commit-freeze statements below describe previous checkpoints.

## Current gate and Q-099 update

The user lifted the lab file-modification freeze and authorized the proposed
Q-099 reconciliation. Local file edits are enabled; commits and pushes remain
frozen until explicitly authorized. Earlier gate statements below are history
where they conflict. Preserve existing changes and deprecated explanations.

Q-099 is AGREED: ownership changes do not automatically change team-held business
authority. The acting administrator needs current authorization; continuing
support is the selected team assignment and its actual lineage, not the original
owner's membership merely because they acted. Explicit personal dependencies
remain required, including upstream dependencies; no automatic migration/rebinding.
See docs/ownership-lineage.md. Q-096's two-owner recommendation and exact ownership
contracts remain open. Do not import Q-097/Q-098 or canonize scratch owner-list JSON
under this scoped approval. No scratch ownership records are changed.

Previous checkpoint, superseded by Q-099 where noted:

Latest: Q-094 orphan definition is agreed; ownership transfer must be explicit,
not automatic. Q-096 proposes team-based ownership and two owners using team-held
authority. Do not silently replace Q-093's assigner-membership dependency: the
proposed distinction between team-held continuing support and the acting owner's
authorization still needs agreement. No two-person approval workflow is adopted.

Latest update: the user explicitly requested recording Q-093. Assignment now
requires separate administrative authority and a valid supporting parent route
available to the assigner directly or through membership. Its combined link fields
remain open. This supersedes provision-without-possession for dependent assignments.
Q-095 records the user's clarification: canonical entities remain teams/groups
and grants; “sub-” is a parent-dependent relationship, not a new entity type.
All child-team authority and each child grant stay within their respective parents;
no separate assignment may bypass the team ceiling. Q-094 orphan lifecycle details
remain open. Keep comments short, record rationale, and do not infer commit/push
authority. See docs/assignment-authority.md and docs/authority-lineage.md.

Previous checkpoint:

Latest recorded approval: Q-092 adopts the initial `auth:group::create/write/delete`
bundle (three explicit permission strings), create includes subteams, write includes
human membership, and grant assignment requires separate authority. Scope `{}` is
tenant-wide. See docs/team-administration.md. User said “agreed, move next”;
continue short, one-step discussion and record approvals with rationale. This does
not reopen commit/push or settle the next assignment permission/bounds proposal.

## Current gate and model — Q-090 / Q-091

Tag `0.0.1` preserves commit `247e8392bb09885a9ff1c8ce94e5205a279e6852` on
local and remote main. The user then explicitly authorized making subteams/
subgroups and recipient-free grants canonical. Recording and reconciling these
two decisions is authorized; do not infer permission to commit/push this new
checkpoint or to finalize unrelated open questions. Earlier gate text below is
historical where it conflicts. The user currently requests short answers and
no unsolicited approval questions.

Q-090 separates reusable grant definitions from recipient-bearing assignments.
Q-091 permits explicit dependent subgroup authority, not membership inheritance.
Preserve parent-assignment dependencies; the combined derived-assignment wire
format, hierarchy record, owner rules, and definition revision fields remain open.
See docs/grant-assignments.md and docs/subgroups.md. Preserve deprecated text and
scratch examples; do not migrate runtime state or invent approved fields.

## Commit gate — reopened by the user

The user explicitly instructed: “lets commit and push, and continue with the
rest.” The checkpoint is authorized for commit and push after verification.
Continue the agreed handbook discussion and checkpoint workflow, respecting any
later user restriction. Do not force-push or treat this as authority to change
unrelated repositories or systems.

Current cadence (user instruction after Q-083): record each decision and its
rationale locally immediately, but commit and push roughly once per ten
questions, or when explicitly requested. Do not commit after every answer.
Q-083 is the first approved question recorded after checkpoint `01d68a5`.
The user explicitly requested a checkpoint after approving Q-087-B, including
its rationale for deliberate `sub`/`human_id` duplication. This request is within
the explicit-checkpoint exception; roughly ten questions remains the default.

Latest gate instruction: after Q-089 was left undecided, the user froze recording
and commits. The subsequent “record this commit and push” explicitly authorizes
recording Q-089-B's revision/adoption decision, reconciling affected current docs,
and committing/pushing this checkpoint. Do not infer unrestricted recording of
later discussion from this scoped reopening; await further user direction.

Previous gate state, retained as history: the user had frozen commits and pushes
while reviewing local reconciliation. Freeze notices in archived snapshots and
completed plans describe that earlier state, not the current instruction.

## Source of truth and preservation

The user's approved discussion decisions take precedence over older lab prose,
`src/content`, interactive simulations, and diagrams. Do not turn an open
proposal into an agreed rule during reconciliation.

Preserve superseded material with clear deprecation labels or linked archives.
Record rationale, examples/counterexamples, consequences, and open questions,
not just conclusions. Keep reference numbers on discussion questions/proposals.

When proposing a contract or format, include its exact versioned JSON/YAML block
in the same discussion message as its meaning, rationale, and approval question.
The user requested this after Q-087-A so semantics and representation can be
reviewed together. Label excerpts versus complete contracts explicitly.

## Discussion pacing — impact-first horizontal coverage

The user explicitly asked for horizontal coverage based on impact. Ask one
question at a time and continue after each answer without waiting for “proceed.”
Prioritize unresolved authority and enforcement decisions across branches before
extended field-by-field validation. Keep lower-level gaps visible for a later
contract-completion pass; parking them does not approve or exclude them. Do not
mark a whole branch complete after settling only its governing principle.

Keep architectural context visible: explain where each substantive proposal fits,
what is already agreed, what would change, and its rationale and trade-offs.
Progress does not remove the need for proposals or architecture discussion.
Update architectural prose/diagrams when approved decisions change the flow or
responsibility boundaries; do not redraw them for every field-level decision.

## Handbook scope — audit belongs to another layer

Q-076 explicitly excludes audit-system policy and design from this handbook.
Do not treat event selection, audit schemas, delivery/failure guarantees, storage,
retention, or disclosure as remaining handbook deliverables. Preserve the rejected
proposal as history. Authorization-result evidence and explicit authorization of
administrative changes remain in scope; earlier audit references are integration
context, not authority to design the external audit layer here.

## Handbook scope — authorization, not business logic

The user's Q-084 correction prohibits expanding this handbook into business-rule
design. Layer 2 means application-specific authorization integration: establishing
trusted authorization material and enforcing the evaluated boundaries on the
actual operation. It is not a business-logic layer within this handbook.
Q-084 is explicitly DISAPPROVED: record the intent to prevent business-logic
scope creep, not merely a wording correction. The assistant's revised question
is withdrawn, not pending. Do not infer approval either to remove generic
authorization conditions or to introduce a condition engine. Preserve rejected
proposals and rationale without automatically re-proposing them in narrower form.

## Role references — Q-089-B supersedes live updates

Published role revisions are immutable. Role-based grants explicitly select
`role_revision` alongside `role_id`; publication does not silently update grants.
Adoption is an authorized grant change with boundary validation. Resolve the
adopted revision, not latest. Preserve ROLE-002/live-role text as deprecated history.
Auth's own protected APIs must follow the same authorization framework. Publishing
a narrower revision does not withdraw permissions from older adopted revisions.
