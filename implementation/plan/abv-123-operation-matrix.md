# ABV-123 D1 — operation, permission, material and result matrix

**Status: design draft; no mapping or route below is newly approved.** This
inventory reconciles the nine current `abv.Facade` methods with approved Auth
rules. `Implemented` means the in-process Go/SQLite prototype exists; it does
not mean an HTTP contract, production Auth Evaluator, or public JSON envelope
exists. The authority-boundary validator remains the core ABV responsibility;
endpoint binding, identity establishment, the Auth Evaluator call and response
translation remain controller/integration responsibilities.

Facade inventory: [`abv.go`](../abv/abv.go),
[`catalog.go`](../abv/catalog.go), [`role.go`](../abv/role.go), and
[`grant_revision.go`](../abv/grant_revision.go). The corresponding coordinator
checks are in [`internal/mutation/`](../abv/internal/mutation/); these sources,
not the CLI command count, establish the nine implemented semantics below.

## Rules applied to every row

- A concrete endpoint has one literal, statically registered permission and
  declared material. A route variable or body field must not select either the
  operation or permission. Different endpoints may deliberately reuse one
  approved permission when their semantics warrant it.
- The Auth Evaluator answers whether the actor may perform the administrative
  operation inside its administrative scope. The ABV core validator separately
  checks the proposed/resulting authority and its dependencies where the
  operation can distribute or restore authority. Neither result substitutes for
  the other.
- Tenant/application is trusted outer context. IDs in a request locate records;
  they do not prove identity, ownership, membership, support or current state.
- Protected writes use coherent evidence through persistence. A deny or
  evaluation error produces no write; errors are not denials. A diagnostic or
  inspection result is never a reusable authorization ticket.

Sources: [one-permission endpoint policy](../../docs/endpoint-policy-format.md),
[two-check Auth gate](../../docs/auth-service-authority-gate.md),
[core validator material and operation rules](../../docs/authority-boundary-validation.md),
[write consistency](../../docs/auth-write-consistency.md), and
[decision/result semantics](../../docs/decision-results.md).

## Current facade matrix

Permission state is deliberately explicit: **approved** means the literal is
present in an approved source decision; **pending** means only the required
authority responsibility is canonical. Pending labels below are descriptions,
not proposed permission identifiers.

| Facade operation and proposed concrete route | Delivery state | Declared/trusted material | Fixed endpoint permission | Auth Evaluator check (administrative) | Core ABV validator check | Result and disclosure boundary |
|---|---|---|---|---|---|---|
| `CreateAssignment`; tenant/app `POST execute/assignment.create` | **Implemented**, team recipient only | Trusted area and identity; complete canonical assignment; current catalog, controls, exact latest content/adopted role, teams, memberships, parent/support assignments, root evidence, validity and uniqueness state | **Approved:** `auth:assignment::create`, bounded to the actual recipient (`group`) | Actor may create this assignment for the proposed recipient | Validate canonical shape; latest-only selection; registered content; actual parent-team support; actor source access; permission subset and conjunctive scope; recipient team ceiling; validity; no duplicate; same evidence through write | Current facade returns an assignment-ID receipt. A future wire contract may return only its approved bounded representation. Deny/error/malformed/conflict returns no write; do not expose the resolved route as transferable authority. |
| `SetGrantStatus`; tenant/app `POST execute/grant.status` | **Implemented**, non-root grants | Trusted area/identity; complete proposed canonical grant control; current control; all assignments for uniqueness and affected enabled bindings; content, roles, lineage, catalog, teams, memberships, roots and validity | **Pending:** literal grant-status administration permission | Actor may set this grant's requested enabled/disabled state | Reject trusted-root control. Disable is an authorized withdrawal. Enable stages the new control and validates every enabled affected assignment's complete current route and validity; one broken enabled binding blocks the whole write. Explicitly disabled assignments stay disabled | Return the persisted canonical control only after atomic success. No partial-recipient enablement, implicit assignment enablement, revision adoption or topology disclosure. |
| `SetAssignmentStatus`; tenant/app `POST execute/assignment.status` | **Implemented**, team recipient/non-root only | Trusted area/identity; assignment ID and requested status bound to the complete current canonical assignment; reverse dependent assignments; current adopted content, controls, roles, catalog, teams, membership, roots and validity | **Pending:** literal assignment-status administration permission | Actor may change this assignment for its actual recipient and grant | Disable discovers all dependent team assignments and rejects while an enabled dependent remains. Enable validates the current adopted revision, actual parent-team support, full narrowed route and validity. It neither selects latest content nor changes grant status | Return the exact persisted assignment. No implicit adoption, dependent cascade, grant-wide change or partial write. |
| `PublishGrantRevision`; tenant/app `POST execute/grant.publish` | **Implemented**, existing non-root identity with unchanged parent | Trusted area/identity; transient source-assignment ID; complete canonical proposed content; current grant control and all content revisions; catalog, roles, actual source route, assignments, teams, memberships, roots and validity | **Pending:** literal grant-publication administration permission | Actor may publish this revision for this existing grant; the transient source selector is part of the checked operation | Validate immutable new revision, strictly newer number, unchanged existing parent, registered permission/scope or exact role revision, no unsupported `$self` copy, source assignment and actor access, no cycle, complete narrowing and source validity | Return the newly persisted canonical content. Publication does not assign, adopt, enable or create a new grant identity. Source assignment is evidence, never persisted issuer lineage. |
| `PublishRole`; tenant/app `POST execute/role.publish` | **Implemented** as an internal projection, not canonical wire JSON | Trusted area/identity; proposed role ID/revision/permission bundle; current role revisions and application catalog | **Pending:** literal role-publication administration permission | Actor may publish this role ID and complete bundle | Reject duplicate `(role, revision)` and validate the complete bundle against the catalog. The immutable publication itself does not change grants; authority-changing adoption is a separate pending operation requiring whole-result validation | Return only the published role projection until a versioned role wire contract is approved. No implicit adoption, assignment or claim that internal fields are canonical JSON. |
| `RegisterPermission`; application-wide `POST execute/permission.register` | **Implemented** as an internal projection, not canonical wire JSON | Trusted application/identity; complete permission definition and supported scope keys; current shared application catalog and compatibility mode | **Pending:** exact existing application platform-management permission; do not revive the non-canonical `auth:application:capability::write` example | Application platform administrator may update this application's definitions; tenant business authority is irrelevant | Validate definition identity, add-only registration, supported keys and catalog compatibility rules. This is structural/catalog validation, not tenant authority distribution | Return only the accepted definition projection until its versioned public contract is approved. No tenant access, installation/adoption, selective tenant version or implicit grant rewrite. |
| `RegisterScope`; application-wide `POST execute/scope.register` | **Implemented** as an internal projection, not canonical wire JSON | Trusted application/identity; complete scope definition/tokens; current shared application catalog and compatibility mode | **Pending:** exact existing application platform-management permission; recommendation is to reuse the same approved platform application-write authority as permission registration if its canonical contract covers both definition kinds | Same application-bound platform authority check as permission registration | Validate definition identity, token form and add-only catalog consistency. ABV does not infer application-domain meaning or facts | Return only the accepted scope projection until its versioned public contract is approved. No tenant business authority or unrestricted fallback. |
| `CheckAssignment`; tenant/app `POST validate/assignment.create` | **Implemented**, read-only team-assignment diagnosis | Trusted area and complete canonical assignment bytes; current catalog, exact latest content/adopted role, teams, parent/support assignments, controls, roots and validity needed for structural/lineage diagnosis. Current method accepts no identity | **Pending. Recommendation:** statically reuse `auth:assignment::create` for this concrete preflight only if approved visibility matches creation; otherwise approve a distinct diagnostic permission. Never choose from the body | **Missing from current facade.** Controller must establish identity and complete the fixed visibility/administrative check before returning protected details | Current core checks canonical shape, latest content, registration, parent-team route and narrowing only. By design it does **not** establish actor administration or source authority and performs no write | Return a bounded diagnostic. It must say that administration/source authority are unestablished and must not expose unrelated records or authorize later execution. Deny/error returns no diagnostic authority. |
| `Inspect`; one tenant/app `GET records/{fixed-kind}/{id}` registration per kind | **Implemented** as one internal `kind` dispatcher for assignment, grant content, grant control, permission, scope, role, team and membership | Trusted area; endpoint-fixed kind and exact non-wildcard ID; only the authoritative record/projection and integrity context needed for that kind. Current method accepts no identity and reads the whole snapshot | **Pending per concrete kind:** each registered kind route needs one literal read permission. A free-form `records/{kind}` policy is not acceptable | **Missing from current facade.** Controller evaluates the route's fixed kind-specific read permission and scope before calling the wrapper | No authority-change validation. Core verifies area/application consistency, exact identity and stored representation integrity; grant lookup currently returns latest published content, not an adopted revision | Return only the authorized kind/ID representation. Canonical JSON is valid only for assignment/grant content/grant control; other rows are internal projections. No list fallback, wildcard, cross-kind visibility or execute authority follows from inspect. |

The proposed route suffixes above come from the reconciliation examples. The
tenant/application prefixes themselves remain unapproved: tenant operations are
proposed under `/api/v1/{tenant}/abv/applications/{application}/...`, while the
shared application catalog is proposed under
`/api/v1/applications/{application}/abv/...`.

## Canonical record and material boundary

The approved public record shapes available to these operations are the v1 grant
control, immutable grant content, and revision-adopting assignment in
[Q-107](../../docs/grant-revision-format.md), including the approved exclusive
direct-permissions versus `role_id`/`role_revision` variants in
[Q-118](../../docs/role-grant-contract.md). The assignment-create body may reuse
that canonical assignment directly. Grant publication may reuse canonical
content; grant status may reuse canonical control. Identity, tenant/application
binding and the publication source-assignment ID are trusted/transient operation
material, not additions to those records.

`RoleContent`, `PermissionDefinition`, `ScopeDefinition`, `Diagnostic`, `Record`
and `Receipt` are explicitly internal projections in
[`domain/records.go`](../abv/domain/records.go). Their current Go shapes must not
be promoted to approved JSON. Exact nested endpoint-input notation, versioned
success/deny/error envelopes, HTTP mapping, reason codes and safe field-level
disclosure remain D2/controller decisions. Approved semantics still require
allow/deny to remain distinct from evaluation failure, a reason for deny, and
the evaluator-supplied `error_message` plus `error_message_reason` to reach the
UI under the eventual safe wire contract.

## Functional coverage outside the nine methods

| State | Operations/capabilities | Boundary |
|---|---|---|
| **Canonical but pending** | Kind-specific/list reads with bounded filters and keyset pagination; explicit assignment adoption of a newer grant revision (including a role-bearing revision); permission retirement and remaining definition lifecycle; general protected team and membership management; legitimate bootstrap/new grant identity creation; direct-human assignment support discovery | Preserve approved semantics, but do not route any through generic inspect/execute or count storage records as APIs. Team membership uses approved `auth:group::write`; team create/write/delete use their approved distinct permissions. Other literal mappings stay pending. |
| **Deferred for this build** | Deletion operations; parent-change/reparenting; PostgreSQL provider/migration | Preserve plans and bottom-up/lineage rules. Deferral is neither implementation nor permission to emulate them with status or generic writes. |
| **Excluded from this build** | Production external Auth service/evaluator integration | Current administrative ports and fixed lab premises preserve the two-gate seam but do not prove authentication or a production authorization decision. ABV remains reusable and the validator remains core. |

## Decisions still required

1. **Approve or revise concrete context/routes.** Recommendation: keep the two
   explicit application bindings above and statically register every operation
   and inspect kind. This prevents wildcard tenant/application lookup and body-
   selected authorization while retaining shared handlers.
2. **Approve literal permissions for the eight unmapped facade semantics.** Start
   from the canonical Auth/application catalogs, not names invented in this lab.
   Reuse one platform application-write permission for both catalog-definition
   routes only if its approved semantics cover both; otherwise keep them distinct.
   Consider assignment validation and creation sharing `auth:assignment::create`
   because validation is creation preflight, but approve the diagnostic disclosure
   explicitly. Give each inspect-kind route its own fixed read mapping or an
   explicitly approved shared read permission whose scope covers that exact kind.
3. **Specify protected read/diagnostic visibility.** `Inspect` and
   `CheckAssignment` currently have no identity or Auth Evaluator call. Decide
   their administrative material and safe outputs before exposing HTTP routes.
4. **Approve wire/result contracts and nested material binding.** Reuse the three
   approved core records; do not infer the other envelopes from Go structs. Bind
   every declared request value once and return no more evidence than the caller
   is authorized to see.
5. **Define targeted provider evidence without weakening checks.** Current writes
   give administrative callbacks cloned whole snapshots. A consolidated store may
   fetch less only after each row's complete evidence and affected-branch queries
   are represented and equivalence-tested; pagination is never authorization
   evidence.

No middleware API, component rename/merge, new permission string, migration or
runtime replacement is selected by this draft.
