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

`161` lines captured, 13 commands. Reproduce with:

```sh
$(sess path)/verify-two-processes.sh
```

## SETUP — BOTH WOMEN ARE GIVEN A ROUTE

> (the seeded fixture leaves Team2's assignment proposed, so until this
> runs nutan holds nothing at all. It is captured rather than assumed,
> because otherwise her refusal below could not be told apart from
> having no grant — and the requests are what say what her route selects.)

```console
abv scenario seed team-fin-c17
LAB ONLY: fixed fixture identity; not authenticated administration
scenario seeded
```

```console
abv assign --file a2.json   (Team2, so nutan has a route at all)
LAB ONLY: fixture-context is not authenticated identity; both authority gates are rechecked
assignment fm5b7t4pan0d created
```

## TWO PROCESSES

> (the Auth service holds the records. The application holds none —
> it is given a URL where it used to be given a database path.)

```console
auth-service --authority authority.db --registry registry.db --listen 127.0.0.1:8080
auth-service listening on 127.0.0.1:8080 for acme/hrms
```

```console
HRMS_AUTH_TOKEN=... hrms --auth http://127.0.0.1:8080 --listen 127.0.0.1:8081 --client agent_hrms --allow-cleartext
hrms listening on 127.0.0.1:8081, asking http://127.0.0.1:8080 as agent_hrms
```

## THE REQUEST REACHES THE HANDLER, OR IT DOES NOT

> (maya holds payslip read and write within dept=FIN)

```console
curl -X GET  /api/v1/acme/FIN/C17                             (inside her boundary)
{
    "tenant_id": "acme",
    "department_id": "FIN",
    "certificate_id": "C17",
    "employee_id": "fi7io4lvjqio",
    "owner_id": "fi7io4lvjqio",
    "title": "FIN annual"
}
200
```

```console
curl -X GET  /api/v1/acme/ENG/C18                             (outside it)
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
curl -X GET  /api/v1/acme/FIN/C17                             (nutan may read it)
{
    "tenant_id": "acme",
    "department_id": "FIN",
    "certificate_id": "C17",
    "employee_id": "fi7io4lvjqio",
    "owner_id": "fi7io4lvjqio",
    "title": "FIN annual"
}
200
```

```console
curl -X PUT  /api/v1/acme/certificates/C17                    (maya may write)
{
    "tenant_id": "acme",
    "department_id": "FIN",
    "certificate_id": "C17",
    "employee_id": "fi7io4lvjqio",
    "owner_id": "fi7io4lvjqio",
    "title": "revised"
}
200
```

```console
curl -X PUT  /api/v1/acme/certificates/C17                    (nutan may not)
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
curl -X GET  /api/v1/acme/certificates                        (maya, every department)
{
    "version": "1",
    "decision": "deny",
    "error_code": "NO_AUTHORIZING_GRANT",
    "error_message": "You do not have access to this resource.",
    "error_message_reason": "No grant authorizes this operation within the requested boundary."
}
403
```

```console
curl -X GET  /api/v1/acme/departments/FIN/certificates        (maya, within FIN)
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

> (one question per request, and the body is the claim worth checking:
> the application asks what a human holds — never whether to allow, and
> no longer even which permission it is about)

**auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve**

**{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":"fi7io4lvjqio"},"options":{}}**

**auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve**

**{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":"fi7io4lvjqio"},"options":{}}**

**auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve**

**{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":"fi7io4lvjwu8"},"options":{}}**

**auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve**

**{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":"fi7io4lvjqio"},"options":{}}**

**auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve**

**{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":"fi7io4lvjwu8"},"options":{}}**

**auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve**

**{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":"fi7io4lvjqio"},"options":{}}**

**auth  <- POST /api/v1/acme/abv/applications/hrms/authority.resolve**

**{"version":"1","identity":{"version":"1","actor":{"type":"service_account","id":"agent_hrms"},"human_id":"fi7io4lvjqio"},"options":{}}**

## WHAT THE APPLICATION LINKS

> (the claim the split exists to make, checked rather than trusted)

```console
go list -deps ./cmd/hrms | grep agentlabs
agentlabs.local/authmiddleware
agentlabs.local/apps/hrms
agentlabs.local/authclient
agentlabs.local/apps/hrms/cmd/hrms
```

## AND WHEN AUTH IS DOWN

> (an outage is an evaluation failure, never a denial — Q-051. A gate
> that answered 403 here would fail closed and lie about why.)

```console
curl -X GET  /api/v1/acme/FIN/C17                             (auth service stopped)
{
    "version": "1",
    "error_code": "AUTHORITY_UNREACHABLE",
    "error_message": "We could not check your access.",
    "error_message_reason": "the authority service did not answer"
}
503
```

## What the run establishes

| Claim | The evidence above |
|---|---|
| The gate runs on the client | `hrms` answers 200 and 403 itself; Auth is never asked whether to allow |
| The application links no authority domain | `go list -deps` names `authmiddleware`, `authclient` and `apps/hrms` — not `abv`, not `registry` |
| The wire carries authority, not decisions | the seven bodies are printed in full: each names a human, `"options":{}`, and **no** permission, endpoint, method, resource or verdict |
| Permissions are selected, scope is inherited | nutan reads the certificate and cannot write it; maya writes the same one through the same endpoint and the same policy |
| An all-values ask is a deny | the listing endpoint names no department, and Q-071 refuses rather than quietly narrowing to what maya could have seen |
| An outage is not a denial | Auth stopped gives `503` `AUTHORITY_UNREACHABLE`, never `403` — Q-051 / DECISION-003 |

The application's policy table declares **four** endpoints across **two**
permissions (`hrms:payroll:payslip::read` and `…::write`), and all four appear
above.

The seven bodies differ in exactly **one** field — `human_id`. They are otherwise
the same question, and it is a question about a person, not about a request.
That is the architectural claim of the whole lab, and here it is the capture
rather than the prose that makes it.

They used to differ in two: the question named the permission the endpoint was
guarding, which narrowed the reply to that one request and made it worthless to
anyone else. `options` is now empty, and the answer describes the person — the
shape that can be cached against them, though nothing caches yet.

## Where the policy comes from

Each endpoint declares its own policy — method, path, the permission it requires,
and where the material comes from — and hands it to the middleware as a
parameter:

```go
authmiddleware.Wrap(policy, identity, evaluator, bind, renderFailure)
```

Ownership is the endpoint's, which is the property worth having: the permission
that guards a route is edited in the same place as the route and the handler, and
cannot drift from them. Nothing administers it remotely, so no record change can
alter who gets through — only authority can.

## What it does not establish

**That the policy's route parameters are bound.** The gate binds the path's
tenant to the trusted area only when the route parameter is spelled exactly
`tenant`. `hrms` spells it that way; a policy that says `tenantId` loses tenant
isolation silently, and the gate cannot tell. That is a contract question — a
policy would have to declare which of its inputs carries the tenant — and it is
recorded rather than answered here.

**Who the caller is.** `hrms` sends a bearer token it is given out of band, and
the lab's credential gate compares it — in constant time, because that is the
line a deployment replaces. Nothing here issues a credential, and the id the
application is known by is deliberately not the secret it authenticates with.

**Freshness.** Nothing here carries an epoch, and nothing caches. Each request
asks again from scratch, which is correct and says nothing about what a cache
would have to invalidate.

**That the policy is right.** The gate enforces the permission the policy names,
and nothing checks that it exists in the application's catalog — deliberately.
A filter is a narrowing, not an assertion: the client never asks whether a
permission exists, so a typo denies every request to that endpoint, permanently
and without saying why. The endpoint owns its policy and is tested beside the
handler it guards, which is where a typo is caught.

## Held by tests, not only by this capture

A capture proves something ran once. These hold it:

| Test | What fails without it |
|---|---|
| `wiring.TestEveryEndpointForBothHumans` | the matrix above, as assertions — every endpoint for both women |
| `wiring.TestAnAuthThatMisbehavesCannotDecideAnything` | an Auth that answers about another human, another area, another version, or not in JSON at all can decide something |
| `wiring.TestAGrantForAnotherPermissionDeniesRatherThanFailing` | a grant for a permission this request is not about is reported as a broken Auth rather than denied |
| `wiring.TestARedirectNeverReachesTheAttacker` | a `Location` header takes the credential and authors the answer |
| `wiring.TestThePathCannotChooseTheTenant` | the path moves the area the question is about |
| `wiring.TestOneRequestAsksExactlyOneQuestion` | a silent cache, or a doubled question |
| `wiring.TestWithdrawingAnAssignmentChangesTheNextAnswer` | authority is a snapshot the application took at startup rather than a live question |
| `wiring.TestSelfResolvesPerHumanOverTheWire` | `$self` stops meaning the person asking — SELF-001 and GROUP-004, over the wire |
| `wiring.TestAHumanWithNoAuthorityIsDeniedRatherThanFailed` | holding nothing is reported as an outage rather than a denial |
| `wiring.TestTheStackIsCorrectUnderConcurrency` | the shared evaluator and store race |
| `apps/hrms.TestTheApplicationLinksNoAuthorityDomain` | `abv` re-enters the application's dependency closure — two lines in a `go.mod` were enough, and every other test stayed green |
| `wiring.TestOnlyAnAnsweredQuestionIsObserved` | an unauthenticated caller writes lines into the record this demonstration reads, or makes the service buffer its body first |
| `auth-service.TestAValidBodyCannotAddLinesToTheRecord` | a body chooses the shape of that record |
| `auth-service.TestAPathCannotAddLinesToTheRecord` | a path does |
| `wiring.TestTheApplicationAsksAsItselfAboutAHuman` | the application asks as the human instead of as itself — the architecture's central claim, and nothing held it |
| `wiring.TestAnApplicationAuthDoesNotRecogniseDecidesNothing` | the application's own credential problem is rendered as the person's denial |
| `wiring.TestAValidityWindowIsHonouredAcrossTheWire` | `validity` is dropped on the wire and grants never expire |
| `wiring.TestTheWriteIsBoundToTheDepartmentTheBodyNames` | the only body-sourced material in the system is replaced by a constant |
| `wiring.TestAMethodThePolicyDoesNotNameIsRefused` | `HEAD`, which Go routes to the `GET` handler, is authorized |
| `wiring.TestTheFixtureAnswerOpensTheGate` | the fixture twenty refusals rest on could not have opened the gate anyway |
