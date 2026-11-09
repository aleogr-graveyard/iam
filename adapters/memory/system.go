package memory

import (
	"context"

	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

// registry is the tenant registry, tenants by ID and by slug.
type registry struct {
	tenants keyed[tenancy.Tenant]
	system  ids.ID // zero until EnsureSystemTenant
}

func newRegistry() registry { return registry{tenants: newKeyed[tenancy.Tenant]("tenant", "slug")} }

func (r registry) clone() registry { return registry{tenants: r.tenants.clone(), system: r.system} }

type systemTx struct {
	txState
	reg registry
}

func (tx *systemTx) CreateTenant(ctx context.Context, t tenancy.Tenant) error {
	if err := tx.check(ctx); err != nil {
		return err
	}
	if err := t.Validate(); err != nil {
		return storage.Invalid(err)
	}
	t.CreatedAt = storage.Microseconds(t.CreatedAt)
	return tx.reg.tenants.insert(t.ID.ID, t.Slug, t)
}

func (tx *systemTx) EnsureSystemTenant(ctx context.Context, id ids.TenantID, name string, clk clock.Clock) (tenancy.Tenant, error) {
	if err := tx.check(ctx); err != nil {
		return tenancy.Tenant{}, err
	}
	if !tx.reg.system.IsZero() {
		return tx.reg.tenants.byID[tx.reg.system], nil
	}
	t, err := tenancy.NewSystemTenant(id, name, clk)
	if err != nil {
		return tenancy.Tenant{}, storage.Invalid(err)
	}
	t.CreatedAt = storage.Microseconds(t.CreatedAt)
	if err := tx.reg.tenants.insert(t.ID.ID, t.Slug, t); err != nil {
		return tenancy.Tenant{}, err
	}
	tx.reg.system = t.ID.ID
	return t, nil
}

func (tx *systemTx) SystemTenant(ctx context.Context) (tenancy.Tenant, error) {
	if err := tx.check(ctx); err != nil {
		return tenancy.Tenant{}, err
	}
	if tx.reg.system.IsZero() {
		return tenancy.Tenant{}, notFound("tenant")
	}
	return tx.reg.tenants.byID[tx.reg.system], nil
}

func (tx *systemTx) TenantBySlug(ctx context.Context, slug string) (tenancy.Tenant, error) {
	if err := tx.check(ctx); err != nil {
		return tenancy.Tenant{}, err
	}
	// The system slug is stored but refused here, with every other invalid one.
	if tenancy.ValidateSlug(slug) != nil {
		return tenancy.Tenant{}, notFound("tenant")
	}
	return tx.reg.tenants.lookup(slug)
}

func (tx *systemTx) Tenant(ctx context.Context, id ids.TenantID) (tenancy.Tenant, error) {
	if err := tx.check(ctx); err != nil {
		return tenancy.Tenant{}, err
	}
	return get(tx.reg.tenants.byID, id.ID, "tenant", same)
}

func (tx *systemTx) ListTenants(ctx context.Context, after ids.TenantID, limit int) ([]tenancy.Tenant, error) {
	if err := tx.check(ctx); err != nil {
		return nil, err
	}
	return page(tx.reg.tenants.byID, after.ID, limit, same)
}
