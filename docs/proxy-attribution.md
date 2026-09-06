# Q-085 — trusted human attribution for proxy requests

Status: **AGREED.** The user answered “i agree” and additionally required canonical
representation across JWT and other places carrying this identity context.
Previous status, retained as history: **PROPOSED, not approved** until that answer.
This question belongs to vocabulary/identity
and the request-context boundary, not business logic or authentication-protocol
design.

## Already agreed

Every agent and service account has human-dependent authority, bounded by the
human's applicable rights and the delegation's limits. Groups contain humans,
not first-class proxies. `$self` retains the authorizing human as its anchor.
Resolving authority must never produce independent proxy grants. These decisions
are not being reopened.

## Agreed rule

For a proxy request, the authorization context must retain both the verified
calling agent/service-account identity and the human whose authority it uses.
Their association must be established through trusted Auth-governed delegation
evidence, not inferred merely from a human identifier supplied by the caller.

A path, body field, or header can carry a requested reference, but cannot itself
prove that the authenticated proxy may act for that human. The evaluator must
not replace the human anchor or fall back to independent proxy authority when
the association cannot be established. Only established human support and the
applicable delegation limits can contribute to an allow.

This is a trust requirement for authorization input, not a requirement for a
remote Auth call on every request. Credential verification, token formats, evidence
transport, and caching mechanisms are not selected here. Existing freshness and
allow/deny/evaluation-error rules continue to apply.

## Example and counterexample

Agent A-17 is verified as the caller. Trusted delegation evidence links its
applicable delegated access to Vinay. Evaluation uses Vinay's applicable authority
and the delegation's limits; it retains A-17 as the actual caller.

If A-17 submits Maya's identifier in the body, that value cannot switch evaluation
to Maya's grants or make `$self` mean Maya. Such an association needs its own valid
trusted evidence; a claimed name is insufficient. Likewise, knowing that Vinay
created an account is not by itself proof of the currently applicable delegation.

## Rationale and alternatives

The subset invariant is meaningful only if its human anchor is trustworthy.
Using an arbitrary caller-supplied human identifier would let a proxy select
someone else's authority. Keeping only the human and discarding the proxy identity
would lose the identity needed to apply that proxy's delegation restrictions.

The recommended approach retains both identities and verifies their association.
It adds no business rules, independent service authority, group nesting, or new
canonical JSON fields. It also does not decide whether one proxy identity may
hold multiple human delegations; association must be established for the authority
used by this request regardless of that later account-model decision.

## Architectural position and open work

The existing flow remains: verified caller/context → endpoint-owned authorization
gate → constrained application operation. The embedded agent needs established
proxy/human attribution before it can use the human's authority. This does not
create a second authorization gate or revive prepared results.

HC-03-05 remains open: tenant mapping, complete attribution contracts, and evidence
integration are not finalized by this decision. No completion credit is claimed.

**Q-085:** Should every proxy request retain the verified proxy identity and its
human authority anchor, with their association established from trusted
Auth-governed evidence rather than caller-supplied identity claims?

**Answer: approved.** The user additionally requires a consistent canonical
representation, beginning with JWT and extending to other contracts that carry
request identity. Rationale: preserving the two identities conceptually is not
enough if each component renames, flattens, or interprets them differently.
The exact fields and JWT mapping were not approved by that requirement; see
[Q-086's now-approved shared identity shape](identity-context.md). Q-087-B now
approves its JWT placement and subject mapping; full integration contracts remain open.
