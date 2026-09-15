# The four terms

**Settled.** Written down because everything else hangs on it. These are the words we use for the parts of authorization, and the line
between the third and the fourth is where the architecture is decided.

![Four inputs make one decision](assets/client-consumes.svg)

---

| | what it does | where it runs |
|---|---|---|
| **administration** | changes authority — grants, assignments, teams, roots | Auth service |
| **authority loading** | supplies what a human holds — the routes | Auth service → client |
| **resolution** | decides *this request*, under that authority plus request material | **the client** |
| **enforcement** | binds actual execution to the decision | **the client** |

---

## The split that matters

The tempting version of this has three terms, with resolution described as
*"what a human is entitled to do."* That collapses two different things, and the
collapse decides the architecture in the wrong direction.

**What a human is entitled to do is authority loading** — SVG-001's step 4,
*"valid human memberships, direct/group assignments, their reusable grant
definitions, permissions from adopted role revisions, validity/conditions, and
dependencies."* It is about the human, and it is the same answer whatever the
request is.

**Resolution is about the request.** CONTRACT-005 —
*"whether application facts are still needed to determine authorized access, or
whether the handler is applying complete restrictions already determined."* It
needs material Auth does not have and must not have: the department the
certificate is in, the employee a human maps to.

### Why the line has to fall there

If resolution ran inside Auth, the client would ask **"may this human do this?"**
That moves the decision across the wire, and CONTRACT-006 forbids it —
*evaluation and enforcement remain inside one endpoint-owned authorization gate.*

Because resolution runs in the client, Auth is only ever asked **"what does this
human hold?"** Which is why:

- Auth has no decision endpoint, and must not grow one.
- The client never needs grants, teams or memberships — it needs the *result* of
  loading, not its inputs.
- The gate is embeddable at all: it carries the rules, not a round trip.

---

## Resolution and enforcement do talk to each other

Which direction depends on the endpoint's declared mode — CONTRACT-002, and the
caller never chooses it.

| mode | the conversation |
|---|---|
| **middleware-complete** | resolution finishes and hands enforcement a decision. One way. |
| **endpoint-completion** | resolution cannot finish without application facts, so enforcement's fact-gathering feeds back into resolution, which then completes, and only then is execution bound. A loop. |

The loop is the mode with no implementation today: `Policy` has no `mode` field,
so the gate always completes and *"Middleware never returns allow for this mode"*
is a rule with no code that can break it.

**Prepared is not authorized** — ENFORCEMENT-001. The middle of that loop permits
gathering facts. It permits no output and no mutation.

---

## What this settles

| question | answer |
|---|---|
| Does Auth decide? | No. It loads. |
| Does the client see grants, teams, memberships? | No. It sees routes — loading's result. |
| Where does request material enter? | Only in the client, only at resolution. |
| Who may call authority loading? | An enforcing client. Administration is a separate surface and a separate story. |
| What does the composition root serve? | Two things: administration, and authority loading. Not resolution — that ships to the client. |
