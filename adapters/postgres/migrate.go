package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MigrationLockKey is the pg_advisory_xact_lock key that serializes
// migration runs on one database.
const MigrationLockKey = 7234717

// appRolePlaceholder is replaced in every migration with the quoted name of
// the service's role.
const appRolePlaceholder = "{{app_role}}"

//go:embed migrations/*.sql
var migrationFiles embed.FS

var (
	// roleNamePattern is a lower-case unquoted Postgres identifier.
	roleNamePattern = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)
	// migrationName is <UTC timestamp yyyymmddhhmmss>_<description>.sql.
	migrationName = regexp.MustCompile(`^[0-9]{14}_[a-z0-9_]+\.sql$`)
)

// ErrInvalidAppRole reports an app role Migrate refuses: a name that is not a
// plain identifier starting with iam_ (the service never touches roles of
// other projects, ADR 0001), or a role that does not exist, is a superuser,
// bypasses row level security, is the migrating role or a member of it, or
// owns something in the iam schema.
var ErrInvalidAppRole = errors.New("postgres: invalid app role")

// migration is one embedded SQL file.
type migration struct {
	version string // the file name without .sql
	sql     string
}

// Migrate applies the embedded migrations (migrations/*.sql) in name order,
// each in its own transaction under pg_advisory_xact_lock(MigrationLockKey),
// and records each in iam.schema_migrations. Applied versions are skipped, so
// running it again is a no-op and concurrent runs serialize. It runs as the
// migration role, which owns the iam schema and everything in it; appRole,
// the service's role, gets USAGE on the schema and DML on the tables.
//
// Errors never carry SQL values or the connection URL.
func Migrate(ctx context.Context, pool *pgxpool.Pool, appRole string) error {
	if !validAppRoleName(appRole) {
		return ErrInvalidAppRole
	}
	ms, err := loadMigrations(migrationFiles)
	if err != nil {
		return err
	}
	role := pgx.Identifier{appRole}.Sanitize()
	for _, m := range ms {
		err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error { return applyMigration(ctx, tx, m, appRole, role) })
		if err != nil {
			return fmt.Errorf("postgres: migration %s: %w", m.version, sanitize(err))
		}
	}
	return nil
}

func validAppRoleName(name string) bool {
	return roleNamePattern.MatchString(name) && strings.HasPrefix(name, "iam_")
}

// loadMigrations reads every migration, checks its name and its placeholders
// and returns them in name (= time) order.
func loadMigrations(files fs.FS) ([]migration, error) {
	names, err := fs.Glob(files, "migrations/*.sql")
	if err != nil {
		return nil, fmt.Errorf("postgres: list migrations: %w", err)
	}
	slices.Sort(names)
	out := make([]migration, 0, len(names))
	for _, name := range names {
		base := path.Base(name)
		if !migrationName.MatchString(base) {
			return nil, fmt.Errorf("postgres: migration file %q is not named <yyyymmddhhmmss>_<name>.sql", base)
		}
		b, err := fs.ReadFile(files, name)
		if err != nil {
			return nil, fmt.Errorf("postgres: read migration %s: %w", base, err)
		}
		sql := string(b)
		if strings.Contains(strings.ReplaceAll(sql, appRolePlaceholder, ""), "{{") {
			return nil, fmt.Errorf("postgres: migration %s has an unknown placeholder", base)
		}
		out = append(out, migration{version: strings.TrimSuffix(base, ".sql"), sql: sql})
	}
	return out, nil
}

// bootstrapSQL prepares the schema and the version table. It runs under the
// lock in every migration transaction, so concurrent first runs cannot race
// on CREATE SCHEMA.
const bootstrapSQL = `
CREATE SCHEMA IF NOT EXISTS iam;
REVOKE ALL ON SCHEMA iam FROM PUBLIC;
CREATE TABLE IF NOT EXISTS iam.schema_migrations (
    version    text        PRIMARY KEY,
    applied_at timestamptz NOT NULL
);`

func applyMigration(ctx context.Context, tx pgx.Tx, m migration, appRole, quotedRole string) error {
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", int64(MigrationLockKey)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, bootstrapSQL); err != nil {
		return err
	}
	if err := checkAppRole(ctx, tx, appRole); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "GRANT USAGE ON SCHEMA iam TO "+quotedRole); err != nil {
		return err
	}
	var applied bool
	err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM iam.schema_migrations WHERE version = $1)", m.version).Scan(&applied)
	if err != nil || applied {
		return err
	}
	if _, err := tx.Exec(ctx, strings.ReplaceAll(m.sql, appRolePlaceholder, quotedRole)); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "INSERT INTO iam.schema_migrations (version, applied_at) VALUES ($1, now())", m.version)
	return err
}

// checkAppRole refuses an app role that could get around row level security
// or the table ownership rules (see ErrInvalidAppRole).
const checkAppRoleSQL = `
SELECT r.rolsuper, r.rolbypassrls, r.rolname = current_user,
       pg_has_role(r.oid, current_user, 'MEMBER'),
       EXISTS (SELECT 1 FROM pg_namespace n WHERE n.nspname = 'iam' AND n.nspowner = r.oid)
       OR EXISTS (SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
                  WHERE n.nspname = 'iam' AND c.relowner = r.oid)
FROM pg_roles r WHERE r.rolname = $1`

func checkAppRole(ctx context.Context, tx pgx.Tx, appRole string) error {
	var super, bypass, isSelf, member, owns bool
	err := tx.QueryRow(ctx, checkAppRoleSQL, appRole).Scan(&super, &bypass, &isSelf, &member, &owns)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: role does not exist", ErrInvalidAppRole)
	}
	if err != nil {
		return err
	}
	switch {
	case super:
		return fmt.Errorf("%w: role is a superuser", ErrInvalidAppRole)
	case bypass:
		return fmt.Errorf("%w: role bypasses row level security", ErrInvalidAppRole)
	case isSelf || member:
		return fmt.Errorf("%w: role is the migration role or a member of it", ErrInvalidAppRole)
	case owns:
		return fmt.Errorf("%w: role owns objects in the iam schema", ErrInvalidAppRole)
	}
	return nil
}
