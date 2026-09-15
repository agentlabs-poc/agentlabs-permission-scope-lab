# Demonstration — a decision

Captured from a real run of `verify-decision.sh`, then rendered into this file **from
that capture**. The SVG beside it is generated from the same capture, so the
image and this file cannot drift.

It is the first of thirty demonstrations to capture a **decision**. Every other
one shows a record being written, read or refused; this one shows the thing the
records exist to produce.

`114` lines captured, 10 commands in 10 blocks. Reproduce with:

```sh
$(sess path)/verify-decision.sh
```

## THE CHAIN AUTHORITY TRAVELS

> (the root bounds the area; Team1's grant narrows it to Finance; Team2's
> narrows again to one certificate. maya is in Team1, nutan in Team2.)

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

## ALLOWED — and on whose authority

> (grant_ids is the chain, root first. This is the whole point of every
> record: a human asked, and the answer came back with its evidence.)

```console
auth-evaluate --human fi7io4lvjqio --permission hrms:payroll:payslip::read --boundary dept=FIN
{
    "version": "1",
    "decision": "allow",
    "grant_ids": [
        "fk3x9r2m0dq3",
        "fk3x9r2m5iv8"
    ]
}
rc=0
```

## DENIED OUTSIDE THE BOUNDARY

> (same human, same permission, one value changed. The deny carries two
> messages: one for the person, one for the operator — Q-053.)

```console
auth-evaluate --human fi7io4lvjqio --permission hrms:payroll:payslip::read --boundary dept=ENG
{
    "version": "1",
    "decision": "deny",
    "error_code": "NO_AUTHORIZING_GRANT",
    "error_message": "You do not have access to this resource.",
    "error_message_reason": "No grant authorizes this operation within the requested boundary."
}
rc=3
```

## A DEEPER ROUTE NEEDS EVERY PREDICATE

> (nutan reaches the same permission through one more hop, so her route
> carries both narrowings. Finance alone is not enough; the chain is
> three grants rather than two, and the ids say so.)

```console
auth-evaluate --human fi7io4lvjwu8 --permission hrms:payroll:payslip::read --boundary dept=FIN --boundary cert=C17
{
    "version": "1",
    "decision": "allow",
    "grant_ids": [
        "fk3x9r2m0dq3",
        "fk3x9r2m5iv8",
        "fk3x9r2man0d"
    ]
}
rc=0
```

```console
auth-evaluate --human fi7io4lvjwu8 --permission hrms:payroll:payslip::read --boundary dept=FIN --boundary cert=C18
{
    "version": "1",
    "decision": "deny",
    "error_code": "NO_AUTHORIZING_GRANT",
    "error_message": "You do not have access to this resource.",
    "error_message_reason": "No grant authorizes this operation within the requested boundary."
}
rc=3
```

## SELECTED, NOT INHERITED — the asymmetry, at a decision

> (Team1 holds read AND write. Team2's grant selects read only, so nutan
> cannot write where maya can — permissions are chosen by each child,
> while scope is inherited and narrowed.)

```console
auth-evaluate --human fi7io4lvjqio --permission hrms:payroll:payslip::write --boundary dept=FIN
{
    "version": "1",
    "decision": "allow",
    "grant_ids": [
        "fk3x9r2m0dq3",
        "fk3x9r2m5iv8"
    ]
}
rc=0
```

```console
auth-evaluate --human fi7io4lvjwu8 --permission hrms:payroll:payslip::write --boundary dept=FIN --boundary cert=C17
{
    "version": "1",
    "decision": "deny",
    "error_code": "NO_AUTHORIZING_GRANT",
    "error_message": "You do not have access to this resource.",
    "error_message_reason": "No grant authorizes this operation within the requested boundary."
}
rc=3
```

## ALL-VALUES DENIES RATHER THAN DERIVING A SUBSET — Q-071

> (asking for every department is not a request for the ones she may see.)

```console
auth-evaluate --human fi7io4lvjqio --permission hrms:payroll:payslip::read --all dept
{
    "version": "1",
    "decision": "deny",
    "error_code": "NO_AUTHORIZING_GRANT",
    "error_message": "You do not have access to this resource.",
    "error_message_reason": "No grant authorizes this operation within the requested boundary."
}
rc=3
```

## $self IS WRITABLE AND UNREACHABLE

> (the evaluator binds $self to the authorizing human, and the codec
> accepts it as a scope value — but resolution refuses any content
> carrying it, at the proposal and at every step of the walk. So a grant
> can be created and can never be assigned. Captured, not asserted.)

```console
abv grants create --parent fk3x9r2m5iv8 --permissions hrms:payroll:payslip::read --scope user=$self
internal projection: grant
id  fyc2tdaf2y2o
status  enabled
kind  child
revision  1
parent  fk3x9r2m5iv8
permissions  hrms:payroll:payslip::read
scope  user=$self
rc=0
```

```console
abv assign --file self.json
unsupported operation
rc=5
```
