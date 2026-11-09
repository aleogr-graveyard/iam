// Package database opens the service's Postgres connection pools, sized by the
// connection budget (budget.go).
package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aleogr/iam/internal/httpserver"
)

// MaxConns is the per-instance pool of the service on the shared lab Postgres
// (ADR 0001).
const MaxConns = ServicePool

const connectTimeout = 5 * time.Second

// PoolKind selects the size of a pool from the connection budget.
type PoolKind int

// Pool kinds.
const (
	// ServicePoolKind is the pool of the running service: ServicePool connections.
	ServicePoolKind PoolKind = iota
	// MigrationPoolKind is the pool of the migration job: MigrationJob connections.
	MigrationPoolKind
)

var errInvalidURL = errors.New("database: invalid connection URL")

// Limits bound what one connection may hold (shared/config DatabaseLimits).
// Every field is required: Postgres reads 0 as "no limit".
type Limits struct {
	StatementTimeout         time.Duration // statement_timeout
	IdleInTransactionTimeout time.Duration // idle_in_transaction_session_timeout
	LockTimeout              time.Duration // lock_timeout
	MaxConnIdleTime          time.Duration // the pool closes a connection idle this long
}

func (l Limits) validate() error {
	for _, d := range []time.Duration{l.StatementTimeout, l.IdleInTransactionTimeout, l.LockTimeout, l.MaxConnIdleTime} {
		if d < time.Millisecond {
			return errors.New("database: every limit must be at least 1ms")
		}
	}
	return nil
}

// PoolConfig parses url, sizes the pool for kind and applies limits as
// session settings of every connection. The error is a fixed text: it never
// carries the URL, so credentials cannot leak through it.
func PoolConfig(url string, kind PoolKind, limits Limits) (*pgxpool.Config, error) {
	if err := limits.validate(); err != nil {
		return nil, err
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, errInvalidURL
	}
	switch kind {
	case ServicePoolKind:
		cfg.MaxConns = ServicePool
	case MigrationPoolKind:
		cfg.MaxConns = MigrationJob
	default:
		return nil, errors.New("database: unknown pool kind")
	}
	cfg.MinConns = 0
	cfg.MaxConnIdleTime = limits.MaxConnIdleTime
	cfg.ConnConfig.ConnectTimeout = connectTimeout
	params := cfg.ConnConfig.RuntimeParams
	params["statement_timeout"] = milliseconds(limits.StatementTimeout)
	params["idle_in_transaction_session_timeout"] = milliseconds(limits.IdleInTransactionTimeout)
	params["lock_timeout"] = milliseconds(limits.LockTimeout)
	return cfg, nil
}

func milliseconds(d time.Duration) string { return strconv.FormatInt(d.Milliseconds(), 10) }

// Open opens the service pool (ServicePool connections) and verifies the
// database answers. Errors never carry the URL.
func Open(ctx context.Context, url string, limits Limits) (*pgxpool.Pool, error) {
	return open(ctx, url, ServicePoolKind, limits)
}

// OpenMigrator opens the migration job's pool (MigrationJob connections) and
// verifies the database answers. Errors never carry the URL.
func OpenMigrator(ctx context.Context, url string, limits Limits) (*pgxpool.Pool, error) {
	return open(ctx, url, MigrationPoolKind, limits)
}

func open(ctx context.Context, url string, kind PoolKind, limits Limits) (*pgxpool.Pool, error) {
	cfg, err := PoolConfig(url, kind, limits)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("database: create pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}
	return pool, nil
}

// ReadinessCheck reports the pool as ready while a single ping succeeds. The
// caller (/readyz) bounds it with the readiness timeout.
func ReadinessCheck(pool *pgxpool.Pool) httpserver.ReadinessCheck {
	return httpserver.ReadinessCheck{
		Name:  "database",
		Check: func(ctx context.Context) error { return pool.Ping(ctx) },
	}
}
