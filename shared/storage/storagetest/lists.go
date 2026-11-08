package storagetest

import (
	"context"
	"fmt"
	"slices"

	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

// listSpec adapts one List method to plain IDs, so that every list is held to
// the same rules.
type listSpec struct {
	// seed creates n rows, in one transaction, in a fresh store; it returns
	// their IDs in creation order (ascending, see fixture.newID).
	seed func(f *fixture, n int) []ids.ID
	// list calls the List method in the scope seed used.
	list func(f *fixture, after ids.ID, limit int) ([]ids.ID, error)
}

func listTests() []test {
	specs := []struct {
		name string
		spec func(f *fixture) listSpec
	}{
		{"Tenants", tenantList},
		{"Organizations", orgKind.spec},
		{"Users", userKind.spec},
		{"Applications", appKind.spec},
	}
	var out []test
	for _, s := range specs {
		out = append(out,
			test{"List/" + s.name + "/OrderAndExclusiveAfter", func(f *fixture) { testListOrder(f, s.spec(f)) }},
			test{"List/" + s.name + "/Pages", func(f *fixture) { testListPages(f, s.spec(f)) }},
			test{"List/" + s.name + "/RefusesNonPositiveLimit", func(f *fixture) { testListLimit(f, s.spec(f)) }},
			test{"List/" + s.name + "/ClampsToMaxListLimit", func(f *fixture) { testListClamp(f, s.spec(f)) }},
		)
	}
	return out
}

func testListOrder(f *fixture, s listSpec) {
	want := s.seed(f, 5)
	got := mustList(f, s, ids.ID{}, 100)
	if !slices.Equal(got, want) {
		f.t.Errorf("List(zero) = %v, want %v (ascending ID = creation order)", got, want)
	}
	if got := mustList(f, s, want[1], 100); !slices.Equal(got, want[2:]) {
		f.t.Errorf("List(after #1) = %v, want %v (after is exclusive)", got, want[2:])
	}
	if got := mustList(f, s, want[4], 100); len(got) != 0 {
		f.t.Errorf("List(after last) = %v, want none", got)
	}
	beyond := f.newID()
	if got := mustList(f, s, beyond, 100); len(got) != 0 {
		f.t.Errorf("List(after an unknown later ID) = %v, want none", got)
	}
}

func testListPages(f *fixture, s listSpec) {
	want := s.seed(f, 5)
	var got []ids.ID
	after := ids.ID{}
	for range 10 {
		page := mustList(f, s, after, 2)
		if len(page) > 2 {
			f.t.Fatalf("page of %d rows with limit 2", len(page))
		}
		if len(page) == 0 {
			break
		}
		got = append(got, page...)
		after = page[len(page)-1]
	}
	if !slices.Equal(got, want) {
		f.t.Errorf("paging by 2 gave %v, want %v", got, want)
	}
}

func testListLimit(f *fixture, s listSpec) {
	s.seed(f, 1)
	for _, limit := range []int{0, -1} {
		_, err := s.list(f, ids.ID{}, limit)
		wantErr(f.t, fmt.Sprintf("limit %d", limit), err, storage.ErrInvalid)
	}
}

func testListClamp(f *fixture, s listSpec) {
	want := s.seed(f, storage.MaxListLimit+1)
	got := mustList(f, s, ids.ID{}, storage.MaxListLimit+50)
	if !slices.Equal(got, want[:storage.MaxListLimit]) {
		f.t.Errorf("List(limit MaxListLimit+50) returned %d rows, want the first %d", len(got), storage.MaxListLimit)
	}
}

func mustList(f *fixture, s listSpec, after ids.ID, limit int) []ids.ID {
	f.t.Helper()
	got, err := s.list(f, after, limit)
	if err != nil {
		f.t.Fatalf("List: %v", err)
	}
	return got
}

func tenantList(*fixture) listSpec {
	return listSpec{
		seed: func(f *fixture, n int) []ids.ID {
			out := make([]ids.ID, n)
			f.inSystem(func(ctx context.Context, tx storage.SystemTx) error {
				for i := range n {
					t := f.tenantValue(fmt.Sprintf("tenant-%d", i))
					out[i] = t.ID.ID
					if err := tx.CreateTenant(ctx, t); err != nil {
						return err
					}
				}
				return nil
			})
			return out
		},
		list: func(f *fixture, after ids.ID, limit int) (out []ids.ID, err error) {
			err = f.store.InSystem(f.ctx, func(ctx context.Context, tx storage.SystemTx) error {
				rows, err := tx.ListTenants(ctx, ids.TenantID{ID: after}, limit)
				out = idsOf(rows)
				return err
			})
			return out, err
		},
	}
}

// scopedKind describes one tenant-scoped repository through method
// expressions, so that every list is driven by the same code.
type scopedKind[R, T any, K ~struct{ ids.ID }] struct {
	repo   func(storage.Tx) R
	create func(R, context.Context, T) error
	list   func(R, context.Context, K, int) ([]T, error)
	get    func(R, context.Context, K) (T, error)
	row    func(f *fixture, i int) T // a valid, distinct i-th row
}

var (
	orgKind = scopedKind[storage.OrganizationRepo, tenancy.Organization, ids.OrganizationID]{
		storage.Tx.Organizations, storage.OrganizationRepo.Create, storage.OrganizationRepo.List, storage.OrganizationRepo.Get,
		func(f *fixture, i int) tenancy.Organization { return f.org(fmt.Sprintf("org-%d", i)) },
	}
	userKind = scopedKind[storage.UserRepo, tenancy.User, ids.UserID]{
		storage.Tx.Users, storage.UserRepo.Create, storage.UserRepo.List, storage.UserRepo.Get,
		func(f *fixture, i int) tenancy.User { return f.user(fmt.Sprintf("user%d@example.com", i)) },
	}
	appKind = scopedKind[storage.ApplicationRepo, tenancy.Application, ids.ApplicationID]{
		storage.Tx.Applications, storage.ApplicationRepo.Create, storage.ApplicationRepo.List, storage.ApplicationRepo.Get,
		func(f *fixture, i int) tenancy.Application { return f.app(fmt.Sprintf("App %d", i)) },
	}
)

// spec builds the listSpec of k for one tenant ("acme").
func (k scopedKind[R, T, K]) spec(f *fixture) listSpec {
	tenant := f.tenant("acme")
	return listSpec{
		seed: func(f *fixture, n int) []ids.ID {
			out := make([]ids.ID, n)
			f.inTenant(tenant, func(ctx context.Context, tx storage.Tx) error {
				for i := range n {
					r := k.row(f, i)
					out[i] = rowID(r)
					if err := k.create(k.repo(tx), ctx, r); err != nil {
						return err
					}
				}
				return nil
			})
			return out
		},
		list: func(f *fixture, after ids.ID, limit int) (out []ids.ID, err error) {
			err = f.store.InTenant(f.ctx, tenant, func(ctx context.Context, tx storage.Tx) error {
				rows, err := k.list(k.repo(tx), ctx, K(struct{ ids.ID }{after}), limit)
				out = idsOf(rows)
				return err
			})
			return out, err
		},
	}
}

// rowID returns the ID of any entity the suite lists.
func rowID(row any) ids.ID {
	switch r := row.(type) {
	case tenancy.Tenant:
		return r.ID.ID
	case tenancy.Organization:
		return r.ID.ID
	case tenancy.User:
		return r.ID.ID
	case tenancy.Application:
		return r.ID.ID
	}
	panic(fmt.Sprintf("storagetest: no ID for %T", row))
}

// rowTenant returns the tenant of a tenant-scoped entity.
func rowTenant(row any) ids.TenantID {
	switch r := row.(type) {
	case tenancy.Organization:
		return r.TenantID
	case tenancy.User:
		return r.TenantID
	case tenancy.Application:
		return r.TenantID
	}
	panic(fmt.Sprintf("storagetest: no tenant for %T", row))
}

// withID and withTenant return a copy of row with its ID or tenant replaced.
func withID[T any](row T, id ids.ID) T {
	var out any
	switch r := any(row).(type) {
	case tenancy.Organization:
		r.ID.ID = id
		out = r
	case tenancy.User:
		r.ID.ID = id
		out = r
	case tenancy.Application:
		r.ID.ID = id
		out = r
	}
	return out.(T)
}

func withTenant[T any](row T, tenant ids.TenantID) T {
	var out any
	switch r := any(row).(type) {
	case tenancy.Organization:
		r.TenantID = tenant
		out = r
	case tenancy.User:
		r.TenantID = tenant
		out = r
	case tenancy.Application:
		r.TenantID = tenant
		out = r
	}
	return out.(T)
}

func idsOf[T any](rows []T) []ids.ID {
	out := make([]ids.ID, len(rows))
	for i, r := range rows {
		out[i] = rowID(r)
	}
	return out
}
