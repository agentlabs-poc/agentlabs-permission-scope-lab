# Root establishment — the operation that gives the lifecycle a beginning

**Implemented.** `EstablishAuthRoot` and `EstablishRoot`, the namespace slice on a
root's ceiling, and the two refusals that keep a root a root.

It is not a record type. Every record it writes — grant head, grant revision,
assignment — already existed; what did not exist was any way to write the first
one. This document is the operation, and the two bugs implementing it uncovered.

---

## 1 · The problem, stated once

**Nothing could create a root.** The only writer of `TrustedRoot: true` in the
tree was `internal/storage/sqlite/fixture.go` — a lab seeder.

So every grant in the system traced back to seeded data. A grant a fixture wrote
could be amended, narrowed, assigned, and have its adopted revision upgraded —
but never *started*. The lifecycle had a middle and an end and no beginning.

Three consequences, all carried openly since the grant slice, and all now closed:

- Q-113 (trusted establishment), Q-114 (registration first) and Q-115 (authority
  to a group, human as member) were asserted by absence rather than demonstrated;
- neither lineage demonstration could run, because a human acting needs a route
  and every route needs a root;
- an authority gate that no operation exercises is a gate nothing tests.

The only non-test writer of `TrustedRoot: true` is now
`internal/mutation/establish.go`.

---

## 2 · The two boundaries

This is the framing that settles the rest, and it is the user's rather than the
implementation's.

| | its boundary |
|---|---|
| **the Auth agent** — enforcement, running at the application | exactly `(tenant, application)`. `localadapter` builds an `Area` from the request and evaluates only there. |
| **the tenant administrator** — writes, logged into the Auth service | **two.** The Auth service boundary, *and* the application boundary. |

The second row is the important one. A tenant administrator does not act *inside*
one boundary — they **hold Auth-boundary authority and use it to write records in
the application boundary**. That is why the two lineages connect through a person
and never through a grant.

---

## 3 · A root's ceiling is its own namespace, not the whole catalog

An application's catalog is read as its own permissions **union every platform
permission**:

```sql
WHERE tenant_id='' AND key1='abv' AND key2='permission'
  AND ((boundary='application' AND key3=?) OR boundary='platform')
```

and `rootRoute` used to take **all** of it. So an HRMS root contained every
`auth:*` permission — **including whichever one authorises establishing an
application root.**

> **Whoever held the HRMS root could establish more roots.** The thing created by
> the authority became able to create more of that authority. That is
> self-amplification, and it is the reason this was fixed — not neatness.

And independently: the Auth agent at HRMS is bounded to `(tenant, hrms)`. An
`auth:*` permission is not in that application, so the agent could never
legitimately be asked to enforce it. It has no business in the ceiling the agent
resolves against.

The union stays correct for **evaluation** — a request in HRMS may legitimately
require a platform permission. Only the **root's ceiling** is sliced, and one
rule covers both roots:

> A root takes the permissions registered under **its own namespace** — `key3`.

| Root | namespace | takes |
|---|---|---|
| Auth root | `auth` | the platform permissions, which are registered there |
| application root | `hrms` | HRMS's own, and no `auth:*` |

No discriminator is needed and none was invented. `rootRoute` compares
`definition.Namespace` against `s.Catalog.ApplicationID`, which is what the root
is a root *of*.

### What this cost, against the estimate

This was described to the user as "four lines in `rootRoute`". It was not.

`domain.PermissionDefinition` was `{ID, Active}`, and the snapshot reader read
each row's boundary and namespace and **discarded both**. `rootRoute` cannot
filter on information it never receives. So it was three changes: the domain type
gained `Boundary` and `Namespace`, the reader stopped discarding them, and then
the four lines. Non-breaking — all the literals are keyed, none positional.

`TestAnApplicationRootDoesNotCarryPlatformPermissions` plants
`auth:tenant:application::admin` at the platform boundary and asserts the HRMS
root does not carry it, then asserts the two ceilings are disjoint. It was
verified to fail without the filter.

---

## 4 · Who establishes

**The tenant administrator establishes an application root**, using
**Auth-boundary authority** — held from the Auth root, not from anything in HRMS.

That is what makes it non-circular: the authority to create HRMS's ceiling comes
from outside HRMS entirely. **§3 is what keeps that true** — without the slice,
HRMS's ceiling would contain the authority that made it.

**The Auth root stays platform-established**, because that one genuinely is
circular: it is what makes someone a tenant administrator at all.

| | Auth root | application root |
|---|---|---|
| established by | Auth platform administration | the tenant administrator |
| acting at | the platform boundary | the Auth service boundary |
| writes records at | `boundary=tenant` | `boundary=tenant` |
| triggered by | tenant creation | installation |
| once per | tenant | `(tenant, application)` |

---

## 5 · The operations

```go
EstablishAuthRoot(ctx, area, identity, holderTeamID) (Grant, GrantContent, error)
EstablishRoot    (ctx, area, identity, holderTeamID) (Grant, GrantContent, error)
```

`EstablishAuthRoot`'s area names the platform's namespace rather than an
application, which is how one implementation serves both: the ceiling it computes
is the platform catalog because that is the namespace it is in.

**Three records, one transaction, all or nothing:**

```
abv.grant            head, status=enabled, trusted_root=true   ← the trust evidence
abv.grant_revision   revision 1 — no parent, no permissions, scope {}
abv.assignment       to the holder team, adopting revision 1
```

Two of the three were expressible before; the assignment is why this waited for
the assignment record.

**Atomicity is the mechanism, not a workflow.** Q-117 requires that incomplete
setup provide no authority. One transaction gives that for free: there is no
moment where a root exists without its assignment, because a half-established
root cannot be committed.

### Preconditions

| requires | because | else |
|---|---|---|
| the application is **installed** | a root is otherwise authority inside an application the tenant does not hold. Answered by the existing `Registry` port — every snapshot load already asks | `ErrNotFound` |
| that namespace's catalog has ≥ 1 permission | Q-114: registration precedes acceptance. An empty catalog computes an empty ceiling, and a ceiling of nothing is not a ceiling | `ErrRejected` |
| the holder team exists | | `ErrNotFound` |
| the holder team has **no parent** | `rootRoute` refuses a parented holder at resolution; refusing at the write means the record is never stored in a shape resolution would reject | `ErrRejected` |
| no root exists for this area | add-only; a second root is a second ceiling | `ErrConflict` |
| `scope` is `{}` | a root narrower than its area is not a ceiling | `ErrMalformed` |

The holder id is validated with `codec.ValidRoleID` — team ids are base-36
Snowflakes, so a holder that cannot be a team id is **malformed** rather than
merely absent.

### The gate names no permission

Q-113 leaves the procedure open, so `RootEstablishment` is its own interface with
nothing chosen — the shape ownership took. An adapter answers it; the handbook
settles it later without changing the signature.

`RootAPI` is separate from `GrantAPI` for the same reason, and the CLI puts
establishment under its own `root` command rather than under `grants`. **No grant
operation can reach trust evidence**, and that is the whole of Q-113's
requirement expressed as vocabulary rather than as a check. `grants create`
without `--parent` says so in its refusal.

---

## 6 · Two bugs the implementation found

Both were invisible while the only root in the corpus was a fixture that stored
an explicit permission list. That list was never read — `rootRoute` computes the
ceiling and overwrites it — so it could only go stale and mislead whoever read
the row. Making the fixture root sourceless, like the roots the operation writes,
surfaced both at once.

**An established root supported nothing.** `selectedPermissions` looked the empty
role key up in the role map, missed, and returned `ErrRejected`. So every root the
new operation wrote was correct and inert: no child could hang from it. Root
content selects nothing *locally* — its coverage is the registered catalog,
computed by `rootRoute` (Q-122) — so nothing was named and there is nothing to
check the catalog against. `TestAnEstablishedRootSupportsAChild` establishes a
root, hangs a child, and follows the route back to the assignment written in the
same transaction; it fails without the fix.

**A root could not be read back.** `DecodeContent` demanded a direct list or the
complete role pair, which made root content undecodable: writable and resolvable
but never readable, so `inspect grant` on an established root answered
`malformed`. What belongs there is the *mixtures* — a list beside a role, or half
a role pair. `ValidateContent` then holds sourceless content to the root shape
throughout: no parent, no narrowing.

One existing test changed with the fixture.
`TestRootCatalogPreservesScopeValidityAndRevalidatesSource` narrowed the root to
prove scope reached the route; a root may not narrow, and only the legacy shape
let it. It is now `TestRootCarriesValidityButNeverNarrowing`, asserting both
halves of one rule: time may bound a ceiling, because an expired ceiling still
describes the whole area while it lasts; a predicate may not, because a ceiling
lower than the region it bounds is not a ceiling.

---

## 7 · What the lab needed

**Auth's own catalog was empty** — no `auth:*` permission was registered anywhere,
because nothing had needed one. The demonstration registers them through the
existing `RegisterPlatformPermission`, which is the first time Auth's own catalog
is populated and what makes *"the Auth root computes from the platform catalog"*
visible rather than asserted.

**No fixture had an area without a root**, and "exactly one root per area" refuses
before any other rule is reached — so the operation could not be demonstrated at
all against the worked fixture. `tenant-genesis` is `TeamFINC17` with the grants,
revisions, controls and assignments stripped: a tenant's teams and an
application's catalog, and not one grant. It is what a tenant looks like when it
has just been created.

---

## 8 · The demonstration

`demos/demo-17-root-establishment.md`, captured from a real run.

The line that matters is the same proposal checked against two roots. A child
grant carrying `auth:grant::create` is created under the HRMS root and under the
Auth root — creation accepts both, because creation checks the *catalog*. Then
`check assignment` resolves each:

```
under the HRMS root   operation rejected or record not found        rc=3
under the Auth root   permissions     auth:grant::create
                      assignment_ids  <the Auth root's own assignment>
```

That is §3 made visible, and it cannot be faked — the two answers come from the
two catalogs. The catalog listing beside it shows all five permissions, `auth:*`
included, in the HRMS area: **the catalog is not the ceiling.**

---

## 9 · What is still open

**The registry has no notion of the platform namespace.** Every snapshot load
asks `Installed(tenant, application)` before reading a tenant's authority, and
the Auth area's application is the platform namespace — which is never installed,
because Auth is not a registered application. The lab's `FixedRegistry` answers
for whichever area it was built with, so the demonstration runs; a real registry
would refuse the Auth area and `EstablishAuthRoot` would never reach its
preconditions.

This sits inside the space Q-113 already leaves open — *"the trusted
operator/procedure ... still need discussion"* — and it is a decision about the
registry's contract rather than Auth-AL's, so it is named here rather than
guessed at.

**Also still open, and listed by Q-113:** root rotation, re-establishment after a
failed setup, and withdrawal. The add-only rule refuses rather than guessing.
