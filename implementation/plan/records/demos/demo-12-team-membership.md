# Demonstration — teams and membership

Captured from a real run of `verify-team.sh`, then rendered into this file **from
that capture**. The SVG beside it is generated from the same capture, so the
image and this file cannot drift.

`69` lines captured, 11 commands. Reproduce with:

```sh
$(sess path)/verify-team.sh
```

## CREATE — the id is issued, the caller names the team

```console
abv team create --name Finance
internal projection: team
id  fyb0tqwvmnls
name  Finance
parent  (root)
rc=0
```

## DELETE refuses while anything depends on the team

```console
abv team delete fibggi2juubk
operation conflict
rc=4
```

```console
abv team delete fibggi2juxhc
operation conflict
rc=4
```

## …and succeeds on a leaf nobody is in

```console
abv team delete fyb0tqwvmnls
internal projection: team
removed  fyb0tqwvmnls
rc=0
```

## REPARENT refuses a cycle at the write, not at resolution

```console
abv team reparent fibggi2jur5s --parent fibggi2juxhc
operation rejected or record not found
rc=3
```

```console
abv team reparent fibggi2juxhc --roots
internal projection: team
id  fibggi2juxhc
name  fp8h2w6yan0d
parent  (root)
rc=0
```

## MEMBERSHIP — the pair is the identity, so a repeat is a conflict

```console
abv team add-member --id fibggi2juxhc --human fi7io4lvjqio
internal projection: membership
added  team=fibggi2juxhc  human=fi7io4lvjqio
rc=0
```

```console
abv team add-member --id fibggi2juxhc --human fi7io4lvjqio
operation conflict
rc=4
```

```console
abv team remove-member --id fibggi2juxhc --human fi7io4lvjqio
internal projection: membership
removed  team=fibggi2juxhc  human=fi7io4lvjqio
rc=0
```

```console
abv team remove-member --id fibggi2juxhc --human fi7io4lvjqio
operation rejected or record not found
rc=3
```

## the lost foreign key is now a rule

```console
abv team add-member --id fy6x28qcdrxx --human fi7io4lvjqio
operation rejected or record not found
rc=3
```

## THE ROWS

> boundary  tenant_id     key2         key3            key4                   value
> --------  ---------  ----------  ------------  ----------------  ----------------------------
> tenant    acme       membership  fibggi2juubk  fi7io4lvjqio      {}
> tenant    acme       membership  fibggi2juxhc  fi7io4lvjwu8      {}
> tenant    acme       membership  fibggi2jv0n4  fi7io4lvjqio      {}
> tenant    acme       team        fibggi2jur5s  fp8h2w6ykxan      {"parent_id":""}
> tenant    acme       team        fibggi2juubk  fp8h2w6y5iv8      {"parent_id":"fibggi2jur5s"}
> tenant    acme       team        fibggi2juxhc  fp8h2w6yan0d      {"parent_id":""}
> tenant    acme       team        fibggi2jv0n4  AssignmentAdmins  {"parent_id":""}
