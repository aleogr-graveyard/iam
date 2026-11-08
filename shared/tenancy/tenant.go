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

	"github.com/aleogr/iam/shared/ids"
)

// SystemTenantSlug is the slug of the system tenant, which hosts the master
// and the operators and is invisible to other tenants (brief 4.23). It is
// reserved: ValidateSlug refuses it, so no other tenant can take it.
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

// IsSystem reports whether t is the system tenant.
func (t Tenant) IsSystem() bool { return t.Slug == SystemTenantSlug }

// Validate checks every field. The reserved SystemTenantSlug is accepted here
// so that the system tenant itself validates; creating a tenant from user
// input must also pass the slug through ValidateSlug.
func (t Tenant) Validate() error {
	if t.ID.IsZero() {
		return invalid("tenant", "id", errZero)
	}
	if !t.IsSystem() {
		if err := ValidateSlug(t.Slug); err != nil {
			return invalid("tenant", "slug", err)
		}
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
