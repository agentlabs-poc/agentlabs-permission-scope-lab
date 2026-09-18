# Authority loading transport — Q-139 / CONTRACT-014

Status: **AGREED.** This chapter did not exist. Nothing in this handbook named
the request an application makes to load authority, or the answer it receives —
the one contract a migration cannot avoid inventing. It is adopted from a
reference implementation that speaks it over HTTP and in-process, with both sides
agreeing on every refusal, and with the traffic captured rather than described.

An application asks **what one human holds in one area**. It never asks whether
to allow. The answer is the complete authority of that human there, which is what
makes it worth caching against them.

## The question

```json
{
  "version": "1",
  "identity": {
    "version": "1",
    "actor": { "type": "service_account", "id": "agent_hrms" },
    "human_id": "fi7io4lvjqio"
  },
  "options": {}
}
```

`actor` is the asking application's own credential; `human_id` is the person it
asks about. They are different parties, and that separation is the point: an
application asks as itself about many humans and is none of them.

Nothing in the question names an endpoint, a method, a resource, or a verdict.

`options` carries two optional narrowings and nothing else:

| Option | Meaning |
|---|---|
| `permissions` | Narrow the answer to these. A narrowing, never an assertion — Q-136. Omitted means everything the human holds, which is the cacheable answer. |
| `omit_source` | Drop the explanation. A consumer that must produce allow evidence cannot use it, because the contributing chain lives in the explanation. |

## The answer

```json
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
      "permissions": ["hrms:payroll:payslip::read", "hrms:payroll:payslip::write"],
      "scope": { "dept": "FIN" },
      "validity": { "not_before": null, "expires_at": null },
      "source": {
        "assignment_id": "fm5b7t4p5iv8",
        "team_id": "fibggi2juubk",
        "via": "membership",
        "lineage": [
          { "grant_id": "fk3x9r2m0dq3", "revision": 1,
            "assignment_id": "fm5b7t4p0dq3", "team_id": "fibggi2jur5s", "root": true },
          { "grant_id": "fk3x9r2m5iv8", "revision": 1,
            "assignment_id": "fm5b7t4p5iv8", "team_id": "fibggi2juubk" }
        ]
      }
    }
  ]
}
```

| Field | Meaning |
|---|---|
| `version` | Contract version, on the envelope and on every grant. The envelope's says how to read the envelope; a grant's says how to read its scope and validity, which are the fields that decide a boundary. |
| `tenant_id`, `application_id`, `human_id` | The three boundaries, echoed. They exist so a consumer can tell "nothing here" from "answered about somebody else". |
| `resolved_grants` | Every grant the human holds in this area, after the chain is walked. Empty is a completed answer, never a failure. |
| `permissions` | What this grant selects, already expanded from any adopted role. |
| `scope` | The boundary, already folded down the chain — a consumer never folds one itself. |
| `validity` | The narrowest window in the chain; absent means no automatic expiry. |
| `source` | Why the human holds it: the binding, the group, and the lineage root-first. `adopted_role` appears when the grant selects through a role revision. |

## Rules a consumer must apply

- **Reject an unsupported version** rather than guessing a default — CONTRACT-010.
- **Corroborate the three boundaries** against the question. An answer about a
  different tenant, application or human is unusable in whole, not in part.
- **Reject unknown fields, duplicate keys and trailing content.** Drift between
  the two sides is how one of them silently starts meaning something else.
- **Bound the answer.** A response larger than the consumer accepts is a failure
  to establish authority, not a denial.
- **Do not follow redirects.** A redirect is not an authority service: following
  one hands the credential to whoever set the header and then believes the reply.
- **Every failure here is an evaluation error**, never a denial — Q-051.

## What is deliberately not adopted

**The route.** The reference implementation serves this at a path containing a
segment named for the lab's own module, which is an artifact of where the work
was done and not a contract. Operation routing stays open, as
[wire contracts](../implementation/plan/abv-123-wire-contracts.md) already say.
What is adopted here is the request and the answer.

**A batch form.** Q-130 requires per-item complete routes and explicitly adds no
batch schema, and no batch transport has been built or demonstrated. HC-07-09
therefore **advances and does not close** — its residue named versioned
transports *and* batch evidence, and only the first is supplied.

**Freshness fields.** `authority_epoch` and `resolved_at` are unbuilt. Nothing
caches, so nothing can be stale; a cache and an epoch are adopted together or
not at all.

## Rationale / conscious tradeoff

The alternative was to leave the wire to each deployment. It was rejected because
this is the boundary the whole architecture rests on: the application holds no
authority records and asks a question it must be able to trust the shape of. Two
teams inventing two answers to that is the failure this handbook exists to
prevent.

The conscious cost is that `source` is rich — assignment, group and lineage per
step. A consumer needs the chain to produce allow evidence under Q-134, and
`omit_source` exists for one that does not. It is explanation, not authorization:
none of it may be used to widen what the grant already permits.
