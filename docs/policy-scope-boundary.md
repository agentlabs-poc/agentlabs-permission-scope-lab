# Policy-declared scope boundary — Q-144, OPEN

Status: **RAISED, NOT APPROVED.** Framing only. No rule here is adopted, no
policy field is added, and nothing in the lab implements it. This chapter exists
so the question is not re-derived from the case that exposed it.

Raised by the user while a narrower question was being asked: *"this challenges
the existing design itself… the policy should also mention the boundary itself."*

## The narrow question that exposed it

Maya holds one grant:

```json
{ "grant_id": "fk3x9r2m5iv8",
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": { "dept": "FIN" } }
```

Three requests against an application's certificate collection:

| Request | What it asks for | Today |
|---|---|---|
| `GET /departments/FIN/certificates` | one department, the one she holds | allowed |
| `GET /certificates` | every department | denied — [Q-071](collection-enforcement.md) |
| `GET /certificates/count` | every department | **undecided** — [counts are named open](collection-enforcement.md) |

The count looked like a small follow-on: deny it like the listing, allow it
because no record is returned, or answer a number narrowed to her boundary.

**It is not a small follow-on.** All three options argue about the *shape of the
output* when the thing that differs is the *boundary of the request*. The listing
and the count ask for exactly the same records. Answering them differently would
mean the boundary a caller must hold depends on how the answer is formatted.

## The problem underneath

**Nothing in a policy says what boundary its endpoint operates at.** A policy
declares `version`, `method`, `path`, one `permission`, its `inputs`, and — since
[Q-133](endpoint-policy-format.md) — its `trusted` correlations. The boundary
reaching evaluation comes from the binder, at request time, as material.

So the endpoint decides both what it will read *and* what boundary it claims to
be reading at, and the two are not tied together. A collection endpoint may
declare the material of a single department and then read every department. The
gate allows it, because the only thing it can compare is what the binder said.

That is the same class as [Q-138](endpoint-authorization.md) — a duty the endpoint
owns and the gate cannot check — with one important difference. Response
disclosure genuinely cannot be checked by a gate. **This one can**, if the policy
declares it, because the declaration is available at mount and the comparison is
against the caller's scope.

## The proposal, as raised

The policy declares the scope boundary the endpoint operates at. The gate then
asks one structural question: **does the caller's scope cover the boundary this
endpoint needs?**

- Maya's authority is bounded at `{dept: FIN}`.
- The listing and the count both operate at the tenant-wide boundary.
- Tenant-wide is not covered by `{dept: FIN}`, so both are refused — **and the
  count question dissolves**, because it was never about counts.
- A policy may declare the widest boundary explicitly, which only an unrestricted
  grant covers.

There is a symmetry worth keeping: `{}` already means "no local restriction, the
whole application boundary" in a grant. The same token read from the other side
would mean "this endpoint needs the whole application boundary". One vocabulary,
two directions.

## What it would change

**[Q-071](collection-enforcement.md) becomes a consequence rather than a rule.**
"An all-values selection denies rather than deriving an authorized subset" would
fall out of coverage: an all-values request is a request at the wider boundary,
and a narrower grant does not cover it. A simplification that removes a rule
rather than adding one.

**A third declared field on a published contract.** After `inputs` and `trusted`.
That is not a tidy-up; CONTRACT-012's field list would be amended a second time,
and the same reservation the user recorded against Q-133 applies here.

**The binder's material becomes checkable.** Either it is verified against the
declaration, or the declaration replaces it.

## The five questions to settle

1. **A new field, or derivable?** A policy already declares `dept` and `cert` as
   inputs — and also `title`, which is not a boundary. Something must distinguish
   a scope key from an ordinary input: a new field, or a mark on the inputs.
2. **Keys, or keys and values?** The expectation is that the policy declares the
   *keys* that bound it and the request supplies the values, so a policy
   declaring no keys is at the widest boundary. Unconfirmed.
3. **What does "covers" mean?** `{dept: FIN}` covering a `{dept}` boundary is
   clear. Whether anything covers the widest boundary except an unrestricted
   grant is not.
4. **Does the binder still supply material?** Checked against the declaration, or
   replaced by it.
5. **What happens to Q-071's chapter?** Superseded as a rule and retained as
   history, or kept and cross-referenced as the same rule stated twice.

## What this chapter does not do

It does not answer the count. Counts remain open where they were, and the
recommendation here is **not to answer them on their own** — an answer reached
about output shape would have to be revisited the moment a boundary declaration
exists.

It does not weaken Q-071. Until this is settled, an all-values selection denies,
exactly as agreed.
