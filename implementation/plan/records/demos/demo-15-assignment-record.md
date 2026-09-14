# Demonstration — the assignment record

Captured from a real run of `verify-assignment.sh` against a real SQLite store,
then rendered into this file **from that capture**. `demo-15-assignment-record.svg`
is generated from the same capture, so the image and this file cannot drift.

`105` lines captured, 13 commands. Reproduce with:

```sh
$(sess path)/verify-assignment.sh
```

## BOTH DIRECTIONS — and neither is not a question

```console
abv assignments list --grant G1
internal projection: assignments
count  1
total  1
G1             group  fibggi2juubk   rev 1    enabled   A1
rc=0
```

```console
abv assignments list --recipient fibggi2juubk --recipient-type group
internal projection: assignments
count  1
total  1
G1             group  fibggi2juubk   rev 1    enabled   A1
rc=0
```

```console
abv assignments list
assignments list requires exactly one of --grant and --recipient: a listing of every assignment is unbounded in the dimension that grows fastest
rc=2
```

## ONE RECORD — the binding identifies it, the id is a handle

```console
abv assignments get A1
internal projection: assignment
grant  G1
recipient  group fibggi2juubk
id  A1
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

> revisions 2 and 3 of G2 published — the assignment has not moved:

```console
abv assign --file /tmp/abvdemo/a2.json --tenant acme --app hrms --db /tmp/abvdemo/asgdemo.db --fixture-context maya-team1
assignment A2 created
rc=0
```

```console
abv assignments get A2
internal projection: assignment
grant  G2
recipient  group fibggi2juxhc
id  A2
adopted revision  1
status  enabled
rc=0
```

```console
abv assignments get A2
internal projection: assignment
grant  G2
recipient  group fibggi2juxhc
id  A2
adopted revision  1
status  enabled
rc=0
```

```console
abv assignments upgrade A2
internal projection: assignment
grant  G2
recipient  group fibggi2juxhc
id  A2
adopted revision  3
status  enabled
rc=0
```

```console
abv assignments upgrade A2
internal projection: assignment
grant  G2
recipient  group fibggi2juxhc
id  A2
adopted revision  3
status  enabled
rc=0
```

## DELETE — refuses while a dependent route rests on it

```console
abv assignments delete A1
operation conflict
rc=4
```

```console
abv assignments delete absent
operation rejected or record not found
rc=3
```

```console
abv assignments delete A2
internal projection: assignment
deleted  A2
rc=0
```

## THE ROWS — the binding is the key path

```text
╭────────────┬──────┬──────────┬───────┬──────────────┬───────────────────────────────────────────────────╮
│    key2    │ key3 │ grant_id │ rtype │  recipient   │                       value                       │
╞════════════╪══════╪══════════╪═══════╪══════════════╪═══════════════════════════════════════════════════╡
│ assignment │ hrms │ G0       │ group │ fibggi2jur5s │ {"id":"A0","grant_revision":1,"status":"enabled"} │
│ assignment │ hrms │ G1       │ group │ fibggi2juubk │ {"id":"A1","grant_revision":1,"status":"enabled"} │
╰────────────┴──────┴──────────┴───────┴──────────────┴───────────────────────────────────────────────────╯
  tables: abv_metadata abv_l1_records applications installations abv_lab_metadata
```

---

## What the run shows

| | Seen above |
|---|---|
| both directions, and neither is not a question | `--grant` and `--recipient` each answer; no filter is refused **with the reason**, not a missing-flag error |
| the binding identifies the record | `get` prints grant and recipient first, the id after — the id is a handle |
| publication does not move an adoption | revisions 2 and 3 of G2 are published and `A2` still reads revision 1 — Q-102 |
| upgrade takes the **latest** | `A2` goes 1 → **3**, never 2. Q-105: *"do not silently select revision 2 instead"* |
| upgrading again is a no-op | the caller asked for the latest and has it; a conflict would make callers special-case it |
| delete refuses while depended on | `A1` supports `G2`'s route from below → `rc=4` |
| the key path is the binding | `key4`=grant, `key5`=recipient type, `key6`=recipient id; the id sits in the value |
| tables 5 → 4 | `assignments` is gone; `abv_metadata`, `abv_l1_records`, `applications`, `installations` remain |

**What this run does not show, and where it is shown instead.** The duplicate
binding above is refused by the lab's administrative gate, which never reaches
storage. That Q-104 is *also* the envelope's primary key — so a writer past the
gate still cannot create a duplicate — is asserted directly in
`internal/storage/sqlite/assignment_l1_test.go`, which calls `insertAssignment`
with no gate in front of it. Saying the CLI proves it would be claiming the
primary key did work it was never asked to do.