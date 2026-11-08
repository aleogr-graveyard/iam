package storagetest

import (
	"context"
	"errors"

	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
)

// isolationTests are the cross-tenant abuse tests of TM-TEN-001: a
// transaction bound to tenant B never reads, changes or links tenant A's rows.
func isolationTests() []test {
	return []test{
		{"CrossTenantIsolation/Reads", testIsolationReads},
		{"CrossTenantIsolation/Memberships", testIsolationMemberships},
		{"CrossTenantIsolation/CreateForcesTxTenant", testIsolationCreateForcesTxTenant},
		{"CrossTenantIsolation/ListAfterFromOtherTenant", testIsolationListAfterFromOtherTenant},
		{"CrossTenantIsolation/SameIDInOtherTenant", testIsolationSameID},
	}
}

// testIsolationReads: rows of A are invisible to every read in B, and their
// unique keys (slug, e-mail) do not collide with B's.
func testIsolationReads(f *fixture) {
	a, b := f.tenant("tenant-a"), f.tenant("tenant-b")
	o, u, app := f.org("sales"), f.user("ana@example.com"), f.confidentialApp("App")
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		return errors.Join(tx.Organizations().Create(ctx, o), tx.Users().Create(ctx, u), tx.Applications().Create(ctx, app))
	})
	f.inTenant(b, func(ctx context.Context, tx storage.Tx) error {
		_, err := tx.Organizations().Get(ctx, o.ID)
		wantErr(f.t, "Organizations().Get in B", err, storage.ErrNotFound)
		_, err = tx.Organizations().BySlug(ctx, "sales")
		wantErr(f.t, "Organizations().BySlug in B", err, storage.ErrNotFound)
		_, err = tx.Users().Get(ctx, u.ID)
		wantErr(f.t, "Users().Get in B", err, storage.ErrNotFound)
		_, err = tx.Users().ByEmail(ctx, "ana@example.com")
		wantErr(f.t, "Users().ByEmail in B", err, storage.ErrNotFound)
		_, err = tx.Applications().Get(ctx, app.ID)
		wantErr(f.t, "Applications().Get in B", err, storage.ErrNotFound)
		return nil
	})
	f.wantEmptyTenant(b)
	f.inTenant(b, func(ctx context.Context, tx storage.Tx) error {
		wantErr(f.t, "same slug in B", tx.Organizations().Create(ctx, f.org("sales")), nil)
		wantErr(f.t, "same e-mail in B", tx.Users().Create(ctx, f.user("ana@example.com")), nil)
		return nil
	})
}

func testIsolationMemberships(f *fixture) {
	a, b := f.tenant("tenant-a"), f.tenant("tenant-b")
	oa, ua := f.seedOrgUser(a, "sales", "ana@example.com")
	ob, ub := f.seedOrgUser(b, "sales", "bia@example.com")
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		return tx.Memberships().Add(ctx, f.membership(oa.ID, ua.ID))
	})
	// Linking across tenants: organization of A with user of B, in either Tx.
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		wantErr(f.t, "org A + user B in A", tx.Memberships().Add(ctx, f.membership(oa.ID, ub.ID)), storage.ErrNotFound)
		return nil
	})
	f.inTenant(b, func(ctx context.Context, tx storage.Tx) error {
		wantErr(f.t, "org A + user B in B", tx.Memberships().Add(ctx, f.membership(oa.ID, ub.ID)), storage.ErrNotFound)
		wantErr(f.t, "org B + user A in B", tx.Memberships().Add(ctx, f.membership(ob.ID, ua.ID)), storage.ErrNotFound)
		wantErr(f.t, "org A + user A in B", tx.Memberships().Add(ctx, f.membership(oa.ID, ua.ID)), storage.ErrNotFound)
		got, err := tx.Memberships().ListByUser(ctx, ua.ID)
		if err != nil || len(got) != 0 {
			f.t.Errorf("ListByUser(user A) in B = %+v, %v; want none", got, err)
		}
		wantErr(f.t, "Remove A's membership in B", tx.Memberships().Remove(ctx, oa.ID, ua.ID), storage.ErrNotFound)
		got, err = tx.Memberships().ListByUser(ctx, ub.ID)
		if err != nil || len(got) != 0 {
			f.t.Errorf("ListByUser(user B) in B = %+v, %v; want none", got, err)
		}
		return nil
	})
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		got, err := tx.Memberships().ListByUser(ctx, ua.ID)
		if err != nil || len(got) != 1 || got[0].OrganizationID != oa.ID || got[0].TenantID != a {
			f.t.Errorf("ListByUser(user A) in A = %+v, %v; want the one membership", got, err)
		}
		return nil
	})
}

// testIsolationCreateForcesTxTenant: a TenantID field naming another tenant is
// overwritten with the Tx's tenant.
func testIsolationCreateForcesTxTenant(f *fixture) {
	a, b := f.tenant("tenant-a"), f.tenant("tenant-b")
	o, u, app := f.org("sales"), f.user("ana@example.com"), f.app("App")
	o.TenantID, u.TenantID, app.TenantID = a, a, a
	f.inTenant(b, func(ctx context.Context, tx storage.Tx) error {
		if err := errors.Join(tx.Organizations().Create(ctx, o), tx.Users().Create(ctx, u), tx.Applications().Create(ctx, app)); err != nil {
			return err
		}
		m := f.membership(o.ID, u.ID)
		m.TenantID = a
		return tx.Memberships().Add(ctx, m)
	})
	f.inTenant(b, func(ctx context.Context, tx storage.Tx) error {
		gotO, err1 := tx.Organizations().Get(ctx, o.ID)
		gotU, err2 := tx.Users().Get(ctx, u.ID)
		gotA, err3 := tx.Applications().Get(ctx, app.ID)
		ms, err4 := tx.Memberships().ListByUser(ctx, u.ID)
		if err := errors.Join(err1, err2, err3, err4); err != nil {
			return err
		}
		if gotO.TenantID != b || gotU.TenantID != b || gotA.TenantID != b || len(ms) != 1 || ms[0].TenantID != b {
			f.t.Errorf("rows created in B carry tenants %v, %v, %v, %+v; want B", gotO.TenantID, gotU.TenantID, gotA.TenantID, ms)
		}
		return nil
	})
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		_, err := tx.Organizations().Get(ctx, o.ID)
		wantErr(f.t, "organization in A", err, storage.ErrNotFound)
		_, err = tx.Users().Get(ctx, u.ID)
		wantErr(f.t, "user in A", err, storage.ErrNotFound)
		_, err = tx.Applications().Get(ctx, app.ID)
		wantErr(f.t, "application in A", err, storage.ErrNotFound)
		ms, err := tx.Memberships().ListByUser(ctx, u.ID)
		if err != nil || len(ms) != 0 {
			f.t.Errorf("memberships in A = %+v, %v; want none", ms, err)
		}
		return nil
	})
}

// testIsolationListAfterFromOtherTenant: an `after` taken from tenant B is a
// position only; the page holds A's rows past it and nothing of B.
func testIsolationListAfterFromOtherTenant(f *fixture) {
	a, b := f.tenant("tenant-a"), f.tenant("tenant-b")
	var aIDs []ids.OrganizationID
	var bID ids.OrganizationID
	for _, slug := range []string{"a-1", "a-2", "b-1", "a-3", "a-4"} {
		o := f.org(slug)
		tenant := a
		if slug == "b-1" {
			tenant, bID = b, o.ID
		} else {
			aIDs = append(aIDs, o.ID)
		}
		f.inTenant(tenant, func(ctx context.Context, tx storage.Tx) error { return tx.Organizations().Create(ctx, o) })
	}
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		got, err := tx.Organizations().List(ctx, bID, storage.MaxListLimit)
		if err != nil {
			return err
		}
		var gotIDs []ids.OrganizationID
		for _, o := range got {
			if o.TenantID != a {
				f.t.Errorf("row of tenant %v in A's page", o.TenantID)
			}
			gotIDs = append(gotIDs, o.ID)
		}
		want := []ids.OrganizationID{aIDs[2], aIDs[3]}
		if len(gotIDs) != len(want) || gotIDs[0] != want[0] || gotIDs[1] != want[1] {
			f.t.Errorf("List(after B's ID) = %v, want %v", gotIDs, want)
		}
		return nil
	})
}

// testIsolationSameID: creating a row whose ID exists in another tenant either
// succeeds in the Tx's tenant or fails with ErrConflict; either way the other
// tenant's row is untouched and stays invisible.
func testIsolationSameID(f *fixture) {
	a, b := f.tenant("tenant-a"), f.tenant("tenant-b")
	orig := f.org("sales")
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error { return tx.Organizations().Create(ctx, orig) })
	orig.TenantID = a
	twin := f.org("twin")
	twin.ID = orig.ID
	err := f.store.InTenant(f.ctx, b, func(ctx context.Context, tx storage.Tx) error { return tx.Organizations().Create(ctx, twin) })
	switch {
	case err == nil:
		f.inTenant(b, func(ctx context.Context, tx storage.Tx) error {
			got, err := tx.Organizations().Get(ctx, twin.ID)
			if err == nil && (got.TenantID != b || got.Slug != "twin") {
				f.t.Errorf("Get in B = %+v, want B's own row", got)
			}
			return err
		})
	case !errors.Is(err, storage.ErrConflict):
		f.t.Errorf("same ID in B: error = %v, want nil or ErrConflict", err)
	}
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		got, err := tx.Organizations().Get(ctx, orig.ID)
		wantErr(f.t, "Get in A", err, nil)
		wantEqual(f.t, "A's row", got, orig)
		return nil
	})
}
