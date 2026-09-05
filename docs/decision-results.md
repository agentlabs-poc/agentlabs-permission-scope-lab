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

Status: **PROPOSED, not approved.** One question only: should a completed deny
include a machine-readable reason for the calling endpoint and internal
diagnostics? Recommend yes; a bare deny identifies the outcome but not its cause.

Example: Vinay's authority loads successfully, and evaluation conclusively finds
no complete grant route authorizing `certificate.read` for the request. The result
is deny with the meaning “no authorizing grant.” The reason must describe the
established conclusion, not infer one from a timeout or a single failed candidate.

The alternative is a bare deny with explanation available only through separate
logs. It keeps the result smaller, but the endpoint cannot directly distinguish
known causes. A machine-readable reason would support consistent diagnosis
without requiring callers to parse free-form text.

This proposal does not require exposing internal reasons to the client. Exact
reason names, fields, precedence, public responses, and reasons for other outcomes
remain separate questions. No JSON schema or exhaustive reason catalogue is
proposed here.
