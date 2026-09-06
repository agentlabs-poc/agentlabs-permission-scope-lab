# Concurrent enforcement — impact-first discussion

## Q-074 / ENFORCEMENT-008 — preserve evaluated boundaries at use

Status: **PROPOSED, not approved.** This addresses an application record changing
between an authorization check and the protected operation. It does not settle
in-flight grant revocation, which is distinct from a record changing departments.

### Example and recommendation

Vinay has the required update permission scoped to Finance. Within the trusted
tenant, his request asks to update C-17 through the Finance endpoint:

1. Evaluation allows the request with department material FIN; application data
   also establishes that C-17 currently belongs to Finance.
2. Another authorized operation moves C-17 to Engineering before Vinay's update
   takes effect.
3. Vinay's original operation attempts to perform its update.

Recommend that step 3 **must not update C-17 under the earlier Finance-bound
allow**. The protected data operation must preserve the evaluated tenant,
department, and requested-record binding. If those constraints no longer hold,
stop that attempt; do not fall back to an ID-only write or silently substitute ENG.

This is not a new instruction to resolve grants in the handler. The application
already owns enforcing request-bound constraints; here it must preserve them
through use, not merely check them earlier and assume they remain true.

### Rationale, alternative, and trade-off

An earlier observation that C-17 belonged to Finance cannot authorize changing
it after a move into Engineering. Otherwise a valid initial check could be
followed by an out-of-boundary effect. Trusting the initial observation alone
simplifies the write path but leaves this gap.

The trade-off is that an otherwise valid request may fail because of a concurrent
change. Conditional writes, locking, and suitable transactional mechanisms are
possible implementation approaches, not prescribed mechanisms or interchangeable
guarantees. The handbook needs the boundary guarantee without embedding one
database's concurrency model in the canonical authorization contract.

No second canonical gate, new result fields, relationship block, or retry loop
is adopted. This proposal does not require another Auth-service call immediately
before every write. Any future retry or changed request must obtain authority
appropriate to its own material; the original allow cannot be transplanted.

### Remaining decisions

Exact conflict/error reporting, transaction isolation, version binding, automatic
retry policy, bulk rollback, streaming reads, and changes to Auth-side authority
while work is in flight remain open. A zero-row constrained write alone does not
prove why it failed; do not invent a policy-deny reason or disclose an inaccessible
record's new department. HC-09-07 remains open pending the full consistency contract.

**Q-074:** If a concurrent change invalidates an operation's evaluated application
boundary before its effect, must that attempt stop rather than use the earlier
allow to perform an out-of-boundary operation?
