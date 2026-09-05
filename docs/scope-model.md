# Working handbook chapter: scope and target

Q-025 authorizes settling the shared scope model before returning to
administrative grants. [The decision log](handbook-roadmap.md) records agreement
status; [the discussion tree](discussion-tree.md) preserves the return to
ADMIN-004/005 and the other unfinished grant branches. This chapter is working
design, not a finalized scope schema or implementation claim.

## Discussion sequence

1. Separate the meaning of a scope from values supplied by a grant; decide how
   scope meanings are defined and recognized.
2. Establish the minimal representation and justify every field, including
   references, self relationships, and implicit tenant context.
3. Define composition: multiple references, dimensions, alternatives, and
   intersections; address missing, empty, invalid, and incompatible scopes.
4. Define resolution and enforcement using trusted facts, including reads,
   collections, creation, updates, moves, and dependent human/proxy context.
5. Test the shared model against ordinary business grants and administrative
   grants. Scope matching and scope containment are distinct problems; do not
   treat containment as implemented by giving it a field name.

These steps do not approve a syntax. PROCESS-005 requires justification and
reuse checks before any field is adopted. Cross-stage dependencies, including
fact freshness and scope changes over time, remain explicit in the tree.

## Agreed foundation

SCOPE-001: scope describes the resource set over which a grant may apply, using
explicit references or trusted attributes/relationships. PERMISSION-001:
permission identifies the operation and resource type; scope determines reach.
TENANT-001: the enclosing trusted tenant remains an implicit enforced boundary.
SELF-001: self resolves for the human being authorized, including group-derived
access; a proxy remains anchored to its authorizing human and delegation limits.

When reading payslip P-17 under a Finance department scope, P-17 is the protected
target. Finance is a reference used to determine whether that payslip falls
within scope. The request is not thereby an operation on the department itself.
Whether a payslip belongs to Finance requires the appropriate trusted facts;
the presence of FIN in the URI does not establish that relationship.

## SCOPE-002 / Q-026 — declared scope meanings, proposed

The recommendation is that a grant selects an explicitly defined scope type and
supplies only the inputs that definition permits. A grant does not invent the
scope's meaning or add arbitrary interpreted fields. Definition, supplied
values, and evaluation result are conceptual distinctions, not a proposal for
three new canonical records or extra fields.

Examples already used in the handbook are sufficient to examine this:

```json
{ "type": "department", "id": "FIN" }
```

In the motivating payroll example, the declared meaning determines which
relationship makes a payslip belong to a department. FIN supplies the chosen
department; it does not define that relationship. The meaning must be explicit
for compatible resource types rather than guessed from a matching field name.

```json
{ "type": "employee_self" }
```

The definition selects payslips belonging to the human's employee identity using
trusted relationships. There is no caller-selected employee ID: the existing
SELF-001 rule supplies the human reference. Exact data contracts remain open.

These reuse earlier illustrative syntax. Neither type/id nor the scope type
names become finalized wire schema merely through repetition here.

| Approach | Benefit | Cost or open concern |
|---|---|---|
| Explicitly defined scope types with validated inputs (recommended) | Makes accepted inputs and meaning reviewable; grants choose values rather than author new rules. | Requires definition governance and an approach for composition/new business relationships. |
| A general scope-expression grammar | Can express many combinations through shared operators. | Requires specifying typing, operators, validation, complexity limits, and safe evaluation before use; the existing examples do not establish that grammar. |

These need not be mutually exclusive forever. The current question is the
grant-authoring contract, not whether implementations may internally share
generic expression machinery. No arbitrary code or database query execution is
proposed in either approach.

The original lab concept page, section 5, also describes typed scope descriptors.
That is useful prior design material, not evidence that its catalog, ownership,
registration, or resolution claims have been reconciled and agreed here.

If SCOPE-002 is accepted, the next discussion must still settle who defines
scope types, their resource/permission compatibility, reusable versus
application-specific definitions, and the minimum parameter representation.
It does not authorize introducing a special scope type for every administrative
combination or assume that existing department/self forms already suffice.

## Return-point tests, not adopted syntax

- Read a particular certificate or payslip within an explicit selection.
- Read Finance resources versus the requesting human's own resources.
- Keep scope semantics intact across group-derived and human-proxy access.
- Create a grant for an eligible recipient with permitted capabilities/reach,
  without introducing an unreviewed second authority format.
- Establish when a proposed grant's scope is contained within assignable reach;
  testing one resource's membership does not prove general scope containment.

Return to ADMIN-004/005 after these scope questions are settled. The withdrawn
grant_selector, permissions_subset_of, and scope_within fields remain withdrawn.
