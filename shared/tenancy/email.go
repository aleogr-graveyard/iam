package tenancy

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// ErrInvalidEmail reports an address refused by NormalizeEmail. It wraps
// ErrInvalid. Its messages never contain the address.
var ErrInvalidEmail = fmt.Errorf("%w e-mail", ErrInvalid)

const (
	maxEmailInput = 1024 // raw input, checked before any trimming or normalization
	maxEmailLen   = 254  // RFC 5321 section 4.5.3.1.3, forward-path minus the brackets
	maxLocalLen   = 64   // RFC 5321 section 4.5.3.1.1
)

// NormalizeEmail returns the canonical form of an e-mail address, used for
// storage and for every comparison, so that every spelling of one mailbox
// yields the same string:
//
//   - input over 1 KiB is refused before any other work;
//   - surrounding whitespace is trimmed; invalid UTF-8, and whitespace or
//     control characters inside the address, are refused;
//   - exactly one '@', separating a non-empty local part and a domain;
//   - the local part is NFC-normalized and keeps its case (RFC 5321 lets the
//     receiving host treat it as case-sensitive); invisible characters in it
//     (format characters, category Cf, and other default-ignorable code
//     points such as U+200B, U+200E, U+FEFF, U+061C, U+00AD) are refused;
//   - the domain is converted to its lowercase A-label form (see
//     canonicalDNSName: UTS #46 mapping, so the uppercase, lowercase, NFD
//     and A-label spellings of one IDN domain are one domain) and must have
//     at least two labels; address literals ("[192.0.2.1]") are refused;
//   - limits (local part 64 octets, address 254 octets) are measured on the
//     stored form.
//
// Confusable characters are accepted: this is syntax normalization only, and
// ownership is proven by a verification message, never by the spelling. The
// domain is displayed in its A-label form, so a look-alike domain does not
// pass for the real one.
func NormalizeEmail(s string) (string, error) {
	if len(s) > maxEmailInput {
		return "", fmt.Errorf("%w: input longer than %d bytes", ErrInvalidEmail, maxEmailInput)
	}
	s = strings.TrimSpace(s)
	if !utf8.ValidString(s) {
		return "", fmt.Errorf("%w: invalid UTF-8", ErrInvalidEmail)
	}
	if strings.IndexFunc(s, isSpaceOrControl) >= 0 {
		return "", fmt.Errorf("%w: whitespace or control character", ErrInvalidEmail)
	}
	local, domain, ok := strings.Cut(s, "@")
	if !ok || strings.Contains(domain, "@") {
		return "", fmt.Errorf("%w: must contain exactly one '@'", ErrInvalidEmail)
	}
	local, err := normalizeLocal(local)
	if err != nil {
		return "", err
	}
	if domain, err = normalizeDomain(domain); err != nil {
		return "", err
	}
	out := local + "@" + domain
	if len(out) > maxEmailLen {
		return "", fmt.Errorf("%w: longer than %d octets", ErrInvalidEmail, maxEmailLen)
	}
	return out, nil
}

func normalizeLocal(local string) (string, error) {
	if strings.IndexFunc(local, isInvisible) >= 0 {
		return "", fmt.Errorf("%w: invisible character in the local part", ErrInvalidEmail)
	}
	local = norm.NFC.String(local)
	if local == "" || len(local) > maxLocalLen {
		return "", fmt.Errorf("%w: local part must be 1 to %d octets", ErrInvalidEmail, maxLocalLen)
	}
	return local, nil
}

var errSingleLabel = errors.New("domain needs at least two labels")

func normalizeDomain(domain string) (string, error) {
	ascii, err := canonicalDNSName(domain)
	if err == nil && !strings.Contains(ascii, ".") {
		err = errSingleLabel
	}
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidEmail, err)
	}
	return ascii, nil
}

func isSpaceOrControl(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }
