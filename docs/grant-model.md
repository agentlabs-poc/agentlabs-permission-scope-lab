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

This rule does not create an unlimited administrator role. ADMIN-002 separately
settles that personal possession of the assigned business access is not required.
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

## Administration without business access — ADMIN-002, ADMIN-003 / Q-022

ADMIN-001 separates using a capability from administering its assignment.
ADMIN-002 establishes the reverse separation too: authorization to assign
business access does not require, or itself confer, that business access.

The agreed model is bounded administrative authority that does not
require personal business access. For example, Maya may administer payroll-read
grants for approved payroll groups within Finance without being able to read
Finance payslips herself. Her administrative authority must specify which
recipients, business capabilities, scopes, and applicable constraints she may
assign. It must not merely say that she can manage any grant in the tenant.

The alternative considered required both administrative authority and personal possession
of the business access being assigned. This adds a possession ceiling, but
would require a payroll-access administrator to receive payroll-reading access
in order to provision it. Q-022 rejected this universal possession prerequisite.
Business access alone remains insufficient to administer grants.

### Explicit self-assignment and audit

ADMIN-003 does not categorically prohibit Maya from assigning herself access.
It requires an explicit, authorized, audited operation. Before that operation,
administration alone does not permit reading payroll. If Maya is an eligible
recipient within her administrative bounds, she may explicitly assign herself
Finance payroll-read access. Subsequent business requests are evaluated against
that grant and all applicable restrictions, not against her administrator status.

If her bounds permit assigning only to specified groups, she cannot instead
create a direct grant to herself. Recording an unauthorized operation does not
make it authorized. Audit makes changes attributable and reviewable; enforcement
must reject changes outside administrative authority before they take effect.

The distinction is deliberate access acquisition versus implicit administrator
access, not a guarantee that an administrator can never acquire business access.
No mandatory second approver or separate self-assignment permission is adopted.

An illustrative audit narrative is: Maya requested payroll-read access for Maya
within Finance; the system evaluated her administrative authority, recorded the
outcome and resulting grant change, and later access used that grant. Exact
event fields, before/after evidence, timestamps, integrity, retention, and
behavior when audit recording fails remain open audit-design work.

Self-benefiting group membership or role edits must not become unexamined routes
around administrative bounds. ADMIN-001 already requires authorization for those
operations; exact controls and recording of their indirect access effects remain
open. Bootstrap and whether human/group grants survive the issuing administrator
losing authority are also still open.

These decisions concern administering human/group access. They do not introduce
independent service/agent authority: a proxy remains bounded by its human's
applicable authority and delegation restrictions under AUTHORITY-002.

## Branches still to conclude

### Whole-grant administrative bounds — ADMIN-004 / Q-023, proposed

The administrative action and the business permission being assigned are two
different layers. Maya asks to create a grant; the proposed grant would let its
recipient read payroll. Authorizing the first does not mean Maya can exercise
the second, or create any arbitrary grant.

The proposal checks the whole proposed grant against administrative bounds:
eligible recipients, assignable permissions, permitted resource reach, and any
required validity or conditions, alongside the allowed administrative operation.
The trusted enclosing tenant remains implicit and enforced.

### Same grant structure for administration — ADMIN-005 / Q-024, proposed

The original Q-023 illustration showed administrative_operation,
allowed_recipients, assignable_permissions, and maximum_resource_scope as a
standalone description. It was not intended as a canonical record, but looked
like a second authority format. The user challenged that shape. The refined
proposal uses the existing grant model, not a new administrative grant entity.

Auth resources are resources too. An administrative grant binds a recipient to
an operation on grants, groups, memberships, or roles, within a scope. For grant
creation, the authorization target is the proposed grant; it need not already
exist in storage. Scope selects which proposed grants may be created. Exact
create-target and scope-evaluation contracts remain stage 6/7 work.

The follow-up G-11 JSON also introduced unreviewed syntax: grant_selector,
recipient nested inside scope, permissions_subset_of, and scope_within. The
user challenged each addition. That example is withdrawn from the working
chapter, not adopted as a scope grammar; its history remains in Git and this
explanation. Keeping the outer grant structure does not by itself establish
that newly invented scope fields are justified.

PROCESS-005 requires explaining the need for each new field and checking whether
existing concepts already express it before adopting it. The following table
records intent, not accepted field definitions:

| Withdrawn syntax | Intended meaning | Why the representation is not settled |
|---|---|---|
| grant_selector | The target being selected is a grant. | Target resource type may already follow from the permission/operation declaration; another discriminator may be redundant. |
| recipient inside scope | Restrict the recipient of the proposed target grant, not the administrator receiving authority. | Recipient is an existing target-grant attribute; its use in a scope expression does not justify a new special nested field. |
| permissions_subset_of | Prevent the target grant from assigning business permissions outside the administrator's assignable set. | This describes a set relation; a dedicated field has not been shown necessary. Role references and current-role expansion also need treatment. |
| scope_within | Prevent the target grant from reaching resources beyond the administrator's assignable reach. | Semantic containment between scopes is not mere field or label comparison. Relationship/self scopes and future changes make this a substantive open problem, not a solved operator. |

For the same motivating example, Maya is a member of finance-access-admins and
may create grants for finance-payroll-readers that contain only payroll-read
and reach only Finance resources. This is a proposed requirement expressed in
plain language, not a finalized administrative scope or four mandatory fields.
It does not itself give Maya payroll-reading access.

SCOPE-001 already allows target selection by references, attributes, and
relationships. It does not yet define a common expression grammar, supported
relations, evidence requirements, or scope-containment procedure. The next
approved detour, Q-025, is to settle that shared scope model before writing more
administrative JSON, then return to ADMIN-004/005 to test whether it expresses
the required bounds. Neither a general predicate language nor admin-specific
fields have been adopted. A bare department label has not been shown sufficient
to express the proposed recipient and permission limits either.

Continue in [scope and target](scope-model.md), SCOPE-002 / Q-026, then return here.

The user's bootstrap direction is compatible with seeding initial ordinary
grants through a trusted initialization process. It does not require a second
runtime grant format or imply that a grant can authorize its own creation.
The bootstrap trust root, seed ownership, scope of initial authority, and
subsequent administration remain to be designed. No universal superuser bypass
is adopted by this proposal.

| Proposed action | Result under the proposed plain-language bounds |
|---|---|
| Create a Finance payroll-read grant for finance-payroll-readers | Within the shown bounds; other applicable checks still apply. |
| Create a tenant-wide payroll-read grant for that group | Outside the resource-scope bound. |
| Create a Finance payroll-edit grant for that group | Outside the assignable-permission bound. |
| Create a Finance payroll-read grant directly for Maya | Outside the recipient bound. |
| Edit or revoke an existing grant | Not authorized by grant.create alone. |

Maya's self-assignment is not universally prohibited: different authority could
explicitly include her among eligible recipients under ADMIN-003. Nor does this
example authorize her to add herself to the permitted group; that is a separate
administrative operation. Specifying a group recipient does not by itself settle
who may join it or whether business-department membership must be synchronized.

Scope comparison must establish semantic containment, not compare labels or
assume a hierarchy. A narrower target selection is acceptable only when shown
to fit the permitted reach. Relationship-based scopes, current role permissions,
later role changes, and recipient-relative self scopes need explicit treatment.

Following GRANT-001, administrative bounds stay associated with the authority
that supplies them. Permission to provision Finance payroll-read and separate
authority to provision Engineering certificate-read cannot be mixed into
Engineering payroll-read authority. More general multi-route administrative
combination rules remain open.

This proposal does not introduce a second required approval, finalized grant
schema, universal time limit, or authority to pass administrative powers onward.
Those details and update/revoke semantics remain separate discussion branches.

### Other open branches

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
