package tenancy

import (
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
	maxEmailLen = 254 // RFC 5321 section 4.5.3.1.3, forward-path minus the brackets
	maxLocalLen = 64  // RFC 5321 section 4.5.3.1.1
)

// NormalizeEmail returns the canonical form of an e-mail address, used for
// storage and for every comparison: surrounding whitespace trimmed, Unicode
// NFC, domain lowercased. The local part keeps its case: RFC 5321 lets the
// receiving host treat it as case-sensitive.
//
// It refuses invalid UTF-8, any whitespace, control or bidirectional character
// inside the address, anything but exactly one '@', an empty local part, a
// domain with an empty label, a local part over 64 octets and an address over
// 254 octets after normalization.
//
// Internationalized domains are lowercased and NFC-normalized but not converted
// to punycode: the address is an identifier here, not something this package
// delivers mail to, and NFC plus lowercase already gives one spelling per
// domain for the comparisons the service makes. IDNA mapping can be added
// later behind the same function without changing stored ASCII addresses.
//
// This is syntax normalization only; it says nothing about deliverability or
// ownership, which only a verification message proves.
func NormalizeEmail(s string) (string, error) {
	s = strings.TrimSpace(s)
	if !utf8.ValidString(s) {
		return "", fmt.Errorf("%w: invalid UTF-8", ErrInvalidEmail)
	}
	if strings.IndexFunc(s, isUnsafeInEmail) >= 0 {
		return "", fmt.Errorf("%w: whitespace or control character", ErrInvalidEmail)
	}
	local, domain, ok := strings.Cut(s, "@")
	if !ok || strings.Contains(domain, "@") {
		return "", fmt.Errorf("%w: must contain exactly one '@'", ErrInvalidEmail)
	}
	local = norm.NFC.String(local)
	domain = norm.NFC.String(strings.ToLower(domain))
	if local == "" || len(local) > maxLocalLen {
		return "", fmt.Errorf("%w: local part must be 1 to %d octets", ErrInvalidEmail, maxLocalLen)
	}
	if err := checkDomainLabels(domain); err != nil {
		return "", err
	}
	out := local + "@" + domain
	if len(out) > maxEmailLen {
		return "", fmt.Errorf("%w: longer than %d octets", ErrInvalidEmail, maxEmailLen)
	}
	return out, nil
}

func checkDomainLabels(domain string) error {
	for label := range strings.SplitSeq(domain, ".") {
		if label == "" {
			return fmt.Errorf("%w: empty domain or domain label", ErrInvalidEmail)
		}
	}
	return nil
}

func isUnsafeInEmail(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsControl(r) || isBidiControl(r)
}
