# Explanation coverage audit — through Q-043

Q-050-F follow-up: INPUT-003 records application-owned value validation with no
new policy fields. The policy chapter preserves duplicate-schema rationale,
five department_id cases, required-versus-nullable distinction, same validated
meaning for authorization/execution, and continued canonical checks. User's
reminder that the shape was already approved is recorded; it is not reopened.

Q-050-E follow-up: INPUT-002 is approved. The policy chapter captures the fixed
input-contract rationale, optional-input alternative not selected, five PUT
presence/source cases including {}, and the limits for types/nullability/errors.
No new policy fields, silent defaults, or complete-validation claim are introduced.

Q-050-D follow-up: ENFORCEMENT-003 is now approved. The policy chapter records
the exact review wording, its rationale, Finance conjunction example, five
review cases, the {} qualification, and limits against a blanket security
guarantee. Earlier proposed-status coverage below is preserved history.

Q-050-C follow-up: the policy chapter records mandatory endpoint boundary
enforcement without a relationship block, the simplicity rationale and conscious
trust tradeoff, a grant/GET/constrained-query example, empty/self/PUT distinctions,
and the original resolver proposal as not adopted. Q-050-D records the user's
review suggestion and the proposed constraint-enforcement refinement, including
why input usage is not a guarantee against all breaches. That refinement is open.

Q-050-B follow-up: endpoint-policy-format.md records the approved partial fields,
each field's rationale, explicit source/name versus shorthand, GET and PUT JSON,
body/local-name mapping, the Engineering-to-Finance counterexample, and open
relationship/update/move semantics. The user requested the PUT example; its body
is clearly a proposed value, not current-state proof or a new published contract.

Q-050-A follow-up: contract-publication.md records the approved version syntax,
string/type-local interpretation, distinction from document revision, rationale,
alternative field name, five validation cases, and remaining schema/compatibility
work. The endpoint policy contract remains explicitly open; a version metadata
fragment is not presented as a complete published policy.

Publication follow-up: contract-publication.md records CONTRACT-009's version
requirement, rationale, working-example status, and unselected version syntax/
compatibility rules. Endpoint policy requirements are explicitly distinguished
from its undiscussed JSON/YAML contract (Q-050). No example has been silently
promoted to a complete versioned contract.

Q-049 follow-up: the endpoint chapter records exactly-one-permission validation,
the simplicity rationale, complete-operation design obligation, read/download/
revoke mapping, no implicit permission hierarchy, preserved multi-permission
grants, and the original AND proposal explicitly not adopted. Earlier plural
endpoint wording and open-combination notes are qualified rather than removed.
Validation timing/schema and multi-object enforcement remain open.

Q-048 follow-up: the grant chapter records the resolved-grant definition and
Vinay's membership/direct/group retrieval flow, the G-17 Finance/self example,
what resolution establishes versus what evaluation must still check, rationale
against flattening/independent assignment, four counterexamples, and open retrieval,
cache/freshness, schema, and decision details. No whole-handbook completion claim.

Q-047 / Q-047-A follow-up: the endpoint chapter records the approved declaration
wording, request/resolved-request/decision distinction, and fixed declaration
versus scope-dependent material needs. Three scope cases include `{}`, with
the request-identification rationale, Finance/Engineering counterexample, tenant
and application-contract safeguards, and limits for conditions/delegation.
No wire schema, multi-permission combination rule, or eager-fetch rule is implied.

Q-046 follow-up: the grant chapter records ADMIN-006, the distinction between
issuance provenance and continuing authority dependency, the cascading alternative
not adopted, Maya/Vinay's example, four lifecycle outcomes, explicit-revocation
consequences, and limits for invalid issuance and incident-response mechanics.
Human-dependent automation and group-membership dependencies remain explicit.
This is a focused coverage update, not a declaration that the handbook is complete.

Q-045 follow-up: [groups and membership](groups-and-membership.md) records the
two approved policies, reasons for central membership authority and human-only
groups, the alternative not selected, five example/counterexample cases,
membership/automation dependencies, and remaining sync/freshness details.
The user explicitly reaffirmed PROCESS-003's rationale requirement.

Q-044 follow-up: the grant chapter now records all five approved administration
rules, the Finance grant JSON, five expected outcomes, separation from personal
access, and the still-open scope encoding/containment. Earlier statements below
that administration rules are unapproved are historical to the Q-043 audit.

The user asked whether the handbook captures sufficient explanation, then
explicitly asked for a check against the actual text. This audit covers the
recent discussions and the core chapters they depend on. It is not a claim
that every branch of the handbook is complete or that the lab implements it.

## What was checked

For each topic: agreed meaning, reason for the distinction, example or
counterexample, safety consequence, and explicit remaining questions. The
decision log alone is not sufficient; substantive explanations belong in the
chapters and the tree records the return point.

| Discussion | Detailed explanation checked | Coverage and limits |
|---|---|---|
| Grant/assignment, direct/group grants, roles and expansion | [Grant chapter](grant-model.md), [grant formats](grant-format.md) | Binding semantics, source dependencies, human-relative self, role expansion, examples, and distinction from independent authority are explained. Lifecycle/transport mechanics remain open. |
| SCOPE-006/007/008 | [Scope boundaries](scope-model.md) | Boundary definition, canonical JSON, AND versus separate grants, explicit empty scope versus missing/null, rejection rules, and examples are captured. Exact containment and temporal meanings are unfinished. |
| CONTRACT-006/007 | [Endpoint authorization](endpoint-authorization.md) | One gate, source declarations, request inputs versus trusted facts, bounded fact gathering, enforcement, and deprecated alternatives are explained. Complete wire schemas and concurrency mechanics remain open. |
| Q-036/037 | [System overview](system-overview.md) | Both responsibility layers, embedded integration, a payslip walkthrough, and the distinction from two decision locations are present. Deployment/API details are not implied. |
| Q-038 | [Scope boundaries](scope-model.md) | Application-owned boundary meanings, explicit supported relationships, endpoint bindings, department examples, and why matching names is insufficient are captured. No universal department catalog is adopted. |
| Q-039/040 | [Application registration](application-registration.md) | Registration, domain-agnostic validation, separate issuance authority, a grant fragment, counterexamples, logical registration flow, and optional relationships are explained. Exact schema remains open. |
| Q-041/042 | [Application registration](application-registration.md) | Explicit application mode, mandatory checks for all grants when enabled, continuing runtime enforcement, existing-grant activation rejection, and preserved prior configuration are captured. Further lifecycle detail is parked by user direction. |
| Q-043 | [Authorization vocabulary](authorization-vocabulary.md) | Canonical wording, why scope alone does not describe a request, five operation cases, P-17/P-18 and caller-department counterexamples, and the limits of the tenability check are documented. Administration bounds are not silently approved. |
| Cross-domain challenge cases | [Worked examples](use-case-examples.md) | Sixteen scenario groups include grants, endpoint material, assumptions, and expected outcomes across Git hosting, ticketing, HRMS, and accounting. These are not production-engine tests. |

## Findings

The checked chapters contain substantive rationale and examples, not just a
list of approvals. Q-043 now has a dedicated explanation instead of relying on
word substitution throughout the handbook. A reader can recover the reason for
the vocabulary decision and the safety distinction it must preserve.

The check also exposed status and wording drift: older passages continued to
describe Q-043 as open, and earlier scope phrasing used the retired abstraction.
The current log, chapter introductions, and tree now distinguish the agreed
vocabulary correction from the unfinished administration model. Exact displaced
wording is retained in [deprecated history](history/q043-vocabulary.md), and
the prior SVG is preserved separately. Existing history is not new approval.

## What this does not claim

The remaining glossary, administrative scope encoding/containment, concrete
request and resolved-data contracts, collection/mutation semantics, and
freshness/audit mechanisms still need discussion. Some older working proposals
need deliberate consolidation; explanation quality is not permission to mark
them agreed. No handbook-wide completion percentage is asserted.

This is a focused editorial and semantic coverage check. Automated checks can
verify links, JSON, reference IDs, and preservation, but cannot alone prove
that every explanation is clear or that an implementation is secure. The
handbook remains open to the user's requests for deeper examples or correction.
