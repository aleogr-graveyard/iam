package tenancy

import (
	"errors"
	"fmt"
	"time"

	"github.com/aleogr/iam/shared/crypto"
	"github.com/aleogr/iam/shared/ids"
)

// MaxRedirectURIs bounds the redirect URIs of one application. Exact matching
// makes each one a separate entry; 20 covers production, staging and local
// development callbacks while keeping every authorization request's lookup and
// every stored row small.
const MaxRedirectURIs = 20

// Application is an OAuth client of a tenant. PKCE is always required, for
// both types (ADR 0012).
type Application struct {
	ID           ids.ApplicationID
	TenantID     ids.TenantID
	Name         string
	Type         ClientType
	RedirectURIs []string
	// SecretHash is crypto.HashSecret of the client secret: empty for a public
	// client, exactly crypto.SecretHashBytes for a confidential one.
	SecretHash []byte
	CreatedAt  time.Time
}

var (
	errSecretOnPublic   = errors.New("public client has a secret")
	errSecretLength     = errors.New("secret hash has the wrong length")
	errNoRedirect       = errors.New("no redirect URI")
	errTooManyRedirects = fmt.Errorf("more than %d redirect URIs", MaxRedirectURIs)
	errDuplicate        = errors.New("duplicate")
)

// Validate checks every field. allowInsecureHTTP is passed through to
// ValidateRedirectURI and comes from configuration (development only).
func (a Application) Validate(allowInsecureHTTP bool) error {
	if a.ID.IsZero() {
		return invalid("application", "id", errZero)
	}
	if a.TenantID.IsZero() {
		return invalid("application", "tenant_id", errZero)
	}
	if err := ValidateName(a.Name); err != nil {
		return invalid("application", "name", err)
	}
	if err := a.validateSecret(); err != nil {
		return invalid("application", "secret_hash", err)
	}
	if err := validateRedirectURIs(a.RedirectURIs, allowInsecureHTTP); err != nil {
		return invalid("application", "redirect_uris", err)
	}
	if a.CreatedAt.IsZero() {
		return invalid("application", "created_at", errZero)
	}
	return nil
}

func (a Application) validateSecret() error {
	switch a.Type {
	case ClientPublic:
		// A non-nil empty slice is refused too: "no secret" has one spelling.
		if a.SecretHash != nil {
			return errSecretOnPublic
		}
	case ClientConfidential:
		if len(a.SecretHash) != crypto.SecretHashBytes {
			return errSecretLength
		}
	default:
		return fmt.Errorf("type: %w", errUnknown)
	}
	return nil
}

// validateRedirectURIs names the index of a bad entry, never its content.
func validateRedirectURIs(uris []string, allowInsecureHTTP bool) error {
	if len(uris) == 0 {
		return errNoRedirect
	}
	if len(uris) > MaxRedirectURIs {
		return errTooManyRedirects
	}
	seen := make(map[string]struct{}, len(uris))
	for i, u := range uris {
		if err := ValidateRedirectURI(u, allowInsecureHTTP); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
		if _, dup := seen[u]; dup {
			return fmt.Errorf("[%d]: %w", i, errDuplicate)
		}
		seen[u] = struct{}{}
	}
	return nil
}
