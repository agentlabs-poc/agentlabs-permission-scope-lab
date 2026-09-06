# Authority freshness and revocation — impact-first discussion

**Q-129 approved:** [an already-allowed ordinary synchronous application operation may finish](concurrent-enforcement.md)
within its evaluated boundaries despite a later authority withdrawal. This is
not stale authority for new checks or retries. Q-074 and Q-110 remain required;
queued, streamed, long-running, and not-yet-allowed cases are not covered by it.

## Q-128 — All confirmed authority reductions (approved)

**APPROVED.** The user agreed to extend Q-069's no-stale-authority guarantee
beyond grant deletion to all confirmed authority reductions. Authorization
checks started after Auth confirms the reduction must not rely on the withdrawn
authority from an older cached or resolved representation.

Examples include membership removal, grant/assignment disablement, withdrawal of
a delegation, and adoption of a narrower boundary. Effective permission retirement
under Q-125 likewise cannot be bypassed by a cached older source. Publishing an
ordinary narrower role/grant revision without adoption is not itself withdrawal
of an assignment's current authority; existing explicit-adoption rules still apply.

```text
Auth confirms Vinay's Finance membership removal
                         ↓
A new endpoint authorization check starts
                         ↓
Cached Finance membership cannot supply that access
```

A different complete valid authority route may still allow the operation. The
rule removes the withdrawn support, not every other grant the human holds.
If sufficient freshness cannot be established, old evidence alone cannot justify
allow. Failure to establish evidence remains an evaluation failure, distinct from
a policy denial based on established facts.

**Rationale:** membership withdrawal, assignment disablement, or delegation
withdrawal must not have weaker effects than deleting a grant. The live dependency
and subset rules must hold across cached and resolved representations, not only
inside Auth's current database state.

**Core-philosophy check:** no cached copy becomes independent authority, and no
old membership or parent boundary can silently survive a confirmed reduction for
new checks. Ordinary valid alternatives remain available, while fail-closed
handling prevents uncertain evidence from being promoted into allow.

**Operational cost:** implementations need coordination sufficient to establish
freshness. They may need to delay confirmation or refuse evaluation when they
cannot establish it. A local TTL does not create a grace period after confirmed
reduction. Confirmation here is that the change took effect under this guarantee,
not merely that a request was received or queued.

No Auth call per request, invalidation transport, cache protocol, confirmation
field, new error code, or distributed transaction is prescribed. These mechanisms
must satisfy the guarantee; the documentation is not runtime verification.
Automatic validity expiry remains governed by the existing time-bound rules and
does not require an administrative confirmation to become effective.

Checks already in flight, previously allowed work, and exact acknowledgment/
ordering/evidence contracts remain separate. This approval does not cancel or
roll back business operations, and does not relax Q-110's stronger checked-write
guarantee for Auth's own authority mutations or Q-074's application-boundary rule.

Earlier membership/delegation propagation-policy gaps below are narrowed by
Q-128. The original Q-069 rationale and cache alternatives are retained as history;
the old grace-window alternative is not reopened for these additional reductions.

Terminology update after Q-082: the canonical permanent-removal operation is
**delete**, not a separate revoke operation. The historical revocation wording
below means permanent grant withdrawal. Its agreed timing guarantee still holds:
after Auth confirms deletion, new checks cannot use that grant via stale cache.
See [grant lifecycle](grant-lifecycle.md). No cache protocol or in-flight timing
rule is newly selected by this terminology reconciliation.

## Q-069 / FRESHNESS-001 — revoked grants and new authorization checks

Status: **AGREED.** The user answered “yes” to Q-069. Original status retained
as history: ~~PROPOSED, not approved~~. Existing rules keep resolved grants
dependent on their source authority; this decision establishes the timing
boundary for confirmed grant revocation, not a new grant or result format.

### Concrete example

1. The application has cached G-17, which grants Vinay Finance certificate-read
   authority.
2. An authorized administrator revokes G-17, and Auth confirms the revocation
   has completed.
3. A new endpoint authorization check starts afterward for Vinay's next request.

This new check **must not authorize through G-17**, even if an old copy is still
cached. This defines the agreed guarantee at the boundary between
a completed revocation and authorization checks started afterward. Checks already
in flight, or operations already authorized, are separate concurrency questions.

The rule removes this authority route, not every route the human may hold. A
different complete valid grant can still authorize the request. If sufficiently
fresh authority cannot be established, an old copy alone cannot justify allow;
the existing fail-closed evaluation-error distinction applies. A timeout does not
prove that no other grant exists.

### Rationale and alternative

The decision gives revocation a clear effect for new authorization checks:
an administrator does not receive confirmation while an unacknowledged cache
window continues to supply the revoked authority. It preserves the existing
source-dependency requirement across cached and resolved representations.

The alternative is a documented bounded-staleness window: cached G-17 could
continue authorizing until a known deadline after revocation. That can reduce
coordination and improve availability, but deliberately retains revoked access
for that window. That alternative is not adopted for new checks after confirmed
grant revocation.

The chosen guarantee also has a cost: implementations need a way to
establish freshness, and may have to stop when they cannot. It does not mandate
one Auth call per request, prescribe a cache invalidation/versioning mechanism,
or claim that the current runtime provides this guarantee. The implementation
must justify any cache use against this timing contract. A cache entry's local
expiry time alone cannot provide a grace period after confirmed revocation.

Counterexample: if G-17's cached copy still has five minutes left before local
expiry, it nevertheless cannot authorize a check started after Auth confirms
revocation. This does not prevent authorization through a different complete
valid grant. Failure to establish fresh authority remains an evaluation failure,
not proof of policy denial.

### Limits and return points

Confirmation semantics, concurrent checks, already-authorized operations,
membership/role/delegation change propagation, and the exact cache protocol remain
open. Approval of this principle alone does not close HC-09-06 or HC-09-07.
Do not interpret this as canceling or rolling back completed business operations.

**Q-069 — answered yes:** once Auth confirms grant revocation, checks started
afterward cannot use that grant, with no stale-cache grace period. Next in the
impact-first pass is [delegation reactivation](delegation-lifecycle.md), not
cache-protocol field design. Already-in-flight work stays tracked separately.
