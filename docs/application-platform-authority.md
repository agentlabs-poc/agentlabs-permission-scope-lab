# Application platform authority — Q-121

**Bootstrap follow-up:** [Q-117 is now approved](bootstrap-initial-assignment.md);
initial authority is unavailable before complete validated establishment.
The parked-status reference in the earlier snapshot below is superseded.

**Q-123 correction:** [one shared version/catalog per application](root-permission-evolution.md)
serves everyone. Earlier per-installation version-selection gaps below are
superseded; source/application binding is still required, tenant release pins are not.

**Q-122 follow-up:** [computed root permission coverage](root-permission-evolution.md)
now settles the mechanism left open by this Q-121 snapshot. The publishing
responsibility below is unchanged. Exact encoding and applicable catalog/version
binding remain open; automatic materialized root revisions are not selected.

**APPROVED.** The user approved using the existing application platform
administrator to authorize publication of application permission and scope
definitions, rather than introducing an unrelated publisher administrator or
requiring tenant business grants to perform that publication.

## Responsibility and boundary

Auth checks the acting administrator's application-management authority for
the application being updated. This is a platform administration operation,
outside a customer's tenant-level HRMS business scope. Outside that scope does
not mean unbounded: the operation remains constrained by the administrator's
actual platform-management authority and the application binding.

| Responsibility | Governing boundary |
|---|---|
| Auth platform administration | Auth platform governance and infrastructure; not implicit tenant business access. |
| Application platform administration | Application definitions, capabilities, and releases within the administrator's authorized management reach. |
| Tenant application administration | The particular tenant installation and its authorized administration. |
| Application business access | Valid tenant grants, assignments, membership, lineage, and the endpoint's enforced permission/scope requirements. |

These are responsibility distinctions, not new canonical recipient types or
an assertion that a single person cannot hold several explicit authorities.
An application platform administrator need not be restricted to just HRMS;
their actual administrative authority determines which applications they manage.

The existing trusted tenant boundary remains mandatory for tenant operations.
It must not be invented for platform publication or confused with a department
scope. Exact canonical platform-context and grant representations remain open.
Automated actors are not granted independent authority by this distinction.

## Concrete example

1. An application platform administrator publishes HRMS's new
   `hrms:payroll:payslip::delete` permission.
2. Auth establishes that the administrator may update the HRMS application.
   Holding a tenant's payslip-delete permission is not the required authority
   for defining the capability.
3. The accepted capability update feeds automatic permission growth of the
   applicable HRMS root under Q-120A, without a separate manual root-expansion
   approval. Root-growth representation and application-to-installation update
   binding still need their contracts.
4. Existing child grants do not automatically select delete. Ordinary revision
   publication/adoption and all inherited boundaries continue to apply.

Publication does not give the administrator access to a tenant's payslips,
install or enable an application for a tenant, create membership, or re-enable
disabled/expired authority. Nor may an HRMS update affect an unrelated root.
The root update path is not an ordinary child grant pretending its old parent
already supplied a newly introduced permission.

## Rationale and philosophy check

The user identified the missing context: Auth already distinguishes platform
application management from tenant application administration. The actor who
defines what the application supports is not acting as a tenant employee or
distributing authority from their personal HRMS business grants.

Reusing this existing responsibility avoids inventing another administrator
category. Checking the correct administrative authority avoids a circular
requirement that a publisher possess a business permission before that permission
can exist. Explicit management bounds prevent this separation from becoming a
bypass. Tenant access still requires its own valid authority, and child grants
remain within parent permission/scope ceilings.

No new JSON field, permission spelling, scope key, wildcard, or independent
automation identity is approved. Approval settles this responsibility separation;
it does not certify that the existing service implements the new handbook model.

## Existing-service evidence, not an imported contract

Read-only inspection used the local `agentlabs-auth-local` checkout at
`64d084d` (the existing Auth service, not this lab's implementation):

- `internal/adminauthz/service.go` defines platform, publisher, tenant, and
  application administration scopes; `platform.application.write` supports
  platform application management. Publisher-bound administration is narrower,
  while explicit platform application authority can cross publisher boundaries.
- `internal/routecontract/catalog.go` binds application-release creation to
  platform application-write authority.
- `docs/architecture/decisions/application-definition-vs-tenant-installation.md`
  separates a global application definition from each tenant installation.

These observations support the administrative distinction. They do not import
the old capability grammar, database bindings, token formats, or bootstrap and
approval policies into the new canonical handbook without review.

## History — earlier Q-121 proposal not adopted

The assistant initially proposed a tenant-shaped grant named
`G-HRMS-PUBLISHER`, parent `G-AUTH-ADMIN-ROOT`, permission
`auth:application:capability::write`, and scope `{"application":"hrms"}`.
The intent was to separate publication authority from ordinary grant assignment.
However, that proposal had not first accounted for the existing application
platform administrator or their non-tenant administration boundary.

The user corrected the responsible authority and explicitly noted that it sits
outside the tenant and HRMS business scope. Q-121 as approved reuses that existing
authority. The earlier grant example, permission name, and scope key are not
canonical and must not be promoted by this approval.

## Remaining work

Automatic-root computation versus materialized revisions remains D02. Initial
platform authority establishment, exact management-authority evidence, binding
of accepted capability updates to applicable roots/installations, capability
removal, and versioned contracts remain visible in D01/D04/C04. Q-117 remains
parked. This decision narrows D01; it does not complete that whole topic or a
whole closure criterion.

Sources in the lab: [Q-120A](root-permission-evolution.md),
[registration](application-registration.md), [Auth's two responsibilities](auth-service-authority-gate.md),
and [remaining agenda](discussion-assessment.md).
