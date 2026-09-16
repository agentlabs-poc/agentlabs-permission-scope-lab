# What the lab lives with, and the service has to settle

This repository is a lab. It models the authority architecture; it never becomes
the Auth service, and the application side migrates into HRMS. So some things
are deliberately left alone here because the place to answer them is the service
this work moves into — and leaving them alone is cheaper and more honest than
inventing a mechanism the handbook has not agreed.

Each entry says what the lab does, why that is tolerable here, and what the
migration owes.

---

## 1 · ~~The gate finds the route's tenant by the placeholder's spelling~~ — built

**Closed in the lab, open for the contract.** The gate bound the route's tenant
to the trusted area by reading `PathValue("tenant")`, so a policy whose path said
`{tenant_id}`, `{org}` or `{tenant_slug}` was not bound at all and nothing said
so — trusted area `acme`, `GET /api/v2/globex/FIN/C17`, **200**, handler running
against `globex`.

**What the lab now does.** A policy carries a `trusted` block correlating a field
of the trusted context with a *declared input*:

```json
"inputs":  { "tenant": { "source": "path", "name": "tenant_id" } },
"trusted": { "tenant": "tenant" }
```

read as *the input I call `tenant` must equal the trusted tenant*. The gate
compares the input's resolved value, so no path spelling is load-bearing, and a
policy declaring no tenant correlation cannot be mounted. That is what turns the
silent case into an impossible one.

**Why a declaration rather than a rule.** The alternative was for the gate to
refuse any path segment it could not account for, which needs it to guess which
segments are tenant-shaped — a heuristic that catches `tenant_id` and misses
`org`. A declaration has nothing to guess: the endpoint already knows.

**What the migration owes.** The published policy format is a handbook chapter,
so the `trusted` field is **a proposal**, not an adopted contract. It is not the
machinery CONTRACT-012 declined — that was a relationship language between
application *records*, needing a resolver interface; a correlation between
trusted context and a declared input needs neither, since both are already in the
gate's hand. What the migration has to settle is whether the published format
adopts this shape, whether a correlation may name a source other than the path,
and whether more than one trusted field is ever correlated at once. The lab
supports `tenant` and `application`; neither open question blocks the tenant case.

---

## 2 · ~~Whether a trusted root may be deleted~~ — answered by Q-132

**Closed.** The lab refused to disable a trusted root, which contradicted
`bootstrap-authority.md:151` — an established root *"remains an ordinary grant
subject to status, validity, revisions, assignments"* — and refused to delete
one, which no rule supported.

[Q-132 / GRANT-010](../../docs/grant-lifecycle.md) replaced both with one
condition: **a grant with a child grant can be neither disabled nor deleted**,
and a root is not a subject of its own. Delete additionally refuses while an
assignment names the grant, which is the older rule against leaving a dangling
reference rather than part of Q-132 — reading "child" to include assignments
would make disable unavailable for every grant anybody holds.

The migration inherits the rule, not a lab default. What the handbook still
leaves open is named in the entry: operation permission names, how "has a
dependent" is represented in published contracts, and bulk dismantle.

---

## 3 · The credential is modelled, not issued

**What the lab does.** `lab.WorkloadClient` is an id and `lab.WorkloadToken` a
secret, compared in constant time against one constant. The application is given
the token out of band.

**Why the lab lives with it.** Issuing credentials is the Auth service's
business, and the shape the credential arrives in is what this work is about.

**What the migration owes.** Real issuance, rotation and revocation, and deriving
the calling application from the token rather than from a constant.

---

## 4 · Freshness has nothing to invalidate

**What the lab does.** Nothing caches. Every request asks again, so nothing can
be stale, and `authority.epoch` is unbuilt.

**Why the lab lives with it.** Correct by construction, and the client is strict
about unknown fields — so the epoch cannot be added to the wire without a
version bump, which is itself worth knowing.

**What the migration owes.** A cache and the epoch together, or neither. The
invariant to preserve is the one demonstration 21 captures: withdrawing
authority changes the next answer.

---

## 5 · ~~The allow result's evidence does not reach the endpoint~~ — built

**Closed.** `authmiddleware` computed the contributing grant ids and
`BoundOperation.Execute` had no way to receive them, so an endpoint could record
*what* it did and nothing about why it was permitted to.

`Execute` now takes the `Result` alongside the context and the writer, which is
what `decision-results.md:306-309` asks for — the references *available in the
result*, independently of whether the request is recorded. It also bounds the
other direction: an endpoint cannot record evidence it was never handed, and a
denial never reaches the effect at all. hrms demonstrates it on the write path.

**What the migration owes.** Somewhere for it to go. Audit recording is still
unbuilt, and it is the consumer this evidence exists for.
