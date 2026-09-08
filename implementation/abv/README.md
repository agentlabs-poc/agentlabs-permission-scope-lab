# ABV — authority-boundary validation

An isolated Go module for the authorization model developed in this repository.
CP1 is delivered; CP2 storage is being implemented and reviewed. This is not a
production Auth service or working database CLI.
See the [implementation plan](../plan/abv-implementation-plan.md) and
[checkpoint evidence](docs/acceptance.md).

## Current components

- `domain`: canonical core grant/control/assignment/identity records and internal
  context/error/relationship projections. `Area` requires tenant and application.
- `internal/codec`: strict supported-core JSON decoding. Representation validity
  does not make a record authorized, trusted or safe to persist.
- `internal/validation`: registered definition checks and non-expanding permission/
  scope composition. Child scope is appended as AND predicates, never map-overwritten.
- `internal/storage`: SQL-free snapshot/write-set provider boundary; the SQLite
  implementation and reusable provider conformance suite are under CP2 review.
- `internal/lab`: controlled fixtures for new disposable SQLite databases only,
  not a seed-into-live-database or ordinary bootstrap API.
- `application`: interface between reusable CLI and later in-process facade.
- `cli`: testable command shell with injected streams, no SQL or process exit.
  Help is available; data operations explicitly return unsupported at CP1.

## Run the checks

From this directory, using Go 1.25 or later:

```sh
go test ./... -count=1
go test -race ./... -count=1
go vet ./...
```

CP1 introduced no external dependencies. CP2 adds pinned `modernc.org/sqlite`
and its dependencies; see [provider checkpoint evidence](docs/sqlite-provider.md).
Actual lineage resolution, the executable and successful protected CLI mutation
commands remain CP3 work.

## Input and authority boundaries

Grant content has no recipient or tenant/application scope fields. The context
is passed separately to authority checks, and retained on internal routes. Exact
IDs are preserved; there is no default tenant, case normalization or wildcard
context. The provider must establish the installation before catalog use.

Only supported string version `"1"` core records are decoded. Unsupported
extensions are rejected, not discarded. Null values, duplicate JSON keys,
alternative field casing and mixed direct/role permission sources are not
accepted. Empty scope `{}` remains meaningful; missing/null scope is invalid.
Core decoding represents absent-parent root-shaped records but does not trust
them. Ordinary root establishment must never be inferred from parent omission.

Prototype parser limits: 1 MiB per record and 64 nested levels. The supported
subset requires non-empty, duplicate-free permission lists and positive signed
64-bit revision values; these limits are not approval of full public schemas.
An empty validity object is unsupported; contradictory time windows are malformed.
Validity instants are parsed and preserved; current time eligibility is later
resolver/mutation work. The application owns domain values and `$self` meaning.

`CheckContent` assumes a provider-established area's catalog and role snapshot.
`Narrow` assumes a valid supported parent and checked child. It derives the full
selected permission list from direct content or the exact role-revision record;
callers cannot substitute or trim a separate expansion. Neither
is an authorization result or validation ticket. Do not expose these helpers as
an unprotected write API. Cross-recipient `$self` containment remains unresolved;
copying token strings is not proof of authority to distribute them.
