# CP5-A — Application registration and computed root coverage

Status: CP5-P01 and written specification approved for execution by “keep moving”.
This is an implementation design, not a new canonical wire contract.

## Goal and bounds

Deliver add-only permission/scope registration for an existing application,
through reusable ABV and the in-process lab CLI. Legitimate roots compute their
permission ceiling from that application's active catalog. Ordinary grants keep
their explicit selections and adopted revisions.

The user subsequently requested working implementation first with no review
passes. Retain Sol-medium coding and verification, but do not dispatch reviewers.
Execution uses independently testable units: catalog writes, root computation,
then CLI acceptance. Catalog writes are split into persistence and protected
coordination in the [provider plan](abv-cp5a-provider-plan.md) to clarify ownership. Each
unit receives a concrete time cap in its execution plan; a missed cap triggers
reassessment, not an automatic extension. No runtime work is claimed here.

## Sources and rationale

- [Registration](../../docs/application-registration.md): Q-039–Q-042 require
  registered definitions and application-wide compatibility validation.
- [Platform authority](../../docs/application-platform-authority.md): Q-121
  separates application management from tenant business access.
- [Root evolution](../../docs/root-permission-evolution.md): Q-122/Q-123 compute
  roots from one shared application catalog; children do not silently expand.

Reuse the existing catalog projections, permission grammar, token/key checks,
SQLite tables, bounded reads and transaction pattern. Do not introduce a second
catalog, fake tenant, new dependency, wildcard grant encoding or business resolver.

## Request flow and responsibility

```text
CLI: explicit application + lab publisher context + definition
  -> reusable catalog operation
  -> SQLite write transaction / application catalog snapshot
     -> independent application-administration check
     -> ABV definition and compatibility validation
     -> insert definition (and permission's supported keys atomically)
  -> commit, then return success

Tenant operation: explicit tenant + application
  -> existing administrative gate and lineage validation
     -> trusted root reads the same application's active catalog
     -> child retains its exact permission selection and AND scope narrowing
```

Catalog operations use a validated application-only context. Existing `Area`
remains mandatory for tenant operations; its zero value never becomes valid.
The catalog provider path is SQL-free above the SQLite adapter and reads/writes
only the selected application. It must not construct a placeholder tenant.

Application-administration approval is required inside the protected operation,
before persistence. The lab supplies an explicitly application-bounded publisher
premise, separate from Maya's tenant/team administration. Missing capability,
wrong application or rejected authority fails closed. Lab identity is not real
authentication and tenant membership cannot imply catalog publishing authority.
Evidence supplied to an administration adapter must not let it mutate the
snapshot later used by ABV validation.

## Add-only registration behavior

- The application must already exist; no implicit application/bootstrap creation.
- Register one new active permission, or one new scope definition, per operation.
- Reuse current permission syntax and scope-key/token validation. Only currently
  supported tokens are accepted; application business meaning is not interpreted.
- A permission may include its initial supported scope keys. Every supplied key
  must be registered and unique, even when compatibility enforcement is disabled.
- Preserve the application's existing explicit compatibility mode. When enabled,
  the registered support list governs subsequent grant checks. An empty list
  supports no scope keys; it is not a wildcard. Existing declarations are unchanged.
- Register scope keys before permission declarations that reference them.
- Duplicate identifiers conflict, including retired permission identifiers; never
  overwrite, reactivate, rename, or silently change meaning through registration.
- Validation, authorization, cancellation, storage errors and conflicts produce
  no partial write. Permission plus support rows are one atomic insertion.
- Reuse SQLite write serialization so tenant checks see a coherent catalog
  snapshot, not half a publication. Apply existing bounded-load conventions.

The CLI uses explicit arguments and labeled output for internal projections.
It must not publish a guessed canonical registration JSON/YAML schema. Existing
versioned grant and assignment contracts stay unchanged.

## Computed roots

Only the existing trusted-root evidence, actual root team and valid root lineage
establish root status. Omitting `parent_grant_id` is insufficient. Compute a
deterministically ordered permission set from active definitions in that root's
application catalog. Do not rewrite stored root content or assignment adoption.

Preserve root scope, validity, grant/assignment enablement and tenant isolation.
Catalog growth does not create or enable a root, grant publisher business access,
or bypass compatibility checks. Missing/malformed evidence fails closed.
Ordinary direct/role selections still resolve exactly as before. All existing
root-resolution callers must be checked for assumptions about the stored list.

Example: the catalog gains `hrms:payroll:payslip::export`. Otherwise valid HRMS
roots in tenants A and B gain that ceiling. A child selecting only
`hrms:payroll:payslip::read` stays read-only. An unrelated application's root does
not gain export. A disabled root remains ineffective.

## Acceptance matrix

| Case | Required result |
|---|---|
| Authorized new permission/scope; close and reopen SQLite | Definition persists |
| Missing publisher capability or wrong application | Reject; database unchanged |
| Tenant administrator without publisher premise | Reject catalog mutation |
| Malformed ID, token, duplicate key or unregistered supported key | Reject; no partial rows |
| Duplicate permission/scope registration, including concurrent attempts | At most one success; no overwrite |
| Cancellation or injected write failure | Rollback; no success output |
| Compatibility enabled/disabled | Mode and existing declarations unchanged |
| New permission in one shared application catalog | Its valid tenant roots reflect it |
| Existing ordinary child grant and assignment | Stored records and authority unchanged |
| Parentless untrusted grant or root in another application | No computed coverage |
| Disabled/expired/missing-support root route | No activation or bypass |
| Root scope and adopted revisions after catalog growth | Unchanged |

Use existing Go test infrastructure, temporary SQLite databases and CLI harness.
Run focused red/green tests per unit, then full tests, race, vet and build before
claiming slice completion. Tests must exercise both authority gates and no-write
failures. Independent review is paused by the user, not claimed complete.
No performance project or external service is necessary.

## Explicitly outside this slice

Role publication is the next CP5 unit, not completed by registration. Permission
retirement, scope-definition edits and compatibility-mode changes need separate
bounded coverage; add-only APIs do not implicitly implement them. Root retirement
semantics remain as canonically agreed, not superseded by this additive slice.

CP4 publication/adoption remains pending. PostgreSQL, deletion and reparenting
remain user-deferred. Production Auth-service integration remains outside this
build. No canonical root encoding or platform grant schema is decided here.
