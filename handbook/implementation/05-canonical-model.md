# 5. Canonical records and identity

[Contents](../README.md) · [Previous: theory](../theory/04-resolution-and-enforcement.md) · [Next](06-auth-service.md)

**Part II applies the repository's chosen model.** The requirements in these
chapters describe that model, not every possible authorization implementation.
The examples are approved core shapes and clearly identified excerpts, not a
complete deployable schema set.

The [Foundations reference](../theory/canonical-terms.md) now owns the term-by-term
explanation and representation guide. This chapter retains the original record
examples as an integration walkthrough: definitions need not wait until here,
and the repeated G1/A1 examples describe the same immutable content.

## From canonical concepts to implementation responsibilities

| Foundation concept | What an implementation must do with it |
|---|---|
| Identity and tenant context | Establish trustworthy context; do not authorize from an unverified body or decoded token alone. |
| Grant control and revision | Associate live grant-wide state with the exact immutable content an assignment adopted. |
| Assignment and membership | Establish the recipient route; a stored definition is not access and a group assignment is not human membership. |
| Role revision | Load the explicitly referenced permission bundle, not whichever role content is newest now. |
| Parent/team lineage | Establish actual eligible support, retain inherited boundaries and validate relevant team ceilings. |
| Resolved views and result | Keep computed authority dependent and request-bound; do not persist it as an independent grant. |

For example, loading A1 must lead to G1 revision 2 and G1's live control, not
merely the newest G1 record. Loading G1's parent must establish actual supporting
lineage, not search for any broader holding elsewhere. If A1's recipient is a
group, the implementation must establish the human's applicable membership too.
Only then can that route contribute to evaluation of the declared permission
and material. These are logical responsibilities, not a prescribed sequence of
network calls, database tables or an invented evaluator interface.

## The framework's deliberate choices

The model uses positive, complete authority routes. Ordinary grants are derived
from valid parent support. Each child selects permitted operations and adds
scope constraints with AND. A reusable grant does not contain its recipient;
an assignment supplies that relationship and adopts a specific grant revision.

Teams and groups mean the same thing. Authorization membership is owned by Auth
and contains humans, not agents or service accounts. Prefer group-based human
access, while allowing legitimate direct assignments. An application's own
department or employee grouping does not silently become Auth membership; any
necessary synchronization is an explicit integration responsibility.

Agents and service accounts are human-dependent proxies in this model. Their
effective authority must remain within both the human's current applicable
authority and their delegation limits. V1 supports direct human-to-proxy
delegation, not proxy-to-proxy authority chains.

These restrictions keep the chosen model explicit. They are not claims that a
generic authorization theory forbids machine principals, nested membership or
different permission-combination mechanisms.

## Permission names

Use a namespaced noun followed by a verb:

```text
<app>:<domain>[:<subdomain-or-resource>...]::<verb>

hrms:employee:certificate::read
hrms:employee:certificate::write
accounting:ledger:entry::post
codehost:repository:branch::write
```

The namespace can have different depths. A colon separates noun segments;
the double colon separates the operation. Department and user identifiers are
not encoded into each operation name. A shared prefix does not confer descendant
permissions. V1 has no permission aliases or wildcard permission names.

Register names before grants use them. An existing identifier must not acquire
a materially different authorization meaning, even after retirement. Descriptive
labels may change without changing the protected operation. Detailed character
validation and the full catalog lifecycle remain [pending](../appendices/pending.md).

## Scope format and meaning

Scope is a required flat object with registered keys and non-empty string values.
Entries combine with AND. The application defines each key's boundary meaning;
there are no universal department, repository or certificate types built into
the grant merely because examples use those names.

For example, the scope value `{"dept":"FIN","user":"$self"}` means the
defined Finance boundary AND the authorizing human's self boundary, where the
application supports that combination. `$self` is a reserved token for that
human. It does not mean the group receiving the assignment or automatically the
agent sending the request.

Reject missing or null scope, unsupported keys/tokens, duplicate keys, empty or
non-string values, arrays, nested objects and wildcard/query operators. Do not
repair invalid input by dropping a restriction. The required validation follows
the registered definitions without making Auth an application database interpreter.

The explicit empty object `{}` adds no local restriction. At an otherwise
legitimate tenant root it adds no narrower boundary inside that tenant. On a
derived grant it retains the inherited boundary. It never discards a parent
constraint. Tenant is the trusted implied outer context, not an ordinary scope
entry that callers can replace.

Keep scope predicates intact during resolution. Merging parent and child objects
with “last value wins” is incorrect when both constrain the same key.

## Three records, three responsibilities

The following example supplies Finance certificate read/write to Team1. Assume
registered permissions and scope definitions, valid G0 support, and successful
administrative/source-boundary validation. Revision 2 is latest when A1 is created
or explicitly upgraded. These assumptions are required checks, not bootstrap
shortcuts.

### Grant identity and live control

```json
{
  "version": "1",
  "id": "G1",
  "status": "enabled"
}
```

This live control applies across G1's revisions and assignments. It supplies no
authority without content and valid assignment/support. A grant-wide disable
cannot be overridden by an enabled assignment.

### Immutable authority content

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

The pair `(grant_id, revision)` identifies published content. Parent, permission
source, scope and any local validity window belong to that immutable content.
Changing them requires new content, not editing an already adopted revision.

`parent_grant_id` is the declared grant-lineage link. Actual supporting assignments
and their adopted revisions still have to be established. No independent
`parent_assignment_id` or `parent_grant_revision` field is adopted.

### Assignment and explicit adoption

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

The assignment identifies who receives the content and which revision it has
adopted. Publishing revision 3 does not update A1. Enabled means eligible for
evaluation, not an unconditional allow. The three-record separation is logical;
it does not mandate exactly three database tables.

## Roles are an alternative permission source

A grant revision either contains explicit `permissions` or contains both
`role_id` and `role_revision`. Do not combine the two forms or accept half a role
reference. There is no latest-role fallback.

This independent example assumes R-CERTIFICATE-READER revision 1 contains the
certificate-read permission and fits valid G0 support:

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

The role revision and grant revision identify different content. Publishing a
new role revision neither rewrites this grant nor changes its assignments.
Explicit adoption and boundary validation prevent a reusable bundle from
silently changing recipients' authority.

## Identity is not assignment

A term such as **principal** is useful as general identity vocabulary, but it is
not an additional canonical record or field here. Specify the actual role:
actor/caller, authorizing human, or assignment recipient. The JWT subject has the
particular meaning explained below. These identities can be related without
being interchangeable.

A request's identity describes who is acting and whose human authority anchors
the request. It is not the recipient of a grant or a list of entitlements.

Approved identity block for Agent A-17 acting for Vinay:

```json
{
  "version": "1",
  "actor": {"type": "agent", "id": "A-17"},
  "human_id": "U-17"
}
```

Actor types are `user`, `agent` and `service_account`; `user` denotes a human.
For a direct human, actor ID and human ID must agree. For a proxy, trusted
evidence must establish the association and delegation. Submitting this JSON
does not prove the relationship.

The chosen JWT profile keeps the human subject and carries the same block:

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

This is an identity-related payload excerpt, **not a complete JWT or token
validation procedure**. Issuer, audience, expiry and cryptographic verification
remain necessary integration concerns. The profile requires `sub` to equal
`identity.human_id`; mismatches cannot be repaired by choosing the more powerful
identity. Keeping the repeated human ID makes the identity block self-contained
outside JWT as well.

Compatibility of the human subject does not establish safe delegation in legacy
software. Protected proxy operations must still receive the applicable evaluator
checks; a consumer that ignores the proxy and uses unrestricted human authority
would bypass the model. No separate duplicate `agent_id` claim is selected here.

## Format versions are not authority revisions

Every published contract has a supported top-level string `version`, initially
`"1"`. A missing, malformed or unsupported version is rejected, not guessed.
`revision`, `grant_revision` and `role_revision` select authority content; they
do not replace the format version.

**Pending:** full field/default/timestamp schemas, computed-root representation,
direct-human support discovery, cross-recipient `$self` binding and complete
identity/delegation transports are not settled by these core examples. Do not
fill the gaps by treating a parentless ordinary grant as a root or by copying
recipient-relative scope text without preserving its meaning.

**Sources:** [permission](../../docs/permission-model.md),
[scope](../../docs/scope-model.md), [core records](../../docs/grant-revision-format.md),
[role variant](../../docs/role-grant-contract.md), [identity](../../docs/identity-context.md),
[JWT mapping](../../docs/jwt-identity-mapping.md).

[Next: registration, bootstrap and Auth administration](06-auth-service.md)
