# ABV-123-07 — Auth-model reconciliation and contract design

**Status: proposed refinements; existing auth decisions remain authoritative.**
User direction: design a plan and reconcile 123 against the auth model we have.
No new permission identifier, endpoint contract, record schema, migration or
runtime implementation is approved by this document. The 123 direction is pinned;
conflicting details must be corrected explicitly, not treated as exceptions.

## Decision summary

Keep the five-table storage direction and shared read/validate/execute mechanics.
Do **not** equate these families with four universal authorization endpoints.
Recommend statically bound operation/kind routes, each declaring one permission,
using shared transport code and existing typed operations. There remain distinct
operations even if their parsing, routing and storage mechanisms are shared.

The earlier generic `/execute` and `/records/{kind}` suggestions did not establish
a safe one-permission mapping. Retain them as the historical family notation;
the concrete routes below are a proposed refinement, not silently approved aliases.

Rationale: fewer handwritten handlers and tables are useful. Fewer visible URLs
are not useful if they hide different powers behind one broad permission. Do not
invent an all-powerful `abv.execute` grant or derive an endpoint permission from
arbitrary body fields. A static registered operation may share implementation
with another operation without sharing its authorization policy.

## Reconciliation against approved rules

| Auth-model rule / source | 123 consequence |
|---|---|
| Exactly one permission and declared inputs/sources per endpoint — Q-049, Q-050; docs/endpoint-policy-format.md | Each concrete registered operation/kind route has a fixed policy. A registry cannot accept unknown kinds/actions or weaken inputs. |
| Resolve once at endpoint with complete material; no prepared handoff — docs/endpoint-authorization.md | Shared wrappers establish the declared material and complete the check before protected output/effects. No unfinished validate result authorizes execute. |
| Administrative authority is separate from ability to distribute authority — Q-093/Q-100; docs/assignment-authority.md, docs/auth-service-authority-gate.md | Auth Evaluator checks the operation's administrative permission/scope. ABV validates the proposed authority and affected dependencies. Both remain inside Auth for its own mutations. |
| Scope is an AND boundary; child permission subset — docs/grant-record-reference.md | Storage folding cannot overwrite inherited predicates or combine unrelated routes. Arbitrary scope interpretation remains outside Auth. |
| Tenant/application authority and application-wide catalog are distinct — Q-121–Q-123; docs/root-permission-evolution.md | Context is explicit; app-wide records cannot be accessed through a fabricated tenant or a wildcard tenant. |
| Immutable content, explicit adoption, current actual parent support — Q-102–Q-109 | Keep controls and revisions as separate record kinds; latest only on create/upgrade, actual adopted revisions during resolution. Generic state cannot replace these rules. |
| Assignment uniqueness and bottom-up dependency guards — Q-101/Q-104 | Keep exact recipient uniqueness and indexed reverse discovery, including disabled bridges and all affected branches. |
| Teams are human membership, not departments or ownership — docs/team-administration.md | Team writes can manage membership under the approved authority; assigning grants remains a separate operation. Membership, ownership and source authority are not conflated. |
| Explicit legitimate root trust — docs/bootstrap-authority.md | Root-trust record insertion is not an ordinary generic write or inferred from absent parent. Preserve bootstrap authority and installation prerequisites. |
| Proxies remain human-dependent; no proxy chains — docs/identity-context.md, docs/delegation-lifecycle.md | New transport accepts only trusted identity context and supported execution paths; it cannot make unsupported proxy paths work by storing JSON. |
| Conflict stops protected Auth write — Q-110; docs/auth-write-consistency.md | Evidence, revision selection and persistence stay coherent. No stale validation ticket or silent substitution/retry. |
| Allow/deny distinct from evaluation failure; reasons required — Q-051 onward; docs/decision-results.md | Contract errors preserve these meanings and approved message fields. Do not manufacture HTTP/result schemas from Go error values alone. |
| Canonical versions and unchanged meaning — Q-107/Q-118/Q-126 | Existing record JSON is reused. Any new envelope has its own version; URL v1 does not replace record version. No permission aliases or semantic repurposing. |

This is an architectural coverage map, not evidence that every handbook path is
implemented. Direct-human recipient support discovery and proxy enforcement remain
outside the current prototype's implemented subset; rejection is not lost canonical
support. They must not be silently marked complete or erased by consolidation.

## Proposed context and endpoint binding

Keep the pinned tenant prefix `/api/v1/{tenant}/abv/`. Recommendation: establish
application explicitly in a path segment under it, rather than adding a default
or a new implicit header convention:

```text
Tenant/application operations:
  /api/v1/{tenant}/abv/applications/{application}/...

Application-wide catalog administration:
  /api/v1/applications/{application}/abv/...
```

Both are proposals. Path values select the requested context; they are not proof
of authority. The endpoint must establish the application/installation and check
the caller's authority there. Target application context does not change the
namespace/meaning of the Auth administrative permission. No new inner tenant or
application scope field is added to grant JSON.

Concrete route examples beneath the tenant/application prefix:

```text
POST execute/assignment.create
POST execute/grant.publish
POST execute/grant.status
POST execute/assignment.status
POST execute/role.publish
POST validate/assignment.create
GET  records/assignment/{id}
GET  records/grant/{id}
GET  records/assignment
```

These are separately registered endpoint policies, not a free-form
`POST execute/{anything}` policy that selects permission from request data.
Record-kind routes follow the same rule. Multiple routes may deliberately use
the same permission when their approved semantics warrant it; one permission
per endpoint does not mean one new permission per operation.

Do not add request operation fields duplicating the concrete route. Reuse the
existing versioned record where that fully specifies the proposal. Transient
support context remains operation input, never a persisted issuer dependency.

### One complete existing request-body example

Proposed route: `POST /api/v1/acme/abv/applications/hrms/execute/assignment.create`.
Request body reuses the approved Q-107 assignment shape, not a new envelope:

```json
{
  "version": "1",
  "id": "A2",
  "grant_id": "G2",
  "grant_revision": 2,
  "recipient": {"type": "group", "id": "Team2"},
  "status": "enabled"
}
```

`auth:assignment::create` is already illustrated in Q-093. Its administrative
scope must permit assignment to Team2. The body does not establish Maya's identity
or prove she has G2's underlying authority. The trusted adapter establishes Maya;
Auth checks her administrative grant; ABV checks registration, latest selection,
actual required supporting team route, her source authority and all relevant
constraints in the protected transaction. It writes A2 only after all pass.

The body example does not settle nested endpoint-input notation or response JSON.
Do not add unapproved `inputs` syntax for `recipient.id`: settle its exact binding
once, using the existing endpoint-policy model rather than a second policy DSL.

## Operation inventory — no hidden completion claims

| Operation family | Existing implementation | Reconciled contract work |
|---|---|---|
| Permission/scope registration | Add-only Go/SQLite/CLI | App-platform authority, exact wire definitions and read visibility; keep support validation mandatory when enabled. |
| Role publication | Immutable Go/SQLite/CLI | Publish capability stays separate from assignment; exact versioned role envelope not inferred from internal Go projection. |
| Grant publication | Existing non-root identity, same parent | Preserve two gates and transient supporting context; does not create a new grant identity or bootstrap a root. |
| Assignment creation | Team-recipient prototype | Reuse Q-093 authority and canonical assignment; unsupported recipient routes remain explicit. |
| Grant/assignment status | Separate protected controls | No collapsed generic record status; retain their different affected-binding rules. |
| Assignment diagnosis | Read-only structural diagnosis | Protected visibility; cannot be advertised as a reusable authorization or write guarantee. |
| Inspect | Existing lab inspection | Kind-specific read authority and output boundaries; no automatic permission to inspect everything from ability to execute. |
| Listings/upgrade candidates | Not complete | Fixed filters, keyset pagination, visibility, advisory-only latest revision information. |
| Explicit adoption | Not implemented | Latest-only selection and complete resulting/affected authority; no implicit enablement. |
| Permission retirement/further definition lifecycle | Not complete | Preserve approved retirement semantics; do not infer generic disable/delete rules. |
| Team/membership management | Records and resolution exist, general protected CRUD does not | Q-092 group create/write permissions already defined; general CRUD delivery is a separate scope choice, not an incidental storage task. |
| Bootstrap/new grant identity creation | Controlled fixtures are not general APIs | Preserve legitimate minimal bootstrap, registration and root rules; expose no new operation by generic dispatch. |

Do not invent permission strings to fill rows: for each concrete operation,
reference an approved permission or mark the mapping proposed and obtain approval.
Q-092 already associates membership management with `auth:group::write`; do not
resurrect separate membership permissions merely because separate records exist.
Endpoint design does not authorize production Auth integration, deletion,
reparenting or PostgreSQL work deferred/excluded by the user.

## Storage reconciliation

Five physical tables remain the target, subject to the performance/integrity proof:
one typed L1 record store; assignments; teams; memberships; schema metadata.
Co-location changes none of the canonical record identities, immutability,
validity, root-trust provenance or authority boundaries above. Canonical JSON
must not get new key/state/recipient fields to fit a generic envelope.

Index exact record identities and hot relationships; fetch complete operation
evidence, not whole unrelated-area state. Partial evidence must not be represented
as the complete Snapshot currently passed to administration callbacks. That
provider/callback contract is a prerequisite for targeted writes, not an optional
optimization. Preserve the full-read path until equivalence is established.

## Bounded plan to finish the design, then prove it

| Step | Maximum first attempt | Exit artifact |
|---|---:|---|
| R1: reconcile and expose conflicts (this document) | 20 min | Existing rules mapped, unsupported/incomplete capabilities distinguished, route/context refinement proposed. |
| R2: operation permission/material matrix | 25 min | Every current operation maps to one declared permission, explicit sources, appropriate ABV checks and output visibility. Reuse approved IDs; unresolved mappings are decisions, not guessed code. |
| R3: wire examples and behavior | 25 min | Complete versioned success/denial/error examples per distinct shape; canonical record reuse, revision selection, bounded list filters and explicit conflict/retry rules. |
| R4: storage keys and indexes | 25 min | Exact one-store outer-boundary key, numeric revisions, constraint/reference map, reverse queries and administration evidence contract. |
| R5: bounded SQLite comparison | Use S2–S4 in abv-123-storage-plan.md | Current loader vs indexed current schema vs indexed consolidated schema; equivalent authorization outcomes, measured reads/writes/space. |

R2–R4 are design tasks, not additional feature implementation. Each stops on its
artifact or cap; one 10-minute correction maximum before regrouping. No automatic
extension, repeated review loops or speculative framework. Ask about only genuine
new policy choices. User approval of the concrete route/context and permission
matrix precedes replacement contracts. A failed performance/integrity comparison
retains the old implementation. Migration needs its own reversible approved plan.

Functional design is complete only when every in-scope operation has a defined
permission/material/ABV/result contract and all required capabilities are accounted
for as implemented, planned, deferred or excluded. Counting endpoint families or
tables does not meet that condition. Existing authorization decisions are not
superseded by implementation convenience.
