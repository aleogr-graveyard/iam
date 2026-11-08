package tenancy_test

import (
	"net/netip"
	"net/url"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/aleogr/iam/shared/tenancy"
)

// FuzzValidateRedirectURI asserts that every accepted redirect URI satisfies
// the invariants of ADR 0012 and RFC 8252, whatever the input.
func FuzzValidateRedirectURI(f *testing.F) {
	for _, s := range []string{
		"https://app.example.com/cb", "http://127.0.0.1:8080/cb", "http://[::1]/cb",
		"http://localhost/cb", "https://*.example.com/", "https://a@b/", "https://a/#f",
		"javascript:alert(1)", "com.example.app:/cb", "https://ex%61mple.com/", "https://[fe80::1%25x]/",
		"http://0x7f.1/", "https://example.com/a/../b", "https://example.com:/", "HTTPS://A/",
	} {
		f.Add(s, false)
		f.Add(s, true)
	}
	f.Fuzz(func(t *testing.T, raw string, insecure bool) {
		if tenancy.ValidateRedirectURI(raw, insecure) != nil {
			return
		}
		checkAccepted(t, raw, insecure)
	})
}

func checkAccepted(t *testing.T, raw string, insecure bool) {
	t.Helper()
	if why := textViolation(raw); why != "" {
		t.Fatalf("accepted %q: %s", raw, why)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("accepted unparsable %q: %v", raw, err)
	}
	if why := urlViolation(u); why != "" {
		t.Fatalf("accepted %q: %s", raw, why)
	}
	if why := schemeViolation(u, insecure); why != "" {
		t.Fatalf("accepted %q: %s", raw, why)
	}
}

func textViolation(raw string) string {
	if len(raw) > tenancy.MaxRedirectURILen || !utf8.ValidString(raw) {
		return "over-long or invalid UTF-8"
	}
	if strings.IndexFunc(raw, func(r rune) bool { return r <= ' ' || r >= 0x7f }) >= 0 {
		return "non-printable or non-ASCII"
	}
	if strings.ContainsAny(raw, "*#\\") {
		return "wildcard, fragment or backslash"
	}
	return ""
}

func urlViolation(u *url.URL) string {
	if u.User != nil || u.Opaque != "" || u.Fragment != "" || u.Hostname() == "" {
		return "userinfo, opaque, fragment or empty host"
	}
	if strings.ContainsAny(u.Host, "%@") || strings.ToLower(u.Host) != u.Host {
		return "bad host"
	}
	if segs := strings.Split(u.Path, "/"); slices.Contains(segs, "..") || slices.Contains(segs, ".") {
		return "dot segment"
	}
	return ""
}

func schemeViolation(u *url.URL, insecure bool) string {
	switch {
	case u.Scheme == "https", u.Scheme == "http" && insecure:
		return ""
	case u.Scheme == "http":
		addr, err := netip.ParseAddr(u.Hostname())
		if err != nil || (addr != netip.MustParseAddr("127.0.0.1") && addr != netip.IPv6Loopback()) {
			return "non-loopback http"
		}
		return ""
	default:
		return "scheme " + u.Scheme
	}
}

// FuzzNormalizeEmail asserts that a normalized address is a fixed point and
// satisfies the length and character rules, whatever the input.
func FuzzNormalizeEmail(f *testing.F) {
	for _, s := range []string{
		"alice@example.com", " Alice@EXAMPLE.com ", "jose\u0301@CAFE\u0301.br", "a@@b", "@b", "a@",
		"a\u202e@b.c", "a@b..c", "a@\u0130.com", "\xff@x.y",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		out, err := tenancy.NormalizeEmail(in)
		if err != nil {
			return
		}
		again, err := tenancy.NormalizeEmail(out)
		if err != nil || again != out {
			t.Fatalf("not a fixed point: %q -> %q -> %q (%v)", in, out, again, err)
		}
		if !emailInvariantsHold(out) {
			t.Fatalf("invariant broken for %q -> %q", in, out)
		}
	})
}

func emailInvariantsHold(out string) bool {
	local, domain, _ := strings.Cut(out, "@")
	return len(out) <= 254 && local != "" && len(local) <= 64 && domain != "" &&
		strings.Count(out, "@") == 1 && utf8.ValidString(out) && strings.ToLower(domain) == domain
}
