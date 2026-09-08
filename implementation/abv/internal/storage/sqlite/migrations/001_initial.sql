PRAGMA foreign_keys = ON;

CREATE TABLE abv_metadata (
    marker TEXT PRIMARY KEY CHECK (marker = 'agentlabs-abv'),
    schema_version INTEGER NOT NULL CHECK (schema_version = 1)
);

CREATE TABLE applications (
    application_id TEXT PRIMARY KEY,
    compatibility_enabled INTEGER NOT NULL CHECK (compatibility_enabled IN (0, 1))
);
CREATE TABLE installations (
    tenant_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    PRIMARY KEY (tenant_id, application_id),
    FOREIGN KEY (application_id) REFERENCES applications(application_id)
);
CREATE TABLE permissions (
    application_id TEXT NOT NULL,
    permission_id TEXT NOT NULL,
    active INTEGER NOT NULL CHECK (active IN (0, 1)),
    PRIMARY KEY (application_id, permission_id),
    FOREIGN KEY (application_id) REFERENCES applications(application_id)
);
CREATE TABLE scope_definitions (
    application_id TEXT NOT NULL,
    scope_key TEXT NOT NULL,
    allowed_tokens_json BLOB NOT NULL,
    PRIMARY KEY (application_id, scope_key),
    FOREIGN KEY (application_id) REFERENCES applications(application_id)
);
CREATE TABLE supported_scope_keys (
    application_id TEXT NOT NULL,
    permission_id TEXT NOT NULL,
    scope_key TEXT NOT NULL,
    ordinal INTEGER NOT NULL CHECK (ordinal >= 0),
    PRIMARY KEY (application_id, permission_id, scope_key),
    UNIQUE (application_id, permission_id, ordinal),
    FOREIGN KEY (application_id, permission_id) REFERENCES permissions(application_id, permission_id),
    FOREIGN KEY (application_id, scope_key) REFERENCES scope_definitions(application_id, scope_key)
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
INSERT INTO abv_metadata(marker, schema_version) VALUES ('agentlabs-abv', 1);
