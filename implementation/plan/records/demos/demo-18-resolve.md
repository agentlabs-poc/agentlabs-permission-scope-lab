# Demonstration — resolve

Captured from a real run of `verify-resolve.sh`, then rendered into this file **from
that capture**. The SVG beside it is generated from the same capture, so the
image and this file cannot drift.

`343` lines captured, 13 commands. Reproduce with:

```sh
$(sess path)/verify-resolve.sh
```

## ONE GRANT — and the chain that carries it

> (maya is a member of Team1, which holds the FIN grant. The lineage runs
> root-first: the trusted root, then Team1's narrowing to dept=FIN.)

```console
abv resolve --human fi7io4lvjqio
{
  "version": "1",
  "tenant_id": "acme",
  "application_id": "hrms",
  "human_id": "fi7io4lvjqio",
  "resolved_grants": [
    {
      "version": "1",
      "grant_id": "fk3x9r2m5iv8",
      "revision": 1,
      "parent_grant_id": "fk3x9r2m0dq3",
      "permissions": [
        "hrms:payroll:payslip::read",
        "hrms:payroll:payslip::write"
      ],
      "scope": {
        "dept": "FIN"
      },
      "source": {
        "assignment_id": "fm5b7t4p5iv8",
        "team_id": "fibggi2juubk",
        "via": "membership",
        "lineage": [
          {
            "grant_id": "fk3x9r2m0dq3",
            "revision": 1,
            "assignment_id": "fm5b7t4p0dq3",
            "team_id": "fibggi2jur5s",
            "root": true
          },
          {
            "grant_id": "fk3x9r2m5iv8",
            "revision": 1,
            "assignment_id": "fm5b7t4p5iv8",
            "team_id": "fibggi2juubk"
          }
        ]
      }
    }
  ]
}
rc=0
```

## A SECOND ROUTE, ONE STEP DEEPER

> (Team2 sits under Team1 and narrows further to cert=C17. Adding maya to
> Team2 and assigning its grant gives her a second, three-step route.)

```console
abv team add-member --id fibggi2juxhc --human fi7io4lvjqio
internal projection: membership
added  team=fibggi2juxhc  human=fi7io4lvjqio
```

```console
abv assign --file a2.json  (Team2 grant → Team2)
assignment fm5b7t4pan0d created
```

## TWO GRANTS, TWO LINEAGES, ONE DOCUMENT

> (the second grant's scope is dept=FIN AND cert=C17 — folded down the
> chain, so the client never folds one itself)

```console
abv resolve --human fi7io4lvjqio
{
  "version": "1",
  "tenant_id": "acme",
  "application_id": "hrms",
  "human_id": "fi7io4lvjqio",
  "resolved_grants": [
    {
      "version": "1",
      "grant_id": "fk3x9r2m5iv8",
      "revision": 1,
      "parent_grant_id": "fk3x9r2m0dq3",
      "permissions": [
        "hrms:payroll:payslip::read",
        "hrms:payroll:payslip::write"
      ],
      "scope": {
        "dept": "FIN"
      },
      "source": {
        "assignment_id": "fm5b7t4p5iv8",
        "team_id": "fibggi2juubk",
        "via": "membership",
        "lineage": [
          {
            "grant_id": "fk3x9r2m0dq3",
            "revision": 1,
            "assignment_id": "fm5b7t4p0dq3",
            "team_id": "fibggi2jur5s",
            "root": true
          },
          {
            "grant_id": "fk3x9r2m5iv8",
            "revision": 1,
            "assignment_id": "fm5b7t4p5iv8",
            "team_id": "fibggi2juubk"
          }
        ]
      }
    },
    {
      "version": "1",
      "grant_id": "fk3x9r2man0d",
      "revision": 1,
      "parent_grant_id": "fk3x9r2m5iv8",
      "permissions": [
        "hrms:payroll:payslip::read"
      ],
      "scope": {
        "cert": "C17",
        "dept": "FIN"
      },
      "source": {
        "assignment_id": "fm5b7t4pan0d",
        "team_id": "fibggi2juxhc",
        "via": "membership",
        "lineage": [
          {
            "grant_id": "fk3x9r2m0dq3",
            "revision": 1,
            "assignment_id": "fm5b7t4p0dq3",
            "team_id": "fibggi2jur5s",
            "root": true
          },
          {
            "grant_id": "fk3x9r2m5iv8",
            "revision": 1,
            "assignment_id": "fm5b7t4p5iv8",
            "team_id": "fibggi2juubk"
          },
          {
            "grant_id": "fk3x9r2man0d",
            "revision": 1,
            "assignment_id": "fm5b7t4pan0d",
            "team_id": "fibggi2juxhc"
          }
        ]
      }
    }
  ]
}
rc=0
```

## THE SAME CALL, NARROWED TO ONE PERMISSION

> (the gate passes no filter at all — it asks what she holds, which is the
> shape that could be cached against her. Nothing caches yet. A narrowing
> is for a caller that wants less than everything.)

```console
abv resolve --human fi7io4lvjqio --permissions hrms:payroll:payslip::write
{
  "version": "1",
  "tenant_id": "acme",
  "application_id": "hrms",
  "human_id": "fi7io4lvjqio",
  "resolved_grants": [
    {
      "version": "1",
      "grant_id": "fk3x9r2m5iv8",
      "revision": 1,
      "parent_grant_id": "fk3x9r2m0dq3",
      "permissions": [
        "hrms:payroll:payslip::read",
        "hrms:payroll:payslip::write"
      ],
      "scope": {
        "dept": "FIN"
      },
      "source": {
        "assignment_id": "fm5b7t4p5iv8",
        "team_id": "fibggi2juubk",
        "via": "membership",
        "lineage": [
          {
            "grant_id": "fk3x9r2m0dq3",
            "revision": 1,
            "assignment_id": "fm5b7t4p0dq3",
            "team_id": "fibggi2jur5s",
            "root": true
          },
          {
            "grant_id": "fk3x9r2m5iv8",
            "revision": 1,
            "assignment_id": "fm5b7t4p5iv8",
            "team_id": "fibggi2juubk"
          }
        ]
      }
    }
  ]
}
rc=0
```

## WITHOUT THE EXPLANATION — what a bearer token would carry

```console
abv resolve --human fi7io4lvjqio --no-source
{
  "version": "1",
  "tenant_id": "acme",
  "application_id": "hrms",
  "human_id": "fi7io4lvjqio",
  "resolved_grants": [
    {
      "version": "1",
      "grant_id": "fk3x9r2m5iv8",
      "revision": 1,
      "parent_grant_id": "fk3x9r2m0dq3",
      "permissions": [
        "hrms:payroll:payslip::read",
        "hrms:payroll:payslip::write"
      ],
      "scope": {
        "dept": "FIN"
      }
    },
    {
      "version": "1",
      "grant_id": "fk3x9r2man0d",
      "revision": 1,
      "parent_grant_id": "fk3x9r2m5iv8",
      "permissions": [
        "hrms:payroll:payslip::read"
      ],
      "scope": {
        "cert": "C17",
        "dept": "FIN"
      }
    }
  ]
}
rc=0
```

## A FILTER IS A NARROWING, NOT AN ASSERTION

> (three permissions that resolve to nothing: one never registered, one
> registered and then retired, one registered and active that she simply
> does not hold. All three are the same answer — empty, rc=0 — and the
> gate denies on each. Asking whether a permission exists is not a
> question this service answers.)

```console
abv catalog set-permission-status hrms:payroll:payslip::write --active false
internal projection: permission
id  hrms:payroll:payslip::write
active  false
```

**never registered**

```console
abv resolve --human fi7io4lvjqio --permissions hrms:payroll:payslip::export
{
  "version": "1",
  "tenant_id": "acme",
  "application_id": "hrms",
  "human_id": "fi7io4lvjqio",
  "resolved_grants": []
}
rc=0
```

**registered, then retired**

```console
abv resolve --human fi7io4lvjqio --permissions hrms:payroll:payslip::write
{
  "version": "1",
  "tenant_id": "acme",
  "application_id": "hrms",
  "human_id": "fi7io4lvjqio",
  "resolved_grants": []
}
rc=0
```

**registered and active, simply not held**

```console
abv resolve --human fi7io4lvjqio --permissions hrms:payroll:payslip::delete
{
  "version": "1",
  "tenant_id": "acme",
  "application_id": "hrms",
  "human_id": "fi7io4lvjqio",
  "resolved_grants": []
}
rc=0
```

## WHAT RETIREMENT ACTUALLY DID

> (Q-143: the retirement withdraws the retired permission and nothing
> else. The stored grant is untouched — Q-125 rewrites nothing — and it
> still supplies ::read, so her route narrows rather than closing. The
> deeper grant never selected ::write and is untouched. A route closes
> only when nothing it selects survives.)

```console
abv resolve --human fi7io4lvjqio --no-source
{
  "version": "1",
  "tenant_id": "acme",
  "application_id": "hrms",
  "human_id": "fi7io4lvjqio",
  "resolved_grants": [
    {
      "version": "1",
      "grant_id": "fk3x9r2m5iv8",
      "revision": 1,
      "parent_grant_id": "fk3x9r2m0dq3",
      "permissions": [
        "hrms:payroll:payslip::read"
      ],
      "scope": {
        "dept": "FIN"
      }
    },
    {
      "version": "1",
      "grant_id": "fk3x9r2man0d",
      "revision": 1,
      "parent_grant_id": "fk3x9r2m5iv8",
      "permissions": [
        "hrms:payroll:payslip::read"
      ],
      "scope": {
        "cert": "C17",
        "dept": "FIN"
      }
    }
  ]
}
rc=0
```

```console
abv inspect grant fk3x9r2m5iv8 | jq .
{
  "version": "1",
  "grant_id": "fk3x9r2m5iv8",
  "revision": 1,
  "parent_grant_id": "fk3x9r2m0dq3",
  "permissions": [
    "hrms:payroll:payslip::read",
    "hrms:payroll:payslip::write"
  ],
  "scope": {
    "dept": "FIN"
  }
}
```

## AND A HUMAN WITH NOTHING

> (a subject this tenant has no record of is a refusal, not an empty
> answer: there is no set to narrow.)

```console
abv resolve --human fn2q6v8sbo1e
operation rejected or record not found
rc=3
```
