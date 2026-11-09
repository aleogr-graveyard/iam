package postgres_test

import (
	"context"
	"testing"
	"time"

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
	if err := migrate(t, db.AdminURL, db.AppRole); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return db
}

// migrateConfig is the migration configuration of the tests.
func migrateConfig(role string) postgres.MigrateConfig {
	return postgres.MigrateConfig{AppRole: role, StatementTimeout: time.Minute, LockTimeout: 10 * time.Second}
}

func migrate(t *testing.T, url, role string) error {
	t.Helper()
	pool, err := database.OpenMigrator(t.Context(), url, testdb.Limits())
	if err != nil {
		t.Fatalf("OpenMigrator: %v", err)
	}
	defer pool.Close()
	return postgres.Migrate(t.Context(), pool, migrateConfig(role))
}

// newStore is NewStore with its startup check required to pass.
func newStore(t *testing.T, pool *pgxpool.Pool, opts ...postgres.Option) *postgres.Store {
	t.Helper()
	s, err := postgres.NewStore(t.Context(), pool, opts...)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

// adminExec runs statements on db as its admin (a superuser). It does not use
// t.Context(), so cleanups may call it.
func adminExec(t *testing.T, db testdb.AppDB, sqls ...string) {
	t.Helper()
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, db.AdminURL)
	if err != nil {
		t.Fatalf("admin: connect: %v", err)
	}
	defer func() { _ = conn.Close(ctx) }()
	for _, sql := range sqls {
		if _, err := conn.Exec(ctx, sql); err != nil {
			t.Fatalf("admin: %s: %v", sql, err)
		}
	}
}

func quote(name string) string { return pgx.Identifier{name}.Sanitize() }

// appPool opens a pool as the app role, sized as the service's (database.Open).
func appPool(t *testing.T, db testdb.AppDB) *pgxpool.Pool {
	t.Helper()
	pool, err := database.Open(t.Context(), db.AppURL, testdb.Limits())
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
