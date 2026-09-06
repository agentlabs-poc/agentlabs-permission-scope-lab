# Working handbook chapter: grants, assignments, and roles

Role update — Q-089-B / ROLE-003: grants select an immutable published role
revision explicitly. Publication does not change existing grants; adoption is an
authorized, boundary-validated grant change. This supersedes ROLE-002's live-role
behavior and the “current/latest role” interpretation in older examples below.
RESOLUTION-003 still produces a dependent view of the same grant, now using its
adopted revision. See [role revisions](role-revisions.md) for formats and rationale.

Lifecycle terminology after Q-082: the canonical grant operations are create,
enable, disable, and delete. A separate revoke operation is superseded; older
references to permanent grant revocation mean deletion from usable authority.
The permanence, dependent-subset, and complete-binding rules remain. See
[grant lifecycle](grant-lifecycle.md) for current meanings and retained history.

Q-049 / CONTRACT-008 now mandates exactly one required permission per protected
endpoint, validated by Auth. This does not limit the number of permissions in a
grant or the number of direct/group grants available to a human. Complete grants
remain alternative routes for the endpoint's one permission. See
[endpoint authorization](endpoint-authorization.md) for rationale and examples.
Earlier multi-permission endpoint-contract gaps below are historical after Q-049;
grant packaging, concrete schemas, and other lifecycle/enforcement gaps remain.

## Resolved grants and membership-based retrieval — RESOLUTION-006 / Q-048, agreed

A resolved grant is an evaluation-ready view of an existing grant, with the
relevant references and human context established, while preserving its original
authority, restrictions, and dependencies. It is not a new assignment or a bare
permission list. Resolved does not mean that the request is allowed.

### Logical retrieval flow

The user approved Q-048 and added that we need Vinay's memberships, then resolved
grants for Vinay and the teams of which he is a member:

1. Establish Vinay's verified identity and enclosing tenant context.
2. Obtain his valid authorization-group memberships from Auth-owned authority.
3. Retrieve direct grants assigned to Vinay and group grants assigned to those
   groups. Direct grants remain supported, though group grants are preferred.
4. Establish evaluation views for Vinay, retaining each source grant and the
   membership supporting each group-derived route. Resolve relevant role
   references and human-relative scope meanings without widening authority.
5. Evaluate the relevant complete grants against the request material, with
   validity, conditions, tenant, and any delegation constraints enforced.

This is a logical dependency flow, not a requirement for separate API calls,
a particular service response, or eagerly loading every grant before any denial.
Fetching already computed views is compatible with it if those views preserve
the dependencies and satisfy the eventual freshness contract. A caller-supplied
list of group names is not membership proof. An empty valid membership list does
not remove legitimate direct grants; failure to obtain membership is not evidence
that group membership is absent. Detailed retrieval/error contracts remain open.

### Example: Finance and self through Employees

This uses the existing working grant layout; lifecycle details are omitted for
focus, not removed from evaluation:

```json
{
  "id": "G-17",
  "recipient": { "type": "group", "id": "employees" },
  "permissions": ["hrms:employee:certificate::read"],
  "scope": { "dept": "FIN", "user": "$self" }
}
```

For Vinay requesting C-17, the view remains G-17 with group recipient Employees.
Valid membership establishes why this authority is available through Vinay.
`$self` means Vinay's self relationship, not the group collectively. Finance and
self remain AND-bound restrictions on this grant's certificate-read permission.
For a role-referencing grant, permissions and role provenance are established
from its adopted revision under RESOLUTION-003 as updated by ROLE-003;
scope and other constraints remain associated.

Resolution has clarified the authority source, recipient relationship, capability
references, and human-relative boundary meaning. It has not established merely
by doing so that C-17 belongs to Vinay or Finance. Relevant trusted request
material must establish boundary satisfaction for the decision. Scope `{}`
introduces no narrower department/self restriction to resolve, while other
mandatory checks remain. No universal lookup or new handoff state is introduced.

### Rationale, counterexamples, and dependencies

Flattening this into "Vinay can read certificates" would lose Finance/self
restrictions and the membership dependency. Converting it into a direct grant
would let the derived route survive membership removal. Both alternatives
conflict with the agreed complete-grant and dependency invariants.

| Change or attempted shortcut | Consequence |
|---|---|
| Vinay loses Employees membership | G-17 no longer supplies authority through that membership; unrelated valid routes remain separate. |
| A different grant supplies tenant-wide certificate-read | It does not donate broader scope to a different permission in G-17 or another grant. Evaluate each complete binding. |
| C-17 belongs to another employee | G-17's self boundary is not satisfied, even though its meaning is resolved for Vinay. |
| Vinay's agent acts on his behalf | Use the human's supporting authority with delegation limits; the agent does not acquire first-class group membership. |

Source identity and constraints survive computation and caching; a view is not
independent authority after supporting state changes. Q-046's issuer-provenance
rule is distinct: ordinary issuance history is not a live membership/delegation
dependency. Exact freshness, cache invalidation, wire fields, delegation encoding,
and multi-permission decision contracts remain open. This approval consolidates
meaning and retrieval sources, not those implementation mechanics.

## Ordinary grant lifecycle — ADMIN-006 / Q-046, agreed

A validly issued ordinary human/group grant does not become invalid merely
because its issuing administrator later loses grant-administration authority.
The grant remains subject to its own scope, validity, conditions, revocation,
and applicable recipient dependencies. Issuance provenance is retained for
audit; it is not, by itself, a continuing authority dependency.

### Rationale and alternative

Creating an ordinary grant is an authorized administrative action, not lending
the administrator's personal business access. ADMIN-002 already separates
providing access from using it. Routine rotation of administrative duties
should therefore not unexpectedly withdraw organizational access.

The alternative considered was to invalidate every grant issued through an
administrator when that administrator loses issuance authority. That would
provide cascading withdrawal, but make all those grants depend on the
administrator's continuing role and create potentially disruptive chains.
Q-046 does not adopt that dependency for ordinary human/group grants.

### Example and dependency distinctions

Maya is authorized to create a Finance payslip-read grant for
`finance-payroll-readers`. She validly creates it; Vinay receives access through
membership. Maya subsequently changes departments and loses her issuance
permission. The group's grant remains valid subject to its own constraints;
Vinay must still have valid membership. Maya can no longer use the lost
administrative authority to create further grants.

| Subsequent event | Consequence |
|---|---|
| Maya loses the authority she used to issue the grant | This alone does not invalidate the ordinary group grant. |
| Vinay loses the supporting group membership | Access derived through that membership is no longer available; unrelated valid routes remain separate. |
| The grant expires or is explicitly revoked | It no longer supplies authority, regardless of Maya's current permissions. |
| Vinay's agent relied on the affected group-derived access | Its access remains bounded by Vinay's current authority and delegation limits; Q-046 does not create independent automation. |

### Security consequence and remaining limits

Removing an administrator's permissions is not cleanup of previously issued
grants. Where those grants must be withdrawn, explicit authorized revocation is
required. Preserve who issued them and the relevant issuance evidence for audit
and review; audit does not itself authorize or revoke anything.

This rule concerns grants that were validly issued. It does not legitimize an
unauthorized issuance, waive other validity checks, freeze a referenced role's
permissions, or remove source-grant/membership dependencies from computed
views. Incident-response procedures, exact audit records, revocation propagation,
and other lifecycle mechanics remain open. No new grant field or format is
introduced. Earlier Q-046-proposed text below is preserved as checkpoint history.

Q-045 also settles group ownership and human-only membership. See
[groups and membership](groups-and-membership.md) for the rationale, alternative,
examples, and lifecycle limits. Q-046 asks about ordinary human/group grants
after the issuing administrator loses administration authority; that lifecycle
rule remains proposed and does not change human-dependent automation.

Current: Q-044 approves ADMIN-004/005 at the governing-rule level. The ordinary
grant model applies to administration, with complete-assignment and associated
boundary checks. Exact administrative scope encoding and containment remain
open. Earlier proposed-status notes below are preserved historical positions.

## Grant administration — Q-044, agreed rules

1. **Ordinary grant model.** Administrative authority is a grant binding a user
   or group to administrative permissions within a scope boundary. No separate
   authority format or new administrative scope keys are introduced here.
2. **Explicit administrative operation.** Permission to create grants does not
   automatically permit updating or revoking them. The required permission
   must correspond to the administrative operation actually requested.
3. **Complete requested assignment.** The proposed recipient, permissions,
   scope, and applicable validity/conditions must fit the administrator's
   authority. For changes, consider relevant existing and proposed material.
   Concrete before/after contracts remain to be specified.
4. **Associated bounds.** Unrelated administrative grants cannot donate fields
   to manufacture authority. Finance payroll-read provisioning plus Engineering
   certificate-read provisioning does not authorize Engineering payroll-read.
5. **Normal validation.** Administration does not bypass registered permission
   and scope contracts, enabled relationship validation, canonical validity
   checks, or tenant boundaries.

Providing access still does not confer personal access (ADMIN-002), and
self-assignment remains explicit, authorized, and audited (ADMIN-003). This
approval does not impose personal business-access possession or a second approver.

### Example: Finance payroll provisioning

Assume Maya may create payroll-read grants for `finance-payroll-readers` within
Finance. She requests this business grant; lifecycle fields are omitted for focus:

```json
{
  "recipient": {
    "type": "group",
    "id": "finance-payroll-readers"
  },
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": { "dept": "FIN" }
}
```

| Request under the stated authority | Outcome |
|---|---|
| Create the displayed grant | Permitted subject to other required checks. |
| Use Engineering or tenant-wide scope | Outside Maya's authorized reach. |
| Add payroll-write | Outside her assignable permissions. |
| Make Maya the direct recipient | Outside her permitted recipient boundary. |
| Revoke an existing grant | Requires corresponding administrative permission. |

Maya cannot read payslips merely because she can create the first grant. Nor
does this authority permit changing group membership to obtain indirect access;
membership and role administration require their own authorization.

Maya's administrative authority is described in prose because its exact scope
encoding is not settled. A bare `dept` entry has not been shown sufficient to
express recipient, assignable-permission, and containment limits together.

### What this closes and what remains

The complete-assignment/associated-bound rules of ADMIN-004 and ordinary-grant
model of ADMIN-005 are agreed. Scope encoding, containment, bootstrap, onward
administration, and detailed role/lifecycle mechanics remain concrete-contract
work. No withdrawn G-11 field or operator is revived. Q-045 next consolidates
the earlier group ownership and human-only membership proposals.

Q-043 settles TERM-005's vocabulary correction, not the proposed administrative
scope model. [Authorization vocabulary](authorization-vocabulary.md) explains
the boundary/request-material approach and its safety requirements. Earlier
wording revised here is retained in the [deprecated record](history/q043-vocabulary.md).

Current scope format is SCOPE-007 / Q-034: a required flat key-value object,
AND within scope, with explicit {} for tenant-wide reach and no missing/null
default. See [scope boundaries](scope-model.md). Earlier grant examples retain
their historical scope syntax; GRANT-EX-007 shows the canonical scope form.
The [current grant formats](grant-format.md) also restate direct/group, role,
and expanded-view layouts without deleting the earlier examples.

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
3. Evaluation expands G-6's explicitly adopted role revision to find its permissions.
4. That computed view still represents G-6; it is not another assignment.

The reason for this terminology is to make identity and lifecycle unambiguous.
Calling the stored binding an assignment and its computed view a different grant
could incorrectly suggest a second independently revocable or independently
authorized object. This decision concerns the logical model, not table layout.

## Tenant encloses the grant — TENANT-001

Every example is evaluated within a trusted enclosing tenant. We omit repeated
tenant fields from individual grant examples. Recipients, groups, scope
references, and resources must resolve inside that tenant boundary.

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

**Current rule — ROLE-003 / Q-089-B:** a grant adopts a particular immutable role
revision. Adding or removing permissions in a new revision does not change that
grant until an authorized adoption passes boundary validation. The role/recipient/
scope distinction above survives; older unpinned examples are not complete current
contracts. See [revision and adoption examples](role-revisions.md).

<details>
<summary>ROLE-002 live-role behavior — deprecated by Q-089-B, retained as history</summary>

Existing referencing grants use the current role definition. Removing download
withdraws that capability through both G-6 and G-7. Adding revoke supplies revoke
within Finance through G-6 and throughout the tenant through G-7. Grant scopes
and other constraints stay attached; unrelated grants still apply independently.

An authorized role edit changes declared authority. Resolution only interprets
it. Role-edit permissions, scope compatibility, revision evidence, cache
invalidation, and propagation guarantees remain to be specified. Live references
do not mean those questions are already solved.

</details>

## Expanded view of the same grant — RESOLUTION-003

Role expansion looks up the adopted revision's permission set and exposes it in the same
permission-set form that evaluation uses for explicitly listed permissions.
The original grant identity and restrictions remain intact, and its role source
remains available for explanation.

Historical wording, superseded by Q-089-B: “looks up the current permission set”
meant the live role definition. The computed-view principle remains; latest-role
substitution does not. Preserve the adopted revision in the resolution evidence;
the exact resolved-view transport schema remains open.

GRANT-EX-005 shows G-6 with explicit read/download permissions and Finance scope.
The view is computed; it is not another stored grant created by assigning access.

"Expanded" is deliberately narrower than "fully resolved." Knowing a role's
permissions does not establish which employee a self selector refers to, or
whether a requested certificate is in Finance. Application facts and a concrete
resource may still be needed. It is also not a completed allow decision.

## Consequences for services and agents — AUTHORITY-002, DELEGATION-002

The accepted model has no independently authorized service-account path. Both
service accounts and agents depend on a human and remain within that human's
applicable authority and the delegation's limits. Their identities can remain
distinct for authentication and attribution.

If the human loses supporting rights, resolution removes access no longer
covered; unrelated retained rights may continue to support access. Examples of
direct human grants must not be read as permitting independent service grants.
Exact delegation encoding and evaluation contracts remain open.

The impact-first [delegation lifecycle discussion](delegation-lifecycle.md)
records Q-070's agreed automatic reactivation: affected delegated access is
inactive while human support is absent and works again when support returns,
provided the delegation itself remains valid. Explicit renewal was not adopted.
The human-subset and affected-authority-only rules remain; lifecycle mechanics
are still open.

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

Scope clarification after Q-076: the explicit authorization and separation of
administration from business access below remain in scope. Audit mentions retain
the earlier integration intent; event design, evidence storage, retention, and
audit-failure behavior belong to another layer and are not handbook deliverables.
Historical references below to open audit-design work are superseded on scope,
not a reason to reopen or weaken administrative authorization.

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
creation, request material describes the proposed grant; it need not already
exist in storage. Scope selects which proposed grants may be created. Exact
creation-request and scope-evaluation contracts remain stage 6/7 work.

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
| grant_selector | The resource being selected is a grant. | Resource resource type may already follow from the permission/operation declaration; another discriminator may be redundant. |
| recipient inside scope | Restrict the recipient of the proposed resource grant, not the administrator receiving authority. | Recipient is an existing proposed-grant attribute; its use in a scope expression does not justify a new special nested field. |
| permissions_subset_of | Prevent the resource grant from assigning business permissions outside the administrator's assignable set. | This describes a set relation; a dedicated field has not been shown necessary. Role references and current-role expansion also need treatment. |
| scope_within | Prevent the resource grant from reaching resources beyond the administrator's assignable reach. | Semantic containment between scopes is not mere field or label comparison. Relationship/self scopes and future changes make this a substantive open problem, not a solved operator. |

For the same motivating example, Maya is a member of finance-access-admins and
may create grants for finance-payroll-readers that contain only payroll-read
and reach only Finance resources. This is a proposed requirement expressed in
plain language, not a finalized administrative scope or four mandatory fields.
It does not itself give Maya payroll-reading access.

SCOPE-001 already allows boundary selection by references, attributes, and
relationships. It does not yet define a common expression grammar, supported
relations, evidence requirements, or scope-containment procedure. The next
approved detour, Q-025, is to settle that shared scope model before writing more
administrative JSON, then return to ADMIN-004/005 to test whether it expresses
the required bounds. Neither a general predicate language nor admin-specific
fields have been adopted. A bare department label has not been shown sufficient
to express the proposed recipient and permission limits either.

~~Continue in [scope boundaries](scope-model.md), SCOPE-002 / Q-026, then return here.~~

Current return after Q-042: core scope/registration principles are agreed and
the user has parked further registration lifecycle detail. Resume ADMIN-005
with the ordinary grant/permission/scope concepts and material describing the
proposed creation or change. Q-043 settles vocabulary only; ADMIN-004's
assignable bounds and ADMIN-005's representation still require discussion.
ADMIN-001/002/003 separation and explicit audited self-assignment remain agreed.
No withdrawn G-11 fields or containment algorithm are adopted by this return;
the canonical administration model remains a proposal after Q-043's vocabulary decision.

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
assume a hierarchy. A narrower boundary selection is acceptable only when shown
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
