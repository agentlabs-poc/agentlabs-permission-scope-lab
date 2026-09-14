PRAGMA foreign_keys = ON;

CREATE TABLE abv_metadata (
    marker TEXT PRIMARY KEY CHECK (marker = 'agentlabs-abv'),
    schema_version INTEGER NOT NULL CHECK (schema_version = 10)
);

-- The ABV-123 L1 record store. Permissions and scopes live here; the remaining
-- concepts still have their own tables and move deliberately, one at a time.
--
-- tenant_id is '' rather than NULL for a record that is not tenant-scoped:
-- both SQLite and PostgreSQL treat NULLs in a unique index as distinct, which
-- would let duplicate records coexist. An empty string is a real, comparable
-- absent value, and the boundary column names which case applies.
CREATE TABLE abv_l1_records (
    -- Who owns the record. This is the one fact the key path cannot carry:
    -- a tenant record is told by its tenant, but an application record and a
    -- platform record both have no tenant and both put a namespace in key3.
    -- Deriving it from key3 would mean holding the platform's reserved
    -- namespace list, which belongs to the auth service rather than here.
    boundary TEXT NOT NULL CHECK (boundary IN ('platform', 'application', 'tenant')),
    tenant_id TEXT NOT NULL,
    key1 TEXT NOT NULL, key2 TEXT NOT NULL,
    key3 TEXT NOT NULL DEFAULT '', key4 TEXT NOT NULL DEFAULT '',
    key5 TEXT NOT NULL DEFAULT '', key6 TEXT NOT NULL DEFAULT '',
    key7 TEXT NOT NULL DEFAULT '', key8 TEXT NOT NULL DEFAULT '',
    key9 TEXT NOT NULL DEFAULT '', key10 TEXT NOT NULL DEFAULT '',
    value TEXT NOT NULL,
    ts TEXT NOT NULL DEFAULT (datetime('now')),
    state TEXT NOT NULL DEFAULT 'enabled' CHECK (state IN ('enabled', 'disabled', 'deleted')),
    -- Padding must be contiguous: a record occupies a prefix of the slots and
    -- leaves the rest empty. A gap means the row was not written by the codec,
    -- and its identity cannot be trusted. The codec enforces this too; the
    -- constraint is what stops a future writer bypassing it.
    CHECK ((key3  <> '' OR (key4 = '' AND key5 = '' AND key6 = '' AND key7 = '' AND key8 = '' AND key9 = ''))
       AND (key4  <> '' OR (key5 = '' AND key6 = '' AND key7 = '' AND key8 = '' AND key9 = ''))
       AND (key5  <> '' OR (key6 = '' AND key7 = '' AND key8 = '' AND key9 = ''))
       AND (key6  <> '' OR (key7 = '' AND key8 = '' AND key9 = ''))
       AND (key7  <> '' OR (key8 = '' AND key9 = ''))
       AND (key8  <> '' OR  key9 = '')),
    CHECK ((boundary = 'tenant' AND tenant_id <> '')
        OR (boundary <> 'tenant' AND tenant_id = '')),
    PRIMARY KEY (boundary, tenant_id, key1, key2, key3, key4, key5, key6, key7, key8, key9, key10)
);

CREATE INDEX abv_l1_records_prefix ON abv_l1_records
    (boundary, tenant_id, key1, key2, key3, key4, key5);

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
-- The ownership marker is written last, so an incomplete initialization is
-- never accepted as an ABV database on a later open.
INSERT INTO abv_metadata(marker, schema_version) VALUES ('agentlabs-abv', 10);
