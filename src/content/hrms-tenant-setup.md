# HRMS: grants, declared inputs, and bounded execution

This is a worked explanation of the agreed model, not a deployable HRMS policy
or a claim about another repository. Permission identifiers and boundary
relationships below are illustrative application registrations. The
[original HRMS guide](../../docs/history/reconciliation-2026-09-05/src/content/hrms-tenant-setup.md.txt)
is preserved as deprecated history.

## Employee membership and self-scoped payslips

Auth stores the employee authorization group and Vinay's human membership. HRMS
may synchronize that membership from its business records, but an HRMS department
or reporting line does not automatically confer Auth membership.

This abbreviated working grant binds payslip read to each member's own boundary:

```json
{
  "version": "1",
  "recipient": { "type": "group", "id": "employees" },
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": { "user": "$self" }
}
```

Vinay's self is Vinay, not the employees group. HRMS defines which employee
record corresponds to that human and how a payslip belongs to that employee.
For P-17, it must establish or enforce the ownership relationship before
returning the payslip. Possession of the ID alone is not proof.

**Why groups are preferred:** the same grant supports every valid member without
creating one grant per employee. Direct human grants remain permitted. An agent
acting for Vinay can use only the delegated subset of Vinay's currently valid
authority; it cannot join the group independently.

## Finance certificate readers

```json
{
  "version": "1",
  "recipient": { "type": "group", "id": "certificate-readers" },
  "permissions": ["hrms:employee:certificate::read"],
  "scope": { "dept": "FIN" }
}
```

These examples omit lifecycle details for clarity, not because those restrictions
can be discarded. Membership-derived grants remain dependent views with their
source binding intact.

The endpoint declares one permission and exact input sources:

```json
{
  "version": "1",
  "method": "GET",
  "path": "/api/v1/{tenant}/{dept}/{cert}",
  "permission": "hrms:employee:certificate::read",
  "inputs": {
    "tenant": { "source": "path", "name": "tenant" },
    "dept": { "source": "path", "name": "dept" },
    "cert": { "source": "path", "name": "cert" }
  }
}
```

For `GET /api/v1/acme/FIN/C-17`, bind the requested tenant to trusted context.
Within the one endpoint-owned gate, complete-grant evaluation establishes
Finance read authority. The endpoint then restricts the actual data operation
to that tenant AND Finance AND C-17.

| Case | Required consequence |
|---|---|
| C-17 belongs to Finance in the trusted tenant | The boundary can be satisfied; all other authority and application checks still apply. |
| C-17 actually belongs to Engineering | Do not disclose it under Finance-only authority; the path claim is not proof. |
| C-17 belongs to another tenant | Do not disclose it; scope never removes the outer tenant boundary. |
| Department appears only in a log | Not enforcement; the returned data must actually be constrained. |
| Applicable scope is `{}` | No department restriction comes from this grant; tenant isolation and the fixed endpoint input contract still apply. |

A constrained lookup can enforce the relationship directly. A separate
relationship declaration, named resolver, or eager lookup is not required.
Missing/out-of-scope response details are not finalized by this example.

## PUT: selected body material is a proposed value

Request: `PUT /api/v1/acme/certificates/C-17`.

The business body below is sample input, not a newly published contract:

```json
{
  "department_id": "FIN",
  "display_name": "Employment certificate"
}
```

The approved partial endpoint policy is:

```json
{
  "version": "1",
  "method": "PUT",
  "path": "/api/v1/{tenant}/certificates/{cert}",
  "permission": "hrms:employee:certificate::write",
  "inputs": {
    "tenant": { "source": "path", "name": "tenant" },
    "cert": { "source": "path", "name": "cert" },
    "proposed_dept": { "source": "body", "name": "department_id" }
  }
}
```

`proposed_dept` is a local input name, not a reserved scope key. It comes only
from the declared body field. Omission rejects the request, even with a
tenant-wide grant; a query parameter cannot substitute for it.

HRMS owns type, nullability, format, and domain validation. A present value is
not necessarily valid. If HRMS normalizes an identifier, authorization and the
mutation must use the same validated meaning.

**Critical distinction:** requesting Finance does not prove that the existing
certificate is in Finance. The exact source/destination rules for updates and
moves remain open. This example approves input binding, not permission to move
an Engineering certificate into Finance.

## Read the detailed rationale

- [Endpoint policy and review cases](../../docs/endpoint-policy-format.md)
- [Grants and human-dependent authority](../../docs/grant-model.md)
- [Group ownership and optional synchronization](../../docs/groups-and-membership.md)
- [Canonical scope format](../../docs/scope-model.md)
- [More HRMS, Git, ticketing, and accounting cases](../../docs/use-case-examples.md)

The former interactive HRMS explorer has been removed from the active site.
Its [archived source](../../docs/history/retired-hrms-explorer-2026-09-05/README.md)
is historical comparison material, not the specification or implementation evidence.
