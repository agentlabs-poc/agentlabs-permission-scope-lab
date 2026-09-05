# Grant examples

## GRANT-EX-001 — capability and scope remain bound

This JSON illustrates the agreed GRANT-001 concept. Field names, permission
catalog entries, scope encoding, and validity encoding are illustrative, not a
finalized wire schema or a claim about current Auth implementation.

Both grants are evaluated inside an already established, trusted tenant context
T-1 and apply to Maya. Tenant is the outer boundary (TENANT-001), inherited by
each grant rather than repeated as a field. The example validity interval is
September 2026. `conditions: []` means this example declares no additional
conditions; scope and validity remain mandatory parts of the illustrated binding.

```json
[
  {
    "id": "G-1",
    "recipient": { "type": "user", "id": "maya" },
    "permission": "hrms:employee:certificate::read",
    "scope": { "type": "tenant" },
    "status": "active",
    "validity": {
      "not_before": "2026-09-01T00:00:00Z",
      "expires_at": "2026-10-01T00:00:00Z"
    },
    "conditions": []
  },
  {
    "id": "G-2",
    "recipient": { "type": "user", "id": "maya" },
    "permission": "hrms:employee:certificate::revoke",
    "scope": { "type": "department", "id": "FIN" },
    "status": "active",
    "validity": {
      "not_before": "2026-09-01T00:00:00Z",
      "expires_at": "2026-10-01T00:00:00Z"
    },
    "conditions": []
  }
]
```

The outer tenant binding applies to both grants. In this illustrative encoding,
`{ "type": "tenant" }` means the entire enclosing tenant, never an arbitrary
tenant selected inside the grant. G-2's department scope does not permit access
to another tenant's department with the same identifier.

Implicit in the grant payload does not mean absent from the authorization
system. Storage, lookups, caches, and portable representations must preserve the
tenant binding. A detached grant without its required tenant context is not
enough to authorize anything. Physical storage and envelope schemas remain open.

Assuming these are the only grants, Maya has valid tenant membership, the
evaluation occurs within the illustrated validity interval, and all other
required checks succeed:

| Requested action | Target facts | Result supported by these grants |
|---|---|---|
| Read | Certificate in Finance, T-1 | Allowed through G-1 |
| Read | Certificate in Engineering, T-1 | Allowed through G-1 |
| Revoke | Certificate in Finance, T-1 | Allowed through G-2 |
| Revoke | Certificate in Engineering, T-1 | Denied: G-2 does not reach the target |
| Read or revoke | Certificate in T-2 | Denied: outside the grants' tenant boundary |

Resolution must never construct `certificate::revoke @ tenant:T-1` by combining
G-2's capability with G-1's scope. That authority was not granted.

This example does not settle role expansion, combining multiple grants for the
same permission, grant-versus-assignment terminology, or how human-dependent
service/agent delegation is represented.

## GRANT-EX-002 — group recipient

This separate, abbreviated example is also inside trusted tenant context T-1.
It is not added to the grant set used in the outcome table above. Status,
validity, and conditions are omitted here only to focus on the recipient.

```json
{
  "id": "G-3",
  "recipient": { "type": "group", "id": "payroll-readers" },
  "permission": "hrms:employee:certificate::read",
  "scope": { "type": "department", "id": "FIN" }
}
```

The grant applies through the recipient group's human memberships. Resolve the
group and department inside the enclosing tenant. The request still identifies
the acting human or dependent agent/service; group membership does not turn the
group into the authenticated caller. The grant's scope and conditions remain
attached when its authority is made applicable through membership.

## GRANT-EX-003 — multiple permissions and multiple grant sources

This separate example illustrates agreed concepts GRANT-002 and GRANT-003 inside
trusted tenant context T-1. Maya is a valid member of `certificate-operators`.
Only the binding fields are shown; status, validity, and additional conditions
are omitted for brevity, not removed from the grant model.

```json
[
  {
    "id": "G-4",
    "recipient": { "type": "user", "id": "maya" },
    "permissions": [
      "hrms:employee:certificate::read",
      "hrms:employee:certificate::download"
    ],
    "scope": { "type": "tenant" }
  },
  {
    "id": "G-5",
    "recipient": { "type": "group", "id": "certificate-operators" },
    "permissions": [
      "hrms:employee:certificate::read",
      "hrms:employee:certificate::revoke"
    ],
    "scope": { "type": "department", "id": "FIN" }
  }
]
```

Every permission in G-4 has that grant's tenant-wide reach. Every permission in
G-5 has that grant's Finance reach. Shared validity and conditions, when supplied,
likewise apply to every permission within the binding. This array is a proposed
encoding, not a finalized API contract alongside the earlier singular field.

Assuming all required validity and authorization checks succeed:

- Maya has G-4 directly: tenant-wide read and download.
- Maya has G-5 through group membership: Finance-only read and revoke.
- These grants do not authorize revoking an Engineering certificate.
- If Maya leaves `certificate-operators`, she loses G-5's group-derived route.
  G-4's independently valid direct user assignment still applies. This does not
  imply independently authorized service accounts or agents.

Resolution must retain each contributing grant's source. The example does not
specify conflict/deny precedence or the general algebra for combining grants.

## GRANT-EX-004 — reusable role, separately scoped grants

This illustrates agreed concept ROLE-001. The JSON groups a role
definition and two grants for explanation; it is not a finalized API payload.
Tenant context is implicit. Grant status, validity, and conditions are omitted
for brevity and remain part of the binding.

```json
{
  "role": {
    "id": "certificate-reader",
    "name": "Certificate Reader",
    "permissions": [
      "hrms:employee:certificate::read",
      "hrms:employee:certificate::download"
    ]
  },
  "grants": [
    {
      "id": "G-6",
      "recipient": { "type": "user", "id": "maya" },
      "role_id": "certificate-reader",
      "scope": { "type": "department", "id": "FIN" }
    },
    {
      "id": "G-7",
      "recipient": { "type": "group", "id": "auditors" },
      "role_id": "certificate-reader",
      "scope": { "type": "tenant" }
    }
  ]
}
```

The same role supplies read/download capabilities in both grants. G-6 binds
those capabilities to Maya within Finance. G-7 binds them to human members of
Auditors across the enclosing tenant. The role itself creates no access until
an applicable grant binds it to a recipient. Under agreed ROLE-002, authorized
role edits change permissions available through these grants while their scopes,
validity, and conditions stay attached. Freshness mechanics remain open.

## GRANT-EX-005 — expanded role-based grant

This illustrates agreed RESOLUTION-003, using G-6 from the preceding example.
The current Certificate Reader definition contains read and download. Expansion
makes those permissions explicit for evaluation and retains the role source.
It remains grant G-6: this is a computed view of that grant, not a second record
created by assigning new access. The exact computed-view schema is illustrative.

```json
{
  "id": "G-6",
  "recipient": { "type": "user", "id": "maya" },
  "permissions": [
    "hrms:employee:certificate::read",
    "hrms:employee:certificate::download"
  ],
  "scope": { "type": "department", "id": "FIN" },
  "source": {
    "role_id": "certificate-reader"
  }
}
```

Tenant is still the trusted enclosing context. Status, validity, and conditions
are omitted from this abbreviated display but must be preserved from G-6.
Tracking the particular role revision used is required for a reproducible trace;
its representation is not chosen by this example. Such tracking does not pin
future evaluations to an old role version.

An explicitly listed permission grant can use the same evaluation form, with
its own grant source and no role source. This is an expanded view, not a new
independent assignment and not necessarily a fully resolved authorization input.
For example, an employee-self scope would still need its application relationship
established, and the target may still require an application lookup.

## GRANT-EX-006 — one group grant, per-human self scope

This illustrates SELF-001 inside the enclosing trusted tenant. The scope encoding
is illustrative; status, validity, and conditions are omitted for brevity.

```json
{
  "id": "G-10",
  "recipient": { "type": "group", "id": "employees" },
  "permissions": [
    "hrms:payroll:payslip::read"
  ],
  "scope": { "type": "employee_self" }
}
```

Assume Vinay and Maya are valid Employees members. HRMS establishes Vinay maps
to employee E-5 and Maya maps to E-18. Then:

| Evaluated human / actor | Self refers to | Reach supported by G-10 |
|---|---|---|
| Vinay acting for himself | E-5 | Payslips owned by E-5 |
| Maya acting for herself | E-18 | Payslips owned by E-18 |
| Vinay's agent acting under valid delegation | E-5 | E-5 payslips, further limited by that delegation |

The group is the grant recipient, not the reference identity for self. The grant
creator's employee identity is also irrelevant to this selector. Membership
does not allow Vinay to read Maya's payslips through G-10. Missing trusted
user-to-employee information leaves this scope unresolved, never group-wide.

The original grant and membership dependency remain attached to each evaluation;
no independent per-user grant is created.
