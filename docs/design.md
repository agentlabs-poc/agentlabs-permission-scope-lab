# Permission Quest — working design

Permission Quest is an executable design conversation. Payroll is the first scenario pack; the authorization mechanics must remain reusable by other products and domains.

## Working equation

```text
authority = principal × capability × assigned scope × trusted target context
```

The permission grammar is:

```text
<namespaced-noun>::<verb>
```

For example:

```text
hrms:payroll:ledger::read
```

The string identifies a stable capability. It does not contain the user, tenant, employee, department, validity window, or target record.

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

## The game as a decision tool

Each challenge combines a principal, requested permission, resource target, and expected decision. A correct outcome earns points; obtaining broad access is not itself success. The game rewards least privilege and denies tenant escape, owner substitution, naked mutations, and un-delegated agents.

The design board deliberately distinguishes `OPEN`, `PROPOSED`, and `AGREED`. Its JSON export captures the current grants, completed challenges, score, and question statuses. Export is a discussion artifact, not a production policy format.

## Extension seam

Scenario data lives in `src/scenario-pack.ts`. A pack declares principals, permissions, resources, initial grants, and adversarial challenges. The UI and decision loop consume that pack. Payroll-specific display fields still exist in this first spike and should become generic typed resource attributes when the second domain is added; doing that now would be speculative.

## Open design questions

The live board begins with these questions:

- Should assignment scope live in Auth or an HRMS governance binding?
- How are dynamic scopes represented canonically?
- Can roles carry a default scope, or only assignments?
- How does device-login delegation constrain a runner to the user's authority?
- What enters the access token and what stays server-side?
- Do explicit deny assignments override every allow?

No answer is canonical until it is deliberately agreed and moved into the implementation design.
