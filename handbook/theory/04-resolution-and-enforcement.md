# 4. Resolution, decisions and enforcement

[Contents](../README.md) · [Previous](03-relationships-and-delegation.md) · [Next: the concrete model](../implementation/05-canonical-model.md)

**Definitions and JSON alongside this chapter:**
[endpoint policy](canonical-terms.md#endpoint-declaration),
[request/material/resolved request](canonical-terms.md#request-and-material),
[resolved grants and resolution](canonical-terms.md#resolution-and-evaluation),
[allow/deny/error](canonical-terms.md#decision-and-enforcement).
Unapproved request/resolved-view envelopes are not invented to illustrate their meanings.

## From request claims to a justified effect

An incoming request usually contains claims: a document identifier, department
name, intended update or requested tenant. Those values describe what the caller
wants. They do not automatically establish who the caller is, who owns a record,
or which authority is available.

Authorization needs material appropriate to the question being asked. Some
comes from verified identity, some from authority records, some from application
facts, and some from constraints that execution will enforce directly. The
important distinction is not simply whether a value came from a path or body;
it is what that value establishes and what remains to be proved.

![General authorization reasoning from a request and established evidence to bounded execution](../assets/authorization-reasoning.svg)

This is a logical flow. It does not prescribe a central Auth service, a particular
middleware position, one network call per request or a database product.

## Request, resolved request and decision

A **request** identifies the intended operation and supplied inputs. A
**resolved request** is an evaluation-ready view in which the meanings and
required context have been established sufficiently for the checks being made.
Resolution does not mean that every imaginable application fact must be fetched.
It means that the evidence needed for the actual authorization question is
available or that execution is bound to enforce the necessary relationship.

Similarly, a **resolved grant** is an evaluation view of applicable authority,
including the restrictions and dependencies needed to interpret it. It is not a
new independent entitlement. Losing the support from which it was resolved can
matter according to the system's freshness and consistency rules.

A **decision** says whether the established authority permits the requested
operation. These three ideas should not be collapsed. Normalizing identifiers
does not itself produce an allow. Loading grants does not prove applicability.
Returning allow does not guarantee that later code obeys it.

## Two ways to establish a data boundary

Suppose a request asks for certificate C17 under Finance.

One approach is to load trustworthy information about C17, establish its actual
department, evaluate against that fact, and preserve the checked relationship
through use. Another approach is to evaluate the caller's authority within a
Finance boundary and constrain the actual lookup to both Finance and C17. In
the latter case, an Engineering certificate cannot be returned through that
lookup.

Both require the check and the effect to refer to the same intended data. A
route parameter saying Finance followed by an unrestricted C17 lookup is not
the second approach; it leaves the relationship unenforced. Fetching Finance
facts and then mutating a different row is not the first approach either.

This distinction lets application integrations stay practical. Authorization
need not require an eager extra database query when the operation itself can
enforce the boundary. But avoiding that query does not make enforcement optional.

## Completed rejection versus inability to evaluate

When sufficient evidence establishes that no complete applicable route permits
an operation, evaluation can complete with a rejection. When required evidence
cannot be obtained, the system has not established the same fact. An authority
service timeout is an example of an evaluation failure, not proof that the user
has no grant.

Both situations must prevent unchecked protected execution. They can still have
different diagnostic and operational meanings. A caller may be able to correct
an entitlement problem; a service failure may require operational recovery.
The [Foundations reference](canonical-terms.md#decision-and-enforcement)
explains the repository's minimal result JSON; Part II explains endpoint handling.

A diagnostic code catalog is worth one design note, because it is usually decided by
accident. A **closed** list lets a consumer be exhaustive, which is the reason to
want one — and it makes every newly discovered failure mode a contract version, since
failure modes are learned by the layer that meets them rather than by the document.
An **open** list avoids that and costs exhaustiveness: no consumer can handle every
code, so each must fall back to the *class* the result arrived in. Part II chooses
open, with published names fixed forever, and accepts the consequence — a consumer
that logs an unrecognized code without alerting on it will swallow a new failure mode
silently. Either way, the code explains; the class decides.

A failed candidate route also does not establish a final rejection if another
complete valid route authorizes the operation. Conversely, a successful partial
check cannot authorize an operation whose mandatory requirements remain unmet.

A third case sits between the two and is the one most often collapsed into a
rejection: **a route the evaluator cannot read.** Malformed, internally
inconsistent, carrying identifiers it cannot parse — that is precisely a route that
*might* have authorized. So it costs that route, and it costs the certainty of a
rejection, but not the whole answer:

| Situation | Result |
|---|---|
| A route is unusable, and another complete route authorizes | **Allow.** The unusable route says nothing about the one that did. |
| A route is unusable, and no other route authorizes | **Failure to evaluate.** The rejection is not established. |
| Every route is readable, and none authorizes | **Completed rejection.** |

This matters more, not less, as an authority answer grows: if the answer describes
everything a subject holds, then failing the whole evaluation on one unreadable entry
takes away every other authority they have. Reporting a failure rather than a
rejection does tell a person "we could not check" when the honest answer might have
been "you have no access" — and that is the correct trade, because a rejection
asserts something about their authority, and asserting it from evidence that was
never read is the error being prevented.

An answer describing a different subject or a different boundary is not one unusable
route. It means the answer is about somebody else, and nothing in it may be used.

## Time belongs in the reasoning

Authority, membership and application data can change while a request is being
processed. A useful design names its relevant ordering points: when evidence was
established, when a decision was made, when a withdrawal became effective and
when the protected effect occurred.

“Use a cache” does not define the freshness guarantee. “Use a transaction” does
not identify which dependencies must be protected. Those mechanisms need a
stated semantic contract. A system may distinguish an operation already allowed
from a new check after withdrawal, and may require stronger consistency for
authority-changing writes than for an ordinary read.

Queued work presents the same issue over a longer interval. Permission to submit
work now is not automatically proof of authority to execute it later. Recurring
jobs and long-running streams need an explicit rule about which subsequent
effects remain covered; a single old allow is not self-explanatory evidence.

Part II holds one rule here and **defers the rest explicitly**: queued work is
authorized when it executes, not when it was enqueued, and everything else about
queues, schedules and streams is out of scope for its first version. A deferral that
says so is more useful than a specification nobody has run — and it is honest about
the cost, which is that a deployment with background work has one rule and will
invent the remainder.

## Evidence should explain the route, not replace it

Useful decision evidence lets the recipient understand what was decided and,
where required, which authority contributed. Evidence may include grant
references or supporting context. It must not become an unrestricted reusable
capability accidentally detached from the evaluated request.

Providing evidence is also different from designing a complete audit system.
An audit consumer may record authorization outcomes, but retention, storage,
delivery and disclosure policy are separate responsibilities in this handbook.
Specifying the *producing* half without the consuming half does create one
obligation, though: whoever builds the recorder must be told what it may rely on,
or it will infer its requirements from an implementation — which is how a layering
decision turns into an accident.

There is one further duty that belongs to the endpoint and not to the gate, and it
is easy to miss because no rule is broken when it fails. Consider a caller holding
write and not the matching read who issues an update: the write is authorized and
succeeds, and the response carries the whole record — fields the caller never
supplied and now learns by writing. A gate cannot prevent that. It does not know what
a body contains and should not be given the job of finding out. **What a response
discloses is the endpoint's duty**, assigned by the same split that gives the
endpoint execution: it keeps its output, as well as its effect, inside the authorized
boundary. One layer cannot be made answerable for the correctness of every layer
above it, and attempting it is how a gate accretes obligations that look like safety
and are not.

**Chapter takeaway:** resolution prepares justified meaning; evaluation decides
applicability; enforcement constrains the actual effect. A sound integration
keeps the connection among all three visible.

[Continue to Part II: canonical records and identity](../implementation/05-canonical-model.md)
