# Q-086 — shared request-identity representation

**Q-127 approved:** [v1 supports direct human-to-proxy delegation, not proxy chains](delegation-lifecycle.md).
This narrows the eligible supporting relationship without changing the identity
block below. Matching actor/human IDs still require trusted delegation evidence.

Status: **AGREED.** The user answered Q-086 “approve.”
Previous status, retained as history: **PROPOSED, not approved** until that answer.
Q-085 approves trusted proxy/human attribution;
the user additionally requires canonical representation across JWT and other
contracts. This chapter records the approved representation, not a new authority entity.

## Approved shape

Use one versioned `identity` block wherever a contract carries request identity.
Its two identity fields are:

- `actor`: a typed reference identifying the actual caller. Approved type values
  are `user`, `agent`, and `service_account`; `user` denotes a human, consistent
  with existing user-recipient illustrations.
- `human_id`: the Auth human identifier anchoring the applicable authority and
  `$self`. It is required for all three actor types, not inferred from arbitrary
  application material.

Canonical identity block for an agent request:

```json
{
  "version": "1",
  "actor": {"type": "agent", "id": "A-17"},
  "human_id": "U-17"
}
```

Canonical identity block for Vinay acting directly:

```json
{
  "version": "1",
  "actor": {"type": "user", "id": "U-17"},
  "human_id": "U-17"
}
```

For a direct human request, `actor.id` and `human_id` must agree. For a proxy,
trusted Auth-governed evidence must establish the actor/human association and
applicable delegation. This identity block describes that association; merely
possessing or submitting matching JSON does not prove it or confer authority.

## Why these fields

`actor` preserves the actual caller and its kind instead of impersonating the
human in the representation. `human_id` preserves the distinct authority anchor.
A single ambiguous `user_id` cannot express both for an automated request.
`version` lets consumers identify this shared contract's version independently
of its enclosing transport contract.

Using the same shape for direct and delegated calls makes the interpretation
stable. The alternative of omitting `human_id` for direct calls saves a repeated
identifier but requires consumers to derive it based on actor type. The approved
explicit repetition has one consistency rule rather than two representations.

No delegation identifier, role list, grant list, or new scope key is added here.
Those describe authority or supporting evidence, not just who is calling. Exact
delegation-reference representation remains a separate open contract question.

## Reuse and architecture

The approved block is carried under `identity` by contracts that represent the
requester. Candidate mappings to settle are:

| Location | Responsibility |
|---|---|
| JWT payload | Carry the identity representation under a defined issuer/verification contract. Exact claims and mapping remain open. |
| Verified server request context | Expose the established identities, not merely decoded or caller-submitted claims. |
| Authorization request and resolved-request context | Preserve the same identity meanings while obtaining current applicable authority. |
| Background-operation context | Preserve intended attribution and establish its validity again for execution-time authorization. |

Canonical does not mean every object gets an identity block. A grant's `recipient`
still identifies whom the grant is assigned to, including a group. It is not
replaced by requester identity. Minimal decision results are not expanded by this
proposal. Identity reuse also does not impose a transport or storage contract on
the excluded audit layer.

Tenant remains the mandatory enclosing trusted authorization context; it is not
introduced into individual scopes or grants by this identity proposal. Exact
JWT/tenant mappings remain open. Reusing identity values does not reuse an earlier
allow, freeze memberships, or bypass current human/delegation restrictions.

The existing single authorization gate remains unchanged. This decision aligns
its inputs across carriers rather than adding a resolver, prepared state, or
authentication protocol. HC-03-04/05 and full contract checkpoints remain open.

**Q-086:** Should we use this versioned `actor` plus `human_id` identity shape
across requester-bearing contracts, including direct human and proxy requests?

**Answer: approved.** The shared block uses `version`, typed `actor`, and required
`human_id`, including the equal-ID rule for direct human requests. This approval
does not freeze every enclosing contract, establish JWT claim mappings, or make
the block proof of current authority. [Q-087-B](jwt-identity-mapping.md) now approves
the JWT placement and human-subject mapping, including the rationale for keeping
`human_id` in the reusable block despite its duplication of JWT `sub`.
