package storagetest

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

func txTests() []test {
	return []test{
		{"Tx/RollsBackOnError", testTxRollsBackOnError},
		{"Tx/RollsBackOnPanic", testTxRollsBackOnPanic},
		{"Tx/SystemRollsBackOnError", testSystemTxRollsBackOnError},
		{"Tx/SystemRollsBackOnPanic", testSystemTxRollsBackOnPanic},
		{"Tx/NestingIsRefused", testTxNestingIsRefused},
		{"Tx/UseAfterReturnFails", testTxUseAfterReturnFails},
		{"Tx/CanceledContext", testTxCanceledContext},
		{"Tx/CanceledDuringFnRollsBack", testTxCanceledDuringFn},
		{"Tx/ConcurrentTenants", testTxConcurrentTenants},
	}
}

var errBoom = errors.New("storagetest: boom")

// createAll creates one row of every kind in tx.
func (f *fixture) createAll(ctx context.Context, tx storage.Tx) error {
	o, u := f.org("sales"), f.user("ana@example.com")
	return errors.Join(
		tx.Organizations().Create(ctx, o),
		tx.Users().Create(ctx, u),
		tx.Applications().Create(ctx, f.app("App")),
		tx.Memberships().Add(ctx, f.membership(o.ID, u.ID)),
	)
}

// wantEmptyTenant checks that tenant has no organization, user or application.
func (f *fixture) wantEmptyTenant(tenant ids.TenantID) {
	f.t.Helper()
	f.inTenant(tenant, func(ctx context.Context, tx storage.Tx) error {
		orgs, err1 := tx.Organizations().List(ctx, ids.OrganizationID{}, storage.MaxListLimit)
		users, err2 := tx.Users().List(ctx, ids.UserID{}, storage.MaxListLimit)
		apps, err3 := tx.Applications().List(ctx, ids.ApplicationID{}, storage.MaxListLimit)
		if err := errors.Join(err1, err2, err3); err != nil {
			return err
		}
		if len(orgs)+len(users)+len(apps) != 0 {
			f.t.Errorf("tenant holds %d organizations, %d users, %d applications after rollback; want none", len(orgs), len(users), len(apps))
		}
		return nil
	})
}

func testTxRollsBackOnError(f *fixture) {
	a := f.tenant("acme")
	err := f.store.InTenant(f.ctx, a, func(ctx context.Context, tx storage.Tx) error {
		if err := f.createAll(ctx, tx); err != nil {
			f.t.Fatalf("createAll: %v", err)
		}
		return fmt.Errorf("wrapped: %w", errBoom)
	})
	wantErr(f.t, "InTenant", err, errBoom)
	f.wantEmptyTenant(a)
}

func testTxRollsBackOnPanic(f *fixture) {
	a := f.tenant("acme")
	recovered := func() (r any) {
		defer func() { r = recover() }()
		_ = f.store.InTenant(f.ctx, a, func(ctx context.Context, tx storage.Tx) error {
			if err := f.createAll(ctx, tx); err != nil {
				f.t.Fatalf("createAll: %v", err)
			}
			panic(errBoom)
		})
		return nil
	}()
	if recovered != errBoom { //nolint:errorlint // the panic value itself must propagate
		f.t.Fatalf("recovered %v, want the original panic value", recovered)
	}
	f.wantEmptyTenant(a) // also proves the store is usable after a panic
}

func testSystemTxRollsBackOnError(f *fixture) {
	t := f.tenantValue("acme")
	err := f.store.InSystem(f.ctx, func(ctx context.Context, tx storage.SystemTx) error {
		if err := tx.CreateTenant(ctx, t); err != nil {
			f.t.Fatalf("CreateTenant: %v", err)
		}
		if _, err := tx.EnsureSystemTenant(ctx, ids.TenantID{ID: f.newID()}, systemName, f.clk); err != nil {
			f.t.Fatalf("EnsureSystemTenant: %v", err)
		}
		return errBoom
	})
	wantErr(f.t, "InSystem", err, errBoom)
	f.wantNoTenants()
}

func testSystemTxRollsBackOnPanic(f *fixture) {
	t := f.tenantValue("acme")
	recovered := func() (r any) {
		defer func() { r = recover() }()
		_ = f.store.InSystem(f.ctx, func(ctx context.Context, tx storage.SystemTx) error {
			if err := tx.CreateTenant(ctx, t); err != nil {
				f.t.Fatalf("CreateTenant: %v", err)
			}
			panic(errBoom)
		})
		return nil
	}()
	if recovered != errBoom { //nolint:errorlint // the panic value itself must propagate
		f.t.Fatalf("recovered %v, want the original panic value", recovered)
	}
	f.wantNoTenants()
}

func (f *fixture) wantNoTenants() {
	f.t.Helper()
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		all, err := tx.ListTenants(ctx, ids.TenantID{}, storage.MaxListLimit)
		if err == nil && len(all) != 0 {
			f.t.Errorf("ListTenants after rollback = %+v, want none", all)
		}
		_, sysErr := tx.SystemTenant(ctx)
		wantErr(f.t, "SystemTenant after rollback", sysErr, storage.ErrNotFound)
		return err
	})
}

func testTxNestingIsRefused(f *fixture) {
	a := f.tenant("acme")
	noop := func(context.Context, storage.Tx) error { return nil }
	noopSys := func(context.Context, storage.SystemTx) error { return nil }
	o := f.org("sales")
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		wantErr(f.t, "InTenant in InTenant", f.store.InTenant(ctx, a, noop), storage.ErrNestedTx)
		wantErr(f.t, "InSystem in InTenant", f.store.InSystem(ctx, noopSys), storage.ErrNestedTx)
		return tx.Organizations().Create(ctx, o) // the outer Tx still works
	})
	t := f.tenantValue("other")
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		wantErr(f.t, "InTenant in InSystem", f.store.InTenant(ctx, a, noop), storage.ErrNestedTx)
		wantErr(f.t, "InSystem in InSystem", f.store.InSystem(ctx, noopSys), storage.ErrNestedTx)
		return tx.CreateTenant(ctx, t)
	})
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		_, err := tx.Organizations().Get(ctx, o.ID)
		wantErr(f.t, "outer write committed", err, nil)
		return nil
	})
}

func testTxUseAfterReturnFails(f *fixture) {
	a := f.tenant("acme")
	var kept storage.Tx
	var users storage.UserRepo
	f.inTenant(a, func(_ context.Context, tx storage.Tx) error {
		kept, users = tx, tx.Users()
		return nil
	})
	ctx := f.ctx
	wantErr(f.t, "Organizations().Create", kept.Organizations().Create(ctx, f.org("late")), storage.ErrTxDone)
	_, err := kept.Organizations().List(ctx, ids.OrganizationID{}, 10)
	wantErr(f.t, "Organizations().List", err, storage.ErrTxDone)
	wantErr(f.t, "kept UserRepo Create", users.Create(ctx, f.user("late@example.com")), storage.ErrTxDone)
	_, err = kept.Applications().Get(ctx, ids.ApplicationID{ID: f.newID()})
	wantErr(f.t, "Applications().Get", err, storage.ErrTxDone)
	_, err = kept.Memberships().ListByUser(ctx, ids.UserID{ID: f.newID()})
	wantErr(f.t, "Memberships().ListByUser", err, storage.ErrTxDone)

	var keptSys storage.SystemTx
	f.inSystem(func(_ context.Context, tx storage.SystemTx) error { keptSys = tx; return nil })
	wantErr(f.t, "CreateTenant", keptSys.CreateTenant(ctx, f.tenantValue("late")), storage.ErrTxDone)
	_, err = keptSys.Tenant(ctx, a)
	wantErr(f.t, "Tenant", err, storage.ErrTxDone)
	_, err = keptSys.EnsureSystemTenant(ctx, ids.TenantID{ID: f.newID()}, systemName, f.clk)
	wantErr(f.t, "EnsureSystemTenant", err, storage.ErrTxDone)
	f.wantEmptyTenant(a)
}

func testTxCanceledContext(f *fixture) {
	a := f.tenant("acme")
	ctx, cancel := context.WithCancel(f.ctx)
	cancel()
	ran := false
	err := f.store.InTenant(ctx, a, func(context.Context, storage.Tx) error { ran = true; return nil })
	wantErr(f.t, "InTenant", err, context.Canceled)
	err = f.store.InSystem(ctx, func(context.Context, storage.SystemTx) error { ran = true; return nil })
	wantErr(f.t, "InSystem", err, context.Canceled)
	if ran {
		f.t.Error("fn ran with a canceled context")
	}
}

// testTxCanceledDuringFn: a ctx that ends while fn runs makes the commit
// fail, as it would for a database transaction bound to ctx.
func testTxCanceledDuringFn(f *fixture) {
	a := f.tenant("acme")
	ctx, cancel := context.WithCancel(f.ctx)
	defer cancel()
	err := f.store.InTenant(ctx, a, func(ctx context.Context, tx storage.Tx) error {
		if err := f.createAll(ctx, tx); err != nil {
			f.t.Fatalf("createAll: %v", err)
		}
		cancel()
		return nil
	})
	wantErr(f.t, "InTenant", err, context.Canceled)
	f.wantEmptyTenant(a)
}

// testTxConcurrentTenants runs transactions on two tenants from several
// goroutines at once (run with -race); every write lands, in its own tenant.
func testTxConcurrentTenants(f *fixture) {
	const workers, perWorker = 4, 10
	tenants := []ids.TenantID{f.tenant("tenant-a"), f.tenant("tenant-b")}
	users := make([][]tenancy.User, workers)
	for w := range workers {
		for i := range perWorker {
			users[w] = append(users[w], f.user(fmt.Sprintf("w%d-u%d@example.com", w, i)))
		}
	}
	errs := make(chan error, workers*perWorker)
	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			for _, u := range users[w] {
				errs <- f.store.InTenant(f.ctx, tenants[w%2], func(ctx context.Context, tx storage.Tx) error {
					_, err := tx.Users().List(ctx, ids.UserID{}, storage.MaxListLimit)
					return errors.Join(err, tx.Users().Create(ctx, u))
				})
			}
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		wantErr(f.t, "concurrent InTenant", err, nil)
	}
	for _, tenant := range tenants {
		f.wantUsersOnlyOf(tenant, workers/2*perWorker)
	}
}

// wantUsersOnlyOf checks that tenant lists n users, all its own.
func (f *fixture) wantUsersOnlyOf(tenant ids.TenantID, n int) {
	f.t.Helper()
	f.inTenant(tenant, func(ctx context.Context, tx storage.Tx) error {
		list, err := tx.Users().List(ctx, ids.UserID{}, storage.MaxListLimit)
		if err != nil {
			return err
		}
		if len(list) != n {
			f.t.Errorf("tenant has %d users, want %d", len(list), n)
		}
		for _, u := range list {
			if u.TenantID != tenant {
				f.t.Errorf("user of tenant %v listed in %v", u.TenantID, tenant)
			}
		}
		return nil
	})
}
