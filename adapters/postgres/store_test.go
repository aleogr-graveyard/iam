package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/aleogr/iam/adapters/postgres"
	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/storage/storagetest"
	"github.com/aleogr/iam/shared/tenancy"
)

// TestPostgresStore_Contract runs the C3 contract suite with a fresh,
// migrated database per subtest (the suite expects an empty store, and the
// tenant registry is global, so per-subtest tenants would not do). The store
// connects as that database's own app role, which owns nothing and cannot
// bypass RLS, so every subtest, the cross-tenant ones included, runs with
// row level security in force under the explicit tenant filters.
func TestPostgresStore_Contract(t *testing.T) {
	storagetest.Run(t, func(t *testing.T) storage.Store {
		return postgres.NewStore(appPool(t, migratedDB(t)))
	})
}

type fixture struct {
	t   *testing.T
	gen ids.Generator
	now time.Time
}

func newFixture(t *testing.T) *fixture {
	now := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	return &fixture{t: t, gen: ids.Generator{Clock: &clock.Fixed{T: now}}, now: now}
}

func (f *fixture) id() ids.ID {
	f.t.Helper()
	id, err := f.gen.New()
	if err != nil {
		f.t.Fatal(err)
	}
	return id
}

func (f *fixture) createTenant(s storage.Store, slug string) ids.TenantID {
	f.t.Helper()
	tenant := tenancy.Tenant{ID: ids.TenantID{ID: f.id()}, Slug: slug, Name: "Tenant " + slug, Status: tenancy.StatusActive, CreatedAt: f.now}
	if err := s.InSystem(f.t.Context(), func(ctx context.Context, tx storage.SystemTx) error {
		return tx.CreateTenant(ctx, tenant)
	}); err != nil {
		f.t.Fatalf("CreateTenant: %v", err)
	}
	return tenant.ID
}

// The contract suite checks the default (insecure http refused); the option
// set from configuration in development accepts it.
func TestStore_AllowInsecureHTTPOption(t *testing.T) {
	f := newFixture(t)
	db := migratedDB(t)
	pool := appPool(t, db)
	for i, tc := range []struct {
		store *postgres.Store
		ok    bool
	}{
		{postgres.NewStore(pool), false},
		{postgres.NewStore(pool, postgres.WithAllowInsecureHTTP(false)), false},
		{postgres.NewStore(pool, postgres.WithAllowInsecureHTTP(true)), true},
	} {
		tenant := f.createTenant(tc.store, fmt.Sprintf("acme-%d", i))
		app := tenancy.Application{
			ID: ids.ApplicationID{ID: f.id()}, Name: "Dev", Type: tenancy.ClientPublic,
			RedirectURIs: []string{"http://app.example.com/cb"}, CreatedAt: f.now,
		}
		err := tc.store.InTenant(t.Context(), tenant, func(ctx context.Context, tx storage.Tx) error {
			return tx.Applications().Create(ctx, app)
		})
		if tc.ok && err != nil {
			t.Errorf("allowed: Create = %v, want nil", err)
		}
		if !tc.ok && !errors.Is(err, storage.ErrInvalid) {
			t.Errorf("refused: Create = %v, want ErrInvalid", err)
		}
	}
}

// TestStore_CorruptRowIsAnError: rows written behind the adapter's back that
// fail validation are reported, never returned or skipped.
func TestStore_CorruptRowIsAnError(t *testing.T) {
	f := newFixture(t)
	db := migratedDB(t)
	store := postgres.NewStore(appPool(t, db))
	tenant := f.createTenant(store, "acme")
	admin := connect(t, db.AdminURL) // superuser: not bound by RLS
	orgID, userID := f.id(), f.id()
	for sql, id := range map[string]ids.ID{
		"INSERT INTO iam.organizations VALUES ($1, $2, 'Bad Slug', 'Org', now())":          orgID,
		"INSERT INTO iam.users VALUES ($1, $2, 'Ana@EXAMPLE.com', false, 'active', now())": userID,
	} {
		if _, err := admin.Exec(t.Context(), sql, id.String(), tenant.String()); err != nil {
			t.Fatalf("seed corrupt row: %v", err)
		}
	}
	err := store.InTenant(t.Context(), tenant, func(ctx context.Context, tx storage.Tx) error {
		_, err := tx.Organizations().Get(ctx, ids.OrganizationID{ID: orgID})
		wantCorrupt(t, "Organizations().Get", err)
		_, err = tx.Organizations().List(ctx, ids.OrganizationID{}, 10)
		wantCorrupt(t, "Organizations().List", err)
		_, err = tx.Users().Get(ctx, ids.UserID{ID: userID})
		wantCorrupt(t, "Users().Get", err)
		_, err = tx.Users().List(ctx, ids.UserID{}, 10)
		wantCorrupt(t, "Users().List", err)
		return nil
	})
	if err != nil {
		t.Fatalf("InTenant: %v", err)
	}
}

func wantCorrupt(t *testing.T, what string, err error) {
	t.Helper()
	if !errors.Is(err, postgres.ErrCorrupt) {
		t.Errorf("%s: error = %v, want ErrCorrupt", what, err)
	}
	if errors.Is(err, storage.ErrNotFound) || errors.Is(err, tenancy.ErrInvalid) {
		t.Errorf("%s: corrupt row reported as %v", what, err)
	}
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "example.com") {
		t.Errorf("%s: error carries the stored value: %v", what, err)
	}
}

// TestStore_ConflictErrorsCarryNoDriverText: the driver's message, its detail
// (which quotes the duplicate value) and the constraint name never reach the
// caller.
func TestStore_ConflictErrorsCarryNoDriverText(t *testing.T) {
	f := newFixture(t)
	store := postgres.NewStore(appPool(t, migratedDB(t)))
	tenant := f.createTenant(store, "acme")
	user := func() tenancy.User {
		return tenancy.User{ID: ids.UserID{ID: f.id()}, Email: "zq7mark@example.com", Status: tenancy.StatusActive, CreatedAt: f.now}
	}
	err := store.InTenant(t.Context(), tenant, func(ctx context.Context, tx storage.Tx) error {
		if err := tx.Users().Create(ctx, user()); err != nil {
			return err
		}
		err := tx.Users().Create(ctx, user())
		if !errors.Is(err, storage.ErrConflict) {
			t.Errorf("duplicate e-mail: %v, want ErrConflict", err)
		}
		for _, leak := range []string{"zq7mark", "duplicate key", "users_tenant_id_email_key", "SQLSTATE", "Key ("} {
			if err != nil && strings.Contains(err.Error(), leak) {
				t.Errorf("error %q contains %q", err, leak)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("InTenant: %v", err)
	}
}

// TestSanitize covers the driver errors the database tests cannot provoke on
// demand.
func TestSanitize(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505", Message: "duplicate zq7mark", Detail: "Key (email)=(zq7mark@example.com) already exists.", ConstraintName: "users_tenant_id_email_key"}
	for _, tc := range []struct {
		name   string
		in     error
		is     error
		absent string
	}{
		{"server error", fmt.Errorf("wrapped: %w", pgErr), nil, "zq7mark"},
		{"canceled", fmt.Errorf("query: %w", context.Canceled), context.Canceled, ""},
		{"deadline", context.DeadlineExceeded, context.DeadlineExceeded, ""},
		{"closed tx", pgx.ErrTxClosed, storage.ErrTxDone, ""},
		{"other", errors.New("cannot encode zq7mark"), nil, "zq7mark"},
	} {
		got := postgres.Sanitize(tc.in)
		if got == nil {
			t.Errorf("%s: nil", tc.name)
			continue
		}
		if tc.is != nil && !errors.Is(got, tc.is) {
			t.Errorf("%s: %v is not %v", tc.name, got, tc.is)
		}
		if tc.absent != "" && strings.Contains(got.Error(), tc.absent) {
			t.Errorf("%s: %q leaks the cause", tc.name, got)
		}
	}
	if postgres.Sanitize(nil) != nil {
		t.Error("Sanitize(nil) != nil")
	}
	var asPg *pgconn.PgError
	if !errors.As(postgres.Sanitize(pgErr), &asPg) || asPg.Code != "23505" {
		t.Error("the cause is no longer reachable with errors.As")
	}
}
