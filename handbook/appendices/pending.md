# Pending decisions and contracts

[Handbook contents](../README.md)

These items remain **pending while the manuscript is drafted**. This is not
approval, an implementation roadmap, an exclusion from v1, or a count of questions
the reader must answer. Several rows combine related contract work. The existing
[assessment](../../docs/discussion-assessment.md) and
[criterion register](../../docs/handbook-completion-audit.md) retain the detailed
tracking and history. This appendix explains how those gaps affect the book.

The general theory can be read independently of these framework decisions.
The [Foundations reference](../theory/canonical-terms.md) explains every concept's
representation status, including approved core JSON and formats still pending.
Both Foundations and implementation chapters state supported behavior without
selecting an unapproved policy or field to fill a gap.

| Ref | Pending item | What is already usable | What the manuscript does not assume |
|---|---|---|---|
| P-01 | Direct-human parent-support eligibility | Parent ceiling, subset/AND, actual adopted lineage, preferred group access and the original assignment-survival rule. | Newest, arbitrary or all tenant holdings as automatic support; a new parent-assignment/revision selector. |
| P-02 | Recipient-relative source binding, especially `$self` | Runtime self binds to the authorizing human; groups are not self; proxy access retains its human ceiling. | Copying Maya's self selector proves permission to distribute Nutan's self access. |
| P-03 | Root/catalog/trusted setup representation and deliberate recovery | Registration first; maximum intended initial authority; explicit group and membership; coherent visibility; same-intent continuation; shared computed roots. | Stored wildcard, new root flag/source field, ordinary parent-omission bypass, selective tenant catalog versions or silent recovery replacement. |
| P-04 | Full grant, role, assignment and lifecycle contracts | Approved core record variants; immutable revisions; explicit adoption; latest on new assignment/upgrade; separate controls and revision-local validity. | Invented defaults, timestamp/ID syntax, rollback mode, revision-specific controls or complete CRUD/error schemas. |
| P-05 | Membership, hierarchy, ownership and optional-sync contracts | Human groups, no inherited human membership, bounded team hierarchy, explicit ownership changes and unchanged team-held support. | An automatic owner bundle, owner-transfer permission inferred from team write, silent authority import or an approved two-owner policy. |
| P-06 | Identity adapters and complete direct-delegation evidence/lifecycle | Actor/human block, chosen JWT mapping, direct human-to-proxy only, dynamic human ceiling and restoration of still-valid delegation. | Verified universal legacy compatibility, proxy chains, independent service authority or completed account-transfer/multiple-human rules. |
| P-07 | Complete endpoint policy and embedded-agent interfaces | One permission; explicit path/body inputs; required source binding; application value validation; one endpoint-owned gate. | New nested selector syntax, arbitrary source kinds, a relationship block, prepared handoff or a complete SDK/API signature. |
| P-08 | Full request/resolved/evidence/result contracts | Conceptual distinctions; complete routes; approved minimal result variants, messages and contributing grant IDs. | Exhaustive reason catalog, HTTP mapping, all value/disclosure rules, complete lineage/batch evidence or required returned scope. |
| P-09 | Remaining time, conflict and operation-ordering contracts | Confirmed reductions govern new checks; bounded prior-allowed synchronous completion; stronger Auth write consistency. | A cache grace period, chosen consistency protocol, silent retries or automatic extension to not-yet-allowed/long-running operations. |
| P-10 | Remaining operation coverage | Both move boundaries; deny uncovered collections; complete batch preflight with per-item routes; execution-time authorization for queued work. | Settled move grant composition, complete count/export/pagination contracts, or universal stream/recurrence/running-job behavior. |
| P-11 | Scope/reference evolution and residual authorization-only restrictions | Registered meanings, abstract validation, optional declared compatibility checking, permission retirement and stable permission meaning. | A generic business-rule engine, arbitrary subtree/reference semantics, full permission restoration/mixed-grant behavior or silent scope-definition migration. |
| P-12 | Final applicability, governance, scenarios and publication acceptance | Existing scope exclusions and approved principles; source-backed examples; an authored first draft. | New mandatory consumers, a rule-change exception/bypass process, exhaustive conformance or final handbook acceptance. |

## Items not being reopened

- External audit-system design is outside this handbook under Q-076. Decision
  evidence remains in scope; retention/storage policy does not.
- The proposed business-rule expansion was disapproved under Q-084. Do not
  reintroduce it by calling it application-specific authorization.
- Proxy-to-proxy delegation chains are not supported in v1 under Q-127.
- Assignment-specific validity was explicitly deferred under Q-108.
- The independent `parent_grant_revision` proposal was withdrawn under Q-112A;
  an additional parent-assignment reference was not adopted for the settled model.
- The current model has no prepared handoff, canonical target wrapper, endpoint
  relationship block, automatic permission-prefix inheritance or permission alias.

Keeping these exclusions distinct from pending work prevents the draft from
quietly adding scope or treating an unanswered question as already excluded.

## Returning to a pending item

Resume from its recorded source, explain the affected invariant, and compare
concrete records or operations. If a new choice is required, present it with a
reference, recommendation, rationale and core-principle check. If the answer is
already implied by an approved rule, complete the writing directly instead of
requesting another confirmation. Keep a decision and its representation separate
unless both were actually reviewed together.

**Primary sources:** [current status](../../docs/current-status.md),
[C01 limits](../../docs/authority-boundary-validation.md),
[direct-human trace](../../docs/direct-human-parent-context.md),
[contract inventory](../../docs/grant-contract-closure.md),
[scope rejection](../../docs/grant-conditions.md),
[audit boundary](../../docs/authority-change-audit.md).
