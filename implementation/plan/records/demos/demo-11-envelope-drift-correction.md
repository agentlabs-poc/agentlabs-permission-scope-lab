# Demonstration — the envelope drift correction

Captured from a real run of `verify-drift.sh`, then rendered into this file **from
that capture**. The SVG beside it is generated from the same capture, so the
image and this file cannot drift.

`62` lines captured, 3 commands. Reproduce with:

```sh
$(sess path)/verify-drift.sh
```

## THE RULE: a permission's first noun IS its application

```console
abv hrms:employee:certificate::read --app hrms --db drift.db --fixture-context application-publisher
internal projection: permission
id  hrms:employee:certificate::read
active  true
rc=0
```

```console
abv billing:invoice::read --app hrms --db drift.db --fixture-context application-publisher
operation rejected or record not found
rc=3
```

```console
abv reporting:ledger:entry::export --app hrms --db drift.db --fixture-context application-publisher
operation rejected or record not found
rc=3
```

## EVERY RECORD TYPE, ONE SHAPE

> tenant  key1       key2           key3            key4           key5           key6       key10
> ------  ----  --------------  ------------  ----------------  -----------  --------------  ------
> acme    abv   assignment      hrms          fk3x9r2m0dq3      group        fibggi2jur5s
> acme    abv   assignment      hrms          fk3x9r2m5iv8      group        fibggi2juubk
> (none)  abv   catalog         hrms
> acme    abv   grant           hrms          fk3x9r2m0dq3
> acme    abv   grant           hrms          fk3x9r2m5iv8
> acme    abv   grant           hrms          fk3x9r2man0d
> acme    abv   grant_revision  hrms          fk3x9r2m0dq3      0000000001
> acme    abv   grant_revision  hrms          fk3x9r2m5iv8      0000000001
> acme    abv   grant_revision  hrms          fk3x9r2man0d      0000000001
> acme    abv   membership      fibggi2juubk  fi7io4lvjqio
> acme    abv   membership      fibggi2jv0n4  fi7io4lvjqio
> acme    abv   membership      fibggi2juxhc  fi7io4lvjwu8
> (none)  abv   permission      hrms          employee          certificate                  read
> (none)  abv   permission      hrms          payroll           payslip                      delete
> (none)  abv   permission      hrms          payroll           payslip                      read
> (none)  abv   permission      hrms          payroll           payslip                      write
> acme    abv   role            hrms          fi9jvxobqsxs      0000000001   payslip-reader
> (none)  abv   role            hrms          fyb0tqqfe3gg      0000000001   Viewer
> (none)  abv   scope           hrms          cert
> (none)  abv   scope           hrms          dept
> (none)  abv   scope           hrms          region
> (none)  abv   scope           hrms          user
> acme    abv   team            fibggi2jv0n4  AssignmentAdmins
> acme    abv   team            fibggi2juubk  fp8h2w6y5iv8
> acme    abv   team            fibggi2juxhc  fp8h2w6yan0d
> acme    abv   team            fibggi2jur5s  fp8h2w6ykxan

## key3 is the application in every row — no key3/key4 duplication

> --------------  ----  -------------  ------------  -------------
> assignment         2              1  hrms                      2
> catalog            1              1  hrms                      1
> grant              3              1  hrms                      3
> grant_revision     3              1  hrms                      3
> membership         3              3  fibggi2juubk              2
> permission         4              1  hrms                      2
> role               2              1  hrms                      2
> scope              4              1  hrms                      4
> team               4              4  fibggi2jur5s              4

```text
     key2       rows  distinct_key3      key3      distinct_key4
```

## THE ENVELOPE

> columns: boundary tenant_id key1 key2 key3 key4 key5 key6 key7 key8 key9 key10 value ts state
> boundary        STILL PRESENT
> application_id  gone
> revision        gone

```text
  tables:  2 from the migration, +1 lab marker the scenario seed adds
```
