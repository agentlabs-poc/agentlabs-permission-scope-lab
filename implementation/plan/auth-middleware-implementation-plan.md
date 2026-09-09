# AUTH-MW-01 — API-side Auth Middleware / Evaluator implementation plan

**AUTH-MW-04 — user-approved local test adapter:** use an adapter that reads the
existing ABV SQLite database directly for local testing. ABV's existing SQLite
provider already tests its own validator; for the new middleware/evaluator,
the adapter supplies current authority without requiring Auth HTTP integration.
SQL access remains outside API-side evaluation logic. Reuse existing schema and
read-only resolution where possible; do not embed the ABV mutation engine or
invent a second authority model. This supersedes the fixture-only authority
source restriction below, not the trusted-source/transaction/freshness rules.

> For implementers: use bounded subagent-driven development after the contract
> checkpoint is approved. No review passes per user direction; test-first and
> verification remain mandatory. This request pins placement and requests a plan,
> not implementation of unfinished transport or authority-evidence policy.

**Goal:** implement reusable API-side request authorization with canonical
endpoint policies, complete current authority and request-bound enforcement.

**Architecture:** API services host the Auth Middleware/evaluator. Auth Service
owns authority and hosts ABV, the core Auth Validator for definitions and
authority-changing operations. The API-side component consumes authority through
a read-only adapter, not direct Auth tables or an embedded mutation validator.

**Tech stack proposal:** Go 1.25-compatible, standard-library net/http adapter,
in-process reusable evaluator, injected trusted identity/authority adapters and
clock. No new database, web framework, cache, policy DSL or network service.

**Spec:** handbook/implementation/05-canonical-model.md,
06-auth-service.md, 07-application-integration.md; handbook/theory/canonical-terms.md;
docs/collection-enforcement.md, decision-results.md, auth-write-consistency.md,
authority-freshness.md and concurrent-enforcement.md. Read their current refinements,
not historical prepared or live-role proposals.

## A. Pinned placement — AUTH-ARCH-02

User explicitly confirmed: ABV is used inside Auth Service itself; other components
are at the API-service end, then requested this be pinned.

```text
API service
  Auth Middleware / evaluator
    verified identity + endpoint policy + request material + Auth authority
  Endpoint handler
    application facts and constrained data/effects

Auth Service
  Its own API-side authorization using the same framework
  Core Auth Validator / ABV
    registered definitions, proposed authority, source and lineage integrity
  Protected coordinator and authorization storage
```

API services do not install ABV's mutation engine or SQLite provider. Sharing
canonical records or read-only resolution rules is not deploying ABV there.
Auth's own writes retain both administrative and ABV checks through persistence;
an earlier middleware allow cannot be used as a stale write authorization.

## B. Understanding that governs implementation

1. The endpoint declares one permission covering its entire protected operation.
   The policy is server-owned; HTTP method alone or request body cannot select it.
2. Inputs have exact declared sources and are all required. Request schema/value
   validation belongs to the API, not duplicated policy fields or scope rules.
3. Identity is verified actor plus authorizing human. Tenant/application must be
   established and match the selected request context. Body claims are not proof.
4. A grant definition is not access. Authority comes through current direct/group
   assignments, human membership, selected grant/role revisions, controls, validity
   and actual required parent/team support. Resolution never manufactures authority.
5. Evaluate complete routes. Parent restrictions remain AND; unrelated permission/
   scope fragments cannot be combined. Another complete route can allow even when
   one is inapplicable. An orphaned or disabled required route supplies no authority.
6. Scope keys have registered application meaning. `$self` is the authorizing
   human, including group-derived access. `{}` removes no inherited/outer limit.
7. Auth does not learn certificate ownership from a URL or query the application
   database. Endpoint code supplies needed facts or binds execution to the requested
   boundary. It does not inspect grants or turn grant scopes into bespoke SQL.
8. One endpoint-owned gate; no prepared handoff. Middleware is a route-aware wrapper
   callable where sufficient material exists, not necessarily only an early filter.
9. An explicitly FIN-bounded collection may pass; an all-departments request cannot
   be quietly narrowed to FIN. Partial coverage does not become partial success.
10. Allow/deny are completed decisions. Evaluation error is separate and fails
    closed. Preserve approved version/messages/grant references; no returned-scope
    field or reusable authorization ticket is introduced.
11. New checks cannot use authority withdrawn before them. Existing bounded
    synchronous allow has Q-129's limited completion behavior; retries need a new
    decision and changed data must still obey Q-074. Auth writes obey stronger Q-110.
12. Team administration, ownership, business access and grant assignment remain
    different powers. Ordinary application request evaluation is not ABV issuance
    validation. 123 storage changes none of these rules.

## C. Contract checkpoint M0 — settle before writing the evaluator

**Bound: 25 minutes; output:** auth-middleware-contract.md in this plan directory.
One 10-minute correction maximum, then isolate the specific unresolved choice.
No new fields may be called canonical merely because a Go interface needs them.

The handbook already settles policy and minimal results, but not the full
Auth-to-API evidence transport or embedded-agent interfaces. M0 must specify:

| Seam | Required explicit contract |
|---|---|
| Trusted identity adapter | Canonical actor/human context, verified tenant/application; never decode JWT as proof or trust caller identity JSON. Real JWT integration is a separate adapter deliverable. |
| Authority source | Read-only complete applicable authority for the established human/actor/context, exact adopted lineage and freshness. Timeout, malformed or incomplete evidence cannot masquerade as a complete empty result. |
| Evidence location | Recommend Auth supplies dependent resolved-route views; API evaluator matches the request. This is a proposed internal seam, not a new wire grant type or independent entitlement. If raw records are supplied instead, identify shared read-only resolver reuse without importing ABV mutation/storage. Choose one for the first slice. |
| Route evidence | Context, contributing grant references, supported permission selection, accumulated predicates and mandatory lifecycle/delegation constraints; specify which checks Auth proves and which the API must recheck at decision time. No omitted condition is assumed satisfied. |
| Material binding | Explicit application code maps policy inputs/established facts to registered boundary meanings. Equal names alone do not prove mapping. No new relationship/resolver block or expression language. |
| Request scope | Distinguish a concrete record/explicit boundary from an all-boundaries collection; don't infer broad authority from today's returned rows. Exact internal representation is required. |
| Decision transport | Reuse Q-062–Q-067 allow/deny/error shapes; complete variant validation. HTTP status/error-code catalogue additions remain proposed until approved. |
| Package reuse | Inventory existing domain/codec/lineage helpers. Reuse canonical types/semantics without bringing ABV writes/SQLite into API services; any minimal extraction must preserve ABV tests. No speculative shared-framework rewrite. |

An authority provider is not ordinary client input. Only a trusted adapter can
supply evidence. Any snapshot/freshness metadata introduced internally must be
labeled internal and have an actual enforcement contract; a timestamp is not a
proof of fresh authority. No new Auth HTTP endpoint or guessed response shape is
implemented to fill this seam. A deterministic in-process provider can prove
consumer behavior, but cannot prove production network freshness/integration.

## D. First working slice — proposed acceptance scope

Recommended code home after M0: `implementation/authmiddleware/`, separate from
ABV mutation/provider code; choose exact module/type reuse in M0 before scaffolding.
Use one reusable evaluator and one net/http wrapper, not separate engines per API.

First slice: synchronous requests, verified direct-human callers, trusted complete
team-held route evidence, registered literal boundaries and runtime `$self`,
single-record GET, selected-body PUT without ambiguous move composition, and an
explicitly bounded collection with broader-request rejection. Preserve the model's
other capabilities as pending consumer support, not deleted canonical semantics.

Source issuance of recipient-relative self grants remains ABV's unresolved
distribution question; do not treat it as blocking the meaning of `$self` on an
already valid group-derived runtime route. If M0 cannot supply a safe complete
evidence contract for a case, keep that case unsupported and visible, not guessed.

Direct-recipient support discovery, full proxy delegation evidence, move-route
composition, arbitrary batch/stream/queued execution and caches need their own
bounded slices/contracts. Unsupported cases fail closed with diagnostic meaning.
This scope is a proposal to get the first harness working, not a handbook deferral.

## E. Implementation tasks after M0 approval

Each implementation unit has a 20-minute first attempt and one 10-minute focused
correction maximum. If Sol-medium cannot deliver, report the blocker and escalate
once to Astra-medium on the narrowed remainder, at most 15 minutes. A larger model
cannot approve missing policy. Reassess the slice after 60 minutes of coding.

Task files below are responsibilities; M0 fixes exact exported type signatures
and dependency reuse before generating the worker briefs. Do not dispatch workers
against undefined interfaces. This plan is not an executable SDK contract yet.

### M1 — canonical policy and result handling

**Files:** policy.go, policy_test.go, result.go, result_test.go in the new component.

- [ ] RED: one declared permission; missing/unsupported versions; missing policy;
  unknown/duplicate fields under supported schemas; exact path/body source presence;
  valid allow/deny/error and rejected mixed/truncated variants.
- [ ] Implement strict supported policy/result contracts, reusing proven codec
  logic where accessible. No relationship block, type/nullability fields, caller
  permission choice or scope fields in allow. Test result messages reach the adapter.
- [ ] GREEN focused tests, package tests and diff check; report exact APIs and limits.

The exact existing policy example in section F and the handbook's three result
variants are fixtures. Deliberately unsupported source/nested selectors are errors,
not silently ignored fields. Strictness for unfinalized schema extensions is a
prototype limit, not a newly approved general contract.

### M2 — evaluator over trusted complete evidence

**Files:** evaluate.go, evaluate_test.go, evidence.go, evidence_test.go.
Consumes M0 evidence/material/context types and M1 result types.

- [ ] RED: complete applicable route allows; wrong permission/boundary denies;
  cross-area mismatch stops; parent AND survives; no cross-grant permission/scope
  mixing; group self uses human identity; exact adopted role/revision behavior;
  disabled/expired/orphaned support cannot allow; a separate valid route can allow.
- [ ] Implement pure request evaluation using the chosen M0 evidence contract.
  No grant writes, SQL, HTTP, application-DB lookup or inferred business meaning.
  Validate evidence provenance/context/shape/completeness at the declared seam.
- [ ] GREEN with deterministic clock/provider tests: complete empty evidence means
  no applicable authority; failed/incomplete loading means evaluation error. Enforce
  finite work bounds and reject unsupported constraints rather than dropping them.

### M3 — API-side wrapper and binding

**Files:** http.go, http_test.go; small identity/provider adapter definitions only
if not already in the M0 contracts. Consumes M1/M2 stable signatures.

- [ ] RED: missing/invalid policy or identity never calls the protected handler;
  path tenant mismatch; body source cannot fall back to query/path; unknown action;
  deny/error/malformed result stops output/effects; only matching allow calls once.
- [ ] Implement one route wrapper over existing authenticated context, request
  schema validation and explicit application binding functions. Parse body once
  or preserve exact validated values; handler and evaluator cannot see different
  interpretations. Validate bindings at route registration where possible.
- [ ] GREEN with net/http/httptest: GET, selected-body PUT, cancellation, failure
  propagation, no unverified body identity, no second permission selection, and
  bounded input sizes. Authentication cryptography is not invented here.

### M4 — executable API-service harness and integration proof

**Files:** cmd/authmiddleware-demo/main.go; integration_test.go; README.md.
Test provider/fixtures remain explicitly lab-only, not Auth authority publication.

- [ ] RED: FIN/C17 request cannot disclose an ENG certificate; group self cannot
  disclose another human's record; insufficient all-department request is denied
  rather than filtered; body claim does not prove current record ownership.
- [ ] Implement a minimal in-memory application-data demo with constrained handler
  operations and trusted fixture authority source. Include a provider timeout and
  an authority reduction between requests; do not claim a distributed freshness
  protocol from an in-memory test. No production Auth server integration.
- [ ] GREEN: full/race tests, vet/build, actual HTTP request/response walkthrough;
  M1–M3 rejection cases prevent side effects. ABV regression suite still passes if
  shared contracts/helpers were touched. Record limits and pending adapters.

Parallelism: after M0, M1 and an M4 fixture-only scaffold may run independently
when truly useful; don't create speculative scaffolding just to fill agent slots.
M2 depends on policy/evidence types; M3 depends on the evaluator; final M4 depends
on all. Source-contract decisions are not parallelized into inconsistent designs.
While coding runs, documentation and independent adapter tests can be separate
owned-file tasks. No independent review passes, as explicitly requested.

## F. Canonical API policy — unchanged

```json
{
  "version": "1",
  "method": "GET",
  "path": "/api/v1/{tenant}/{dept}/{cert}",
  "permission": "hrms:employee:certificate::read",
  "inputs": {
    "tenant": {"source": "path", "name": "tenant"},
    "dept": {"source": "path", "name": "dept"},
    "cert": {"source": "path", "name": "cert"}
  }
}
```

The application's request schema validates values; its binding/enforcement code
connects registered dept semantics to the operation. `FIN` in the URL is a requested
boundary, not evidence of certificate ownership. The handler's constrained read
uses the same tenant, FIN and C17. Required inputs remain required even under `{}`.
The pinned `/api/v1/{tenant}/abv/` base belongs to ABV APIs, not every application
route this middleware protects. Middleware is independent of the 123 table layout.

## G. Definition of done and exclusions

First slice is done only with the working reusable evaluator/wrapper, executable
bounded harness, negative tests and explicit remaining integration contracts.
No claim of general production Auth integration, new canonical wire schema,
complete delegation support or whole-handbook completion follows.

The first source may now be a local SQLite-backed test adapter under AUTH-MW-04.
It must load coherent actual stored authority, not treat a grant row as access.
Use disposable existing-schema fixtures; verify assignment/membership/control
changes between requests affect evaluation, and prove the adapter performs no
authority writes. Its exact package seam and driver dependency are fixed in M0;
the core evaluator imports neither SQL nor the ABV mutation coordinator.

Real Auth authority transport/freshness, established JWT adapter wiring and rollout
need separate approved integration work. Core ABV and its protected coordinator
remain inside Auth Service, unchanged. Storage consolidation and this API-side
component are separate plans; neither waits for a speculative 123 framework nor
silently implements the other's scope.
