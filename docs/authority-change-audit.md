# Authority-change audit — impact-first discussion

## Q-076 / AUDIT-001 — mandatory records for authority administration changes

### Current conclusion — outside handbook scope

The user answered: “This itself is different layer question. handbook as no scope
for this.” The mandatory audit proposal is **NOT ADOPTED**. Audit policy and
system design belong to another layer, not to this canonical authorization
handbook. This is a scope exclusion, not a deferred handbook decision or a
completed audit contract.

Rationale: defining who may perform an operation and how authority is evaluated
is distinct from selecting audit events and specifying their capture, storage,
delivery, retention, and disclosure. The earlier proposal crossed that boundary.
Those questions must not be used as outstanding requirements for finishing this
handbook. HC-09-08 is retained as EXCLUDED, not marked DONE.

This does not remove the agreed result evidence such as supporting grant IDs,
or the requirement to authorize grant/membership/role/delegation changes. Earlier
explicit-self-assignment audit references remain historical/integration context;
they do not expand this handbook into an audit-system specification. No claim
that auditing is unnecessary, optional in every system, or forbidden is made.

The whole original proposal follows as history; its recommendations and open
questions are not current handbook requirements.

### Original proposal — not adopted, retained as history

Status: **PROPOSED, not approved.** Explicit authorized self-assignment already
requires audit under ADMIN-003. Q-060 does not require every access request to
be persistently logged. This proposal distinguishes authority administration
changes from ordinary access evaluations without reversing either decision.

### Example and recommendation

Maya adds Vinay to the Finance group. Vinay gains access through its existing
grants even though no grant was created or edited. Later, Maya removes him.
Logging only grant edits would miss both changes to his membership-derived access.

Recommend mandatory audit records for successful authority-administration
mutations involving grants, authorization-group memberships, live roles, and
delegations. This covers creating, modifying, and revoking/removing those
relationships or records as applicable. It includes additions and removals,
direct and group-based routes, and authorized changes made through synchronization
or automation; it is not restricted to self-assignment or manual UI actions.

For the example, retain evidence that Maya added Vinay to Finance and later
removed him, in the relevant tenant, with the timing and change identifiable.
These are explanatory contents, not new canonical field names or a wire schema.
Human/proxy attribution must not collapse into an unexplained worker identity.

### Rationale, alternative, and trade-off

Access may change through membership, role, or delegation administration without
editing the final grants used by a request. Auditing these changes supports
accountability and investigation of how authority was gained or lost. Optional
authority-change logging would leave that history incomplete; auditing only
self-assignment would miss changes an administrator makes for someone else.

The cost is mandatory evidence capture and storage for these administrative
mutations. This does not require logging every read, every evaluator call, or
every dependent resolved-grant recomputation. Recording a role or membership
mutation does not itself require one duplicate event per affected user's derived
authority. Audit is evidence, not a substitute for authorizing the mutation.

### Remaining decisions

This proposal selects which administrative mutations require audit, not the
complete audit guarantee. Atomic capture, storage or delivery failures, integrity,
retention, access to audit data, event versions, before/after representation,
supporting authorization evidence, and correlation remain open. No best-effort
exception or specific synchronous remote logging requirement is adopted.

Denied/failed administrative attempts, catalog or configuration changes,
application-domain changes that affect boundaries, and time-driven expiry events
need separate coverage decisions; they are not silently included or excluded
from the eventual audit contract. HC-09-08 remains open even if this principle
is approved.

**Q-076:** Should successful grant, membership, role, and delegation administration
changes always be audited, while ordinary access-request logging remains selective?

## Q-141 / AUDIT-002 — what an audit consumer must be given, without specifying one

Status: **AGREED.** Audit stays excluded from this handbook — Q-076 placed it in
another layer and that is unchanged. This adds only what a consumer in that layer
must be able to rely on, because the *producing* half is now specified here and
the two were drifting apart.

[Q-060](decision-results.md) requires an allow to carry references to the routes
that justified it, and [Q-134](decision-results.md) fixes what those references
are. So this handbook specifies a producer for a consumer it excludes. Whoever
builds the consumer would otherwise infer its requirements from an
implementation, which is how a layering decision turns into an accident.

**What a consumer is given, and may rely on:**

| | |
|---|---|
| The contributing chain | Ordered root first, per Q-134. The order is the dependency. |
| Availability | Present on every allow, whether or not the request is recorded — Q-060. |
| Timing | Handed to the effect *before* it runs, so a record can name the change it authorized. |
| Nothing else | No scope echo (Q-061), no revision, assignment or group per step (Q-134). |

**What a consumer must not assume:**

- **That grant ids stay resolvable.** They name records that may later be
  deleted. A consumer needing to reconstruct a decision after the fact must
  capture what it needs at the time, not resolve the ids later.
- **That the evidence is a complete account of the decision.** It names the route
  that authorized, not the routes considered, not the boundary evaluated, and not
  the material the endpoint bound.
- **That producing it implies recording it.** Q-060 is explicit that evidence is
  available regardless of whether a request is selected for recording.

### Rationale / conscious tradeoff

The alternative was to reopen audit as a criterion here. It was rejected for the
reason Q-076 gave: retention, correlation and disclosure belong to a layer this
handbook does not govern, and drafting them here would make this document
answerable for a system it cannot see.

The conscious cost is that this chapter describes an interface with no consumer
in view, so nothing tests whether what is handed over is sufficient. The first
real audit implementation may find it is not — and the honest place to discover
that is against a real recorder, not in a chapter written in advance of one.

