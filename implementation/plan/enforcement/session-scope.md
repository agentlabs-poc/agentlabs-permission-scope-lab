# The session surface, and two words we must not lose

**For review. A thought, not a design.** Raised in discussion and recorded
because it is cheap to settle now and expensive to reconstruct later. Nothing is
being built from this.

---

## 1 · The surface we had not discussed

When a human logs in, the login asks for specific capabilities — and that request
can span **several authority boundaries**. Read HRMS payroll *and* write in asset
management is one login, two applications.

Everything designed so far is per `(tenant, application)`. A session is not.

| | scope of the question | shape |
|---|---|---|
| **resolved authority** | one `(tenant, application)`, one human | everything the human holds — [94](resolved-authority.md) |
| **the session** | one login, possibly several applications | what this login asked for |

**The session set is always a subset.** Asking can narrow, never widen — the same
principle as RESOLUTION-001: more information cannot invent authority.

---

## 2 · It is a fifth input, not a redesign

The four inputs in [93](terms.md) become five: identity, permission,
material, resolved authority, and **what this session asked for**.

Nothing already agreed breaks. A cross-application session is *N* resolve calls,
one per application, because the document is already per application. The resolve
contract needs no change.

---

## 3 · Q-128 decides how it is built

| | |
|---|---|
| **a** — the token carries the authority Auth resolved at login | fast, and **forbidden**. A token outlives a confirmed reduction, which is exactly what Q-128 prevents |
| **b** — the token carries the *request*; the agent intersects it with freshly resolved authority at each check | every freshness rule stays intact, and costs nothing extra because the agent resolves anyway |

It has to be **b**, and it yields something pleasing:

> **effective = requested ∩ resolved**

So an **over-broad session request is harmless by construction**. Auth never has
to police what a login asks for, because asking for more cannot obtain more.

**Auth issues no token and never sees the login.** The requested set is a claim
the identity provider puts in the token; the agent intersects. That is consistent
with authentication already being a port the application fills rather than
something Auth provides.

---

## 4 · Two words mean different things in login-land

One of these was spotted in discussion; the second is worse, because it is the
same word for two things a token would plausibly carry at once.

| word | in our model | in login-land |
|---|---|---|
| **grant** | an authority record — revisions, lineage, a root | the OAuth authorization grant, and grant *type* |
| **scope** | the boundary predicates — `dept=FIN` | the list of permissions being requested |

**We keep both words and qualify the login side.** Renaming OAuth's terms would
confuse everyone who knows OAuth and leave us owning a translation layer forever.

| term | meaning |
|---|---|
| **login grant**, **login grant type** | the OAuth authorization grant |
| **requested scope** | the permission list a login asked for |

*Requested scope* rather than *login scope*: it is self-describing, and it is what
the OAuth request parameter actually means.

### The invariant that makes qualification safe

**The unqualified word always means ours.** Bare `grant` is an Auth-AL authority
record; bare `scope` is the boundary. The login-side concepts are never written
unqualified — not in the handbook, not in a commit message, not in conversation.

We already hold this rule elsewhere without strain: `grant`, `grant_revision` and
`grant control` only stay distinct because the qualifier is mandatory.

### And one code-level rule, so it cannot leak

Inside a token, OAuth's fields keep their own names — `scope`, `grant_type` —
because that payload is theirs. But **our types never carry an unqualified
login-side field**:

```go
type Session struct {
    RequestedScope []string  // not Scope
    LoginGrantType string     // not GrantType
}
```

The collision then exists only at the one boundary where we parse their payload,
which is precisely where someone is already thinking about it.

---

## 5 · Not being decided here

Token format, who issues it, how the requested scope is expressed across
applications, consent, refresh, and what a session means for an agent acting on a
human's behalf. All of it waits. What is settled is the shape of the answer —
intersection, not issuance — and the words.

---

## 6 · Parked, with the simplifying assumption that makes it safe to park

**Decision:** assume a login asks for **full access** — everything the human is
entitled to. The whole session surface then collapses, and we implement `resolve`
first.

It is not a shortcut. It is what a login means when it does not narrow:

- the session ceiling equals what the human holds, which is exactly what `resolve`
  returns, so **there is no intersection step to build**;
- the token carries **no authority**, so there is no snapshot and nothing can go
  stale — Q-128 is satisfied trivially rather than managed, and the epoch becomes
  a performance concern rather than a correctness one;
- the fifth input disappears and [93](terms.md)'s four stand
  unchanged.

**It forecloses nothing, which is the test of a safe deferral.** When narrow
logins arrive, the token gains `requested_scope`, the agent intersects, and
`resolve` is untouched — the narrowing sits on top of it rather than inside it.

**One thing to put in the token now anyway:**

```json
"authority_request": "full"
```

An explicit claim, not an absent field. Otherwise a later missing
`requested_scope` means two different things — an old full-access token, or a
narrow token that asked for nothing — and that is a migration nobody needs to
inherit.
