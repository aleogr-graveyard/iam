package postgres_test

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5"

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

func migrate(t *testing.T, url, role string) error {
	t.Helper()
	pool, err := database.OpenMigrator(t.Context(), url)
	if err != nil {
		t.Fatalf("OpenMigrator: %v", err)
	}
	defer pool.Close()
	return postgres.Migrate(t.Context(), pool, role)
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
		pool, err := database.OpenMigrator(t.Context(), db.AdminURL)
		if err != nil {
			t.Fatalf("OpenMigrator: %v", err)
		}
		defer pool.Close()
		pools[i] = func() error { return postgres.Migrate(t.Context(), pool, db.AppRole) }
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
		if err := postgres.Migrate(t.Context(), nil, name); !errors.Is(err, postgres.ErrInvalidAppRole) {
			t.Errorf("Migrate(%q) = %v, want ErrInvalidAppRole", name, err)
		}
	}
}

// TestMigrate_RefusesPrivilegedAppRole: an app role that could get around
// RLS or the ownership rules is refused, and nothing is applied.
func TestMigrate_RefusesPrivilegedAppRole(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(role string) string
	}{
		{"missing", func(string) string { return "" }},
		{"superuser", func(role string) string { return "ALTER ROLE " + role + " SUPERUSER" }},
		{"bypassrls", func(role string) string { return "ALTER ROLE " + role + " BYPASSRLS" }},
		{"member of the migrator", func(role string) string {
			return "DO $$ BEGIN EXECUTE format('GRANT %I TO %s', current_user, '" + role + "'); END $$"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := testdb.NewWithAppRole(t)
			role := db.AppRole
			if sql := tc.setup(pgx.Identifier{role}.Sanitize()); sql != "" {
				if _, err := connect(t, db.AdminURL).Exec(t.Context(), sql); err != nil {
					t.Fatalf("setup: %v", err)
				}
			} else {
				role = fmt.Sprintf("iam_test_app_%d_missing", os.Getpid())
			}
			err := migrate(t, db.AdminURL, role)
			if !errors.Is(err, postgres.ErrInvalidAppRole) {
				t.Fatalf("Migrate = %v, want ErrInvalidAppRole", err)
			}
			var exists bool
			if err := connect(t, db.AdminURL).QueryRow(t.Context(), "SELECT to_regclass('iam.tenants') IS NOT NULL").Scan(&exists); err != nil || exists {
				t.Fatalf("iam.tenants exists after a refused run (%v)", err)
			}
		})
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
