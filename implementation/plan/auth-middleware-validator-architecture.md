# AUTH-ARCH-01 — Auth Middleware and core Auth Validator

**AUTH-ARCH-02 — placement pinned by the user:** ABV/core Auth Validator is used
inside Auth Service itself. Auth Middleware/evaluator and application enforcement
are at API services. Auth Service protects its own APIs through that same API-side
authorization before ABV. Do not embed ABV's mutation engine or storage provider
in every application. The [implementation plan](auth-middleware-implementation-plan.md)
preserves this split; its evidence/SDK details remain proposed until specified.

## Handbook-grounded clarification — current recommendation

After the user requested the consolidated handbook be understood, chapters 5,
6 and 7 and the canonical responsibility definitions were read directly. They
already distinguish the embedded request evaluator from Auth's authority-boundary
validator. **Do not group both under a newly enlarged Auth Validator by default.**
The preliminary grouping below is withdrawn as the current recommendation and
retained only as a proposal history. No canonical handbook change occurred.

Map the user's component names onto those existing responsibilities:

| Component | Existing handbook responsibility |
|---|---|
| Auth Middleware | API-side embedded Auth Agent/evaluator integration: bind the endpoint policy and material, obtain usable authority, evaluate the complete request, stop deny/error and preserve request bindings. |
| Core Auth Validator | Auth's authority-boundary validator (ABV): validate definitions and protected authority changes under canonical registration, source, subset, lineage, revision and lifecycle rules. |
| Endpoint implementation | Establish application facts when needed and constrain actual data/effects. No handler-side grant interpretation. |

Ordinary application request: endpoint-owned gate using middleware/evaluator,
then constrained application execution. It does not call the authority-change
validator simply because it reads a certificate. Auth's own authority-changing
request: the same middleware/evaluator establishes administrative authority,
then core Auth Validator validates the proposed authority before persistence.
The protected Auth coordinator preserves both checks through the write; internal
callers cannot bypass them by skipping HTTP middleware.

The evaluator checks administrative scope, not only an action name. Core ABV
checks a different boundary: the authority being issued or changed. Shared
canonical resolution helpers are appropriate; duplicate semantics or merging
these checks into one undifferentiated success are not.

"Middleware" must support the endpoint-owned gate when needed material is only
available there. It is not a pre-handler final decision followed by prepared
completion. Existing GET/PUT policy shapes and allow/deny/error JSON remain;
no relationship block, returned scope query, universal execute power or new
business-rule engine is introduced.

![Handbook-aligned middleware and validator flows](assets/auth-middleware-validator.svg)

Primary handbook references:
[canonical responsibilities](../../handbook/theory/canonical-terms.md#responsibility-layers),
[canonical model](../../handbook/implementation/05-canonical-model.md),
[Auth administration](../../handbook/implementation/06-auth-service.md),
[application integration](../../handbook/implementation/07-application-integration.md).

<details>
<summary>Earlier preliminary proposal — combined-core grouping withdrawn; examples retained as draft history</summary>

**Status: understanding and architecture proposal for user discussion.**
The user proposes an API-side agent named Auth Middleware and core named Auth
Validator. Component names follow that request. The capability grouping, SDK
entry points and deployment arrangement below are recommendations, not approved
new canonical schemas. No code, migration or published endpoint is changed.

## 1. Recommendation

Two reusable components, with application enforcement still explicit:

- **Auth Middleware:** API integration. Bind a server-owned endpoint declaration,
  establish trusted context, extract selected inputs, coordinate material gathering,
  invoke the core, stop failures and call the handler with the same validated inputs.
- **Auth Validator:** core canonical rules. Resolve current applicable authority
  and evaluate a request; for Auth's own authority-changing operations, separately
  validate the proposed definitions/authority changes and affected dependencies.
- **Endpoint code:** owns application meaning and constrained execution. This is
  not a third authorization engine and does not interpret grants.

The proposed core contains two distinct logical capabilities: the existing
authorization evaluator and the authority-boundary validator (ABV). They can share
resolution primitives but cannot substitute for one another. Calling the whole
core Auth Validator must not erase Q-100's distinction or falsely imply the current
ABV prototype already implements general application request evaluation.

![Proposed request flows](assets/auth-middleware-validator.svg)

## 2. Source reconciliation

| Approved foundation | Preserved architecture consequence |
|---|---|
| CONTRACT-006, docs/endpoint-authorization.md | One endpoint-owned gate; no middleware allow/prepared handoff. |
| Q-049/Q-050, docs/endpoint-policy-format.md | One fixed permission, selected inputs/sources, one server-owned declaration. No relationship/resolver DSL or duplicated type schema. |
| Q-093/Q-100, docs/auth-service-authority-gate.md | Administrative evaluation and authority-change boundary validation both required where applicable. Administrative scope is checked, not merely the verb. |
| Q-090–Q-112A, docs/grant-record-reference.md | Recipient-free grants, separate assignments, explicit revisions, actual parent/team lineage, separate controls and validity. |
| Q-039–Q-042, docs/application-registration.md | Registered permissions/scope meanings; optional support validation mandatory once enabled. No arbitrary application-domain interpretation by Auth. |
| Q-062–Q-067, docs/decision-results.md | Existing result variants/messages; no new returned-scope or prepared contract. |
| Q-071/Q-072, docs/collection-enforcement.md | Deny insufficiently covered requested collections; no automatic grant-derived filtering to salvage success. |
| Q-074/Q-110/Q-128/Q-129, docs/concurrent-enforcement.md | Preserve data boundary at use; Auth writes preserve checked authority through persistence; new checks cannot reuse withdrawn authority. |

123 organizes persistence and composition. The core validator is L1 functionality;
API binding is an adapter to core, not permission to move integrity into consumers.
The Auth/application fact-ownership layers remain a different classification.

## 3. Who does what

| Responsibility | Auth Middleware | Core Auth Validator | Endpoint / provider |
|---|---|---|---|
| Authentication | Uses established verifier/adapter; rejects invalid identity context | Requires trusted canonical identity; does not trust body identity | Authentication system proves actor/human association. |
| Policy | Selects registered method/route policy; checks declaration availability | Enforces fixed permission and required material contract | Application owns declaration and request value schema. |
| Input binding | Extracts only declared inputs from exact sources; no fallback | Consumes consistent material; missing necessary evidence is not allow | Application validates meanings and any needed facts. |
| Authority loading | Coordinates via the supported provider/client | Resolves complete applicable routes, adopted roles and restrictions | Auth owns authority records; no client-supplied grants treated as evidence. |
| Application facts | Calls endpoint-owned preparation when needed | Uses supplied trusted material under registered scope meanings | Constrained internal fact queries, no protected effects/disclosure before allow. |
| Decision | Calls core, validates result, stops deny/error/malformed response | Evaluates permission and boundaries; supplies approved result/reasons | Handler does not inspect grants or recombine permissions/scopes. |
| Execution | Passes exactly evaluated material to protected handler | Does not generate business SQL or reinterpret domain relationships | Handler constrains actual reads/writes; provider guards atomic Auth changes. |

Core transport independence is recommended: no HTTP router or application database
inside rule code. A shared authority provider supplies the complete required
evidence; a coordinator establishes consistency. Start with in-process invocation
where deployed, without requiring a new remote validation service. This does not
prescribe authority-fetch transports or introduce a cache policy. Remote deployment
can be evaluated separately; a network hop is not part of the canonical model.

## 4. Middleware location — wrap the endpoint gate

An early HTTP middleware can authenticate and establish context. It cannot always
finish authorization before application-specific material is available. Therefore
the proposed Auth Middleware is also the route-aware endpoint wrapper: preparation,
core invocation and constrained handler execution are one logical endpoint gate.

```text
Client request
  -> established identity and tenant/application context
  -> select endpoint policy, bind and validate declared inputs
  -> obtain sufficient authority and necessary application material
  -> core request evaluation
  -> allow: handler performs the SAME constrained operation
     deny/error: stop
```

Preparation is not a new prepared result/state. It is internal work before the
decision. Do not require eager metadata fetch when denial is already conclusive
or direct constrained execution enforces the relationship. An allow is not a
reusable token for a retry, another endpoint or broader material.

Implementation pseudocode, not approved SDK signatures:

```text
bindProtectedRoute(policy, requestSchema, endpointPreparation, handler):
    establish verified identity and outer context
    bind required inputs at declared sources; validate request values
    establish sufficient request material using endpoint preparation if needed
    result = core.EvaluateRequest(policy, context, material, authorityEvidence)
    validate result; on deny/error stop
    handler(context, sameMaterial) // enforced query/effect boundaries remain mandatory
```

Policy admission should reject missing/invalid policy on a protected route before
serving requests; runtime absence also fails closed. This is a recommended adapter
safeguard, not a new public endpoint-policy field. Do not add automatic public-route
exceptions, permission inference from HTTP verb or request-selected policies.

## 5. Canonical API writing example — certificate read

These JSON blocks reuse approved shapes. IDs and values are illustrative; all
definitions, assignments, team context and upstream support are assumed valid.
HRMS application context is established by the deployment/route binding, not
inferred from the permission string. The policy's tenant input must agree with
the verified requested tenant boundary.

### Grant content and assignment

```json
{
  "version": "1",
  "grant_id": "G-CERT-FIN",
  "revision": 1,
  "parent_grant_id": "G-CERT-PARENT",
  "permissions": ["hrms:employee:certificate::read"],
  "scope": {"dept": "FIN"}
}
```

```json
{
  "version": "1",
  "id": "A-CERT-EMPLOYEES",
  "grant_id": "G-CERT-FIN",
  "grant_revision": 1,
  "recipient": {"type": "group", "id": "Employees"},
  "status": "enabled"
}
```

Grant control is enabled; Vinay is an actual human member of Employees; the
required parent grant is held in the actual supporting team context. These
assumptions are required evidence, not facts established by the displayed JSON.

### Endpoint declaration — existing canonical partial policy shape

```json
{
  "version": "1",
  "method": "GET",
  "path": "/api/v1/{tenant}/{dept}/{cert}",
  "permission": "hrms:employee:certificate::read",
  "inputs": {
    "tenant": {"source": "path", "name": "tenant"},
    "dept": {"source": "path", "name": "dept"},
    "cert": {"source": "path", "name": "cert"}
  }
}
```

For `/api/v1/acme/FIN/C-17`, middleware supplies the verified Vinay identity,
established acme/HRMS context and selected request material. The core evaluates
the registered request boundary using current complete authority. It does not
assume that the path proves C-17's department. Endpoint code enforces the
relationship through a constrained operation, for example:

```sql
SELECT certificate_data
FROM certificates
WHERE tenant_id = :verified_tenant
  AND department_id = :evaluated_dept
  AND certificate_id = :requested_cert;
```

Here application isolation is supplied by this example's application-specific
database; a shared database must constrain application too. No unconstrained
second read is permitted. If C-17 is in ENG, the query returns no protected row.
Exact not-found/HTTP disclosure policy is a separate contract, not invented here.

Scope need not repeat every input. `{}` removes the grant's department restriction,
not required-input checks, tenant isolation or the operation's request binding.
`user: "$self"` refers to Vinay, not Employees, and the endpoint must enforce the
application-defined self relationship against the actual data.

### Decisions — keep the existing contract

```json
{"version":"1","decision":"allow","grant_ids":["G-CERT-FIN"]}
```

```json
{
  "version":"1",
  "decision":"deny",
  "error_code":"NO_AUTHORIZING_GRANT",
  "error_message":"You do not have access to this certificate.",
  "error_message_reason":"No grant authorizes this certificate read within Finance."
}
```

Evaluation failure uses the agreed error fields with no decision field; it is
not `decision: error`. Code spellings above are illustrative, not a finalized
catalogue. Both evaluator-provided messages reach the UI under the approved rule;
the producer must make them disclosure-appropriate, not include raw secrets/traces.
No returned scope/query plan is introduced, and the handler does not use grant_ids
to fetch and interpret grants. General HTTP wrappers/mappings remain to be finalized.

### PUT/body and collection consequences

An API can declare a selected body parameter using existing source/name bindings.
The application request schema validates the value; there is no automatic binding
of every body field. Current and proposed boundaries must remain distinguishable:
a submitted FIN department does not establish that the existing record is in FIN.
For the approved move case, the same endpoint permission must cover both current
and proposed boundaries. The write enforces the evaluated state at use.

An explicitly FIN-bounded collection can be allowed with FIN authority. A request
for all departments cannot be quietly rewritten to FIN. No middleware grant-to-SQL
filter generator, relationship DSL, business workflow or policy-specific handler
grant inspection is introduced. Bulk authorization must cover the complete request
according to the approved operation semantics before protected effects occur.

## 6. Auth's own APIs use the same machinery, with the extra core check

```text
Assignment-create request to Auth
  -> Auth Middleware binds declared permission, context and proposed assignment
  -> core EvaluateRequest: may Maya assign to this recipient?
  -> core ValidateAuthorityChange: is the complete proposed authority supported?
  -> preserve both checked states through one protected persistence operation
  -> commit assignment, or stop without partial writes
```

The two core names are proposed logical entry points, not new HTTP APIs.
Registration/role publication/control changes use their relevant checks; not every
operation distributes new business authority or requires an identical subset test.
An internal caller or CLI must also use the protected coordinator, never bypass
administration because middleware is absent. The current ABV facade's injected
administration capability remains necessary until an equivalent core contract is
established. No early middleware allow can authorize a later stale Auth write.

Example: Maya's administrative grant permits assignment to Team2. Her actual
supporting route supplies FIN read only. Assigning ENG write still fails ABV even
though the first check passes. Team-member administration remains distinct from
grant assignment; ownership never supplies unrelated source authority implicitly.

## 7. What exists and what this proposal adds

Current Go/SQLite/CLI provides authority-boundary validation for its documented
prototype subset, definition checks and protected mutations with injected lab
administration. It is not a complete general application request evaluator or
HTTP Auth Middleware. Renaming it cannot establish that missing implementation.

Keep 123 storage design independent of the API adapter. Middleware must not know
whether authority occupies 13 tables or one L1 store plus relationships. Any
targeted evidence provider must preserve completeness, administrative requirements
and freshness; a partial old Snapshot is not a safe new contract.

The pinned ABV base `/api/v1/{tenant}/abv/` applies to ABV service contracts, not
every protected HRMS/payroll/application API. Operation-specific route and explicit
application binding proposals in ABV-123-07 remain proposals, not prerequisites
for this illustrative existing certificate route.

## 8. Proposed next bounded design unit

After agreement on this responsibility split: one 25-minute unit specifies the
middleware-to-core in-process request/evidence contract and endpoint binding
example, reusing existing identity, policy and result shapes. At most one10-minute
correction, then regroup. Do not add runtime implementation while full schemas,
scope/material bindings or authority evidence requirements remain ambiguous.
Core capability naming/grouping approval does not finalize those contracts.

Required acceptance examples: simple GET; self via group; body-selected PUT;
current/proposed move; broader collection denied; missing material; expired or
orphaned lineage; unsupported proxy path; Auth assignment admin-pass/boundary-fail;
concurrent Auth mutation conflict. These are conformance cases, not invitations
to reopen approved rules. No migration, external Auth integration, new condition
engine, audit system or generic business-rule interpreter is included.

</details>
