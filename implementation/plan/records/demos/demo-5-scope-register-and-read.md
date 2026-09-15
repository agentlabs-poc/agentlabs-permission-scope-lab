# Demonstration — the scope record

Captured from a real run of `verify-scope.sh`, then rendered into this file **from
that capture**. The SVG beside it is generated from the same capture, so the
image and this file cannot drift.

`69` lines captured, 13 commands. Reproduce with:

```sh
$(sess path)/verify-scope.sh
```

## register: add-only, wildcard rejected anywhere in the key

```console
abv catalog register-scope region
internal projection: scope
key  region
rc=0
```

```console
abv catalog register-scope region
operation conflict
rc=4
```

```console
abv catalog register-scope bad*
malformed input
rc=2
```

```console
abv catalog register-scope *
malformed input
rc=2
```

## read: typed, by exact key

```console
abv catalog get-scope dept
internal projection: scope
key  dept
rc=0
```

```console
abv catalog get-scope missing
operation rejected or record not found
rc=3
```

```console
abv catalog get-scope $self
operation rejected or record not found
rc=3
```

## list: ordered by key, offset paged, bounded

```console
abv catalog list-scopes
internal projection: scopes
count  4
total  4
generation  1
cert
dept
region
user
rc=0
```

```console
abv catalog list-scopes --offset 2 --limit 2
internal projection: scopes
count  2
total  4
generation  1
region
user
rc=0
```

```console
abv catalog list-scopes --offset 99
internal projection: scopes
count  0
total  4
generation  1
rc=0
```

```console
abv catalog list-scopes --limit 501
malformed input
rc=2
```

## one L1 record store, two record types

```console
abv catalog register-permission hrms:payroll::read
internal projection: permission
id  hrms:payroll::read
active  true
rc=0
```

```console
abv catalog list-permissions
internal projection: permissions
count  4
total  4
generation  2
hrms:payroll::read  active=true
hrms:payroll:payslip::delete  active=true
hrms:payroll:payslip::read  active=true
hrms:payroll:payslip::write  active=true
rc=0
```
