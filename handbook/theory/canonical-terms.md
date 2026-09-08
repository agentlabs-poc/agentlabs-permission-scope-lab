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
| [Scope and self](#scope-and-self) | Scope, boundary, selector, scope key/value, `$self`, additional/effective scope | Nested flat object; effective composition is not an object overwrite. |
| [Restrictions and conditions](#restrictions-and-conditions) | Constraint, restriction, condition | Existing limits and rules; no generic condition JSON is adopted. |
| [Registration](#registration) | Catalog, permission definition, scope definition, compatibility declaration | Meanings agreed; registration payload pending. |
| [Grant records](#grant-records) | Grant, grant identity/control, grant revision, recipient, assignment, role, role revision | Approved control, revision, assignment and role-reference shapes. |
| [Version and adoption](#version-and-adoption) | Contract version, revision, publication, adoption, latest revision | Existing version/revision fields; activities are not new entities. |
| [Teams and administration](#teams-and-administration) | Team/group, membership, ownership, owner/administrator, issuer, assignment authority | Recipient reference exists; full relationship/owner records pending. |
| [Dependent relationships](#dependent-relationships) | Parent/child, subteam/subgroup, subgrant, grant/team/scope lineage, support, authority route, binding | Grant parent link and assignments; full team hierarchy format pending. |
| [Delegation](#delegation) | Delegation, human ceiling, delegation limits | Identity block identifies the participants; delegation-evidence format pending. |
| [Lifecycle](#lifecycle) | Enablement, disablement, effectiveness, validity, expiry, deletion/revocation, orphan | Live controls and revision-local validity; no invented orphan state field. |
| [Roots and bootstrap](#roots-and-bootstrap) | Root grant, bootstrap, computed root coverage | Trusted-root behavior agreed; complete root/setup representation pending. |
| [Endpoint declaration](#endpoint-declaration) | Endpoint policy, method, path, required permission, input, source, local input name | Approved GET and PUT policy core JSON. |
| [Request and material](#request-and-material) | Request, input, material, domain/application fact, relationship, resolved request | Policy/identity examples available; full request/resolved envelopes pending. |
| [Resolution and evaluation](#resolution-and-evaluation) | Resolution, resolved grant, resolved grants, evaluation, complete route, non-amplification | Computed views and activities; full resolved-grant transport pending. |
| [Decision and enforcement](#decision-and-enforcement) | Allow, deny, evaluation error, result, reason, supporting evidence, enforcement | Approved minimum result variants; evidence details remain incomplete. |
| [Responsibility layers](#responsibility-layers) | Auth Service, auth agent/evaluator, canonical layer, application layer, authority-boundary validator | Logical responsibilities, not new JSON entities. |

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
wildcards are supported in this model. Reusing a retired identifier for a
different authorization meaning is not permitted.

**Counterexample:** certificate-read does not confer certificate-write, even
though the names share a prefix. A Finance scope cannot supply the missing verb.

Source: [permission model](../../docs/permission-model.md).

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

**Counterexample:** recognizing `dept = FIN` in a request is not proof that the
certificate returned belongs to Finance. The endpoint must establish or enforce
that relationship against the actual operation's data.

Runtime self meaning is settled. Whether one person's self-scoped source
authorizes distributing another person's self access remains a distinct
source-binding question; copying `$self` text does not answer it.

Source: [scope model](../../docs/scope-model.md).

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

**Counterexample:** registering both `dept` and repository-read does not alone
prove that a department selector is a supported repository boundary. Nor does
registering read automatically assign read to the publisher.

Source: [application registration](../../docs/application-registration.md).

## Grant records

A **grant** is a reusable authority definition: a permission source, scope and
applicable restrictions and dependencies. A **recipient** is the human or group
to whom an **assignment** binds a grant. Separating them lets one definition be
reused without pretending that creating a definition gives someone access.

Our model distinguishes live grant control, immutable content and assignment.
The following is one consistent running example. Assume registered certificate
permissions and scope keys, legitimate upstream G0 support, existing Team1, and
successful authorization/boundary checks. G1 revision 2 is latest when A1 is
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

**Representation:** A1's approved nested `recipient` identifies Team1. Complete
team, membership and owner records remain pending; earlier tentative scratch
JSON must not be presented as finalized contracts. The relationship can be
persisted in database tables without that choosing a public wire schema.

**Example:** Nutan is a member of Team2; Maya administers an assignment; Om may
hold separate Team1 administration. Those are three different relationships.
Maya's name on an issuance record does not make her all three participants.

**Counterexample:** rotating Team1's owner must not import the new owner's
personal permissions into Team1. Team-held supporting authority normally
continues unchanged when its actual support remains intact.

Sources: [groups](../../docs/groups-and-membership.md),
[team administration](../../docs/team-administration.md),
[assignment authority](../../docs/assignment-authority.md),
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

The grant parent link is approved; complete team/support evidence contracts and
some direct-human support eligibility remain pending. Current structural guards
also apply: inspect affected bindings, disable/remove them bottom-up as required,
and validate current reality on explicit re-enablement. Ancestor ineffectiveness
alone is not equivalent to disabling a child's own binding. Cycles are rejected,
including in disabled structures.

Sources: [lineage and orphans](../../docs/authority-lineage.md),
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

**Deletion/revocation** permanently removes the affected authority binding or
record in this model; it is not temporary disablement. No separate reversible
`revoked` control value or delete API is introduced by that wording. Structural
guards still govern deletions that would break dependent bindings.

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

Sources: [lifecycle](../../docs/grant-lifecycle.md),
[validity](../../docs/grant-validity.md),
[assignment validity](../../docs/assignment-validity.md),
[orphan definition](../../docs/authority-lineage.md).

## Roots and bootstrap

A **root grant** is legitimately established initial authority with no required
parent grant. **Bootstrap** is the trusted procedure establishing that starting
authority, its intended administrator group, legitimate human membership and
explicit assignment. Ordinary callers cannot manufacture roots by omitting a
parent from grant content.

Our intended bootstrap starts with the maximum intended permissions and scope
within its authorized boundary, using a minimal coherent setup. Minimal setup
does not mean arbitrarily underpowered authority that cannot administer the
system. Registration precedes acceptance; partial setup must not expose partial
authority.

**Computed root coverage** follows the applicable registered application catalog.
The application has one shared catalog, not selectively adopted tenant releases.
Catalog growth does not silently add permissions to ordinary child revisions,
create membership or enlarge scopes.

**Representation: format pending.** Parent omission for trusted roots is agreed;
the complete computed-root source encoding and bootstrap trust payload are not.
No `*`, `is_root`, or catalog-source field is invented to fill that gap.

**Counterexample:** an ordinary derived grant with a missing parent is not a
bootstrap shortcut. A platform administrator publishing a permission does not
thereby receive tenant business access.

Sources: [bootstrap](../../docs/bootstrap-authority.md),
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
  }
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
  }
}
```

`proposed_dept` is the local name; `department_id` is the selected body field.
The body proposes a department. It does not establish the existing certificate's
current department. The application owns value validation and must establish or
enforce required relationships. No canonical `relationships` block is adopted.

Every declared input must be present at its declared source. A query parameter
cannot silently replace a missing body field. A broad `{}` grant does not make
the endpoint's required inputs optional. Extra source kinds, nested selection
and the complete policy schema remain pending.

**Counterexample:** the GET path claims Finance, but a later ID-only lookup
returns an Engineering certificate. Correct declaration and extraction have not
enforced the authorized boundary.

Source: [endpoint policy](../../docs/endpoint-policy-format.md).

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

**Counterexample:** using `dept` only in logging does not make it enforced
material. Conversely, a grant's `{}` does not invent a Finance constraint simply
because the endpoint has a department input. Request bindings and actual
authority constraints both need their proper meaning.

Sources: [request vocabulary](../../docs/authorization-vocabulary.md),
[endpoint gate](../../docs/endpoint-authorization.md).

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

This is an explanatory view, **not a canonical resolved-grant JSON contract**.
Its complete serialization and provenance evidence remain pending. The approved
grant/assignment JSON explains the inputs without inventing output fields.

**Complete route** means keeping the operation, boundary and all required
support together. Finance-write from one route and Engineering-read from another
cannot become Engineering-write. Independent alternatives remain independent;
the system does not intersect every unrelated grant into a globally narrowest
scope either. For an approved finite batch, different complete routes may cover
different items, with every item covered before protected effects.

**Counterexample:** copying G2's local `{"cert":"C17"}` into a new independent
grant drops inherited Finance and source dependencies. That is not resolution;
it is unauthorized authority expansion.

Sources: [vocabulary](../../docs/authorization-vocabulary.md),
[lineage](../../docs/authority-lineage.md),
[decisions](../../docs/decision-results.md), [batch coverage](../../docs/bulk-enforcement.md).

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
  "grant_ids": ["G2"]
}
```

`grant_ids` is a non-empty array of non-empty supporting grant identifiers,
not all the human's grants. These references are **supporting evidence** for
traceability, not new authority or a complete lineage snapshot. Returning them
does not require logging every request or supplying a returned scope field.

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
server-only diagnostic channel. Exact code catalogs, disclosure/value rules and
HTTP mappings remain pending.

Approved minimum evaluation-error result, with illustrative timeout code:

```json
{
  "version": "1",
  "error_code": "AUTH_SERVICE_TIMEOUT",
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

**Counterexample:** returning a timeout as `NO_AUTHORIZING_GRANT` claims a
completed policy conclusion the evaluator never reached. Returning allow with
an empty supporting list violates the approved minimum evidence contract.

Source: [result contracts and rationale](../../docs/decision-results.md).

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

No new service API, evaluator signature or validator-result JSON is adopted in
this reference. Those integration details belong to the implementation guide
and remaining contract work.

Sources: [responsibilities](../../docs/system-overview.md),
[Auth boundary gate](../../docs/authority-boundary-validation.md).

## What this reference deliberately does not invent

The core vocabulary needs neither a canonical target entity nor a prepared
decision state. Subteams and subgrants remain teams and grants. Orphaning and
effectiveness remain derived meanings, not extra live status fields. Scope
composition remains AND, not a new query language or JSON merge operation.

Membership/owner records, registration, full role publication, delegation
evidence, resolved-request/resolved-grant transports and computed-root encoding
are explicitly marked pending where discussed. Explaining them here does not
approve a missing schema. See the [pending register](../appendices/pending.md)
for the remaining choices, not a new question quota.

**Foundation takeaway:** a reader should now be able to distinguish what names
an operation, what bounds it, what defines authority, how someone receives it,
what keeps it supported, and how a request becomes a constrained effect—while
recognizing which JSON carries each agreed part.

[Return to the Foundations narrative](01-foundations.md) · [Continue to implementation](../implementation/05-canonical-model.md)
