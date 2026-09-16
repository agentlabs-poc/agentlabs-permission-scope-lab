# What the lab lives with, and the service has to settle

This repository is a lab. It models the authority architecture; it never becomes
the Auth service, and the application side migrates into HRMS. So some things
are deliberately left alone here because the place to answer them is the service
this work moves into — and leaving them alone is cheaper and more honest than
inventing a mechanism the handbook has not agreed.

Each entry says what the lab does, why that is tolerable here, and what the
migration owes.

---

## 1 · The gate finds the route's tenant by the placeholder's spelling

**What the lab does.** `authmiddleware` binds the route's tenant to the trusted
area by reading `PathValue("tenant")`. A policy whose path says `{tenant_id}`,
`{org}` or `{tenant_slug}` is not bound at all, and the gate cannot tell the
difference. A request for another tenant's record then reaches the handler.

**Why the lab lives with it.** One application, one policy set, and every path
spells it `tenant`. The endpoint owns its policy and is written beside the
handler it guards, so the spelling is not something a caller can influence.

**Why it is not fixed here.** The handbook requires the binding — *"route tenant
claims must still be bound to trusted context; field names alone do not prove
relationships"* (`endpoint-policy-format.md:352`) — and does not say how a policy
declares which of its inputs carries the tenant. `endpoint-policy-format.md`
closes with *"Full validation rules … remain open"* and leaves that to Q-050-C.
Adding a `Policy.TenantInput` field here would settle an open contract by
implementation, which is the same error as reading the tenant from a name.

**What the migration owes.** Not the declined machinery — something that makes
the duty CONTRACT-012 assigns to the endpoint checkable. A gate that silently
skips the binding when it cannot find the input is the worst of both: the
handbook put the responsibility on the endpoint, and the endpoint gets no signal
when it has not discharged it. At minimum, a policy carrying a route segment the
gate cannot account for should be refused where policies are mounted, rather than
accepted and left unchecked.

---

## 2 · Whether a trusted root may be deleted

**What the lab does.** Refuses. `DeleteGrant` and `DeleteAssignment` return
`ErrUnsupported` for a trusted root and for any binding of it, matching the
refusals already on the disable paths.

**Why the lab lives with it.** Nothing in the lab deletes a root, so the refusal
costs nothing, and an area whose root is gone has no ceiling for anything —
every demonstration built on it stops meaning what it says.

**Why it is not settled.** No handbook rule prohibits it, and the nearest ones
lean the other way. Q-113 (`bootstrap-authority.md:134-136`) is about
*conferring* root authority — *"ordinary grant creation or modification must not
confer root authority merely by omitting/removing a parent reference"* — and says
nothing about removing one properly established. `bootstrap-authority.md:151-152`
goes further: an established root *"remains an ordinary grant subject to status,
validity, revisions, assignments, and its explicit boundaries."*

Deletion is not named in that list, which is the only reason this refusal is not
flatly against the text. **The lab's refusal to *disable* a root is** — that is
pre-existing behaviour, it contradicts "subject to status", and it is recorded
here rather than quietly kept. The authorized root-change procedure is open
(`root-grant-format.md:90-91`), so both are defaults in unfinished space.

**What the migration owes.** The root-change procedure itself: how a root is
retired, replaced or repaired, under what authority, and what evidence it leaves.
Until that exists, refusing is a default and not an answer.

---

## 3 · The credential is modelled, not issued

**What the lab does.** `lab.WorkloadClient` is an id and `lab.WorkloadToken` a
secret, compared in constant time against one constant. The application is given
the token out of band.

**Why the lab lives with it.** Issuing credentials is the Auth service's
business, and the shape the credential arrives in is what this work is about.

**What the migration owes.** Real issuance, rotation and revocation, and deriving
the calling application from the token rather than from a constant.

---

## 4 · Freshness has nothing to invalidate

**What the lab does.** Nothing caches. Every request asks again, so nothing can
be stale, and `authority.epoch` is unbuilt.

**Why the lab lives with it.** Correct by construction, and the client is strict
about unknown fields — so the epoch cannot be added to the wire without a
version bump, which is itself worth knowing.

**What the migration owes.** A cache and the epoch together, or neither. The
invariant to preserve is the one demonstration 21 captures: withdrawing
authority changes the next answer.

---

## 5 · The allow result's evidence does not reach the endpoint

**What the lab does.** `authmiddleware` computes the contributing grant ids and
`BoundOperation.Execute` has no way to receive them.

**Why it is not fixed here.** `decision-results.md:306-309` requires the
references to be *available in the result*, independently of whether the request
is recorded. Closing it changes a public signature, and audit recording — the
thing that would consume it — is itself unbuilt.

**What the migration owes.** Both together: the evidence on the result, and
somewhere for it to go.
