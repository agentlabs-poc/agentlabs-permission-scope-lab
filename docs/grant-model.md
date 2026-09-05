# Working handbook chapter: grants, assignments, and roles

This chapter consolidates agreed concepts with rationale, examples, consequences,
and unresolved boundaries. It is not a final implementation specification.
The [decision log](handbook-roadmap.md) records approvals; the
[discussion tree](discussion-tree.md) tracks remaining work. JSON illustrations
are in [grant-examples.md](grant-examples.md); their field names are not finalized.

## What a grant means — GRANT-001, TERM-004

A grant is an authorization binding: a recipient receives capabilities within
an assigned scope, subject to validity and conditions.

```text
Grant = recipient + permissions/role + scope + validity/conditions
```

"Grant" is the canonical name of this binding record. "Assignment" refers to
the same binding when describing a permission or role being assigned. "Assign"
is the action creating that binding. We do not introduce two logical records
solely because both words occur in explanations.

Example:

1. An authorized administrator assigns Certificate Reader to Maya within Finance.
2. The resulting binding is grant G-6.
3. Evaluation expands G-6's role to find its current permissions.
4. That computed view still represents G-6; it is not another assignment.

The reason for this terminology is to make identity and lifecycle unambiguous.
Calling the stored binding an assignment and its computed view a different grant
could incorrectly suggest a second independently revocable or independently
authorized object. This decision concerns the logical model, not table layout.

## Tenant encloses the grant — TENANT-001

Every example is evaluated within a trusted enclosing tenant. We omit repeated
tenant fields from individual grant examples. Recipients, groups, scope
references, and targets must resolve inside that tenant boundary.

A tenant-wide scope means the enclosing tenant. A Finance scope cannot reach a
different tenant's Finance department simply because their identifiers match.
Omitting a field from the grant does not remove the need to preserve tenant
binding in storage, lookup, caching, and transport. These physical contracts are
still open.

## Permission and scope stay associated — PERMISSION-001, GRANT-001

Permission identifies the operation on a resource type. Scope identifies the
reach over instances. Scope may name resources explicitly or use attributes and
relationships, as agreed in SCOPE-001.

GRANT-EX-001 gives Maya:

| Grant | Capability | Reach |
|---|---|---|
| G-1 | Read certificates | Enclosing tenant |
| G-2 | Revoke certificates | Finance department |

These grants do not authorize revoking an Engineering certificate. Resolution
must not take G-2's revoke capability and attach G-1's wider tenant scope.
It must likewise preserve each binding's validity and conditions.

This is a concrete application of RESOLUTION-001: resolution may make authority
explicit or restrict it, but cannot manufacture a capability/scope combination
that was never authorized.

## Multiple permissions and grant sources — GRANT-002, GRANT-003

A grant may bundle several permissions when they share the same recipient,
scope, validity, and conditions. If those constraints differ, use separate
bindings. GRANT-EX-003 illustrates a permissions array as one possible encoding.

A human may have multiple directly assigned grants and grants applicable through
valid group memberships. Group-derived applicability does not detach the scope
from the group grant. Losing one group membership does not erase an unrelated,
still-valid direct user grant.

Under agreed RESOLUTION-004, group-derived access must always remain derived
from the original group grant and valid membership. For example, G-7 belongs to
Auditors; Maya's membership makes G-7 applicable to her. Evaluation retains G-7's
identity and recipient together with the human being evaluated and the
supporting membership. It does not create a new independent grant for Maya.

The rationale is continued dependency: withdrawing the group grant or ending
membership must remove that route without requiring an independent user grant
to be found and separately revoked. A computed or cached view cannot sever the
dependency. Exact cache freshness and membership-evidence fields remain open.
Other independently valid direct human grants are evaluated separately.

Team/group
synonymy is settled in TERM-001; the Auth-owned, human-only group model remains
the working direction under GROUP-001-A and GROUP-003 pending consolidation.

### Self refers to the human, not the group — SELF-001, agreed

An Employees group may receive one self-scoped payslip-read grant. Vinay's group
membership makes that grant applicable to Vinay, and self resolves through his
trusted user-to-employee relationship. Maya's membership makes the same grant
applicable to Maya, with her own employee relationship. Neither gets the union
of all group members' payslips. The scope rule is shared; the resolved reach is
specific to the human being authorized.

For a dependent agent or service, self remains anchored to the human supplying
the authority, not to the automated actor's identity. Delegation limits still
apply. GRANT-EX-006 illustrates the human and proxy cases.

GROUP-004 establishes group-based access as the preferred practice, including a
single Employees self-service grant instead of one direct grant per employee.
The reason is shared administration: the grant defines the common rule, while
membership determines which humans receive it. The self selector supplies the
individual reach without creating separate grants for every employee.

This preference is not a prohibition. Q-021 explicitly retained direct human
grants under GRANT-003. A direct grant is not invalid merely because a group
could have been used, and no additional exception-approval process was agreed.
Changes to group membership remain authority-changing operations requiring
ADMIN-001's administrative controls.

### Combining positive grants — DECISION-001, agreed

Consider two valid grants for the same read permission:

| Grant | Source for Maya | Reach |
|---|---|---|
| G-8 | Direct human assignment | Finance |
| G-9 | Valid group membership | Engineering |

Either complete applicable grant may support access: Maya
can read Finance or Engineering, assuming no other rule prohibits it. When the
Engineering membership ends, that route disappears while Finance access remains
supported by G-8. Each route keeps its own conditions, validity, and provenance.

This combines the authority supplied by complete grants; it does not mix a
permission from one grant with scope from another or make the resulting access
independent of its sources. It does not change the subset/intersection limits
on human-dependent service/agent access. DECISION-002 excludes explicit deny
grants from v1.

### Positive-only grants — DECISION-002, agreed

A denied authorization decision is not the same concept as a stored deny grant.
V1 uses positive-only grants: grants supply authority,
while absence of a complete applicable grant or a mandatory boundary failure
prevents access. Q-019 concluded that explicit deny-grant objects and their
conflict-precedence rules are outside v1.

The discussed alternative was an explicit prohibition saying Maya must not read
C-17 even though a valid group grant permits reading all Finance certificates
and C-17 is in Finance. That alternative was not selected for v1.

Without explicit deny grants, withdrawing this access requires removing or
narrowing every route that authorizes C-17. Removing one direct grant does not
cancel a valid group-derived route. This tradeoff was stated before the user
accepted DECISION-002; it is not an implication that expired or removed grants
somehow invalidate unrelated valid grants.
Grant-local conditions likewise constrain their own grant; they do not by
themselves become global prohibitions against other valid grants.

### Administering grants — ADMIN-001, agreed

The agreed rule separates exercising a capability from assigning it to
someone else. Maya's permission to read Finance certificates does not by itself
permit assigning Finance certificate access to Arjun. Creating, changing, or
revoking grants requires explicit administrative authority within its allowed
bounds. Editing role permissions and changing group membership are also
access-changing operations that require the corresponding authorization.

This rule does not create an unlimited administrator role or settle whether
an administrator must personally possess the business authority being assigned.
The allowed recipient set, assignable capabilities, maximum scope, tenant
boundary, and lifecycle constraints need explicit decisions in this branch.

## Roles provide reusable capability definitions — ROLE-001, ROLE-002

A role is a named permission bundle. A referencing grant supplies recipient,
assigned scope, validity, and conditions. A role definition alone does not give
anyone access.

In GRANT-EX-004, Certificate Reader contains read and download:

| Grant | Recipient | Role | Reach |
|---|---|---|---|
| G-6 | Maya | Certificate Reader | Finance |
| G-7 | Auditors group | Certificate Reader | Enclosing tenant |

The role is reusable because the same capabilities can be assigned at different
scopes to different recipients.

Existing referencing grants use the current role definition. Removing download
withdraws that capability through both G-6 and G-7. Adding revoke supplies revoke
within Finance through G-6 and throughout the tenant through G-7. Grant scopes
and other constraints stay attached; unrelated grants still apply independently.

An authorized role edit changes declared authority. Resolution only interprets
it. Role-edit permissions, scope compatibility, revision evidence, cache
invalidation, and propagation guarantees remain to be specified. Live references
do not mean those questions are already solved.

## Expanded view of the same grant — RESOLUTION-003

Role expansion looks up the current permission set and exposes it in the same
permission-set form that evaluation uses for explicitly listed permissions.
The original grant identity and restrictions remain intact, and its role source
remains available for explanation.

GRANT-EX-005 shows G-6 with explicit read/download permissions and Finance scope.
The view is computed; it is not another stored grant created by assigning access.

"Expanded" is deliberately narrower than "fully resolved." Knowing a role's
permissions does not establish which employee a self selector refers to, or
whether a requested certificate is in Finance. Application facts and a concrete
target may still be needed. It is also not a completed allow decision.

## Consequences for services and agents — AUTHORITY-002, DELEGATION-002

The accepted model has no independently authorized service-account path. Both
service accounts and agents depend on a human and remain within that human's
applicable authority and the delegation's limits. Their identities can remain
distinct for authentication and attribution.

If the human loses supporting rights, resolution removes access no longer
covered; unrelated retained rights may continue to support access. Examples of
direct human grants must not be read as permitting independent service grants.
Exact delegation encoding and evaluation contracts remain open.

## Branches still to conclude

- Group-applicability view fields and membership evidence (the dependency itself
  is settled by RESOLUTION-004).
- Administrative bounds and delegation permissions; the need for explicit
  grant-administration authority is settled by ADMIN-001.
- Role/permission schema details, scope compatibility, role revision evidence.
- Status, validity boundaries, conditions, missing/empty scope, and failures.
- Group membership lifecycle, synchronization, nesting, and freshness.
- Other decision-combination constraints; positive-grant alternatives and the
  exclusion of explicit deny grants from v1 are settled by DECISION-001/002.
- Declared, expanded, prepared, and fully resolved grant/request contracts.
- Dependency growth, restoration, delegation chains, and revocation propagation.

Do not mark the grants stage complete until these required branches are settled
or explicitly excluded from v1. Scope and decision-algebra details are shared
dependencies with stages 6–9 in the discussion tree.
