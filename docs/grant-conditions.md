# Q-084 — DISAPPROVED: preserve the authorization boundary

## Decision and rationale

**Status: DISAPPROVED.** The user explicitly instructed that Q-084 be recorded
as disapproved, with rationale capturing the intent behind the rejection.

**Intent:** this is an authorization handbook. Its model and responsibilities
must not expand into business logic, directly or by placing business-rule
enforcement inside its description of Layer 2. The original proposal blurred
that boundary by using application business rules to frame a canonical grant
decision. Calling those rules application-owned does not justify bringing their
design into the authorization architecture.

Application-specific authorization integration remains in scope: establishing
trusted authorization material and ensuring the actual operation respects the
evaluated boundaries. That is not permission to define business workflows.

**Consequences:** the original proposal is rejected, not approved with a wording
change. The assistant's subsequent narrowed question is withdrawn rather than
left awaiting approval. This rejection neither adopts a generic condition engine
nor approves removing existing conceptual grant restrictions. Any genuine gap
in authorization conditions remains an open model issue, not an unanswered Q-084.
The existing authorization request flow is unchanged.

<details>
<summary>Intermediate scope-correction record and revised question — superseded; revised question withdrawn</summary>

## Recorded scope correction

The user stated: “We are crossing the boundry of auth here. we cannot creep in
to leak in to business logic.” This is recorded as a scope constraint, **not** as
approval of the original proposal to omit a generic conditions field.

Both layers discussed in this handbook concern authorization. Layer 1 supplies
canonical authority. Layer 2 supplies application-specific authorization
integration: trusted material and enforcement of evaluated boundaries. Business
workflow rules are outside this handbook; it does not prescribe their design,
evaluation, or outcomes.

For example, establishing that certificate C-17 belongs to Finance is necessary
to enforce a Finance-scoped grant. It remains in scope even though the application
owns that fact. Defining the certificate's business workflow is outside scope.
Rationale: application ownership of data does not make every use of that data
business logic, and authorization integration must not grow into a business-rule
engine. The approved endpoint-owned authorization gate remains unchanged.

## Q-084 revised — PROPOSED, not approved

Recommend keeping v1 grant restrictions to the agreed permission/scope, status,
validity, and membership/delegation dependency model, without adding a generic
`conditions` field. The alternative is an additional mechanism specifically for
authorization restrictions, which would need its own justified use case and
evaluation/subset semantics. Business rules are not a reason to add that mechanism.

**Q-084 revised:** Should v1 use the agreed authorization restrictions without
adding a generic grant `conditions` field?

Existing conceptual conditions and historical illustrations are not removed or
silently ignored. HC-08-03 remains open until this narrower question is settled
and any approved reconciliation is performed. No completeness credit is claimed
for excluding business logic, which was not a canonical authorization deliverable.

</details>

<details>
<summary>Original Q-084 proposal — DISAPPROVED; retained as history</summary>

# Q-084 — does v1 need generic grant conditions?

Status: **PROPOSED, not approved.** This is a Layer 1 / Layer 2 responsibility
decision, not another field-validation question.

## Existing foundation

Layer 1 supplies canonical authority: permission, scope, grant eligibility,
membership, and human-dependent delegation. The embedded authorization agent
works across that authority and the trusted application material at the
endpoint-owned gate. Layer 2 establishes application facts, binds the actual
operation to evaluated material, and enforces application business rules.
Neither layer may use an application rule to broaden granted authority.

Q-083 now approves optional finite grant validity windows, independently of
enabled/disabled status. Existing conceptual grant definitions also mention
“conditions,” and some older illustrative JSON contains `conditions: []`.
Those references do not provide a complete condition language or wire contract.
They remain intact while this proposal is discussed; no previously applicable
restriction is silently removed or treated as satisfied.

## Recommendation and alternatives

Recommend **no generic `conditions` field or expression language in canonical v1
grants**. Keep the agreed scope, validity, status, and dependency checks. Keep
mandatory application checks in Layer 2. Registered application facts can still
participate in the agreed scope mechanism; this proposal does not restrict scope
to a fixed list of built-in application concepts.

The alternative is to define an additional grant-condition mechanism now. That
could express restrictions outside the current model, but would require agreement
on vocabulary, registration, evidence sources, evaluation failures, and how
administrative and delegated authority remain subsets under those restrictions.
An unspecified extension block would defer those questions without providing an
interoperable contract, so it is not recommended as a shortcut.

## Concrete example

Suppose an application registers an invoice-posting permission and a department
scope key. An enabled grant supplies that permission within Finance during its
validity window. The endpoint establishes and enforces the actual invoice's
Finance boundary. The application also requires that the invoice is still a
draft before posting it.

Under this proposal, permission/scope/validity/dependency evaluation answers the
authority question. The draft-state precondition remains an application rule;
it does not require a generic grant-condition expression. An authorization allow
does not promise business-operation success or excuse skipping that rule.

Counterexample: if the product needs a grant-specific restriction such as access
only during recurring working hours, a finite validity window is not equivalent.
We must discuss whether v1 needs that capability; we cannot silently ignore it,
pretend the current scope format expresses it, or move it into an unenforced note.

## Architecture and recording consequences

If approved, the client → authenticated context → endpoint-owned authorization
gate → constrained application operation flow stays unchanged. The decision
clarifies responsibility and the grant contract; it does not add another gate,
a prepared state, or an independent application grant system.

Approval would require marking conflicting older `conditions` illustrations as
superseded for v1 while retaining their history and their core whole-binding
principle. That reconciliation has **not** been performed before approval.
HC-08-03 remains open while this proposal is unanswered.

**Q-084:** Should canonical v1 omit a generic grant-condition mechanism, retaining
the agreed authority checks and mandatory Layer 2 application rules?

</details>
