# ABV-123 D3 — proposed five-table SQLite storage contract

**Status:** design candidate, not approved or applied. This document owns only
the storage/key/query proof requested by ABV-123-08. It does not define public
wire JSON, permissions, middleware, the Auth Evaluator, or the authority-boundary
validator. Existing typed operations and the full-snapshot provider remain the
working contract until equivalence is proven.

Sources: `abv-123-storage-plan.md`, `abv-123-auth-reconciliation.md`, current
`migrations/001_initial.sql`, `snapshot.go`, `update.go`, `catalog.go`,
`grant_revision.go`, `role.go`, and the canonical revision/root rules cited by
those plans.

## 1. Keys and representations

The one L1 table uses a non-null outer-boundary key:

```text
(tenant_id, application_id, record_kind, record_id, revision)
```

- `tenant_id = ''` means the explicitly application-wide boundary; it is valid
  only for `application`, `permission`, `scope`, and `permission_support`.
- A non-empty `tenant_id` means the tenant/application boundary; it is valid only
  for `installation`, `role_revision`, `grant_control`, `grant_revision`, and
  `root_trust`.
- `application_id` is always explicit. There is no NULL/default-tenant lookup,
  wildcard boundary, or cross-boundary fallback.
- `revision` is an SQLite INTEGER: positive for role/grant revisions and exactly
  zero for unrevisioned kinds. Latest selection therefore sorts numerically.
- `record_kind` is closed storage metadata. It cannot be supplied as authority
  or added to canonical public records.
- `record_id` remains the actual domain identity: application ID, permission ID,
  scope key, role ID, or grant ID. Installation uses its application ID; a
  permission-support row uses its permission ID.

`payload_json` stores the bytes already owned by the current representation:
canonical grant-control, grant-revision and assignment JSON; the existing role
permission array; allowed-token array; supported-scope-key array; and the current
application/installation/root internal projections. It is not a new public
generic record envelope. Writes must decode once, validate, derive indexed
columns, encode canonically, and compare decoded identity/index fields with the
columns before entering SQL. Reads repeat that comparison and treat disagreement
as malformed storage, as the current provider does.

`active`, `status`, `parent_grant_id`, `role_id`, and `role_revision` are private
query columns, constrained to the kinds that own them. They do not add fields to
canonical JSON. `permission_support` is likewise an internal typed row used to
retain the ordered support list; it does not create a public record kind.

Examples:

```text
('',     'hrms', 'permission',        'payslip:read', 0)
('',     'hrms', 'scope',             'department',   0)
('',     'hrms', 'permission_support','payslip:read', 0)
('acme', 'hrms', 'installation',      'hrms',         0)
('acme', 'hrms', 'role_revision',     'reader',       12)
('acme', 'hrms', 'grant_control',     'G1',            0)
('acme', 'hrms', 'grant_revision',    'G1',            3)
('acme', 'hrms', 'root_trust',        'G0',            0)
```

## 2. Candidate DDL

This is comparison DDL, not migration SQL. It assumes SQLite JSON functions;
S3 must verify they exist in the actual driver before this candidate is eligible.

```sql
PRAGMA foreign_keys = ON;

CREATE TABLE abv_metadata (
    marker TEXT PRIMARY KEY CHECK (marker = 'agentlabs-abv'),
    schema_version INTEGER NOT NULL CHECK (schema_version > 1)
) STRICT;

CREATE TABLE abv_l1_records (
    tenant_id TEXT NOT NULL,
    application_id TEXT NOT NULL CHECK (application_id <> ''),
    record_kind TEXT NOT NULL CHECK (record_kind IN (
        'application', 'permission', 'scope', 'permission_support',
        'installation', 'role_revision', 'grant_control',
        'grant_revision', 'root_trust'
    )),
    record_id TEXT NOT NULL CHECK (record_id <> ''),
    revision INTEGER NOT NULL,
    payload_json BLOB NOT NULL,
    active INTEGER CHECK (active IN (0, 1)),
    status TEXT CHECK (status IN ('enabled', 'disabled')),
    parent_grant_id TEXT,
    role_id TEXT,
    role_revision INTEGER,

    app_ref_tenant TEXT GENERATED ALWAYS AS ('') STORED,
    app_ref_kind TEXT GENERATED ALWAYS AS ('application') STORED,
    app_ref_id TEXT GENERATED ALWAYS AS (application_id) STORED,
    app_ref_revision INTEGER GENERATED ALWAYS AS (0) STORED,
    install_ref_tenant TEXT GENERATED ALWAYS AS (
        CASE WHEN tenant_id <> '' THEN tenant_id END
    ) STORED,
    install_ref_kind TEXT GENERATED ALWAYS AS ('installation') STORED,
    install_ref_id TEXT GENERATED ALWAYS AS (application_id) STORED,
    install_ref_revision INTEGER GENERATED ALWAYS AS (0) STORED,

    PRIMARY KEY (tenant_id, application_id, record_kind, record_id, revision),
    FOREIGN KEY (app_ref_tenant, application_id, app_ref_kind,
                 app_ref_id, app_ref_revision)
      REFERENCES abv_l1_records
        (tenant_id, application_id, record_kind, record_id, revision)
      DEFERRABLE INITIALLY DEFERRED,
    FOREIGN KEY (install_ref_tenant, application_id, install_ref_kind,
                 install_ref_id, install_ref_revision)
      REFERENCES abv_l1_records
        (tenant_id, application_id, record_kind, record_id, revision)
      DEFERRABLE INITIALLY DEFERRED,

    CHECK ((tenant_id = '' AND record_kind IN
             ('application','permission','scope','permission_support')) OR
           (tenant_id <> '' AND record_kind IN
             ('installation','role_revision','grant_control',
              'grant_revision','root_trust'))),
    CHECK ((record_kind IN ('role_revision','grant_revision') AND revision > 0) OR
           (record_kind NOT IN ('role_revision','grant_revision') AND revision = 0)),
    CHECK (record_kind <> 'application' OR record_id = application_id),
    CHECK (record_kind <> 'installation' OR record_id = application_id),
    CHECK ((record_kind = 'permission' AND active IS NOT NULL) OR
           (record_kind <> 'permission' AND active IS NULL)),
    CHECK ((record_kind = 'grant_control' AND status IS NOT NULL) OR
           (record_kind <> 'grant_control' AND status IS NULL)),
    CHECK (role_revision IS NULL OR role_revision > 0),
    CHECK ((role_id IS NULL) = (role_revision IS NULL)),
    CHECK (record_kind = 'grant_revision' OR
           (parent_grant_id IS NULL AND role_id IS NULL AND role_revision IS NULL))
) STRICT, WITHOUT ROWID;

-- Only kinds with an approved mutable control are updateable. Their providers
-- still constrain which fields may change and use compare-and-write.
CREATE TRIGGER abv_l1_immutable_record_update
BEFORE UPDATE ON abv_l1_records
WHEN OLD.record_kind NOT IN ('application','permission','grant_control')
BEGIN
    SELECT RAISE(ABORT, 'immutable ABV record');
END;

-- The folded ordered list loses supported_scope_keys' two ordinary FKs.
-- This trigger restores exact permission/scope existence and duplicate safety.
CREATE TRIGGER abv_l1_permission_support_insert
BEFORE INSERT ON abv_l1_records
WHEN NEW.record_kind = 'permission_support'
BEGIN
    SELECT CASE WHEN json_valid(NEW.payload_json) = 0
                      OR json_type(NEW.payload_json) <> 'array'
      THEN RAISE(ABORT, 'invalid supported scope key list') END;
    SELECT CASE WHEN NOT EXISTS (
        SELECT 1 FROM abv_l1_records p
         WHERE p.tenant_id = '' AND p.application_id = NEW.application_id
           AND p.record_kind = 'permission' AND p.record_id = NEW.record_id
           AND p.revision = 0
    ) THEN RAISE(ABORT, 'missing permission') END;
    SELECT CASE WHEN EXISTS (
        SELECT 1 FROM json_each(NEW.payload_json) j
         WHERE j.type <> 'text' OR NOT EXISTS (
            SELECT 1 FROM abv_l1_records s
             WHERE s.tenant_id = '' AND s.application_id = NEW.application_id
               AND s.record_kind = 'scope' AND s.record_id = j.value
               AND s.revision = 0
         )
    ) THEN RAISE(ABORT, 'missing scope definition') END;
    SELECT CASE WHEN EXISTS (
        SELECT value FROM json_each(NEW.payload_json)
         GROUP BY value HAVING count(*) <> 1
    ) THEN RAISE(ABORT, 'duplicate supported scope key') END;
END;

CREATE TABLE abv_teams (
    tenant_id TEXT NOT NULL CHECK (tenant_id <> ''),
    application_id TEXT NOT NULL CHECK (application_id <> ''),
    team_id TEXT NOT NULL CHECK (team_id <> ''),
    parent_id TEXT,
    install_kind TEXT GENERATED ALWAYS AS ('installation') STORED,
    install_id TEXT GENERATED ALWAYS AS (application_id) STORED,
    install_revision INTEGER GENERATED ALWAYS AS (0) STORED,
    PRIMARY KEY (tenant_id, application_id, team_id),
    FOREIGN KEY (tenant_id, application_id, install_kind,
                 install_id, install_revision)
      REFERENCES abv_l1_records
        (tenant_id, application_id, record_kind, record_id, revision),
    FOREIGN KEY (tenant_id, application_id, parent_id)
      REFERENCES abv_teams (tenant_id, application_id, team_id)
      DEFERRABLE INITIALLY DEFERRED,
    CHECK (parent_id IS NULL OR parent_id <> team_id)
) STRICT, WITHOUT ROWID;

CREATE TABLE abv_memberships (
    tenant_id TEXT NOT NULL CHECK (tenant_id <> ''),
    application_id TEXT NOT NULL CHECK (application_id <> ''),
    team_id TEXT NOT NULL CHECK (team_id <> ''),
    human_id TEXT NOT NULL CHECK (human_id <> ''),
    PRIMARY KEY (tenant_id, application_id, team_id, human_id),
    FOREIGN KEY (tenant_id, application_id, team_id)
      REFERENCES abv_teams (tenant_id, application_id, team_id)
) STRICT, WITHOUT ROWID;

CREATE TABLE abv_assignments (
    tenant_id TEXT NOT NULL CHECK (tenant_id <> ''),
    application_id TEXT NOT NULL CHECK (application_id <> ''),
    assignment_id TEXT NOT NULL CHECK (assignment_id <> ''),
    grant_id TEXT NOT NULL CHECK (grant_id <> ''),
    grant_revision INTEGER NOT NULL CHECK (grant_revision > 0),
    recipient_type TEXT NOT NULL CHECK (recipient_type IN ('user', 'group')),
    recipient_id TEXT NOT NULL CHECK (recipient_id <> ''),
    status TEXT NOT NULL CHECK (status IN ('enabled', 'disabled')),
    canonical_json BLOB NOT NULL,
    grant_kind TEXT GENERATED ALWAYS AS ('grant_revision') STORED,
    recipient_team_id TEXT GENERATED ALWAYS AS (
        CASE WHEN recipient_type = 'group' THEN recipient_id END
    ) STORED,
    PRIMARY KEY (tenant_id, application_id, assignment_id),
    UNIQUE (tenant_id, application_id, grant_id, recipient_type, recipient_id),
    FOREIGN KEY (tenant_id, application_id, grant_kind, grant_id, grant_revision)
      REFERENCES abv_l1_records
        (tenant_id, application_id, record_kind, record_id, revision),
    FOREIGN KEY (tenant_id, application_id, recipient_team_id)
      REFERENCES abv_teams (tenant_id, application_id, team_id)
) STRICT, WITHOUT ROWID;

CREATE INDEX abv_l1_parent_grant
    ON abv_l1_records
       (tenant_id, application_id, parent_grant_id, record_id, revision)
    WHERE record_kind = 'grant_revision' AND parent_grant_id IS NOT NULL;
CREATE INDEX abv_l1_role_reference
    ON abv_l1_records
       (tenant_id, application_id, role_id, role_revision, record_id, revision)
    WHERE record_kind = 'grant_revision' AND role_id IS NOT NULL;
CREATE INDEX abv_assignments_recipient
    ON abv_assignments
       (tenant_id, application_id, recipient_type, recipient_id, assignment_id);
CREATE INDEX abv_assignments_grant
    ON abv_assignments
       (tenant_id, application_id, grant_id, assignment_id);
CREATE INDEX abv_teams_parent
    ON abv_teams (tenant_id, application_id, parent_id, team_id)
    WHERE parent_id IS NOT NULL;
CREATE INDEX abv_memberships_human
    ON abv_memberships (tenant_id, application_id, human_id, team_id);
```

`abv_metadata` is the fifth table and retains the ownership-marker-last rule.
The exact next `schema_version` belongs to a later migration plan; `> 1` above
intentionally does not claim a version number.

## 3. Constraint survival and known limits

| Existing invariant | Candidate enforcement |
|---|---|
| Installation references application | Deferred L1 self-FK to the exact application row. |
| Permission/scope reference application | Same exact application self-FK; app-wide kinds require `tenant_id=''`. |
| Supported key references permission and scope; keys and ordinals unique | One ordered JSON array per permission (`permission_support` PK); insert trigger checks exact permission, every scope and duplicate values. Array order is the ordinal. |
| Role, grant control/revision, team and root trust reference installation | Tenant L1 rows use the deferred installation self-FK; teams use an ordinary FK. |
| Membership references exact team | Ordinary composite FK. |
| Assignment references exact numeric grant revision and installation | Ordinary composite FK to `grant_revision`; that row itself references the installation. |
| Assignment ID unique in area | Assignment primary key. |
| Grant/recipient unique even while disabled | Unconditional assignment unique constraint; status is absent from it. |
| Role/grant revision identity is numeric and unique | L1 PK plus positive-revision check; immutable-update trigger. |
| Grant-wide status is distinct from immutable revision content | Separate `grant_control` kind at revision 0; only it owns `status`. |
| Root trust is explicit | Separate `root_trust` kind at revision 0. Its creation remains available only to the trusted initialization operation; parent omission never creates it. |
| Indexed metadata agrees with payload | Existing single codec/validation path before write and after read; SQL CHECKs cannot safely prove arbitrary canonical JSON equivalence. |

The supported-list trigger is the one place where folding weakens ordinary FK
mechanics: SQLite cannot attach an FK to every JSON array element. It supplies
insert-time database enforcement, but future scope/permission deletion would need
reverse trigger guards because JSON children do not participate in FK delete
checks. Deletion is deferred in this build; any later deletion design must add
and adversarially test those guards before this candidate is equivalent. If the
actual driver lacks the required JSON behavior or trigger tests find a gap, retain
the current `supported_scope_keys` table. That would make six core tables and is
the only warranted alternative found; integrity wins over the five-table count.

Team parentage had no FK in schema v1. The candidate adds a deferred self-FK and
maps the current empty root parent to SQL NULL; cycle rejection still belongs to
the validator because a self-FK cannot prove acyclicity. Conditional group-recipient
FKs strengthen current storage; direct-human assignments intentionally have no
team FK and remain subject to their explicit supported/unsupported operation path.

SQL cannot distinguish a trusted bootstrap caller from an ordinary writer. The
provider must expose no generic L1 write: root-trust insertion is a dedicated
trusted-initialization capability, in the same transaction as its prerequisites.
Direct database access is therefore privileged infrastructure, as it is today.

Mutable grant-control, assignment-status, application compatibility and future
permission-lifecycle writes require compare-and-write against the exact prior
canonical value/state inside `BEGIN IMMEDIATE`. Immutable inserts reject duplicate
keys. Validation evidence and the write share that transaction; no stale validation
ticket, fallback revision, silent retry, or partial write is allowed.

## 4. Indexed query contract

All statements bind exact boundary values supplied by the typed operation. A
caller-selected kind is accepted only after a closed operation route selects the
allowed literal.

```sql
-- Exact L1 record/revision (the L1 primary key).
SELECT payload_json, active, status, parent_grant_id, role_id, role_revision
  FROM abv_l1_records
 WHERE tenant_id=? AND application_id=? AND record_kind=?
   AND record_id=? AND revision=?;

-- Latest numeric revision; never status-based fallback.
SELECT revision, payload_json, parent_grant_id, role_id, role_revision
  FROM abv_l1_records
 WHERE tenant_id=? AND application_id=? AND record_kind='grant_revision'
   AND record_id=?
 ORDER BY revision DESC LIMIT 1;

-- Human's direct teams / team's humans.
SELECT team_id FROM abv_memberships
 WHERE tenant_id=? AND application_id=? AND human_id=? AND team_id>?
 ORDER BY team_id LIMIT ?;
SELECT human_id FROM abv_memberships
 WHERE tenant_id=? AND application_id=? AND team_id=? AND human_id>?
 ORDER BY human_id LIMIT ?;

-- Recipient holdings / all holdings of a grant. Disabled rows remain visible.
SELECT assignment_id, grant_id, grant_revision, status, canonical_json
  FROM abv_assignments
 WHERE tenant_id=? AND application_id=?
   AND recipient_type=? AND recipient_id=? AND assignment_id>?
 ORDER BY assignment_id LIMIT ?;
SELECT assignment_id, recipient_type, recipient_id, grant_revision, status,
       canonical_json
  FROM abv_assignments
 WHERE tenant_id=? AND application_id=? AND grant_id=? AND assignment_id>?
 ORDER BY assignment_id LIMIT ?;

-- Deterministic required parent holding (unique index).
SELECT assignment_id, grant_revision, status, canonical_json
  FROM abv_assignments
 WHERE tenant_id=? AND application_id=? AND grant_id=?
   AND recipient_type='group' AND recipient_id=?;

-- Child teams.
SELECT team_id FROM abv_teams
 WHERE tenant_id=? AND application_id=? AND parent_id=? AND team_id>?
 ORDER BY team_id LIMIT ?;

-- Actual adopted assignments whose exact revision references a parent.
-- Disabled assignments remain structural evidence; no latest substitution.
SELECT a.assignment_id, a.grant_id, a.grant_revision, a.recipient_type,
       a.recipient_id, a.status, a.canonical_json, g.payload_json
  FROM abv_l1_records AS g
  JOIN abv_assignments AS a
    ON a.tenant_id=g.tenant_id AND a.application_id=g.application_id
   AND a.grant_id=g.record_id AND a.grant_revision=g.revision
 WHERE g.tenant_id=? AND g.application_id=?
   AND g.record_kind='grant_revision' AND g.parent_grant_id=?
   AND a.assignment_id>?
 ORDER BY a.assignment_id LIMIT ?;
```

The L1 primary key covers exact application permission/scope/support reads, exact
role/grant/control/root reads, and reverse-order latest scans; no duplicate latest
index is proposed before `EXPLAIN QUERY PLAN` proves one necessary. Listings use
keyset cursors over the shown index suffixes. Pagination never applies to evidence
required for an authorization decision.

Complete lineage evidence is a bounded frontier traversal, not one recursive SQL
query assumed correct: batch-fetch adopted grant revisions by exact keys, their
controls and role revisions, exact parent-team holdings, child teams, parent-grant
references, memberships, catalog definitions/support lists, installation and root
trust. Follow both team and grant edges, retaining disabled structural bridges and
all affected branches. Deduplicate exact typed keys. If the operation's depth/work/
time bound is reached, fail the operation; never return a partial allow. Every
batch runs on the same read transaction/connection as the decision, and protected
writes use the same `BEGIN IMMEDIATE` view through commit.

## 5. Administration callback dependency

Today `Provider.Update` passes a complete `storage.Snapshot`, and administration
callbacks may inspect any map in it. A targeted reader cannot pass a partial map
under that type and claim behavioral equivalence. Before replacing the loader,
each operation must have an explicit dependency declaration and receive a new
evidence type that is complete for that operation, or the provider must retain
the full-snapshot path. The first option is a provider/core contract change and
requires separate design approval; this D3 draft does not name or redefine the
validator or middleware.

Required proof: instrument every current administrative callback against hostile
fixtures, record every key/query class it reads, compare decisions and write sets
between full snapshot and targeted evidence, and reject undeclared access. Until
that succeeds for an operation, its write continues through the coherent full
snapshot. Exact inspection and bounded listings can be proven independently.

## 6. Bounded comparison design (no benchmark claims)

Compare exactly the three storage-plan configurations: current schema/current
loader, current schema/indexed targeted retrieval, and this five-table candidate/
equivalent retrieval. Use identical canonical bytes, fixtures, operation results,
driver, pragmas, durability, machine and fixed seeds.

Run 100/1,000/10,000 loader-equivalent records plus 100,000 unrelated background
rows; report the old loader's configured-limit failure rather than timing it as a
success. Hold a selected route fixed while varying unrelated rows, then separately
vary depth 1/8/32, branching, revisions 1/10/100 and human membership count.
Measure exact inspection, membership-plus-assignment retrieval, complete diagnosis,
assignment creation, grant publication, and both status changes, including allows,
rejections, a paired reader/writer and competing duplicate inserts.

Correctness runs first: isolation, exact revision, immutability, uniqueness,
orphaned support, disabled bridges, every affected branch, root trust, hostile
callback access, parent scope AND, stale evidence, cancellation and rollback.
Then capture query plans, statements, rows/payload bytes loaded, latency,
throughput, allocations, transaction/lock duration, contention errors, DB/index
size and write cost. Use five fixed repeats and report median/variation; use at
least 200 operations for p50/p95 where the time cap permits and label insufficient
samples. Fresh mutation fixtures prevent duplicate rejection from masquerading
as throughput. Warm and reopened connections are separate; neither is called a
cold OS-cache result.

Apply the proposed gates already recorded in `abv-123-storage-plan.md`; they are
accept/reject criteria, not current performance claims. Failure of integrity,
callback completeness, targeted-scaling, or the comparison gates retains the
current schema. No cache, migration, dual write, PostgreSQL claim, or tuning loop
is part of this design.

## 7. Decisions and uncompleted work

No policy decision is made here. Before implementation, owners still must approve
the physical contract and decide whether the JSON-trigger dependency is acceptable
or the sixth `supported_scope_keys` table is retained. The exact schema version,
migration/cutover/rollback, per-operation evidence interface, evidence limits,
public read visibility, and wire/permission mappings remain separate work.

No DDL has been executed, no database migrated, no benchmark run, and no
authorization behavior approved by producing this draft.
