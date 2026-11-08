package tenancy

import (
	"time"

	"github.com/aleogr/iam/shared/ids"
)

// User belongs to one tenant: one person, one account, any number of
// organizations through memberships.
type User struct {
	ID            ids.UserID
	TenantID      ids.TenantID
	Email         string // in the form NormalizeEmail returns
	EmailVerified bool
	Status        Status
	CreatedAt     time.Time
}

// Validate checks every field. Email must already be normalized: an address
// that NormalizeEmail would change is refused, so two spellings of one address
// can never be stored as different users.
func (u User) Validate() error {
	if u.ID.IsZero() {
		return invalid("user", "id", errZero)
	}
	if u.TenantID.IsZero() {
		return invalid("user", "tenant_id", errZero)
	}
	norm, err := NormalizeEmail(u.Email)
	if err != nil {
		return invalid("user", "email", err)
	}
	if norm != u.Email {
		return invalid("user", "email", errNotNormalized)
	}
	if !u.Status.Valid() {
		return invalid("user", "status", errUnknown)
	}
	if u.CreatedAt.IsZero() {
		return invalid("user", "created_at", errZero)
	}
	return nil
}
