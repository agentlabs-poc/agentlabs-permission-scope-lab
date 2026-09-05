# Authorization Handbook — working edition

## Current position

The approved discussion through **Q-050-F** is the source of truth. Reader pages
and the [shared system diagram](system-overview.md) are locally reconciled to
that model. This is not Handbook v1 publication or an authorization implementation
migration. The user has now authorized committing and pushing this checkpoint
and continuing the handbook discussion. The prior commit/push freeze is lifted.

The [previous entry page](history/reconciliation-2026-09-05/docs/handbook.md.txt)
preserves the earlier sequential checkpoint summaries. Those earlier “next” and
“open” labels are historical, not competing current directions.

The reconciliation condensed this page's individual Q-041–Q-050-F checkpoints;
it did not revoke their decisions or remove the underlying chapters/log. The
visible decision trail below restores those references with current status,
rationale, examples, and links. Navigation cleanup must not hide agreed decisions.

## The agreed shape

- Permission identifies the operation; scope selects its boundary within the
  trusted implicit tenant. No additional canonical “target” entity is required.
- Scope is a required flat string-value object: AND within one scope, alternatives
  through complete grants, explicit `{}` for no narrower tenant-local restriction.
  Missing/null scope is invalid. Applications register supported key meanings.
- Auth owns human authorization groups/memberships. Team means group.
  Group grants are preferred, not exclusive; self resolves per authorizing human.
- Services/agents are human-dependent subsets, never independent group members.
  Resolved grants preserve their source bindings and dependencies.
- A protected method/route declares exactly one permission and selected inputs
  with exact sources. Every declared input is required; the application validates
  values. Published JSON/YAML contracts have top-level string `version: "1"`.
- One endpoint-owned gate combines shared authority evaluation and mandatory
  application enforcement. No prepared handoff or canonical relationship block.
- Authority within Finance does not by itself prove a requested certificate
  belongs to Finance. Actual output/effects must remain constrained to authorized
  boundaries and relevant request bindings.

These summaries do not replace the rationale, alternatives, examples,
counterexamples, and open questions in the chapters below (PROCESS-003).

## Decision trail — Q-041 through Q-050-F

These decisions remain part of the handbook. “Agreed” applies to the stated
rule, not every related schema or implementation. The
[chronological log](handbook-roadmap.md) retains the original questions,
alternatives, and approvals; the entries here reflect subsequent refinements.

### Q-041 — Explicit registration-validation mode, agreed

REGISTRATION-003: the application declares upfront whether permission–scope
support-relationship validation is enabled. If enabled, all grants—including
existing and role-based grants—must pass it. Missing metadata is not a way to
disable checks. If disabled, other canonical/registration/issuance checks and
runtime boundary enforcement still apply.

**Rationale:** one explicit application-level choice is predictable; inferring
the mode from absent metadata would permit ambiguity or bypass. This is
registration compatibility, not the endpoint relationship block later rejected.
Representation and detailed lifecycle mechanics remain open.
[Explanation and alternatives](application-registration.md).

### Q-042 — Existing grants before enabling validation, agreed

REGISTRATION-004: reject activation if existing grants are incompatible. Report
the incompatibilities and preserve the previous configuration until authorized
corrections are made. Do not silently delete, rewrite, or grandfather grants.

**Rationale:** enabling a security rule must not secretly alter existing authority
or retain exceptions to an all-grants rule. For example, an incompatible role
grant must be corrected before activation succeeds. Further registration
lifecycle mechanics are parked by the user's direction, not silently settled.
[Activation example and consequences](application-registration.md).

### Q-043 — Permission/boundary/request vocabulary, agreed as reformulated

TERM-005: no additional canonical “target” entity or wrapper is required.
Describe authorization using permissions, scope boundaries, requests, and
sufficient trusted request material, bound to the actual operation.

**Rationale:** the extra abstraction did not add a necessary distinction across
read/list/create/move/administrative examples. Removing it does not remove
relationship proof or enforcement: FIN in a path still does not prove C-17
belongs to Finance. The earlier administration framing was withdrawn; this
vocabulary decision did not finalize administrative scope encoding.
[Tenability checks and counterexamples](authorization-vocabulary.md).

### Q-044 — Five grant-administration rules, agreed at rule level

ADMIN-004/005: use the ordinary grant model; require the specific administrative
operation; check the complete requested assignment; keep its bounds associated;
and apply normal canonical/registration/tenant validation.

**Rationale:** granting access is not using access, and fields from unrelated
administrative grants cannot be mixed to manufacture broader issuance authority.
Finance payroll-read provisioning cannot donate its permission to Engineering
certificate-read provisioning. Self-assignment remains explicit, authorized, and
audited. Exact admin scope encoding, containment mechanics, and before/after
change contracts remain open.
[Five rules, Finance example, and counterexamples](grant-model.md#grant-administration--q-044-agreed-rules).

### Q-045 — Auth-owned human groups and membership, agreed

GROUP-001-A / GROUP-003: Auth owns authorization groups and memberships; team
and group mean the same thing. Groups contain humans, not first-class services
or agents. Applications may sync business groups into Auth or keep them separate.
Group-based human access remains preferred, not the only permitted route.

**Rationale:** one authorization-membership authority avoids treating application
business relationships as implicit grants. A proxy can receive only a delegated
subset through its human, preserving the dependency on that human's membership.
Nesting, synchronization timing, and detailed delegation encoding remain open.
[Membership examples and ownership rationale](groups-and-membership.md).

### Q-046 — Ordinary grants after issuer-authority loss, agreed

ADMIN-006: an ordinary human/group grant that was validly issued does not become
invalid solely because its issuing administrator later loses issuance authority.
Its own validity, conditions, revocation, and recipient dependencies still apply.

**Rationale:** routine administrator rotation should not unexpectedly revoke
organizational access. Withdrawing an issued grant requires explicit authorized
revocation. This differs from a service/agent's continuing dependence on its
authorizing human: the latter remains a live subset relationship. Invalid
issuance and incident-response mechanics are not settled by this rule.
[Lifecycle example and cascading alternative](grant-model.md#ordinary-grant-lifecycle--admin-006--q-046-agreed).

### Q-047 and Q-047-A — Request, resolved material, and declaration, agreed

RESOLUTION-005: the authorization request identifies the declared operation,
verified identity/tenant context, and selected inputs. A resolved request is its
evaluation view with sufficient trusted material for the relevant boundaries;
resolved does not mean allowed.

Q-047-A clarified that the endpoint predeclares its permission, inputs, and
sources. Later Q-049 fixes exactly one permission, and Q-050-C assigns required
relationship enforcement to implementation rather than a policy block.

**Rationale:** inputs identify the request but do not automatically become scope
boundaries or established facts. With `{}`, department adds no narrower scope
restriction, yet the fixed input contract and actual request binding remain.
No prepared handoff or mandatory eager fact lookup is reinstated.
[Declared-material and empty-scope examples](endpoint-authorization.md#endpoint-declaration-and-request-resolution--q-047--q-047-a-agreed).

### Q-048 — Dependent resolved grants and membership retrieval, agreed

RESOLUTION-006: obtain Vinay's valid memberships, then grants for Vinay directly
and for those groups. Resolved views establish relevant references/human context
while retaining source identity, associated permissions/scope/conditions/validity,
and membership/delegation dependencies.

**Rationale:** flattening to a permission list loses restrictions; copying a group
grant into an independent direct grant loses its membership dependency. Live role
expansion and self anchoring do not produce a new assignment or an allow.
Transport, cache/freshness, and concrete resolved-view schemas remain open; this
is a logical dependency flow, not a required API-call count.
[Resolved-grant example and counterexamples](grant-model.md#resolved-grants-and-membership-based-retrieval--resolution-006--q-048-agreed).

### Q-049 — Exactly one required permission per endpoint, agreed

CONTRACT-008: Auth mandates exactly one required permission for each protected
method/route. Missing, zero/multiple-permission, or invalid declarations cannot
permit execution. That permission must cover the complete protected operation.

**Rationale:** one fixed requirement keeps the endpoint contract simple and
reviewable. The earlier multiple-required-permission combination proposal was
not adopted. A grant may still contain multiple permissions, and a human may
have multiple direct/group grants; those are different concerns.
[Endpoint examples and preserved alternative](endpoint-authorization.md#one-required-permission-per-protected-endpoint--contract-008--q-049-agreed).

### Q-050 — Endpoint policy contract, partially settled; overall open

CONTRACT-009 requires versions in every published JSON/YAML contract. Q-050-A
through Q-050-F below settle the convention, partial policy shape, relationship
responsibility, review requirement, input presence, and value-validation ownership.

**Rationale:** agreeing requirements is not the same as publishing a complete
schema. The remaining structural validation, nested input syntax, further source
kinds, publication/compatibility details, and update/move semantics are still
open. Settled subquestions must not vanish or be repeatedly reopened because
their parent remains unfinished.
[Policy chapter](endpoint-policy-format.md) and [publication](contract-publication.md).

### Q-050-A — Shared contract version convention, agreed

CONTRACT-010: use a required top-level string `version`, initially `"1"`; quote
it in YAML. Reject missing, malformed, or unsupported versions without guessing
or fallback. Interpret the version within its contract type, not as a document
edit revision.

**Rationale:** explicit interpretation avoids ambiguity while allowing contract
types to evolve independently. Numeric `1` is not the adopted string value;
adding `"version": "1"` alone does not complete the remaining schema.
[Version examples and naming alternative](contract-publication.md).

### Q-050-B — Partial policy structure and GET/PUT bindings, agreed

CONTRACT-011: `version`, `method`, `path`, one `permission`, and named
`inputs` with explicit `source`/`name`. A local input name can differ from
its source field and is not automatically a scope key.

**Rationale:** explicit bindings distinguish requested material from established
facts without a shorthand selector grammar. In the approved PUT example,
`proposed_dept` comes from body field `department_id`; requesting FIN does
not prove the existing certificate belongs to Finance or authorize a move.
[GET/PUT JSON and field rationale](endpoint-policy-format.md#get-example-selected-path-parameters).

### Q-050-C — No relationship block; mandatory endpoint enforcement, agreed

CONTRACT-012: policy does not add `relationships`, named resolvers, or argument
maps. The endpoint implementation must establish or enforce relationships needed
to keep actual output/effects inside authorized boundaries.

**Rationale:** this consciously keeps the canonical contract small and places
application-specific enforcement with its owner. Finance-read authority does
not authorize an unchecked C-17 lookup. A tenant AND Finance AND C-17 query can
enforce the relationship directly. Auth alone cannot detect every defective
handler; enforcement is mandatory, not optional business validation.
[Grant/query example and rejected proposal](endpoint-policy-format.md#no-relationship-block--contract-012--q-050-c-agreed).

### Q-050-D — Review actual constraints, not merely input usage, agreed

ENFORCEMENT-003: review whether boundaries and request bindings constrain the
data returned or changed. Logging department, using an ineffective OR filter,
or returning unchecked data after a correct lookup does not meet this requirement.

**Rationale:** protection comes from the constraint on the actual operation,
not a variable appearing in code. `{}` does not invent department restrictions
just to use an input. This is central review guidance, not a guarantee that it
catches wrong permissions, stale dependencies, or every other breach.
[Review cases and limits](endpoint-policy-format.md#endpoint-review-requirement--enforcement-003--q-050-d-agreed).

### Q-050-E — Required inputs at their declared sources, agreed

INPUT-002: every declared input must be present at its exact source. Do not
silently omit/default it or fall back to another source, even when a tenant-wide
grant applies. Additional payload fields do not automatically become auth inputs.

**Rationale:** the endpoint has a fixed input contract independent of grant
breadth. Omitting PUT's body `department_id` is not repaired by a query
parameter or broader authority. Presence alone does not establish validity;
Q-050-F separately settles who validates the value.
[Presence, omission, and fallback cases](endpoint-policy-format.md#required-inputs-at-their-declared-sources--input-002--q-050-e-agreed).

### Q-050-F — Application-owned input value validation, agreed

INPUT-003: the application's request contract defines and validates types,
nullability, format, and domain meaning. Do not duplicate these as type/nullable/
validation-expression fields in endpoint policy. Authorization and execution
must use the same validated meaning after parsing or normalization.

**Rationale:** two input schemas can disagree. Required presence is distinct
from valid value: a required field may accept null only when its application
contract explicitly allows it, and null never means unrestricted authority.
Auth retains its own canonical checks. This confirms the existing policy shape;
it is not a new schema or an invitation to seek the same approval again.
[Value cases, counterexamples, and responsibility split](endpoint-policy-format.md#application-owned-value-validation--input-003--q-050-f-agreed).

## Read the chapters

| Chapter | Purpose |
|---|---|
| [System overview and SVG](system-overview.md) | Current responsibility layers, single gate, inputs, complete authority, and actual-use enforcement. |
| [Vocabulary](authorization-vocabulary.md) | Permission/boundary/request-material terms, tenability checks, and the retired vocabulary's rationale. |
| [Permission](permission-model.md) | Restored detailed operation/reach explanation, namespaced naming examples, rationale, and explicitly open grammar rules. |
| [Scope boundaries](scope-model.md) | Canonical scope semantics/format, self, AND/alternatives, empty/invalid cases, and earlier proposals. |
| [Grants and roles](grant-model.md) | Complete bindings, groups, live roles, dependent resolution, administration, and lifecycle distinctions. |
| [Grant formats](grant-format.md) | Working current-scope examples and deprecated layout correspondence; complete schemas remain open. |
| [Groups and membership](groups-and-membership.md) | Auth ownership, human-only membership, optional application sync, and preferred group access. |
| [Application registration](application-registration.md) | Abstract validation, application meanings, optional support checks, and activation rules. |
| [Endpoint authorization](endpoint-authorization.md) | One endpoint-owned gate, request versus resolved material, and retained historical decisions. |
| [Endpoint policy format](endpoint-policy-format.md) | Approved version/method/path/permission/inputs shape; GET/PUT, value validation, and mandatory enforcement. |
| [Operation-specific enforcement](operation-enforcement.md) | Impact-first discussion of boundary-changing writes, with source/destination examples and open concurrency/composition details. |
| [Authority freshness](authority-freshness.md) | Impact-first revocation timing discussion, stale-cache trade-offs, and in-flight-operation questions kept distinct. |
| [Contract publication](contract-publication.md) | Version convention, rejection behavior, and still-open schema publication work. |
| [Cross-domain use cases](use-case-examples.md) | Sixteen scenario groups across Git, ticketing, HRMS, and accounting; illustrative, not runtime tests. |
| [Explanation audit](explanation-audit.md) | Coverage evidence and explanation gaps. |

## Follow the whole discussion

Q-051 / DECISION-003 is **agreed**: completed allow/deny decisions are separate
from evaluation errors. The approved Auth-service timeout example and rationale
are in [decision results](decision-results.md). Existing fail-closed and
single-gate rules remain unchanged. Q-052 / DECISION-004 is also **agreed**:
every completed deny must include an internal machine-readable reason. The
chapter records the rationale and consequences. Q-053 / DECISION-005 settles
evaluator-provided `error_message` and `error_message_reason`, both delivered to
the UI for presentation. Q-054 / DECISION-006 uses these same fields for evaluation
errors without conflating errors with completed denials. Q-055 / DECISION-007
approves a stable `error_code` alongside the two readable messages. Full schemas,
reason-code definitions, and safe-content rules remain open. The permission
explanation has been restored at the user's request; Q-056 adopts its naming
convention with variable-depth application namespaces. Q-057 agrees that name
prefixes do not automatically confer deeper permissions. Q-058 excludes wildcard
permission names from v1. Q-059 excludes permission aliases from v1. Character
validation and catalog evolution remain open. Q-060 approves supporting-grant
references in allow results, available for audit without requiring every request
to be logged. Q-061's proposed returned-boundary fields are not required;
endpoint enforcement remains unchanged. Q-062 approves a minimal allow JSON
shape with version, decision, and grant_ids. Q-063 approves the deny variant
using version, decision, and the agreed error code/two message fields. Q-064
approves an evaluation-error variant with no decision field. Q-065 requires
rejecting mixed variant-specific fields. Q-066 requires a non-empty supporting-grant
ID list for allow. Q-067 requires rejecting unknown result fields. Further format
details are parked under the user's impact-first horizontal instruction, not
excluded. Q-068 requires move authority over both current and proposed boundaries;
the operation chapter records the approved detailed example and rationale.
Q-069 now proposes revocation freshness for new checks; no stale-cache grace
period is recommended, but that timing contract is not yet agreed.
The chapter preserves rationale and distinguishes available evidence from
persistent audit.

The [discussion tree](discussion-tree.md) is the mind map: settled principles,
unfinished branches, and return points. The [decision log](handbook-roadmap.md)
retains stable references and agreement status.

The user requested horizontal progress on high-impact decisions. The next
selection should come from decision results, enforcement across operation types,
or freshness/dependency behavior—not another approval of the settled endpoint
policy shape. The exact priority order is still a proposal, not an adopted rule.

Full contracts, update/move/list/bulk behavior, administrative scope encoding,
delegation mechanics, revocation/concurrency, audit, conformance testing, and
publication/governance remain unfinished. The subsequent
[MEASURE-001 audit](handbook-completion-audit.md) now counts 35 of 69 checkpoints
closed (50.7%), up from 34/69 after Q-059 closes HC-04-04; this supersedes the
earlier absence of a measured baseline, not
the distinction between documented agreement and implementation readiness.

## Preservation and implementation status

The [reconciliation register](reconciliation.md) distinguishes current text,
retained history, and implementation gaps. The refreshed
[concept](../src/content/authorization-concept.md),
[HRMS](../src/content/hrms-tenant-setup.md), and
[projects](../src/content/projects-repositories-teams.md) pages summarize the
approved model and link the detailed chapters.

Original prose and diagrams remain in the
[reconciliation archive](history/reconciliation-2026-09-05/README.md).
All guided explorers and the standalone enforcement trace have been removed
from the active site by user approval. The homepage is now the handbook;
source is preserved for the [HRMS explorer](history/retired-hrms-explorer-2026-09-05/README.md)
and [earlier retired pages](history/retired-pages-2026-09-05/README.md).
Their algorithms and external implementation claims have not been migrated or
reverified. The Projects handbook chapter is retained.

Earlier [two-mode flow](authorization-flow.md),
[endpoint-completion cases](endpoint-completion-cases.md), and
[grant layouts](grant-examples.md) retain rationale with deprecation labels.
Historical options are not live alternatives unless deliberately reopened.

## Working discipline

Record meaning, rationale, alternatives, examples/counterexamples, consequences,
and open dependencies. Reconcile chapters, log, tree, examples, and diagrams
periodically. Preserve superseded material instead of deleting it.

PROCESS-004's checkpoint commit/push practice was suspended during the user's
explicit review freeze. The user has now reopened the gate with “lets commit
and push, and continue with the rest.” Verification and preservation remain
required; later user instructions can change the gate again.
