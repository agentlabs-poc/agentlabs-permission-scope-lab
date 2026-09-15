# Demonstration — the assignment record

Captured from a real run of `verify-assignment.sh`, then rendered into this file **from
that capture**. The SVG beside it is generated from the same capture, so the
image and this file cannot drift.

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

**Two tables.** The assignment is keyed by what it binds — `key4` the grant,
`key5` and `key6` the recipient — so the one-per-grant-and-recipient rule is the
envelope's own primary key rather than a constraint this record type had to add
to a table it shares with seven others.