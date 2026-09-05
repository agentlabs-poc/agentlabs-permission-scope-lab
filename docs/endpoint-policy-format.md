# Endpoint policy format — approved partial structure

## No relationship block — CONTRACT-012 / Q-050-C, agreed

The adopted policy retains `version`, `method`, `path`, one `permission`, and
selected `inputs` with their sources. No `relationships` block, named resolver,
or argument-mapping contract is adopted.

> The endpoint predeclares one required permission and selected inputs with
> their sources. The endpoint implementation must establish or enforce the
> application relationships necessary to keep execution within the authorized
> boundaries.

This refines CONTRACT-007/008's earlier predeclared-relationship wording. The
relationship responsibility remains mandatory in implementation, not policy
syntax. The GET/PUT examples below remain structurally applicable. Full policy
validation, missing-input handling, nested-body selection, and publication remain
open; relationship-block design is no longer an unanswered v1 option.

### Rationale and conscious tradeoff

The user chose simplicity by placing relationship enforcement with the endpoint.
Auth establishes authority for the supplied boundary under the complete grant
and mandatory constraints. The endpoint keeps actual execution within that
boundary, without needing a canonical relationship language or resolver interface.

Auth cannot independently detect every incorrectly implemented endpoint. This
is a conscious responsibility split, not permission to trust arbitrary claims
about existing records. Endpoint enforcement is mandatory authorization work,
not optional business validation. Review/tests must cover mismatched tenant,
department, certificate, and applicable self boundaries. This responsibility
split introduces neither independent downstream authority nor a prepared handoff.

### Grant, request, and enforcement example

Assume valid membership and other grant constraints. This working grant example
supplies Finance certificate-read; lifecycle details are omitted for focus:

```json
{
  "version": "1",
  "recipient": { "type": "group", "id": "certificate-readers" },
  "permissions": ["hrms:employee:certificate::read"],
  "scope": { "dept": "FIN" }
}
```

For `GET /api/v1/acme/FIN/C-17`, Auth can establish that Vinay may read within
Finance. That does not establish that unchecked C-17 belongs to Finance. The
endpoint must bind execution to the trusted tenant, requested certificate, and
authorized Finance boundary. A constrained lookup can enforce this directly:

```sql
WHERE tenant_id = :trusted_tenant
  AND department_id = :requested_department
  AND certificate_id = :requested_certificate
```

This illustrates application enforcement, not a query grammar in policies or
grants. An Engineering certificate cannot be returned by this operation. An
unchecked ID-only lookup followed by disclosure violates the contract even if
scope evaluation accepted the supplied Finance value.

With `{}`, no additional department restriction comes from that grant's scope.
Tenant isolation, validity, conditions, human/delegation limits, and application
requirements remain. A self-scoped operation must similarly remain within the
authorizing human's defined self relationship. PUT's requested Finance is not
proof of current Finance membership; exact update/move rules remain open.

### Earlier relationship proposal — not adopted

The original Q-050-C proposal added this block to the versioned GET policy.
This is a historical fragment, not an approved standalone contract:

```json
{
  "relationships": {
    "dept": {
      "resolver": "certificate.department",
      "arguments": { "certificate": "cert" }
    }
  }
}
```

Its rationale was explicit binding of registered boundary meanings to application
facts, separate from request extraction. The argument named a local input; the
resolver supplied a fact, not an authorization decision. After requesting a
grant example, the user chose mandatory endpoint enforcement to avoid this added
machinery. The proposal was never approved. Earlier Q-050-C-open notes below are
preserved history, superseded by CONTRACT-012; the actual relationship still matters.

### Review follow-up — Q-050-D, proposed refinement

The user suggested reviewing whether all material is used to generate output,
with the expectation that this would catch any breach. The proposed refinement
is to verify that every authorization constraint relied on actually restricts
the output or mutation, not merely that its input appears somewhere in code.
Logging the department, or using it in an OR branch that still admits other
departments, does not enforce Finance containment. Conversely, `{}` imposes no
department scope restriction merely because a department input is declared.

This is a strong review criterion, not a guarantee of detecting every breach.
Wrong permissions, untrusted context, stale dependencies, or output paths bypassing
the checked operation can escape a simple input-usage review. The refinement is
proposed, not a new approved rule or replacement for mandatory checks/tests.

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
