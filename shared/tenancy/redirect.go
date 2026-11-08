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
	// uriTailChars is the RFC 3986 set allowed in the path and the query:
	// unreserved, sub-delims except '*' (wildcard) and the single quote, ":@/?",
	// and '%' (which must start a well-formed %HH).
	uriTailChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~!$&()+,;=:@/?%"
)

var (
	// The loopback IP literals of RFC 8252 section 7.3, in canonical form.
	loopbackHosts = []string{"127.0.0.1", "[::1]"}
	defaultPorts  = map[string]string{"https": "443", "http": "80"}
	// Percent-encoded characters refused in the path: NUL, '#', '%' (double
	// encoding), '*', '.', '/', '?' and '\'. Each can hide a wildcard, a dot
	// segment, a separator or a fragment from one parser but not from another
	// (threat model TM-OIDC-002).
	encodedTricks = []string{"%00", "%23", "%25", "%2a", "%2e", "%2f", "%3f", "%5c"}
	broadcastIPv4 = netip.AddrFrom4([4]byte{255, 255, 255, 255})
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
//   - the host is a DNS name already in canonical form (lowercase LDH, valid
//     A-labels that the IDNA Lookup profile maps to themselves, no numeric-
//     looking last label such as 127.1), a canonical IPv4 literal or a
//     canonical bracketed IPv6 literal without zone; IP literals that are
//     unspecified, broadcast, multicast, link-local, IPv4-mapped or
//     IPv4-compatible are refused; no percent-encoding in the host;
//   - if present, a port from 1 to 65535 without leading zeros, and not the
//     scheme's default port (443 for https, 80 for http);
//   - path and query use only RFC 3986 characters (unreserved, sub-delims
//     except '*' and the single quote, ":@/?") and well-formed %HH escapes;
//   - the path does not start with "//", has no "." segment, no ".." anywhere
//     (which also covers "..;" tricks), and no encoded NUL, '#', '%', '*',
//     '.', '/', '?' or '\'.
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
	authority, tail := splitAuthority(rest)
	if err := checkAuthority(scheme, authority); err != nil {
		return err
	}
	if err := checkTail(tail); err != nil {
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
// tail (path and query).
func splitAuthority(rest string) (authority, tail string) {
	end := strings.IndexAny(rest, "/?")
	if end < 0 {
		return rest, ""
	}
	return rest[:end], rest[end:]
}

func checkAuthority(scheme, authority string) error {
	if strings.ContainsAny(authority, "@%") {
		return redirectErr("userinfo or percent-encoding in the authority")
	}
	host, port, err := splitHostPort(authority)
	if err != nil {
		return err
	}
	if err := checkPort(scheme, port); err != nil {
		return err
	}
	return checkHost(host)
}

func checkPort(scheme, port string) error {
	if port == "" {
		return nil
	}
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 || strconv.Itoa(n) != port {
		return redirectErr("port must be 1 to 65535 without leading zeros")
	}
	if port == defaultPorts[scheme] {
		return redirectErr("explicit default port")
	}
	return nil
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
		return checkIPv6Literal(strings.TrimSuffix(inner, "]"))
	}
	// ParseAddr accepts only dotted-quad decimal without leading zeros.
	if addr, err := netip.ParseAddr(host); err == nil && addr.Is4() {
		return checkIP(addr)
	}
	canonical, err := canonicalDNSName(host)
	if err != nil {
		return fmt.Errorf("%w: host: %w", ErrInvalidRedirectURI, err)
	}
	if canonical != host {
		return redirectErr("host is not in canonical lowercase A-label form")
	}
	return nil
}

func checkIPv6Literal(inner string) error {
	addr, err := netip.ParseAddr(inner)
	if err != nil || !addr.Is6() || addr.Zone() != "" || addr.String() != inner {
		return redirectErr("IPv6 literal must be canonical, without zone")
	}
	return checkIP(addr)
}

// checkIP refuses addresses that never name a single reachable client host,
// and IPv6 forms that embed an IPv4 address (a second spelling of it).
func checkIP(a netip.Addr) error {
	if a.IsUnspecified() || a == broadcastIPv4 || a.IsMulticast() ||
		a.IsLinkLocalUnicast() || a.IsLinkLocalMulticast() || a.Is4In6() || isIPv4Compatible(a) {
		return redirectErr("unspecified, broadcast, multicast, link-local or IPv4-embedding address")
	}
	return nil
}

// isIPv4Compatible reports the deprecated ::a.b.c.d form (RFC 4291 section
// 2.5.5.1), other than :: and ::1.
func isIPv4Compatible(a netip.Addr) bool {
	if !a.Is6() || a.IsLoopback() || a.IsUnspecified() {
		return false
	}
	b := a.As16()
	return [12]byte(b[:12]) == [12]byte{}
}

// checkTail checks the path and the query.
func checkTail(tail string) error {
	if err := checkTailCharacters(tail); err != nil {
		return err
	}
	path, _, _ := strings.Cut(tail, "?")
	return checkPath(path)
}

func checkTailCharacters(tail string) error {
	for i := 0; i < len(tail); i++ {
		if !strings.ContainsRune(uriTailChars, rune(tail[i])) {
			return redirectErr("character outside RFC 3986 at offset %d of the path or query", i)
		}
		if tail[i] == '%' && (i+2 >= len(tail) || !isHex(tail[i+1]) || !isHex(tail[i+2])) {
			return redirectErr("malformed percent-encoding")
		}
	}
	return nil
}

func isHex(c byte) bool { return '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' }

func checkPath(path string) error {
	if strings.HasPrefix(path, "//") {
		return redirectErr("path starts with //")
	}
	lower := strings.ToLower(path)
	for _, t := range encodedTricks {
		if strings.Contains(lower, t) {
			return redirectErr("encoded NUL, '#', '%%', '*', '.', '/', '?' or '\\' in the path")
		}
	}
	if strings.Contains(path, "..") {
		return redirectErr("'..' in the path")
	}
	if slices.Contains(strings.Split(path, "/"), ".") {
		return redirectErr("dot segment in the path")
	}
	return nil
}

func isLoopbackHost(authority string) bool {
	host, _, err := splitHostPort(authority)
	return err == nil && slices.Contains(loopbackHosts, host)
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
