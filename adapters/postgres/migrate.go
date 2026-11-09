package postgres

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aleogr/iam/shared/crypto"
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

// ErrMigrationChanged reports an applied migration whose embedded file no
// longer has the checksum recorded when it was applied: migrations are
// immutable once applied; change the schema with a new one.
var ErrMigrationChanged = errors.New("postgres: applied migration was changed")

// MigrateConfig is what Migrate needs from configuration.
type MigrateConfig struct {
	// AppRole is the service's role (IAM_DATABASE_APP_ROLE).
	AppRole string
	// StatementTimeout and LockTimeout bound each statement and each lock
	// wait of a migration transaction, the migration lock included (at least
	// 1ms each).
	StatementTimeout time.Duration
	LockTimeout      time.Duration
}

// migration is one embedded SQL file.
type migration struct {
	version  string // the file name without .sql
	sql      string
	checksum []byte // crypto.Checksum of the file
}

// Migrate applies the embedded migrations (migrations/*.sql) in name order,
// each in its own transaction under pg_advisory_xact_lock(MigrationLockKey),
// and records each in iam.schema_migrations with the checksum of its file.
// Applied versions are skipped, so running it again is a no-op and
// concurrent runs serialize; an applied version whose file changed is
// ErrMigrationChanged. It runs as the migration role, which owns the iam
// schema and everything in it. Last, in one more transaction, it grants the
// app role USAGE on the schema and verifies that the privileges the app role
// can reach are exactly appPrivileges (ErrInvalidAppRole otherwise), so a
// changed IAM_DATABASE_APP_ROLE fails instead of starting a service that
// cannot work.
//
// Errors never carry SQL values or the connection URL.
func Migrate(ctx context.Context, pool *pgxpool.Pool, cfg MigrateConfig) error {
	if !validAppRoleName(cfg.AppRole) {
		return ErrInvalidAppRole
	}
	if cfg.StatementTimeout < time.Millisecond || cfg.LockTimeout < time.Millisecond {
		return errors.New("postgres: migration timeouts must be at least 1ms")
	}
	ms, err := loadMigrations(migrationFiles)
	if err != nil {
		return err
	}
	for _, m := range ms {
		err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error { return applyMigration(ctx, tx, cfg, m) })
		if err != nil {
			return fmt.Errorf("postgres: migration %s: %w", m.version, sanitize(err))
		}
	}
	if err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error { return grantAndVerify(ctx, tx, cfg) }); err != nil {
		return fmt.Errorf("postgres: app role grants: %w", sanitize(err))
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
		out = append(out, migration{version: strings.TrimSuffix(base, ".sql"), sql: sql, checksum: crypto.Checksum(b)})
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
    checksum   bytea       NOT NULL,
    applied_at timestamptz NOT NULL
);`

// beginMigration sets the transaction's timeouts, takes the migration lock, prepares
// the schema and checks the app role; it returns the role's OID.
func beginMigration(ctx context.Context, tx pgx.Tx, cfg MigrateConfig) (uint32, error) {
	_, err := tx.Exec(ctx, `SELECT set_config('statement_timeout', $1, true), set_config('lock_timeout', $2, true)`,
		milliseconds(cfg.StatementTimeout), milliseconds(cfg.LockTimeout))
	if err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", int64(MigrationLockKey)); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx, bootstrapSQL); err != nil {
		return 0, err
	}
	return checkAppRole(ctx, tx, cfg.AppRole)
}

func milliseconds(d time.Duration) string { return strconv.FormatInt(d.Milliseconds(), 10) }

func applyMigration(ctx context.Context, tx pgx.Tx, cfg MigrateConfig, m migration) error {
	if _, err := beginMigration(ctx, tx, cfg); err != nil {
		return err
	}
	var stored []byte
	err := tx.QueryRow(ctx, "SELECT checksum FROM iam.schema_migrations WHERE version = $1", m.version).Scan(&stored)
	switch {
	case err == nil && bytes.Equal(stored, m.checksum):
		return nil // applied
	case err == nil:
		return ErrMigrationChanged
	case !errors.Is(err, pgx.ErrNoRows):
		return err
	}
	quoted := pgx.Identifier{cfg.AppRole}.Sanitize()
	if _, err := tx.Exec(ctx, strings.ReplaceAll(m.sql, appRolePlaceholder, quoted)); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "INSERT INTO iam.schema_migrations (version, checksum, applied_at) VALUES ($1, $2, now())",
		m.version, m.checksum)
	return err
}

func grantAndVerify(ctx context.Context, tx pgx.Tx, cfg MigrateConfig) error {
	role, err := beginMigration(ctx, tx, cfg)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "GRANT USAGE ON SCHEMA iam TO "+pgx.Identifier{cfg.AppRole}.Sanitize()); err != nil {
		return err
	}
	return verifyPrivileges(ctx, tx, role)
}
