# `abv.ownership` — canonical record

**Implemented.** The last record type the handbook has settled for Auth-AL. After it,
what remains is root establishment, which is blocked on decisions rather than
work.

---

## 0 · What is approved, and what is explicitly not

Q-099 is **agreed at rule level**, and unusually explicit about where the rule
stops:

> *"It does not finalize ownership permissions, owner-list JSON, or parent-link
> fields."*

So this document can settle the **record** and must not settle the **permission**.

| Approved | |
|---|---|
| Ownership is authority to **administer** a team | Q-099 |
| It is **separate** from the team's held business authority | Q-099 |
| Changing owners changes nothing else — not grants, assignments, adopted revisions, parent links, or other memberships | Q-099 |
| Owners are **plural** — *"explicitly replaces Team1's owners"* | Q-099 |
| Being an owner supplies **no** business authority and **no** permission to assign grants | Q-099 |
| A team-held route's continuing source is its supporting assignment, **not** the owner who created it | Q-099 |

**Not approved, and left alone here:**

> *"It does not decide whether `auth:group::write` authorizes ownership transfer;
> that operation's exact permission remains open."* — `team-administration.md`

So the operations take an administrative gate, as every other write does, and
**which** permission that gate demands is the handbook's to settle, not ours.

---

## 1 · The record

Identity is the pair, presence is the fact — the shape membership already has,
for the same reason: a relationship has no id of its own.

```
key1=abv  key2=ownership  key3=<team id>  key4=<human id>
value={}
boundary=tenant, tenant_id=<tenant>
```

### The complete canonical path

```
abv.ownership:fp8h2w6y5iv8:fi7io4lvjqio
└┬┘ └───┬───┘ └─────┬────┘ └─────┬────┘
 │      │           │            └─ the human id · key4
 │      │           └────────────── the team id · key3
 │      └────────────────────────── record type · key2
 └───────────────────────────────── domain namespace · key1

                 key3 and key4 are the relationship, and the
                 relationship IS the identity — one row per owner,
                 and a team's owners are its rows
```

**`key3` is the team, not the application**, and that is the one place this
record departs from the rest. A team is a tenant's, not an application's — the
team record already keys that way, and ownership follows the thing it owns.

**The value is empty**, as a membership's is. Presence is the fact; there is
nothing about an ownership except that it exists.

### Membership and ownership are the same shape and different facts

```
abv.membership:fp8h2w6y5iv8:fi7io4lvjqio    this human receives the team's authority
abv.ownership:fp8h2w6y5iv8:fi7io4lvjqio     this human may administer the team
```

Identical paths but for `key2`, and that is exactly right: they are two different
relationships between the same two things. Q-099 exists to say they are not the
same relationship — **an owner is not thereby a member, and a member is not
thereby an owner.**

> A human may be both, one, or neither. Nothing in the record couples them, and
> nothing should: Q-099's whole point is that owning a team gives you no business
> authority, which is what membership gives.

---

## 2 · The contract

Four operations, the shape membership's already has.

```go
AddOwner    (ctx, area, identity, teamID, humanID) error
RemoveOwner (ctx, area, identity, teamID, humanID) error
ListOwners  (ctx, area, identity, filter)          (OwnerPage, error)
IsOwner     (ctx, area, identity, teamID, humanID) (bool, error)
```

| Rule | |
|---|---|
| `AddOwner` is add-only | adding an existing ownership is `ErrConflict`, not a quiet no-op |
| `RemoveOwner` says whether it did anything | removing an absent one is `ErrNotFound` |
| both halves must exist | the team, and the human — the foreign keys the fold removed, as rules |
| `ListOwners` answers **both directions** | a team's owners, or a human's teams; exactly one filter required, the rule `ListMembers` and `ListAssignments` already hold |

**No `SetOwners`.** Q-099 says an authorized change *"explicitly replaces Team1's
owners"*, which reads like a whole-list operation — but a replace is an add and a
remove that cannot be told apart afterwards, and it can empty a team's owner list
in one call. Add and remove compose to the same effect and say what they did.

> **Open — may a team have zero owners?** Removing the last owner leaves a team
> nobody may administer, which is recoverable only by whoever administers teams
> tenant-wide. Refusing it means a team can never be fully handed over in one
> step. **Recommend allowing it**, because `auth:group::write` is tenant-wide
> team administration (`scope: {}`) and is not lost when a team's own owners go —
> so there is always a way back. Recorded rather than assumed.

### What ownership does NOT do

Worth stating in the contract, because Q-099 exists to prevent exactly this:

- it does not make the owner a member
- it does not give the owner the team's business authority
- it does not let the owner assign grants — *"ownership is not an implicit
  grant-assignment permission"*
- it does not become the continuing source of any route the owner created

---

## 3 · Validation

| Layer | Rule | Fails with |
|---|---|---|
| shape | team id and human id are base-36 Snowflakes | `ErrMalformed` |
| record | the team exists in this tenant | `ErrRejected` |
| record | the human exists | `ErrRejected` |
| record | no existing ownership for this pair — the primary key | `ErrConflict` |
| authority | the administrative gate | `ErrRejected` |

**Nothing else.** There is no ceiling to check and no lineage to walk: ownership
grants no authority, so there is nothing to keep within anything. That makes it
the lightest record in the model — and the reason is Q-099's rule, not an
oversight.

---

## 4 · Scope

1. `abv.ownership` and its storage
2. The four operations
3. CLI verbs, with a Go test against the compiled binary
4. A captured demonstration
5. **Markdown for the four older demonstrations** — role, scope, envelope-drift
   and team-membership have SVGs in the repo and no text, because demo markdown
   only started at the grant. `demos/` is inconsistent and this closes it.

**No table is dropped**, because there are none left. This is the first record
added to a domain that already holds everything in one store.

---

## 5 · What was decided

1. `key2=ownership`, keyed by `(team, human)` at the tenant boundary — the
   membership shape, deliberately. The rows show both record types side by side,
   identical but for `key2`.
2. Add and remove, not a whole-list replace.
3. **Yes** — a team may have zero owners, because team administration is
   tenant-wide and there is always a way back.

The administrative permission is still **not** chosen here, because the handbook
says it is open. `OwnershipAdministration` is its own interface so an adapter may
answer it either way and the handbook can settle it later without changing the
signature.
