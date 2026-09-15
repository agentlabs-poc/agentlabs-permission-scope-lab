# Demonstration — root establishment

Captured from a real run of `verify-root.sh`, then rendered into this file **from
that capture**. The SVG beside it is generated from the same capture, so the
image and this file cannot drift.

`165` lines captured, 20 commands. Reproduce with:

```sh
$(sess path)/verify-root.sh
```

## GENESIS — a tenant with teams and not one grant

```console
abv team list --roots --app hrms
internal projection: teams
count  2
total  2
generation  0
fibggi2jur5s  fp8h2w6ykxan       parent=(root)
fibggi2jv0n4  AssignmentAdmins   parent=(root)
rc=0
```

```console
abv grants list --app hrms
internal projection: grants
count  0
total  0
rc=0
```

## ESTABLISHING IS NOT A GRANT OPERATION — Q-113 as vocabulary

```console
abv grants create --permissions hrms:payroll:payslip::read --scope dept=FIN --app hrms
grants create requires --parent: a parentless grant is not a root, and establishing one is not a grant operation
rc=2
```

## LINEAGE A — the Auth root, written from outside the tenant

> (Q-114: an empty catalog computes an empty ceiling, and a ceiling of
> nothing is not a ceiling. Registration precedes acceptance.)

```console
abv root establish-auth --team fibggi2jur5s --app auth
operation rejected or record not found
rc=3
```

```console
abv catalog register-platform-permission auth:tenant:application::install --namespace auth
internal projection: permission
id  auth:tenant:application::install
namespace  auth
boundary  platform
active  true
rc=0
```

```console
abv catalog register-platform-permission auth:grant::create --namespace auth
internal projection: permission
id  auth:grant::create
namespace  auth
boundary  platform
active  true
rc=0
```

```console
abv root establish-auth --team fibggi2jur5s --app auth
internal projection: grant
id  fyb9ahvgyfpc
status  enabled
kind  root
revision  1
parent  (none — root)
permissions  (computed from the catalog)
scope  {} (adds no narrowing)
rc=0
```

```console
abv assignments list --grant fyb9ahvgyfpc --app auth
internal projection: assignments
count  1
total  1
fyb9ahvgyfpc   group  fibggi2jur5s   rev 1    enabled   fyb9ahvgyfpd
rc=0
```

```console
abv root establish-auth --team fibggi2jur5s --app auth
operation conflict
rc=4
```

## LINEAGE B — the application root, written by the tenant administrator

> (A root is held by a top-level team. rootRoute refuses a parented
> holder at resolution; establishment refuses it at the write, so the
> record is never stored in a shape resolution would reject.)

```console
abv root establish --team fibggi2juubk --app hrms
operation rejected or record not found
rc=3
```

```console
abv root establish --team fp8h2w6yzzzz --app hrms
operation rejected or record not found
rc=3
```

```console
abv root establish --team not-a-snowflake --app hrms
malformed input
rc=2
```

```console
abv root establish --team fibggi2jur5s --app hrms
internal projection: grant
id  fyb9ai0y8f0g
status  enabled
kind  root
revision  1
parent  (none — root)
permissions  (computed from the catalog)
scope  {} (adds no narrowing)
rc=0
```

## THE CATALOG IS NOT THE CEILING — decision 3, made visible

> (hrms's catalog is its own permissions UNION every platform one, which
> is right for evaluation: a request inside hrms may legitimately need a
> platform permission. It is wrong for a ceiling. Creation checks the
> catalog; resolution checks the ceiling — so both children below are
> written, and only one of them resolves.)

```console
abv catalog list-permissions
internal projection: permissions
count  5
total  5
generation  2
auth:grant::create  active=true
auth:tenant:application::install  active=true
hrms:payroll:payslip::delete  active=true
hrms:payroll:payslip::read  active=true
hrms:payroll:payslip::write  active=true
rc=0
```

```console
abv grants create --parent fyb9ai0y8f0g --permissions hrms:payroll:payslip::read --scope dept=FIN --app hrms
internal projection: grant
id  fyb9ai453qww
status  enabled
kind  child
revision  1
parent  fyb9ai0y8f0g
permissions  hrms:payroll:payslip::read
scope  dept=FIN
rc=0
```

```console
abv check assignment --file /tmp/abvdemo/proposal.json --app hrms
READ-ONLY DIAGNOSIS: proposal is structurally and lineally valid; administrative and source authority are not established
This does not authorize a later write.
internal projection: proposed route
field           value                       source
grant_id        fyb9ai453qww
permissions     hrms:payroll:payslip::read
scope.dept      FIN                         fyb9ai453qww
assignment_ids  fyb9ai0y8f0h
rc=0
```

```console
abv grants create --parent fyb9ai0y8f0g --permissions auth:grant::create --app hrms
internal projection: grant
id  fyb9ai6ailmo
status  enabled
kind  child
revision  1
parent  fyb9ai0y8f0g
permissions  auth:grant::create
scope  {} (adds no narrowing)
rc=0
```

```console
abv check assignment --file /tmp/abvdemo/proposal.json --app hrms
operation rejected or record not found
rc=3
```

## THE SAME PERMISSION, THE OTHER ROOT

> (Identical request, identical permission, one area across. The route
> traces to the Auth root's own assignment — the lineage the tenant
> administrator holds, which is not the one hrms bounds.)

```console
abv grants create --parent fyb9ahvgyfpc --permissions auth:grant::create --app auth
internal projection: grant
id  fyb9ai8axnnk
status  enabled
kind  child
revision  1
parent  fyb9ahvgyfpc
permissions  auth:grant::create
scope  {} (adds no narrowing)
rc=0
```

```console
abv check assignment --file /tmp/abvdemo/proposal.json --app auth
READ-ONLY DIAGNOSIS: proposal is structurally and lineally valid; administrative and source authority are not established
This does not authorize a later write.
internal projection: proposed route
field           value               source
grant_id        fyb9ai8axnnk
permissions     auth:grant::create
assignment_ids  fyb9ahvgyfpd
rc=0
```

## THE ROWS — three records per root, one transaction

```console
╭──────┬────────────────┬──────────────┬─────────────────────────────────────────────────────────────╮
│ area │      type      │      id      │                            value                            │
╞══════╪════════════════╪══════════════╪═════════════════════════════════════════════════════════════╡
│ hrms │ assignment     │ fyb9ai0y8f0g │ {"id":"fyb9ai0y8f0h","grant_revision":1,"status":"enabled"} │
│ hrms │ grant          │ fyb9ai0y8f0g │ {"status":"enabled","trusted_root":true}                    │
│ hrms │ grant_revision │ fyb9ai0y8f0g │ {"scope":{}}                                                │
│ auth │ assignment     │ fyb9ahvgyfpc │ {"id":"fyb9ahvgyfpd","grant_revision":1,"status":"enabled"} │
│ auth │ grant          │ fyb9ahvgyfpc │ {"status":"enabled","trusted_root":true}                    │
│ auth │ grant_revision │ fyb9ahvgyfpc │ {"scope":{}}                                                │
╰──────┴────────────────┴──────────────┴─────────────────────────────────────────────────────────────╯
  roots: auth=fyb9ahvgyfpc  hrms=fyb9ai0y8f0g
```
