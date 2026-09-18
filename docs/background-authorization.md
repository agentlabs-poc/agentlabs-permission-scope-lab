# Background authorization — impact-first discussion

Terminology update after Q-082: the example's permanent grant revocation is now
the canonical **delete** operation. Execution-time authorization still applies;
the old wording is retained, not a separate grant lifecycle operation.

## Q-075 / ENFORCEMENT-009 — queued work needs execution-time authorization

Status: **AGREED.** The user answered Q-075 “agree” after the queued Finance
export example. Delayed execution requires authorization for the protected
operation at execution time; queue acceptance does not preserve earlier authority.
This is not a full non-HTTP adapter or queue contract.

Previous status, retained as history: **PROPOSED, not approved** until that answer.

### Agreed example and outcome

1. At 10:00, Vinay submits a Finance certificate export. Submission is authorized,
   and the job is queued; it has not yet read or exported the protected records.
2. At 10:05, Auth confirms revocation of Vinay's only grant supporting that export.
3. At 10:10, a worker picks up the queued job and is about to read the records.

Require a new authorization evaluation for the protected job operation when
the worker starts it, using the applicable current authority and trusted job
material. In this example, no supporting authority remains: deny execution and
do not produce the export. Do not treat the submission's stored allow as ongoing
permission to execute later.

### Rationale and alternative

Queue acceptance is not a durable grant. Treating it as execution authority
would allow delayed work to bypass later authority removal and would effectively
freeze a proxy's authority independently of its human. The alternative, retaining
submission-time authority until completion, makes scheduling more predictable but
conflicts with the chosen current-human-subset model unless separately designed
and approved; it is not adopted here.

The trade-off is that an accepted job may subsequently be unable to execute.
If current authority cannot be established, preserve the existing evaluation-
error distinction and block protected execution; a timeout is not proof of denial.
No automatic cancellation, queue deletion, retry, or job-status vocabulary is
selected by this decision.

### Ownership and limits

The worker's application adapter supplies trusted tenant, authorizing-human/proxy
context, the operation's permission, and its declared material to evaluation;
the application enforces the resulting request boundaries. Merely copying a user
ID from untrusted job content is not trusted attribution. A worker cannot use
unrelated broader authority to bypass the submitting human's limits. No independent
service-account authority or new delegation is created merely by queuing work.

Submission and delayed execution are distinct operations, not middleware and
endpoint halves of one prepared decision. This does not revive the deprecated
prepared state or mandate a remote Auth call for every execution: authority
loading must meet the agreed freshness requirements, whatever mechanism is used.

Exact job/adapter schemas, identity binding, submission-versus-execution permission
mapping, retries, recurring schedules, already-running jobs, streaming exports,
result delivery, and job lifecycle reporting remain open. HC-09-09 remains open
despite approval of this governing timing rule; adapter integration is unfinished.

**Q-075:** Must queued work obtain execution-time authorization, rather than
reusing the allow obtained when the job was submitted?

**Answer: agreed.** The example, rationale, alternative, and trade-off above are
part of the record. The subsequent Q-076 [audit detour](authority-change-audit.md)
was excluded by the user as another layer's responsibility. Return to Q-077 in
[groups and membership](groups-and-membership.md). Running-job and retry details
remain tracked rather than implicitly approved.

## Q-142 / ENFORCEMENT-011 — non-HTTP and background integration is deferred for v1

Status: **AGREED, as an explicit deferral.** The criterion covering this asked
for integration requirements *or an explicit deferral*, and the user chose the
deferral. Recorded as a decision so that silence is not later read as an
oversight.

Everything this handbook settles assumes a synchronous request arriving at one
endpoint-owned gate, with the effect bounded by the evaluated material. Queues,
scheduled jobs, streams and long-running work are **out of scope for v1**.

This does not weaken what is already settled about them.
[Q-075](#q-075--enforcement-009--queued-work-needs-execution-time-authorization)
stands: queued work is authorized when it executes, not when it was enqueued.
[Q-129](concurrent-enforcement.md) stands, and its own limits are already stated
— it covers an ordinary synchronous operation already allowed, and explicitly not
queues, streams or long-running cases.

**What a deployment doing this anyway must not assume:** that an allow travels
with the work. It does not. An evaluation is about a request, a boundary and a
moment; carrying its result forward to a later execution is the thing Q-075
refuses.

### Rationale / conscious tradeoff

Specifying it now would mean designing freshness, cancellation and
re-authorization for execution paths nothing here has built or demonstrated —
the same objection that keeps caching out until an epoch exists to invalidate
against. A deferral that says so is more useful than a specification nobody has
run.

The cost is real and worth naming: a deployment with background work has no
guidance beyond Q-075's single rule, and will invent the rest. That is preferable
to inventing it here, where nothing could check it.

