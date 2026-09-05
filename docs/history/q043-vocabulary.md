# Deprecated vocabulary — exact wording before Q-043

### docs/scope-model.md — original line 5

```text
Latest: Q-039 / REGISTRATION-001 now agrees application registration of supported
```

### docs/scope-model.md — original line 15

```text
governance, representation, and validation mechanics remain open under Q-039.
```

This is historical wording, not current handbook vocabulary. The user directed
removing “target” entirely from the authorization vocabulary and approved the
tenability check in Q-043. Changed original lines are preserved verbatim below
with their original file and line number; unchanged surrounding text remains
in those files. Vocabulary changes do not silently approve old proposals.
The previous diagram is preserved as [deprecated SVG](../assets/history/authorization-system-pre-q043.svg).

## docs/scope-model.md

### Original line 1

```text
# Working handbook chapter: scope and target
```

### Original line 12

```text
explicit supported target relationships, with endpoint-specific trusted fact
```

### Original line 19

```text
> Scope is a boundary selector that identifies a set of targets within the
```

### Original line 20

```text
> enclosing tenant. A target satisfies the scope when it falls within the
```

### Original line 24

```text
binds recipient, permission(s), and scope with validity/conditions. Target
```

### Original line 40

```text
value selects that boundary. The above illustrates targets inside dept-1 AND
```

### Original line 87

```text
For the existing motivating meanings, the target must be both inside dept-1
```

### Original line 89

```text
target boundary; it is not the recipient of the grant. The scope does not grant
```

### Original line 135

```text
The application defines a scope key's boundary meaning and the target
```

### Original line 155

```text
An endpoint supplies established target facts through its declared binding.
```

### Original line 163

```text
the application's defined target relationships and trusted facts. The embedded
```

### Original line 217

```text
target. Finance is a reference used to determine whether that payslip falls
```

### Original line 281

```text
## SCOPE-003 / Q-027 — scope-owned target selection, agreed
```

### Original line 292

```text
In the semantic sense, scope acts as a selector: it describes which targets are
```

### Original line 304

```text
| Scope | Own the meaning of target selection and the valid inputs needed to express it. |
```

### Original line 306

```text
| Request/target | Identify the operation sought and the resource or proposed resource it acts on; representation remains open. |
```

### Original line 321

```text
| Shared semantic concept | Self | The human anchor is shared under SELF-001, but the target relationship must be explicitly defined for each compatible resource. |
```

### Original line 345

```text
composition, target matching, and scope containment remain open.
```

### Original line 351

```text
scope may be used only where its target relationship has been defined, never
```

### Original line 457

```text
complete request includes the department-qualified target; if request and
```

### Original line 493

```text
SCOPE-005 proposes treating scope as declarative target selection that the
```

### Original line 498

```text
Returned targets = requested targets intersect authorized targets
```

## docs/grant-model.md

### Original line 5

```text
default. See [scope and target](scope-model.md). Earlier grant examples retain
```

### Original line 46

```text
references, and targets must resolve inside that tenant boundary.
```

### Original line 220

```text
target may still be needed. It is also not a completed allow decision.
```

### Original line 311

```text
creation, the authorization target is the proposed grant; it need not already
```

### Original line 313

```text
create-target and scope-evaluation contracts remain stage 6/7 work.
```

### Original line 328

```text
| grant_selector | The target being selected is a grant. | Target resource type may already follow from the permission/operation declaration; another discriminator may be redundant. |
```

### Original line 329

```text
| recipient inside scope | Restrict the recipient of the proposed target grant, not the administrator receiving authority. | Recipient is an existing target-grant attribute; its use in a scope expression does not justify a new special nested field. |
```

### Original line 330

```text
| permissions_subset_of | Prevent the target grant from assigning business permissions outside the administrator's assignable set. | This describes a set relation; a dedicated field has not been shown necessary. Role references and current-role expansion also need treatment. |
```

### Original line 331

```text
| scope_within | Prevent the target grant from reaching resources beyond the administrator's assignable reach. | Semantic containment between scopes is not mere field or label comparison. Relationship/self scopes and future changes make this a substantive open problem, not a solved operator. |
```

### Original line 339

```text
SCOPE-001 already allows target selection by references, attributes, and
```

### Original line 348

```text
~~Continue in [scope and target](scope-model.md), SCOPE-002 / Q-026, then return here.~~
```

### Original line 353

```text
created or changed as target. Then revisit ADMIN-004's assignable bounds.
```

### Original line 380

```text
assume a hierarchy. A narrower target selection is acceptable only when shown
```

## docs/grant-examples.md

### Original line 68

```text
| Requested action | Target facts | Result supported by these grants |
```

### Original line 73

```text
| Revoke | Certificate in Engineering, T-1 | Denied: G-2 does not reach the target |
```

### Original line 237

```text
established, and the target may still require an application lookup.
```

### Original line 315

```text
The user scope key is not the group recipient; it selects a target boundary.
```

### Original line 321

```text
validation and target facts are required; this example is not a runtime test.
```

## docs/grant-format.md

### Original line 110

```text
grant or proof that target-specific authorization is complete. This retains
```

## docs/use-case-examples.md

### Original line 27

```text
  alone does not establish an existing target's relationships.
```

### Original line 222

```text
| Declared inputs versus actual target facts | UC-GIT-001, UC-TICKET-004. |
```

## docs/discussion-tree.md

### Original line 53

```text
- Current as of Q-038: application-owned scope meanings, supported target
```

### Original line 63

```text
  explicitly defined scope-key meanings and supported target relationships.
```

### Original line 71

```text
- Active branch: **6. Scope and target → scope-key definitions and governance**.
```

### Original line 80

```text
  The detailed chapter is [scope and target](scope-model.md).
```

### Original line 180

```text
│   ├── Resource type versus instance; target; context [open]
```

### Original line 204

```text
│   │   └── Ordinary model with grant-as-target [active: Q-043]
```

### Original line 211

```text
├── 6. Scope and target [formerly active; remaining detail parked after Q-042]
```

### Original line 220

```text
│   │   ├── Explicit key meanings and target mappings [formerly active; settled: Q-038; refines SCOPE-004]
```

### Original line 230

```text
│   └── Targets for read/list/create/update/move; relationship timing [open]
```

### Original line 306

```text
Q-027 → scope-owned target selection and grant binding (SCOPE-003 agreed).
```

### Original line 317

```text
Q-038 → scope-key meanings and explicit supported target relationships (open; return to stage 6).
```

### Original line 327

```text
Q-043 → canonical grant administration with grant-as-target (open; ADMIN-005 before ADMIN-004 bounds).
```

## docs/application-registration.md

### Original line 39

```text
| A scope key is registered for the relevant application. | What target relationship that boundary key represents. |
```

### Original line 40

```text
| Scope follows canonical syntax and declared constraints. | Trusted facts establishing whether a particular target is inside the selected boundary. |
```

### Original line 121

```text
endpoint-owned gate must still establish supported target relationships and
```

### Original line 141

```text
In either mode, the endpoint-owned gate must establish that the actual target
```

## docs/handbook-roadmap.md

### Original line 27

```text
| 3. Core vocabulary | Actor, principal, subject, identity, tenant membership, group, resource type, resource instance, operation, permission, role, assignment, grant, policy, scope, target, context, and authority. | A glossary and relationship map, including deliberate distinctions and synonyms. |
```

### Original line 30

```text
| 6. Scope and target | Scope type, descriptor, referenced resource, resolved scope, effective reach, static selectors, dynamic relationships, exact and subtree reach, multiple dimensions, empty and missing scope; targets for reads, lists, creation, updates, and moves. | Scope semantics, target semantics, and containment rules. |
```

### Original line 31

```text
| 7. Requests and resolution | Incoming application request versus authorization request; resolved request, grant bundle, resolved grants, resolved target; inputs, outputs, owners, sources of facts, timing, and failures for each resolution operation. | Explicit contracts and a complete resolution flow. |
```

### Original line 34

```text
| 10. Challenge the model | HRMS and repositories end to end; conflicting grants, stale membership, cross-tenant targets, missing relationships, delegated agents, partial bulk access; verify relevant implementation and external references. | An agreed scenario suite and an evidence-backed implementation gap register. |
```

### Original line 93

```text
- Distinguish a caller-selected target identifier from established target facts;
```

### Original line 120

```text
| SCOPE-006 | AGREED | Scope is a boundary selector that identifies a set of targets within the enclosing tenant. A target satisfies the scope when it falls within the selected boundary. Permission covers the operation; a valid applicable grant must cover both operation and target. | User defined scope as a boundary selector, affirmed the permission distinction, and requested a canonical definition. Q-030/Q-031 capture the pause-period discussion. No query language or new fields follow from this definition. |
```

### Original line 125

```text
| INPUT-001 | AGREED | The server-owned method/route declaration maps the endpoint to an action and required permission, and identifies which path and body parameters contribute authorization request inputs. Combine those inputs with verified identity/tenant context, Auth-owned authority, and required application facts at the endpoint-owned gate. | User clarified material is usually method-to-action plus path and identified body parameters. Selected inputs do not prove existing target relationships. HTTP method alone does not determine every operation's permission. Exact binding schema and current-versus-proposed state semantics remain open. |
```

### Original line 126

```text
| SCOPE-005 | DEPRECATED | Scope describes authorized target selection; where supported the application may translate it into a safe query restriction, intersected with the requested selection. A fully determined restriction can be enforced by a middleware-complete endpoint; endpoint-completion obtains facts and completes evaluation under its fixed contract. | User asks for route/JSON examples and proposes selector/query-like scope. Q-029 tests the flow in scope-model.md, reusing agreed CONTRACT-002/005. Does not adopt arbitrary SQL in grants, a query grammar, universal query compilation, or dynamic endpoint modes. |
```

### Original line 127

```text
| SCOPE-003 | AGREED | Scope owns target-selection semantics; a grant binds its use to recipient and permissions without defining its domain meaning. Application-defined concepts such as department are optional; shared concepts such as self require explicit application/resource relationships. Authorization validates and evaluates scope through its definition. | User answered Q-027: "agree." Establishes responsibility separation, not a new entity, field, fixed catalog, general expression grammar, independent storage requirement, or mutable scope-reference model. SCOPE-002's representation remains open. See scope-model.md. |
```

### Original line 128

```text
| SCOPE-004 | ~~PROPOSED~~ PARTIALLY AGREED | A scope may be used for a resource only where its target-selection meaning for that resource is explicitly defined. Do not infer compatibility from matching field names or use an unrelated relationship. Validate scope/resource compatibility when binding permissions; unsupported or unresolved use must not establish authority. | Q-028 tests department scope on payslips, certificates, and repositories. Supports reusable meanings with explicit resource mappings, not a mandatory per-resource scope type. Multi-permission compatibility, role evolution, declaration location, and failure handling require detailed follow-up. Q-038 now agrees the application-owned meaning, explicit supported target relationships, endpoint fact bindings, and no implicit unrestricted access for unsupported relationships. The earlier proposed grant-binding validation point and detailed compatibility mechanisms remain open; Q-039 addresses shared definitions. |
```

### Original line 131

```text
| ADMIN-005 | PROPOSED | Use the same canonical grant binding for administration: recipient is the administrator user/group, permission is an operation on Auth resources such as grant creation, and scope selects the grant targets that may be administered. How to express target-grant recipient, permission, and scope limits must be justified through the shared scope model, not silently introduced as new syntax. | User challenged both Q-023's standalone format and Q-024's new nested fields. G-11 syntax is withdrawn under PROCESS-005. The unified model remains the working direction; exact scope grammar, bootstrap, containment, and lifecycle rules are not finalized. |
```

### Original line 142

```text
| PRINCIPLE-001 | AGREED | A protected operation may proceed only when the system establishes valid authority covering the requested action and target within the applicable tenant boundary. If required information is missing, invalid, or unavailable and authority cannot be established, the operation must not proceed. | User agreed on 2026-09-05: "yes, this is right." Authentication or knowledge of a resource ID does not by itself establish access. Failure to establish authority must remain distinguishable from a completed evaluation that denies access. Exact grant-combination, freshness, and response rules remain open. |
```

### Original line 144

```text
| TENANT-001 | AGREED | Tenant is the enclosing authorization boundary and is implicit in a grant evaluated within trusted tenant context. Recipients, groups, scope references, and targets resolve within that boundary; grants cannot select or widen it. The authorization system must preserve the binding even when the grant payload omits a tenant field. | User instructed: "Tenant should be more implicit as this is the outer most boundry is implied." Grant examples omit repeated tenant IDs. Tenant-wide scope means the enclosing tenant. Storage, cache, and transport representations remain to be designed; implicit context is not permission to drop isolation or trust a caller's tenant assertion. |
```

### Original line 149

```text
| ENFORCEMENT-001 | DEPRECATED | A prepared context is not an allow decision. For a prepared request, the endpoint may perform only the internal, tenant/request-constrained work needed to establish required facts before completing authorization. Missing facts or an omitted required check cannot yield protected output or a business mutation. A completed middleware allow still requires enforcement of its action, target, and restrictions by the application. | Accepted as part of Q-010 with ARCH-002's refinement. Bind facts and the decision to actual use; avoid changing relevant facts between checking and use. Query predicates may combine final evaluation and access for collection operations; exact enforcement contracts remain open. |
```

### Original line 153

```text
| CONTRACT-001 | DEPRECATED | Each protected operation declares its permission mapping, the meaning of its request bindings, and how the required authorization facts or enforceable predicates are obtained. Applicable scope/delegation evaluators report requirements still unresolved. Middleware relies on those contracts to distinguish proven conditions from pending ones; it cannot infer arbitrary business facts from URL structure. | Explicit contracts allow predictable handling of supported operations. Undeclared or unsupported checks fail closed rather than being silently interpreted as satisfied. This is a contract proposal, not a finalized schema or a requirement to enumerate all target IDs. |
```

### Original line 177

```text
| SCOPE-001 | AGREED | A scope describes the set of resources over which an assigned permission may apply. It may express explicit selections or selectors based on trusted resource attributes and relationships; it need not enumerate IDs or follow one hierarchy. A target is what the request seeks to act on, checked against that reach. | User answered Q-009: "Yes agree to this." Illustrations: exact payslip P-17; payslips owned by the principal's employee; payslips associated with Finance. Final selector grammar, temporal relationship semantics, and targets for collection/create operations remain open. A scope description is not itself an allow decision. |
```

### Original line 183

```text
| RESOLUTION-003 | AGREED | Expand a role reference into its current permission set, producing the same permission-set grant form used to evaluate explicitly listed permissions. Preserve the original grant identity, scope, recipient, validity, conditions, and source provenance, including the role definition used. Expansion is a computed view of the same grant, not a new assignment, and does not merge unrelated grants. | User agreed after clarification that the earlier JSON represented the existing G-6 after role lookup. GRANT-EX-005 retains G-6 visibly. Expanded does not mean fully resolved: application relationships or target facts may still be missing. Exact schema, role-revision encoding, membership expansion, and final resolved-grant representation remain open. |
```

### Original line 223

```text
| Q-027 | ANSWERED | Yes: scope owns target-selection meaning; grant binds its use without defining it. | SCOPE-003 agreed; storage, scope catalog, grammar, and representation remain open. |
```

### Original line 224

```text
| Q-028 | ~~OPEN~~ ANSWERED BY Q-038 | Should a scope be usable only with resources for which its target-selection meaning has been explicitly defined, rather than inferred from a shared name or field? | SCOPE-004 proposed; distinguish scope compatibility from authorization to perform the operation. Later Q-038 agrees this principle; it does not settle all of SCOPE-004's validation mechanisms. |
```

### Original line 234

```text
| Q-038 | ~~OPEN~~ ANSWERED | Should an application define a scope key's boundary meaning and supported target relationships explicitly, with endpoints binding trusted material to that meaning rather than inventing a meaning or inferring it from matching field names? | Refines existing SCOPE-003 and proposed SCOPE-004 under ARCH-004/005. Use dept across payslips, certificates, and unsupported repositories as the example. No definition-registry format or new grant fields proposed. User agreed. Application-owned meanings and endpoint-specific trusted fact bindings are now settled; illustrative department relationships are not a universal application catalog. |
```

### Original line 235

```text
| Q-039 | ~~OPEN~~ ANSWERED AS REFINED | Should application-owned scope definitions be shared as one contract for grant validation and request evaluation, so neither side independently invents accepted keys or supported target meanings? | Historical proposal: Proposed governance principle only. It does not require Auth to query application data, copy domain records, or execute application code. Definition transport, storage, versioning, concrete-reference checks, and exact validation timing remain open. No new grant fields proposed. Subsequent refinement and approval: applications register both supported scopes and permissions; Auth validates grants before accepting, without application-domain interpretation. REGISTRATION-001 records this approved principle; compatibility declarations and registration syntax remain open. |
```

### Original line 237

```text
| Q-041 | ~~OPEN~~ ANSWERED AS REFINED | Historical question: If relationship metadata is omitted, should Auth perform its other registration/canonical/issuance checks without claiming compatibility, leaving supported target relationships and scope satisfaction to the endpoint-owned gate; when metadata is supplied, additionally validate against it? | Historical rationale: Recommended interpretation of optional declarations, not yet agreed. Optionality granularity, absent versus empty, partial/invalid metadata, multi-key combinations, and change semantics remain open. Runtime unsupported relationships cannot establish authority under Q-038. User instead required an upfront declaration and approved the explicit application-level choice: enabled makes relationship validation mandatory for all grants, including role-based and existing grants. REGISTRATION-003 records the refined approval; no omission-based mode inference is adopted. |
```

### Original line 239

```text
| Q-043 | OPEN | Should grant administration use the ordinary grant/permission/scope model, with the grant being created or changed as the authorization target? | Resume ADMIN-005 before ADMIN-004's exact assignable bounds. Example: Maya can issue Finance payslip-read grants to payroll-readers without thereby gaining payslip access. No new admin fields, scope operators, or containment mechanism proposed. |
```

## docs/system-overview.md

### Original line 60

```text
operations, scope-key meanings for targets, trusted fact sources, additional
```

### Original line 117

```text
| Application fact source | Establish target attributes and domain relationships needed by the operation and scope, through bounded internal access. |
```

### Original line 164

```text
of permission before fetching unnecessary target metadata. Material collection
```

## docs/endpoint-authorization.md

### Original line 39

```text
fetching unnecessary target metadata. Shared role expansion and authority loading
```

### Original line 56

```text
| Application, when required | Trusted target attributes and relationships needed to evaluate the selected boundaries. |
```

### Original line 62

```text
target/change, subject to validation and any required current-state facts.
```

### Original line 79

```text
Required application facts establish actual target membership, rather than
```

### Original line 89

```text
The target must be both within dept-1 and
```

## docs/reconciliation.md

### Original line 86

```text
| Scope meaning and ownership | SCOPE-003/006 | [Scope and target](scope-model.md): boundary selector; scope owns its meaning, grant binds its use. |
```

### Original line 87

```text
| Application meanings and target compatibility | Q-038; conceptual core of SCOPE-004 | Applications define supported boundary relationships; endpoints bind trusted facts rather than infer meanings from matching names. Grant-validation mechanics remain open. |
```

### Original line 129

```text
| Q-028 and scope chapter introductions left explicit target relationships open. | Q-038 agrees application-defined meanings and supported target relationships. SCOPE-004 remains only partially agreed because validation mechanisms are still open. |
```

### Original line 155

```text
  Q-038 settles application ownership and explicit target meanings. Shared
```

## docs/handbook.md

### Original line 43

```text
| [Scope and target](scope-model.md) | Canonical boundary-selector definition and v1 key-value format, AND within scope, alternatives through grants, and empty/invalid scope rules. |
```

## Additional preserved lines — docs/grant-model.md

### Original line 352

```text
with Q-043: ordinary grant/permission/scope concepts, with the grant being
```

### Original line 356

```text
the canonical administration model remains a proposal pending Q-043.
```

## Additional preserved lines — docs/scope-model.md

### Original line 21

```text
> selected boundary.
```

### Original line 25

```text
membership satisfies the scope check, not the entire authorization decision.
```
