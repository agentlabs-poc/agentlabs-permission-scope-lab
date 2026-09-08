# 7. Application integration

[Contents](../README.md) · [Previous](06-auth-service.md) · [Next](08-lifecycle-and-scenarios.md)

**Framework-specific integration guidance.** Authorization is performed within
one endpoint-owned gate. Middleware may authenticate, establish trusted context
and load usable authority, but it does not hand the handler a provisional
allow/prepared result. There is no prepared state in the selected model.

## Follow the request

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

## Declare authorization before writing the handler

Every protected method/route declares exactly one permission in this framework.
That permission covers the endpoint's complete protected operation. The caller
does not choose it. A grant may contain several permissions and a user may have
many grants; neither changes the endpoint's single-permission contract.

When an endpoint combines several responsibilities, its declared operation must
genuinely cover the complete effect, or the endpoint must be redesigned. Checking
read and then performing an unrelated privileged write is not authorized because
both happen inside one function.

The approved partial policy contains `version`, `method`, `path`, `permission`
and named `inputs` with explicit `source` and `name` bindings. It does not contain
a relationship block or duplicate application value-validation rules.

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
  }
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
  }
}
```

Here `proposed_dept` is a local input name, not a new canonical scope key. Its
value comes from the body field `department_id`. A value of FIN states an
intention; it does not prove the record currently belongs to Finance. A body
field naming an employee likewise does not establish the caller's identity.

The handler may process other business fields, but selecting only one field as
authorization material does not authorize arbitrary extra effects. The complete
operation must remain within the permission and boundary actually evaluated.
Nested input selectors, further source kinds and complete policy validation
remain pending; do not introduce an unapproved path-expression syntax.

## Resolve authority without making the endpoint inspect grants

Authority loading considers the human's direct assignments and valid group
memberships, group assignments, adopted grant and role content, controls,
validity, parent support and applicable delegation. Resolved grants remain
dependent views of these records, not permanent copied authority.

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
operation escape its evaluated boundary. If the operation moves data between
departments, the agreed model requires authority over both current and proposed
boundaries. How separate grants compose for that move remains pending; the PUT
example does not silently decide it.

## Results: allow, deny and evaluation failure

The approved minimal allow shape identifies contributing grants:

```json
{
  "version": "1",
  "decision": "allow",
  "grant_ids": ["G2"]
}
```

The list is non-empty and identifies supporting grants for this evaluation, not
every grant held by the human. It is not complete lineage evidence or a reusable
allow for another request. A returned-scope field is not required by the chosen
minimal contract; enforcement remains mandatory without it.

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
  "error_code": "AUTH_SERVICE_TIMEOUT",
  "error_message": "We could not check your access.",
  "error_message_reason": "The authorization service did not respond in time."
}
```

The error variant has no `decision` because no authorization decision completed.
Both deny and error prevent protected execution. The code spellings are
illustrative; the exhaustive catalog and HTTP mapping remain pending. Absence
of `decision` does not make arbitrary truncated JSON a valid error response.
Reject mixed variants and unknown result fields under the selected v1 rules.

Both evaluator-provided messages reach the UI; the UI chooses their presentation.
The second message is therefore not private merely because a screen hides it.
Avoiding secrets and sensitive server detail is necessary integration care;
the final detailed disclosure contract remains pending. The endpoint should not
guess a reason by reimplementing evaluation.

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

**Sources:** [system flow](../../docs/system-overview.md),
[endpoint policy](../../docs/endpoint-policy-format.md),
[decisions](../../docs/decision-results.md),
[concurrent enforcement](../../docs/concurrent-enforcement.md).

[Next: lifecycle and worked scenarios](08-lifecycle-and-scenarios.md)
