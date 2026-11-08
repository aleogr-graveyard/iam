package tenancy

import (
	"time"

	"github.com/aleogr/iam/shared/ids"
)

// Membership links a user to an organization of the same tenant. Roles inside
// the organization are relations in AuthZ, not fields here.
type Membership struct {
	TenantID       ids.TenantID
	OrganizationID ids.OrganizationID
	UserID         ids.UserID
	CreatedAt      time.Time
}

// Validate checks every field. That the organization and the user belong to
// TenantID is a storage invariant (tenant-scoped foreign keys), not checked here.
func (m Membership) Validate() error {
	if m.TenantID.IsZero() {
		return invalid("membership", "tenant_id", errZero)
	}
	if m.OrganizationID.IsZero() {
		return invalid("membership", "organization_id", errZero)
	}
	if m.UserID.IsZero() {
		return invalid("membership", "user_id", errZero)
	}
	if m.CreatedAt.IsZero() {
		return invalid("membership", "created_at", errZero)
	}
	return nil
}
