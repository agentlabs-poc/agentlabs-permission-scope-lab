# ABV, reusable CLI and SQLite Provider Implementation Plan

## Current execution note — 8 September 2026

The user approved execution, starting with CP1, and authorized verified commits
and pushes. Coding now uses `gpt-5.6-sol` subagents with medium reasoning at the
user's request. Earlier planning-only/publication gates below describe the plan
before that approval; preserve them as history, not current blockers.
See [checkpoint progress](progress.md) for delivered work and evidence.

Internal-interface refinement: `Route` carries an `Area`; `CheckContent` and
`Narrow` take an explicit `Area` and reject mismatched context. This implements
the mandatory outer-boundary requirement; it adds no tenant/application fields
to canonical grant JSON. Syntax-only decoding does not resolve authority and
cannot authorize a write.

**CP1 review correction:** `Narrow` receives the area's role-revision records,
not a free-standing caller-supplied permission expansion. It derives the entire
selected permission list from direct content or the exact role revision through
the same helper as definition validation. This prevents substituting another
parent permission or silently trimming the selected role bundle. The earlier
loose expansion parameter is superseded; canonical grant JSON is unchanged.

**CP3 interface refinement:** `ResolveParentTeam` receives the exact selected
child `GrantContent`, not a bare child-grant ID with implicit revision selection.
This binds parent discovery/cycle checks to the proposal. The coordinator checks
latest-only creation; traversal follows actual parent-team adoptions. The older
two-string sketch is superseded only at this internal seam, not in public JSON.

**Task 6 composition refinement:** the CLI's parsed `--db` must select the actual
store. Replace its prebound `application.API` argument with a small injected
connector function, returning the API and its close function for the validated
area/path. Keep the three API operations unchanged. This supersedes only the
earlier `Run` signature sketch below; no duplicate parsing in `main` or ignored
database flag is permitted. A function is sufficient; no factory hierarchy is
needed. Help, malformed commands and scenarios do not open the ordinary API.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> or superpowers:executing-plans to implement this plan task-by-task. The default
> handoff is inline execution; do not infer approval to spawn agents, commit,
> push, create a remote repository or modify the existing Auth service.

**Goal:** Build a reusable in-process Go ABV library and reusable CLI testing
interface, with SQLite persistence behind a transaction-aware provider boundary.

**Architecture:** CLI commands call an application interface implemented by the
ABV facade. A mutation coordinator runs administrative authorization and ABV on
consistent provider evidence, then persists only the exact validated change.
The validator contains no SQL and the CLI contains no authorization rules.

**Tech Stack:** Go module compatibility 1.25.0; standard-library CLI/JSON/testing;
`database/sql`; `modernc.org/sqlite` v1.58.0. PostgreSQL is a later provider,
not a dependency of the first slice.

**Spec:** [ABV design and rationale](abv-design.md). Read it before execution,
especially the first-slice limits, transaction guarantees and source contracts.

## Global constraints

- Tenant/application is the mandatory outer boundary for this component now.
- Every read/write/validation requires both context IDs; inner scope cannot widen them.
- Both CLI and ABV are reusable; initial CLI calls are in-process, with no HTTP/RPC server.
- Grants remain recipient-free; assignments remain separate.
- Required revision fields and approved semantics remain, despite their omission from the slide lesson.
- Parent constraints accumulate with AND; no map overwrite, union of unrelated sources or silent permission trimming.
- Administrative authorization and ABV are distinct mandatory checks for protected writes.
- No application business database, rule engine, network authority cache or audit-retention subsystem.
- Unsupported cases fail explicitly without writes; that is not a new canonical prohibition.
- No ordinary parent-omission bootstrap, always-allow evaluator or skip-validation flag.
- No existing Auth-service changes, external schema promotion, commit or push under planning approval.
- Use file-backed temporary databases and at least two independent connections for concurrency tests.

## Scope and ordering

This is a detailed execution plan for **CP1–CP3**, producing a usable local
testing slice. CP4–CP6 are bounded follow-on milestones with entry requirements;
they must receive their own exact operation/interface plans before execution.
This distinction avoids claiming a complete Auth implementation while public
contracts and source-binding decisions are still open.

```text
T1 context + types + CLI seam
    → T2 core decoding / pure checks
    → T3 SQLite provider + persistence conformance
    → T4 actual parent-team resolution
    → T5 transactional assignment mutation
    → T6 reusable CLI commands + controlled lab scenarios
    → T7 adversarial / race / end-to-end acceptance
    → CP4 lifecycle → CP5 real Auth / catalog management → CP6 PostgreSQL
```

The first visible demo is not an HTTP server. It is:
seed a disposable scenario → inspect records → check A2 → assign A2 through
both checks → reopen the database and inspect the saved result.

## Source layout

All planned code paths are relative to `implementation/abv/`, which does not exist
yet. Do not create it while only reviewing this plan.

```text
go.mod / go.sum                      isolated module and pinned dependencies
abv.go                              public construction and library facade
domain/area.go                      mandatory outer context
domain/records.go                    approved record types; internal projections
domain/errors.go                     internal typed failures, not public API codes
application/api.go                  interface used by CLI and future adapters
cli/run.go                          reusable command runner; injected streams
cli/inspect.go / check.go / assign.go command-specific parsing and presentation
cmd/abv/main.go                      dependency wiring and process exit only
internal/codec/                      strict approved core-record decoding
internal/validation/                 pure definition and subset/AND checks
internal/lineage/                    eligible parent-team routes / source access
internal/mutation/                   both gates and consistent persistence
internal/storage/provider.go         snapshot / write-set / transaction contract
internal/storage/sqlite/             connection, migrations, projections, queries
internal/storage/contracttest/       same behavioral tests for each provider
internal/lab/                        explicit disposable scenario fixtures
testdata/                           approved JSON and adversarial input files
docs/local-testing.md                CLI walkthrough and limits
```

No `cli` imports of `internal/storage/sqlite` or `internal/validation`. The
binary's composition root wires them. No imports from CLI into the public ABV
library. Use no generic ORM or runtime SQL strings in the validation package.

## Proposed internal interfaces

These signatures fix implementation seams; they are **not new canonical wire
contracts**. Types for pending representations stay internal or labeled diagnostic.
Field names in canonical JSON remain those in the handbook, not automatically
the names of every internal Go structure.

```go
// domain/area.go
type Area struct { tenantID, applicationID string }
func NewArea(tenantID, applicationID string) (Area, error)
func (a Area) TenantID() string
func (a Area) ApplicationID() string
func (a Area) Validate() error

// domain/records.go — normal Go structs with canonical JSON tags where approved.
type Identity struct { Version string; Actor Actor; HumanID string }
type Actor struct { Type, ID string }
type Recipient struct { Type, ID string }
type GrantControl struct { Version, ID, Status string }
type Validity struct { NotBefore, ExpiresAt *time.Time }
type GrantContent struct {
    Version, GrantID string
    Revision int64
    ParentGrantID string
    Permissions []string
    RoleID string
    RoleRevision int64
    Scope map[string]string
    Validity *Validity
}
type Assignment struct {
    Version, ID, GrantID string
    GrantRevision int64
    Recipient Recipient
    Status string
}
// Internal projections: not approved standalone JSON APIs.
type RoleContent struct { ID string; Revision int64; Permissions []string }
type Team struct { ID, ParentID string }
type Membership struct { TeamID, HumanID string }
type PermissionDefinition struct { ID string; Active bool }
type ScopeDefinition struct { Key string; AllowedTokens []string }
type Catalog struct {
    ApplicationID string
    Permissions map[string]PermissionDefinition
    Scopes map[string]ScopeDefinition
    CompatibilityEnabled bool
    SupportedKeys map[string][]string // permission ID -> registered scope keys
}
type GrantKey struct { ID string; Revision int64 }
type RoleKey struct { ID string; Revision int64 }
type Predicate struct { Key, Value, SourceGrantID string }
type Route struct {
    Area Area
    GrantID string
    Permissions []string
    Predicates []Predicate
    AssignmentIDs []string
    Validities []Validity
}
type Diagnostic struct { Summary string; Route *Route }
type Record struct { Kind, ID string; CanonicalJSON []byte; Rows [][]string }
type Receipt struct { AssignmentID string }
type FixtureContext struct { Name string } // lab selection, never authenticated identity
```

`ParentGrantID == ""` is representable so decoding can distinguish a root-shaped
record, but ordinary mutation must reject it without legitimate trusted setup.
Integers above express selected internal storage types, not a settled public
maximum or default. `Scope == nil` is invalid; `{}` is a non-nil empty map.
Do not serialize a `Diagnostic`, `Route`, `Catalog`, `Team`, `RoleContent` or
`Receipt` as a newly canonical JSON contract.

```go
// internal/storage/provider.go
type Snapshot struct {
    Area domain.Area
    Catalog domain.Catalog
    Controls map[string]domain.GrantControl
    Contents map[domain.GrantKey]domain.GrantContent
    Assignments map[string]domain.Assignment
    Roles map[domain.RoleKey]domain.RoleContent
    Teams map[string]domain.Team
    Memberships []domain.Membership
    TrustedRoots map[string]bool // internal established-root projection, not a grant field
}
type WriteSet struct { NewAssignments []domain.Assignment }
type Provider interface {
    Read(context.Context, domain.Area, func(Snapshot) error) error
    Update(context.Context, domain.Area, func(Snapshot) (WriteSet, error)) error
    Close() error
}
```

The initial write set deliberately supports only assignment creation. Do not
advertise arbitrary grant/catalog CRUD before its mutation checks exist. All
record kinds can be stored/retrieved through the controlled fixture loader for
provider conformance. CP4/CP5 add typed write commands with corresponding checks.
`TrustedRoots` is populated only by controlled test setup in this slice, never
derived from parent omission or an ordinary JSON input.

```go
// internal/validation, internal/lineage
func CheckContent(domain.Area, domain.Catalog, domain.GrantContent,
    map[domain.RoleKey]domain.RoleContent) error
func Narrow(domain.Area, domain.Route, domain.GrantContent,
    map[domain.RoleKey]domain.RoleContent) (domain.Route, error)
func ResolveParentTeam(storage.Snapshot, domain.GrantContent, string, time.Time) (domain.Route, error)
func HasSource(storage.Snapshot, domain.Identity, domain.Route, time.Time) error

// internal/mutation — administrative adapter must use state bound to this write.
type Administration interface {
    CheckAssignment(context.Context, storage.Snapshot,
        domain.Identity, domain.Assignment, time.Time) error
}
type Clock interface { Now() time.Time }
type Service struct { /* private provider, administration and clock dependencies */ }
func New(storage.Provider, Administration, Clock) (*Service, error)
func (s *Service) CreateAssignment(context.Context, domain.Area,
    domain.Identity, domain.Assignment) (domain.Receipt, error)

// application/api.go — shared seam for reusable CLI and future adapters.
type API interface {
    Inspect(context.Context, domain.Area, string, string) (domain.Record, error)
    CheckAssignment(context.Context, domain.Area, []byte) (domain.Diagnostic, error)
    Assign(context.Context, domain.Area, domain.FixtureContext, []byte) (domain.Receipt, error)
}
type ScenarioRunner interface {
    Seed(context.Context, domain.Area, string, string) error // name, new database path
    Run(context.Context, domain.Area, string, string, string) error // name, case, new path
}
// cli/run.go — no os.Exit, globals, implicit tenant, SQL or rule logic.
func Run(context.Context, []string, io.Reader, io.Writer, io.Writer,
    application.API, application.ScenarioRunner) int
```

`FixtureContext` belongs to the local application adapter, not ABV's trusted
identity API. A later authenticated CLI adapter replaces fixture attribution;
ABV and command parsing remain reusable. Production Auth integration must not
map arbitrary CLI names to trusted humans.

Internal error categories are malformed, rejected, unsupported, not-found,
conflict, unavailable and cancelled. Use typed errors and `errors.Is`/`errors.As`;
preserve causes internally without leaking raw database details to CLI output.

## Task 1 — mandatory context, records and reusable CLI seam [CP1]

**Create:** `go.mod`, `domain/area.go`, `domain/records.go`, `domain/errors.go`,
`application/api.go`, `cli/run.go`, `domain/area_test.go`, `cli/run_test.go`.

**Consumes:** approved records and the exact proposed interfaces above.
**Produces:** context validation, typed record/error definitions and a testable
CLI entry function. No database or generic mutation command yet.

- [x] Create the isolated module only after design review and implementation
  authorization. Use `module agentlabs.local/abv`, `go 1.25.0`; no parent workspace
  `go.work` or website package change.
- [x] Write this failing boundary test before implementing the constructor:

```go
func TestAreaRequiresBothBoundaries(t *testing.T) {
    for _, ids := range [][2]string{{"", "hrms"}, {"acme", ""}, {"", ""}} {
        if _, err := NewArea(ids[0], ids[1]); err == nil {
            t.Fatalf("accepted missing boundary: %q", ids)
        }
    }
    if err := (Area{}).Validate(); err == nil { t.Fatal("accepted zero Area") }
    a, err := NewArea("acme", "hrms")
    if err != nil || a.TenantID() != "acme" || a.ApplicationID() != "hrms" {
        t.Fatalf("lost exact context: %#v %v", a, err)
    }
}
```

- [x] Run `go test ./domain -run TestAreaRequiresBothBoundaries -count=1`; confirm
  the missing implementation fails, then implement constructor/accessors/validation.
  Reject empty/all-whitespace context and wildcard `*`; do not silently trim IDs.
- [x] Add table tests for exact same tenant/different app and same app/different
  tenant equality distinctions. Implement canonical JSON tags using the source
  blocks, not Go default field names.
- [x] Add CLI tests: missing `--tenant`, missing `--app`, unknown command and
  malformed flags return 2 without calling a spy API. Implement only help,
  boundary parsing and dispatch seams; unsupported commands return 5, not success.
- [x] Run `go test ./domain ./cli -count=1` and `go vet ./...`.
- [x] Review diff and record CP1 progress; do not commit without authorization.

## Task 2 — exact JSON and pure authority checks [CP1]

**Create:** `internal/codec/assignment.go`, `content.go`, `duplicates.go`,
`codec_test.go`; `internal/validation/content.go`, `narrow.go`, `validation_test.go`;
`testdata/g1.json`, `g2.json`, `a1.json`, `a2.json`.

**Consumes:** domain types. **Produces:**
`codec.DecodeAssignment([]byte) (domain.Assignment, error)`,
`codec.DecodeContent([]byte) (domain.GrantContent, error)`, `CheckContent`, `Narrow`.

- [x] Copy G1/G2/A1/A2 from the Foundations reference as exact versioned fixtures,
  preserving revision fields. Include valid upstream premises in test setup.
- [x] Add failing tests for duplicate JSON keys, missing version, null/missing
  scope, scope arrays, empty values, recipient on content, missing role half,
  mixed role/direct fields and trailing JSON values. Decode token-by-token to
  detect duplicates before ordinary unmarshalling; never repair malformed input.
- [x] Run `go test ./internal/codec -count=1`; implement strict core decoding.
  Extra unresolved extensions return unsupported/malformed without being discarded;
  document this as the prototype's supported input contract, not a newly complete
  canonical schema. Version must be the supported string `"1"`.
- [x] Add the non-amplification tests before implementing `Narrow`:

```go
func TestNarrowPreservesConflictingPredicates(t *testing.T) {
    area, err := domain.NewArea("acme", "hrms")
    if err != nil { t.Fatal(err) }
    parent := domain.Route{Area: area, GrantID: "G1", Permissions: []string{"read", "write"},
        Predicates: []domain.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "G1"}}}
    child := domain.GrantContent{Version: "1", GrantID: "G2", Revision: 1,
        ParentGrantID: "G1", Permissions: []string{"read"}, Scope: map[string]string{"dept": "ENG"}}
    got, err := Narrow(area, parent, child, nil)
    if err != nil { t.Fatal(err) }
    if len(got.Predicates) != 2 || got.Predicates[0].Value != "FIN" {
        t.Fatalf("lost parent restriction: %#v", got)
    }
    child.Permissions = []string{"delete"}
    if _, err := Narrow(area, parent, child, nil); err == nil {
        t.Fatal("accepted permission expansion")
    }
}
```

`read`/`write`/`delete` here are internal algebra test atoms, not registered public
permission identifiers. Registration tests use the full handbook strings.

- [x] Implement permission subset by set membership; build a fresh predicate
  slice from parent predicates plus sorted child keys. Preserve provenance and
  all parent validities. Do not use last-write-wins maps for effective scope.
- [x] Test unknown permission/key/token, role expansion, explicitly disabled
  compatibility mode versus enabled unsupported combinations, and supported empty
  child scope. Publication of contradictory scopes is not decided by `Narrow`.
- [x] Run `go test ./internal/codec ./internal/validation -count=1`; review
  unsupported-case classifications and checkpoint the evidence without committing.

## Task 3 — SQLite provider and reusable conformance tests [CP2]

**Create:** `internal/storage/provider.go`, `errors.go`;
`internal/storage/sqlite/open.go`, `migrate.go`, `snapshot.go`, `update.go`,
`migrations/001_initial.sql`; `internal/storage/contracttest/suite.go`;
`internal/storage/sqlite/provider_test.go`; `internal/lab/fixture.go`.

**Consumes:** context/types; provider interface above.
**Produces:** `sqlite.Open(ctx, filePath) (storage.Provider, error)` and
`contracttest.Run(t, factory)` where `factory` opens providers against a supplied
temporary file. Provider callbacks are invoked exactly once and never auto-retried.

- [x] Add dependency with `go get modernc.org/sqlite@v1.58.0`; retain checksums
  and matching transitive versions. Do not upgrade unrelated modules.
- [x] Write provider tests for rollback, close/reopen persistence, consistent
  snapshot reads, missing-area rejection, per-connection foreign-key enforcement
  and both dimensions of isolation. Expect failure before tables/provider exist.
- [x] Create provider migrations for the logical tables in the design. Use
  composite tenant/application keys for every ordinary authority relation and
  explicit installation membership before shared catalog lookup. Add the
  grant/recipient uniqueness constraint including disabled assignments.
  No destructive cascade, public revision default, or unqualified ID lookup.
- [x] Implement pinned-connection transactions in this order:

```text
validate Area → acquire connection → BEGIN IMMEDIATE
  → establish installation → read shared catalog + bounded authority snapshot
  → invoke callback once → check write-set partition and uniqueness
  → write all changes → COMMIT
on any error/panic/cancellation → ROLLBACK → release connection
```

- [x] Test a write set of two assignments where the second violates uniqueness:
  neither may appear afterward. Repeat after closing/reopening the file.
  Mutating the callback's snapshot must not mutate persisted records unless an
  explicit permitted write set is committed.
- [x] For each kind, round-trip full values (including roles, scope, validity,
  Unicode IDs and disabled assignments) with controlled fixtures. Catalogs must
  not become per-tenant copies. Deliberately reuse G1/Team1/A1 IDs in a different
  tenant and a different application; prove isolation for reads and writes.
- [x] Open the same file using two provider instances. Hold one mutation with
  channels, issue a competing update and verify consistent ordering or explicit
  lock/conflict failure—never partial visibility or callback replay. Use context
  deadlines, not sleep-dependent timing.
- [x] Run `go test ./internal/storage/... -count=1` and
  `go test -race ./internal/storage/... -count=1`. Fix connection leaks and
  rollback failures before proceeding. Record SQLite-only coverage honestly.

Task 3 completion: independently approved at `8ac7593` after review corrections.
See [CP2 evidence](../abv/docs/sqlite-provider.md), including uncertainty-safe
transaction start/migration, corruption/error handling and cross-query read tests.

## Task 4 — deterministic parent-team lineage [CP3]

**Create:** `internal/lineage/resolve.go`, `source.go`, `resolve_test.go`;
extend `internal/lab/fixture.go` and add `internal/lab/cases.go`.

**Consumes:** snapshot and pure validation. **Produces:** `ResolveParentTeam`
and `HasSource` with no provider or CLI dependency.

- [x] Seed the exact baseline internally: G0 legitimate fixture root; Team1
  holding G1 through A1; Team2 child of Team1; G2 parent G1; Maya in Team1;
  Nutan in Team2. A2 is proposed and absent initially. Role catalog, permission
  registrations, controls and explicit administrative test premises are present.
- [x] Write a baseline test resolving G2 support for Team2: find Team1, then
  Team1's actual G1 assignment, then its selected content and G0 support.
  Assert effective permissions include read/write at the parent and Finance
  predicates remain. Create no permanent dependency on issuer Maya.
- [x] Write table cases before implementation:

| Case | Expected result |
|---|---|
| G1 only at TeamX | Rejected missing eligible support; no substitution. |
| TeamX holds broader content | Actual Team1-held content still determines the ceiling. |
| G1 disabled, A1 enabled | Required support unusable. |
| Team1 assignment disabled | Required support unusable. |
| Team2 parent relationship missing | Cannot infer it from names or scope. |
| Parent graph cyclic, including disabled edges relevant to the proposal | Reject cycle; do not recurse indefinitely. |
| Same IDs in another tenant/application | Cannot supply support. |
| Maya lacks source membership | `HasSource` rejects; Nutan's membership does not substitute. |
| Direct-human differing-support case | Explicit unsupported gap, not guessed authority. |
| Self binding would change across recipients | Explicit unsupported gap, not literal-equality proof. |

- [x] Implement DFS with a recursion-stack cycle check and bounded traversal;
  resolve actual adopted support top-down. Preserve complete route associations.
  Do not union historical versions into a synthetic graph or inherit membership.
- [x] Test expiry boundary using an injected timestamp; exact expiry is
  ineligible, inclusive start is eligible. Copy all inherited validity limits.
- [x] Run `go test ./internal/lineage ./internal/validation -count=1`; checkpoint
  the passing cases and record unsupported cases separately from canonical denials.

Task 4 completion: independently approved at `4550860`; see
[lineage evidence](../abv/docs/lineage-resolution.md), including the source-binding
uniqueness correction and genuine exact-expiry revalidation.

## Task 5 — both checks and transaction-safe assignment creation [CP3]

**Create:** `internal/mutation/service.go`, `assignment.go`, `service_test.go`,
`consistency_test.go`; `abv.go`; `internal/lab/administration.go`.

**Consumes:** provider, decoder, lineage, validation, injected clock and
administrative adapter. **Produces:** `CreateAssignment`, facade retrieval and
diagnostic methods. Public callers do not receive provider raw-write methods.

- [x] Write failing tests asserting no rows are written when administration
  rejects, the source is missing, the proposal is malformed, the grant is too
  broad, the recipient is wrong or validation cannot finish. The lab evaluator
  checks its explicit fixture permission/recipient boundary; never use a general
  `return nil` evaluator in the production-intended composition.
- [x] Implement the operation sequence:

```text
CreateAssignment(ctx, area, identity, proposed)
  validate context and typed proposal
  provider.Update(ctx, area, callback(snapshot))
    require snapshot.Area == area
    validate identity and operation's supported shape
    administration.CheckAssignment(ctx, snapshot, identity, proposed, now)
    require selected grant content exists and is latest for this creation
    CheckContent(area, snapshot.Catalog, selectedContent, snapshot.Roles)
    establish proposed recipient team, parent route and assigner source
    Narrow(area, parentRoute, selectedContent, snapshot.Roles)
    check complete recipient/team boundary and relevant live controls
    reject existing current grant/recipient pair even if disabled
    recheck time eligibility before returning exact NewAssignments write set
  only return Receipt after provider confirms commit
```

Snapshot mutation is not proof that a proposal passed validation. The coordinator
builds the write set itself and never accepts a caller's `validated` flag.

- [x] Add tests that a successful diagnostic check followed by parent disablement
  does not permit a later assign. A new assign must re-read and recheck. A failed
  commit produces no success receipt, and a failed attempt does not silently
  select another content revision.
- [x] Use two SQLite handles and channel barriers to test a competing parent
  disablement before mutation acquisition versus after commit. Establish and
  assert ordering; do not call a legitimate later update a retroactive failure.
  Verify busy/conflict/cancellation paths never replay the callback automatically.
- [x] Advance a fake clock across the content's expiry during validation and
  assert no assignment write is issued. Record the precise time-check point;
  a production commit-time expiry contract still needs explicit review.
- [x] Run `go test ./internal/mutation -count=1`, then `go test -race ./...`.
  CP3's guarantee covers SQLite and the controlled administrative port, not
  unreviewed external Auth evidence or a complete production evaluator.

Task 5 completion: independently approved at `d7fea85`, including the
deterministic-ordering test correction. See [protected assignment evidence](../abv/docs/assignment-creation.md).

## Task 6 — reusable CLI with in-process adapter [CP3]

**Create:** `cli/inspect.go`, `check.go`, `assign.go`, `scenario.go`, corresponding
`*_test.go`; `cmd/abv/main.go`; `internal/lab/application.go`, `scenarios.go`;
`docs/local-testing.md`.

**Consumes:** `application.API` and `ScenarioRunner`. **Produces:** the exact
commands listed in design section 8A, backed by an in-process facade.

Internal binding seam (supersedes the prebound API argument in the earlier sketch):

```go
// application/api.go
type Connect func(context.Context, domain.Area, string) (API, func() error, error)

// cli/run.go
func Run(context.Context, []string, io.Reader, io.Writer, io.Writer,
    application.Connect, application.ScenarioRunner) int
```

Validate command/context before connecting, forward the exact path and area,
close a successful connection exactly once, and report close/output failures.
The lab connector implements this seam; neither CLI parsing nor the reusable
facade imports lab identity assumptions.

The generic CP2 database marker is not a lab-scenario marker. Store a distinct
internal scenario marker in the new lab database, bound to its scenario name
and exact tenant/application. Only exclusive scenario creation may establish it;
ordinary opening must never add or repair it. Missing/mismatched/unsupported
markers cannot enable fixture-identity assignment. Use an optional lab-owned
metadata table, not a changed core schema-v1 contract or new canonical field.
Create it only after successful new-file fixture loading; a failure leaves an
unusable-for-lab-assignment file and reports failure without deletion. Verify the
marker on the lab assignment path. This guards accidental use of a generic ABV
database; it does not resist a hostile database/filesystem owner. All authority
reads and assignment writes still use the facade/provider and both checks.

- [ ] Write CLI tests before parsing: capture stdout/stderr in `bytes.Buffer`,
  inject an API spy and call `Run`. Assert exact tenant/app forwarding and that
  absent context causes zero API calls. No test must rely on process globals.
- [ ] Implement inspect dispatch for permission, scope, role, grant, assignment,
  team and membership. Parameterize all lookups by Area. Print approved core
  JSON only where it exists; print internal relationship projections as tables.
- [ ] Implement `check assignment` as read-only diagnosis, prominently saying
  it does not authorize a later write. Implement `assign` through the coordinator,
  never by calling provider SQL or treating a prior check as a permit.
- [ ] Implement fixed lab scenario initialization using exclusive file creation.
  Refuse existing files and non-lab databases without deletion, reset or overwrite.
  A missing/wrong `fixture-context` cannot mutate. Display test-identity limitations
  clearly; never accept an arbitrary `--actor` as authenticated authority.
- [ ] Map failures to the proposed CLI statuses: 2 malformed, 3 rejected,
  4 unavailable/conflict/cancelled, 5 unsupported; successful action is 0.
  A scenario runner may return 0 when an expected rejection is correctly observed,
  but must print the observed rejection rather than claiming the mutation succeeded.
- [ ] Test file-not-found, duplicate JSON key, missing context, invalid scenario,
  existing seed database, cross-application lookup and pipe-safe record output.
- [ ] Compile with `go build -o ./bin/abv ./cmd/abv`. In tests, use `t.TempDir()`
  paths and `exec.CommandContext` to exercise the binary's seed → inspect → check
  → assign → reopen → inspect workflow and the no-write negative cases.
- [ ] Run `go test ./cli ./internal/lab ./cmd/abv -count=1`. Document the commands,
  supported prototype boundaries and eventual authenticated adapter requirement.

## Task 7 — security and portability acceptance [CP1–CP3 exit]

**Create:** `acceptance_test.go`, `docs/acceptance.md`,
`internal/storage/contracttest/concurrency.go`; add tests in the responsible packages.

**Consumes:** all first-slice components. **Produces:** reproducible evidence,
not a claim that the entire handbook or PostgreSQL is implemented.

- [ ] Map source cases C01-T01–T28 to tests, unsupported implementation limits
  or later checkpoints. C01-T27 is application actual-data enforcement, not an
  ABV success claim. C01-T07 publication policy and C01-T28 source binding remain
  visibly unresolved where applicable. No blanket “28/28 pass” from fixtures.
- [ ] Make same-tenant/different-app and different-tenant/same-app cases mandatory
  for every read/write/resolve test family. Verify catalog access requires a valid
  installation in the explicit Area, with no scope/ID override or fallback.
- [ ] Run `go test ./... -count=1`, `go test -race ./... -count=1`, `go vet ./...`
  and `go build ./cmd/abv`. Run provider conformance with two independent handles
  and restart the database between persistence assertions.
- [ ] Add benchmarks for 100, 1,000 and 10,000 authority records in one bounded
  graph; report time and allocation counts without inventing a latency SLA.
  Above the configured snapshot limit must return an explicit error, never a
  partial route or a successful result based on a truncated graph.
- [ ] Check package imports: CLI only depends on the application interface/domain;
  validation/lineage import no SQL or CLI; the provider imports no CLI; facade
  imports no presentation layer. Changing a CLI flag must not change validator tests.
- [ ] Record exact command outcomes and remaining gaps in `docs/acceptance.md`.
  Review repository diff; no Auth server files or handbook decisions change.
  Stop at the checkpoint for review; commit/push only on explicit authorization.

## Follow-on checkpoints — bounded work, not silently omitted

### CP4 — lifecycle and structural writes

Add typed operations for grant/assignment enable-disable, authorized deletion,
parent changes, publication and explicit upgrades. Entry: CP3 passing and exact
internal operation payloads reviewed. Extend provider writes without public raw CRUD.

Required cases: G2 disable makes enabled G3 ineffective; re-enable restores only
otherwise valid children; explicitly disabled G3 stays disabled; binding changes
require bottom-up removal/disablement across every shared branch; cycles are
rejected even when disabled; orphaned routes cannot authorize; invalid re-enable
does not partially enable a shared grant. Latest-only creation/upgrades and
ordinary re-enable preserving adoption remain mandatory.

**Exit:** mutation and provider tests establish these outcomes, including races,
with new CLI commands calling the same coordinator. No auto-repair, cascade
deletion or issuer-ownership authority import is introduced.

### CP5 — definition management and real Auth integration

Add permission/scope/role operations with reviewed exact contracts and dependency
effects. Required cases: registration before use; role expansion stays bounded
when used by grants; enabled compatibility checks every affected grant; permission
retirement removes effective permission without rewriting references; no semantic
identifier reuse or tenant-specific shared-catalog copies.

Entry: review unresolved public record/definition changes, root establishment,
direct-human/self binding support needed by the selected operations, and the
existing Auth registry's compatibility mapping. Do not infer their answers from
the lab fixture. Platform publication is a separate explicitly bounded integration;
do not add a context-free escape to the initial tenant/application component.

Replace fixture identity/administration with trusted real adapters. Ensure
administrative grants and membership evidence share the validated write ordering
or a proved conditional-write protocol; a prior remote allow is insufficient.
The CLI can reuse its parsing/application interface with authenticated composition.

**Exit:** no fixture identity on a production path; both real gates and storage
consistency pass; existing Auth behavior has an explicit migration/rollback plan.

### CP6 — PostgreSQL provider and data portability

Implement PostgreSQL-specific migrations, queries, transaction/lock protocol and
error translation beneath the unchanged provider port. Run all contract and ABV
tests against both backends, including two writers, catalog/tenant ordering,
absent-row uniqueness races, foreign references and cancellation.

Perform a SQLite → PostgreSQL rehearsal preserving exact IDs, content selections,
controls and relations; compare record counts/content and behavioral decisions.
Plan a maintenance/quiescence boundary or separately reviewed live migration;
do not assume dual writes are safe. Keep the source database recoverable.

**Exit:** tests demonstrate backend equivalence for supported operations and
verified transfer. Until then the claim is “provider boundary prepared,” not
“PostgreSQL supported.”

## Planning review and execution gate

The source-backed architecture, CLI reuse and mandatory tenant/application pair
are captured. The plan contains no new canonical wire schemas and no runtime
implementation. Review the proposed internal interfaces and local-test scope
before task execution. The user need not re-answer already settled authority rules.

**Recommended execution:** inline, one checkpoint at a time, starting T1.
Subagent-driven execution remains an optional explicit choice, not automatic
delegation. Neither choice authorizes committing the currently uncommitted
handbook/presentation or changing the existing Auth service.
