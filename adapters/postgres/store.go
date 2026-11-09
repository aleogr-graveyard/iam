// Package postgres is the Postgres storage adapter (contract C3) and the
// owner of the database schema (Migrate, migrations/*.sql).
//
// # Two barriers
//
// Every tenant-owned query filters by tenant_id explicitly, and row level
// security repeats the rule in the database: InTenant sets the
// transaction-local variable iam.tenant_id, and the policy of every
// tenant-owned table (FORCE ROW LEVEL SECURITY, so the owner is bound too)
// admits only rows of that tenant, on reads and on writes (WITH CHECK). The
// service connects as a role that owns nothing and cannot bypass RLS
// (Migrate refuses any other); either barrier alone keeps tenants apart.
//
// # Tenant setting
//
// InTenant checks that the tenant exists and sets iam.tenant_id in one
// statement, at the start of its transaction. InSystem always overwrites
// iam.tenant_id with the system tenant's ID, so the system's own audit
// appends (Task 10) obey RLS; before EnsureSystemTenant has created the
// system tenant it sets it to the empty string, which the policies read as
// "no tenant" (ruling R1). The setting is transaction-local (set_config with
// is_local), so it never survives on a pooled connection.
//
// # Transactions
//
// Each InTenant and InSystem is one database transaction at READ COMMITTED
// on one pool connection. Every write runs inside a savepoint, so that a
// refused write (a conflict, a missing organization) leaves the transaction
// usable, as the contract requires. A unique violation is ErrConflict and a
// foreign key violation ErrNotFound; errors never carry driver messages,
// constraint values or the connection URL. Rows read back are validated, and
// a row that fails is ErrCorrupt, never skipped.
package postgres

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

// rollbackTimeout bounds the rollback of a transaction whose ctx has ended.
const rollbackTimeout = 5 * time.Second

// Store is safe for concurrent use. It never opens connections of its own:
// every transaction takes one connection of the pool it was given.
type Store struct {
	pool              *pgxpool.Pool
	allowInsecureHTTP bool
}

var _ storage.Store = (*Store)(nil)

// Option configures a Store.
type Option func(*Store)

// WithAllowInsecureHTTP sets whether applications may have http redirect URIs
// other than loopback (tenancy.Application.Validate), on create and when rows
// are read back. It comes from configuration and is for development only; the
// default is false.
func WithAllowInsecureHTTP(allow bool) Option {
	return func(s *Store) { s.allowInsecureHTTP = allow }
}

// NewStore returns a Store over pool, which must connect as the service's
// role (the app role given to Migrate) or a login role that is a member of
// it. It fails closed (ErrInvalidAppRole) when that role is, or can become,
// a superuser, a BYPASSRLS role or an owner of anything in the iam schema:
// such a role would get around row level security, whatever Migrate checked.
func NewStore(ctx context.Context, pool *pgxpool.Pool, opts ...Option) (*Store, error) {
	if err := checkConnectionRole(ctx, pool); err != nil {
		return nil, err
	}
	s := &Store{pool: pool}
	for _, o := range opts {
		o(s)
	}
	return s, nil
}

// setTenantSQL sets iam.tenant_id only when the tenant exists: no row means
// an unknown tenant.
const setTenantSQL = `SELECT set_config('iam.tenant_id', id::text, true) FROM iam.tenants WHERE id = $1`

// setSystemSQL sets iam.tenant_id to the system tenant, or to ” (no tenant)
// before it exists, overwriting whatever the connection held.
const setSystemSQL = `SELECT set_config('iam.tenant_id', coalesce((SELECT id::text FROM iam.tenants WHERE slug = $1), ''), true)`

// InTenant implements storage.Store.
func (s *Store) InTenant(ctx context.Context, tenant ids.TenantID, fn func(context.Context, storage.Tx) error) error {
	txCtx, err := begin(ctx)
	if err != nil {
		return err
	}
	if tenant.IsZero() {
		return storage.ErrInvalidTenant
	}
	return s.run(txCtx, func(ctx context.Context, pgTx pgx.Tx) error {
		var set string
		err := pgTx.QueryRow(ctx, setTenantSQL, uuidArg(tenant.ID)).Scan(&set)
		if errors.Is(err, pgx.ErrNoRows) {
			return storage.ErrInvalidTenant
		}
		if err != nil {
			return sanitize(err)
		}
		tx := &tenantTx{txConn: txConn{tx: pgTx}, tenant: tenant, allowInsecureHTTP: s.allowInsecureHTTP}
		defer tx.finish()
		return fn(ctx, tx)
	})
}

// InSystem implements storage.Store.
func (s *Store) InSystem(ctx context.Context, fn func(context.Context, storage.SystemTx) error) error {
	txCtx, err := begin(ctx)
	if err != nil {
		return err
	}
	return s.run(txCtx, func(ctx context.Context, pgTx pgx.Tx) error {
		var set string
		if err := pgTx.QueryRow(ctx, setSystemSQL, tenancy.SystemTenantSlug).Scan(&set); err != nil {
			return sanitize(err)
		}
		tx := &systemTx{txConn: txConn{tx: pgTx}}
		defer tx.finish()
		return fn(ctx, tx)
	})
}

// begin refuses a canceled ctx and a nested transaction.
func begin(ctx context.Context) (context.Context, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return storage.EnterTx(ctx)
}

// run opens a transaction, runs body and commits when body returns nil and
// ctx is still live. On an error the transaction rolls back and body's error
// is returned unchanged; on a panic it rolls back and the panic propagates.
func (s *Store) run(ctx context.Context, body func(context.Context, pgx.Tx) error) error {
	pgTx, err := s.pool.Begin(ctx)
	if err != nil {
		return sanitize(err)
	}
	defer func() {
		// A no-op after Commit. The ctx may have ended: roll back without it.
		rbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), rollbackTimeout)
		defer cancel()
		_ = pgTx.Rollback(rbCtx)
	}()
	if err := body(ctx, pgTx); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return sanitize(pgTx.Commit(ctx))
}

// txConn is what the tenant and system transactions share: the database
// transaction and the end-of-fn marker.
type txConn struct {
	done atomic.Bool
	tx   pgx.Tx
}

func (c *txConn) finish() { c.done.Store(true) }

// check is called first by every method of a Tx or SystemTx.
func (c *txConn) check(ctx context.Context) error {
	if c.done.Load() {
		return storage.ErrTxDone
	}
	return ctx.Err()
}

// write runs one statement inside a savepoint, in a single round trip when
// it succeeds. A refused statement is rolled back to the savepoint, so the
// transaction stays usable, and mapped by writeError.
func (c *txConn) write(ctx context.Context, entity, sql string, args ...any) (pgconn.CommandTag, error) {
	if err := c.check(ctx); err != nil {
		return pgconn.CommandTag{}, err
	}
	var tag pgconn.CommandTag
	b := &pgx.Batch{}
	b.Queue("SAVEPOINT iam_write")
	b.Queue(sql, args...).Exec(func(ct pgconn.CommandTag) error { tag = ct; return nil })
	b.Queue("RELEASE SAVEPOINT iam_write")
	err := c.tx.SendBatch(ctx, b).Close()
	if err == nil {
		return tag, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if _, rbErr := c.tx.Exec(ctx, "ROLLBACK TO SAVEPOINT iam_write; RELEASE SAVEPOINT iam_write"); rbErr != nil {
			return tag, sanitize(rbErr)
		}
	}
	return tag, writeError(entity, err)
}

// scanner reads one row into a validated value.
type scanner[T any] func(pgx.Row) (T, error)

// queryOne runs a query expected to return at most one row.
func queryOne[T any](ctx context.Context, c *txConn, entity string, scan scanner[T], sql string, args ...any) (T, error) {
	var zero T
	if err := c.check(ctx); err != nil {
		return zero, err
	}
	v, err := scan(c.tx.QueryRow(ctx, sql, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return zero, notFound(entity)
	}
	if err != nil {
		return zero, sanitize(err)
	}
	return v, nil
}

// queryAll runs a query and scans every row; any invalid row fails the call.
func queryAll[T any](ctx context.Context, c *txConn, scan scanner[T], sql string, args ...any) ([]T, error) {
	if err := c.check(ctx); err != nil {
		return nil, err
	}
	rows, err := c.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, sanitize(err)
	}
	defer rows.Close()
	out := []T{}
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, sanitize(err)
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, sanitize(err)
	}
	return out, nil
}

// uuidArg passes an ID to a query; the zero ID is a valid UUID that names no
// row, so a lookup with it finds nothing.
func uuidArg(id ids.ID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }

// idDest scans a non-null uuid column into an ID.
type idDest struct{ id *ids.ID }

func (d idDest) ScanUUID(v pgtype.UUID) error {
	if !v.Valid {
		return errors.New("postgres: null uuid")
	}
	*d.id = v.Bytes
	return nil
}

// utcDest scans a timestamptz into a time.Time in UTC.
type utcDest struct{ t *time.Time }

func (d utcDest) ScanTimestamptz(v pgtype.Timestamptz) error {
	if !v.Valid || v.InfinityModifier != pgtype.Finite {
		return errors.New("postgres: null or infinite timestamp")
	}
	*d.t = v.Time.UTC()
	return nil
}
