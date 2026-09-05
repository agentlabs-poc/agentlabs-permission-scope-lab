# Collection enforcement — impact-first discussion

## Q-071 / ENFORCEMENT-005 — authorized subsets for collection reads

Status: **PROPOSED, not approved.** Existing rules require actual output to stay
within authorized boundaries. This question chooses collection behavior when
the caller has authority over only part of the requested collection, rather than
introducing another permission or policy schema.

### Example

Vinay has the endpoint's registered listing permission with Finance scope.
The tenant contains Finance and Engineering certificates. He invokes a listing
endpoint designed to return certificates within the caller's authorized reach.

Recommend returning only the Finance certificates matching the request, rather
than denying the entire listing merely because Engineering certificates also
exist in that collection. Engineering certificates must not be returned.

This assumes valid authority for the listing operation. Absence of any applicable
authority is not equivalent to an authorized Finance listing with no matching
records; it cannot be silently turned into successful access by filtering.

### Rationale and alternative

A scoped listing lets the same endpoint serve callers with different boundaries
without granting access to the entire tenant. Requiring authority over every row
in the broader collection would make a valid Finance listing fail just because
unrelated departments have records.

The alternative is all-or-nothing authorization of the broader requested set.
That may be appropriate for an explicitly atomic business operation, but is not
the recommended semantics for this ordinary scoped list. Bulk mutations and
explicit-ID batch requests remain separate contracts, not implicitly filtered.

The application must constrain the actual collection data path; authorizing one
sample row does not authorize all rows. A partial-authority caller does not gain
broader reach because the handler uses an unconstrained query. This does not
require the evaluator to return scope fields, introduce a relationship block,
or change the one-permission/one-gate model.

### Remaining decisions

Pagination, counts, sorting, empty-result response details, exports, field-level
visibility, multiple-grant query composition, and bulk-write semantics remain
open. This question selects the governing behavior of scoped collection reads,
not a complete query-generation contract or a production implementation claim.
HC-09-03 remains open even if this principle is approved.

**Q-071:** Should a scoped list return the authorized subset, rather than denying
the whole listing because other out-of-scope records exist?
