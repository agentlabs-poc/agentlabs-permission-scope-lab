PRAGMA foreign_keys = ON;

CREATE TABLE abv_metadata (
    marker TEXT PRIMARY KEY CHECK (marker = 'agentlabs-abv'),
    schema_version INTEGER NOT NULL CHECK (schema_version = 4)
);

-- The ABV-123 L1 record store. Permissions and scopes live here; the remaining
-- concepts still have their own tables and move deliberately, one at a time.
--
-- tenant_id is '' rather than NULL for a record that is not tenant-scoped:
-- both SQLite and PostgreSQL treat NULLs in a unique index as distinct, which
-- would let duplicate records coexist. An empty string is a real, comparable
-- absent value, and the boundary column names which case applies.
CREATE TABLE abv_l1_records (
    boundary TEXT NOT NULL CHECK (boundary IN ('application', 'tenant')),
    tenant_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    key1 TEXT NOT NULL, key2 TEXT NOT NULL,
    key3 TEXT NOT NULL DEFAULT '', key4 TEXT NOT NULL DEFAULT '',
    key5 TEXT NOT NULL DEFAULT '', key6 TEXT NOT NULL DEFAULT '',
    key7 TEXT NOT NULL DEFAULT '', key8 TEXT NOT NULL DEFAULT '',
    key9 TEXT NOT NULL DEFAULT '', key10 TEXT NOT NULL DEFAULT '',
    revision INTEGER NOT NULL DEFAULT 0 CHECK (revision >= 0),
    value TEXT NOT NULL,
    ts TEXT NOT NULL DEFAULT (datetime('now')),
    state TEXT NOT NULL DEFAULT 'enabled' CHECK (state IN ('enabled', 'disabled', 'deleted')),
    CHECK ((boundary = 'application' AND tenant_id = '')
        OR (boundary = 'tenant' AND tenant_id <> '')),
    PRIMARY KEY (boundary, tenant_id, application_id,
                 key1, key2, key3, key4, key5, key6, key7, key8, key9, key10, revision),
    FOREIGN KEY (application_id) REFERENCES applications(application_id)
);

-- One index for every structural prefix and revision ordering: a btree serves
-- any left-anchored prefix of its columns, so the identity key above already
-- covers them. No per-record-type index is added here.
CREATE INDEX abv_l1_records_prefix ON abv_l1_records
    (boundary, tenant_id, application_id, key1, key2, key3, key4, key5);

CREATE TABLE applications (
    application_id TEXT PRIMARY KEY,
    compatibility_enabled INTEGER NOT NULL CHECK (compatibility_enabled IN (0, 1)),
    -- Bumped in the same transaction as any catalog write. A reader that sees
    -- the same generation before and after an offset walk knows nothing moved
    -- between its pages; a different one means retry. Per application, so one
    -- application's change does not invalidate another's cached catalog.
    generation INTEGER NOT NULL DEFAULT 0 CHECK (generation >= 0)
);
CREATE TABLE installations (
    tenant_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    PRIMARY KEY (tenant_id, application_id),
    FOREIGN KEY (application_id) REFERENCES applications(application_id)
);
CREATE TABLE roles (
    tenant_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK (revision > 0),
    permissions_json BLOB NOT NULL,
    PRIMARY KEY (tenant_id, application_id, role_id, revision),
    FOREIGN KEY (tenant_id, application_id) REFERENCES installations(tenant_id, application_id)
);
CREATE TABLE grant_controls (
    tenant_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    grant_id TEXT NOT NULL,
    version TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('enabled', 'disabled')),
    canonical_json BLOB NOT NULL,
    PRIMARY KEY (tenant_id, application_id, grant_id),
    FOREIGN KEY (tenant_id, application_id) REFERENCES installations(tenant_id, application_id)
);
CREATE TABLE grant_contents (
    tenant_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    grant_id TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK (revision > 0),
    canonical_json BLOB NOT NULL,
    PRIMARY KEY (tenant_id, application_id, grant_id, revision),
    FOREIGN KEY (tenant_id, application_id) REFERENCES installations(tenant_id, application_id)
);
CREATE TABLE teams (
    tenant_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    team_id TEXT NOT NULL,
    parent_id TEXT NOT NULL,
    PRIMARY KEY (tenant_id, application_id, team_id),
    FOREIGN KEY (tenant_id, application_id) REFERENCES installations(tenant_id, application_id)
);
CREATE TABLE memberships (
    tenant_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    team_id TEXT NOT NULL,
    human_id TEXT NOT NULL,
    PRIMARY KEY (tenant_id, application_id, team_id, human_id),
    FOREIGN KEY (tenant_id, application_id, team_id) REFERENCES teams(tenant_id, application_id, team_id)
);
CREATE TABLE trusted_roots (
    tenant_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    grant_id TEXT NOT NULL,
    PRIMARY KEY (tenant_id, application_id, grant_id),
    FOREIGN KEY (tenant_id, application_id) REFERENCES installations(tenant_id, application_id)
);
CREATE TABLE assignments (
    tenant_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    assignment_id TEXT NOT NULL,
    grant_id TEXT NOT NULL,
    grant_revision INTEGER NOT NULL CHECK (grant_revision > 0),
    recipient_type TEXT NOT NULL CHECK (recipient_type IN ('user', 'group')),
    recipient_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('enabled', 'disabled')),
    canonical_json BLOB NOT NULL,
    PRIMARY KEY (tenant_id, application_id, assignment_id),
    UNIQUE (tenant_id, application_id, grant_id, recipient_type, recipient_id),
    FOREIGN KEY (tenant_id, application_id, grant_id, grant_revision)
        REFERENCES grant_contents(tenant_id, application_id, grant_id, revision),
    FOREIGN KEY (tenant_id, application_id) REFERENCES installations(tenant_id, application_id)
);

-- The ownership marker is written last, so an incomplete initialization is
-- never accepted as an ABV database on a later open.
INSERT INTO abv_metadata(marker, schema_version) VALUES ('agentlabs-abv', 4);
