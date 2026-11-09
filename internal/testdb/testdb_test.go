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

func TestNewWithAppRole_ConnectsAsUnprivilegedRole(t *testing.T) {
	db := testdb.NewWithAppRole(t)
	if !strings.HasPrefix(db.AppRole, "iam_test_app_") {
		t.Fatalf("role %q lacks the iam_test_app_ prefix", db.AppRole)
	}
	app := connect(t, db.AppURL)
	var user string
	var super, bypass bool
	err := app.QueryRow(context.Background(),
		"SELECT current_user, rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user").Scan(&user, &super, &bypass)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if user != db.AppRole || super || bypass {
		t.Fatalf("connected as %q (superuser %v, bypassrls %v); want %q, unprivileged", user, super, bypass, db.AppRole)
	}
	if _, err := app.Exec(context.Background(), "CREATE TABLE public.x (id int)"); err == nil {
		t.Fatal("app role created a table in the public schema")
	}
}

func TestNewWithAppRole_DropsRoleOnCleanup(t *testing.T) {
	var role string
	t.Run("sub", func(t *testing.T) { role = testdb.NewWithAppRole(t).AppRole })

	admin := connect(t, testdb.AdminURL(t))
	var exists bool
	err := admin.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)", role).Scan(&exists)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if exists {
		t.Fatalf("role %q still exists after cleanup", role)
	}
}

func TestNewRole_CreatesAndDrops(t *testing.T) {
	var role string
	t.Run("sub", func(t *testing.T) {
		role = testdb.NewRole(t, testdb.New(t), "NOLOGIN BYPASSRLS")
		var bypass bool
		err := connect(t, testdb.AdminURL(t)).QueryRow(context.Background(),
			"SELECT rolbypassrls FROM pg_roles WHERE rolname = $1", role).Scan(&bypass)
		if err != nil || !bypass || !strings.HasPrefix(role, "iam_test_app_") {
			t.Fatalf("role %q: bypassrls %v, %v", role, bypass, err)
		}
	})
	var exists bool
	err := connect(t, testdb.AdminURL(t)).QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)", role).Scan(&exists)
	if err != nil || exists {
		t.Fatalf("role %q still exists after cleanup (%v)", role, err)
	}
}

// The role's cleanup runs while the database it owns objects in still exists
// (the database was registered first, so it is dropped after): it must drop
// what the role owns there, and the role with it.
func TestNewRole_DropsRoleThatOwnsObjectsInALiveDatabase(t *testing.T) {
	var role string
	t.Run("sub", func(t *testing.T) {
		url := testdb.New(t)
		role = testdb.NewRole(t, url, "NOLOGIN")
		conn := connect(t, url)
		for _, sql := range []string{
			"CREATE TABLE owned (id int)",
			"ALTER TABLE owned OWNER TO " + pgx.Identifier{role}.Sanitize(),
			"CREATE TABLE granted (id int)",
			"GRANT SELECT ON granted TO " + pgx.Identifier{role}.Sanitize(),
		} {
			if _, err := conn.Exec(context.Background(), sql); err != nil {
				t.Fatalf("%s: %v", sql, err)
			}
		}
	})
	var exists bool
	err := connect(t, testdb.AdminURL(t)).QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)", role).Scan(&exists)
	if err != nil || exists {
		t.Fatalf("role %q still exists after cleanup (%v)", role, err)
	}
}
