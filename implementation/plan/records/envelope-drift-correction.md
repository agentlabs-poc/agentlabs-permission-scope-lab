# Envelope drift — correcting `abv_l1_records`

What is wrong, what changes, where every change lands, and what the correction
costs. **Implemented.** The demonstration is `demos/demo-11-envelope-drift-correction.svg`.

---

## 1 · What drifted

The canonical 123 L1 envelope, as the payroll lab holds it:

```sql
CREATE TABLE payroll_l1_records (
  tenant TEXT NOT NULL,
  key1 … key10 TEXT NOT NULL,
  value TEXT NOT NULL CHECK(json_valid(value) AND json_type(value)='object'),
  ts TEXT NOT NULL,
  state TEXT NOT NULL CHECK(state IN ('enabled','disabled','deleted')),
  UNIQUE(tenant, key1, …, key10)
);
```

`abv_l1_records` carries **three columns beyond it**:

| Column | Why it is redundant | Evidence |
|---|---|---|
| `revision` | Roles put theirs in `key5`; permission and scope only ever write `0`. Nothing reads it for meaning. | `TestRolesAreL1Records` asserts it stays `0` |
| `boundary` | Says whether a record is tenant-scoped — which `tenant_id` already says by being empty or not | every write derives it from the tenant it was handed |
| `application_id` | Duplicates `key3`, which every record type carries | permission, scope and role all write the application into `key3` |

**The columns are frozen.** Structure belongs in the key path; a column per
concern rebuilds the per-record-type tables the fold exists to remove. That is
the whole premise of one canonical store, and three columns have quietly
contradicted it.

### Why now

Every record type added from here copies the current shape. **Eight tables remain
outside the store**, so deferring this multiplies the correction by eight. The
schema is pre-production, so this is the cheapest it will ever be.

---

## 2 · The target

```sql
CREATE TABLE abv_l1_records (
    tenant_id TEXT NOT NULL,
    key1 TEXT NOT NULL, key2 TEXT NOT NULL,
    key3 … key10 TEXT NOT NULL DEFAULT '',
    value TEXT NOT NULL,
    ts TEXT NOT NULL DEFAULT (datetime('now')),
    state TEXT NOT NULL DEFAULT 'enabled'
        CHECK (state IN ('enabled', 'disabled', 'deleted')),
    PRIMARY KEY (tenant_id, key1, key2, key3, key4, key5,
                 key6, key7, key8, key9, key10)
);

CREATE INDEX abv_l1_records_prefix ON abv_l1_records
    (tenant_id, key1, key2, key3, key4, key5);
```

`tenant_id` stays `''` rather than NULL for an application-scoped record, for the
reason already settled: both SQLite and PostgreSQL treat NULLs in a unique index
as distinct, which would let duplicates coexist.

**The primary key gets simpler**, not more complex: fourteen columns become
eleven, and it is exactly the canonical index rather than a variant of it.

---

## 3 · Every place that changes

The SQL surface is small — **three files, eight statements**. Most of the wider
grep noise is unrelated uses of the word *revision*.

### 3.1 · Schema — 1 file

`internal/storage/sqlite/migrations/001_initial.sql`

- drop the three columns from `abv_l1_records`
- drop the `boundary`/`tenant_id` CHECK
- drop the `application_id` foreign key
- rebuild the primary key and the prefix index
- `schema_version` 5 → 6, and `supportedSchemaVersion` in `migrate.go`

### 3.2 · Writes — 2 files, 3 statements

| Where | Statement | Change |
|---|---|---|
| `sqlite/catalog.go:194` | `insertPermissionRecord` | drop `boundary`, `application_id`, `revision` from the column list; `application_id` was already written into `key3` |
| `sqlite/catalog.go:244` | `insertScopeRecord` | same |
| `sqlite/role.go:49` | `insertRole` | same, and the `boundary` variable disappears — the tenant argument alone decides |

### 3.3 · Reads — 2 files, 5 statements

| Where | Change |
|---|---|
| `sqlite/catalog.go:160` | `WHERE boundary='application' AND tenant_id='' AND application_id=?` → `WHERE tenant_id='' AND key1='abv' AND key2=… AND key3=?` |
| `sqlite/catalog.go:257` | same shape |
| `sqlite/snapshot.go:94` | permissions read |
| `sqlite/snapshot.go:131` | scopes read |
| `sqlite/snapshot.go:258` | roles read — **the interesting one**, see below |

**The role read is where the correction earns its keep.** Today it asks for
either boundary:

```sql
AND ((boundary='tenant' AND tenant_id=?) OR boundary='application')
```

After, the same question is asked of the column that actually holds the answer:

```sql
AND (tenant_id=? OR tenant_id='')
```

And `RoleContent.Managed` is derived from `tenant_id` being empty rather than
from a `boundary` string — one fact, read from one place.

### 3.4 · Fixtures — 1 file

`sqlite/fixture.go` seeds through the same writers, so it follows them. The
`boundary` derivation in `seedArea` goes.

### 3.5 · Tests — 3 files

| Where | Change |
|---|---|
| `sqlite/role_l1_test.go` | `TestRolesAreL1Records` asserts `revision` stays `0` — that assertion is **replaced** by one that the column does not exist |
| `sqlite/role_l1_test.go` | `TestApplicationAndTenantRolesDifferOnlyInTheBoundary` — renamed, and asserts on `tenant_id` rather than `boundary` |
| `sqlite/catalog_test.go`, `provider_test.go` | queries naming the dropped columns |

Everything else in the suite reads through the contract and does not name a
column, which is the point of having gone through the CLI all along.

---

## 4 · What the correction costs

Three real losses, all of them enforcement moving from the database into
validation. None is a surprise; two are already recorded elsewhere.

| Lost | What enforced it | Where it goes |
|---|---|---|
| **`boundary` / `tenant_id` CHECK** | *application-scoped means empty tenant, tenant-scoped means a real one* | a validation rule at the write path |
| **`application_id` → `applications` FK** | *this application exists* | validation, reading the catalog it already loads |
| **per-row application typing** | a query could name `application_id` directly | `key3` answers it, on the same index |

> **The installation foreign key is already lost**, and `00-open-questions.md`
> records it: a composite key cannot reference `installations`. This correction
> does not make that worse — it puts `application_id` in the same position, which
> at least makes the situation uniform rather than half-enforced.

**Nothing about authority changes.** No contract signature, no validation rule
about permissions or scopes or bundles, no evaluation path. This is storage
shape only, and the test suite is the proof: every test that is not a storage
test should pass untouched.

---

## 5 · Settled — no rendered-value column

Payroll's envelope carries an eleventh column holding the assembled canonical
string, built by the database on every insert:

```sql
key TEXT GENERATED ALWAYS AS (canonical_record_key(key1,…,key10)) VIRTUAL,
UNIQUE(tenant, key)
```

**We are not adding it.** Decided.

| | |
|---|---|
| **The slots are the stored truth** | ten columns, and nothing else |
| **The rendered string is produced on demand** | in Go, at the boundary, and never stored |
| **One renderer, one direction each way** | parse coming in, render going out |

Storing the value twice means two things that must agree forever, and when they
disagree the database is silently wrong. It is the same rule that removed
escaping from the permission identifier: exactly one piece of code renders a
canonical string, and nothing else does.

There is a second reason, and it is the stronger one. For the database to *build*
that string it must know how every record type is shaped — that a permission is a
namespace plus nouns plus a verb, that a scope is one flat key, that a role is an
id plus a revision plus a name. **That is a record-type registry expressed in
SQL.** It is not application knowledge, so it does not break the rule that no
application vocabulary lives in the engine — but it is the same kind of mistake
one level up, putting canonical structure somewhere a second implementation has
to maintain it.

Two practical costs follow. The database stops being readable without the
function loaded — payroll records this themselves: *"A plain external SQLite
connection cannot evaluate the virtual key without that function."* And it must
be written twice, since PostgreSQL generated columns cannot call an application
function at all and would need a trigger instead.

**And we gain nothing.** Payroll needs the string for ten partial indexes that
match on its prefix. Every query we have is slot equality, already answered by
the identity index.

## 6 · What the work found — `key3` had to become the application

**`application_id` could not simply be dropped.** A permission's `key3` held the
*leading noun*, and PR #3 established that nothing requires the leading noun to be
the application. Without the column, two applications registering
`billing:invoice::read` would have produced an identical key path and collided on
the primary key.

Scope and role were already safe — both write the application id into `key3`
directly. Permissions were not.

**So the correction settles what PR #3 left open:** `key3` is the application, in
every record type, written from the application id. The permission noun path moves
to `key4`…`key9`, **six slots instead of seven**, and the leading noun stays in it,
so an identifier still renders exactly as its author wrote it:

```
hrms:employee:certificate::read
  key3=hrms   ← the application, from the application id
  key4=hrms   ← noun 1, as written
  key5=employee   key6=certificate   key10=read
```

The application appears twice, deliberately. Dropping the duplication would mean
assuming an identifier's first segment equals its application id — the assumption
the permission document has rejected from the start.

PR #3 recommended *against* this change, on the grounds that it bought only
tidiness and cost a noun slot. That reasoning was correct at the time and is now
superseded: it buys correctness.

---

## 7 · Shape of the work

**One PR, not three.** The three columns are entangled: the primary key names all
of them, so removing one means rebuilding it anyway, and a half-corrected
envelope is worse than either end state. The demonstration is one table, before
and after.

| Step | |
|---|---|
| 1 | Schema: drop three columns, rebuild the key and index, bump to 6 |
| 2 | Writes: three insert statements |
| 3 | Reads: five select statements, and `Managed` derived from `tenant_id` |
| 4 | Fixtures follow the writers |
| 5 | Validation takes the two checks the database gives up |
| 6 | Tests: three files; the rest must pass untouched |
| 7 | Demonstration: the row shape before and after, through the CLI |

**The success criterion is that nothing outside storage changes.** If a contract
test needs editing, the correction has reached further than it should and that
is worth stopping over.

---

## 8 · What this does not do

- It does not fold any of the eight remaining tables.
- It does not touch `applications`, which keeps `generation` and
  `compatibility_enabled` — those are that record's own fields, not envelope
  columns.
- It does not resolve the `state`-versus-`{"active": true}` inconsistency. That
  is a modelling question for every record type, and this is a shape correction.
