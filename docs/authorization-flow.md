# Working handbook chapter: authorization flow and endpoint contracts

> Vocabulary below is historical as well as the flow design. TERM-005 / Q-043
> supplies the current [request-material explanation](authorization-vocabulary.md).

> DEPRECATED DESIGN — preserved as requested. Q-033 approved CONTRACT-006's
> [endpoint-owned authorization gate](endpoint-authorization.md). The two-mode
> flow and prepared handoff below are historical, not current requirements.
> Auth-first declarations, trusted facts, and fail-closed enforcement remain;
> the current chapter states them without the deprecated mode mechanism.

This chapter preserves the endpoint discussion, its rationale, examples, and
open details. Agreed decisions are identified by reference ID. Related proposed
implementation contracts are explicitly identified; no API schema or current
production implementation is claimed.

## Shared rules, application facts — CHARTER-001, CHARTER-002

All AgentLabs applications follow the same core authorization semantics. Auth
holds shared identity, tenant membership, roles, and authorization assignments.
Applications own their business facts, such as which employee corresponds to a
user and which department contains a certificate.

The application's authorization layer evaluates access using both sources.
The application enforces the result in protected handlers, jobs, and data access.
This logical separation does not mandate a network round-trip for every fact.
The working integration direction uses a shared evaluator with application
adapters (ARCH-001); its exact packaging and interfaces remain open.

The middleware in the agreed endpoint model does not access the application
database. An endpoint needing domain facts obtains them through its application
service or repository and supplies them for authorization completion.

## Auth-first design — PRINCIPLE-002, CONTRACT-004

Design the operation, route, authorization declaration, handler, and data access
together. Each HTTP method plus route template has one authoritative contract
and one declared resolution mode. Middleware and handler follow that contract.

For example, if the route identifies an opaque certificate whose ownership must
be discovered, its contract accounts for that application lookup. The handler
must be written to complete the declared checks before protected output.

A different authorization workflow may warrant a separate endpoint or a
deliberate contract revision. Different recipients or multiple checks within
one workflow do not require a separate route per role, user, or grant. Separate
routes may still map to the same permission when the operation is the same.

## Two declared resolution modes — CONTRACT-002

| Mode | Middleware outcome | Handler obligation |
|---|---|---|
| Middleware-complete | Final allow or deny | On allow, enforce the decision's action, target, and restrictions while executing. No prepared state is part of this endpoint's contract. |
| Endpoint-completion | Prepared or deny | On prepared, obtain the required application facts, complete authorization, and enforce the completed result. Middleware never returns allow for this mode. |

The endpoint's declared mode determines where authorization finishes; the
request determines whether authorization succeeds. The caller does not choose
the mode, and middleware does not switch modes because one user's grants appear
easier to evaluate than another's.

The reason for declaring the mode is predictability: both parties know what
middleware's success means. An endpoint-completion handler always expects
unfinished authorization. A middleware-complete handler receives a completed
decision, while still being responsible for enforcement.

The earlier universal allow/deny/prepared model was superseded by this declared
contract. A universal rule that every handler must complete authorization was
also superseded; only endpoint-completion mode requires that second resolution.

PRINCIPLE-001 still applies to failures. Unavailable required authority,
incompatible inputs, or failed resolution cannot permit protected execution.
These failures remain diagnostically different from a completed policy denial.

## Certificate example: enough identifiers is not always enough facts

Consider this illustrative endpoint-completion route:

```http
GET /api/v1/T-1/FIN/certificates/C-17
```

The requested identifiers are available, but the application may still need to
establish their relationships. FACT-001 and CONTRACT-001 record the proposed
fact-provenance/declaration details behind this distinction:

| Input or fact | Meaning |
|---|---|
| Verified identity context | Identifies the human or dependent actor and applicable tenant context. |
| Operation mapping | Establishes the required permission from the server-owned endpoint declaration. |
| T-1, FIN, C-17 in the route | Select the requested tenant boundary, department, and certificate. |
| Stored certificate attributes | Establish C-17's actual tenant, department, and any required ownership relationship. |

The string FIN in the route does not prove C-17 belongs to Finance. The endpoint
may establish that fact by a lookup or a constrained query, depending on the
declared enforcement contract. A self-scoped grant needs the relevant human
ownership relationship; a department-wide grant does not automatically require
the certificate to be personally owned by that human.

For an endpoint-completion contract:

1. Middleware binds the authenticated context and operation and obtains applicable
   authority information. Complete positive grants remain alternative routes;
   one failed candidate alone does not prove a global denial (DECISION-001).
2. If it can establish a conclusive denial under the applicable rules, it stops.
3. Otherwise it forwards prepared authorization context, preserving restrictions
   and unresolved requirements. It does not return final allow in this mode.
4. The application obtains the required trusted resource/relationship facts.
5. The evaluator completes the outstanding authorization checks using those facts.
6. The handler returns protected data or performs the protected action only under
   the resulting authorization and enforced restrictions.

The exact fields of the prepared context, expanded/resolved request, and
expanded/resolved grants remain open. The flow does not establish those schemas.

## Prepared is not authorized — ENFORCEMENT-001

Preparation permits continuation of the internal work needed to establish
required authorization facts. It does not authorize protected output or a
business mutation. Fact retrieval must remain constrained by the established
tenant and request boundaries. If a required completion check is omitted or
cannot be completed, the protected operation must not succeed.

A middleware-complete allow also has limits. It authorizes the declared operation
within its target and restrictions, not arbitrary data the handler might access.
The application must bind actual use to that decision. Concurrent changes
between checking and use, revocation freshness, and enforcement APIs remain
topics for the lifecycle branch.

## Resolution versus enforcement — CONTRACT-005

A database call alone does not determine the endpoint's mode. The distinction
is whether application facts are still needed to determine authorized access,
or whether the handler is applying complete restrictions already determined.

For example, where middleware can fully establish permitted access and supply
a complete tenant restriction for a supported operation, the handler's query
within that tenant is enforcement. It does not automatically imply an unfinished
authorization decision.

Conversely, a self selector with an unknown human-to-employee relationship may
need completion. Resource containment, reporting relationships, approval
separation, amount-based limits, moves, and bulk requests can also need facts.
The seven [candidate cases](endpoint-completion-cases.md) explain their required
facts and assumptions. Those illustrative policies are not automatically adopted
as product requirements.

A business validation that applies to everyone, such as rejecting malformed
input, is not automatically an authorization condition. Whether the actor's
authority permits an operation is the question this contract addresses.

## Why resolution cannot create authority — RESOLUTION-001

More information can expose or narrow existing authority; it cannot invent it.
Expanding a role to several permission entries increases representational detail
without granting new capabilities. Resolving ownership can establish whether a
target is covered without widening the grant's reach.

For dependent service/agent access (AUTHORITY-002), the result remains bounded
by both the human's applicable authority and delegation limits. Loss of an
unrelated human permission does not invalidate still-covered access.

## Details still open

The [scope chapter's endpoint walkthrough](scope-model.md) (SCOPE-005 / Q-029)
illustrates a department-qualified certificate request, an existing-form grant,
and a constrained lookup. It distinguishes query enforcement of a fully known
restriction from endpoint completion; query grammar and interfaces remain open.

- CONTRACT-001: exact operation declaration and required-fact interfaces.
- CONTRACT-003: mode validation across every supported scope/delegation form;
  handling incompatible configuration without silently changing modes.
- FACT-001: field provenance, request selectors, established facts, and binding.
- Prepared context, decision, resolved request/grants, and revision schemas.
- Query predicates, collection semantics, partial bulk results, and mutations.
- Cached authority, synchronization, revocation and role-change freshness.
- Consistency between authorization checks and resource use; audit evidence.

The estimate that most endpoints can be middleware-complete remains a hypothesis.
We must inspect actual operation contracts and supported scopes to measure it.
