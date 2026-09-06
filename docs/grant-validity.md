# Grant validity placement — Q-109

**APPROVED.** The user answered “yes, approved” after reviewing the exact JSON,
explicit-adoption rationale, and the trade-off for shortening expiry.
[Q-083](grant-lifecycle.md) already approves optional
grant time validity. [Q-108](assignment-validity.md) retains it and defers
assignment-specific windows in v1. Q-109 settles placement after
[Q-107's split](grant-revision-format.md): optional grant validity belongs to the
immutable grant revision, not the assignment or live grant-control record.

<details>
<summary>Earlier proposal status — superseded by explicit approval</summary>

Previous status: PROPOSED / NOT APPROVED. The question was placement after
Q-107's split, not whether grants can expire. The alternative live grant-wide
window was considered but not adopted; the reasoning remains below.

</details>

## Agreed placement and exact approved excerpt

Place optional `validity` in the immutable grant revision, alongside permissions
and scope. Changing or removing a time bound requires new published content and
explicit authorized adoption under existing latest-only and boundary rules.
Keep grant-wide enable/disable as live control, unchanged by this proposal.

Versioned grant-revision excerpt, not a complete API schema:

```json
{
  "version": "1",
  "grant_id": "G1",
  "revision": 2,
  "parent_grant_id": "G0",
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": {"dept": "FIN"},
  "validity": {
    "expires_at": "2026-09-30T00:00:00Z"
  }
}
```

Assume valid required G0 support and a valid subset. Assignments adopting this
revision could supply authority only before its expiry and while all other
checks pass. `validity` and `expires_at` already exist in Q-083; their placement
in immutable revision content is now approved. Optional `not_before` retains
Q-083's inclusive-start meaning; expiry is exclusive. Absence of a window means
no additional local time limit, not authority independent of upstream support.

## Rationale, alternatives, and consequences

- **Selected: revision content.** Lifetime is part of what a recipient adopts.
  A later expiry or removed bound must not silently extend adopted authority.
  For example, publishing revision 3 with October 31 expiry would leave an
  assignment on revision 2 constrained by September 30 until valid explicit
  adoption. Re-enabling the grant or assignment does not extend either window.
- **Alternative not adopted: live grant-wide window.** One edit could affect every adopted
  revision, simplifying shared deadline management, but an extension could
  silently expand the lifetime of existing assignments. This would need a
  separately agreed exception or mutation policy, not an implied Q-107 feature.
- **Accepted trade-off:** shortening a newly published window does not
  shorten older adopted revisions. Existing live grant-wide disable remains
  available for immediate logical withdrawal across revisions, subject to
  runtime freshness guarantees that still need specification. This decision
  does not introduce a separate grant-wide scheduled cutoff or scheduler.

**Core-philosophy check:** revision placement preserves explicit adoption and
prevents silent expansion of lifetime. Current required upstream validity still
constrains every revision, so an extended local window cannot bypass a parent's
expiry. It does not turn scope into a time-policy object or give assignments
their own validity window. Different adopted revisions may naturally have
different windows; that is revision adoption, not independent recipient timers.

**Q-109, answered “yes, approved”:** approve keeping grant validity inside its
immutable revision, so changing that validity requires explicit revision adoption?

Timestamp parsing, trusted clocks, concurrent checks/writes, complete update and
recovery contracts, and directly assigned human parent-revision selection remain
open. No new expiry/reset mechanism or runtime implementation is adopted here.
