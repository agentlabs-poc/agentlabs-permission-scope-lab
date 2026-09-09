# AUTH-TEST-01 — model coverage expansion

## Scope and stop condition

User approved covering C01–C14 after the local middleware slice. This pass maps
existing evidence and adds tests for supported behavior, not unsupported features
or new policy. Two independent Sol-medium test-only units have 15-minute bounds:
middleware evidence/evaluation tests and SQLite adapter/handler tests. Controller
owns this map and full verification. One focused correction is allowed if needed;
an implementation defect is reported before broadening into a production change.
No reviews, schema changes, new dependencies or production integration.

Characterization tests may pass against existing correct behavior. They are not
claimed as RED/GREEN feature implementation. Each test must name a real boundary
failure it catches; success is not a claim of exhaustive model coverage.

## Coverage map

Paths below are relative to `implementation/`. Existing test files are evidence
anchors, not a claim that every combination within a row is tested.

| Ref | Existing executable evidence | Remaining distinction |
|---|---|---|
| C01 direct users | `abv/internal/lineage/human_test.go`: `TestResolveHumanRejectsInvalidRequestAndDirectUserAssignment`; `source_test.go`: `TestHasSourceLeavesDirectHumanAndSelfBindingExplicitlyUnsupported` | Rejection covered; successful direct-recipient discovery is not implemented. |
| C02 proxy/service accounts | `authmiddleware/evidence_test.go`: `TestEvaluateRejectsMalformedRequestBeforeLoading` | Unsupported identity rejection is not proxy delegation/lifecycle coverage. |
| C03 group self | `authmiddleware/evaluate_test.go`: `TestEvaluateUsesDirectHumanForGroupSelf`; `abv/internal/httpdemo/httpdemo_test.go` | Trusted resolved fixtures only; SQLite issuance/resolution remains unsupported. |
| C04 lineage combinations | `abv/internal/lineage/human_test.go`, `resolve_test.go`, `dependents_test.go`; `authmiddleware/evaluate_test.go` | Existing independent-route, parent AND, missing support and traversal tests; strengthen combinations without inferring membership inheritance. |
| C05 lifecycle | `abv/internal/mutation/grant_status_test.go`: `TestSetGrantStatusPreservesDescendantStateAndEffectiveness`; `assignment_status_test.go`: bottom-up, fork, restore and expiry tests | Extend runtime SQLite-to-evaluator sequences; assignment status and grant control remain separate. |
| C06 revisions | `abv/internal/lineage/resolve_test.go`: actual team/root role revision tests; `abv/internal/mutation/grant_revision_test.go`; `abv/grant_revision_test.go` | Publication and pinned reads tested; explicit adoption implementation remains separate, not closed by these tests. |
| C07 two authority gates | `abv/internal/mutation/service_test.go`: `TestCreateAssignmentPersistsExactProposalAfterBothChecks`, `TestCreateAssignmentFailuresDoNotWrite`, administrative-evidence isolation | Tests independently remove administration or source authority and assert no receipt/write; not a real Auth administrative middleware integration. |
| C08 consistency | `abv/internal/mutation/consistency_test.go`: writer conflict and committed disablement; SQLite reader tests | Mutation transaction races and local snapshot behavior are narrower than distributed runtime freshness. |
| C09 application state | `abv/internal/httpdemo/httpdemo_test.go`: FIN/C18 GET/PUT and unchanged ENG data | Constrained synchronous demo only; cross-boundary moves and concurrent application transactions are not fully modeled. |
| C10 catalog/root | `abv/internal/mutation/catalog_test.go`; `abv/internal/lineage/root_catalog_test.go`: computed root expansion, unchanged ordinary selection, context isolation and compatibility | Controlled lab setup is not the complete production bootstrap/recovery lifecycle. |
| C11 batches | No batch endpoint in the local slice | Needs implementation before positive end-to-end coverage; do not claim per-request tests prove atomic batches. |
| C12 moves | Misleading PUT/current-row mismatch rejection in HTTP demo | Positive move composition needs an exact contract, not a guessed test oracle. |
| C13 pagination/retry/streaming | `authmiddleware/http_test.go`: `TestAllowedSynchronousEffectCompletesAfterAuthorityWithdrawalAndNextRequestDenies` | Fresh subsequent checks tested; no pagination, streaming or long-running authorization contract implemented here. |
| C14 robustness/performance | Codec rejection, traversal/material ceilings, race suites; `abv/acceptance_test.go`: `BenchmarkBoundedAuthorityDiagnosis` | Add bounded generated checks; benchmarks are diagnostics, not an agreed throughput/SLA guarantee. |

## Explicit exclusions

Production JWT/Auth transport/distributed freshness are outside the local build.
PostgreSQL, deletion and reparenting remain user-deferred. A fail-closed rejection
test documents an unsupported path; it does not complete that path's model cases.
No percentage is inferred from test counts or from listing fourteen categories.

## Execution evidence

C03/C04/C14 middleware additions:

- `TestEvaluateAddingPredicateOnlyNarrowsRoute`: adding FIN preserves FIN access
  but denies ENG and all-department requests previously allowed by the broad route.
- Expanded `TestEvaluateUsesDirectHumanForGroupSelf`: exact human allows;
  another identity and all-humans selection deny.
- `TestEvaluateAcceptsExactWorkLimits`: the allowed maximum material, routes,
  contributing grants and predicates work, complementing existing overflow tests.
- `TestEvaluateRejectsMalformedRequestBeforeLoading`: attributed service actor is
  rejected before loading authority; no proxy support is claimed.

Existing complete-route isolation tests are reused rather than duplicated.
Controller reran middleware full uncached tests, full race, vet and build: pass.

C05/C08 added `TestSQLiteAuthoritySourceSeesProtectedDescendantStatusChanges`
and `TestSQLiteHTTPDemoTracksProtectedDescendantAssignmentAndGrantControls`.
Both create A2 through the protected lab API and observe Nutan's C17 access:
allow → assignment disabled/deny → explicitly enabled/allow → grant disabled/deny
→ explicitly enabled/allow. HTTP denial must not disclose the protected title.
This proves committed local reads through real controls, not a cached fixture.
Original parent A1/G1 status tests remain: their direct database edits are
explicit test-state injection, not an authorized public mutation workflow.

C14 added `authmiddleware/codec_fuzz_test.go` with
`FuzzDecodePolicyPreservesBindings` and `FuzzDecodeResultPreservesVariant`.
They check accepted wire round trips preserve bindings, grants and messages;
errors return zero policy/result, including the distinct evaluation-error variant.
Six seeds each include valid and hostile inputs. Bounded runs use
`go test -run '^$' -fuzz '^<name>$' -fuzztime=5s -parallel=2 -timeout=30s`:
122,805 policy and 88,562 result executions pass. These counts are generated
inputs, not distinct model use cases. Seed regressions also run in ordinary tests.

Controller baseline: full uncached ABV suite passes. Existing bounded diagnostic
benchmark rerun with `go test -run '^$' -bench '^BenchmarkBoundedAuthorityDiagnosis$'
-benchtime=1x -count=1 -timeout=60s .`: 100 / 1,000 / 10,000 records complete in
approximately 0.85 / 5.78 / 58.51 ms per diagnostic, respectively. These single
iterations are a smoke check, not statistically stable performance claims or
API evaluator throughput measurements. No performance optimization is implied.

Final controller verification: both Go modules pass `go test -count=1 ./...`,
`go test -race -count=1 ./...`, `go vet ./...` and `go build ./...`; diff check
passes. No production behavior changes or defects were found by this pass.
Two Sol-medium units completed within their bounds; parent-state injection tests
were explicitly retained alongside the new protected-lifecycle tests.

This bounded pass is complete, not all C01–C14. Positive direct/proxy/self SQLite
paths, adoption, batch/move/streaming contracts, concurrent application-data moves
and distributed freshness remain as qualified above. Real SQLite alternative-route
mutation and deterministic mid-loop cancellation are additional test gaps; existing
unit route/cancellation tests are not mislabeled as those integration proofs.
