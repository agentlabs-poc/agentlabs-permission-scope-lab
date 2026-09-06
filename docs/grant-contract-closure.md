# Grants and assignments — contract closure inventory

This is a consolidation, not a new schema proposal. It shows exactly which
parts of HC-05-08/11/12/13 and HC-07-08 can be written from approval and which
must not be invented. Full-contract checkpoints remain open. Existing sources
take precedence over older recipient-bearing examples.

**Closure work:** [current approved record reference](grant-record-reference.md)
assembles the existing three records and their lifecycle consequences without
new policy. [Q-118](role-grant-contract.md) now approves the assembled role-based
revision and exclusive permission-source validation. Direct permissions exclude
both role fields; role-based content requires the pair. Original pending status
is preserved in that chapter. This advances variant work without full schema closure.

## Approved representation, ready to reuse

| Record or field | Settled meaning | Source |
|---|---|---|
| All published records: `version` | Required supported string, initially `"1"`; not a content revision. | [Q-050-A](contract-publication.md) |
| Grant control: `id`, `status` | Stable grant identity; enabled/disabled is live and grant-wide. | [Q-107](grant-revision-format.md) |
| Grant content: `grant_id`, `revision` | Identifies immutable published authority content. | [Q-106/107](grant-revision-format.md) |
| Content: `parent_grant_id` | Explicit grant-lineage link. Actual support must still be established. No extra parent-assignment or independently pinned parent-revision field is adopted. | [Q-101](parent-grant-bindings.md), [Q-112A](direct-human-parent-context.md) |
| Content: `permissions`, `scope` | Explicit operation set plus required flat AND boundary constraints. Children retain the effective parent ceiling. | [Q-107](grant-revision-format.md), [scope](scope-model.md) |
| Optional content: `validity` | Optional `not_before` and/or `expires_at`; inclusive start, exclusive expiry. Changes require new content and explicit adoption. | [Q-083](grant-lifecycle.md), [Q-109](grant-validity.md) |
| Assignment: `id`, `grant_id`, `grant_revision` | Assignment identity and the content explicitly adopted for its recipient. | [Q-107](grant-revision-format.md) |
| Assignment: `recipient.type`, `recipient.id` | Human/group recipient binding, not part of reusable grant content. | [Q-090](grant-assignments.md), [Q-107](grant-revision-format.md) |
| Assignment: `status` | Live route control, distinct from global grant control and effective support. | [Q-101](parent-grant-bindings.md), [Q-107](grant-revision-format.md) |
| No assignment-specific validity in v1 | Explicitly deferred, not forgotten or implicitly inherited as a new field. | [Q-108](assignment-validity.md) |
| Role reference: `role_id`, `role_revision` | Explicitly adopted immutable role revision; publication does not silently change the grant. | [Q-089-B](role-revisions.md) |

Q-107 approves the three core blocks; Q-109 approves the validity extension.
Q-118 now approves the assembled role-based content shape after the Q-107 split;
full validation/publication contracts remain open.
No implied fourth identity record or new field is needed to explain this inventory.

## Approved lifecycle behavior, not questions to repeat

| Operation or event | Settled rule | Exact remaining contract work |
|---|---|---|
| Ordinary creation/assignment | Administrative authority AND valid source-boundary support; no parent-omission escape. | Operation payloads, source discovery/evidence, recipient-bound validation. |
| New assignment | One current assignment per tenant/grant/recipient including disabled; latest published revision must pass current bounds, no old fallback. | Complete key/ID validation, duplicate/conflict responses. |
| Publish revised authority | Published content immutable; later publication does not change existing adoption. | Full role/direct variants, revision allocation/publication contract. |
| Explicit adoption | Upgrade to latest under actual current lineage, top-down. Existing assignments may remain on older content. | Adoption request/result and consistent lookup/write boundary. |
| Disable grant | All authority through that grant becomes ineffective. | Initial status default and complete operation responses. |
| Disable assignment | Withdraw that recipient route without globally disabling the grant. | Full route-operation contract and binding validation. |
| Parent change | Relevant bound assignments must be removed or disabled, affected changes bottom-up. | Exact hierarchy records, discovery of affected bindings, operation payloads. |
| Re-enable | Explicitly disabled records need explicit enablement, validating current reality. Still-enabled descendants can regain effectiveness when support returns. | Complete transition/precondition representation, not a new automatic repair rule. |
| Missing required support | Affected orphan lineage ineffective; no automatic reparenting/deletion. Fetch failure is not proof of absence. | Provenance/error representation and governed repair. |
| Grant deletion | Canonical permanent withdrawal; no separate revoke operation. | Complete deletion/dependency response contract and propagation proof. |
| Expiry | Adopted revision's local window and required upstream limits govern; re-enable does not extend time. | Timestamp syntax, malformed windows, trusted clock behavior. |
| Concurrent invalidating change | Do not persist authority checked against invalidated state. | Conflict/retry/consistency contract; no chosen database mechanism. |

Sources: [revisions](grant-revisions.md), [bindings](parent-grant-bindings.md),
[lifecycle](grant-lifecycle.md), [write consistency](auth-write-consistency.md),
[lineage](authority-lineage.md).

## Current drafting and contract-review gaps — through Q-130

**C01-D1 progress:** [combined boundary-validation behavior](authority-boundary-validation.md)
is now written with deterministic parent-team lookup, operation coverage and
28 sourced cases. The support-discovery item below is narrowed to direct-human
routes, recipient-relative source binding and exact interface/evidence contracts;
the governing parent/subset rule is not an unanswered question.

These are unfinished deliverables, not all blockers requiring new user decisions.
Draft from approved fields and rules; present coherent versioned examples for
genuinely new choices instead of one question for each unresolved detail.

1. **Computed-root encoding and trusted establishment evidence.** Q-119 parent
   omission, Q-121 platform publication authority, Q-122 computation, Q-123 shared
   catalog, and Q-117/Q-124 coherent/same-intent bootstrap are settled. Specify the
   exact trusted source representation and recovery/intent evidence without a
   wildcard, tenant release pin, new root flag, or revision churn by implication.
2. **Actual support discovery and evidence.** Reuse parent ID, actual assignments,
   team context, top-down adopted lineage and complete-route validation. Finish
   evidence/input/output representation; do not re-ask the selection principle.
3. **Complete validation and operation records.** Ordinary direct/role variants
   are approved under Q-118. Consolidate defaults/IDs/timestamps/lifecycle/adoption
   requirements and identify only unsupported format choices for review. Q-128
   supplies reduction freshness; Q-110 still protects Auth writes.
4. **Human relationship and direct-delegation records.** Complete membership,
   team/ownership and direct human-proxy evidence. Q-127 explicitly excludes proxy
   chains; they are not a v1 schema requirement. Q-070 restoration remains agreed.
5. **Residual authorization-only restrictions.** Inventory existing references
   before asking whether any new mechanism is needed. Do not revive Q-084's
   rejected business-rule proposal or silently drop restrictions.

Q-125/Q-126 settle permission retirement and stable meaning; Q-130 settles finite
batch per-item complete-route composition. Their exact evidence/operation formats
still belong to contract drafting, not repeated votes on the approved behavior.

<details>
<summary>Previous closure-gap inventory — retained history, not a current blocker list</summary>

## Genuine completion blockers — historical label

These are gaps, not approved answers or a mandate to ask one question per cell.
The assistant should draft a coherent contract from existing rules, marking only
the unsupported choices for review with exact versioned JSON where applicable.

1. **Root representation and trust boundary.** Q-113–Q-116 establish the intended
   setup and replay behavior. The complete root record, trusted provisioning
   identity/evidence, and partial/concurrent setup handling are not specified.
   This blocks claiming a full root-to-leaf contract, not the ordinary subset rule.
   [Q-119](root-grant-format.md) now approves root-content parent omission; it
   does not complete establishment evidence or recovery. [Q-120](root-permission-evolution.md)
   and Q-120A accept automatic new-permission root growth without a separate
   manual expansion step. Neither wildcard encoding nor the root-update trust/
   revision contract is settled. Ordinary child selections remain bounded.
2. **Support discovery and evidence.** Define inputs/output and validation for
   finding actual eligible parent support, including direct-human and team routes.
   Use already-approved lineage-supported latest; do not reintroduce withdrawn
   parent fields or an issuer-membership dependency merely to fill the gap.
3. **Complete record variants and validation.** Q-118 settles direct/role
   permission-source shapes; finish root/derived representation, full variant
   validation, IDs, revision values, timestamps, initial status,
   and operation contracts. Unknown details must not become executable schema
   defaults by editorial choice.
4. **Membership, hierarchy, and delegation contracts.** Human-only membership,
   team ceilings, and human-dependent automation are settled. Their full shared
   records, delegation chain limits, and support-change mechanics are not.
5. **Existing conceptual conditions.** Q-084 rejected business-logic expansion
   and withdrew the generic-condition proposal. Old condition references do not
   justify either silently ignoring restrictions or inventing a condition engine.
   Inventory any actual authorization-only requirement before raising a bounded
   decision; do not re-ask the rejected business-workflow question.

</details>

## Acceptance checks for the finished package

- Every published grant/control/revision/assignment variant has its required
  version, fields, validation, and source decision documented.
- Root establishment and ordinary derived creation cannot be confused by a
  missing field or caller-controlled flag.
- One worked user → group → child-group → human/proxy route retains actual
  assignments, adopted support, permission subsets, and AND scope throughout.
- Disabled, expired, deleted, orphaned, cyclic, stale, duplicate, and conflicting
  mutations have explicit expected outcomes and no amplification fallback.
- Role publication, grant publication, and assignment adoption remain different
  events; no example silently upgrades authority.
- No historical recipient-bearing, live-role, prepared-state, or extra-parent-ID
  format is presented as a second current schema.

These acceptance checks do not require building the Auth engine to finish the
handbook. Implementation and evidence tasks are separately tracked in the
[implementation roadmap](implementation-roadmap.md).
