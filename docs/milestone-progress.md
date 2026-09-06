# Milestone progress — through Q-111

## What these percentages measure

Treat the eleven existing roadmap stages as milestones. For each stage:

`remaining percentage = OPEN checkpoints / (DONE + OPEN checkpoints) × 100`

The source is the fixed [MEASURE-001 closure register](handbook-completion-audit.md).
No subjective effort weights or fractional credit are added. A partially finished
checkpoint stays OPEN until its entire criterion is documented and settled.
HC-09-08, external audit-system design, remains EXCLUDED under Q-076; it is not
completed work. Q-108 defers one capability within a broader lifecycle/schema
checkpoint, not that whole checkpoint, so it does not reduce the denominator.

This report recalculates per-stage percentages and reconciles affected evidence
through Q-111. It is a checkpoint-closure view, not a full new semantic acceptance
audit, an estimate of hours/questions remaining, or a production-security score.
The stage granularity is the existing tracking rubric, not a new canonical policy.

## Remaining work per milestone

| Milestone | Closed / in scope | Open | Remaining | Main outstanding deliverables |
|---|---:|---:|---:|---|
| 1. Purpose and architecture | 2 / 4 | 2 | 50.0% | Audience/applicability and rule/exception governance; not a missing request-flow diagram. |
| 2. Principles | 4 / 5 | 1 | 20.0% | Consolidated numbered principle catalog, mandatory versus guidance wording, compliance examples. |
| 3. Vocabulary and identity | 3 / 5 | 2 | 40.0% | Complete glossary and trusted identity/tenant mapping contracts. |
| 4. Permissions | 3 / 4 | 1 | 25.0% | Catalog ownership, naming governance, and evolution rules; grammar and no wildcard/alias decisions already exist. |
| 5. Grants and authority | 7 / 13 | 6 | 46.2% | Administrative boundary contracts, bootstrap, delegation, full lifecycle, team/membership/sync, and revision mechanics. |
| 6. Scope and registration | 5 / 7 | 2 | 28.6% | Registration/change/distribution contracts and reference/existence or application-specific boundary handling. |
| 7. Requests and resolution | 6 / 10 | 4 | 40.0% | Complete endpoint/grant/role schemas, request/resolved transports, and handler-agent integration contracts. |
| 8. Decision semantics | 1 / 4 | 3 | 75.0% | Finish result/reason validation, authorization-only condition treatment, and dependency/provenance representation. |
| 9. Enforcement and time | 2 / 8 | 6 | 75.0% | Detailed collection/write/bulk contracts, freshness, concurrency, and background-operation integration. |
| 10. Challenge and verify | 2 / 4 | 2 | 50.0% | Final-rule adversarial expected outcomes and end-to-end HRMS/repository review. |
| 11. Publish the foundation | 0 / 4 | 4 | 100.0% | Final editorial reconciliation, contract package, explicit v1 acceptance, and implementation roadmap. |
| **Total** | **35 / 68** | **33** | **48.5%** | **One additional checkpoint remains excluded, not completed.** |

Overall closure is **51.5%**. Sum checkpoint counts to obtain the overall result;
do not average the eleven percentages, because stage sizes differ. Display values
are rounded to one decimal place from integer counts.

## Why recent approvals have not moved the strict closure score

The old rows deliberately bundle complete deliverables. Recent work advances
several of them substantially, but none of the following rows is wholly closed:

| Recent settled work | Affected open rows | What still prevents closure |
|---|---|---|
| Separate Auth administrative and source-boundary gates; parent-dependent narrowing, actual adopted lineage | HC-05-08 | Complete validator/administrative contracts and eligible-support discovery/evidence. Q-112A subsequently reaffirms lineage-supported latest; the governing selection rule is not an open vote. |
| Assignment uniqueness, revision adoption, latest-only creation/upgrades, immutable content | HC-05-13 / HC-07-08 | Full role/grant variants, lifecycle/compatibility and operation/validation contracts. |
| Core three-record JSON; grant validity on revisions, assignment windows deferred | HC-05-11 / HC-07-08 | Complete lifecycle transitions, initial/default rules, timestamp validation, and full schemas. |
| Orphan consequences, bottom-up binding changes, ownership/support separation, cycle prohibition | HC-05-12 | Full hierarchy/membership/ownership/sync contracts and revision-aware validation details. |
| Auth checked-state preservation through assignment writes | HC-09-07 | Persistence/conflict/retry contracts and general in-flight authority consistency. |

Several other stages also have approved governing rules inside OPEN rows—for
example allow/deny/error outcomes, collection denial instead of automatic filtering,
and complete-batch authorization. **75% remaining in stages 8 or 9 does not mean
75% of their fundamental ideas are missing.** It means three of four, or six of
eight, complete-deliverable checkpoints remain unclosed.

This is the limitation of the coarse rubric. For finer progress within an OPEN
row, define stable independently closable sub-checkpoints first and measure them
separately. Do not assign “80% done” by impression or silently change the existing
denominator; any new breakdown should be labeled as a new measurement series.

## Recommended remaining-work order, not new policy approval

1. Close authority risks: direct-human supporting-revision selection, complete
   Auth boundary validation, bootstrap source authority, and delegation limits.
2. Finish freshness and remaining concurrent-change guarantees.
3. Complete canonical schemas, result contracts, and integration boundaries.
4. Run cross-domain/adversarial review, reconcile terminology, and obtain v1
   acceptance with every remaining item settled or explicitly deferred/excluded.

No next authorization question is introduced during this status response.
Discussion remains one referenced proposal at a time when resumed. No completion
score here authorizes implementation, commit, push, or release.
