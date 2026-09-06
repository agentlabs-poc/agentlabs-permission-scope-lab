# Remaining discussions and autonomous work — ASSESS-001

Assessment date: 2026-09-06. Scope: the existing handbook closure criteria,
current approved decisions through Q-120A, and their source chapters. The user
asked how many discussions remain and what the assistant can run, resolve, and
record. This is a documented planning assessment, not new policy, a scope cut,
or permission to self-approve unresolved canonical contracts.

## Answer and counting convention

| Category | Count | Responsibility |
|---|---:|---|
| Substantive design-discussion topics | **13** | Assistant frames a concrete recommendation from evidence; user decides the unresolved behavior or governance. |
| Contract-review packages | **5** | Assistant drafts the full package and separates already-approved fields from genuinely new choices; user approves the latter. |
| Final v1 acceptance | **1** | User accepts the reconciled package and any explicit exclusions/deferrals. |
| Autonomous writing/verification work packages | **8** | Assistant can perform the work without another approval of settled rules; some final outputs depend on the decisions above. |

There are **18 identified discussion/review topics, plus final acceptance**.
This is a finite agenda at topic/package granularity, not 18 guaranteed single
questions or messages. Each design topic lists its remaining choices below;
multi-part topics may need more than one short question, asked one at a time.
The five contract reviews are not five undiscovered architectural principles.

The eight assistant work packages are the **execution side of the same work**,
not eight additional discussions to add to 18. Five produce the five review
packages. They can be drafted now with unsupported choices labeled; they cannot
be declared fully resolved merely because a draft exists.

This replaces the earlier incomplete conversational picture of one active topic
and one parked question, which described only the current branch. For example,
Q-096's residual ownership policy and several earlier parked contract areas also
remain. It does not retrospectively change their approval status.

### What is counted once

- One topic groups closely related choices at a common authority or integration
  boundary. It is not one checklist row or one occurrence of the word “open.”
- If a decision affects several schemas/chapters, count the decision once; list
  those consumers in the coverage map rather than opening duplicate questions.
- Already-approved behavior is writing/evidence work, even when old prose still
  says proposed. No re-vote on parent ceilings, role/direct variants, root parent
  omission, automatic root growth, or required input presence is counted.
- New fields, defaults, compatibility behavior, trust assumptions, and scope
  exclusions require review. “I can draft this” does not mean “I can approve this.”
- No extra chapter, feature, implementation platform, or external audit design
  is added to the existing scope. Runtime implementation is not a handbook exit condition.

## Thirteen substantive discussion topics

D numbers are local assessment references, not new canonical Q decisions. Their
listed choices define the assessed scope; this is not an unbounded invitation
to expand each topic. Recommendations are to be presented with evidence when
that topic is taken up, not silently adopted by this list.

| Ref | Remaining discussion and concrete unresolved choices | Already settled; do not reopen | Evidence / existing criteria |
|---|---|---|---|
| D01 | **Trusted capability and root ownership:** how initial provisioning and later application capability updates are authorized and bound to the correct tenant/application; which root receives which registered capabilities. | Registration-first setup, maximum intended initial authority, explicit user/group membership, no personal bootstrap bypass, automatic root growth requirement. | [Q-113–Q-119](bootstrap-authority.md), [Q-120A](root-permission-evolution.md), [registration](application-registration.md); HC-05-09, HC-06-06. |
| D02 | **Automatic-root mechanism:** live/computed coverage versus automated materialized revisions/adoption; reconcile the selected mechanism with immutable content and existing assignments. | Root must grow without a separate manual expansion gate; no ordinary child automatic permission selection. No `*` is approved yet. | [Q-120A](root-permission-evolution.md); HC-05-09, HC-05-13, HC-07-08. |
| D03 | **Incomplete setup and recovery:** Q-117 authority visibility/completion, partial/concurrent setup attempts, and governed recovery when initial setup cannot finish. | Completed bootstrap replay changes no authority under Q-116. Omission alone cannot prove root status. | [Q-117, parked](bootstrap-initial-assignment.md), [root trust](bootstrap-authority.md); HC-05-09. |
| D04 | **Registered definition evolution:** permission/scope definition removal or changed meaning, effect on existing grants and automatic roots, reference/existence responsibilities, and exact/subtree semantics where applications need them. | Application owns domain meaning; Auth validates registered contracts; upfront compatibility mode and activation safeguard are agreed. No universal department type or restored relationship block. | [registration open contracts](application-registration.md), [scope](scope-model.md), [permission governance](permission-model.md); HC-04-03, HC-06-06, HC-06-07. |
| D05 | **Delegation lifecycle limits:** chains/redelegation, applicable lifetime representation, authority growth within a delegation, and corresponding administration. | Every proxy is human-dependent; affected access is a subset; still-valid delegation can regain access when human support returns. | [delegation lifecycle](delegation-lifecycle.md), [attribution](proxy-attribution.md); HC-05-10. |
| D06 | **Ownership and membership/sync lifecycle:** remaining Q-096 recommendation, ownership transfer/absence handling, and optional-sync authority/lifecycle boundaries. | Explicit human membership, team versus grant lineage, no transitive human membership, no automatic personal-authority import on owner rotation; no implicit two-person approval. | [Q-096 remainder](authority-lineage.md), [ownership](ownership-lineage.md), [groups](groups-and-membership.md); HC-05-12. |
| D07 | **Identity trust integration:** trusted tenant mapping, allowed identity sources/adapters, and legacy-token normalization without a proxy bypass. | Shared actor/human block, direct-human equal IDs, human JWT `sub`, and proxy enforcement coverage are approved. | [identity](identity-context.md), [JWT mapping](jwt-identity-mapping.md); HC-03-05. |
| D08 | **Authority-change timing:** confirmation/visibility for membership and dependency changes, in-flight authorization changes, and authorization conflict/retry behavior. | No stale-grant grace for checks begun after confirmed deletion; preserve application boundaries through use and checked authority through Auth writes. | [freshness](authority-freshness.md), [concurrency](concurrent-enforcement.md), [Auth writes](auth-write-consistency.md); HC-09-06, HC-09-07. |
| D09 | **Operation coverage and composition:** whether/how complete routes cover both move states or multiple batch items; remaining collection/count/export and mutation coverage cases; distinguish authorization failure from later execution failure. | Both move boundaries need authority; deny uncovered collection requests; authorize the complete batch before protected effects. No automatic subset filtering or implicit rollback guarantee. | [moves](operation-enforcement.md), [collections](collection-enforcement.md), [bulk](bulk-enforcement.md); HC-09-03, HC-09-04, HC-09-05. |
| D10 | **Background/running operations:** retry and running-job authorization boundaries, recurring/streaming integration, or explicit v1 deferral of unsupported cases. | Queued execution reauthorizes with trusted human/proxy context and current support; enqueue-time allow is not sufficient. | [background](background-authorization.md); HC-09-09. |
| D11 | **Residual authorization conditions:** inventory existing condition references, identify any real authorization-only requirement not covered by the agreed mechanisms, then decide its v1 treatment. | Q-084 business-rule proposal is disapproved and its revised question withdrawn. Neither a condition engine nor silent removal of restrictions is approved. | [Q-084](grant-conditions.md); HC-08-03. |
| D12 | **Handbook applicability:** products/consumers governed and mandatory compliance scope. | Auth's own APIs and application integration follow the agreed architecture; no implementation certification is implied. | [roadmap stage 1](handbook-roadmap.md), [closure register](handbook-completion-audit.md); HC-01-03. |
| D13 | **Rule-change governance:** who approves canonical changes and how any exceptions are governed without quietly bypassing core bounds. | User-approved decisions remain source of truth; no automatic deletion or rewriting of earlier decisions. | [roadmap stage 1](handbook-roadmap.md), [closure register](handbook-completion-audit.md); HC-01-04. |

D03 contains the existing parked Q-117; it is not excluded or silently approved.
D06 includes the earlier ownership remainder, not a reopened Q-099 decision.
D11 requires an assistant inventory before any question: do not simply repeat
the rejected proposal under a new number. D09/D10 cover authorization integration,
not business transaction-engine or workflow design. D12/D13 can wait until the
high-impact authority decisions are addressed.

## Five contract reviews — assistant drafts, user approves new choices

| Ref | Package to prepare | What I can resolve and record from current approval | What still needs review / dependencies |
|---|---|---|---|
| C01 | **Authority support and boundary-validator contract** | Actual supporting assignments, top-down lineage-supported latest, complete-route provenance, subset/AND checks, team ceiling and no issuer-dependency invention; diagnostic expected outcomes. | Eligible-support discovery/evidence inputs/outputs, ambiguous/corrupt support representation, cycle/depth validation limits and exact admin request boundary. Do not add withdrawn parent selector fields. D01/D02/D05/D06 can affect these contracts. |
| C02 | **Grant/role/revision/assignment lifecycle contract** | Assemble Q-107/109/118/119 fields and approved live/immutable/assignment distinctions; encode already-agreed alternatives and lifecycle cases. | Full ID/type/numeric/timestamp/default validation, unknown fields, deletion/adoption operation contracts, compatibility and retention. Automatic root mechanics depend on D02; residual restrictions on D11. Existing shapes do not settle these remaining defaults. |
| C03 | **Membership, hierarchy, ownership and delegation records** | Keep each relationship distinct, apply parent/team and human ceilings, and align examples with Q-099/101. Historical scratch JSON is available as evidence, not automatically canonical. | Exact new record shapes/fields and operation contracts, depending on D05/D06. Do not infer a new recipient type, automatic owner grant, or sync behavior. |
| C04 | **Identity and registration contracts** | Reuse approved identity/JWT blocks and registration ownership/validation rules; document trusted versus claimed inputs and existing rejection meanings. | Full registration publication/change format, namespace/naming validation, trusted tenant mapping, provider/legacy adapters, supported scope definitions, and compatibility/distribution contracts. Depends on D01/D04/D07. |
| C05 | **Endpoint/request/resolved/result/agent contracts** | Reuse one-permission declarations, path/body binding, required presence, application-owned value validation, allow/deny/error blocks, both UI messages, and result-field rejection rules. | Nested input syntax/additional sources; complete request/resolved transports and dependency evidence; reason/value validation, handler invocation/failure contract and operation/batch/background bindings. Depends on C01 and D08/D09/D10/D11. Do not add returned scope rejected by Q-061. |

These reviews need not ask about every field individually. Present a coherent
versioned block, label each already-approved field, and ask only about the new
contract choices. A review may shrink if further source reconciliation proves
a listed choice was already settled; record that evidence rather than re-ask.

## Eight autonomous work packages

| Ref | Work I can run and record | Can start now? | Completion boundary |
|---|---|---|---|
| A01 | Draft C01 from actual lineage/source rules, with source-linked examples and unresolved evidence slots explicitly labeled. | Yes | I can settle consequences of existing rules, not invent governing source selection or approve new evidence fields. |
| A02 | Draft C02 and its field/lifecycle validation matrix. | Yes | Already-approved field behavior can be recorded immediately; unsupported defaults/variants remain proposals. |
| A03 | Draft C03 and reconcile historical membership/ownership examples. | Yes | Preserve original records; new canonical relationship fields need review. |
| A04 | Draft C04 with existing identity and registration semantics separated from open trust choices. | Yes | No invented provider trust, tenant mapping, registration authority, or new root rights. |
| A05 | Draft C05, including exact approved result variants and endpoint examples. | Yes | New transports, source grammar, reason values, and boundary-specific choices remain review items. |
| A06 | Build and check the adversarial and HRMS/repository expected-outcome matrices against agreed rules. | Yes, for settled cases | Final coverage depends on the accepted contracts. Diagnostic checks/documentary scenarios are not a production Auth engine test. Missing-policy cases must not be marked passing. |
| A07 | Reconcile stale open labels and current prose/reader/diagrams, preserving source decisions and superseded rationale. | Yes | No meaning change without approval; final whole-handbook consistency waits for the last decisions. |
| A08 | Assemble the versioned contract package, verify links/JSON/reader build, and recalculate criterion closure with evidence. | Structure/checks now; final package after reviews | Package only approved contracts as canonical. Do not self-approve v1 acceptance, commit/push, or deployment. |

**What I can fully resolve without you:** contradictions that are already
answered by an approved later decision, faithful consolidation, consequences/
examples of those decisions, documentary and static checks, and evidence-based
status updates. **What I cannot resolve alone:** genuine new authority behavior,
trust/ownership choices, new canonical fields or defaults, scope deferrals, and
final acceptance. All eight packages involve real work I can perform, but are
not eight guaranteed whole-checkpoint closures without intervening decisions.

## Already settled — do not count as pending discussions

| Earlier open/proposed wording | Current evidence and disposition |
|---|---|
| Role versus explicit grant content | Q-118 settles exclusive sources and both-role-fields requirement. Full value validation remains C02, not a re-vote. |
| Root parent representation | Q-119 settles omission for trusted roots. Initial trust remains D01/D03, not a repeated format question. |
| Whether root expansion needs a separate manual approval | Q-120A says no. Only mechanism/trust boundary remains D01/D02. |
| Parent ceilings or an independent parent-revision/assignment field | Q-101/Q-112A retain parent ID and actual supported lineage; extra fields are withdrawn. C01 defines evidence, not a new principle. |
| Single permission, relationship block, required inputs, value ownership | Q-049/Q-050-B–F settle these. Remaining source grammar/transport validation belongs to C05. |
| Optional compatibility validation mode and activation | Q-040–Q-042 settle upfront mode and safeguards. Definition evolution remains D04. |
| Mandatory returned scope/boundaries with allow | Q-061 says not required; Q-062 gives the minimal allow. Do not treat old returned-boundary proposals as pending requirements. Supporting evidence work must respect that decision. |
| Nested human membership or independent automated authority | Not supported by current rules. Subteam authority is supported, without transitive human membership. |
| Owner changes necessarily break team-held authority | Q-099 answers no; actual required personal dependencies remain. Q-096's remaining ownership policy is D06. |
| Revoke as a separate operation, assignment validity window, or automatic live roles | Delete is the canonical permanent operation; assignment validity is deferred in v1; role/grant adoptions remain explicit for ordinary grants. |
| Business-rule engine and external audit design | Q-084 rejects business-logic expansion; Q-076 excludes external audit design. Neither is a new implementation task here. Residual authorization-only conditions remain D11. |

## Coverage of every currently open criterion

This maps the **30 OPEN rows** in the original register to the agenda. It does
not close rows or change their criteria. A D/C/A reference here is a dependency,
not additional question-count credit. Final acceptance is F01.

| Checkpoint | Discussion/review dependencies | Assistant work / final action |
|---|---|---|
| HC-01-03 | D12 | A07 |
| HC-01-04 | D13 | A07 |
| HC-03-05 | D07, C04 | A04, A06 |
| HC-04-03 | D01, D04, C04 | A04, A07 |
| HC-05-08 | C01 | A01, A06 |
| HC-05-09 | D01, D02, D03, C02 | A02, A06 |
| HC-05-10 | D05, C01, C03 | A01, A03, A06 |
| HC-05-11 | C02 | A02, A06 |
| HC-05-12 | D06, C01, C03 | A01, A03, A06 |
| HC-05-13 | D02, C01, C02 | A01, A02, A06 |
| HC-06-06 | D01, D04, C04 | A04, A06 |
| HC-06-07 | D04, C01, C04, C05 | A01, A04, A05, A06 |
| HC-07-07 | C05 | A05, A06 |
| HC-07-08 | D02, D11, C02 | A02, A06 |
| HC-07-09 | C01, C04, C05 | A01, A04, A05 |
| HC-07-10 | C05 | A05, A06 |
| HC-08-02 | C05 | A05, A06 |
| HC-08-03 | D11, C02, C05 | A02, A05, A06 |
| HC-08-04 | C01, C05 | A01, A05 |
| HC-09-03 | D09, C05 | A05, A06 |
| HC-09-04 | D09, C05 | A05, A06 |
| HC-09-05 | D08, D09, C05 | A05, A06 |
| HC-09-06 | D08, C01, C03, C04 | A01, A03, A04, A06 |
| HC-09-07 | D08, C01, C02, C05 | A01, A02, A05, A06 |
| HC-09-09 | D10, C04, C05 | A04, A05, A06 |
| HC-10-03 | Finalized relevant D/C topics | A06 |
| HC-10-04 | Finalized relevant D/C topics | A06 |
| HC-11-01 | Finalized relevant D/C topics | A07 |
| HC-11-02 | Accepted C01–C05 | A08 |
| HC-11-03 | F01 | User final acceptance after A06–A08 |

## Execution recommendation and limits

Proceed with A01–A05 drafts and the settled portions of A06/A07 while discussing
one high-impact decision at a time: D01/D02 first, then source/delegation/identity
and definition lifecycle, followed by operational timing and final governance.
Use A08 for evidence-backed closure, not to inflate the score through question
counts. Do not ask a new question for an already-resolved rule just to fill a
planned discussion slot.

This is the current assessed agenda, not a promised number of messages or a
finish-time estimate. A newly found contradiction or a user-selected new feature
can alter it; record additions/removals with reasons rather than silently moving
the goalposts. Where a review shows no new semantic choice remains, resolve and
record it directly using its existing approval.

The handbook measurement remains **38/68 closed, 30 open, one excluded**.
No pending topic is approved, deferred, excluded, or implemented by assessment.
Historical source content and tag `0.0.2` remain unchanged. New recording stays
in the lab; no commit, push, or scratchpad mutation is included.
