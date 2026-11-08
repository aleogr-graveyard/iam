// Package database opens the service's Postgres connection pools, sized by the
// connection budget (budget.go).
package database

import (
	"context"
	"errors"
	"fmt"
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

// PoolConfig parses url and sizes the pool for kind. The error is a fixed
// text: it never carries the URL, so credentials cannot leak through it.
func PoolConfig(url string, kind PoolKind) (*pgxpool.Config, error) {
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
	cfg.ConnConfig.ConnectTimeout = connectTimeout
	return cfg, nil
}

// Open opens the service pool (ServicePool connections) and verifies the
// database answers. Errors never carry the URL.
func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	return open(ctx, url, ServicePoolKind)
}

// OpenMigrator opens the migration job's pool (MigrationJob connections) and
// verifies the database answers. Errors never carry the URL.
func OpenMigrator(ctx context.Context, url string) (*pgxpool.Pool, error) {
	return open(ctx, url, MigrationPoolKind)
}

func open(ctx context.Context, url string, kind PoolKind) (*pgxpool.Pool, error) {
	cfg, err := PoolConfig(url, kind)
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
