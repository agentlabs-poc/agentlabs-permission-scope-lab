# Demonstration — the ownership record

Captured from a real run of `verify-ownership.sh`, then rendered into this file **from
that capture**. The SVG beside it is generated from the same capture, so the
image and this file cannot drift.

`97` lines captured, 14 commands. Reproduce with:

```sh
$(sess path)/verify-ownership.sh
```

## OWNERS ARE PLURAL — a team has several, a human owns several

```console
abv owners add --team fibggi2juubk --human fi7io4lvjqio
internal projection: ownership
fi7io4lvjqio owns fibggi2juubk
rc=0
```

```console
abv owners add --team fibggi2juubk --human fi7io4lvjwu8
internal projection: ownership
fi7io4lvjwu8 owns fibggi2juubk
rc=0
```

```console
abv owners add --team fibggi2juxhc --human fi7io4lvjqio
internal projection: ownership
fi7io4lvjqio owns fibggi2juxhc
rc=0
```

```console
abv owners list --team fibggi2juubk
internal projection: owners
count  2
total  2
team=fibggi2juubk   human=fi7io4lvjqio
team=fibggi2juubk   human=fi7io4lvjwu8
rc=0
```

```console
abv owners list --human fi7io4lvjqio
internal projection: owners
count  2
total  2
team=fibggi2juubk   human=fi7io4lvjqio
team=fibggi2juxhc   human=fi7io4lvjqio
rc=0
```

## ADD IS ADD-ONLY, REMOVE SAYS WHETHER IT DID ANYTHING

```console
abv owners add --team fibggi2juubk --human fi7io4lvjqio
operation conflict
rc=4
```

```console
abv owners remove --team fibggi2juubk --human fi7io4lvjwu8
internal projection: ownership
fi7io4lvjwu8 no longer owns fibggi2juubk
rc=0
```

```console
abv owners remove --team fibggi2juubk --human fi7io4lvjwu8
operation rejected or record not found
rc=3
```

## BOTH HALVES MUST EXIST, AND NEITHER DIRECTION IS OPTIONAL

```console
abv owners add --team fp8h2w6yzzzz --human fi7io4lvjqio
operation rejected or record not found
rc=3
```

```console
abv owners add --team fibggi2juubk --human not-a-snowflake
malformed input
rc=2
```

```console
abv owners list
owners list requires exactly one of --team and --human: a team's owners, or a human's teams
rc=2
```

## OWNING IS NOT MEMBERSHIP — Q-099, the whole point

> (maya owns Team2 without being one of its members, and nutan is a member
> of Team2 while owning nothing at all. Neither relationship implies the
> other, which is what Q-099 exists to say.)

```console
abv owners list --team fibggi2juxhc
internal projection: owners
count  1
total  1
team=fibggi2juxhc   human=fi7io4lvjqio
rc=0
```

```console
abv team members --id fibggi2juxhc
internal projection: members
count  1
total  1
generation  0
team=fibggi2juxhc  human=fi7io4lvjwu8
rc=0
```

```console
abv owners list --human fi7io4lvjwu8
internal projection: owners
count  0
total  0
rc=0
```

## THE ROWS — membership and ownership differ only in key2

```text
╭────────────┬──────────────┬──────────────┬───────╮
│ key2 type  │  key3 team   │  key4 human  │ value │
╞════════════╪══════════════╪══════════════╪═══════╡
│ membership │ fibggi2juubk │ fi7io4lvjqio │ {}    │
│ membership │ fibggi2juxhc │ fi7io4lvjwu8 │ {}    │
│ membership │ fibggi2jv0n4 │ fi7io4lvjqio │ {}    │
│ ownership  │ fibggi2juubk │ fi7io4lvjqio │ {}    │
│ ownership  │ fibggi2juxhc │ fi7io4lvjqio │ {}    │
╰────────────┴──────────────┴──────────────┴───────╯
  tables: abv_metadata abv_l1_records abv_lab_metadata
```

---

**Membership and ownership differ only in `key2`.** Identical paths, two
different relationships between the same two things — which is what Q-099 exists
to say. An owner may administer the team and has none of its business authority;
a member receives that authority and may not administer the team. Neither implies
the other, and the last section shows a human on each side of that line.