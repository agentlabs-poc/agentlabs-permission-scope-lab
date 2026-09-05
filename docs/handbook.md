# Authorization Handbook — working edition

Current: Q-050-E approves INPUT-002: every declared input is required at its
specified source, without silent omission/default/fallback. `{}` does not waive
input requirements or invent narrower scope restrictions. [PUT cases and rationale](endpoint-policy-format.md)
distinguish input presence from value validity. Next: input value validation and
remaining policy validation/publication. Earlier missing-input-open notes are
retained history where this approval now settles presence/source behavior.

Current: Q-050-D approves ENFORCEMENT-003. Endpoint review must verify that
boundaries and request bindings actually constrain returned/changed data, not
merely that inputs appear in code. [Rationale and counterexamples](endpoint-policy-format.md)
include ineffective filters and `{}`; this is not a guarantee against every
breach. Q-050's remaining input/policy validation and publication are next.
Older Q-050-D-open notes below are preserved history.

Current: Q-050-C approves CONTRACT-012: no relationship block; endpoint
implementation must establish or enforce actual execution within authorized
boundaries. [Policy rationale and examples](endpoint-policy-format.md) preserve
the conscious tradeoff and earlier proposal not adopted. Q-050-D is an open
review refinement: constraints must restrict output/effects, not merely appear
in input usage. Full policy validation/publication and update/move contracts
remain open. Older relationship-binding-open notes below are history.

Current: Q-050-B approves CONTRACT-011's [partial endpoint policy format](endpoint-policy-format.md),
including explicit path/body input bindings. GET and PUT examples preserve the
rationale and proposed-versus-current-fact distinction. Q-050 remains unfinished;
next is Q-050-C, relationship bindings. Earlier policy-structure-open notes below
are checkpoint history; no complete policy schema has been published.

Current: Q-050-A approves the [shared version convention](contract-publication.md):
required top-level `version: "1"` as a string; missing, malformed, or unsupported
versions are rejected. CONTRACT-010 distinguishes contract version from document
revision. Q-050's full endpoint policy schema remains open. Older statements
that version syntax is unselected are history, not the current convention.

Publication requirement: **every published JSON/YAML contract includes a version**
(CONTRACT-009). Existing JSON examples remain working illustrations until full
versioned contracts are settled. [Contract publication](contract-publication.md)
records rationale and open details. The endpoint policy contract has not yet
been discussed; Q-050 is next, before returning to decision results/enforcement.

Current: Q-049 approves CONTRACT-008. Each protected endpoint predeclares
**one required permission**, inputs, sources, and how to establish any required
relationship. Auth rejects zero/multiple-permission or missing/invalid declarations;
they cannot permit execution. [Endpoint authorization](endpoint-authorization.md)
records rationale, examples, design obligations, and the earlier proposal not
adopted. Multiple permissions per grant remain supported. Earlier plural/open
endpoint-combination notes are historical, not unfinished v1 choices.
Next branch: decision results and enforcement contracts; exact schemas remain open.

Current: Q-048 approves RESOLUTION-006, the dependent resolved-grant view and
membership-based retrieval flow. Obtain Vinay's valid memberships, then grants
for Vinay directly and for those groups, preserving every source binding and
dependency. The [grant chapter](grant-model.md) records rationale, JSON example,
counterexamples, and open mechanics. Next branch: decision semantics, including
multiple required permissions; no additional combination policy is approved yet.
Earlier checkpoint summaries below are retained history.

Current: Q-047 and Q-047-A agree request/resolved-request meanings and clarify
CONTRACT-007: the endpoint predeclares permissions, inputs, sources, and how to
establish any required relationship. [Endpoint authorization](endpoint-authorization.md)
explains fixed declarations, scope-dependent material needs, and the `{}` case.
Resolved does not mean allowed. Next branch: resolved grants (Q-048, open).
Earlier checkpoint summaries below are retained history.

Current: Q-046 approves ordinary grant lifecycle independence from the issuing
administrator's later loss of issuance authority (ADMIN-006). The
[grant chapter](grant-model.md#ordinary-grant-lifecycle--admin-006--q-046-agreed)
records rationale, the cascading alternative not adopted, examples, and explicit
revocation consequences. Human-dependent automation remains unchanged.
Next: Q-047, request versus resolved request at the one endpoint-owned gate.
Administrative encoding and detailed lifecycle mechanics remain open.
Checkpoint summaries below are retained history, not additional live positions.

Q-045 settles Auth-owned authorization groups/memberships, optional application
sync, and human-only groups. [Groups and membership](groups-and-membership.md)
records the rationale, rejected alternative, examples, and remaining mechanics.
Q-046 is next in authority-model closure; older group-open notes are history.

Current: Q-044 approves the five [grant-administration rules](grant-model.md).
ADMIN-004/005 are settled at rule level; exact encoding remains open. Q-045
next consolidates group ownership/membership directions. Older administration
open-status notes below describe earlier checkpoints.

Q-043 now agrees TERM-005: use permissions, scope boundaries, requests, and
trusted request material without an additional canonical entity. The detailed
[vocabulary explanation](authorization-vocabulary.md) preserves the reasoning,
case checks, security counterexamples, and remaining gaps. Administration
remains open; the earlier Q-043 framing is withdrawn.

Current discussion: stage 5, grant administration, Q-043 under ADMIN-004/005.
Q-042 is agreed; further registration lifecycle details are parked by user
direction. Earlier checkpoint summaries below are retained context.

This is the entry point for the handbook being developed through our discussions.
Chapters preserve the explanation, rationale, examples, counterexamples, and
remaining questions, not just short decision summaries.

This is an evolving working edition. Agreement on a concept does not finalize
its JSON schema, implementation, or every related branch.

Current checkpoint: scope is a boundary selector, scope requirements combine
with AND, and alternative authority comes through separate complete grants.
Authorization uses one endpoint-owned gate (CONTRACT-006), not the earlier
two-mode/prepared model. SCOPE-007's flat string-value scope is canonical:
explicit {} means tenant-wide reach; missing/null scope is invalid.

Q-039 adds REGISTRATION-001: applications register supported permissions and
scope contracts; Auth validates grant acceptance without interpreting domain
meaning. Registration format and permission-scope compatibility remain open.

Q-040 subsequently approves optional permission-scope support-relationship
declarations through registration (REGISTRATION-002). Representation and
omission behavior (Q-041) remain open; this qualifies the earlier checkpoint.

Q-041 then settles REGISTRATION-003: applications explicitly enable or disable
relationship validation upfront. Enabled applies to all grants, including
existing and role-based grants; disabled does not weaken runtime enforcement.
This supersedes omission-based mode proposals. Change/revalidation mechanics
remain open, beginning with Q-042.

## Read the current chapters

| Material | What it preserves |
|---|---|
| [Groups and membership](groups-and-membership.md) | Ownership and human-only membership decisions, their rationale, optional-sync distinction, examples, and dependent automated access. |
| [Explanation coverage audit](explanation-audit.md) | Evidence of which recent explanations were checked, where reasoning/examples live, and what remains incomplete. |
| [Authorization vocabulary and request material](authorization-vocabulary.md) | Q-043's approved explanation, why it is tenable across operations, and the evidence-to-execution requirement. |
| [Application registration](application-registration.md) | Agreed registration/validation responsibilities, permission and scope examples, and the boundary between canonical checks and application interpretation. |
| [Cross-domain use cases](use-case-examples.md) | Sixteen worked scenario groups across Git hosting, ticketing, HRMS, and accounting, with grants, endpoint material, and expected outcomes. |
| [Reconciliation register](reconciliation.md) | Current-versus-historical status, retained deprecations, and unresolved implementation/design gaps. |
| [System block diagram](system-overview.md) | Scalable SVG and text views of the agreed endpoint-owned gate, explicit source declarations, shared evaluator, and enforcement. |
| [Current grant formats](grant-format.md) | Current direct/group, role-reference, and expanded-view examples using canonical scope, with a map to deprecated layouts. |
| [Scope boundaries](scope-model.md) | Canonical boundary-selector definition and v1 key-value format, AND within scope, alternatives through grants, and empty/invalid scope rules. |
| [Grants, assignments, and roles](grant-model.md) | Definitions, tenant context, permission/scope binding, groups, per-human self, role changes and expansion, dependency, positive grants, and administrative authority. |
| [Endpoint-owned authorization](endpoint-authorization.md) | Current single-gate model, selected endpoint inputs, authority and application facts, enforcement, and the deprecation map. |
| [Earlier authorization flow](authorization-flow.md) | Deprecated two-mode design, preserved with its rationale and examples. |
| [Grant JSON examples](grant-examples.md) | GRANT-EX-007 uses canonical v1 scope; six earlier examples preserve historical syntax and grant/role/dependency explanations. |
| [Earlier endpoint-completion cases](endpoint-completion-cases.md) | Seven application-fact cases; their two-mode classification is deprecated and preserved. |

## Follow progress and decisions

- [Discussion tree and mind map](discussion-tree.md): a compact whole-handbook
  overview followed by all eleven stages, concluded questions, open siblings,
  the active branch, and explicit return points.
- [Roadmap and decision log](handbook-roadmap.md): stable proposal/question IDs,
  agreement status, user decisions, and historical alternatives.

PROCESS-003 requires definitions, rationale, examples/counterexamples,
consequences, and unresolved details to remain reconstructable. We update the
chapters when a decision changes, keeping superseded alternatives in the log.

## Checkpoint and reconciliation practice

PROCESS-004: commit and push meaningful documentation checkpoints as the
discussion progresses. At branch closures and periodically during longer
branches, review the following together:

- Chapters and examples agree with the decision log, including terminology.
- The discussion tree retains all open questions, dependencies, and return points.
- Explanations preserve rationale and counterexamples, not just conclusions.
- Contradictions with the original lab handbook are reconciled where decisions
  are settled; unresolved choices and implementation gaps remain explicit.

Keep superseded decisions as history. Reconciliation does not authorize silently
deciding open questions or changing application behavior.

PROCESS-006 explicitly resumes recording after the discussion-only pause and
requires earlier designs to be retained with deprecation labels.

## Relationship to the original lab handbook

The original [lab concept page](../src/content/authorization-concept.md) predates
this discussion ~~and has not yet been reconciled with the working edition.~~ The
lab's interactive evaluator and enforcement trace also have not been updated
to implement our decisions. Treat their behavior and claims as source material
for later review, not as evidence that the new model is implemented.

Reconciliation checkpoint: original lab pages and the older design now carry
historical/deprecation notices, with the substantive prose preserved. Current
chapters and examples are aligned to approved scope and endpoint decisions.
This is not a full rewrite or an implementation migration; see the
[reconciliation register](reconciliation.md) for resolved drift and open gaps.

The working chapters currently consolidate the grants and authorization-flow
branches and begin the scope-model discussion. Several other branches still have only partial discussion or decision
notes. We will develop those chapters as their questions are settled and review
coverage against the entire tree before declaring v1 complete.
