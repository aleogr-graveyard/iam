package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

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

// insert validates a row (its tenant and time already forced by the caller)
// and inserts it.
func insert(ctx context.Context, tx *tenantTx, entity string, validate func() error, sql string, args ...any) error {
	if err := tx.check(ctx); err != nil {
		return err
	}
	if err := validate(); err != nil {
		return storage.Invalid(err)
	}
	_, err := tx.write(ctx, entity, sql, args...)
	return err
}

// scanRow reads one row into dest and checks the value it filled: it must be
// valid and belong to the Tx's tenant (a third check, after the query filter
// and RLS, that costs nothing). Neither failure is ever skipped: ErrCorrupt.
func scanRow[T any](tx *tenantTx, row pgx.Row, entity string, v *T, tenant *ids.TenantID, validate func(T) error, dest ...any) (T, error) {
	var zero T
	if err := row.Scan(dest...); err != nil {
		return zero, err
	}
	if err := validate(*v); err != nil {
		return zero, corrupt(entity, err)
	}
	if *tenant != tx.tenant {
		return zero, corrupt(entity, errForeignRow)
	}
	return *v, nil
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
	var o tenancy.Organization
	return scanRow(r.tx, row, "organization", &o, &o.TenantID, tenancy.Organization.Validate,
		idDest{&o.ID.ID}, idDest{&o.TenantID.ID}, &o.Slug, &o.Name, utcDest{&o.CreatedAt})
}

func (r orgRepo) Create(ctx context.Context, o tenancy.Organization) error {
	o.TenantID, o.CreatedAt = r.tx.tenant, storage.Microseconds(o.CreatedAt)
	return insert(ctx, r.tx, "organization", o.Validate,
		"INSERT INTO iam.organizations ("+orgColumns+") VALUES ($1, $2, $3, $4, $5)",
		uuidArg(o.ID.ID), uuidArg(o.TenantID.ID), o.Slug, o.Name, o.CreatedAt)
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
	var u tenancy.User
	return scanRow(r.tx, row, "user", &u, &u.TenantID, tenancy.User.Validate,
		idDest{&u.ID.ID}, idDest{&u.TenantID.ID}, &u.Email, &u.EmailVerified, &u.Status, utcDest{&u.CreatedAt})
}

func (r userRepo) Create(ctx context.Context, u tenancy.User) error {
	u.TenantID, u.CreatedAt = r.tx.tenant, storage.Microseconds(u.CreatedAt)
	return insert(ctx, r.tx, "user", u.Validate,
		"INSERT INTO iam.users ("+userColumns+") VALUES ($1, $2, $3, $4, $5, $6)",
		uuidArg(u.ID.ID), uuidArg(u.TenantID.ID), u.Email, u.EmailVerified, string(u.Status), u.CreatedAt)
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
	var m tenancy.Membership
	return scanRow(r.tx, row, "membership", &m, &m.TenantID, tenancy.Membership.Validate,
		idDest{&m.TenantID.ID}, idDest{&m.OrganizationID.ID}, idDest{&m.UserID.ID}, utcDest{&m.CreatedAt})
}

// Add relies on the composite foreign keys (tenant_id, organization_id) and
// (tenant_id, user_id): an organization or a user of another tenant is
// missing in this one, ErrNotFound.
func (r memberRepo) Add(ctx context.Context, m tenancy.Membership) error {
	m.TenantID, m.CreatedAt = r.tx.tenant, storage.Microseconds(m.CreatedAt)
	return insert(ctx, r.tx, "membership", m.Validate,
		"INSERT INTO iam.memberships ("+memberColumns+") VALUES ($1, $2, $3, $4)",
		uuidArg(m.TenantID.ID), uuidArg(m.OrganizationID.ID), uuidArg(m.UserID.ID), m.CreatedAt)
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

// validate uses the Store's allowInsecureHTTP, on create and on read: a store
// that refuses insecure redirect URIs does not serve a client stored with one.
func (r appRepo) validate(a tenancy.Application) error { return a.Validate(r.tx.allowInsecureHTTP) }

func (r appRepo) scan(row pgx.Row) (tenancy.Application, error) {
	var a tenancy.Application
	return scanRow(r.tx, row, "application", &a, &a.TenantID, r.validate,
		idDest{&a.ID.ID}, idDest{&a.TenantID.ID}, &a.Name, &a.Type, &a.RedirectURIs, &a.SecretHash, utcDest{&a.CreatedAt})
}

func (r appRepo) Create(ctx context.Context, a tenancy.Application) error {
	a.TenantID, a.CreatedAt = r.tx.tenant, storage.Microseconds(a.CreatedAt)
	return insert(ctx, r.tx, "application", func() error { return r.validate(a) },
		"INSERT INTO iam.applications ("+appColumns+") VALUES ($1, $2, $3, $4, $5, $6, $7)",
		uuidArg(a.ID.ID), uuidArg(a.TenantID.ID), a.Name, string(a.Type), a.RedirectURIs, a.SecretHash, a.CreatedAt)
}

func (r appRepo) Get(ctx context.Context, id ids.ApplicationID) (tenancy.Application, error) {
	return getByID(ctx, r.tx, "application", "applications", appColumns, r.scan, id.ID)
}

func (r appRepo) List(ctx context.Context, after ids.ApplicationID, limit int) ([]tenancy.Application, error) {
	return listPage(ctx, r.tx, "applications", appColumns, r.scan, after.ID, limit)
}
