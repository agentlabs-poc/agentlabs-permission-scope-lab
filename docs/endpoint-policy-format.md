# Endpoint policy format — approved partial structure

Approval follow-up: the user approved the presented PUT example after requesting
it. This confirms the body-to-local-input illustration under Q-050-B, including
the distinction between a requested department and established current facts.
It does not approve relationship syntax or source/destination update rules.

Q-050-B approves CONTRACT-011: `version`, `method`, `path`, one `permission`,
and named `inputs` with explicit `source`/`name` bindings. Relationship bindings
are not yet settled (Q-050-C); these examples are approved structure illustrations,
not complete deployable or published endpoint policies. Including a version does
not by itself complete a contract.

## GET example: selected path parameters

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

| Field | Meaning and rationale |
|---|---|
| `version` | The shared top-level string contract version, initially `"1"`. |
| `method` and `path` | Bind the policy to an endpoint within the application; the same path with another method may require a different permission. |
| `permission` | Exactly one required permission, mapping the operation directly without a redundant separate action field. |
| `inputs` | Locally name selected request inputs that authorization may use. These names are not automatically scope boundary keys. |
| `source` and `name` | Explicitly identify the source location and parameter/field name independently of the local input name. |

For a route parameter named `certificateId`, the local input could still be
`cert`, with `source: path` and `name: certificateId`. Explicit source/name was
selected over shorthand such as `path.cert`, avoiding a punctuation-based selector
grammar and making source validation clearer. Path and selected body fields are
illustrated; nested body selection syntax and additional source kinds remain open.

## PUT example: path identifiers and selected body values

Assume an application-defined certificate-write permission covers this endpoint's
complete update operation and is registered by the application. The identifier
below is an illustrative catalog entry, not a universal permission definition.

The request is `PUT /api/v1/acme/certificates/C-17` with this illustrative business
body (a sample payload, not a newly published application request contract):

```json
{
  "department_id": "FIN",
  "display_name": "Employment certificate"
}
```

Its partial endpoint policy uses the approved structure:

```json
{
  "version": "1",
  "method": "PUT",
  "path": "/api/v1/{tenant}/certificates/{cert}",
  "permission": "hrms:employee:certificate::write",
  "inputs": {
    "tenant": { "source": "path", "name": "tenant" },
    "cert": { "source": "path", "name": "cert" },
    "proposed_dept": { "source": "body", "name": "department_id" }
  }
}
```

| Local input | Source | Meaning in this example |
|---|---|---|
| `tenant` | Path `tenant`, yielding `acme` | Requested tenant identifier; must agree with trusted tenant context. |
| `cert` | Path `cert`, yielding `C-17` | Certificate identified by the request; execution must remain bound to it. |
| `proposed_dept` | Body `department_id`, yielding `FIN` | Requested department value, not proof of the certificate's current department. |

`proposed_dept` is a local name chosen for clarity, not a new canonical scope key
or reserved keyword. There is no `department_id` path parameter: it is explicitly
read from the body. `display_name` is not selected as an authorization input in
this example; the handler still validates and processes it under its application
contract. Selection does not authorize arbitrary additional body fields or
automatically turn a body value into a trusted relationship fact.

### Safety rationale and counterexample

If C-17 currently belongs to Engineering, submitting `department_id: FIN` does
not establish that Finance-only authority permits modifying it. Relevant current
facts and any required proposed-state checks must still be established and
enforced. The exact update/move authorization contract is still open; this input
binding example does not choose source-only, destination-only, or both-boundary
rules. It shows how the proposed value enters evaluation without masquerading
as an existing fact.

The same principle applies to a body employee identifier: it describes a requested
value, not verified caller identity or proof of current ownership. Those remain
distinct from authenticated context and application-established relationships.

## What stays implicit and what remains open

The declaration is server-owned and fixed. Verified identity/tenant context and
shared Auth authority loading need not be copied into every policy. Route tenant
claims must still be bound to trusted context; field names alone do not prove
relationships. An applicable `{}` scope does not invent department restrictions
merely because department input is declared; other mandatory checks remain.

Q-050-C must define how relationship bindings establish required facts and connect
them to registered scope meanings. Full validation rules, missing-input handling,
nested-body syntax, additional sources, schema publication, and update/move
enforcement remain open. This approval does not finalize the whole endpoint policy.
