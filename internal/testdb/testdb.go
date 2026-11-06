// Package testdb gives integration tests an isolated, throwaway Postgres
// database on the server named by IAM_TEST_DATABASE_URL.
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
)

const (
	adminEnv   = "IAM_TEST_DATABASE_URL"
	requireEnv = "IAM_REQUIRE_DB"
	opTimeout  = 10 * time.Second
)

var counter atomic.Uint64

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

	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	admin, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		t.Fatalf("testdb: connect to admin database: %v", err)
	}
	defer func() { _ = admin.Close(context.Background()) }()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+quoted); err != nil {
		t.Fatalf("testdb: create %s: %v", name, err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
		defer cancel()
		conn, err := pgx.Connect(ctx, adminURL)
		if err != nil {
			t.Errorf("testdb: connect for cleanup: %v", err)
			return
		}
		defer func() { _ = conn.Close(context.Background()) }()
		if _, err := conn.Exec(ctx, "DROP DATABASE IF EXISTS "+quoted+" WITH (FORCE)"); err != nil {
			t.Errorf("testdb: drop %s: %v", name, err)
		}
	})

	u, err := url.Parse(adminURL)
	if err != nil {
		t.Fatalf("testdb: parse %s: invalid URL", adminEnv)
	}
	u.Path = "/" + name
	return u.String()
}
