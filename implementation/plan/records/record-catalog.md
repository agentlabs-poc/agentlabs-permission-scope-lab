# `abv.catalog` — the last fold

**Implemented.** After this, `abv_l1_records` and `abv_metadata` are the only tables
Auth-AL has, and the 123 shape is complete: one canonical key/value record store
per domain per layer, plus the ownership marker.

---

## 1 · The theory — why a table survives a fold

A table survives a fold for exactly one reason: **it holds a fact nobody has
found a home for.** Not because it is useful, not because dropping it is
awkward — because the fact is real and the envelope has no slot for it yet.

That is the whole test, and it is worth stating because these two tables fail it
differently.

**The envelope holds records.** A record is a fact with an identity: something
you can name, that persists, and that a canonical path can address. `key1…key10`
is the name; `value` is the fact. Anything shaped that way belongs in the
envelope, and anything not shaped that way is either not a record — it is an
event, a counter, a cache — or it is a record nobody has named yet.

So the question for each table is not "can we delete it" but **"what fact is in
it, and is that fact a record?"**

---

## 2 · `installations` — no fact of its own

```sql
CREATE TABLE installations (tenant_id, application_id, PRIMARY KEY (...));
```

Presence is the entire content. It says *this tenant has this application*, and
**that is the registry's fact, not Auth-AL's** — `application_registry_l1_records`
already holds it as an `installation` record, and the `Registry` port already
answers it.

Auth-AL reads it in one place:

```go
snapshot.go:43   SELECT 1 FROM installations WHERE tenant_id=? AND application_id=?
```

and that read **already routes through the port when one is wired**. The table is
the fallback for an Auth-AL opened without a registry.

> **So this is not a fold at all.** There is nothing to move. It is a second copy
> of another domain's record, kept as a fallback, and the cost of keeping it is
> that two stores can disagree about whether a tenant has an application — with
> Auth-AL believing its own.

**What it takes:** make the port mandatory. `OpenSQLite` without a registry
either becomes an error, or keeps working with `Installed` always false, which is
fail-closed and honest. Then delete the table.

---

## 3 · `applications` — two facts, and only one is a record

```sql
CREATE TABLE applications (
    application_id TEXT PRIMARY KEY,
    compatibility_enabled INTEGER NOT NULL,
    generation INTEGER NOT NULL DEFAULT 0
);
```

Three columns, three different situations.

### 3.1 · `application_id` — the registry's, like installations

Read at `catalog.go:189` to answer *does this application exist*, and already
routed through the port's `ApplicationExists`. Same story as §2: not a fold, a
duplicate.

### 3.2 · `compatibility_enabled` — a record, and it has been one all along

Q-041, approved: *"each application explicitly declares upfront whether
relationship validation is enabled or disabled."*

Read that sentence as a record and it names itself. A **declaration**, made by
the application, at the application boundary, persisting until changed. It has an
identity — the application — and a value. It is a record that has been living in
a column.

```
key1=abv  key2=catalog  key3=<application>
value={"compatibility_enabled": true}
boundary=application, tenant_id=''
```

**Why `boundary=application` and not `tenant`:** the declaration is the
application's, identical for every tenant, exactly as a permission definition is.

### 3.3 · `generation` — not a record, and this is the interesting one

`generation` is a counter bumped in the same transaction as any catalog write:

```go
catalog.go:286   UPDATE applications SET generation = generation + 1 WHERE application_id = ?
catalog.go:72    UPDATE applications SET generation = generation + 1          -- every application
```

It exists for one guarantee, and the code says it plainly:

> *"Generation is the application catalog's version at the moment of the read. It
> is unchanged across pages exactly when nothing was written between them, which
> is what makes an offset walk safe to cache: read a generation, page through,
> read it again, and retry if it moved."*

**A counter is not a record.** It has no independent existence — it is a
derived property of *the catalog*, meaningful only as "has this changed since I
looked". Nothing addresses it, nothing reads it for its own sake, and its value
carries no information beyond inequality with a previous value.

Putting it in the envelope as a value field works mechanically and is wrong for
the same reason `state` and `ts` sitting unread is wrong: it makes the envelope
carry something that is not a record because there was nowhere else to put it.

**Three options, and the third is the one to argue for.**

| | | |
|---|---|---|
| **A** | a field on the `catalog` record | works; puts a counter in a record store, and a catalog write becomes a read-modify-write of an unrelated record |
| **B** | keep one small table for it | honest, but it is the table we are trying to delete, renamed |
| **C** | **derive it** — `MAX(ts)` or `COUNT(*)` over that application's catalog rows | no counter at all; the guarantee comes from the records themselves |

**C is right if it holds.** The guarantee needed is *"unchanged across two reads
exactly when nothing was written between them."* A count of an application's
catalog rows gives that for insertions but **not** for a status change, which
rewrites a value without changing the count. `MAX(ts)` catches both — but `ts`
has one-second resolution, and two writes in the same second would be
indistinguishable.

So C needs one of: a monotonic write stamp finer than a second, or
`COUNT(*) || MAX(ts)` together, which is still defeatable by a same-second status
flip on a fixed number of rows.

> **This is the real decision in this proposal.** `generation` is where the 123
> architecture meets a guarantee that is not about records, and it is the first
> such case. The answer sets the precedent for every domain that later wants a
> change-detection signal — and the registry already has the same need
> (`RegisterApplication` invalidates a cached listing exactly as a catalog write
> does).
>
> **Recommendation: A, with the reason recorded.** Take the counter into the
> `catalog` record, note plainly that it is a counter in a record store
> and why, and revisit it when `ts` earns a finer stamp — which the deferred audit
> question will force anyway. C is better and cannot be adopted honestly until
> `ts` can distinguish two writes in the same second.

---

## 4 · What this actually costs

| # | Step | |
|---|---|---|
| 1 | `abv.catalog` record + storage | small — a two-field value, the shape `abv.grant`'s head already has |
| 2 | `readCatalog`'s compat and generation reads re-pointed | two queries |
| 3 | `bumpGeneration` becomes a value update | the per-application one is direct; the **all-applications** one at `catalog.go:72` becomes an `UPDATE … WHERE key2='catalog'` |
| 4 | Make the `Registry` port mandatory | the decision, not the code |
| 5 | Drop both tables, assert they stay gone | one migration, one test |
| 6 | Re-capture the demonstrations | tables 4 → 2 is visible in every one |

**Step 4 is the only one that is a judgement call.** The rest follows.

---

## 5 · What this does NOT do

It does not make Auth-AL depend on the registry. The port stays a port; making it
mandatory means *a registry must be wired*, not *this registry must be wired* —
the legacy tables behind an adapter remain a valid answer, which was the whole
point of the seam.

---

## 6 · What was decided

1. **`installations`** — the port is mandatory and the table is gone. There is
   one way to open a store and it takes a registry, so the requirement is
   structural rather than a runtime check.
2. **`compatibility_enabled`** — `abv.catalog` at the application boundary. The
   name is `catalog`, not `catalog`: `domain.Catalog` already names
   exactly these fields, and every other record type here is a single noun.
3. **`generation`** — option **A**, with the reason in the code rather than only
   here. C is better and cannot be adopted while `ts` has one-second resolution.

After all three: **two tables**, `abv_metadata` and `abv_l1_records`. Which is
what the 123 architecture says a domain is.
