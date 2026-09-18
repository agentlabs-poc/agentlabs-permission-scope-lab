# Pending decisions and contracts

[Handbook contents](../README.md)

These items remain **pending, deferred or excluded** while the manuscript stands.
This is not approval, an implementation roadmap, an exclusion from v1, or a count of
questions the reader must answer. Several rows combine related contract work. The
existing [assessment](../../docs/discussion-assessment.md) and
[criterion register](../../docs/handbook-completion-audit.md) retain the detailed
tracking and history. This appendix explains how those gaps affect the book.

The general theory can be read independently of these framework decisions.
The [Foundations reference](../theory/canonical-terms.md) explains every concept's
representation status, including approved core JSON and formats still pending.
Both Foundations and implementation chapters state supported behavior without
selecting an unapproved policy or field to fill a gap.

## Closed since the first draft

**Three rows closed, and two moved without closing.** They are listed here rather
than deleted, because a reader returning to the first draft needs to know what moved —
and because a row that is half closed is the easiest kind to over-read.

| Was pending | Now |
|---|---|
| The authority-loading transport — the request an application makes and the answer it receives | **Closed.** Stated in [Chapter 7](../implementation/07-application-integration.md) and in the [Foundations reference](../theory/canonical-terms.md#resolution-and-evaluation). Its route, a batch form and freshness fields are not adopted, and are listed below. |
| Administrative bounds, and how administrative containment is validated | **Closed.** Administration is an ordinary grant on the Auth root chain — [Chapter 6](../implementation/06-auth-service.md). No second authority model exists to specify. |
| The handler and embedded-agent integration contract | **Closed.** Four things supplied, one received, and a fixed order — [Chapter 7](../implementation/07-application-integration.md). An SDK and the HTTP status mapping are deliberately not specified. |
| Root establishment **authority** | **Closed** — a tenant operates in two namespaces, and establishment is the platform-namespace authority's closing act. Its **trust and intent evidence is not**, nor is the computed-root encoding or deliberate recovery. See P-03 and Chapter 6. |
| The ownership **record** and ownership **absence** | **Superseded and settled respectively.** Ownership is a grant, so no owner record is left to design, and no team is ownerless. The **permission that authorizes an ownership transfer has still never been chosen** — see P-05. |

## Still open

| Ref | Pending item | What is already usable | What the manuscript does not assume |
|---|---|---|---|
| P-01 | Direct-human parent-support eligibility | Parent ceiling, subset/AND, actual adopted lineage, preferred group access and the original assignment-survival rule. | Newest, arbitrary or all tenant holdings as automatic support; a new parent-assignment/revision selector. |
| P-02 | Recipient-relative source binding, especially `$self` | Runtime self binds to the authorizing human; groups are not self; proxy access retains its human ceiling. | Copying Maya's self selector proves permission to distribute Nutan's self access. |
| P-03 | Computed-root encoding, **trusted setup evidence and deliberate recovery** | Parent omission for a trusted root, the trust marker on the grant head, establishment as the platform-namespace authority's act, coverage computed from the application's catalog, one shared catalog, and recovery being establishment again after a bottom-up dismantling. | A stored wildcard, an `is_root` or catalog-source field, an ordinary parent-omission bypass, or a specified proof that a trusted setup must carry. The authority is settled; the evidence it presents is not. |
| P-04 | Full grant, role, assignment and lifecycle contracts | Approved core record variants; immutable revisions; explicit adoption; latest on new assignment/upgrade; separate controls and revision-local validity; publication free and adoption evaluated; a dependent blocking disable and delete. | Invented defaults, timestamp/ID syntax, rollback mode, revision-specific controls or complete CRUD/error schemas. |
| P-05 | Membership and hierarchy records, and **the permission that authorizes an ownership transfer** | Human groups, no inherited human membership, bounded team hierarchy, synchronization as an ordinary authorized caller, unchanged team-held support, ownership as a grant, and no ownerless team. | A wire schema for the team or membership record, silent authority import, an approved two-owner policy, or that team-write authorizes handing a team's administration to somebody else. Appointing another administrator is creating and assigning a grant, and which permission covers that act is not chosen. |
| P-06 | Identity adapters and complete direct-delegation evidence | Actor/human block, chosen JWT mapping, direct human-to-proxy only, a delegation's own validity window, a delegation tracking the human rather than a snapshot, and restoration of still-valid delegation. | Verified universal legacy compatibility, proxy chains, independent service authority or completed account-transfer/multiple-human rules. |
| P-07 | **The boundary an endpoint operates at** | One permission; explicit path/body inputs at their declared sources; required trusted correlation; mount-time structural validation; one endpoint-owned gate. | That a policy declares the scope boundary it reads at, or that the gate can check the binder's material against a declaration. Until this is settled, an all-values selection denies. |
| P-08 | Remaining result and evidence contracts | Approved minimal result variants, messages, the published code catalogue, and the ordered contributing chain. | An exhaustive reason catalog a consumer may switch on, an HTTP mapping, complete disclosure rules, or batch evidence. |
| P-09 | Freshness, caching and remaining ordering contracts | Confirmed reductions govern new checks; bounded prior-allowed synchronous completion; stronger Auth write consistency. | A cache, an authority epoch, a grace period, a chosen consistency protocol or silent retries. A cache and an epoch are adopted together or not at all. |
| P-10 | Collection shape, and non-HTTP work | Both move boundaries evaluated at one gate; deny uncovered collections; per-item batch preflight; execution-time authorization for queued work. | Complete count/export/pagination contracts — and see P-07, which may dissolve the count question rather than answer it. Queues, schedules, streams and long-running work are **explicitly deferred for v1**, not unspecified by oversight. |
| P-11 | Scope evolution and residual authorization-only restrictions | Registered meanings; values opaque and matched exactly, except a platform key naming Auth's own records; subtree and pattern scope **excluded**; optional declared compatibility checking; retirement narrowing rather than closing. | A generic business-rule engine, subtree or reference semantics arriving later, or silent scope-definition migration. |
| P-12 | An administrative wire contract, and **which permission each administrative operation requires** | That administrative authority is a grant, what bounds each side, and the three group permissions. One authority-loading endpoint is published. | That administrative operations have a transport — they do not, and operation routing stays open. Nor that every administrative operation's permission has been chosen: the group family is named, and the rest are not. |
| P-13 | Final applicability, governance, scenarios and publication acceptance | Existing scope exclusions and approved principles; source-backed examples; a recompiled manuscript. | New mandatory consumers, a rule-change exception/bypass process, exhaustive conformance or final handbook acceptance. |

## Items not being reopened

- External audit-system design is outside this handbook. Decision evidence remains
  in scope, and what a consumer in that layer may rely on is now stated in
  [Chapter 6](../implementation/06-auth-service.md); retention and storage policy
  do not come back.
- The proposed business-rule expansion was disapproved. Do not reintroduce it by
  calling it application-specific authorization — and note that missing, invalid and
  unsupported *evidence* is settled without one.
- Proxy-to-proxy delegation chains are not supported in v1.
- Assignment-specific validity was explicitly deferred.
- The independent `parent_grant_revision` proposal was withdrawn; an additional
  parent-assignment reference was not adopted for the settled model.
- Whether required authority arrives in one grant or several is **not a question this
  model has an opinion about**. What is required is that the resolved authority covers
  the permission and the boundary; composition is endpoint design.
- A nested body selector is **refused, not deferred**. So is a hierarchical scope
  value, and so is an identifier rename.
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
[criterion register](../../docs/handbook-completion-audit.md),
[boundary validation limits](../../docs/authority-boundary-validation.md),
[direct-human trace](../../docs/direct-human-parent-context.md),
[contract inventory](../../docs/grant-contract-closure.md),
[the open endpoint boundary](../../docs/policy-scope-boundary.md),
[evidence cases](../../docs/grant-conditions.md),
[background deferral](../../docs/background-authorization.md),
[audit boundary](../../docs/authority-change-audit.md).
