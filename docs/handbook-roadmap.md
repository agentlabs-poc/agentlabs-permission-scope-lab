# Authorization handbook — agreed roadmap

## Current reconciliation and process override

The user's approved discussion takes precedence over older `src/content` prose,
diagrams, and lab implementations. The [current tree](discussion-tree.md) and
[system overview](system-overview.md) reconcile through Q-050-F. Historical
decision rows retain their original rationale; later refinements govern where
their earlier open/next wording has changed.

**Commit/push gate reopened:** the user explicitly instructed “lets commit and
push, and continue with the rest.” PROCESS-004's checkpoint publishing can resume
after verification. The earlier review freeze is preserved in historical notes.
Publishing a checkpoint does not approve an open policy choice or authorize
runtime migration.

The user requested horizontal progress on critical/high-impact decisions.
Decision results, operation-specific enforcement, and freshness/dependencies
remain candidates; no exact ordering or new answer is recorded as approved.

## Original roadmap and chronological decision record

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
| 3. Core vocabulary | Actor, principal, subject, identity, tenant membership, group, resource type, resource instance, operation, permission, role, assignment, grant, policy, scope, context, and authority. | A glossary and relationship map, including deliberate distinctions and synonyms. |
| 4. Permissions and operations | Meaning, naming, operation-to-permission mapping, multiple required permissions, resource hierarchy, wildcards, and compatibility as capabilities evolve. | Permission semantics and catalog rules. |
| 5. Grants and their lifecycle | Direct, role, and group assignment; who may grant what; validity; provenance; role changes; revocation; delegation; assignment versus grant. | An assignment model and rules for creating, changing, and withdrawing authority. |
| 6. Scope boundaries | Scope type, descriptor, referenced resource, resolved scope, effective reach, static selectors, dynamic relationships, exact and subtree reach, multiple dimensions, empty and missing scope; request material for reads, lists, creation, updates, and moves. | Scope semantics, resource semantics, and containment rules. |
| 7. Requests and resolution | Incoming application request versus authorization request; resolved request, grant bundle, resolved grants, resolved request material; inputs, outputs, owners, sources of facts, timing, and failures for each resolution operation. | Explicit contracts and a complete resolution flow. |
| 8. Decision semantics | Combining permissions, scopes, grants, tenant boundaries, conditions, and delegation; union versus intersection; deny precedence; missing information; contributing-grant provenance. | Decision rules with unambiguous examples. |
| 9. Enforcement and change over time | Row and field restrictions; list, count, export, bulk and mutation behavior; caches; revocation freshness; resource moves; relationship changes; audit evidence. | Enforcement obligations and consistency guarantees. |
| 10. Challenge the model | HRMS and repositories end to end; conflicting grants, stale membership, cross-tenant resources, missing relationships, delegated agents, partial bulk access; verify relevant implementation and external references. | An agreed scenario suite and an evidence-backed implementation gap register. |
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
- Distinguish a caller-selected resource identifier from established resource facts;
  extracting a path parameter does not establish ownership or membership.
- Define whether "resolved grants" means loaded assignments, expanded roles and
  memberships, evaluated relationships, combined restrictions, or another stage.
- Verify the trace's reported mismatch between role-binding scope and
  permission-assignment scope against the relevant code revisions.

## Decision log

TERM-005 / Q-043 aligns current wording with permission, scope boundaries,
requests, and trusted request material. [Original superseded lines](history/q043-vocabulary.md)
are preserved verbatim; other decision statuses are not changed by this wording
alignment. [Detailed reasoning](authorization-vocabulary.md) explains the
tenability check and the continuing evidence-to-execution requirement.

Current endpoint authority is CONTRACT-006; see endpoint-authorization.md.
CONTRACT-007 supplies the agreed permission/material/source declaration. The
[reconciliation register](reconciliation.md) maps historical rationale to later
approvals, without deleting earlier text or finalizing open policy questions.
Historical decision text is preserved where deprecated: status labels and the
new chapter govern current interpretation. CONTRACT-004 retains the single
auth-first declaration principle, but not the old resolution-mode choice.

ADMIN-002 and ADMIN-003 are agreed under Q-022; grant-model.md explains the
separation of administration, business access, and explicit audited self-assignment.

PROCESS-004 governs checkpoint commits and periodic reconciliation; see the
working handbook's checkpoint practice for the review checklist.

| ID | Status | Decision | Rationale / evidence |
|---|---|---|---|
| DECISION-005 | AGREED AS REFINED | The evaluator supplies error_message and error_message_reason, both reach the UI, and the UI controls presentation; error_message is more user-facing. | User corrected the application-authored mapping proposal and explicitly rejected keeping the second message server-side in Q-053-A. Evaluator ownership avoids reconstructing the explanation in applications; both delivered fields are client-visible regardless of display. Previous proposals and rationale remain in decision-results.md. Exact value encoding, full schema, and safe-content rules remain open. |
| DECISION-006 | AGREED | Evaluation errors use the same two message fields as denials, without collapsing the distinction between them. | User answered Q-054 “yes.” Consistent UI presentation does not imply equivalent authorization meaning: a timeout is not proof of missing permission. Both cases still stop protected execution. See decision-results.md for example, rationale, counterexample, and outstanding schema questions. |
| DECISION-007 | AGREED | The evaluator supplies error_code for a stable machine-readable cause alongside readable error_message and error_message_reason. | User explicitly answered “Q-055 yes.” Software can identify the cause without parsing wording that may change or be translated. Using the reason message itself as a code is not adopted. Exact codes, catalogue evolution, and full result envelope remain open; see decision-results.md for rationale and counterexample. |
| PERMISSION-002 | AGREED WITH DEPTH | Canonical naming is namespaced-noun::verb, with variable-depth colon-separated application/domain/subdomain-or-resource segments before the double-colon verb separator. | User confirmed Q-056 “this is right” and requested deeper namespace explanation from earlier Markdown. Its Resource-type depth section contains ledger/entry/attachment examples, now restored in permission-model.md. Variable depth avoids forcing every application's domain into identical levels; Auth does not interpret those levels. Character rules, depth limits, catalog evolution, and inheritance/wildcards remain separate. |
| PERMISSION-003 | AGREED | A parent permission name does not automatically authorize a deeper permission merely because of a shared namespace prefix. | User answered Q-057 “agree.” Adding a specialized operation must not silently expand existing grants through naming. Ordinary grants/roles can explicitly list both permissions, subject to their own authority and lifecycle rules. Wildcards and aliases remain separate. See permission-model.md for rationale, alternative, and counterexample. |
| PERMISSION-004 | AGREED — OUTSIDE V1 | Exclude wildcard permission names from v1; use explicit registered permissions, optionally bundled in roles. | User answered Q-058 “not in v1.” Avoid new matching-depth and future-permission expansion rules in the foundation. Wildcard patterns cannot be accepted or expanded into v1 authority. The alternative one-segment wildcard remains preserved as history. Aliases and live-role lifecycle are not changed. See permission-model.md for rationale and counterexample. |
| PERMISSION-005 | AGREED — OUTSIDE V1 | Exclude permission aliases from v1; registered permission identifiers are not interchangeable through alias mappings. | User answered Q-059 “no alias.” Keep references unambiguous across grants, endpoints, registration, and audit, without extra equivalence/migration rules. UI display labels are unaffected; separately registered names remain distinct. See permission-model.md for rationale, alternative, and counterexample. |
| DECISION-008 | AGREED | Return supporting-grant references with an allow result to the server-side endpoint, whether or not the request is selected for persistent audit. | User approved Q-060 and clarified audit may be needed, but not all requests will be tracked. Preserve evidence at evaluation time without mandating universal request logging. No promise of retrospective reconstruction without retained evidence; snapshots, storage, selection, and retention remain open. References do not return all grants, replace scope enforcement, or mandate UI disclosure. See decision-results.md for rationale and counterexamples. |
| DECISION-009 | NOT ADOPTED | Do not require evaluated-boundary/scope fields in the allow result; supporting-grant references remain required. | User answered Q-061 “not required.” Keep the result minimal without adding a redundant boundary-delivery contract. The endpoint's existing correctly bound enforcement material and actual-use constraints remain mandatory; missing returned scope is not tenant-wide authority. No second lookup or prepared state follows. Original proposal and rationale remain in decision-results.md. |
| DECISION-010 | AGREED | Minimal allow representation: version "1", decision "allow", and grant_ids containing supporting grant identifiers. | User answered “062 agree.” An ID array captures Q-060's evidence without full grant records or boundary fields rejected in Q-061. It identifies supporting grants, not all the human's grants, and does not mandate logging every request. Full schema, historical evidence, correlation, and other variants remain open. See decision-results.md for rationale and counterexample. |
| DECISION-011 | AGREED | Minimal deny representation: version "1", decision "deny", error_code, error_message, and error_message_reason. | User answered Q-063 “agree.” Reuses the completed-decision field and separates machine cause from readable messages. A cause must reflect established evidence; the example code is not a finalized catalogue. No grant_ids field is required in this minimum; detailed rejected-candidate traces remain separate. See decision-results.md for rationale and timeout counterexample. |
| DECISION-012 | AGREED | Minimal evaluation-error representation: version "1" and the three agreed error fields, without decision. | User answered Q-064 “agreed.” Neither allow nor deny was established, so omit decision instead of adding a third decision value or status field. The approved timeout scenario illustrates it; the code spelling is not finalized. Missing decision alone does not validate malformed output, and protected execution stays blocked. Full validation and transport remain open. See decision-results.md. |
| DECISION-013 | AGREED | Reject responses mixing known variant-specific fields from allow, deny, and evaluation-error shapes. | User answered Q-065 “agree.” An allow containing error_code cannot authorize execution by ignoring the conflicting field. Reject as malformed and stop protected execution, not as a completed policy denial. This keeps consumers consistent and exposes integration defects. Unknown extensions, exact value validation, and code compatibility remain open. See decision-results.md for rationale and counterexample. |
| DECISION-014 | AGREED | Allow requires a non-empty grant_ids array of non-empty supporting-grant string identifiers. | User answered Q-066 “yes.” Q-060's supporting evidence remains required even with logging disabled or tenant-wide scope. Missing/null/empty references do not satisfy the result contract; malformed allow blocks execution, not a completed policy denial. Exact ID grammar, historical evidence, and route selection remain open. See decision-results.md for rationale and counterexample. |
| DECISION-015 | AGREED | Reject fields not defined in the supported authorization-result contract rather than ignoring them. | User explicitly answered “067 agree.” Strict rejection exposes typos and contract drift; ignoring extras favors compatibility but may conceal unconsumed information. This concerns the result object, not unrelated application data. Mixed known-variant fields are separately covered by Q-065. See decision-results.md for rationale and counterexample. |
| PROCESS-007 | AGREED | Prioritize horizontal coverage by impact, one question at a time, before extended field-level contract polishing. | User explicitly requested “move horizontally” and “coverage based on impact,” while separately approving Q-067. Remaining lower-level details stay tracked, not approved or excluded. Earlier instruction to continue after answers without waiting for proceed remains in force. The tree records the next recommended impact pass. |
| ENFORCEMENT-004 | AGREED | For the same-tenant boundary-changing move, require move authority covering both current and proposed boundaries. | User requested the fuller Q-068 Finance-to-Engineering example and answered “agree.” Source-only checks can place data into an unauthorized boundary; destination-only checks can pull data from one. The approved explanation includes a Finance-only grant, trusted current state versus proposed input, and tenant-wide coverage. One permission/one gate remains; extra read permissions are not required by this rule. Composition and concurrency stay open. See operation-enforcement.md. |
| FRESHNESS-001 | AGREED | After Auth confirms a grant's revocation, authorization checks started afterward cannot use that grant through a stale cache. | User answered Q-069 “yes.” Confirmed revocation has no stale-cache grace period for new checks, even if local cache expiry is later. Other valid grants remain usable; unavailable freshness cannot justify stale allow. The coordination/availability cost and rejected bounded-staleness alternative are recorded. Cache protocol, in-flight work, and other change propagation remain open. See authority-freshness.md. |
| DELEGATION-003 | AGREED AS CORRECTED | Affected delegated access becomes inactive when human support is lost and automatically works again when support returns, provided the delegation itself remains valid and within its limits. | User clarified Q-070: delegation is a subset of Vinay; step 2 becomes inactive, at step 3 it works again. This rejects the proposed explicit-renewal requirement. Dynamic dependence avoids an added renewal workflow; revoked/expired delegations and unrelated access are not revived or changed. No new status field adopted. Rationale, trade-off, and original proposal remain in delegation-lifecycle.md. |
| ENFORCEMENT-005 | AGREED AS CORRECTED | Deny the partially authorized collection request instead of automatically deriving and returning an authorized subset. | User rejected Q-071's filtering recommendation: “too much intelligence in endpoint about auth. it shold just deny.” Keep grant/scope resolution in evaluation; do not require the endpoint to translate grants into a narrower request. Application data-path enforcement remains mandatory. Rationale, trade-off, and original proposal retained in collection-enforcement.md. |
| ENFORCEMENT-006 | AGREED | Distinguish an explicitly bounded collection request from a broader request; Finance-only authority can cover an explicit Finance list but cannot silently narrow an all-departments list. | User answered Q-072 “yes” after the Finance-only grant and two-request explanation. The endpoint declares department material and constrains data access to that same evaluated tenant/department; it does not inspect grants to discover accessible departments. Rationale and previous proposal status retained in collection-enforcement.md. |
| ENFORCEMENT-007 | AGREED | For one bulk request, establish authorization and required boundary checks for the complete batch before effects; an uncovered item blocks the whole operation rather than triggering partial execution. | User answered Q-073 “Agree” to the Finance deletion batch containing an Engineering certificate. One authorization outcome avoids surprise partial changes, at the cost of one uncovered item blocking the batch. This is not a database rollback guarantee; concurrency, transactions, retries, representation, and error mapping remain open. Rationale and prior proposal status retained in bulk-enforcement.md. |
| ENFORCEMENT-008 | AGREED | Preserve evaluated application boundaries through the protected data operation; a concurrent boundary change must not turn an earlier allow into an out-of-boundary effect. | User answered Q-074 “Agree” to C-17 moving from Finance to Engineering between authorization and update. Stop the original attempt if its required boundary no longer holds; no ID-only fallback or silent material substitution. This is application enforcement, not grant interpretation. Rationale and original proposal status retained in concurrent-enforcement.md; mechanisms, conflict reporting, retries, and in-flight Auth authority changes remain open. |
| ENFORCEMENT-009 | AGREED | Queued work obtains authorization for the protected operation at execution time; submission's stored allow is not a durable execution grant. | User answered Q-075 “agree” to the queued Finance export whose sole supporting grant is revoked before the worker starts. A new evaluation preserves current human-dependent authority; accepted jobs may therefore fail to execute. Rationale, trade-off, and previous proposal status retained in background-authorization.md. Job schemas, retries, and running-job behavior remain open; no independent worker authority or prepared state. |
| AUDIT-001 | PROPOSED | Successful grant, authorization-membership, live-role, and delegation administration mutations require audit; ordinary access-request logging remains selective. | Q-076 uses Maya adding/removing Vinay from Finance: effective access changes without a grant edit. Auditing only grant edits or self-assignment misses indirect authority changes. Mandatory capture has a storage/reliability cost; schemas, failure guarantees, retention, and additional event categories remain open. See authority-change-audit.md. |
| DECISION-003 | AGREED | Completed authorization decisions are allow/deny; inability to complete evaluation is a separate error, not a third authorization decision. Deny and error both prevent protected execution. | User approved Q-051, requested a concrete Auth-service timeout example, and confirmed it. Required authority loading times out with no sufficient valid authority already available: this proves inability to evaluate, not absence of permission. No prepared handoff, error transport, JSON schema, HTTP mapping, or cache policy adopted. See decision-results.md for rationale, alternatives, and consequences. |
| DECISION-004 | AGREED | Every completed deny must include an internal machine-readable reason for the calling endpoint and diagnostics. | User approved Q-052: “yes deny should include reason, very important” and reaffirmed recording rationale. A bare deny loses known cause information, forces callers to guess or reconstruct evaluation, and weakens diagnosis/audit. Reasons must reflect established conclusions, not mislabel timeouts or individual candidate failures. Exact codes/fields, precedence, and client disclosure remain open. See decision-results.md for rationale, alternatives, examples, and consequences. |
| SVG-001 | APPROVED PRESENTATION | Show one client request through authentication middleware, endpoint handler, embedded Auth Agent, Auth authority loading, constrained application database access, and the response. Keep middleware, handler, and agent inside the application boundary and retain the one endpoint-owned authorization gate. | User found the responsibility-first SVG difficult to follow, preferred the earlier flow style, and approved the proposed client-to-handler layout. Short numbered arrows show execution order; detailed principles stay in the overview. The Finance certificate example does not revive prepared results, mandatory resolvers, fixed remote calls, or a new decision schema. Previous SVG/overview are archived; see system-overview.md. This is a presentation decision, not a new authorization rule. |
| INPUT-003 | AGREED | Application request contracts define and validate input types, nullability, format, and domain meaning; endpoint policy does not duplicate type/nullable/validation-expression fields. Presence/source requirements remain INPUT-002. Auth retains canonical authority/scope checks, and authorization and execution must use the same validated meaning after application-defined parsing/normalization. | User confirmed Q-050-F and noted the endpoint shape was already discussed and approved. Rationale: avoid duplicate request schemas that can disagree. A required field may be nullable only under its explicit application contract; null never means unrestricted authority or permission to skip a boundary. The department_id example distinguishes wrong type/null/empty/missing cases. No universal string-only input rule or new policy field is adopted. See endpoint-policy-format.md. |
| INPUT-002 | AGREED | Every input listed in the endpoint policy must be present at its declared source. Reject a request missing a declared input; do not silently omit/default it or obtain it from another source. Required input presence is independent of the caller's grants, including tenant-wide {}. | User approved Q-050-E. Rationale: fixed predictable input contracts; optional-input absent behavior adds complexity and can weaken checks if mishandled. The PUT department_id example covers presence, omission, query fallback, broader grants, and additional application fields. Scope restrictions and actual endpoint enforcement remain separate from presence. Types, nullability, validation ordering, and error representation remain open; no optional/default fields or HTTP status are adopted. See endpoint-policy-format.md. |
| ENFORCEMENT-003 | AGREED | Endpoint review must verify that authorization boundaries and request bindings actually constrain the data returned or changed, not merely that their inputs are used somewhere. | User approved Q-050-D. Rationale: input usage is not effective enforcement; logging, ineffective OR filters, and unchecked output paths are counterexamples. Review Finance certificate-read for trusted tenant AND Finance AND requested certificate containment. {} does not create a department restriction. This is a central review requirement, not a guarantee against incorrect permissions, untrusted context, stale dependencies, or every other breach; mandatory checks/tests remain. See endpoint-policy-format.md. |
| CONTRACT-012 | AGREED | No relationship block, named-resolver, or argument-mapping contract is required in the adopted endpoint policy. The endpoint predeclares one required permission and selected inputs with sources. Auth establishes authority within supplied boundaries under all mandatory constraints; endpoint implementation must establish or enforce that actual execution stays within those boundaries. | User approved revised Q-050-C for simplicity, consciously assigning application relationship enforcement to endpoints. This is mandatory authorization enforcement, not optional business validation or permission to return unchecked data. The Finance grant/constrained-lookup example and original proposal not adopted are in endpoint-policy-format.md. It refines earlier CONTRACT-007/008 predeclared-relationship wording without changing registered meanings, tenant/self/delegation constraints, or the single endpoint-owned gate. Full policy validation/publication and update/move contracts remain open. |
| CONTRACT-011 | AGREED PARTIAL STRUCTURE | Endpoint policy fields are version, method, path, exactly one permission, and named inputs with explicit source/name bindings. Method/path identify the endpoint within the application; permission directly maps its operation without a separate action field. Local input names may differ from source field names. Path and selected body inputs are supported by this partial shape. | User approved Q-050-B and requested a PUT body example. Rationale: each field has one responsibility; explicit source/name avoids a shorthand selector grammar and does not confuse request inputs with established relationships. See endpoint-policy-format.md for GET/PUT, field justification, and counterexamples. Nested body selectors, further source kinds, full validation, relationship bindings, and update/move rules remain open. This is not a complete published policy schema. |
| CONTRACT-010 | AGREED | Published JSON/YAML contracts use a required top-level version field with a string value, initially "1" (quoted in YAML). It identifies the contract format/meaning, not the document edit revision, and is interpreted within its contract type. Reject missing, malformed, or unsupported versions without guessing or fallback. | User approved Q-050-A. Rationale: consistent representation and identifiable interpretation; independent contract-type evolution avoids unnecessary coupling. The alternative schema_version name was considered; version was selected for simplicity. Version is metadata, not a scope boundary key. Full schemas, compatibility/migration rules, document revisions, and future numbering remain open. See contract-publication.md for examples and counterexamples. |
| CONTRACT-008 | AGREED | Every protected endpoint predeclares exactly one required permission, inputs, sources, and how to establish any required relationship. Auth mandates and enforces exactly one permission: zero/multiple required permissions and missing/invalid declarations cannot permit execution. The permission must cover the complete protected operation. | User approved revised Q-049 for design simplicity. The earlier AND-across-multiple-permissions proposal was not adopted. Multi-permission grants and multiple direct/group grants remain supported; complete grants remain alternative routes for the one requirement. Narrow permission checks cannot authorize unrelated privileged work. CONTRACT-007's plural wording and earlier open-combination notes are historical after this refinement. Endpoint policy syntax and validation timing remain open. See endpoint-authorization.md for rationale, examples, and the preserved alternative. |
| CONTRACT-009 | AGREED | Every published JSON/YAML contract, including grants and endpoint policies, must include a version. Existing working illustrations are not complete published contracts. | User explicitly required versioning and noted the endpoint policy contract has not yet been discussed. Rationale: consumers must be able to identify the contract definition used, avoiding ambiguous interpretation as contracts evolve. This does not select a field name, version value/format, placement, compatibility policy, or scope key. See contract-publication.md; endpoint policy contract discussion is Q-050. |
| CONTRACT-007 | AGREED | Each protected endpoint's single server-owned declaration identifies its required permission and required authorization material, including where that material is obtained. Distinguish request parameters/proposed values from verified context and authoritative application facts; shared authority/context infrastructure need not be duplicated. | User answered Q-035: "yes." Extends INPUT-001/CONTRACT-006 without adopting new fields, declaration syntax, or another decision phase. The gate gathers sufficient material rather than eagerly fetching every possible fact. system-overview.md and its SVG map the agreed logical model. Q-047-A explicitly clarifies: the endpoint predeclares permissions, inputs, sources, and how to establish any required relationship. This declaration is fixed; relevant scope and other mandatory checks determine the material needed. Scope need not repeat every input; {} adds no narrower boundary within the tenant. Rationale: distinguish request identification from authority restriction and avoid unnecessary fact gathering without losing operation binding. See endpoint-authorization.md. Plural permissions does not settle their combination semantics. |
| RESOLUTION-005 | AGREED | Authorization request means the declared operation, verified identity/tenant context, and selected request inputs. Resolved request means an evaluation view of that request with sufficient trusted material for the relevant authorization boundaries, not an allow result. Keep request claims distinguishable from established facts. | User agreed to Q-047 and approved its Q-047-A clarification. Rationale: request inputs are not relationship proof, and fact gathering does not create authority. Fixed endpoint declarations coexist with selective material resolution: {} does not require narrower scope facts, while tenant, validity, conditions, delegation, and application-contract requirements remain. No new persisted entity, prepared handoff, JSON schema, or mandatory eager lookup is adopted. |
| PROCESS-006 | AGREED | Resume recording and checkpoint commits/pushes after the discussion pause. Preserve earlier designs and mark them deprecated rather than deleting or silently replacing them. | User approved Q-033: "we should start recoding. do not replace earlier one just mark them as depricated. i approve the above." Discussion-only Q-030–Q-033 is now captured with its actual approval status. |
| SCOPE-006 | AGREED | Scope is a boundary selector that defines the reach of granted permissions within the enclosing tenant. Request material must establish that the requested operation remains within the selected boundary. Permission covers the operation; boundary satisfaction alone is not the entire authorization decision. | User defined scope as a boundary selector, affirmed the permission distinction, and requested a canonical definition. Q-030/Q-031 capture the pause-period discussion. No query language or new fields follow from this definition. Q-043 / TERM-005 aligns vocabulary without weakening trusted-fact or actual-use requirements; original wording is preserved in history/q043-vocabulary.md. |
| SCOPE-007 | AGREED | Canonical v1 scope is a required flat JSON object of defined boundary keys and non-empty string values: concrete references or reserved $self for the authorizing human where supported. Entries combine with AND; order is irrelevant. Explicit {} is tenant-wide reach within the trusted enclosing tenant. Missing/null/non-object scope, unknown/duplicate keys, empty/non-string values, unsupported symbols, arrays, nested objects, and wildcard operators are invalid. Never drop invalid restrictions or default invalid/missing scope to {}. | User answered Q-034: "agreed" after the minimal format and empty-object distinction were presented. Scope syntax is canonical; key definitions/governance, containment, lifecycle, complete grant/request schemas, and implementation remain separate. GRANT-EX-007 uses current scope syntax; earlier typed scope examples are preserved as deprecated history. |
| SCOPE-008 | AGREED | Entries/requirements within one scope combine with AND. Alternative authority is supplied through separate complete applicable grants, each with its own scope and other restrictions. Do not merge unrelated grant fields. | User proposed AND and separate grants for OR in Q-032, then approved the reviewed model in Q-033. Consistent with DECISION-001 and GRANT-001. Does not finalize the candidate key-value grammar or override mandatory tenant/delegation limits. |
| CONTRACT-006 | AGREED | One endpoint-owned authorization gate gathers sufficient material, invokes a shared evaluator, and enforces its decision before protected execution. Middleware may authenticate, establish context, and load authority, but does not issue a business-authorization decision requiring endpoint completion. No cross-layer prepared outcome or two-mode selection. Keep one authoritative auth-first endpoint declaration. | User approved Q-033 after review of earlier cases. Sufficient does not mean every fact: the gate can deny without unnecessary lookups. Deprecates CONTRACT-002/003 and the mode-specific parts of earlier contracts; earlier records/prose are preserved. Shared semantics, non-amplification, and human-dependent authority remain. |
| ENFORCEMENT-002 | AGREED | Before authorization, constrain internal fact gathering to necessary tenant/request-bound work; no protected output, mutation, or business side effect may occur. Enforce the decision against the actual operation and applicable facts. Missing required material or evaluation failure cannot permit execution. | Approved with CONTRACT-006 in Q-033. Carries ENFORCEMENT-001's safety invariant forward without prepared/middleware-allow wording. Collection enforcement need not preload all rows; exact contracts, consistency, and freshness remain open. |
| INPUT-001 | AGREED | The server-owned method/route declaration maps the endpoint to an action and required permission, and identifies which path and body parameters contribute authorization request inputs. Combine those inputs with verified identity/tenant context, Auth-owned authority, and required application facts at the endpoint-owned gate. | User clarified material is usually method-to-action plus path and identified body parameters. Selected inputs do not prove existing resource relationships. HTTP method alone does not determine every operation's permission. Exact binding schema and current-versus-proposed state semantics remain open. |
| SCOPE-005 | DEPRECATED | Scope describes authorized boundary selection; where supported the application may translate it into a safe query restriction, intersected with the requested selection. A fully determined restriction can be enforced by a middleware-complete endpoint; endpoint-completion obtains facts and completes evaluation under its fixed contract. | User asks for route/JSON examples and proposes selector/query-like scope. Q-029 tests the flow in scope-model.md, reusing agreed CONTRACT-002/005. Does not adopt arbitrary SQL in grants, a query grammar, universal query compilation, or dynamic endpoint modes. |
| SCOPE-003 | AGREED | Scope owns boundary-selection semantics; a grant binds its use to recipient and permissions without defining its domain meaning. Application-defined concepts such as department are optional; shared concepts such as self require explicit application/resource relationships. Authorization validates and evaluates scope through its definition. | User answered Q-027: "agree." Establishes responsibility separation, not a new entity, field, fixed catalog, general expression grammar, independent storage requirement, or mutable scope-reference model. SCOPE-002's representation remains open. See scope-model.md. |
| SCOPE-004 | ~~PROPOSED~~ PARTIALLY AGREED | A scope may be used for a resource only where its boundary-selection meaning for that resource is explicitly defined. Do not infer compatibility from matching field names or use an unrelated relationship. Validate scope/resource compatibility when binding permissions; unsupported or unresolved use must not establish authority. | Q-028 tests department scope on payslips, certificates, and repositories. Supports reusable meanings with explicit resource mappings, not a mandatory per-resource scope type. Multi-permission compatibility, role evolution, declaration location, and failure handling require detailed follow-up. Q-038 now agrees the application-owned meaning, explicit supported resource relationships, endpoint fact bindings, and no implicit unrestricted access for unsupported relationships. The earlier proposed grant-binding validation point and detailed compatibility mechanisms remain open; Q-039 addresses shared definitions. |
| SCOPE-002 | PARTIALLY DEPRECATED | A grant selects an explicitly defined scope type and supplies only its permitted inputs; it does not invent the scope meaning or introduce arbitrary interpreted fields. Scope definitions must make applicable resource relationships and accepted inputs explicit. | Q-026 compares declared scope types with a general expression grammar using existing department and employee_self examples. No new fields or finalized schema are adopted. Definition ownership, registration, compatibility, composition, and containment remain open. Its type-plus-parameters representation is historical after SCOPE-007; definition governance remains open. See scope-model.md. |
| PROCESS-005 | AGREED | Justify every proposed new field: establish the needed distinction, explain its semantics, and check whether existing concepts already express it before adoption. Illustrative JSON is not approval of new canonical syntax. | User challenged permissions_subset_of, scope_within, recipient inside scope, and grant_selector: "we need to be sure why we add any new field". G-11's unreviewed scope syntax is withdrawn; semantic needs and representation remain separate. |
| ADMIN-005 | ~~PROPOSED~~ AGREED AT RULE LEVEL | Use the same canonical grant binding for administration: recipient is the administrator user/group, permission is an operation on Auth resources such as grant creation, and scope selects the grants that may be administered. How to express proposed-grant recipient, permission, and scope limits must be justified through the shared scope model, not silently introduced as new syntax. | User challenged both Q-023's standalone format and Q-024's new nested fields. G-11 syntax is withdrawn under PROCESS-005. The unified model remains the working direction; exact scope grammar, bootstrap, containment, and lifecycle rules are not finalized. Q-044 now approves the ordinary model, distinct administrative operations, and continued normal validation. Exact encoding stays open. |
| ADMIN-004 | ~~PROPOSED~~ AGREED AT RULE LEVEL | Authorizing grant creation requires the requested administrative operation and the whole proposed grant to fit applicable administrative authority: eligible recipient, assignable permissions, resource reach, and required validity/conditions. Bounds remain associated; unrelated administrative grants cannot donate fields to manufacture broader authority. | Q-023 tests Finance payroll-read provisioning to one approved group. Administrative operation and business permission being assigned are different layers. Exact bound representation, role expansion/change handling, relationship-scope containment, multi-route administration, update/revoke rules, and onward administration remain open. Q-044 now approves complete requested-assignment checks, relevant existing/proposed material for changes, and associated administrative bounds. Encoding and containment mechanics stay open. |
| ADMIN-002 | AGREED | Authority to assign access is separate from authority to use that access. An administrator need not personally possess the business access being assigned, but assignment must remain within authorized administrative bounds. | User answered Q-022: "we should keep them seperate can provide access does not mean can access". Rejects requiring personal business-access possession as a universal prerequisite. Does not relax human-dependent service/agent limits. |
| ADMIN-003 | AGREED | Administrative authority must not implicitly supply business access to the administrator. Self-assignment is not categorically forbidden: it must be an explicit access-changing operation, authorized within the administrator's bounds, and captured in the audit trail. | User noted that providing access may include providing it to themselves, "but then that has to be explicit, which audit will capture". Audit supplies accountability, not authorization or prevention by itself. No additional approval workflow is adopted. Audit schema, delivery/integrity guarantees, indirect group/role effects, and administrative bounds remain to be specified. |
| ADMIN-006 | AGREED | A validly issued ordinary human/group grant does not become invalid merely because its issuing administrator later loses grant-administration authority. The grant's own validity, scope, conditions, revocation, and applicable recipient dependencies continue to govern it. Retain issuance provenance for audit, not as an implicit continuing authority dependency. | User agreed to Q-046. Rationale: authorized administration is not lending personal access; routine administrator rotation should not unexpectedly withdraw organizational access. Cascading withdrawal on issuer-authority loss was considered and not selected. Removing issuance authority does not clean up earlier grants; intended withdrawal requires explicit authorized revocation. Group-derived and human-dependent automated authority retain their live dependencies. Invalid issuance and incident-response mechanics are not settled here. See grant-model.md for the Maya/Vinay example and counterexamples. |
| RESOLUTION-006 | AGREED | A resolved grant is an evaluation-ready view of an existing grant with relevant references and human context established, preserving source identity, associated permissions/scope/validity/conditions, and applicable membership/delegation dependencies. Obtain the human's valid Auth-owned memberships and retrieve direct grants plus grants for those groups; views remain tied to the original bindings. Resolved does not mean allowed. | User approved Q-048 and explicitly added Vinay's membership list followed by grants for Vinay and his membership teams. Rationale: flattening loses restrictions; creating independent direct grants loses membership dependency. Role expansion and self anchoring clarify authority meaning, not request boundary satisfaction. Logical flow does not mandate separate API calls, eager fetching, a transport schema, or a cache/freshness mechanism. The grant chapter records G-17, counterexamples, and open mechanics. |
| PROCESS-004 | AGREED | Commit and push meaningful documentation checkpoints throughout the discussion. Periodically reconcile the chapters, decision log, discussion tree, examples, and original handbook; retain superseded history and explicitly unresolved questions. | User instructed: "keep commiting and pusing. every now and then we need to reconcile". Reconciliation must not silently settle open policy choices or imply the application implements the handbook. |
| PLAN-001 | AGREED | Follow the eleven-stage roadmap and refine the handbook through focused discussions, worked examples, and recorded decisions. The assistant steers the discussion. | User approved the proposed roadmap on 2026-09-05: "looks good pin it. you steer the discussion." |
| PROCESS-001 | AGREED | Give every proposed rule and discussion question a stable reference ID. Treat suggestions from either participant as candidates until deliberately agreed; use examples to assess the model and the need for canonical terms. | User requested reference numbers and clarified: "even i am proposing and we need to see if the shape is getting right." |
| PROCESS-002 | AGREED | Maintain a traversable discussion tree covering all eleven stages, with conclusions, open siblings, detours, and explicit return points. Revisit every unfinished branch before declaring the handbook complete. | User requested a "traverse tree" so branches conclude and the entire tree is covered. The navigation artifact is discussion-tree.md. |
| PROCESS-003 | AGREED | Notes must retain sufficient detail to reconstruct the discussion: definitions, rationale, examples/counterexamples, consequences, links to decision IDs, and unresolved questions. Consolidate settled branches into working handbook chapters while retaining the decision log and tree. | User instructed: "make sure what you note down have suffeint details." grant-model.md consolidates the grant/role branch; it does not finalize unresolved schemas or implementation contracts. Reaffirmed after Q-045 as a standing instruction across sessions: "keep this in you memory always record with rationale." Every new or revised decision must explain what was agreed, why, relevant alternatives/tradeoffs, examples, consequences, and what remains open; a bare approval or status change is insufficient. This durable repository rule preserves reasoning when conversation context is unavailable. |
| CHARTER-001 | AGREED | The handbook defines mandatory shared authorization semantics for every AgentLabs application. Applications define their own resources, permissions, and relationships within those rules; storage and implementation may differ. | User answered "Yes" on 2026-09-05 to requiring the same core authorization rules across applications. Individual rules remain to be settled in the following stages. This decision alone does not require a central decision service or one policy engine. |
| CHARTER-002 | AGREED | Auth is the shared authority for identity, tenant membership, roles, and authorization assignments. Each application's authorization layer resolves domain facts and evaluates access under the shared handbook rules; the application's protected execution and data-access paths enforce the result. | User answered "yes" on 2026-09-05 to this responsibility boundary. Shared authorization libraries are compatible with it. This does not settle deployment topology, wire formats, or the details of grant resolution. |
| PRINCIPLE-001 | AGREED | A protected operation may proceed only when the system establishes valid authority covering the requested action and resource within the applicable tenant boundary. If required information is missing, invalid, or unavailable and authority cannot be established, the operation must not proceed. | User agreed on 2026-09-05: "yes, this is right." Authentication or knowledge of a resource ID does not by itself establish access. Failure to establish authority must remain distinguishable from a completed evaluation that denies access. Exact grant-combination, freshness, and response rules remain open. |
| PRINCIPLE-002 | AGREED | Applications and endpoints are designed with an auth-first approach: authorization requirements guide the operation, URI, required domain facts, handler contract, and data-access enforcement from the start. | User instructed: "the applications or endpoints should be written with auth first approach." Extends CONTRACT-004's one-declaration rule to application design. This does not authorize implementation changes to existing applications in this discussion. |
| TENANT-001 | AGREED | Tenant is the enclosing authorization boundary and is implicit in a grant evaluated within trusted tenant context. Recipients, groups, scope references, and resources resolve within that boundary; grants cannot select or widen it. The authorization system must preserve the binding even when the grant payload omits a tenant field. | User instructed: "Tenant should be more implicit as this is the outer most boundry is implied." Grant examples omit repeated tenant IDs. Tenant-wide scope means the enclosing tenant. Storage, cache, and transport representations remain to be designed; implicit context is not permission to drop isolation or trust a caller's tenant assertion. |
| ARCH-001 | PROPOSED | Use a reusable authorization core with application-specific operation mappings and trusted domain-fact resolvers. The agent consumes Auth-owned authority information and application-owned facts; this does not require two network calls per request. | Clarifies CHARTER-002. ARCH-002 refines where facts enter: the generic middleware does not access the application database; the endpoint/application service supplies domain facts to the shared evaluator. Neither proposal is an implementation-status claim. |
| ARCH-004 | AGREED | Two responsibility layers jointly establish effective authorization: Layer 1 supplies canonical authority rules and shared evaluation semantics; Layer 2 supplies application-specific meanings, trusted facts, additional restrictions, and enforcement. Layer 2 cannot manufacture authority or override Layer 1 constraints. | User approved Q-036. These are responsibility layers, not two decision locations: CONTRACT-006's single endpoint-owned gate remains current. See system-overview.md for rationale and the payslip example. Auth-service versus application runtime placement is a follow-up clarification, not a new deployment contract. |
| ARCH-005 | AGREED | Auth primarily supplies authority material; the application supplies domain meanings, facts, declarations, and additional restrictions. The application-embedded auth agent integrates both using shared canonical evaluation rules; the endpoint enforces the result at one gate. Layer 1 is not confined to the Auth service. | User approved Q-037. Reusable integration consumes application-provided bindings rather than knowing every application database. Auth middleware may describe this overall integration, but pre-handler HTTP middleware alone need not have sufficient material. No new network-call count, integration API, or two-stage decision is adopted. Clarifies ARCH-004; older ARCH-001 remains historical proposal text rather than silently approved wholesale. |
| ARCH-002 | DEPRECATED | Generic middleware obtains Auth-owned authority, evaluates what it can without application-database access, and stops conclusive denials. The application supplies remaining domain facts and enforces authorization. Whether middleware returns a final allow versus prepared context is under renewed discussion. | User subsequently clarified: "I tentively agree to allow/prepare model, deny model is clear." Earlier acceptance of the three-outcome formulation must not be treated as final. See ARCH-003 and CONTRACT-001 for the certificate-route example. |
| ENFORCEMENT-001 | DEPRECATED | A prepared context is not an allow decision. For a prepared request, the endpoint may perform only the internal, tenant/request-constrained work needed to establish required facts before completing authorization. Missing facts or an omitted required check cannot yield protected output or a business mutation. A completed middleware allow still requires enforcement of its action, resource, and restrictions by the application. | Accepted as part of Q-010 with ARCH-002's refinement. Bind facts and the decision to actual use; avoid changing relevant facts between checking and use. Query predicates may combine final evaluation and access for collection operations; exact enforcement contracts remain open. |
| RESOLUTION-002 | SUPERSEDED | Earlier candidate: middleware dynamically chooses completed allow, deny, or prepared for any endpoint. | CONTRACT-002 replaces the universal three-outcome proposal with two endpoint-declared modes. Early conclusive denial and fail-closed evaluation errors remain applicable. |
| ARCH-003 | SUPERSEDED | Earlier candidate: every admitted request is treated as prepared for mandatory handler completion. | CONTRACT-002 limits mandatory authorization completion to endpoints explicitly declaring it; middleware-complete endpoints receive final allow or denial, never prepared. Endpoint enforcement remains required in both modes. |
| FACT-001 | PROPOSED | A route parameter supplies a requested identifier or boundary, not proof of relationships. A server-owned route mapping gives the parameter meaning; the application establishes that the actual resource satisfies the bound tenant, department, and any applicable ownership constraints through trusted facts or a constrained query. | For GET /api/v1/{tenant}/{dept}/{certificate}, the route does not prove that the certificate belongs to the named department or user. Ownership is required for self-scoped access, not automatically for every department-scoped grant. |
| CONTRACT-001 | DEPRECATED | Each protected operation declares its permission mapping, the meaning of its request bindings, and how the required authorization facts or enforceable predicates are obtained. Applicable scope/delegation evaluators report requirements still unresolved. Middleware relies on those contracts to distinguish proven conditions from pending ones; it cannot infer arbitrary business facts from URL structure. | Explicit contracts allow predictable handling of supported operations. Undeclared or unsupported checks fail closed rather than being silently interpreted as satisfied. This is a contract proposal, not a finalized schema or a requirement to enumerate all resource IDs. |
| CONTRACT-002 | DEPRECATED | Each endpoint/operation declares one authorization resolution mode. Middleware-complete mode permits final allow or deny; there is no prepared result, and the handler enforces the completed decision. Endpoint-completion mode permits middleware deny or prepared only, never middleware allow; the handler supplies declared application facts, completes evaluation, and enforces the result. The mode is fixed by the server-owned contract, not selected by the caller or switched per request. | User accepted the single-declaration shape: "looks good," then asked for cases beyond self. The estimated 90–95% coverage is an unverified hypothesis, not a requirement or measured result. Final field names and schemas remain open. |
| CONTRACT-003 | DEPRECATED | Validate an endpoint's declared mode against all supported authorization requirements, including scope and delegation forms. Middleware-complete mode cannot silently permit or fall back to prepared when a required fact is unavailable; incompatible configuration or runtime evaluation failure prevents protected execution. Endpoint-completion mode always follows its completion contract even if a particular grant appears broad enough for an early allow. | Makes CONTRACT-002 predictable while preserving PRINCIPLE-001. Changes to operations or supported scopes require checking the mode contract again. Mode declaration does not prove resource ownership or eliminate mandatory data restrictions. |
| CONTRACT-004 | PARTIALLY DEPRECATED | Design the URI, handler, and authorization declaration together. Each HTTP method plus route template has one authoritative authorization contract with one resolution mode. It declares the operation/permission mapping, request bindings, supported authorization requirements, and mandatory handler completion/enforcement. Middleware and handler use that same contract rather than separate conflicting declarations. | User replied "looks good" to the refined one-endpoint/one-workflow rule. Different grants or additional checks within the same declared contract do not require extra endpoints; incompatible workflows require a separate endpoint or deliberate contract revision. A different endpoint does not automatically require a different permission. |
| CONTRACT-005 | DEPRECATED | Select endpoint-completion when required application facts are still needed to finish deciding authorized access. A handler merely applying already determined, complete mandatory restrictions is enforcement and does not alone require endpoint-completion. Evaluate this distinction against every supported authorization form for the endpoint. | User responded "looks good" to the beyond-self review. Examples EC-001 through EC-007 are in endpoint-completion-cases.md; acceptance of the distinction does not adopt every illustrative policy as a product requirement. Business validation is not automatically authorization, and no coverage percentage is established. |
| GROUP-001 | ~~OPEN~~ AGREED: GROUP-001-A | Decide ownership and resolution of groups and teams. GROUP-001-A (refined working proposal): Auth owns generic authorization groups and their memberships; applications may synchronize business memberships into Auth or keep them independent. Authorization group membership is obtained from Auth under the freshness contract, not independently inferred from application business membership. GROUP-001-B (alternative): applications own some authorization-group memberships, consumed through explicit resolver integrations. | User proposed relying on Auth for groups/teams and membership, with optional application-owned synchronization. GROUP-001-A is the recommended working shape, still a proposal under PROCESS-001. Business department, ownership, and containment facts remain application-owned. Current implementation support remains unverified. Later Q-045 approves GROUP-001-A; the earlier proposed-status wording is historical and GROUP-001-B is not selected. Rationale: one authorization-membership authority, with deliberate optional sync rather than inference from business membership. See groups-and-membership.md. |
| SYNC-001 | OPEN | For a group explicitly synchronized from business membership, define the authorized writer, update delay, removals, retry/reconciliation, and behavior when synchronization is stale or fails. | Synchronization is optional to configure, but a configured relationship needs a correctness contract. Until a change reaches Auth and relevant authorization caches, old membership may still confer access. Default denial alone cannot detect an unknown upstream change. Resolve timing and failure guarantees in stage 9. |
| TERM-001 | AGREED | "Team" and "group" mean the same authorization concept. Use "group" as the handbook's canonical term and "team" as a synonym, with no separate authorization behavior. | User answered Q-002 on 2026-09-05: "team/group should mean same." This settles synonymy; member types, nesting, and exact grant-inheritance rules are separate decisions. |
| GROUP-002 | NOT ADOPTED | Original assistant proposal: a group may contain human users and service/agent principals with valid memberships in the group's tenant. | User's Q-003 response favors human-only groups; replaced as the working direction by GROUP-003. Preserve this entry as discussion history. |
| GROUP-003 | ~~PROPOSED~~ AGREED | Groups/teams contain human users only. Service accounts and agents are not first-class group members. | User's proposed direction: "group should be purly for humans"; automated access uses explicit service-account assignments or human delegation. Nested-group behavior is not established by this statement. Later Q-045 approves human-only membership; historical service-assignment wording cannot revive independent automation under AUTHORITY-002. Rationale: group-derived automated access stays through the human, preserving membership dependency and proxy limits. See groups-and-membership.md. |
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
| REGISTRATION-001 | AGREED | Applications register supported permissions and scope contracts with Auth. Auth validates grant acceptance against registrations and canonical authority rules without interpreting application-specific business meaning. Application bindings supply domain interpretation and trusted facts for embedded-agent evaluation. | User approved refined Q-039. Auth still checks authority to create/change grants; valid syntax is not issuance authority. Registered references do not prove domain-object existence, and separately registered permissions/keys do not imply compatibility. Registration format and permission-scope compatibility declarations remain open. See application-registration.md. |
| REGISTRATION-002 | AGREED WITH QUALIFICATION | The application's registration flow binds permissions and scopes to declared support relationships. Declaring these relationships is optional; supplied contracts support Auth-side compatibility validation without domain interpretation. | User approved Q-040 and required the relationship-declaration feature to be optional. Individual permission/scope registration remains required under Q-039. No grant fields or API sequence adopted. Q-041 separately proposes omitted-declaration behavior; optionality granularity and detailed compatibility mechanics remain open. |
| REGISTRATION-003 | AGREED | The application explicitly declares upfront whether relationship validation is enabled. When enabled, every grant must pass declared permission-scope relationship checks, without per-grant bypass; this includes role-based and existing grants. When disabled, other registration/canonical/issuance checks remain, and runtime scope enforcement is mandatory in either mode. | User approved refined Q-041. Missing metadata cannot disable an enabled mode. Supersedes choosing behavior merely from metadata omission. No field name, missing-choice default, or migration mechanism adopted. Registration/role-change revalidation details remain open; Q-042 begins enabling validation with existing grants. |
| REGISTRATION-004 | AGREED | All existing grants must pass before relationship validation is enabled. Otherwise reject activation, report incompatible grants, and preserve the previous configuration pending explicit authorized correction of grants or declarations. | Q-042 agreed. No silent grant deletion/modification or grandfathering of incompatible grants. User directed moving branches; remaining registration lifecycle mechanics are parked. |
| SCOPE-001 | AGREED | A scope describes the set of resources over which an assigned permission may apply. It may express explicit selections or selectors based on trusted resource attributes and relationships; it need not enumerate IDs or follow one hierarchy. A resource is what the request seeks to act on, checked against that reach. | User answered Q-009: "Yes agree to this." Illustrations: exact payslip P-17; payslips owned by the principal's employee; payslips associated with Finance. Final selector grammar, temporal relationship semantics, and request material for collection/create operations remain open. A scope description is not itself an allow decision. |
| GRANT-001 | AGREED | A grant binds a recipient, permission or role reference, scope, and validity/conditions as one authority unit. Resolution must preserve which scope and conditions belong to which capability; it must not combine a capability from one grant with broader reach supplied only by a different-capability grant. | User answered Q-012: "agree" and requested JSON. GRANT-EX-001 in grant-examples.md illustrates tenant-wide read and Finance-only revoke. This does not settle grant versus assignment terminology, role expansion, same-permission grant combination, or dependent service grant representation. |
| GRANT-002 | AGREED | A grant may bundle one or more permissions when they share its recipient, scope, validity, and conditions. Permissions requiring different reach or conditions use separate bindings. | User said "proceed" after this proposal and its example. GRANT-EX-003 illustrates the concept; the exact permissions-array wire encoding remains illustrative. Role references and role expansion remain separate discussions. |
| GRANT-003 | AGREED | A human may have multiple directly assigned grants and grants applicable through valid group memberships in the enclosing tenant. Resolution gathers applicable bindings and retains their permission/scope/condition associations and direct/group provenance. | User said "proceed" after the direct/group grant example. Group-derived access depends on membership; loss of that membership does not erase separately valid direct grants. This does not settle explicit-deny precedence, general grant-combination semantics, or nested groups. |
| ROLE-001 | AGREED | A role is a named reusable permission bundle. A grant references either such a role or an explicit permission set and binds it to a recipient, scope, validity, and conditions. A role definition supplies capabilities; the grant supplies their assigned reach. Group is the collection of human recipients, role is the collection of capabilities, and grant binds recipient to scoped authority. | User answered Q-013: "makes sense." GRANT-EX-004 illustrates this. Role-change behavior is settled by ROLE-002. Role templates/defaults, scope compatibility, and revision mechanics remain open. |
| ROLE-002 | AGREED | A role-referencing grant uses the role's current permission definition. Authorized role edits therefore change capabilities available through existing referencing grants while preserving each grant's scope, validity, and conditions. Unrelated direct grants remain separate. | User answered Q-014: "yes." Grants do not require manual upgrades to adopt a role edit. Who may make such edits, compatibility validation, revision evidence, propagation freshness, and dependent delegation limits remain separate requirements. Role editing changes declared authority; resolution must not independently invent permissions. |
| RESOLUTION-003 | AGREED | Expand a role reference into its current permission set, producing the same permission-set grant form used to evaluate explicitly listed permissions. Preserve the original grant identity, scope, recipient, validity, conditions, and source provenance, including the role definition used. Expansion is a computed view of the same grant, not a new assignment, and does not merge unrelated grants. | User agreed after clarification that the earlier JSON represented the existing G-6 after role lookup. GRANT-EX-005 retains G-6 visibly. Expanded does not mean fully resolved: application relationships or resource facts may still be missing. Exact schema, role-revision encoding, membership expansion, and final resolved-grant representation remain open. |
| TERM-004 | AGREED | Use "grant" as the canonical name for the authorization binding record. "Assignment" refers to that same binding when describing a permission/role being assigned; "assign" is the action creating the binding. Do not introduce a second logical assignment object solely because both words appear in explanations. | User answered Q-016: "agree." Rationale, examples, and consequences are in grant-model.md. This is a logical vocabulary choice, not a requirement for one physical database table. |
| TERM-005 | AGREED | Describe authorization through permissions, scope boundaries, requests, and sufficient trusted request material, without an additional canonical entity or wrapper. Material must remain bound to the actual operation; all grant, tenant, validity, condition, and delegation constraints still apply. | User rejected the extra vocabulary in Q-043, requested a tenability check, and agreed after the read/list/create/move/grant-administration explanation. The earlier administration framing is withdrawn, not renamed into another mandatory object. Canonical administration bounds remain open. Exact superseded wording is preserved in history/q043-vocabulary.md; reasoning is in authorization-vocabulary.md. |
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
| Q-022 | ANSWERED | Yes: providing access and using access are separate. Providing oneself access must be explicit and audited, not implicitly inherited from administration. | ADMIN-002 and ADMIN-003 agreed; exact administrative bounds and audit guarantees remain open. |
| Q-023 | ANSWER INCORPORATED | User challenges the apparent new format and proposes expressing administration with the existing grant, permission, and scope concepts; some initial grants may be seeded during bootstrap. | ADMIN-005 / Q-024 refines the representation. ADMIN-004's bounds remain proposed, not silently agreed by this feedback. |
| Q-024 | ANSWER INCORPORATED | User challenges the newly introduced scope fields and requires justification for every addition. | PROCESS-005 records the requirement. ADMIN-005 remains proposed; withdraw G-11 syntax and explain intended constraints independently of representation. |
| Q-025 | ANSWERED | Yes: settle the shared scope model before returning to administrative JSON. | Stage 6 is active; return to ADMIN-004/005. scope-model.md retains the discussion sequence, definitions, examples, alternatives, and open decisions. |
| Q-026 | ANSWER INCORPORATED | User notes that department is not universal, suggests standard and application-specific concepts, and proposes that scope owns its definition independently of grants. Asks whether treating scope as a selector should change the canonical arrangement. | SCOPE-003 / Q-027 develops the proposal. SCOPE-002's type/parameter shape is not treated as agreed. |
| Q-027 | ANSWERED | Yes: scope owns boundary-selection meaning; grant binds its use without defining it. | SCOPE-003 agreed; storage, scope catalog, grammar, and representation remain open. |
| Q-028 | ~~OPEN~~ ANSWERED BY Q-038 | Should a scope be usable only with resources for which its boundary-selection meaning has been explicitly defined, rather than inferred from a shared name or field? | SCOPE-004 proposed; distinguish scope compatibility from authorization to perform the operation. Later Q-038 agrees this principle; it does not settle all of SCOPE-004's validation mechanisms. |
| Q-029 | CLOSED BY MODEL CHANGE | Does scope as authorized selection, narrowed by the request and enforced through a constrained query where supported, match the intended middleware/endpoint model? | SCOPE-005 was not agreed and is now deprecated under CONTRACT-006. Q-028 remains open after requests for clearer explanation; the concrete route/JSON walkthrough does not introduce new scope fields. |
| Q-030 | ANSWERED | User defines scope as a boundary selector and confirms permissions separately determine the operation. | SCOPE-006. Discussed while recording was paused; now recorded under PROCESS-006. |
| Q-031 | ANSWER INCORPORATED | User requests canonical boundary-based scope definition and proposes minimal key-value syntax. | SCOPE-006 agreed; SCOPE-007 representation proposed. |
| Q-032 | ANSWERED | AND within one scope; OR through separate grants, each with its own scope. | SCOPE-008 agreed through the subsequent approval of the reviewed model. |
| Q-033 | ANSWERED | Yes: approve one endpoint-owned authorization gate, shared evaluator, and no prepared authorization outcome. Resume recording and deprecate, rather than delete, earlier designs. | CONTRACT-006, ENFORCEMENT-002, PROCESS-006. INPUT-001 captures the accompanying endpoint-material clarification. |
| Q-034 | ANSWERED | Agreed: adopt the minimal v1 format; explicit {} means tenant-wide reach for the grant's permissions, while omitted or null scope is invalid. | SCOPE-007 canonical. No automatic default to {} or silent dropping of invalid restrictions. Earlier syntax remains preserved as historical examples. |
| Q-035 | ANSWERED | Yes: the endpoint declaration identifies required permission, material, and sources. | CONTRACT-007 agreed. User additionally requested an SVG diagram; assets/authorization-system.svg visualizes the logical model, not a new deployment requirement. |
| Q-036 | ANSWERED | Agreed: canonical Layer 1 and application-specific Layer 2 jointly establish authorization; Layer 2 supplies meanings, facts, and restrictions within Layer 1 authority. | ARCH-004 agreed. User then suggested that most Layer 1 material comes from Auth, Layer 2 from the application, and the embedded auth agent works across both. Clarify integration terminology without reviving the deprecated middleware/endpoint decision split. |
| Q-037 | ANSWERED | Agreed: Auth supplies authority, the application supplies domain meaning and facts, and the embedded auth agent evaluates across both with endpoint enforcement. | ARCH-005 agreed. Distinguish logical responsibility from physical placement and embedded integration from pre-handler-only middleware. Return to scope-key definitions and ownership. |
| Q-038 | ~~OPEN~~ ANSWERED | Should an application define a scope key's boundary meaning and supported resource relationships explicitly, with endpoints binding trusted material to that meaning rather than inventing a meaning or inferring it from matching field names? | Refines existing SCOPE-003 and proposed SCOPE-004 under ARCH-004/005. Use dept across payslips, certificates, and unsupported repositories as the example. No definition-registry format or new grant fields proposed. User agreed. Application-owned meanings and endpoint-specific trusted fact bindings are now settled; illustrative department relationships are not a universal application catalog. |
| Q-039 | ~~OPEN~~ ANSWERED AS REFINED | Should application-owned scope definitions be shared as one contract for grant validation and request evaluation, so neither side independently invents accepted keys or supported boundary meanings? | Historical proposal: Proposed governance principle only. It does not require Auth to query application data, copy domain records, or execute application code. Definition transport, storage, versioning, concrete-reference checks, and exact validation timing remain open. No new grant fields proposed. Subsequent refinement and approval: applications register both supported scopes and permissions; Auth validates grants before accepting, without application-domain interpretation. REGISTRATION-001 records this approved principle; compatibility declarations and registration syntax remain open. |
| Q-040 | ~~OPEN~~ ANSWERED WITH QUALIFICATION | Should the application explicitly declare which scope keys are supported for each permission, so Auth can reject unsupported permission-scope combinations without interpreting domain meaning? | Historical proposal: Next compatibility proposal under SCOPE-004 and REGISTRATION-001, not yet approved. Registering a permission and key separately is not evidence of compatibility. No payload fields adopted; multiple permissions, role evolution, and permitted multi-key combinations need subsequent detail. User approved with the qualification that registration binds these relationships and declaring them is optional. REGISTRATION-002 records that qualification; omission behavior remains Q-041. |
| Q-041 | ~~OPEN~~ ANSWERED AS REFINED | Historical question: If relationship metadata is omitted, should Auth perform its other registration/canonical/issuance checks without claiming compatibility, leaving supported resource relationships and scope satisfaction to the endpoint-owned gate; when metadata is supplied, additionally validate against it? | Historical rationale: Recommended interpretation of optional declarations, not yet agreed. Optionality granularity, absent versus empty, partial/invalid metadata, multi-key combinations, and change semantics remain open. Runtime unsupported relationships cannot establish authority under Q-038. User instead required an upfront declaration and approved the explicit application-level choice: enabled makes relationship validation mandatory for all grants, including role-based and existing grants. REGISTRATION-003 records the refined approval; no omission-based mode inference is adopted. |
| Q-042 | ~~OPEN~~ ANSWERED | When enabling relationship validation would leave existing grants incompatible, should activation be rejected until those grants have been explicitly corrected? | Existing-grant activation question under REGISTRATION-003. No grandfathering, automatic deletion/narrowing/suspension, or migration mechanism is adopted. Later role/relationship changes and propagation mechanics remain open. User agreed: reject activation, report incompatible grants, preserve the previous configuration, and require explicit authorized correction of grants or declarations before retrying. REGISTRATION-004 records this rule. User requested moving branches; remaining registration lifecycle details are parked. |
| Q-043 | ANSWERED AS REFORMULATED | Remove the unnecessary vocabulary abstraction and explain authorization through permission, scope boundaries, requests, and trusted material, after checking tenability across operations. | TERM-005 agreed after the user requested the check and answered "agree". Original administration framing is withdrawn and preserved verbatim in history/q043-vocabulary.md. This answers the vocabulary question, not ADMIN-004/005's still-open representation or bounds. User also asked to verify sufficient explanatory detail; see explanation-audit.md. |
| Q-044 | ANSWERED | Approve five grant-administration rules: ordinary grant model, explicit administrative operation, complete requested assignment, associated bounds without cross-grant field mixing, and normal validation. | User approved. ADMIN-004/005 are settled at rule level; ADMIN-002/003 stay unchanged. The Finance group example and counterexamples are in grant-model.md. No new scope fields or containment algorithm adopted. |
| Q-045 | ~~OPEN~~ ANSWERED | Consolidate the earlier directions that Auth owns authorization groups/memberships, applications may sync business membership or keep it separate, and groups contain humans rather than first-class services/agents? | GROUP-001-A and GROUP-003 remain proposed pending this closure. Team/group synonymy, group-preferred human grants, and human-dependent automation are already agreed and are not reopened. No nesting, sync timing, or delegation encoding policy adopted. User now approved both policies and reaffirmed that all recording must capture rationale. The earlier proposed-status sentence is historical; groups-and-membership.md records reasoning, alternatives, examples, counterexamples, and open details under PROCESS-003. |
| Q-046 | ~~OPEN~~ ANSWERED | Should an ordinary human/group grant remain valid when its issuing administrator loses grant-administration authority, provided the grant's own validity, conditions, and recipient dependencies still hold? | Proposed lifecycle distinction: issuance authority is checked when making the change, while the resulting human/group grant has its own lifecycle. Human-dependent service/agent delegation remains bounded by the human's current authority. No policy is yet adopted for issuer departure, retroactive invalidity, or incident-response revocation. Update: user agreed; ADMIN-006 now settles later loss of issuance authority. The earlier proposed/open statement is historical. Detailed incident response and invalid issuance remain open; rationale, examples, and explicit-revocation consequences are in grant-model.md. |
| Q-047 | ~~OPEN~~ ANSWERED WITH CLARIFICATION | Should authorization request mean the declared operation with verified identity/tenant context and selected request inputs, while resolved request means its evaluation view with sufficient trusted material to assess the relevant boundaries, rather than an authorization result? | Proposed semantic distinction before JSON: a path department identifies requested context, not proof of an existing certificate's department. Resolution establishes needed meaning and evidence, not permission to execute. No new schema, mandatory eager lookup, prepared handoff, or separate persisted entity is proposed. Update: user agreed, then refined the relationship between declared material and selected scope boundaries. Q-047-A explicitly approved the wording; RESOLUTION-005 and clarified CONTRACT-007 record the agreement. The earlier proposed-status wording is history. |
| Q-047-A | ANSWERED | The endpoint predeclares permissions, inputs, sources, and how to establish any required relationship. Declared inputs remain available; relevant scope and other mandatory checks determine the trusted material needed for evaluation. | User explicitly approved this wording after the {} / department / certificate comparison. Rationale: do not make every input a selected boundary or discard operation identity when a scope does not use an input. Empty scope remains tenant-wide, with no narrower scope restriction; all other mandatory constraints remain. Detailed examples and counterexamples are in endpoint-authorization.md. |
| Q-048 | ~~OPEN~~ ANSWERED | Should a resolved grant be a request-evaluation view of an existing grant, preserving its source identity, associated permissions/scope/conditions, and applicable membership/delegation dependencies, rather than a new independent assignment or a bare permission list? | Next consolidation of RESOLUTION-003/004 and the subset invariant, not a new wire schema. Concrete resolved-grant fields and the distinction between authority expansion, request-specific evaluation, and the final decision remain to be discussed. Update: user approved the evaluation-ready dependent view and added obtaining Vinay's memberships, then grants for Vinay and his membership teams. RESOLUTION-006 records the meaning and logical flow. The earlier open-status wording is history; concrete fields and freshness remain open. The grant chapter captures rationale and examples. |

## Resume here

Current direction: **PROCESS-007 is agreed** — horizontal coverage based on
impact. **Q-067 / DECISION-015 is agreed**, not parked awaiting an answer.
Further low-level result-validation details are parked for the contract-completion
pass, not excluded from v1. **Q-068 / ENFORCEMENT-004 is agreed**: move authority
covers both current and proposed boundaries. Its explanation and rationale are in
operation-enforcement.md; detailed composition remains open. Move horizontally
to **Q-069 / FRESHNESS-001, now agreed**: no stale-cache use of a revoked grant by
authorization checks started after Auth confirms revocation. The freshness chapter
records the coordination/availability trade-off without selecting a cache protocol.
**Q-070 / DELEGATION-003 is agreed as corrected by the user**: affected delegation
is inactive when human support is absent and works again when support returns,
provided the delegation itself remains valid. Explicit renewal is not required
for this case. **Q-071 / ENFORCEMENT-005 is agreed as corrected:** deny rather
than automatically narrowing a partially authorized collection request. The user
rejected the endpoint complexity of grant-derived subset filtering.
**Q-072 / ENFORCEMENT-006 is agreed:** an explicitly Finance-bounded request can
be authorized with Finance-only authority; an all-departments request cannot.
**Q-073 / ENFORCEMENT-007 is agreed:** complete-batch authorization and boundary
checks before effects, without automatic partial execution.
**Q-074 / ENFORCEMENT-008 is agreed:** preserve evaluated application boundaries
through use despite concurrent record changes.
**Q-075 / ENFORCEMENT-009 is agreed:** queued work needs execution-time authorization,
not reuse of submission's allow. Next is **Q-076 / AUDIT-001, proposed**: mandatory
audit of successful grant, membership, role, and delegation administration changes,
distinct from selective access-request logging. Keep in-flight authority changes
and detailed adapter schemas open rather than drilling further here.
The audit remains 35/69 closed; changing discussion order earns no completion credit.

Historical result-contract checkpoint before the horizontal-coverage instruction:

Current: **Q-059 / PERMISSION-005 is agreed**: no permission aliases in v1.
Together with Q-057 and Q-058, this closes HC-04-04 under MEASURE-001:
35 of 69 checkpoints closed, 34 open (50.7% closure). Naming/catalog evolution
remains open; feature exclusions do not remove this decision checkpoint from
the denominator. **Q-060 / DECISION-008 is agreed**: return supporting-grant
references with allow; audit availability does not require tracking every request.
**Q-061 / DECISION-009 is not adopted**: no returned boundary fields are required;
existing endpoint enforcement remains mandatory. **Q-062 / DECISION-010 is
agreed**: minimal allow JSON with version, decision, and grant_ids. **Q-063 /
DECISION-011 is agreed**: minimal deny JSON with version, decision, and the
agreed error code/two messages. **Q-064 / DECISION-012 is agreed**: evaluation-error
JSON using the same error fields without decision. **Q-065 / DECISION-013 is
agreed**: reject mixtures of known result-variant fields. **Q-066 / DECISION-014
is agreed**: require at least one supporting grant ID in an allow result.
**Q-067 / DECISION-015 is proposed**: reject unknown result fields. Ask one
question at a time, with rationale in the chapter. The broader result-contract
checkpoint remains open; the measured count stays 35/69.

Historical checkpoint before Q-059 approval:

Latest: Q-056 / PERMISSION-002 is **agreed with variable depth**. The earlier
Resource-type depth section was checked directly and its full sequence restored
in permission-model.md and the reader. Preserve `:` within the namespace and
`::` before the verb. **Q-057 / PERMISSION-003 is agreed**: no automatic
parent-to-child permission inheritance from name prefixes. **Q-058 / PERMISSION-004
is agreed**: wildcard permission names are outside v1. **Q-059 / PERMISSION-005
is proposed**: exclude permission aliases from v1. Ask this one question, then
return to decision-result completion; detailed catalog validation remains open.

Historical checkpoint before Q-056 approval:

Current: Q-055 / DECISION-007 is **agreed**. The user also requested retention of
the earlier permission explanation; permission-model.md and the active reader
now restore that detail and link the unchanged original. **Q-056 / PERMISSION-002
is proposed**: formally retain namespaced-noun::verb naming. This sidebar does
not reopen operation-versus-reach semantics or settle wildcard/inheritance rules.
After the naming question, return to the unfinished decision-result contract.

Historical checkpoint before Q-055 approval and the permission-retention sidebar:

Latest: Q-053 / DECISION-005 is **agreed as refined**: the evaluator provides
`error_message` and `error_message_reason`; both reach the UI, which controls
presentation. Q-053-A's server-only second message proposal is not adopted.
Both values are client-visible; complete schemas, reason-value encoding, and
safe-content rules remain open. The chapter preserves earlier proposals and
rationale. **Q-054 / DECISION-006 is agreed**: use the same two message fields for
evaluation errors without collapsing the distinction from completed denials.
**Q-055 / DECISION-007 is proposed**: add error_code to carry the stable
machine-readable cause while keeping the two messages readable. No new field is
approved until that question is answered. Continue one question at a time.

Historical checkpoint, superseded by the refinement above:

Current direction: Q-053 / DECISION-005 was revised by the user: the evaluator
provides both client and internal messages; the UI decides presentation. The
original application-authored mapping proposal is preserved as superseded history
in decision-results.md. **Q-053-A remains proposed**: retain the internal message
server-side and send only the client-safe message to the requesting UI. The
chapter records rationale and the response-inspection counterexample. Exact
record layout and HTTP mapping remain open.

Latest update: Q-052 / DECISION-004 is **agreed**. Every completed deny includes
an internal machine-readable reason. The user emphasized its importance and the
standing requirement to record rationale. The chapter explains why, with examples
and counterexamples. Earlier Q-052 proposed/next labels below are history.
Continue one question at a time within decision results; exact reason contracts
and public disclosure remain open. HC-08-02 is not yet closed.

Update: Q-051 is **agreed**, including the concrete timeout example. The original
proposal below is historical. Q-052 / DECISION-004 is now the sole next question:
an internal machine-readable reason for deny, **proposed, not approved**.
Keep discussion one question at a time as the user requested. Full decision-result
contracts remain unfinished; HC-08-02 and the 34/69 closure score stay unchanged.

Historical proposal: Q-051 / DECISION-003 was **PROPOSED, not approved**. Recommend
completed allow/deny decisions with evaluation errors represented separately.
The existing distinction and fail-closed obligation remain settled; exact result
fields and error transport remain open. [Decision results](decision-results.md)
records rationale, alternatives, examples, consequences, and remaining gaps.
Publishing this proposal does not close an audit checkpoint.

Current: Q-050-F approves INPUT-003, application-owned value validation with no
type/nullability fields added to endpoint policy. The policy shape was already
approved; do not ask to approve it again. Examples and rationale are in the policy
chapter. Return next to decision-result contracts. Remaining structural policy
validation/publication details and update/move semantics stay tracked as open;
they do not reopen settled ownership or the no-relationship-block choice.

Current: Q-050-E approves INPUT-002. Every declared input is required at its
specified source, independent of grant breadth; no omission/default/fallback.
The policy chapter records rationale, the optional-input alternative, PUT cases,
and the distinction from value validity. Next: input value-validation responsibilities
and remaining policy validation/publication. Earlier missing-input-open notes are
history for the now-settled presence/source requirement, not all validation details.

Current: Q-050-D approves ENFORCEMENT-003's actual-enforcement review requirement.
The policy chapter records rationale, positive and negative examples, and limits.
Next within Q-050: remaining input/policy validation before full publication.
Update/move and decision-result contracts remain open; relationship-block design
and the review criterion are settled, not further v1 options. Older positions
below are retained history.

Current: Q-050-C approves CONTRACT-012: no relationship block; actual boundary
enforcement belongs to endpoint implementation. The policy chapter preserves
rationale, grant/query example, and the prior proposal not adopted. Q-050-D
captures the user's review suggestion and a proposed constraint-enforcement
refinement; it remains open. Q-050's remaining validation/publication work and
update/move contracts are not complete. Earlier relationship-open notes are history.

Q-050-B follow-up: user approved the PUT body-input example as presented.
The existing rationale remains: explicit extraction does not prove an existing
relationship. Q-050-C remains open; no update/move rule is adopted by this approval.

Current: Q-050-B approves CONTRACT-011's partial endpoint policy structure.
The dedicated format chapter includes GET and PUT body examples with rationale.
Next: Q-050-C relationship bindings. Q-050 as a whole remains open; versioned
partial illustrations are not complete published policies. Earlier notes below
retain their checkpoint status, not current undisputed gaps.

Current: Q-050-A approves CONTRACT-010. Version metadata is settled; the full
endpoint policy contract (Q-050) is still open. Next work is its method/route
binding, one permission, inputs/sources, and relationship bindings. Earlier
version-field/unsupported-version open notes below are checkpoint history.

<!-- Continuation of the question log; the resume position follows this table. -->

| ID | Status | Question / conclusion | Rationale / evidence |
|---|---|---|---|
| Q-050-F | ANSWERED | Keep input value validation in the application request contract, without additional type/nullability fields in endpoint policy. | User considered the shape already discussed/approved and explicitly approved the clarification. INPUT-003 records validation ownership, duplicate-schema rationale, department_id cases, nullable-versus-missing distinction, and consistent validated meaning between authorization and execution. The Q-050-B policy structure remains unchanged; do not reopen it as a new proposal. |
| Q-050-E | ANSWERED | Make every declared input required at its specified source, without silent omission, defaults, or source fallback, independently of applicable grant scope. | User approved. INPUT-002 records the fixed-contract rationale, optional-input alternative not selected for v1, PUT examples, and {} qualification. Present does not mean valid; types, nullability, and other validation/error details remain open. |
| Q-050-B | ANSWERED | Adopt the partial endpoint policy structure: version, method, path, one permission, and named inputs with explicit source/name bindings. | User approved and requested more examples for PUT using body parameters. CONTRACT-011 and endpoint-policy-format.md capture the GET example, PUT body-field mapping, each field's rationale, and current-versus-proposed-state safeguards. The complete policy and relationship bindings are not thereby approved. |
| Q-050-C | ~~OPEN~~ ANSWERED AS REVISED | Define endpoint relationship bindings: how required application facts are established and explicitly connected to registered boundary meanings without treating request claims as proof. | Next part of Q-050 after the version convention and input structure. Need explicit references/inputs and trusted results without inventing scope meanings, a query language, eager resolution of all facts, or another authorization gate. No binding schema or resolver interface is adopted yet. Update: after a grant example the user rejected the relationship block for simplicity and approved mandatory endpoint enforcement of actual boundaries instead. CONTRACT-012 records this revision; the original resolver proposal was not adopted and is preserved in endpoint-policy-format.md. Earlier binding-requirement/open wording is history. |
| Q-050-D | ~~OPEN~~ ANSWERED | Refine endpoint review to verify that every authorization constraint relied on actually restricts output or mutation, rather than merely checking that all material appears in code. | User suggested reviewing use of all material to generate output, with the expectation this catches any breach. Proposed refinement: logging or an ineffective OR filter uses material without enforcing it; {} does not create a department restriction. This is a useful review criterion, not a guarantee against wrong permissions, untrusted context, stale dependencies, or bypass paths. The refinement is not yet approved and does not replace mandatory checks/tests. Update: user approved the refined wording; ENFORCEMENT-003 records the review requirement, examples, rationale, and limits. The earlier proposed-status sentence is preserved history. |
| Q-050-A | ANSWERED | Adopt required top-level version: "1" as a string in published JSON/YAML contracts, reject missing/malformed/unsupported versions, and distinguish contract version from document revision. | User approved the shared convention. CONTRACT-010 records semantics, the schema_version alternative, per-contract-type interpretation, rationale, and rejection examples. Q-050's full endpoint policy contract remains open; its earlier version-representation gaps are settled by this subquestion, not the whole schema. |
| Q-049 | ANSWERED AS REVISED | Exactly one required permission per protected endpoint, mandated by Auth. The permission covers the complete protected operation; inputs, sources, and relationship establishment remain predeclared. | The assistant first proposed AND across multiple required permissions with different complete grants potentially satisfying them. User requested exactly one for design simplicity and approved the revision. CONTRACT-008 records the adopted rule. The original rationale and Finance read/download example are preserved as not adopted in endpoint-authorization.md. Multi-permission grants remain supported; this is not approval of an endpoint policy schema. |
| Q-050 | OPEN | Define the endpoint policy JSON/YAML contract, including version representation, one required permission, inputs, sources, and relationship bindings. | User correctly noted that the endpoint policy contract has not yet been discussed and required versions in all published contracts (CONTRACT-009). Requirements do not settle schema. Exact version field/value/placement, policy structure, binding syntax, validation timing, and compatibility remain open. Discuss this before returning to decision results/enforcement; no draft payload is adopted here. |

- Current: Q-049 approves CONTRACT-008, exactly one required permission per
  protected endpoint; the multi-permission alternative was not adopted.
  The user additionally requires versions in all published JSON/YAML contracts
  (CONTRACT-009) and correctly notes endpoint policy syntax is undiscussed.
  Next: Q-050, endpoint policy contract, before returning to decision results
  and enforcement. Older current-position and plural/open-combination notes
  below are preserved history. Working examples are not published contracts.

- Current: Q-048 approves RESOLUTION-006 and the membership/direct/group
  retrieval flow. Resolved grants remain dependent views, not new assignments
  or allow decisions. The grant chapter captures rationale, examples, and open
  mechanics. Next branch: decision semantics and multiple required permissions;
  no new combination policy is adopted. Earlier resume positions are history.

- Current: Q-047 / Q-047-A agree RESOLUTION-005 and clarify CONTRACT-007.
  The endpoint predeclares permissions, inputs, sources, and how to establish
  any required relationship. The endpoint chapter explains the rationale and
  selective-material examples. Next: Q-048, resolved-grant meaning. Exact
  declaration and request/result schemas remain open; older resume notes are
  retained checkpoint history.

- Current: Q-046 approves ADMIN-006. Ordinary human/group grants do not depend
  merely on their issuing administrator retaining issuance authority. The grant
  chapter records rationale, alternatives, examples, and revocation consequences.
  Next: Q-047 returns to requests/resolution, distinguishing request inputs from
  trusted evaluation material. Administrative encoding and detailed lifecycle
  work remain open. Earlier current-position labels below are checkpoint history.

- Current: Q-045 approves Auth-owned authorization membership with optional
  application sync and human-only groups. The rationale and examples are in
  groups-and-membership.md. Q-046 is next: distinguish the lifecycle of an
  ordinary human/group grant from its administrator's later role changes.
  This does not reopen independent service/agent authority or parked sync detail.

- Current: Q-044 approved ADMIN-004/005's five governing rules and Finance
  example. Exact scope encoding and containment are not thereby solved.
  Q-045 next consolidates earlier group ownership/human-only membership
  proposals. Earlier Q-043/current-position notes below are history.

- Current: Q-043 settled TERM-005 after a tenability check. The vocabulary
  correction is agreed, while ADMIN-004/005 remain open. Stay in grant
  administration; do not reopen parked registration mechanics. Earlier
  Q-043-open notes below are preserved history. The detailed explanation and
  [coverage audit](explanation-audit.md) distinguish settled reasoning from gaps.

- **Current: stage 5, grant administration — ADMIN-004/005, Q-043.** Q-042
  is agreed as REGISTRATION-004. User directed ending the registration detour;
  park remaining lifecycle mechanics and return to the saved administration
  branch. Do not continue Q-042 implementation detail. Older positions are history.

- Current: Q-041 approved REGISTRATION-003. The application explicitly chooses
  enabled/disabled upfront; enabled requires relationship validation for all
  grants, including role-based and existing grants. Q-042 asks about enabling
  validation when incompatible grants already exist. Older omission-based
  Q-041 proposals and active-position notes below are historical.

- Current: Q-040 approved REGISTRATION-002 with optional relationship
  declarations bound through registration. Q-041 proposes omission behavior;
  it is not yet agreed. Earlier Q-040-next notes below are historical.

- Latest checkpoint: Q-039 approved REGISTRATION-001. Applications register
  permission/scope contracts; Auth validates grant acceptance without domain
  interpretation. [Application registration](application-registration.md)
  preserves rationale and examples. Q-040 is next: compatibility declarations.
  Earlier Q-039-open notes are historical, not current status.

- Current position: [discussion tree](discussion-tree.md), stage 6 →
  Reconciliation checkpoint adds preservation notices and
  [16 cross-domain cases](use-case-examples.md); no new policy is adopted and
  the current discussion/return path remains unchanged.
  SCOPE-006/008 boundary semantics and SCOPE-007's canonical v1 format are agreed.
  Q-034 is answered; next address scope-key definitions and governance, with no
  new scope question issued yet. Sidebar CONTRACT-007 / Q-035 is agreed;
  system-overview.md includes the requested SVG logical block diagram.
  Return to scope-key definitions now. Q-033 approves CONTRACT-006's
  endpoint-owned gate; INPUT-001 captures
  selected endpoint request inputs. Recording has resumed under PROCESS-006.
  SCOPE-004 / Q-028 remains open; explanation did not establish agreement.
  SCOPE-002's typed representation is now historical; definition governance
  remains open under SCOPE-007. Current grant formats are in grant-format.md;
  earlier layouts are explicitly deprecated and retained. discussion-tree.md
  supplies the whole-handbook mind map plus the detailed branch inventory.
  Q-025 approved the detour. Return to ADMIN-004/005 after settling shared scope.
  GROUP-004 / Q-021 is closed. handbook.md indexes the detailed working chapters,
  while the original lab page remains unreconciled.
  After grant authority/lifecycle
  questions, define the request and its
  resolved forms (stage 7). Prepared and two endpoint modes are deprecated;
  concrete evaluator inputs/outputs still need precision.
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
- Historical, now DEPRECATED: CONTRACT-002 — two endpoint-declared modes. Middleware-complete
  endpoints use allow/deny; endpoint-completion endpoints use deny/prepared in
  middleware and must complete authorization in the handler. Mode is contractual,
  not dynamically inferred per request. RESOLUTION-002 and ARCH-003 are superseded.
- Historical, now DEPRECATED: CONTRACT-003 validated the mode against scope/delegation
  requirements; missing required facts cannot silently produce allow or switch
  a middleware-complete endpoint to prepared. Enforcement remains mandatory.
- Historical CONTRACT-004 (PARTIALLY DEPRECATED) — URI, handler, and a single authoritative
  authorization declaration are designed together per HTTP method/route. Keep
  ~~one mode per endpoint.~~ Distinguish different contracts from multiple checks
  or grant scopes handled under one contract; avoid creating a route per grant.
  Current: CONTRACT-006/007 retains one authoritative declaration without a
  two-mode choice or prepared handoff.
- Explain FACT-001 using the certificate route: requested tenant and department
  IDs do not prove certificate containment or ownership. Required ownership
  checks depend on the applicable scope, not simply on the existence of a user.
- Historical, now DEPRECATED as mode-selection guidance: CONTRACT-005 — missing authorization facts versus applying
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
- Q-022 settled separate administration and explicit audited self-assignment.
  Next discussion topic: the exact boundaries on what access an administrator
  may assign, and to whom. Q-024's scope-field challenge is recorded under
  PROCESS-005. Q-025 approved settling the shared scope model before returning
  to ADMIN-004/005; administrative bounds and representation remain open.
  After grant authority/lifecycle questions, compare
  declared and resolved forms under CONTRACT-006; prepared remains deprecated.
  Keep other branches in the tree open.
- Coverage percentages require a later endpoint audit and are not assumed proven.
- ~~The generic middleware has no application-database dependency. The
  shared evaluator can be invoked again with domain facts; handlers should not
  each invent scope-combination semantics. Preserve unresolved constraints.~~
  Current: application facts are gathered under the one endpoint-owned gate;
  shared evaluation is not a second completion stage after middleware authority.
  No handler invents its own scope semantics (CONTRACT-006/007).
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
