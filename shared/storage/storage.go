// Package storage defines the storage ports of the service (contract C3): the
// interfaces every adapter implements (adapters/memory, adapters/postgres) and
// the domain cores depend on. The contract suite in storagetest is the
// executable form of the rules below; every adapter must pass it.
//
// # Tenant scope
//
// Every tenant-owned read and write happens inside Store.InTenant, which opens
// one transaction bound to one tenant (Postgres: one database transaction with
// the RLS variable iam.tenant_id set transaction-locally). Repository methods
// never take a tenant: the tenant is implicit in the Tx, no method ever returns
// another tenant's row, and a row created through a Tx always gets the Tx's
// tenant, whatever its TenantID field says (the field is overwritten before
// validation). The tenant registry is reachable only through Store.InSystem.
//
// IDs are unique across tenants, per entity kind: creating an organization,
// user or application whose ID exists in another tenant is ErrConflict, and the
// other tenant's row is untouched. An ID names one row of the whole service:
// the application ID is the OAuth client_id, and IDs appear in audit events
// and pagination cursors, where a tenant-local ID would be ambiguous. (Postgres
// target: PRIMARY KEY (id) plus UNIQUE (tenant_id, id), the latter the target
// of the tenant-scoped composite foreign keys such as memberships'.)
//
// A tenant named in a request ({tenant} in the admin API) is resolved only
// through SystemTx.TenantBySlug, never from a tenant ID sent by a client.
//
// # Transactions
//
//   - fn runs inside the transaction; the transaction commits when fn returns
//     nil and rolls back when fn returns an error (returned unchanged) or
//     panics (the panic propagates after the rollback).
//   - fn must use the ctx it receives. Nesting is not supported: InTenant or
//     InSystem called with that ctx, or with any ctx derived from it, returns
//     ErrNestedTx, and the outer transaction is unaffected. The marker is a
//     ctx value, which gives two limits. It survives context.WithoutCancel and
//     goroutines started from fn's ctx, so work deliberately detached from the
//     transaction (an audit flush after it, say) gets a false ErrNestedTx: it
//     must start from a ctx that does not descend from fn's. And nesting
//     through an unrelated ctx is not detected at all (the memory adapter
//     deadlocks; Postgres takes a second pool connection); the forbidigo rule
//     on context.Background and context.TODO outside entry points makes such a
//     ctx hard to come by.
//   - A Tx, a SystemTx and their repositories are valid only while fn runs and
//     only on fn's goroutine; afterwards every method returns ErrTxDone.
//
// # Arguments and errors
//
//   - Entities are validated on create, after the Tx's tenant has been forced
//     onto them; a validation failure is an ErrInvalid error that also wraps
//     tenancy.ErrInvalid, never ErrConflict or ErrNotFound.
//   - A zero ID given to a lookup (Get, Tenant, Remove) finds nothing:
//     ErrNotFound. A zero ID inside an entity to create is ErrInvalid. A zero
//     tenant given to InTenant is ErrInvalidTenant.
//   - Lookups by slug or e-mail take raw input: BySlug and TenantBySlug with a
//     slug refused by tenancy.ValidateSlug, and ByEmail with an address refused
//     by tenancy.NormalizeEmail, give ErrNotFound. ByEmail normalizes its input.
//   - Error messages name entities and fields, never values: no e-mail, slug,
//     redirect URI or secret ever appears in them.
//
// # Lists
//
// List methods return rows in ascending ID order (UUIDv7 IDs sort by creation
// time, to the millisecond), starting strictly after `after`; the zero ID
// starts from the beginning. `after` is only a position: an ID from another
// tenant, or one that does not exist, returns this tenant's rows past that
// position and nothing else. A limit of zero or less is ErrInvalid; a limit
// above MaxListLimit is clamped to it (see ClampLimit).
//
// # Carried over to later tasks
//
//   - Postgres adapter (Task 9): the RLS tests are mandatory, not optional:
//     a query without the tenant filter inside InTenant sees only that tenant,
//     iam.tenant_id never leaks across pooled transactions, and InSystem
//     resets iam.tenant_id (to the system tenant or to unset, ruling R1) on a
//     connection that just served InTenant.
//   - MembershipRepo.ListByUser is unbounded here. The tenancy service
//     (Task 11) enforces a configurable cap on memberships per user, refusing
//     an Add beyond it (fail closed), so the list stays small.
package storage

import (
	"context"

	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/tenancy"
)

// MaxListLimit is the largest page a List method returns.
const MaxListLimit = 200

// Store opens tenant-scoped and system transactions. Implementations are safe
// for concurrent use.
type Store interface {
	// InTenant runs fn in one transaction bound to tenant. A zero or unknown
	// tenant is ErrInvalidTenant and fn does not run. The tenant's status is
	// not checked here: lifecycle rules (a suspended tenant) belong to callers.
	InTenant(ctx context.Context, tenant ids.TenantID, fn func(ctx context.Context, tx Tx) error) error
	// InSystem runs fn in one transaction over the tenant registry. Its audit
	// events go to the system tenant's chain once it exists (Task 10).
	InSystem(ctx context.Context, fn func(ctx context.Context, tx SystemTx) error) error
}

// SystemTx is the tenant registry. Only the platform's own operations use it;
// tenants never reach it.
type SystemTx interface {
	// CreateTenant validates t with tenancy.Tenant.Validate, which refuses the
	// system slug, and stores it. A duplicate ID or slug is ErrConflict.
	CreateTenant(ctx context.Context, t tenancy.Tenant) error
	// EnsureSystemTenant returns the system tenant, creating it through
	// tenancy.NewSystemTenant(id, name, clk) when it does not exist yet. It is
	// idempotent: once the system tenant exists, id, name and clk are ignored
	// and every call returns the same tenant, also under concurrent calls, so
	// callers must use the returned tenant's ID, never the id they passed. An
	// id that already names an ordinary tenant is ErrConflict: Ensure never
	// adopts an existing tenant as the system tenant.
	EnsureSystemTenant(ctx context.Context, id ids.TenantID, name string, clk clock.Clock) (tenancy.Tenant, error)
	// SystemTenant returns the system tenant, or ErrNotFound before
	// EnsureSystemTenant has created it.
	SystemTenant(ctx context.Context) (tenancy.Tenant, error)
	// TenantBySlug finds a tenant by its slug. A slug refused by
	// tenancy.ValidateSlug, the system slug included, is ErrNotFound: the
	// system tenant is not addressable by slug (use SystemTenant).
	TenantBySlug(ctx context.Context, slug string) (tenancy.Tenant, error)
	// Tenant finds a tenant by ID, the system tenant included.
	Tenant(ctx context.Context, id ids.TenantID) (tenancy.Tenant, error)
	// ListTenants lists every tenant, the system tenant included.
	ListTenants(ctx context.Context, after ids.TenantID, limit int) ([]tenancy.Tenant, error)
}

// Tx is a transaction bound to one tenant.
type Tx interface {
	Organizations() OrganizationRepo
	Users() UserRepo
	Memberships() MembershipRepo
	Applications() ApplicationRepo
}

// OrganizationRepo stores the organizations of the Tx's tenant. Slugs are
// unique within the tenant.
type OrganizationRepo interface {
	Create(ctx context.Context, o tenancy.Organization) error
	Get(ctx context.Context, id ids.OrganizationID) (tenancy.Organization, error)
	BySlug(ctx context.Context, slug string) (tenancy.Organization, error)
	List(ctx context.Context, after ids.OrganizationID, limit int) ([]tenancy.Organization, error)
}

// UserRepo stores the users of the Tx's tenant. E-mail addresses, in the form
// tenancy.NormalizeEmail returns, are unique within the tenant.
type UserRepo interface {
	// Create requires u.Email to be normalized already (tenancy.User.Validate).
	Create(ctx context.Context, u tenancy.User) error
	Get(ctx context.Context, id ids.UserID) (tenancy.User, error)
	// ByEmail normalizes email with tenancy.NormalizeEmail before the lookup.
	ByEmail(ctx context.Context, email string) (tenancy.User, error)
	List(ctx context.Context, after ids.UserID, limit int) ([]tenancy.User, error)
}

// MembershipRepo links users to organizations of the Tx's tenant.
type MembershipRepo interface {
	// Add requires the organization and the user to exist in the Tx's tenant
	// (ErrNotFound otherwise); a duplicate pair is ErrConflict.
	Add(ctx context.Context, m tenancy.Membership) error
	// Remove deletes the pair, or returns ErrNotFound.
	Remove(ctx context.Context, org ids.OrganizationID, user ids.UserID) error
	// ListByUser returns the user's memberships in ascending organization ID
	// order; an unknown user has none.
	ListByUser(ctx context.Context, user ids.UserID) ([]tenancy.Membership, error)
}

// ApplicationRepo stores the OAuth clients of the Tx's tenant. Whether
// Application.Validate accepts insecure http redirect URIs is fixed when the
// adapter is constructed, from configuration (development only), and is off
// by default.
type ApplicationRepo interface {
	Create(ctx context.Context, a tenancy.Application) error
	Get(ctx context.Context, id ids.ApplicationID) (tenancy.Application, error)
	List(ctx context.Context, after ids.ApplicationID, limit int) ([]tenancy.Application, error)
}
