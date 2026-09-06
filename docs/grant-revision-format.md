# Grant, revision, and assignment JSON — Q-107

**Consolidated companion:** [current approved record reference](grant-record-reference.md)
puts these records and later approved lifecycle rules together. The
[Q-118 role variant](role-grant-contract.md) is now approved: use explicit
permissions or the complete role pair; direct permissions exclude both role
fields. The earlier pending label is superseded; Q-107 and its rationale remain.

**APPROVED CORE SHAPE.** The user answered “Approve” after reviewing all three
JSON blocks, the new `revision` / `grant_revision` fields, rationale, and
core-philosophy check. This adopts the core separation and shown fields, not
complete validity, variant, or API schemas. The semantic foundation remains
[Q-102–Q-106](grant-revisions.md).

<details>
<summary>Earlier draft status — superseded by explicit Q-107 approval</summary>

Previous status: DRAFT / NOT YET REVIEWED / NOT APPROVED. Moving recording into
the lab did not approve the layouts; approval followed their explicit review.

</details>

## Approved separation

Represent the grant's live identity/control separately from its immutable
authority revision and the assignment that adopts a revision. This is a logical
contract separation, not a requirement for three database tables or a new kind
of authority entity. Tenant remains implied.

### Grant identity and live control

```json
{
  "version": "1",
  "id": "G1",
  "status": "enabled"
}
```

The live status governs G1 across every revision and assignment. This control-
only excerpt supplies no permissions by itself.

### Immutable published content

```json
{
  "version": "1",
  "grant_id": "G1",
  "revision": 2,
  "parent_grant_id": "G0",
  "permissions": ["hrms:payroll:payslip::read", "hrms:payroll:payslip::write"],
  "scope": {"dept": "FIN"}
}
```

Assume valid required upstream G0 authority and a valid subset; root setup is
omitted, not bypassed. The `(grant_id, revision)` pair identifies this
immutable content. Its parent link and scope remain effective within current
support; immutability does not create independent authority.

### Assignment adopting revision 2

```json
{
  "version": "1",
  "id": "A1",
  "grant_id": "G1",
  "grant_revision": 2,
  "recipient": {"type": "group", "id": "Team1"},
  "status": "enabled"
}
```

Revision 2 is assumed latest at this assignment's validated creation/adoption.
Publication of revision 3 later would not update A1. Enabled does not mean allow.

## Why each field/responsibility exists

| Representation | Purpose |
|---|---|
| `revision` on content — approved | Identifies immutable published content versions of the grant. |
| `grant_revision` on assignment — approved | Stores this recipient's explicitly adopted content; grant identity alone cannot express independent adoption. |
| Grant control `status` | Mutable grant-wide switch without rewriting a published revision. |
| Assignment `status` | Existing separate control for the recipient route. |
| `version` | Existing contract-format version, not the authority content revision. |

**Rationale / core-philosophy check:** explicit selection prevents silent upgrades;
immutable content prevents edits beneath an adoption; separate controls preserve
global suspension and per-assignment disablement. Parent links and boundary
checks remain mandatory, and no parent-assignment reference is introduced.
The cost is explicit revision-selection state and a separate immutable-content
representation; the user accepted that cost when approving these core layouts.

**Q-107, answered “Approve”:** approve these core excerpts and the revision-
selection fields as the representation of the agreed separation?

## Not settled by this core-shape approval

[Q-108](assignment-validity.md) now **defers assignment-specific validity in v1**
while retaining Q-083 grant validity. No validity field is added to assignments.
[Q-109](grant-validity.md) **approves optional grant validity in immutable revision
content**, with its exact extended JSON excerpt. Changing/removing a bound requires
new content and explicit adoption; live enablement does not reset the window.

Q-118 above now settles the direct/role permission-source variants and exclusive
field presence. The earlier variant-open item is narrowed, not full schema closure.
Remaining: complete field/type validation; timestamp
validation and condition placement; root shape; latest selection under concurrency; full API
payloads; and eligible-support discovery/evidence for direct-human routes.
[Q-112A](direct-human-parent-context.md) reaffirms lineage-supported latest and
withdraws the proposed `parent_grant_revision` field; none is added here. No revision-specific
disable/delete, rollback, draft lifecycle state, or physical storage schema is
adopted. Earlier pre-revision examples remain historical checkpoints where their
layout differs. This approval is not publication of every remaining schema rule.
