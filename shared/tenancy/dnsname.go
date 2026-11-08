package tenancy

import (
	"errors"
	"strings"

	"golang.org/x/net/idna"
)

const (
	maxHostLen  = 253
	maxLabelLen = 63
)

var (
	errDNSName       = errors.New("not a valid domain name")
	errDNSNameLength = errors.New("domain name longer than 253 octets")
	errNumericLabel  = errors.New("last label looks like a numeric IPv4 address")
)

// canonicalDNSName returns the one stored spelling of a domain name: the
// UTS #46 Lookup profile of golang.org/x/net/idna maps it (case folding, NFC,
// width folding, U+3002 and other dots to '.', ignored characters such as the
// soft hyphen removed), validates it (STD3 rules, hyphens, joiners, bidi) and
// converts each label to its lowercase A-label ("xn--..."); an A-label in the
// input must be valid punycode that maps back to itself.
//
// The result must then be lowercase LDH ASCII (RFC 1123): labels of 1 to 63
// octets, no leading or trailing hyphen, at most 253 octets in all, and a last
// label that is neither all digits nor "0x..." (the WHATWG URL parser reads
// such a name as an IPv4 address: 127.1, 2130706433, 0x7f.0.0.1).
//
// The A-label is also the form shown to people: confusable U-labels (a
// Cyrillic U+0430 for a Latin "a") are not refused, but they never display as
// the look-alike.
func canonicalDNSName(name string) (string, error) {
	ascii, err := idna.Lookup.ToASCII(name)
	if err != nil {
		return "", errDNSName
	}
	if len(ascii) > maxHostLen {
		return "", errDNSNameLength
	}
	labels := strings.Split(ascii, ".")
	for _, l := range labels {
		if !isLDHLabel(l) {
			return "", errDNSName
		}
	}
	if last := labels[len(labels)-1]; strings.HasPrefix(last, "0x") || strings.Trim(last, "0123456789") == "" {
		return "", errNumericLabel
	}
	return ascii, nil
}

func isLDHLabel(l string) bool {
	if l == "" || len(l) > maxLabelLen || l[0] == '-' || l[len(l)-1] == '-' {
		return false
	}
	for i := 0; i < len(l); i++ {
		if !isLowerAlnum(l[i]) && l[i] != '-' {
			return false
		}
	}
	return true
}
