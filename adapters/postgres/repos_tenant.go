package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

// tenantTx is a transaction bound to one tenant. Every statement names the
// tenant explicitly (the first barrier); RLS, through iam.tenant_id, is the
// second.
type tenantTx struct {
	txConn
	tenant            ids.TenantID
	allowInsecureHTTP bool
}

func (tx *tenantTx) Organizations() storage.OrganizationRepo { return orgRepo{tx} }
func (tx *tenantTx) Users() storage.UserRepo                 { return userRepo{tx} }
func (tx *tenantTx) Memberships() storage.MembershipRepo     { return memberRepo{tx} }
func (tx *tenantTx) Applications() storage.ApplicationRepo   { return appRepo{tx} }

var errForeignRow = errors.New("row of another tenant")

// own checks that a row read back belongs to the Tx's tenant: a third check,
// after the query filter and RLS, that costs nothing.
func (tx *tenantTx) own(entity string, got ids.TenantID) error {
	if got != tx.tenant {
		return corrupt(entity, errForeignRow)
	}
	return nil
}

// getByID reads the row of table with id, in the Tx's tenant.
func getByID[T any](ctx context.Context, tx *tenantTx, entity, table, columns string, scan scanner[T], id ids.ID) (T, error) {
	return queryOne(ctx, &tx.txConn, entity, scan,
		"SELECT "+columns+" FROM iam."+table+" WHERE tenant_id = $1 AND id = $2",
		uuidArg(tx.tenant.ID), uuidArg(id))
}

// listPage reads a page of table in ascending ID order, strictly after after.
func listPage[T any](ctx context.Context, tx *tenantTx, table, columns string, scan scanner[T], after ids.ID, limit int) ([]T, error) {
	if err := tx.check(ctx); err != nil {
		return nil, err
	}
	n, err := storage.ClampLimit(limit)
	if err != nil {
		return nil, err
	}
	return queryAll(ctx, &tx.txConn, scan,
		"SELECT "+columns+" FROM iam."+table+" WHERE tenant_id = $1 AND id > $2 ORDER BY id LIMIT $3",
		uuidArg(tx.tenant.ID), uuidArg(after), n)
}

// --- organizations

type orgRepo struct{ tx *tenantTx }

const orgColumns = "id, tenant_id, slug, name, created_at"

func (r orgRepo) scan(row pgx.Row) (tenancy.Organization, error) {
	var id, tenant pgtype.UUID
	var o tenancy.Organization
	if err := row.Scan(&id, &tenant, &o.Slug, &o.Name, &o.CreatedAt); err != nil {
		return tenancy.Organization{}, err
	}
	o.ID, o.TenantID, o.CreatedAt = ids.OrganizationID{ID: fromUUID(id)}, ids.TenantID{ID: fromUUID(tenant)}, o.CreatedAt.UTC()
	if err := o.Validate(); err != nil {
		return tenancy.Organization{}, corrupt("organization", err)
	}
	return o, r.tx.own("organization", o.TenantID)
}

func (r orgRepo) Create(ctx context.Context, o tenancy.Organization) error {
	if err := r.tx.check(ctx); err != nil {
		return err
	}
	o.TenantID = r.tx.tenant
	if err := o.Validate(); err != nil {
		return storage.Invalid(err)
	}
	_, err := r.tx.write(ctx, "organization",
		"INSERT INTO iam.organizations ("+orgColumns+") VALUES ($1, $2, $3, $4, $5)",
		uuidArg(o.ID.ID), uuidArg(o.TenantID.ID), o.Slug, o.Name, o.CreatedAt)
	return err
}

func (r orgRepo) Get(ctx context.Context, id ids.OrganizationID) (tenancy.Organization, error) {
	return getByID(ctx, r.tx, "organization", "organizations", orgColumns, r.scan, id.ID)
}

// BySlug refuses an invalid slug before the query: it is never stored, and
// some invalid text (a NUL byte) would make the query fail.
func (r orgRepo) BySlug(ctx context.Context, slug string) (tenancy.Organization, error) {
	if err := r.tx.check(ctx); err != nil {
		return tenancy.Organization{}, err
	}
	if tenancy.ValidateSlug(slug) != nil {
		return tenancy.Organization{}, notFound("organization")
	}
	return queryOne(ctx, &r.tx.txConn, "organization", r.scan,
		"SELECT "+orgColumns+" FROM iam.organizations WHERE tenant_id = $1 AND slug = $2",
		uuidArg(r.tx.tenant.ID), slug)
}

func (r orgRepo) List(ctx context.Context, after ids.OrganizationID, limit int) ([]tenancy.Organization, error) {
	return listPage(ctx, r.tx, "organizations", orgColumns, r.scan, after.ID, limit)
}

// --- users

type userRepo struct{ tx *tenantTx }

const userColumns = "id, tenant_id, email, email_verified, status, created_at"

func (r userRepo) scan(row pgx.Row) (tenancy.User, error) {
	var id, tenant pgtype.UUID
	var u tenancy.User
	if err := row.Scan(&id, &tenant, &u.Email, &u.EmailVerified, &u.Status, &u.CreatedAt); err != nil {
		return tenancy.User{}, err
	}
	u.ID, u.TenantID, u.CreatedAt = ids.UserID{ID: fromUUID(id)}, ids.TenantID{ID: fromUUID(tenant)}, u.CreatedAt.UTC()
	if err := u.Validate(); err != nil {
		return tenancy.User{}, corrupt("user", err)
	}
	return u, r.tx.own("user", u.TenantID)
}

func (r userRepo) Create(ctx context.Context, u tenancy.User) error {
	if err := r.tx.check(ctx); err != nil {
		return err
	}
	u.TenantID = r.tx.tenant
	if err := u.Validate(); err != nil {
		return storage.Invalid(err)
	}
	_, err := r.tx.write(ctx, "user",
		"INSERT INTO iam.users ("+userColumns+") VALUES ($1, $2, $3, $4, $5, $6)",
		uuidArg(u.ID.ID), uuidArg(u.TenantID.ID), u.Email, u.EmailVerified, string(u.Status), u.CreatedAt)
	return err
}

func (r userRepo) Get(ctx context.Context, id ids.UserID) (tenancy.User, error) {
	return getByID(ctx, r.tx, "user", "users", userColumns, r.scan, id.ID)
}

func (r userRepo) ByEmail(ctx context.Context, email string) (tenancy.User, error) {
	if err := r.tx.check(ctx); err != nil {
		return tenancy.User{}, err
	}
	norm, err := tenancy.NormalizeEmail(email)
	if err != nil {
		return tenancy.User{}, notFound("user")
	}
	return queryOne(ctx, &r.tx.txConn, "user", r.scan,
		"SELECT "+userColumns+" FROM iam.users WHERE tenant_id = $1 AND email = $2",
		uuidArg(r.tx.tenant.ID), norm)
}

func (r userRepo) List(ctx context.Context, after ids.UserID, limit int) ([]tenancy.User, error) {
	return listPage(ctx, r.tx, "users", userColumns, r.scan, after.ID, limit)
}

// --- memberships

type memberRepo struct{ tx *tenantTx }

const memberColumns = "tenant_id, organization_id, user_id, created_at"

func (r memberRepo) scan(row pgx.Row) (tenancy.Membership, error) {
	var tenant, org, user pgtype.UUID
	var m tenancy.Membership
	if err := row.Scan(&tenant, &org, &user, &m.CreatedAt); err != nil {
		return tenancy.Membership{}, err
	}
	m.TenantID, m.OrganizationID, m.UserID = ids.TenantID{ID: fromUUID(tenant)}, ids.OrganizationID{ID: fromUUID(org)}, ids.UserID{ID: fromUUID(user)}
	m.CreatedAt = m.CreatedAt.UTC()
	if err := m.Validate(); err != nil {
		return tenancy.Membership{}, corrupt("membership", err)
	}
	return m, r.tx.own("membership", m.TenantID)
}

// Add relies on the composite foreign keys (tenant_id, organization_id) and
// (tenant_id, user_id): an organization or a user of another tenant is
// missing in this one, ErrNotFound.
func (r memberRepo) Add(ctx context.Context, m tenancy.Membership) error {
	if err := r.tx.check(ctx); err != nil {
		return err
	}
	m.TenantID = r.tx.tenant
	if err := m.Validate(); err != nil {
		return storage.Invalid(err)
	}
	_, err := r.tx.write(ctx, "membership",
		"INSERT INTO iam.memberships ("+memberColumns+") VALUES ($1, $2, $3, $4)",
		uuidArg(m.TenantID.ID), uuidArg(m.OrganizationID.ID), uuidArg(m.UserID.ID), m.CreatedAt)
	return err
}

func (r memberRepo) Remove(ctx context.Context, org ids.OrganizationID, user ids.UserID) error {
	tag, err := r.tx.write(ctx, "membership",
		"DELETE FROM iam.memberships WHERE tenant_id = $1 AND organization_id = $2 AND user_id = $3",
		uuidArg(r.tx.tenant.ID), uuidArg(org.ID), uuidArg(user.ID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return notFound("membership")
	}
	return nil
}

func (r memberRepo) ListByUser(ctx context.Context, user ids.UserID) ([]tenancy.Membership, error) {
	return queryAll(ctx, &r.tx.txConn, r.scan,
		"SELECT "+memberColumns+" FROM iam.memberships WHERE tenant_id = $1 AND user_id = $2 ORDER BY organization_id",
		uuidArg(r.tx.tenant.ID), uuidArg(user.ID))
}

// --- applications

type appRepo struct{ tx *tenantTx }

const appColumns = "id, tenant_id, name, type, redirect_uris, secret_hash, created_at"

// scan validates with the Store's allowInsecureHTTP: a store that refuses
// insecure redirect URIs does not serve a client stored with one.
func (r appRepo) scan(row pgx.Row) (tenancy.Application, error) {
	var id, tenant pgtype.UUID
	var a tenancy.Application
	if err := row.Scan(&id, &tenant, &a.Name, &a.Type, &a.RedirectURIs, &a.SecretHash, &a.CreatedAt); err != nil {
		return tenancy.Application{}, err
	}
	a.ID, a.TenantID, a.CreatedAt = ids.ApplicationID{ID: fromUUID(id)}, ids.TenantID{ID: fromUUID(tenant)}, a.CreatedAt.UTC()
	if err := a.Validate(r.tx.allowInsecureHTTP); err != nil {
		return tenancy.Application{}, corrupt("application", err)
	}
	return a, r.tx.own("application", a.TenantID)
}

func (r appRepo) Create(ctx context.Context, a tenancy.Application) error {
	if err := r.tx.check(ctx); err != nil {
		return err
	}
	a.TenantID = r.tx.tenant
	if err := a.Validate(r.tx.allowInsecureHTTP); err != nil {
		return storage.Invalid(err)
	}
	_, err := r.tx.write(ctx, "application",
		"INSERT INTO iam.applications ("+appColumns+") VALUES ($1, $2, $3, $4, $5, $6, $7)",
		uuidArg(a.ID.ID), uuidArg(a.TenantID.ID), a.Name, string(a.Type), a.RedirectURIs, a.SecretHash, a.CreatedAt)
	return err
}

func (r appRepo) Get(ctx context.Context, id ids.ApplicationID) (tenancy.Application, error) {
	return getByID(ctx, r.tx, "application", "applications", appColumns, r.scan, id.ID)
}

func (r appRepo) List(ctx context.Context, after ids.ApplicationID, limit int) ([]tenancy.Application, error) {
	return listPage(ctx, r.tx, "applications", appColumns, r.scan, after.ID, limit)
}
