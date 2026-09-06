# Authorization: the agreed foundation

**Q-120A — automatic root growth:** legitimate application capability upgrades
should expand applicable root coverage without a separate manual root approval.
Ordinary child permissions remain explicitly selected and bounded. [Behavior
versus mechanism](../../docs/root-permission-evolution.md): no live `*` is yet
approved; root trust/binding and revision mechanics remain unfinished.

## Current closure pass — CLOSURE-001

[Closure checklist](../../docs/v1-closure.md): **38/68 criteria closed (55.9%),
30 open** after three documentary completions. This is not implementation or
final v1 acceptance. Approved decisions remain intact; Q-117 stays unapproved.

- [Principles and compliance examples](../../docs/principle-catalog.md).
- [Identity and authority glossary](../../docs/identity-glossary.md).
- [Grant/assignment contract closure inventory](../../docs/grant-contract-closure.md).
- [Separate implementation roadmap](../../docs/implementation-roadmap.md).

**Q-115 agreed:** trusted setup creates a selected legitimate human user and
administrators group, explicitly assigns initial root authority to the group,
and explicitly adds the human as a member. No special identity is required.
[Arrangement, JSON, and rationale](../../docs/bootstrap-initial-assignment.md).
**Q-116 agreed:** repeating completed bootstrap reports already initialized
without changing authority. Q-117 separately proposes withholding initial
authority until complete setup is durably established; it is not yet approved.

**Q-114 — registration-first bootstrap:** register the relevant permission and
scope contracts, including Auth's own administration, before initial grant
acceptance. Trusted setup establishes maximum intended tenant authority and
explicitly makes it available to the initial human administrator. Normal
administration then creates users/groups, distributes equal/narrower grants,
and manages membership. Earlier FIN-read-only seed framing is superseded.

[Bootstrap rationale](../../docs/bootstrap-authority.md) · [Open setup flow SVG](../../docs/assets/bootstrap-registration-flow.svg).

## Current revision decisions — Q-102–Q-106

[Grant revisions and adoption](../../docs/grant-revisions.md) records approved
independent adoption, update suggestions, current-adopted top-down lineage,
one assignment per grant/recipient, latest-only creation/upgrades, and immutable
authority content with separate live controls. Existing assignments do not
automatically upgrade. All current parent boundaries remain mandatory.

[Q-107 core JSON](../../docs/grant-revision-format.md) is **approved at core-shape
level**; full schemas remain open. All discussion and exploration now belong in
the lab; [scratchpad sources](../../docs/history/scratchpad-import/README.md) are history.

[Q-112A](../../docs/direct-human-parent-context.md) reaffirms lineage-supported
latest. The extra `parent_grant_revision` proposal is withdrawn; eligible-support
discovery and validation/evidence contracts remain separate unfinished work.

[Q-108](../../docs/assignment-validity.md) defers assignment-specific validity in v1.
[Q-109](../../docs/grant-validity.md) puts optional grant validity in immutable
revision content. Changing the window requires explicit adoption; publication
alone neither extends nor shortens an existing assignment's adopted window.
Current upstream validity and live grant-wide disablement still constrain access.

[Q-110](../../docs/auth-write-consistency.md) preserves the checked revision and
authority through Auth's assignment write. A conflicting change stops that
attempt, without saving stale approval or silently substituting another revision.
Persistence mechanics and conflict-result formatting remain open.

[Q-111](../../docs/lineage-cycles.md) rejects grant/team self-parenting and ancestor
loops, including disabled records/bindings. Revision-aware validation details
remain open. [Milestone progress](../../docs/milestone-progress.md) separates
remaining checkpoint percentages from recent partial advances and effort estimates.

## Latest pinned discussion — Q-101

[Parent-grant bindings](../../docs/parent-grant-bindings.md) records the current
framing, rationale, and 31 review cases. A child grant uses `parent_grant_id`;
actual assignments and parent-team context establish required support. No
additional parent-assignment lineage reference is required for the settled cases.

[Open the four-part grant/team binding SVG](../../docs/assets/parent-grant-bindings.svg).

Assignment, grant enablement, and effective authority are different controls.
Affected assignment changes proceed bottom-up. Relevant bindings may be removed
OR disabled before parent changes; retained assignments require explicit
re-enablement validated against current reality.

- [Direct/shared-route comparison SVG](../../docs/assets/parent-grant-routes.svg)
- [Structural-change and re-enablement SVG](../../docs/assets/binding-change-lifecycle.svg)

Earlier specific-supporting-assignment wording is superseded where it conflicts;
real parent-team support, narrowing, and the endpoint-owned gate remain required.

Current authority model: [Q-090](../../docs/grant-assignments.md) separates
recipient-free grants from assignments; [Q-091](../../docs/subgroups.md) makes
explicit dependent subgroups canonical without inherited membership. The
`0.0.1` tag preserves the preceding model. Full new contracts remain open.

[Q-099: ownership and lineage](../../docs/ownership-lineage.md) separates owner
rotation from continuing team-held support. The acting administrator must still
be authorized; actual personal dependencies remain required. No automatic
parent rebinding or merging of personal grants occurs. See the
[ownership/lineage SVG](../../docs/assets/ownership-lineage.svg).

[Q-100: inside Auth Service](../../docs/auth-service-authority-gate.md) explains
why permission to administer an assignment is not permission to distribute
arbitrary authority. Auth's administrative evaluator and authority-boundary
validator must both pass before the protected write. See the
[Auth Service internal-flow SVG](../../docs/assets/auth-service-authority-gate.svg)
for shared records, failure paths, and persistence. Exact contracts remain open.

<details>
<summary>Earlier reader checkpoint — preserved history</summary>

This reader summarizes the approved discussion through **Q-050-F**. It is a
working handbook, not a claim that the historical simulators implement the model.
Detailed rationale remains in the [decision log](../../docs/handbook-roadmap.md)
and [current chapters](../../docs/handbook.md). The
[Q-041 through Q-050-F decision trail](../../docs/handbook.md#decision-trail--q-041-through-q-050-f)
keeps those agreements, rationale, and remaining questions visible. The
[original concept page](../../docs/history/reconciliation-2026-09-05/src/content/authorization-concept.md.txt)
is preserved as deprecated history.

</details>

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
list explicit registered permissions, not wildcard patterns. Q-059 also excludes
permission aliases from v1: different registered identifiers remain distinct,
regardless of similar wording or UI labels.

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

The grant defines reusable authority; a separate assignment identifies who
receives it. These are minimal examples, not complete lifecycle/update schemas.

```json
{
  "version": "1",
  "id": "G-SELF-PAYSLIP-READ",
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": {"user": "$self"}
}
```

```json
{
  "version": "1",
  "id": "A-EMPLOYEES-PAYSLIP-READ",
  "grant_id": "G-SELF-PAYSLIP-READ",
  "recipient": {"type": "group", "id": "employees"},
  "status": "enabled"
}
```

For Vinay, self still means Vinay. His valid membership supplies the Employees
assignment, which references the grant. Creating the definition alone supplies
no access. Keep both record identities and membership dependencies in resolution.
Roles provide permissions from explicitly adopted immutable revisions, not live
updates. Shared definition revision/adoption mechanics remain open.

A Finance-write route and tenant-wide-read route cannot combine into tenant-wide
write. [Meaning, reuse, and restrictions](../../docs/grant-assignments.md).

<details>
<summary>Earlier recipient-bearing grant and live-role prose — deprecated, preserved</summary>

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

</details>

## 3. Human membership and dependent automation

Auth owns authorization groups and memberships; team and group mean the same
thing. Groups contain humans. Group-based access is preferred, not mandatory:
direct human grants remain supported. Applications may synchronize business
groups to Auth or keep them separate.

Subteam and subgroup are synonymous. A subgroup receives only explicitly selected
dependent authority: a parent grant supported in its parent-team context, a permission
subset, and parent scope AND child constraints. Child scope `{}` adds no further
restriction; it never discards the parent boundary. Membership is not inherited.
See [Team1, Team2, and Nutan](../../docs/subgroups.md).

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

For Vinay, authority loading obtains valid memberships, direct/group assignments,
their grant definitions, and supporting assignment/delegation dependencies.
Adopted role expansion and dependency resolution retain provenance and restrictions.
A resolved grant is an evaluation-ready dependent view—not an allow decision.

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
