package storagetest

import (
	"bytes"
	"context"
	"time"

	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

func membershipTests() []test {
	return []test{
		{"Membership/AddListRemove", testMembershipAddListRemove},
		{"Membership/AddRequiresOrganizationAndUserInTenant", testMembershipRequiresOrgAndUser},
		{"Membership/DuplicateConflicts", testMembershipDuplicate},
		{"Membership/AddValidates", testMembershipAddValidates},
		{"Membership/ZeroOrUnknownIDsFindNothing", testMembershipZeroOrUnknown},
	}
}

// seedOrgUser creates an organization and a user in tenant and returns them as
// stored.
func (f *fixture) seedOrgUser(tenant ids.TenantID, slug, email string) (tenancy.Organization, tenancy.User) {
	f.t.Helper()
	o, u := f.org(slug), f.user(email)
	f.inTenant(tenant, func(ctx context.Context, tx storage.Tx) error {
		if err := tx.Organizations().Create(ctx, o); err != nil {
			return err
		}
		return tx.Users().Create(ctx, u)
	})
	o.TenantID, u.TenantID = tenant, tenant
	return o, u
}

func testMembershipAddListRemove(f *fixture) {
	a := f.tenant("acme")
	o1, u := f.seedOrgUser(a, "sales", "ana@example.com")
	o2, _ := f.seedOrgUser(a, "support", "bia@example.com")
	m1, m2 := f.membership(o1.ID, u.ID), f.membership(o2.ID, u.ID)
	// Added in reverse order: ListByUser must still sort by organization ID.
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		if err := tx.Memberships().Add(ctx, m2); err != nil {
			return err
		}
		return tx.Memberships().Add(ctx, m1)
	})
	m1.TenantID, m2.TenantID = a, a
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		got, err := tx.Memberships().ListByUser(ctx, u.ID)
		wantErr(f.t, "ListByUser", err, nil)
		wantEqual(f.t, "ListByUser", got, []tenancy.Membership{m1, m2})
		if bytes.Compare(o1.ID.ID[:], o2.ID.ID[:]) >= 0 {
			f.t.Fatal("fixture: organization IDs are not ascending")
		}
		wantErr(f.t, "Remove", tx.Memberships().Remove(ctx, o1.ID, u.ID), nil)
		return nil
	})
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		got, err := tx.Memberships().ListByUser(ctx, u.ID)
		wantErr(f.t, "ListByUser after Remove", err, nil)
		wantEqual(f.t, "ListByUser after Remove", got, []tenancy.Membership{m2})
		wantErr(f.t, "second Remove", tx.Memberships().Remove(ctx, o1.ID, u.ID), storage.ErrNotFound)
		return nil
	})
}

func testMembershipRequiresOrgAndUser(f *fixture) {
	a := f.tenant("acme")
	o, u := f.seedOrgUser(a, "sales", "ana@example.com")
	unknownOrg := ids.OrganizationID{ID: f.newID()}
	unknownUser := ids.UserID{ID: f.newID()}
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		wantErr(f.t, "unknown organization", tx.Memberships().Add(ctx, f.membership(unknownOrg, u.ID)), storage.ErrNotFound)
		wantErr(f.t, "unknown user", tx.Memberships().Add(ctx, f.membership(o.ID, unknownUser)), storage.ErrNotFound)
		got, err := tx.Memberships().ListByUser(ctx, u.ID)
		if err == nil && len(got) != 0 {
			f.t.Errorf("ListByUser = %+v, want none", got)
		}
		return err
	})
}

func testMembershipDuplicate(f *fixture) {
	a := f.tenant("acme")
	o, u := f.seedOrgUser(a, "sales", "ana@example.com")
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		return tx.Memberships().Add(ctx, f.membership(o.ID, u.ID))
	})
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		wantErr(f.t, "duplicate", tx.Memberships().Add(ctx, f.membership(o.ID, u.ID)), storage.ErrConflict)
		return nil
	})
}

func testMembershipAddValidates(f *fixture) {
	a := f.tenant("acme")
	o, u := f.seedOrgUser(a, "sales", "ana@example.com")
	cases := map[string]tenancy.Membership{
		"zero organization": f.membership(ids.OrganizationID{}, u.ID),
		"zero user":         f.membership(o.ID, ids.UserID{}),
		"zero created_at":   {OrganizationID: o.ID, UserID: u.ID, CreatedAt: time.Time{}},
	}
	for what, m := range cases {
		f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
			wantInvalid(f.t, what, tx.Memberships().Add(ctx, m))
			return nil
		})
	}
}

func testMembershipZeroOrUnknown(f *fixture) {
	a := f.tenant("acme")
	o, u := f.seedOrgUser(a, "sales", "ana@example.com")
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		return tx.Memberships().Add(ctx, f.membership(o.ID, u.ID))
	})
	f.inTenant(a, func(ctx context.Context, tx storage.Tx) error {
		wantErr(f.t, "Remove(zero org)", tx.Memberships().Remove(ctx, ids.OrganizationID{}, u.ID), storage.ErrNotFound)
		wantErr(f.t, "Remove(zero user)", tx.Memberships().Remove(ctx, o.ID, ids.UserID{}), storage.ErrNotFound)
		for what, user := range map[string]ids.UserID{"zero": {}, "unknown": {ID: f.newID()}} {
			got, err := tx.Memberships().ListByUser(ctx, user)
			if err != nil || len(got) != 0 {
				f.t.Errorf("ListByUser(%s) = %+v, %v; want none", what, got, err)
			}
		}
		return nil
	})
}
