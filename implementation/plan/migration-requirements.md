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
policy declaring no tenant correlation cannot be mounted.

Three rules, not two — the third was missing on the first attempt and the entry
claimed the silent case was impossible while it was not. Nothing required the
correlation to name the input that actually carries the route's claim, so a
policy could hold `{tenant}` in its path, point its correlation elsewhere, and
the segment the handler reads went unchecked. A path placeholder spelled with a
trusted field's own name must now be read by the input that field correlates.
That fires only on this gate's two field names, spelled exactly; a path that
calls it `{org}` is covered by the mandatory correlation instead.

`trusted` is a required field, to decode and to mount. A policy document carrying
only version, method, path, permission and inputs is refused.

**Why a declaration rather than a rule.** The alternative was for the gate to
refuse any path segment it could not account for, which needs it to guess which
segments are tenant-shaped — a heuristic that catches `tenant_id` and misses
`org`. A declaration has nothing to guess: the endpoint already knows.

**What the migration owes.** The published policy format is a handbook chapter,
so the `trusted` field is **a proposal**, not an adopted contract. Note also what
was given up: the gate previously held *every* route tenant to the trusted one
unconditionally, by spelling. The declaration is stronger where the spelling
differed and weaker where it matched, and only the mount rule above closes the
gap. A deployment adopting this shape has to decide whether a correlation is
required for every trusted field a path exposes. It is not the
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

Two assignment-side root special cases went with the grant-side ones: delete of
an assignment, and its status change, each refused while the grant was a trusted
root. Q-132 does not cover them — it is about grants — and their old
justification, that deleting one takes the root's holder away and leaves a
ceiling nobody holds, is not replaced by anything. **Open for the migration:**
whether a root's last holder may be removed, and what it means if it is.

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

---

## 6 · What a retirement does to a mixed grant

**What the lab does.** Retiring one permission withdraws every route through a
grant that selects it — including routes *beneath* that grant which select only
permissions still supplied. Demonstration 18 captures it: maya holds a grant for
`::read` and `::write` at `dept=FIN`, and a deeper one for `::read` alone at
`cert=C17`; retiring `::write` leaves her holding nothing at all, because the
deeper route's chain runs through the mixed grant.

**Why it is not a defect.** [Permission lifecycle](../../docs/permission-lifecycle.md)
says so in terms under Q-125's remaining contract boundaries: *"This decision
specifies loss of the retired permission; it does not settle every consequence
for other still-supported permissions in a mixed grant."* The route-wise outcome
is defensible under `authority-lineage.md:169`, and it is fail-closed.

**What the migration owes.** The settlement. The available answers are that a
mixed grant loses only the retired permission and keeps narrowing beneath it,
or that it is withdrawn whole as it is here. Either is defensible; the lab
implements the second because it is what the chain walk already did, not because
it was chosen.
