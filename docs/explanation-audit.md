# Explanation coverage audit — through Q-043

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
