# CP2 — SQLite provider

**Implementation in progress; verification and independent review pending.**
This checkpoint adds persistence, not production Auth integration. The source
requirements are [Task 3](../../plan/abv-implementation-plan.md) and the
[provider design](../../plan/abv-design.md).

## Responsibility boundary

The provider stores and retrieves authority records within a mandatory tenant
and application context. It supplies one consistent snapshot and commits one
explicit write set. It does not decide whether Maya may administer assignments
or whether a proposed grant fits a valid parent route: CP3's coordinator will
run the authorization evaluator and ABV before producing a write set.

No CLI command calls raw storage. The provider port is internal to the Go module;
there is no SQL object in its interface and no public `authorized: true` flag.
CP2's storage tests are not evidence that the two authorization gates exist yet.

## Required transaction flow

![SQLite provider transaction flow within tenant/application context](assets/sqlite-transaction.svg)

For a write, acquire a connection and transaction protection before reading
authority evidence. Establish the tenant/application installation, then load
its application-owned catalog and area-owned records. Invoke the callback once.
Only its explicit supported write set may persist; changing the returned
snapshot's maps does not itself change the database.

A successful return follows commit. Failure before commit must leave no partial
write; rollback and connection cleanup also apply to callback panics and cancelled
contexts. A lock conflict is a distinguishable failure, not a callback retry or
evidence that the parent is absent. An external caller cannot save an earlier
validation result as a reusable authorization ticket.

SQLite's write transactions serialize writers; starting with `BEGIN IMMEDIATE`
obtains that protection before validation reads. WAL permits concurrent readers
to continue with their prior snapshot. These behaviors are the basis of the
provider tests, not assumptions that every database shares the same SQL.
[SQLite isolation documentation](https://www.sqlite.org/isolation.html).

Foreign-key enforcement is connection-specific, so provider connections must
enable and verify it. [SQLite foreign-key documentation](https://www.sqlite.org/foreignkeys.html).
Neither database locks nor foreign keys establish current-time grant validity;
the future coordinator must perform that authorization check.

## Isolation and data ownership

The current migration separates these record families:

| Ownership | Tables |
|---|---|
| Provider metadata, not authority | `abv_metadata` |
| Shared application definitions | `applications`, `permissions`, `scope_definitions`, `supported_scope_keys` |
| Tenant/application installation and authority | `installations`, `roles`, `grant_controls`, `grant_contents`, `assignments`, `teams`, `memberships`, `trusted_roots` |

The provider's ordinary write set currently supports only new assignments.
Other families are supplied by controlled test setup until their authorized
mutation operations are implemented. `Open` and `Close` manage storage resources;
all record reads and writes require the explicit area.

Application catalog definitions are shared by that application. Tenant-specific
authority records and relationships carry composite tenant/application keys.
The installation check precedes catalog access; there is no global grant-ID
search to infer missing context and no independently editable per-tenant catalog.

Persist grant control separately from immutable content and assignments. The
same grant/recipient binding cannot be duplicated merely because its existing
assignment is disabled. Indexed columns and their canonical JSON payload must
agree. Corrupt stored evidence is an evaluation/storage error, not a valid empty
authority set.

Do not use cascading deletion to erase dependent authority. References whose
absence represents canonical orphanhood must remain representable; foreign keys
must not make an orphan impossible to preserve. Physical deletion/lifecycle
policy remains later work.

## Controlled fixtures

Test setup may create coherent typed fixtures in a **new disposable database**.
It must refuse an existing path, conflicting shared catalogs or an unrelated
database. Ordinary provider operations cannot call a seed-into-live-database
method. The fixture's explicit established-root evidence is not inferred from
a missing parent field and is not a production bootstrap trust procedure.

The provider may use internal JSON for non-canonical projections such as team
and catalog records. That is a storage encoding, not publication of a new
canonical authorization API. Approved grant/control/assignment representations
keep their version and adopted revision fields.

## Acceptance checklist

| Required evidence | Status |
|---|---|
| Full-value persistence and reopen for all planned record families | Pending |
| Same IDs in different tenants and different applications remain isolated | Pending |
| Unknown installation and zero context never invoke callbacks | Pending |
| Shared application catalog, not independent tenant copies | Pending |
| Duplicate grant/recipient rejection includes disabled bindings | Pending |
| Two-row failure rolls back both rows, including after reopen | Pending |
| Callback snapshot mutation cannot bypass persistence checks | Pending |
| Callback error, panic and cancellation release the transaction | Pending |
| Two independent providers: write ordering/conflict and consistent reads | Pending |
| WAL and foreign keys verified on provider connections | Pending |
| Corrupt evidence is an error, never a partial successful snapshot | Pending |
| Configured snapshot bound errors rather than truncating proof | Pending |
| Provider-neutral conformance separate from SQLite-specific tests | Pending |
| Full tests, race detector, vet and independent review | Pending |

SQLite-only evidence will be recorded here once available. PostgreSQL, real
Auth administration, actual lineage resolution, lifecycle mutation and working
CLI commands are not claimed by this checkpoint.
