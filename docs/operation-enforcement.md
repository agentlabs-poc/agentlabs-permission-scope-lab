# Operation-specific enforcement — impact-first discussion

**Q-130 related approval:** [different batch items may use different complete grant routes](bulk-enforcement.md).
That does not yet settle Q-068's distinct open question of separate grants for
the source and destination of one move. Both move boundaries remain mandatory.

Q-090 separates [grants and assignments](grant-assignments.md). Recipient-bearing
grant examples below are deprecated layouts, retained to explain operation
boundaries. Their enforcement requirements remain; this chapter does not provide
the new full assignment or subgroup contract.

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

## Q-150 / ENFORCEMENT-013 — composition is not a grant question

Status: **AGREED, and it answers Q-131 by dissolving it.** The user declined the
framing: *"it doesn't really matter if it is one grant or two grant as long as the
resolved grant has the permission… and the required scope. So the question is more
about designing the API than the grant itself."*

Q-068 requires authority over both the current and the proposed boundary of a
move. Q-131 asked whether those two may be covered by two different grants. That
question is withdrawn: **the model has no opinion about composition.** What is
required is that the resolved authority covers the permission and the boundary
being evaluated. Whether it arrived from one grant or several is not a property
this model reasons about, and no rule should depend on it.

### How one operation covers two boundaries

A move is **one endpoint and one gate**, and the gate evaluates **once per
boundary**. The same permission is asked about each, and the effect runs only if
every evaluation allows:

```json
{ "permission": "hrms:payroll:payslip::write", "material": { "dept": "FIN" } }
{ "permission": "hrms:payroll:payslip::write", "material": { "dept": "ENG" } }
```

Each evaluation is a complete route against one boundary, so no fragment is ever
mixed — the concern Q-130 names is structurally absent rather than guarded
against. And because both evaluations happen in one request at one gate, "both
hold" means at one moment, which is what a boundary-changing operation needs.

**Expressing the two boundaries is the endpoint's obligation.** A boundary is
identified by a registered scope key, and a request carries one value per key, so
a single evaluation cannot mean two departments at once. An endpoint that moves
data declares both, and asks for each. Getting that wrong produces an endpoint
that authorizes one end of a move and not the other, which is exactly what Q-068
forbids — and it is endpoint design, not a gap in the authorization model.

### Create and update follow from the same rule

| Operation | Boundaries evaluated |
|---|---|
| Create | the **proposed** boundary only — there is no current one |
| Update that does not change the boundary | the **current** boundary only |
| Update that changes the boundary | **both** — it is a move, whatever the endpoint calls it |
| Move | **both** |

The last row is the one worth stating: an update is a move when it changes the
boundary, and an endpoint cannot escape the both-boundary requirement by naming
the operation differently.

### Rationale / conscious tradeoff

Refusing composition as a concept is what keeps this simple. The alternative —
requiring a single grant to cover both ends — would push administrators toward
broader grants: to let somebody move records between two departments they already
administer, they would have to grant tenant-wide authority instead. A rule that
makes legitimate use require more access than it needs is the wrong rule.

The cost falls where the user placed it. The model cannot check that an endpoint
asked about both boundaries; it can only answer what it is asked. An endpoint that
asks about one end is wrong, and this handbook cannot detect it — the same duty
split as [Q-138](endpoint-authorization.md), and the same reason
[Q-144](policy-scope-boundary.md) is worth settling, since a declared boundary is
what would make it checkable.

