# ABV-123-03 — Storage and retrieval simplification plan

**Auth-model reconciliation:** [ABV-123-07](abv-123-auth-reconciliation.md)
maps the approved rules, proposes concrete operation-bound routes instead of
universal authorization endpoints, and defines the bounded contract-design steps.
Its route/context refinements are proposals, not yet approved replacements.

**Current pin — ABV-123-04/05/06:** the user requested “pin these down”, then
clarified that completeness means **functional** completeness. The current
direction is one L1 record store and the endpoint families below. Earlier
six-table/two-store proposals are retained as superseded alternatives, not the
current target. Exact schemas, payloads, permission mapping and migration are
not approved by this architectural pin.

## Current pinned direction and functional coverage

### ABV-123-04 — five-table target

`abv_l1_records`, `abv_assignments`, `abv_teams`, `abv_memberships`, and
`abv_metadata`. The lab-only fixture marker is separate. The L1 record store
holds the record families previously mapped to catalog/tenant stores below:
applications, installations, permission/scope registration and supported-key
lists, role revisions, grant controls, grant revisions and explicit root trust.

One physical store does not merge application-wide catalog authority with
tenant/application authority. Every key, read, write and reference must establish
its correct boundary. The exact representation is an S1 deliverable; no nullable
tenant fallback or caller-selected record kind can confer authority.

Rationale: one store for typed supporting records avoids a table per concept;
indexed assignments, teams and memberships retain frequent relationship queries
and uniqueness constraints. Separate grant-control and immutable-revision
semantics survive co-location. This is a storage change, not a new grant model.

### ABV-123-05/06 — endpoint direction

Tenant-scoped base: `/api/v1/{tenant}/abv/` (`v1`, not `vi`). Application context
must also be explicitly and authoritatively established. Its exact binding and
the separate application-wide catalog-management base are still to be defined.

| Relative endpoint | Pinned purpose |
|---|---|
| GET records/{kind}/{id} | Retrieve a canonical record with declared revision selection. |
| GET records/{kind} | Supported indexed filters and bounded listing. |
| POST validate | Typed proposal validation; not authority to commit. |
| POST execute | Typed canonical operation through authoritative checks. |

No public free-form L1-record put/delete or SQL/query-language interface follows
from these paths. Request/response envelopes must be versioned and approved.
Existing typed methods remain the working contract until a replacement is proven.

**Unresolved consistency point:** the handbook requires one declared permission
per endpoint. A shared execute route must not infer a different endpoint permission
from arbitrary caller input or collapse distinct administrative powers into an
unrestricted execute permission. Settle the exact permission/operation mapping
before treating this endpoint layout as implementation-ready. If preserving the
existing rule requires operation-specific paths, propose that refinement explicitly
rather than silently changing the rule. The same check applies to heterogeneous
record reads and validation. No endpoint layout removes the separate ABV gate.

### Functional coverage assessment

| Required capability | Home under 123 | Current qualification |
|---|---|---|
| Permission/scope registration and optional support validation | Typed L1 records + protected registration operations | Existing additive implementation; catalog boundary unchanged. |
| Roles and pinned revisions | Typed L1 records + role publication | Existing implementation; publication does not update grants. |
| Grants, immutable revisions, live controls and validity | Typed L1 records + protected operations | Existing publication/control support; storage folding cannot change lifecycle. |
| Assignments and selected revisions | Assignment relationships + protected assignment operations | Creation/control implemented; explicit adoption and advisory candidates still pending. |
| Teams, memberships and dependent hierarchy | Indexed team/membership relationships + L1 lineage checks | Resolution exists; storing these facts does not imply ordinary team/membership management APIs exist. |
| Source authority, root trust, orphaning and descendant effects | Core ABV checks using complete indexed evidence | Preserve both gates and every affected branch; no generic write bypass. |
| Inspect, list and validate | Read/validate contracts over the same L1 records | Inspection/assignment diagnosis exist; generalized list/filter/payload contracts are not implemented. |
| Higher-level administrator/agent workflows | L2 maintained composition or L3 consumer composition | Must call protected primitives; no new L2/L3 persistence required by this build. |

Conclusion: the new structure provides a home for the identified in-scope
capabilities; no functional loss is intended. It does **not** establish complete
contracts or implementation. Remaining functional contracts include exact command/
record kinds, versioned inputs/results/errors, operation-specific administration,
read visibility, application-context binding and safe retry/conflict behavior.
Existing pending adoption/advisory and definition-lifecycle work remains pending;
123 does not complete it. New grant/bootstrap/team-management public operations
are not conjured by a generic execute endpoint. Previously deferred deletion,
reparenting and PostgreSQL and excluded production Auth integration stay unchanged.

### Earlier planning baseline

The earlier proposal below remains useful for its complete record inventory,
query/index map, performance controls and bounded proof. References to six tables
or two stores in sections 1–2 are superseded by the five-table pin above. During
S1 update the proof candidate to the one-store boundary contract, preserving the
separate-context guarantees. No proof or migration has run.

**Earlier status: proposal for discussion, not an approved schema or migration.**
The user requests fewer APIs/tables, effective storage/retrieval and performance.
Pause adoption implementation while this direction is settled. Preserve all
approved authorization semantics, current APIs and existing databases meanwhile.

**Goal:** reduce physical storage and integration complexity while retaining
indexed retrieval, complete authorization evidence and transactional integrity.

**Architecture:** adapt 123's canonical-record ownership, not payroll's table
names or lifecycle. Use typed record stores for supporting definitions and keep
hot relationships explicitly indexed. Core operations remain the only authority
mutation boundary; generic storage is internal, not permission to write any JSON.

**Tech stack:** existing Go, SQLite and provider seam. No dependency, production
Auth integration, PostgreSQL implementation, cache, L2/L3 store or generic query
language in this proof.

**Sources:** current `implementation/abv/internal/storage/sqlite/migrations/001_initial.sql`,
`snapshot.go`, `update.go`, `internal/storage/provider.go`, `abv.go`,
`docs/acceptance.md`; canonical `docs/grant-revisions.md`,
`docs/parent-grant-bindings.md`, `docs/auth-write-consistency.md` and
`docs/root-permission-evolution.md`. Architecture comparison:
`agentlabs-payroll-ledger-lab/docs/design/123-architecture.md` in the sibling repo.
Its newer application-owned L3 clarification supersedes older exclusions there.
These 123 layers are not the handbook's Auth/application facts split.

## 1. Evidence and recommended direction

Current schema: 13 ABV tables, including metadata; the lab adds one fixture table.
Current facade: 9 functional in-process methods, not 9 HTTP endpoints.
The area snapshot loader fetches all grant revisions, assignments, roles, teams
and memberships in the selected tenant/application, plus the application catalog.
The default limit is 10,000 loader-counted records. Even exact inspection currently
uses that full snapshot path. This is a scaling concern independent of table count.

The retained CP3 diagnosis benchmark reports about 0.72/6.05/58.61 ms and
0.23/2.61/25.28 MB allocated for 100/1,000/10,000 records. Those are historical,
three-iteration local measurements, not current latency claims or mutation SLAs.
Rerun a reproducible baseline before comparing a replacement.

Recommend a **six-table candidate**, including metadata. This is a target to
prove, not a promise that six is optimal. Compare it against indexed retrieval
on the existing schema so table folding does not receive credit for an unrelated
query improvement. Do not build a third, universal-record-store design unless
the bounded comparison demonstrates a concrete need.

## 2. Complete table disposition

Proposed physical names below are not new canonical domain records.

| Current table | Proposed home | Reason / constraint |
|---|---|---|
| applications | abv_catalog_records: application record | Application-wide ownership; compatibility setting remains authoritative. |
| permissions | abv_catalog_records: permission record | Exact registered identity; active status retains existing meaning. |
| scope_definitions | abv_catalog_records: scope-definition record | Exact key/token validation; application owns domain-value interpretation. |
| supported_scope_keys | Permission record's internal support list | Exact permission reads fetch its supported keys; preserve list order, registration references and all-or-nothing validation. This is not a new public permission JSON. |
| installations | abv_tenant_records: installation record | Tenant/app existence remains explicit before authority is usable. |
| roles | abv_tenant_records: role-revision record | Preserve current tenant/app ownership and numeric immutable revision identity. |
| grant_controls | abv_tenant_records: grant-control record | Separate mutable grant-wide control; never a revision status. |
| grant_contents | abv_tenant_records: grant-revision record | Exact immutable revision and parent reference; no recipient. |
| trusted_roots | abv_tenant_records: root-trust record | Explicit protected trust evidence; ordinary generic writes cannot create it. |
| assignments | abv_assignments | Hot recipient lookup, exact adopted revision, grant/recipient uniqueness and reverse dependency lookup. |
| teams | abv_teams | Actual team parent and child discovery, not grant ownership or inherited membership. |
| memberships | abv_memberships | Hot human-to-team and team-to-human lookups; human-only semantics retained. |
| abv_metadata | abv_metadata | Database ownership and schema version are infrastructure, not authority. |

Result: two typed record stores + three relationship tables + metadata = six.
The lab-only fixture marker stays separate and is not counted as core storage.

Why two record stores: catalog records are application-wide; authority records
are tenant/application-bound. Do not mix these through nullable tenant shortcuts
or fabricate a tenant for application administration. This consciously adapts
123 storage to two existing trust boundaries rather than copying one envelope.
L2/L3 persistence is not needed for the current ABV deliverable.

## 3. Storage contract to settle before code

Use explicit context, registered record kind, opaque ID, numeric revision where
applicable, and canonical payload. Keep query-critical identity/reference fields
indexable; deserialize only selected payloads. A record-kind discriminator is
storage metadata, not authority and not a new public grant field.

Do not mechanically add key1..key10, timestamp or generic state to every row.
ABV's real lookup patterns decide columns. Revision must sort numerically, IDs
must round-trip unchanged, and unused revision identity must have one unambiguous
internal representation. Select latest revision before applying eligibility;
never fall back to an older enabled revision. Optional record timestamps cannot
become revision, validity or ordering authority.

Checkpoint S1 must produce exact DDL, key examples and a constraint map:

- Unique record identity by actual boundary/kind/ID/revision.
- Unique grant/recipient assignment, including disabled assignments.
- Exact assignment-to-grant-revision and membership-to-team references.
- Installation/application existence and registered scope-support references.
- Immutable revision inserts; compare-and-write for mutable records.
- Indexed metadata agrees with payload, derived through one validation path.
- Root trust cannot be established by parent omission or ordinary record writes.

Prefer native constraints. Where folding removes an existing foreign key, show
the equivalent atomic enforcement and adversarial test explicitly. If preserving
it makes the shared table materially more complex, retain that dedicated table
and explain why. Table-count reduction is not a license to weaken integrity.
No public canonical JSON is silently extended to accommodate physical storage.

## 4. Retrieval and index contract

| Required operation | Indexed lookup / retrieval requirement |
|---|---|
| Exact permission/scope | Application + record kind + ID; fetch only requested definitions. |
| Exact role/grant revision | Tenant + application + kind + ID + numeric revision. |
| Latest published revision | Same identity prefix with numeric revision ordering; no status-based fallback. |
| Human's teams | Membership index: tenant, application, human, team. |
| Team's humans | Membership uniqueness/index: tenant, application, team, human. |
| Recipient's assignments | Tenant, application, recipient type, recipient ID; keyset-paginated listing. |
| Required parent holding | Unique tenant/application/grant/recipient tuple; actual parent's adopted revision. |
| Assignments of a grant | Tenant, application, grant ID; disabled records remain discoverable. |
| Child teams | Tenant, application, parent team ID. |
| Grants referencing a parent | Tenant, application, parent grant ID plus revision identity; join actual adopted assignments, not every latest revision. |
| Affected descendants | Follow BOTH team and grant dependencies; include disabled bridges and every relevant branch. |

An identity primary key is not automatically sufficient for reverse queries.
Measure the listed secondary indexes; count their space and write amplification.
No blanket JSON index or scan/filter of every payload to find a single record.
Avoid an arbitrary query language; supported query shapes and cursors are bounded.

Read only the **complete evidence required by the operation**: actual supporting
assignments, adopted grant/role revisions, controls, catalog definitions, teams,
memberships and root evidence. Batch available frontiers to avoid one SQL call
per graph node; no N+1 promise until measured. Changed lineage requires complete
affected-branch discovery, not merely checking the selected route.

All evidence for a decision must share a coherent transaction view. Current
administration callbacks receive a full snapshot and may inspect other records;
do not silently hand them partial maps. S1 must specify how their required evidence
is supplied. Preserve the full-snapshot path until each operation's dependency
set is explicit and tested; do not guess missing administrative requirements.

Pagination is for listings, never incomplete authorization. If required evidence
exceeds bounded records/depth/work/time, return an explicit failure, not a partial
allow. Unrelated records should not consume a targeted lookup's evidence budget.
Any change from whole-area corruption checking to operation-scoped validation
must be stated and tested; it is not automatically behavior-identical.

## 5. API simplification without a universal write bypass

Propose three integration families: **read, validate, execute**. These are a
conceptual surface, not approval of a new generic request schema or exactly three
Go methods. Existing public methods remain compatible wrappers during a proof.

| Existing method | Family | Rule retained |
|---|---|---|
| Inspect | Read | Exact kind/identity, bounded visibility; list semantics declared separately. |
| CheckAssignment | Validate | Diagnostic only; not authorization to commit. |
| RegisterPermission | Execute: register permission | Separate application-publisher gate and registration contract. |
| RegisterScope | Execute: register scope | Same boundary, distinct scope schema and checks. |
| PublishRole | Execute: publish role | Immutable pinned role revisions; no automatic grant change. |
| PublishGrantRevision | Execute: publish grant revision | Administrative and source-boundary checks, immutable insert. |
| CreateAssignment | Execute: assign | Both gates, latest selection, actual lineage and unique recipient holding. |
| SetGrantStatus | Execute: grant status | Grant-wide control, preserved revisions and affected-route validation. |
| SetAssignmentStatus | Execute: assignment status | Separate recipient control, bottom-up binding rules and current re-enable checks. |

Use a closed typed command set and existing handlers if one Execute entry point
actually simplifies CLI/provider integration. Do not substitute maps of arbitrary
fields or untyped callbacks. Count operations honestly: nine semantics do not
become three semantics by placing them behind a dispatcher. Keep typed functions
if a generic envelope introduces more branching, boilerplate or worse diagnostics.
Core state cannot be changed by free-form put/delete. Exact public envelopes and
versions require separate approval before replacement; no HTTP layer is required.

## 6. Bounded proof checkpoints — after design approval

The present task delivers this plan only. The following are execution budgets,
not delivery estimates. Each checkpoint stops on its exit condition or time cap;
at most one 10-minute targeted correction, then report and regroup. Coding uses
Sol-medium; no review passes as directed. Preserve runnable tests and evidence.

| Checkpoint | Cap | Deliverable / exit condition |
|---|---:|---|
| S1: contract + query map | 20 min | Exact candidate DDL, keys, constraints, callback evidence needs and API compatibility map; unresolved safety choices surfaced before coding. |
| S2: baseline | 25 min | Reproducible current-schema data generator and workload measurements; environments/seeds/settings recorded, correctness assertions precede timings. |
| S3: SQLite proof | 45 min, split into <=20-min coding tasks | Targeted-read baseline on current schema, then six-table candidate using the same requests/fixtures; no live-data migration, no dual-write. |
| S4: performance + integrity comparison | 30 min | Query plans, end-to-end latency/memory/write metrics, unchanged approved outcomes and no-write rejection proof. |
| S5: decision | 15 min | Retain or reject each consolidation; final table/API count and trade-offs; separate reversible migration plan only if justified. |

If S1 cannot establish complete operation evidence without a broad new framework,
restrict the first proof to exact retrieval and one existing protected operation.
Do not hide that limitation or claim full targeted-lineage support.

### Fixed benchmark matrix

Compare three configurations, changing one axis at a time:
1. Current schema and current snapshot loader.
2. Current schema and indexed operation-specific retrieval.
3. Consolidated candidate and equivalent indexed retrieval.

Use 100 / 1,000 / 10,000 loader-equivalent records for directly comparable cases.
Add a 100,000-record unrelated-background case to test targeted scaling; the
default old provider must report its limit rather than be timed as successful.
Any benchmark-only increase of that limit is labeled separately, not deployed.
Keep a fixed selected route while growing unrelated records; separately vary
depth 1/8/32, branching, revisions per grant 1/10/100, and human memberships.
Use representative small/large payloads with identical bytes across providers.

Workloads: exact inspection; human membership + assignment retrieval; complete
assignment diagnosis; protected assignment creation; grant publication; status
change with branching dependencies. Measure successful and rejected operations.
Pair point reads with a writer, and test competing duplicate publications/writes.
Use fresh fixtures for mutations so duplicate failures are not mistaken for
successful write throughput. Report affected row counts and outcomes.

Run on the same machine, driver, indexes, transaction semantics, SQLite pragmas
and durability settings. Warm and reopened-connection runs are separate; reopening
does not prove cold OS cache. Exclude fixture construction from timed operations.
Use five repeats with a fixed workload; report median and variation, plus p50/p95
from at least 200 measured operations per workload where the cap permits. Mark
insufficient samples explicitly. No repeated tuning loop to achieve a threshold.

Record latency, throughput, allocations/bytes, rows and payload bytes loaded,
SQL statements, transaction/lock duration, contention errors, DB/index size and
write cost. Capture EXPLAIN QUERY PLAN for each hot lookup and reverse query.
No hidden in-process cache in just one configuration. Existing SQLite write
serialization remains visible; table folding is not a concurrency solution.

### Proposed acceptance gates, not measured claims

- All approved authority outcomes and isolation/rollback/immutability/uniqueness
  tests pass. Include orphaned support, disabled bridges, every affected branch,
  hostile callback changes, parent scope AND, stale revisions and cancellation.
- Exact reads avoid whole-area payload loading. Fixed-route evidence volume does
  not grow with unrelated rows; legitimate affected-branch work may grow.
- Candidate p95 on named hot reads and representative writes is no worse than
  the indexed current-schema alternative by more than 10%, beyond observed noise.
  Report write/index overhead even when the read target passes.
- Targeted retrieval at 10,000 records aims for at least 2x lower diagnosis time
  and 50% fewer allocated bytes than the full-snapshot baseline. These are proof
  goals to accept/reject, not promised production performance or invented SLAs.
- A missed performance gate stops migration and produces a concrete trade-off;
  do not silently weaken gates or add caches/distributed infrastructure.

## 7. Later migration gate — not authorized by this plan

Keep the current database and APIs working. If the proof wins, propose a new
schema version and copy into a separate destination; never overwrite the source
or silently reinterpret version 1. Validate canonical content, numeric revision
order, controls, assignments, root trust, memberships, references and complete
lineage outcomes before switching. Preserve source for rollback and inventory
the lab marker separately. No dual writes in the first migration.

Exact quiescence/cutover/rollback handling requires its own approved plan. Do not
promise rollback after new-schema writes without reverse transfer or an explicit
write-free cutover window. PostgreSQL portability remains a design check, not
implementation or measured cross-provider performance.

## Completion definition

The planning task is complete when this proposed mapping, retrieval design,
measurement protocol and migration gate are available for discussion. The proof
is complete only with comparable evidence and an explicit keep/reject decision.
Reducing table/API counts alone is not completion. No existing decision, table,
record, method or history is deleted by this proposal.
