package testdb_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/aleogr/iam/internal/testdb"
)

func connect(t *testing.T, url string) *pgx.Conn {
	t.Helper()
	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}

func TestNew_IsolatedDatabases(t *testing.T) {
	ctx := context.Background()
	a := connect(t, testdb.New(t))
	b := connect(t, testdb.New(t))

	if _, err := a.Exec(ctx, "CREATE TABLE only_in_a (id int)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	var found bool
	if err := b.QueryRow(ctx, "SELECT to_regclass('only_in_a') IS NOT NULL").Scan(&found); err != nil {
		t.Fatalf("query: %v", err)
	}
	if found {
		t.Fatal("table created in one database is visible in the other")
	}
}

func TestNew_DropsDatabaseOnCleanup(t *testing.T) {
	var name string
	t.Run("sub", func(t *testing.T) {
		u, err := url.Parse(testdb.New(t))
		if err != nil {
			t.Fatalf("parse url: %v", err)
		}
		name = strings.TrimPrefix(u.Path, "/")
		if !strings.HasPrefix(name, "iam_test_") {
			t.Fatalf("database name %q lacks the iam_test_ prefix", name)
		}
	})

	admin := connect(t, testdb.AdminURL(t))
	var exists bool
	err := admin.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if exists {
		t.Fatalf("database %q still exists after cleanup", name)
	}
}
