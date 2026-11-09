-- Tenancy: the tenant registry and the tenant-owned tables of contract C3.
--
-- Applied by Migrate as the migration role, which owns everything created
-- here. {{app_role}} is replaced with the quoted name of the service's role,
-- which gets the grants at the end (Migrate grants USAGE on the schema) and
-- owns nothing.
-- This is the first migration; every later one must be backward compatible
-- with the running code (expand/contract).
--
-- IDs are unique across tenants (ruling R6): PRIMARY KEY (id), plus
-- UNIQUE (tenant_id, id) as the target of the tenant-scoped composite foreign
-- keys, so that a membership cannot link rows of two tenants.
--
-- Row level security is the second barrier: every query of the adapter also
-- filters by tenant_id. FORCE makes the policies apply to the table owner too.
-- The policy compares with NULLIF(..., ''): once a transaction-local
-- set_config has ended, the setting reads as '' on that connection, and a
-- connection without a tenant must see no row rather than fail on ''::uuid.

CREATE TABLE iam.tenants (
    id         uuid        PRIMARY KEY,
    slug       text        NOT NULL,
    name       text        NOT NULL,
    status     text        NOT NULL CHECK (status IN ('active', 'suspended')),
    created_at timestamptz NOT NULL,
    CONSTRAINT tenants_slug_key UNIQUE (slug)
);

CREATE TABLE iam.organizations (
    id         uuid        PRIMARY KEY,
    tenant_id  uuid        NOT NULL REFERENCES iam.tenants (id),
    slug       text        NOT NULL,
    name       text        NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT organizations_tenant_id_id_key UNIQUE (tenant_id, id),
    CONSTRAINT organizations_tenant_id_slug_key UNIQUE (tenant_id, slug)
);

CREATE TABLE iam.users (
    id             uuid        PRIMARY KEY,
    tenant_id      uuid        NOT NULL REFERENCES iam.tenants (id),
    email          text        NOT NULL,
    email_verified boolean     NOT NULL,
    status         text        NOT NULL CHECK (status IN ('active', 'suspended')),
    created_at     timestamptz NOT NULL,
    CONSTRAINT users_tenant_id_id_key UNIQUE (tenant_id, id),
    CONSTRAINT users_tenant_id_email_key UNIQUE (tenant_id, email)
);

CREATE TABLE iam.memberships (
    tenant_id       uuid        NOT NULL REFERENCES iam.tenants (id),
    organization_id uuid        NOT NULL,
    user_id         uuid        NOT NULL,
    created_at      timestamptz NOT NULL,
    CONSTRAINT memberships_pkey PRIMARY KEY (tenant_id, organization_id, user_id),
    CONSTRAINT memberships_organization_fkey FOREIGN KEY (tenant_id, organization_id)
        REFERENCES iam.organizations (tenant_id, id),
    CONSTRAINT memberships_user_fkey FOREIGN KEY (tenant_id, user_id)
        REFERENCES iam.users (tenant_id, id)
);

-- ListByUser, and the foreign key from memberships to users.
CREATE INDEX memberships_tenant_id_user_id_idx ON iam.memberships (tenant_id, user_id, organization_id);

CREATE TABLE iam.applications (
    id            uuid        PRIMARY KEY,
    tenant_id     uuid        NOT NULL REFERENCES iam.tenants (id),
    name          text        NOT NULL,
    type          text        NOT NULL CHECK (type IN ('public', 'confidential')),
    redirect_uris text[]      NOT NULL,
    secret_hash   bytea,
    created_at    timestamptz NOT NULL,
    CONSTRAINT applications_tenant_id_id_key UNIQUE (tenant_id, id),
    CONSTRAINT applications_secret_hash_check CHECK ((type = 'public') = (secret_hash IS NULL))
);

ALTER TABLE iam.organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam.organizations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam.organizations
    USING (tenant_id = NULLIF(current_setting('iam.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('iam.tenant_id', true), '')::uuid);

ALTER TABLE iam.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam.users FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam.users
    USING (tenant_id = NULLIF(current_setting('iam.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('iam.tenant_id', true), '')::uuid);

ALTER TABLE iam.memberships ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam.memberships FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam.memberships
    USING (tenant_id = NULLIF(current_setting('iam.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('iam.tenant_id', true), '')::uuid);

ALTER TABLE iam.applications ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam.applications FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam.applications
    USING (tenant_id = NULLIF(current_setting('iam.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('iam.tenant_id', true), '')::uuid);

-- Least privilege: exactly what the store uses (appPrivileges in roles.go,
-- which Migrate verifies after every run). Later features widen it with
-- their own expand migrations. The tenant registry has no tenant_id and no
-- row level security: it is reachable only through Store.InSystem, and the
-- app role can neither change nor delete its rows.
GRANT SELECT, INSERT ON iam.tenants, iam.organizations, iam.users, iam.applications TO {{app_role}};
GRANT SELECT, INSERT, DELETE ON iam.memberships TO {{app_role}};
