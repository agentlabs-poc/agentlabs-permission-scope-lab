# Operation-specific enforcement — impact-first discussion

## Q-068 / ENFORCEMENT-004 — moving data across an authorization boundary

Status: **PROPOSED, not approved.** The existing endpoint-policy chapter explicitly
leaves source-only, destination-only, and both-boundary move rules open. Its PUT
example establishes where a proposed value comes from, not authority to perform
the move. This discussion addresses that security decision, not new policy fields.

### Concrete question

Certificate C-17 currently belongs to Finance. A request proposes moving it to
Engineering within the same trusted tenant. Assume the application registers one
permission for this operation, illustrated by `hrms:employee:certificate::move`.

Recommend requiring authority for that move covering **both the current Finance
boundary and the proposed Engineering boundary**. Source authority alone would
not authorize placing data in another boundary; destination authority alone would
not authorize taking data from a boundary the caller cannot administer.

| Available move authority | Proposed result for FIN to ENG |
|---|---|
| Finance only | Deny: destination is not covered. |
| Engineering only | Deny: current/source boundary is not covered. |
| Both boundaries covered under the eventual complete authority rules | May allow only if every other mandatory constraint is satisfied. |

This does not assume two different grants can be combined for one move. That
composition question remains open; Q-068 asks which states need authority, not
how multiple grants satisfy them. A valid tenant-wide move grant can illustrate
coverage of both boundaries, without eliminating other restrictions.

### Rationale and alternatives

Changing a scope-relevant value changes the reach of future access. Checking
only the destination could let someone bring inaccessible data into their own
boundary. Checking only the source could let someone place data into a boundary
where they have no authority. Both-state authorization addresses both directions.

Source-only and destination-only checks are simpler but protect only one side of
the transition. Their narrower enforcement is not assumed sufficient here.
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
open. This proposal does not authorize cross-tenant moves or claim that the runtime
implements the rule. Approval of this principle alone will not close HC-09-04.

**Q-068:** For this boundary-changing move, should authorization cover both the
current and proposed boundaries?
