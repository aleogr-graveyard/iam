package storagetest

import (
	"context"
	"sync"
	"time"

	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

const systemName = "Platform"

func tenantTests() []test {
	return []test{
		{"Tenant/CreateAndGet", testTenantCreateAndGet},
		{"Tenant/Conflicts", testTenantConflicts},
		{"Tenant/CreateValidates", testTenantCreateValidates},
		{"Tenant/NotFound", testTenantNotFound},
		{"InTenant/RejectsZeroOrUnknownTenant", testInTenantRejectsZeroOrUnknownTenant},
		{"SystemTenant/EnsureIsIdempotent", testEnsureSystemTenantIsIdempotent},
		{"SystemTenant/EnsureConcurrently", testEnsureSystemTenantConcurrently},
		{"SystemTenant/EnsureValidates", testEnsureSystemTenantValidates},
		{"SystemTenant/EnsureRefusesExistingTenantID", testEnsureRefusesExistingTenantID},
		{"SystemTenant/SlugRefusedFromInput", testSystemSlugRefusedFromInput},
		{"SystemTenant/UsableInTenant", testSystemTenantUsableInTenant},
	}
}

func testTenantCreateAndGet(f *fixture) {
	want := f.tenantValue("acme")
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		if err := tx.CreateTenant(ctx, want); err != nil {
			return err
		}
		got, err := tx.Tenant(ctx, want.ID)
		wantErr(f.t, "Tenant in the same tx", err, nil)
		wantEqual(f.t, "Tenant in the same tx", got, want)
		return nil
	})
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		got, err := tx.Tenant(ctx, want.ID)
		wantErr(f.t, "Tenant", err, nil)
		wantEqual(f.t, "Tenant", got, want)
		got, err = tx.TenantBySlug(ctx, "acme")
		wantErr(f.t, "TenantBySlug", err, nil)
		wantEqual(f.t, "TenantBySlug", got, want)
		return nil
	})
}

func testTenantConflicts(f *fixture) {
	first := f.tenantValue("acme")
	sameSlug := f.tenantValue("acme")
	sameID := f.tenantValue("other")
	sameID.ID = first.ID
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error { return tx.CreateTenant(ctx, first) })
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		wantErr(f.t, "duplicate slug", tx.CreateTenant(ctx, sameSlug), storage.ErrConflict)
		wantErr(f.t, "duplicate ID", tx.CreateTenant(ctx, sameID), storage.ErrConflict)
		return nil
	})
}

func testTenantCreateValidates(f *fixture) {
	f.tenant("acme")
	cases := map[string]func(*tenancy.Tenant){
		"zero ID":          func(t *tenancy.Tenant) { t.ID = ids.TenantID{} },
		"bad slug":         func(t *tenancy.Tenant) { t.Slug = "Bad Slug" },
		"reserved slug":    func(t *tenancy.Tenant) { t.Slug = "admin" },
		"empty name":       func(t *tenancy.Tenant) { t.Name = "" },
		"unknown status":   func(t *tenancy.Tenant) { t.Status = "deleted" },
		"zero created_at":  func(t *tenancy.Tenant) { t.CreatedAt = time.Time{} },
		"invalid with dup": func(t *tenancy.Tenant) { t.Slug, t.Name = "acme", "" },
	}
	for what, mutate := range cases {
		bad := f.tenantValue("fresh")
		mutate(&bad)
		f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
			wantInvalid(f.t, what, tx.CreateTenant(ctx, bad))
			return nil
		})
	}
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		_, err := tx.TenantBySlug(ctx, "fresh")
		wantErr(f.t, "no invalid tenant stored", err, storage.ErrNotFound)
		return nil
	})
}

func testTenantNotFound(f *fixture) {
	f.tenant("acme")
	unknown := ids.TenantID{ID: f.newID()}
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		_, err := tx.Tenant(ctx, ids.TenantID{})
		wantErr(f.t, "Tenant(zero)", err, storage.ErrNotFound)
		_, err = tx.Tenant(ctx, unknown)
		wantErr(f.t, "Tenant(unknown)", err, storage.ErrNotFound)
		for _, slug := range []string{"nope", "", "Acme", "acme ", "admin", tenancy.SystemTenantSlug} {
			_, err = tx.TenantBySlug(ctx, slug)
			wantErr(f.t, "TenantBySlug("+slug+")", err, storage.ErrNotFound)
		}
		return nil
	})
}

func testInTenantRejectsZeroOrUnknownTenant(f *fixture) {
	f.tenant("acme")
	for what, id := range map[string]ids.TenantID{"zero": {}, "unknown": {ID: f.newID()}} {
		ran := false
		err := f.store.InTenant(f.ctx, id, func(context.Context, storage.Tx) error { ran = true; return nil })
		wantErr(f.t, what+" tenant", err, storage.ErrInvalidTenant)
		if ran {
			f.t.Errorf("%s tenant: fn ran", what)
		}
	}
}

func (f *fixture) ensureSystem(id ids.TenantID) tenancy.Tenant {
	f.t.Helper()
	var got tenancy.Tenant
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		var err error
		got, err = tx.EnsureSystemTenant(ctx, id, systemName, f.clk)
		return err
	})
	return got
}

func testEnsureSystemTenantIsIdempotent(f *fixture) {
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		_, err := tx.SystemTenant(ctx)
		wantErr(f.t, "SystemTenant before Ensure", err, storage.ErrNotFound)
		return nil
	})
	first := f.ensureSystem(ids.TenantID{ID: f.newID()})
	if !first.IsSystem() || first.Status != tenancy.StatusActive || first.Name != systemName {
		f.t.Fatalf("EnsureSystemTenant = %+v, want the active system tenant", first)
	}
	second := f.ensureSystem(ids.TenantID{ID: f.newID()})
	wantEqual(f.t, "second EnsureSystemTenant", second, first)
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		got, err := tx.SystemTenant(ctx)
		wantErr(f.t, "SystemTenant", err, nil)
		wantEqual(f.t, "SystemTenant", got, first)
		got, err = tx.Tenant(ctx, first.ID)
		wantErr(f.t, "Tenant(system)", err, nil)
		wantEqual(f.t, "Tenant(system)", got, first)
		_, err = tx.TenantBySlug(ctx, tenancy.SystemTenantSlug)
		wantErr(f.t, "TenantBySlug(system)", err, storage.ErrNotFound)
		all, err := tx.ListTenants(ctx, ids.TenantID{}, storage.MaxListLimit)
		wantErr(f.t, "ListTenants", err, nil)
		if len(all) != 1 || all[0].ID != first.ID {
			f.t.Errorf("ListTenants = %+v, want only the system tenant", all)
		}
		return nil
	})
}

func testEnsureSystemTenantConcurrently(f *fixture) {
	const n = 8
	candidates := make([]ids.TenantID, n)
	for i := range candidates {
		candidates[i] = ids.TenantID{ID: f.newID()}
	}
	got := make([]ids.TenantID, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			errs[i] = f.store.InSystem(f.ctx, func(ctx context.Context, tx storage.SystemTx) error {
				st, err := tx.EnsureSystemTenant(ctx, candidates[i], systemName, f.clk)
				got[i] = st.ID
				return err
			})
		})
	}
	wg.Wait()
	for i := range n {
		if errs[i] != nil || got[i] != got[0] || got[0].IsZero() {
			f.t.Errorf("call %d: ID %v, error %v; want %v, nil", i, got[i], errs[i], got[0])
		}
	}
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		all, err := tx.ListTenants(ctx, ids.TenantID{}, storage.MaxListLimit)
		if err == nil && len(all) != 1 {
			f.t.Errorf("ListTenants has %d tenants, want 1", len(all))
		}
		return err
	})
}

func testEnsureSystemTenantValidates(f *fixture) {
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		_, err := tx.EnsureSystemTenant(ctx, ids.TenantID{}, systemName, f.clk)
		wantInvalid(f.t, "zero ID", err)
		_, err = tx.EnsureSystemTenant(ctx, ids.TenantID{ID: f.newID()}, "", f.clk)
		wantInvalid(f.t, "empty name", err)
		_, err = tx.EnsureSystemTenant(ctx, ids.TenantID{ID: f.newID()}, systemName, clock.Clock(nil))
		wantInvalid(f.t, "nil clock", err)
		_, err = tx.SystemTenant(ctx)
		wantErr(f.t, "SystemTenant after refused Ensure calls", err, storage.ErrNotFound)
		return nil
	})
}

func testSystemSlugRefusedFromInput(f *fixture) {
	input := f.tenantValue(tenancy.SystemTenantSlug)
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		wantInvalid(f.t, "system slug before Ensure", tx.CreateTenant(ctx, input))
		return nil
	})
	sys := f.ensureSystem(ids.TenantID{ID: f.newID()})
	again := f.tenantValue(tenancy.SystemTenantSlug)
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		wantInvalid(f.t, "system slug after Ensure", tx.CreateTenant(ctx, again))
		got, err := tx.SystemTenant(ctx)
		wantErr(f.t, "SystemTenant", err, nil)
		wantEqual(f.t, "SystemTenant", got, sys)
		return nil
	})
}

func testSystemTenantUsableInTenant(f *fixture) {
	sys := f.ensureSystem(ids.TenantID{ID: f.newID()})
	op := f.user("operator@example.com")
	f.inTenant(sys.ID, func(ctx context.Context, tx storage.Tx) error { return tx.Users().Create(ctx, op) })
	f.inTenant(sys.ID, func(ctx context.Context, tx storage.Tx) error {
		got, err := tx.Users().ByEmail(ctx, "operator@example.com")
		if err == nil && got.TenantID != sys.ID {
			f.t.Errorf("operator TenantID = %v, want the system tenant", got.TenantID)
		}
		return err
	})
}

// testEnsureRefusesExistingTenantID: Ensure never adopts an ordinary tenant as
// the system tenant, whatever ID it is given.
func testEnsureRefusesExistingTenantID(f *fixture) {
	t := f.tenantValue("acme")
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error { return tx.CreateTenant(ctx, t) })
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		_, err := tx.EnsureSystemTenant(ctx, t.ID, systemName, f.clk)
		wantErr(f.t, "Ensure with an existing tenant's ID", err, storage.ErrConflict)
		return nil
	})
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		_, err := tx.SystemTenant(ctx)
		wantErr(f.t, "SystemTenant", err, storage.ErrNotFound)
		got, err := tx.Tenant(ctx, t.ID)
		wantErr(f.t, "Tenant", err, nil)
		wantEqual(f.t, "existing tenant", got, t)
		if got.IsSystem() {
			f.t.Error("existing tenant became the system tenant")
		}
		return nil
	})
}
