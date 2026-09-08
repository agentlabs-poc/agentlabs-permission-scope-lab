# 8. Lifecycle and worked scenarios

[Contents](../README.md) · [Previous](07-application-integration.md) · [Pending items](../appendices/pending.md)

**Framework-specific behavior.** The following scenarios explain approved rules
and their limits. They are expected outcomes for review, not results from an
implemented Auth engine or an exhaustive security test suite.

## Publication, adoption and resolution are different events

A published grant or role revision is immutable. Publishing newer content does
not silently change an existing assignment's adopted grant revision or a grant's
selected role revision. Authorized administrators can discover older assignments
and review suggested updates; the suggestion is advisory, not permission to adopt.

New assignments and explicit assignment upgrades must select the latest published
grant revision and pass current validation. Existing assignments may remain on
older content. If the latest content is not supported, reject the new assignment
or upgrade; do not quietly select an intermediate revision that happens to fit.

Parent resolution follows actual adopted support from top to bottom. Team1's G1
revision 2 is the ceiling for Team2's route while A1 adopts that revision. A newer
revision merely published elsewhere—or adopted by TeamX—does not substitute for
Team1's required support. Re-enabling an assignment is not an upgrade: it retains
its adopted content while validating current reality.

Parent adoption may change inherited scope even when child scope remains `{}`.
Therefore validate the complete resulting authority and affected bindings;
unchanged child JSON is not evidence that its effective boundary is unchanged.
The computed-root catalog behavior described in Chapter 6 is a specific exception
to ordinary explicit permission selection, not a reason to make all revisions live.

## Controls, time and effectiveness

| Event | Stored state | Effective consequence |
|---|---|---|
| Disable G2 | G2 becomes disabled; A2 and descendants need not be rewritten. | No assignment can obtain authority through G2. |
| Disable A2 | A2 becomes disabled; G2 remains enabled. | That recipient route stops; unrelated assignments are not globally disabled. |
| Disable parent G2 of enabled G3 | G3 can remain stored enabled. | G3 is ineffective through missing usable support. |
| Validly re-enable G2 | Explicit enablement succeeds only after current checks. | Still-enabled, otherwise valid G3 may work again. |
| G3 itself was explicitly disabled | Its disabled state remains. | Parent restoration does not enable it. |
| Delete a grant | Permanent withdrawal from usable authority. | It cannot be enabled back; delete is the selected permanent operation, not a separate revoke state. |

Grant validity belongs to immutable revision content. This independent core
example has a local expiry; it still requires valid upstream support and an
authorized assignment:

```json
{
  "version": "1",
  "grant_id": "G-TEMP-READER",
  "revision": 1,
  "parent_grant_id": "G1",
  "permissions": ["hrms:employee:certificate::read"],
  "scope": {"cert": "C17"},
  "validity": {"expires_at": "2026-09-30T00:00:00Z"}
}
```

Optional `not_before` is inclusive; `expires_at` is exclusive. No local window
means no additional local time limit, not exemption from upstream limits.
Extending or removing a window requires new content and explicit adoption.
Enabling an expired revision does not reset its clock. Assignment-specific time
windows are explicitly deferred in v1; do not add them by analogy.

## Orphans and structural changes

When required parent support is absent, the affected dependent lineage is
ineffective even if its definitions and assignments remain stored. A warning or
preventive workflow above resolution can help avoid creating orphans, but runtime
resolution must still refuse unsupported authority. A timeout is not established
absence, and an ordinary reusable definition with no recipient is not automatically
an orphan. A legitimate root does not need an invented parent.

Structural changes have specific guards. An affected enabled assignment binding
prevents changing its bound parent underneath it. Handle affected assignments
bottom-up by explicit disablement or removal, inspecting every affected branch
and shared use. Temporary ancestor disablement and derived ineffectiveness do
not substitute for the required structural handling.

After a permitted change, retained disabled assignments stay disabled. Explicit
re-enablement checks new parentage, current support, adopted content and complete
boundaries. Former support at Team1 does not justify enabling a route now requiring
TeamX support. Self-parenting and ancestor cycles are invalid even while records
are disabled. Historical revision edges are not blindly merged into one graph;
the relevant proposed/adopted lineage is what must be validated.

## Freshness and operations already in progress

New checks after a confirmed authority reduction cannot use stale evidence to
retain the withdrawn support. This includes membership removal, grant/assignment
disablement, delegation withdrawal, adopted narrowing and effective permission
retirement. A cache lifetime is not a grace period. Another complete valid route
may still authorize; inability to establish required freshness is not proof of
a policy denial.

An ordinary synchronous application operation that already received allow before
later withdrawal may finish within the evaluated boundaries. This does not permit
new data, retries or unrelated operations to reuse that allow. If the actual
data boundary changes, it must not escape the check. Auth's authority-changing
writes retain their stronger consistency-through-persistence requirement.

Still-evaluating requests, long-running work and streams are not silently covered
by that synchronous-completion rule. Queued work requires execution-time
authorization rather than unconditional reuse of enqueue-time approval. The
remaining ordering, streaming and recurring-work contracts stay pending.

## Scenario A — employee self access in HRMS

An Employees group can receive a grant whose scope uses `user: "$self"` for
payslip-read. Vinay's membership makes the group's valid route applicable, while
`$self` binds to Vinay. For Nutan's request it binds to Nutan. It never binds to
Employees as a group or to an agent instead of its authorizing human.

Illustrative immutable content, assuming valid supporting payroll authority,
registration and explicit assignment to Employees:

```json
{
  "version": "1",
  "grant_id": "G-EMPLOYEE-PAYSLIP",
  "revision": 1,
  "parent_grant_id": "G-PAYROLL-BASE",
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": {"user": "$self"}
}
```

The endpoint constrains actual payslips to the authorizing human. Group-based
distribution avoids separately defining an ordinary grant for every employee.
This established runtime meaning does **not** settle whether Maya can distribute
her own self-scoped source to Nutan by copying the token. That issuance-boundary
question remains explicitly pending.

## Scenario B — repository and ticket boundaries

The same construction can apply to a registered repository boundary. Assume the
required source supplies repository-write within Repo-A. A valid child selects
that permission and adds `branch: release` under an application-defined branch
relationship. Effective reach is Repo-A AND its release branch—not any branch
with that name in another repository.

The endpoint must bind the branch lookup and mutation to Repo-A. The name
`codehost:repository:branch::write` does not itself prove a branch relationship,
and a Team1 membership does not prove which repositories Team1 can write. The
grant, assignment, membership and actual-data constraints together explain it.

For ticketing, substitute an explicitly registered project boundary and a
ticket-update permission. A ticket ID requested under Project-A must not update
Project-B data through an ID-only lookup. No built-in project or subtree traversal
is implied by choosing these example keys. Exact hierarchy/subtree semantics are
application-definition work that the current framework has not universally fixed.

## Scenario C — accounting permission/scope separation

Assume individually valid parent support, controls and assignments for these
independent authority contents:

```json
{
  "version": "1",
  "grant_id": "G-ACCOUNTING-FIN-POST",
  "revision": 1,
  "parent_grant_id": "G-ACCOUNTING-BASE",
  "permissions": ["accounting:ledger:entry::post"],
  "scope": {"dept": "FIN"}
}
```

```json
{
  "version": "1",
  "grant_id": "G-ACCOUNTING-ENG-READ",
  "revision": 1,
  "parent_grant_id": "G-ACCOUNTING-BASE",
  "permissions": ["accounting:ledger:entry::read"],
  "scope": {"dept": "ENG"}
}
```

The pair does not authorize posting in Engineering. Posting comes from the FIN
route; ENG reach comes from a read route. Combining them would create unsupported
authority. Nor does permission to assign these grants automatically permit ledger
posting: administration and business execution retain separate checks.

This example does not define accounting approval workflows, posting periods or
financial validation. Those are business rules outside this handbook's model.

## Scenario D — a batch needs complete coverage

### First distinguish an explicit collection from automatic filtering

A Finance-only reader may request a listing explicitly bounded to Finance,
subject to the required listing permission and other checks. The handler queries
that same tenant and Finance boundary. Engineering records elsewhere in the
database do not by themselves invalidate this explicitly bounded request.

An endpoint defined to list all tenant certificates asks a broader question.
Finance-only authority must not turn it into a successful Finance-only result.
The evaluator rejects insufficient authority; the handler does not inspect
grants and silently rewrite the request. Finding only Finance rows today also
does not prove entitlement to an all-departments listing. This is the chosen
explicit-collection rule; general count, export and pagination contracts remain
pending rather than inferred from this example.

### Then evaluate every item in a finite batch

Suppose a delete endpoint has two complete applicable routes: FIN-delete and
ENG-delete. A finite batch containing a Finance item and an Engineering item may
use different complete routes for those items. Every item needs the endpoint's
same declared permission with its full scope and lineage support before effects.

If the ENG route is read-only, it cannot borrow delete from the FIN route. One
uncovered item denies the whole batch before protected effects; inability to
evaluate required evidence produces an error instead. The endpoint does not
silently delete an authorized subset or reconstruct grant combinations itself.

Authorization preflight does not guarantee that subsequent business execution
cannot fail. Transaction, conflict and partial execution behavior require their
own application contract; they are not permission to knowingly include an
unauthorized item. Batch material and exact result-evidence transports remain
pending even though this governing coverage rule is agreed.

## Scenario E — ownership rotation and dependent agents

Om replaces Maya as an authorized Team1 owner while Team1's grants, assignments
and actual supporting lineage remain unchanged. Team2's business authority does
not automatically change. Om's personal ENG-write grant is not imported into the
team. Future administration still checks the acting person's authority and source.
Exact ownership-transfer permission/records remain pending.

Separately, Vinay's read-only agent loses effective Finance access when required
human Finance support disappears. If that support returns and the delegation
itself remains valid, the access can become usable again. An explicitly withdrawn
or expired delegation is not revived by that restoration. This is a chosen
human-dependent lifecycle, not an independent machine entitlement.

## Scenario F — the unresolved direct-human context

Team1 holds G1 revision 1 with FIN-read; TeamX holds G1 revision 2 with ENG-read.
G2 refers to G1 and is directly assigned to Nutan. The local narrowing operation
works with either supplied parent, but the supplied records do not establish
which holding is eligible continuing support for that direct route.

Do not silently pick the newest holding, the first row, the issuer's history or
every holding in the tenant. The original direct-assignment survival example
used the same parent content at both teams; it did not settle this competing-
revision case. This first draft deliberately retains the open question instead
of inventing a field or declaring direct assignments unsupported.

**Chapter takeaway:** review both normal access and transitions. Explain which
records change, which effective routes stop or resume, and which dependencies
must still be proved. Keep unresolved cases explicit rather than extending a
nearby example beyond its assumptions.

**Sources:** [revisions](../../docs/grant-revisions.md),
[binding lifecycle](../../docs/parent-grant-bindings.md),
[validity](../../docs/grant-validity.md), [freshness](../../docs/authority-freshness.md),
[bulk](../../docs/bulk-enforcement.md), [delegation](../../docs/delegation-lifecycle.md),
[direct-human trace](../../docs/direct-human-parent-context.md).

[Return to contents](../README.md) · [Read the pending register](../appendices/pending.md)
