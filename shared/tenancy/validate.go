package tenancy

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ErrInvalid reports an entity or value that breaks a domain rule. Every error
// of this package wraps it.
var ErrInvalid = errors.New("tenancy: invalid")

// ErrInvalidSlug reports a slug refused by ValidateSlug. It wraps ErrInvalid.
var ErrInvalidSlug = fmt.Errorf("%w slug", ErrInvalid)

var (
	errZero          = errors.New("zero value")
	errUnknown       = errors.New("unknown value")
	errNotNormalized = errors.New("not in normalized form")
)

// invalid returns the error for a field of an entity. err must never carry the
// field's content.
func invalid(entity, field string, err error) error {
	if errors.Is(err, ErrInvalid) {
		return fmt.Errorf("%s: %s: %w", entity, field, err)
	}
	return fmt.Errorf("%w: %s: %s: %w", ErrInvalid, entity, field, err)
}

// Status is the lifecycle state of a tenant or a user.
type Status string

// Statuses.
const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

// Valid reports whether s is a declared status.
func (s Status) Valid() bool { return s == StatusActive || s == StatusSuspended }

// ClientType is the OAuth client type (RFC 6749 section 2.1).
type ClientType string

// Client types.
const (
	ClientPublic       ClientType = "public"
	ClientConfidential ClientType = "confidential"
)

// Valid reports whether c is a declared client type.
func (c ClientType) Valid() bool { return c == ClientPublic || c == ClientConfidential }

const (
	minSlugLen = 2
	maxSlugLen = 63
)

// reservedSlugs are refused as tenant and organization slugs: the system
// tenant, and names that a route or a future subdomain issuer (ADR 0007) may
// need. The issuer prefix "t" (/t/<tenant>) needs no entry: one-character
// slugs are already refused by the length floor.
var reservedSlugs = []string{SystemTenantSlug, "admin", "api", "well-known"}

// ValidateSlug accepts ^[a-z0-9][a-z0-9-]{1,62}$ without a trailing dash (so a
// slug is also a valid DNS label for a future subdomain issuer), and refuses
// the reserved slugs.
func ValidateSlug(s string) error {
	if len(s) < minSlugLen || len(s) > maxSlugLen {
		return fmt.Errorf("%w: length must be %d to %d", ErrInvalidSlug, minSlugLen, maxSlugLen)
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !isLowerAlnum(c) && (c != '-' || i == 0 || i == len(s)-1) {
			return fmt.Errorf("%w: only a-z, 0-9 and inner dashes", ErrInvalidSlug)
		}
	}
	if slices.Contains(reservedSlugs, s) {
		return fmt.Errorf("%w: reserved", ErrInvalidSlug)
	}
	return nil
}

func isLowerAlnum(c byte) bool { return 'a' <= c && c <= 'z' || '0' <= c && c <= '9' }

// MaxNameLen is the longest display name, in characters (runes).
const MaxNameLen = 128

// ValidateName accepts a display name of 1 to MaxNameLen characters of valid
// UTF-8, without leading or trailing whitespace, control characters or
// bidirectional formatting characters (which can disguise a name, such as
// "Acme" + U+202E + "fdp.exe"). Names are refused, never silently rewritten.
func ValidateName(s string) error {
	if s == "" || strings.TrimSpace(s) != s {
		return fmt.Errorf("%w name: empty or surrounded by whitespace", ErrInvalid)
	}
	if !utf8.ValidString(s) || utf8.RuneCountInString(s) > MaxNameLen {
		return fmt.Errorf("%w name: invalid UTF-8 or longer than %d characters", ErrInvalid, MaxNameLen)
	}
	if strings.IndexFunc(s, isUnsafeInName) >= 0 {
		return fmt.Errorf("%w name: control or bidirectional character", ErrInvalid)
	}
	return nil
}

func isUnsafeInName(r rune) bool { return unicode.IsControl(r) || isBidiControl(r) }

// isBidiControl reports the explicit bidirectional formatting characters
// (Unicode UAX #9): embeddings, overrides and isolates. Marks such as U+200E
// are left alone: they do not reorder text on their own.
func isBidiControl(r rune) bool {
	return r >= 0x202a && r <= 0x202e || r >= 0x2066 && r <= 0x2069
}
