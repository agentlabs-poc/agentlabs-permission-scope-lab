# 6. Registration, bootstrap and Auth administration

[Contents](../README.md) · [Previous](05-canonical-model.md) · [Next](07-application-integration.md)

**Framework-specific implementation guidance.** This chapter explains the agreed
logical responsibilities. It does not supply a deployed Auth engine, a complete
registration API or a new privileged bypass.

## Register meaning before accepting authority

An application registers its supported permissions and scope definitions with
Auth. Auth validates grants against those registrations and the canonical model.
The application continues to own the domain meaning and the facts needed to
enforce it. Registering `dept` does not teach Auth how to query an HRMS database,
and registering certificate-read does not grant that permission to a user.

Auth's own administrative permissions and scope definitions must also be
registered. Bootstrap is not a reason to accept unknown permissions or skip
validation. Definition management, initial authority and later business use are
different operations with different authorization questions.

**An application registers only in its own namespace.** A permission's first noun
segment is the application: `hrms` may register `hrms:payroll:payslip::read` and may
not register anything under `auth:` or `system:`. This is a boundary rule, not a
naming convention, and the reason is exact. A root grant's ceiling is computed as
every active permission *in its own namespace*, while an application's evaluation
catalog is wider — its own permissions union the platform's, because a request
inside an application may legitimately require a platform permission. The ceiling
is sliced where the catalog is not. Without the slice, an application root would
carry every platform permission, including whichever one authorizes establishing an
application root: the thing created by an authority could then create more of that
authority.

**An identifier is never renamed.** Correct a description or a display label
freely; the identifier itself is immutable. A rename is strictly worse than a
repurpose, because the old identifier stops resolving and every grant referencing it
narrows silently. Register the right identifier and retire the wrong one — both
grants keep working and nothing changes without a record.

**A scope key registered at the platform boundary is Auth's own.** Its values name
Auth's own records, so they are resolved rather than treated as opaque: `team`
carries a team identifier, and a grant naming a team that does not exist is refused
at the write. An application's keys stay opaque, and must — Auth cannot know what
`dept=FIN` denotes. A key claimed at both boundaries is ambiguous, and the catalog
read refuses it by name rather than choosing a winner.

An application explicitly declares whether permission/scope compatibility
validation is enabled. The feature is optional; its mode is not guessed for
individual grants. When enabled, every relevant grant must satisfy the declared
support relationships. Activating it requires the existing population to pass;
otherwise reject activation and report incompatibilities without silently
rewriting grants or grandfathering exceptions. Disabling this optional check
does not disable ordinary scope syntax, registration or authority validation.

## Platform management is not tenant business access

The existing application platform administrator publishes an application's
permission and scope definitions through its own bounded management authority.
This authority is outside the application's tenant business scope; it is not
unbounded administration of everything.

A publisher can introduce an HRMS payslip permission without holding a particular
tenant's payslip-read grant. Publication does not itself permit reading those
payslips. Auth platform administration, application platform administration,
tenant administration and business access must not be collapsed into one vague
administrator role.

## Shared catalog, separately bounded roots

The agreed model has one current application version/catalog for everyone.
There are no selective tenant release pins or catalog-adoption operations.
Each legitimate root computes its effective permission coverage from the
applicable registered catalog, while retaining its own tenant boundary and
lifecycle requirements.

![Shared application catalog feeds separately bounded tenant roots without automatically selecting new child permissions](../../docs/assets/shared-root-catalog-flow.svg)

If the shared catalog adds certificate-delete, otherwise valid roots reflect
that permission. A read-only child does not acquire delete: its explicit
permission selection remains read. Catalog growth does not widen scope, create
memberships, re-enable records or rewrite ordinary adopted revisions.

This is an intentional root behavior. It does not approve a stored `*`, a new
root-source field or automatic root revision/adoption churn. The exact root wire
representation is still pending. When a permission is effectively retired,
root coverage no longer supplies it; retained grant references cannot preserve
that retired permission. The full restoration and scope-evolution contracts
remain unresolved.

## A tenant operates in two namespaces

One tenant holds **two authorities**, and neither implies the other. This is the
distinction the rest of this chapter rests on:

```
                     ACME  (one tenant, two authorities)
  ┌────────────────────────────────────┬────────────────────────────────────┐
  │  acme in the PLATFORM namespace    │  acme in the APPLICATION namespace │
  │           auth: / system:          │               hrms:                │
  ├────────────────────────────────────┼────────────────────────────────────┤
  │  enable hrms for acme              │  create grants beneath the root    │
  │  ▸ establish acme/hrms root        │  create teams, add members         │
  │      names the holder team ────────┼──▶ authorized AGAINST the root     │
  └────────────────────────────────────┴────────────────────────────────────┘
       creates the authority ───────────────▶ which authorizes everything here
```

Holding the platform-namespace authority does not confer the application-namespace
one, and holding the application-namespace authority never confers the
platform-namespace one. The same tenant appears twice in the fourfold split this
chapter keeps apart: Auth platform administration, application platform
administration, tenant administration and business access.

**Root establishment is the tenant's platform-namespace authority.** Establishing
an application's root in a tenant is the closing step of *enabling that application
for that tenant*, and it is authorized by the same authority that enabled it.

Not by a permission, because requiring one is circular and the circle does not
close:

```
  to establish acme/hrms's root
      you need a grant carrying that permission
          a grant must have a parent
              …up to some root
                  which had to be established
                      which needs that permission …
```

Nothing terminates that chain inside the grant model. It has to be started from
outside it, and the authority already outside it — the authority that has already
decided this application should run in this tenant — is the tenant's
platform-namespace authority. A permission could only ever be held by whoever is
already there.

And not by a separate platform operator, because the decision is the tenant's. The
party that turned the application on is the party that says what authority it starts
with, and to whom. The handover is explicit: establishment names the team that will
hold the root, so the act creating the authority also names its first holder.
Afterwards the platform-namespace authority is finished.

The cost is worth stating. The platform-namespace authority is genuinely powerful
and has no grant chain constraining it: whoever holds it for a tenant decides that
tenant's starting authority. That is inherent — something has to start the chain —
and it is bounded elsewhere, by registration being a prerequisite and by root
coverage being computed from the application's catalog rather than chosen.

## Bootstrap establishes a beginning, not a permanent exception

Trusted setup creates or identifies a legitimate human, creates an administrators
group, explicitly assigns the intended root authority to that group, and
explicitly admits the human as a member. No special username is required.

The starting authority is the maximum intended authority within its established
boundary so subsequent ordinary administration can create teams and distribute
equal or narrower authority. Minimal setup means a minimal explicit arrangement,
not an artificially inadequate starting permission set.

The root is legitimate because trusted establishment occurred, not because its
JSON omits `parent_grant_id`. Ordinary bounded creation cannot manufacture root
authority by omitting a parent. After setup, the human uses normal assignments,
membership and lifecycle rules; there is no personal evaluator bypass.

Initial bootstrap authority is unavailable until the entire intended arrangement
is validated and durably established. Partial records may exist without supplying
partial access. A retry after completed setup reports the completed state without
restoring removed memberships or changing authority. An authorized incomplete
setup may continue only for the same revalidated intent, not silently substitute
a different administrator or merge conflicting attempts.

**No separate ceremony record is introduced.** The root grant carries its own trust
marker and the establishing actor is recorded as the actor of that write. An
establishment is an authorized write like any other; what makes it special is what
it creates, not how it is recorded.

**Deliberate recovery is establishment again**, not a distinct operation. A grant
with a dependent can be neither disabled nor deleted, so replacing a root requires
the subtree to be dismantled bottom-up first — which means there is no state in
which two roots of one application coexist in one tenant, and no separate recovery
path is needed to reach a clean one. The exact trust and intent evidence a setup
must carry remains pending.

## Administration is a grant

Administration needed a rule and did not have one. Administrative operations were
declared as separate checks and every implementation compared an identifier to a
constant, so nothing said what such a check should compute — and nothing could be
wrong. Separately, the ownership relation carried **no authority at all**: nothing
in resolution or validation consulted it, so "who may administer this team" was
recorded in a table that no decision read.

**Administrative authority is an ordinary grant.** Its permissions are in the
platform namespace, its chain begins at the tenant's Auth root, and it is created,
assigned, narrowed, disabled, deleted and resolved by the mechanisms Chapter 5
already describes. A tenant has two chains and one set of rules:

```
  ACME'S AUTH ROOT                        ACME/HRMS'S APPLICATION ROOT
  area     acme / auth                    area     acme / hrms
  ceiling  every active auth: permission  ceiling  every active hrms: permission
      │ parent                                │ parent
      ▼                                       ▼
  auth:group::write  { team: Team2 }       hrms:employee:certificate::read
      │ parent                                │   { dept: FIN }
      ▼                                       ▼
  narrower administrative authority        narrower business authority

  SAME walk · SAME narrowing · SAME containment · SAME ceiling rule
```

The namespace slice is what stops the two chains leaking into one another: an
application root's ceiling is its own namespace, so it can never carry an `auth:`
permission.

An administrative operation therefore needs no bespoke check. It declares the
permission it requires and the boundary it acts at, and one evaluation answers it —
the same evaluator the business side uses:

```
  authorize(acme/auth, identity, "auth:group::write", { team: "Team2" })
```

What each operation must state is which permission and which boundary. What nothing
must state is a rule of its own.

### What bounds an operation depends on which side it is

| | Takes | Bounded by |
|---|---|---|
| **Tenant administration** — teams, ownership, grants, assignments, roles, establishment | an area plus something naming a team | **a team**, carried as the platform scope key `team` |
| **Platform administration** — permission and scope registration, catalog reads, application role publication | an application | **the area itself**; no scope key is needed |

Registering a payslip permission into the shared catalog has nothing to do with any
team — it is application platform authority, and one catalog serves every tenant.
So only the tenant side needs the scope key.

### Sideways escalation is impossible rather than guarded

Scope predicates accumulate conjunctively along a chain, and a request carries one
value per key. That single fact supplies the whole containment property:

| Proposed grant | Result |
|---|---|
| `{team: X}` beneath a parent scoped `{team: X}` | **allowed** — the predicates agree, and this is how an owner appoints another owner of the same team |
| `{team: Y}` beneath a parent scoped `{team: X}` | the route carries `team=X` **and** `team=Y`, which no request can satisfy — it authorizes nothing, ever |
| `{team: X}` beneath a parent scoped `{}` | **allowed** — and only an unrestricted grant from the Auth root has `{}`, which is the tenant administrator |

A team's administrator can therefore hand that team to someone else and cannot reach
a sibling. Nothing enforces this; it is what the model already computes.

The third row has a consequence worth stating on its own: an operation with **no
bounding team** — creating a top-level team, or moving one up to become top-level —
is satisfied only by a route carrying no predicate at all. A holder scoped to one
team must not be able to reach that state by moving what it may already write.

**One wart, recorded rather than hidden.** Narrowing *accepts* the contradictory
re-scope in the middle row and builds a route that never matches, rather than
refusing the write. The safety is "authorizes nothing", not "cannot be written". A
nonsense administrative grant can therefore be stored, and grant health is where it
should surface.

### Ownership is a grant, and the relation is superseded

Creating a team confers its ownership: the creator receives an administrative grant
over the team it just created. A team must always retain at least one enabled
administrative assignment, so no team is ownerless. The earlier ownership relation
is not pending work — it is superseded, because ownership is a grant rather than a
table.

### The recursion, and where it terminates

The grant administering a team is held by a team, which is itself administered.
That terminates where the business chain terminates — at a root — and the first link
is the platform-namespace authority above, which is outside the grant model by
construction. It is a chain someone has to be able to read, and authority loading
already carries lineage for exactly that.

## Two checks inside Auth's own endpoint gate

**Placement clarification — AUTH-ARCH-02:** the user names the Auth-service-side
authority-boundary validator **core Auth Validator / ABV**. API services host
the **Auth Middleware / evaluator**. Auth Service's own APIs use the same
API-side authorization and then ABV where the operation requires authority-change
validation. ABV is not deployed into every application API service by this model.
Shared canonical primitives do not merge these two responsibilities. See the
[implementation planning pin](../../implementation/plan/auth-middleware-implementation-plan.md).

Auth Service protects its administrative APIs with the same framework. The
endpoint declares one required administrative permission and its input bindings.
Its handler must establish both operation authority and the proposed authority's
boundary before the protected change takes effect.

![Auth request passes administrative evaluation and authority-boundary validation before persistence](../../docs/assets/auth-service-authority-gate.svg)

| Responsibility | What it establishes |
|---|---|
| Administrative evaluator | The caller may perform the declared management operation within its administrative scope, including the relevant recipient boundary. |
| Authority-boundary validator | The proposed authority has valid permitted source support, preserves permission/scope limits and applicable team ceilings, and respects affected bindings. |

The first evaluator is not scope-blind: it evaluates administrative scope. The
second responsibility concerns the business or administrative authority being
distributed. Neither substitutes for the other. Shared resolution primitives
may serve both; two responsibilities do not mandate two deployed services.

## Worked assignment: Maya, Team1 and Team2

Continue Chapter 5's G1/A1 example. Team1 holds G1 revision 2, supplying FIN
certificate read/write under valid upstream support. Team2 is explicitly a child
of Team1. Maya is a valid Team1 member and separately has assignment-create
authority for Team2. Nutan is a member of Team2. None of these is an inferred
ownership relationship.

The administrative revision below is assumed validly assigned to Maya through
her administration group, with current support under G-AUTH-ROOT. It is an
approved-core-shape example, not a new owner bundle:

```json
{
  "version": "1",
  "grant_id": "G-ASSIGN-TEAM2",
  "revision": 1,
  "parent_grant_id": "G-AUTH-ROOT",
  "permissions": ["auth:assignment::create"],
  "scope": {"team": "Team2"}
}
```

The proposed child content and resulting assignment are:

```json
{
  "version": "1",
  "grant_id": "G2",
  "revision": 1,
  "parent_grant_id": "G1",
  "permissions": ["hrms:employee:certificate::read"],
  "scope": {"cert": "C17"}
}
```

```json
{
  "version": "1",
  "id": "A2",
  "grant_id": "G2",
  "grant_revision": 1,
  "recipient": {"type": "group", "id": "Team2"},
  "status": "enabled"
}
```

Assume the G2 control is enabled, revision 1 is latest for G2, both scope keys
are supported for this operation, and every required lifecycle condition holds.
The proposed assignment is not considered persisted merely because it is shown.

![Grant and team lineage for the worked assignment](../assets/model-lineage.svg)

Auth reads Team2's real parent Team1, finds the current G1 assignment there and
uses A1's adopted revision 2. At most one current assignment may exist for a
tenant/grant/recipient combination, including disabled records. G1 held at an
unrelated TeamX does not replace the required Team1 holding.

Maya's source access and administrative authority are checked separately. Read
is a subset of G1's read/write; C17 adds narrowing to FIN. Nutan's resulting
route is therefore FIN AND C17 read, subject to all support and endpoint checks.
It is not write, all Finance, or C17 regardless of department.

Changing the submitted permission to delete fails the source-boundary check.
Changing the recipient to Team3 cannot pass the shown administrative route.
The distinctions explain the outcome without adding ownership powers.

## Validate the whole result and preserve it through the write

The validator needs current controls, selected content, actual supporting
assignments, relevant relationships, registration and affected reverse
dependencies. A grant definition found by ID is not proof of usable source
authority. Parent resolution is top-down through actual adopted support;
structural disable/removal work is bottom-up through affected bindings.

If a conflicting publication, membership change, withdrawal or binding change
invalidates the checked state before persistence, stop that attempt without a
partial assignment write. Do not silently choose a different revision or apply
stale approval. A fresh attempt must validate current authority. A transaction
or conditional-write mechanism can implement this guarantee; the handbook does
not choose a database product or new version-token field.

Membership administration is a separate authorized operation: `auth:group::write`
includes management of human members in its administrative scope. It distributes
the team's existing valid access; it does not authorize changing the team's
grants. This chapter does not add an unapproved business-permission-possession
check to membership writes.

**A synchronization is an ordinary authorized caller.** An application that
synchronizes its business membership into Auth does so as a service account holding
team-write authority within a definite scope, calling ordinary endpoints. It has no
privileged path and cannot write the store directly — nothing can; authority records
are changed through authorized operations or not at all. Whether a deployment offers
one membership write per call or one call that changes many is endpoint design, not
an authorization question: a bulk write inside one team is one boundary and one
evaluation, and one spanning several teams is governed by the existing per-item
coverage rule — complete support for every item before any effect, no fragment
mixing, no partial success. A directory that disagrees with Auth is not thereby
right.

**A team's dependants include the administrative ones.** Deleting a team is refused
while an assignment names it, and moving a team is refused while an enabled binding
sits at or beneath it. Both rules read *both* chains, because an administrative
assignment names the same tenant-wide team record a business assignment does. Reading
only one would make the team holding a tenant's Auth root deletable whenever it
happened to hold no business assignment — which takes away all administrative
authority for that tenant, with no record of why.

## What an audit consumer may rely on

Audit belongs to another layer and stays there. But the *producing* half is
specified here — an allow carries the ordered contributing chain — so what a consumer
in that layer may rely on has to be stated, or whoever builds it will infer its
requirements from an implementation.

A consumer is given the contributing chain, ordered root first; it is present on
every allow whether or not the request is recorded; it is handed to the effect
*before* the effect runs, so a record can name the change it authorized; and nothing
else — no scope echo, and no revision, assignment or group per step.

A consumer must not assume that grant ids stay resolvable, because they name records
that may later be deleted: anything needed to reconstruct a decision after the fact
must be captured at the time. Nor that the evidence is a complete account of the
decision — it names the route that authorized, not the routes considered, not the
boundary evaluated, and not the material the endpoint bound. Nor that producing it
implies recording it.

The cost is honest: this describes an interface with no consumer in view, so nothing
tests whether what is handed over is sufficient. The first real recorder may find it
is not, and that is the right place to discover it.

**Pending:** direct-human source eligibility, cross-recipient `$self` binding, a wire
contract for administrative operations, and root/registration evidence are tracked
in [the pending register](../appendices/pending.md). No new field or guessed
support route fills those gaps.

**Sources:** [registration and scope opacity](../../docs/application-registration.md),
[namespace ownership and permanence](../../docs/permission-lifecycle.md),
[bootstrap and the two namespaces](../../docs/bootstrap-authority.md),
[administration is a grant](../../docs/administrative-authority.md),
[the platform scope key](../../docs/scope-model.md),
[ownership](../../docs/ownership-lineage.md),
[membership synchronization](../../docs/groups-and-membership.md),
[dependents and dismantling](../../docs/grant-lifecycle.md),
[root evolution](../../docs/root-permission-evolution.md),
[platform authority](../../docs/application-platform-authority.md),
[combined validator](../../docs/authority-boundary-validation.md),
[audit consumers](../../docs/authority-change-audit.md),
[write consistency](../../docs/auth-write-consistency.md).

[Next: application integration](07-application-integration.md)
