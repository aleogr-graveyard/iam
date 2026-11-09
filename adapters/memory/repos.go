package memory

import (
	"context"
	"slices"

	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/storage"
	"github.com/aleogr/iam/shared/tenancy"
)

// tenantTx is a transaction bound to one tenant: its repositories see only
// data, the working copy of that tenant's rows.
type tenantTx struct {
	txState
	tenant            ids.TenantID
	data              *tenantData
	all               map[ids.ID]*tenantData // committed rows of every tenant; read only
	allowInsecureHTTP bool
}

// takenElsewhere reports whether another tenant holds a row with this ID, as
// found by has: IDs are unique across tenants (storage package doc).
func (tx *tenantTx) takenElsewhere(has func(*tenantData) bool) bool {
	for id, d := range tx.all {
		if id != tx.tenant.ID && has(d) {
			return true
		}
	}
	return false
}

func (tx *tenantTx) Organizations() storage.OrganizationRepo { return orgRepo{tx} }
func (tx *tenantTx) Users() storage.UserRepo                 { return userRepo{tx} }
func (tx *tenantTx) Memberships() storage.MembershipRepo     { return memberRepo{tx} }
func (tx *tenantTx) Applications() storage.ApplicationRepo   { return appRepo{tx} }

// createKeyed validates v (its tenant and time already forced by the caller)
// and stores it in the index pick selects, in the Tx's tenant, refusing an ID
// that any tenant already holds or a duplicate key.
func createKeyed[V any](ctx context.Context, tx *tenantTx, validate func() error, pick func(*tenantData) keyed[V], id ids.ID, key string, v V) error {
	if err := tx.check(ctx); err != nil {
		return err
	}
	if err := validate(); err != nil {
		return storage.Invalid(err)
	}
	idx := pick(tx.data)
	if tx.takenElsewhere(func(d *tenantData) bool { _, ok := pick(d).byID[id]; return ok }) {
		return conflict(idx.entity, "id")
	}
	return idx.insert(id, key, v)
}

// --- organizations

type orgRepo struct{ tx *tenantTx }

func (r orgRepo) Create(ctx context.Context, o tenancy.Organization) error {
	o.TenantID, o.CreatedAt = r.tx.tenant, storage.Microseconds(o.CreatedAt)
	return createKeyed(ctx, r.tx, o.Validate, func(d *tenantData) keyed[tenancy.Organization] { return d.orgs }, o.ID.ID, o.Slug, o)
}

func (r orgRepo) Get(ctx context.Context, id ids.OrganizationID) (tenancy.Organization, error) {
	if err := r.tx.check(ctx); err != nil {
		return tenancy.Organization{}, err
	}
	return get(r.tx.data.orgs.byID, id.ID, "organization", same)
}

// BySlug needs no slug validation: an invalid slug is never stored.
func (r orgRepo) BySlug(ctx context.Context, slug string) (tenancy.Organization, error) {
	if err := r.tx.check(ctx); err != nil {
		return tenancy.Organization{}, err
	}
	return r.tx.data.orgs.lookup(slug)
}

func (r orgRepo) List(ctx context.Context, after ids.OrganizationID, limit int) ([]tenancy.Organization, error) {
	if err := r.tx.check(ctx); err != nil {
		return nil, err
	}
	return page(r.tx.data.orgs.byID, after.ID, limit, same)
}

// --- users

type userRepo struct{ tx *tenantTx }

func (r userRepo) Create(ctx context.Context, u tenancy.User) error {
	u.TenantID, u.CreatedAt = r.tx.tenant, storage.Microseconds(u.CreatedAt)
	return createKeyed(ctx, r.tx, u.Validate, func(d *tenantData) keyed[tenancy.User] { return d.users }, u.ID.ID, u.Email, u)
}

func (r userRepo) Get(ctx context.Context, id ids.UserID) (tenancy.User, error) {
	if err := r.tx.check(ctx); err != nil {
		return tenancy.User{}, err
	}
	return get(r.tx.data.users.byID, id.ID, "user", same)
}

func (r userRepo) ByEmail(ctx context.Context, email string) (tenancy.User, error) {
	if err := r.tx.check(ctx); err != nil {
		return tenancy.User{}, err
	}
	norm, err := tenancy.NormalizeEmail(email)
	if err != nil {
		return tenancy.User{}, notFound("user")
	}
	return r.tx.data.users.lookup(norm)
}

func (r userRepo) List(ctx context.Context, after ids.UserID, limit int) ([]tenancy.User, error) {
	if err := r.tx.check(ctx); err != nil {
		return nil, err
	}
	return page(r.tx.data.users.byID, after.ID, limit, same)
}

// --- memberships

type memberRepo struct{ tx *tenantTx }

func (r memberRepo) Add(ctx context.Context, m tenancy.Membership) error {
	if err := r.tx.check(ctx); err != nil {
		return err
	}
	m.TenantID, m.CreatedAt = r.tx.tenant, storage.Microseconds(m.CreatedAt)
	if err := m.Validate(); err != nil {
		return storage.Invalid(err)
	}
	d := r.tx.data
	if _, ok := d.orgs.byID[m.OrganizationID.ID]; !ok {
		return notFound("organization")
	}
	if _, ok := d.users.byID[m.UserID.ID]; !ok {
		return notFound("user")
	}
	key := memberKey{org: m.OrganizationID.ID, user: m.UserID.ID}
	if _, dup := d.members[key]; dup {
		return conflict("membership", "organization_id, user_id")
	}
	d.members[key] = m
	return nil
}

func (r memberRepo) Remove(ctx context.Context, org ids.OrganizationID, user ids.UserID) error {
	if err := r.tx.check(ctx); err != nil {
		return err
	}
	key := memberKey{org: org.ID, user: user.ID}
	if _, ok := r.tx.data.members[key]; !ok {
		return notFound("membership")
	}
	delete(r.tx.data.members, key)
	return nil
}

func (r memberRepo) ListByUser(ctx context.Context, user ids.UserID) ([]tenancy.Membership, error) {
	if err := r.tx.check(ctx); err != nil {
		return nil, err
	}
	var out []tenancy.Membership
	for k, m := range r.tx.data.members {
		if k.user == user.ID {
			out = append(out, m)
		}
	}
	slices.SortFunc(out, func(a, b tenancy.Membership) int {
		return compareIDs(a.OrganizationID.ID, b.OrganizationID.ID)
	})
	return out, nil
}

// --- applications

type appRepo struct{ tx *tenantTx }

func (r appRepo) Create(ctx context.Context, a tenancy.Application) error {
	if err := r.tx.check(ctx); err != nil {
		return err
	}
	a.TenantID, a.CreatedAt = r.tx.tenant, storage.Microseconds(a.CreatedAt)
	if err := a.Validate(r.tx.allowInsecureHTTP); err != nil {
		return storage.Invalid(err)
	}
	_, dup := r.tx.data.apps[a.ID.ID]
	if dup || r.tx.takenElsewhere(func(d *tenantData) bool { _, ok := d.apps[a.ID.ID]; return ok }) {
		return conflict("application", "id")
	}
	r.tx.data.apps[a.ID.ID] = copyApp(a)
	return nil
}

func (r appRepo) Get(ctx context.Context, id ids.ApplicationID) (tenancy.Application, error) {
	if err := r.tx.check(ctx); err != nil {
		return tenancy.Application{}, err
	}
	return get(r.tx.data.apps, id.ID, "application", copyApp)
}

func (r appRepo) List(ctx context.Context, after ids.ApplicationID, limit int) ([]tenancy.Application, error) {
	if err := r.tx.check(ctx); err != nil {
		return nil, err
	}
	return page(r.tx.data.apps, after.ID, limit, copyApp)
}
