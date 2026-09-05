# Authorization handbook — agreed roadmap

Roadmap approved: 2026-09-05.

Use the [discussion tree](discussion-tree.md) to navigate branches, conclusions,
open questions, and the current return point. This file holds the decision log.
Read the [working handbook](handbook.md) for the detailed chapters and examples.

## Purpose

Develop a shared foundation for authorization across AgentLabs: precise concepts,
agreed principles, and rules that implementations can follow. Refine the model
through discussion and worked examples, documenting decisions as they are made.

The user approved this roadmap and asked the assistant to steer the discussions.
Approval of the roadmap does not settle the individual authorization decisions.

## Work stages

These are work stages, not a fixed number of meetings. Draft incrementally;
publication in stage 11 consolidates the decisions made throughout the process.

| Step | Discuss and settle | Finished result |
|---|---|---|
| 1. Purpose and authority | Audience, products governed, mandatory shared rules, application-owned choices, and responsibility boundaries. | A short charter and boundary of responsibility. |
| 2. Principles | Tenant isolation, default denial, least privilege, trusted information, explicit authority, delegation limits, enforcement, and permitted exceptions. | Numbered principles with examples of compliance and violation. |
| 3. Core vocabulary | Actor, principal, subject, identity, tenant membership, group, resource type, resource instance, operation, permission, role, assignment, grant, policy, scope, target, context, and authority. | A glossary and relationship map, including deliberate distinctions and synonyms. |
| 4. Permissions and operations | Meaning, naming, operation-to-permission mapping, multiple required permissions, resource hierarchy, wildcards, and compatibility as capabilities evolve. | Permission semantics and catalog rules. |
| 5. Grants and their lifecycle | Direct, role, and group assignment; who may grant what; validity; provenance; role changes; revocation; delegation; assignment versus grant. | An assignment model and rules for creating, changing, and withdrawing authority. |
| 6. Scope and target | Scope type, descriptor, referenced resource, resolved scope, effective reach, static selectors, dynamic relationships, exact and subtree reach, multiple dimensions, empty and missing scope; targets for reads, lists, creation, updates, and moves. | Scope semantics, target semantics, and containment rules. |
| 7. Requests and resolution | Incoming application request versus authorization request; resolved request, grant bundle, resolved grants, resolved target; inputs, outputs, owners, sources of facts, timing, and failures for each resolution operation. | Explicit contracts and a complete resolution flow. |
| 8. Decision semantics | Combining permissions, scopes, grants, tenant boundaries, conditions, and delegation; union versus intersection; deny precedence; missing information; contributing-grant provenance. | Decision rules with unambiguous examples. |
| 9. Enforcement and change over time | Row and field restrictions; list, count, export, bulk and mutation behavior; caches; revocation freshness; resource moves; relationship changes; audit evidence. | Enforcement obligations and consistency guarantees. |
| 10. Challenge the model | HRMS and repositories end to end; conflicting grants, stale membership, cross-tenant targets, missing relationships, delegated agents, partial bulk access; verify relevant implementation and external references. | An agreed scenario suite and an evidence-backed implementation gap register. |
| 11. Publish the foundation | Reconcile terminology, rules, examples, and contracts; distinguish required behavior, guidance, and current implementation; establish versioning and decision changes. | Authorization Handbook v1 and a separate implementation roadmap. |

## Discussion method

For each discussion:

1. Pick one concrete question and show the current interpretation.
2. Work through an example and a counterexample.
3. Compare alternatives and their consequences; give a recommendation.
4. Ask the user to settle the definition or rule.
5. Record the decision, rationale, examples, and remaining dependencies.

The assistant guides the sequence, keeps questions focused, and updates this
record after decisions. Do not treat silence or a proposal as agreement.

Every proposal and question presented to the user must have a stable reference
ID. Use the decision ID for a proposal and a Q-NNN ID for a discussion question;
reference the same IDs in later discussion. Label alternative proposals with
suffixes when useful. User suggestions are candidates to examine, not automatic
decisions or preferences to implement. Test the shape with examples and
counterexamples before seeking agreement. Adopt a canonical term only when its
meaning and useful distinction have been established.

For every concept, capture:

- Meaning and distinctions from related concepts.
- Owner and representation.
- Lifecycle: when it is created, used, changed, and retired.
- Behavior when missing, invalid, or unavailable.
- Examples and counterexamples.

For anything called "resolved," also state what uncertainty has been removed
and what remains to be evaluated.

A chapter is finished when terminology is consistent, rules determine expected
outcomes, and dependencies are settled or explicitly deferred outside v1.
An implementation gap becomes tracked work; it does not automatically invalidate
an agreed principle.

## Evidence and decision discipline

Keep these separate:

- Agreed requirements: deliberately accepted authorization rules.
- Proposed rules: options awaiting discussion and acceptance.
- Current implementation: verified behavior tied to a repository and revision.
- Illustrative examples: explanations that do not establish implementation status.
- Open questions and gaps: uncertainty or divergence requiring further work.

Existing lab prose and code are starting evidence, not automatically authoritative.
Claims in the enforcement trace about other repositories need verification against
those repositories before being described as current facts. External comparisons
need primary sources when we assess them.

## Initial questions to revisit

- The existing request example restricts caller input to method, path, and token.
  Generalize the model to bodies, filters, creation, bulk work, and background jobs.
- Distinguish a caller-selected target identifier from established target facts;
  extracting a path parameter does not establish ownership or membership.
- Define whether "resolved grants" means loaded assignments, expanded roles and
  memberships, evaluated relationships, combined restrictions, or another stage.
- Verify the trace's reported mismatch between role-binding scope and
  permission-assignment scope against the relevant code revisions.

## Decision log

PROCESS-004 governs checkpoint commits and periodic reconciliation; see the
working handbook's checkpoint practice for the review checklist.

| ID | Status | Decision | Rationale / evidence |
|---|---|---|---|
| PROCESS-004 | AGREED | Commit and push meaningful documentation checkpoints throughout the discussion. Periodically reconcile the chapters, decision log, discussion tree, examples, and original handbook; retain superseded history and explicitly unresolved questions. | User instructed: "keep commiting and pusing. every now and then we need to reconcile". Reconciliation must not silently settle open policy choices or imply the application implements the handbook. |
| PLAN-001 | AGREED | Follow the eleven-stage roadmap and refine the handbook through focused discussions, worked examples, and recorded decisions. The assistant steers the discussion. | User approved the proposed roadmap on 2026-09-05: "looks good pin it. you steer the discussion." |
| PROCESS-001 | AGREED | Give every proposed rule and discussion question a stable reference ID. Treat suggestions from either participant as candidates until deliberately agreed; use examples to assess the model and the need for canonical terms. | User requested reference numbers and clarified: "even i am proposing and we need to see if the shape is getting right." |
| PROCESS-002 | AGREED | Maintain a traversable discussion tree covering all eleven stages, with conclusions, open siblings, detours, and explicit return points. Revisit every unfinished branch before declaring the handbook complete. | User requested a "traverse tree" so branches conclude and the entire tree is covered. The navigation artifact is discussion-tree.md. |
| PROCESS-003 | AGREED | Notes must retain sufficient detail to reconstruct the discussion: definitions, rationale, examples/counterexamples, consequences, links to decision IDs, and unresolved questions. Consolidate settled branches into working handbook chapters while retaining the decision log and tree. | User instructed: "make sure what you note down have suffeint details." grant-model.md consolidates the grant/role branch; it does not finalize unresolved schemas or implementation contracts. |
| CHARTER-001 | AGREED | The handbook defines mandatory shared authorization semantics for every AgentLabs application. Applications define their own resources, permissions, and relationships within those rules; storage and implementation may differ. | User answered "Yes" on 2026-09-05 to requiring the same core authorization rules across applications. Individual rules remain to be settled in the following stages. This decision alone does not require a central decision service or one policy engine. |
| CHARTER-002 | AGREED | Auth is the shared authority for identity, tenant membership, roles, and authorization assignments. Each application's authorization layer resolves domain facts and evaluates access under the shared handbook rules; the application's protected execution and data-access paths enforce the result. | User answered "yes" on 2026-09-05 to this responsibility boundary. Shared authorization libraries are compatible with it. This does not settle deployment topology, wire formats, or the details of grant resolution. |
| PRINCIPLE-001 | AGREED | A protected operation may proceed only when the system establishes valid authority covering the requested action and target within the applicable tenant boundary. If required information is missing, invalid, or unavailable and authority cannot be established, the operation must not proceed. | User agreed on 2026-09-05: "yes, this is right." Authentication or knowledge of a resource ID does not by itself establish access. Failure to establish authority must remain distinguishable from a completed evaluation that denies access. Exact grant-combination, freshness, and response rules remain open. |
| PRINCIPLE-002 | AGREED | Applications and endpoints are designed with an auth-first approach: authorization requirements guide the operation, URI, required domain facts, handler contract, and data-access enforcement from the start. | User instructed: "the applications or endpoints should be written with auth first approach." Extends CONTRACT-004's one-declaration rule to application design. This does not authorize implementation changes to existing applications in this discussion. |
| TENANT-001 | AGREED | Tenant is the enclosing authorization boundary and is implicit in a grant evaluated within trusted tenant context. Recipients, groups, scope references, and targets resolve within that boundary; grants cannot select or widen it. The authorization system must preserve the binding even when the grant payload omits a tenant field. | User instructed: "Tenant should be more implicit as this is the outer most boundry is implied." Grant examples omit repeated tenant IDs. Tenant-wide scope means the enclosing tenant. Storage, cache, and transport representations remain to be designed; implicit context is not permission to drop isolation or trust a caller's tenant assertion. |
| ARCH-001 | PROPOSED | Use a reusable authorization core with application-specific operation mappings and trusted domain-fact resolvers. The agent consumes Auth-owned authority information and application-owned facts; this does not require two network calls per request. | Clarifies CHARTER-002. ARCH-002 refines where facts enter: the generic middleware does not access the application database; the endpoint/application service supplies domain facts to the shared evaluator. Neither proposal is an implementation-status claim. |
| ARCH-002 | PARTIALLY AGREED | Generic middleware obtains Auth-owned authority, evaluates what it can without application-database access, and stops conclusive denials. The application supplies remaining domain facts and enforces authorization. Whether middleware returns a final allow versus prepared context is under renewed discussion. | User subsequently clarified: "I tentively agree to allow/prepare model, deny model is clear." Earlier acceptance of the three-outcome formulation must not be treated as final. See ARCH-003 and CONTRACT-001 for the certificate-route example. |
| ENFORCEMENT-001 | AGREED | A prepared context is not an allow decision. For a prepared request, the endpoint may perform only the internal, tenant/request-constrained work needed to establish required facts before completing authorization. Missing facts or an omitted required check cannot yield protected output or a business mutation. A completed middleware allow still requires enforcement of its action, target, and restrictions by the application. | Accepted as part of Q-010 with ARCH-002's refinement. Bind facts and the decision to actual use; avoid changing relevant facts between checking and use. Query predicates may combine final evaluation and access for collection operations; exact enforcement contracts remain open. |
| RESOLUTION-002 | SUPERSEDED | Earlier candidate: middleware dynamically chooses completed allow, deny, or prepared for any endpoint. | CONTRACT-002 replaces the universal three-outcome proposal with two endpoint-declared modes. Early conclusive denial and fail-closed evaluation errors remain applicable. |
| ARCH-003 | SUPERSEDED | Earlier candidate: every admitted request is treated as prepared for mandatory handler completion. | CONTRACT-002 limits mandatory authorization completion to endpoints explicitly declaring it; middleware-complete endpoints receive final allow or denial, never prepared. Endpoint enforcement remains required in both modes. |
| FACT-001 | PROPOSED | A route parameter supplies a requested identifier or boundary, not proof of relationships. A server-owned route mapping gives the parameter meaning; the application establishes that the actual resource satisfies the bound tenant, department, and any applicable ownership constraints through trusted facts or a constrained query. | For GET /api/v1/{tenant}/{dept}/{certificate}, the route does not prove that the certificate belongs to the named department or user. Ownership is required for self-scoped access, not automatically for every department-scoped grant. |
| CONTRACT-001 | PROPOSED | Each protected operation declares its permission mapping, the meaning of its request bindings, and how the required authorization facts or enforceable predicates are obtained. Applicable scope/delegation evaluators report requirements still unresolved. Middleware relies on those contracts to distinguish proven conditions from pending ones; it cannot infer arbitrary business facts from URL structure. | Explicit contracts allow predictable handling of supported operations. Undeclared or unsupported checks fail closed rather than being silently interpreted as satisfied. This is a contract proposal, not a finalized schema or a requirement to enumerate all target IDs. |
| CONTRACT-002 | AGREED | Each endpoint/operation declares one authorization resolution mode. Middleware-complete mode permits final allow or deny; there is no prepared result, and the handler enforces the completed decision. Endpoint-completion mode permits middleware deny or prepared only, never middleware allow; the handler supplies declared application facts, completes evaluation, and enforces the result. The mode is fixed by the server-owned contract, not selected by the caller or switched per request. | User accepted the single-declaration shape: "looks good," then asked for cases beyond self. The estimated 90–95% coverage is an unverified hypothesis, not a requirement or measured result. Final field names and schemas remain open. |
| CONTRACT-003 | PROPOSED | Validate an endpoint's declared mode against all supported authorization requirements, including scope and delegation forms. Middleware-complete mode cannot silently permit or fall back to prepared when a required fact is unavailable; incompatible configuration or runtime evaluation failure prevents protected execution. Endpoint-completion mode always follows its completion contract even if a particular grant appears broad enough for an early allow. | Makes CONTRACT-002 predictable while preserving PRINCIPLE-001. Changes to operations or supported scopes require checking the mode contract again. Mode declaration does not prove resource ownership or eliminate mandatory data restrictions. |
| CONTRACT-004 | AGREED | Design the URI, handler, and authorization declaration together. Each HTTP method plus route template has one authoritative authorization contract with one resolution mode. It declares the operation/permission mapping, request bindings, supported authorization requirements, and mandatory handler completion/enforcement. Middleware and handler use that same contract rather than separate conflicting declarations. | User replied "looks good" to the refined one-endpoint/one-workflow rule. Different grants or additional checks within the same declared contract do not require extra endpoints; incompatible workflows require a separate endpoint or deliberate contract revision. A different endpoint does not automatically require a different permission. |
| CONTRACT-005 | AGREED | Select endpoint-completion when required application facts are still needed to finish deciding authorized access. A handler merely applying already determined, complete mandatory restrictions is enforcement and does not alone require endpoint-completion. Evaluate this distinction against every supported authorization form for the endpoint. | User responded "looks good" to the beyond-self review. Examples EC-001 through EC-007 are in endpoint-completion-cases.md; acceptance of the distinction does not adopt every illustrative policy as a product requirement. Business validation is not automatically authorization, and no coverage percentage is established. |
| GROUP-001 | OPEN | Decide ownership and resolution of groups and teams. GROUP-001-A (refined working proposal): Auth owns generic authorization groups and their memberships; applications may synchronize business memberships into Auth or keep them independent. Authorization group membership is obtained from Auth under the freshness contract, not independently inferred from application business membership. GROUP-001-B (alternative): applications own some authorization-group memberships, consumed through explicit resolver integrations. | User proposed relying on Auth for groups/teams and membership, with optional application-owned synchronization. GROUP-001-A is the recommended working shape, still a proposal under PROCESS-001. Business department, ownership, and containment facts remain application-owned. Current implementation support remains unverified. |
| SYNC-001 | OPEN | For a group explicitly synchronized from business membership, define the authorized writer, update delay, removals, retry/reconciliation, and behavior when synchronization is stale or fails. | Synchronization is optional to configure, but a configured relationship needs a correctness contract. Until a change reaches Auth and relevant authorization caches, old membership may still confer access. Default denial alone cannot detect an unknown upstream change. Resolve timing and failure guarantees in stage 9. |
| TERM-001 | AGREED | "Team" and "group" mean the same authorization concept. Use "group" as the handbook's canonical term and "team" as a synonym, with no separate authorization behavior. | User answered Q-002 on 2026-09-05: "team/group should mean same." This settles synonymy; member types, nesting, and exact grant-inheritance rules are separate decisions. |
| GROUP-002 | NOT ADOPTED | Original assistant proposal: a group may contain human users and service/agent principals with valid memberships in the group's tenant. | User's Q-003 response favors human-only groups; replaced as the working direction by GROUP-003. Preserve this entry as discussion history. |
| GROUP-003 | PROPOSED | Groups/teams contain human users only. Service accounts and agents are not first-class group members. | User's proposed direction: "group should be purly for humans"; automated access uses explicit service-account assignments or human delegation. Nested-group behavior is not established by this statement. |
| SERVICE-001 | SUPERSEDED | Earlier proposal: service accounts obtain their own authority through explicit assignments rather than group membership. | AUTHORITY-002 replaces independently authorized service accounts with human-dependent service authority. How a user explicitly limits that authority remains open. |
| DELEGATION-001 | PROPOSED | An agent acting on behalf of a human may use explicitly delegated portions of that human's authority, including group-derived authority where delegation is permitted. The agent does not thereby become a group member. Delegated authority remains bounded by the human's applicable authority and the delegation's constraints; preserve both acting-agent and delegating-human identities. | Refines the user's indirect route. Recommended consequence: loss of the human's underlying group-derived authority removes that route for the agent too. Who may delegate what, independent service authority, future grant expansion, and revocation freshness remain unresolved. |
| DELEGATION-002 | AGREED | Delegated authority depends on the continuing validity of the authority and relationships supporting it. When a required upstream relationship or supporting authority ceases to be valid, authority derived through that dependency is no longer valid. This applies transitively if further delegation is supported. | User answered Q-004: "delegation is highly dependent relation. once the higher relation breaks. lower relation automatically has no value." This establishes semantic invalidity without requiring manual downstream revocation. It does not establish support for redelegation, exact revocation propagation time, unrelated access paths, automatic reactivation, or substitution of supporting grants. |
| SERVICE-002 | NOT ADOPTED | Earlier proposal: independent service assignments survive their creator losing access. | User rejected the complexity and risk of independently authorized service accounts. Replaced by AUTHORITY-002. |
| AUTHORITY-001 | SUPERSEDED | Earlier candidate model distinguished independent service assignments from proxy authority. | Replaced by AUTHORITY-002 at the user's direction. Do not carry the independent-service branch into the canonical model. |
| AGENT-001 | SUPERSEDED | Earlier proposal: agents always act as proxies with no independent authority. | Incorporated into AUTHORITY-002, which applies the human-dependency requirement to both agents and service accounts. Separate actor identity remains compatible with dependent authority. |
| TERM-003 | SUPERSEDED | Earlier vocabulary proposal compared independent assignments with delegated authority. | The independent-service distinction is no longer part of the working model. Identity versus authority remains a useful distinction for the later glossary; no independent-service vocabulary is needed. |
| AUTHORITY-002 | AGREED | Every service account and agent is under a human user and has only dependent authority that is a subset of that user's applicable authority. There is no independent service-account assignment path. Resolution restricts access to the authority still supported by the user and the applicable delegation constraints. Loss of unrelated user rights does not invalidate still-covered service/agent access. | User required human-dependent service and agent authority, then answered Q-007: "it is obvisouly latter i.e 2. service account/agent will alway be subset ... resolution should limit the access." This supersedes SERVICE-001, SERVICE-002, AUTHORITY-001, and AGENT-001. It establishes restriction of affected authority, not blanket account invalidation on every rights change. |
| RESOLUTION-001 | AGREED | Authorization resolution cannot create authority beyond the applicable authorized inputs and constraints. The resolved authority is a subset of the authority available under those inputs; equality is allowed. Dependent service/agent access must stay within the user's applicable authority. | User stated: "resulution is alway a subset." This is a semantic authority bound, not a requirement that output JSON contain fewer fields or entries. Role expansion and relationship resolution may make existing authority explicit without creating new authority. Exact resolution contracts and combination rules remain to be defined. |
| PERMISSION-001 | AGREED | A permission identifies an operation on a resource type; scope determines which instances an assignment can reach. Reading one's own payroll and reading a department's or tenant's payroll use the same permission when the underlying operation is the same. Self, department, and tenant reach stay outside the permission name. | User answered Q-008: "agree." Does not finalize permission grammar, scope types, or how different business operations should be mapped. |
| SCOPE-001 | AGREED | A scope describes the set of resources over which an assigned permission may apply. It may express explicit selections or selectors based on trusted resource attributes and relationships; it need not enumerate IDs or follow one hierarchy. A target is what the request seeks to act on, checked against that reach. | User answered Q-009: "Yes agree to this." Illustrations: exact payslip P-17; payslips owned by the principal's employee; payslips associated with Finance. Final selector grammar, temporal relationship semantics, and targets for collection/create operations remain open. A scope description is not itself an allow decision. |
| GRANT-001 | AGREED | A grant binds a recipient, permission or role reference, scope, and validity/conditions as one authority unit. Resolution must preserve which scope and conditions belong to which capability; it must not combine a capability from one grant with broader reach supplied only by a different-capability grant. | User answered Q-012: "agree" and requested JSON. GRANT-EX-001 in grant-examples.md illustrates tenant-wide read and Finance-only revoke. This does not settle grant versus assignment terminology, role expansion, same-permission grant combination, or dependent service grant representation. |
| GRANT-002 | AGREED | A grant may bundle one or more permissions when they share its recipient, scope, validity, and conditions. Permissions requiring different reach or conditions use separate bindings. | User said "proceed" after this proposal and its example. GRANT-EX-003 illustrates the concept; the exact permissions-array wire encoding remains illustrative. Role references and role expansion remain separate discussions. |
| GRANT-003 | AGREED | A human may have multiple directly assigned grants and grants applicable through valid group memberships in the enclosing tenant. Resolution gathers applicable bindings and retains their permission/scope/condition associations and direct/group provenance. | User said "proceed" after the direct/group grant example. Group-derived access depends on membership; loss of that membership does not erase separately valid direct grants. This does not settle explicit-deny precedence, general grant-combination semantics, or nested groups. |
| ROLE-001 | AGREED | A role is a named reusable permission bundle. A grant references either such a role or an explicit permission set and binds it to a recipient, scope, validity, and conditions. A role definition supplies capabilities; the grant supplies their assigned reach. Group is the collection of human recipients, role is the collection of capabilities, and grant binds recipient to scoped authority. | User answered Q-013: "makes sense." GRANT-EX-004 illustrates this. Role-change behavior is settled by ROLE-002. Role templates/defaults, scope compatibility, and revision mechanics remain open. |
| ROLE-002 | AGREED | A role-referencing grant uses the role's current permission definition. Authorized role edits therefore change capabilities available through existing referencing grants while preserving each grant's scope, validity, and conditions. Unrelated direct grants remain separate. | User answered Q-014: "yes." Grants do not require manual upgrades to adopt a role edit. Who may make such edits, compatibility validation, revision evidence, propagation freshness, and dependent delegation limits remain separate requirements. Role editing changes declared authority; resolution must not independently invent permissions. |
| RESOLUTION-003 | AGREED | Expand a role reference into its current permission set, producing the same permission-set grant form used to evaluate explicitly listed permissions. Preserve the original grant identity, scope, recipient, validity, conditions, and source provenance, including the role definition used. Expansion is a computed view of the same grant, not a new assignment, and does not merge unrelated grants. | User agreed after clarification that the earlier JSON represented the existing G-6 after role lookup. GRANT-EX-005 retains G-6 visibly. Expanded does not mean fully resolved: application relationships or target facts may still be missing. Exact schema, role-revision encoding, membership expansion, and final resolved-grant representation remain open. |
| TERM-004 | AGREED | Use "grant" as the canonical name for the authorization binding record. "Assignment" refers to that same binding when describing a permission/role being assigned; "assign" is the action creating the binding. Do not introduce a second logical assignment object solely because both words appear in explanations. | User answered Q-016: "agree." Rationale, examples, and consequences are in grant-model.md. This is a logical vocabulary choice, not a requirement for one physical database table. |
| RESOLUTION-004 | AGREED | Access through a group grant is always derived from the original group grant and the human's valid membership. Preserve the group grant identity/recipient and the requesting human and supporting membership in the evaluation context. It must never become an independent direct user assignment. | User answered Q-017: "it should never be independent. should alway be derived." Removal of the supporting membership or grant withdraws that derived route, without erasing unrelated valid grants. Computed/cached views must retain this dependency. Exact fields, membership evidence, and freshness remain open. |
| DECISION-001 | AGREED | For positive grants supporting the same required permission, treat each complete applicable grant as an alternative authorization route. The authorized resource set is the union of the routes that remain valid under their own scope, conditions, and dependencies, subject to mandatory outer restrictions. Preserve route provenance; do not mix fields across grants or create an independent grant from the aggregate. | User answered Q-018: "agree." Direct Finance read plus group-derived Engineering read covers Finance or Engineering, assuming no other rule prohibits access. If Engineering membership ends, only that route disappears. This does not settle explicit-deny semantics or override delegated service/agent limits. |
| DECISION-002 | AGREED | V1 grants are positive-only: a grant supplies bounded authority, not an explicit prohibition. Access may still be denied when no complete valid grant authorizes it or a mandatory boundary/constraint prevents it. Explicit deny-grant objects and their conflict-precedence rules are outside v1. | User answered Q-019: "yes." Withdrawing access requires removing/narrowing every route supplying it; removing one grant does not cancel others. A grant-local condition failure does not veto another complete valid grant. Tenant/delegation limits remain enforced; this decision does not add a configurable global-policy mechanism. |
| ADMIN-001 | AGREED | Possessing a business permission does not by itself authorize assigning it to another recipient. Creating, changing, or revoking grants requires explicit grant-administration authority in the enclosing tenant and within its permitted administrative bounds. Role edits and group membership changes that affect access also require their respective administrative authorization. | User answered Q-020: "Agree." Exact recipient/capability/scope bounds, whether grantors must personally possess assigned rights, and delegation permission remain open. This does not define an unlimited administrator. |
| SELF-001 | AGREED | In a group-derived self-scoped grant, self refers to the human whose authority is being evaluated, not the group recipient, group creator, or grant creator. The same scope descriptor resolves separately for each human member. A dependent agent/service uses the authorizing human's self relationship, within the delegation's limits. | User stressed that an Employees group with self scope means each user as self, not the group. This applies SCOPE-001 and AUTHORITY-002 without making all members' resources available to every member. GRANT-EX-006 illustrates it. Exact scope encoding and relationship-resolution timing remain open. |
| GROUP-004 | AGREED | Prefer assigning human access to groups and deriving it through membership, including self-service access through an Employees group. Direct human grants remain supported; group-based access is a preferred practice, not an exclusive permitted mechanism. | User answered Q-021: "we should put it as prefered not permit only." This preserves GRANT-003 and does not introduce a special approval workflow for direct human grants. Human-dependent service/agent delegation remains separate from group membership. |
| TERM-002 | PROPOSED | Use "domain fact" as explanatory vocabulary for an application-owned attribute or relationship used in authorization, such as a payslip's owner or a repository's parent project. | Test whether this category helps explain ownership and resolution. It need not be a separate stored entity or wire type; its relation to trusted context remains to be defined. |

## Referenced discussion questions

| ID | Status | Question | Related proposals |
|---|---|---|---|
| Q-001 | ANSWER INCORPORATED | May business membership and access-group membership differ? User's proposed answer: rely on Auth for authorization groups and memberships; applications decide whether to synchronize business memberships or leave them independent. | GROUP-001-A, SYNC-001; ownership proposal awaits finalization. |
| Q-002 | ANSWERED | Team and group mean the same authorization concept. | TERM-001 agreed. |
| Q-003 | ANSWER INCORPORATED | User proposes human-only groups; service accounts need explicit assignments, and agents may act through a human's delegation rather than first-class group membership. | GROUP-002 not adopted; GROUP-003, SERVICE-001, DELEGATION-001 capture the revised direction. |
| Q-004 | ANSWERED | Yes: delegation is a dependent relationship; loss of required upstream authority invalidates downstream derived authority. | DELEGATION-002 agreed; freshness guarantees remain a separate stage 9 decision. |
| Q-005 | ANSWER INCORPORATED | User proposes classifying authority as first-class or proxy: independent grants remain under their own lifecycle, proxy authority becomes invalid when its supporting relation breaks; agents should always be proxies. | AUTHORITY-001, AGENT-001, TERM-003; model proposed for assessment, not yet finalized. |
| Q-006 | CLOSED BY MODEL CHANGE | The independent-service premise was rejected. All service-account and agent authority must depend on a human user. | AUTHORITY-002 agreed. This does not establish support for chains between dependent services and agents. |
| Q-007 | ANSWERED | The service retains still-covered repository access. Resolution limits effective authority to a subset; unrelated rights loss does not invalidate the entire account. | AUTHORITY-002 clarified; RESOLUTION-001 agreed. |
| Q-008 | ANSWERED | Use the same read permission with different scopes; keep self, department, and tenant reach out of the permission name. | PERMISSION-001 agreed. |
| Q-009 | ANSWERED | Yes: scope supports explicit selections and attribute/relationship selectors, without requiring enumeration of every resource ID. | SCOPE-001 agreed. |
| Q-010 | PARTIALLY ANSWERED | Early conclusive middleware denial is accepted. Subsequent discussion replaced a universal allow/prepared split with endpoint-declared modes. | See CONTRACT-002 through CONTRACT-004 for the current proposals; RESOLUTION-002 and ARCH-003 are superseded. ENFORCEMENT-001 remains applicable. |
| Q-011 | ANSWER INCORPORATED | User proposes an explicit endpoint mode: middleware-complete endpoints have no prepared result; endpoints requiring further resolution receive only deny/prepared from middleware, never allow. | CONTRACT-002 and CONTRACT-003 capture the working proposal for assessment. |
| Q-012 | ANSWERED | Yes: a grant is a binding whose capability, scope, and conditions stay associated throughout resolution. | GRANT-001 agreed; JSON example GRANT-EX-001 saved in grant-examples.md. |
| Q-013 | ANSWERED | Yes: role supplies reusable permissions; the referencing grant supplies recipient, scope, validity, and conditions. | ROLE-001 agreed. |
| Q-014 | ANSWERED | Yes: role edits update capabilities supplied through existing references while grant scope, validity, and conditions remain attached. | ROLE-002 agreed. |
| Q-015 | ANSWERED | Yes: the expanded view of the same grant exposes permissions for evaluation while preserving original identity, source, and restrictions. | RESOLUTION-003 agreed; exact JSON schema remains illustrative. |
| Q-016 | ANSWERED | Yes: grant is the canonical binding; assignment refers to the same binding, and assign is the action creating it. | TERM-004 agreed; detailed rationale and examples in grant-model.md. |
| Q-017 | ANSWERED | Group-derived access must always remain derived from the original grant and membership, never independent. | RESOLUTION-004 agreed. |
| Q-018 | ANSWERED | Yes: valid grants for the same permission provide alternative routes, preserving each route's conditions and dependencies. | DECISION-001 agreed. |
| Q-019 | ANSWERED | Yes: v1 grants are positive-only; explicit deny grants are excluded. Denied decisions and mandatory restrictions still apply. | DECISION-002 agreed. |
| Q-020 | ANSWERED | Yes: administering grants requires explicit authority, not mere possession of the business capability. | ADMIN-001 agreed; administrative bounds remain open. |
| Q-021 | ANSWERED | No prohibition: group-based access is preferred, while direct human grants remain supported. | GROUP-004 agreed; GRANT-003 unchanged. |

## Resume here

- Current position: [discussion tree](discussion-tree.md), returning to stage 5 →
  ADMIN-001's administrative bounds after closing GROUP-004 / Q-021.
  Current user request also checks handbook detail; handbook.md indexes the
  detailed working chapters, while the original lab page remains unreconciled.
  After grant authority/lifecycle
  questions, define the request and its
  prepared/resolved forms (stage 7). The endpoint-mode discussion established
  where these objects are consumed; their contents and semantics need precision.
  Earlier stages retain open questions; this move does not mark them complete.
- Settled: CHARTER-001 and CHARTER-002 — shared rules across applications;
  Auth holds identity and assignments; applications resolve domain facts,
  evaluate authorization, and enforce the result.
- Also settled: PRINCIPLE-001 — protected operations require established,
  applicable authority; an inability to establish it must not permit execution.
- Explain ARCH-001: a shared core can use application-specific adapters to obtain
  domain facts; logical information ownership does not dictate network topology.
- Working proposal: GROUP-001-A, refined with optional application-owned
  synchronization into Auth. Do not treat the working shape as a finalized rule.
- Q-001's input is incorporated. Preserve SYNC-001 for stage 9 rather than
  deciding synchronization guarantees implicitly during vocabulary discussion.
- Also settled: TERM-001 — team and group are synonymous; group is the canonical
  handbook term. No separate team authorization semantics.
- Q-003's response replaces GROUP-002 with the human-only working proposal
  GROUP-003. Independent service assignments have since been rejected.
- Explain DELEGATION-001: human membership can support delegated agent access
  without making the agent a group member. Do not assume that all human authority
  is delegable or combine independent service grants with delegated authority.
- Also settled: DELEGATION-002 — downstream delegated authority has no validity
  when a required upstream dependency breaks. This does not authorize redelegation.
- Latest settled direction: AUTHORITY-002 — both service accounts and agents are
  human-dependent and cannot exceed the user's authority. Superseded independent
  service proposals are history, not current options.
- Q-007 is answered: restrict effective authority; preserve still-covered access.
- Also settled: RESOLUTION-001 — resolution cannot amplify authority. Explain
  this as a bound on meaning, not on the size of its data representation.
- Also settled: PERMISSION-001 — permission identifies the operation; scope
  expresses self, department, or tenant reach separately.
- Also settled: SCOPE-001 — explicit selections and attribute/relationship
  selectors can describe scope without enumerating every resource ID.
- Agreed: CONTRACT-002 — two endpoint-declared modes. Middleware-complete
  endpoints use allow/deny; endpoint-completion endpoints use deny/prepared in
  middleware and must complete authorization in the handler. Mode is contractual,
  not dynamically inferred per request. RESOLUTION-002 and ARCH-003 are superseded.
- Explain CONTRACT-003: validate the mode against supported scope/delegation
  requirements; missing required facts cannot silently produce allow or switch
  a middleware-complete endpoint to prepared. Enforcement remains mandatory.
- Also agreed: CONTRACT-004 — URI, handler, and a single authoritative
  authorization declaration are designed together per HTTP method/route. Keep
  one mode per endpoint. Distinguish different contracts from multiple checks
  or grant scopes handled under one contract; avoid creating a route per grant.
- Explain FACT-001 using the certificate route: requested tenant and department
  IDs do not prove certificate containment or ownership. Required ownership
  checks depend on the applicable scope, not simply on the existence of a user.
- Also agreed: CONTRACT-005 — missing authorization facts versus applying
  complete enforcement restrictions. Candidate cases EC-001 through EC-007 are
  in [endpoint-completion-cases.md](endpoint-completion-cases.md).
- Also agreed: GRANT-001 — capability/scope/conditions remain bound during
  resolution. [grant-examples.md](grant-examples.md) contains GRANT-EX-001 with
  illustrative JSON and expected outcomes; the wire schema is not finalized.
- Also agreed: PRINCIPLE-002 — auth-first application and endpoint design.
- Also agreed: TENANT-001 — trusted tenant context encloses grants; examples
  omit repeated tenant fields and use tenant-wide scope relative to that context.
- Also agreed: GRANT-002 and GRANT-003 — multiple permissions per shared
  binding, and multiple grants applicable directly or through groups.
  GRANT-EX-003 illustrates both in [grant-examples.md](grant-examples.md).
- Also agreed: ROLE-001 — reusable role definitions supply capabilities;
  referencing grants supply recipient and constraints. GRANT-EX-004 illustrates it.
- Also agreed: ROLE-002 — existing referencing grants follow current role
  permissions, retaining each grant's own scope, validity, and conditions.
- Also agreed: RESOLUTION-003 — role expansion produces a computed permission-set
  view of the same grant, retaining its identity, provenance, and restrictions.
- Also agreed: TERM-004 — grant/assignment denote one binding. The detailed
  [grant model chapter](grant-model.md) consolidates definitions, rationale,
  examples, consequences, and open branches as required by PROCESS-003.
- Also agreed: RESOLUTION-004 — group-derived applicability always depends on
  the original group grant and membership; it never becomes an independent grant.
- Also agreed: DECISION-001 — positive grants are alternative authorization
  routes, each retaining its own conditions and dependencies.
- Also agreed: DECISION-002 — positive-only grants; explicit deny grants and
  associated precedence rules are outside v1. Denial decisions remain essential.
- Also agreed: ADMIN-001 — grant administration requires separate authority.
- Also agreed: SELF-001 — self is evaluated per human, even for group grants;
  proxy access retains that human reference and the delegation's limits.
- Also agreed: GROUP-004 — prefer membership-based access; direct human grants
  remain supported. Q-021 closes this sidebar; return to administration bounds.
- Documentation detail has been reviewed: grant-model.md and authorization-flow.md
  preserve narrative rationale, examples, consequences, and open contracts;
  handbook.md is the working entry point. Original lab prose is not yet reconciled.
- Next discussion topic: boundaries on grant administration. No new question ID
  has been issued in this documentation-status turn.
  After grant authority/lifecycle questions, compare
  declared, prepared, and resolved forms. Keep other branches in the tree open.
- Coverage percentages require a later endpoint audit and are not assumed proven.
- The generic middleware has no application-database dependency. The
  shared evaluator can be invoked again with domain facts; handlers should not
  each invent scope-combination semantics. Preserve unresolved constraints.
- Do not resume independent service models or blanket invalidation of accounts
  on unrelated rights changes.
- Defer exact revocation timing, reactivation, and alternate supporting-grant
  semantics to their later lifecycle and grant-resolution discussions.
- TERM-002 remains provisional explanatory vocabulary for application-owned facts.
- Preserve the distinction between a policy denial and an authorization
  evaluation failure; both prevent the protected operation, but are not the
  same diagnostic outcome.
- Stage 1 follow-ups, to settle before v1 publication: audience, explicit
  exceptions, and ownership of changes to the shared contract. Moving the
  discussion forward does not mark those questions agreed or complete.
