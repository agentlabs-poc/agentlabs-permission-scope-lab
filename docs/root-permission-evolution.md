# Q-120 — New application permissions and root authority

**Q-125 approved:** [authorized retirement withdraws permission coverage](permission-lifecycle.md)
even if existing grants still reference the permission. The computed root no
longer supplies it once retirement is effective; references do not preserve access.
No automatic grant/assignment edit occurs. Earlier removal gaps below are narrowed.

**Bootstrap follow-up:** the previously parked [Q-117 is now approved](bootstrap-initial-assignment.md).
Incomplete setup provides no initial administrator authority. Earlier Q-117
pending notices below are historical; this does not change root computation.

## Q-123 — One application version for everyone, user correction accepted

**AGREED AS CORRECTED.** The user clarified that the application has one version
for everyone. Selective tenant updates do not exist and are not required.
The proposed per-tenant installed-release catalog selection is not adopted.

For each application, its roots use the same current authoritative registered
capability catalog. When the shared application update introduces delete, all
otherwise valid roots for that application reflect that catalog addition through
Q-122 computation. There is no tenant-specific release pin, catalog adoption,
or separate root-update approval in this model.

```text
HRMS shared catalog: read, write
Tenant A HRMS root:  read, write
Tenant B HRMS root:  read, write

HRMS shared update adds delete
Tenant A HRMS root:  read, write, delete
Tenant B HRMS root:  read, write, delete
Existing child selecting read: still read
```

**Rationale:** the earlier question introduced a deployment/versioning feature
the product neither provides nor needs. The shared version removes that catalog
selection branch rather than requiring new installation-version binding records.
This is a per-application rule: HRMS and an unrelated application still have
separate catalogs and management boundaries.

**Core-philosophy check:** a shared capability catalog is not shared tenant
authority. Each root remains in its own tenant/application boundary with its own
assignments, enablement, validity, and membership. A catalog addition does not
install an application, create a root, grant a publisher tenant data access, or
select new permissions for ordinary children. Independent ordinary grant/role
revision adoption remains supported; that is not selective application versioning.

No per-tenant release/catalog field is introduced. Canonical root-source encoding,
authoritative publication visibility, failure handling, and definition removal
remain to be finalized. These must not reintroduce tenant release selection by
another name. One shared logical version does not alone prove distributed
updates become visible atomically; that is a consistency contract, not a rollout
feature adopted here.

**History — Q-123 recommendation not adopted:** the assistant proposed Tenant A
remaining on HRMS v1/read-write and Tenant B using v2/read-write-delete, with each
root following its installation's release. The intent was to preserve definition
versus installation boundaries. The user corrected the unnecessary assumption
of selective upgrades. That proposal and earlier release-selection gaps below
are superseded by the single-version rule, not parked v1 requirements.

## Q-122 — Computed root permission coverage, approved

**APPROVED.** The user agreed with computing a legitimate root's effective
permission coverage from its applicable registered application capabilities.
Automatic publication of new root grant revisions and adoption by root
assignments is not the chosen mechanism for capability growth.

```text
Applicable registered HRMS catalog: read, write
Effective root permissions:         read, write
Ordinary child selects:             read

Applicable catalog gains delete
Effective root permissions:         read, write, delete
Ordinary child still selects:       read
```

The registered source must belong to the root's trusted application binding.
It is not every permission registered anywhere in Auth. A caller cannot make
an ordinary grant into a computed root by omitting its parent, selecting a
different catalog, or claiming platform administrator status.

**Rationale:** roots represent the application's maximum source ceiling.
Computing that permission coverage avoids generating root revisions and changing
root assignments merely to reflect legitimate application capability additions.
It directly satisfies Q-120A's automatic-growth requirement without a separate
manual root update or adoption gate. Q-121 supplies the responsible publishing
authority; it does not give publishers tenant business access.

**Core-philosophy check:** dynamic permission coverage is explicit root behavior,
not permission to rewrite immutable grant content. Ordinary grants and roles
retain explicit permission selection and revision adoption. Parent/team ceilings,
scope AND composition, membership, assignment and grant enablement, validity,
and human/proxy restrictions still apply. Catalog growth does not enlarge scope
or reactivate an otherwise ineffective route.

**Trade-off and dependency:** effective root permissions depend on authoritative
registration state. Required catalog/binding evidence must be established; a
lookup failure cannot produce allow. This follows the existing fail-closed
evaluation rule, not a new error code or cache policy. The catalog version,
freshness, and exact publication-to-installation/root binding remain open.

### Representation and remaining boundaries

This approval selects the computation mechanism, not a wire format. No stored
`*`, new permission-source field, root flag, or implicit omitted-field behavior
is approved. Q-119 parent omission for trusted roots remains agreed; its explicit
permission-list example does not alone encode a computed catalog source.
The complete root record and resolved representation still need review against
Q-118's existing direct/role variants. Do not silently reinterpret an ordinary
stored `permissions` list as a dynamic catalog reference.

Capability removal/retirement, changing scope definitions, which release is
applicable to each installation, and concurrent visibility are not settled by
the additive-growth example. Existing assignment/validity rules are not waived.
The content/revision mechanism for changes other than catalog permission growth
is not specified by this decision.

### Earlier positions retained below

The following Q-120A/Q-121 snapshots predate Q-122. Their statements that the
computed-versus-materialized mechanism is undecided are superseded by the
approval above; their rationale and unresolved trust/encoding questions remain.
The initial Q-120 manual-update recommendation remains superseded as previously
marked. The alternatives are retained for explanation, not reopened decisions.

**Q-121 approved refinement:** [application platform authority](application-platform-authority.md)
authorizes capability publication outside tenant business scope. The actor is
bounded by existing application-management authority, not required to hold the
new tenant business permission. Earlier unspecified-updater wording below is
narrowed by this decision; precise bindings and root-growth mechanism remain open.

## Q-120A — Automatic root growth requirement, agreed direction

**USER DIRECTION ACCEPTED; MECHANISM OPEN.** The user clarified: “the point is
the root expansion cannot be deliberate, lets accept this? ... live* or not is
a different thing.” This means no separate manual root-expansion decision for
each legitimate application capability upgrade, not untrusted registration or
permission to bypass tenant/application ownership.

Applicable root permission coverage must grow automatically with legitimate new
application permissions within its already-authorized tenant/application boundary.
An advisory list awaiting another manual root approval would not meet that need.
The new coverage becomes usable through an otherwise valid root route; it does
not re-enable disabled/expired/deleted authority or restore absent membership.

Live wildcard, a computed root set, or automated explicit revision/publication/
adoption are possible mechanisms, **not selected contracts**. No stored `*`
syntax is approved. Q-058 remains the wire-format rule until explicitly changed.

| Event | Intended consequence |
|---|---|
| Valid application upgrade introduces and registers payroll-delete within its established boundary | Applicable root coverage gains delete automatically; no separate manual root expansion/adoption gate. |
| Existing child explicitly selects payroll-read | Still read-only; a larger parent ceiling does not select new child permissions. |
| An ordinary referenced role publishes a new revision | Ordinary role/grant adoption rules remain; this is not a return to live roles. |
| Root/assignment disabled or required membership absent | No activation or recreated membership merely because capabilities grow. |
| Unrelated application/tenant or unauthorized caller supplies a permission | No authority to expand this root beyond its trusted boundary. |

**Rationale:** the root is the application's maximum source ceiling. Requiring
a manual edit for each legitimate capability release creates recurring operational
maintenance and can leave new features unavailable for distribution. Accepting
automatic growth does not choose its encoding.

**Core-philosophy qualification:** the earlier requirement that root coverage
remain frozen until a separate manual approval/adoption is superseded for this
case. The original Q-120 recommendation below is not adopted in that respect.
Ordinary child selection, parent/team ceilings, AND scope, and adoption remain.
This does not authorize editing immutable published records in place; root
revision/representation mechanics still need resolution.

**Security consequence:** the trusted capability-update/registration flow now
influences effective root authority. For this path, registration is not merely
inert catalog maintenance. Who may introduce capabilities, which tenant/app
root they affect, and how platform/customer boundaries are enforced must be
specified. Generic registration permission is not automatically a completed
root-updater trust contract. The rule does not claim those safeguards exist.

Still open: live versus materialized representation, automated root revision
mechanics, capability ownership/binding, removal behavior, concurrency/visibility,
and trusted updater evidence. Q-117 remains parked; Q-116 bootstrap replay remains
non-mutating. No additional full checkpoint is closed by this behavioral direction.

<details>
<summary>History — initial Q-120 exploration; separate deliberate root-update recommendation superseded by Q-120A</summary>

**DISCUSSION / NOT APPROVED.** After approving Q-119, the user asked how
permissions added by a later application release reach the root, and suggested
possibly supporting `*` only in root grants. This is a real root-evolution gap,
not permission to change the approved v1 wildcard rule during reconciliation.
Q-058, Q-089-B, and Q-102–Q-107 remain current unless explicitly superseded.

## The underlying problem

An application's first release registers payroll read/write. Its intended root
authority explicitly includes those permissions. A later release adds payroll
delete. Registration makes delete a valid permission identifier; it does not
by itself give any human/group a grant for it. Initial maximum authority under
Q-114 is not an already-approved promise of automatic future permission growth.

There is also an issuance problem: an ordinary administrator whose whole source
ceiling contains only read/write cannot manufacture delete by editing a grant.
Q-093/Q-100 still apply. The authority to establish or expand the root source
requires its own governed trust procedure; merely possessing permission to
register definitions, publish a role, or administer grants is not enough.

## Recommended direction — explicit root revisions, not live wildcard expansion

1. The new application release registers its new permission under the relevant
   registration contract. Existing authority is unchanged by that registration.
2. A **governed root-authority update**, explicitly authorized under its still-to-be-
   defined trust procedure, publishes new immutable root content listing the
   intended added permissions. This is not an ordinary derived-grant subset check
   pretending the old root already possessed a newly introduced permission.
3. The administrators-group assignment explicitly adopts the new root revision
   through the authorized update/adoption workflow. Publication alone does not
   change existing assignments; latest-only and checked-state guarantees remain.
4. Any downstream role/grant that needs the new permission must deliberately
   select it through its normal validated revision/adoption process. An explicit
   read-only child does not acquire delete merely because its root now has delete.

This is a proposal for root evolution, not approval of a new root-update API,
permission name, trust identity, transaction scheme, or recovery procedure.
It is also not replay of completed bootstrap under Q-116.

## Concrete record progression using existing fields

Assume these are the payroll portion of the seed authority, with other required
initial administrative permissions supplied by the full bundle. Tenant remains
implicit and all listed permissions are registered before the corresponding
content is accepted. Existing role/direct-source exclusivity applies.

Root content before the new permission:

```json
{
  "version": "1",
  "grant_id": "G0-PAYROLL",
  "revision": 1,
  "permissions": [
    "hrms:payroll:payslip::read",
    "hrms:payroll:payslip::write"
  ],
  "scope": {}
}
```

Proposed governed root update after delete is registered:

```json
{
  "version": "1",
  "grant_id": "G0-PAYROLL",
  "revision": 2,
  "permissions": [
    "hrms:payroll:payslip::read",
    "hrms:payroll:payslip::write",
    "hrms:payroll:payslip::delete"
  ],
  "scope": {}
}
```

| State | Administrators' assignment | Effective permission change through this root |
|---|---|---|
| Application registers delete | Still adopts root revision 1 | None. Registration is not a grant. |
| Authorized root update publishes revision 2 | Still adopts root revision 1 | None from publication alone. |
| Authorized assignment adoption selects revision 2 | Adopts revision 2, with all current checks satisfied | Delete becomes available through this root route. |
| Existing explicit read-only child remains unchanged | Still selects read only | No automatic delete permission. |

Application release version, contract `version: "1"`, and grant `revision: 2`
are different concepts. Adding a registered permission does not automatically
change the contract format version or every assignment's adoption.

## Root-only wildcard alternative

A **live** wildcard could deliberately mean all present and future permissions
within some defined application/tenant boundary. Its operational appeal is that
root permission lists need not be updated for each new registration. Limiting it
to roots avoids directly introducing wildcard child grants, but does not remove
the root expansion effect: existing root assignments gain the new permission
without adopting changed authority content.

That would be a deliberate exception to both no-wildcard v1 behavior and the
current no-silent-expansion/adoption model. It couples permission-registration
authority to growth of root recipients' access, even if a registrant cannot
otherwise assign new authority. The registration service, tenant/application
binding, and customer/platform authority separation would need explicit review.
It need not imply cross-tenant access, but tenant isolation alone does not resolve
this future-permission expansion within the tenant.

The wildcard also does not automatically answer who may add permissions to
downstream grants; child selection and source/administrative checks still apply.
It must not be advertised as solving every distribution or upgrade step.

A separate convenience option is to generate an explicit candidate permission
list from the registry for review and revision publication. Such a helper can
reduce list-maintenance work without storing a live wildcard. It is not a new
canonical syntax or permission to automatically publish/adopt future additions.
No helper or root-specific expansion tool is adopted in this discussion.

## Rationale, trade-offs, and remaining blocker

**Recommendation:** retain explicit root permissions and deliberate root revision
updates/adoption. This preserves registered vocabulary versus granted authority,
immutable content, explicit recipient adoption, stable existing assignments,
parent/team ceilings, and the Auth administrative/source distinction.

**Trade-off:** new capabilities require a root upgrade step before they can be
distributed from that root. Authority to perform that expansion cannot be inferred
from the ordinary administrator's old bounded grants. The exact trusted updater,
tenant/application consent and bounds, evidence, and root-update endpoint or
provisioning contract remain an explicit blocker—not solved by writing “authorized.”
No permanent bootstrap-user bypass, blanket platform-to-customer privilege, or
automatic permission increase is proposed as the missing trust mechanism.

**Q-120:** prefer explicit new root revisions and deliberate adoption for new
application permissions, keeping live `*` out of root grants, with governed
root-expansion authorization specified separately?

Sources: [registration](application-registration.md), [Q-058](permission-model.md),
[trusted root establishment](bootstrap-authority.md), [Q-119](root-grant-format.md),
[immutable adoption](grant-revisions.md), [Auth's two checks](auth-service-authority-gate.md).

</details>
