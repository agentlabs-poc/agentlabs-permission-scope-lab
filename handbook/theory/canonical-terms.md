# Foundations reference — canonical concepts and their representations

[Contents](../README.md) · [Foundations narrative](01-foundations.md) · [Implementation](../implementation/05-canonical-model.md)

Foundations must do more than introduce authorization: a reader should be able
to explain every core term, recognize its representation, and distinguish it
from neighboring concepts. This reference belongs to Foundations. Read it beside
Chapters 1–4 or use its index when a term first appears.

General concepts come first; statements introduced as **our model** describe
this handbook's chosen framework. The JSON documents that model, not a universal
authorization standard. A different implementation can preserve a distinction
without using the same records or field names.

## How to read the representations

There are four representation categories:

- **Approved core JSON:** agreed fields and meanings, but not necessarily a
  complete deployable schema, API, or validation specification.
- **Nested value:** a field or value inside an approved contract, such as scope
  or recipient. It does not need an invented standalone record or version field.
- **Concept, not a separate record:** a distinction or activity represented by
  existing records, their relationships, or evaluation behavior.
- **Format pending:** the meaning is explained, but its full wire representation
  has not been approved. A diagram or explanatory table is not a proposed schema.

Every standalone framework-contract example below includes `"version": "1"`.
This identifies its format, not proof that the example is trusted, complete, or
authorized. JSON fragments mentioned within prose remain nested values.

## Term index

Each row links to an explanation, representation status, rationale and example.
Related words are grouped for comparison; grouping does not make them synonyms.

| Concept family | Terms covered | Representation |
|---|---|---|
| [Principles and rules](#principles-and-rules) | Authentication, authorization, principle, policy, authority, effective authority | Concepts; concrete authority uses the records below. |
| [Identity](#identity) | Identity, principal, actor/caller, human/user, authorizing human, subject, agent, service account, proxy | Approved identity block and JWT profile excerpt. |
| [Enclosing context](#enclosing-context) | Tenant, tenant context, tenant membership, application, platform administration | Context/responsibility concepts; complete mappings pending. |
| [Permission](#permission) | Operation, action, permission, permission namespace | Registered string, used in grant and endpoint JSON. |
| [Scope and self](#scope-and-self) | Scope, boundary, selector, scope key/value, `$self`, additional/effective scope | Nested flat object, exact match, values opaque except a platform key naming Auth's own records. |
| [Restrictions and conditions](#restrictions-and-conditions) | Constraint, restriction, condition | Existing limits and rules; no generic condition JSON is adopted. |
| [Registration](#registration) | Catalog, permission definition, scope definition, compatibility declaration | Meanings agreed; registration payload pending. |
| [Grant records](#grant-records) | Grant, grant identity/control, grant revision, recipient, assignment, role, role revision | Approved control, revision, assignment and role-reference shapes. |
| [Version and adoption](#version-and-adoption) | Contract version, revision, publication, adoption, latest revision | Existing version/revision fields; activities are not new entities. |
| [Teams and administration](#teams-and-administration) | Team/group, membership, ownership, owner/administrator, issuer, assignment authority | Administration is an ordinary grant; ownership is a grant and its relation is superseded; membership records pending. |
| [Dependent relationships](#dependent-relationships) | Parent/child, subteam/subgroup, subgrant, grant/team/scope lineage, support, authority route, binding | Grant parent link and assignments; full team hierarchy format pending. |
| [Delegation](#delegation) | Delegation, human ceiling, delegation limits | Not a grant and not a chain step; carries its own window and tracks the human. Evidence format pending. |
| [Lifecycle](#lifecycle) | Enablement, disablement, effectiveness, validity, expiry, deletion/revocation, orphan | Live controls and revision-local validity; no invented orphan state field. |
| [Roots and bootstrap](#roots-and-bootstrap) | Root grant, bootstrap, computed root coverage, the tenant's two namespaces | Trusted-root behavior and establishment authority agreed; computed-root encoding pending. |
| [Endpoint declaration](#endpoint-declaration) | Endpoint policy, method, path, required permission, input, source, local input name, trusted correlation | Approved GET and PUT policy core JSON, plus its mount-time structural rules. |
| [Request and material](#request-and-material) | Request, input, material, domain/application fact, relationship, resolved request | Policy/identity examples available; full request/resolved envelopes pending. |
| [Resolution and evaluation](#resolution-and-evaluation) | Resolution, resolved grant, resolved grants, evaluation, complete route, non-amplification | Computed views, plus the approved authority-loading question and answer. |
| [Decision and enforcement](#decision-and-enforcement) | Allow, deny, evaluation error, result, reason, supporting evidence, enforcement | Approved result variants, the ordered contributing chain, and the published code catalogue. |
| [Responsibility layers](#responsibility-layers) | Auth Service, auth agent/evaluator, canonical layer, application layer, authority-boundary validator | Logical responsibilities, plus the handler integration contract. |

## Principles and rules

**Authentication** establishes an identity under a trust arrangement.
**Authorization** establishes whether an operation is permitted and ensures
that its protected effects remain within the applicable authority. Recognizing
Vinay does not establish that he can read Nutan's payslip.

A **principle** is a design constraint, such as not deriving authority from
missing evidence. A **policy** expresses applicable authorization rules. A
policy is not necessarily a file; an endpoint declaration is one concrete policy
record in our model, while parent-subset rules also govern the system.

**Authority** means permitted operations together with their reach and required
support. **Effective authority** means what the currently valid routes actually
provide after restrictions and dependencies are applied. It is not the list of
all permission strings mentioned in stored records.

These are concepts, not additional canonical JSON objects. Grant, assignment,
identity and endpoint records make particular aspects concrete. The reason to
keep the concepts is to explain why a well-formed, enabled record can still
supply no usable authority: its support may be absent, its validity may have
ended, or the requested operation may lie outside its scope.

**Counterexample:** creating an `authority` field containing every permission a
user has ever held would neither establish current entitlement nor replace
evaluation. No such field is introduced here.

Sources: [principles](../../docs/principle-catalog.md),
[vocabulary](../../docs/authorization-vocabulary.md).

## Identity

**Identity** describes who is acting and, in our dependent-proxy model, whose
human authority supports that action. **Actor** and **caller** mean the actual
requesting person or program. **Human** and **user** refer to a human identity;
the approved actor type for a human is `user`.

The **authorizing human**, or **human anchor**, bounds proxy authority and
anchors `$self`. **Principal** is useful general identity vocabulary, not an
extra entity or wire field. Use the precise role—actor, human anchor or
assignment recipient—when describing a contract.

Approved core identity JSON for Vinay acting directly:

```json
{
  "version": "1",
  "actor": {"type": "user", "id": "U-17"},
  "human_id": "U-17"
}
```

Approved core shape for an **agent** acting for Vinay:

```json
{
  "version": "1",
  "actor": {"type": "agent", "id": "A-17"},
  "human_id": "U-17"
}
```

The same approved shape instantiated for a **service account** acting for Vinay:

```json
{
  "version": "1",
  "actor": {"type": "service_account", "id": "SA-17"},
  "human_id": "U-17"
}
```

In our model, both automated actor types are dependent
**proxies**, not independent human substitutes or first-class team members.
Other authorization systems may deliberately choose independent machine
principals; that is not the policy adopted here.

| Field | Meaning | Why it is separate |
|---|---|---|
| `version` | Identity-contract format. | Identity can be reused outside a JWT. |
| `actor.type` | `user`, `agent` or `service_account`. | Preserves the actual kind of caller. |
| `actor.id` | Actual caller identifier. | An agent's ID is not its human's ID. |
| `human_id` | Human anchoring authority and self. | Preserves the source even when caller and human differ. |

Direct-human actor ID and human ID must agree. For proxies, trusted evidence
must establish the association and the applicable delegation. Submitted JSON
alone proves neither identity nor delegation.

**Subject** has a specific mapping in our JWT profile: `sub` identifies the
human and equals `identity.human_id`. Approved identity-related payload excerpt:

```json
{
  "version": "1",
  "sub": "U-17",
  "identity": {
    "version": "1",
    "actor": {"type": "agent", "id": "A-17"},
    "human_id": "U-17"
  }
}
```

This is not a complete JWT or token-verification procedure. Cryptographic
verification, issuer, audience and expiry concerns are omitted for focus,
not made optional. The profile deliberately repeats the human ID so the nested
identity block remains self-contained. It adds no duplicate `agent_id` claim.

**Counterexample:** an old consumer accepting Vinay's subject while ignoring
A-17's narrower delegation can amplify authority. Preserving `sub` is not proof
of universal safe backward compatibility. Full trust adapters remain pending.

Sources: [identity glossary](../../docs/identity-glossary.md),
[identity block](../../docs/identity-context.md), [JWT](../../docs/jwt-identity-mapping.md).

## Enclosing context

A **tenant** is the enclosing isolation context for ordinary tenant authority.
**Tenant context** must be established through trusted bindings. A path input
named `tenant` identifies what the caller requested; it does not establish that
the caller belongs to that context or may administer it.

In our grant examples, tenant is implied, not an ordinary `scope` key. This
keeps a caller from treating an enclosing boundary as a removable child
restriction. The complete shared tenant-context representation is pending;
no `tenant_id` field is invented in the canonical records below.

**Tenant membership** and authorization-group membership are not interchangeable.
Association with a tenant does not by itself confer group or administrator
authority. Its complete identity mapping is also pending.

An **application** owns the meanings of its registered operations and boundary
keys. **Application platform administration** governs capability publication
outside tenant business scope, within its own management boundary. **Tenant
application administration** manages the tenant installation; business access
requires its own valid authority. These are responsibilities, not new actor types.

**Counterexample:** the administrator who publishes HRMS's supported permissions
does not thereby acquire the ability to read every tenant's payroll data.

Sources: [identity glossary](../../docs/identity-glossary.md),
[platform authority](../../docs/application-platform-authority.md).

## Permission

An **operation** or **action** is the protected work being requested. A
**permission** is the registered name of that operation in our model. It says
what kind of work authority covers, not who receives it or which department it
covers. Endpoint `method` maps to a declared permission; the HTTP method alone
is not the permission.

The canonical permission namespace is:

```text
<app>:<domain>[:<subdomain-or-resource>...]::<verb>

hrms:employee:certificate::read
hrms:employee:certificate::write
accounting:ledger:entry::post
```

The string is a **nested value**, visible in the full grant and endpoint JSON
below. A grant's `permissions` array may contain multiple strings; an endpoint's
`permission` is exactly one string in our chosen model. There is no need for a
new `action` object or separate permission-value wrapper.

The colon-separated nouns identify the operation namespace; the double colon
separates the verb. Namespace depth can vary. Department IDs and people do not
belong in the operation name. No prefix inheritance, permission aliases or
wildcards are supported in this model. The noun path *reads* like a hierarchy and
is not one: evaluation compares whole identifiers, and `hrms:payroll:*` is not a
thing the model recognizes.

**The first noun segment is the application, and it is load-bearing.** An
application registers only in its own namespace: `hrms` may register
`hrms:payroll:payslip::read` and may not register anything under `auth:` or
`system:`. The reason is not tidiness. A root grant's ceiling is computed as every
active permission *in its own namespace*, while an application's evaluation
catalog is wider — its own permissions union the platform's, because a request
inside an application may legitimately require a platform permission. The ceiling
is sliced where the catalog is not, and without that slice an application root
would carry every platform permission, including whichever one authorizes
establishing an application root. The thing created by an authority could then
create more of that authority.

**An identifier is never renamed.** Reusing a retired identifier for a different
authorization meaning is not permitted, and the identifier itself is immutable;
a description or display label is not. A rename is strictly worse than a
repurpose, because the old identifier stops resolving and every grant referencing
it narrows silently — a tenant's grant quietly supplies less and nothing says so.
If a name is wrong, register the right identifier and retire the wrong one.

**Retirement withdraws the permission, not the route.** A grant supplies what it
selects and the catalog still supplies; it stops only when nothing it selects
survives. A grant selecting payslip-read and payslip-write whose write is retired
still supplies read, and a route beneath it that never selected write is
untouched. The write path is unchanged: nothing may be authored, revised, assigned
**or re-enabled** while it selects a permission the catalog does not supply — so
**disabling a narrowed grant is one-way** until the permission is restored or the
grant is revised. The read path keeps such a grant working; the write path will not
take it back.

**Counterexample:** certificate-read does not confer certificate-write, even
though the names share a prefix. A Finance scope cannot supply the missing verb.

Sources: [permission model](../../docs/permission-model.md),
[namespace and permanence](../../docs/permission-lifecycle.md).

## Scope and self

**Scope** is a **boundary selector**: it describes the reach of authority.
The **boundary** is the resulting permitted region, not another record between
scope and the operation. **Selector** here does not mean an arbitrary query
language. A scope key names a registered boundary dimension, and its value
selects a boundary under the application's defined meaning.

Our canonical scope is a required flat object of registered keys and non-empty
string values. Within one scope, entries combine with AND. For example, the
nested value `{"dept":"FIN","cert":"C17"}` means Finance AND certificate C17,
under the application's supported meanings. The complete grant JSON below shows
where this value lives; it is not a standalone versioned scope record.

`$self` selects the **authorizing human**, not the recipient group, its owner,
the grant's creator or the automated actor. A grant with nested scope
`{"user":"$self"}` assigned to Employees applies self relative to each requesting
human. Vinay's agent still uses Vinay as self, subject to valid delegation.
Group assignment therefore supports personal self-access without separate
copies of the grant for each employee.

**Additional scope** is a child's locally declared restriction. **Effective
scope** preserves inherited restrictions AND that additional scope. It is a
computed boundary, not an instruction to overwrite the parent's object.

| Scope situation | Meaning |
|---|---|
| Parent `dept = FIN`; child `cert = C17` | Finance AND C17. |
| Parent `dept = FIN`; child `{}` | Finance, with no additional local restriction. |
| Parent `dept = FIN`; child `dept = ENG` | Both restrictions remain; ENG does not replace FIN. |
| Legitimate tenant root with `{}` | No narrower local boundary inside that tenant; other requirements remain. |

Flat scope syntax rejects arrays, nested objects, duplicate keys, empty or
non-string values, unsupported keys/tokens and wildcard/query operators.
Missing or null scope is not the explicit empty object. OR alternatives use
separate complete grant routes, not a scope array or permission/scope mixing.

**A value matches exactly, and it is opaque.** Nothing in this model resolves a
scope value to a record, checks that the record exists, or asks what it refers to.
`dept=FIN` authorizes within a boundary named `FIN`; whether such a department
exists, and whether certificate C17 belongs to it, is the application's to
establish. Subtree and pattern scope are **excluded rather than deferred**: there
is no prefix, wildcard or hierarchy, and a hierarchical scope would give every
existing boundary implied children on the day it was introduced, silently
changing the reach of every stored grant.

| Written | Means |
|---|---|
| `{"dept": "FIN"}` | exactly the boundary named `FIN` |
| `{}` | no local restriction — the whole application boundary |
| `{"dept": "FIN*"}` | **refused.** Not a boundary. |
| `{"dept": "$self"}` | the authorizing human, the one reserved token |

`{}` is not a wildcard by another name: it adds no restriction, which is different
from matching many values. And a *key* never declares either form — `{}` and
`$self` are properties of a value.

**One exception, and it is the platform's own key.** A scope key registered at the
platform boundary names something in Auth's own records, and its value therefore
*is* resolved: the key `team` carries a team identifier, and a grant naming a team
that does not exist is refused at the write. Opacity holds for an application's
keys and does not hold for a key naming Auth's own records — Auth cannot know what
`dept=FIN` denotes, and does know whether a team exists. Matching stays exact
either way; the subtree exclusion above is unaffected.

A scope key is a bare word, so one catalog holds one entry per key and an
application's catalog is its own keys union every platform key. **A key claimed at
both boundaries is ambiguous and the read refuses it**, naming the key. There is
no correct winner to choose, and choosing by storage order would mean an
application's opaque values being validated as Auth record identifiers for some
application names and not others.

**Counterexample:** recognizing `dept = FIN` in a request is not proof that the
certificate returned belongs to Finance. The endpoint must establish or enforce
that relationship against the actual operation's data.

Runtime self meaning is settled. Whether one person's self-scoped source
authorizes distributing another person's self access remains a distinct
source-binding question; copying `$self` text does not answer it.

Sources: [scope model](../../docs/scope-model.md),
[opacity and exclusion](../../docs/application-registration.md).

## Restrictions and conditions

A **constraint** or **restriction** is a limit that must remain satisfied for
authority to apply. Scope predicates, revision validity, current support and
delegation limits are examples, but they do not all belong inside `scope`.
Keeping them distinct avoids making a boundary selector carry every policy rule.

Earlier discussion also uses **condition** for additional authorization
restrictions. That vocabulary does not approve an arbitrary `conditions` object,
expression language or application workflow engine. Scope and validity use the
approved representations shown here; remaining authorization-only restriction
treatment is still pending.

**Counterexample:** a business rule saying that a payslip calculation must
balance is not a new grant condition merely because it can prevent an operation.
The handbook's authorization boundary is not a generic business-rule system.

Sources: [authorization-condition boundary](../../docs/grant-conditions.md),
[contract gaps](../../docs/grant-contract-closure.md).

## Registration

A **permission definition** gives a registered operation its application-owned
meaning. A **scope definition** gives a key its boundary meaning and supported
constraints. A **catalog** is the application's registered capabilities, not
a grant, group or list of entitlements possessed by every user.

**Registration** makes these definitions available for validation before grants
use them. Auth validates canonical syntax, registration and declared constraints;
the application interprets its domain meanings. Valid syntax alone neither
authorizes issuance nor proves facts about application data.

A **compatibility declaration** specifies supported permission/scope
combinations when that feature is enabled. Our model requires an upfront
application-level enabled/disabled choice. Enabled means all grants must conform;
it is not a per-grant escape hatch. Enabling it over existing incompatible grants
must fail without silently changing or grandfathering those grants.

**Representation: format pending.** Registration meanings are approved, but a
complete registration payload is not. The permission and scope values shown in
this reference are consumers of registration, not a disguised registration API.

An application registers **only within its own namespace**, and the platform
registers within the platform's. The two boundaries are separate operations with
separate authority, and a key or permission the platform owns cannot be claimed by
an application — see [Permission](#permission) for why the first noun segment is a
boundary rule rather than a convention.

**Counterexample:** registering both `dept` and repository-read does not alone
prove that a department selector is a supported repository boundary. Nor does
registering read automatically assign read to the publisher.

Sources: [application registration](../../docs/application-registration.md),
[namespace ownership](../../docs/permission-lifecycle.md).

## Grant records

A **grant** is a reusable authority definition: a permission source, scope and
applicable restrictions and dependencies. A **recipient** is the human or group
to whom an **assignment** binds a grant. Separating them lets one definition be
reused without pretending that creating a definition gives someone access.

Our model distinguishes live grant control, immutable content and assignment.
The following is one consistent running example. Assume registered certificate
permissions and scope keys, existing Team1, and successful authorization/boundary
checks. G0 is the trusted root, held by Team0 through assignment A0, and Team1 is
Team0's child — the example needs a named root holder because a resolved answer
carries the root step, and a chain that begins mid-way cannot be shown honestly. G1 revision 2 is latest when A1 is
created or explicitly upgraded. These are required premises, not bypasses.

### Grant identity and control

Approved core JSON:

```json
{
  "version": "1",
  "id": "G1",
  "status": "enabled"
}
```

`id` identifies the reusable grant. `status` is its live grant-wide switch.
`version` identifies the format. This block has no permission content: it cannot
authorize a request by itself. The global switch applies across G1's revisions
and assignments, so one recipient cannot override a global disable.

### Grant revision

A **grant revision** is immutable published authority content:

```json
{
  "version": "1",
  "grant_id": "G1",
  "revision": 2,
  "parent_grant_id": "G0",
  "permissions": [
    "hrms:employee:certificate::read",
    "hrms:employee:certificate::write"
  ],
  "scope": {"dept": "FIN"}
}
```

| Field | Meaning and rationale |
|---|---|
| `grant_id` | Connects the content to the grant's identity/control. |
| `revision` | Identifies this immutable publication; `(grant_id, revision)` identifies content. |
| `parent_grant_id` | Declares required parent grant lineage; does not prove current supporting assignments exist. |
| `permissions` | Explicit operations selected from supported parent authority. |
| `scope` | Local boundary constraints, retaining inherited constraints. |

No `recipient` belongs on this reusable revision. Changing its permission
source, parent or scope requires new content, not an edit beneath an existing
adoption. Immutability preserves the published content; it does not freeze live
upstream support or make the grant independent.

### Assignment and recipient

Approved core JSON:

```json
{
  "version": "1",
  "id": "A1",
  "grant_id": "G1",
  "grant_revision": 2,
  "recipient": {"type": "group", "id": "Team1"},
  "status": "enabled"
}
```

`id` identifies this assignment; `grant_id` chooses the grant and
`grant_revision` states its explicit adoption. `recipient.type` and
`recipient.id` identify the receiver, here Team1. The assignment's `status`
controls this route, not every assignment of G1. Its `version` describes this
record's format independently of the grant's content revision.

For a direct human recipient, the nested recipient type is `user`. Prefer
group-based access, but direct assignments are permitted where required support
is established. This does not settle competing parent-support eligibility for
every direct-human case. Agents receive human-dependent delegated authority;
do not invent a first-class independent agent assignment route here.

**Counterexample:** G1, A1 and Team1 existing does not make Nutan a member of
Team1. Assignment is not membership, and `enabled` is not a completed allow.

### Role and role revision

A **role** is a reusable permission bundle. A **role revision** identifies
immutable bundle content. A role does not identify recipients, confer membership,
or replace a grant's scope. Our grant revision uses either explicit
`permissions` or both `role_id` and `role_revision`, never both forms.

Approved role-reference grant shape, assuming the referenced role revision
contains certificate-read and fits legitimate G0 support:

```json
{
  "version": "1",
  "grant_id": "G-ROLE-READER",
  "revision": 1,
  "parent_grant_id": "G0",
  "role_id": "R-CERTIFICATE-READER",
  "role_revision": 1,
  "scope": {"dept": "FIN"}
}
```

`role_id` selects the bundle and `role_revision` selects its exact published
content. No implicit latest-role fallback exists. The role reference is approved;
the full standalone role-publication schema remains pending. Do not invent that
schema merely because a role is referenced here.

**Counterexample:** publishing a role revision that adds write must not silently
give write to assignments adopting this unchanged reader grant.

Sources: [core records](../../docs/grant-revision-format.md),
[record reference](../../docs/grant-record-reference.md),
[role variant](../../docs/role-grant-contract.md).

## Version and adoption

**Contract version** identifies the rules for interpreting a record's format.
An **authority revision** identifies immutable content. **Publication** makes
new content available; **adoption** explicitly selects content for use through
an assignment. These are different events, not synonyms for an update.

| Field/event | Example | Consequence |
|---|---|---|
| `version` | String `"1"` | Supported format; missing/unsupported versions are not guessed. |
| `revision` | G1 content revision `2` | Identifies the grant's published content. |
| `grant_revision` | A1 selects `2` | A1 uses that content while its route remains valid. |
| `role_revision` | Reader bundle revision `1` | The role-based grant selects that bundle content. |
| Publish G1 revision 3 | A new immutable publication | Does not rewrite A1's selection. |
| Explicitly upgrade A1 | Validated adoption of latest content | Changes the selection only after applicable checks. |

**Latest revision** must be qualified. New assignments and explicit upgrades
select the grant's latest published revision. Required parent support uses the
actual adopted content in the eligible lineage, resolved top-down—not whichever
broader parent revision happens to exist elsewhere. Re-enabling an assignment
does not implicitly upgrade it.

The fields already shown express these distinctions; no extra adoption object
or parent-revision selector is introduced. `parent_grant_revision` and
`parent_assignment_id` are not approved additions to grant content.

**Counterexample:** fetching the newest parent definition globally, rather than
the actual supporting lineage, can change a child's authority without the
required adoption and boundary checks.

Sources: [versions](../../docs/contract-publication.md),
[revisions](../../docs/grant-revisions.md),
[direct-human context](../../docs/direct-human-parent-context.md).

## Teams and administration

**Team** and **group** are synonyms in our model: an Auth-owned collection of
explicit human members. **Membership** is the human-to-group relationship through
which a human receives that group's valid assigned authority. Applications may
keep separate business groupings; synchronization into Auth is explicit when
those groupings must affect authorization.

**Ownership** and **administration** describe explicit administrative authority,
not membership or automatic business access. An **owner** or **administrator**
must be authorized for the actual operation and administrative boundary.
**Issuer** describes who created or assigned authority at a particular time;
it does not automatically make that person the permanent support of a team-held
grant route.

**Administrative authority is an ordinary grant.** This is the model's answer to
"who may administer this?", and it needed no new mechanism: an administrative
grant carries platform-namespace permissions, its chain begins at the tenant's
**Auth root**, and it is created, assigned, narrowed, disabled, deleted and
resolved by the machinery already described here. A tenant therefore has two
chains and one set of rules:

```
  ACME'S AUTH ROOT                        ACME/HRMS'S APPLICATION ROOT
  every active auth: permission           every active hrms: permission
      │ parent                                │ parent
      ▼                                       ▼
  auth:group::write  { team: Team2 }       hrms:employee:certificate::read
      │ parent                                │   { dept: FIN }
      ▼                                       ▼
  narrower administrative authority        narrower business authority

  SAME walk · SAME narrowing · SAME containment · SAME ceiling rule
```

What bounds an administrative operation depends on which side it is.
**Tenant administration** — teams, grants, assignments, roles, establishment —
is bounded by **a team**, carried as the platform scope key `team`.
**Platform administration** — permission and scope registration, catalog reads,
application role publication — is bounded by **the area itself**, and needs no
scope key: registering a permission into a shared catalog has nothing to do with
any one team.

**Sideways escalation is impossible rather than guarded.** Scope predicates
accumulate conjunctively along a chain and a request carries one value per key,
which supplies the whole containment property:

| Proposed grant | Result |
|---|---|
| `{team: X}` beneath a parent scoped `{team: X}` | **allowed** — the predicates agree; this is how one owner appoints another owner of the same team |
| `{team: Y}` beneath a parent scoped `{team: X}` | the route demands `team=X` **and** `team=Y`, which no request satisfies — it authorizes nothing, ever |
| `{team: X}` beneath a parent scoped `{}` | **allowed** — and only an unrestricted grant from the Auth root has `{}`, which is the tenant administrator |

Note what that third row means: an operation with **no bounding team** — creating
a top-level team, or moving one up to become top-level — is satisfied only by a
route carrying no predicate at all. Nothing enforces any of this; it is what the
model already computes.

**One wart, recorded rather than hidden.** Narrowing *accepts* the contradictory
re-scope in the middle row and builds a route that never matches, rather than
refusing the write. The safety is "authorizes nothing", not "cannot be written".

**Ownership is a grant too, and the separate relation is superseded.** Creating a
team confers its ownership — the creator receives an administrative grant over it
— and a team must always retain at least one enabled administrative assignment, so
no team is ownerless. The earlier ownership *relation* carried no authority at
all: nothing in resolution or validation consulted it, so "who may administer this
team" was recorded in a table no decision read.

**A synchronization is an ordinary authorized caller.** An application
synchronizing its business membership into Auth does so as a service account
holding team-write authority within a definite scope, calling ordinary endpoints.
It has no privileged path and cannot write the store directly; nothing can.
Whether a deployment offers one membership write per call or one call changing
many is endpoint design, not an authorization question — a bulk write inside one
team is one boundary and one evaluation, and one spanning several teams is
governed by the existing per-item coverage rule.

Team create covers creating teams and subteams; team write includes human
membership management; team delete removes teams. These do not by themselves
authorize assigning business grants. **Assignment authority** requires the
separate administrative operation/recipient authority plus the proposed
authority fitting valid source support available to the assigner. Child-team
ceilings still apply.

Membership administration is powerful: adding Nutan distributes the team's
existing effective access. The approved rule does not invent an additional
requirement that this membership administrator personally possess each of the
team's business permissions. Changing the team's grants is a different operation.

**Representation:** A1's approved nested `recipient` identifies Team1, and an
administrative grant uses the approved grant and assignment shapes unchanged —
that is the point of administration being a grant rather than a second model.
Complete team and membership *records* remain pending; earlier tentative scratch
JSON must not be presented as finalized contracts. The relationship can be
persisted in database tables without that choosing a public wire schema. The
ownership relation is not pending — it is superseded.

**Example:** Nutan is a member of Team2; Maya administers an assignment; Om may
hold separate Team1 administration. Those are three different relationships.
Maya's name on an issuance record does not make her all three participants.

**Counterexample:** rotating Team1's owner must not import the new owner's
personal permissions into Team1. Team-held supporting authority normally
continues unchanged when its actual support remains intact.

Sources: [groups and synchronization](../../docs/groups-and-membership.md),
[team administration](../../docs/team-administration.md),
[assignment authority](../../docs/assignment-authority.md),
[administration is a grant](../../docs/administrative-authority.md),
[ownership](../../docs/ownership-lineage.md).

## Dependent relationships

**Parent** and **child** describe a dependency with a non-expansion rule.
**Subteam/subgroup** means a team relative to its parent; **subgrant** means a
grant relative to its parent grant. They are not new entity types.

**Grant lineage** follows declared grant parents and actual supporting authority.
**Team lineage** follows team parentage and imposes the parent team's ceiling on
all child-team authority. **Scope lineage** describes accumulation of restrictions
along supported grant lineage; it does not create a stored scope entity.

**Support** is the required live authority and relationships making a route
usable. An **authority route** ties an assignment and its adopted content to
the applicable parent/team, membership and delegation support. A **binding**
describes those relationships jointly; it is not an extra grant field.

Continue G1/A1/Team1 from above. Assume Team2 is established as Team1's child,
G2's global control is enabled, and G2 revision 1 is latest at validated adoption.
Approved grant-revision and assignment shapes instantiated for the child:

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

`parent_grant_id` declares G1, while A1 establishes Team1's actual adopted G1
content. A2 supplies G2 to Team2. The team-parent relationship is a required
premise here: its full JSON is pending, not encoded by inventing a `parent_team_id`
field. Nutan separately needs explicit Team2 membership.

![Separate grant and team lineages joined by assignments, with explicit membership](../assets/model-lineage.svg)

The grant selects read from read/write and adds C17 to Finance. Effective child
reach is Finance AND C17. Every Team2 authority route must also fit Team1's
ceiling. Subset includes equality: selecting all permitted operations or adding
`{}` does not require artificial further narrowing.

**Counterexamples:** Team2 parentage does not make Nutan a Team1 member. A grant
held by unrelated TeamX cannot replace Team1's required support. Matching JSON
keys must not overwrite a parent's restriction. Separate permission and scope
fragments from unrelated routes cannot be combined into broader authority.

**A node that cannot be reached closes its route.** Walking a route means
reaching every node in its chain; one that cannot be reached means the route
supplies nothing, and that is an ordinary completed answer rather than a failure.
The scattered cases were always this one rule:

| Why the node cannot be reached | |
|---|---|
| Its supporting binding is absent or disabled | the assignment no longer carries it |
| Its parent no longer carries what it selects | narrowing fails at that step |
| A permission it selects is **unregistered** | the catalog never supplied it |
| **Every** permission it selects is retired | nothing it selects survives — see below |
| Its validity window has not opened, or has closed | the revision is not in force |
| The group holding it no longer has the member | membership was withdrawn |

None of these is an error and none reaches beyond its own route: other routes the
human holds are unaffected.

**Retirement is the case that does not belong on that list without qualification.**
A grant supplies what it selects and the catalog still supplies, so retiring one of
several selected permissions **narrows** the route rather than closing it — the route
closes only when nothing it selects survives. An earlier statement of the
unreachable rule listed a retired permission flatly as a closing cause; the later
retirement decision governs where the two differ.

**Unreachable is not unreadable, and the distinction is the whole safety of the
rule.** A record that cannot be *read* — one that does not parse, or is
structurally invalid — is not a closed route; the answer fails. Closing a route
can only ever remove authority, so skipping an unreachable node is safe. Silently
closing a route because a row was damaged would mean answering authorization
questions from a store we have admitted we cannot fully read, and the quiet denial
would look exactly like a correct one.

**A grant with a dependent can be neither disabled nor deleted.** While another
grant names it as parent, both operations are refused, for every grant including a
trusted root — there is no root exception. A *dependent* here means a child grant
and not an assignment: holding a grant is an assignment, so counting assignments
would make every grant anybody holds undisablable. Delete separately refuses while
an assignment still names the grant, which is the older rule against leaving a
dangling reference. Dismantling is therefore bottom-up.

The grant parent link is approved; complete team/support evidence contracts and
some direct-human support eligibility remain pending. Current structural guards
also apply: inspect affected bindings, disable/remove them bottom-up as required,
and validate current reality on explicit re-enablement. Ancestor ineffectiveness
alone is not equivalent to disabling a child's own binding. Cycles are rejected,
including in disabled structures.

Sources: [lineage, orphans and unreachable nodes](../../docs/authority-lineage.md),
[dependents and dismantling](../../docs/grant-lifecycle.md),
[subgroups](../../docs/subgroups.md),
[four-part bindings](../../docs/parent-grant-bindings.md),
[cycles](../../docs/lineage-cycles.md).

## Delegation

**Delegation** is an authority dependency through which a proxy acts for a human
under limits. The **human ceiling** is that human's currently applicable
authority; **delegation limits** can narrow it further. Being the account's
creator is not a substitute for establishing this current support.

In our model, proxy authority must fit both. V1 supports direct human-to-proxy
delegation, not proxy-to-proxy chains. An agent does not become a first-class
member of Employees when Vinay delegates access supported by that group.

**A delegation is not a grant and not a step in a lineage.** It adds no authority
and never appears in a contributing chain. It is permission for an **actor** to ask
about a **human**, and the authority that comes back is the human's:

```
  vinay ──▶ agent A                agent A may ask about vinay
                                   the answer is VINAY'S authority
                                   A adds nothing to it
```

That is why a delegation needs no representation in a resolved answer: the
authority-loading question already names the asking actor and the subject
separately, and a delegated request differs from a direct one only in the actor.

**Lifetime.** A delegation carries its own validity window, and the effective
lifetime is the narrower of the delegation's and the human's authority. This is
the existing rule applied rather than a new one — a route's validity is already
the narrowest window in its chain, and a delegation is another constraint on the
same request. Without a window of its own, a delegation would be the only thing in
the model that is revocable but never expires.

**Growth.** A delegation *tracks* the human's authority: what a proxy may do is
resolved from what its human holds **at the time of the request**, not at the time
the delegation was created. A subset is resolved, not copied; freezing a snapshot
would let a delegation diverge from the authority it depends on with nothing
reconciling them.

**Representation:** the identity JSON above identifies actor and human. It is
not the delegation grant, evidence or lifecycle record. The complete delegation
representation remains pending; adding a guessed `delegation_id` or permissions
array to identity would silently invent a contract.

**Example:** Vinay has effective read/write, while A-17's delegation permits
only read. A-17 cannot write. If Vinay loses required read support, A-17 loses
that route too. Restoring support can restore a still-valid dependent delegation;
it does not resurrect one explicitly revoked, disabled or expired.

Source: [delegation lifecycle](../../docs/delegation-lifecycle.md).

## Lifecycle

**Enablement** is an explicit control making a record eligible to participate.
**Effectiveness** is the computed outcome of all required controls, content,
support and restrictions. **Disablement** withdraws participation without
deleting the record. These distinctions explain why assignment does not imply
enablement, and enablement does not imply access.

Approved grant-control shape in its disabled state:

```json
{
  "version": "1",
  "id": "G2",
  "status": "disabled"
}
```

This is an alternative lifecycle snapshot, not a second simultaneous G2 control.
Disabling G2 affects all routes through it. Disabling A2 instead affects that
assignment. If G3 depends on G2, G3 becomes ineffective when G2 is disabled; a
still-enabled, otherwise valid G3 can resume when support resumes. An explicitly
disabled G3 stays disabled until explicitly enabled under current checks.

**Validity** is an optional local time restriction on immutable grant content.
**Expiry** is its exclusive end; an optional `not_before` is an inclusive start.
Approved extended grant-revision shape, using an independent temporary example:

```json
{
  "version": "1",
  "grant_id": "G-TEMP-READER",
  "revision": 1,
  "parent_grant_id": "G1",
  "permissions": ["hrms:employee:certificate::read"],
  "scope": {"cert": "C17"},
  "validity": {"expires_at": "2026-09-30T00:00:00Z"}
}
```

Assume validated assignment and required live support. `validity` belongs to
the revision, not live control or assignment. The shown `expires_at` excludes
access at and after that instant. Absence of a local window does not remove
inherited time limits. Changing the window requires new content and explicit
adoption. Assignment-specific validity is deferred in v1.

**Publication is free; adoption is where the evaluation happens.** A new role
revision is validated against its own shape, the catalog and the publisher's
authority. It does not consult the grants that adopted earlier revisions and
cannot be refused because one of them would be affected — a role an application
ships to every tenant would otherwise let one tenant's private grant structure
block a shared catalog change, and the publisher has no authority to repair it.
When a grant *adopts* a newer revision, every enabled dependent beneath it is
re-evaluated against the revision being adopted; if a dependent that resolved
before would not resolve after, the adoption is refused and the grant is left
exactly as it was. The evaluation is differential rather than absolute — a
dependent that did not resolve before does not block the adoption — and a disabled
dependent is skipped, to be revalidated when it is enabled.

**Deletion/revocation** permanently removes the affected authority binding or
record in this model; it is not temporary disablement. No separate reversible
`revoked` control value or delete API is introduced by that wording. Structural
guards still govern deletions that would break dependent bindings, and a grant
another grant names as parent can be neither disabled nor deleted while that is
true — see [dependent relationships](#dependent-relationships).

An **orphan grant/route** lacks required parent support in its declared lineage.
Its affected descendants cannot supply authority. Orphaning is assessed for the
affected route, not automatically every use of a reusable definition. No new
`status: orphan` JSON field is approved: orphaning is a derived condition.

| Situation | Interpretation |
|---|---|
| Unassigned reusable grant | Not automatically orphaned. |
| Legitimate root without required parent | Not orphaned. |
| Required parent support removed | Affected lineage is orphaned and ineffective. |
| Required support disabled or expired | Ineffective; not proof that support records are missing. |
| Auth lookup times out | Evidence unavailable; not proof of orphaning or completed denial. |

**Counterexample:** enabling the temporary grant after September 30 cannot
reset its expiry. Likewise, an orphan's stored child cannot repair its support
by automatically choosing a different parent.

Sources: [lifecycle and dependents](../../docs/grant-lifecycle.md),
[role publication and adoption](../../docs/role-revisions.md),
[validity](../../docs/grant-validity.md),
[assignment validity](../../docs/assignment-validity.md),
[orphan definition](../../docs/authority-lineage.md).

## Roots and bootstrap

A **root grant** is legitimately established initial authority with no required
parent grant. **Bootstrap** is the trusted procedure establishing that starting
authority, its intended administrator group, legitimate human membership and
explicit assignment. Ordinary callers cannot manufacture roots by omitting a
parent from grant content.

**A tenant operates in two namespaces, and establishment belongs to the first.**
One tenant holds two authorities and neither implies the other:

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

Establishing an application's root in a tenant is the **closing step of enabling
that application for that tenant**, authorized by the same authority that enabled
it. Not by a permission, because requiring one is circular — a grant needs a
parent, up to some root, which had to be established, which would need that
permission. Nothing terminates that chain inside the grant model; it has to be
started from outside it. And not by a separate platform operator, because the
decision is the tenant's: the party that turned the application on is the party
that says what authority it starts with and to whom.

The handover is explicit. Establishment names the team that will hold the root, so
the act creating the authority also names who first receives it. Afterwards the
platform-namespace authority is finished, and every subsequent act is authorized
against the root.

No separate ceremony record is introduced. The root grant carries its own trust
marker and the establishing actor is recorded as the actor of that write; what
makes an establishment special is what it creates, not how it is recorded.
Replacing a lost or wrong root is establishment again rather than a distinct
operation — the subtree is dismantled bottom-up first, so two roots of one
application never coexist in one tenant.

Our intended bootstrap starts with the maximum intended permissions and scope
within its authorized boundary, using a minimal coherent setup. Minimal setup
does not mean arbitrarily underpowered authority that cannot administer the
system. Registration precedes acceptance; partial setup must not expose partial
authority.

**Computed root coverage** follows the applicable registered application catalog.
The application has one shared catalog, not selectively adopted tenant releases.
Catalog growth does not silently add permissions to ordinary child revisions,
create membership or enlarge scopes.

**Representation: format pending.** Parent omission for trusted roots is agreed,
as is the trust marker on the grant head; the complete computed-root source
encoding remains unsettled. No `*`, `is_root`, or catalog-source field is invented
to fill that gap.

**Counterexample:** an ordinary derived grant with a missing parent is not a
bootstrap shortcut. A platform administrator publishing a permission does not
thereby receive tenant business access.

Sources: [bootstrap and the two namespaces](../../docs/bootstrap-authority.md),
[initial setup](../../docs/bootstrap-initial-assignment.md),
[root evolution](../../docs/root-permission-evolution.md).

## Endpoint declaration

An **endpoint policy** is a server-owned declaration connecting an endpoint's
method/path to one required permission and selected inputs. It describes what
must be evaluated; it is not an allow result or a client-selected grant.

Approved GET policy core JSON:

```json
{
  "version": "1",
  "method": "GET",
  "path": "/api/v1/{tenant}/{dept}/{cert}",
  "permission": "hrms:employee:certificate::read",
  "inputs": {
    "tenant": {"source": "path", "name": "tenant"},
    "dept": {"source": "path", "name": "dept"},
    "cert": {"source": "path", "name": "cert"}
  },
  "trusted": {"tenant": "tenant"}
}
```

| Field | Meaning and rationale |
|---|---|
| `version` | Supported endpoint-policy format. |
| `method`, `path` | The declared operation's route binding. |
| `permission` | Exactly one required registered operation, not a permission list. |
| `inputs` | Selected local input names; not every body field becomes authorization material. |
| `source` | Where the value must be obtained. |
| `name` | Parameter/field name at that source; can differ from the local input name. |
| `trusted` | Correlates a field of the trusted request context with one of this policy's declared inputs. Required for the tenant. |

`"trusted": {"tenant": "tenant"}` reads as: *the input this policy calls `tenant`
must equal the trusted tenant.* It is the mechanism for an obligation stated
everywhere and previously left unspecified — a route's tenant claim must be bound
to trusted context, and a field name alone proves nothing. A policy declaring no
tenant correlation **cannot be mounted**, so forgetting becomes impossible rather
than silent, and a path carrying `{tenant}` or `{application}` that no correlation
names is refused for the same reason. This is not the relationship language the
model declined: both values are already in the gate's hand, nothing is looked up,
and no record relationship is asserted.

Approved PUT policy core JSON makes the local/source distinction visible:

```json
{
  "version": "1",
  "method": "PUT",
  "path": "/api/v1/{tenant}/certificates/{cert}",
  "permission": "hrms:employee:certificate::write",
  "inputs": {
    "tenant": {"source": "path", "name": "tenant"},
    "cert": {"source": "path", "name": "cert"},
    "proposed_dept": {"source": "body", "name": "department_id"}
  },
  "trusted": {"tenant": "tenant"}
}
```

`proposed_dept` is the local name; `department_id` is the selected body field.
The body proposes a department. It does not establish the existing certificate's
current department. The application owns value validation and must establish or
enforce required relationships. No canonical `relationships` block is adopted.

Every declared input must be present at its declared source. A query parameter
cannot silently replace a missing body field. A broad `{}` grant does not make
the endpoint's required inputs optional. A declared input the request does not
supply is a **refusal of the request** — there is no implicit default, no empty
string and no fallback to another source.

**A policy is validated structurally at mount time, not at request time.** A
policy that cannot be mounted cannot guard an endpoint, so the endpoint does not
serve rather than serving unguarded:

| | |
|---|---|
| Version | Exactly the supported contract version; no default is guessed. |
| Method | A valid uppercase HTTP token, and the request's method must equal it. |
| Path | Absolute, no query or fragment, every placeholder well formed and unique. |
| Permission | Exactly one, canonical, no wildcard and no list. |
| Path inputs | Must name a placeholder the path actually declares. |
| Body inputs | **Top level only.** A selector containing `.`, `[`, `]` or `/` is refused. |
| Sources | Exactly two: `path` and `body`. |
| Trusted | Required for the tenant. |
| Unknown fields | A field this version does not define is **refused, not ignored**. |

**Nested body selection is refused rather than deferred.** Once `a.b` is admitted,
arrays, filters and absence semantics follow, and the policy becomes a place where
application structure is described. An application needing a nested value reads it
in its binder, validates it, and supplies it as material under its own local name.
Value validation stays with the application: a policy may be perfectly valid and
still name an input whose value the application rejects.

**What a policy still does not declare** is the scope *boundary* its endpoint
operates at, so the boundary reaching evaluation is whatever the binder supplies at
request time. An endpoint may therefore claim a narrow boundary and read a wide
one, and nothing can check it. That question is framed and **not approved**.

**Counterexample:** the GET path claims Finance, but a later ID-only lookup
returns an Engineering certificate. Correct declaration and extraction have not
enforced the authorized boundary.

Sources: [endpoint policy, trusted correlations and mount-time validation](../../docs/endpoint-policy-format.md),
[the open boundary question](../../docs/policy-scope-boundary.md).

## Request and material

A **request** is the particular operation sought under the endpoint declaration.
A **request input** is a selected value obtained from its specified source.
**Request material** includes inputs, verified context and trusted facts or
enforceable constraints sufficient to evaluate and bind that operation.

A **domain/application fact** is application-owned knowledge, such as the
department of C17. A **relationship** is an association relevant to a boundary,
such as C17 belonging to Finance. These are necessary meanings in some checks,
not a required canonical relationship entity or policy block.

A **resolved request** is the evaluation-ready view: the required operation and
material have been established with their meanings and bindings. It is not
already authorized and does not create a reusable entitlement.

Continue Nutan's Team2 example and the GET policy above:

| Stage | Concrete meaning | What is not established yet |
|---|---|---|
| Request | Nutan asks for `GET /api/v1/acme/FIN/C17`. | The path is not proof of tenant access or certificate department. |
| Extracted inputs | `tenant = acme`, `dept = FIN`, `cert = C17` from the declared path. | Matching text is not an application fact. |
| Trusted context and data binding | Establish Nutan's identity/tenant; establish or constrain the operation to Finance and C17. | This still does not supply the read permission. |
| Resolved request | Evaluate certificate-read for this human and this bound operation/material. | Resolution is not allow. |

**Representation: full request/resolved-request formats pending.** The identity
and endpoint policy above are approved building blocks. This table illustrates
meaning; it does not approve fields such as `resolved_inputs` or a new envelope.

The endpoint may establish a fact before evaluation or preserve a constraint
through actual execution. The model does not require Auth Service to query the
application database. It does require that the operation cannot escape the
boundary on which authorization relied.

**Evidence that is missing, invalid or unsupported has three different answers,
and none of them is an allow:**

| | Result |
|---|---|
| Required material missing — a declared input absent, or the binder cannot produce it | failure to establish; the request is refused before a decision is reached |
| A route's own evidence invalid — an ill-formed predicate, content that does not validate, an unreadable chain | that route closes |
| …and no other route authorizes | failure to establish, not a denial — a route that could not be read might have authorized |
| Evidence of an unsupported kind — a reserved token other than `$self`, a wildcard where a boundary belongs, a contract version this consumer does not speak | refused, never interpreted, and never treated as absent |
| Every route readable, none authorizes | completed denial |

An unsupported value is not an empty one. None of this is a condition engine:
every case concerns evidence *this model already defines* — its own material, its
own predicates, its own contract versions — and nothing here evaluates a business
fact or is extensible by an application.

**Counterexample:** using `dept` only in logging does not make it enforced
material. Conversely, a grant's `{}` does not invent a Finance constraint simply
because the endpoint has a department input. Request bindings and actual
authority constraints both need their proper meaning.

Sources: [request vocabulary](../../docs/authorization-vocabulary.md),
[endpoint gate](../../docs/endpoint-authorization.md),
[missing, invalid and unsupported evidence](../../docs/grant-conditions.md).

## Resolution and evaluation

**Resolution** establishes usable request and authority views while preserving
their source support and restrictions. A **resolved grant** is a computed view
of an existing authority route, not a newly issued grant, independent copy,
assignment, or result. **Resolved grants** are the applicable route views;
plural does not mean flattening all permissions and scopes into one object.

**Evaluation** determines whether complete applicable authority covers the
required operation and boundary. **Non-amplification** means that resolution
cannot create authority beyond its valid sources. It is a rule about the
meaning of the result, not merely the number of fields or permissions returned.

For Nutan, assume the example's current support, membership and controls are
valid. The resolved G2 route retains:

| Component | Established meaning |
|---|---|
| Recipient route | Nutan's explicit Team2 membership → A2 → G2 revision 1. |
| Parent/team support | Team2's parent Team1 → A1 → G1 revision 2 → required G0 support. |
| Permission | Certificate-read only; parent write was not selected. |
| Boundary | Parent restrictions, including Finance, AND child C17. |
| Remaining dependencies | Applicable controls, time windows and any human/proxy limits. |

**This view now has an approved contract.** An application asks what one human
holds in one area, and never asks whether to allow:

```json
{
  "version": "1",
  "identity": {
    "version": "1",
    "actor": {"type": "service_account", "id": "agent_hrms"},
    "human_id": "U-17"
  },
  "options": {}
}
```

`actor` is the asking application's own credential; `human_id` is the person it
asks about. They are different parties, and that separation is the point — an
application asks as itself about many humans and is none of them. Nothing in the
question names an endpoint, a method, a resource or a verdict. `options` carries
two optional narrowings and nothing else: `permissions` restricts the answer (a
narrowing, never an assertion — omitted means everything the human holds, which is
the cacheable answer), and `omit_source` drops the explanation.

The answer echoes the three boundaries and lists every grant the human holds there,
after the chain has been walked:

```json
{
  "version": "1",
  "tenant_id": "acme",
  "application_id": "hrms",
  "human_id": "U-17",
  "resolved_grants": [
    {
      "version": "1",
      "grant_id": "G2",
      "revision": 1,
      "parent_grant_id": "G1",
      "permissions": ["hrms:employee:certificate::read"],
      "scope": {"dept": "FIN", "cert": "C17"},
      "validity": {"not_before": null, "expires_at": null},
      "source": {
        "assignment_id": "A2",
        "team_id": "Team2",
        "via": "membership",
        "lineage": [
          {"grant_id": "G0", "revision": 1, "assignment_id": "A0", "team_id": "Team0", "root": true},
          {"grant_id": "G1", "revision": 2, "assignment_id": "A1", "team_id": "Team1"},
          {"grant_id": "G2", "revision": 1, "assignment_id": "A2", "team_id": "Team2"}
        ]
      }
    }
  ]
}
```

| Field | Meaning |
|---|---|
| `version` | On the envelope and on every grant. A grant's version says how to read its scope and validity — the fields that decide a boundary. |
| `tenant_id`, `application_id`, `human_id` | The three boundaries, echoed, so a consumer can tell "nothing here" from "answered about somebody else". |
| `resolved_grants` | Every grant the human holds in this area. **Empty is a completed answer, never a failure.** |
| `permissions` | What this grant selects, already expanded from any adopted role. |
| `scope` | The boundary, **already folded down the chain** — a consumer never folds one itself. |
| `validity` | The narrowest window in the chain; absent means no automatic expiry. |
| `source` | Why the human holds it: the binding, the group, and the lineage root-first — G0 carries `"root": true` because G0 is the root, and the chain is complete rather than elided. |

A consumer must reject an unsupported version rather than guessing a default;
corroborate the three boundaries against its own question, since an answer about a
different tenant, application or human is unusable *in whole* rather than in part;
reject unknown fields, duplicate keys and trailing content; bound the answer's
size; and never follow a redirect, because following one hands the credential to
whoever set the header and then believes the reply. **Every failure here is an
evaluation error and never a denial.**

What is deliberately *not* adopted: the route it is served at, a batch form, and
freshness fields — nothing caches yet, so nothing can be stale, and a cache and an
epoch are adopted together or not at all. `source` is explanation, not
authorization: none of it may be used to widen what the grant already permits.

**Complete route** means keeping the operation, boundary and all required
support together. Finance-write from one route and Engineering-read from another
cannot become Engineering-write. Independent alternatives remain independent;
the system does not intersect every unrelated grant into a globally narrowest
scope either. For an approved finite batch, different complete routes may cover
different items, with every item covered before protected effects.

**An unusable route does not establish a denial.** A route the consumer cannot
read — malformed, internally inconsistent, carrying identifiers it cannot parse —
is exactly a route that might have authorized:

| Situation | Result |
|---|---|
| A route is unusable, and another complete route authorizes | **Allow.** The unusable route says nothing about the one that did. |
| A route is unusable, and no other route authorizes | **Evaluation error.** The denial is not established. |
| Every route is readable, and none authorizes | **Completed denial.** |

An unusable route therefore costs that route, and costs the *certainty* of a
denial — not the whole answer. This matters more now that an answer describes
everything a human holds in an area: failing the whole evaluation on any bad route
would take away every other grant that human held.

**Counterexample:** copying G2's local `{"cert":"C17"}` into a new independent
grant drops inherited Finance and source dependencies. That is not resolution;
it is unauthorized authority expansion.

Sources: [vocabulary](../../docs/authorization-vocabulary.md),
[lineage](../../docs/authority-lineage.md),
[authority loading transport](../../docs/authority-resolve-transport.md),
[decisions and unusable routes](../../docs/decision-results.md),
[batch coverage](../../docs/bulk-enforcement.md).

## Decision and enforcement

A **decision** is the completed authorization answer for the evaluated operation.
**Allow** means sufficient authority was established. **Deny** means evaluation
completed and the request was not permitted. An **evaluation error** means a
required evaluation could not finish; it is not proof that no entitlement exists.
Both deny and error stop protected execution.

Approved minimum allow result, for the G2 route above:

```json
{
  "version": "1",
  "decision": "allow",
  "grant_ids": ["G0", "G1", "G2"]
}
```

`grant_ids` is the **contributing chain of the route that authorized the request,
ordered root first**. Read left to right, that is the trusted root, the grant
narrowing it to one department, and the grant narrowing that to one certificate.
**The order is the dependency**: each entry is bounded by the one before it. It is
a non-empty array of supporting grant identifiers for *this* evaluation — not all
the human's grants, and not a reusable authorization for another request. These
references are **supporting evidence** for traceability, not new authority.
Returning them does not require logging every request or supplying a returned scope
field.

The richer alternative — assignment, team and revision per step — was considered
and not adopted: authority loading already answers in that shape, so a caller
needing it asks the service that owns those records. The order carries the
structure, and everything the richer form adds is history rather than structure.
The cost is stated rather than hidden: an endpoint holding only grant ids cannot
later say which *assignment* carried the authority if that assignment has since
been deleted. That is acceptable while the evidence exists to let an endpoint
account for its own effect, and would not be if the array were ever made the record
of last resort — which is the audit layer's question.

Approved minimum deny result, with illustrative reason/code text:

```json
{
  "version": "1",
  "decision": "deny",
  "error_code": "NO_AUTHORIZING_GRANT",
  "error_message": "You do not have access to this certificate.",
  "error_message_reason": "No grant authorizes this certificate read within Finance."
}
```

`error_code` identifies the cause programmatically. `error_message` is the
user-facing explanation; `error_message_reason` gives explanatory detail.
Both are evaluator-provided and reach the UI. The second field is not a private
server-only diagnostic channel.

**The code catalogue is open and its names are fixed.** A published code never
changes meaning; new codes may appear at any time. A consumer must tolerate a code
it does not recognize and fall back to the *class* the result arrived in — a
completed denial, or a failure to establish authority. So a consumer may not switch
exhaustively on codes and must not make a security-relevant choice from one: that
choice is already carried by the allow / deny / evaluation-error distinction, which
is the contract a consumer branches on. **A code explains; it does not decide.**

A `AUTHORITY_` prefix means the authority answer failed or failed to arrive; an
unprefixed name identifies the decision itself, the caller's own configuration, or
the specific way an answer disagreed with the question it was asked.

| Code | Meaning |
|---|---|
| `AUTHORITY_UNREACHABLE` | The authority service did not answer. |
| `AUTHORITY_TIMEOUT` | It did not answer in time. |
| `AUTHORITY_REFUSED` | It answered, and the answer was not a success. |
| `AUTHORITY_UNAVAILABLE` | An in-process authority store could not answer. |
| `AUTHORITY_UNREADABLE` | The answer did not match the contract, or was not the contract at all. |
| `AUTHORITY_AMBIGUOUS` | The answer did not mean exactly one thing. |
| `AUTHORITY_OVERSIZED` | The answer exceeded the accepted size. |
| `AUTHORITY_MALFORMED` | The answer was read, and this gate cannot use it. |
| `UNSUPPORTED_VERSION` | The answer states a contract version this consumer does not speak. |
| `WRONG_AREA` | The answer describes a different tenant or application. |
| `WRONG_SUBJECT` | The answer describes a different human. |
| `NO_AUTHORIZING_GRANT` | A completed denial: no complete route authorizes this operation. |
| `MISCONFIGURED` | The asking application's own setup is wrong, before any question is sent. |

`AUTHORITY_UNREADABLE` and `AUTHORITY_MALFORMED` are deliberately distinct: one
answer could not be read, the other was read and could not be used, and they occur
in different layers. The cost of an open list is stated plainly — **no consumer can
write an exhaustive handler**, and one that logs an unrecognized code without
alerting on it will swallow a new failure mode silently. That is accepted, because
freezing the list buys exhaustiveness in exchange for a contract version every time
a layer learns something new about how authority can fail to arrive.

Disclosure and value rules, and HTTP mappings, remain outside this contract.

Approved minimum evaluation-error result, with a published timeout code:

```json
{
  "version": "1",
  "error_code": "AUTHORITY_TIMEOUT",
  "error_message": "We could not check your access.",
  "error_message_reason": "The authorization service did not respond in time."
}
```

This assumes required authority could not be loaded and no sufficient valid
material was already available. There is no `decision` because evaluation did
not establish allow or deny. Missing `decision` alone does not make arbitrary
JSON a valid error. Validate the expected variant, reject unknown/mixed result
fields and do not repair a malformed result into allow.

**Enforcement** prevents unchecked execution and keeps actual data/effects within
the evaluated boundaries. It is an activity, not another result field. An allow
for Finance/C17 cannot authorize an unchecked read of a different certificate.
Evidence, evaluation and execution must remain about the same operation.

**What a response discloses is the endpoint's duty, not the gate's.** A caller
holding write and not the matching read issues an update: the write is authorized
and succeeds, and the response carries the whole record — fields the caller never
supplied and now learns by writing. No rule here is violated, and the gate could
not have prevented it; it cannot know what a body contains and should not be given
the job of finding out. That does not make disclosure nobody's duty. It is
authorization work, assigned where the split already puts it: the endpoint keeps
execution, and its output, inside the authorized boundary. The tradeoff is worth
naming — an endpoint author who reads only the policy contract may never think
about the response at all, which is why it is stated as a rule with a name to point
at during review.

**Counterexample:** returning a timeout as `NO_AUTHORIZING_GRANT` claims a
completed policy conclusion the evaluator never reached. Returning allow with
an empty supporting list violates the approved minimum evidence contract.

Sources: [result contracts, the code catalogue and the contributing chain](../../docs/decision-results.md),
[response disclosure](../../docs/endpoint-authorization.md),
[what an audit consumer may rely on](../../docs/authority-change-audit.md).

## Responsibility layers

These terms explain ownership, not additional authority records:

- **Auth Service** owns shared authorization state and canonical validation.
  Its own administrative endpoints are also protected by the framework.
- **Canonical layer** supplies shared identity, grant, assignment, lineage and
  evaluation rules. It does not know application database relationships merely
  because a scope contains a registered key.
- **Application layer** supplies the meanings and trusted facts or constrained
  execution that bind those rules to actual application data.
- **Auth agent / evaluator** applies the rules with the required material at the
  endpoint-owned gate. Middleware may establish identity or preload authority;
  no prepared handoff or provisional application allow is required.
- **Authority-boundary validator** is Auth's internal check that proposed grant
  or assignment authority stays within valid sources and team ceilings. Auth's
  ordinary endpoint permission check establishes authority to perform the
  administrative operation; it does not alone prove this containment.

The two Auth checks both run inside Auth Service when Auth authority is changed.
They are not permission to consult an application database for every grant
write, and not a second business-rule engine. Their rationale is that authority
to perform an assignment operation and authority to distribute its proposed
content answer different questions.

**One gate in front of one endpoint, and the order is the authorization.** An
application supplies four things and receives one:

| | What it is | What it must not do |
|---|---|---|
| **Policy** | The endpoint's static declaration. | Change after mounting. |
| **Identity source** | Establishes the trusted request context. | Read the business body — it is handed a request whose body is empty. |
| **Binder** | Validates application schema, selects the material, and returns the effect as a closure over those same validated values. | Publish output, or perform the effect. |
| **Failure handler** | Renders a denial or an evaluation failure. | Turn one into the other, or proceed. |

And it receives the **Result** — the decision and, on an allow, the contributing
chain. The order is: method and route, then identity (from a request with no
business body, so a credential can never be taken from the payload it protects),
then the trusted correlations, then the binder, then authority loading and
evaluation, and **only on an allow** the effect, which is handed the Result and no
HTTP request — so it cannot reparse a path or a body after the decision. The binder
runs *before* evaluation deliberately: the effect must close over the same
validated values the decision was made about, and a binder running afterwards could
bind to something else.

An application linking this gate links **no authority records, no schema and no
authority store**. It needs the gate and a source of answers, and nothing else.
Deliberately unspecified: an SDK, the HTTP status mapping for each failure kind,
and any transport for administration — one endpoint is published, and
administrative operations have no wire contract.

No new service API, evaluator signature or validator-result JSON is adopted in
this reference. Those integration details belong to the implementation guide
and remaining contract work.

Sources: [responsibilities](../../docs/system-overview.md),
[Auth boundary gate](../../docs/authority-boundary-validation.md),
[handler integration contract](../../docs/handler-integration-contract.md).

## What this reference deliberately does not invent

The core vocabulary needs neither a canonical target entity nor a prepared
decision state. Subteams and subgrants remain teams and grants. Orphaning and
effectiveness remain derived meanings, not extra live status fields. Scope
composition remains AND, not a new query language or JSON merge operation.

Membership records, registration, full role publication, delegation evidence, the
resolved-*request* envelope and computed-root encoding are explicitly marked
pending where discussed. Explaining them here does not approve a missing schema.
See the [pending register](../appendices/pending.md) for the remaining choices, not
a new question quota.

Two things previously listed there are no longer pending, and one is superseded
rather than completed. **Authority loading has a contract** — the question and its
answer are approved, above. **Administration is an ordinary grant**, so it needed no
administrative authority model of its own. And **the ownership relation is
superseded**: ownership is a grant, not a table, so there is no owner record left
to specify.

What remains genuinely open is worth naming precisely rather than gesturing at: the
boundary an endpoint operates at is not declared, so a policy can claim a narrow
boundary and read a wide one; administrative operations have no wire contract;
freshness fields and caching are adopted together or not at all, and neither is;
non-HTTP and background integration is explicitly deferred for v1; and direct-human
parent-support eligibility is unchanged.

**Foundation takeaway:** a reader should now be able to distinguish what names
an operation, what bounds it, what defines authority, how someone receives it,
what keeps it supported, and how a request becomes a constrained effect—while
recognizing which JSON carries each agreed part.

[Return to the Foundations narrative](01-foundations.md) · [Continue to implementation](../implementation/05-canonical-model.md)
