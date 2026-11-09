package storagetest

import (
	"context"
	"errors"
	"time"

	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

func timeTests() []test {
	return []test{
		{"Times/StoredAtMicrosecondPrecision", testTimesMicrosecond},
	}
}

// testTimesMicrosecond: a CreatedAt with nanoseconds reads back truncated to
// the microsecond (toward the past, never rounded), for every entity. 999ns
// would round up into the next microsecond.
func testTimesMicrosecond(f *fixture) {
	want := f.clk.Now() // fixture times have microsecond precision
	at := want.Add(999 * time.Nanosecond)
	tn := f.tenantValue("acme")
	tn.CreatedAt = at
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error { return tx.CreateTenant(ctx, tn) })
	var sys tenancy.Tenant
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		var err error
		sys, err = tx.EnsureSystemTenant(ctx, ids.TenantID{ID: f.newID()}, systemName, &clock.Fixed{T: at})
		return err
	})
	wantTime(f, "EnsureSystemTenant result", sys.CreatedAt, want)
	o, u, app := f.org("sales"), f.user("ana@example.com"), f.app("App")
	o.CreatedAt, u.CreatedAt, app.CreatedAt = at, at, at
	m := f.membership(o.ID, u.ID)
	m.CreatedAt = at
	f.inTenant(tn.ID, func(ctx context.Context, tx storage.Tx) error {
		return errors.Join(tx.Organizations().Create(ctx, o), tx.Users().Create(ctx, u),
			tx.Applications().Create(ctx, app), tx.Memberships().Add(ctx, m))
	})
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
		gotT, err1 := tx.Tenant(ctx, tn.ID)
		gotS, err2 := tx.SystemTenant(ctx)
		wantTime(f, "tenant", gotT.CreatedAt, want)
		wantTime(f, "system tenant", gotS.CreatedAt, want)
		return errors.Join(err1, err2)
	})
	f.inTenant(tn.ID, func(ctx context.Context, tx storage.Tx) error {
		gotO, err1 := tx.Organizations().Get(ctx, o.ID)
		gotU, err2 := tx.Users().Get(ctx, u.ID)
		gotA, err3 := tx.Applications().Get(ctx, app.ID)
		gotM, err4 := tx.Memberships().ListByUser(ctx, u.ID)
		if err := errors.Join(err1, err2, err3, err4); err != nil {
			return err
		}
		wantTime(f, "organization", gotO.CreatedAt, want)
		wantTime(f, "user", gotU.CreatedAt, want)
		wantTime(f, "application", gotA.CreatedAt, want)
		if len(gotM) != 1 {
			f.t.Fatalf("memberships = %+v, want 1", gotM)
		}
		wantTime(f, "membership", gotM[0].CreatedAt, want)
		return nil
	})
}

func wantTime(f *fixture, what string, got, want time.Time) {
	f.t.Helper()
	if !got.Equal(want) {
		f.t.Errorf("%s: CreatedAt = %v, want %v (truncated to the microsecond)", what, got, want)
	}
}
