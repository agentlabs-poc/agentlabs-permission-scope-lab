# Concurrent enforcement — impact-first discussion

## Q-129 — Already-authorized synchronous operation may finish (approved)

**APPROVED.** The user approved allowing the same ordinary synchronous application
operation to finish within its evaluated boundaries when it already received
allow before a later administrative authority withdrawal. The completed decision
is not retroactively canceled merely because that withdrawal is subsequently
confirmed.

```text
1. Vinay's payslip-read request receives allow.
2. Auth confirms removal of his Finance membership.
3. That same bounded synchronous request may finish.
4. New checks/retries/subsequent operations cannot reuse the withdrawn support.
```

**Rationale:** requiring cancellation of every already-authorized application
operation would require coordination between Auth and each running application
operation. Q-129 places the authorization boundary at the completed allow for
this bounded case, while Q-128 governs checks started after confirmed withdrawal.

**Security trade-off and philosophy check:** a previously authorized operation
can have its effect after authority withdrawal. That limited timing behavior is
explicitly accepted, not a stale-cache grace window for later requests. Authority
was valid when evaluated, and the completed allow cannot authorize different
data, broader scope, a retry, or additional operations. It is not a reusable
access token or indefinite authorization lease.

### Existing stronger guarantees remain

- Q-074 still requires the actual operation to preserve evaluated application
  boundaries. If the record moves from FIN to ENG before use, the original
  FIN-bound operation must stop; Q-129 does not excuse the changed reach.
- Q-110 still requires Auth's own authority-changing writes to preserve the
  validated authority/revision state through persistence. A conflicting reduction
  can stop that write even if its earlier checks passed. The Q-129 allowance is
  not a relaxation of Auth's administrative boundary validator.
- Q-128 still forbids withdrawn stale authority for checks started after Auth
  confirms the reduction. A new retry or subsequent operation needs a fresh
  authorization decision; the earlier allow cannot be transplanted.

Queued jobs, streams, and long-running operations are outside this approved
allowance. Their continuing authorization contracts remain separate, including
Q-075's already-agreed execution-time check for queued work. This decision does
not promise rollback of completed effects, specify a maximum request duration,
or prescribe a cancellation protocol.

An evaluation that has not yet issued allow when withdrawal is confirmed is
not the already-authorized case illustrated here. Exact evaluation ordering,
conflict/failure reporting, and evidence contracts must distinguish these cases;
do not treat merely starting a request as receiving allow.

No new identity/result field, prepared state, business workflow, or extra endpoint
authorization gate is introduced. Earlier blanket in-flight-work gaps below are
narrowed by this approved synchronous case, not eliminated for all operation types.

**Q-128 approved:** [new checks cannot use stale authority after any confirmed reduction](authority-freshness.md),
not just grant deletion. That timing boundary does not yet settle already-in-flight
or already-allowed work. Q-074's application-data boundary rule below and Q-110's
Auth-write consistency remain independently required.

## Q-074 / ENFORCEMENT-008 — preserve evaluated boundaries at use

Status: **AGREED.** The user answered Q-074 “Agree” after the Finance-to-Engineering
concurrent-move example and the recommendation to stop the original update.
This addresses an application record changing between an authorization check
and the protected operation. It does not settle in-flight grant revocation,
which is distinct from a record changing departments.

Previous status, retained as history: **PROPOSED, not approved** until that answer.

### Agreed example and outcome

Vinay has the required update permission scoped to Finance. Within the trusted
tenant, his request asks to update C-17 through the Finance endpoint:

1. Evaluation allows the request with department material FIN; application data
   also establishes that C-17 currently belongs to Finance.
2. Another authorized operation moves C-17 to Engineering before Vinay's update
   takes effect.
3. Vinay's original operation attempts to perform its update.

Step 3 **must not update C-17 under the earlier Finance-bound
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
is adopted. This decision does not require another Auth-service call immediately
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

**Answer: agreed.** Application-side enforcement must preserve the evaluated
boundary through execution; the earlier allow cannot authorize a changed reach.
The rationale, alternative, and implementation-neutral guarantee above are part
of the decision. Continue horizontally to Q-075 in
[background authorization](background-authorization.md).
