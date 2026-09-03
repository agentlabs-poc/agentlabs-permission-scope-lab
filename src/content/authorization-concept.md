# Permission, scope, and authority

This document is about **authorization**, not authentication. Authentication already happened by the time any of this runs: a principal has been identified, in a tenant, and that identity is this document's starting input—not something it resolves. What's still open is authorization: given that identified principal, what are they actually allowed to do, and to which data?

HRMS serves many tenants at once, and every tenant has its own employees, managers, and payroll administrators. A payroll ledger is sensitive: an employee should see their own, a payroll administrator should see every employee's—but only inside their own tenant, never another company's. The same shape of problem repeats for code repositories, documents, and every other resource this platform will host. Getting it wrong in either direction is a real failure: too loose leaks another employee's salary or another tenant's data; too strict blocks a payroll administrator from doing their job. A permission string alone can't express that difference—"can read payroll ledgers" says nothing about *whose*—so the rest of this document builds up what has to sit alongside it.

The Authorization Explanation Bench exists to answer a deceptively simple question:

> When we give a user a permission string, what exactly have we authorized them to do—and to which data?

The short answer is that a permission string is only a **capability name**. It is not a complete authority.

```text
Authority = Principal ∩ Permission ∩ Scope ∩ Target
```

All four parts must intersect: the *assigned* scope, not any scope; the *trusted* target, not a caller-asserted one. If any part is absent or cannot be resolved, access fails closed.

- **Principal** — who is asking: a user, a group/team, or a service/agent.
- **Permission** — what capability is being invoked: the stable `<namespaced-noun>::<verb>` string.
- **Scope** — how far the principal's grant reaches: a typed descriptor (`employee_self`, `department:<id>`, `tenant:<id>`, `resource_exact`, `resource_subtree`, …) attached at assignment time, not inside the permission string.
- **Target** — the specific resource instance being requested, plus its trusted attributes (tenant, owner, department/project)—resolved server-side from the id, never asserted by the caller.

The rest of this document follows one request through its eight resolution stages (§1), states who owns each stage (§2), then gives each of stages 2–6 its own detailed section (§3–§7), before closing with reference material and how the vocabulary maps onto industry terms (§13).

## 1. The shape of a request

Suppose Vinay calls:

```http
GET /api/payroll/ledger/me
```

Nothing about this request is self-authorizing. It becomes a decision by passing through eight stages:

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
6. Intersect permission, scope, and target
             ↓
7. Apply the resulting row predicate or deny
             ↓
8. Record the authorization and business audit events
```

Stages 2–6 are where the equation's four terms actually get established:

| Stage | Resolves | Detail |
|---|---|---|
| 2. Map operation to permission | Permission | §3 |
| 3. Load assignments | The grant (permission + scope, as one record) | §4 |
| 4. Resolve scope | Scope | §5 |
| 5. Resolve the resource | Target | §6 |
| 6. Intersect | Authority itself—the containment check | §7 |

Stage 1 (Principal) happens once at authentication and isn't specific to this bench. Stages 7–8 (enforcement, audit) are covered by §2 (who owns them) and §10 (the audit receipt shape).

For Vinay's self grant, stage 7 produces this enforced predicate:

```sql
WHERE tenant_id = :authenticated_tenant_id
  AND ledger_owner_id = :authenticated_employee_id
```

The request cannot broaden this predicate. Supplying another `employee_id` in a URL, body, header, CLI flag, or UI state does not change Vinay's reach—none of stages 1–6 read anything from the caller except which operation and which target id they're naming.

## 2. Who owns which stage

No single system runs all eight stages:

| Stage | Owner |
|---|---|
| 1. Authenticate principal and tenant | Auth |
| 2. Map operation to permission | HRMS Authorization Agent |
| 3. Load active, versioned assignments | Auth |
| 4. Resolve assignment scope | HRMS Authorization Agent |
| 5. Resolve target owner and tenant | HRMS Authorization Agent |
| 6. Intersect permission, scope, target | HRMS Authorization Agent |
| 7. Apply the row predicate | Payroll API |
| 8. Record the audit events | Payroll API |

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

## 3. Naming the operation

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

The permission string should stay stable across users. It does **not** say which employee, department, company, or tenant the caller can reach—stage 2 only names *what kind* of operation this is, nothing about *who* or *which one*.

### Resource-type depth

The grammar permits zero or more subresource segments, to express real resource relationships:

```text
<application>:<domain>:<resource>[:<subresource>...]::<verb>
```

```text
hrms:payroll::read
hrms:payroll:ledger::read
hrms:payroll:ledger:entry::read
hrms:payroll:ledger:entry:attachment::read
```

A name such as `salary_earning` is one resource segment, while `ledger:entry` describes an entry below a ledger. This is depth in the *permission string*—separate from depth in the target instance (§6) or the assignment scope (§5), and the three must not be collapsed into one.

The working safety proposal is that parent capabilities do **not** automatically grant child capabilities:

```text
hrms:payroll:ledger::read
```

does not implicitly grant:

```text
hrms:payroll:ledger:entry::read
```

Any inheritance must be declared explicitly by a role, operation mapping, or wildcard. A proposed one-segment wildcard is:

```text
hrms:payroll:ledger:*::read
```

It would match:

```text
hrms:payroll:ledger:entry::read
hrms:payroll:ledger:summary::read
```

but would not match the deeper:

```text
hrms:payroll:ledger:entry:attachment::read
```

This avoids a shallow wildcard silently gaining authority over deeper resource types introduced later. Parent inheritance and wildcard depth remain design decisions until deliberately agreed.

The same shape names operations in a completely different domain:

```text
agentforge:space::read
agentforge:project::read
agentforge:repository::read
agentforge:repository::write
agentforge:repository::admin
```

and the same hierarchy pattern repeats across products:

```text
HRMS        Tenant → Legal entity → Department → Employee → Payroll ledger
AgentForge  Tenant → Space → Project → Repository → Job
Documents   Tenant → Folder → Subfolder → Document
Commerce    Tenant → Store → Catalogue → Product
```

## 4. Loading the grant

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

### Roles group capabilities; assignments carry reach

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

## 5. Resolving scope

The first payroll scenario uses these scope types:

| Scope | Meaning |
|---|---|
| `employee_self` | The employee linked to the authenticated principal |
| `employee:<id>` | One explicitly identified employee |
| `department:<id>` | Employees belonging to one department |
| `legal_entity:<id>` | Resources belonging to one legal entity |
| `tenant:<id>` | Applicable resources inside one tenant |

Scopes are typed descriptors—not arbitrary SQL expressions. HRMS owns the meaning of HRMS-specific scope types.

`employee_self` is dynamic. It does not mean "use the employee ID supplied in the request." HRMS resolves it from a trusted identity relationship:

```text
authenticated user:vinay
        ↓ trusted mapping
employee EMP-005
```

### Where scopes come from

Three related objects must not be collapsed into one:

```text
Scope type       Created by platform or application designers
Scope target     Created by the owning business application
Assignment scope Created by a tenant administrator
```

An application registers its scope vocabulary when it integrates with Auth. For example, HRMS registers `employee_self`, `department`, and `legal_entity`. Auth can provide generic types such as `tenant_self`, `resource_exact`, and `resource_subtree`.

The owning application creates scopeable resources during normal business operation. HRMS creates departments and employees; AgentForge creates Spaces, Projects, and Repositories. Auth does not create those resources.

The tenant administrator creates an assignment scope by selecting a registered scope type and, where required, an existing target:

```text
Role:       Payroll Administrator
Principal:  user:maya
Scope type: department
Target:     FIN
```

The owning application validates that the target exists and belongs to the authenticated tenant. Auth stores the tenant-rooted reference. A tenant administrator cannot invent arbitrary scope semantics or use knowledge of another tenant's resource ID to grant access to it.

A role's `tenant_self` or `employee_self` value can be a suggested assignment template. It becomes an effective scope only when the role is assigned. `tenant_self` is resolved to the concrete active tenant; `employee_self` remains a dynamic HRMS relationship resolved from trusted identity context.

### Scope depth

This is depth in the *assignment scope*—separate from depth in the permission string (§3) or the target instance (§6). Scope is not necessarily one rigid tree:

```text
tenant:TENANT-001
└── legal_entity:INDIA-PVT-LTD
    └── department:ENG
        └── employee:EMP-005
```

A payroll target can simultaneously carry tenant, legal-entity, department, employee-owner, and other trusted attributes. Authorization checks those dimensions against the assigned scope—which is exactly what §7 does, once the target itself is resolved (§6).

## 6. Resolving the target

This is depth in the *target instance*—separate from depth in the permission string (§3) or the assignment scope (§5). Stage 5 identifies the particular object being accessed:

```text
tenant/TENANT-001/payroll-ledger/PAY-000005/entry/ENTRY-017
```

Resource instance IDs never belong inside the permission string (§3).

The client only ever supplies the id. It does not, and cannot, assert what that id means:

```http
GET /api/payroll/ledger/PAY-000005
```

HRMS looks the id up in its own payroll store and reads back its actual attributes:

```json
{
  "resource_type": "payroll:ledger:entry",
  "resource_id": "ENTRY-017",
  "tenant_id": "TENANT-001",
  "legal_entity_id": "INDIA-PVT-LTD",
  "department_id": "ENG",
  "ledger_owner_id": "EMP-005"
}
```

Nothing in that object came from the request. The client named which record to look up; HRMS supplied what it contains. An id is therefore never secret—knowing `PAY-000023` is not what stops Vinay from reading Arjun's ledger. Stage 6 is what stops it, by checking this resolved target against the assigned scope from §5.

An unknown or unresolvable target id fails closed, the same as an unknown permission or scope (§11).

## 7. Checking containment

Stage 6 asks one question: does the assigned scope, resolved in §5, contain the target, resolved in §6? The rule is specific to each scope type:

| Scope type | Contains the target when |
|---|---|
| `employee_self` | target's owner equals the principal's resolved employee |
| `employee:<id>` | target's owner equals `<id>` |
| `department:<id>` | target's department equals `<id>` |
| `tenant:<id>` | target's tenant equals `<id>` |
| `resource_exact:<type>:<id>` | target's type and id equal `<type>` and `<id>` exactly |
| `resource_subtree:<type>:<id>` | target descends from `<type>:<id>` in the owning application's containment graph |

Only the owning application knows what a given resource descends from—Space = a domain resource, Scope = the reach of an authorization assignment, and the two are not the same thing. Take the AgentForge hierarchy:

```text
Space SPACE-01
├── Project PROJECT-A
│   ├── Repository REPO-API
│   └── Repository REPO-UI
└── Project PROJECT-B
    └── Repository REPO-DOCS
```

Access to exactly one repository uses an exact scope:

```json
{
  "principal": "user:vinay",
  "permission": "agentforge:repository::read",
  "scope": {
    "type": "resource_exact",
    "resource_type": "agentforge:repository",
    "resource_id": "REPO-API"
  }
}
```

This allows `REPO-API`, but not `REPO-UI` or `REPO-DOCS`. Access to every repository inside one project uses a subtree scope instead:

```json
{
  "principal": "user:vinay",
  "permission": "agentforge:repository::read",
  "scope": {
    "type": "resource_subtree",
    "resource_type": "agentforge:project",
    "resource_id": "PROJECT-A"
  }
}
```

| Target | Contained by `PROJECT-A` | Decision |
|---|---:|---:|
| `REPO-API` | Yes | Allow |
| `REPO-UI` | Yes | Allow |
| `REPO-DOCS` | No | Deny |

An assignment covering the complete Space follows the same pattern, one level higher:

```json
{
  "principal": "user:maya",
  "permission": "agentforge:repository::read",
  "scope": {
    "type": "resource_subtree",
    "resource_type": "agentforge:space",
    "resource_id": "SPACE-01"
  }
}
```

When `REPO-API` is requested, AgentForge resolves its trusted target context the same way §6 describes:

```json
{
  "resource_type": "agentforge:repository",
  "resource_id": "REPO-API",
  "tenant_id": "TENANT-001",
  "space_id": "SPACE-01",
  "project_id": "PROJECT-A"
}
```

and containment—the table above—is checked against it:

```text
Permission: agentforge:repository::read              ✓
Containment: PROJECT-A contains REPO-API             ✓
Tenant: target and principal are in TENANT-001       ✓
Decision                                              ALLOW
```

The owning application—not generic Auth—resolves whether a repository belongs to a project or Space.

Containment only ever broadens *where* a granted capability applies; it never invents a capability that was not explicitly granted. Granting:

```text
agentforge:project::read
```

does not create this different capability:

```text
agentforge:repository::read
```

A reusable role can explicitly contain both:

```text
Project Viewer
├── agentforge:project::read
└── agentforge:repository::read
```

Assigning that role at `resource_subtree:PROJECT-A` gives both declared capabilities reach inside Project A—the scope broadens where they apply, exactly as §3's wildcard rule says a parent permission never implicitly creates a child one:

> A parent scope can extend the reach of an explicitly granted capability, but a parent capability does not implicitly create child capabilities.

## 8. Example decisions

| Principal | Permission | Scope | Target | Decision |
|---|---|---|---|---|
| Vinay | ledger read | employee self | Vinay's ledger | Allow |
| Vinay | ledger read | employee self | Arjun's ledger | Deny |
| Maya | ledger read | tenant 001 | Vinay's ledger in tenant 001 | Allow |
| Maya | ledger read | tenant 001 | Neha's ledger in tenant 002 | Deny |
| Vinay | ledger post | employee self | Vinay's ledger | Deny: capability absent |
| PayBot | ledger read | no delegation | Vinay's ledger | Deny: no complete grant |

## 9. Delegated jobs and agents

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

This is deliberately still a question in the bench. The worked examples must help us test and finalize it before implementation.

## 10. Audit receipt

A useful decision record—stage 8's output—contains the whole explanation:

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

## 11. Failure rules

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

## 12. What the explanation bench is for

The bench is not merely a static document and is not a production authorization engine. It is an executable design instrument:

- create assignments;
- target real-looking domain resources;
- predict allow or deny;
- inspect the explanation and enforced predicate;
- attack the proposed model with adversarial scenarios;
- preserve unresolved questions;
- move candidate answers from `OPEN` to `PROPOSED` to `AGREED`;
- export the current design session.

Payroll is the first interactive scenario pack. Later packs can cover employee documents, leave, expenses, recruiting, AgentForge jobs, or other products while keeping the same authorization equation.

## 13. Canonical terms across the industry

Nothing in this vocabulary is invented in isolation. Every term maps onto an established access-control concept, most formally onto NIST/ANSI **RBAC** (Role-Based Access Control, INCITS 359-2004):

| Term used here | NIST RBAC | What it means |
|---|---|---|
| Principal | User | Anything that authenticates and can hold a grant—a person, a group, or a service/agent |
| Permission | Permission (operation + object) | A stable capability name; the `<namespaced-noun>::<verb>` string |
| Role | Role | A named bundle of permissions. Carries no scope, no target, no principal |
| Grant / assignment | User assignment (UA) | The binding `principal + role/permission + scope + validity` that gives a role reach |
| Resource | Protected object class | The *type* named inside the permission string, e.g. `ledger`, `repository` |
| Target | Protected object instance | The specific *instance* being requested, plus its trusted attributes (tenant, owner, project) |
| Scope | Not in base RBAC; closest is ABAC's attribute constraints | The reach a grant covers over target instances |
| Group / team | Group | A named set of principals; an assignment made to a group is inherited by every member |

A role is only ever the second row plus the third—the same split §4 already draws between what a role says and what its assignment says. The detail worth adding here: the assignment's principal can be an individual or a group, and the role is redefined for neither. "Project Viewer" assigned to Maya's team at one scope is the same role as one assigned to a single user at another.

### Where real platforms diverge

"Role" and "scope" do not mean the same thing everywhere, and it is worth knowing the differences before this model meets an external system:

- **Kubernetes RBAC** matches this bench closely: a `Role` / `ClusterRole` is a pure permission bundle with no principal or target attached. A separate `RoleBinding` / `ClusterRoleBinding` supplies `{subject, role, namespace}`—structurally identical to this bench's `Grant`.
- **Azure RBAC** matches even more closely: a Role Definition is the permission bundle; a Role Assignment is `{principal, role, scope}`, and Azure's `scope` is literally a resource-hierarchy path (`/subscriptions/.../resourceGroups/...`)—the same "reach over a resource tree" idea as this bench's `department:<id>`, `tenant:<id>`, and `resource_subtree` scopes.
- **AWS IAM diverges.** An AWS "Role" is not a permission bundle—it is an *assumable identity* (a special kind of principal). The actual permission bundle in AWS is called a **Policy**, attached to a Role, User, or Group. Anyone arriving from an AWS background will not mean the same thing by "role" that this bench does.
- **OAuth 2.0 already owns the word "scope"**, and means something different by it: an OAuth scope such as `read:contacts` is closer to this bench's *permission* than to its *reach*. If this system ever issues or accepts OAuth tokens, "scope" becomes ambiguous between the two meanings in the same sentence—worth a disambiguating term (this bench's equation already says `reach`) before that becomes a real integration surface rather than bench vocabulary.

## 14. Questions still open

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
11. Do parent permissions imply child-resource permissions?
12. Should wildcards match exactly one resource segment or support recursive depth?

These are not implementation details. Their answers determine the final Auth/Authz shape, so they remain visible until deliberately agreed.
