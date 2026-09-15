# Demonstration — the service credential

Captured from a real run of `verify-service-actor.sh`, then rendered into this file **from
that capture**. The SVG beside it is generated from the same capture, so the
image and this file cannot drift.

`114` lines captured, 6 commands in 7 blocks. Reproduce with:

```sh
$(sess path)/verify-service-actor.sh
```

## SETUP — nutan is given a route, so there is something to ask about

> (the seeded fixture leaves Team2's assignment proposed and absent. Without
> this, asking about nutan returns an empty document — which would look the
> same as asking about somebody who does not exist.)

```console
abv assign --file testdata/a2.json
assignment fm5b7t4pan0d created
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

## REFUSALS — the credential is bound, and a human is not a credential

> (first: a credential this deployment never issued. Second: nutan asking as
> herself, refused because the fixture gate admits one human — a real
> deployment would admit many, and then the actor rules below decide.)

```console
abv resolve --human fi7io4lvjqio --client agent_crm
operation rejected or record not found
rc=3
```

```console
abv resolve --human fi7io4lvjwu8
operation rejected or record not found
rc=3
   (the rule the CLI cannot reach, and the one that matters if a gate admits
several humans: a user actor naming somebody else is impersonation rather
than delegation, and validateReadingIdentity refuses it before any gate
runs. TestValidateReadingIdentity covers it; --human always names both
the actor and the subject, so no command line can construct it.)
```

## NO NEW RECORD TYPE — this moved a gate, not a record

```console
   (the one extra assignment is the setup above. Nothing else in this change
    writes anything: who may ask is not a record.)
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
