# `abv.permission` — canonical record

What a permission is, the functions that operate on it, how it is stored, and
where it is consulted.

**Draft for review.** Behaviour marked *implemented* is read from
`implementation/abv`. Storage layout is proposed; see `20-storage-encoding.md`.

---

## 1 · What it is

> A **permission** names an operation that authority can cover. It says what may
> be done — never who may do it, and never which records it reaches.

Reach comes from a grant's scope, registered separately. Separating the two is
why the model needs no new permission per department or repository: *read
payslip* is one operation whether it reaches one person's payslips or a whole
department's.

A permission belongs to an **application**, never a tenant. Under Q-123 there is
one shared catalog per application. Adding a permission gives every tenant a new
*word*, never new access.

It is **not** authority, **not** a boundary, **not** hierarchical — a shared
prefix confers nothing — **not** revisioned, and **not** deletable.

### Independent from scope

![Permissions and scopes are registered independently; the only link is compatibility mode](record-permission-model.svg)

Two validation loops that never consult each other: a grant's permissions must
each be registered and active, its scope keys must each be registered. Neither
asks anything of the other. The single exception is section 5.

### The identifier

```
hrms:employee:certificate::read
accounting:ledger:entry::post
```

Convention is `<app>:<domain>::<verb>`. Under Q-126 an identifier can never be
repurposed for a different authorization meaning, even after retirement — which
is why the record is add-only and never deleted.

> **Open.** Auth-AL does not enforce the shape. `codec.PermissionList` rejects
> only blank strings, invalid UTF-8, `*` and duplicates. The Agent is stricter —
> `authmiddleware.validPermission` requires exactly one `::` with non-empty
> halves — so Auth-AL will store an identifier the Agent later refuses.

---

## 2 · The contract

Four functions. Nothing else is exposed for this record.

```go
RegisterPermission (ctx, app, identity, def)             (PermissionDefinition, error)
GetPermission      (ctx, app, identity, id)              (PermissionDefinition, error)
ListPermissions    (ctx, app, identity, filter)          (PermissionPage, error)
SetPermissionStatus(ctx, app, identity, id, active bool) (PermissionDefinition, error)
```

```go
type PermissionDefinition struct {
    ID     string
    Active bool
}
```

**No tenant argument anywhere** — a permission has no tenant dimension.
**Identity on every call, including reads**, so each function runs the injected
administrative check itself rather than assuming a caller did; an application's
catalog is its capability surface. **Returns are self-contained** — nothing
implies a further call.

Shared errors: `ErrMalformed` shape · `ErrUnsupported` identity or provider ·
`ErrRejected` a failed rule · `ErrConflict` identifier exists · `ErrNotFound`
missing read. An error is never a denial, and nothing writes on either.

### 2.1 · `RegisterPermission` *(implemented)*

The only way a permission comes into existence. Add-only.

| Parameter | |
|---|---|
| `app` | which application's catalog |
| `identity` | acting human; version `"1"`, actor type `user`, actor id = human id |
| `def.ID` | non-blank, valid UTF-8, no `*` |
| `def.Active` | must be `true` — no pre-retired registration |

```go
app, _ := domain.NewApplication("hrms")
def := domain.PermissionDefinition{ID: "hrms:employee:certificate::read", Active: true}

f.RegisterPermission(ctx, app, identity, def)
```

![Registration validation chain: each check, and the error it returns](record-permission-register.svg)

Validation and write share one transaction; any failure writes nothing. Gate 1 is
the injected application-platform check, not Auth-AL's. Re-registering is
`ErrConflict` — there is no update.

### 2.2 · `GetPermission` *(proposed)*

Typed, application-scoped read of one identifier. No wildcard, no prefix.

```go
def, err := f.GetPermission(ctx, app, identity, "hrms:employee:certificate::read")
// {ID: "...", Active: true}
```

Retired permissions return with `Active: false` rather than being hidden — a
caller asking for a specific identifier is entitled to learn it exists and is
retired. That is how Q-126's permanence is observable.

### 2.3 · `ListPermissions` *(proposed)*

```go
type PermissionFilter struct {
    Prefix     string // "hrms:employee:"
    ActiveOnly bool
    After      string // cursor: last identifier of the previous page
    Limit      int    // server caps it
}

type PermissionPage struct {
    Permissions []domain.PermissionDefinition // ordered by identifier
    NextAfter   string                        // "" on the last page
}
```

Three filters and no more. `Prefix` is the one selective dimension a permission
has; `ActiveOnly` is a cheap predicate once the range is narrow; `After`+`Limit`
bound the page.

```go
f.ListPermissions(ctx, app, identity, domain.PermissionFilter{Limit: 100})
f.ListPermissions(ctx, app, identity, domain.PermissionFilter{Prefix: "hrms:employee:", Limit: 100})
f.ListPermissions(ctx, app, identity, domain.PermissionFilter{ActiveOnly: true, Limit: 100})
f.ListPermissions(ctx, app, identity, domain.PermissionFilter{After: page.NextAfter, Limit: 100})
```

> **Prefix confers nothing.** It groups identifiers for a human reading a
> catalog. A shared prefix grants no descendant permission, and evaluation never
> matches on a prefix. Keeping this filter out of the authorization path is what
> keeps that true.

**Absent:** mid-string wildcards, regular expressions, query language, any sort
order but identifier, unbounded results. A list matching nothing returns an empty
page, never a fallback.

> **Open.** On PostgreSQL a `LIKE 'x%'` prefix uses a btree index only under the
> `C` collation or a `text_pattern_ops` index. Otherwise it silently degrades to
> a sequential scan — correct results, passing tests, and a catalog that slows as
> it grows. Decide at checkpoint 2.

### 2.4 · `SetPermissionStatus` *(proposed)*

Retirement is reversible, so this is live state, not a one-way transition — the
same shape as `SetGrantStatus` and `SetAssignmentStatus`.

```go
f.SetPermissionStatus(ctx, app, identity, "hrms:employee:certificate::read", false) // retire
f.SetPermissionStatus(ctx, app, identity, "hrms:employee:certificate::read", true)  // restore
```

Idempotent — setting the status it already holds succeeds. The argument is a
`bool` because the record's own field is `Active bool`, unlike the
enabled/disabled strings grants use.

**Nothing cascades in either direction.** Grants and assignments referencing the
identifier are never rewritten, deleted or disabled; they simply stop resolving,
and resume when it is restored. The row itself always stays, because Q-126 makes
the identifier permanent.

**Restoring resumes every route silently.** Nothing is revalidated, because the
record has no dependencies of its own. That is accepted: requiring each affected
grant to re-adopt would make retirement destructive while pretending to be a
flag. The safeguard is that setting the status is itself protected.

### Not in this contract

**`Inspect` is excluded.** The facade's `Inspect(area, kind, id)` does return a
permission today, but it is a generic dispatcher over eight kinds, takes an
`Area` requiring a tenant a permission does not have, returns display rows rather
than a typed definition, and — decisively — **one function covering eight kinds
cannot carry one fixed permission per endpoint**, which ABV-123-05/06 requires.
It stays a CLI and diagnostic tool. `GetPermission` replaces it here.

Also absent: no update, no delete, no bulk register, and no separate accessor for
supported keys — they ride in `PermissionDefinition`.

---

## 3 · The stored row

```
boundary        = application        ← no tenant
tenant_id       = NULL
application_id  = hrms
key1            = abv
key2            = permission
key3            = hrms:employee:certificate::read
key4..key10     = ''
revision        = 0                  ← not revisioned
value           = {"active": true}
state           = enabled
```

The identifier is stored raw in `key3`, colons included. There is no escaping:
the columns carry the structure, and every consumer knows `key2 = permission`
means `key3` is a permission identifier. Where a caller wants one canonical
string, the wrapper renders and parses it at the boundary.

```sql
list    boundary = 'application' AND tenant_id IS NULL
        AND application_id = $1
        AND key1 = 'abv' AND key2 = 'permission' AND key4 = ''
        AND key3 LIKE $2 || '%'
        AND (NOT $3 OR value->>'active' = 'true')
        AND key3 > $4
        ORDER BY key3 LIMIT $5

get     ... AND key3 = $2 AND key4 = ''
```

Range scans on `abv_l1_prefix`. `key4 = ''` is kept in the predicate so the
listing stays correct if a sibling record type is ever added under the same
`key3`; today nothing shares it.

---

## 4 · Where it is consulted

![A permission is consulted at five points: four inside Auth-AL, one on the Agent side](record-permission-resolve.svg)

| | Point | What it checks |
|---|---|---|
| ① | `ResolveHuman` | The requested permission is in the catalog and **active**, before any grant is examined. |
| ② | `CheckContent` | Every permission in a grant revision — listed directly or via an adopted role — is registered and active. |
| ③ | `rootRoute` | A root's permissions are **computed** as every currently active permission, never stored (Q-122). |
| ④ | `Narrow` | A child grant's permissions each appear in the parent's. A child can never introduce one. |
| ⑤ | `validateRoute` *(Agent)* | Route permission equals request permission **exactly**. No prefix, no wildcard, no hierarchy. |

① is why retirement works without touching a grant: the flag is read before
anything else, so flipping it stops every future evaluation across every tenant
immediately.

---

## 5 · Permission–scope relationship validation

The handbook records the feature and nothing else:

> An application explicitly declares whether permission/scope compatibility
> validation is enabled. The feature is optional; its mode is not guessed for
> individual grants. When enabled, every relevant grant must satisfy the declared
> support relationships.
> — `handbook/implementation/06-auth-service.md`

**How the support relationships are represented is not canonical.** No record,
no field and no format is approved. Appendix P-11 lists *optional declared
compatibility checking* as pending, so the representation was left open
deliberately.

The prototype carries `Catalog.SupportedKeys` and a `supported_scope_keys` table,
consulted only when `CompatibilityEnabled` is set. That is implementation ahead
of the model, not a settled shape — and it is a relationship *between*
permissions and scopes, so hanging it off the permission record is itself an
unapproved modelling choice.

**Nothing about it belongs in this document until P-11 is decided.** The four
functions above are unaffected: registration does not take supported keys, and
`PermissionDefinition` does not carry them.

> **Gap for the model, not for this record.** Q-041 requires the mode to be
> declared upfront per application, and Q-042 requires the entire existing grant
> population to pass before it can be enabled — rejecting activation and
> reporting incompatibilities rather than rewriting or grandfathering. No
> operation does that. Belongs with the application record and the relationship
> contract, wherever P-11 settles it.

---

## 6 · Open questions

1. **Grammar choice** — this form, the Auth registry's
   `<namespace>:<resource-path>:<action>`, or `adminauthz` typed constants.
   Settle before any identifier is registered: Q-126 makes them permanent.
2. **Grammar enforcement** — Auth-AL accepts identifiers the Agent rejects.
3. **Prefix index** — `C` collation or `text_pattern_ops`, at checkpoint 2.
4. **P-11 — relationship representation.** The feature is canonical, its shape
   is not. Until it is decided, no permission-side field or record exists.

**Settled:** retirement is reversible, so status is one operation. The contract
is four functions.
