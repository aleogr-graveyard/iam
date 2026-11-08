// Package tenancy is the domain model shared by AuthN and AuthZ (design,
// section 5): tenant, organization, user, membership and application (OAuth
// client), with the validation rules every boundary applies to them.
//
// Each entity is a plain value with a Validate method rather than a
// constructor: the same check then runs on input from the API and on rows read
// from storage, and a zero value never passes as valid. Validate never mutates
// its receiver; callers normalize first (NormalizeEmail) and then validate.
//
// Errors wrap ErrInvalid (and a more specific sentinel where one exists) and
// name the offending field, never its content: e-mail addresses are personal
// data and redirect URIs are attacker-controlled text.
package tenancy

import (
	"time"

	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/ids"
)

// SystemTenantSlug is the slug of the system tenant, which hosts the master
// and the operators and is invisible to other tenants (brief 4.23). It is
// reserved: ValidateSlug and Tenant.Validate refuse it, so no tenant built from
// input can take it. Only NewSystemTenant builds the system tenant.
const SystemTenantSlug = "system"

// Tenant is a fully isolated namespace; its issuer is https://<host>/t/<slug>
// (ADR 0007).
type Tenant struct {
	ID        ids.TenantID
	Slug      string
	Name      string
	Status    Status
	CreatedAt time.Time
}

// NewSystemTenant builds the system tenant, active, created at clk.Now(). It is
// for the bootstrap path only, the single writer of the system tenant; every
// other tenant comes from input and goes through Validate.
func NewSystemTenant(id ids.TenantID, name string, clk clock.Clock) (Tenant, error) {
	if clk == nil {
		return Tenant{}, invalid("tenant", "created_at", errNoClock)
	}
	t := Tenant{ID: id, Slug: SystemTenantSlug, Name: name, Status: StatusActive, CreatedAt: clk.Now()}
	if err := t.validateFields(); err != nil {
		return Tenant{}, err
	}
	return t, nil
}

// IsSystem reports whether t is the system tenant.
func (t Tenant) IsSystem() bool { return t.Slug == SystemTenantSlug }

// Validate checks every field of a tenant built from input. It fails closed:
// every reserved slug is refused, the system slug included.
func (t Tenant) Validate() error {
	if err := ValidateSlug(t.Slug); err != nil {
		return invalid("tenant", "slug", err)
	}
	return t.validateFields()
}

// ValidateStored is ONLY for rows read back from storage: it is Validate,
// except that it also accepts the system tenant, which NewSystemTenant created.
// Never call it on input.
func (t Tenant) ValidateStored() error {
	if t.IsSystem() {
		return t.validateFields()
	}
	return t.Validate()
}

// validateFields checks everything but the slug rule.
func (t Tenant) validateFields() error {
	if t.ID.IsZero() {
		return invalid("tenant", "id", errZero)
	}
	if err := ValidateName(t.Name); err != nil {
		return invalid("tenant", "name", err)
	}
	if !t.Status.Valid() {
		return invalid("tenant", "status", errUnknown)
	}
	if t.CreatedAt.IsZero() {
		return invalid("tenant", "created_at", errZero)
	}
	return nil
}
