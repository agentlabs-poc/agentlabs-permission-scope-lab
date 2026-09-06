# Q-118 — Role-based immutable grant revision

## Agreed shape and clarification

**AGREED.** The user answered: “agree, role_revision does not have meeing when
permission is included directly.” A grant revision has exactly one permission
source: explicit `permissions`, or the complete `role_id`/`role_revision` pair.
Direct permissions exclude **both** role fields; a stray `role_revision` is not
accepted as meaningful metadata or silently ignored. The original proposal and
its rationale remain below as history.

Approved direct-permission grant-revision example:

```json
{
  "version": "1",
  "grant_id": "G1",
  "revision": 2,
  "parent_grant_id": "G0",
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": {"dept": "FIN"}
}
```

Approved role-based alternative, assuming the referenced role revision contains
the same payslip-read permission:

```json
{
  "version": "1",
  "grant_id": "G1",
  "revision": 2,
  "parent_grant_id": "G0",
  "role_id": "R-PAYROLL-READER",
  "role_revision": 1,
  "scope": {"dept": "FIN"}
}
```

These are alternative illustrations, not two records that may coexist under the
same immutable `(grant_id, revision)` identity. Assume valid G0 support,
registration, live controls, validity, and authorized assignment adoption.
The grant's own `revision` remains required in either variant. `role_revision`
identifies a role's content only; it is not another name for the grant revision.

| Permission-source fields present | Agreed structural outcome |
|---|---|
| `permissions` only | Direct-permission variant. |
| `role_id` and `role_revision` only | Role-based variant; resolve the exact adopted role content. |
| `permissions` plus either or both role fields | Reject mixed permission-source fields, including a lone stray `role_revision`. |
| Only one of the two role fields, without `permissions` | Reject incomplete role reference; no latest-role fallback. |
| Neither source | Reject missing authority content. |

**Rationale:** each field has one purpose and each grant has one unambiguous
permission source. There is no merge, override, or precedence rule to invent.
To supply read and write, explicitly list both, use a role revision containing
both, or establish separately authorized complete grants. Each route remains
bounded by its actual parent support; separate grants do not permit field mixing.

**Core-philosophy check / accepted trade-off:** the assembled role variant uses
existing fields and the same parent, scope, validity, and assignment model.
Role publication cannot silently change this grant. Ad hoc extra permissions
beside a role are deliberately unsupported. Live status and recipients remain
outside immutable grant content, as previously agreed.

This completes the permission-source variant choice, not all of HC-07-08 or
HC-05-13. Root representation, full field/type validation, timestamp and lifecycle
contracts, support evidence, and compatibility remain unfinished. No empty-array,
universal null-handling, or arbitrary unknown-field policy is inferred here.

<details>
<summary>History — original Q-118 proposal, approved and clarified above</summary>

**PROPOSED / NOT APPROVED.** This is a bounded representation decision contributing
to HC-07-08 and HC-05-13, not a new role capability, root policy, or field family.
Q-089-B approved `role_id` and `role_revision`; Q-090 removed the recipient from
grant definitions; Q-106/107 split live controls, immutable content, and assignment
adoption. Their complete assembled role variant has not yet been approved.

## Recommendation and exact proposed shape

An immutable grant revision has **one permission source**: either explicit
`permissions`, or the pair `role_id` plus `role_revision`, not both. The role pair
belongs to immutable authority content alongside `parent_grant_id` and `scope`.
Existing live grant controls and recipient-bearing assignments are unchanged.

Existing approved role-revision shape, with illustrative identifiers:

```json
{
  "version": "1",
  "id": "R-PAYROLL-READER",
  "revision": 1,
  "permissions": ["hrms:payroll:payslip::read"]
}
```

**Proposed assembled grant-revision variant:**

```json
{
  "version": "1",
  "grant_id": "G1",
  "revision": 2,
  "parent_grant_id": "G0",
  "role_id": "R-PAYROLL-READER",
  "role_revision": 1,
  "scope": {"dept": "FIN"}
}
```

Assume valid current G0 support for FIN-read, registered relevant contracts, valid
live controls, and an authorized assignment adopting G1 revision 2. The example
omits root setup, not its validation. G1's `revision: 2` and the role's
`role_revision: 1` are intentionally different: they identify different records.
Optional grant validity retains Q-109's placement in this immutable content.

## Why this shape

No field is new. `role_id` selects the bundle; `role_revision` fixes the exact
immutable permission content. Placing the pair in grant content preserves the
recipient-free model and prevents mutable control changes from silently switching
permissions. `grant_revision` on the existing assignment still selects G1's
content; it does not select the role independently.

The explicit exclusivity rule makes field interpretation unambiguous. Alternatives
are to combine explicit permissions with a role or prefer one source over the
other. Combining needs new union/intersection semantics; precedence risks hiding
authority-changing fields. Neither is necessary to represent the existing role
bundle model. The recommendation adds a validation rule, not new authority.

## Validation consequences proposed with this variant

| Permission-source material on a grant revision | Proposed structural outcome |
|---|---|
| Explicit `permissions` with no role fields | Direct-permission variant; existing permission and boundary checks remain. |
| `role_id` and `role_revision`, with no `permissions` | Role variant; resolve that exact role revision and validate the resulting permission set against support. |
| `role_id` without `role_revision`, or vice versa | Reject incomplete role reference; no latest-role fallback. |
| Both explicit `permissions` and the role pair | Reject ambiguous permission sources; do not merge or silently choose. |
| Neither permission source | Reject missing authority content. |

This does not settle empty permission arrays, ID grammar, numeric bounds, null
handling for every field, or all unknown-field rules. Those remain full-schema
work rather than being silently adopted through this table.

## Adoption example under already-approved rules

Publishing role revision 2 with read/write does not change G1 revision 2, which
selects role revision 1. To change that role selection, publish new immutable G1
content after the applicable validation, then explicitly adopt the latest G1
revision for the intended assignment. Existing assignments do not switch by
publication alone. Validate the whole resulting authority against current actual
support and team ceilings; a role bundle is not an exception to permission subsets.

**Core-philosophy check:** preserves one permission source, parent-bounded
authority, flat AND scope, recipient-free grants, immutable role/grant content,
explicit assignment adoption, registration validation, and separate admin/source
checks. It adds no parent-assignment ID, parent-revision field, target wrapper,
prepared mode, business-rule engine, or independent automated access.

**Trade-off:** a grant cannot add ad hoc permissions beside a role. Use an
explicit permission set or an appropriately published/adopted role revision;
any additional authority route remains separately authorized and bounded. This
is the proposed exclusivity consequence, not a newly approved restriction yet.

**Q-118:** approve this role-based grant-revision shape, with exactly one
permission source—explicit permissions or an explicitly revised role, never both?

Original decisions and rationale remain in [role revisions](role-revisions.md),
[grant separation](grant-assignments.md), and [Q-107](grant-revision-format.md).
This proposal does not supersede them before approval or claim that the entire
grant schema/checkpoint is finished.

</details>
