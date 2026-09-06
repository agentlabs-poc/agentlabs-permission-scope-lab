# Bulk enforcement — impact-first discussion

## Q-130 — Complete grant routes may cover different batch items (approved)

**APPROVED after the expanded two-grant explanation.** Every item in a batch
must have complete valid authority for the endpoint's one required permission.
Different items may rely on different complete grant routes; a single grant
covering the whole batch is not required.

### Exact grant examples

Vinay has these two grant revisions available through valid assignments and
membership, with current valid parent support and all other restrictions met.
Tenant is implied; G-ROOT supports both permissions/boundaries shown.
These are existing-format grant revisions, not a new batch or assignment schema.

G-FIN supplies Finance certificate-delete authority:

```json
{
  "version": "1",
  "grant_id": "G-FIN",
  "revision": 1,
  "parent_grant_id": "G-ROOT",
  "permissions": ["hrms:employee:certificate::delete"],
  "scope": {"dept": "FIN"}
}
```

G-ENG supplies Engineering certificate-delete authority:

```json
{
  "version": "1",
  "grant_id": "G-ENG",
  "revision": 1,
  "parent_grant_id": "G-ROOT",
  "permissions": ["hrms:employee:certificate::delete"],
  "scope": {"dept": "ENG"}
}
```

The request is one batch deleting C-17 and C-18. The application establishes
their actual department facts, not merely caller-claimed department labels.

| Item | Actual department | Complete supporting route |
|---|---|---|
| C-17 | FIN | G-FIN and its required lineage/assignment/membership support |
| C-18 | ENG | G-ENG and its required lineage/assignment/membership support |

```text
C-17 has complete delete authority through G-FIN
                       AND
C-18 has complete delete authority through G-ENG
                        ↓
One authorization outcome for the complete batch
```

Both items must be covered and all mandatory checks must succeed before any
protected effects. Including C-19 in HR, without a complete route covering its
deletion, denies the whole batch. Do not delete C-17/C-18 or silently omit C-19.
Inability to complete evaluation is an error, not a fabricated completed deny;
it also prevents execution.

### No permission/scope fragment mixing

If G-ENG supplies only read, it cannot authorize deleting C-18. The evaluator
must not borrow delete from G-FIN and combine it with Engineering scope from
G-ENG. Each item needs a complete permission-and-scope route with all applicable
lineage, validity, membership, assignment, and delegation restrictions.

AND between required item checks does not change AND within each grant's scope.
The evaluator is not merging FIN and ENG into a broader synthetic grant or
inventing an OR/array scope operator. Different complete routes provide different
items' authority while retaining their own complete boundaries.

### Rationale, ownership, and trade-offs

Vinay could already delete either certificate in a separate authorized request.
Batching those operations should not require a broader grant merely to package
them together. Requiring one grant for the entire batch would create that extra
authority-distribution burden; the user approved per-item complete-route coverage.

The evaluator performs coverage resolution. The endpoint still declares one
permission, supplies trusted material for the requested items, and enforces the
same boundaries during execution. It does not inspect grants, derive a different
query, or return an automatically filtered successful subset. There is one final
authorization decision, not per-item partial execution permission.

The cost is complete batch preflight and preservation of each item's evaluated
boundary through use. Q-073's distinction between authorization failure and later
execution failure remains: this does not promise database rollback after every
possible execution error. Q-074 concurrent application-boundary protection remains.

### Remaining boundaries

Exact batch material/policy/request and provenance contracts remain open; no new
JSON transport or endpoint source syntax is adopted here. This finite requested-
item rule does not authorize an unbounded collection merely because all currently
visible rows happen to be covered (Q-072). It does not yet settle using different
grants for the source and destination states of one move (Q-068).
Existing historical batch-contract gaps below are narrowed by this approval,
not evidence that every bulk implementation requirement is complete.

## Q-073 / ENFORCEMENT-007 — authorize the complete batch before effects

Status: **AGREED.** The user answered Q-073 “Agree” after the three-certificate
example and the distinction between authorization failure and later database
failure. This concerns one bulk request submitted as one operation, not several
independent requests. Q-071 and Q-072 settled reads; this separate approval
settles the governing bulk authorization rule.

Previous status, retained as history: **PROPOSED, not approved** until that answer.

### Agreed example and outcome

Vinay has the endpoint's required delete permission with Finance scope. A bulk
request asks to delete C-17, C-18, and C-19 within Finance. Trusted application data
establishes that C-17 and C-18 belong to Finance but C-19 belongs to Engineering.
The request's Finance label is not proof that every selected certificate belongs
there; this is the existing application enforcement responsibility.

Reject the whole operation without deleting any of the three.
Do not silently omit C-19, and do not delete C-17 before discovering that the
remaining batch cannot be authorized or satisfy its required boundary checks.
If all selected items are covered and all required checks succeed, the batch
may proceed to execution.

### Rationale, alternative, and ownership

This gives the caller one authorization outcome for the requested operation and
avoids partially performing a request that should have been rejected. The
alternative is per-item authorization with partial success, which needs an
explicit contract for item results, retries, and effects; it is not adopted by
this decision. The trade-off is that one uncovered item blocks otherwise
permitted items in that batch.

The endpoint still declares one permission. The evaluator resolves authority;
the application binds the requested IDs to trusted tenant/department data and
enforces those constraints. No grant inspection, grant-derived query rewriting,
new canonical scope arrays, per-item permission declarations, or prepared state
is introduced. Batch request representation and evaluator integration are open;
this is an agreed governing behavior, not a finalized algorithm.

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

HC-09-05 remains open despite approval of this principle because transaction,
concurrency, retry, and representation contracts remain unfinished.

**Q-073:** For one bulk request, should failure of authorization or required
boundary checks for any selected item reject the whole operation before any
protected changes, rather than automatically processing a permitted subset?

**Answer: agreed.** The explanation, rationale, alternative, and distinction
from execution rollback above are retained as part of the approved record.
Continue horizontally to Q-074 in [concurrent enforcement](concurrent-enforcement.md).
