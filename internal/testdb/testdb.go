// Package testdb gives integration tests an isolated, throwaway Postgres
// database on the server named by IAM_TEST_DATABASE_URL, and optionally a
// throwaway login role for it.
package testdb

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/aleogr/iam/internal/database"
	"github.com/aleogr/iam/shared/crypto"
)

const (
	adminEnv   = "IAM_TEST_DATABASE_URL"
	requireEnv = "IAM_REQUIRE_DB"
	opTimeout  = 10 * time.Second
)

var counter atomic.Uint64

// Limits returns the database limits tests open pools with: the shape of the
// configuration defaults, generous enough for a loaded CI runner.
func Limits() database.Limits {
	return database.Limits{
		StatementTimeout: 30 * time.Second, IdleInTransactionTimeout: 30 * time.Second,
		LockTimeout: 2 * time.Second, MaxConnIdleTime: time.Minute,
	}
}

// AdminURL returns the admin connection URL from IAM_TEST_DATABASE_URL. When
// it is unset the test is skipped, or fails if IAM_REQUIRE_DB=1.
func AdminURL(t testing.TB) string {
	t.Helper()
	u := os.Getenv(adminEnv)
	if u == "" {
		if os.Getenv(requireEnv) == "1" {
			t.Fatalf("%s is not set but %s=1", adminEnv, requireEnv)
		}
		t.Skipf("%s not set; skipping database test", adminEnv)
	}
	return u
}

// New creates a fresh database named iam_test_<pid>_<n>, drops it when the
// test ends and returns a connection URL for it. Randomness is avoided on
// purpose: the process id plus a counter is unique per run on one server.
func New(t testing.TB) string {
	t.Helper()
	adminURL := AdminURL(t)
	name := fmt.Sprintf("iam_test_%d_%d", os.Getpid(), counter.Add(1))
	quoted := pgx.Identifier{name}.Sanitize()

	adminExec(t, adminURL, "CREATE DATABASE "+quoted)
	t.Cleanup(func() { cleanupExec(t, adminURL, "DROP DATABASE IF EXISTS "+quoted+" WITH (FORCE)") })
	return withPath(t, adminURL, name)
}

// AppDB is a fresh database plus a login role for it that owns nothing and is
// granted nothing: the migration grants it what the service's role gets.
type AppDB struct {
	AdminURL string // the new database, as the admin user (owner of what it creates)
	AppURL   string // the new database, as AppRole
	AppRole  string // iam_test_app_<pid>_<n>
}

// NewWithAppRole creates a fresh database (as New) and a login role named
// iam_test_app_<pid>_<n> with a random password, NOSUPERUSER and NOBYPASSRLS.
// The database is dropped first and the role after it when the test ends.
func NewWithAppRole(t testing.TB) AppDB {
	t.Helper()
	adminURL := AdminURL(t)
	role := fmt.Sprintf("iam_test_app_%d_%d", os.Getpid(), counter.Add(1))
	quoted := pgx.Identifier{role}.Sanitize()
	raw, err := crypto.RandomBytes(24)
	if err != nil {
		t.Fatalf("testdb: role password: %v", err)
	}
	password := fmt.Sprintf("%x", raw)

	adminExec(t, adminURL, "CREATE ROLE "+quoted+
		" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS NOINHERIT PASSWORD '"+password+"'")
	// Registered before New's cleanup, so it runs after the database is gone
	// (cleanups run last-in, first-out) and no privilege still depends on it.
	t.Cleanup(func() { cleanupExec(t, adminURL, "DROP ROLE IF EXISTS "+quoted) })

	db := AppDB{AdminURL: New(t), AppRole: role}
	u, err := url.Parse(db.AdminURL)
	if err != nil {
		t.Fatalf("testdb: parse %s: invalid URL", adminEnv)
	}
	u.User = url.UserPassword(role, password)
	db.AppURL = u.String()
	return db
}

// NewRole creates a role iam_test_app_<pid>_<n> with the given attributes
// (for example "NOLOGIN BYPASSRLS"), dropped when the test ends. Objects it
// comes to own in a database must be dropped by the test first, or that
// database must be dropped before it (register this before the database).
func NewRole(t testing.TB, attributes string) string {
	t.Helper()
	adminURL := AdminURL(t)
	role := fmt.Sprintf("iam_test_app_%d_%d", os.Getpid(), counter.Add(1))
	quoted := pgx.Identifier{role}.Sanitize()
	adminExec(t, adminURL, "CREATE ROLE "+quoted+" "+attributes)
	t.Cleanup(func() { cleanupExec(t, adminURL, "DROP ROLE IF EXISTS "+quoted) })
	return role
}

// adminExec runs one statement on the admin database and fails the test on
// error. Errors from these statements never carry the admin password.
func adminExec(t testing.TB, adminURL, sql string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	admin, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		t.Fatalf("testdb: connect to admin database: %v", err)
	}
	defer func() { _ = admin.Close(context.Background()) }()
	if _, err := admin.Exec(ctx, sql); err != nil {
		t.Fatalf("testdb: setup statement failed: %v", err)
	}
}

// cleanupExec is adminExec for cleanups: it reports instead of stopping.
func cleanupExec(t testing.TB, adminURL, sql string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	conn, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		t.Errorf("testdb: connect for cleanup: %v", err)
		return
	}
	defer func() { _ = conn.Close(context.Background()) }()
	if _, err := conn.Exec(ctx, sql); err != nil {
		t.Errorf("testdb: cleanup statement failed: %v", err)
	}
}

func withPath(t testing.TB, adminURL, name string) string {
	t.Helper()
	u, err := url.Parse(adminURL)
	if err != nil {
		t.Fatalf("testdb: parse %s: invalid URL", adminEnv)
	}
	u.Path = "/" + name
	return u.String()
}
