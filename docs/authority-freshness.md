# Authority freshness and revocation — impact-first discussion

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
