# Decision results — working discussion

## Q-051 / DECISION-003 — completed decision versus evaluation error

Status: **AGREED.** The user approved the distinction, requested an Auth-service
timeout example, and then confirmed that example. Original status retained as
history: ~~PROPOSED, awaiting user approval~~. This closes the representation
principle only; it does not finalize a JSON contract or change runtime behavior.

### Existing foundation, not a new question

CONTRACT-006 places authorization at one endpoint-owned gate. ENFORCEMENT-002
already prevents protected execution when required material is missing or
evaluation fails. Failure remains diagnostically different from a completed
denial. DECISION-001/002 require complete positive authority routes and preserve
mandatory restrictions. None of these rules is reopened here.

### Agreed representation principle

Keep the completed authorization decision binary: **allow or deny**. Report
inability to complete evaluation as a separate **evaluation error**, not a third
authorization decision. An error supplies no authorization and cannot permit
protected output, mutations, or business side effects. This is not a prepared
handoff and does not let the handler proceed with unfinished authorization.

“Separate” is a semantic distinction, not a requirement to throw exceptions,
make another network call, or use a particular transport envelope.

| Situation | Agreed meaning | Endpoint consequence |
|---|---|---|
| Sufficient valid authority establishes Finance certificate-read access | Completed allow within the established boundary | Enforce the trusted tenant, Finance boundary, and requested certificate against the actual read before returning data. |
| Sufficient evidence conclusively establishes that no complete applicable route authorizes the operation | Completed deny | Stop protected execution. One failed grant alone does not establish this if another route could authorize it. |
| Required authority cannot be obtained and no sufficient valid preloaded evidence is available | Evaluation error; no completed decision | Stop protected execution; retain the operational failure for diagnosis. |

The first case does not prove that C-17 belongs to Finance: the endpoint's
constrained read must enforce that relationship. Missing declared inputs and
invalid application values retain their existing rejection rules; this decision
does not reclassify every request-validation failure as an evaluator error.

### Concrete timeout example — approved under Q-051

Vinay requests `GET /api/v1/acme/FIN/C-17`; the endpoint requires
`certificate.read`. Authentication establishes Vinay's identity and tenant.
The Auth Agent calls Auth to obtain the required authority, but the call times
out. No sufficient valid authority is already available.

Evaluation therefore ends with an error, not a completed deny. The endpoint
stops and returns no certificate. The timeout establishes inability to finish
checking, not that Vinay lacks Finance access. It does not authorize a fallback
read. The absence of sufficient valid preloaded authority matters: this example
does not make every failed optional or redundant service call fatal, nor approve
any cache freshness or fallback policy.

### Rationale and alternatives

The agreed distinction separates **what authority permits** from **whether evaluation
could finish**. This makes access decisions and operational failures distinguishable
for handlers, audit, and diagnosis without weakening fail-closed behavior.

An alternative is a single result discriminator containing allow, deny, and
error. That can be implemented safely, but mixes completed decisions with
evaluation status. Treating every failure as an undifferentiated denial is
simpler for a boolean caller but loses the diagnostic distinction already
required by the handbook. A client-facing response may still conceal internal
details; internal classification does not require disclosing grants or failures.

Counterexample: an Auth outage must not be recorded as proof that Vinay lacks
Finance access. Equally, an error must never authorize an unconstrained database
read or revive the deprecated prepared state.

### Remaining questions

Exact versioned result shape, reason codes, returned restrictions/provenance,
public versus internal details, HTTP mapping, and failure aggregation across
candidate routes remain open. Freshness and retry policy are separate branches.
Approval of this principle alone does not close HC-08-02 in
[MEASURE-001](handbook-completion-audit.md).

**Q-051 — answered yes:** Should the contract keep completed decisions as allow/deny and report
evaluation errors separately, with both deny and error blocking protected execution?

## Q-052 / DECISION-004 — an internal reason for denial

Status: **AGREED.** The user confirmed: “yes deny should include reason, very
important” and explicitly reaffirmed recording rationale. Original status
retained as history: ~~PROPOSED, not approved~~.

### Agreed rule

Every completed deny must include a machine-readable reason for the calling
endpoint and internal diagnostics. The reason is required, not optional.
A bare deny identifies the outcome but not its cause.

### Example and counterexample

Example: Vinay's authority loads successfully, and evaluation conclusively finds
no complete grant route authorizing `certificate.read` for the request. The result
is deny with the meaning “no authorizing grant.” The reason must describe the
established conclusion, not infer one from a timeout or a single failed candidate.

Counterexample: returning only deny and expecting the endpoint to guess whether
authority was absent or an outer boundary prevented access does not satisfy this
rule. Nor may a timeout be mislabeled “no authorizing grant”: Q-051 requires
an evaluation error when the required check could not complete.

### Rationale, alternatives, and consequences

The reason preserves why the evaluator rejected the request at the point where
that conclusion is known. It lets the endpoint and internal diagnostics explain
the result consistently, supports audit and troubleshooting, and avoids callers
guessing the cause or reimplementing evaluation to recover it. Machine-readable
reasons avoid dependence on parsing free-form prose. This explains the user's
emphasis that the reason is an important part of a denial, not merely helpful
optional logging.

The alternative is a bare deny with explanation available only through separate
logs. It keeps the result smaller, but the endpoint cannot directly distinguish
known causes. That alternative is not adopted. Implementations must preserve the
reason alongside the denial; including a reason does not change the denial's
effect or authorize any protected execution.

### Remaining questions

This decision does not require exposing internal reasons to the client. Exact
reason names, fields, precedence, public responses, and reasons for other outcomes
remain separate questions. No JSON schema or exhaustive reason catalogue is
adopted here. “No authorizing grant” is an explanatory example, not an approved
wire-level code. HC-08-02 remains open until its full closure criterion is met.

## Q-053 / DECISION-005 — client-facing disclosure of denial reasons

### Agreed refinement — both evaluator-provided messages reach the UI

The user explicitly answered Q-053-A: “the messages should reach ui. with
error_message, error_message_reason. the error message is more user facing.”
The evaluator supplies both named fields, and both reach the requesting UI.
`error_message` is the more user-facing message; `error_message_reason` supplies
the reason. The UI decides how to present them. The earlier recommendation to
keep the second message server-side is **not adopted**.

Rationale: the evaluator knows the established cause and can supply a consistent
explanation without applications reconstructing it. Delivering both gives the UI
the explanatory information as well as the primary display message. Calling the
second field “internal” no longer describes its visibility: it is client-visible
even if the UI does not display it. The deny/error distinction remains Q-051;
message delivery does not permit execution or alter the decision.

For illustration, the primary message could say “You do not have access to this
certificate,” and the reason could explain “No grant authorizes this certificate
read within Finance.” This describes meaning, not a finalized choice between
reason prose and a machine-readable code. Q-052's machine-readable reason
requirement remains; how it maps to this format is still to be clarified, without
silently adding a third field.

Security consequence: both delivered fields are inspectable by the recipient.
Recommendation, not a separately approved redaction contract: the evaluator
should keep secrets, unrelated users' data, and sensitive server diagnostics out
of both fields. UI hiding is not a confidentiality boundary. Detailed safe-content
rules remain open; the delivery decision itself is settled.

The field names and delivery are agreed; the complete versioned result envelope,
exact value contracts, and HTTP mapping remain open.

### Earlier direction — evaluator supplies both messages, retained as history

The user corrected the original proposal: “everything is provided from evaluator,
it can have 2 records client message and internal message. ui decides how to show
it.” Record evaluator ownership of both messages and UI ownership of presentation
as the user's direction. Do not require the application to invent the client
message from an internal reason. Exact record/field layout is not finalized.

Rationale for this arrangement: the evaluator already knows the established
cause, so it can produce consistent client-facing wording and internal diagnostic
detail together. The UI can choose how to present the client-facing information
without reproducing authorization reasoning. Q-052's machine-readable reason
remains required; the messages do not silently replace that requirement.

Example, explanatory wording only:

- Client message: “You do not have access to this certificate.”
- Internal message: “No complete grant authorizes certificate.read within Finance.”

### Q-053-A — server-only second message proposal, not adopted; historical

Recommend that both messages be available to the server-side endpoint, but only
the client-safe message be delivered to the ordinary requesting client. The UI
controls its presentation. Server-side diagnostic access remains separate.
Hiding an internal message in the UI would not protect it if it were already
included in a browser response: the recipient could inspect the response itself.

This delivery boundary is **proposed, not yet approved**. The user's statement
does not establish that internal diagnostics should be sent to the browser.
Exact public response, localization, internal-access policy, and JSON schema
remain open; no runtime behavior is implemented.

**Q-053-A:** Should the endpoint keep the internal message server-side and send
only the client-safe message to the requesting UI?

### Original proposal — superseded by the user's direction, retained as history

Status: **PROPOSED, not approved.** Q-052 requires an internal reason. This
question concerns only who controls what the external client receives.

Recommend that the evaluator provide the required reason to the endpoint, while
the application explicitly maps it to a safe client-facing response. Do not
automatically forward internal reasons or diagnostic details. The mapping may
be shared application infrastructure; this does not require each handler to
invent its own policy or add fields to the endpoint declaration.

Example: the internal denial means “no authorizing grant for certificate.read
within Finance.” The application could return “You do not have access to this
certificate.” These are explanatory messages, not finalized codes or response
schemas. The application must not turn the denial into authorization or replace
the internally established reason with its public wording.

Rationale: internal diagnosis needs the actual cause, but permission structure,
membership, or other contextual details may not be suitable for the requesting
client. The application knows the disclosure needs of its API and audience.
Separating these responsibilities preserves internal explanation without making
all internal detail public by default.

The alternative is to forward the evaluator's reason directly. That is simpler
but couples diagnostic detail to the public API and provides no explicit
disclosure boundary. A deliberately approved safe public reason may still be
specific; the proposal does not require every client to receive identical text.

Counterexample: serializing the entire internal denial object to a client without
an explicit safe mapping would bypass this boundary. Mapping the public response
must not erase the internal reason or permit protected execution.

Exact public codes, HTTP status, logging/redaction rules, and whether to conceal
the existence of particular records remain open. No runtime change is adopted.

**Q-053:** Should the application control client-facing disclosure through an
explicit safe mapping, rather than forwarding internal denial reasons automatically?

## Q-054 / DECISION-006 — same message fields for evaluation failures

Status: **AGREED.** The user answered “yes” to using the same two fields for
evaluation errors. Original status retained as history: ~~PROPOSED, not approved~~.

Evaluation errors use the same `error_message` and `error_message_reason` fields
as denials. The evaluator provides both, both reach the UI, and the UI controls
presentation. Their distinct meaning under Q-051 remains: a completed denial is
not inability to finish evaluation, even when their message formats match.

Rationale: the UI can present both kinds of failure consistently without separate
display contracts, while the actual outcome still distinguishes lack of authority
from an operational evaluation failure. Sharing a message structure neither
creates a third completed decision nor revives a prepared state. Both cases still
block protected execution.

Example: an Auth timeout with no sufficient valid authority could carry the
primary message “We could not check your access,” with
the reason meaning “The authorization service did not respond in time.” These
are illustrative meanings, not reason-code or retry-policy decisions.

The alternative is separate message-field names for denial and evaluation
failure. That distinguishes their presentation structurally, but forces the UI
to handle two message formats for the same display purpose. Reusing message
fields must not erase the distinct decision/error status. No particular status
field or transport wrapper is selected by this question.

Counterexample: identical message keys do not justify logging an Auth timeout as
a completed denial or proceeding with protected execution. Complete versioned
envelopes, exact reason encoding, and safe-content rules remain open. This narrow
agreement does not finish the full HC-08-02 checkpoint.

## Q-055 / DECISION-007 — stable code alongside readable messages

Status: **AGREED.** The user explicitly answered “Q-055 yes.” Original status
retained as history: ~~PROPOSED, not approved~~. The evaluator supplies an
`error_code` alongside the two readable messages, to represent the
machine-readable cause required by Q-052. The code accompanies the messages
delivered to the UI; it is not inferred by parsing their text.
For an Auth timeout the illustrative code could be `AUTH_SERVICE_TIMEOUT`, while
`error_message` and `error_message_reason` remain readable explanations.

Rationale for the additional field: software can identify the cause
without parsing wording that might be edited or translated. UI explanation and
machine classification then have distinct responsibilities. The code is
provided by the evaluator, not inferred by the UI from text. This does not by
itself settle the separate completed-decision versus evaluation-error envelope.

The alternative is to make `error_message_reason` itself a stable code, avoiding
a new field but giving up its readable explanation, or to parse its prose, which
makes machine behavior depend on wording. Exact code names, catalogue, evolution,
and the full versioned schema are not finalized by this decision. These
alternatives are not adopted: `error_message_reason` remains readable text, and
clients should use the code rather than matching that text for machine handling.

Counterexample: changing the explanation's wording must not change how software
identifies the cause. The illustrative `AUTH_SERVICE_TIMEOUT` spelling is not
yet a finalized catalogue entry, and a code does not authorize automatic retry
or collapse the distinction between evaluation error and completed denial.

**Q-055 — answered yes:** Should we add `error_code` for the stable machine-readable cause while
keeping both agreed message fields readable?

## Q-060 / DECISION-008 — supporting-grant references with allow

Status: **AGREED.** The user approved Q-060 and clarified that the evidence may
be needed for audit even though not every request will be tracked. Original
status retained as history: ~~PROPOSED, not approved~~.

An allow result provides the server-side endpoint with references to the complete
grant routes actually used to justify that allow. The evidence is available in
the result regardless of whether that request is selected for audit recording.

DECISION-001 already preserves route provenance during evaluation; this question
is specifically whether that evidence is available in the returned allow result,
rather than remaining only inside the evaluator. It does not create new grants
or require returning every grant the human holds.

Example: Vinay's Finance group grant G-17 supplies the required read permission
within Finance, under its applicable restrictions and dependencies. The allow
result identifies G-17 as supporting authority so the endpoint can associate the
operation with why it was authorized. This reference does not prove that C-17
belongs to Finance or replace the endpoint's constrained read.

Rationale: the endpoint and audit path should not have to reconstruct which
authority justified access. A bare allow with evidence held only in evaluator
logs is smaller, but makes that association depend on a separate lookup. An
allow result with supporting references makes the association explicit while
remaining a dependent evaluation result, not a reusable authorization grant.

### Audit availability is separate from recording every request

The user clarified: “eventually may need it for audit. we may not track all
request. but incase required. we should have it.” This decision makes supporting
references available; it does **not** require persistent audit records for every
request. Selection, storage, retention, and any operation-specific mandatory
audit rules remain separate decisions.

Rationale for the distinction: producing evidence when evaluation knows the
supporting authority enables an audit path without forcing all requests into
persistent storage. A caller that chooses to record the result can retain that
evidence instead of rerunning authorization later against potentially changed
grants. Conversely, if evidence is not persisted, this rule does not promise
that a historical decision can later be reconstructed. Reference-only results
also do not settle grant snapshots or historical version retention.

Counterexample: omitting supporting references because audit logging is currently
disabled loses the result's agreed traceability. Persisting every request is not
the only way to make those references available to an audit-capable caller.

This decision concerns server-side result evidence, not automatic UI disclosure.
It does not select field names, grant snapshot/version format, provenance-chain
encoding, audit storage, or how many alternative routes to evaluate and return.
Returning a grant identifier is not a substitute for enforcing all restrictions.

**Q-060 — answered yes:** Should the allow result return its supporting-grant references to the
endpoint for traceability?

## Q-061 / DECISION-009 — evaluated boundary information with allow

### Current conclusion — not required

Status: **NOT ADOPTED.** The user answered “not required.” Do not require scope
or evaluated-boundary fields in the allow result. Q-060's supporting-grant
references remain required; they serve traceability rather than a new boundary
delivery contract.

Design rationale: this keeps the result focused on the decision and its
supporting references rather than duplicating enforcement information in a new
return format. This explains the consequence of the user's choice; the user
did not separately prescribe how internal endpoint/evaluator context is stored.

The existing endpoint-owned gate and actual-use enforcement obligations remain
unchanged. The endpoint must use its correctly bound request/evaluation material
to constrain output and effects. It must not treat the absence of returned scope
as tenant-wide authority or use a later changed grant to broaden a prior decision.
No second lookup, prepared state, or new policy field is required by this choice.

Example: an allow supported by G-17 does not need to repeat `{"dept":"FIN"}`
in its result. The Finance certificate endpoint must still constrain the actual
read to the trusted tenant, applicable Finance boundary, and requested certificate.
Returning that certificate through an unchecked ID-only read remains incorrect.

The original proposal and its trade-offs follow as history, not current rules.

### Original proposal — not adopted, retained for rationale

Status: **PROPOSED, not approved.** Recommend that the allow result return the
evaluated boundary information the endpoint must enforce, associated with each
supporting authority route. Supporting references explain why access was allowed;
the associated boundaries explain the allowed reach for execution.

Example: a Finance read authorized through G-17 returns its Finance boundary,
illustrated by the scope fragment `{"dept":"FIN"}`, alongside the supporting
reference. The trusted tenant remains the mandatory outer boundary. The endpoint
must still constrain its actual read to that tenant, Finance, and the requested
certificate; the boundary is not proof that an unchecked certificate belongs to
Finance.

Rationale: references alone would require the endpoint to recover boundaries
from other state or reload grants, potentially obtaining values different from
those evaluated. Returning the evaluated boundaries keeps enforcement tied to
the decision. This carries forward the existing enforcement obligation; it is
not a prepared state or a second business-authorization decision.

The alternative is to keep the boundary only in shared request/evaluator context
and return references alone. That can work if the context remains correctly bound,
but leaves the return contract less explicit. The proposed result must not flatten
different grant routes into a broader invented scope or discard mandatory limits.

Exact fields, resolved-value representation, condition/delegation evidence,
multi-route encoding, and UI disclosure remain open. No new permission or scope
authority is created by returning this dependent evaluation information.

**Q-061:** Should an allow result carry its evaluated boundaries alongside the
supporting-grant references for endpoint enforcement?

## Q-062 / DECISION-010 — minimal allow-result JSON

Status: **AGREED.** The user answered “062 agree.” Original status retained as
history: ~~PROPOSED, not approved~~. The minimal allow representation is:

```json
{
  "version": "1",
  "decision": "allow",
  "grant_ids": ["G-17"]
}
```

The concrete `decision` and `grant_ids` field names are agreed here. `version`
follows CONTRACT-010. The grant IDs identify supporting grants from this evaluation,
not every grant held by the human or a new grant assignment. The existing trusted
tenant/request context still binds the result. Q-061 adds no scope field, and
denial/evaluation-error message fields are not part of this allow example.

Rationale: an ID array represents the references required by Q-060 without
embedding complete grant records or redundant boundary data. The alternative is
a list of reference objects, useful if required reference metadata is later
identified but adding structure not yet justified by the current minimum.

Counterexample: `grant_ids` must not be a dump of the human's unrelated grants,
and the allow must not be reused to authorize a different request. Omitting scope
does not remove the endpoint's existing enforcement responsibility. Producing the
references does not mandate persistent audit logging for every request (Q-060).

This is an approved base shape, not the complete published result schema.
Grant-version/snapshot evidence, full validation, request correlation, and other
result variants remain separate questions. It does not make a result reusable
outside the evaluation that produced it or promise historical grant recovery.

**Q-062 — answered yes:** this is the minimal allow-result shape.

## Q-063 / DECISION-011 — minimal deny-result JSON

Status: **AGREED.** The user answered “agree” to Q-063. Original status retained
as history: ~~PROPOSED, not approved~~. The minimal deny representation is:

```json
{
  "version": "1",
  "decision": "deny",
  "error_code": "NO_AUTHORIZING_GRANT",
  "error_message": "You do not have access to this certificate.",
  "error_message_reason": "No grant authorizes this certificate read within Finance."
}
```

The example assumes sufficient valid evidence conclusively establishes that no
complete authority route authorizes this request. `NO_AUTHORIZING_GRANT` is an
illustrative code, not an approved exhaustive catalogue or a reason for every
denial. Other denial causes must preserve their actual established meaning.

Rationale: this uses the same `decision` field as allow, keeps the stable code
separate from readable messages, and implements the agreed requirement that a
deny explain its cause. The evaluator provides both messages for UI presentation.
No new field names are introduced beyond the already agreed fields.

No `grant_ids` field is required for this minimum: Q-060 requires supporting
references for an allow, not a list of rejected candidates for a deny. An empty
array or separate evaluation trace would add data without establishing why any
candidate failed; detailed denial traces remain a separate question.

Counterexample: an Auth timeout cannot use this completed deny shape to claim
that no permission exists. Evaluation failure remains separate under Q-051; its
concrete representation is now agreed in Q-064 below. Historically, that variant
was the next open question when Q-063 was approved.
This decision does not require returning sensitive diagnostics to the client.

Full validation, exact code catalogue, HTTP mapping, and complete schema publication
remain open. **Q-063 — answered yes:** this is the minimal deny-result shape.

## Q-064 / DECISION-012 — minimal evaluation-error JSON

Status: **AGREED.** The user answered “agreed” to Q-064. Original status retained
as history: ~~PROPOSED, not approved~~. Evaluation errors use the agreed error
fields without a `decision` field because evaluation could not complete:

```json
{
  "version": "1",
  "error_code": "AUTH_SERVICE_TIMEOUT",
  "error_message": "We could not check your access.",
  "error_message_reason": "The authorization service did not respond in time."
}
```

This uses the approved Q-051 timeout case: required authority could not be loaded
and no sufficient valid authority was already available. The code spelling is
illustrative pending the catalogue. Both messages are evaluator-provided and
reach the UI under Q-053/054; their presentation does not create authority.

Rationale: absence of a completed decision reflects the fact that neither allow
nor deny was established. Reusing the existing error fields avoids a third
`decision` value or another status field in this minimal format. The endpoint
must stop protected execution, as already agreed for evaluation failure.

The alternative is an explicit evaluation-status field or a distinct nested
error envelope. That makes the variant visibly tagged but adds structure beyond
the current fields. These alternatives are not adopted for the minimal shape.
This decision does not select exception-based transport,
HTTP mapping, or a retry policy.

Validation consequence: omission of `decision` alone must not make an arbitrary
or truncated response a valid evaluation-error result. Consumers must validate
the complete expected error shape; malformed results cannot become allow or a
completed policy denial. Full validation, variant/code compatibility, and the
code catalogue remain open and must be finalized before publishing the schema.

**Q-064 — answered yes:** this is the minimal evaluation-error shape, with no
`decision` field because no authorization decision was reached. The example's
`AUTH_SERVICE_TIMEOUT` spelling remains illustrative, not a finalized code entry.

## Q-065 / DECISION-013 — reject mixtures of result variants

Status: **PROPOSED, not approved.** Recommend rejecting a result that mixes
variant-specific fields from the agreed allow, deny, and evaluation-error shapes,
rather than choosing one interpretation and ignoring the contradictory fields.

Intentionally invalid example under this proposal:

```json
{
  "version": "1",
  "decision": "allow",
  "grant_ids": ["G-17"],
  "error_code": "AUTH_SERVICE_TIMEOUT"
}
```

This claims a completed allow while also carrying an error field that does not
belong to the allow variant. The proposed rule rejects this response as malformed
and blocks protected execution; it does not infer either completed denial or a
verified timeout from the conflicting payload.

Rationale: the same response must not lead one consumer to authorize execution
while another treats it as failed evaluation. The alternative, accepting the
leading discriminator and ignoring conflicting fields, permits inconsistent
handling and can hide integration defects.

Under this proposal, allow does not carry the three error fields; deny does not
carry `grant_ids`; evaluation-error results carry neither `decision` nor
`grant_ids`. Denial traces and success warnings are not implicitly introduced by
reusing another variant's fields. Fail-closed behavior is already agreed; the
new question is the explicit rejection of mixed known-variant fields.

Required value types, grant-ID cardinality, unrelated unknown extension fields,
code/variant compatibility, and the full schema remain separate validation
questions. **Q-065:** Should mixed result variants be rejected rather than
partially interpreted?
