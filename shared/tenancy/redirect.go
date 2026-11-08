package tenancy

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

// ErrInvalidRedirectURI reports a redirect URI refused by ValidateRedirectURI.
// It wraps ErrInvalid. Its messages never contain the URI.
var ErrInvalidRedirectURI = fmt.Errorf("%w redirect URI", ErrInvalid)

// MaxRedirectURILen is the longest accepted redirect URI, in bytes. Browsers
// and proxies commonly cap whole URLs near 2 KiB, and a redirect URI is only a
// part of the authorization request.
const MaxRedirectURILen = 2048

const (
	httpsPrefix = "https://"
	httpPrefix  = "http://"
	maxHostLen  = 253
	maxLabelLen = 63
)

var (
	// The loopback IP literals of RFC 8252 section 7.3, in canonical form.
	loopbackHosts = []string{"127.0.0.1", "[::1]"}
	// Percent-encoded '*', '.', '/' and '\' in the path: an encoded wildcard,
	// an encoded dot segment or a separator that some servers decode before
	// routing (threat model TM-OIDC-002).
	encodedTricks = []string{"%2a", "%2e", "%2f", "%5c"}
)

// ValidateRedirectURI enforces ADR 0012, RFC 9700 section 4.1 and RFC 8252 on
// a redirect URI an application registers. Matching at authorization time is
// exact string comparison, so the registered form must be canonical:
//
//   - absolute, at most MaxRedirectURILen bytes of printable ASCII (a non-ASCII
//     host must be registered in its punycode form);
//   - scheme "https", lowercase; or "http" for the loopback IP literals
//     127.0.0.1 and [::1] on any port (RFC 8252 section 7.3); any other http
//     only when allowInsecureHTTP (development configuration);
//   - "localhost" gets no loopback exemption: RFC 8252 section 8.3 recommends
//     against it because the name may resolve elsewhere;
//   - private-use schemes of native apps (com.example.app:/cb, RFC 8252
//     section 7.1) are refused for now: native apps are out of milestone 1
//     (design, section 6) and will be accepted by a later, reviewed change;
//   - no wildcard '*' anywhere, no fragment, no userinfo, no backslash;
//   - a lowercase host that is a DNS name, a canonical IPv4 literal or a
//     canonical bracketed IPv6 literal without zone; no percent-encoding in the
//     host, and no name whose last label looks numeric (127.1, 0x7f.1), which
//     browsers would read as an IPv4 address;
//   - if present, a port from 1 to 65535 without leading zeros;
//   - no dot segments and no encoded '*', '.', '/' or '\' in the path.
func ValidateRedirectURI(raw string, allowInsecureHTTP bool) error {
	if raw == "" || len(raw) > MaxRedirectURILen {
		return redirectErr("empty or longer than %d bytes", MaxRedirectURILen)
	}
	if err := checkURICharacters(raw); err != nil {
		return err
	}
	scheme, rest, ok := splitScheme(raw)
	if !ok {
		return redirectErr("scheme must be https or http, lowercase, with an authority")
	}
	authority, path := splitAuthority(rest)
	if err := checkAuthority(authority); err != nil {
		return err
	}
	if err := checkPath(path); err != nil {
		return err
	}
	if scheme == "http" && !allowInsecureHTTP && !isLoopbackHost(authority) {
		return redirectErr("http is allowed only for 127.0.0.1 and [::1]")
	}
	return checkParsesTheSame(raw, scheme, authority)
}

func redirectErr(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrInvalidRedirectURI}, args...)...)
}

// checkURICharacters refuses anything but printable ASCII, and the wildcard,
// fragment and backslash characters anywhere.
func checkURICharacters(raw string) error {
	for i := 0; i < len(raw); i++ {
		if raw[i] <= ' ' || raw[i] >= 0x7f {
			return redirectErr("non-printable or non-ASCII byte at offset %d", i)
		}
	}
	if strings.ContainsAny(raw, "*#\\") {
		return redirectErr("wildcard, fragment or backslash")
	}
	return nil
}

func splitScheme(raw string) (scheme, rest string, ok bool) {
	if rest, ok = strings.CutPrefix(raw, httpsPrefix); ok {
		return "https", rest, true
	}
	if rest, ok = strings.CutPrefix(raw, httpPrefix); ok {
		return "http", rest, true
	}
	return "", "", false
}

// splitAuthority splits what follows "scheme://" into the authority and the
// path (the query, if any, is not inspected beyond the character checks).
func splitAuthority(rest string) (authority, path string) {
	end := strings.IndexAny(rest, "/?")
	if end < 0 {
		return rest, ""
	}
	authority, rest = rest[:end], rest[end:]
	path, _, _ = strings.Cut(rest, "?")
	return authority, path
}

func checkAuthority(authority string) error {
	if strings.ContainsAny(authority, "@%") {
		return redirectErr("userinfo or percent-encoding in the authority")
	}
	host, port, err := splitHostPort(authority)
	if err != nil {
		return err
	}
	if port != "" {
		if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 || strconv.Itoa(n) != port {
			return redirectErr("port must be 1 to 65535 without leading zeros")
		}
	}
	return checkHost(host)
}

var errBadAuthority = redirectErr("malformed host or port")

// splitHostPort separates a bracketed IPv6 literal or a name from an optional
// ":port"; a ':' with no port after it is refused.
func splitHostPort(authority string) (host, port string, err error) {
	if strings.HasPrefix(authority, "[") {
		end := strings.IndexByte(authority, ']')
		if end < 0 {
			return "", "", errBadAuthority
		}
		host, rest := authority[:end+1], authority[end+1:]
		if rest == "" {
			return host, "", nil
		}
		port, ok := strings.CutPrefix(rest, ":")
		if !ok || port == "" {
			return "", "", errBadAuthority
		}
		return host, port, nil
	}
	host, port, hasPort := strings.Cut(authority, ":")
	if hasPort && port == "" {
		return "", "", errBadAuthority
	}
	return host, port, nil
}

func checkHost(host string) error {
	if host == "" {
		return redirectErr("empty host")
	}
	if inner, ok := strings.CutPrefix(host, "["); ok {
		inner = strings.TrimSuffix(inner, "]")
		addr, err := netip.ParseAddr(inner)
		if err != nil || !addr.Is6() || addr.Is4In6() || addr.Zone() != "" || addr.String() != inner {
			return redirectErr("IPv6 literal must be canonical, without zone")
		}
		return nil
	}
	if addr, err := netip.ParseAddr(host); err == nil && addr.Is4() {
		return nil // ParseAddr accepts only dotted-quad decimal without leading zeros
	}
	return checkDNSName(host)
}

// checkDNSName accepts lowercase LDH labels (RFC 1123) and refuses a name
// whose last label is numeric or hexadecimal-looking: the WHATWG URL parser
// reads such a host as an IPv4 address (127.1, 2130706433, 0x7f.0.0.1).
func checkDNSName(host string) error {
	if len(host) > maxHostLen {
		return redirectErr("host longer than %d bytes", maxHostLen)
	}
	labels := strings.Split(host, ".")
	for _, l := range labels {
		if !isLDHLabel(l) {
			return redirectErr("host must be lowercase letters, digits, dashes and dots")
		}
	}
	if last := labels[len(labels)-1]; strings.HasPrefix(last, "0x") || strings.Trim(last, "0123456789") == "" {
		return redirectErr("host looks like a numeric IPv4 address")
	}
	return nil
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

func checkPath(path string) error {
	lower := strings.ToLower(path)
	for _, t := range encodedTricks {
		if strings.Contains(lower, t) {
			return redirectErr("encoded wildcard, dot or separator in the path")
		}
	}
	for seg := range strings.SplitSeq(path, "/") {
		if seg == "." || seg == ".." {
			return redirectErr("dot segment in the path")
		}
	}
	return nil
}

func isLoopbackHost(authority string) bool {
	host, _, err := splitHostPort(authority)
	if err != nil {
		return false
	}
	return slices.Contains(loopbackHosts, host)
}

var errParserMismatch = errors.New("parser disagreement")

// checkParsesTheSame is defense in depth: the standard parser, which other
// components use, must see the same scheme and host, no userinfo, no opaque
// part and no fragment.
func checkParsesTheSame(raw, scheme, authority string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != scheme || u.Host != authority || u.User != nil || u.Opaque != "" || u.Fragment != "" {
		return fmt.Errorf("%w: %w", ErrInvalidRedirectURI, errParserMismatch)
	}
	return nil
}
