PRAGMA foreign_keys = ON;

CREATE TABLE registry_metadata (
    marker TEXT PRIMARY KEY CHECK (marker = 'agentlabs-application-registry'),
    schema_version INTEGER NOT NULL CHECK (schema_version = 1)
);

-- The application registry's L1 record store.
--
-- One canonical key/value table per domain and layer is the 123 shape, and the
-- domain is in the table name: this is application_registry, layer L1. It holds
-- which applications exist and which tenants have them installed, and nothing
-- else — Auth-AL's records live in abv_l1_records, a different table in a
-- different domain.
--
-- The envelope is the frozen 123 shape plus boundary, which is the one fact no
-- key slot can carry: an application record and a tenant's installation both
-- put the slug in key3, and only the tenant tells them apart.
CREATE TABLE application_registry_l1_records (
    boundary TEXT NOT NULL CHECK (boundary IN ('application', 'tenant')),
    tenant_id TEXT NOT NULL,
    key1 TEXT NOT NULL, key2 TEXT NOT NULL,
    key3 TEXT NOT NULL DEFAULT '', key4 TEXT NOT NULL DEFAULT '',
    key5 TEXT NOT NULL DEFAULT '', key6 TEXT NOT NULL DEFAULT '',
    key7 TEXT NOT NULL DEFAULT '', key8 TEXT NOT NULL DEFAULT '',
    key9 TEXT NOT NULL DEFAULT '', key10 TEXT NOT NULL DEFAULT '',
    value TEXT NOT NULL,
    ts TEXT NOT NULL DEFAULT (datetime('now')),
    state TEXT NOT NULL DEFAULT 'enabled' CHECK (state IN ('enabled', 'disabled', 'deleted')),
    CHECK ((boundary = 'tenant' AND tenant_id <> '')
        OR (boundary <> 'tenant' AND tenant_id = '')),
    -- Padding is contiguous: a record occupies a prefix of the slots. A gap
    -- means the row was not written by this domain's writer.
    CHECK ((key3 <> '' OR (key4 = '' AND key5 = '' AND key6 = '' AND key7 = '' AND key8 = '' AND key9 = ''))
       AND (key4 <> '' OR (key5 = '' AND key6 = '' AND key7 = '' AND key8 = '' AND key9 = ''))),
    PRIMARY KEY (boundary, tenant_id, key1, key2, key3, key4, key5, key6, key7, key8, key9, key10)
);

CREATE INDEX application_registry_l1_records_prefix ON application_registry_l1_records
    (boundary, tenant_id, key1, key2, key3);

INSERT INTO registry_metadata(marker, schema_version) VALUES ('agentlabs-application-registry', 1);
