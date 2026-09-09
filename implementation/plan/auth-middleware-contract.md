# AUTH-MW M0 — internal first-slice contract

Status: implementation contract for the local harness. The Go values below are
internal seams, not new canonical JSON, grant, evidence, result, or Auth wire
fields. Production identity and Auth transports remain unimplemented.

## 1. Placement and trust boundary

```text
API service                         Auth Service / local test composition
-----------                         -------------------------------------
trusted identity adapter            current Auth records
static endpoint policy                        |
application binding code            read-only route resolver
          |                                    |
          +---- Request + AuthoritySource -----+
                         |
                     Evaluator
                         |
                  allow / deny / error
                         |
              constrained application effect
```

The evaluator and `net/http` wrapper belong to the API service. They are
read-only consumers. ABV mutation, the core Auth Validator, its protected
coordinator, and authority storage stay inside Auth. The API package must not
import ABV mutation or SQLite packages, inspect Auth tables, or validate issuance
again. Auth's own APIs still require their administrative evaluator and, for an
authority-changing operation, ABV through persistence.

For the local harness, an adapter may read the existing ABV SQLite database and
return resolved routes. SQLite is therefore a composition/test dependency, not
an evaluator dependency and not a second authority model. No remote Auth endpoint
is implied.

## 2. Exact Go seam

Recommended module: `implementation/authmiddleware`, module path
`agentlabs.local/authmiddleware`. Keep its core package standard-library-only.

```go
package authmiddleware

import (
	"context"
	"net/http"
	"time"
)

type Area struct {
	TenantID      string
	ApplicationID string
}

type Actor struct {
	Type string // first slice: "user" only
	ID   string
}

type Identity struct {
	Version string
	Actor   Actor
	HumanID string
}

type RequestContext struct {
	Area     Area
	Identity Identity
}

// IdentitySource establishes these values from authenticated server context.
// Request JSON and an unverified/merely decoded token cannot implement this port.
type IdentitySource interface {
	Establish(context.Context, *http.Request) (RequestContext, error)
}

type Source string

const (
	SourcePath Source = "path"
	SourceBody Source = "body"
)

type Input struct {
	Source Source `json:"source"`
	Name   string `json:"name"`
}

// Policy is server-owned static configuration. Inputs retain the approved
// policy meaning; Bindings are application code, not another policy/wire block.
type Policy struct {
	Version    string           `json:"version"`
	Method     string           `json:"method"`
	Path       string           `json:"path"`
	Permission string           `json:"permission"`
	Inputs     map[string]Input `json:"inputs"`
}

type SelectionKind uint8

const (
	SelectionExact SelectionKind = iota + 1
	SelectionAll
)

// Selection is the boundary requested by this operation. All means the entire
// trusted Area for that registered key; it never means “rows I can access”.
type Selection struct {
	Kind  SelectionKind
	Value string // required only for SelectionExact
}

// Material is produced by explicit application binding code after exact-source
// extraction and application schema validation. Keys are registered scope keys.
type Material map[string]Selection

type Request struct {
	Context    RequestContext
	Permission string
	Material   Material
}

type AuthorityQuery struct {
	Context    RequestContext
	Permission string
}

type Predicate struct {
	Key           string
	Value         string // registered literal or "$self"
	SourceGrantID string
}

// Route is an internal dependent resolved view. GrantIDs is the ordered,
// duplicate-free contributing chain for traceability. ValidUntil is the earliest
// automatic expiry in the route; nil means no automatic expiry was present.
type Route struct {
	Area        Area
	HumanID     string
	Permission  string
	GrantIDs    []string
	Predicates  []Predicate
	ValidUntil  *time.Time
}

// Authority is complete only by the successful-return contract of Load; there
// is deliberately no caller-set Complete or Trusted boolean.
type Authority struct {
	Routes []Route
}

type AuthoritySource interface {
	Load(context.Context, AuthorityQuery) (Authority, error)
}

type Decision string

const (
	Allow Decision = "allow"
	Deny  Decision = "deny"
)

// Result is the agreed minimal result shape. For allow, GrantIDs is non-empty
// and error fields are empty. For deny, GrantIDs is empty and all error fields
// are non-empty. Evaluation failure is returned as *EvaluationError, not Result.
type Result struct {
	Version            string   `json:"version"`
	Decision           Decision `json:"decision"`
	GrantIDs           []string `json:"grant_ids,omitempty"`
	ErrorCode          string   `json:"error_code,omitempty"`
	ErrorMessage       string   `json:"error_message,omitempty"`
	ErrorMessageReason string   `json:"error_message_reason,omitempty"`
}

// EvaluationError has the approved evaluation-error fields and no decision.
// Code values remain opaque because the exhaustive catalogue is unresolved.
type EvaluationError struct {
	Version       string `json:"version"`
	Code          string `json:"error_code"`
	Message       string `json:"error_message"`
	MessageReason string `json:"error_message_reason"`
	Cause         error  `json:"-"`
}

func (e *EvaluationError) Error() string
func (e *EvaluationError) Unwrap() error

type Clock interface { Now() time.Time }

type Evaluator struct { /* source AuthoritySource; clock Clock */ }

func New(source AuthoritySource, clock Clock) (*Evaluator, error)
func (e *Evaluator) Evaluate(context.Context, Request) (Result, error)
func (p Policy) Validate() error
func (r Result) Validate() error
func DecodeResult([]byte) (Result, error)
```

Only `Policy`, `Input`, `Result`, and `EvaluationError` above participate in
the selected JSON codecs. The other structs are in-process values, and the
identity copies preserve canonical meaning without establishing another wire
identity contract. `Evaluate` itself has no HTTP dependency; `net/http` appears
only on the host identity/wrapper port.

`Evaluate` returns `(allow, nil)` or `(deny, nil)` only after complete
evaluation. A source timeout, cancellation, malformed/incomplete evidence,
unsupported evidence condition, or inability to establish required freshness
returns a zero `Result` and `*EvaluationError`. Protected execution requires
exactly `result.Decision == Allow`, `err == nil`, and `result.Validate() == nil`.
The HTTP status mapping and exhaustive error-code/message catalogue are not
selected here.

`Result.Validate` enforces version `"1"`, the exact allow and deny variants from
Q-062--Q-067, a non-empty list of non-empty grant IDs for allow, all three error
fields for deny, and rejection of mixed variants. `DecodeResult` also recognizes
the complete evaluation-error variant and returns it as `*EvaluationError` with
a zero `Result`; its strict JSON decoder rejects unknown fields. The example
`NO_AUTHORIZING_GRANT` may be used only for the established no-route deny;
`AUTH_SERVICE_TIMEOUT` may be used only for an actual timeout. They are examples,
not a complete catalogue. Other failures remain fail-closed Go errors until the
catalogue supplies an honest external code/message; they must not be relabeled as
either example. This is the remaining blocker to a general result renderer, not
a blocker to the bounded evaluator harness.

## 3. Responsibility contract

### Identity adapter

A successful `Establish` means authentication, tenant binding, application
binding, and direct-human identity were established by a trusted host adapter.
The first slice accepts only `version == "1"`, actor type `user`, non-empty IDs,
and `Actor.ID == HumanID`. Tenant/application must be non-empty, non-wildcard,
valid UTF-8 values. The route/path tenant must equal the established tenant.

JWT verification, issuer/audience/time checks, proxy association, and delegation
evidence belong to a later production adapter. The harness may inject a fixed
identity, but must label it trusted test setup rather than authentication proof.

### Authority source

A successful `Load(q)` makes one indivisible assertion: `Routes` is the complete
set of currently usable routes applicable to `q.Context.Identity.HumanID` in the
exact area for `q.Permission`, including direct-human membership in groups that
hold assignments. An empty successful set conclusively means no applicable route;
an error never means an empty set.

`Authority` intentionally does not echo area, identity, query, trust, or
completeness fields. The injected source is a trusted port, and success is bound
to the exact `AuthorityQuery` argument by the method contract. The evaluator
still validates every returned route against that query. A source that cannot
make the complete query-bound assertion must return an error.

Before returning a route, the Auth-side adapter proves all Auth-owned facts:

- the human's current applicable membership and the exact group-held assignment;
- exact adopted grant and role revisions, registered/active permission selection,
  grant-wide and assignment controls, validity at the load instant;
- actual required parent-team support and every inherited predicate, without
  last-value-wins merging or cross-route combination;
- tenant/application equality, actual lineage, root legitimacy, and bounded,
  cycle-free complete resolution; and
- completeness/coherence of the read and freshness sufficient for this new check.

The API evaluator rechecks the route area, human, requested permission, shape,
predicate support, and `ValidUntil` against its decision-time clock. It must
reject an expired route and continue considering other complete routes. It does
not recheck memberships, controls, adoption, roles, root status, or lineage from
raw records. A provider must return an error rather than omit a condition it
cannot represent or prove. No positive route may contain a proxy/delegation,
generic condition, or other constraint unsupported by this first-slice type.

`ValidUntil` represents an authority rule's real automatic expiry, not a cache
TTL, lease, or proof of freshness. `Load` success itself carries the freshness
obligation. The production mechanism needed to ensure that checks begun after a
confirmed reduction cannot use withdrawn authority is a genuine integration gap;
no timestamp, cache duration, or invented version token closes it.

### Application policy, material, and handler

The endpoint owns one static `Policy`. Startup/registration validation requires
version `"1"`, one non-empty permission, exact method/path, non-empty unique local
input names, and only first-slice `path` and top-level `body` sources. Every input
must be present at exactly its declared source. Callers cannot select permission.
No query fallback, nested selector, relationship block, or policy DSL exists.

Application code explicitly maps validated local inputs to registered scope keys
and constructs `Material`. Equal names do not create that mapping. Tenant is
checked against trusted `Area`, not converted into an ordinary scope predicate.
The same parsed values must reach evaluation and constrained execution.

For each predicate in one route:

- literal `v` matches only `Selection{Kind: SelectionExact, Value: v}` for that
  key;
- `$self` matches only an exact selection whose value equals
  `Request.Context.Identity.HumanID`;
- `SelectionAll`, a missing key, an empty value, an unsupported token, or a
  conflicting repeated predicate does not match that route; and
- predicates with different keys all match (AND). Predicates from different
  routes are never combined. Any one complete matching route may allow.

A route with no predicate for a key adds no restriction on that dimension; it
does not remove the trusted `Area` or the endpoint's concrete record binding.
All contributing grant IDs for the selected route are returned. If several
routes independently match, the first slice chooses the first route after stable
lexicographic ordering by its `GrantIDs` tuple. This is deterministic selection,
not a policy preference or an exhaustive audit trace.

Material describes the requested boundary; it is not a factual relationship.
For example, path `dept=FIN, cert=C17` can match FIN/C17 authority, but does not
prove C17 belongs to FIN. Only the handler/application datastore can establish
that fact. After allow, the handler must constrain the actual read/write by the
same tenant, department, record, and self values. An ID-only fallback is forbidden.
If a concurrent factual change breaks that relation before use, the attempt stops.
There is no prepared state, reusable ticket, second authorization decision, or
grant-derived SQL/query filtering.

An explicitly FIN-bounded collection uses `dept: Exact("FIN")`. A tenant-wide
collection uses `dept: All`; FIN-only authority cannot match it, even if today's
rows all happen to be in FIN. The endpoint must deny rather than rewrite the
request into an authorized subset.

## 4. Concrete local providers

### Existing-ABV SQLite adapter (preferred integration harness)

Add a read-only Auth-side adapter over the existing ABV SQLite provider/schema;
do not read the tables in `authmiddleware` and do not create a second schema.
The consumer-facing implementation shape is:

```go
// package local harness adapter, not package authmiddleware
type SQLiteAuthoritySource struct { /* existing ABV read-only resolver */ }

func Open(path string, clock authmiddleware.Clock) (*SQLiteAuthoritySource, error)
func (s *SQLiteAuthoritySource) Load(context.Context, authmiddleware.AuthorityQuery) (authmiddleware.Authority, error)
func (s *SQLiteAuthoritySource) Close() error
```

Internally it takes one coherent read transaction over the existing schema and
uses Auth-owned read-only resolution primitives. It must discover the direct
human's memberships, resolve each matching group assignment with actual adopted
content/support, retain route predicates and contributing grant IDs, and return
no partial set on malformed rows, traversal limits, cancellation, or read error.
No mutation API or administrative fixture is exposed to the evaluator.

The current SQLite `Open`/`OpenWithOptions` path may create or migrate a missing
database and configure WAL, and `abv.OpenSQLite` constructs the mutation facade
and requires an administration port. The read adapter must not use either as a
shortcut. Its Auth-side read-only open must reject an absent file, an empty or
non-ABV database, and an unsupported schema; it must not seed, migrate, or change
journal settings. Reuse the existing schema decoding and lineage logic behind an
explicit read transaction seam rather than duplicating lineage in SQL.

Exactly where that Auth-owned read seam lives is a bounded implementation
prerequisite: it must be placed under the ABV module so Go permits reuse of its
internal storage/lineage code, while the conversion/composition code must not
make the core evaluator module depend on ABV. Pick the smallest package/module
layout satisfying those two compile-time constraints before M2/M4; this document
does not falsely claim the current internal packages are already callable.

Implementation refinement (autonomous continuation): use the existing ABV module's
`internal/storage/sqlite.Reader` for read-only Open/Read/Close and
`internal/lineage.ResolveHuman` for group-held route discovery. The thin conversion
adapter will live at `implementation/abv/localadapter`, importing the separate
`agentlabs.local/authmiddleware` module through a local Go replacement. Dependency
direction is ABV lab adapter → middleware, never middleware → ABV. This avoids a
third module or shared-domain extraction. It is local harness wiring, not an Auth
transport API. Reader/query are bounded prerequisites, not already implemented.

The read query must distinguish established inactivity (disabled, outside validity,
missing required parent assignment) from invalid or unsupported evidence. Existing
issuance rejection remains intact; the read query may skip only the explicit
inactive classification, never every generic rejection. Otherwise a corrupt record
could be reported to the evaluator as a legitimate absence of authority.

The current `agentlabs.local/abv/internal/lineage` helpers are Auth-internal and
were built for issuance. In particular they reject `$self` routes and do not yet
provide the complete human-to-group route query required here. The adapter may
reuse their literal route mechanics inside the ABV module, but must not widen or
mislabel the issuer validator merely to make consumer tests pass. The SQLite
adapter's first implementation is therefore limited to literal predicates until
a separate Auth read resolver supplies runtime `$self` routes.

The SQLite test proves coherent reads of current committed local state and that
a later read sees a committed disablement/removal. It does not prove production
cross-process reduction confirmation or transport freshness.

### Deterministic route fixture (consumer `$self` proof)

Keep one tiny in-memory `AuthoritySource` in `_test.go` that returns copied,
pre-resolved `Authority` or a configured error. It is trusted fixture evidence,
not a record resolver, authentication adapter, production provider, or alternate
database. Use it to test direct-human/team-held `$self` matching, completeness
errors, timeout, and expiry until the Auth read resolver supports those routes.

Do not let the fixture accept raw grants or infer membership/lineage; doing so
would create a second validator. Its successful response is already-resolved
evidence satisfying the `Load` contract.

## 5. Reuse decision and module consequence

The public `agentlabs.local/abv/domain` package has useful `Area`, `Identity`,
`Predicate`, and `Route` shapes, but importing it from the evaluator would couple
the new module to `agentlabs.local/abv`. That module currently requires
`modernc.org/sqlite`, so even domain-only reuse expands the API component's module
graph toward the storage implementation. Its current `Route` also lacks the
human/completeness contract required here and its single `GrantID` is insufficient
for the agreed contributing-grant result.

Therefore the first slice uses the small local internal types above and explicit
adapter conversion. Any local harness module may carry local requirements/replaces
for both modules; the production evaluator module must not require ABV.
Do not extract a shared module or refactor ABV merely for type identity. Revisit
only if a real Auth transport supplies a stable shared read contract.

## 6. Precisely supported first slice

Supported:

- synchronous direct-human requests only;
- authority held by a group through a current direct human membership;
- one server-owned permission per endpoint and exact area isolation;
- literal predicates through the SQLite adapter;
- literal and runtime `$self` matching through trusted resolved-route fixtures;
- one exact-record GET, one selected-top-level-body PUT that cannot move an
  existing record across boundaries, and one explicitly bounded collection;
- independent complete-route evaluation, deny for no matching route, and
  evaluation error for incomplete/unavailable authority; and
- Q-129 completion only for the same bounded synchronous operation after allow,
  while Q-074 still requires factual boundaries to hold through the effect.

Unsupported and fail closed:

- proxy actors/delegation, direct-recipient assignment discovery, recipient-
  relative `$self` issuance/copying, nested group membership, and arbitrary
  conditions;
- move operations requiring authority over current and proposed boundaries,
  batch/stream/queued/long-running work, pagination/count/export semantics;
- partial collection success, grant-derived query generation, cross-route
  coverage or predicate/permission mixing;
- caches, leases, retries, remote Auth calls, JWT validation, production
  freshness/withdrawal coordination, and HTTP status/catalogue policy; and
- new canonical policy, grant, evidence, relationship, snapshot, or result fields.

## 7. Handbook decisions versus real gaps

The handbook already decides: placement; one permission; exact input sources;
complete positive routes; parent AND; `$self` as the authorizing human; explicit
collection boundaries; no automatic subset filtering; allow/deny/error
distinction; supporting grant IDs; strict result variants; endpoint factual
enforcement; no prepared state; and the Q-128/Q-129/Q-074 timing boundaries.

Actual remaining integration/policy gaps are: production Auth evidence transport;
how its success proves Q-128 freshness; proxy/delegation evidence; direct-recipient
support discovery; recipient-relative `$self` issuance across recipients; move
composition; and the complete error-code/message/HTTP catalogue. None is guessed
by this contract. They do not block the bounded literal SQLite harness plus
trusted resolved-route `$self` consumer tests described above.
