# Authorization: the agreed foundation

This reader summarizes the approved discussion through **Q-050-F**. It is a
working handbook, not a claim that the historical simulators implement the model.
Detailed rationale remains in the [decision log](../../docs/handbook-roadmap.md)
and [current chapters](../../docs/handbook.md). The
[Q-041 through Q-050-F decision trail](../../docs/handbook.md#decision-trail--q-041-through-q-050-f)
keeps those agreements, rationale, and remaining questions visible. The
[original concept page](../../docs/history/reconciliation-2026-09-05/src/content/authorization-concept.md.txt)
is preserved as deprecated history.

## 1. Permission answers “what”; scope answers “within which boundaries”

A permission identifies an operation. A scope selects the boundary within which
a grant can supply that permission. Both are needed, along with an applicable
recipient, valid authority, and all mandatory constraints.

The earlier permission explanation is retained in the active
[permission chapter](../../docs/permission-model.md). A permission string is an
operation name, not complete authority: the same payroll-ledger read permission
can be granted for self, Finance, or tenant-wide reach through different scopes.

The earlier naming convention separates the namespaced noun from its operation:

```text
hrms:payroll:ledger::read
└ namespaced noun ┘  verb
```

Here `hrms` identifies the application, `payroll` the domain, `ledger` the resource
type, and `read` the operation. Employee IDs, departments, and tenant reach stay
out of the permission name. Q-056 adopts the convention with variable depth:

```text
<app>:<domain>[:<subdomain-or-resource>...]::<verb>

hrms:payroll::read
hrms:payroll:ledger::read
hrms:payroll:ledger:entry::read
hrms:payroll:ledger:entry:attachment::read
```

The application defines those levels; Auth does not hardcode their business
meaning. `:` separates namespace levels and `::` separates the verb. Naming
depth does not establish scope reach. Q-057 agrees that parent permission names
do not automatically authorize deeper permissions; a grant or role can explicitly
list both. Q-058 excludes wildcard permission names from v1: grants and roles
list explicit registered permissions, not wildcard patterns. Permission-alias
exclusion is a separate open proposal under Q-059.

There is no additional canonical “target” entity or required resource wrapper
(TERM-005). A request can identify a certificate, list, or proposed change using
ordinary application inputs. Removing a wrapper does not remove the obligation
to constrain the actual operation.

Scope is a required flat key-value object (SCOPE-007/008):

| Scope fragment | Meaning, where the application supports these keys |
|---|---|
| `{"dept":"FIN"}` | Finance boundary. |
| `{"user":"$self"}` | The authorizing human's defined self boundary. |
| `{"dept":"FIN","user":"$self"}` | Finance AND that human's self boundary. |
| `{}` | No narrower scope restriction inside the trusted tenant. |

Keys and their meanings are registered by the application; department is not a
universal platform type. Values are single non-empty string references or a
supported `$self` reference. No arrays, nested expressions, wildcard, or arbitrary
query language is adopted. Missing or null scope is invalid, never shorthand
for `{}`. Alternatives use separate complete grants, not OR inside one scope.

**Rationale:** this keeps the boundary format small without confusing a stored
scope with SQL. An application may enforce a boundary using a constrained query,
but that does not make the grant a query program.
[Scope definitions and rejected formats](../../docs/scope-model.md).

## 2. A grant keeps permission and boundary together

This abbreviated working grant gives members of the employee group permission
to read their own payslips. Lifecycle fields are omitted for focus; this is not
a complete published grant schema.

```json
{
  "version": "1",
  "recipient": { "type": "group", "id": "employees" },
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": { "user": "$self" }
}
```

For Vinay, self means Vinay—not the group, its creator, or another member. The
application defines the employee-to-human relationship; a matching field name
does not establish it.

A human can have many direct and group-derived grants; a grant can bind many
permissions to the same scope and restrictions. Roles supply a live permission
bundle, not a new scope. Resolving a role does not create another independent
assignment.

Each candidate grant is evaluated as a complete binding. A Finance-write grant
and a tenant-wide-read grant do **not** combine into tenant-wide write. Valid
positive grants are alternative routes; explicit deny grants are excluded from
v1. [Grant meaning and rationale](../../docs/grant-model.md).

## 3. Human membership and dependent automation

Auth owns authorization groups and memberships; team and group mean the same
thing. Groups contain humans. Group-based access is preferred, not mandatory:
direct human grants remain supported. Applications may synchronize business
groups to Auth or keep them separate.

Services and agents always depend on a human's authority and remain a subset of
it. They are not first-class group members. Losing a required membership or
upstream authorization invalidates the affected derived route. Another valid
route is not automatically removed.

Ordinary grants to humans/groups are different: a valid grant does not disappear
solely because its issuing administrator later loses issuance authority.
Withdrawing that grant requires explicit revocation (ADMIN-006).
[Membership](../../docs/groups-and-membership.md) and
[dependent authority](../../docs/grant-model.md) retain the detailed distinctions.

## 4. Two responsibility layers, one endpoint-owned gate

Layer 1 supplies canonical authority rules: grants, permissions, memberships,
roles, dependency limits, tenant isolation, and scope composition. Most authority
material comes from Auth.

Layer 2 supplies application meanings, selected inputs, valid values, facts
where needed, additional restrictions, and enforcement. The application-embedded
auth agent works across both layers using shared rules.

For Vinay, authority loading logically obtains valid memberships and then direct
grants plus grants for those groups. Role expansion and dependency resolution
retain provenance and restrictions. A resolved grant is an evaluation-ready,
dependent view—not an allow decision.

Pre-handler middleware may authenticate and load context or authority. Business
authorization has **one endpoint-owned gate**, with no middleware allow/prepared
handoff. Facts may be established before evaluation or relationships enforced
through constrained execution; there is no mandatory eager database lookup or
canonical resolver API.
[Current system overview](../../docs/system-overview.md).

## 5. Endpoints are designed auth-first

Each protected method/route declares exactly one permission and selected inputs
with their sources. The server owns this policy; the caller does not choose its
permission.

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

This approved partial structure has no relationship block, named resolver, or
duplicate type/nullability schema. Every declared input is required at its exact
source; no silent default or fallback. The application validates values, and
authorization and execution use the same validated meaning.

The path tenant must agree with trusted tenant context. Local input names do not
automatically become scope keys. A tenant-wide grant does not make declared
inputs optional or invent department restrictions.
[Policy examples and rationale](../../docs/endpoint-policy-format.md).

## 6. Boundary authority must constrain actual execution

Suppose Vinay has certificate-read with `{"dept":"FIN"}` and requests
`GET /api/v1/acme/FIN/C-17`. Auth can establish read authority within Finance.
It cannot infer from that URI that C-17 belongs to Finance.

The endpoint must establish or enforce that relationship before disclosing data.
A parameterized, constrained lookup can enforce it directly:

```sql
WHERE tenant_id = :trusted_tenant
  AND department_id = :authorized_department
  AND certificate_id = :requested_certificate
```

Here the authorized department is Finance, consistent with the request. An
Engineering certificate cannot pass this lookup. An unchecked ID-only read after
accepting Finance in the path violates the contract.

Review the actual data returned or changed, not mere input usage: logging the
department, an ineffective OR condition, or a later unchecked output path does
not enforce the boundary. This is mandatory authorization work, not optional
business validation. It is a conscious endpoint responsibility, not a guarantee
that Auth alone detects every application defect (CONTRACT-012 / ENFORCEMENT-003).

## 7. Registration and administration

Applications register permissions and scope contracts. Auth validates registered
identifiers, canonical shape, and issuance authority without interpreting the
application database. Optional permission–scope support validation is selected
upfront per application; when enabled it applies to all grants. This registration
feature is distinct from the rejected endpoint relationship block.

The ability to grant access is separate from the ability to use it. An
administrator's self-assignment must be explicit, authorized, and auditable.
Exact administrative scope encoding remains open; no separate unapproved grant
format or containment field is introduced.
[Registration](../../docs/application-registration.md) and
[administration](../../docs/grant-model.md).

## 8. What remains unfinished

All published JSON/YAML contracts require a top-level string `version`, initially
`"1"`; missing, malformed, or unsupported versions are rejected. Adding a version
does not finish the rest of a schema.

Decision-result contracts, full schema publication, update/move and bulk/list
semantics, freshness/revocation/concurrency, delegation mechanics, audit formats,
and comprehensive conformance tests remain open. The historical explorers are
not conformance tests. Follow the [discussion tree](../../docs/discussion-tree.md)
for the whole-handbook map and [cross-domain cases](../../docs/use-case-examples.md)
for Git, ticketing, HRMS, and accounting examples.
