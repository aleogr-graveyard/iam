package database_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aleogr/iam/internal/database"
	"github.com/aleogr/iam/internal/httpserver"
	"github.com/aleogr/iam/internal/testdb"
)

func TestOpen_MaxConnsIsTwo(t *testing.T) {
	pool, err := database.Open(context.Background(), testdb.New(t), testdb.Limits())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer pool.Close()
	if got := pool.Config().MaxConns; got != 2 {
		t.Fatalf("MaxConns = %d, want 2", got)
	}
}

func TestOpen_RejectsUnreachableDatabase(t *testing.T) {
	const password = "s3cr3t-pw"
	start := time.Now()
	pool, err := database.Open(context.Background(),
		"postgres://iam:"+password+"@127.0.0.1:1/iam?sslmode=disable", testdb.Limits())
	if err == nil {
		pool.Close()
		t.Fatal("Open succeeded against an unreachable address")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("Open took %v, want under 5s", elapsed)
	}
	if strings.Contains(err.Error(), password) {
		t.Fatalf("error leaks the password: %v", err)
	}
}

func TestReadyz_ClosedPoolIsUnavailable(t *testing.T) {
	pool, err := database.Open(context.Background(), testdb.New(t), testdb.Limits())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	srv := httpserver.New(":0", httpserver.WithReadinessChecks(database.ReadinessCheck(pool)))

	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("open pool: status = %d, want 200", rec.Code)
	}

	pool.Close()
	rec = httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("closed pool: status = %d, want 503", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"unavailable","failed":["database"]}` {
		t.Fatalf("body = %q", got)
	}
}

func TestOpenMigrator_MaxConnsIsOne(t *testing.T) {
	pool, err := database.OpenMigrator(context.Background(), testdb.New(t), testdb.Limits())
	if err != nil {
		t.Fatalf("OpenMigrator: %v", err)
	}
	defer pool.Close()
	if got := pool.Config().MaxConns; got != 1 {
		t.Fatalf("MaxConns = %d, want 1", got)
	}
}

func TestOpen_InvalidURLDoesNotLeakCredentials(t *testing.T) {
	const password = "s3cr3t-pw"
	raw := "postgres://iam:" + password + "@[::1:5432/iam"
	for name, open := range map[string]func(context.Context, string, database.Limits) (*pgxpool.Pool, error){
		"Open": database.Open, "OpenMigrator": database.OpenMigrator,
	} {
		pool, err := open(context.Background(), raw, testdb.Limits())
		if err == nil {
			pool.Close()
			t.Fatalf("%s accepted an invalid URL", name)
		}
		if strings.Contains(err.Error(), password) {
			t.Fatalf("%s: error leaks the password: %v", name, err)
		}
	}
}

// The session settings reach the server on both pool kinds, and a statement
// longer than the limit is cancelled by the server.
func TestOpen_LimitsApplyOnTheServer(t *testing.T) {
	url := testdb.New(t)
	limits := testdb.Limits()
	limits.StatementTimeout = 200 * time.Millisecond
	for name, open := range map[string]func(context.Context, string, database.Limits) (*pgxpool.Pool, error){
		"Open": database.Open, "OpenMigrator": database.OpenMigrator,
	} {
		pool, err := open(context.Background(), url, limits)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		var stmt, lock string
		if err := pool.QueryRow(context.Background(), "SELECT current_setting('statement_timeout'), current_setting('lock_timeout')").Scan(&stmt, &lock); err != nil {
			t.Fatalf("%s: read settings: %v", name, err)
		}
		if stmt != "200ms" || lock != "2s" {
			t.Errorf("%s: statement_timeout %q, lock_timeout %q; want 200ms, 2s", name, stmt, lock)
		}
		start := time.Now()
		_, err = pool.Exec(context.Background(), "SELECT pg_sleep(5)")
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "57014" {
			t.Errorf("%s: long statement: %v, want query_canceled (57014)", name, err)
		}
		if elapsed := time.Since(start); elapsed > 3*time.Second {
			t.Errorf("%s: cancelled after %v", name, elapsed)
		}
		pool.Close()
	}
}
