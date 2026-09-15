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
revision.** Disabling fk3x9r2m5iv8 makes authority through fk3x9r2m5iv8 ineffective *across every
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

> **These are our records in the handbook's shape, not quotations of it.** The
> handbook's own excerpts use illustrative identifiers — `G1`, `A1` — and `docs/`
> keeps them. Every identifier Auth-AL issues is a base-36 Snowflake, so the
> examples here carry real ones.

Unchanged from `grant-record-reference.md`. Reproduced so the mapping below has
something to map.

**The control**

```json
{
  "version": "1",
  "id": "fk3x9r2m5iv8",
  "status": "enabled"
}
```

**The revision — explicit permissions**

```json
{
  "version": "1",
  "grant_id": "fk3x9r2m5iv8",
  "revision": 2,
  "parent_grant_id": "fk3x9r2m0dq3",
  "permissions": ["hrms:payroll:payslip::read", "hrms:payroll:payslip::write"],
  "scope": {"dept": "FIN"}
}
```

**The revision — role reference.** Q-118: either `permissions`, or both role
fields, never a mixture.

```json
{
  "version": "1",
  "grant_id": "fk3x9r2man0d",
  "revision": 1,
  "parent_grant_id": "fk3x9r2m5iv8",
  "role_id": "fi8c8111kow0",
  "role_revision": 1,
  "scope": {"dept": "FIN"}
}
```

---

## 3 · The id — issuance is strict, acceptance is not yet

`fk3x9r2m5iv8`, `fk3x9r2m0dq3`, `G-17` are the handbook's *illustrative* ids, the way `maya` and
`Team1` were before the sweep. Every id Auth-AL issues is a base-36 Snowflake —
settled for roles, teams and human ids — and a grant id is issued by Auth-AL.

```
CreateGrant issues   "id": "fy85p22i8glc"
the corpus holds     "id": "fk3x9r2m0dq3"
```

**What shipped splits the two.** `CreateGrant` issues a Snowflake and never
accepts a caller's identifier. But `insertGrantHead` admits any non-blank
identifier, deliberately: the fixtures and the handbook's worked examples use
`fk3x9r2m0dq3`/`fk3x9r2m5iv8`/`fk3x9r2man0d`, and enforcing the alphabet in storage would reject the corpus
before the sweep that converts it. **Issuance strict, acceptance after the
sweep** — and the sweep is not done.

**It is a sweep, not a rename.** `fk3x9r2m0dq3`/`fk3x9r2m5iv8` appear across fixtures, scenarios,
tests and demo scripts, and the last such sweep touched 69 files rather than the
23 estimated. Two specific traps from that one, recorded here so they are not
rediscovered:

- **single-quoted SQL literals escaped the first pass**, so a `WHERE grant_id='fk3x9r2m5iv8'`
  matched nothing and silently turned a negative test green;
- the handbook's own prose keeps `fk3x9r2m5iv8` and must not be rewritten — `docs/` is
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
  "parent_grant_id": "fk3x9r2m0dq3"     ──▶   value.parent_grant_id
  "permissions": [ 2 items ]  ──▶   value.permissions
  "scope":     {"dept":"FIN"} ──▶   value.scope
  "version":   "1"            ──▶   see §7 — the open question comes due here
```

### The rows

```
boundary  tenant_id  key1  key2            key3  key4          key5        value
─────────────────────────────────────────────────────────────────────────────────────────────────
tenant    acme       abv   grant           hrms  fi8c8111kow0              {"status":"enabled","trusted_root":false}
tenant    acme       abv   grant_revision  hrms  fi8c8111kow0  0000000001  {"parent_grant_id":"…","permissions":[1],"scope":{}}
tenant    acme       abv   grant_revision  hrms  fi8c8111kow0  0000000002  {"permissions":[2],"scope":{"dept":"FIN"}}
```

The head has an empty `key5` while its revisions fill it — which the contiguity
CHECK permits, because the head occupies a **prefix** of the slots and stops.

---

## 5 · The trusted root is a flag on the head

`trusted_roots` was a three-column table — `(tenant_id, application_id, grant_id)`
— whose job is **trust evidence**, not encoding. Q-119 is explicit that the two
differ: a root is encoded by *omitting* `parent_grant_id`, and "omitting a field
proves nothing." The table said the omission was legitimate.

Being a trusted root does not make it a different grant, so by §1's test it is
value on the head, not a key slot:

```json
{"status": "enabled", "trusted_root": true}
```

It is **not** in the canonical JSON. Q-119 refuses a root flag in submitted
*content*, because content is what a caller sends and a self-asserted flag would
be a parentless escape. The head is what Auth recorded, which is where evidence
belongs — `domain.Grant` carries it, `domain.GrantControl` (the wire form) does
not.

> Whether that reading of Q-119 is right is [the proposal's question](proposal-root-source.md),
> not this document's. This records what was built on it.

**Nothing establishes a trusted root.** The only writer is
`internal/storage/sqlite/fixture.go`; there is no operation, no CLI verb and no
gate. A root can be *seeded* and never *established*. That gap is unchanged by
this record and is the first thing the assignment slice closes — establishment
writes four things and the fourth is a holder assignment.

## 6 · The contract

### What exists

| Operation | |
|---|---|
| `CreateGrant` | head **and revision 1**, one transaction, id issued |
| `GetGrant` | head alone at `revision=0`, or head plus one revision |
| `ListGrants` | status filter, tri-state root filter, offset paging |
| `ListGrantRevisions` | one grant's revisions, newest first |
| `DeleteGrant` | head and every revision, refusing while depended on |
| `PublishGrantRevision` | pre-existing — amends a grant |
| `SetGrantStatus` | pre-existing — `enabled \| disabled`, reversible |

Q-082 names create, enable, disable and delete. All four now exist.

### `PublishGrantRevision` is amendment-only

Reading `insertGrantRevision` rather than its name, it requires **all** of:

| Precondition | else |
|---|---|
| the head exists | `ErrNotFound` |
| the grant is **not** a trusted root | `ErrRejected` |
| **at least one prior revision exists** | `ErrNotFound` |
| `proposed.Revision > latest` | `ErrConflict` |
| `parent_grant_id` identical to the latest revision | `ErrRejected` |

Two consequences follow, and neither is visible from the name. **Revision 1 can
never be published** — the third precondition asks for a predecessor. **A root's
revisions can never be published at all**, deliberately: a root's coverage is
computed from the catalog, so there is nothing to amend.

### Which is why `CreateGrant` writes both records

One operation, not a head followed by a publish. A create-then-publish pair would
leave a window where a grant exists, can be enabled, and supplies nothing.

**The id is issued, never accepted** — `fy85p22i8glc`, not a name the caller
chose. A caller who can name a grant is one step from naming a root.

**A parentless create is `ErrRejected`, not a root.** The guarantee Q-113 asks
for is not a check inside creation; it is that **no grant operation writes trust
evidence at all**, and nothing in `GrantAdministration` can. The CLI refuses it
at the command line with a message that says so, rather than reporting a missing
flag.

**The catalog and role reads happen inside the writing transaction**, so a
permission retired between check and write cannot be admitted.

### `DeleteGrant`

Q-082 folded revoke into delete, so this is the one permanent removal. It
destroys the grant whole — head and every revision — because a head with no
content is a grant that can be enabled and supplies nothing, and content with no
head is authority with no live switch.

It refuses while a child grant or an assignment depends on it, the rule teams
settled. That refusal is deliberately the conservative direction: it is loosened
when assignments are their own record, never tightened.

### Not in this contract

Assignments, and root establishment. Both are the next slice.

## 7 · One question settled, one left open

**`version` is settled: wire-only.** `record-role.md` raised it and left it open —
the envelope has no version column, and `abv_metadata.schema_version` versions the
schema rather than a record's format. The role could defer it because nothing
stored `version`; `grant_controls` had a column, so the fold had to answer.

Of the three options — inside the value, a new envelope column, or wire-only in
the codec — **wire-only** is what shipped. `version` is always `"1"`
(`ValidateContent` returns `ErrUnsupported` for anything else), so a column or a
value field would persist a constant. It is rebuilt on the way out by
`decodeRevision`, exactly as `grant_id` and `revision` are rebuilt from their key
slots.

The cost is real and worth naming: **a stored row cannot say which format wrote
it.** That is acceptable while there is one format and the codec refuses every
other, and it stops being acceptable the day a second one exists. The answer
binds every record type with a canonical JSON, not just this one.

**Still open: the head's `state` against the value's `status`.** The envelope
carries `state ∈ enabled|disabled|deleted`; the grant's contract carries
`status ∈ enabled|disabled`. Two switches with one meaning is drift of the kind
the envelope correction removed.

**What shipped keeps both**: `status` lives in the head's value and `state` is
left at its `enabled` default, unread. So the envelope column is currently a lie
for this record type. The better answer is that the head's `status` leaves the
value and *is* `state`, with `deleted` serving Q-082's delete — but that touches
every record type's relationship to `state`, not just this one, so it is recorded
here rather than decided here.

---

## 8 · What this fold removes

| Table | Becomes |
|---|---|
| `grant_controls` | `abv.grant` rows |
| `grant_contents` | `abv.grant_revision` rows |
| `trusted_roots` | a flag in the `abv.grant` value |

**Tables 8 → 5** at the time; two tables remain today, and §9 of
[the assignment record](record-assignment.md) records the rest of the journey. The precedent is exact and verifiable: the permission, scope,
role, team and membership folds each *removed* their table rather than leaving it
beside the envelope, and `role_l1_test.go` asserts `SELECT 1 FROM roles` still
fails. This fold carries the same assertion.

`assignments` went with the assignment record; `applications` and `installations`
went when the catalog's own state became a record. **Two tables remain.**

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
