# Canonical format — resolved authority

**Partly implemented — read §0 before trusting a field.** The shape Auth
delivers when a client asks what a human is entitled to.

Companions: [terms.md](terms.md) for where this sits, and the handbook's
[system overview](../../../docs/system-overview.md) for the request flow it
serves.

---

## 0 · What exists, and what is written here ahead of itself

This document describes the contract. Only part of it is built, and the parts
differ in kind rather than in polish.

| | state |
|---|---|
| `domain.ResolvedAuthority` and the grant shape | **built** — `Facade.ResolveAuthority`, demonstrated in [demo 18](../records/demos/demo-18-resolve.md) |
| effective scope and validity, expanded permissions | **built**, including the contradiction rule in §4 |
| `source` and its lineage, and `OmitSource` | **built** |
| the `permissions` filter | **built** |
| `authority_epoch` and `resolved_at` | **not emitted.** The envelope below shows them because the freshness design needs a place to land; nothing computes an epoch yet |
| `POST authority.resolve`, `GET authority.epoch` | **not built.** There is no HTTP surface at all — the call today is the Go method and the `abv resolve` verb, whose `--client` flag supplies the credential in place of a token |
| a caller who is not the subject | **built** — `--client`, and §2's identity rules |
| `expand_roles: false` and `permissions_ref` | **not built**, deliberately — §6 |

**The caller no longer has to be the subject.** That gap is closed: an
application resolves other people through its own credential, modelled on the
Auth service's workload client — an id and a secret exchanged for a token their
contract describes as *"bound to `auth.registry.read` and one tenant
application"*. The binding is the area, so the gate is a comparison rather than a
policy, and nothing about the subject is checked because an application asks
about many humans and is none of them. See [demo 19](../records/demos/demo-19-service-credential.md).

The lab models the shape and not the issuance. This repository never becomes that
service, so a real credential — the secret, its rotation, the token exchange —
belongs to the migration. What must survive the migration is the rule: **who may
ask is the gate's question, and who is asked about is the walk's.**

---

## 1 · The rule the shape follows

**It is the canonical grant block, with the role resolved away and one `source`
object added.** No invented vocabulary: `scope` is `scope`, `permissions` is
`permissions`. A reader who knows the grant record knows this.

Two deliberate differences from a stored grant, and the name says so —
`resolved_grants`, not `grants`:

| | stored | resolved |
|---|---|---|
| `permissions` | may be a `role_id` + `role_revision` pair | always the expanded list |
| `scope` | this grant's **local** narrowing | the **effective** boundary, ANDed down the chain |
| `validity` | this grant's own window | the **narrowest** window across the chain |

The client must never fold a chain itself. That is what keeps every lineage rule
we have settled — the root's namespace slice, selected-versus-inherited
permissions, inherited-and-ANDed scope — on Auth's side of the wire.

---

## 2 · The call

Three boundaries are named explicitly, every time: **tenant, application, human.**
Nothing is inferred and nothing has a default. The response echoes all three, so
a client can confirm it was answered about what it asked about.

### As a function — the L1 form

```go
// ResolveAuthority answers what one human is entitled to inside one area.
// Area carries the tenant and the application; Identity carries the actor and
// the human whose authority is being resolved.
func (f *Facade) ResolveAuthority(
    ctx context.Context,
    area domain.Area,          // tenant + application
    identity domain.Identity,  // actor{type,id} + human_id — Q-086
    opts domain.ResolveOptions,
) (domain.ResolvedAuthority, error)
```

This is the form Auth's own gate uses in-process, and the form the HTTP handler
wraps. It is the only new read on the Facade; every other exported method is
administrative.

### As an endpoint — proposed, not built

There is no HTTP surface yet. The shape below is what the Go form becomes when
there is one.

```http
POST /api/v1/{tenant}/abv/applications/{application}/authority.resolve
```

```json
{
  "version": "1",
  "identity": {
    "version": "1",
    "actor": {"type": "user", "id": "fi7io4lvjqio"},
    "human_id": "fi7io4lvjqio"
  },
  "options": {
    "include_source": true,
    "expand_roles": true,
    "permissions": null
  }
}
```

| | where it comes from | why there |
|---|---|---|
| **tenant** | the path | the pinned base is `/api/v1/{tenant}/abv/` |
| **application** | the path | permissions are application-scoped, so the answer is |
| **human** | the body, inside the identity block | it is identity, and identity is a body-level contract — Q-086 |

**Path selects context; it does not prove it.** The wire-contracts draft already
pins this: *"Path values select requested context; they do not prove identity,
installation, or authority."* The caller must be established as entitled to ask,
and the tenant's installation of the application is checked before any record is
read.

**The boundaries are not repeated in the body.** Two copies invite a mismatch and
a rule for resolving it. One copy in the path, echoed once in the response, has
neither.

**Why the whole identity block and not just `human_id`.** A request may be an
agent acting for a human. AUTHORITY-002 bounds the answer by *"both the human's
applicable authority and delegation limits"*, and the delegation is Auth's fact.
So Auth resolves **for this actor acting for this human** and returns something
already bounded — the client never has to narrow it further.

### Options, all defaulting to the complete answer

| option | default | effect |
|---|---|---|
| `include_source` | `true` | omit to drop the explanation and keep the hot path lean |
| `expand_roles` | `true` | see §6 — `false` is a designed-in future option, not built |
| `permissions` | `null` | a filter, not a requirement. `null` means everything the human holds |

On the CLI the same three boundaries appear as `--tenant`, `--app` and `--human`,
and `--client` names the credential asking. Without it the caller is the subject
— which is why no command line can express impersonation: `--human` supplies the
actor and the subject at once.

`permissions` exists because a gate deciding one request needs one permission,
and a menu needs all of them. The same call serves both; the default is the
complete answer because that is the cacheable one.

### The freshness companion — proposed, not built

```http
GET /api/v1/{tenant}/abv/authority.epoch
```

Tenant-scoped, not application-scoped — teams and memberships are keyed by tenant
with no application, so a membership removal reduces authority in every
application the tenant holds. Cheap to call; the expensive call is the resolve.
A cached document may be used only while its `authority_epoch` still matches.

---

## 3 · The response document


```json
{
  "version": "1",
  "tenant_id": "acme",
  "application_id": "hrms",
  "human_id": "fi7io4lvjqio",
  "authority_epoch": "fyb94x7tldds",
  "resolved_at": "2026-09-15T08:48:02Z",

  "resolved_grants": [
    {
      "version": "1",
      "grant_id": "fk3x9r2m5iv8",
      "revision": 1,
      "parent_grant_id": "fk3x9r2m0dq3",
      "permissions": ["hrms:payroll:payslip::read", "hrms:payroll:payslip::write"],
      "scope": {"dept": "FIN"},
      "validity": {"not_before": null, "expires_at": null},
      "source": {
        "assignment_id": "fm5b7t4p5iv8",
        "team_id": "fibggi2juubk",
        "via": "membership",
        "lineage": [
          {"grant_id":"fk3x9r2m0dq3","revision":1,"assignment_id":"fm5b7t4p0dq3","team_id":"fibggi2jur5s","root":true},
          {"grant_id":"fk3x9r2m5iv8","revision":1,"assignment_id":"fm5b7t4p5iv8","team_id":"fibggi2juubk"}
        ]
      }
    },
    {
      "version": "1",
      "grant_id": "fk3x9r2man0d",
      "revision": 1,
      "parent_grant_id": "fk3x9r2m5iv8",
      "permissions": ["hrms:payroll:payslip::read"],
      "scope": {"dept": "FIN", "cert": "C17"},
      "validity": {"not_before": null, "expires_at": null},
      "source": {
        "assignment_id": "fm5b7t4pan0d",
        "team_id": "fibggi2juxhc",
        "via": "membership",
        "lineage": [
          {"grant_id":"fk3x9r2m0dq3","revision":1,"assignment_id":"fm5b7t4p0dq3","team_id":"fibggi2jur5s","root":true},
          {"grant_id":"fk3x9r2m5iv8","revision":1,"assignment_id":"fm5b7t4p5iv8","team_id":"fibggi2juubk"},
          {"grant_id":"fk3x9r2man0d","revision":1,"assignment_id":"fm5b7t4pan0d","team_id":"fibggi2juxhc"}
        ]
      }
    },
    {
      "version": "1",
      "grant_id": "fk3x9r2mc4p1",
      "revision": 3,
      "parent_grant_id": "fk3x9r2m5iv8",
      "permissions": ["hrms:payroll:payslip::read"],
      "scope": {"dept": "FIN", "region": "APAC"},
      "validity": {"not_before": null, "expires_at": null},
      "source": {
        "assignment_id": "fm5b7t4pc4p1",
        "team_id": "fibggi2jv3ko",
        "via": "membership",
        "adopted_role": {"role_id": "fi9jvxobqsxs", "revision": 2},
        "lineage": [
          {"grant_id":"fk3x9r2m0dq3","revision":1,"assignment_id":"fm5b7t4p0dq3","team_id":"fibggi2jur5s","root":true},
          {"grant_id":"fk3x9r2m5iv8","revision":1,"assignment_id":"fm5b7t4p5iv8","team_id":"fibggi2juubk"},
          {"grant_id":"fk3x9r2mc4p1","revision":3,"assignment_id":"fm5b7t4pc4p1","team_id":"fibggi2jv3ko"}
        ]
      }
    },
    {
      "version": "1",
      "grant_id": "fk3x9r2md7zq",
      "revision": 1,
      "parent_grant_id": "fk3x9r2m0dq3",
      "permissions": ["hrms:payroll:payslip::approve"],
      "scope": {"dept": "ENG"},
      "validity": {"not_before": null, "expires_at": "2026-12-31T00:00:00Z"},
      "source": {
        "assignment_id": "fm5b7t4pd7zq",
        "team_id": "fibggi2jv6bs",
        "via": "membership",
        "lineage": [
          {"grant_id":"fk3x9r2m0dq3","revision":1,"assignment_id":"fm5b7t4p0dq3","team_id":"fibggi2jur5s","root":true},
          {"grant_id":"fk3x9r2md7zq","revision":1,"assignment_id":"fm5b7t4pd7zq","team_id":"fibggi2jv6bs"}
        ]
      }
    }
  ]
}
```

### What the four demonstrate

| | via team | depth | shows |
|---|---|---|---|
| `…5iv8` | Team1 | 2 | the short chain — one hop off the root |
| `…an0d` | Team2 | 3 | the AND — `dept=FIN` from Team1 **and** `cert=C17` from Team2 |
| `…c4p1` | Team3 | 3 | a role-based grant, permissions resolved, `adopted_role` kept as explanation |
| `…d7zq` | Approvers | 2 | a separate branch off the same root, different permission, a validity window |

**The root is in every lineage and is nobody's grant here.** It is the ceiling
all four descend from. It appears in `resolved_grants` only for a member of the
root-holding team — the tenant administrator.

**Two grants carry `payslip::read` with different scopes.** Both ship. Q-130 says
different routes may cover different items in one batch, so the client needs the
set, not a winner.

---

## 4 · Field rules

| field | rule |
|---|---|
| `version` | required string, rejected if missing or unsupported — CONTRACT-010 |
| `authority_epoch` | **not emitted yet.** The tenant's epoch this was resolved at; a cached document may be used only while it still matches — Q-128 |
| `resolved_at` | **not emitted yet.** Evidence, not a validity input — time is decided by `validity` |
| `permissions` | expanded, never a role reference. Non-empty |
| `scope` | effective. `{}` means the whole area, which is a complete scope, not an absent one |
| `validity` | effective — the narrowest window in the chain. `null` bounds mean unbounded |
| `source.via` | how the human reaches it. `membership` today; groups-only is deliberate, and a direct human assignment is refused at both write and read |

| `source.adopted_role` | present only when the grant adopted a role. Explanation, never authority |
| `source.lineage` | ordered root-first. Every entry names the grant revision, the assignment that carried it, and the team it went to |

### Who may ask, which is not a field of the answer

| actor type | admitted by the identity rule | admitted by any gate here |
|---|---|---|
| `user` | only when it names itself — naming another human is impersonation, refused before any gate | itself only, in the lab |
| `service_account` | yes, naming anyone | yes, within the area its credential is bound to |
| `agent` | yes, naming anyone | **no** — Q-086 admits the type and no gate in this repository implements delegation for it, which is unsupported rather than refused |

**`source` is annotation, never a decision input.** The gate matches
`permissions`, `scope` and `validity` and nothing else. If an application starts
reasoning over `lineage`, every rule we have settled becomes a rule it must track
in step — which is the whole reason resolution stays on Auth's side.

### A chain may contradict itself, and then the route is dropped

`Narrow` appends each child's predicates to its parent's, so a chain can state
one key twice with different values — `dept=FIN` above, `dept=ENG` below. Every
predicate must hold against one material value, so such a route authorizes
nothing.

**Folding that into a scope object without noticing would turn an unmatchable
route into a matchable one.** So the fold reports the contradiction rather than
overwriting, and the route is left out of the document entirely. This is the one
place where making `scope` an object rather than a predicate list costs
something, and it is why the fold is Auth's job rather than a client's.

### Lineage repeats shared prefixes, deliberately

`…0dq3 → …5iv8` appears three times above. The alternative — a shared assignment
map with references — makes a client join before it can explain one grant. The
repetition compresses well and reads correctly on its own.

---

## 5 · What comes back when there is nothing, and when it fails

These are three different answers and they must stay distinguishable —
`decision-results.md` Q-051 separates a completed decision from an evaluation
error, and the same separation applies here.

| situation | answer |
|---|---|
| the human holds nothing here | **success**, `resolved_grants: []`. A complete answer, not an error. The gate turns it into a deny. |
| every grant expired or disabled | **success**, `resolved_grants: []`. Same reason — resolution completed. |
| the tenant has not installed the application | **not found**. There is no area to resolve in. |
| the caller may not ask about this human | **rejected**. |
| the identity block is missing a part, or names a wildcard | **malformed**. Never a partial document. |
| the version is not `"1"`, or the actor type is outside Q-086, or a `user` actor names another human | **unsupported** — "we do not do that", decided before any gate runs |
| authority could not be established — store unavailable, snapshot ceiling exceeded | **evaluation error**. Not a deny, and the client must not treat it as one — Q-128. |

**An empty document is the most important row.** It is the normal answer for most
humans and most permissions, and reading it as a failure would make every
unauthorised request look like an outage.

---

## 6 · Roles: expanded now, referenced later if size demands it

A role revision is **immutable** — `PublishRole` is add-only and a duplicate
`(role_id, revision)` is `ErrConflict`. So `(role_id, revision) → permissions` is
a cache key that never needs invalidation, which makes a reference form genuinely
attractive for a role with many permissions adopted by many grants.

**It is still not the canonical form**, for three reasons:

1. A reference puts a second fetch on the hot path, and a cache miss before a
   decision becomes an evaluation error rather than a decision.
2. Repeated permission strings compress extremely well. Measure before designing
   a cache protocol.
3. `source.adopted_role` already answers *"why does she have this?"* without the
   client needing the expansion at all.

**And retirement is handled server-side**, which is what makes the option safe to
hold in reserve: if any permission in a role is retired, `CheckContent` rejects
the whole grant at resolve and it never ships. A client cannot allow on a stale
expansion because it never receives the grant.

**The escape hatch, designed in and not built:** the request gains
`expand_roles: false`, and a grant then carries `permissions_ref: {role_id,
revision}` in place of `permissions` — exactly one of the two, never both. Opt-in,
so clients that do not care never implement it.

---

## 7 · Open

1. **`resolved_grants` as the array name** — carries the stored-versus-effective
   distinction in one word. Accepted unless a better name appears.
2. **Does `source` ship always, or on request?** The gate ignores it. A lean hot
   path would omit it and the audit path would ask for it — same contract,
   fewer bytes.
3. **Is the document per `(tenant, application)`?** I believe yes: permissions
   are application-scoped, so the set is, while the epoch is per tenant because
   teams and memberships are.
4. **Does the tenant administrator's own document include the root grant?** It
   follows from the rules — they hold it — but it means one document can contain
   the entire ceiling, and that is worth seeing before it surprises anyone.
