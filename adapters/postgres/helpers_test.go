package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aleogr/iam/adapters/postgres"
	"github.com/aleogr/iam/internal/database"
	"github.com/aleogr/iam/internal/testdb"
)

// migratedDB returns a fresh database migrated as its admin user, with its
// own non-owner app role.
func migratedDB(t *testing.T) testdb.AppDB {
	t.Helper()
	db := testdb.NewWithAppRole(t)
	pool, err := database.OpenMigrator(t.Context(), db.AdminURL)
	if err != nil {
		t.Fatalf("OpenMigrator: %v", err)
	}
	defer pool.Close()
	if err := postgres.Migrate(t.Context(), pool, db.AppRole); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return db
}

// appPool opens a pool as the app role, sized as the service's (database.Open).
func appPool(t *testing.T, db testdb.AppDB) *pgxpool.Pool {
	t.Helper()
	pool, err := database.Open(t.Context(), db.AppURL)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// singleConnPool opens a pool of one connection as the app role, so that
// consecutive transactions are known to share the connection.
func singleConnPool(t *testing.T, db testdb.AppDB) *pgxpool.Pool {
	t.Helper()
	cfg, err := pgxpool.ParseConfig(db.AppURL)
	if err != nil {
		t.Fatalf("parse app URL: %v", err)
	}
	cfg.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(t.Context(), cfg)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// connect opens one connection to url.
func connect(t *testing.T, url string) *pgx.Conn {
	t.Helper()
	conn, err := pgx.Connect(t.Context(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}
