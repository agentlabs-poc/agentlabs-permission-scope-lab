# Q-119 — Root grant-revision representation

## Agreed representation

**AGREED.** The user answered “agree” and then separately raised future
application permissions and a possible root-only wildcard. Q-119 approves the
same immutable grant-revision record for a root, omitting `parent_grant_id`
rather than using null, an empty string, a synthetic parent, or a root flag.
Derived content retains its required parent ID. Trusted establishment must be
verified independently of field omission; no ordinary parentless escape is added.

The exact example, rationale, alternatives, and trade-offs from the approved
proposal are preserved below. Grant controls, recipient-bearing assignments,
maximum intended registered initial authority, and normal root lifecycle remain
unchanged. Full establishment evidence and governed root updates remain open.

The [Q-120/Q-120A discussion](root-permission-evolution.md) now accepts automatic
root growth as the user's behavioral requirement; its mechanism is open. Q-119
does not adopt `*` or supersede Q-058's stored-format rule. The wildcard suggestion
remains exploration, not a selected encoding.

<details>
<summary>History — exact Q-119 proposal, approved above</summary>

**PROPOSED / NOT APPROVED.** This addresses the root/derived record variant in
HC-07-08 and contributes to HC-05-09. It does not reopen Q-113's trusted-root
principle, Q-114's maximum initial authority, or Q-115's group-held setup.
Q-117 incomplete-setup visibility stays parked.

## Recommendation and exact proposed shape

Use the same immutable grant-revision format for a root, with `parent_grant_id`
**omitted**. Do not add a root flag, synthetic parent, or new entity type. A
derived grant retains its required parent ID. Null and empty-string parent
values are not alternative root encodings under this proposal.

One payroll root grant in the full bootstrap bundle:

```json
{
  "version": "1",
  "grant_id": "G0-PAYROLL",
  "revision": 1,
  "permissions": [
    "hrms:payroll:payslip::read",
    "hrms:payroll:payslip::write"
  ],
  "scope": {}
}
```

This is a field-focused root-content example, **not the entire seed bundle**.
Setup still supplies maximum intended registered permissions and scope inside
the tenant, including Auth administration. Other initial grants supply remaining
intended permissions; this example neither limits the initial administrator to
payroll read/write nor prescribes the number of roots. Registration precedes
grant acceptance. Scope `{}` never removes the implied tenant boundary.

Root control/status and recipient-bearing assignments remain separate Q-107
records. Initial authority is assigned to the administrators group with explicit
human membership under Q-115. The parent-omission convention would also apply
to Q-118's role-based content, retaining its exact adopted role reference.

## Format is not trust evidence

Under the existing Q-113 rule, Auth must establish legitimate trusted root
origin for the authority being accepted before terminating lineage there.
Omitting a field proves nothing. Ordinary bounded operations cannot manufacture
root authority by submitting parentless JSON. A missing required parent creates
an ineffective orphaned route, not a root. No self-parent sentinel is permitted
by the existing cycle rule.

The established root still needs valid status, time, registration, assignments,
adoption and tenant context. Root definition existence alone gives nobody access.
The exact trusted-establishment evidence and authorized root-change procedure
remain open; this proposal is not their implementation.

## Rationale, alternatives, and boundary review

Omission expresses that there is no required parent without inventing a fake
identity or a new discriminator. Alternatives are a null parent or an explicit
root flag. Neither alternative proves trusted establishment; the recommendation
keeps one root encoding and the existing field vocabulary.

**Accepted only if approved — trade-off:** JSON shape alone cannot distinguish a
legitimate root from an invalid parentless submission. Auth must use trusted
establishment evidence bound to the authority, not a client assertion or mere
recognition of a familiar ID. That trust obligation already exists under Q-113;
no runtime check or boundary is waived by choosing omission.

**Core-philosophy check:** same grant entity, recipient separation, implicit
tenant, registration-first maximum initial authority, immutable content, explicit
adoption, ordinary lifecycle, and no personal bootstrap bypass. A valid root
terminates lineage; missing/unproven support never acquires that privilege.

**Q-119:** approve the same grant-revision record for roots, with
`parent_grant_id` omitted rather than null or a new flag, while trusted
establishment remains required independently of that omission?

Sources: [root establishment](bootstrap-authority.md), [initial assignment](bootstrap-initial-assignment.md),
[grant records](grant-record-reference.md), [role variant](role-grant-contract.md),
[cycles](lineage-cycles.md).

</details>
