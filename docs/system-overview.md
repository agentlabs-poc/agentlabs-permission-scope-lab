# Authorization system — canonical logical overview

## SVG-001 — follow a request, not a responsibility checklist

For ownership changes rather than request flow, see the separate
[Q-099 ownership/lineage explanation](ownership-lineage.md) and
[ownership SVG](assets/ownership-lineage.svg). Owner rotation does not redesign
this endpoint-owned gate or replace the selected team-held authority source.

The user approved a client-to-handler request-flow layout. The diagram follows
one certificate read: client, authentication middleware, endpoint handler,
embedded Auth Agent, authority loading from Auth, constrained application data
access, and the response back to the client. Middleware, handler, and embedded
agent sit inside the application boundary. Evaluation and enforcement remain
inside one endpoint-owned authorization gate.

This replaces the less readable responsibility-first layout, not the approved
authorization rules. Its [previous SVG](assets/history/authorization-system-pre-svg001.svg)
and [overview text](history/system-overview-pre-svg001.md.txt) remain archived.

**Authority loading updated through Q-090/Q-091.** This diagram represents the approved discussion,
not the historical lab implementation or a verified deployment topology.
The [pre-reconciliation overview](history/reconciliation-2026-09-05/docs/system-overview.md.txt)
and [previous SVG](history/reconciliation-2026-09-05/docs/assets/authorization-system.svg)
are preserved as deprecated history. The
[earlier pre-Q-043 diagram](assets/history/authorization-system-pre-q043.svg)
is also retained.

[Open the shared scalable SVG](assets/authorization-system.svg).

![Certificate-read request from client through middleware, handler, embedded Auth Agent, Auth authority, constrained database access, and back to the client](assets/authorization-system.svg)

### Reading the numbered request flow

1. The client submits the request and token. It does not select the endpoint's
   permission or supply authoritative grant/membership facts.
2. Authentication middleware establishes verified identity and trusted tenant
   context. It does not finish business authorization with an allow/prepared result.
3. The handler binds and validates required inputs from the server-owned endpoint
   policy, then invokes the embedded Auth Agent with the declared permission and
   relevant material.
4. Authority loading supplies valid human memberships, direct/group assignments,
   their reusable grant definitions, permissions from adopted role revisions,
   validity/conditions, and dependencies including supporting parent assignments
   for subgroup-derived routes. Shared loading
   or valid preloaded material may satisfy this; the arrows do not mandate a
   fresh remote call for every request. The embedded agent resolves and evaluates
   complete assigned authority using the shared canonical rules.
5. In this example a valid route establishes certificate-read authority within
   Finance. No valid authority means stop. Finance authority does not establish
   that unchecked C-17 belongs to Finance; the result labels are explanatory,
   not a new decision-result schema.
6. The handler enforces that relationship with a constrained lookup: trusted
   tenant AND authorized Finance department AND requested certificate. This is
   still mandatory endpoint authorization work, not optional business validation.
7. Only matching authorized data can enter the handler's response. A missing or
   out-of-boundary row cannot be disclosed; exact response-status semantics remain
   open. Authentication, input, authority, or enforcement failure must prevent
   protected disclosure/effects.
8. The application returns the response to the client.

This is a concrete logical example, not a verified trace of an existing service.
Applications may instead establish necessary facts before evaluation and ensure
consistent enforcement at actual use. The main diagram uses direct constrained
execution for clarity; no mandatory eager lookup/resolver or separate Auth access
to the application's database is introduced.

The reader imports this exact SVG from `docs/assets`; there is no second
hand-maintained current diagram in `src/content`. The standalone enforcement
trace's different historical illustration is now
[archived source only](history/retired-pages-2026-09-05/README.md), removed from
the active site by user approval.

## One gate, not two business-authorization locations

CONTRACT-006 places business authorization at one endpoint-owned gate.
Pre-handler middleware may authenticate, establish trusted context, or load
authority. It does not hand the endpoint a middleware allow/prepared outcome.

The gate combines the canonical evaluation rules with application-owned meaning
and enforcement. “Auth agent” describes the reusable embedded integration across
both responsibility layers; it does not mandate a middleware-only decision,
separate service, fixed number of network calls, or duplicated per-handler clients.

| Responsibility | What it supplies or guarantees |
|---|---|
| Authentication/context | Verified actor and authorizing human, and trusted tenant context. Caller path/body values do not become trusted facts merely by extraction. |
| Auth / Layer 1 authority | Valid human memberships, direct/group assignments, grant definitions, adopted role revisions, validity/conditions, and required parent-assignment/delegation dependency evidence. |
| Application / Layer 2 | Server-owned endpoint policy, valid request values, registered boundary meanings, facts where needed, and constrained execution. |
| Embedded evaluator | Complete-grant applicability, one required permission, supported scope meaning, and tenant/human/dependency limits. |
| Endpoint enforcement | Actual returned data and mutation effects remain within the authorized boundary and relevant request bindings. |

These are responsibilities, not an approved SDK interface or result schema.
Layer 2 cannot manufacture authority or override canonical limits.

Q-090 separates [grant definitions and assignments](grant-assignments.md).
Q-091 permits [explicitly dependent subgroups](subgroups.md), without membership
inheritance. These change authority representation/loading, not the request path
or the endpoint's single gate. The older direct/group “grant” loading description
is preserved in baseline `0.0.1`; assignment-aware wire contracts remain incomplete.

## Declaration and required inputs

CONTRACT-008/010/011 and INPUT-002/003 establish the current partial policy:

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

Every protected endpoint declares exactly one permission covering its complete
operation. The declaration is server-owned. Missing or invalid declarations
cannot permit protected execution.

Every selected input must be present at its exact declared source, without
silent default or fallback, even under a tenant-wide grant. The application
request contract owns types, nullability, format, and domain validation. The
endpoint policy does not duplicate these with type/nullable fields. Authorization
and execution must use the same validated meaning.

The path tenant must agree with trusted tenant context. Local input names do not
automatically establish scope-key mappings or application relationships.
[GET/PUT examples and rationale](endpoint-policy-format.md) distinguish current
facts from proposed body values.

## Authority resolution is not an independent assignment or an allow

Under RESOLUTION-006 / Q-048, logically obtain Vinay's valid memberships, then
grants for Vinay directly and for those groups. Expand adopted role-revision permissions
as needed while preserving each grant's scope, conditions, validity, provenance,
and dependencies.

Q-089-B supersedes the earlier “current role permissions” wording: publication of
a new revision does not update a grant. An authorized, boundary-validated grant
adoption changes the selected revision. Auth's own administrative APIs, including
publication and adoption, are protected by the same framework. The request flow
above is unchanged; its role-permission source is now revision-pinned. See
[role revisions](role-revisions.md).

Resolved grants are dependent evaluation views. Do not pool the permission of
one grant with the scope of another, or turn group-derived access into an
independent direct grant. Resolution can narrow authority, not amplify it.
This retrieval flow does not mandate a particular transport, cache, or number
of Auth calls. Freshness mechanics remain open.

## Boundary authority and actual execution

CONTRACT-012 deliberately avoids a canonical relationship block, named resolver,
or argument-map language in endpoint policy. It does **not** remove the endpoint's
relationship-enforcement responsibility.

For a Finance certificate-read grant and
`GET /api/v1/acme/FIN/C-17`, Auth can establish read authority within Finance.
That is not unchecked access to C-17. The endpoint must constrain the actual
operation to the trusted tenant AND authorized Finance boundary AND requested
certificate. A constrained query can enforce the relationship directly; a
separate eager application lookup is not mandatory.

Alternatively, the application can establish necessary trusted facts and use
them during evaluation, while keeping actual use consistent with those facts.
Neither pattern permits protected output or effects before the necessary
constraints are enforced. Validation, authority, or required enforcement failure
must prevent protected execution.

With `{"user":"$self"}`, the endpoint must similarly enforce the application's
defined self relationship to the authorizing human. With `{}`, no narrower
department/self boundary arises from that scope; tenant, dependency, validity,
conditions, declared input requirements, and application obligations remain.

**Rationale:** the small canonical format delegates application-specific
relationship enforcement to the code that owns the operation. This is a conscious
trust boundary: Auth alone cannot detect every incorrect handler.

ENFORCEMENT-003 requires review of actual data/effect constraints. Logging Finance,
using it in an ineffective OR clause, or returning an unchecked second lookup
does not enforce Finance containment. This review is necessary but does not
guarantee detection of every security defect.

## Registration is a lifecycle responsibility, not another request stage

Applications register supported permissions and scope contracts; Auth validates
acceptance without interpreting application-domain meaning. Optional registered
permission–scope support validation is explicitly enabled/disabled upfront.
When enabled it applies to all grants, including existing and role-based grants;
activation with incompatible grants is rejected until explicit correction.

This optional registration feature is separate from the endpoint relationship
block that was not adopted. [Registration rationale](application-registration.md).

## What remains open

- Full policy validation/publication beyond the approved partial structure.
- Decision-result, resolved-request/grant transport, and audit contracts.
- Administrative scope encoding and containment proofs.
- List/count/export/bulk and update/move semantics.
- Freshness, revocation, concurrency, dependency-chain mechanics, and SDK APIs.
- Runtime migration and a comprehensive executable conformance suite.

The diagram is a logical explanation, not evidence those branches are finished.
The [discussion tree](discussion-tree.md) is the separate whole-handbook mind map.
