# ABV-123-09 — proposed wire contracts

**Status: D2 design draft; no contract or route below is approved unless it is
explicitly identified as an existing approved canonical block.** This document
does not authorize runtime integration, migration, a generic write API, or use of
a validation result as authority to execute.

This draft applies the current decisions in
[grant records](../../docs/grant-record-reference.md),
[assignment authority](../../docs/assignment-authority.md),
[identity context](../../docs/identity-context.md),
[decision results](../../docs/decision-results.md), and
[Auth write consistency](../../docs/auth-write-consistency.md). It does not pick
the operation-to-permission mappings owned by D1, storage keys/indexes owned by
D3, or names/boundaries for evaluator, validator, or middleware components.

## 1. Contract boundary and route status

The tenant base `/api/v1/{tenant}/abv/` is pinned. Everything following that
base in this document is a route proposal. In particular, the explicit
application segment and operation-bound routes are not canonical yet:

```text
/api/v1/{tenant}/abv/applications/{application}/...
```

Application-wide catalog administration remains a separate proposed family:

```text
/api/v1/applications/{application}/abv/...
```

Path values select requested context; they do not prove identity, installation,
or authority. The adapter must establish trusted tenant, application, and
identity before an endpoint's operation-specific administrative gate and ABV
checks run. A caller-submitted identity block is never proof of those facts.

Every operation route is statically registered. There is no free-form
`POST execute/{operation}`, arbitrary `records/{kind}` write, body-selected
permission, or universal execute permission. D1 supplies each route's one
declared permission and material mapping; this draft deliberately does not.

## 2. Existing blocks versus proposed envelopes

| Block | Status here | Wire use |
|---|---|---|
| Grant control `{version,id,status}` | Existing approved canonical core block | Request/response body where the complete operation proposal is a grant-wide status change. |
| Grant content `{version,grant_id,revision,...}` | Existing approved canonical core block | Immutable authority content nested unchanged in publication input; returned unchanged after commit. |
| Assignment `{version,id,grant_id,grant_revision,recipient,status}` | Existing approved canonical core block | Assignment-create request and committed-record response. |
| Identity `{version,actor,human_id}` | Existing approved canonical block | Established request context, not accepted as self-authenticating operation data. |
| Allow, deny, and evaluation-error result blocks | Existing approved minimal blocks | Authorization result semantics; not mutation-success or reusable authorization tickets. |
| Publication, status-command, validation, diagnostic, list, and conflict wrappers | Proposed by this draft | Transport-only envelopes; their fields require contract approval. They do not alter nested canonical records. |

The URL `v1`, an envelope's `version`, a record's `version`, a grant content
`revision`, and an assignment's `grant_revision` have different meanings. None
defaults or substitutes for another.

## 3. Concrete assignment-create request

Proposed route:

```http
POST /api/v1/acme/abv/applications/hrms/execute/assignment.create
Content-Type: application/json
```

The request body is the existing approved canonical Assignment block, with no
additional transport wrapper:

```json
{
  "version": "1",
  "id": "A2",
  "grant_id": "G2",
  "grant_revision": 2,
  "recipient": {"type": "group", "id": "Team2"},
  "status": "enabled"
}
```

For this example, the trusted adapter has separately established the existing
canonical direct-human identity context:

```json
{
  "version": "1",
  "actor": {"type": "user", "id": "U-MAYA"},
  "human_id": "U-MAYA"
}
```

That second block illustrates established server context; it is not a second
body and cannot be supplied to impersonate Maya. The endpoint must run its D1
operation-specific administrative check for Team2 and independently validate
the proposed authority, latest-on-create selection, current actual support,
permission subset, inherited scope, team ceiling, registrations, validity,
uniqueness, and other applicable constraints. `grant_id: "G2"` alone proves no
support. The assignment is written only if both gates pass against coherent
state.

Proposed successful transport: return `201` and the exact committed canonical
Assignment block above. The body is not an allow result. No success is returned
before commit, and no other assignment, membership, grant control, or revision
is changed. Exact HTTP status selection remains a contract choice; `201` is a
proposal, not an approved mapping.

## 4. Concrete grant-publication request

Proposed route:

```http
POST /api/v1/acme/abv/applications/hrms/execute/grant.publish
Content-Type: application/json
```

Grant publication needs transient source evidence that does not belong in the
canonical GrantContent record. This draft therefore proposes the smallest
operation-specific envelope:

```json
{
  "version": "1",
  "support_assignment_id": "A1",
  "grant_revision": {
    "version": "1",
    "grant_id": "G2",
    "revision": 2,
    "parent_grant_id": "G1",
    "permissions": [
      "hrms:payroll:payslip::read",
      "hrms:payroll:payslip::write"
    ],
    "scope": {"cert": "C17"}
  }
}
```

The outer object is proposed. The nested `grant_revision` is the existing
canonical explicit-permissions GrantContent block. `support_assignment_id` is
request evidence only: it must resolve to the actual eligible parent route held
in the same tenant/application context and must never be persisted as grant
content, issuer history, adoption state, or a permanent dependency.

The operation-specific administrative gate and the authority-boundary checks
both run. The candidate must be a newer immutable revision of the existing
non-root G2, retain the applicable parent identity, use registered content, and
fit the resolved current parent route. Publication does not create or update an
assignment, change a control, select the revision for a recipient, or establish
a root.

Proposed successful transport: return `201` and only the committed canonical
GrantContent block nested above. Omitting the request envelope from the response
prevents transient support evidence from looking canonical. As with assignment
creation, `201` remains a proposed HTTP mapping.

## 5. Operation contract map

All route suffixes below are proposals beneath the pinned tenant base plus the
proposed `applications/{application}/` segment. “Gate” deliberately names an
operation-specific requirement without choosing its permission identifier.

| Proposed route | Request | Success result | Required behavior / limit |
|---|---|---|---|
| `POST execute/assignment.create` | Canonical Assignment | Committed canonical Assignment | Separate administrative and ABV checks; latest revision only; no partial write. |
| `POST execute/grant.publish` | Proposed publication envelope containing canonical GrantContent | Committed canonical GrantContent | Transient support assignment; immutable insert only; no adoption. |
| `POST execute/grant.status` | Canonical GrantControl | Committed canonical GrantControl | Grant-wide operation; enabling checks every required enabled binding, while an explicitly disabled broken assignment remains disabled and does not alone block. |
| `POST execute/assignment.status/{id}` | Proposed `{"version":"1","status":"disabled"}` command | Full committed canonical Assignment | Status only; preserve identity, recipient, and adopted revision. Re-enable validates current reality and does not adopt latest. |
| `POST validate/assignment.create` | Same canonical Assignment as create | Proposed validation result | Performs this validation operation's own read/diagnostic authorization and current ABV analysis; never writes and never authorizes a later execute call. |
| `GET records/assignment/{id}` | Path ID | Canonical Assignment | Kind-specific read gate and visibility; no power inferred from assignment-create or generic inspect access. |
| `GET records/grant/{id}` | Path ID plus a fixed representation selector if separate identity/content/control views are approved | Canonical selected grant block | Do not collapse control and immutable revisions into an ambiguous merged record. Exact selector/routes remain open. |
| `GET diagnostics/assignment/{id}` | Path ID | Proposed diagnostic envelope | Structural/advisory only; not an authorization result, reusable validation, or write guarantee. |
| `GET records/assignment` | Fixed allow-listed filters and bounded keyset cursor | Proposed typed list envelope | No arbitrary predicate language; visibility applies to every item and metadata. |
| `GET adoption-candidates/assignment` | Fixed allow-listed filters and bounded keyset cursor | Proposed advisory list envelope | May report adopted and available revision; does not prove adoption will pass or authorize it. |

Grant and assignment status routes remain distinct because their affected-binding
rules differ. The status-command envelope exists only because assignment status
has no approved standalone control record; it does not propose one. Whether IDs
belong in path or body, and the final route verbs, remain choices.

Role publication, permission/scope registration, team/membership operations,
explicit adoption, retirement, and bootstrap need the same operation-specific
method. They are intentionally not filled with guessed request schemas or D1
permission mappings here. Deletion, reparenting, PostgreSQL, and production Auth
integration remain deferred or excluded from the current build.

## 6. Proposed validation and diagnostic envelopes

A successful validation response needs to be unmistakably non-executable. The
minimal proposal is:

```json
{
  "version": "1",
  "validation": "valid",
  "authorizes_execution": false
}
```

An invalid proposal uses `validation: "invalid"` and the agreed readable/error
field pattern, but its exact code catalogue is still open. It is not a completed
authorization deny about the caller's ability to invoke the validation endpoint.
Conversely, failure of that endpoint's own authorization gate uses the approved
deny/evaluation-error semantics in section 8.

A diagnostic read uses the same non-authorizing marker but must not be called
valid merely because only partial evidence was loaded. Proposed minimum:

```json
{
  "version": "1",
  "diagnosis": "eligible",
  "authorizes_execution": false
}
```

`eligible` means only that the declared diagnostic completed for the inspected
state. Exact issue/detail fields and code values remain open. A caller must send
a new execute request; the execute endpoint re-establishes identity, authority,
support, freshness, and consistency and may reach a different result.

## 7. Proposed read and list envelopes

Single-record reads return the existing canonical block directly when the route
selects an unambiguous record type. This avoids a wrapper that adds no meaning.
Grant identity/control and immutable content must have unambiguous operation-
specific routes or a fixed selector; a generic merged object is not proposed.

Example proposed assignment list request:

```http
GET /api/v1/acme/abv/applications/hrms/records/assignment?recipient_type=group&recipient_id=Team2&limit=50&cursor=opaque
```

Example proposed response, with canonical Assignment values unchanged:

```json
{
  "version": "1",
  "items": [
    {
      "version": "1",
      "id": "A2",
      "grant_id": "G2",
      "grant_revision": 2,
      "recipient": {"type": "group", "id": "Team2"},
      "status": "enabled"
    }
  ],
  "next_cursor": "opaque-next"
}
```

The envelope and filter names are proposed. Each list route defines its fixed
filters, stable order, cursor meaning, maximum/default limit, and item visibility;
those values are unresolved, not inferred here. Pagination is keyset/opaque to
the caller. Unknown kinds, filters, sort expressions, or free-form predicates are
rejected rather than passed to a generic storage query. Empty/exhausted cursor
representation and total-count disclosure remain choices.

Adoption-candidate items are advisory projections, not canonical assignments.
Their future contract may expose assignment/recipient, adopted revision,
available revision, and reviewable differences as approved by Q-102A, but it
must not return `decision: "allow"`, auto-adopt, or promise that a later adoption
will pass current checks.

## 8. Authorization result and error map

These three blocks are existing approved minimal authorization-result contracts,
not newly proposed operation envelopes.

Completed allow, available to the server-side endpoint:

```json
{
  "version": "1",
  "decision": "allow",
  "grant_ids": ["G-17"]
}
```

Completed deny:

```json
{
  "version": "1",
  "decision": "deny",
  "error_code": "NO_AUTHORIZING_GRANT",
  "error_message": "You do not have access to this operation.",
  "error_message_reason": "No complete grant route authorizes this operation."
}
```

Evaluation error, with no `decision` field:

```json
{
  "version": "1",
  "error_code": "AUTH_SERVICE_TIMEOUT",
  "error_message": "We could not check your access.",
  "error_message_reason": "The authorization service did not respond in time."
}
```

The field layouts and variant rules are approved; the shown code spellings and
wording are illustrative because the exhaustive catalogue/value rules remain
open. A valid allow has a non-empty `grant_ids` array. Allow has no error fields;
deny has no `grant_ids`; evaluation error has neither `decision` nor `grant_ids`.
Mixed or unknown result fields are rejected and block protected execution.
Neither deny nor evaluation error permits protected output or effects.

| Failure class | Meaning | Mutation effect | Retry consequence |
|---|---|---|---|
| Malformed body, unknown field, bad version/value, untrusted/mismatched context | Request could not be accepted at its contract/trust boundary | None | Correct the request; repeating unchanged is not useful. Exact public error/HTTP shape is open. |
| Completed authorization deny | Sufficient evaluation established no complete authorizing route | None | Do not automatically retry unchanged. A materially changed/current authority state requires a new request and evaluation. |
| Evaluation error | Required evaluation could not complete; not proof of policy denial | None | Retry only under an operation/client policy for a retryable finalized code; no code catalogue or generic automatic retry is approved. |
| ABV validation rejection | Proposed authority/resulting binding violates a required boundary or is unsupported | None | Correct proposal/state; never reinterpret as authorization allow. Exact public envelope is open. |
| Not found or not visible | Record cannot be returned under that route's visibility contract | None | No cross-kind or cross-area fallback. Concealment versus explicit forbidden/not-found remains open. |
| Consistency conflict | Relevant checked revision or authority state changed before persistence | None; transaction stops | Fresh explicit attempt must rerun both gates against current state. Never reuse validation or silently substitute latest. |
| Commit/response uncertainty | Client did not receive a known success | Server may or may not have committed | Reconcile by operation-specific read before another mutation; do not assume failure or replay blindly. |

HTTP status mapping and any outer error wrapper remain unresolved. A public
transport may carry the approved deny/evaluation-error bodies directly, but it
must preserve their semantic distinction and the agreed UI-visible message
fields. A malformed evaluator result is an integration/evaluation failure, not
a completed deny.

## 9. Retry and replay contract

1. Reads, lists, and diagnostics may be retried as new evaluations. A retry sees
   current state and does not promise the same page, diagnosis, or visibility.
2. Validation may be repeated, but each result is advisory for that call only.
   It is never a ticket, capability, prepared transaction, or authority to call
   execute without all checks again.
3. Assignment-create replay after a lost response is not automatically
   idempotent. Read A2 through its protected read route and compare the complete
   canonical record. A duplicate/conflict does not authorize overwriting or
   treating a different record as success. Whether an explicit idempotency key
   is added later remains open.
4. Grant-publication replay likewise reads `(G2,2)` through its protected exact-
   revision route and compares canonical content. Immutability prevents overwrite,
   but a duplicate error alone does not prove that this caller's attempt created
   the stored revision. No support evidence is recovered from the record.
5. Same-state control requests remain protected operations. Retrying still runs
   the current administrative and applicable structural/enable checks; an old
   success or validation cannot bypass them.
6. A consistency conflict always requires a fresh request/evaluation. Do not
   silently adopt a newer revision, choose another support route, or weaken an
   operation-specific gate to make the retry succeed.
7. The bounded bootstrap continuation rules do not create a general mutation
   replay rule. Bootstrap may resume only the same intended incomplete setup
   after revalidation; ordinary ABV writes need their own explicit replay design.

## 10. Actual choices remaining

- Approve or revise the proposed explicit application path and each operation-
  bound route; the tenant base alone is pinned.
- Decide D1's one-permission and declared-material mapping for every route.
- Approve request-envelope names, especially `grant_revision`,
  `support_assignment_id`, assignment-status path/body placement, and whether
  grant record views use separate routes or a fixed selector.
- Approve success HTTP statuses and the public mapping/wrapper for approved
  allow, deny, evaluation-error, validation-rejection, not-found, and conflict
  semantics; finish the error-code/value catalogue.
- Approve validation/diagnostic field names and detail representation while
  preserving `authorizes_execution: false` semantics.
- Fix each list's filters, order, cursor encoding rules, maximum/default limit,
  end-of-list representation, visibility, and total-count policy.
- Decide whether ordinary mutations need idempotency keys. Until then, use the
  protected read-and-compare reconciliation described above, not blind replay.
- Complete direct-human/proxy support discovery and recipient-relative binding;
  canonical support is retained, but the current prototype's unsupported paths
  are not claimed complete.
- D3 separately decides storage keys, indexes, constraints, and evidence reads.
  Component naming and middleware/validator ownership remain deliberately open.

No item in this list is approved by appearing in this draft.
