# Identity and authority glossary — consolidated meanings

This completes the documentary glossary requested by HC-03-04 using approved
Q-043, Q-085–Q-087, and Q-090 onward. It adds no identity entity, JSON field,
membership API, authentication protocol, or new authority rule. Where older
prose uses an ambiguous umbrella word, the precise approved identity roles below
take precedence. Full trust/tenant mappings remain HC-03-05, not this glossary.

## Identity terms

| Term | Meaning in this handbook | Important distinction / source |
|---|---|---|
| Identity | Established information about who is acting and, for dependent access, whose human authority supports the request. | Matching submitted JSON is not proof. [Q-085/086](identity-context.md). |
| Actor / caller | The actual caller, represented by typed `identity.actor`: `user`, `agent`, or `service_account`. | The actor of a proxy call is not silently replaced by its human. [Q-086](identity-context.md). |
| Human / user | The human identity denoted by actor type `user` in this model. | “User” in the approved identity/recipient examples does not mean an independent service account. [Identity](identity-context.md), [groups](groups-and-membership.md). |
| Authorizing human / human anchor | The human identified by `identity.human_id`, whose applicable authority bounds the request and anchors `$self`. | It is required for direct and proxy calls. Trusted evidence binds a proxy to this human. [Q-085](proxy-attribution.md). |
| Principal | General identity terminology in earlier prose, not an additional canonical entity or approved wire field. | When specifying a contract, say **actor**, **authorizing human**, or **assignment recipient**, according to the role meant. Do not use this umbrella term to merge those roles. [Q-086](identity-context.md), [Q-090](grant-assignments.md). |
| Subject | In the agreed JWT profile, the human identified by `sub`, equal to `identity.human_id`. | For a proxy, this is not its actor ID. This is our profile's mapping, not a rule imposed on arbitrary third-party tokens. [Q-087-B](jwt-identity-mapping.md). |
| Agent / service account | An automated actor with human-dependent authority, constrained by applicable human support and delegation limits. | Neither is a first-class group member or an independent authority source. [AUTHORITY-002](proxy-attribution.md). |
| Proxy / delegation | Acting through supported human authority under additional delegation restrictions. Delegation names the authority dependency, not merely an account's creator. | Exact delegation reference and chain schemas remain open. [Attribution](proxy-attribution.md), [lifecycle](delegation-lifecycle.md). |
| Tenant context | The trusted enclosing authorization boundary. | Not an ordinary grant scope key; a caller-provided tenant reference does not establish it. Its complete identity mapping remains open. [Q-086](identity-context.md). |
| Tenant membership | Any application/identity-system relationship used to establish a human's tenant context; its full shared representation is not yet settled here. | Do not infer administrator rights or authorization-group membership from tenant association. [Identity](identity-context.md), [Q-115](bootstrap-initial-assignment.md). |
| Self / `$self` | The authorizing human under the registered scope key's supported application relationship. | Not the group, group owner, grant creator, or automated actor. [SELF-001](scope-model.md). |

“Principal” and “tenant membership” above explain the limits of existing wording;
they do not introduce new canonical storage types. The current shared identity
fields remain exactly the approved actor/human block.

## Membership, ownership, and authority terms

| Term | Meaning | Not equivalent to |
|---|---|---|
| Team / group | Synonyms for an Auth-owned collection of human members receiving its valid assigned authority. | A department scope value or automatic ancestor membership. |
| Membership | An explicit human-to-group relationship supporting that human's group-derived access. | Group ownership, grant assignment, or inherited membership through a parent team. |
| Owner / administrator | A human exercising explicit administrative authority for the relevant operation and boundary. | Automatic business access, automatic assignment power, or an authority source merely because they created a team. |
| Recipient | The human or group identified by an assignment's `recipient`. | The requester, the authorizing human, or a field on a reusable grant definition. |
| Grant | A reusable authority definition with permissions or an adopted role revision, scope, and applicable restrictions/dependencies. | Access by itself; a recipient-bearing record; an allow result. |
| Assignment | A recipient binding to a grant and an explicitly adopted revision, with its own live enablement control. | Grant-wide enablement or proof that every dependency is effective. |
| Role | A reusable permission bundle; role revisions are immutable and explicitly selected by grants. | Group membership, scope reach, or a recipient. |
| Grant revision | Immutable authority content identified by grant ID and revision. | Contract `version`, live grant status, or automatic adoption by existing assignments. |
| Resolved grant | A computed evaluation view of an existing authority route, retaining its source bindings and restrictions. | A copied independent grant, a new assignment, or an allow. |
| Authority route / lineage | The actual assignment, adopted content, and required upstream/team/membership/delegation support through which authority is available. | A parent definition merely existing somewhere, or a permanent dependency on whoever issued the grant. |
| Subteam / subgroup | A team relative to its parent, whose total authority remains within that parent's authority. | A new entity type or transitive human membership. |
| Subgrant | A grant relative to its parent grant, with a permission subset and accumulated AND scope. | An independent copy of the parent's authority. |
| Orphan grant/route | A grant's affected lineage whose required parent support no longer exists. | An unassigned definition, a legitimate root, merely disabled support, or failure to fetch evidence. |
| Effective authority | What valid current routes can actually supply after mandatory restrictions and dependencies are applied. | Stored `enabled` flags alone, all permissions mentioned in a token, or the globally narrowest unrelated grant. |

Sources and rationale: [groups](groups-and-membership.md), [assignment separation](grant-assignments.md),
[lineage/orphans](authority-lineage.md), [ownership](ownership-lineage.md),
[revisions](grant-revisions.md), [core JSON](grant-revision-format.md).

## Request-side terms

| Term | Meaning and distinction |
|---|---|
| Permission | The registered operation name; says what can be done, not who or how far. |
| Scope / boundary | The selector restricting granted reach inside the tenant; not a universal application object or arbitrary query language. |
| Endpoint policy | The server-owned declaration of one required permission and selected inputs/sources for a method/route. |
| Request | The operation sought under that declaration. |
| Request input | A declared value obtained from its specified source; presence does not prove its relationship to application data. |
| Request material | Inputs, verified context, and trusted facts sufficient for authorization and its binding to actual execution. |
| Resolved request | The evaluation-ready view with the required meaning/material established; not a completed authorization decision. |
| Resolution | Establishing the applicable request and authority views while preserving dependencies and non-amplification. |
| Evaluation | Deciding whether a complete applicable authority route covers the required operation and boundary. |
| Enforcement | Preventing protected execution without sufficient authorization and keeping actual data/effects within evaluated boundaries. |
| Allow / deny / evaluation error | Allow and deny are completed decisions. Error means required evaluation could not finish; both deny and error prevent protected execution. |

Sources: [request vocabulary](authorization-vocabulary.md), [endpoint gate](endpoint-authorization.md),
[permission](permission-model.md), [scope](scope-model.md), [results](decision-results.md).
No canonical “target” entity is introduced.

## One example keeping identities separate

```text
Agent A-17                    actual actor
    │ trusted delegation
    ▼
Vinay U-17                    human anchor; JWT subject; $self
    │ explicit membership
    ▼
Employees                     group; assignment recipient
    │ assignment adopts grant revision
    ▼
Self-read grant               reusable authority definition
```

Assume a valid supported self-read route and delegation. A-17 is not an Employees
member. Employees is not the human anchor. A separate administrator who created
the assignment is not automatically the request's subject or continuing support.
For Vinay's direct request, actor ID and human ID coincide. For A-17's request
they differ. A body value claiming another human cannot change either established
identity. These distinctions prevent applying the wrong person's authority while
preserving the existing recipient-free grant model.

This glossary closes terminology consolidation only. Trusted tenant mapping,
legacy-token normalization, complete delegation/membership contracts, and deployed
identity verification remain explicit work in the [closure checklist](v1-closure.md).
