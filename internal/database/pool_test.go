package database_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aleogr/iam/internal/database"
	"github.com/aleogr/iam/internal/httpserver"
	"github.com/aleogr/iam/internal/testdb"
)

func TestOpen_MaxConnsIsTwo(t *testing.T) {
	pool, err := database.Open(context.Background(), testdb.New(t))
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
		"postgres://iam:"+password+"@127.0.0.1:1/iam?sslmode=disable")
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
	pool, err := database.Open(context.Background(), testdb.New(t))
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
	pool, err := database.OpenMigrator(context.Background(), testdb.New(t))
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
	for name, open := range map[string]func(context.Context, string) (*pgxpool.Pool, error){
		"Open": database.Open, "OpenMigrator": database.OpenMigrator,
	} {
		pool, err := open(context.Background(), raw)
		if err == nil {
			pool.Close()
			t.Fatalf("%s accepted an invalid URL", name)
		}
		if strings.Contains(err.Error(), password) {
			t.Fatalf("%s: error leaks the password: %v", name, err)
		}
	}
}
