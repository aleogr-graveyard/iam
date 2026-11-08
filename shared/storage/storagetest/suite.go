// Package storagetest is the contract suite of the storage ports (contract
// C3). Every adapter runs it from its own tests:
//
//	func TestStore_Contract(t *testing.T) {
//		storagetest.Run(t, func(t *testing.T) storage.Store { return memory.New() })
//	}
//
// Subtest names are stable, so a failure names the same rule in every
// adapter. Production code must not import this package (depguard).
package storagetest

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

// t0 has microsecond precision, the precision of a Postgres timestamptz.
var t0 = time.Date(2026, 11, 7, 12, 0, 0, 123456000, time.UTC)

type test struct {
	name string
	fn   func(f *fixture)
}

// Run runs every contract test against fresh stores from newStore: each
// subtest calls it once (more than once only where it says so) and expects an
// empty store.
func Run(t *testing.T, newStore func(t *testing.T) storage.Store) {
	t.Helper()
	groups := [][]test{tenantTests(), repoTests(), membershipTests(), listTests(), isolationTests(), scopedIsolationTests(), txTests(), errorTests()}
	for _, g := range groups {
		for _, tc := range g {
			t.Run(tc.name, func(t *testing.T) {
				tc.fn(newFixture(t, newStore(t)))
			})
		}
	}
}

// fixture builds valid entities and runs transactions that must succeed.
type fixture struct {
	t     *testing.T
	ctx   context.Context
	store storage.Store
	clk   *clock.Fixed
	gen   ids.Generator
}

func newFixture(t *testing.T, s storage.Store) *fixture {
	clk := &clock.Fixed{T: t0}
	return &fixture{t: t, ctx: t.Context(), store: s, clk: clk, gen: ids.Generator{Clock: clk}}
}

// newID advances the clock by one millisecond first, so that IDs sort in the
// order they were made.
func (f *fixture) newID() ids.ID {
	f.t.Helper()
	f.clk.Advance(time.Millisecond)
	id, err := f.gen.New()
	if err != nil {
		f.t.Fatal(err)
	}
	return id
}

func (f *fixture) tenantValue(slug string) tenancy.Tenant {
	return tenancy.Tenant{ID: ids.TenantID{ID: f.newID()}, Slug: slug, Name: "Tenant " + slug, Status: tenancy.StatusActive, CreatedAt: f.clk.Now()}
}

func (f *fixture) org(slug string) tenancy.Organization {
	return tenancy.Organization{ID: ids.OrganizationID{ID: f.newID()}, Slug: slug, Name: "Org " + slug, CreatedAt: f.clk.Now()}
}

func (f *fixture) user(email string) tenancy.User {
	return tenancy.User{ID: ids.UserID{ID: f.newID()}, Email: email, Status: tenancy.StatusActive, CreatedAt: f.clk.Now()}
}

func (f *fixture) app(name string) tenancy.Application {
	return tenancy.Application{
		ID: ids.ApplicationID{ID: f.newID()}, Name: name, Type: tenancy.ClientPublic,
		RedirectURIs: []string{"https://app.example.com/callback", "http://127.0.0.1:8080/cb"},
		CreatedAt:    f.clk.Now(),
	}
}

func (f *fixture) confidentialApp(name string) tenancy.Application {
	a := f.app(name)
	a.Type = tenancy.ClientConfidential
	a.SecretHash = make([]byte, 32)
	for i := range a.SecretHash {
		a.SecretHash[i] = byte(i + 1)
	}
	return a
}

func (f *fixture) membership(org ids.OrganizationID, user ids.UserID) tenancy.Membership {
	return tenancy.Membership{OrganizationID: org, UserID: user, CreatedAt: f.clk.Now()}
}

// tenant creates a tenant with slug and returns its ID.
func (f *fixture) tenant(slug string) ids.TenantID {
	f.t.Helper()
	t := f.tenantValue(slug)
	f.inSystem(func(ctx context.Context, tx storage.SystemTx) error { return tx.CreateTenant(ctx, t) })
	return t.ID
}

func (f *fixture) inTenant(tenant ids.TenantID, fn func(ctx context.Context, tx storage.Tx) error) {
	f.t.Helper()
	if err := f.store.InTenant(f.ctx, tenant, fn); err != nil {
		f.t.Fatalf("InTenant: %v", err)
	}
}

func (f *fixture) inSystem(fn func(ctx context.Context, tx storage.SystemTx) error) {
	f.t.Helper()
	if err := f.store.InSystem(f.ctx, fn); err != nil {
		f.t.Fatalf("InSystem: %v", err)
	}
}

// --- assertions

func wantErr(t *testing.T, what string, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: error = %v, want %v", what, err, target)
	}
}

// wantInvalid checks a validation error: ErrInvalid and tenancy.ErrInvalid,
// never ErrConflict or ErrNotFound.
func wantInvalid(t *testing.T, what string, err error) {
	t.Helper()
	if !errors.Is(err, storage.ErrInvalid) || !errors.Is(err, tenancy.ErrInvalid) {
		t.Errorf("%s: error = %v, want storage.ErrInvalid wrapping tenancy.ErrInvalid", what, err)
	}
	if errors.Is(err, storage.ErrConflict) || errors.Is(err, storage.ErrNotFound) {
		t.Errorf("%s: validation error %v is also ErrConflict or ErrNotFound", what, err)
	}
}

// wantEqual compares after normalizing times to UTC: adapters may return
// another location for the same instant.
func wantEqual[T any](t *testing.T, what string, got, want T) {
	t.Helper()
	if g, w := normalize(got), normalize(want); !reflect.DeepEqual(g, w) {
		t.Errorf("%s:\n got  %+v\n want %+v", what, g, w)
	}
}

func normalize(v any) any {
	switch e := v.(type) {
	case tenancy.Tenant:
		e.CreatedAt = e.CreatedAt.UTC()
		return e
	case tenancy.Organization:
		e.CreatedAt = e.CreatedAt.UTC()
		return e
	case tenancy.User:
		e.CreatedAt = e.CreatedAt.UTC()
		return e
	case tenancy.Application:
		e.CreatedAt = e.CreatedAt.UTC()
		return e
	case tenancy.Membership:
		e.CreatedAt = e.CreatedAt.UTC()
		return e
	case []tenancy.Membership:
		out := make([]any, len(e))
		for i, m := range e {
			out[i] = normalize(m)
		}
		return out
	}
	return v
}
