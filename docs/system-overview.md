# Authorization system — canonical logical overview

This block diagram maps the agreed logical model: CHARTER-002, CONTRACT-006,
INPUT-001, ENFORCEMENT-002, and SCOPE-006/007/008. It does not introduce a new
authorization stage, deployment topology, network call count, or wire schema.
The detailed source-binding declaration refinement is proposed as CONTRACT-007.

```text
                            Endpoint declaration
                       Required permission + material sources
                                      |
                                      v
Request --> Authenticate --> Endpoint-owned authorization gate
            and establish             |
            trusted context           v
Auth authority -----------> Gather sufficient material <--- Application facts
                                      |
                                      v
                              Shared evaluator
                          Complete grant evaluation
                             /              \
                          allow          deny / failure
                            |                  |
                            v                  v
                     Enforce decision         Stop
                            |
                            v
                   Protected operation / response
```

The application-facts arrow feeds material gathering; it is not a direct route
to protected execution. Identity/tenant context and selected request inputs
also enter the gate. The declaration is server-owned configuration, not a
caller-controlled payload. Failures in authentication or required material
gathering likewise cannot reach protected execution, even when they occur before
the evaluator in the main-path diagram.

## Responsibilities

| Component | Responsibility |
|---|---|
| Endpoint declaration | Identify the required permission and relevant material/source bindings for this operation. Detailed binding schema is still open. |
| Authentication/context | Establish verified actor/human identity and trusted tenant context; do not issue a partial business-authorization result. |
| Endpoint-owned gate | Gather sufficient material and invoke the shared evaluator; do not perform protected output or mutation first. |
| Auth authority source | Supply applicable grants, roles, memberships, and delegation dependencies under the still-to-be-finalized freshness contract. |
| Application fact source | Establish target attributes and domain relationships needed by the operation and scope, through bounded internal access. |
| Shared evaluator | Check recipient applicability, required permission, scope boundary membership, validity/conditions, and mandatory outer/dependency limits without mixing unrelated grant fields. |
| Enforcement/operation | Bind execution to the decision and its restrictions; keep check and actual use consistent. |

Auth authority may be loaded through shared infrastructure or valid caches; the
diagram does not require every handler to implement its own Auth integration or
perform a new network call. Scope has no independent network service implied by
this diagram: it owns boundary semantics, and the shared evaluation uses those
definitions. Definition storage/registration remains open.

## Declaration refinement — CONTRACT-007, proposed

Every protected endpoint has one authoritative declaration covering:

1. Its action/required permission, mapped from the server-owned method and route.
2. The authorization material that operation needs or may need for its supported
   scopes, with an explicit way to obtain that material.

Inputs are not just a bag of similarly named fields. The source gives each item
its provenance and role: requested identifier, proposed value, verified context,
Auth-owned authority, or trusted application fact. These are explanatory
categories, not newly adopted schema fields.

For GET /api/v1/tenant/{dept}/{cert}, an illustrative declaration's meaning is:

| Material | Where obtained |
|---|---|
| Required certificate-read permission | The method/route's server-owned action mapping. |
| Requested department | Path parameter dept. |
| Requested certificate | Path parameter cert. |
| Human/actor and tenant | Verified context, not arbitrary body fields. |
| Applicable authority | Shared Auth authority loading. |
| Actual certificate department, if needed | The application's bounded certificate fact lookup. |

For an operation with a body, only explicitly identified body inputs contribute
request material. Their meaning must distinguish proposed changes from facts
about existing data. A caller-submitted department does not prove the stored
certificate belongs to that department. Shared identity/authority infrastructure
need not be duplicated in every endpoint; the exact declaration/reference
mechanism is still open.

"Required material" does not mean eagerly load the union of every conceivable
fact. The endpoint contract must provide a defined source for what its supported
authorization needs, and the gate obtains enough to decide. It can deny for lack
of permission before fetching unnecessary target metadata. Material collection
does not become an independent middleware authorization phase or a prepared
outcome.

## What this diagram does not finalize

- Declaration JSON, decorators, source-binding vocabulary, or resolver APIs.
- The scope-key governance branch or an automatic path-key-to-scope-key mapping.
- Freshness, caching, audit event schemas, or transaction mechanisms.
- Exact list/create/update/move/bulk request and result contracts.
- Any change to application implementation or the deprecated two-mode design.

The [discussion tree](discussion-tree.md) remains the handbook-progress mind
map. This diagram answers a different question: how the agreed system's logical
components cooperate. This source-binding sidebar returns to scope-key
definitions, then administrative bounds, then request/resolved-data contracts.
