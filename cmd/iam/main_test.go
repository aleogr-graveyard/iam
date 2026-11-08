package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/aleogr/iam/internal/testdb"
)

func setBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("IAM_ENVIRONMENT", "development")
	t.Setenv("IAM_PUBLIC_BASE_URL", "http://127.0.0.1:8080")
	t.Setenv("IAM_CURSOR_KEY", "")
	t.Setenv("IAM_CONFIG_FILE", "")
}

func TestRun_RejectsUnknownArguments(t *testing.T) {
	for _, args := range [][]string{{"serve"}, {"migrate", "now"}, {"--migrate"}} {
		if err := run(slog.New(slog.DiscardHandler), args); !errors.Is(err, errUsage) {
			t.Errorf("run(%q) = %v, want the usage error", args, err)
		}
	}
}

func TestRun_MigrateRequiresDatabaseURL(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("IAM_DATABASE_URL", "")
	if err := run(slog.New(slog.DiscardHandler), []string{"migrate"}); err == nil {
		t.Fatal("migrate without IAM_DATABASE_URL succeeded")
	}
}

// An invalid or unreachable database URL never puts its password in the
// error or in the logs (roadmap: inherited from milestone 0).
func TestRun_MigrateDoesNotLeakCredentials(t *testing.T) {
	const password = "s3cr3t-pw"
	setBaseEnv(t)
	for _, url := range []string{
		"postgres://iam:" + password + "@[::1:5432/iam",
		"postgres://iam:" + password + "@127.0.0.1:1/iam?sslmode=bogus",
		"postgres://iam:" + password + "@127.0.0.1:1/iam?sslmode=disable",
	} {
		t.Setenv("IAM_DATABASE_URL", url)
		var logs bytes.Buffer
		err := run(slog.New(slog.NewJSONHandler(&logs, nil)), []string{"migrate"})
		if err == nil {
			t.Fatal("migrate succeeded against an invalid database")
		}
		if strings.Contains(err.Error(), password) || strings.Contains(logs.String(), password) {
			t.Fatalf("password leaked: error %q, logs %q", err, logs.String())
		}
	}
}

func TestRun_MigrateAppliesMigrations(t *testing.T) {
	db := testdb.NewWithAppRole(t)
	setBaseEnv(t)
	t.Setenv("IAM_DATABASE_URL", db.AdminURL)
	t.Setenv("IAM_DATABASE_APP_ROLE", db.AppRole)
	for range 2 { // the second run is a no-op
		if err := run(slog.New(slog.DiscardHandler), []string{"migrate"}); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}
	conn, err := pgx.Connect(context.Background(), db.AppURL)
	if err != nil {
		t.Fatalf("connect as app role: %v", err)
	}
	defer func() { _ = conn.Close(context.Background()) }()
	var n int
	if err := conn.QueryRow(context.Background(), "SELECT count(*) FROM iam.tenants").Scan(&n); err != nil || n != 0 {
		t.Fatalf("app role reads iam.tenants: %d rows, %v; want 0, nil", n, err)
	}
}
