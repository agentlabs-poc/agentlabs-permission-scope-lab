# Working handbook chapter: scope boundaries

**Current refinement — Q-158 / SCOPE-012:** [a scope value is a non-empty list](#q-158--scope-012--a-scope-value-is-a-list-a-child-narrows-by-subset);
a record satisfies a key when its value is a member of the list, keys combine with
AND, and a child narrows by per-key subset — a child that is not a subset is
refused at write. This supersedes the single-string-value format and the "no
arrays / no OR" statements below, which are preserved as history. JSON examples
written before Q-158 show `"dept": "FIN"`; read them as `"dept": ["FIN"]`.

## Post-0.0.1 update — Q-090 / Q-091

A [grant definition](grant-assignments.md) now carries permissions and scope
without a recipient; assignments carry recipients. Older recipient-bearing
wording below is deprecated as a layout, not as a complete-authority invariant.

For [dependent subgroup authority](subgroups.md), effective scope is parent scope
AND child additional constraints. Additional `{}` means no further narrowing,
not removal of the parent boundary. Conflicting same-key restrictions do not
overwrite each other. This does not add arrays, OR, nested scope objects, or
new built-in department/group meanings. Ordinary root `{}` still adds no narrower
restriction inside the implied tenant. The combined derived wire format is open.

Q-050-C / CONTRACT-012 keeps scope as a boundary selector while assigning
application relationship enforcement to endpoint implementation, without a
relationship block in policy. Auth may establish authority within supplied
Finance material; the endpoint must ensure the actual operation stays there,
using trusted facts or constrained execution. Matching a path value alone is
not proof about an unchecked certificate. Scope meanings, tenant isolation,
self anchoring, and other mandatory constraints remain unchanged. See
[the grant and constrained-lookup example](endpoint-policy-format.md).
Earlier requirements for declarative relationship bindings are historical;
application meanings and mandatory enforcement remain necessary.

Q-043 / TERM-005 aligns this chapter's vocabulary with permissions, scope
boundaries, requests, and request material. [Reasoning and case checks](authorization-vocabulary.md)
explain why no additional canonical entity is required. [Superseded wording](history/q043-vocabulary.md)
is preserved exactly; existing safety requirements remain unchanged.

## Current agreed foundation — recording resumed after Q-033

Historical checkpoint after Q-039: REGISTRATION-001 agrees application registration of supported
permissions and scope contracts, with Auth validating grant acceptance without
domain interpretation. See [application registration](application-registration.md).
Earlier Q-039-proposed statements are preserved history; registration format,
compatibility declarations (Q-040), and lifecycle remain open.

Q-038 subsequently settles application ownership of scope-key meanings and
explicit supported resource relationships, with endpoint-specific trusted fact
bindings (SCOPE-004's conceptual core). Earlier broad statements that key
definitions remain open must now be read with that agreement. Shared definition
governance, representation, and validation mechanics remained open at Q-039.
Current update: Q-040/041 settle optional relationships with an explicit
application-level validation choice; Q-042 settles the activation safeguard.
Remaining registration lifecycle detail is parked. These historical notes
do not reopen those decisions or the Q-038 ownership agreement.

SCOPE-006 supplies the canonical definition:

> Scope is a boundary selector that defines the reach of granted permissions
> within the enclosing tenant. Request material must establish that the
> requested operation remains within the selected boundary.

Permission determines the operation. Scope defines its permitted reach. A grant
binds recipient, permission(s), and scope with validity/conditions. Satisfying
the boundary requirements completes the scope check, not the entire decision.
Scope owns its boundary semantics under SCOPE-003; a grant does not define them.

SCOPE-008: requirements within one scope combine with AND. Alternative authority
is represented through separate complete grants, each with its own scope and
restrictions. DECISION-001's alternative-route semantics and mandatory outer
tenant/delegation limits continue to apply.

SCOPE-007 is the agreed canonical v1 scope format, approved in Q-034:

```json
{ "dept": "dept-1", "user": "$self" }
```

Under the canonical key-value shape, a key identifies a defined boundary and its
value selects that boundary. The above illustrates resources inside dept-1 AND
inside the authorizing human's self boundary. The user key is not the grant's
recipient field. SELF-001 anchors self to the human, including group-derived
and human-proxy access; $self is the canonical reserved token for that human.

The boundary meaning, AND/alternative-grant semantics, flat string-value format,
$self token, empty-object semantics, and basic rejection rules below are settled.
Application key definitions and their governance, scope containment, lifecycle,
and persistence remain open. No arrays, nested objects, wildcards, or arbitrary
query syntax are part of v1.

CONTRACT-006 establishes one endpoint-owned authorization gate, with no prepared
handoff. The current [endpoint authorization chapter](endpoint-authorization.md)
explains input material, evaluation, and enforcement. Endpoint placement is not
part of scope's definition.

## SCOPE-007 — canonical v1 format, agreed

The user approved the presented minimal format and Q-034's explicit empty-scope
choice. The following is canonical v1 scope syntax, without new wrapper fields
or operators. This agreement does not finalize the complete grant schema or
claim the application already implements this format.

### Shape and interpretation

- Scope is a required JSON object mapping defined boundary keys to single,
  non-empty string values. Key order has no semantic effect.
- Each key identifies a scope-owned boundary meaning, not an arbitrary request
  or database field. Applications need not support a universal department key.
- A concrete string selects a concrete boundary reference under that key's
  meaning and the trusted tenant/application context.
- The format reserves the exact token $self for the authorizing human, resolved
  through trusted context and the key's defined relationship. It is not a
  caller-supplied identity or automatically the automated actor. Unsupported
  key/token combinations are invalid; the token does not define arbitrary
  relationship traversals. Literal use of the reserved token is not supported
  in v1.
- All entries apply with AND (already agreed under SCOPE-008). Alternative
  grants supply OR, preserving each grant's complete restrictions.

```json
{
  "dept": "dept-1",
  "user": "$self"
}
```

For the existing motivating meanings, the requested resources must lie both inside dept-1
and inside the authorizing human's personal boundary. The user key selects a
scope boundary; it is not the recipient of the grant. The scope does not grant
an operation by itself.

### Required validation

Reject missing/null/non-object scope, unknown keys, duplicate keys, empty or
non-string values, and unsupported symbolic references. Do not drop an invalid
entry and evaluate the remainder: that could widen the boundary. Validation
must retain every required restriction. Failure cannot establish authority
through that invalid scope; general error handling remains part of resolution.

No arrays, nested objects, OR operators, or wildcard operators are in this
v1 format. The bare * token is rejected, not treated as tenant-wide access.
Do not coerce types or guess key meanings. Concrete-reference validation follows
the defined boundary; endpoint material is not automatically trusted evidence.

### Q-034 — explicit empty object versus absence, agreed

An explicitly supplied empty object means no additional boundary
restriction inside the trusted enclosing tenant:

```json
{}
```

That represents tenant-wide reach for the grant's permissions, subject to its
other restrictions and all mandatory outer bounds. It does not itself supply
permissions or administration authority to issue such a grant. Omitting scope
or supplying null remains invalid; no missing-scope default to {} is permitted.

This avoids adding a tenant selector field when tenant is already implicit.
The alternative considered was rejecting {} and designing another explicit
tenant-wide form. That alternative was not selected in Q-034.
The risk of the agreed shape is that losing entries can widen a scope;
strict validation, authorized grant changes, and audit must not silently turn
malformed or omitted scope into {}. Canonical documentation is not evidence of
implementation support or authority for an administrator to widen a grant.

A canonical wire shape does not by itself settle scope-definition lifecycle,
scope containment for administration, or evaluator/endpoint interfaces. Earlier
type/id and employee_self examples below are retained as historical syntax,
not the canonical v1 scope representation. GRANT-EX-007 in
[grant examples](grant-examples.md) uses the current canonical scope format.

## Scope-key meaning and endpoint fact bindings — Q-038, agreed

The application defines a scope key's boundary meaning and the resource
relationships for which that meaning is supported. The endpoint declares how
to obtain trusted material needed to evaluate that meaning. It does not invent
an independent interpretation merely because a request parameter or database
column has the same name. This agrees the conceptual core of SCOPE-004 and
answers the earlier Q-028; the exact compatibility-validation mechanism is not
thereby adopted.

For example, a grant scope may be:

```json
{ "dept": "dept-1" }
```

Illustrative application definitions could interpret `dept` for a payslip as
the department responsible for its payroll record, and for a certificate as
the department owning that certificate. A repository has no department boundary
unless its application explicitly defines a supported relationship. These
examples are not universal meanings or a newly adopted resource catalog.

An endpoint supplies established resource facts through its declared binding.
A path segment named `dept` does not prove a payslip belongs to that department;
nor may a handler silently use the requesting employee's current department
instead of the defined payroll-record relationship. Unsupported relationships
cannot become unrestricted access by ignoring the scope entry. Missing required
evidence cannot establish authorization through that route.

Layer 1 supplies canonical scope format and evaluation rules. Layer 2 supplies
the application's defined resource relationships and trusted facts. The embedded
auth agent evaluates the boundary using both, under the single endpoint-owned
gate. Application-owned meanings remain subject to canonical rules, including
the human anchor of `$self`; application ownership is not permission to redefine
that token or widen authority.

Q-039 proposes sharing those definitions as one contract used for grant
validation and request evaluation. This remains a proposal: it does not yet
adopt a registry, definition schema, distribution mechanism, or requirement for
Auth to access the application's database. Multi-permission compatibility,
role evolution, reference validation, and administrative scope containment
remain separate open details.

## Historical development and still-open proposals

The earlier discussions below are retained. SCOPE-002 and SCOPE-004 were not
approved as schemas or compatibility mechanisms. SCOPE-005's two-mode query
walkthrough is DEPRECATED and was never agreed. Later agreed boundary language
above is authoritative; the historical wording does not silently finalize fields.

Q-025 authorizes settling the shared scope model before returning to
administrative grants. [The decision log](handbook-roadmap.md) records agreement
status; [the discussion tree](discussion-tree.md) preserves the return to
ADMIN-004/005 and the other unfinished grant branches. This chapter is working
design, not a finalized scope schema or implementation claim.

## Discussion sequence

1. Separate the meaning of a scope from values supplied by a grant; decide how
   scope meanings are defined and recognized.
2. Establish the minimal representation and justify every field, including
   references, self relationships, and implicit tenant context.
3. Define composition: multiple references, dimensions, alternatives, and
   intersections; address missing, empty, invalid, and incompatible scopes.
4. Define resolution and enforcement using trusted facts, including reads,
   collections, creation, updates, moves, and dependent human/proxy context.
5. Test the shared model against ordinary business grants and administrative
   grants. Scope matching and scope containment are distinct problems; do not
   treat containment as implemented by giving it a field name.

These steps do not approve a syntax. PROCESS-005 requires justification and
reuse checks before any field is adopted. Cross-stage dependencies, including
fact freshness and scope changes over time, remain explicit in the tree.

## Agreed foundation

SCOPE-001: scope describes the resource set over which a grant may apply, using
explicit references or trusted attributes/relationships. PERMISSION-001:
permission identifies the operation and resource type; scope determines reach.
TENANT-001: the enclosing trusted tenant remains an implicit enforced boundary.
SELF-001: self resolves for the human being authorized, including group-derived
access; a proxy remains anchored to its authorizing human and delegation limits.

When reading payslip P-17 under a Finance department scope, P-17 is the protected
resource. Finance is a reference used to determine whether that payslip falls
within scope. The request is not thereby an operation on the department itself.
Whether a payslip belongs to Finance requires the appropriate trusted facts;
the presence of FIN in the URI does not establish that relationship.

## SCOPE-002 / Q-026 — historical typed-format proposal, partially deprecated

> The earlier type-plus-parameters representation below is historical after
> SCOPE-007. Its definition-governance questions remain open. Original reasoning
> is preserved; it is not an alternative current canonical scope format.

Q-026's response challenges treating department as a universal type and asks
whether a scope-owned selector model is the better canonical arrangement.
SCOPE-003 / Q-027 below settles that responsibility separation. SCOPE-002 is not yet agreed;
neither a fixed catalog nor a type-plus-parameters representation is finalized.

The recommendation is that a grant selects an explicitly defined scope type and
supplies only the inputs that definition permits. A grant does not invent the
scope's meaning or add arbitrary interpreted fields. Definition, supplied
values, and evaluation result are conceptual distinctions, not a proposal for
three new canonical records or extra fields.

Examples already used in the handbook are sufficient to examine this:

```json
{ "type": "department", "id": "FIN" }
```

In the motivating payroll example, the declared meaning determines which
relationship makes a payslip belong to a department. FIN supplies the chosen
department; it does not define that relationship. The meaning must be explicit
for compatible resource types rather than guessed from a matching field name.

```json
{ "type": "employee_self" }
```

The definition selects payslips belonging to the human's employee identity using
trusted relationships. There is no caller-selected employee ID: the existing
SELF-001 rule supplies the human reference. Exact data contracts remain open.

These reuse earlier illustrative syntax. Neither type/id nor the scope type
names become finalized wire schema merely through repetition here.

| Approach | Benefit | Cost or open concern |
|---|---|---|
| Explicitly defined scope types with validated inputs (recommended) | Makes accepted inputs and meaning reviewable; grants choose values rather than author new rules. | Requires definition governance and an approach for composition/new business relationships. |
| A general scope-expression grammar | Can express many combinations through shared operators. | Requires specifying typing, operators, validation, complexity limits, and safe evaluation before use; the existing examples do not establish that grammar. |

These need not be mutually exclusive forever. The current question is the
grant-authoring contract, not whether implementations may internally share
generic expression machinery. No arbitrary code or database query execution is
proposed in either approach.

The original lab concept page, section 5, also describes typed scope descriptors.
That is useful prior design material, not evidence that its catalog, ownership,
registration, or resolution claims have been reconciled and agreed here.

If SCOPE-002 is accepted, the next discussion must still settle who defines
scope types, their resource/permission compatibility, reusable versus
application-specific definitions, and the minimum parameter representation.
It does not authorize introducing a special scope type for every administrative
combination or assume that existing department/self forms already suffice.

## SCOPE-003 / Q-027 — scope-owned boundary selection, agreed

User agreed to this responsibility boundary in Q-027. This does not settle
SCOPE-002's representation, the standard/application-specific catalog, or storage.

The user proposes that scope owns its definition and grants neither define nor
interpret it; applications need different scope concepts. Department is an
application-specific example, not a mandatory type for every application.
The user also suggests shared types such as self and asks us to reconsider the
canonical arrangement before adding more structure.

In the semantic sense, scope acts as a selector: it describes which resources are
within a grant's reach under the relevant trusted context. This does not mean
arbitrary query syntax, a user-authored database filter, a new selector entity,
or a precomputed enumeration. Selection alone does not authorize an operation;
recipient applicability, permission, validity, conditions, tenant, and dependent
authority limits remain part of the complete authorization decision.

The agreed responsibilities are:

| Concept | Responsibility |
|---|---|
| Permission | Identify the operation on a resource type. |
| Scope | Own the meaning of boundary selection and the valid inputs needed to express it. |
| Grant | Bind a recipient to permissions over that scope, preserving scope use and other restrictions. |
| Request and material | Identify the operation sought and the resource or proposed resource it acts on; representation remains open. |
| Authorization resolution/enforcement | Obtain and validate appropriate facts, interpret scope through its defined semantics, and enforce the complete decision. |

"Grant does not understand scope" means grant semantics must not special-case
department, employee ownership, or future application relationships. It does
not mean accepting opaque unvalidated content: the authorization system must
validate and evaluate scope through its definition. An unknown or incompatible
scope cannot be used as authority. This does not decide a deployment topology,
central service, new catalog, or physical storage layout.

The standard/application-defined distinction is a candidate organization of
scope meanings, not yet a canonical field or a required list of types:

| Candidate kind | Example | Important limitation |
|---|---|---|
| Shared semantic concept | Self | The human anchor is shared under SELF-001, but the resource relationship must be explicitly defined for each compatible resource. |
| Application-defined concept | Department-based reach | Only applications/resources with a declared department relationship can use it; others need not implement it. |

For payslips, self may use a trusted employee-ownership relationship. A document
application could use a defined document-owner relationship instead; that is an
illustrative possibility, not an adopted policy. Self must not automatically
mean creator, assignee, owner, or token subject by guessing a field. Existing
human/group/proxy rules still apply. A reusable shared meaning therefore does
not guarantee a universal resolver or identical data model.

ADMIN-005 remains a useful test: Auth's grants are its domain resources, so the
same scope contract must be tested against selecting proposed grants. No
grant_selector or other withdrawn field is reinstated by that observation.

Canonical semantic ownership is not the same as requiring every scope value to
be a separately stored mutable object. Whether a grant embeds a scope value or
references one, and how definition/value edits affect existing grants, are
separate open decisions. Avoid importing live-update or revocation behavior
from the role model without agreement.

The recommendation is to settle this separation first, then compare declared
scope definitions with shared selector building blocks using concrete cases.
Do not finalize either a closed universal type list or a free-form expression
language yet. Compatibility, definition authority, registration/naming,
composition, boundary evaluation, and scope containment remain open.

## SCOPE-004 / Q-028 — explicit resource compatibility, proposed

Given SCOPE-003, the next question is when a scope's selection meaning is valid
for a particular resource. The recommendation is explicit compatibility: a
scope may be used only where its resource relationship has been defined, never
guessed from a familiar name or a coincidentally matching field.

For an illustrative Finance department scope:

| Resource | Definition needed before department scope can select it |
|---|---|
| Payslip | The declared relationship establishing which department the payslip belongs to. |
| Certificate | The declared relationship establishing which department the certificate belongs to. |
| Repository | An explicit repository-to-department relationship; without one this scope is unsupported. |

If a repository creator happens to work in Finance, that alone does not prove
the repository belongs to Finance. An application could deliberately define a
creator-based relationship, but the evaluator must not invent that rule. The
same caution applies to interpreting self as owner, creator, or assignee.

This permits reuse: one scope meaning may support several resource types when
their mappings are explicitly defined. It does not require creating a new scope
type for every resource or forcing every application to implement department.
No compatible_resources field, registry format, or endpoint syntax is proposed.
Where compatibility is declared remains to be decided.

Compatibility does not confer a permission. A valid department scope paired
with payslip-read still does not authorize payslip-edit. Conversely, having the
read permission does not make an unsupported scope meaningful. Validate the
pairing when constructing a grant; resolution cannot treat an unsupported or
unresolvable pairing as established authority. Exact validation interfaces and
error responses remain open.

GRANT-002's multi-permission binding must retain a meaningful scope for the
included permissions. How that is validated across resources, and how later
ROLE-002 role changes interact with compatibility, are explicit follow-ups;
neither silently broadening scope nor silently inventing a mapping is justified.

## SCOPE-005 / Q-029 — selectors and the endpoint flow, deprecated proposal

> Deprecated by CONTRACT-006 / Q-033. Retained for history; middleware-complete,
> endpoint-completion, and prepared below are no longer the current model.

The user requested explanation through GET /api/v1/tenant/{dept}/{cert}, with
some material available to middleware and other facts available only at the
endpoint. The user also proposes that scope acts as a selector and may become
query-like. Q-028 is still open: do not treat this feedback as approval of a
compatibility registry or a new field. The following example connects that
proposal to the already agreed CONTRACT-002/005 endpoint modes.

### Grant and requested path

Assume verified human Vinay, trusted enclosing tenant T-1, and this sole
applicable grant, with validity and other required constraints satisfied. This
abbreviated grant reuses existing illustrative binding/scope syntax. It is not a
new schema; status, validity, and conditions are omitted only for focus.

```json
{
  "id": "G-13",
  "recipient": { "type": "user", "id": "vinay" },
  "permissions": ["hrms:employee:certificate::read"],
  "scope": { "type": "department", "id": "FIN" }
}
```

For the user's route template, tenant is the trusted enclosing context in this
illustration; if a route also carries a tenant identifier, bind and validate it
against that context. The actual request here is:

```http
GET /api/v1/tenant/FIN/C-17
```

The following JSON only displays extracted route parameters, not a proposed
authorization-request record or new canonical fields:

```json
{ "dept": "FIN", "cert": "C-17" }
```

Middleware knows identity, tenant, operation mapping, grant scope, and requested
department/certificate. FIN is a requested selection. It does not prove the
stored certificate C-17 has a Finance department relationship. The proposed
selector meaning for this example is certificate department equals FIN; that
meaning and its trusted data mapping are application-defined, not inferred by
matching the URI variable name to a column name.

If the requested department were ENG, the sole FIN grant cannot authorize this
department-qualified request. Middleware may deny conclusively without querying
the application database. With other applicable grants, all valid alternative
routes must be considered before a conclusive denial.

### Middleware-complete contract: apply a fully determined restriction

For a contract that can establish the complete allowed restriction in
middleware, the handler enforces it in the resource lookup. Under the explicit
example mapping certificate department to department_id, a parameterized query
could look like this; table/column names are application data, not scope fields:

```sql
SELECT * FROM certificates
WHERE tenant_id = :trusted_tenant
  AND department_id = :authorized_department
  AND id = :requested_certificate;
```

Here trusted_tenant is T-1, authorized_department is FIN established from the
grant and permitted request selection, and requested_certificate is C-17. Do
not substitute an unchecked caller department for authorized_department. The
complete request includes the department-qualified resource; if request and
authority differ, do not silently drop the requested department constraint.

The handler is not receiving unrestricted authorization for C-17. It is allowed
to retrieve C-17 only inside the established tenant and Finance selection. If
C-17 is actually in Engineering, this lookup returns no matching row and no
certificate data is disclosed. Missing/out-of-scope response semantics remain
an endpoint contract detail; the query does not establish global nonexistence.

This is enforcement of fully determined restrictions, not a second
authorization-completion step. It therefore can fit middleware-complete mode
under CONTRACT-005, if the endpoint and all its supported scopes/delegations
satisfy that fixed contract. An ID-only lookup after checking FIN in the path
would not enforce the restriction.

### Endpoint-completion contract: obtain facts to finish authorization

For a separately declared endpoint-completion contract, middleware supplies
prepared context unless it can deny. The handler obtains the bounded internal
facts and invokes the evaluator to complete authorization before protected
output. It never receives middleware allow, even if one request is easy.

For example, the handler may establish the actual department of the requested
certificate from application metadata and submit it for final evaluation. A
different supported scope might require an application-only human-to-employee
relationship before the authorized selection can be resolved. The fact-access
path must remain appropriately tenant/request constrained; preparation does
not authorize returning the certificate or an unbounded database read.

These are alternative design-time contracts, not runtime modes that middleware
chooses per caller. A database lookup alone is not a reason for completion:
applying an already complete restriction is enforcement; obtaining facts to
finish deciding the restriction or access is completion.

### Query-like meaning without premature query syntax

SCOPE-005 proposes treating scope as declarative boundary selection that the
application may translate into a safe query restriction where supported. The
requested selection narrows, never replaces, the authorized selection:

```text
Returned resources = requested resources intersect authorized resources
```

The complete authorization still checks permission, recipient applicability,
tenant, validity, conditions, and delegation bounds. A query predicate alone is
not the entire decision. For a single-item request the intersection can contain
at most that item; a list contract may enforce the same authorized restriction
across its result set. Mutation/create/aggregate semantics remain separate work.

Stored scope, an interpreted selector, and an executable query are different
representations, not three approved canonical entities. A query-capable selector
does not require storing arbitrary SQL or executable code in a grant, nor does
it show that every scope can be compiled to one query. Exact grammar, mapping
ownership, supported operations, and the evaluator/handler interface remain open.

Q-028's substantive concern is that the declared scope meaning and the query
mapping agree: a department selector for a certificate must not silently test
the creator's department instead. An auth-aware endpoint/application mapping
can establish this; a separate compatibility registry has not been adopted.

## Return-point tests, not adopted syntax

- Read a particular certificate or payslip within an explicit selection.
- Read Finance resources versus the requesting human's own resources.
- Keep scope semantics intact across group-derived and human-proxy access.
- Create a grant for an eligible recipient with permitted capabilities/reach,
  without introducing an unreviewed second authority format.
- Establish when a proposed grant's scope is contained within assignable reach;
  testing one resource's membership does not prove general scope containment.

Return to ADMIN-004/005 after these scope questions are settled. The withdrawn
grant_selector, permissions_subset_of, and scope_within fields remain withdrawn.

## Q-156 / SCOPE-010 — a platform scope key, and values that name Auth's own records

Status: **AGREED.** [Q-155](administrative-authority.md) makes administrative
authority an ordinary grant scoped to the team it administers. That needs a scope
key the platform owns, and it needs one thing Q-148 did not allow.

### The key

`team` is registered at the **platform boundary**, and its value is a team
identifier:

```json
{ "permissions": ["auth:group::write"], "scope": { "team": "fibggi2juubk" } }
```

Until now every scope key was application-registered — `RegisterScope` takes an
application, and the catalog reader loads only application-boundary scopes. A
platform key is new machinery, and it is the only new machinery
[Q-155](administrative-authority.md) requires.

There is no per-team registration. The **key** is registered once; the **value** is
the team's id, like any other scope value.

### A value naming an Auth record is validated, not opaque

[Q-148](application-registration.md) settled that scope values are opaque to
authorization: nothing resolves one to a record, because resolving would require
application facts this model does not hold.

**A team is not an application fact.** It is Auth's own record, in Auth's own
store. So for a platform scope key whose value names an Auth record, the value
**is** validated — a grant scoped to a team that does not exist is refused.

This refines Q-148 rather than contradicting it. Q-148's reason was that Auth
cannot know whether a department exists; it can always know whether a team does.
The rule stated exactly:

- **Application-boundary keys** — values stay opaque. `dept=FIN` is not resolved.
- **Platform-boundary keys naming an Auth record** — values are validated to exist.

### What does not change

Matching is still **exact**, and subtree scope is still excluded. Administering a
hierarchy is not expressed by a pattern: `{team: X}` administers X. Reaching more
teams means holding more grants, or holding one with `{}` from the Auth root.

That is not a limitation worked around — it is what makes sideways escalation
unsatisfiable under [Q-155](administrative-authority.md), because two values of one
key can never both be satisfied.

### Rationale / conscious tradeoff

Validating the value costs a lookup on a write path that already reads the
snapshot, so the cost is close to nothing and it buys the invariant that an
administrative grant always names a real team.

The conscious cost is the asymmetry: two scope keys now behave differently, and a
reader has to know which boundary a key belongs to before knowing whether its value
means anything. That is accepted because the alternative — validating application
values — is the resolver interface this handbook has declined three times.

---

## A key claimed at two boundaries is refused — Q-156 / SCOPE-011

An application's catalog is its own scope keys union every platform key, and a
scope key is a bare word: unlike a permission id it carries no namespace, so one
catalog holds one entry per key. A key declared by both an application and the
platform is therefore genuinely ambiguous, and the catalog read **refuses** rather
than choosing.

It chose before, by the SQL ordering — namespace then key — so an application whose
id sorted before the platform's namespace had its own declaration overwritten and
its own opaque values validated as Auth record ids, while an application whose id
sorted later kept working. The same records, two answers, decided by a string
comparison nobody wrote down.

`CheckScopeRegistration` already refuses a registration that collides with a
platform key, so the reachable state is a key an application registered *before*
Auth owned that key. That is a migration, and the right behaviour for a migration
is a refusal naming the key rather than a silent reinterpretation of every value
under it.

---

## Q-158 / SCOPE-012 — a scope value is a list; a child narrows by subset

Status: **AGREED.** Raised while proving the sample applications, after a child
grant scoped `dept=ENG` beneath a parent scoped `dept=FIN` was written and then
never matched. The user asked whether scope narrowing could be enforced without
understanding the scope — *"the parent may have three of them, finance,
engineering, production, and if you make it for a child, it has to be one of
them"* — then chose to allow lists now, inside v1, *"because nothing is there"*,
and gave the go-ahead to change the handbook, the Auth service and the sample
applications. The groundwork and safety analysis were reviewed before any change.

### The rule

**A scope value is a non-empty list of strings.** A record satisfies a key when
its value is **a member** of that key's list. Keys still combine with AND. Values
stay opaque and are compared exactly; Auth never interprets one.

```json
{ "dept": ["ENG", "FIN"], "user": ["$self"] }
```

reads: the record's department is ENG or FIN, **and** the record lies in the
authorizing human's own boundary.

**A child narrows by subset.** For every key the child states, its list must be a
subset of the parent's effective list for that key. A key the parent's chain does
not constrain may be added freely; a key the child omits is inherited unchanged.
A child that is not a subset is **refused at write**. Nothing is trimmed, and lists
are never combined by union.

| Parent effective scope | Child scope | Result |
|---|---|---|
| `{dept: [FIN, PROD]}` | `{dept: [FIN]}` | **allowed** — a subset |
| `{dept: [FIN, PROD]}` | `{}` | **allowed** — inherits `[FIN, PROD]` |
| `{dept: [FIN]}` | `{dept: [FIN], loc: [BLR]}` | **allowed** — a new key narrows further |
| `{dept: [FIN]}` | `{dept: [ENG]}` | **refused** — ENG is not within FIN |
| `{dept: [FIN]}` | `{dept: [ENG, FIN]}` | **refused** — not trimmed to FIN |
| any | `{dept: []}` | **refused** — malformed, never "unrestricted" |

The effective scope of a route is the per-key **intersection** of every link's
list. Because each child is a subset, the intersection is the child's list; it is
still computed rather than assumed, so a stored state that violates the rule can
only narrow, never widen, and an empty intersection authorizes nothing.

### Canonical form and validation

- A list is **non-empty, strictly ascending, and free of duplicates**. A
  non-canonical list is **rejected**, not normalised — normalising on one write
  path and comparing bytes on another would let the two disagree.
- `[]`, `null`, a bare string, or any non-string element is **malformed**. An empty
  list is never read as an absent key and never as `{}`. This extends Q-034: losing
  entries can widen a scope, and an empty list is the new way to lose one.
- One invalid element invalidates the whole scope; it is never dropped while the
  rest is kept (SCOPE-007's existing rule, now per element).
- Every rule that applied to a value applies to each element: the bare `*` is
  rejected, `$self` is the only reserved token, and a `team` element must name an
  existing team (Q-156, above).
- `{}` keeps its meaning under Q-034: no additional restriction, including values
  that do not yet exist. A list is a closed enumeration; it never means "all".

### `$self` is compared as a token at write

At write time `$self` is an opaque element like any other: `[$self]` beneath
`[maya-id]` is refused even though it would match Maya. The check can refuse a
grant that would have been satisfiable; it can never widen one. Identical text
with a different human binding is not a containment proof (C01-T28). At request
time `$self` is bound to the requesting human, element by element.

### Requests stay single-valued

Lists exist only in **grant** scope. Request material still carries one value per
key ([Q-150](operation-enforcement.md)), so a list never turns one request into a
question about several departments. A handler still queries with the value it was
asked about and never with a grant's list — [Q-072](collection-enforcement.md)'s
"does not inspect grants" is what keeps lists out of query translation.

### What it settles and what it amends

- **Settles** the creation-time rejection left open by [Q-091](subgroups.md) ("exact
  creation-time rejection rules remain open") and by
  [authority-boundary-validation](authority-boundary-validation.md) ("whether every
  unsatisfiable definition must be rejected at publication is not settled"): a
  child that is not a subset of its parent is refused at write.
- **Retires** the wart recorded under [Q-155](administrative-authority.md): a
  contradictory re-scope is no longer written and left inert; it is refused.
  Sideways escalation stays impossible for the same reason as before — a request
  value must be a member of every list on the chain — and is now also refused at
  the write.
- **Amends** SCOPE-007 (single string values → lists), SCOPE-008 and PC-04
  (alternatives *within one key* may be a list; alternatives across keys or
  permissions still use separate complete grants), PC-05 (narrowing is per-key
  subset), and the "no arrays" statements in this chapter and its summaries. Earlier
  wording is preserved below and in the summaries as history.
- **Keeps** exact matching, the exclusion of subtree and wildcard scope
  ([Q-148](application-registration.md)), `{}` semantics (Q-034), the contract
  version `"1"` (no version had been published to a consumer; renaming before
  publication follows [Q-135](decision-results.md)'s precedent), and the
  single-valued request.

### Consequences settled by existing rules

| Situation | Outcome | Settled by |
|---|---|---|
| A parent revision narrows its list so an enabled child is no longer a subset | Adoption of that revision is **refused** | [Q-105](grant-revisions.md) adoption re-validates dependants; [B22](parent-grant-bindings.md) "validate the complete result" |
| A stored chain that is not a subset (legacy data, a direct store write) | The route is **closed**, not widened | [Q-140](authority-lineage.md); the effective intersection is empty or narrower |
| Delete team X while a grant names `{team: [X, Y]}` | **Refused** until the grant is revised | [Q-157](ownership-lineage.md): no write passes through an unreachable state; Q-156: no grant names a team that does not exist |
| The refusal's code | `SCOPE_NOT_WITHIN_PARENT`, given only to a caller who holds the parent | [Q-135](decision-results.md): the catalogue is open, a code explains and does not decide |

### Rationale / conscious tradeoff

The format change costs one rule — set inclusion on strings — and needs no
meaning: whether `ENG` lies within `[FIN, PROD]` is a string comparison, so the
subset check works for every key an application will ever register. It expresses
"one of these departments" in one grant instead of one grant per department, and it
turns an inert, misleading grant into a refusal with a reason.

The alternative was to keep single values and add a write-time contradiction check.
It was rejected because the user's motivating case — a parent holding several
departments and a child choosing among them — would still need one grant per
department, and the check would exist anyway.

The conscious costs: every consumer's matcher changes (membership instead of
equality), every stored scope is migrated to one-element lists, and JSON examples
written before this decision show the earlier string form. Those examples are
preserved as written; read `"dept": "FIN"` there as `"dept": ["FIN"]`.
