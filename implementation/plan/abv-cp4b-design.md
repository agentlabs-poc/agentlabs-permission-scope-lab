# CP4-B01 — assignment enable/disable design

Status: CP4-B01 approved for bounded implementation by the user's “proceed”.
Runtime work is in progress, not yet delivered. The governing
model is already approved. This document chooses a bounded implementation scope
and interfaces, not a new canonical assignment format or new authorization policy.

## Scope and source

Add protected enable/disable of existing **team-held assignments**, through the
same facade, two gates and SQLite transaction used by CP4-A. Tenant/application
remains mandatory outer context. Preserve grant identity, adopted revision,
recipient and all other records. No deletion, parent changes, upgrades, root
management, automatic cascading writes or production Auth integration.

Sources: [Q-101 binding lifecycle](../../docs/parent-grant-bindings.md),
[operation obligations](../../docs/authority-boundary-validation.md#6-operation-coverage),
[Q-102–106 adoption](../../docs/grant-revisions.md) and
[approved CP4 scope](abv-implementation-plan.md#cp4--lifecycle-and-structural-writes).

## The important difference from CP4-A

![Existing canonical four-part binding](../../docs/assets/parent-grant-bindings.svg)

In the existing lab, A1 assigns G1 to Team1; A2 assigns child G2 to child Team2.
After A2 is created, disabling A1 must fail while A2 remains enabled. Disable
A2 first, then A1. Enabling A2 while required A1 is disabled must fail. Restore
A1 explicitly, then A2. Every call needs its own current administrative check.

Grant disablement or expiry does not substitute for explicit assignment
disablement in the structural guard. This follows Q-101's bottom-up rule, not
the effective-access result. An enabled dependent binding still counts even if
its grant is disabled or its authority is otherwise ineffective.

Disabled bindings remain in the inventory. Traverse affected branches through
them to detect cycles and any lower enabled bindings; do not use an already
inconsistent disabled-parent/enabled-child arrangement to evade bottom-up checks.
Other assignments of the same grant at unrelated teams are not automatically
children. Membership and owner relationships do not identify these bindings.

![Existing bottom-up and revalidation flow](../../docs/assets/binding-change-lifecycle.svg)

Re-enablement validates the assignment's own exact adopted content against its
actual current parent-team support. It does not pick latest, restore a grant
control or enable another assignment. Explicitly disabled descendants remain
disabled. Runtime still refuses every descendant lacking its complete support.
As in CP4-A, restoration is not permanently tied to the historical issuer's
membership; the current administrator and actual continuing support are separate.

## Recommended implementation scope and alternatives

Recommend full reverse-binding checks for supported team-held routes, including
branching and shared grants. A leaf-only first implementation would be smaller
but defer the core bottom-up behavior this slice is intended to demonstrate.
General direct-human/proxy dependency discovery would broaden the slice into
unimplemented source-selection work; keep that outside CP4-B.

Do not silently skip uncertainty: if an enabled non-group assignment derives
from a grant in the affected team route and the implementation cannot establish
whether it depends on that holding, return unsupported with no write. This may
conservatively reject an unrelated direct-human route. It is a disclosed
prototype coverage limit, not a canonical claim that all such routes depend on
the team. Explicitly disabled non-group records still count for shape/uniqueness,
but supply no active binding. Disabling known team-held self-scoped bindings does
not require interpreting self; enabling them remains unsupported by this resolver.

## Exact interface proposal

Reuse the existing versioned Assignment as output. Take assignment ID and desired
status as ordinary Go arguments, then load the complete record **inside** Update.
No caller-side read is required; no new `AssignmentControl` JSON is introduced.

```go
// Both mutation.Service and abv.Facade.
SetAssignmentStatus(context.Context, domain.Area, domain.Identity,
    string /* assignmentID */, string /* status */) (domain.Assignment, error)

// Optional internal administration capability; public port uses Evidence alias.
type AssignmentStatusAdministration interface {
    CheckAssignmentStatus(context.Context, storage.Snapshot, domain.Identity,
        domain.Assignment /* exact proposed record */, time.Time) error
}

// Optional application seam; existing API stays unchanged.
type AssignmentStatusAPI interface {
    SetAssignmentStatus(context.Context, domain.Area, domain.FixtureContext,
        string /* assignmentID */, string /* status */) (domain.Assignment, error)
}
```

Successful output after disabling A2, within the implied selected tenant/app:

```json
{"version":"1","id":"A2","grant_id":"G2","grant_revision":1,"recipient":{"type":"group","id":"Team2"},"status":"disabled"}
```

Old assignment-create or grant-status adapters do not gain this new capability.
Missing/typed-nil capability fails unsupported. Same-state calls still pass the
ordinary administrative and relevant structural/enable checks. Failure returns
zero Assignment; success returns the exact committed record, only after commit.

## Core-philosophy check and outcome

- Explicit tenant/app and complete provider evidence prevent boundary fallback.
- Separate administration and ABV prevent possession or ownership implying control.
- Actual parent-team holdings, permission subset and scope AND remain unchanged.
- One conditional status write preserves adoption and prevents silent repairs.
- Bottom-up guards inspect stored binding state, not merely current effectiveness.
- Unsupported discovery fails explicitly; no new canonical restrictions are asserted.

The [execution draft](abv-cp4b-implementation-plan.md) separates persistence,
reverse-binding discovery, the protected operation and CLI acceptance. CP4-B01's
scope/interfaces are approved; each implementation task still requires tests and
independent review. Existing canonical decisions remain unchanged.
