# Demonstration — two processes

Captured from a real run of `verify-two-processes.sh`, then rendered into this
file **from that capture**. The SVG beside it is generated from the same
capture, so the image and this file cannot drift.

Every demonstration before this one ran inside a single process, and most ran
through the `abv` CLI. This one starts **two programs on two ports** and makes
ordinary HTTP requests into an application's endpoints. Nothing about the
decision is new — the chain walk is the same code demonstration 20 captured.
What is new is *where it runs*: the question crosses a network, and the
application that asks holds no authority records at all.

`127` lines captured, 10 commands. Reproduce with:

```sh
$(sess path)/verify-two-processes.sh
```

## TWO PROCESSES

> (the Auth service holds the records. The application holds none —
> it is given a URL where it used to be given a database path.)

```console
$ auth-service --authority authority.db --registry registry.db --listen 127.0.0.1:8080
auth-service listening on 127.0.0.1:8080 for acme/hrms
$ hrms --auth http://127.0.0.1:8080 --listen 127.0.0.1:8081 --client agent_hrms
hrms listening on 127.0.0.1:8081, asking http://127.0.0.1:8080 as agent_hrms
```

## THE REQUEST REACHES THE HANDLER, OR IT DOES NOT

> (maya holds payslip read and write within dept=FIN)

```console
$ curl -X GET  /api/v1/acme/FIN/C17                             (inside her boundary)
{
    "tenant_id": "acme",
    "department_id": "FIN",
    "certificate_id": "C17",
    "employee_id": "fi7io4lvjqio",
    "owner_id": "fi7io4lvjqio",
    "title": "FIN annual"
}
200
$ curl -X GET  /api/v1/acme/ENG/C18                             (outside it)
{
    "version": "1",
    "decision": "deny",
    "error_code": "NO_AUTHORIZING_GRANT",
    "error_message": "You do not have access to this resource.",
    "error_message_reason": "No grant authorizes this operation within the requested boundary."
}
403
```

## SAME ENDPOINT, SAME POLICY, TWO PEOPLE

> (Team1's grant carries read AND write. Team2's child selects read
> only — so nutan reaches the same endpoint and is refused. Permissions
> are chosen by each child; scope is inherited and narrowed.)

```console
$ curl -X PUT  /api/v1/acme/certificates/C17                    (maya may write)
{
    "tenant_id": "acme",
    "department_id": "FIN",
    "certificate_id": "C17",
    "employee_id": "fi7io4lvjqio",
    "owner_id": "fi7io4lvjqio",
    "title": "revised"
}
200
$ curl -X PUT  /api/v1/acme/certificates/C17                    (nutan may not)
{
    "version": "1",
    "decision": "deny",
    "error_code": "NO_AUTHORIZING_GRANT",
    "error_message": "You do not have access to this resource.",
    "error_message_reason": "No grant authorizes this operation within the requested boundary."
}
403
```

## A THIRD ENDPOINT, AND AN ALL-VALUES ASK

> (the listing endpoint names no department, so it asks for every one.
> Q-071: that is a deny, not a quiet narrowing to the FIN certificates
> maya could have seen.)

```console
$ curl -X GET  /api/v1/acme/certificates                        (maya, every department)
{
    "version": "1",
    "decision": "deny",
    "error_code": "NO_AUTHORIZING_GRANT",
    "error_message": "You do not have access to this resource.",
    "error_message_reason": "No grant authorizes this operation within the requested boundary."
}
403
$ curl -X GET  /api/v1/acme/departments/FIN/certificates        (maya, within FIN)
[
    {
        "tenant_id": "acme",
        "department_id": "FIN",
        "certificate_id": "C17",
        "employee_id": "fi7io4lvjqio",
        "owner_id": "fi7io4lvjqio",
        "title": "FIN annual"
    },
    {
        "tenant_id": "acme",
        "department_id": "FIN",
        "certificate_id": "C19",
        "employee_id": "fi7io4lvjwu8",
        "owner_id": "fi7io4lvjqio",
        "title": "FIN supplemental"
    }
]
200
```

## WHAT CROSSED THE BOUNDARY

> (one question per request, and it is always the same question)

```console
auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve
auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve
auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve
auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve
auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve
auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve
```

## WHAT THE APPLICATION LINKS

> (the claim the split exists to make, checked rather than trusted)

```console
$ go list -deps ./cmd/hrms | grep agentlabs
agentlabs.local/authmiddleware
agentlabs.local/apps/hrms
agentlabs.local/authclient
agentlabs.local/apps/hrms/cmd/hrms
```

## AND WHEN AUTH IS DOWN

> (an outage is an evaluation failure, never a denial — Q-128. A gate
> that answered 403 here would fail closed and lie about why.)

```console
$ curl -X GET  /api/v1/acme/FIN/C17                             (auth service stopped)
{
    "version": "1",
    "error_code": "AUTH_UNREACHABLE",
    "error_message": "We could not check your access.",
    "error_message_reason": "the authority service did not answer"
}
503
```
## What the run establishes

| Claim | The evidence above |
|---|---|
| The gate runs on the client | `hrms` answers 200 and 403 itself; Auth only ever answers the same `authority.resolve` question |
| The application links no authority domain | `go list -deps` names `authmiddleware`, `authclient` and `apps/hrms` — not `abv`, not `registry` |
| The wire carries authority, not decisions | six requests, six identical questions; no request names an endpoint, a method or a resource |
| Permissions are selected, scope is inherited | same endpoint, same policy, same certificate: maya writes, nutan is refused |
| An all-values ask is a deny | the listing endpoint names no department, and Q-071 refuses rather than quietly narrowing to what maya could have seen |
| An outage is not a denial | Auth stopped gives `503` `AUTH_UNREACHABLE`, never `403` — Q-128 |

The application's policy table declares **four** endpoints across **two**
permissions (`hrms:payroll:payslip::read` and `…::write`), and all four appear
above.

## What it does not establish

**Where a policy comes from.** The four policies above are built by a Go helper
at startup, in the application's own code. That is the honest state of the lab:
a policy is a compile-time value beside the handler it guards. Whether it should
instead be a record — registered with the application, versioned, and read at
boot — is the question this demonstration forces and does not answer.

**Who the caller is.** `hrms` sends `Authorization: Bearer agent_hrms`, and the
lab's credential gate compares that string. A deployment issues credentials and
derives the calling application from one; this does not.

**Freshness.** Nothing here carries an epoch, and nothing caches. Each request
asks again from scratch, which is correct and says nothing about what a cache
would have to invalidate.
