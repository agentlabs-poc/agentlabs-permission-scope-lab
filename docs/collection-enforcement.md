# Collection enforcement — impact-first discussion

## Q-071 / ENFORCEMENT-005 — deny instead of deriving an authorized subset

### Current conclusion — agreed as corrected by the user

The user rejected automatic subset filtering: “too much intelligence in endpoint
about auth. it shold just deny.” For the partially authorized collection request
under discussion, return **deny**, not a successful result narrowed to the rows
the caller could access. The earlier recommendation below is **not adopted**.

Rationale: the endpoint should not inspect grants and construct a different,
caller-specific collection query to salvage an unauthorized request. It supplies
its declared request material to evaluation and obeys the result. Permission and
scope resolution remain the evaluator's responsibility; the application still
owns enforcement of its actual data access against the authorized request.

In the example, Finance-only authority must not turn a broader Finance-and-
Engineering listing into a successful Finance-only response. A completed lack
of authority produces the agreed deny contract and reasons. Inability to
complete evaluation remains an error, not a fabricated deny.

The conscious trade-off is less automatic discovery of the caller's accessible
records in exchange for a simpler, explicit request contract and no grant-to-
query translation in the endpoint. This does not remove the endpoint's existing
responsibility to bind its data query to the evaluated material. It does not
introduce a second authorization gate, returned scope fields, or a query language.

Q-072 below now records the approved distinction for an explicitly Finance-limited
request. It was proposed separately and subsequently approved by the user.
Bulk writes, counts, pagination, exports, and schema details remain open.
HC-09-03 remains open; this governing decision does not close the full checkpoint.

## Q-072 / ENFORCEMENT-006 — explicit collection boundaries

Status: **AGREED.** The user answered “yes” to the Finance-only grant and two-request
explanation. The request itself establishes the collection boundary through
endpoint-declared inputs, without inferring a narrower boundary from grants.

Previous status, retained as history: **PROPOSED, not approved** until that answer.

For a caller with the endpoint's required permission scoped to Finance, within
the same trusted tenant:

| Requested collection | Agreed result |
|---|---|
| `GET /api/v1/{tenant}/departments/FIN/certificates`, explicitly requesting Finance | Allow if the remaining required checks succeed; the handler queries only the evaluated tenant and Finance department. |
| `GET /api/v1/{tenant}/certificates`, defined here as requesting all tenant certificates | Deny with Finance-only authority; do not rewrite it into a Finance request. |

These are illustrative endpoint contracts, not a new route standard. Each
endpoint still declares exactly one permission and its required material. For
this example, both use the same registered listing permission. The broader
endpoint has no undeclared department filter or implicit “my accessible rows”
meaning. Tenant-wide authority could authorize that broader request, subject to
the other required checks; no multiple-grant coverage algorithm is selected.

Rationale: request-bound data constraints are already an endpoint responsibility;
discovering caller-specific constraints from grants would add the intelligence
the user rejected. Records outside an explicitly requested Finance collection
should not by themselves make that Finance request unauthorized. Conversely,
finding only Finance records today must not establish authority for an all-
departments request.

**Q-072:** Is this the intended distinction: an explicitly Finance-bounded list
can be allowed, while an all-departments list is denied for Finance-only authority?

**Answer: yes.** The endpoint declares department material from the path, supplies
it to the evaluator, and after allow constrains its query to the same evaluated
tenant and department. It does not inspect grants to discover accessible
departments. This settles request-bound collection behavior, not general query
generation, transport schemas, or the remaining collection checkpoint.

Continue horizontally with Q-073 in [bulk enforcement](bulk-enforcement.md).

## Original Q-071 authorized-subset proposal — not adopted, retained as history

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
