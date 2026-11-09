package postgres_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/aleogr/iam/adapters/postgres"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

// The RLS tests prove the second barrier on its own: they run statements the
// repositories never issue (no tenant_id filter, a foreign tenant_id) as the
// app role and check that the database refuses or hides other tenants' rows.

var tenantTables = []string{"organizations", "users", "memberships", "applications"}

const codeInsufficientPrivilege = "42501"

// seeded is one tenant with one row in every tenant-owned table.
type seeded struct {
	tenant ids.TenantID
	org    ids.OrganizationID
	user   ids.UserID
}

func (f *fixture) seed(s storage.Store, slug string) seeded {
	f.t.Helper()
	r := seeded{tenant: f.createTenant(s, slug), org: ids.OrganizationID{ID: f.id()}, user: ids.UserID{ID: f.id()}}
	err := s.InTenant(f.t.Context(), r.tenant, func(ctx context.Context, tx storage.Tx) error {
		return errors.Join(
			tx.Organizations().Create(ctx, tenancy.Organization{ID: r.org, Slug: "sales", Name: "Sales", CreatedAt: f.now}),
			tx.Users().Create(ctx, tenancy.User{ID: r.user, Email: "ana@example.com", Status: tenancy.StatusActive, CreatedAt: f.now}),
			tx.Memberships().Add(ctx, tenancy.Membership{OrganizationID: r.org, UserID: r.user, CreatedAt: f.now}),
			tx.Applications().Create(ctx, tenancy.Application{
				ID: ids.ApplicationID{ID: f.id()}, Name: "App", Type: tenancy.ClientPublic,
				RedirectURIs: []string{"https://app.example.com/cb"}, CreatedAt: f.now,
			}),
		)
	})
	if err != nil {
		f.t.Fatalf("seed %s: %v", slug, err)
	}
	return r
}

// rawInTenant runs fn with the raw transaction of InTenant(tenant).
func rawInTenant(t *testing.T, s storage.Store, tenant ids.TenantID, fn func(ctx context.Context, raw pgx.Tx)) {
	t.Helper()
	if err := s.InTenant(t.Context(), tenant, func(ctx context.Context, tx storage.Tx) error {
		fn(ctx, postgres.RawTx(tx))
		return nil
	}); err != nil {
		t.Fatalf("InTenant: %v", err)
	}
}

func rawInSystem(t *testing.T, s storage.Store, fn func(ctx context.Context, raw pgx.Tx)) {
	t.Helper()
	if err := s.InSystem(t.Context(), func(ctx context.Context, tx storage.SystemTx) error {
		fn(ctx, postgres.RawTx(tx))
		return nil
	}); err != nil {
		t.Fatalf("InSystem: %v", err)
	}
}

type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func count(t *testing.T, ctx context.Context, q querier, table string) int {
	t.Helper()
	var n int
	if err := q.QueryRow(ctx, "SELECT count(*) FROM iam."+table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func setting(t *testing.T, ctx context.Context, q querier) (tenant string, backend int) {
	t.Helper()
	if err := q.QueryRow(ctx, "SELECT current_setting('iam.tenant_id', true), pg_backend_pid()").Scan(&tenant, &backend); err != nil {
		t.Fatalf("read setting: %v", err)
	}
	return tenant, backend
}

func TestRLS_BlocksQueryWithoutTenantFilter(t *testing.T) {
	f := newFixture(t)
	store := newStore(t, appPool(t, migratedDB(t)))
	a, b := f.seed(store, "tenant-a"), f.seed(store, "tenant-b")
	for _, tenant := range []ids.TenantID{a.tenant, b.tenant} {
		rawInTenant(t, store, tenant, func(ctx context.Context, raw pgx.Tx) {
			for _, table := range tenantTables {
				var total, own int
				err := raw.QueryRow(ctx, "SELECT count(*), count(*) FILTER (WHERE tenant_id = $1) FROM iam."+table, tenant.ID.String()).Scan(&total, &own)
				if err != nil {
					t.Fatalf("%s: %v", table, err)
				}
				if total != 1 || own != 1 {
					t.Errorf("%s without a tenant filter sees %d rows (%d own); want only the 1 of its tenant", table, total, own)
				}
			}
		})
	}
}

func TestRLS_SettingDoesNotLeakAcrossTransactions(t *testing.T) {
	f := newFixture(t)
	db := migratedDB(t)
	pool := singleConnPool(t, db)
	store := newStore(t, pool)
	a := f.seed(store, "tenant-a")
	var inside int
	rawInTenant(t, store, a.tenant, func(ctx context.Context, raw pgx.Tx) {
		var got string
		got, inside = setting(t, ctx, raw)
		if got != a.tenant.String() {
			t.Fatalf("iam.tenant_id inside InTenant = %q, want %v", got, a.tenant.ID)
		}
	})
	ctx := t.Context()
	got, backend := setting(t, ctx, pool)
	if backend != inside {
		t.Fatalf("pool of one served backend %d after %d", backend, inside)
	}
	if got != "" {
		t.Errorf("iam.tenant_id after commit = %q, want empty", got)
	}
	for _, table := range tenantTables {
		if n := count(t, ctx, pool, table); n != 0 {
			t.Errorf("%s on the pooled connection without a setting sees %d rows, want 0", table, n)
		}
	}
	// A failed transaction must not leave the setting behind either.
	err := store.InTenant(ctx, a.tenant, func(context.Context, storage.Tx) error { return errors.New("boom") })
	if err == nil {
		t.Fatal("InTenant returned nil for a failing fn")
	}
	if got, _ := setting(t, ctx, pool); got != "" {
		t.Errorf("iam.tenant_id after rollback = %q, want empty", got)
	}
}

func TestRLS_InSystemResetsTenantAfterInTenant(t *testing.T) {
	f := newFixture(t)
	store := newStore(t, singleConnPool(t, migratedDB(t)))
	a := f.seed(store, "tenant-a")
	var backend int
	check := func(wantSetting string, wantUsers int) {
		t.Helper()
		rawInTenant(t, store, a.tenant, func(ctx context.Context, raw pgx.Tx) { _, backend = setting(t, ctx, raw) })
		rawInSystem(t, store, func(ctx context.Context, raw pgx.Tx) {
			got, b := setting(t, ctx, raw)
			if b != backend {
				t.Fatalf("InSystem ran on backend %d, InTenant on %d", b, backend)
			}
			if got != wantSetting {
				t.Errorf("iam.tenant_id in InSystem = %q, want %q", got, wantSetting)
			}
			if n := count(t, ctx, raw, "users"); n != wantUsers {
				t.Errorf("users visible in InSystem = %d, want %d", n, wantUsers)
			}
		})
	}
	// Ruling R1: before the system tenant exists, InSystem has no tenant.
	check("", 0)

	var sys tenancy.Tenant
	if err := store.InSystem(t.Context(), func(ctx context.Context, tx storage.SystemTx) error {
		var err error
		sys, err = tx.EnsureSystemTenant(ctx, ids.TenantID{ID: f.id()}, "Platform", f.gen.Clock)
		return err
	}); err != nil {
		t.Fatalf("EnsureSystemTenant: %v", err)
	}
	// Afterwards it is bound to the system tenant, which has no users yet.
	check(sys.ID.String(), 0)
}

func TestRLS_WithCheckRefusesForeignTenantInsert(t *testing.T) {
	f := newFixture(t)
	store := newStore(t, appPool(t, migratedDB(t)))
	a, b := f.seed(store, "tenant-a"), f.seed(store, "tenant-b")
	bt := b.tenant.String()
	statements := map[string][]any{
		"INSERT INTO iam.organizations VALUES ($1, $2, 'other', 'Other', now())":                             {f.id().String(), bt},
		"INSERT INTO iam.users VALUES ($1, $2, 'bia@example.com', false, 'active', now())":                   {f.id().String(), bt},
		"INSERT INTO iam.memberships VALUES ($1, $2, $3, now())":                                             {bt, b.org.String(), b.user.String()},
		"INSERT INTO iam.applications VALUES ($1, $2, 'X', 'public', '{https://x.example/cb}', NULL, now())": {f.id().String(), bt},
	}
	rawInTenant(t, store, a.tenant, func(ctx context.Context, raw pgx.Tx) {
		for sql, args := range statements {
			sp, err := raw.Begin(ctx)
			if err != nil {
				t.Fatalf("savepoint: %v", err)
			}
			_, err = sp.Exec(ctx, sql, args...)
			var pgErr *pgconn.PgError
			// The privilege error has the same code: the message tells the
			// policy apart.
			if !errors.As(err, &pgErr) || pgErr.Code != codeInsufficientPrivilege || !strings.Contains(pgErr.Message, "row-level security") {
				t.Errorf("%s: error %v, want the RLS refusal (42501)", sql, err)
			}
			_ = sp.Rollback(ctx)
		}
		// B's membership cannot be deleted from A (DELETE is granted only on
		// memberships): it does not match.
		tag, err := raw.Exec(ctx, "DELETE FROM iam.memberships WHERE user_id = $1", b.user.String())
		if err != nil || tag.RowsAffected() != 0 {
			t.Errorf("DELETE of B's membership from A: %d rows, %v; want 0 rows", tag.RowsAffected(), err)
		}
	})
	rawInTenant(t, store, b.tenant, func(ctx context.Context, raw pgx.Tx) {
		for _, table := range tenantTables {
			if n := count(t, ctx, raw, table); n != 1 {
				t.Errorf("tenant B holds %d %s, want its 1", n, table)
			}
		}
	})
}

// The app role owns nothing, has only DML on the tenancy tables, and can
// neither turn RLS off nor reach around it.

func TestRLS_AppRoleOwnsNothingAndRLSIsForced(t *testing.T) {
	app := connect(t, migratedDB(t).AppURL)
	ctx := t.Context()
	var owned int
	err := app.QueryRow(ctx, `SELECT count(*) FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'iam' AND pg_get_userbyid(c.relowner) = current_user`).Scan(&owned)
	if err != nil || owned != 0 {
		t.Fatalf("app role owns %d objects in iam (%v), want none", owned, err)
	}
	var rls, force bool
	for _, table := range tenantTables {
		err := app.QueryRow(ctx, "SELECT relrowsecurity, relforcerowsecurity FROM pg_class WHERE oid = $1::regclass", "iam."+table).Scan(&rls, &force)
		if err != nil || !rls || !force {
			t.Errorf("%s: row security %v, forced %v (%v); want both", table, rls, force, err)
		}
	}
}

// TestRLS_AppRoleHasLeastPrivilege: exactly what the store uses. Later
// features widen it through expand migrations.
func TestRLS_AppRoleHasLeastPrivilege(t *testing.T) {
	app := connect(t, migratedDB(t).AppURL)
	ctx := t.Context()
	want := map[string][]string{
		"tenants": {"SELECT", "INSERT"}, "organizations": {"SELECT", "INSERT"}, "users": {"SELECT", "INSERT"},
		"applications": {"SELECT", "INSERT"}, "memberships": {"SELECT", "INSERT", "DELETE"}, "schema_migrations": nil,
	}
	for table, granted := range want {
		for _, priv := range []string{"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"} {
			var has bool
			if err := app.QueryRow(ctx, "SELECT has_table_privilege($1, $2)", "iam."+table, priv).Scan(&has); err != nil {
				t.Fatalf("privilege: %v", err)
			}
			if has != slices.Contains(granted, priv) {
				t.Errorf("app role has %s on %s: %v, want %v", priv, table, has, !has)
			}
		}
	}
	var usage, create bool
	err := app.QueryRow(ctx, "SELECT has_schema_privilege('iam', 'USAGE'), has_schema_privilege('iam', 'CREATE')").Scan(&usage, &create)
	if err != nil || !usage || create {
		t.Errorf("schema iam: USAGE %v, CREATE %v (%v); want USAGE only", usage, create, err)
	}
}

// The tenant registry, which has no RLS, cannot be changed or emptied by the
// app role.
func TestRLS_TenantsUpdateAndDeleteRefused(t *testing.T) {
	f := newFixture(t)
	db := migratedDB(t)
	f.seed(newStore(t, appPool(t, db)), "tenant-a")
	app := connect(t, db.AppURL)
	for _, sql := range []string{"UPDATE iam.tenants SET status = 'suspended'", "DELETE FROM iam.tenants"} {
		_, err := app.Exec(t.Context(), sql)
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != codeInsufficientPrivilege {
			t.Errorf("%s: error %v, want insufficient privilege (42501)", sql, err)
		}
	}
}

func TestRLS_AppRoleCannotAlterDropOrDisableRLS(t *testing.T) {
	app := connect(t, migratedDB(t).AppURL)
	ctx := t.Context()
	for _, sql := range []string{
		"ALTER TABLE iam.users DISABLE ROW LEVEL SECURITY",
		"ALTER TABLE iam.users NO FORCE ROW LEVEL SECURITY",
		"DROP POLICY tenant_isolation ON iam.users",
		"ALTER POLICY tenant_isolation ON iam.users USING (true)",
		"CREATE POLICY open ON iam.users USING (true)",
		"DROP TABLE iam.memberships",
		"ALTER TABLE iam.users ADD COLUMN x int",
		"TRUNCATE iam.users",
		"CREATE TABLE iam.x (id int)",
		"SELECT count(*) FROM iam.schema_migrations",
		`INSERT INTO iam.schema_migrations VALUES ('x', '\x00', now())`,
	} {
		_, err := app.Exec(ctx, sql)
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != codeInsufficientPrivilege {
			t.Errorf("%s: error %v, want insufficient privilege (42501)", sql, err)
		}
	}
	// With row_security off, a role that cannot bypass RLS gets an error, never
	// unfiltered rows.
	if _, err := app.Exec(ctx, "SET row_security = off"); err != nil {
		t.Fatalf("SET row_security: %v", err)
	}
	if _, err := app.Exec(ctx, "SELECT count(*) FROM iam.users"); err == nil {
		t.Error("query with row_security off succeeded; want it refused")
	}
}
