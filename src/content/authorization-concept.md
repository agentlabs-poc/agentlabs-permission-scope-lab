# The Authorization Handbook

> HISTORICAL LAB HANDBOOK — preserved for reference. This is not the current
> canonical working edition. Current chapters are in `docs/handbook.md`;
> the logical diagram is in `docs/system-overview.md`.
>
> Current decisions: scope is a boundary selector, encoded as a required flat
> key-value object (SCOPE-006/007); entries combine with AND and alternative
> grants provide OR (SCOPE-008). Explicit `{}` is tenant-wide; missing/null
> scope is invalid. Authorization uses one endpoint-owned gate without a
> prepared handoff (CONTRACT-006). Each endpoint declares required permission,
> material, and sources, including identified path/body inputs (CONTRACT-007).
> Services and agents remain human-dependent (AUTHORITY-002).
>
> Earlier typed scopes, request/receipt schemas, registration models, identity
> terminology, external comparisons, and implementation claims below remain
> historical or unverified where not explicitly agreed in the decision log.
> The interactive evaluator has not been migrated to the current handbook.

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

### What "the request" actually contains

> HISTORICAL INPUT EXAMPLE — any restriction below to method/path/token alone
> is not the current contract. INPUT-001 and CONTRACT-007 include explicitly
> identified body inputs and declared material sources. Inputs still do not
> automatically prove existing target relationships.

It is easy to smuggle resolved fields back into the request and call them caller-supplied. The caller contributes exactly three things, nothing else, ever:

```json
{
  "verb": "GET",
  "path": "/api/payroll/ledger/me",
  "token": "<bearer at+jwt>"
}
```

Everything a decision needs beyond this is *derived*, not supplied—each field traces to exactly one of the three above, never to the caller directly:

```json
{
  "principal": { "id": "user:vinay", "tenant_id": "TENANT-001" },
  "operation_id": "hrms.payroll.ledger.read",
  "permission": "hrms:payroll:ledger::read",
  "trusted_context": {},
  "grant": { "permission": "hrms:payroll:ledger::read", "scope": { "type": "employee_self" } },
  "resolved_context": { "employee_self": "EMP-005" }
}
```

- `principal` comes from `token` (Auth verifies the signature, returns identity).
- `operation_id` and `permission` come from `verb` + `path` together, through the manifest—a fixed table, no data lookup.
- `trusted_context` comes from `path` again, but a different part of it: whichever path segment the manifest's own binding config names as trusted (empty here—`/me` carries no id of its own; see §6 for a route that does).
- `grant` comes from `principal`, loaded from Auth (§4).
- `resolved_context` comes from `principal` again, via the one relationship lookup dynamic scope needs (§5's "Static vs. dynamic scope").

The decision at §7 is then a pure comparison inside the second object—nothing new enters after this point. A "client request" object that includes `permission` or a resolved employee id has already skipped a step; that field was never sent, it was computed.

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

> HISTORICAL LAYOUT — current equivalents are in `docs/grant-format.md`.
> Grant/role/binding semantics remain relevant; typed scope objects and the
> exact stored/expanded payloads below are not the canonical v1 scope format.

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

> DEPRECATED SCOPE REPRESENTATION — retain the following explanation as history.
> SCOPE-007 now uses flat string-valued boundary keys, the reserved `$self`
> human reference where supported, and explicit `{}` for tenant-wide reach.
> Scope-key governance and definition lifecycle remain open; the scope-type
> catalogs, `tenant_self` conversion, and registration rules below are not adopted.

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

Lined up against permission registration (§3), there are four different registration granularities here, easy to conflate:

| Thing | Registered with Auth? | When |
|---|---|---|
| Permission string (`hrms:payroll:ledger::read`) | Yes—full catalog | Once, at integration/manifest-publish time |
| Scope *type* (`department` as a concept) | Yes—just the type name | Once, at integration time |
| Scope *target* (a specific department, `FIN`) | No | Never bulk-registered |
| One assignment referencing that target (`Maya @ department:FIN`) | Yes—but as one grant record, not a catalog entry | Per-assignment, validated live |

Permission strings and scope types are both small, static, one-time-registered catalogs. Scope targets are the opposite—an open-ended, constantly-changing set the owning application manages entirely on its own, that Auth never mirrors. An assignment is the only place a specific target and Auth ever meet, and even then only as a single reference inside one grant record, not as an addition to any catalog.

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

This validation step only applies when the scope names a concrete target—`employee:<id>`, `department:<id>`, `legal_entity:<id>`, `tenant:<id>`, `resource_exact`, `resource_subtree`. `employee_self` and `tenant_self` carry no target field at all: there is nothing for the tenant administrator to select, nothing for the owning application to confirm exists, because the scope names a relationship, not an instance—so neither is registered as a fixed target the way the six types above are. What they resolve to is not identical in timing, though (see "Static vs. dynamic scope" below): `tenant_self` is inferred once, immediately at assignment creation, and stored from then on as a plain `tenant:<id>`; `employee_self` is inferred fresh at decision time, on every request, with no assignment-time equivalent.

A role's `tenant_self` or `employee_self` value can be a suggested assignment template. It becomes an effective scope only when the role is assigned. `tenant_self` is resolved to the concrete active tenant; `employee_self` remains a dynamic HRMS relationship resolved from trusted identity context.

### Three axes for choosing a scope type

The scope types above aren't one flat list—they differ along three independent axes, and conflating them is the most common source of confusion when designing a new one.

**Axis 1 — who owns the vocabulary, i.e. who has to register the type at all:**

| Category | Types | Who registers it |
|---|---|---|
| Application-specific | `employee_self`, `employee`, `department`, `legal_entity` | The owning app (HRMS)—has to teach Auth this word exists, once, at integration time |
| Auth-generic, built-in | `tenant_self`, `resource_exact`, `resource_subtree` | Nobody—every application gets these for free, cross-application |

**Axis 2 — static value vs. dynamic relationship, i.e. whether containment needs a live lookup:**

| Kind | Types | How containment resolves |
|---|---|---|
| **Static** — a fixed value baked into the grant | `employee:<id>`, `department:<id>`, `legal_entity:<id>`, `tenant:<id>`, `resource_exact`, `resource_subtree` | Compare the fixed value against the target's own attribute. No principal-side lookup needed. |
| **Dynamic at assignment time only** — resolved once, then stored as a plain static value | `tenant_self` | Collapses into `tenant:<id>` the moment the assignment is created (every assignment already lives inside one tenant, so there's nothing left to resolve after); behaves exactly like the static row from then on. |
| **Dynamic at request time, always** — a placeholder meaning "resolve against whoever's asking, right now" | `employee_self` | Requires resolving the principal first (`$self` → which employee holds this token), fresh on every request—the one case that genuinely cannot be answered from the grant and request alone, ever. |

`tenant_self` is not merely cheap to resolve—it does not need resolving at request time at all. Every assignment already lives inside exactly one tenant, so "the tenant I'm currently acting in" is redundant with a fact the assignment record already carries; it collapses into a plain `tenant:<id>` once, at assignment-creation time, and is never re-evaluated after. `employee_self` has no such shortcut: nothing about an assignment record implies which employee a principal maps to, so it stays genuinely dynamic, re-resolved on every single request—employee identity is never in the token (§1, "what is deliberately not in the token"), so this is a real query every time, distinct from resolving the target itself (§6).

**Axis 3 — reach, narrowest to broadest.** This is *why* so many types exist at all: each trades precision against how much it auto-covers as new resources get created, with no assignment change:

```text
resource_exact:<id>            narrowest — one instance, forever, never auto-expands
employee_self / employee:<id>  one person's records — auto-expands to their future records only if "self"
department:<id>                a group — auto-expands as people join or leave the department
legal_entity:<id>               wider group
tenant:<id>                     everything in the tenant
resource_subtree:<type>:<id>    everything under one node — breadth depends entirely on which node is chosen
```

A grant using `resource_exact` never needs re-granting for that one resource, but also never covers a new one. A grant using `employee_self` never needs re-granting *ever*, for any future resource that comes to belong to that principal—the tradeoff explored fully in the payroll self-service example above (§5, "Where scopes come from").

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

### Whether resolving the target needs a lookup depends on the route, not the model

The example above (`PAY-000005` → look it up → learn it's `ENTRY-017`, owned by `EMP-005`) needs a real store read because the id in the path is *opaque*—nothing about the string `PAY-000005` reveals who owns it. That lookup is not a property of stage 5 in general; it is a property of *this route's own design*.

A route can instead bind the ownership attribute directly into the path:

```http
GET /api/payroll/employees/EMP-005/ledger
```

Here the target *is* `EMP-005`, read straight from the path—no store read at all. Stage 5 becomes a pure string read, and stage 6 compares it directly against whatever `employee_self` resolved to in §5. Confirmed against the real system's own manifest, one route at a time: a route bound on the ownership attribute needs "nothing... looking up"; a route bound on an opaque object id "can only be matched by `*_ids` scopes... today," and closing that gap needs a purpose-built context enricher between resolving the operation and evaluating the policy—work that exists as a named, not-yet-built change, not as something this model already does automatically.

So: opaque ids buy a stable, meaningless-until-resolved identifier at the cost of a lookup on every request. Attribute-bound ids buy a lookup-free stage 5 at the cost of exposing that attribute in every URL for that route, forever. Neither is "more correct"—it is a route-design choice made once, per API, with a lasting cost either way.

### Route design decides mechanism, not security

Neither shape is safer than the other. Both are exposed to the same real vulnerability class if stage 6 is ever skipped—OWASP calls it **API1:2023, Broken Object Level Authorization (BOLA)**, and it is the most common real-world API vulnerability, precisely because each route shape invites a different version of the identical mistake:

- An attribute-bound route tempts trusting the path value *as if naming it were authorization*—"the URL says `EMP-005`, so serve `EMP-005`"—skipping the comparison against what scope actually resolved. That is BOLA in zero extra lines of code.
- An opaque-id route tempts doing the lookup, getting a real row back, and returning it—because it resolved, it must be fine—without ever checking whether the resolved owner matches the caller. That is BOLA one step later.

Same root failure; route design only decides where in the code the missing check would go. What actually determines security quality is unconditional discipline: does *every* request run stage 6 (§7) before any data returns, no matter how stage 5 resolved the target? That is an implementation question, orthogonal to URL shape—not something a route-design choice can buy or sell away.

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

> HISTORICAL RECEIPT EXAMPLE — audit accountability is required for explicit
> access-changing self-assignment (ADMIN-003), but this receipt's exact schema,
> persistence, failure behavior, and integrity mechanisms are not finalized.

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

### Auth is mechanism, not policy

Every row in the table above is something Auth stores and returns without interpreting: permission strings, scope, grants, and—if teams are ever added—team membership are all opaque custody data. Auth's own architecture doc says this outright, describing what an Auth Agent may do with a resolved bundle: *"Another application may use relationship-based access control, ordered rules, ACLs, or a policy engine such as Cedar, Rego, or Zanzibar-style tuples. Auth stores the artifacts without imposing one resolution model."* Nothing about *how* to resolve any of it is Auth's decision, for any application, ever.

This is the classic **mechanism, not policy** split from systems design—the same principle behind a microkernel exposing IPC and scheduling while device semantics live in userspace, or TCP moving bytes with no opinion about what they mean. Auth is Layer 1: generic, dumb, reusable across every application. Each application's own Auth Agent is Layer 2: the place all the domain-specific complexity actually lives—HRMS's `employee_self` relationship resolver, code-hosting's containment graph, or nothing at all for a low-stakes application that just needs a flat allow-list.

The real cost of keeping Layer 1 this thin: nothing gets solved once, centrally, for every application. It gets solved N times, once per Auth Agent, at whatever quality and risk tolerance that application's own team chooses to invest. That is also the honest answer to whether Auth should validate a scope target's existence (§5, "Where scopes come from")—validation is a Layer 2 decision, not a Layer 1 one, and Auth staying silent on it is the same choice Zanzibar's own designers made deliberately, not a gap this bench happened to inherit by accident.

### Where real platforms diverge

"Role" and "scope" do not mean the same thing everywhere, and it is worth knowing the differences before this model meets an external system:

- **Kubernetes RBAC** matches this bench closely: a `Role` / `ClusterRole` is a pure permission bundle with no principal or target attached. A separate `RoleBinding` / `ClusterRoleBinding` supplies `{subject, role, namespace}`—structurally identical to this bench's `Grant`.
- **Azure RBAC** matches even more closely: a Role Definition is the permission bundle; a Role Assignment is `{principal, role, scope}`, and Azure's `scope` is literally a resource-hierarchy path (`/subscriptions/.../resourceGroups/...`)—the same "reach over a resource tree" idea as this bench's `department:<id>`, `tenant:<id>`, and `resource_subtree` scopes.
- **AWS IAM diverges.** An AWS "Role" is not a permission bundle—it is an *assumable identity* (a special kind of principal). The actual permission bundle in AWS is called a **Policy**, attached to a Role, User, or Group. Anyone arriving from an AWS background will not mean the same thing by "role" that this bench does.
- **OAuth 2.0 already owns the word "scope"**, and means something different by it: an OAuth scope such as `read:contacts` is closer to this bench's *permission* than to its *reach*. If this system ever issues or accepts OAuth tokens, "scope" becomes ambiguous between the two meanings in the same sentence—worth a disambiguating term (this bench's equation already says `reach`) before that becomes a real integration surface rather than bench vocabulary.
- **Microsoft Entra ID (Azure AD) issues two different claims depending on the Principal**, and neither is scope in this bench's sense. A signed-in user acting through a client app gets a `scp` claim (space-delimited delegated-permission strings, e.g. `"scp": "User.Read Mail.Send"`); an app acting with no user gets a `roles` claim instead (an array of app-role values). Both are this bench's *permission*, not its *reach*—Microsoft folds coarse reach into the permission string itself by naming convention (`User.Read` = self, `User.Read.All` = tenant-wide), rather than keeping it a separate bound attribute the way this bench's `scope` does. Entra ID also has no equivalent to `employee_self`: whether a specific mailbox or file belongs to the caller is resolved entirely by the resource provider (Exchange, SharePoint)—the same split this bench draws between Auth (stops at the permission/scope claim) and the owning application (resolves the target, §6).

### Scope-reach: two real paradigms, not five types

No real system defines this bench's five scope types as a matching, named vocabulary—that list is this bench's own invention. Industry practice instead converges on two different unifying mechanisms:

**Hierarchical resource scoping** collapses `tenant`, `department`, `resource_exact`, and `resource_subtree` into one mechanism: reach is a path prefix into a resource tree, not four separate typed descriptors. Azure RBAC's `scope` is literally a path (`/subscriptions/{s}/resourceGroups/{rg}/...`); GCP IAM inherits down an Organization → Folder → Project → Resource hierarchy; Kubernetes RBAC scopes a `Role` to one namespace or a `ClusterRole` to none; AWS IAM matches a resource ARN exactly or with a wildcard. None of these needs a separate "exact" type versus a "subtree" type—breadth falls out of how much of the path or ARN is specified.

**Relationship-based access control (ReBAC)**—Google's Zanzibar model, and its open implementations SpiceDB, OpenFGA, and Ory Keto—goes further and folds in `employee_self` and `department` too. Reach is expressed as relationship tuples (`document:PAY-000005#owner@user:vinay`, `repo:REPO-API#viewer@group:TEAM-BACKEND#member`) and userset rewrite rules (a folder's `viewer` set includes anyone who is a `viewer` on its parent). One graph-traversal mechanism covers everything this bench splits into five named types; `employee_self` stops being a special case and becomes just the relationship where the subject is the resource's own `owner`.

**Most conventional IAM systems have no first-class equivalent to `employee_self` at all.** AWS approximates it with a policy-variable trick—substituting `${aws:username}` into a resource ARN, so "your own resources" only works when the ARN already contains the username as a literal string—narrower than a true relationship, and nothing like it exists natively in Azure RBAC, GCP IAM, or Kubernetes RBAC. This gap is exactly why Zanzibar-style ReBAC exists: hierarchical path-scoping never solved "whichever resource happens to belong to whoever is asking" natively.

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
