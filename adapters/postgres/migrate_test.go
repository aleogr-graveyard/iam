package postgres_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/aleogr/iam/adapters/postgres"
	"github.com/aleogr/iam/internal/database"
	"github.com/aleogr/iam/internal/testdb"
)

func appliedVersions(t *testing.T, url string) []string {
	t.Helper()
	conn := connect(t, url)
	rows, err := conn.Query(t.Context(), "SELECT version FROM iam.schema_migrations ORDER BY version")
	if err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	got, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	return got
}

func embeddedVersions(t *testing.T) []string {
	t.Helper()
	v, err := postgres.MigrationVersions(postgres.EmbeddedMigrations)
	if err != nil {
		t.Fatalf("embedded migrations: %v", err)
	}
	if len(v) == 0 {
		t.Fatal("no embedded migrations")
	}
	return v
}

func TestMigrate_IsIdempotent(t *testing.T) {
	db := migratedDB(t)
	conn := connect(t, db.AdminURL)
	var before string
	if err := conn.QueryRow(t.Context(), "SELECT string_agg(version || applied_at::text, ',') FROM iam.schema_migrations").Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err := migrate(t, db.AdminURL, db.AppRole); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
	var after string
	if err := conn.QueryRow(t.Context(), "SELECT string_agg(version || applied_at::text, ',') FROM iam.schema_migrations").Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatalf("second run changed schema_migrations:\n before %s\n after  %s", before, after)
	}
	if got := appliedVersions(t, db.AdminURL); !slices.Equal(got, embeddedVersions(t)) {
		t.Fatalf("applied %v, want %v", got, embeddedVersions(t))
	}
}

func TestMigrate_ConcurrentRunsSerialize(t *testing.T) {
	db := testdb.NewWithAppRole(t)
	const runs = 2
	errs := make([]error, runs)
	pools := make([]func() error, runs)
	for i := range runs {
		pool, err := database.OpenMigrator(t.Context(), db.AdminURL, testdb.Limits())
		if err != nil {
			t.Fatalf("OpenMigrator: %v", err)
		}
		defer pool.Close()
		pools[i] = func() error { return postgres.Migrate(t.Context(), pool, migrateConfig(db.AppRole)) }
	}
	var wg sync.WaitGroup
	for i := range runs {
		wg.Go(func() { errs[i] = pools[i]() })
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("run %d: %v", i, err)
		}
	}
	if got := appliedVersions(t, db.AdminURL); !slices.Equal(got, embeddedVersions(t)) {
		t.Fatalf("applied %v, want each of %v once", got, embeddedVersions(t))
	}
}

// Names are checked before any statement, so no database is needed.
func TestMigrate_RejectsBadRoleName(t *testing.T) {
	for _, name := range []string{
		"", "iam", "app", "postgres", "IAM_APP", "iam-app", "iam app", `iam_app"`, "iam_app; DROP TABLE x",
		"iam_ä", "iam_" + strings.Repeat("a", 60), "1iam_app", "pg_iam_app",
	} {
		if err := postgres.Migrate(t.Context(), nil, migrateConfig(name)); !errors.Is(err, postgres.ErrInvalidAppRole) {
			t.Errorf("Migrate(%q) = %v, want ErrInvalidAppRole", name, err)
		}
	}
}

func TestMigrate_SchemaMigrationsInIamSchema(t *testing.T) {
	db := migratedDB(t)
	var schema string
	err := connect(t, db.AdminURL).QueryRow(t.Context(),
		"SELECT table_schema FROM information_schema.tables WHERE table_name = 'schema_migrations'").Scan(&schema)
	if err != nil || schema != "iam" {
		t.Fatalf("schema_migrations in %q (%v), want iam", schema, err)
	}
}

func TestMigrations_AreTimestampNamed(t *testing.T) {
	for _, v := range embeddedVersions(t) {
		if len(v) < 15 || v[14] != '_' {
			t.Errorf("version %q is not <yyyymmddhhmmss>_<name>", v)
		}
	}
	for name, files := range map[string]fstest.MapFS{
		"no timestamp":        {"migrations/tenancy.sql": {Data: []byte("SELECT 1;")}},
		"short timestamp":     {"migrations/2026_tenancy.sql": {Data: []byte("SELECT 1;")}},
		"upper case":          {"migrations/20261107000000_Tenancy.sql": {Data: []byte("SELECT 1;")}},
		"unknown placeholder": {"migrations/20261107000000_x.sql": {Data: []byte("GRANT SELECT ON t TO {{other}};")}},
	} {
		if _, err := postgres.MigrationVersions(files); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	got, err := postgres.MigrationVersions(fstest.MapFS{
		"migrations/20261108000000_b.sql": {Data: []byte("SELECT 1;")},
		"migrations/20261107000000_a.sql": {Data: []byte("GRANT SELECT ON t TO {{app_role}};")},
	})
	if err != nil || !slices.Equal(got, []string{"20261107000000_a", "20261108000000_b"}) {
		t.Errorf("versions = %v, %v; want name order", got, err)
	}
}

// TestMigrate_RefusesEditedMigration: every applied version keeps the
// checksum of its file, and a run against a database where an applied
// migration differs from the embedded one fails.
func TestMigrate_RefusesEditedMigration(t *testing.T) {
	db := migratedDB(t)
	var n int
	err := connect(t, db.AdminURL).QueryRow(t.Context(),
		"SELECT count(*) FROM iam.schema_migrations WHERE length(checksum) = 32").Scan(&n)
	if err != nil || n != len(embeddedVersions(t)) {
		t.Fatalf("%d versions carry a SHA-256 checksum (%v), want %d", n, err, len(embeddedVersions(t)))
	}
	adminExec(t, db, `UPDATE iam.schema_migrations SET checksum = '\x00'`)
	if err := migrate(t, db.AdminURL, db.AppRole); !errors.Is(err, postgres.ErrMigrationChanged) {
		t.Fatalf("Migrate = %v, want ErrMigrationChanged", err)
	}
}

// TestMigrate_VerifiesAppRolePrivileges: after migrating, the app role must
// reach exactly the expected privileges; an extra grant or a different app
// role (IAM_DATABASE_APP_ROLE changed after the grants were made) fails.
func TestMigrate_VerifiesAppRolePrivileges(t *testing.T) {
	t.Run("extra grant", func(t *testing.T) {
		db := migratedDB(t)
		adminExec(t, db, "GRANT UPDATE ON iam.users TO "+quote(db.AppRole))
		if err := migrate(t, db.AdminURL, db.AppRole); !errors.Is(err, postgres.ErrInvalidAppRole) {
			t.Fatalf("Migrate = %v, want ErrInvalidAppRole", err)
		}
	})
	t.Run("grant through a member role", func(t *testing.T) {
		db := migratedDB(t)
		r := testdb.NewRole(t, db.AdminURL, "NOLOGIN")
		adminExec(t, db, "GRANT TRUNCATE ON iam.tenants TO "+quote(r), "GRANT "+quote(r)+" TO "+quote(db.AppRole))
		if err := migrate(t, db.AdminURL, db.AppRole); !errors.Is(err, postgres.ErrInvalidAppRole) {
			t.Fatalf("Migrate = %v, want ErrInvalidAppRole", err)
		}
	})
	// Grants that table privileges do not show: a column, a sequence, a
	// function (EXECUTE, here of a SECURITY DEFINER function, which runs with
	// its owner's rights) and CREATE on the schema.
	for name, grant := range map[string]func(role string) []string{
		"column grant":           func(role string) []string { return []string{"GRANT UPDATE (email) ON iam.users TO " + role} },
		"column grant to PUBLIC": func(string) []string { return []string{"GRANT UPDATE (email) ON iam.users TO PUBLIC"} },
		"sequence grant": func(role string) []string {
			return []string{"CREATE SEQUENCE iam.s", "GRANT USAGE ON SEQUENCE iam.s TO " + role}
		},
		"public security definer function": func(string) []string {
			return []string{
				"CREATE FUNCTION iam.f() RETURNS bigint LANGUAGE sql SECURITY DEFINER AS 'SELECT count(*) FROM iam.users'",
				"GRANT EXECUTE ON FUNCTION iam.f() TO PUBLIC",
			}
		},
		"schema create": func(role string) []string { return []string{"GRANT CREATE ON SCHEMA iam TO " + role} },
	} {
		t.Run(name, func(t *testing.T) {
			db := migratedDB(t)
			adminExec(t, db, grant(quote(db.AppRole))...)
			if err := migrate(t, db.AdminURL, db.AppRole); !errors.Is(err, postgres.ErrInvalidAppRole) {
				t.Fatalf("Migrate = %v, want ErrInvalidAppRole", err)
			}
		})
	}
	// A function the migrator adds later is not executable by PUBLIC: the
	// first migration revoked EXECUTE in the migrator's default privileges.
	t.Run("new function is not public", func(t *testing.T) {
		db := migratedDB(t)
		adminExec(t, db, "CREATE FUNCTION iam.g() RETURNS int LANGUAGE sql AS 'SELECT 1'")
		if err := migrate(t, db.AdminURL, db.AppRole); err != nil {
			t.Fatalf("Migrate = %v, want nil", err)
		}
		var can bool
		if err := connect(t, db.AppURL).QueryRow(t.Context(), "SELECT has_function_privilege('iam.g()', 'EXECUTE')").Scan(&can); err != nil || can {
			t.Fatalf("app role can execute a new function: %v (%v)", can, err)
		}
	})
	t.Run("changed app role", func(t *testing.T) {
		db := migratedDB(t)
		other := testdb.NewRole(t, db.AdminURL, "NOLOGIN")
		if err := migrate(t, db.AdminURL, other); !errors.Is(err, postgres.ErrInvalidAppRole) {
			t.Fatalf("Migrate with another role = %v, want ErrInvalidAppRole", err)
		}
	})
}

// TestMigrate_LockTimeout: a migration that cannot get the migration lock
// gives up after the configured lock timeout instead of waiting forever.
func TestMigrate_LockTimeout(t *testing.T) {
	db := migratedDB(t)
	holder := connect(t, db.AdminURL)
	if _, err := holder.Exec(t.Context(), "SELECT pg_advisory_lock($1)", int64(postgres.MigrationLockKey)); err != nil {
		t.Fatalf("hold the lock: %v", err)
	}
	// The pool's own lock_timeout is long, so only the per-transaction
	// setting can end the wait in time.
	limits := testdb.Limits()
	limits.LockTimeout = time.Minute
	pool, err := database.OpenMigrator(t.Context(), db.AdminURL, limits)
	if err != nil {
		t.Fatalf("OpenMigrator: %v", err)
	}
	defer pool.Close()
	cfg := migrateConfig(db.AppRole)
	cfg.LockTimeout = 200 * time.Millisecond
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	start := time.Now()
	err = postgres.Migrate(ctx, pool, cfg)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "55P03" {
		t.Fatalf("Migrate = %v, want lock_not_available (55P03)", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("Migrate gave up after %v", elapsed)
	}
}

func TestMigrate_RefusesMissingTimeouts(t *testing.T) {
	cfg := migrateConfig("iam_app")
	cfg.StatementTimeout = 0
	if err := postgres.Migrate(t.Context(), nil, cfg); err == nil {
		t.Fatal("Migrate accepted a zero statement timeout")
	}
}

// TestMigrate_StatementTimeout: a migration statement longer than the
// configured statement timeout is cancelled. The pool's own limits are long,
// so only the per-transaction setting can end it; the lock wait is the long
// statement, and its own timeout is longer than the statement timeout.
func TestMigrate_StatementTimeout(t *testing.T) {
	db := migratedDB(t)
	holder := connect(t, db.AdminURL)
	if _, err := holder.Exec(t.Context(), "SELECT pg_advisory_lock($1)", int64(postgres.MigrationLockKey)); err != nil {
		t.Fatalf("hold the lock: %v", err)
	}
	limits := testdb.Limits()
	limits.StatementTimeout, limits.LockTimeout = time.Minute, time.Minute
	pool, err := database.OpenMigrator(t.Context(), db.AdminURL, limits)
	if err != nil {
		t.Fatalf("OpenMigrator: %v", err)
	}
	defer pool.Close()
	cfg := migrateConfig(db.AppRole)
	cfg.StatementTimeout, cfg.LockTimeout = 200*time.Millisecond, time.Minute
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	err = postgres.Migrate(ctx, pool, cfg)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "57014" {
		t.Fatalf("Migrate = %v, want query_canceled (57014)", err)
	}
}
