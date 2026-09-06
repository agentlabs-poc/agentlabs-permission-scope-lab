# Handbook working rules

## Commit gate — reopened by the user

The user explicitly instructed: “lets commit and push, and continue with the
rest.” The checkpoint is authorized for commit and push after verification.
Continue the agreed handbook discussion and checkpoint workflow, respecting any
later user restriction. Do not force-push or treat this as authority to change
unrelated repositories or systems.

Current cadence (user instruction after Q-083): record each decision and its
rationale locally immediately, but commit and push roughly once per ten
questions, or when explicitly requested. Do not commit after every answer.
Q-083 is the first approved question recorded after checkpoint `01d68a5`.
The user explicitly requested a checkpoint after approving Q-087-B, including
its rationale for deliberate `sub`/`human_id` duplication. This request is within
the explicit-checkpoint exception; roughly ten questions remains the default.

Latest gate instruction: after Q-089 was left undecided, the user froze recording
and commits. The subsequent “record this commit and push” explicitly authorizes
recording Q-089-B's revision/adoption decision, reconciling affected current docs,
and committing/pushing this checkpoint. Do not infer unrestricted recording of
later discussion from this scoped reopening; await further user direction.

Previous gate state, retained as history: the user had frozen commits and pushes
while reviewing local reconciliation. Freeze notices in archived snapshots and
completed plans describe that earlier state, not the current instruction.

## Source of truth and preservation

The user's approved discussion decisions take precedence over older lab prose,
`src/content`, interactive simulations, and diagrams. Do not turn an open
proposal into an agreed rule during reconciliation.

Preserve superseded material with clear deprecation labels or linked archives.
Record rationale, examples/counterexamples, consequences, and open questions,
not just conclusions. Keep reference numbers on discussion questions/proposals.

When proposing a contract or format, include its exact versioned JSON/YAML block
in the same discussion message as its meaning, rationale, and approval question.
The user requested this after Q-087-A so semantics and representation can be
reviewed together. Label excerpts versus complete contracts explicitly.

## Discussion pacing — impact-first horizontal coverage

The user explicitly asked for horizontal coverage based on impact. Ask one
question at a time and continue after each answer without waiting for “proceed.”
Prioritize unresolved authority and enforcement decisions across branches before
extended field-by-field validation. Keep lower-level gaps visible for a later
contract-completion pass; parking them does not approve or exclude them. Do not
mark a whole branch complete after settling only its governing principle.

Keep architectural context visible: explain where each substantive proposal fits,
what is already agreed, what would change, and its rationale and trade-offs.
Progress does not remove the need for proposals or architecture discussion.
Update architectural prose/diagrams when approved decisions change the flow or
responsibility boundaries; do not redraw them for every field-level decision.

## Handbook scope — audit belongs to another layer

Q-076 explicitly excludes audit-system policy and design from this handbook.
Do not treat event selection, audit schemas, delivery/failure guarantees, storage,
retention, or disclosure as remaining handbook deliverables. Preserve the rejected
proposal as history. Authorization-result evidence and explicit authorization of
administrative changes remain in scope; earlier audit references are integration
context, not authority to design the external audit layer here.

## Handbook scope — authorization, not business logic

The user's Q-084 correction prohibits expanding this handbook into business-rule
design. Layer 2 means application-specific authorization integration: establishing
trusted authorization material and enforcing the evaluated boundaries on the
actual operation. It is not a business-logic layer within this handbook.
Q-084 is explicitly DISAPPROVED: record the intent to prevent business-logic
scope creep, not merely a wording correction. The assistant's revised question
is withdrawn, not pending. Do not infer approval either to remove generic
authorization conditions or to introduce a condition engine. Preserve rejected
proposals and rationale without automatically re-proposing them in narrower form.

## Role references — Q-089-B supersedes live updates

Published role revisions are immutable. Role-based grants explicitly select
`role_revision` alongside `role_id`; publication does not silently update grants.
Adoption is an authorized grant change with boundary validation. Resolve the
adopted revision, not latest. Preserve ROLE-002/live-role text as deprecated history.
Auth's own protected APIs must follow the same authorization framework. Publishing
a narrower revision does not withdraw permissions from older adopted revisions.
