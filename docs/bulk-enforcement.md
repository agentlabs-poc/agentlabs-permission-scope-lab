# Bulk enforcement — impact-first discussion

## Q-073 / ENFORCEMENT-007 — authorize the complete batch before effects

Status: **PROPOSED, not approved.** Q-071 and Q-072 settle collection-read behavior;
they do not silently settle bulk writes. This proposal concerns one bulk request
submitted as one operation, not several independent requests.

### Example and recommendation

Vinay has the endpoint's required delete permission with Finance scope. A bulk
request asks to delete C-17, C-18, and C-19 within Finance. Trusted application data
establishes that C-17 and C-18 belong to Finance but C-19 belongs to Engineering.
The request's Finance label is not proof that every selected certificate belongs
there; this is the existing application enforcement responsibility.

Recommend rejecting the whole operation without deleting any of the three.
Do not silently omit C-19, and do not delete C-17 before discovering that the
remaining batch cannot be authorized or satisfy its required boundary checks.
If all selected items are covered and all required checks succeed, the batch
may proceed to execution.

### Rationale, alternative, and ownership

This gives the caller one authorization outcome for the requested operation and
avoids partially performing a request that should have been rejected. The
alternative is per-item authorization with partial success, which needs an
explicit contract for item results, retries, and effects; it is not adopted by
this proposal. The trade-off is that one uncovered item blocks otherwise
permitted items in that batch.

The endpoint still declares one permission. The evaluator resolves authority;
the application binds the requested IDs to trusted tenant/department data and
enforces those constraints. No grant inspection, grant-derived query rewriting,
new canonical scope arrays, per-item permission declarations, or prepared state
is introduced. Batch request representation and evaluator integration are open;
this is a governing behavior proposal, not a finalized algorithm.

An evaluation timeout remains an error and prevents execution, not evidence for
a completed deny. The exact response for application-level missing or mismatched
records remains open; do not leak out-of-boundary record details in client reasons.

### Limits and remaining decisions

“No effects when the batch fails authorization or its required pre-execution
boundary checks” is not a claim that every later execution failure rolls back.
Database transaction atomicity, concurrent record changes, check-to-use binding,
retries, missing IDs, batch size limits, and asynchronous execution remain open.
The implementation must still preserve the approved boundaries during use;
preflight alone is not a solution to concurrent changes.

HC-09-05 remains open even if this principle is approved.

**Q-073:** For one bulk request, should failure of authorization or required
boundary checks for any selected item reject the whole operation before any
protected changes, rather than automatically processing a permitted subset?
