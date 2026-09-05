# Authorization Handbook — working edition

This is the entry point for the handbook being developed through our discussions.
Chapters preserve the explanation, rationale, examples, counterexamples, and
remaining questions, not just short decision summaries.

This is an evolving working edition. Agreement on a concept does not finalize
its JSON schema, implementation, or every related branch.

Current checkpoint: scope is a boundary selector, scope requirements combine
with AND, and alternative authority comes through separate complete grants.
Authorization uses one endpoint-owned gate (CONTRACT-006), not the earlier
two-mode/prepared model. Exact key-value scope grammar remains proposed.

## Read the current chapters

| Material | What it preserves |
|---|---|
| [Scope and target](scope-model.md) | Canonical boundary-selector definition, AND within scope, alternatives through grants, and the still-proposed key-value syntax. |
| [Grants, assignments, and roles](grant-model.md) | Definitions, tenant context, permission/scope binding, groups, per-human self, role changes and expansion, dependency, positive grants, and administrative authority. |
| [Endpoint-owned authorization](endpoint-authorization.md) | Current single-gate model, selected endpoint inputs, authority and application facts, enforcement, and the deprecation map. |
| [Earlier authorization flow](authorization-flow.md) | Deprecated two-mode design, preserved with its rationale and examples. |
| [Grant JSON examples](grant-examples.md) | Six worked illustrations, including direct/group recipients, implicit tenant, multiple permissions, roles, expanded grant identity, and Employees-group self scope. |
| [Earlier endpoint-completion cases](endpoint-completion-cases.md) | Seven application-fact cases; their two-mode classification is deprecated and preserved. |

## Follow progress and decisions

- [Discussion tree](discussion-tree.md): all eleven stages, concluded questions,
  open siblings, active branch, and return points.
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
this discussion and has not yet been reconciled with the working edition. The
lab's interactive evaluator and enforcement trace also have not been updated
to implement our decisions. Treat their behavior and claims as source material
for later review, not as evidence that the new model is implemented.

The working chapters currently consolidate the grants and authorization-flow
branches and begin the scope-model discussion. Several other branches still have only partial discussion or decision
notes. We will develop those chapters as their questions are settled and review
coverage against the entire tree before declaring v1 complete.
