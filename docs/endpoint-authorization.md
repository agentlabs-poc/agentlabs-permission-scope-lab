# Working handbook chapter: endpoint-owned authorization

Q-050-D now approves ENFORCEMENT-003: review that authorization boundaries and
request bindings actually constrain returned or changed data, not merely that
inputs appear in code. [The review examples](endpoint-policy-format.md) cover
logging, ineffective OR filters, unchecked output paths, and `{}`. This does not
guarantee detection of every breach or replace other checks/tests. Earlier
Q-050-D-proposed wording below is retained history after this approval.

## Current responsibility split — CONTRACT-012 / Q-050-C

The endpoint predeclares one required permission and selected inputs with their
sources. It must establish or enforce application relationships in implementation
to keep the actual operation within authorized boundaries. There is no adopted
relationship block or named-resolver contract. [The policy chapter](endpoint-policy-format.md)
preserves rationale, the grant/GET example, constrained lookup, tradeoff, and
original proposal not adopted.

Auth can establish authority within a supplied Finance boundary; this is not
permission to return an unchecked certificate merely because its route says
Finance. The endpoint must establish containment or enforce it directly through
the actual data operation. No protected output or effect may escape the authorized
boundary. This stays within CONTRACT-006's single endpoint-owned gate, not a
return to pre-handler business authorization or a prepared response.

The endpoint may use trusted facts already available or a constrained operation;
it need not eagerly fetch a relationship fact before every scope evaluation.
Q-047's distinction remains: request claims do not become factual proof because
their values match a grant. Tenant, grant validity/conditions, and human/delegation
limits remain mandatory. Review/tests must cover relevant mismatches; exact
update/move, consistency, collection, and result contracts remain open.

Earlier predeclared-relationship wording below is retained history, qualified by
CONTRACT-012. Q-050-D's subsequent input-usage review refinement remains proposed,
not an assertion that this review alone detects every security breach.

Q-050-B / CONTRACT-011 now approves the partial endpoint policy fields:
`version`, `method`, `path`, one `permission`, and named `inputs` with
`source`/`name`. [GET and PUT examples](endpoint-policy-format.md) explain field
rationale and path versus body inputs. The full policy remains open: relationship
bindings are Q-050-C. Earlier statements that no policy structure is discussed
are preserved history, not the current partial-approval status.

Q-050-A now settles version metadata under CONTRACT-010: a required top-level
`version` string, initially `"1"`, with rejection of missing/malformed/unsupported
versions. The rest of the endpoint policy JSON/YAML schema remains open under
Q-050. Earlier version-field/value-open notes below are retained history.

The endpoint policy's JSON/YAML contract has **not yet been discussed or
finalized**. The rules below are requirements, not an adopted schema. Under
CONTRACT-009, every published contract must include a version. See
[contract publication](contract-publication.md). Q-050 will address the endpoint
policy contract; no field names or version values are being invented here.

## One required permission per protected endpoint — CONTRACT-008 / Q-049, agreed

> The endpoint predeclares one required permission, inputs, sources, and how
> to establish any required relationship.

Every protected endpoint must declare exactly one required permission in v1.
Auth validates and enforces that declaration: zero or multiple required permissions
are invalid, and missing or invalid declarations must not permit execution.
The permission represents the endpoint's complete protected operation. The
declaration is server-owned; the caller does not select its required permission.
Exact validation timing, transport syntax, and framework integration remain open.

### Rationale and endpoint design obligation

One required permission removes endpoint-level AND/OR permission-expression
logic and gives each endpoint one explicit authorization requirement. It does
not remove scope or other mandatory checks. This keeps the endpoint contract
simple while leaving authority packaging through grants and roles flexible.

The tradeoff is deliberate operation design. If an endpoint bundles distinct
protected operations, a deliberately defined permission must cover the complete
operation, or the endpoint must be redesigned. Checking a narrow permission and
then performing unrelated privileged work is not compliant. This rule does not
settle multi-object, collection, bulk, or cross-application execution contracts.

| Endpoint operation | Its one required permission |
|---|---|
| Read a certificate | `hrms:employee:certificate::read` |
| Download a certificate | `hrms:employee:certificate::download` |
| Revoke a certificate | `hrms:employee:certificate::revoke` |

The download permission's defined meaning must cover the data disclosure inherent
in downloading. The download endpoint does not additionally require read, and
download does not automatically authorize the separate read endpoint. These
examples do not introduce permission hierarchy or implicit permission expansion.

### Grants remain flexible; checks remain complete

A grant may still contain multiple permissions sharing its scope and conditions.
Vinay may still have many direct and group-derived grants. For the endpoint's
one required permission, complete valid applicable grants remain alternative
authority routes under DECISION-001. Each route retains its own scope, validity,
conditions, and dependencies; no permission/scope field mixing is permitted.
Tenant isolation, human/delegation limits, application constraints, and binding
the decision to actual execution remain mandatory.

For example, Finance-scoped read cannot substitute for missing download authority.
A download grant confined to Engineering cannot borrow Finance reach from the
read grant. This follows from complete-grant evaluation, not from requiring
both permissions at the download endpoint.

### Earlier Q-049 proposal — not adopted

The assistant initially proposed AND across multiple endpoint-required permissions,
allowing different complete grants to satisfy different requirements. Its rationale
was to avoid depending on how permissions were packaged into grants. The example
required read AND download: direct G-21 supplied Finance read, and group G-22
supplied Finance download; making G-22 Engineering-only would fail the Finance
download requirement. Grant validity and membership were assumed in that example.

The user instead requested exactly one permission per endpoint, mandated by Auth,
for design simplicity, and approved that revision. The earlier multi-permission
proposal was never approved and is not a v1 option. Multiple permissions in a
grant remain supported. Older plural endpoint wording and open-combination notes
below are retained history, qualified by CONTRACT-008 rather than deleted.

## Endpoint declaration and request resolution — Q-047 / Q-047-A, agreed

Historical plural wording, narrowed to exactly one required permission by Q-049:

> The endpoint predeclares permissions, inputs, sources, and how to establish
> any required relationship.

This is the approved clarification of CONTRACT-007, not a new declaration
format. The single declaration is server-owned and fixed for the endpoint;
it is not rebuilt from each caller's grants. Application-defined scope meanings
remain registered contracts, not meanings invented by the endpoint.

### Request, resolved request, and decision

- **Authorization request:** the declared operation, verified identity/tenant
  context, and selected request inputs.
- **Resolved request:** an evaluation view of the same request with sufficient
  trusted material to assess the relevant authorization boundaries. Request
  claims remain distinguishable from established facts.
- **Decision:** whether applicable authority permits the requested operation
  using that material. Resolved does not mean allowed.

These are conceptual distinctions inside the one endpoint-owned gate, not
separate persisted entities, a prepared handoff, or newly adopted JSON fields.
Resolution does not manufacture authority or require unnecessary lookups before
an obvious denial. Exact schemas and multi-permission combination rules remain
open; the approved plural wording does not silently choose AND or OR for them.

Update: CONTRACT-008 now excludes multi-permission endpoint declarations in v1.
The preceding open-combination statement is historical; exact schemas remain open.

### Declared inputs versus material needed for a scope check

For `GET /api/v1/{tenant}/{dept}/{cert}`, the endpoint predeclares these inputs
and their sources, the required permission, and how to establish relationships
when needed. The route tenant must be bound to trusted tenant context. Declared
inputs remain available even when one grant's scope does not need all of them.

Assuming the application has registered the illustrated boundary keys and
their meanings for certificate-read:

| Grant scope | Scope evaluation needed |
|---|---|
| `{}` | Enclosing tenant only; this scope imposes no additional department or certificate restriction. |
| `{"dept":"FIN"}` | Establish that the operation on the requested certificate stays within Finance. The path department alone need not establish this relationship. |
| `{"dept":"FIN","cert":"C-17"}` | Establish both the Finance boundary and the C-17 restriction. |

The middle case uses the certificate identifier to establish its department
without making `cert` a selected boundary key. A scope need not repeat every
declared input. Conversely, having a path input named `dept` does not establish
that the requested certificate actually belongs to that department.

With an applicable `{}` grant, do not fetch department facts solely to enforce
a department restriction that this scope does not contain. Department and
certificate still identify the requested operation; they have not disappeared
from the declaration. The operation must stay within the trusted tenant and
remain bound to the requested certificate. If the application route contract
requires that certificate to belong to the supplied department, that remains
application validation rather than a restriction invented for the empty scope.

Grant validity, conditions, and human/delegation limits continue to apply.
An empty scope in one grant does not erase other mandatory constraints. Required
material for those constraints may differ from material needed for scope alone.

### Rationale, counterexample, and remaining work

Separate request identification from authority restriction. Treating every
declared input as a mandatory selected boundary would contradict `{}`'s agreed
tenant-wide meaning and cause unnecessary fact gathering. Treating unused inputs
as absent could instead lose the binding to the actual requested operation.

For example, `FIN/C-17` does not prove Finance membership if established facts
place C-17 in Engineering. A Finance-only grant cannot authorize that operation
merely by matching the path string. A tenant-wide grant does not impose that
Finance restriction, but still cannot bypass tenant isolation or the endpoint's
application contract. No universal extra lookup is required when trusted context
or an appropriately constrained operation already establishes what is needed.

The approved principle is selective, sufficient trusted material under a fixed
declaration. Concrete declaration syntax, relationship-binding representation,
conflict handling, collection enforcement, and freshness remain open. Earlier
permission/material wording below remains compatible shorthand, not a second
declaration or a requirement to resolve every possible relationship eagerly.

TERM-005 / Q-043 uses request material and boundary evaluation without an
additional canonical entity. [Detailed rationale](authorization-vocabulary.md)
retains the requirement to bind evidence to actual execution.

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
fetching unnecessary resource metadata. Shared role expansion and authority loading
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
| Application, when required | Trusted resource attributes and relationships needed to evaluate the selected boundaries. |

Request inputs are material, not automatically authoritative claims about an
existing resource. A submitted department can identify the requested boundary
or a proposed value; it does not by itself prove an existing certificate belongs
there. For creation/update, selected body values can describe the proposed
resource or change, subject to validation and any required current-state facts.
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
Required application facts establish that C-17 is inside the relevant boundary, rather than
trusting the request's relationship claims. Permission and scope are evaluated
as parts of each complete grant, never recombined across unrelated grants.

SCOPE-007's canonical key-value syntax illustrates the agreed AND rule:

```json
{ "dept": "dept-1", "user": "$self" }
```

The certificate must be both within dept-1 and
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
