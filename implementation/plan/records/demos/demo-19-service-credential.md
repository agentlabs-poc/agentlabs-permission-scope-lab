# Demonstration — the service credential

Captured from a real run of `verify-service-actor.sh`, then rendered into this file **from
that capture**. The SVG beside it is generated from the same capture, so the
image and this file cannot drift.

`98` lines captured, 5 commands. Reproduce with:

```sh
$(sess path)/verify-service-actor.sh
```

## THE OLD SHAPE — a human asking about themselves

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
    }
  ]
}
rc=0
```

## THE NEW ONE — an application asking about somebody else

> (agent_hrms is the lab's stand-in for a workload credential, bound to
> one tenant application. The binding is what the gate compares; nothing
> about the subject is checked, because the application is none of them.)

```console
abv resolve --human fi7io4lvjwu8 --client agent_hrms --no-source
{
  "version": "1",
  "tenant_id": "acme",
  "application_id": "hrms",
  "human_id": "fi7io4lvjwu8",
  "resolved_grants": [
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

## WHO ASKS DOES NOT CHANGE WHAT IS HELD

```console
abv resolve --human fi7io4lvjqio --client agent_hrms --no-source
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
    }
  ]
}
rc=0
```

## REFUSALS

> (a credential this area does not know; and a human actor naming
> somebody else, which is impersonation rather than delegation)

```console
abv resolve --human fi7io4lvjqio --client agent_crm
operation rejected or record not found
rc=3
```

```console
abv resolve --human fi7io4lvjwu8   (as a user actor — refused by the fixture gate)
operation rejected or record not found
rc=3
```

## THE ROWS ARE UNCHANGED — this moved a gate, not a record

```console
╭────────────────┬──────╮
│      type      │ rows │
╞════════════════╪══════╡
│ assignment     │    3 │
│ catalog        │    1 │
│ grant          │    3 │
│ grant_revision │    3 │
│ membership     │    3 │
│ permission     │    3 │
│ role           │    1 │
│ scope          │    3 │
│ team           │    4 │
╰────────────────┴──────╯
```
