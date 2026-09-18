# Registered permission lifecycle — Q-125

**Q-128 follow-up:** [all confirmed authority reductions have no stale-cache grace for new checks](authority-freshness.md).
Effective retirement cannot be bypassed by an older catalog/root representation.
Exact confirmation/propagation mechanics remain open; grant records are unchanged.

## Q-126 — Stable authorization meaning (approved)

**APPROVED.** An existing permission identifier must not be repurposed for a
materially different authorization meaning, including after retirement. A
different operation requires a new identifier. Descriptions and display labels
may be corrected without changing the operation's authorization meaning.

For example, `hrms:payroll:run::approve` means approving a payroll run. Changing
that same identifier to also authorize transferring money to employees would
silently expand existing authority. Payment execution needs a separate permission,
such as `hrms:payroll:run::pay`. These names illustrate the approved grammar;
they do not register new permissions or assign them to any real recipient.

**Rationale:** immutable grant content and adopted role revisions retain
permission identifiers. They would not protect against silent privilege changes
if the identifier could later stand for a different operation. Stable meaning
preserves what the administrator selected and what the recipient actually received.
Retirement does not make the name available for an unrelated operation.

**Core-philosophy check:** this preserves explicit authority selection, role/grant
revision intent, no implicit aliases, and application ownership of domain meaning.
Auth can preserve identifier history and validate registration contracts, but
does not infer business semantics or prove that new application code implements
the declared operation. The application platform administrator remains responsible
for the definition's meaning; application integration must enforce that meaning.

**Trade-off:** a materially different operation requires explicit naming and
grant/role selection instead of silently inheriting old grants. Cosmetic wording
corrections do not require inventing another permission identifier. The rule is
not a ban on bug fixes or internal implementation changes preserving the same
authorization meaning.

No new permission field, semantic-comparison engine, automatic grant migration,
or alias is introduced. This settles reuse for a *different* meaning; it does
not silently approve same-meaning restoration of a retired permission or settle
scope-definition evolution and exact lifecycle API contracts.

The earlier identifier-reuse gap below is narrowed by Q-126; remaining restoration
and lifecycle questions are distinct from repurposing an identifier.

## Q-125 — Retirement with existing grant references (approved)

**APPROVED.** The user agreed that an authorized application permission retirement
may proceed even when existing grants still reference that permission. Editing
every reference first is not a prerequisite for retirement.

Once retirement takes effect, the application's computed root no longer supplies
the permission. An existing grant reference cannot authorize that retired
operation. Auth does not silently rewrite or delete stored grants or assignments.

Retirement requires the application-management authority described in Q-121;
it is not authorized merely by holding the business permission or by administering
an unrelated tenant grant. Q-123's one shared application catalog applies: this
is not selective permission retirement by tenant application version.

## Grant example and consequence

Before retirement, assume this revision has an otherwise valid assignment and
supporting root route, and all permission/scope definitions are registered:

```json
{
  "version": "1",
  "grant_id": "G-PAYSLIP-DELETE",
  "revision": 1,
  "parent_grant_id": "G-HRMS-ROOT",
  "permissions": ["hrms:payroll:payslip::delete"],
  "scope": {"dept": "FIN"}
}
```

This is an immutable grant-revision example using existing fields, not a
retirement API or the unfinalized computed-root encoding. Tenant is implied.

```text
Before retirement
HRMS catalog supplies delete → root supplies delete → child can supply FIN delete
                                                    subject to all other checks

After effective retirement
HRMS catalog no longer supplies delete → root no longer supplies delete
                                      → stored child cannot authorize delete
```

The stored revision and assignment are retained unchanged. Their mere existence
or enabled status cannot restore the retired permission. An operation requiring
it must not be allowed. A failure to fetch retirement/catalog evidence is still
an evaluation failure, not proof that retirement occurred or permission to allow.

## Rationale, trade-off, and core-philosophy check

Old grant references should not force an application to keep supporting a
capability. Requiring a full reference migration first would make retirement
depend on every existing assignment and grant owner completing a cleanup task.
The approved alternative allows retirement and withdraws the authority it supplies.

This follows the live parent ceiling: stored child content cannot preserve
permission that the valid source no longer supplies. Retaining the records
preserves immutable content and explicit assignment lifecycle; it does not
grandfather access. The trade-off is loss of access for operations relying on
that permission when retirement becomes effective.

No automatic grant-disable mutation, assignment deletion, new status value,
`revoke` operation, or migration permission is introduced. The existing Q-042
rule for activating permission/scope compatibility validation is unchanged;
retirement does not authorize silently enabling an incompatible configuration.

## Q-151 / PERMISSION-008 — namespace ownership and identifier permanence

Status: **AGREED.** Both halves confirmed by the user: an application's
permissions must begin with the application — *"if the application is HRMS, every
permission should start with HRMS… it cannot register any namespace starting with
auth [or] system"* — and an identifier is never renamed, which the user noted is
already the handbook's position.

### An application registers only in its own namespace

At the application boundary, a permission's **first noun segment is the
application**:

```
hrms:payroll:payslip::read      ← hrms may register this
auth:client::read               ← hrms may not; platform namespace
system:user::read               ← hrms may not; platform namespace
```

A platform permission is a separate operation at a separate boundary, gated by
platform authority. An application cannot reach it, and cannot claim a namespace
the platform defines.

**Why this is a boundary rule and not a naming convention.** A root grant's
ceiling is computed as every active permission **in its own namespace**. An
application's *evaluation* catalog is wider — its own permissions union the
platform's — because a request inside an application may legitimately require a
platform permission. The ceiling is sliced where the catalog is not, and the
reason is exact: without the slice an application root would carry every
platform permission, including whichever one authorises establishing an
application root. The thing created by an authority could then create more of
that authority.

So an application registering under `auth:` would be minting capability into a
ceiling it does not own. That is the escalation this rule prevents, and it is why
the first segment is load-bearing rather than tidy.

It also removes a duplication: the application is stored once, as the record's own
key, and the identifier's noun path holds only what follows.

### An identifier is never renamed

[Q-126](#q-126--stable-authorization-meaning-approved) already forbids
repurposing an identifier for a materially different meaning, and permits
correcting descriptions and display labels. This states the consequence that was
left implicit: **the identifier itself is immutable.** A label is not.

A rename is strictly worse than a repurpose. The old identifier stops resolving,
and under [Q-143](#q-143--permission-007--retirement-withdraws-the-permission-not-the-route)
every grant referencing it **narrows silently** — a tenant's grant quietly
supplies less, and nothing tells anybody. A repurpose at least keeps resolving
while meaning the wrong thing; a rename fails quietly.

If a name is wrong: register the right identifier and retire the wrong one. Both
grants keep working, no authority changes without a record, and grant health is
what surfaces the grants still naming the retired one.

### Rationale / conscious tradeoff

Neither half is new machinery — both are already enforced, and this ratifies them
as rules rather than as implementation choices that happened to be made.

The cost of identifier permanence is accumulation: a catalog keeps every
identifier it ever registered, including ones retired for being badly named, and
nothing reclaims them. That is accepted, because the alternative is a name whose
meaning depends on when you read it.

## Q-143 / PERMISSION-007 — retirement withdraws the permission, not the route

Status: **AGREED.** The user chose narrowing over closure and gave the reason:
*"as far as the user is concerned, we cannot silently stop his access."*

Q-125 retires a permission without rewriting the grants that reference it. This
settles what that leaves behind — the case Q-125's own remaining boundaries
parked, *"every consequence for other still-supported permissions in a mixed
grant"*.

**A grant supplies what it selects and the catalog still supplies.** It stops only
when nothing it selects survives.

A grant selecting `payslip::read` and `payslip::write`, whose `write` is retired,
still supplies `read`. Its route narrows rather than closing, and a route hanging
beneath it that never selected `write` is untouched.

| | Before this decision | Under this decision |
|---|---|---|
| The grant that selects the retired permission | Route closes | Narrows to what survives |
| A route beneath it, selecting only survivors | Closes too | Untouched |
| A grant selecting *only* the retired permission | Route closes | Route closes |

### Why not closure

Closure was the behaviour and it needed no new rule: invalid content, and
[Q-140](authority-lineage.md) closes a route whose supporting authority is not
valid. It was rejected because retirement is a deliberate act on **one
permission**, and closure withdrew others that were never retired, plus
descendant routes that never referenced it. A person loses access nobody decided
to take away, and loses it silently.

### The write path is unchanged

A grant may not be **authored, revised or assigned** while it selects a permission
the catalog does not supply. That refusal stays. The narrowing is a property of
resolution, not of authoring: nothing new may reference a retired permission, and
what already references one keeps working for the rest.

### Narrowing changes which routes exist, not only what they supply

Narrowing cannot tell "the parent lost this to retirement" from "the parent never
had it" — it drops the permission either way. So a retirement can turn a route
that was **refused** into a route that **resolves**:

```
  parent supplies {read}
  child selects   {read, write}

  write active   →  the child is outside its parent  →  route REFUSED
  write retired  →  the child narrows to {read}      →  route RESOLVES, carrying read
```

An adversarial sweep found roughly four hundred such routes across three thousand
retirement scenarios. Nothing is amplified — the opened route is bounded by the
parent's live set and by the child's own selection, and every permission in it is
selected *and* active *and* contained. But "dropping a permission can only reduce
what a route supplies" is a statement about permission **sets**, and it is silent
about route **existence**, which is the thing this actually changes. The safety
argument needs both halves.

The state it requires — a child over-selecting relative to its parent — cannot be
authored: every write path refuses it. It arises from a permission being retired
after the fact, and retirement is reversible.

### A narrowed grant cannot be re-enabled while disabled

The read path keeps a narrowed grant working; the write path refuses to author or
re-enable one. So disabling a grant that references a retired permission is
**one-way** until the permission is restored or the grant is revised. Combined with
Q-132's bottom-up dismantle, that is a state an administrator can enter and not
leave by the same route they came in.

This is consistent — the write path is deliberately unchanged — and it is another
reason the grant-health requirement below is not optional.

### Rationale / conscious tradeoff

The cost is that a grant's effect no longer matches its record. The stored content
reads `["read", "write"]` while the grant supplies only `read`, so *"what does this
grant give"* cannot be answered from the record alone — it needs the record and the
catalog together.

That cost is accepted, and it is also the reason for the requirement below.

## Grant health — a stated requirement, not a decision

A grant that still references a retired permission is **unhealthy**. It keeps
working, which is the point of Q-143, and it is also no longer what its author
wrote. The user named the obligation that follows:

> *"There should be a way to detect that it is unhealthy and [surface it] to the
> user so that it is corrected… where we have a catalog of grants, and we show
> which one is healthy and which one is not. And it is always the administrator
> who has to do that."*

**What is required, and is not specified here:** a way to evaluate a grant's
health, a way for an administrator to see which grants in their area are
unhealthy, and correction as an administrative act. Auth does not repair them —
Q-125 already forbids silently rewriting stored grants.

This is recorded as a requirement rather than drafted as a contract, because it is
an administrative surface and no administrative surface is settled yet. It belongs
with HC-05-08 and the collection contracts, not here.

## Q-136 / PERMISSION-006 — a filter is a narrowing, not an assertion

Status: **AGREED.** The user settled this while reviewing what a retired
permission does to a running system, and directed that there be no distinction:
*"I don't need any distinction between retired and no permission. It can just
deny."*

When authority loading is asked about a permission the catalog does not supply —
**retired, or never registered** — the answer is an empty set of routes, not a
refusal. The gate then denies, as it does for any empty answer.

A caller never asks *"does this permission exist"*. It asks what a human holds,
restricted to these permissions, and narrowing a set by something absent yields
an empty set. Treating the narrowing as a claim to be validated is a category
error, and it had a cost: a refusal before any grant was examined, surfacing to
the person as *"we could not check your access"* — an untrue statement, produced
by a deliberate administrative act.

Retired and never-registered are deliberately indistinguishable. Both deny, both
fail closed, and that is what both states mean.

### What this does not change

Q-125 still governs the records: retirement rewrites no grant and deletes no
assignment. This is about the *question*, not the stored authority.

Nor does it weaken enforcement. A retired permission cannot authorize anything —
a grant selecting one no longer validates, so the route carrying it stops. That
is a separate mechanism from the filter, and it is what actually withdraws
access.

### Rationale / conscious tradeoff

The argument for keeping the refusal was a mistyped permission: a silent denial
hides a misconfigured endpoint. It does not survive the principle — the caller
did not ask, so the answer cannot carry it. The endpoint owns its policy, is
written beside the handler it guards and is tested there, and an endpoint that
denies every request is not subtle in practice.

Carrying a *reason* to the operator was considered and dropped. Auth knows why
the set is empty, the gate composes the message, and the only bridge is a new
field on a wire whose consumer rejects unknown fields — a version bump for a
diagnostic.

## Remaining contract boundaries

The retirement request/evidence format, effective visibility and concurrency,
permission-identifier reuse/restoration, and scope-definition evolution remain
open. This decision specifies loss of the retired permission; it does not settle
every consequence for other still-supported permissions in a mixed grant.
Any such outcome must respect the agreed complete-lineage rules, not silently
rewrite the stored permission selection.

Sources: [permission meaning](permission-model.md),
[application registration](application-registration.md),
[publishing authority](application-platform-authority.md),
[computed root and shared catalog](root-permission-evolution.md), and
[grant lifecycle](grant-lifecycle.md).
