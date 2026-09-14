# `abv.grant` and `abv.grant_revision` — canonical record

**Draft.** The next Auth-AL record after teams and membership, and the one the
rest of the model waits on: `assignments` cannot fold before it, because its
foreign key is `(grant_id, grant_revision)`.

---

## 0 · What is already approved, and what is not

Read the handbook first; this document does not decide what it has already
decided.

| Approved | Where |
|---|---|
| A grant has **no recipient** — the assignment carries it | Q-090, `grant-assignments.md` |
| Live control and immutable content are **separate records** | Q-107, `grant-revision-format.md` |
| `status: enabled \| disabled`, grant-wide | Q-081 revised, `grant-lifecycle.md` |
| Lifecycle is **create, enable, disable, delete** — `delete` supersedes revoke | Q-082 |
| `permissions` **XOR** (`role_id` + `role_revision`) — never a mixture | Q-118, `role-grant-contract.md` |
| Validity belongs on the **revision**, not the control or the assignment | Q-109 / Q-108 |
| `parent_grant_id` may be omitted **only for a trusted root** | Q-119 |
| An adopted revision never auto-upgrades | Q-102–Q-106 |

Explicitly **not** settled by the handbook, and therefore not settled here:
complete field validation, id and revision value rules, initial/default status,
timestamp validation, and the root wire representation. `grant-contract-closure.md`
is the standing gap list.

> **Q-107 says this separation is "a logical contract separation, not a
> requirement for three database tables."** That is permission to fold the three
> tables into the envelope — not a licence to merge the two records.

---

## 1 · Why this is two record types, not one

The role fold settled a single immutable record: `key4`=id, `key5`=revision. The
grant looks the same and is not.

**A grant carries a mutable status that belongs to the grant, not to any
revision.** Disabling G1 makes authority through G1 ineffective *across every
revision and assignment* — so the status cannot live on a revision record, which
is immutable by contract, and cannot be duplicated across revisions, which would
let two revisions of one grant disagree about whether it is enabled.

So: one head record holding what changes, one revision record holding what never
does.

```
abv.grant           mutable   status                    one per grant
abv.grant_revision  immutable permissions/role, scope   one per published revision
```

This is the test the envelope has used throughout — *does this change make it a
different record?*

| Change | Different record? | Therefore |
|---|---|---|
| publish revision 3 | **yes** — revision 2 still exists and is still adopted | key slot |
| disable the grant | no — same grant, same revisions | value on the head |
| mark it a trusted root | no — same grant | value on the head |

---

## 2 · Canonical JSON — the approved wire forms

Unchanged from `grant-record-reference.md`. Reproduced so the mapping below has
something to map.

**The control**

```json
{
  "version": "1",
  "id": "G1",
  "status": "enabled"
}
```

**The revision — explicit permissions**

```json
{
  "version": "1",
  "grant_id": "G1",
  "revision": 2,
  "parent_grant_id": "G0",
  "permissions": ["hrms:payroll:payslip::read", "hrms:payroll:payslip::write"],
  "scope": {"dept": "FIN"}
}
```

**The revision — role reference.** Q-118: either `permissions`, or both role
fields, never a mixture.

```json
{
  "version": "1",
  "grant_id": "G2",
  "revision": 1,
  "parent_grant_id": "G1",
  "role_id": "fi8c8111kow0",
  "role_revision": 1,
  "scope": {"dept": "FIN"}
}
```

---

## 3 · The id becomes a base-36 Snowflake

`G1`, `G0`, `G-17` are the handbook's *illustrative* ids, the way `maya` and
`Team1` were before the sweep. Every id Auth-AL issues is a base-36 Snowflake —
settled for roles, teams and human ids — and a grant id is issued by Auth-AL.

```
"id": "G1"   ──▶   "id": "fi8c8111kow0"
```

**This is a sweep, not a rename.** `G0`/`G1` appear across fixtures, scenarios,
tests and demo scripts, and the last such sweep touched 69 files rather than the
23 estimated. Two specific traps from that one, recorded here so they are not
rediscovered:

- **single-quoted SQL literals escaped the first pass**, so a `WHERE grant_id='G1'`
  matched nothing and silently turned a negative test green;
- the handbook's own prose keeps `G1` and must not be rewritten — `docs/` is
  untouched, and the illustrative id there stays illustrative.

The application slug stays a slug for the reason the registry charter gives: it
appears in every canonical path. A grant id does not.

---

## 4 · Canonical key layout

`key3` is the application, as it is in every record type since the drift
correction.

### `abv.grant` — the head

| slot | holds | example |
|---|---|---|
| `key1` | domain namespace | `abv` |
| `key2` | record type | `grant` |
| `key3` | the application | `hrms` |
| `key4` | **the grant id** — base-36 Snowflake | `fi8c8111kow0` |
| `key5` … `key10` | unused → `''` | `''` |

```
value    = {"status": "enabled", "trusted_root": false}
boundary = tenant     tenant_id = acme
```

### `abv.grant_revision` — the content

| slot | holds | example |
|---|---|---|
| `key1` | domain namespace | `abv` |
| `key2` | record type | `grant_revision` |
| `key3` | the application | `hrms` |
| `key4` | **the grant id** | `fi8c8111kow0` |
| `key5` | **the revision** — zero-padded to 10 | `0000000002` |
| `key6` … `key10` | unused → `''` | `''` |

```
value    = {"parent_grant_id": "...", "permissions": [...], "scope": {...}, "validity": {...}}
boundary = tenant     tenant_id = acme
```

Zero-padding is the role's rule and exists for the same reason: `key5` is TEXT,
so `0000000010` must sort after `0000000009`.

### The mapping, field by field

```
  (the operation's Area)      ──▶   boundary       = tenant
                                    tenant_id      = acme
                              ──▶   key1           = abv
                              ──▶   key2           = grant_revision
                                    key3           = hrms
  "grant_id":  "fi8c81…"      ──▶   key4           = fi8c8111kow0
  "revision":  2              ──▶   key5           = 0000000002
  "parent_grant_id": "G0"     ──▶   value.parent_grant_id
  "permissions": [ 2 items ]  ──▶   value.permissions
  "scope":     {"dept":"FIN"} ──▶   value.scope
  "version":   "1"            ──▶   see §7 — the open question comes due here
```

### The rows

```
boundary tenant app  key1 key2           key3 key4          key5        value
─────────────────────────────────────────────────────────────────────────────────────────────────
tenant   acme   hrms abv  grant          hrms fi8c8111kow0              {"status":"enabled","trusted_root":false}
tenant   acme   hrms abv  grant_revision hrms fi8c8111kow0  0000000001  {"parent_grant_id":"…","permissions":[1],"scope":{}}
tenant   acme   hrms abv  grant_revision hrms fi8c8111kow0  0000000002  {"permissions":[2],"scope":{"dept":"FIN"}}
```

The head has an empty `key5` while its revisions fill it — which the contiguity
CHECK permits, because the head occupies a **prefix** of the slots and stops.

---

## 5 · The trusted root becomes a flag, and `trusted_roots` folds away

`trusted_roots` is a three-column table — `(tenant_id, application_id, grant_id)`
— and its job is **trust evidence**, not encoding. Q-119 is explicit that the two
are different things: a root is encoded by *omitting* `parent_grant_id`, and
"omitting a field proves nothing." The table is what says the omission is
legitimate.

Being a trusted root does not make it a different grant, so by §1's test it is
value on the head, not a key slot:

```json
{"status": "enabled", "trusted_root": true}
```

> **Open — does this collide with Q-119's "no root flag"?** Q-119 refuses a root
> flag *in the revision content*, because the content is the thing a caller
> submits and a self-asserted flag would be a parentless escape. The head is not
> submitted content; it is what Auth records about the grant, which is where trust
> evidence belongs. **I read the two as compatible and the fold as legitimate** —
> but it is close enough to the line to settle deliberately rather than by
> assumption, and the answer decides whether `trusted_roots` folds or survives.

> **Finding — nothing creates a trusted root today.** The only
> `INSERT INTO trusted_roots` in the tree is in `internal/storage/sqlite/fixture.go`.
> There is no facade operation, no CLI verb, and no administrative gate. A root
> can be *seeded* by a lab scenario and can never be *established*. That is a
> gap in the contract, not something the fold introduces — but the fold must not
> quietly preserve it by making `trusted_root` a field nothing can ever set.
> **Decide with this record**, because `false` is a defensible default only if
> something can eventually make it `true`.

---

## 6 · The contract

### What exists today

| Operation | State |
|---|---|
| `PublishGrantRevision` | **implemented** — writes `grant_contents` |
| `SetGrantStatus` | **implemented** — writes `grant_controls` |

### `PublishGrantRevision` is amendment-only

Reading `insertGrantRevision` rather than its name: before it writes anything it
requires **all** of

| Precondition | else |
|---|---|
| the control row exists | `ErrNotFound` |
| the grant is **not** a trusted root | `ErrRejected` |
| **at least one prior revision exists** | `ErrNotFound` |
| `proposed.Revision > latest` | `ErrConflict` |
| `parent_grant_id` identical to the latest revision | `ErrRejected` |

Two consequences follow, and neither is visible from the operation's name.

**Revision 1 can never be published.** The third precondition asks for a
predecessor, so the first revision of any grant has to arrive some other way.
**A root's revisions can never be published at all** — the second precondition
refuses them outright, deliberately.

So `PublishGrantRevision` amends grants; it does not originate them. Combined
with the only `INSERT INTO grant_controls` in the tree living in `fixture.go`,
**every grant in the system traces back to a seeded fixture.** Q-082 approves
create / enable / disable / delete; of the four, only enable and disable exist.

### Seeding is right for the root and wrong for everything else

The distinction matters, because "seed it" is the obvious answer and it is half
correct.

**A root must be seeded.** It has no parent by definition, so it cannot be
derived from anything that already exists — it is established by trusted setup,
which is what bootstrap means. The handbook says so and says the trusted-setup
evidence is unfinished (Q-117/Q-124). The code agrees with the handbook here:
refusing to publish revisions of a root is the right refusal. **Nothing in this
record changes that** — `trusted_root` becomes a seeded flag, and establishing a
root stays the bootstrap question it already is.

**An ordinary grant must not be.** A seed is a *starting* state; it cannot be the
only way to *reach* a state. The model exists to show an administrator delegating
narrower authority downward from a root — and today you can amend a grant a
fixture wrote, but never create one. That is the gap, and it is not a lab
convenience issue: `CreateGrant` is a contract operation with an authority gate,
and no gate can be exercised by a fixture.

> **Design consequence.** `CreateGrant` writes the head **and revision 1 in one
> transaction**, because revision 1 cannot go through the publish path. It is one
> operation, not a head followed by a publish — and its parent check is the
> ordinary one, since a child's first revision still needs real upstream support.

| Operation | Why it is needed |
|---|---|
| `CreateGrant` | Q-082's `create`. Issues the Snowflake id, writes the head at `status=enabled`, `trusted_root=false`, **and revision 1**, atomically. Add-only. |
| `GetGrant` | The head plus, on request, one revision. Every other record type has a typed read. |
| `ListGrants` | Offset paging, the bound and the cap the other listings already carry. |
| `ListGrantRevisions` | The role's `--latest` precedent: revisions of one grant, newest first. |
| `DeleteGrant` | Q-082's `delete`, which supersedes revoke. **Must refuse while an assignment adopts any of its revisions** — the rule teams settled. |

**`CreateGrant` before `PublishGrantRevision`, always.** A revision of a grant
that does not exist is the same error as an installation of an application that
does not exist, and gets the same answer.

### Not in this contract

Assignments. They adopt a grant revision and are the next record after this one;
nothing here changes `assignments`, and it keeps its table until then.

---

## 7 · Two open questions this record forces

**`version` has nowhere to go — and the grant is where it stops being deferrable.**
`record-role.md` raised it and left it open: the envelope has no version column,
and `abv_metadata.schema_version` versions the schema, not a record's format. The
role could defer it because nothing stored `version`. **`grant_controls` has a
`version` column today**, so folding the table either finds the field a home or
drops something that is currently persisted. The three options are unchanged —
inside the value, a new envelope column, or wire-only in the codec — and the
answer binds every record type with a canonical JSON, not just this one.

**What is the head's `state`, given the value also has a `status`?** The envelope
carries `state ∈ enabled|disabled|deleted`, and the grant's own contract carries
`status ∈ enabled|disabled`. Two switches with one meaning is drift of exactly
the kind the envelope correction removed. Either the head's `status` leaves the
value and *is* `state` — with `deleted` then serving Q-082's delete — or the
value keeps `status` and `state` stays permanently `enabled`, which makes the
envelope column a lie. **The first is right**, and it is worth saying out loud
because it is the first time a record's own lifecycle and the envelope's have
been the same lifecycle.

---

## 8 · What this fold removes

| Table | Becomes |
|---|---|
| `grant_controls` | `abv.grant` rows |
| `grant_contents` | `abv.grant_revision` rows |
| `trusted_roots` | a flag in the `abv.grant` value |

**Tables 8 → 5.** The precedent is exact and verifiable: the permission, scope,
role, team and membership folds each *removed* their table rather than leaving it
beside the envelope, and `role_l1_test.go` asserts `SELECT 1 FROM roles` still
fails. This fold carries the same assertion.

`assignments` stays until the assignment record. `applications` and
`installations` stay for the reasons the registry charter records.

---

## 9 · The root

The root's shape, permission source, trust establishment, the two lineages that
reach it and the operations that create it are **not in this document**. They are
[the root proposal](proposal-root-source.md), because they are handbook
questions rather than record-layout ones.

What this record must preserve, and the proposal depends on:

| | at the root | downward |
|---|---|---|
| permissions | computed ceiling, grows with the catalog | **selected** — each child states its own, checked against the parent's set |
| scope | stored; `{}` is the whole region | **inherited whole**, the child may only AND more on |

`validation.Narrow` is where that asymmetry lives, four lines apart: the child's
permissions *replace*, the parent's predicates are *cloned then appended to*.
