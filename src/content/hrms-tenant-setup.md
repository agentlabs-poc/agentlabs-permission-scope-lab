# HRMS example: configuring Payroll Administration

> HISTORICAL SCENARIO — preserved, not a finalized provisioning contract.
> Typed scopes and `tenant_self` below are deprecated under SCOPE-007; current
> examples are in `docs/grant-format.md`. Scope-key registration, bootstrap,
> assignment storage, and grant-administration bounds remain open. An
> administrator need not personally possess the business access they assign
> (ADMIN-002). Do not interpret the historical delegation-ceiling language as
> restoring that rejected prerequisite. Current request handling uses the
> endpoint-owned gate and declared material sources (CONTRACT-006/007).

This example follows the full journey from tenant setup to an authorized payroll-ledger request.

```text
Tenant provisioning
      ↓
HRMS creates departments and employees
      ↓
Application registers permissions and scope types
      ↓
Tenant admin creates Payroll Administrator role
      ↓
Tenant admin assigns the role at a scope
      ↓
Payroll API consumes the resulting authority
```

## 1. Existing tenant and HRMS resources

Auth already knows the tenant and its members:

```text
TENANT-001 · Acme India
├── user:tenant-admin
├── user:maya
├── user:vinay
└── user:arjun
```

HRMS owns the business resources and relationships:

```text
TENANT-001
└── Legal entity INDIA-PVT-LTD
    ├── Department FIN
    │   └── EMP-018 · Maya
    └── Department ENG
        ├── EMP-005 · Vinay
        └── EMP-023 · Arjun
```

These resources are not created by Auth. Auth can refer to them after HRMS validates their identity and tenant ownership.

## 2. HRMS registers its authorization vocabulary

The HRMS application publishes stable permission definitions:

```text
hrms:payroll:head::read
hrms:payroll:head::create
hrms:payroll:salary_earning::read
hrms:payroll:salary_earning::create
hrms:payroll:monthly_input::read
hrms:payroll:monthly_input::create
hrms:payroll:draft::read
hrms:payroll:draft::approve
hrms:payroll:ledger::read
hrms:payroll:ledger::post
hrms:payroll:statutory_liability::read
hrms:payroll:statutory_liability::reconcile
```

It also registers supported scope types:

| Scope type | Registered by | Kind | Resolved when | Contains the target when |
|---|---|---|---|---|
| `tenant_self` | Auth (built-in) | Convenience input | Once, at assignment creation—collapses into `tenant:<id>` immediately, see §4 | *(not stored as-is; see `tenant:<id>`)* |
| `tenant:<id>` | Auth (built-in) | Static | — | target's tenant == `<id>` |
| `legal_entity:<id>` | HRMS | Static | — | target's legal entity == `<id>` |
| `department:<id>` | HRMS | Static | — | target's department == `<id>` |
| `employee:<id>` | HRMS | Static | — | target's owner == `<id>` |
| `employee_self` | HRMS | Dynamic | Every request—no assignment-time shortcut exists | target's owner == principal's resolved employee |
| `resource_exact:<type>:<id>` | Auth (built-in) | Static | — | target's type and id == `<type>` and `<id>`, exactly |

The tenant administrator selects from this vocabulary. They cannot invent a new scope resolver or arbitrary query expression.

`tenant_self` and `employee_self` look like the same idea—"whichever X the current principal has"—but they resolve at different times, and that difference is real, not cosmetic. Every assignment already lives inside exactly one tenant (its own `tenant_id` column), so "the tenant I'm currently acting in" is redundant with a fact that's already true; Auth can safely resolve it once, at assignment creation, and store a plain `tenant:<id>` from then on (§4 shows exactly this: the stored row's `scope_type` is `tenant`, not `tenant_self`). Employee identity carries no such structural shortcut—nothing about the assignment record implies which employee `user:vinay` maps to—so `employee_self` cannot collapse the same way, and stays dynamic, re-evaluated against trusted HRMS context on every single request (§7).

## 3. Tenant admin creates the role

The administrator operates inside an authenticated active tenant:

```text
Active tenant: Acme India · TENANT-001 🔒
```

The role form is:

```text
Role name
[ Payroll Administrator                         ]

Description
[ Operates payroll for the assigned HRMS scope ]

Permissions
☑ hrms:payroll:head::read
☑ hrms:payroll:salary_earning::create
☑ hrms:payroll:monthly_input::create
☑ hrms:payroll:draft::approve
☑ hrms:payroll:ledger::read
☑ hrms:payroll:ledger::post
☑ hrms:payroll:statutory_liability::reconcile

Suggested assignment scope
[ Current tenant ▾ ]
```

The role groups capabilities. It does not give anyone access until it is assigned.

### Records created by the role form

```text
tenant_roles
┌────────────────────┬──────────────────────────────┐
│ id                 │ ROLE-PAYROLL-ADMIN           │
│ tenant_id          │ TENANT-001                   │
│ name               │ Payroll Administrator        │
│ status             │ active                       │
│ created_by         │ user:tenant-admin            │
└────────────────────┴──────────────────────────────┘
```

```text
tenant_role_permissions
┌────────────────────┬────────────────────────────────────────────┐
│ role_id            │ permission                                 │
├────────────────────┼────────────────────────────────────────────┤
│ ROLE-PAYROLL-ADMIN │ hrms:payroll:salary_earning::create        │
│ ROLE-PAYROLL-ADMIN │ hrms:payroll:monthly_input::create         │
│ ROLE-PAYROLL-ADMIN │ hrms:payroll:draft::approve                │
│ ROLE-PAYROLL-ADMIN │ hrms:payroll:ledger::read                   │
│ ROLE-PAYROLL-ADMIN │ hrms:payroll:ledger::post                   │
└────────────────────┴────────────────────────────────────────────┘
```

`TENANT-002` may create a role with the same display name, but it is a separate tenant-owned role.

## 4. Tenant admin assigns the role

The assignment screen is separate:

```text
User
[ Maya Rao · user:maya ▾ ]

Role
[ Payroll Administrator ▾ ]

Scope
● Current tenant
○ Legal entity
○ Department
○ Particular employee

Validity
From [ Today ]  Until [ No expiry ]
```

The UI submits the semantic scope:

```json
{
  "principal_id": "user:maya",
  "role_id": "ROLE-PAYROLL-ADMIN",
  "scope": { "type": "tenant_self" }
}
```

Auth resolves `tenant_self` from trusted session context and stores a concrete boundary:

```text
principal_role_assignments
┌──────────────────┬────────────────────────────┐
│ id               │ ASSIGNMENT-902             │
│ tenant_id        │ TENANT-001                 │
│ principal_id     │ user:maya                  │
│ role_id          │ ROLE-PAYROLL-ADMIN         │
│ status           │ active                     │
│ valid_from       │ 2026-09-03T00:00:00Z       │
│ valid_until      │ null                       │
└──────────────────┴────────────────────────────┘
```

```text
assignment_scopes
┌────────────────────┬────────────────────────────┐
│ assignment_id      │ ASSIGNMENT-902             │
│ tenant_id          │ TENANT-001                 │
│ scope_type         │ tenant                     │
│ resource_type      │ auth:tenant                │
│ resource_id        │ TENANT-001                 │
└────────────────────┴────────────────────────────┘
```

No `$TENANT_ID` appears in a permission string. `tenant_self` is resolved server-side; the concrete tenant is stored for enforcement and audit.

## 5. Preventing cross-tenant assignment

Before creating the assignment, Auth verifies:

```text
✓ grantor may create role assignments
✓ grantor administrative scope contains TENANT-001
✓ role belongs to TENANT-001
✓ Maya is an active member of TENANT-001
✓ requested scope is rooted in TENANT-001
✓ selected permissions are inside the grantor's delegation ceiling
```

The request body does not control `tenant_id`. An administrator authenticated in `TENANT-001` cannot create an assignment rooted in `TENANT-002`.

## 6. A narrower department assignment

The same role can instead be assigned to Maya for Finance:

```json
{
  "principal_id": "user:maya",
  "role_id": "ROLE-PAYROLL-ADMIN",
  "scope": {
    "type": "department",
    "resource_id": "FIN"
  }
}
```

HRMS first verifies that `FIN` belongs to `TENANT-001`. Auth then records the reference and its concrete tenant root.

The role's capabilities do not change. Their reach becomes smaller:

```text
Finance employee payroll       ALLOW
Engineering employee payroll   DENY
Another tenant's payroll       DENY
```

## 7. Employee self-service assignment

The Employee Payroll role can contain:

```text
hrms:payroll:ledger::read
hrms:payroll:payslip::read
```

Vinay receives it with:

```json
{
  "principal_id": "user:vinay",
  "role_id": "ROLE-PAYROLL-EMPLOYEE",
  "scope": { "type": "employee_self" }
}
```

The assignment stores the dynamic scope type. During authorization, HRMS resolves `user:vinay → EMP-005` from its trusted relationship.

## 8. Payroll request consumption

Maya requests Vinay's payroll ledger:

```http
GET /api/payroll/employees/EMP-005/ledger
```

```text
Authenticated principal     user:maya
Active tenant               TENANT-001
Required permission         hrms:payroll:ledger::read
Assignment scope            tenant:TENANT-001
Target tenant               TENANT-001
Target ledger owner         EMP-005
Decision                    ALLOW
```

The Payroll API still applies a tenant predicate:

```sql
WHERE tenant_id = :authenticated_tenant_id
  AND ledger_owner_id = :requested_employee_id
```

Vinay making the same operation with `employee_self` receives the stronger predicate:

```sql
WHERE tenant_id = :authenticated_tenant_id
  AND ledger_owner_id = :authenticated_employee_id
```

## 9. Complete ownership schema

```text
AUTH
├── applications
│   ├── permission_definitions
│   └── scope_type_definitions
├── tenants
│   ├── tenant_principal_memberships
│   ├── tenant_roles
│   │   ├── tenant_role_permissions
│   │   └── role_scope_policy
│   ├── principal_role_assignments
│   │   └── assignment_scopes
│   ├── principal_permission_assignments
│   │   └── assignment_scopes
│   └── authorization_assignment_audit

HRMS
├── employees and user_employee_links
├── legal_entities
├── departments and membership relationships
├── operation_permission_manifest
├── scope resolvers
└── payroll resources carrying
    ├── tenant_id
    ├── legal_entity_id
    ├── department_id
    └── ledger_owner_id
```

The complete effective authority remains:

```text
Principal × Role capabilities × Assignment scope × Trusted payroll target
```
