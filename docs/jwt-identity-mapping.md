# Q-087 — JWT placement and subject mapping

Status: **Q-087-A and Q-087-B AGREED.** The user approved
the compatibility/enforcement conclusion and asked that exact canonical blocks
accompany proposals so they can be approved together. Q-086's identity block
remains approved. The user approved the exact JWT mapping after discussing the
intentional `sub`/`human_id` duplication. A complete token issuance/authentication
protocol is not defined.

Previous status, retained as history: **PROPOSED, not approved** before the
compatibility discussion and Q-087-A answer.

## Q-087-A — agreed compatibility and enforcement requirement

The user answered “i approve this” following the explanation that retaining the
human `sub` does not by itself enforce an agent's restricted authority.

Keep the existing human subject unchanged. Every protected operation using a
proxy token, including access to legacy software, must be subject to our evaluator's
applicable delegation checks. Legacy software must not provide a bypass merely
because it recognizes the human subject and ignores the proxy identity.

Rationale: compatibility preserves existing identity interpretation, but must not
turn delegated access into the human's unrestricted access. The evaluator reads
the richer canonical identity while enforcing the human ceiling and proxy limits.
This is an approved architectural requirement, **not verification that existing
deployments meet it or a demonstrated 100% compatibility result**.

No separate `agent_id` field is adopted; the approved typed `identity.actor`
already carries that identifier. Q-087-B now confirms the exact JWT identity
mapping without changing Q-086's canonical representation.

## Compatibility discussion and rationale

The user requires backward compatibility with existing JWT consumers: keep the
existing human `sub`, retain an identity block that our evaluator understands,
and avoid requiring changes to other software that may not be viable. The user
also suggests `agent_id` as the needed identifier. This feedback is recorded as
a compatibility requirement and a field suggestion, not approval of a new field
or a claim that compatibility has been demonstrated. Update: Q-087-A now approves
the enforcement requirement above; Q-087-B subsequently approved the exact mapping.

Preserving a human subject supports compatibility, but token-format compatibility
is not the same as preserving restricted proxy authority. For example, Vinay can
read and delete, while A-17 may only read. An older consumer that authorizes solely
from Vinay's subject could accept A-17's token and permit delete if nothing on
that protected path enforces the delegation limit.

Our evaluator can supply that enforcement only for operations it actually gates.
Controlling one evaluator does not establish that all legacy access paths pass
through it. JWT audience checks and distinct validation rules help prevent tokens
from being accepted in unintended contexts; their actual configuration and
enforcement must be established, not assumed from adding an identity field.
[RFC 8725 §§3.9–3.12](https://www.rfc-editor.org/rfc/rfc8725.html#section-3.9)

The approved `identity.actor.id` already identifies the agent and retains its
type alongside it. The earlier suggestion of a separate `agent_id` claim was
considered, but Q-087-B selects the existing typed actor without a duplicate claim.
Such duplication is not inherently more backward-compatible than the existing
block. No Q-086 fields are deprecated here.

Acceptance of older direct-human tokens without an identity block also needs an
explicit verified normalization rule. Missing identity cannot by itself prove
that a request is direct-human rather than delegated. The original proposal's
required-block rule below is not an approved legacy-token migration policy.

**Q-087-A — clarification:** Can every protected operation using an agent token,
including access to legacy software, be required to pass through our evaluator?
Without that coverage, unrestricted legacy acceptance and enforced proxy subsets
cannot both be promised. No “100% backward-compatible” result has been established.

**Answer: approved as an architectural requirement.** Deployment coverage still
needs verification. This question is no longer pending.

## Q-087-B — consolidated JWT identity mapping, agreed

Carry the approved block unchanged in the JWT payload's `identity` claim. For
this profile, require `sub` to equal `identity.human_id`: the subject is the
human whose authority supports the request. The actual caller remains
`identity.actor`, which may be that human, an agent, or a service account.

Identity-related payload excerpt for A-17 acting for Vinay. This is **not a
complete JWT or sufficient token-validation contract**; issuer, audience,
expiration, cryptographic protection, and their validation are omitted for focus,
not declared unnecessary:

```json
{
  "version": "1",
  "sub": "U-17",
  "identity": {
    "version": "1",
    "actor": {"type": "agent", "id": "A-17"},
    "human_id": "U-17"
  }
}
```

For direct human access, `sub`, `identity.human_id`, and `identity.actor.id`
all identify that human, with actor type `user`. For a service account, the
actor type is `service_account`; its human remains the subject.

The top-level version identifies our enclosing payload contract; the nested
version identifies the reusable identity contract. `identity` and `version`
are profile-specific claims, not claims mandated by the JWT standard. Their
meaning requires agreement between the issuing Auth service and consumers;
this is not a universal mapping for arbitrary third-party JWTs.

## Rationale and alternatives

### Why keep `human_id` when JWT already has `sub`?

The user challenged the duplication, then approved retaining it after the
following rationale was explained. Both values identify the **same human**;
`human_id` is not another identity or an additional authorization requirement.

The JWT's `sub` preserves existing human-subject interpretation. The nested
`identity.human_id` makes the approved canonical block self-contained when used
outside JWT: authorization requests, resolved-request context, and background
operations need the human anchor even when they do not carry a JWT payload.
Keeping only the agent's typed reference in that standalone block would lose
whose authority bounds it.

An alternative was to omit `human_id` from JWT and populate it from verified
`sub` in a transport adapter. That saves duplication in tokens but creates a
different transport shape instead of reusing the same identity block everywhere.
The user selected the self-contained block and accepted the JWT duplication.
The two values must agree; disagreement is invalid, not an opportunity to select
either value or silently repair the token.

This records the reason, trade-off, and alternative—not just the selected fields.
The unchanged block is documented in [canonical identity](identity-context.md).

### Standards context and subject/actor distinction

JWT defines `sub` as the subject of the claims, not invariably the actual caller.
It also requires a subject identifier to be unique within its issuer context or
globally. Issuer/context mapping cannot be discarded when interpreting IDs.
[RFC 7519 §4.1.2](https://www.rfc-editor.org/rfc/rfc7519.html#section-4.1.2)

OAuth token exchange illustrates the separate subject/actor distinction using
the `act` claim. That distinction supports the recommendation for our dependent
authority model, but our `identity` claim is a custom representation, not a
substitute claiming RFC 8693 interoperability.
[RFC 8693 §4.1](https://www.rfc-editor.org/rfc/rfc8693.html#section-4.1)

An alternative profile could make the caller the subject and keep the human
only in `identity.human_id`. The recommendation instead keeps the human subject
stable for direct and delegated access. Its trade-off is explicit: consumers
must use `identity.actor` for the caller and cannot equate `sub` with that caller.

A standards-oriented token-exchange integration could additionally map `act` to
the canonical actor. That would require an explicit adapter and consistency
rules, not a second independent identity source. This proposal does not adopt
`act`, token exchange, or delegation-chain semantics.

## Consistency and authorization boundary

For our agreed identity mapping, disagreement between `sub` and `identity.human_id`
is invalid identity context. Consumers must not pick whichever grants more
authority, silently overwrite one from the other, or use a human-only fallback
when the required identity block is absent.

Claims become established identity input only through the trusted verification
and attribution path. A valid token does not make delegation or human rights
permanent until token expiry. Q-069/Q-085 and the subset rule still require
applicable current authority and trusted proxy/human support. Identity mapping
must never become a replacement for grant resolution.

The existing flow remains: verify and establish identity context → endpoint-owned
authorization gate → constrained operation. No second gate or business-rule
mechanism is introduced. Tenant mapping, full token profile, allowed issuers,
audiences, token types, evidence transport, and external-provider adapters remain
open. No full-contract completion credit is claimed.

**Q-087:** Should our JWT profile carry the approved `identity` block unchanged,
with `sub = identity.human_id` and the actual caller retained in `identity.actor`?

**Q-087-B:** Approve the exact versioned JWT identity excerpt above, preserving
the human `sub` and the unchanged Q-086 `identity` block, with no duplicate
`agent_id` claim? The approved Q-087-A enforcement requirement applies. Other
token claims, legacy-token normalization, and the full token profile remain open.

**Answer: approved.** The user approved retaining `human_id` for a self-contained
canonical block, accepting its deliberate duplication of JWT `sub`, and requested
recording the rationale and committing this checkpoint. Q-087-B is no longer
pending. This does not claim complete JWT-profile approval or deployment validation.
