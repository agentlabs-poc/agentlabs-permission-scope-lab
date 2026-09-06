# Assignment-specific time validity — Q-108

## Current decision — deferred in v1, approved

The user approved the revised recommendation: **retain optional grant validity
under Q-083; defer assignment-specific validity in v1.** No `validity` field is
added to the Q-107 assignment contract. The original proposal below is retained
as superseded discussion, not a supported v1 alternative.

An assignment cannot keep an expired grant effective. Its authority remains
constrained by the applicable grant validity and every required upstream route,
alongside status, permissions, scope, and dependency checks. Scope itself remains
a boundary selector, not an independently expiring entity.

Example: if G1 expires September 30, assignments to Maya and Nutan both follow
that limit, subject to any earlier loss of required support or explicit
disablement/removal. Giving Maya a September 15 deadline while Nutan retains
September 30 would require recipient-specific scheduled expiry, which is
deliberately deferred. Assignment enablement still does not mean effective access.

**Rationale / core-philosophy check:** assignment validity would add flexibility,
not close a correctness gap in grant validity. Deferring it keeps fewer lifecycle
controls and simpler resolution without weakening the subset invariant. The
accepted trade-off is no independently scheduled expiry per recipient sharing a
grant. Explicit disable/remove operations remain subject to existing authority
and binding guards; this decision does not add an automatic scheduling mechanism.

**Approval trail:** the user asked whether grant validity made assignment
validity necessary, then answered “approve” to the revised recommendation to
defer it in v1. Q-108 alone did not settle the shared grant window's placement;
[Q-109](grant-validity.md) subsequently approved immutable revision content,
not live grant control. Assignment-specific validity remains deferred.

<details>
<summary>Superseded initial Q-108 proposal — assignment validity was not adopted</summary>

**PROPOSED / NOT APPROVED.** Q-083 already agrees optional time-window meaning;
Q-107 now separates grant control, immutable grant content, and assignment.
This question concerns placement of a recipient-specific time limit, not
reopening start/expiry semantics or creating a scheduling language.

## Recommendation and exact proposed excerpt

Put a recipient-specific `validity` window on its assignment. This narrows the
authority that assignment can supply without changing every recipient of the
shared grant or requiring a new revision just for a different recipient window.

```json
{
  "version": "1",
  "id": "A1",
  "grant_id": "G1",
  "grant_revision": 2,
  "recipient": {"type": "group", "id": "Team1"},
  "status": "enabled",
  "validity": {
    "not_before": "2026-09-07T00:00:00Z",
    "expires_at": "2026-09-08T00:00:00Z"
  }
}
```

Assume valid supporting authority and revision 2 was latest at the applicable
creation/adoption. This assignment is time-eligible during September 7 UTC,
not before the start or at/after September 8. All other authority checks remain.

`validity`, `not_before`, and `expires_at` already appeared in Q-083. Their
placement on this recipient-bearing assignment is the proposed change; this
is not approval of all timestamp parsing or update endpoint contracts.

## Rationale and core-philosophy check

- A recipient's access duration belongs to its binding; shared grant content
  need not be duplicated or edited for another recipient's different duration.
- The window can only narrow effective authority. It cannot override a shorter
  valid lifetime or any other restriction in required upstream support.
- Expiry is effective time ineligibility, not a new stored status. An assignment
  can remain stored as enabled after expiry while supplying no authority.
- Re-enablement or revision adoption does not implicitly reset or extend the
  window. Changing the window is a separate explicit authorized change with
  applicable boundary validation.
- No window means no additional assignment-specific time restriction, not
  perpetual independent authority. Q-083's inclusive start/exclusive expiry
  and nonempty-interval requirement remain.

**Q-108:** approve this optional `validity` placement on assignments for
recipient-specific time limits?

## Limits

This does not remove grant-level or upstream time restrictions, select where
every shared-authority validity field belongs, migrate historical records, or
settle concurrency/trusted-clock and already-running-operation contracts. It
does not make adoption or enabling a way to bypass expiry. Complete format
validation and administrative update/recovery mechanisms remain open.

</details>
