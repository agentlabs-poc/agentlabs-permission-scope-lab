# Authorization Explanation Bench — working design

> TERM-005 / Q-043 additionally deprecates the older authorization vocabulary
> used below. Preserve this lab design as history; the current explanation is
> [permissions, boundaries, and request material](authorization-vocabulary.md).

> HISTORICAL LAB DESIGN — preserved, not the current canonical contract.
> Scope syntax below is deprecated under SCOPE-007. The current model uses one
> endpoint-owned gate (CONTRACT-006), explicit permission/material/source
> declarations (CONTRACT-007), and human-dependent service/agent authority
> (AUTHORITY-002). Identity terminology, permission grammar, registration,
> schemas, and audit mechanisms below are not silently adopted by this notice.
> See the [working handbook](handbook.md) and [reconciliation register](reconciliation.md).

The Authorization Explanation Bench is an executable design conversation. Markdown pages carry the explanations; payroll is the first interactive scenario pack. The authorization mechanics must remain reusable by other products and domains.

## Working equation

```text
authority = principal ∩ permission ∩ scope ∩ target
```

- **principal** — who is asking: a user, a group/team, or a service/agent.
- **permission** — the stable `<namespaced-noun>::<verb>` capability string.
- **scope** — the reach a grant covers, assigned at grant time (not inside the permission string).
- **target** — the specific resource instance, plus its trusted attributes, resolved server-side—never asserted by the caller.

Each part must intersect; an absent or unresolved part fails closed.

The permission grammar is:

```text
<application>:<domain>:<resource>[:<subresource>...]::<verb>
```

For example:

```text
hrms:payroll:ledger::read
```

The string identifies a stable capability. It does not contain the user, tenant, employee, department, validity window, or target record.

Resource depth, target-instance depth, and assignment-scope depth are separate. Resource instance IDs never enter a permission string. The current safety proposal gives parent permissions no implicit authority over child resources and makes wildcard depth explicit; this remains open until agreed through the bench.

An effective assignment is therefore:

```text
principal + permission/role + scope + validity
```

A permission without an evaluable scope is incomplete for scoped data and must fail closed.

## Consumption

1. Auth authenticates the principal and supplies active, versioned grants.
2. HRMS maps the API operation to its required permission.
3. HRMS resolves dynamic scope from trusted relationships, such as the authenticated user-to-employee link.
4. HRMS resolves the target resource's tenant and ownership attributes.
5. The authorization agent intersects capability, scope and target.
6. The API applies the resulting tenant and row predicate.
7. The decision and business access are audited.

Caller-controlled values may identify a requested target, but they cannot establish the caller's scope or trusted identity relationship.

## Current scope vocabulary

> DEPRECATED VOCABULARY/FORMAT — "current" in this historical heading refers
> to the earlier lab. Canonical v1 scope is the flat key-value format in
> [scope and target](scope-model.md), not the typed descriptors below.

- `employee_self`: dynamic; HRMS resolves the authenticated user's employee.
- `employee:<id>`: one employee.
- `department:<id>`: employees belonging to one department.
- `tenant:<id>`: all applicable resources inside one tenant.

Scope values are typed descriptors. They are not arbitrary user-supplied SQL expressions.

## Responsibility boundary under discussion

### Auth

- principal identity;
- roles and direct grants;
- opaque permission strings;
- assignment scope descriptor;
- status and validity;
- assignment version and invalidation.

### HRMS Authorization Agent

- API operation-to-permission mapping;
- HRMS scope vocabulary and resolution;
- user-to-employee relationship;
- target resource attributes;
- authorization decision and enforced predicate.

### Payroll API

- invokes authorization;
- applies the returned restriction;
- never trusts a caller-provided scope;
- emits the business audit record;
- fails closed when a decision or scope cannot be resolved.

## The bench as a decision tool

Each worked request combines a principal, requested permission, resource target, and expected decision. The guided explorers explain the request story, permission anatomy, role/direct assignment, scope origin, trusted target context, containment, enforcement predicate, and audit receipt in sequence. Separate HRMS and Projects/Repositories paths demonstrate individual and group principals, dynamic and concrete scopes, exact resources, subtrees, and tenant isolation.

The design board deliberately distinguishes `OPEN`, `PROPOSED`, and `AGREED`. Its JSON export captures current grants, explored scenarios, and question statuses. Export is a discussion artifact, not a production policy format.

## Extension seam

Scenario data lives in `src/scenario-pack.ts`. A pack declares principals, permissions, resources, initial grants, and worked requests. The evaluator consumes that pack. Payroll-specific display fields still exist in this first spike and should become generic typed resource attributes when a second interactive domain is added.

## Open design questions

The live board begins with these questions:

- Should assignment scope live in Auth or an HRMS governance binding?
- How are dynamic scopes represented canonically?
- Can roles carry a default scope, or only assignments?
- How does device-login delegation constrain a runner to the user's authority?
- What enters the access token and what stays server-side?
- Do explicit deny assignments override every allow?

No answer is canonical until it is deliberately agreed and moved into the implementation design.
