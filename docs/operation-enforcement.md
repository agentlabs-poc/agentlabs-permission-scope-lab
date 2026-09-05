# Operation-specific enforcement — impact-first discussion

## Q-068 / ENFORCEMENT-004 — moving data across an authorization boundary

Status: **AGREED.** The user requested a fuller Finance-to-Engineering explanation
and then answered “agree.” Original status retained as history: ~~PROPOSED, not
approved~~. The endpoint-policy chapter previously left source-only,
destination-only, and both-boundary move rules open; Q-068 now settles the
both-boundary rule for this move. Its PUT binding example alone did not establish
authority to perform the move. No new policy fields are introduced.

### Concrete question

Certificate C-17 currently belongs to Finance. A request proposes moving it to
Engineering within the same trusted tenant. Assume the application registers one
permission for this operation, illustrated by `hrms:employee:certificate::move`.

Require authority for that move covering **both the current Finance boundary
and the proposed Engineering boundary**. Source authority alone does
not authorize placing data in another boundary; destination authority alone would
not authorize taking data from a boundary the caller lacks move authority over.

| Available move authority | Agreed boundary rule for FIN to ENG |
|---|---|
| Finance only | Deny: destination is not covered. |
| Engineering only | Deny: current/source boundary is not covered. |
| Both boundaries covered under the eventual complete authority rules | May allow only if every other mandatory constraint is satisfied. |

This does not assume two different grants can be combined for one move. That
composition question remains open; Q-068 asks which states need authority, not
how multiple grants satisfy them. A valid tenant-wide move grant can illustrate
coverage of both boundaries, without eliminating other restrictions.

### Approved explanatory grant and material

This simplified example omits lifecycle fields; it is not a full grant schema:

```json
{
  "version": "1",
  "recipient": { "type": "user", "id": "vinay" },
  "permissions": ["hrms:employee:certificate::move"],
  "scope": { "dept": "FIN" }
}
```

It supplies Finance move authority, not Engineering move authority. The endpoint
distinguishes current department FIN, established from trusted application data,
from proposed department ENG, supplied as validated request input. This grant
alone cannot authorize the FIN-to-ENG move. A valid tenant-wide move grant with
scope `{}` could cover both departments, subject to the other mandatory checks.

Authority over both boundaries means authority for the declared move operation,
not an additional requirement for read permissions. Both boundary checks belong
to one endpoint-owned authorization decision; they do not create two gates.

### Rationale and alternatives

Changing a scope-relevant value changes the reach of future access. Checking
only the destination could let someone bring inaccessible data into their own
boundary. Checking only the source could let someone place data into a boundary
where they have no authority. Both-state authorization addresses both directions.

Source-only and destination-only checks are simpler but protect only one side of
the transition. Neither alternative is adopted for this move.
Whether some separately defined business operation has different semantics is
not decided by a generic read or write permission name.

### What remains unchanged and open

The endpoint establishes the actual current state from trusted application data;
the proposed body value does not prove current ownership. The application owns
domain meaning and actual-use enforcement. There is still one declared permission
and one endpoint-owned gate, with no relationship block, canonical resource wrapper,
or returned-scope requirement. Auth does not interpret Finance or Engineering.

Exact source/destination material bindings, grant composition, concurrent-change
protection, atomicity, create/update distinctions, and move-result handling remain
open. This decision does not authorize cross-tenant moves or claim that the runtime
implements the rule. Approval of this principle alone does not close HC-09-04.

**Q-068 — answered yes:** this boundary-changing move requires authorization
covering both current and proposed boundaries. Under the impact-first pass, move
next to [revocation and freshness](authority-freshness.md), keeping the detailed
move-composition and concurrency questions open.
