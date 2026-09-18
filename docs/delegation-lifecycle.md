# Delegation lifecycle — impact-first discussion

**Q-128 approved:** [confirmed reductions cannot survive in stale evidence for new checks](authority-freshness.md).
Delegation withdrawal or loss of required human support is covered by the same
freshness guarantee as grant deletion. Restoration still follows Q-070 when its
requirements are met; exact evidence/propagation mechanisms remain open.

## Q-127 — Direct human-to-proxy delegation only in v1 (approved)

**APPROVED.** The user approved excluding proxy-to-proxy authority delegation
chains from v1. Each supported delegation connects its human authority anchor
directly to the acting agent or service account, without another proxy as an
intermediate delegator.

```text
Supported in v1
Vinay → Agent A
Vinay → Agent B
Vinay → Service account S

Not supported in v1
Vinay → Agent A → Agent B
        B's authority depends on A's delegation as an intermediate source
```

Multiple proxies can use the same human anchor through their own separately
authorized, bounded delegations. B needs its own valid human-linked delegation;
it cannot acquire authority merely because A calls it, supplies a token/context,
or asks it to perform work. Do not flatten a prohibited chain by discarding the
intermediate actor's identity or restrictions and claiming direct human authority.

Agent collaboration is not prohibited. It simply does not implicitly transfer
authority: each protected operation still needs the actual caller's established
delegation, human support, and all required boundaries. This rule applies equally
to agent and service-account combinations.

**Rationale:** intermediate proxy dependencies add withdrawal, restoration,
expiry, and attribution cases. Direct human-linked delegation keeps v1's authority
model smaller while preserving multiple independently bounded proxies per human.
The alternative of supporting chains would require every intermediate delegation
and its restrictions to remain valid; that feature is explicitly not supported
in v1, not an unfinished v1 implementation requirement.

**Core-philosophy check:** proxies remain human-dependent subsets, never independent
authority or first-class group members. Q-070's restoration behavior remains for
still-valid direct delegations. Team/grant hierarchy and parent-supported narrowing
remain supported; those are not intermediate proxy delegators. A human's source
authority can still come through a valid group/grant lineage.

**Trade-off and remaining scope:** B requires an authorized human-linked delegation
instead of inheriting A's delegation. No new JSON field, chain identity block,
creation permission, or automatic delegation grant is introduced. Whether a proxy
may perform an administrative creation operation is not decided by calling it
collaboration; it requires its own authorized operation and boundary contract.
This decision also does not settle account-level multiple-human associations or
ownership transfer. Exact direct-delegation lifecycle/evidence remains open.

Earlier redelegation-chain gaps below are retained as history but are no longer
v1 requirements. Do not reintroduce chains through a transport adapter or implicit
human impersonation.

## Existing foundation

AUTHORITY-002 makes every service account and agent dependent on a human, with
authority bounded by that human's applicable rights and delegation limits.
DELEGATION-002 invalidates authority derived through a required supporting
relationship when it breaks. Losing unrelated rights does not invalidate all
of an agent's access. These rules remain agreed; no independent service authority
or first-class agent group membership is introduced here.

The [grant chapter](grant-model.md) previously left restoration open. Q-070 now
settles automatic restoration when supporting human authority returns; other
delegation lifecycle details remain open. It does not reopen the already-settled
requirement to remove unsupported access.

## Q-070 / DELEGATION-003 — restoring a broken delegation

### Current conclusion — agreed as corrected by the user

The user clarified: “deligation is subset of vinay, 2 is invalid and becomes
in active. at 3-> works again.” This selects **automatic reactivation**, not the
previously recommended explicit renewal. No additional approval step is required
solely because the human temporarily lost and later regained supporting access.

| State | Human's Finance authority | Affected delegated Finance access |
|---|---|---|
| 1. Vinay has supporting access | Valid | Usable within the delegation's limits. |
| 2. Vinay loses supporting access | Not valid | Inactive; cannot authorize Finance access. |
| 3. Vinay regains supporting access | Valid again | Automatically usable again where current authority covers the still-valid delegation. |

Rationale: delegation is a continuing restriction of Vinay's applicable authority,
not independent authority and not automatically a permanently revoked record when
his rights shrink. Removing human support makes the affected effective access
inactive. Restoring that support makes the corresponding access available again
under the existing delegation limits and mandatory constraints. This keeps the
model dependent on current entitlement without a separate renewal workflow.

This does not revive a delegation that has itself been explicitly revoked,
expired, or otherwise remains invalid. Nor does it restore access beyond Vinay's
current rights or beyond the delegation's own limits. Unrelated supported access
is unaffected. “Inactive” describes effective access here; no stored status enum,
new JSON field, or automatic account deletion is adopted.

The trade-off is intentional: old still-valid automation can regain access when
its human regains support. To prevent that return, an authorized actor must
invalidate the delegation itself rather than rely on temporary loss of human
rights. The exact lifecycle operation and administration permissions remain open.
Missing evidence is still not proof that the underlying relationship broke.

**Q-070 — answered with a correction:** no explicit renewal is required for this
case; step 2 is inactive and step 3 works again. Supporting-reference mechanics,
freshness, concurrent changes, expiry representation, growth, and redelegation
remain open. Move horizontally next to [collection access](collection-enforcement.md).

### Original explicit-renewal proposal — not adopted, retained as history

Status: **PROPOSED, not approved.** Recommend that once a required supporting
relationship has broken and the affected delegated authority is invalidated,
that delegation does not automatically regain validity merely because the human
later regains the corresponding access. Require an explicit authorized renewal
of the affected delegation before it becomes usable again.

### Example

1. Vinay belongs to Finance and has delegated Finance certificate-read authority
   to agent A through that supporting membership.
2. Vinay is removed from Finance. That supporting relationship breaks, so A can
   no longer use that delegated Finance authority under the existing rules.
3. Vinay is later added back to Finance and regains his own Finance access.
4. Proposed: A's old affected delegation remains unusable until explicitly renewed
   through an authorized action, evaluated against Vinay's current authority.

This is not a blanket shutdown of A. Unrelated still-supported delegated access
remains governed by its own dependencies. It also does not redefine the ordinary
human/group grant lifecycle or require a new independent grant to the agent.

### Rationale and alternative

The recommendation prevents forgotten or previously invalidated automation from
silently regaining access when the human's membership changes later. Renewing the
delegation makes the restoration intentional rather than an incidental effect
of restoring the human's access.

The alternative is automatic reactivation whenever the original delegation and
current human rights again satisfy all constraints. That is operationally simpler
and may suit temporary membership changes, but can restore old automation without
a fresh deliberate action. Explicit renewal instead adds operational work and
requires lifecycle state capable of recognizing the broken delegation.

### Limits and return points

Who may renew, the concrete renewal operation, whether a replacement record is
created, supporting-grant substitution, detection/propagation timing, expiry,
redelegation chains, and race handling remain open. This does not turn every
transient lookup failure into permanent delegation invalidation: missing evidence
is not proof that the underlying relationship broke. No new JSON fields or
automatic account deletion is proposed.

**Q-070:** After a required supporting relationship breaks, should restoring the
human's access require explicit authorized renewal of the affected delegation,
rather than automatically reactivating it?

## Q-154 / DELEGATION-004 — a delegation's lifetime, growth and administration

Status: **AGREED.** The user's position was that these should already be clear, and
that the recommendations stand if they are not. Two of the three were not stated
anywhere; the middle one is the consequence of a rule already agreed and is written
here as a consequence rather than a choice.

### What a delegation is, first

A delegation is **not a grant and not a step in a lineage.** It does not add
authority and it does not appear in a contributing chain. It is permission for an
**actor** to ask about a **human** — the two parties Q-086 already keeps apart —
and the authority that comes back is the human's.

```
  vinay ──▶ agent A                agent A may ask about vinay
                                   the answer is VINAY'S authority
                                   A adds nothing to it
```

That is why a delegation needs no representation in a resolved answer. The
transport already names the asking actor and the subject separately, and a
delegated request differs from a direct one only in the actor.

### Lifetime

A delegation **carries its own validity window**, and the effective lifetime is
the **narrower** of the delegation's and the human's authority.

This is not a new rule so much as the existing one applied. A route's validity is
already the narrowest window in its chain; a delegation is another constraint on
the same request, so it folds the same way. Without a window of its own a
delegation would be the only thing in the model that outlives nothing — revocable,
but never expiring.

### Growth

A delegation **tracks the human's authority.** What a proxy may do is resolved
from what its human holds **at the time of the request**, not at the time the
delegation was created.

This follows from proxies being *"human-dependent subsets, never independent
authority"*. A subset is resolved, not copied. Freezing a snapshot would make a
delegation diverge from the authority it depends on, with nothing reconciling
them, and would require a delegation write for every grant the human later
receives.

**The cost, stated:** an agent silently gains reach when its human does. That is
inherent in delegation being dependent rather than a copy — and it is the same
property that makes revoking the human's authority remove the proxy's reach with
no delegation write at all.

### Administration

**The human administers their own delegations.** Vinay creates and revokes them,
and an administrator does not create one on his behalf, because the proxy acts as
an extension of him and he is accountable for what it does.

An administrator is not powerless: removing vinay's authority removes the
delegation's reach immediately, through growth above, without touching the
delegation record.

### Records and evidence

A delegation record names the **human**, the **proxy**, and its **validity
window**. Nothing else is required, and no new field enters the resolved answer.

**Left open deliberately:** whether a delegation may also carry its own *scope*,
narrowing the human's authority further. Nothing in the model needs it — the
human's own grants already bound the proxy, and a narrower delegation can be
expressed by delegating from a human who holds less. If a case appears that this
cannot express, that is the moment to add it.

### Rationale / conscious tradeoff

The whole of this rests on a delegation being a permission to *ask*, not a holding
of authority. Every part follows from that: it needs no lineage representation, it
tracks rather than copies, and it is the human's to give because it is the human's
authority being used.

The tradeoff is in growth, and it is deliberate. A model where delegations froze
would be more predictable and would rot: proxies would hold authority their humans
had lost, or lack authority their humans had gained, and reconciling the two would
become an administrative task nobody owns.

