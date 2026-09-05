# Delegation lifecycle — impact-first discussion

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
