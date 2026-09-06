# Background authorization — impact-first discussion

## Q-075 / ENFORCEMENT-009 — queued work needs execution-time authorization

Status: **PROPOSED, not approved.** This question concerns delayed execution of
human-authorized work. It is not a full non-HTTP adapter or queue contract.

### Example and recommendation

1. At 10:00, Vinay submits a Finance certificate export. Submission is authorized,
   and the job is queued; it has not yet read or exported the protected records.
2. At 10:05, Auth confirms revocation of Vinay's only grant supporting that export.
3. At 10:10, a worker picks up the queued job and is about to read the records.

Recommend a new authorization evaluation for the protected job operation when
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
selected by this proposal.

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
even if this governing timing rule is approved.

**Q-075:** Must queued work obtain execution-time authorization, rather than
reusing the allow obtained when the job was submitted?
