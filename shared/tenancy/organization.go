package tenancy

import (
	"time"

	"github.com/aleogr/iam/shared/ids"
)

// Organization lives inside a tenant and has its own policies, administrators
// and branding. Its slug is unique within the tenant.
type Organization struct {
	ID        ids.OrganizationID
	TenantID  ids.TenantID
	Slug      string
	Name      string
	CreatedAt time.Time
}

// Validate checks every field.
func (o Organization) Validate() error {
	if o.ID.IsZero() {
		return invalid("organization", "id", errZero)
	}
	if o.TenantID.IsZero() {
		return invalid("organization", "tenant_id", errZero)
	}
	if err := ValidateSlug(o.Slug); err != nil {
		return invalid("organization", "slug", err)
	}
	if err := ValidateName(o.Name); err != nil {
		return invalid("organization", "name", err)
	}
	if o.CreatedAt.IsZero() {
		return invalid("organization", "created_at", errZero)
	}
	return nil
}
