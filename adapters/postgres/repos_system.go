package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

// systemTx is the tenant registry. iam.tenants has no tenant_id and no row
// level security; only this type reaches it. sys is the system tenant as of
// the start of the transaction (zero before it exists), the chain its audit
// events go to (ruling R1).
type systemTx struct {
	txConn
	sys ids.TenantID
}

func (tx *systemTx) Audit() storage.AuditAppender { return auditRepo{conn: &tx.txConn, tenant: tx.sys} }

const tenantColumns = "id, slug, name, status, created_at"

func scanTenant(row pgx.Row) (tenancy.Tenant, error) {
	var t tenancy.Tenant
	if err := row.Scan(idDest{&t.ID.ID}, &t.Slug, &t.Name, &t.Status, utcDest{&t.CreatedAt}); err != nil {
		return tenancy.Tenant{}, err
	}
	// ValidateStored: the system tenant, created by NewSystemTenant, is a
	// valid stored row although Validate refuses its slug.
	if err := t.ValidateStored(); err != nil {
		return tenancy.Tenant{}, corrupt("tenant", err)
	}
	return t, nil
}

const insertTenantSQL = `INSERT INTO iam.tenants (id, slug, name, status, created_at) VALUES ($1, $2, $3, $4, $5)`

func tenantArgs(t tenancy.Tenant) []any {
	return []any{uuidArg(t.ID.ID), t.Slug, t.Name, string(t.Status), t.CreatedAt}
}

func (tx *systemTx) CreateTenant(ctx context.Context, t tenancy.Tenant) error {
	if err := tx.check(ctx); err != nil {
		return err
	}
	if err := t.Validate(); err != nil {
		return storage.Invalid(err)
	}
	t.CreatedAt = storage.Microseconds(t.CreatedAt)
	_, err := tx.write(ctx, "tenant", insertTenantSQL, tenantArgs(t)...)
	return err
}

// EnsureSystemTenant is race-safe across connections: the insert yields to a
// concurrent one on the system slug (ON CONFLICT (slug) DO NOTHING waits for
// it to commit), and the tenant is then read back by slug, so every caller
// gets the one that won. An id that names an existing tenant violates the
// primary key, which ON CONFLICT (slug) does not absorb: ErrConflict.
func (tx *systemTx) EnsureSystemTenant(ctx context.Context, id ids.TenantID, name string, clk clock.Clock) (tenancy.Tenant, error) {
	existing, err := tx.SystemTenant(ctx)
	if !errors.Is(err, storage.ErrNotFound) {
		return existing, err
	}
	t, err := tenancy.NewSystemTenant(id, name, clk)
	if err != nil {
		return tenancy.Tenant{}, storage.Invalid(err)
	}
	t.CreatedAt = storage.Microseconds(t.CreatedAt)
	if _, err := tx.write(ctx, "tenant", insertTenantSQL+" ON CONFLICT (slug) DO NOTHING", tenantArgs(t)...); err != nil {
		return tenancy.Tenant{}, err
	}
	return tx.SystemTenant(ctx)
}

func (tx *systemTx) SystemTenant(ctx context.Context) (tenancy.Tenant, error) {
	t, err := tx.bySlug(ctx, tenancy.SystemTenantSlug)
	if err == nil && !t.IsSystem() {
		return tenancy.Tenant{}, corrupt("tenant", errors.New("system slug on a non-system row"))
	}
	return t, err
}

func (tx *systemTx) TenantBySlug(ctx context.Context, slug string) (tenancy.Tenant, error) {
	if err := tx.check(ctx); err != nil {
		return tenancy.Tenant{}, err
	}
	// The system slug is stored but refused here, with every other invalid one.
	if tenancy.ValidateSlug(slug) != nil {
		return tenancy.Tenant{}, notFound("tenant")
	}
	return tx.bySlug(ctx, slug)
}

func (tx *systemTx) bySlug(ctx context.Context, slug string) (tenancy.Tenant, error) {
	return queryOne(ctx, &tx.txConn, "tenant", scanTenant,
		"SELECT "+tenantColumns+" FROM iam.tenants WHERE slug = $1", slug)
}

func (tx *systemTx) Tenant(ctx context.Context, id ids.TenantID) (tenancy.Tenant, error) {
	return queryOne(ctx, &tx.txConn, "tenant", scanTenant,
		"SELECT "+tenantColumns+" FROM iam.tenants WHERE id = $1", uuidArg(id.ID))
}

func (tx *systemTx) ListTenants(ctx context.Context, after ids.TenantID, limit int) ([]tenancy.Tenant, error) {
	if err := tx.check(ctx); err != nil {
		return nil, err
	}
	n, err := storage.ClampLimit(limit)
	if err != nil {
		return nil, err
	}
	return queryAll(ctx, &tx.txConn, scanTenant,
		"SELECT "+tenantColumns+" FROM iam.tenants WHERE id > $1 ORDER BY id LIMIT $2", uuidArg(after.ID), n)
}
