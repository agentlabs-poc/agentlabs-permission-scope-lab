# Demonstration — the grant record

Captured from a real run of `verify-grant.sh` against a real SQLite store, then
rendered into this file **from that capture**. Every command and every line of
output below was produced by the run; nothing is transcribed from memory.
`demo-14-grant-record.svg` is generated from the same capture, so the image and
this file cannot drift apart — the issued identifiers match in both.

`138` lines captured, 15 commands. Reproduce with:

```sh
$(sess path)/verify-grant.sh
```

| Exit code | Means |
|---|---|
| `0` | the operation happened |
| `2` | malformed input — refused before the store was touched |
| `3` | rejected, or no such record |
| `4` | conflict — something depends on it, or it already exists |

## WHAT IS HERE — the root is told apart from its children

```console
abv grants list
internal projection: grants
count  3
total  3
G0             enabled   root
G1             enabled   child
G2             enabled   child
rc=0
```

```console
abv grants list --roots
internal projection: grants
count  1
total  1
G0             enabled   root
rc=0
```

```console
abv grants list --children
internal projection: grants
count  2
total  2
G1             enabled   child
G2             enabled   child
rc=0
```

## THE ROOT — no parent, and no narrowing

> (the list below is what the fixture seeded. rootRoute ignores it and
> computes from the catalog; new root content omits the field entirely.)

```console
abv grants get G0 --revision 1
internal projection: grant
id  G0
status  enabled
kind  root
revision  1
parent  (none — root)
permissions  hrms:payroll:payslip::read, hrms:payroll:payslip::write, hrms:payroll:payslip::delete
scope  {} (adds no narrowing)
rc=0
```

## A CHILD — it states its own, and narrows

```console
abv grants get G1 --revision 1
internal projection: grant
id  G1
status  enabled
kind  child
revision  1
parent  G0
permissions  hrms:payroll:payslip::read, hrms:payroll:payslip::write
scope  dept=FIN
rc=0
```

```console
abv grants get G2 --revision 1
internal projection: grant
id  G2
status  enabled
kind  child
revision  1
parent  G1
permissions  hrms:payroll:payslip::read
scope  cert=C17
rc=0
```

## CREATE — head and revision 1 in one act, the id issued

```console
abv grants create --parent G1 --permissions hrms:payroll:payslip::read --scope cert=C99
internal projection: grant
id  fy85p22i8glc
status  enabled
kind  child
revision  1
parent  G1
permissions  hrms:payroll:payslip::read
scope  cert=C99
rc=0
```

```console
abv grants create --parent G1 --permissions hrms:payroll:payslip::read --scope dept=FIN
internal projection: grant
id  fy85p243o4jk
status  enabled
kind  child
revision  1
parent  G1
permissions  hrms:payroll:payslip::read
scope  dept=FIN
rc=0
```

## CREATE REFUSES — and each refusal is its own answer

```console
abv grants create --permissions hrms:payroll:payslip::read
grants create requires --parent: a parentless grant is not a root, and establishing one is not a grant operation
rc=2
```

```console
abv grants create --parent absent --permissions hrms:payroll:payslip::read
operation rejected or record not found
rc=3
```

```console
abv grants create --parent G1 --permissions hrms:payroll:payslip::export
operation rejected or record not found
rc=3
```

```console
abv grants create --parent G1 --permissions hrms:payroll:payslip::read --scope branch=B1
operation rejected or record not found
rc=3
```

## DELETE — refuses while anything depends on it

```console
abv grants delete G1
operation conflict
rc=4
```

```console
abv grants delete absent
operation rejected or record not found
rc=3
```

## REVISIONS — newest first

```console
abv grants revisions G1
internal projection: grant revisions
grant  G1
count  1
total  1
revision 1    2 permissions
rc=0
```

## THE ROWS — one table, two record types, told apart by key2

```text
╭────────────────┬──────┬──────────────┬────────────┬────────────────────────────────────────────────╮
│      key2      │ key3 │     key4     │    key5    │                     value                      │
╞════════════════╪══════╪══════════════╪════════════╪════════════════════════════════════════════════╡
│ grant          │ hrms │ G0           │            │ {"status":"enabled","trusted_root":true}       │
│ grant_revision │ hrms │ G0           │ 0000000001 │ {"permissions":["hrms:payroll:payslip::read"," │
│ grant          │ hrms │ G1           │            │ {"status":"enabled","trusted_root":false}      │
│ grant_revision │ hrms │ G1           │ 0000000001 │ {"parent_grant_id":"G0","permissions":["hrms:p │
│ grant          │ hrms │ G2           │            │ {"status":"enabled","trusted_root":false}      │
│ grant_revision │ hrms │ G2           │ 0000000001 │ {"parent_grant_id":"G1","permissions":["hrms:p │
│ grant          │ hrms │ fy85p22i8glc │            │ {"status":"enabled","trusted_root":false}      │
│ grant_revision │ hrms │ fy85p22i8glc │ 0000000001 │ {"parent_grant_id":"G1","permissions":["hrms:p │
│ grant          │ hrms │ fy85p243o4jk │            │ {"status":"enabled","trusted_root":false}      │
│ grant_revision │ hrms │ fy85p243o4jk │ 0000000001 │ {"parent_grant_id":"G1","permissions":["hrms:p │
╰────────────────┴──────┴──────────────┴────────────┴────────────────────────────────────────────────╯
  tables: abv_metadata abv_l1_records applications installations assignments abv_lab_metadata
```

---

## What the run shows

| | Seen above |
|---|---|
| the root is a different kind of row | `grants list` prints `root` against `G0` and `child` against the rest; `--roots` and `--children` are the same question asked either way |
| a root neither inherits nor narrows | `G0` prints `parent  (none — root)` and `scope  {} (adds no narrowing)` |
| a child states its permissions | `G1` takes two of the root's three — `delete` was available and not taken |
| scope accumulates | `G2` narrows to `cert=C17` **under** `G1`'s `dept=FIN`; it cannot drop the parent's constraint |
| create writes both records | one command, and the rows section shows a `grant` **and** a `grant_revision` for each new id |
| the identifier is issued | `fy85p22i8glc`, not a name the caller chose |
| four refusals, four answers | no parent `rc=2` · unknown parent `rc=3` · unregistered permission `rc=3` · unregistered scope key `rc=3` |
| delete refuses while depended on | `G1` has a child and an assignment → `rc=4`, not a cascade |
| one table, two record types | told apart by `key2`, with the revision zero-padded in `key5` |

**The parentless refusal is the one worth reading twice.** It is refused by the
command line itself, with a message rather than a missing-flag error: a
parentless grant is not a root, and there is no invocation of this verb that
could establish one. That is Q-113 expressed as vocabulary — no grant operation
writes trust evidence, so grant administration cannot confer root authority.

**What the run cannot show.** The root's stored permission list is the fixture's,
and resolution ignores it — but proving that needs a resolved route, which needs
assignments. The same for establishing a root at all: it writes four things and
the fourth is a holder assignment. Both belong to the assignment slice.