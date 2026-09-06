# Auth write consistency — Q-110

**AGREED.** The user answered “agree” after the concurrent publication / parent
disablement example, recommendation, rationale, and trade-off. This is the
Auth Service's own protected write, not
a new application endpoint responsibility. [Q-100](auth-service-authority-gate.md)
already requires administrative authorization and authority-boundary validation.
[Q-104A/Q-105](grant-revisions.md) require latest published grant content for
creation and explicit upgrades. Q-110 requires preserving the checked revision
and relevant authority state through persistence; a conflicting change stops
the attempt, without a stale write or silent revision substitution.

<details>
<summary>Earlier status — proposal before Q-110 approval</summary>

Previous status: PROPOSED / NOT APPROVED. The open question was what happens
when relevant state changes between validation and persistence. The user has
now approved the rule, not a particular storage mechanism or error schema.

</details>

## Agreed rule

The protected Auth write must preserve the revision selection and relevant
authority/binding state on which validation depended. If a conflicting change
wins before the write, stop that attempt without partially creating/updating the
assignment. Do not apply stale approval or silently select a different revision.
An explicit fresh attempt must pass both gates against current state.

This requires a consistency guarantee through persistence, not simply another
read that leaves a new check-to-write gap. Transactions or conditional writes
are possible implementation choices, not a prescribed database or new canonical
wire field. Define a consistent ordering for concurrent operations: a publication
ordered after a successful assignment write does not retroactively invalidate
that earlier latest-only selection.

## Concrete latest-revision race

1. An assignment-creation attempt selects G1 revision 2, currently latest, and
   validates its permissions, scope, validity, and required support.
2. Another operation publishes revision 3 before the assignment write succeeds.
3. The original attempt must not create an assignment to revision 2 as though
   it were still latest. It must also not silently substitute unreviewed revision
   3. Stop the attempt; a fresh attempt can explicitly select and validate 3.

The same preservation principle applies when required parent support, membership
used for administration, enablement, relevant team bindings, or time eligibility
changes so the earlier validation no longer holds. Unrelated changes need not
invalidate an attempt; exact dependency tracking remains an implementation
contract. Existing assignment content stays unchanged on a failed upgrade,
although concurrent upstream changes may independently make it ineffective.

## Rationale, philosophy, and trade-off

Checking a valid boundary earlier is insufficient if a different state is used
to authorize the write. The recommendation preserves explicit adoption, latest-
only creation/upgrades, and subset authority inside Auth's own gate. It adds no
application business logic, new evaluator layer, or implicit repair behavior.

The trade-off is that a request which initially passed can fail during concurrent
administration. Trusting the earlier check is simpler but can admit stale writes;
silently retrying with a newer revision can adopt authority the requester did not
select. Neither alternative is recommended.

This does not settle the global request-time revocation/cache contract, external
side-effect rollback, automatic retry policy, database isolation, or the precise
error/HTTP format. [Q-074](concurrent-enforcement.md) remains the separate approved
application-data boundary-through-use rule. No runtime implementation is included.

**Q-110, answered “agree”:** approve stopping an Auth assignment-write attempt when a conflicting
change invalidates its checked revision or authority before persistence, with no
stale write or silent revision substitution?
