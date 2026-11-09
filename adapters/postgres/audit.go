package postgres

import (
	"context"
	"encoding/binary"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/aleogr/iam/shared/audit"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
)

// Appends to one chain are serialized by a transaction-level advisory lock
// on the tenant, taken before the newest event is read and held until the
// transaction ends, so a second append waits for the first to commit (and
// then, at READ COMMITTED, reads its row) or roll back. The alternatives do
// not fit the append-only grants: SELECT ... FOR UPDATE on the newest event
// needs the UPDATE privilege and has no row to lock for a chain's first
// event, and a chain-head row needs UPDATE on a mutable table. The lock needs
// no privilege, covers the first event and touches no table; UNIQUE
// (tenant_id, sequence) remains the backstop, turning a fork into
// ErrConflict.
//
// The lock is the two-key form pg_advisory_xact_lock(auditLockClass, key),
// whose key space never overlaps the one-key form of MigrationLockKey. The key
// is 32 random bits of the tenant's UUIDv7; two tenants that share it only
// serialize their appends, which stays correct.
const auditLockClass int32 = 0x69616d61 // "iama"

func auditLockKey(tenant ids.TenantID) int32 {
	return int32(binary.BigEndian.Uint32(tenant.ID[12:])) //nolint:gosec // reinterpreting 32 bits; any value is a valid key
}

// auditRepo is the audit chain of one tenant within a transaction, the system
// tenant's in InSystem.
type auditRepo struct {
	conn   *txConn
	tenant ids.TenantID // zero: no chain (InSystem before the system tenant)
}

const auditColumns = "id, tenant_id, sequence, occurred_at, schema_version, action, actor_type, actor_id, target, outcome, data, prev_hash, hash"

const lastEventSQL = `SELECT sequence, hash FROM iam.audit_events WHERE tenant_id = $1 ORDER BY sequence DESC LIMIT 1`

func (r auditRepo) Append(ctx context.Context, e audit.Event) (audit.Event, error) {
	if err := r.conn.check(ctx); err != nil {
		return audit.Event{}, err
	}
	if r.tenant.IsZero() {
		return audit.Event{}, storage.ErrInvalidTenant
	}
	if _, err := r.conn.tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1, $2)", auditLockClass, auditLockKey(r.tenant)); err != nil {
		return audit.Event{}, sanitize(err)
	}
	var seq int64
	var prev []byte
	err := r.conn.tx.QueryRow(ctx, lastEventSQL, uuidArg(r.tenant.ID)).Scan(&seq, &prev)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return audit.Event{}, sanitize(err)
	}
	sealed, err := storage.SealAudit(r.tenant, seq+1, prev, e)
	if err != nil {
		return audit.Event{}, err
	}
	data := sealed.Data
	if data == nil {
		data = map[string]string{}
	}
	_, err = r.conn.write(ctx, "audit event",
		"INSERT INTO iam.audit_events ("+auditColumns+") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
		uuidArg(sealed.ID.ID), uuidArg(sealed.TenantID.ID), sealed.Sequence, sealed.Time, sealed.SchemaVersion,
		sealed.Action, sealed.Actor.Type, sealed.Actor.ID, sealed.Target, sealed.Outcome, data,
		nonNil(sealed.PrevHash), sealed.Hash)
	if err != nil {
		return audit.Event{}, err
	}
	return sealed, nil
}

// nonNil stores an empty PrevHash as an empty bytea, never NULL.
func nonNil(b []byte) []byte {
	if b == nil {
		return []byte{}
	}
	return b
}

func (r auditRepo) Events(ctx context.Context, after int64, limit int) ([]audit.Event, error) {
	if err := r.conn.check(ctx); err != nil {
		return nil, err
	}
	n, err := storage.AuditPage(after, limit)
	if err != nil {
		return nil, err
	}
	return queryAll(ctx, r.conn, r.scan,
		"SELECT "+auditColumns+" FROM iam.audit_events WHERE tenant_id = $1 AND sequence > $2 ORDER BY sequence LIMIT $3",
		uuidArg(r.tenant.ID), after, n)
}

// scan reads one event and checks it as scanRow checks entities: valid
// (audit.Event.Validate; the hash itself is for audit.VerifyChain) and of the
// chain's tenant, or ErrCorrupt.
func (r auditRepo) scan(row pgx.Row) (audit.Event, error) {
	var e audit.Event
	err := row.Scan(idDest{&e.ID.ID}, idDest{&e.TenantID.ID}, &e.Sequence, utcDest{&e.Time}, &e.SchemaVersion,
		&e.Action, &e.Actor.Type, &e.Actor.ID, &e.Target, &e.Outcome, &e.Data, &e.PrevHash, &e.Hash)
	if err != nil {
		return audit.Event{}, err
	}
	if len(e.Data) == 0 {
		e.Data = nil
	}
	if len(e.PrevHash) == 0 {
		e.PrevHash = nil
	}
	if err := e.Validate(); err != nil {
		return audit.Event{}, corrupt("audit event", err)
	}
	if e.TenantID != r.tenant {
		return audit.Event{}, corrupt("audit event", errForeignRow)
	}
	return e, nil
}
