# Grant records — current approved reference

This reference assembles the approved Q-107 records, Q-109 validity placement,
and Q-101–Q-112A lifecycle/lineage rules. It does not replace their decisions or
rationale, introduce new fields, or claim complete schemas. Older record layouts
remain in their source chapters as history. Tenant is the trusted implied boundary.

## 1. Grant identity and live control

```json
{
  "version": "1",
  "id": "G1",
  "status": "enabled"
}
```

This is the grant-wide live switch, not authority content. Disabling it makes
authority through G1 ineffective across all its revisions and assignments.
An enabled value is necessary for that route, not sufficient for allow.

## 2. Immutable authority revision using explicit permissions

```json
{
  "version": "1",
  "grant_id": "G1",
  "revision": 2,
  "parent_grant_id": "G0",
  "permissions": [
    "hrms:payroll:payslip::read",
    "hrms:payroll:payslip::write"
  ],
  "scope": {"dept": "FIN"}
}
```

G0 represents valid actual upstream support, not permission to manufacture a
root by naming it. Assume the registered scope supports these operations and
the effective parent can supply both permissions over the relevant FIN boundary.
The child inherits all parent restrictions; its local scope is additional AND
constraints. The content is immutable after publication.

Optional validity belongs here, not on the assignment or live control. The exact
approved extension is in [Q-109](grant-validity.md). A changed window requires
new revision content and explicit adoption; enabling cannot extend expiry.

## 3. Assignment adopting this revision

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

Revision 2 is assumed latest when this assignment is created/adopted. The group
already exists; a human uses its valid authority through explicit Auth membership.
Creating G1 alone does not create A1, membership, ownership, or access. Direct
human assignments remain supported, with group assignment preferred.

## What the three version-like fields mean

| Field | Selects | Does not mean |
|---|---|---|
| `version` | The expected record's contract format, initially string `"1"`. | Grant authority revision or a default for missing metadata. |
| `revision` | Immutable content of the identified grant. | A live status or a revision automatically adopted by all recipients. |
| `grant_revision` | The content this assignment explicitly adopts. | The parent's revision or the latest publication on every request. |

For a role-based grant, the separately approved `role_revision` identifies the
role bundle adopted in grant content. [Q-118](role-grant-contract.md) now approves
the assembled post-Q-107 role variant and exclusive permission-source rule.
Use either `permissions` or both `role_id` and `role_revision`, never a mixture.
With direct permissions, neither role field belongs in the record. Earlier
pending-variant status is superseded by the explicit Q-118 approval.

## Resolution and mutation examples

```text
G1 live control + G1 revision 2 + A1
                    │
                    ├── actual current supporting G0 lineage
                    ├── applicable team ceiling and explicit membership
                    └── applicable validity and human/proxy restrictions
                                      ↓
                             eligible authority route
                                      ↓
                  requested permission + trusted scope material
                                      ↓
                         decision and actual enforcement
```

An eligible route is not already an allow for every operation. Required material
and actual application-boundary enforcement remain part of the endpoint-owned
gate. Scope keys do not prove relationships merely by matching request values.

| Event / attempted interpretation | Required consequence under approved rules |
|---|---|
| Publish G1 revision 3 | A1 still adopts revision 2. No automatic migration. |
| Upgrade A1 while revision 3 is latest | Validate adoption of revision 3. Do not choose an older intermediate revision as a fallback. |
| Disable A1 | Withdraw this route. Other assignments of G1 are not globally disabled. |
| Disable G1 | No assignment or adopted revision can override the global switch. |
| Re-enable A1 | Retain its adopted revision; validate current reality. This is not an automatic upgrade. |
| Add another G1 assignment to Team1 while A1 exists disabled | Duplicate current grant/recipient binding: reject, do not bypass A1's state. |
| Expiry of the adopted content or required parent support | Cannot supply authority through the expired route; enabling flags do not reset time. |
| G0 exists but required actual support cannot be established | The definition alone is not usable support. Failure to fetch evidence is not proof of nonexistence. |
| Another team holds a broader G0 revision | Do not substitute it for the route's required lineage. Resolve actual current adopted support top-down. |
| Set child scope `{}` | Inherit the parent's constraints without additional narrowing, not tenant-wide escape. |
| Add `dept = ENG` when effective parent requires `dept = FIN` | Preserve both constraints. Never overwrite FIN to obtain ENG authority. |
| Change parent while relevant bindings are active | Existing guarded structural-change rules apply; ordinary editing cannot bypass them. |

These are sourced expected outcomes, not an executable engine test. Approval of
these individual rules does not define all record validation or API error codes.

## Source and unfinished-contract boundaries

- [Q-107 core records](grant-revision-format.md) supplies the three JSON blocks.
- [Q-102–Q-106](grant-revisions.md) supplies adoption, uniqueness, publication,
  and actual current supporting-revision rules.
- [Q-101](parent-grant-bindings.md) supplies enablement and structural guards;
  [Q-112A](direct-human-parent-context.md) rejects extra parent selector fields.
- [Q-109](grant-validity.md) supplies revision-local time validity.
- [Q-090](grant-assignments.md) supplies recipient separation and reuse.

The direct/role permission-source variants are now agreed under Q-118. Still
open: complete variant validation, ID/revision value rules,
initial/default status, full timestamp validation, support-discovery/evidence,
operation/error payloads, and unresolved authorization-condition treatment.
Do not invent defaults or publish a complete machine schema from partial examples.
The [closure inventory](grant-contract-closure.md) remains the gap checklist.

[Q-119](root-grant-format.md) now approves omitting `parent_grant_id` for
legitimately established root content; derived content retains its parent link.
The original representation gap is narrowed, not a completed root trust/update
procedure. [Q-120](root-permission-evolution.md) discusses future permissions and
a root-only wildcard proposal. Q-120A accepts automatic root growth as intended
behavior, while wildcard versus revision mechanism remains open. Current v1
no-wildcard syntax and ordinary dependent-grant selection/adoption remain unchanged.
