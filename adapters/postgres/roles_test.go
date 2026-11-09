package postgres_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aleogr/iam/adapters/postgres"
	"github.com/aleogr/iam/internal/testdb"
)

// A role the app role can become (pg_has_role MEMBER: SET ROLE works even
// without INHERIT) is as good as the app role itself. Each case makes the app
// role reach a privilege that gets around RLS or the ownership rules.
type roleCase struct {
	name string
	// setup changes db (migrated already when migrated is true) so that its app
	// role is unsafe.
	setup    func(t *testing.T, db testdb.AppDB)
	migrated bool
}

func memberOf(attributes string) func(*testing.T, testdb.AppDB) {
	return func(t *testing.T, db testdb.AppDB) {
		r := testdb.NewRole(t, "NOLOGIN "+attributes)
		adminExec(t, db, "GRANT "+quote(r)+" TO "+quote(db.AppRole))
	}
}

// memberOfOwner makes the app role a member of a role that owns an object in
// the iam schema, created by create (%s is the owner role).
func memberOfOwner(create, drop string) func(*testing.T, testdb.AppDB) {
	return func(t *testing.T, db testdb.AppDB) {
		r := testdb.NewRole(t, "NOLOGIN")
		adminExec(t, db, fmt.Sprintf(create, quote(r)), "GRANT "+quote(r)+" TO "+quote(db.AppRole))
		// Registered after NewRole, so it runs first and the role owns nothing
		// when it is dropped.
		t.Cleanup(func() { adminExec(t, db, drop) })
	}
}

var unsafeRoleCases = []roleCase{
	{name: "superuser", setup: func(t *testing.T, db testdb.AppDB) { adminExec(t, db, "ALTER ROLE "+quote(db.AppRole)+" SUPERUSER") }},
	{name: "bypassrls", setup: func(t *testing.T, db testdb.AppDB) { adminExec(t, db, "ALTER ROLE "+quote(db.AppRole)+" BYPASSRLS") }},
	{name: "member of a superuser role", setup: memberOf("SUPERUSER")},
	{name: "member of a bypassrls role", setup: memberOf("BYPASSRLS")},
	{name: "member of the migrator", setup: func(t *testing.T, db testdb.AppDB) {
		adminExec(t, db, "DO $$ BEGIN EXECUTE format('GRANT %I TO %s', current_user, '"+quote(db.AppRole)+"'); END $$")
	}},
	{name: "member of an owner of a table", migrated: true, setup: memberOfOwner(
		"CREATE TABLE iam.extra (id int); ALTER TABLE iam.extra OWNER TO %s", "DROP TABLE iam.extra")},
	{name: "member of an owner of a function", migrated: true, setup: memberOfOwner(
		"CREATE FUNCTION iam.f() RETURNS int LANGUAGE sql AS 'SELECT 1'; ALTER FUNCTION iam.f() OWNER TO %s", "DROP FUNCTION iam.f()")},
	{name: "member of an owner of a type", migrated: true, setup: memberOfOwner(
		"CREATE TYPE iam.e AS ENUM ('a'); ALTER TYPE iam.e OWNER TO %s", "DROP TYPE iam.e")},
	{name: "owner of a function", migrated: true, setup: func(t *testing.T, db testdb.AppDB) {
		adminExec(t, db, "CREATE FUNCTION iam.g() RETURNS int LANGUAGE sql AS 'SELECT 1'",
			"ALTER FUNCTION iam.g() OWNER TO "+quote(db.AppRole))
		t.Cleanup(func() { adminExec(t, db, "DROP FUNCTION iam.g()") })
	}},
}

// TestMigrate_RefusesUnsafeAppRole: Migrate refuses an app role that is, or
// can become, a role that gets around RLS or owns the schema, and applies
// nothing on a fresh database.
func TestMigrate_RefusesUnsafeAppRole(t *testing.T) {
	cases := append([]roleCase{{name: "missing", setup: func(*testing.T, testdb.AppDB) {}}}, unsafeRoleCases...)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := testdb.NewWithAppRole(t)
			if tc.migrated {
				db = migratedInto(t, db)
			}
			tc.setup(t, db)
			role := db.AppRole
			if tc.name == "missing" {
				role = fmt.Sprintf("iam_test_app_%d_missing", os.Getpid())
			}
			if err := migrate(t, db.AdminURL, role); !errors.Is(err, postgres.ErrInvalidAppRole) {
				t.Fatalf("Migrate = %v, want ErrInvalidAppRole", err)
			}
			if !tc.migrated {
				var exists bool
				err := connect(t, db.AdminURL).QueryRow(t.Context(), "SELECT to_regclass('iam.tenants') IS NOT NULL").Scan(&exists)
				if err != nil || exists {
					t.Fatalf("iam.tenants exists after a refused run (%v)", err)
				}
			}
		})
	}
}

// TestNewStore_RefusesUnsafeRole is the startup check: a store whose
// connection role could get around RLS fails closed, whatever Migrate saw (the
// lab service logs in as iam-run, a member of the app role).
func TestNewStore_RefusesUnsafeRole(t *testing.T) {
	for _, tc := range unsafeRoleCases {
		if tc.name == "member of the migrator" {
			continue // the migrator owns the schema: covered by the owner cases
		}
		t.Run(tc.name, func(t *testing.T) {
			db := migratedDB(t)
			tc.setup(t, db)
			if _, err := postgres.NewStore(t.Context(), appPool(t, db)); !errors.Is(err, postgres.ErrInvalidAppRole) {
				t.Fatalf("NewStore = %v, want ErrInvalidAppRole", err)
			}
		})
	}
	t.Run("migrator connection", func(t *testing.T) {
		db := migratedDB(t)
		db.AppURL = db.AdminURL
		if _, err := postgres.NewStore(t.Context(), appPool(t, db)); !errors.Is(err, postgres.ErrInvalidAppRole) {
			t.Fatalf("NewStore as the schema owner = %v, want ErrInvalidAppRole", err)
		}
	})
}

// The reviewer's probe: a member of a BYPASSRLS role that switches to it
// reads every tenant. NewStore refuses such a connection before any query.
func TestRLS_MembershipBypassIsRefused(t *testing.T) {
	db := migratedDB(t)
	bypass := testdb.NewRole(t, "NOLOGIN BYPASSRLS")
	adminExec(t, db, "GRANT "+quote(bypass)+" TO "+quote(db.AppRole))
	app := connect(t, db.AppURL)
	if _, err := app.Exec(t.Context(), "SET ROLE "+quote(bypass)); err != nil {
		t.Fatalf("probe: SET ROLE: %v", err) // the attack is real
	}
	if _, err := postgres.NewStore(t.Context(), appPool(t, db)); !errors.Is(err, postgres.ErrInvalidAppRole) {
		t.Fatalf("NewStore = %v, want ErrInvalidAppRole", err)
	}
}

func migratedInto(t *testing.T, db testdb.AppDB) testdb.AppDB {
	t.Helper()
	if err := migrate(t, db.AdminURL, db.AppRole); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return db
}
