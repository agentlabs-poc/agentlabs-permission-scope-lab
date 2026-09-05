# Working handbook chapter: endpoint-owned authorization

CONTRACT-006 is the current agreed model, approved in Q-033. The user authorized
resuming documentation and explicitly requested preserving earlier designs as
deprecated, not deleting or silently rewriting them. The earlier
[two-mode authorization flow](authorization-flow.md) remains as historical prose.
No application implementation has been changed to match this chapter.

The [canonical logical block diagram](system-overview.md) maps this agreed model.
CONTRACT-007 requires each endpoint's permission, required material, and material
sources to be explicit in the single declaration; detailed syntax remains open.

## One authorization gate — CONTRACT-006

The endpoint owns gathering sufficient authorization material, invoking a shared
evaluator, and enforcing the result before the protected operation. Endpoint
ownership does not mean each handler invents its own grant or scope semantics.

```text
Authenticate and establish trusted context
                    ↓
Endpoint gathers sufficient authorization material
                    ↓
Shared authorization evaluator decides
                    ↓
Endpoint enforces the result and performs the operation
```

Middleware may authenticate, reject invalid authentication/request inputs,
establish context, and load authority information. It does not make a first
business-authorization decision that the endpoint sometimes completes. There
is one endpoint-owned authorization gate and no cross-layer prepared outcome.
Internal incomplete work is not authority; an error or missing required material
cannot permit protected execution.

This is one logical decision location, not a requirement for one database call,
one literal function invocation, or eagerly loading all possible facts. If no
applicable grant supplies the required permission, the gate may deny without
fetching unnecessary target metadata. Shared role expansion and authority loading
may still be reused without being independent authorization decisions.

## Material selected by the endpoint — INPUT-001

The endpoint contributes the method-to-action mapping and explicitly identified
path/body parameters as request inputs. Its server-owned method-plus-route
declaration identifies the action and associated permission; HTTP method alone
is not a universal permission rule. The caller cannot supply the required
permission or expand the set of authorization-relevant fields arbitrarily.

| Source | Material |
|---|---|
| Verified authentication context | Human/actor identity and trusted tenant context. |
| Server-owned endpoint declaration | Mapping from this method and route to action/required permission, and identified input bindings. |
| Request | Selected path parameters and explicitly identified body parameters. |
| Auth | Applicable direct/group grants, role definitions, memberships, and delegation dependencies. |
| Application, when required | Trusted target attributes and relationships needed to evaluate the selected boundaries. |

Request inputs are material, not automatically authoritative claims about an
existing resource. A submitted department can identify the requested boundary
or a proposed value; it does not by itself prove an existing certificate belongs
there. For creation/update, selected body values can describe the proposed
target/change, subject to validation and any required current-state facts.
The exact request representation, field mapping, and conflict handling are open.

The endpoint does not need an application lookup on every request. It obtains
the additional material necessary for the declared operation and applicable
authority. Trusted existing context or an appropriately constrained operation
may already establish the needed relationship. No query language is adopted.

## Boundary example — SCOPE-006, SCOPE-008

For GET /api/v1/tenant/dept-1/C-17, assume the tenant is established through
trusted context. The declaration maps this operation to certificate-read and
identifies the department and certificate path inputs. If a tenant identifier
is carried in the route, it must also be bound to trusted tenant context.

The endpoint obtains applicable authority. A grant's permission must cover
certificate-read; its scope selects the boundary in which C-17 must fall.
Required application facts establish actual target membership, rather than
trusting the request's relationship claims. Permission and scope are evaluated
as parts of each complete grant, never recombined across unrelated grants.

SCOPE-007's canonical key-value syntax illustrates the agreed AND rule:

```json
{ "dept": "dept-1", "user": "$self" }
```

The target must be both within dept-1 and
within the authorizing human's self boundary. These application key meanings
must be defined; the format and $self token are canonical under Q-034. This is
not a new request or resolved-grant schema. An explicit empty scope {} is
tenant-wide; omitted or null scope is invalid, never defaulted to {}.

Alternative authority is supplied through separate complete applicable grants,
each retaining its own scope, conditions, and dependencies (DECISION-001 and
SCOPE-008). A group-derived grant retains its membership dependency; service and
agent authority remains bounded by its human and delegation constraints.

## Safe fact gathering and enforcement — ENFORCEMENT-002

Before authority is established, only the necessary tenant/request-constrained
internal fact work may occur. Do not perform the protected mutation, disclose
protected output, or trigger business side effects while gathering material.
Authority must remain bound to the actual operation and the facts used at the
time of enforcement; consistency and freshness mechanisms remain open.

For a collection, do not infer that every row must be read before authorization.
An endpoint may enforce authorized reach through a constrained data operation
where supported. The exact collection, query, aggregate, create, move, and bulk
contracts remain open and do not revive the two-location decision model.

## Deprecation map

| Earlier material, preserved | Current interpretation |
|---|---|
| ARCH-002 and CONTRACT-001's split middleware/completion workflow | Deprecated; endpoint-owned material gathering and evaluation are current. |
| CONTRACT-002 and CONTRACT-003's two modes and mode validation | Deprecated; there is one endpoint-owned authorization gate. |
| CONTRACT-004 | One authoritative auth-first method/route declaration remains; its requirement to choose a resolution mode is deprecated. |
| CONTRACT-005's choice between completion modes | Deprecated as a mode-selection rule; deciding authority and enforcing it remain distinct responsibilities. |
| ENFORCEMENT-001's prepared/middleware-allow wording | Deprecated; ENFORCEMENT-002 retains the safety invariant without prepared. |
| SCOPE-005's two-mode selector/query walkthrough | Deprecated proposal, never agreed. Query grammar remains open. |

The earlier RESOLUTION-002 and ARCH-003 alternatives stay superseded; endpoint
ownership does not revive a universal prepared result. Resolution itself remains
essential: removing prepared removes a handoff state, not authorization work.

## Still to settle

- Defined scope-key meanings, governance, and enforcement of canonical validation.
- Action/permission mapping, identified path/body input declarations, and conflicts.
- Concrete authorization request, evaluator result, and resolved-grant forms.
- Scope containment for grant administration; ADMIN-004/005 remain open.
- Collection/create/update/move/bulk semantics and non-HTTP operation adapters.
- Audit, consistency, revocation freshness, and safeguards ensuring no handler
  bypasses the shared gate.
