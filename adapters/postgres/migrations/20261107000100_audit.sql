-- Audit: the per-tenant, hash-chained, append-only audit log (contract C4,
-- ADR 0005). Expand only: a new table and its grants.
--
-- One chain per tenant, numbered by sequence from 1: UNIQUE (tenant_id,
-- sequence) refuses a fork even if two appends ever got past the adapter's
-- per-chain lock. Event IDs are unique across tenants (ruling R6): PRIMARY
-- KEY (id). The hash covers every column but hash itself (package audit), so
-- an edit made behind the adapter's back is detected by audit.VerifyChain.
--
-- prev_hash is empty (not NULL) for sequence 1, the genesis, and 32 bytes
-- otherwise. data is a JSON object of short strings (audit.Event.Data).
--
-- Row level security is the second barrier, as on the tenancy tables; for the
-- system tenant's chain, InSystem sets iam.tenant_id to the system tenant
-- (ruling R1).

CREATE TABLE iam.audit_events (
    id             uuid        PRIMARY KEY,
    tenant_id      uuid        NOT NULL REFERENCES iam.tenants (id),
    sequence       bigint      NOT NULL CHECK (sequence >= 1),
    occurred_at    timestamptz NOT NULL,
    schema_version integer     NOT NULL,
    action         text        NOT NULL,
    actor_type     text        NOT NULL,
    actor_id       text        NOT NULL,
    target         text        NOT NULL,
    outcome        text        NOT NULL,
    data           jsonb       NOT NULL CHECK (jsonb_typeof(data) = 'object'),
    prev_hash      bytea       NOT NULL,
    hash           bytea       NOT NULL CHECK (length(hash) = 32),
    CONSTRAINT audit_events_tenant_id_sequence_key UNIQUE (tenant_id, sequence),
    CONSTRAINT audit_events_prev_hash_check CHECK (
        (sequence = 1 AND length(prev_hash) = 0) OR (sequence > 1 AND length(prev_hash) = 32))
);

ALTER TABLE iam.audit_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE iam.audit_events FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON iam.audit_events
    USING (tenant_id = NULLIF(current_setting('iam.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('iam.tenant_id', true), '')::uuid);

-- Append-only for the service: read and insert, never update, delete or
-- truncate (appPrivileges in roles.go).
GRANT SELECT, INSERT ON iam.audit_events TO {{app_role}};
