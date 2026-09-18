# Administrative authority is a grant — Q-155 / ADMIN-007

Status: **AGREED**, and implemented one family deep — see *As implemented* below.
The shape is the user's: *"there should also be some kind of grant and assignment
that should come along with ownership… or something like the scope should cover
it."* What follows is that framing traced through the rules
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

---

## As implemented

The rule is adopted for every administrative operation and **implemented one
family deep** — the team family. That is deliberate: the model is worth proving
against a real chain before twenty-eight call sites are rewritten to it, and a
half-rewritten gate set is worse than an honest boundary.

**What resolves.** `CreateTeam`, `SetTeamParent`, `DeleteTeam`, `AddMember` and
`RemoveMember` no longer ask a deployment whether the caller may act. Each states
the permission it requires and the team that bounds it, and one call answers:

```
  authorize(auth chain, identity, "auth:group::write", { team: "fibggi2juxhc" })
```

`CheckTeamCreate`, `CheckTeamWrite` and `CheckTeamDelete` are **deleted** from the
administration interfaces rather than left unused. A gate nobody calls is a gate
somebody re-wires.

**Where the chain is read.** The administrative chain is in a different *area*
from the operation it authorizes — `acme/auth`, not `acme/hrms` — so the provider
reads both inside the one write transaction (`UpdateAdministered`). Reading it
afterwards on a second connection would be a check that a concurrent revocation
could outrun. A store that has not been told where the platform namespace is
refuses every administrative write rather than resolving `auth:` authority against
an application's own chain, which is the leak [Q-151](permission-lifecycle.md)
exists to prevent.

**The scope value is resolved, not opaque.** A `team` value is checked against
the tenant's teams at the write. [Q-148](scope-model.md)'s opacity still holds for
an application's keys — Auth cannot know what `dept=FIN` denotes — and does not
hold for a key naming Auth's own records.

**What the lab demonstrates.** Two holders, captured from the CLI against a real
store:

```
  fk3x9r2mau01   {"scope":{}}                                    ← the Auth root
  fk3x9r2mau02   create + delete + write, {"scope":{}}           ← maya
  fk3x9r2mau03   write, {"scope":{"team":"fibggi2juxhc"}}        ← priya

  priya  add-member  team=fibggi2juxhc   rc=0   her team
  priya  add-member  team=fibggi2juubk   rc=3   same permission, another team
  maya   add-member  team=fibggi2juubk   rc=0   no predicate, so nothing to fail
  priya  team create --parent fibggi2juxhc  rc=3   write is not create
  priya  team create --name Rootish         rc=3   no bounding team to satisfy
```

And the sideways grant behaves exactly as the wart above records it: a grant
scoped `{team: FIN}` beneath one scoped `{team: C17}` is **written**, resolves,
and produces a route carrying `team` pinned to both values — which no request can
satisfy. A test asserts the two predicates rather than asserting a refusal,
because "authorizes nothing" is the property and "cannot be written" is not.

**What is still a fixture, and why it is not a lie.** Every other administrative
gate — grants, assignments, roles, ownership, establishment, and all of platform
administration — still compares an identifier to a constant. The rule above
governs them; the implementation has not reached them. Two pieces of this decision
are deferred with them, because both need a cross-area *write* rather than the
cross-area read landed here:

- team creation issuing the creator's own administrative grant and assignment
  ([Q-157](ownership-lineage.md)'s "creating a team confers ownership")
- the refusal that keeps a team from losing its last enabled administrative
  assignment

Until those land, ownership remains what [Q-099](ownership-lineage.md) found it
to be: a relation nothing consults.
