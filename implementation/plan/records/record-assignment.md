# `abv.assignment` — canonical record

**Draft.** The last Auth-AL record with a table of its own. When it folds,
`abv_l1_records` holds every canonical record type and only the registry
leftovers remain beside it.

It is also the record that unblocks everything else: root establishment, the
narrowing and resolution code, `DeleteGrant`'s real dependency check, and both
lineage demonstrations all wait on it.

---

## 0 · What is already approved

Unusually much, because assignments are where most of the revision decisions
landed. This document decides none of it again.

| Approved | Where |
|---|---|
| A grant has **no recipient**; the assignment carries it | Q-090 |
| The approved JSON — `version`, `id`, `grant_id`, `grant_revision`, `recipient`, `status` | Q-107 |
| Each assignment adopts its revision **explicitly**; publication never migrates it | Q-102 |
| **One current assignment per (grant, recipient)** — retained disabled ones count | Q-104 |
| Creation and explicit upgrade both select the **latest** published revision | Q-104A / Q-105 |
| Re-enabling is **not** an upgrade: retain the adopted revision, revalidate reality | Q-105 |
| Two separate checks: administrative authority **and** a supporting parent route | Q-093 |
| A child team's total authority stays within its parent team's | Q-095 |
| **No `validity` on an assignment** — deferred in v1 | Q-108 |
| `status: enabled \| disabled`, separate from the grant-wide switch | Q-106 |

Still open in the handbook, and therefore here: the exact administrative
contract, evidence and error payloads. `grant-contract-closure.md` is the list.

---

## 1 · The record

> **These are our records in the handbook's shape, not quotations of it.** The
> handbook's own excerpts use illustrative identifiers — `G1`, `A1` — and `docs/`
> keeps them. Every identifier Auth-AL issues is a base-36 Snowflake, so the
> examples here carry real ones.

```json
{
  "version": "1",
  "id": "fm5b7t4p5iv8",
  "grant_id": "fk3x9r2m5iv8",
  "grant_revision": 2,
  "recipient": {"type": "group", "id": "fibggi2juubk"},
  "status": "enabled"
}
```

**One record type, not two.** The grant needed a head and a revision because it
had immutable content and a mutable switch. An assignment has no immutable
content — it *is* a binding, and every field except `status` is fixed at
creation. So one record, and `status` in the value.

---

## 2 · Canonical key layout

> **The governing constraint, stated first because it decides the rest.**
> `abv_l1_records` is a **shared table**. Every record type lives in it, and no
> record type may change it — no column, no per-type unique constraint, no
> partial index, no expression index over the value. We derive meaning *from* the
> key slots; we do not add keys of our own.
>
> The schema offers exactly two things, and both are generic:
>
> ```sql
> PRIMARY KEY (boundary, tenant_id, key1, key2, key3, key4, key5, key6, key7, key8, key9, key10)
> CREATE INDEX abv_l1_records_prefix ON abv_l1_records
>     (boundary, tenant_id, key1, key2, key3, key4, key5)
> ```
>
> Neither mentions a record type, and neither may be made to. **This proposal
> adds nothing to the schema.**


The identity question is the interesting one, and Q-104 settles it.

**The pair `(grant, recipient)` is the identity; the assignment id is a name for
it.** Q-104: "a given grant identity has at most one current assignment to a
given recipient type/identity", and "retained disabled assignments count." That
is a uniqueness rule on the pair — which is exactly what a key path enforces for
free.

| slot | holds | example |
|---|---|---|
| `key1` | domain namespace | `abv` |
| `key2` | record type | `assignment` |
| `key3` | the application | `hrms` |
| `key4` | **the grant id** | `fi8c8111kow0` |
| `key5` | the recipient **type** | `group` |
| `key6` | the recipient **id** | `fibggi2juubk` |
| `key7` … `key10` | unused → `''` | `''` |

```
value    = {"id": "fm5b7t4p5iv8", "grant_revision": 2, "status": "enabled"}
boundary = tenant     tenant_id = acme
```

**Q-104 becomes the primary key rather than a check.** The envelope's
`PRIMARY KEY (boundary, tenant_id, key1…key10)` *is* "one assignment per
(tenant, application, grant, recipient type, recipient id)" — the same
constraint the current `assignments` table spells as a separate `UNIQUE`. A
duplicate cannot be written, disabled or not, because there is nowhere to write
it.

**Given the constraint above, this is not the better option — it is the only
one.** The `assignments` table enforces Q-104 with a `UNIQUE` of its own. The
envelope cannot: a per-record-type unique constraint is exactly what a shared
table forbids. So the pair either occupies the key path, where the existing
primary key enforces it for free, or **Q-104 stops being enforced by storage at
all** and survives only as a check in Go that a direct writer could bypass.

Every other layout runs into the same wall:

| Layout | Q-104 enforced by | Allowed? |
|---|---|---|
| pair in `key4…key6` | the shared primary key | **yes** — nothing added |
| id in `key4`, pair after | a partial unique index on the pair | no — per-record-type |
| pair, then id in `key7` | nothing — `(grant, recipient, id)` is unique, the pair is not | **breaks the rule** |
| pair, id indexed out of the value | an expression index over `value` | no — per-record-type |

That is also why the id cannot simply join the path: the primary key is **all ten
slots**, so any slot added to the path widens what is unique.

> **This is the first record whose key path is not "its own id first".** A
> permission is keyed by its identifier, a role by its id, a team by its id. An
> assignment is keyed by **what it binds**, because that pair is what must be
> unique — and the id, which is unique anyway, would not enforce anything.

### Then what is the assignment id for?

It is a stable handle for operations and for diagnosis — `SetAssignmentStatus`
takes one, receipts carry one, and the lineage code returns `AssignmentIDs` so a
decision can say which routes it used. It stays, in the **value**, because by
this record's own test it is not identity: reassigning the same grant to the
same recipient is the same binding, whatever it is called.

**Cost and gain, both measured against the code rather than guessed.**

*Cost:* four places index the snapshot by assignment id —
`abv.go:126`, `assignment_status.go:36` and `:64`, `grant_status.go:115`. Keyed
by the pair, an id lookup becomes a scan of the area's assignments. Two honest
options: carry the id in `key7` too, which reintroduces two identities for one
record, or have the operations take the pair. **The pair is right**, and it reads
better: *disable Team1's access to fk3x9r2m5iv8*, not *disable fm5b7t4p5iv8*.

*Gain, and it is on the hot path:* `lineage.resolve` calls `uniqueAssignment`
once per level of every route it walks, and that function **scans every
assignment in the area** looking for one `(grant, recipient)` — counting matches
to enforce Q-104 at read time. Keyed by the pair, it is a direct lookup and the
uniqueness it was counting is guaranteed by the primary key. The scan disappears
along with the reason for it.

So the id lookups that become scans are administrative and occasional; the scan
that stops being a scan is per level, per route, per evaluation.

### The complete canonical path

```
abv.assignment:hrms:fi8c81r9v8w4:group:fibggi2juubk
└┬┘ └───┬────┘ └─┬┘ └─────┬─────┘ └─┬─┘ └────┬────┘
 │      │        │        │         │        └─ the recipient id · key6
 │      │        │        │         └────────── the recipient type · key5
 │      │        │        └──────────────────── the grant id · key4
 │      │        └───────────────────────────── the application · key3
 │      └────────────────────────────────────── record type · key2
 └───────────────────────────────────────────── domain namespace · key1

                 key4 … key6 are the binding, and the binding IS the
                 identity — Q-104 says one per (grant, recipient), so
                 the path enforces it rather than a separate constraint
```

**Read the path aloud and it is the record**: *in HRMS, grant `fi8c81…` is
assigned to the group `fibggi2…`*. Nothing else about the assignment is identity
— not its id, not the revision it adopted, not whether it is enabled. Those can
all change or be renamed without it becoming a different binding.

The assignment id and the adopted revision ride in the value:

```
abv.assignment:hrms:fi8c81r9v8w4:group:fibggi2juubk    "fm5b7t4p5iv8" @ revision 1, enabled
```

Compare the three that came before, and the shape of the difference is the point:

```
abv.permission:hrms:employee:certificate::read      identity = what it permits
abv.role:hrms:fi8c8111kow0:0000000003               identity = its own id + revision
abv.grant:hrms:fi8c8111kow0                         identity = its own id
abv.assignment:hrms:fi8c81r9v8w4:group:fibggi2juubk identity = the two things it joins
```

A permission is identified by what it permits, a role and a grant by their own
ids. An assignment has an id too — and it is **not** what identifies it, because
the thing that must be unique is the pair.

### Canonical JSON ↔ the stored row

The JSON is the contract; the row is the storage. Neither is derived by string
surgery — the wrapper maps field to column and back:

```
  (the operation's Area)        ──▶   boundary   = tenant
                                      tenant_id  = acme
                                ──▶   key1       = abv
                                ──▶   key2       = assignment
                                      key3       = hrms
  "grant_id":      "fi8c81…"    ──▶   key4       = fi8c81r9v8w4
  "recipient": {
      "type":      "group"      ──▶   key5       = group
      "id":        "fibggi2…"   ──▶   key6       = fibggi2juubk
  }
  "id":            "fm5b7t4p5iv8"         ──▶   value.id
  "grant_revision": 1           ──▶   value.grant_revision
  "status":        "enabled"    ──▶   value.status
  "version":       "1"          ──▶   (wire-only — rebuilt by the codec, as the
                                       grant fold settled)
```

**Notice which way the fields move, and it is the opposite of every record so
far.** Elsewhere the record's own id climbs *out* of the payload into `key4`.
Here the id stays *in* the value and what climbs out is the pair the record
binds. The envelope's rule has not changed — identity goes in the slots — only
the answer to what the identity is.

### The rows

There is no `app` column and no `application_id` column: `key3` carries the
application, for every record type, since the drift correction removed it.

```
boundary  tenant_id  key1  key2        key3  key4          key5   key6          value
──────────────────────────────────────────────────────────────────────────────────────────────────────────
tenant    acme       abv   assignment  hrms  fi8c8111kow0  group  fibggi2jur5s  {"id":"fm5b7t4p0dq3","grant_revision":1,"status":"enabled"}
tenant    acme       abv   assignment  hrms  fi8c81r9v8w4  group  fibggi2juubk  {"id":"fm5b7t4p5iv8","grant_revision":1,"status":"enabled"}
```

The full envelope is `boundary | tenant_id | key1…key10 | value | ts | state`.
`ts` and `state` are written by their defaults and read by nothing; §7 records
the decision to keep them.

---

## 3 · The contract

### What exists

| Operation | |
|---|---|
| `CreateAssignment` | implemented — writes `assignments` |
| `SetAssignmentStatus` | implemented |
| `CheckAssignment` | implemented — the evaluation path |

### What does not, and should

| Operation | Why |
|---|---|
| `GetAssignment` | every other record type has a typed read. **Our precedent, not the handbook's** |
| `ListAssignments` | both directions: a grant's recipients, a recipient's grants. **Our precedent, not the handbook's** — the rule `ListMembers` and `ListInstallations` already hold |
| `DeleteAssignment` | Q-101: *"relevant bindings may be **removed** OR disabled before parent changes"*, and Q-104 distinguishes *"historical/deleted records"* from current ones. **Not Q-082** — that decision is about grants (*"cannot enable a deleted grant back"*) and does not name assignments. |
| **`UpgradeAssignment`** | the handbook's own *"authorized adoption operation"* (Q-104), governed by Q-105. **Nothing implements it**, and it is the operation the whole revision model exists for |

`UpgradeAssignment` is the one to get right:

- it selects the **latest** published revision, never an intermediate — *"do not
  silently select revision 2 instead"*;
- if the recipient's permitted authority cannot support the latest, it **rejects
  and leaves the assignment unchanged**, rather than falling back;
- it is not re-enablement, and re-enablement is not it.

---

## 4 · Validation — the heaviest of any record so far

This is the first record whose validation needs a **resolved route**, not just
the catalog. It is why `Narrow` and `lineage.resolve` have been untouched through
two slices.

| Layer | Rule | Fails with |
|---|---|---|
| shape | `version == "1"`, ids non-blank, `grant_revision > 0` | `ErrMalformed` |
| shape | `recipient.type ∈ {user, group}` | `ErrMalformed` |
| record | the grant head exists, and is not deleted | `ErrNotFound` |
| record | the adopted revision exists — the foreign key the fold removed | `ErrRejected` |
| record | the recipient exists: a team, or a human with a membership | `ErrRejected` |
| Q-104 | no current assignment for this (grant, recipient) — **disabled ones count** | `ErrConflict` |
| Q-105 | on create and on upgrade, the revision **is** the latest published | `ErrRejected` |
| Q-093 ① | the actor holds administrative authority for this operation and recipient | `ErrRejected` |
| Q-093 ② | the actor holds a **supporting parent route** for the authority assigned | `ErrRejected` |
| Q-095 | the recipient team's **total** authority stays within its parent team's | `ErrRejected` |

**Q-093's two checks are the substance.** *"Permission to assign does not itself
supply the authority being distributed. Possessing the source authority does not
itself permit assignment."* Two independent checks, and neither implies the
other — this is the rule most likely to be collapsed into one by accident.

**Q-095's team ceiling is the one with no precedent.** It is not a property of
the grant being assigned; it is a property of *everything the recipient team
already holds, plus this*. So it cannot be checked from the proposed record
alone — it needs the recipient's existing assignments resolved. That makes it
the most expensive check in the model, and the one whose cost should be measured
before it is called cheap.

---

## 5 · What this fold removes

| Table | Becomes |
|---|---|
| `assignments` | `abv.assignment` rows |

**Tables 5 → 4**, and `abv_l1_records` then holds **eight** record types:
permission, scope, role, team, membership, grant, grant_revision, assignment.
`applications` and `installations` are the only tables left beside it, and they
leave when Auth-AL's compatibility and generation move into the envelope.

---

## 6 · What this unblocks

Recorded because it is the argument for doing this record next rather than
ownership.

| Blocked today | Freed |
|---|---|
| `EstablishRoot` / `EstablishAuthRoot` | they write four things; the fourth is a holder assignment |
| `validation.Narrow`, `lineage.resolve` | they walk assignments |
| `DeleteGrant`'s real refusal | the dependants are assignments |
| both lineage demonstrations | a human acting needs an assignment and a membership |
| **a trusted root that no operation can create** | the gap carried openly since before the grant slice |

---

## 7 · `ts` and `state` stay — decided

Both are written on every insert and read by nothing, and I proposed settling
them here. **The user's decision is to keep both as they are.** Recorded so it is
not raised a fourth time.

| Column | Status |
|---|---|
| `ts TEXT NOT NULL DEFAULT (datetime('now'))` | **keep**, unread for now |
| `state TEXT NOT NULL CHECK (enabled\|disabled\|deleted)` | **keep**, unread for now |

What follows from keeping them, so the decision is implemented rather than just
tolerated:

- **An assignment writes neither.** `status` stays in the value, as every other
  record type's lifecycle does. Mapping `status` onto `state` for this one record
  would make the column mean something here and nothing everywhere else, which is
  worse than it meaning nothing consistently.
- **Nothing reads them**, so no behaviour depends on a value no writer sets
  deliberately.

They are unused capacity, not drift — `authority-change-audit.md` is unbuilt and
`ts` is the obvious place its answer would start.

---

## 8 · Scope

1. `abv.assignment` and its storage, keyed by the pair
2. `GetAssignment`, `ListAssignments`, `DeleteAssignment`, `UpgradeAssignment`
3. `assignments` dropped — **tables 5 → 4**
4. `Narrow` and `lineage.resolve` re-pointed at the folded rows
5. Root establishment, now that a holder assignment can be written
6. The two envelope questions in §7
7. Demonstrations: the grant demo extended, and **both lineages end to end**

**Success criterion, as with the registry: resolution behaves identically.**
Every existing lineage and evaluation test passes untouched. If one needs
editing, the fold changed meaning rather than location.

---

## 9 · Alignment check against the handbook

Every claim in this document traced to a decision, so nothing here is invention.

| This proposal | Handbook |
|---|---|
| the assignment carries the recipient; the grant does not | Q-090, agreed |
| the record's fields — `version`, `id`, `grant_id`, `grant_revision`, `recipient`, `status` | Q-107, approved |
| adoption is explicit; publishing a revision never migrates an assignment | Q-102, agreed |
| **one current assignment per (grant, recipient type/identity)**, disabled ones counting | Q-104, agreed |
| creation and explicit upgrade both select the **latest** published revision, never an intermediate | Q-104A / Q-105, agreed |
| re-enabling retains the adopted revision and revalidates; it is not an upgrade | Q-105, agreed |
| assignment enable/disable is a live control, separate from the grant-wide switch, and neither publishes a revision | Q-106, agreed |
| no `validity` field on an assignment | Q-108, approved — deferred in v1 |
| two independent checks: administrative authority **and** a supporting parent route | Q-093, agreed at rule level |
| a child team's total authority stays within its parent team's | Q-095, agreed |
| an assignment may be **removed**, not only disabled | Q-101 |
| an upgrade operation exists — the *"authorized adoption operation"* | Q-104 / Q-105 |
| `ts` and `state` stay | **the user's decision**, §7 — not a handbook question |
| **keying the record by `(grant, recipient)` rather than by its id** | **not in the handbook** — it is this proposal's way of making Q-104 structural. Q-104 explicitly says it is *"not a database-index prescription"*, so the handbook neither requires nor forbids it. |
| **`GetAssignment`, `ListAssignments`** | **not in the handbook** — our own precedent from teams and the registry |

### Where the handbook is deliberately silent, and this proposal does not fill it

Q-104's own words: *"This is not a database-index prescription or permission to
recreate deleted support without Q-101's dependency and explicit-enablement
checks."* So the key layout in §2 is an **implementation** of an agreed rule, not
a new rule — and if it is judged wrong, Q-104 still holds and would need
enforcing some other way.

Still open in the handbook and not answered here: the exact administrative
contract for each operation, evidence and error payloads, and condition
placement. `grant-contract-closure.md` remains the gap list.
