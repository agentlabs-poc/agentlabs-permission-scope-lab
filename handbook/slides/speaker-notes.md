# Core concepts & JSON — presenter notes

Revision/adoption mechanics are intentionally outside the presentation.

## 1. Handbook of Authorization

PPT-001: approved vocabulary-first presentation. The user asked to keep revision mechanics aside. This deck teaches definitions and approved representations without changing any canonical contracts. Presenter notes retain qualifications that do not need to dominate the slides.

Source: [handbook reference](../theory/canonical-terms.md).

## 2. Six questions make the model readable

Authentication establishes identity. Authorization connects authority to a protected effect. A principle is a design constraint; a policy expresses applicable rules. Effective authority is what current valid routes can provide. In all JSON, version identifies the format. Revision fields are retained but visually muted and not taught in this deck. No revision/adoption lesson is included.

Source: [handbook reference](../theory/canonical-terms.md#principles-and-rules).

## 3. Human, actor and principal

Identity describes who acts. human_id identifies the authorizing human, anchoring current authority and self. Direct-human equality is required. Matching JSON is not identity proof: the identity must be established through trusted evidence. The typed actor avoids conflating a proxy caller with the human.

Source: [handbook reference](../theory/canonical-terms.md#identity).

## 4. Agent and service account are proxies

The human-dependent service-account rule is specific to this framework. Other systems may choose independent machine principals. Neither JSON block is a delegation record: both identify the participants. Exact delegation evidence remains pending. There is no separate duplicate agent_id claim.

Source: [handbook reference](../theory/canonical-terms.md#identity).

## 5. JWT subject stays human

This identity-related payload excerpt omits issuer, audience, expiry and cryptographic verification for focus, not because they are optional. sub must equal identity.human_id. The repeated human ID keeps identity self-contained. Legacy consumers that ignore the proxy must not bypass applicable delegation checks. Universal backward compatibility has not been demonstrated.

Source: [handbook reference](../theory/canonical-terms.md#identity).

## 6. Tenant is the enclosing boundary

A tenant identifier in a URL is supplied input, not proof of access. Tenant association is not authorization-group membership. Application platform administration governs capability publication outside tenant business scope, within its own management boundary. Publishing HRMS permissions does not grant tenant payroll access. Complete context and platform trust mappings remain pending.

Source: [handbook reference](../theory/canonical-terms.md#enclosing-context).

## 7. Permission names the operation

Permission grammar: <app>:<domain>[:<subdomain-or-resource>...]::<verb>. Permissions must be registered. Department IDs do not belong in operation names. The array is a nested value, not a standalone published contract, so it does not invent a version wrapper. An existing permission name cannot acquire materially different authorization meaning after retirement. HTTP method alone does not define permission.

Source: [handbook reference](../theory/canonical-terms.md#permission).

## 8. Scope selects the boundary

Our scope is a required flat object. Arrays, nested query objects, duplicate keys, wildcard operators, unsupported keys and empty/non-string values are not accepted. Missing/null scope is not an empty object. A child’s {} adds no local restriction but retains parent constraints. Alternative boundaries need separate complete routes. Auth validates registration and syntax; the application owns boundary meaning.

Source: [handbook reference](../theory/canonical-terms.md#scope-and-self).

## 9. $self means the authorizing human

This example assumes an already valid supported self-scoped group route. It does not settle whether Maya’s self-limited source permits distributing Nutan’s self access. Copying the token is not a containment proof. The application must establish or enforce its registered user/self relationship on the actual data. Prefer group-based access without forbidding legitimate direct assignments.

Source: [handbook reference](../theory/canonical-terms.md#scope-and-self).

## 10. Three concepts—not three names for access

A role does not define recipient or scope. A grant uses explicit permissions or a complete role reference, not both. An assignment connects a reusable recipient-free grant to a human or group. The full standalone role-publication schema remains pending; this conceptual comparison deliberately does not invent one. Scope and parent support still constrain role-based authority.

Source: [handbook reference](../theory/canonical-terms.md#grant-records).

## 11. A grant defines bounded authority

G1 matches the handbook running example. Assume registered meanings, valid G0 support, enabled live control and successful administrative/boundary checks. The example is immutable content, not a complete combined record: live control and recipient assignment are separate. Revision mechanics are omitted from teaching, but required fields remain intact. Ordinary grants cannot omit their parent to manufacture roots.

Source: [handbook reference](../theory/canonical-terms.md#grant-records).

## 12. A grant can use a role instead

Assume R-CERTIFICATE-READER contains certificate-read and fits valid G0 support. Full standalone role publication JSON is not approved by showing this reference. role_revision is preserved in the canonical JSON but its mechanics are outside scope. A direct-permission grant contains neither role_id nor role_revision. No latest-role fallback or mixed variant is allowed.

Source: [handbook reference](../theory/canonical-terms.md#grant-records).

## 13. An assignment names the recipient

The canonical assignment also retains its required content-selection field without teaching adoption mechanics. Prefer human access through groups while retaining valid direct-human routes. Assignment enablement is not grant-wide enablement or proof of effective authority. An agent does not get independent first-class group membership through this record.

Source: [handbook reference](../theory/canonical-terms.md#grant-records).

## 14. Membership is not ownership

Auth owns explicit human membership. Team hierarchy does not imply inherited human memberships. Teams are not application departments even when used to organize department staff. Team write includes membership management; assigning grants needs separate authority. Full membership, team hierarchy and owner record formats remain pending, so no tentative scratch JSON is promoted. Authorized membership management intentionally distributes the team’s existing access.

Source: [handbook reference](../theory/canonical-terms.md#teams-and-administration).

## 15. Can assign? And may assign this authority?

The Auth endpoint evaluator checks the administrative operation and recipient boundary. Auth’s authority-boundary validator checks the proposed authority against eligible source support and team ceilings. Q-093 requires a valid direct/group source available to the assigner as well as administrative authority. This is not an application database lookup or a generic business-rule engine. Creator history is not automatically permanent support. Exact owner-transfer permission remains pending.

Source: [handbook reference](../theory/canonical-terms.md#responsibility-layers).

## 16. A child grant can only narrow

Assume Team2 is a child of Team1, A1 supplies G1 to Team1, and A2 supplies G2 to Team2 with valid controls/support. All child-team authority must remain within the parent-team ceiling as well. Adding dept=ENG cannot overwrite dept=FIN. Subgrant means a grant relative to a parent, not another record type. Full team hierarchy wire representation remains pending.

Source: [handbook reference](../theory/canonical-terms.md#dependent-relationships).

## 17. Two lineages, joined by assignments

The graph repeats G1/A1/Team1 and G2/A2/Team2 from the handbook. Required G0 support is assumed. G2 parent_grant_id is G1; A1 establishes actual support in the parent-team context. No parent_assignment_id or parent_grant_revision field is introduced. Nutan’s explicit Team2 membership supplies the valid G2 route, not Team1 write access. Unrelated TeamX holdings cannot replace required Team1 support. Complete team relationship JSON is pending.

Source: [handbook reference](../theory/canonical-terms.md#dependent-relationships).

## 18. Proxy authority stays within the human

Delegation is not simply an account creator field. Both current human authority and delegation limits apply. If support later returns, still-valid delegation can resume; expired, revoked or explicitly disabled delegation is not automatically revived. Full delegation evidence JSON is pending. Teams contain humans, not these automated actors.

Source: [handbook reference](../theory/canonical-terms.md#delegation).

## 19. Enabled does not mean effective

This slide is a later lifecycle snapshot, not a contradiction of the earlier valid-route example. Stored enabled flags are necessary controls, not final authorization. A still-enabled child can become ineffective when required support is disabled and resume when that support returns. A child explicitly disabled stays disabled until explicitly enabled and revalidated. Structural changes require bottom-up handling of affected bindings, not merely making an ancestor ineffective.

Source: [handbook reference](../theory/canonical-terms.md#lifecycle).

## 20. Validity limits when authority applies

Optional validity belongs in immutable grant content, not live grant control or assignment. The full timestamp and lifecycle API contracts remain unfinished. Absence of a local validity window means no extra local time restriction, not independence from upstream expiry. This deck does not teach content revision or adoption mechanics, but preserves approved field placement.

Source: [handbook reference](../theory/canonical-terms.md#lifecycle).

## 21. An orphan has lost required support

Orphaning is assessed for the affected lineage of a reusable grant, not automatically every assignment of that definition. Permanent deletion/revocation differs from temporary disablement; no invented revoked/orphan live status is selected. Required structural guards still apply before breaking bindings. Missing support cannot be replaced with any convenient broader source. Warning/prevention at an upper management layer does not alter the canonical ineffective outcome.

Source: [handbook reference](../theory/canonical-terms.md#lifecycle).

## 22. Register meanings, then establish authority

Applications own domain meaning; Auth validates registered contracts and canonical authority rules. Optional permission/scope compatibility is declared upfront; enabled mode validates every grant. Bootstrap must establish coherent maximum intended authority inside its authorized boundary, with a minimal setup. Computed root permissions follow one shared application catalog; ordinary child permissions do not silently grow. Complete registration/root/trusted-setup JSON remains pending; no stored wildcard or is_root field is invented.

Source: [handbook reference](../theory/canonical-terms.md#registration).

## 23. Declare the required operation and inputs

All declared inputs must be present at their stated sources. The tenant path input must be reconciled with trusted tenant context. Local input names are not automatically scope keys. There is no prepared handoff or canonical relationships block. The endpoint must establish or enforce the relationships needed to constrain actual execution. Full policy validation, nested-body syntax and additional sources remain pending.

Source: [handbook reference](../theory/canonical-terms.md#endpoint-declaration).

## 24. Body field and local input can differ

The body field must be present at the declared source. Application value/type validation is not a new grant-scope field. Current and proposed operation boundaries must be respected; the exact move-grant composition is still pending. A broader {} grant does not waive declared input presence. The one-permission requirement does not mean one grant contains only one permission.

Source: [handbook reference](../theory/canonical-terms.md#endpoint-declaration).

## 25. Request → material → resolved request

A request input is a declared value from a specified source. A domain fact is application-owned knowledge; a relationship is the relevant association, such as C17 belonging to Finance. A supplied path value alone is not a fact. The endpoint may fetch authoritative facts or preserve the required constraints through execution. No canonical target entity, relationship record or resolved_inputs envelope is introduced.

Source: [handbook reference](../theory/canonical-terms.md#request-and-material).

## 26. A resolved grant is a dependent view

Resolved grants are computed views of existing applicable routes, preserving source bindings, scope and restrictions. They are not stored new grants, new assignments or allows. This slide is an explanatory table, not an approved serialization. A FIN-write route and ENG-read route do not provide ENG-write. Independent alternatives remain separate: resolution also does not compute a globally narrowest intersection of every unrelated grant. Full resolved-grant/evidence format remains pending.

Source: [handbook reference](../theory/canonical-terms.md#resolution-and-evaluation).

## 27. Allow identifies supporting authority

grant_ids is a non-empty array of non-empty strings. It identifies supporting grants from this evaluation, not necessarily a complete lineage snapshot. Returned scope is not required. References are available even when not every request is recorded for audit; audit storage/retention design is outside scope. Unknown fields and mixed variants are rejected. Full code/transport/provenance schema work remains pending.

Source: [handbook reference](../theory/canonical-terms.md#decision-and-enforcement).

## 28. Deny is not an evaluation error

Deny means sufficient evidence completed evaluation without authorizing the request. Error means a required evaluation could not complete, such as an Auth timeout with no sufficient valid preloaded authority. Error has no decision field, but omission alone does not validate a malformed response. The code strings and message text here are illustrative, not an exhaustive catalog. Both error_message and error_message_reason are evaluator-provided; disclosure and value rules remain pending. Malformed/unknown/mixed result variants cannot become allow.

Source: [handbook reference](../theory/canonical-terms.md#decision-and-enforcement).

## 29. From Nutan’s request to a constrained read

This is one endpoint-owned gate, not middleware allow followed by an independent prepared decision. Middleware may authenticate and preload sufficient authority. Auth’s canonical layer supplies grants, membership and support; the application layer supplies facts or constrained execution. Inputs appearing only in logs do not enforce boundaries. Already-allowed ordinary synchronous operations and new checks have distinct approved temporal rules; this slide does not invent caching or concurrency protocols.

Source: [handbook reference](../theory/canonical-terms.md#responsibility-layers).

## 30. The vocabulary in one sentence

End with the canonical distinctions rather than claiming every interface is complete. Sources: handbook/theory/canonical-terms.md; implementation/05-canonical-model.md; implementation/06-auth-service.md; implementation/07-application-integration.md; appendices/pending.md. Revision/adoption mechanics were intentionally left out at the user’s request; required fields remain in examples. Full team/membership/owner, standalone role, registration/root, delegation and request/resolved-view representations remain pending. This deck neither changes the model nor closes those design decisions.

Source: [handbook reference](../theory/canonical-terms.md).
