# 7. Application integration

[Contents](../README.md) · [Previous](06-auth-service.md) · [Next](08-lifecycle-and-scenarios.md)

**Framework-specific integration guidance.** Authorization is performed within
one endpoint-owned gate. Middleware may authenticate, establish trusted context
and load usable authority, but it does not hand the handler a provisional
allow/prepared result. There is no prepared state in the selected model.

## Follow the request

**Placement clarification — AUTH-ARCH-02:** Auth Middleware/evaluator is the
API-service-side component. Core Auth Validator / ABV remains inside Auth Service
for its definition and authority-changing operations. The API-side component
consumes Auth authority; it does not embed ABV's mutation engine or database.
This naming clarification preserves the endpoint-owned gate and application
enforcement described below; it does not introduce a prepared handoff.

![Client request through authentication, endpoint input binding, embedded Auth evaluation, constrained data access and response](../../docs/assets/authorization-system.svg)

The endpoint declares its required permission and selected inputs. Its handler
binds and validates those inputs, invokes the embedded Auth Agent, and enforces
the resulting boundary on the actual operation. The agent combines canonical
authority from Auth with the application's authorization material and integration.
It is not necessarily a separate process or a middleware-only component.

Auth supplies current applicable authority and its required support. The
application supplies its registered meanings, request validation, facts where
needed and constrained execution. Neither layer replaces the other: an
application cannot manufacture grants, and Auth cannot prove certificate
ownership merely from a path parameter.

## The gate contract: four things supplied, one received

An application integrates authorization through **one gate in front of one
endpoint**. The gate owns the order of operations.

| | What it is | What it must not do |
|---|---|---|
| **Policy** | The endpoint's static declaration — version, method, path, one permission, selected inputs, trusted correlations. | Change after mounting. A mounted copy decides; the caller's struct is no longer consulted. |
| **Identity source** | Establishes the trusted request context from the request. | Read the business body. It is handed a request whose body is empty. |
| **Binder** | Validates application schema, selects the material authorization will use, and returns the effect as a closure over those same validated values. | Publish output, or perform the effect. It runs before any decision exists. |
| **Failure handler** | Renders a denial or an evaluation failure. | Turn one into the other, or proceed. |

And it receives the **Result**: the decision and, on an allow, the contributing
chain.

**The order is the authorization**, which is why it is specified rather than left to
each deployment:

1. **Method and route** must match the policy, or nothing else runs.
2. **Identity** is established, from a request with no business body — so a
   credential can never be taken from the payload it is meant to protect.
3. **Trusted correlations** are verified against the resolved inputs.
4. **The binder** produces material and the effect closure.
5. **Authority is loaded and evaluated.**
6. **Only on an allow** does the effect run, and it is handed the Result.

The binder runs before evaluation deliberately: the effect must close over the
*same* validated values the decision was made about. A binder running afterwards
could bind to something else, which is the check-to-use gap enforcement exists to
prevent. And **no HTTP request reaches the effect** — it receives a context, a writer
and the Result, so it cannot reparse a path or a body after the decision.

An application linking this gate links **no authority records, no schema and no
authority store**. It needs the gate and a source of answers, and nothing else.

Deliberately not specified: an SDK, which handbook completion does not require; the
HTTP status mapping for a denial versus a failure, which belongs to the application;
and any transport for administrative operations, which stays open.

A deployment with its own middleware chain must place this gate where the order
above still holds, and the contract cannot check that it did. That cost is accepted,
because a deployment that reorders those steps has a gate that looks right and is
not, and nothing in a pure wire contract would catch it.

## Declare authorization before writing the handler

Every protected method/route declares exactly one permission in this framework.
That permission covers the endpoint's complete protected operation. The caller
does not choose it. A grant may contain several permissions and a user may have
many grants; neither changes the endpoint's single-permission contract.

When an endpoint combines several responsibilities, its declared operation must
genuinely cover the complete effect, or the endpoint must be redesigned. Checking
read and then performing an unrelated privileged write is not authorized because
both happen inside one function.

The approved partial policy contains `version`, `method`, `path`, `permission`,
named `inputs` with explicit `source` and `name` bindings, and `trusted`
correlations. It does not contain a relationship block or duplicate application
value-validation rules.

**`trusted` closes an obligation this chapter previously stated and left
unspecified.** A route's tenant claim must be bound to trusted context, and a field
name alone proves nothing. `"trusted": {"tenant": "tenant"}` reads as *the input this
policy calls `tenant` must equal the trusted tenant*, and a policy declaring no
tenant correlation **cannot be mounted** — so forgetting becomes impossible rather
than silent. A path carrying `{tenant}` or `{application}` that no correlation names
is refused for the same reason. A correlation may name any declared source; the check
is on the resolved value. It is not the relationship language the model declined:
both values are already in the gate's hand, nothing is looked up, and no record
relationship is asserted.

The alternative was for the gate to refuse any route segment it could not account
for, which needs a heuristic for which segments are tenant-shaped — one that catches
`tenant_id` and misses `org`. A declaration guesses nothing. The cost is that every
existing policy document must gain the field or stop mounting, which is the intended
consequence: the failure this closes was a policy that omitted the binding and was
accepted anyway.

**A policy is validated structurally at mount time**, not at request time. A policy
that cannot be mounted cannot guard an endpoint, so the endpoint does not serve
rather than serving unguarded:

| | |
|---|---|
| Version | Exactly the supported contract version. No default is guessed. |
| Method | A valid uppercase HTTP token, and the request's method must equal it. |
| Path | Absolute, no query or fragment, every placeholder well formed and unique. |
| Permission | Exactly one, canonical, no wildcard and no list. |
| Inputs | Declared; each local name and source name well formed. |
| Path inputs | Must name a placeholder the path actually declares. |
| Body inputs | **Top level only.** A selector containing `.`, `[`, `]` or `/` is refused. |
| Sources | Exactly two are supported: `path` and `body`. |
| Trusted | Required for the tenant. |
| Unknown fields | A field this version does not define is **refused, not ignored**. |

Every rule there is checkable before a request arrives, and a policy that is wrong
should fail where an operator is looking rather than on a caller's request. The cost
is that adding a field to the policy contract is a breaking change by construction —
a consumer that silently ignores what it does not understand is a consumer that
enforces something other than what was declared.

**Nested body selection is refused, not deferred.** `employee.department_id` is not a
selector. Once `a.b` is admitted, arrays, filters and absence semantics follow, and
the policy becomes a place where application structure is described. An application
needing a nested value reads it in the binder, validates it, and supplies it as
material under its own local name.

**A declared input the request does not supply is a refusal of the request**, not an
absent value passed to the binder. No implicit default, no empty string, and no
fallback to another source.

**What a policy still does not declare** is the scope boundary its endpoint operates
at. The boundary reaching evaluation is whatever the binder supplies at request time,
so an endpoint may claim a narrow boundary and read a wide one, and nothing can check
it. That question is framed and not approved; until it is settled, an all-values
selection denies rather than deriving an authorized subset.

### GET: path-supplied material

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
  },
  "trusted": {"tenant": "tenant"}
}
```

This is an approved-structure example, not a complete deployable policy schema.
For `/api/v1/acme/FIN/C17`, extraction supplies the requested tenant, department
and certificate. The tenant claim must agree with trusted context. `FIN` is not
proof that C17 belongs to Finance. A local input name also does not automatically
establish a registered scope-key relationship.

Every declared input is required at its exact declared source. Do not silently
fall back from body to path or from an absent field to a permissive default.
The application's request contract validates types, nullability, format and
domain values; duplicating those rules as new policy fields would create a
second contract that could disagree.

### PUT: selected body material

```json
{
  "version": "1",
  "method": "PUT",
  "path": "/api/v1/{tenant}/certificates/{cert}",
  "permission": "hrms:employee:certificate::write",
  "inputs": {
    "tenant": {"source": "path", "name": "tenant"},
    "cert": {"source": "path", "name": "cert"},
    "proposed_dept": {"source": "body", "name": "department_id"}
  },
  "trusted": {"tenant": "tenant"}
}
```

Here `proposed_dept` is a local input name, not a new canonical scope key. Its
value comes from the body field `department_id`. A value of FIN states an
intention; it does not prove the record currently belongs to Finance. A body
field naming an employee likewise does not establish the caller's identity.

The handler may process other business fields, but selecting only one field as
authorization material does not authorize arbitrary extra effects. The complete
operation must remain within the permission and boundary actually evaluated.
Nested input selectors and further source kinds are **refused rather than pending** —
exactly two sources are supported, and a body selector containing `.`, `[`, `]` or `/`
is refused at mount. What remains open about a policy is the boundary it operates at,
below; the structural rules are settled.

## Resolve authority without making the endpoint inspect grants

Authority loading considers the human's direct assignments and valid group
memberships, group assignments, adopted grant and role content, controls,
validity, parent support and applicable delegation. Resolved grants remain
dependent views of these records, not permanent copied authority.

### The question and the answer

The application asks **what one human holds in one area.** It never asks whether to
allow.

```json
{
  "version": "1",
  "identity": {
    "version": "1",
    "actor": {"type": "service_account", "id": "agent_hrms"},
    "human_id": "U-17"
  },
  "options": {}
}
```

`actor` is the asking application's own credential; `human_id` is the person it asks
about. They are different parties, and that separation is the point — an application
asks as itself about many humans and is none of them. Nothing in the question names
an endpoint, a method, a resource or a verdict.

`options` carries two optional narrowings and nothing else. `permissions` narrows the
answer, and is a narrowing rather than an assertion: omitted means everything the
human holds, which is the answer worth caching against them. `omit_source` drops the
explanation, and a consumer that must produce allow evidence cannot use it, because
the contributing chain lives in the explanation.

The answer echoes the three boundaries and lists every grant the human holds there,
after the chain has been walked:

```json
{
  "version": "1",
  "tenant_id": "acme",
  "application_id": "hrms",
  "human_id": "U-17",
  "resolved_grants": [
    {
      "version": "1",
      "grant_id": "G2",
      "revision": 1,
      "parent_grant_id": "G1",
      "permissions": ["hrms:employee:certificate::read"],
      "scope": {"dept": "FIN", "cert": "C17"},
      "validity": {"not_before": null, "expires_at": null},
      "source": {
        "assignment_id": "A2",
        "team_id": "Team2",
        "via": "membership",
        "lineage": [
          {"grant_id": "G0", "revision": 1, "assignment_id": "A0", "team_id": "Team0", "root": true},
          {"grant_id": "G1", "revision": 2, "assignment_id": "A1", "team_id": "Team1"},
          {"grant_id": "G2", "revision": 1, "assignment_id": "A2", "team_id": "Team2"}
        ]
      }
    }
  ]
}
```

`lineage` is root-first and complete: G0 is the trusted root, held by Team0 through
A0, which is why it and not G1 carries `"root": true`. That completeness is what lets
a consumer produce the allow evidence below — the same three grants, in the same
order.

`scope` arrives **already folded down the chain** — Finance from the parent, C17
locally — so a consumer never folds one itself. `permissions` arrives already expanded
from any adopted role. `validity` is the narrowest window in the chain. `source`
explains why the human holds it, root-first; it is explanation, not authorization, and
none of it may widen what the grant already permits.

Empty `resolved_grants` is a **completed answer**, never a failure.

### What a consumer must do with it

- **Reject an unsupported version** rather than guessing a default.
- **Corroborate the three boundaries** against the question it asked. An answer about
  a different tenant, application or human is unusable *in whole*, not in part.
- **Reject unknown fields, duplicate keys and trailing content.** Drift between the
  two sides is how one of them silently starts meaning something else.
- **Bound the answer's size.** A response larger than the consumer accepts is a
  failure to establish authority, not a denial.
- **Do not follow redirects.** A redirect is not an authority service: following one
  hands the credential to whoever set the header and then believes the reply.
- **Every failure here is an evaluation error**, never a denial.

Not adopted: the route this is served at, a batch form, and freshness fields.
`authority_epoch` and `resolved_at` are unbuilt — nothing caches, so nothing can be
stale, and a cache and an epoch are adopted together or not at all.

Sufficient valid preloaded authority may satisfy this step. The logical flow
does not mandate a fresh remote Auth call for every request; it mandates the
required correctness and freshness of the evidence actually used.

For the previous chapters' example, Nutan's Team2 membership makes A2/G2
applicable. Its required parent-team support at Team1 supplies G1 revision 2.
The complete route offers certificate read within FIN AND C17. The embedded
evaluator evaluates that route; the handler should not reconstruct which grants
to combine or invent a fallback when required support is ineffective.

Another complete valid route may independently authorize an operation. One
inapplicable grant is not necessarily a final deny. Permissions and scope
fragments from different grants must never be mixed to create a broader route.

**A route the consumer cannot read is not a denial either.** Malformed, internally
inconsistent, carrying identifiers it cannot parse — that is exactly a route that
might have authorized:

| Situation | Result |
|---|---|
| A route is unusable, and another complete route authorizes | **Allow.** The unusable route says nothing about the one that did. |
| A route is unusable, and no other route authorizes | **Evaluation error.** The denial is not established. |
| Every route is readable, and none authorizes | **Completed denial.** |

An unusable route costs that route, and costs the *certainty* of a denial — not the
whole answer. This matters precisely because the answer describes everything the human
holds in the area: failing the whole evaluation on one bad route would take away every
other grant they hold. Reporting an error rather than a denial does tell a person "we
could not check" when the honest answer might have been "you have no access", and that
is the correct trade — a denial asserts something about their authority, and asserting
it from evidence that was never read is the error this prevents.

An answer describing another tenant, application or human is **not** one unusable
route. It means the answer is about somebody else, and nothing in it may be used.

## Enforce the actual data boundary

For the GET above, a constrained lookup can implement the boundary:

```sql
SELECT certificate_id, department_id, display_name
FROM certificates
WHERE tenant_id = :trusted_tenant
  AND department_id = :requested_department
  AND certificate_id = :requested_certificate;
```

This is illustrative SQL, not a canonical query contract or database requirement.
The parameter values must remain bound to the checked request. If C17 belongs to
ENG, this FIN lookup cannot return it. A missing or mismatched result is not a
reason to retry without the department constraint.

Contrast an unsafe path: check Finance scope, log `dept=FIN`, then fetch C17 by
ID alone and return it. The input was used, but not to restrict the effect.
Review must trace the predicates to every protected output or mutation, not
merely search for the variable name.

If a valid grant contributes `{}`, it supplies no local department restriction.
The endpoint declaration alone does not invent one. The inherited/outer bounds
and application contract still apply, and the handler must perform the operation
it actually declared. Self-scoped data must similarly be constrained to the
authorizing human's defined relationship.

For writes, a changed relationship between checking and use must not let the
operation escape its evaluated boundary.

**A move needs authority over both boundaries, and composition is not a grant
question.** The model has no opinion about whether the required authority arrived in
one grant or several: what is required is that the *resolved* authority covers the
permission and the boundary being evaluated. A move is **one endpoint and one gate**,
and the gate evaluates **once per boundary**:

```json
{ "permission": "hrms:payroll:payslip::write", "material": { "dept": "FIN" } }
{ "permission": "hrms:payroll:payslip::write", "material": { "dept": "ENG" } }
```

Each evaluation is a complete route against one boundary, so no fragment is ever
mixed — the concern is structurally absent rather than guarded against. And because
both evaluations happen in one request at one gate, "both hold" means at one moment,
which is what a boundary-changing operation needs.

**Expressing the two boundaries is the endpoint's obligation.** A boundary is
identified by a registered scope key and a request carries one value per key, so a
single evaluation cannot mean two departments at once. An endpoint that moves data
must evaluate the boundary it is moving from and the boundary it is moving to,
separately, and run the effect only if both allow.

## Results: allow, deny and evaluation failure

The approved minimal allow shape identifies contributing grants:

```json
{
  "version": "1",
  "decision": "allow",
  "grant_ids": ["G0", "G1", "G2"]
}
```

`grant_ids` is the **contributing chain of the route that authorized the request,
ordered root first**. Read left to right: the trusted root, the grant narrowing it to
one department, and the grant narrowing that to one certificate. **The order is the
dependency** — each entry is bounded by the one before it. It identifies supporting
grants for *this* evaluation, not every grant held by the human, and it is not a
reusable allow for another request. A returned-scope field is not required by the
chosen minimal contract; enforcement remains mandatory without it.

An endpoint holding only grant ids cannot later say which *assignment* carried the
authority if that assignment has since been deleted. That is acceptable while the
evidence exists to let an endpoint account for its own effect. Authority loading
already answers in the richer shape, so a caller needing it asks the service that
owns those records.

When sufficient evidence establishes that no complete route permits the request:

```json
{
  "version": "1",
  "decision": "deny",
  "error_code": "NO_AUTHORIZING_GRANT",
  "error_message": "You do not have access to this certificate.",
  "error_message_reason": "No grant authorizes this certificate read within Finance."
}
```

When a required Auth lookup times out and no sufficient valid evidence is already
available, evaluation cannot complete:

```json
{
  "version": "1",
  "error_code": "AUTHORITY_TIMEOUT",
  "error_message": "We could not check your access.",
  "error_message_reason": "The authorization service did not respond in time."
}
```

The error variant has no `decision` because no authorization decision completed.
Both deny and error prevent protected execution. Absence of `decision` does not make
arbitrary truncated JSON a valid error response. Reject mixed variants and unknown
result fields.

**The code catalogue is published, its names are fixed, and the list is open.** A
published code never changes meaning; new codes may appear at any time. A consumer
must tolerate a code it does not recognize and fall back to the **class** the result
arrived in — a completed denial, or a failure to establish authority. So a consumer
may not switch exhaustively on codes and must not make a security-relevant choice
from one: that choice is already carried by the allow / deny / evaluation-error
distinction, which is the contract to branch on. A code explains; it does not decide.

An `AUTHORITY_` prefix means the authority answer failed or failed to arrive:
`AUTHORITY_UNREACHABLE`, `AUTHORITY_TIMEOUT`, `AUTHORITY_REFUSED`,
`AUTHORITY_UNAVAILABLE`, `AUTHORITY_UNREADABLE`, `AUTHORITY_AMBIGUOUS`,
`AUTHORITY_OVERSIZED`, `AUTHORITY_MALFORMED`. Unprefixed names the decision itself,
the caller's own configuration, or the way an answer disagreed with its question:
`NO_AUTHORIZING_GRANT`, `UNSUPPORTED_VERSION`, `WRONG_AREA`, `WRONG_SUBJECT`,
`MISCONFIGURED`. `AUTHORITY_UNREADABLE` and `AUTHORITY_MALFORMED` are deliberately
distinct — one answer could not be read, the other was read and could not be used —
and they occur in different layers.

The cost is stated plainly: **no consumer can write an exhaustive handler**, and one
that logs an unrecognized code without alerting on it will swallow a new failure mode
silently. That is accepted, because freezing the list buys exhaustiveness in exchange
for a contract version every time a layer learns something new about how authority can
fail to arrive.

The HTTP status each kind renders as belongs to the application, not to this contract.

Both evaluator-provided messages reach the UI; the UI chooses their presentation.
The second message is therefore not private merely because a screen hides it.
Avoiding secrets and sensitive server detail is necessary integration care;
the final detailed disclosure contract remains pending. The endpoint should not
guess a reason by reimplementing evaluation.

## What the response discloses is the endpoint's duty

The gate authorizes **a request**. What the response body then discloses is not its
question.

The case is concrete. A caller holding write and not the matching read issues an
update. The write is authorized and succeeds, and the response carries the whole
record — fields the caller never supplied and now learns by writing. No rule in this
handbook is violated, and the gate could not have prevented it: it cannot know what a
body contains, and should not be given the job of finding out.

**That does not make it nobody's duty.** Disclosure is authorization work, and it is
assigned where this chapter's split already puts it: the endpoint keeps execution, and
its output, inside the authorized boundary. Endpoint enforcement is mandatory
authorization work, not optional business validation.

A gate that grew a disclosure check would be performing a check it cannot actually
perform, on data it does not model, for every endpoint it guards. One layer cannot be
made answerable for the correctness of every layer above it. The tradeoff is worth
naming: an endpoint author who reads only the policy contract may never think about
the response at all, which is why this is a rule with a name to point at during
review rather than something left implied.

## Missing, invalid and unsupported evidence

Three cases, three different answers, and none of them is an allow:

| | Result |
|---|---|
| Required material missing — a declared input absent, or the binder cannot produce it | failure to establish; refused before a decision is reached |
| A route's own evidence invalid — an ill-formed predicate, content that does not validate, an unreadable chain | that route closes |
| …and no other route authorizes | failure to establish, not a denial |
| Evidence of an unsupported kind — a reserved token other than `$self`, a wildcard where a boundary belongs, a contract version this consumer does not speak | refused, never interpreted, never treated as absent |
| Every route readable, none authorizes | completed denial |

An unsupported value is not an empty one. None of this is a condition engine: every
case concerns evidence this model already defines — its own material, its own
predicates, its own contract versions. Nothing here evaluates a business fact, and
nothing here is extensible by an application.

## Non-HTTP and background work is deferred

Everything in this chapter assumes a synchronous request arriving at one
endpoint-owned gate, with the effect bounded by the evaluated material. Queues,
scheduled jobs, streams and long-running work are **out of scope for v1** — recorded
as a deferral so that silence is not later read as an oversight.

One rule still governs them and is not weakened: queued work is authorized when it
**executes**, not when it was enqueued. A deployment doing this anyway must not assume
that an allow travels with the work. It does not. An evaluation is about a request, a
boundary and a moment.

The cost is real: a deployment with background work has no guidance beyond that single
rule and will invent the rest. That is preferable to inventing it here, where nothing
could check it.

## Integration review checklist

- Identity and tenant context are established, not trusted from arbitrary input.
- The server-owned policy declares the complete protected operation.
- Required values come from the declared sources and pass application validation.
- Evaluation uses complete current authority routes and mandatory proxy limits.
- Deny, malformed results and evaluation errors cannot reach protected effects.
- Every relied-on boundary restricts the actual query, mutation or response.
- Concurrent changes and retries do not detach effects from checked boundaries.
- Diagnostic messages and contributing IDs are interpreted within their stated
  contracts, not as a new authority source.
- The policy declares a tenant correlation, and mounts — a policy that cannot mount
  must stop the endpoint serving rather than serve it unguarded.
- The response's contents, not only the request's authorization, stay inside the
  boundary the decision was made about.
- An unrecognized error code is handled by its class, and raises an alert rather than
  a log line nobody reads.

**Sources:** [system flow](../../docs/system-overview.md),
[endpoint policy, trusted correlations and mount-time validation](../../docs/endpoint-policy-format.md),
[handler integration contract](../../docs/handler-integration-contract.md),
[authority loading transport](../../docs/authority-resolve-transport.md),
[decisions, the code catalogue and unusable routes](../../docs/decision-results.md),
[response disclosure](../../docs/endpoint-authorization.md),
[missing, invalid and unsupported evidence](../../docs/grant-conditions.md),
[composition and boundary-changing operations](../../docs/operation-enforcement.md),
[background work](../../docs/background-authorization.md),
[the open boundary question](../../docs/policy-scope-boundary.md),
[concurrent enforcement](../../docs/concurrent-enforcement.md).

[Next: lifecycle and worked scenarios](08-lifecycle-and-scenarios.md)
