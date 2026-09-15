# Demonstration — the assignment record

Captured from a real run of `verify-assignment.sh` against a real SQLite store, then
rendered into this file **from that capture**. The SVG beside it is generated
from the same capture, so the image and this file cannot drift.

`108` lines captured, 13 commands. Reproduce with:

```sh
$(sess path)/verify-assignment.sh
```

## BOTH DIRECTIONS — and neither is not a question

```console
abv assignments list --grant fk3x9r2m5iv8
internal projection: assignments
count  1
total  1
fk3x9r2m5iv8   group  fibggi2juubk   rev 1    enabled   fm5b7t4p5iv8
rc=0
```

```console
abv assignments list --recipient fibggi2juubk --recipient-type group
internal projection: assignments
count  1
total  1
fk3x9r2m5iv8   group  fibggi2juubk   rev 1    enabled   fm5b7t4p5iv8
rc=0
```

```console
abv assignments list
assignments list requires exactly one of --grant and --recipient: a listing of every assignment is unbounded in the dimension that grows fastest
rc=2
```

## ONE RECORD — the binding identifies it, the id is a handle

```console
abv assignments get fm5b7t4p5iv8
internal projection: assignment
grant  fk3x9r2m5iv8
recipient  group fibggi2juubk
id  fm5b7t4p5iv8
adopted revision  1
status  enabled
rc=0
```

## Q-104 — a duplicate binding is refused

> (refused here by the lab's administrative gate, which never reaches
> storage. That Q-104 is ALSO the envelope's primary key — so a writer
> past the gate still cannot create one — is asserted in
> internal/storage/sqlite/assignment_l1_test.go, not here.)

```console
abv assign --file /tmp/abvdemo/dup2.json --tenant acme --app hrms --db /tmp/abvdemo/asgdemo.db --fixture-context maya-team1
operation rejected or record not found
rc=3
```

## UPGRADE — the latest, never an intermediate

> revisions 2 and 3 of fk3x9r2man0d published — the assignment has not moved:

```console
abv assign --file /tmp/abvdemo/a2.json --tenant acme --app hrms --db /tmp/abvdemo/asgdemo.db --fixture-context maya-team1
assignment fm5b7t4pan0d created
rc=0
```

```console
abv assignments get fm5b7t4pan0d
internal projection: assignment
grant  fk3x9r2man0d
recipient  group fibggi2juxhc
id  fm5b7t4pan0d
adopted revision  1
status  enabled
rc=0
```

```console
abv assignments get fm5b7t4pan0d
internal projection: assignment
grant  fk3x9r2man0d
recipient  group fibggi2juxhc
id  fm5b7t4pan0d
adopted revision  1
status  enabled
rc=0
```

```console
abv assignments upgrade fm5b7t4pan0d
internal projection: assignment
grant  fk3x9r2man0d
recipient  group fibggi2juxhc
id  fm5b7t4pan0d
adopted revision  3
status  enabled
rc=0
```

```console
abv assignments upgrade fm5b7t4pan0d
internal projection: assignment
grant  fk3x9r2man0d
recipient  group fibggi2juxhc
id  fm5b7t4pan0d
adopted revision  3
status  enabled
rc=0
```

## DELETE — refuses while a dependent route rests on it

```console
abv assignments delete fm5b7t4p5iv8
operation conflict
rc=4
```

```console
abv assignments delete absent
operation rejected or record not found
rc=3
```

```console
abv assignments delete fm5b7t4pan0d
internal projection: assignment
deleted  fm5b7t4pan0d
rc=0
```

## THE ROWS — the binding is the key path

> (each heading names the real column and what it holds. there is no
> grant_id or recipient column — key4, key5 and key6 ARE the binding.)

```text
╭────────────┬──────────┬──────────────┬────────────────┬────────────────┬─────────────────────────────────────────────────────────────╮
│ key2 type  │ key3 app │  key4 grant  │ key5 rcpt type │ key6 recipient │                            value                            │
╞════════════╪══════════╪══════════════╪════════════════╪════════════════╪═════════════════════════════════════════════════════════════╡
│ assignment │ hrms     │ fk3x9r2m0dq3 │ group          │ fibggi2jur5s   │ {"id":"fm5b7t4p0dq3","grant_revision":1,"status":"enabled"} │
│ assignment │ hrms     │ fk3x9r2m5iv8 │ group          │ fibggi2juubk   │ {"id":"fm5b7t4p5iv8","grant_revision":1,"status":"enabled"} │
╰────────────┴──────────┴──────────────┴────────────────┴────────────────┴─────────────────────────────────────────────────────────────╯
  tables: abv_metadata abv_l1_records applications installations abv_lab_metadata
```

---

## What the run shows

| | Seen above |
|---|---|
| both directions, and neither is not a question | `--grant` and `--recipient` each answer; no filter is refused **with the reason** |
| the binding identifies the record | `get` prints grant and recipient first, the id after |
| publication does not move an adoption | two revisions published, the assignment still reads revision 1 — Q-102 |
| upgrade takes the **latest** | 1 → **3**, never 2. Q-105: *"do not silently select revision 2 instead"* |
| upgrading again is a no-op | the caller asked for the latest and has it |
| delete refuses while depended on | the supporting assignment cannot go while a route rests on it — `rc=4` |
| every id is a base-36 Snowflake | `key4`, `key6` and the value's `id` alike |
| tables 5 → 4 | `assignments` is gone |

**What this run does not show, and where it is shown instead.** The duplicate
binding is refused by the lab's administrative gate, which never reaches storage.
That Q-104 is *also* the envelope's primary key is asserted in
`internal/storage/sqlite/assignment_l1_test.go`, which calls `insertAssignment`
with no gate in front of it.