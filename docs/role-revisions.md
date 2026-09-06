# Role revisions and explicit grant adoption — Q-089-B / ROLE-003

**Related grant-revision decisions:** [Q-102–Q-106](grant-revisions.md) extends
explicit adoption to reusable grant revisions selected by assignments, with
latest-only creation/upgrades and separate live controls. The role-revision
rules below remain; they are complemented by the approved
[Q-107 core grant JSON](grant-revision-format.md). Remaining schema/operation
mechanics are not closed merely by recording these approved principles.

Post-0.0.1 update: [Q-090](grant-assignments.md) moves recipients into separate
assignments. Recipient-bearing grant excerpts below are deprecated layouts;
immutable role revisions, explicit selection, and boundary-validated adoption
remain required. Shared grant definitions must not silently propagate expanded
authority to their assignments. Definition revision/adoption fields and full
role-based split-record contracts are still open; this chapter's existing role
revision fields do not settle them.

Status: **AGREED.** The user proposed separating role revisions from grant
adoption, with boundary validation when a grant adopts a revision, and instructed
“record this commit and push” after the concrete Q-089-B proposal and examples.

This supersedes ROLE-002's automatic live-role updates. The original Q-089
alternatives were left undecided; neither is retroactively approved. Their shared
assumption of automatic propagation is replaced by the revision/adoption model.
See [the preserved discussion](role-change-authority.md).

## Architectural boundary: Auth is itself protected by the framework

The user clarified that Auth's own APIs must comply with the same authorization
framework used by applications. Role publication and grant adoption are protected
administrative operations, not privileged exceptions because they run inside Auth.
Each protected endpoint has its one declared permission, required material, trusted
sources, and enforced authorization boundary under the existing endpoint contract.

Authority to administer the platform's Auth objects and the permissions assigned
to customers/application users must remain distinct. The evaluator enforces the
declared administrative contract; it must not silently turn role authorship into
permission to change customer assignments. This adds no second grant format or
business-rule layer. The exact administrative scope encoding remains open.

## Three responsibilities

1. **Author and publish a role revision.** It defines a permission bundle and is
   immutable after publication. A new definition creates a new revision rather
   than modifying the contents of an already published revision.
2. **Adopt a revision in a grant.** A role-referencing grant explicitly selects a
   revision. Changing that selection is an authorized grant change, with validation
   of the resulting assignment and its applicable boundaries before taking effect.
3. **Resolve the selected revision.** Evaluation expands that specific revision,
   retaining the grant's identity, recipient, scope, validity, and dependencies.
   It must not silently switch to the latest revision or create a new independent
   grant from the expanded view.

Publishing a revision does not change which revision existing grants adopt.
Separate grants may adopt different revisions of the same role.

## Approved contract excerpts

These fields and meanings are approved; these are excerpts, not the complete
role/grant schemas or publication/adoption endpoint contracts.

Published R-17 revision 1:

```json
{
  "version": "1",
  "id": "R-17",
  "revision": 1,
  "permissions": ["hrms:employee:certificate::read"]
}
```

Newly published R-17 revision 2:

```json
{
  "version": "1",
  "id": "R-17",
  "revision": 2,
  "permissions": [
    "hrms:employee:certificate::read",
    "hrms:employee:certificate::write"
  ]
}
```

Existing G-17 remains on revision 1 after revision 2 is published:

```json
{
  "version": "1",
  "id": "G-17",
  "recipient": {"type": "group", "id": "finance-readers"},
  "role_id": "R-17",
  "role_revision": 1,
  "scope": {"dept": "FIN"},
  "status": "enabled"
}
```

`version` identifies the contract format. The role's `revision` identifies a
particular published definition. The grant's `role_revision` records which
definition it adopts; `role_id` alone cannot make that selection explicit.
Publication of revision 2 does not make revision 1 disappear or become revision 2.
Number allocation, complete type validation, migration, and revision retention
contracts are not finalized by these examples.

## Adoption example and boundary enforcement

Assume G-18 assigns R-17 revision 1 to Engineering readers within Engineering.
After revision 2 is published, both groups still receive read only through their
respective grants, assuming the other required checks succeed.

Vinay requests that G-17 adopt revision 2. The Auth adoption operation validates
his current administrative authority for the resulting Finance read/write
assignment, including its recipient, permissions, scope, validity, and other
applicable constraints. It also applies the relevant registration/compatibility
checks. Role-publication authority alone does not authorize adoption.

If authorized and valid, G-17's `role_revision` changes from 1 to 2. It remains
G-17 with its Finance scope; it can now supply write through normal evaluation.
If not authorized or invalid, the grant remains unchanged. G-18 stays at revision
1 unless a separately authorized adoption changes it. No approval is inferred
for all grants merely from approval of one grant's adoption.

The change is validated before it takes effect, not silently adopted first and
checked later. The adopter need not personally possess the business permissions
being assigned; the existing separation of administration and use still applies.
Human-dependent agent/service authority remains bounded by the human's current
applicable authority and the delegation's own limits.

## Rationale, alternatives, and consequences

The live-role model coupled authoring a reusable definition to changing authority
through every referencing grant. A role edit could expand permissions across
independently administered scopes without an assignment change at those boundaries.

Revision pinning separates those actions. Role authors can publish a new bundle;
grant administrators decide whether a particular binding may adopt it. The same
explicit assignment boundary is checked where authority actually changes.
Immutable revisions are necessary: editing an adopted revision in place would
reintroduce silent changes despite the revision field.

The old alternatives—validate all downstream expansions during a live role edit,
or give role editors power over every current reference—are retained as history,
not selected. No global grant migration or latest-revision fallback is adopted.

This protection works in both directions. If revision 3 removes write, grants
still adopting revision 2 retain write under that revision until their authority
is explicitly changed. Publishing a narrower revision is **not immediate access
withdrawal**. Urgent withdrawal can use existing authorized grant disable/delete
operations; a separate revision-wide withdrawal facility is not introduced here.
Adoption adds operational work and permits multiple revisions to coexist; the
user selected explicit control over silent propagation.

## Remaining work and reconciliation

Publication/adoption endpoint schemas and permission names, administrative bound
encoding, concurrent adoption, bulk adoption, rollback, revision removal/retention,
legacy unpinned-grant migration, full registration compatibility, and propagation
evidence remain open. No concrete implementation or deployment conformance is
claimed. In particular, an old example lacking `role_revision` is not permission
to infer latest-revision behavior in the new model.

ROLE-001's reusable permission bundle remains. ROLE-002's live-reference behavior
is deprecated. RESOLUTION-003 retains its dependent computed-view meaning, but
the permission source is now the explicitly adopted immutable revision.
HC-05-03 remains closed as a replaced governing decision; HC-05-13 and the full
schema/integration checkpoints remain open. This records no new completion credit.

**Q-089-B — approved:** replace live role references with immutable revisions,
explicit grant revision selection, and authorized, boundary-validated adoption.

<details>
<summary>Original decision-log entries — superseded permission source, preserved verbatim</summary>

The historical AGREED labels below describe the earlier decision state. ROLE-003
now governs role revision selection; RESOLUTION-003's dependent-view meaning survives.

| Reference | Historical status | Historical decision | Historical rationale |
|---|---|---|---|
| ROLE-002 | AGREED | A role-referencing grant uses the role's current permission definition. Authorized role edits therefore change capabilities available through existing referencing grants while preserving each grant's scope, validity, and conditions. Unrelated direct grants remain separate. | User answered Q-014: "yes." Grants do not require manual upgrades to adopt a role edit. Who may make such edits, compatibility validation, revision evidence, propagation freshness, and dependent delegation limits remain separate requirements. Role editing changes declared authority; resolution must not independently invent permissions. |
| RESOLUTION-003 | AGREED | Expand a role reference into its current permission set, producing the same permission-set grant form used to evaluate explicitly listed permissions. Preserve the original grant identity, scope, recipient, validity, conditions, and source provenance, including the role definition used. Expansion is a computed view of the same grant, not a new assignment, and does not merge unrelated grants. | User agreed after clarification that the earlier JSON represented the existing G-6 after role lookup. GRANT-EX-005 retains G-6 visibly. Expanded does not mean fully resolved: application relationships or resource facts may still be missing. Exact schema, role-revision encoding, membership expansion, and final resolved-grant representation remain open. |

</details>
