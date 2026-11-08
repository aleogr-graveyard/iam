package storagetest

import (
	"context"
	"errors"
	"fmt"
	"slices"

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
	}
}

// scopedIsolationTests run, for every tenant-scoped kind, the isolation rules
// that apply to all of them alike.
func scopedIsolationTests() []test {
	var out []test
	for _, k := range []struct {
		name              string
		listAfter, sameID func(*fixture)
	}{
		{"Organizations", orgKind.testListAfterFromOtherTenant, orgKind.testSameIDInOtherTenant},
		{"Users", userKind.testListAfterFromOtherTenant, userKind.testSameIDInOtherTenant},
		{"Applications", appKind.testListAfterFromOtherTenant, appKind.testSameIDInOtherTenant},
	} {
		out = append(out,
			test{"CrossTenantIsolation/ListAfterFromOtherTenant/" + k.name, k.listAfter},
			test{"CrossTenantIsolation/SameIDInOtherTenant/" + k.name, k.sameID},
		)
	}
	return out
}

// testIsolationReads: rows of A are invisible to every read in B, and their
// unique keys (slug, e-mail) do not collide with B's. It runs with an
// unverified and a verified user, so a lookup that treats verified addresses
// as global is caught.
func testIsolationReads(f *fixture) {
	for i, verified := range []bool{false, true} {
		a, b := f.tenant(fmt.Sprintf("tenant-a-%d", i)), f.tenant(fmt.Sprintf("tenant-b-%d", i))
		o, u, app := f.org("sales"), f.user("ana@example.com"), f.confidentialApp("App")
		u.EmailVerified = verified
		f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
			return errors.Join(tx.Organizations().Create(ctx, o), tx.Users().Create(ctx, u), tx.Applications().Create(ctx, app))
		})
		f.wantInvisible(b, o.ID, u.ID, app.ID, fmt.Sprintf("verified=%v", verified))
		f.wantEmptyTenant(b)
		f.inTenant(b, func(ctx context.Context, tx storage.Tx) error {
			wantErr(f.t, "same slug in B", tx.Organizations().Create(ctx, f.org("sales")), nil)
			wantErr(f.t, "same e-mail in B", tx.Users().Create(ctx, f.user("ana@example.com")), nil)
			return nil
		})
	}
}

// wantInvisible checks that every read in tenant finds none of the rows
// testIsolationReads created in the other tenant.
func (f *fixture) wantInvisible(tenant ids.TenantID, o ids.OrganizationID, u ids.UserID, app ids.ApplicationID, what string) {
	f.t.Helper()
	f.inTenant(tenant, func(ctx context.Context, tx storage.Tx) error {
		_, err := tx.Organizations().Get(ctx, o)
		wantErr(f.t, what+": Organizations().Get in B", err, storage.ErrNotFound)
		_, err = tx.Organizations().BySlug(ctx, "sales")
		wantErr(f.t, what+": Organizations().BySlug in B", err, storage.ErrNotFound)
		_, err = tx.Users().Get(ctx, u)
		wantErr(f.t, what+": Users().Get in B", err, storage.ErrNotFound)
		_, err = tx.Users().ByEmail(ctx, "ana@example.com")
		wantErr(f.t, what+": Users().ByEmail in B", err, storage.ErrNotFound)
		_, err = tx.Applications().Get(ctx, app)
		wantErr(f.t, what+": Applications().Get in B", err, storage.ErrNotFound)
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

// testListAfterFromOtherTenant: an `after` taken from tenant B is a position
// only; the page holds A's rows past it and nothing of B.
func (k scopedKind[R, T, K]) testListAfterFromOtherTenant(f *fixture) {
	a, b := f.tenant("tenant-a"), f.tenant("tenant-b")
	// B's rows sit both at and after the cursor, so a page that mixed tenants
	// would show one.
	var aIDs, bIDs []ids.ID
	for i, tenant := range []ids.TenantID{a, a, b, a, b, a} {
		r := k.row(f, i)
		if tenant == b {
			bIDs = append(bIDs, rowID(r))
		} else {
			aIDs = append(aIDs, rowID(r))
		}
		f.inTenant(tenant, func(ctx context.Context, tx storage.Tx) error { return k.create(k.repo(tx), ctx, r) })
	}
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		rows, err := k.list(k.repo(tx), ctx, K(struct{ ids.ID }{bIDs[0]}), storage.MaxListLimit)
		if err != nil {
			return err
		}
		for _, r := range rows {
			if rowTenant(r) != a {
				f.t.Errorf("row of tenant %v in A's page", rowTenant(r))
			}
		}
		if got := idsOf(rows); !slices.Equal(got, aIDs[2:]) {
			f.t.Errorf("List(after B's ID) = %v, want %v", got, aIDs[2:])
		}
		return nil
	})
}

// testSameIDInOtherTenant: IDs are unique across tenants. Creating a row whose
// ID exists in another tenant is ErrConflict, and the other tenant's row is
// untouched and stays invisible.
func (k scopedKind[R, T, K]) testSameIDInOtherTenant(f *fixture) {
	a, b := f.tenant("tenant-a"), f.tenant("tenant-b")
	orig := k.row(f, 0)
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error { return k.create(k.repo(tx), ctx, orig) })
	id := K(struct{ ids.ID }{rowID(orig)})
	twin := withID(k.row(f, 1), rowID(orig))
	f.inTenant(b, func(ctx context.Context, tx storage.Tx) error {
		wantErr(f.t, "same ID in B", k.create(k.repo(tx), ctx, twin), storage.ErrConflict)
		_, err := k.get(k.repo(tx), ctx, id)
		wantErr(f.t, "Get in B", err, storage.ErrNotFound)
		return nil
	})
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		got, err := k.get(k.repo(tx), ctx, id)
		wantErr(f.t, "Get in A", err, nil)
		wantEqual(f.t, "A's row", got, withTenant(orig, a))
		return nil
	})
}
