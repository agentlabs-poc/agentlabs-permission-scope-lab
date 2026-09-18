# Administrative authority is a grant — Q-155 / ADMIN-007

Status: **AGREED.** The shape is the user's: *"there should also be some kind of
grant and assignment that should come along with ownership… or something like the
scope should cover it."* What follows is that framing traced through the rules
already agreed, and it turns out to need almost no new mechanism.

This closes the question HC-05-08 asks — how administrative bounds are encoded and
how containment is validated — by **reusing the authority model rather than
building a second one beside it**.

---

## The problem it replaces

Administration had no rule. Twenty-eight administrative gates were declared as
interfaces and every implementation compared an identifier to a constant. Nothing
said what a gate should compute, so nothing could be wrong — and a gate-ordering
defect survived a week of green tests because the gates could not tell an
authorized caller from an unauthorized one.

Separately, [Q-099](ownership-lineage.md)'s ownership relation carried **no
authority at all**: nothing in resolution or validation consulted it. So "who may
administer this team" was recorded in a table that no decision read.

## Two chains, one machinery

A tenant already has two roots, and [Q-153](bootstrap-authority.md) already named
the two authorities they anchor:

```
  ACME'S AUTH ROOT                        ACME/HRMS'S APPLICATION ROOT
  area     acme / auth                    area     acme / hrms
  ceiling  every active auth: permission  ceiling  every active hrms: permission
      │ parent                                │ parent
      ▼                                       ▼
  auth:group::write  { team: juubk }       hrms:payslip::read  { dept: FIN }
      │ parent                                │ parent
      ▼                                       ▼
  narrower administrative authority        narrower business authority

  SAME walk · SAME narrowing · SAME containment · SAME ceiling rule
```

**Administrative authority is an ordinary grant** whose permissions are in the
platform namespace and whose chain begins at the Auth root. It is created,
assigned, narrowed, disabled, deleted and resolved by the mechanisms that already
exist. [Q-151](permission-lifecycle.md)'s namespace slice is what stops the two
chains leaking into one another: an application root's ceiling is its own
namespace, so it can never carry `auth:` permissions.

## The gates collapse into one evaluation

An administrative operation does not need a bespoke check. It declares the
permission it requires and the boundary it acts at, and one evaluation answers it —
the same evaluator the business side uses:

```
  authorize(area{acme, auth}, identity, "auth:group::write", { team: "fibggi2juubk" })
```

Twenty-eight interfaces become one call. What each operation must state is which
permission and which boundary; what nothing must state is a rule of its own.

## What bounds an operation depends on which side it is

The gate signatures already divide, and the division is the answer:

| | Takes | Bounded by |
|---|---|---|
| **Tenant administration** — teams, ownership, grants, assignments, roles, establishment | `domain.Area` plus something naming a team | **a team**, carried as scope |
| **Platform administration** — permission and scope registration, catalog reads, application role publication | `domain.Application` | **the area itself**; no scope key is needed |

Registering `hrms:payroll:payslip::read` into the shared catalog has nothing to do
with any team — it is [Q-121](application-platform-authority.md)'s application
platform authority, and one catalog serves every tenant ([Q-123](root-permission-evolution.md)).
So only the tenant side needs the new scope key.

## Why sideways escalation is impossible rather than guarded

Scope predicates accumulate conjunctively along a chain, and a request carries one
value per key. That single fact supplies the whole containment property:

| Proposed grant | Result |
|---|---|
| `{team: X}` beneath a parent scoped `{team: X}` | **allowed** — the predicates agree, and this is how an owner appoints another owner of the same team |
| `{team: Y}` beneath a parent scoped `{team: X}` | the route carries `team=X` **and** `team=Y`, which no request can satisfy — it authorizes nothing, ever |
| `{team: X}` beneath a parent scoped `{}` | **allowed** — and only an unrestricted grant from the Auth root has `{}`, which is the tenant administrator |

So a team's owner can hand that team to someone else and cannot reach a sibling.
Nothing enforces this; it is what the model already computes.

**One wart, recorded rather than hidden.** Narrowing *accepts* the contradictory
re-scope and builds a route that never matches, rather than refusing the write. The
safety is "authorizes nothing", not "cannot be written". A nonsense administrative
grant can therefore be stored, and grant health is where it should surface.

## Rationale / conscious tradeoff

The alternative was a second authority model for administration — its own bounds,
its own containment, its own validator. It was rejected because the model already
has one that is built, tested and demonstrated, and because a second one would have
to be kept in agreement with the first forever.

The cost is recursion: the grant administering a team is held by a team, which is
itself administered. That terminates where the business chain terminates — at a
root — and the first link is [Q-153](bootstrap-authority.md)'s platform-namespace
authority, which is outside the grant model by construction. It is a chain someone
has to be able to read, and the transport already carries lineage for exactly that.
