# Permission, scope, and authority

Permission Quest exists to answer a deceptively simple question:

> When we give a user a permission string, what exactly have we authorized them to do—and to which data?

The short answer is that a permission string is only a **capability name**. It is not a complete authority.

```text
Authority = Principal × Capability × Assigned Scope × Trusted Target Context
```

All four parts must intersect. If any part is absent or cannot be resolved, access fails closed.

## 1. The permission string

The canonical grammar is:

```text
<namespaced-noun>::<verb>
```

Payroll examples are:

```text
hrms:payroll:ledger::read
hrms:payroll:ledger::post
hrms:payroll:salary_earning::create
```

Everything before `::` identifies the noun. Everything after it identifies the operation.

The namespace prevents unrelated applications from accidentally using the same noun:

```text
hrms       → application
payroll    → domain
ledger     → resource
read       → operation
```

The permission string should stay stable across users. It does **not** say which employee, department, company, or tenant the caller can reach.

## 2. A grant is more than a permission

Assigning only this string is incomplete for scoped business data:

```text
hrms:payroll:ledger::read
```

The effective assignment must carry:

```text
principal + role/permission + scope + validity
```

An employee assignment might be:

```json
{
  "principal": "user:vinay",
  "permission": "hrms:payroll:ledger::read",
  "scope": { "type": "employee_self" },
  "status": "active"
}
```

A payroll administrator can receive the same capability at a wider scope:

```json
{
  "principal": "user:maya",
  "permission": "hrms:payroll:ledger::read",
  "scope": {
    "type": "tenant",
    "id": "TENANT-001"
  },
  "status": "active"
}
```

The capability is identical; the reach is different.

## 3. Scope vocabulary

The first payroll scenario uses these scope types:

| Scope | Meaning |
|---|---|
| `employee_self` | The employee linked to the authenticated principal |
| `employee:<id>` | One explicitly identified employee |
| `department:<id>` | Employees belonging to one department |
| `legal_entity:<id>` | Resources belonging to one legal entity |
| `tenant:<id>` | Applicable resources inside one tenant |

Scopes are typed descriptors—not arbitrary SQL expressions. HRMS owns the meaning of HRMS-specific scope types.

`employee_self` is dynamic. It does not mean “use the employee ID supplied in the request.” HRMS resolves it from a trusted identity relationship:

```text
authenticated user:vinay
        ↓ trusted mapping
employee EMP-005
```

## 4. Roles and assignments

A role groups capabilities:

```text
Employee
├── hrms:payroll:ledger::read
└── hrms:payroll:payslip::read

Payroll administrator
├── hrms:payroll:ledger::read
├── hrms:payroll:salary_earning::create
└── hrms:payroll:ledger::post
```

The user-to-role assignment carries the reach:

```text
Vinay ── Employee role ── employee_self
Maya  ── Payroll administrator role ── tenant:TENANT-001
```

This separation lets the role remain reusable. It also makes privilege review clearer: the role says **what**, while the assignment says **where**.

A direct permission grant should use the same scoped-assignment shape. It must not bypass scope simply because it was assigned directly.

## 5. Request consumption

Suppose Vinay calls:

```http
GET /api/payroll/ledger/me
```

The authorization flow is:

```text
1. Authenticate principal and tenant
             ↓
2. Map API operation to required permission
             ↓
3. Load active, versioned assignments
             ↓
4. Resolve assignment scope using trusted HRMS context
             ↓
5. Resolve the requested resource's owner and tenant
             ↓
6. Intersect capability, scope, and target
             ↓
7. Apply the resulting row predicate or deny
             ↓
8. Record the authorization and business audit events
```

For Vinay's self grant, the enforced predicate is effectively:

```sql
WHERE tenant_id = :authenticated_tenant_id
  AND ledger_owner_id = :authenticated_employee_id
```

The request cannot broaden this predicate. Supplying another `employee_id` in a URL, body, header, CLI flag, or UI state does not change Vinay's reach.

## 6. Auth and authorization-agent boundary

The working boundary is:

### Auth

Auth is generic. It knows:

- the authenticated principal;
- roles and direct assignments;
- opaque permission strings;
- the assignment's scope descriptor;
- validity windows and status;
- assignment versions and invalidation.

Auth does not need to understand payroll ledgers or employee ownership.

### HRMS Authorization Agent

The HRMS Authorization Agent knows:

- which permission an HRMS operation requires;
- what `employee_self`, department, and legal-entity scopes mean;
- how a user maps to an employee;
- the trusted attributes of the requested payroll resource;
- whether scope contains the target;
- which mandatory data predicate follows from the decision.

### Payroll API

The API is an enforcement point. It:

- asks for an authorization decision;
- applies the returned tenant and ownership restrictions;
- never trusts caller-provided authority context;
- records the business access event;
- returns no data when the decision or scope cannot be resolved.

The UI and CLI do not create authority. They call the same protected APIs.

## 7. Example decisions

| Principal | Capability | Assigned scope | Target | Decision |
|---|---|---|---|---|
| Vinay | ledger read | employee self | Vinay's ledger | Allow |
| Vinay | ledger read | employee self | Arjun's ledger | Deny |
| Maya | ledger read | tenant 001 | Vinay's ledger in tenant 001 | Allow |
| Maya | ledger read | tenant 001 | Neha's ledger in tenant 002 | Deny |
| Vinay | ledger post | employee self | Vinay's ledger | Deny: capability absent |
| PayBot | ledger read | no delegation | Vinay's ledger | Deny: no complete grant |

## 8. Delegated jobs and agents

Device login authenticates the human on whose behalf a job runs. Authentication alone must not give the job broader authority.

The unresolved design question is how delegation is represented. The minimum information is likely:

```text
human principal
delegated runner identity
effective permission assignment
effective scope
delegation purpose
validity window
correlation/audit reference
```

The runner's effective authority must be no greater than both the user's grant and the delegation. In set terms:

```text
runner authority = user authority ∩ delegated authority
```

This is deliberately still a question in the lab. The game must help us test and finalize it before implementation.

## 9. Audit receipt

A useful decision record contains the whole explanation:

```json
{
  "principal": "user:vinay",
  "permission": "hrms:payroll:ledger::read",
  "assignment_id": "grant-983",
  "scope": "employee_self",
  "resolved_employee_id": "EMP-005",
  "tenant_id": "TENANT-001",
  "resource": "payroll-ledger:PAY-000005",
  "decision": "allow",
  "policy_version": "v17"
}
```

This record explains which assignment and policy version produced the decision. It must avoid copying sensitive payroll values unnecessarily.

## 10. Failure rules

The working rules are:

1. A permission without a required scope does not authorize scoped data.
2. An unknown permission mapping fails closed.
3. An unknown or unresolvable scope type fails closed.
4. Tenant isolation applies even when another scope matches.
5. `self` is derived from trusted identity relationships.
6. The caller cannot supply or widen their effective scope.
7. UI visibility is not API authorization.
8. A delegated runner cannot exceed the delegating user's authority.
9. Every decision identifies the assignment and policy version used.
10. Broad wildcard grants require deliberate treatment because future operations may otherwise become reachable automatically.

## 11. What the game is for

The game is not merely a visual explanation and is not a production authorization engine. It is an executable design instrument:

- create assignments;
- target real-looking domain resources;
- predict allow or deny;
- inspect the explanation and enforced predicate;
- attack the proposed model with adversarial scenarios;
- preserve unresolved questions;
- move candidate answers from `OPEN` to `PROPOSED` to `AGREED`;
- export the current design session.

Payroll is the first scenario pack. Later packs can cover employee documents, leave, expenses, recruiting, AgentForge jobs, or other products while keeping the same authorization equation.

## 12. Questions still open

The initial question board contains:

1. Should assignment scope live in Auth or an HRMS governance binding?
2. How are dynamic scopes such as `employee_self` represented canonically?
3. Can roles carry a default scope, or only assignments?
4. Can a direct grant narrow or widen a role assignment?
5. How are department changes reflected in scope evaluation?
6. How does device-login delegation constrain a runner?
7. Do explicit deny assignments override allows?
8. What belongs in access tokens versus server-side resolution?
9. How are changed assignments invalidated?
10. What is the canonical audit receipt?

These are not implementation details. Their answers determine the final Auth/Authz shape, so they remain visible until deliberately agreed.
