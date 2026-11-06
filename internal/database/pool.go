// Package database opens the service's Postgres connection pool.
package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aleogr/iam/internal/httpserver"
)

// MaxConns is the per-instance connection budget on the shared lab Postgres
// (ADR 0001).
const MaxConns int32 = 2

const connectTimeout = 5 * time.Second

// Open parses url, caps the pool at MaxConns and verifies the database answers.
// Errors never carry the URL, so credentials cannot leak through them.
func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, errors.New("database: invalid connection URL")
	}
	cfg.MaxConns = MaxConns
	cfg.ConnConfig.ConnectTimeout = connectTimeout

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
