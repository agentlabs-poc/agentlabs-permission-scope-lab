# Demonstration — the grant record

Captured from a real run of `verify-grant.sh`, then rendered into this file **from
that capture**. The SVG beside it is generated from the same capture, so the
image and this file cannot drift.

`138` lines captured, 15 commands. Reproduce with:

```sh
$(sess path)/verify-grant.sh
```

## WHAT IS HERE — the root is told apart from its children

```console
abv grants list
internal projection: grants
count  3
total  3
fk3x9r2m0dq3   enabled   root
fk3x9r2m5iv8   enabled   child
fk3x9r2man0d   enabled   child
rc=0
```

```console
abv grants list --roots
internal projection: grants
count  1
total  1
fk3x9r2m0dq3   enabled   root
rc=0
```

```console
abv grants list --children
internal projection: grants
count  2
total  2
fk3x9r2m5iv8   enabled   child
fk3x9r2man0d   enabled   child
rc=0
```

## THE ROOT — no parent, and no narrowing

> (the list below is what the fixture seeded. rootRoute ignores it and
> computes from the catalog; new root content omits the field entirely.)

```console
abv grants get fk3x9r2m0dq3 --revision 1
internal projection: grant
id  fk3x9r2m0dq3
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
abv grants get fk3x9r2m5iv8 --revision 1
internal projection: grant
id  fk3x9r2m5iv8
status  enabled
kind  child
revision  1
parent  fk3x9r2m0dq3
permissions  hrms:payroll:payslip::read, hrms:payroll:payslip::write
scope  dept=FIN
rc=0
```

```console
abv grants get fk3x9r2man0d --revision 1
internal projection: grant
id  fk3x9r2man0d
status  enabled
kind  child
revision  1
parent  fk3x9r2m5iv8
permissions  hrms:payroll:payslip::read
scope  cert=C17
rc=0
```

## CREATE — head and revision 1 in one act, the id issued

```console
abv grants create --parent fk3x9r2m5iv8 --permissions hrms:payroll:payslip::read --scope cert=C99
internal projection: grant
id  fyahgo3fcuf4
status  enabled
kind  child
revision  1
parent  fk3x9r2m5iv8
permissions  hrms:payroll:payslip::read
scope  cert=C99
rc=0
```

```console
abv grants create --parent fk3x9r2m5iv8 --permissions hrms:payroll:payslip::read --scope dept=FIN
internal projection: grant
id  fyahgo53aeps
status  enabled
kind  child
revision  1
parent  fk3x9r2m5iv8
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
abv grants create --parent fk3x9r2m5iv8 --permissions hrms:payroll:payslip::export
operation rejected or record not found
rc=3
```

```console
abv grants create --parent fk3x9r2m5iv8 --permissions hrms:payroll:payslip::read --scope branch=B1
operation rejected or record not found
rc=3
```

## DELETE — refuses while anything depends on it

```console
abv grants delete fk3x9r2m5iv8
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
abv grants revisions fk3x9r2m5iv8
internal projection: grant revisions
grant  fk3x9r2m5iv8
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
│ grant          │ hrms │ fk3x9r2m0dq3 │            │ {"status":"enabled","trusted_root":true}       │
│ grant_revision │ hrms │ fk3x9r2m0dq3 │ 0000000001 │ {"permissions":["hrms:payroll:payslip::read"," │
│ grant          │ hrms │ fk3x9r2m5iv8 │            │ {"status":"enabled","trusted_root":false}      │
│ grant_revision │ hrms │ fk3x9r2m5iv8 │ 0000000001 │ {"parent_grant_id":"fk3x9r2m0dq3","permissions │
│ grant          │ hrms │ fk3x9r2man0d │            │ {"status":"enabled","trusted_root":false}      │
│ grant_revision │ hrms │ fk3x9r2man0d │ 0000000001 │ {"parent_grant_id":"fk3x9r2m5iv8","permissions │
│ grant          │ hrms │ fyahgo3fcuf4 │            │ {"status":"enabled","trusted_root":false}      │
│ grant_revision │ hrms │ fyahgo3fcuf4 │ 0000000001 │ {"parent_grant_id":"fk3x9r2m5iv8","permissions │
│ grant          │ hrms │ fyahgo53aeps │            │ {"status":"enabled","trusted_root":false}      │
│ grant_revision │ hrms │ fyahgo53aeps │ 0000000001 │ {"parent_grant_id":"fk3x9r2m5iv8","permissions │
╰────────────────┴──────┴──────────────┴────────────┴────────────────────────────────────────────────╯
  tables: abv_metadata abv_l1_records abv_lab_metadata
```

---

**Two tables.** `abv_metadata` and `abv_l1_records` — the ownership marker and
one canonical record store. `abv_lab_metadata` is the lab scenario's own marker,
not Auth-AL's.