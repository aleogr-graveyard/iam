package tenancy_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aleogr/iam/shared/clock"
	"github.com/aleogr/iam/shared/crypto"
	"github.com/aleogr/iam/shared/ids"
	"github.com/aleogr/iam/shared/tenancy"
)

var t0 = time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)

func newID(t *testing.T) ids.ID {
	t.Helper()
	id, err := ids.Generator{Clock: &clock.Fixed{T: t0}}.New()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// ---------------------------------------------------------------- slugs

func TestValidateSlug(t *testing.T) {
	cases := []struct {
		slug string
		ok   bool
	}{
		{"acme", true},
		{"a1", true},
		{"0acme", true},
		{"acme-corp", true},
		{"a-b-c-1", true},
		{strings.Repeat("a", 63), true},
		{strings.Repeat("a", 64), false},
		{"", false},
		{"a", false},
		// "t" is the issuer path prefix (/t/<tenant>, ADR 0007); it is refused
		// by the length floor, so it needs no entry in the reserved list.
		{"t", false},
		{"Acme", false},
		{"ACME", false},
		{"-acme", false},
		{"acme-", false},
		{"ac_me", false},
		{"ac.me", false},
		{"ac me", false},
		{" acme", false},
		{"acme\n", false},
		{"a\u00e7me", false},
		{"system", false},
		{"admin", false},
		{"api", false},
		{"well-known", false},
		{".well-known", false},
		{"systems", true},
	}
	for _, c := range cases {
		err := tenancy.ValidateSlug(c.slug)
		if (err == nil) != c.ok {
			t.Errorf("ValidateSlug(%q) = %v, want ok=%v", c.slug, err, c.ok)
		}
		if err != nil && !errors.Is(err, tenancy.ErrInvalidSlug) {
			t.Errorf("ValidateSlug(%q): %v is not ErrInvalidSlug", c.slug, err)
		}
	}
}

func TestValidateSlug_SystemSlugIsReserved(t *testing.T) {
	if err := tenancy.ValidateSlug(tenancy.SystemTenantSlug); !errors.Is(err, tenancy.ErrInvalidSlug) {
		t.Fatalf("system slug must be reserved, got %v", err)
	}
}

// ---------------------------------------------------------------- e-mail

func TestNormalizeEmail(t *testing.T) {
	nfdLocal := "jose\u0301"  // e + combining acute (NFD)
	nfcLocal := "jos\u00e9"   // precomposed e-acute (NFC)
	nfdDomain := "cafe\u0301" // e + combining acute (NFD)
	cases := []struct {
		name, in, want string
	}{
		{"plain", "alice@example.com", "alice@example.com"},
		{"trim", "  alice@example.com\t", "alice@example.com"},
		{"domain lowercased", "alice@EXAMPLE.Com", "alice@example.com"},
		{"local case kept", "Alice.Smith@example.com", "Alice.Smith@example.com"},
		{"NFD local to NFC", nfdLocal + "@example.com", nfcLocal + "@example.com"},
		{"NFD domain to A-label", "a@" + nfdDomain + ".br", "a@xn--caf-dma.br"},
		{"IDN domain to lowercase A-label", "a@CAF\u00c9.BR", "a@xn--caf-dma.br"},
		{"A-label kept", "a@xn--caf-dma.br", "a@xn--caf-dma.br"},
		{"plus tag kept", "a+tag@example.com", "a+tag@example.com"},
		{"local at 64 octets", strings.Repeat("a", 64) + "@example.com", strings.Repeat("a", 64) + "@example.com"},
	}
	for _, c := range cases {
		got, err := tenancy.NormalizeEmail(c.in)
		if err != nil || got != c.want {
			t.Errorf("%s: NormalizeEmail(%q) = %q, %v; want %q", c.name, c.in, got, err, c.want)
		}
	}
}

func TestNormalizeEmail_NFCAndNFDAreEqual(t *testing.T) {
	a, errA := tenancy.NormalizeEmail("jose\u0301@example.com")
	b, errB := tenancy.NormalizeEmail("jos\u00e9@example.com")
	if errA != nil || errB != nil || a != b {
		t.Fatalf("%q (%v) != %q (%v)", a, errA, b, errB)
	}
}

// Every spelling of one domain must normalize to the same stored address.
func TestNormalizeEmail_EquivalentDomainSpellings(t *testing.T) {
	groups := map[string][]string{
		"a@xn--bcher-kva.de": {
			"a@b\u00fccher.de",                // U-label
			"a@xn--bcher-kva.de",              // A-label
			"a@XN--BCHER-KVA.DE",              // uppercase A-label
			"a@B\u00dcCHER.DE",                // uppercase U-label
			"a@bu\u0308cher.de",               // NFD
			"a@b\u00fccher\u3002de",           // ideographic full stop as the dot
			"a@b\u00fc\u00adcher.de",          // soft hyphen, ignored by IDNA mapping
			"a@\uff42\u00fccher.\uff44\uff45", // fullwidth letters
		},
		"a@example.com": {
			"a@example.com", "a@EXAMPLE.COM", "a@\uff45\uff58\uff41\uff4d\uff50\uff4c\uff45.com",
			"a@example\u3002com", "a@exa\u00admple.com", "a@example\uff0ecom",
		},
	}
	for want, spellings := range groups {
		for _, in := range spellings {
			got, err := tenancy.NormalizeEmail(in)
			if err != nil || got != want {
				t.Errorf("NormalizeEmail(%q) = %q, %v; want %q", in, got, err, want)
			}
		}
	}
}

// Golden results that pin the IDNA profile (UTS #46, non-transitional): an
// x/net or Go upgrade that changes any of them changes stored canonical forms
// and needs a data migration.
func TestNormalizeEmail_IDNAGolden(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a@\u00df.de", "a@xn--zca.de"},           // sharp s kept, not mapped to "ss"
		{"a@\u1e9e.de", "a@xn--zca.de"},           // capital sharp s folds to sharp s
		{"a@ss.de", "a@ss.de"},                    // ... and stays distinct from "ss"
		{"a@\U0001f600.com", "a@xn--e28h.com"},    // emoji label accepted by UTS #46 Lookup
		{"a@\u0130.com", "a@xn--i-9bb.com"},       // capital I with dot above
		{"a@\u0441\u0430.com", "a@xn--80a7a.com"}, // Cyrillic look-alike of "ca": kept, shown as A-label
	}
	for _, c := range cases {
		got, err := tenancy.NormalizeEmail(c.in)
		if err != nil || got != c.want {
			t.Errorf("NormalizeEmail(%q) = %q, %v; want %q", c.in, got, err, c.want)
		}
	}
}

func TestNormalizeEmail_Refused(t *testing.T) {
	// 64 + 1 + 189 = 254 octets is the maximum.
	domain189 := strings.Repeat("d", 63) + "." + strings.Repeat("d", 63) + "." + strings.Repeat("d", 57) + ".com"
	if n := len("@" + domain189); n != 190 {
		t.Fatalf("fixture: %d", n)
	}
	max := strings.Repeat("a", 64) + "@" + domain189
	if got, err := tenancy.NormalizeEmail(max); err != nil || len(got) != 254 {
		t.Fatalf("254-octet address refused: %v", err)
	}
	cases := map[string]string{
		"empty":               "",
		"blank":               "   ",
		"no at":               "alice.example.com",
		"two at":              "alice@bob@example.com",
		"empty local":         "@example.com",
		"empty domain":        "alice@",
		"255 octets":          strings.Repeat("a", 64) + "@" + strings.Repeat("d", 63) + "." + strings.Repeat("d", 63) + "." + strings.Repeat("d", 58) + ".com",
		"local 65 octets":     strings.Repeat("a", 65) + "@example.com",
		"space inside":        "ali ce@example.com",
		"tab inside":          "alice@exa\tmple.com",
		"newline inside":      "alice@example.com\nBcc: x@y",
		"NUL":                 "alice\x00@example.com",
		"DEL":                 "alice\x7f@example.com",
		"C1 control":          "alice\u0085@example.com",
		"no-break space":      "alice\u00a0@example.com",
		"bidi override":       "alice\u202e@example.com",
		"invalid UTF-8":       "alice\xff@example.com",
		"domain leading dot":  "alice@.example.com",
		"domain trailing dot": "alice@example.com.",
		"domain empty label":  "alice@example..com",
		"angle brackets":      "a@<x>.com",
		"slash in domain":     "a@ex/ample.com",
		"quotes in domain":    "a@\"x\".com",
		"underscore domain":   "a@ex_ample.com",
		"single label":        "a@localhost",
		"label leading dash":  "a@-example.com",
		"label trailing dash": "a@example-.com",
		"label 64 octets":     "a@" + strings.Repeat("d", 64) + ".com",
		"invalid A-label":     "a@xn--a.com",
		"numeric last label":  "a@1.2.3.4",
		"IP literal":          "a@[192.0.2.1]",
		"zero-width space":    "al\u200bice@example.com",
		"LRM":                 "alice\u200e@example.com",
		"BOM":                 "\ufeffalice@example.com",
		"arabic letter mark":  "alice\u061c@example.com",
		"soft hyphen local":   "al\u00adice@example.com",
		"hangul filler":       "alice\u3164@example.com",
		"variation selector":  "alice\ufe0f@example.com",
		"input over 1 KiB":    strings.Repeat(" ", 1020) + "a@example.com",
	}
	for name, in := range cases {
		got, err := tenancy.NormalizeEmail(in)
		if err == nil {
			t.Errorf("%s: accepted as %q", name, got)
			continue
		}
		if !errors.Is(err, tenancy.ErrInvalidEmail) {
			t.Errorf("%s: %v is not ErrInvalidEmail", name, err)
		}
	}
}

// M7: a 64-octet local part and a 190-octet domain are each within their own
// limit; only the 254-octet total refuses the address.
func TestNormalizeEmail_TotalLengthBranch(t *testing.T) {
	domain190 := strings.Repeat("d", 63) + "." + strings.Repeat("d", 63) + "." + strings.Repeat("d", 58) + ".com"
	in := strings.Repeat("a", 64) + "@" + domain190
	if len(domain190) != 190 || len(in) != 255 {
		t.Fatalf("fixture: domain %d, total %d", len(domain190), len(in))
	}
	_, err := tenancy.NormalizeEmail(in)
	if !errors.Is(err, tenancy.ErrInvalidEmail) || !strings.Contains(err.Error(), "longer than 254 octets") {
		t.Fatalf("err = %v, want the total-length refusal", err)
	}
}

// Errors never echo the address: it is personal data.
func TestNormalizeEmail_ErrorHasNoPersonalData(t *testing.T) {
	for _, in := range []string{"secret.person@", "secret.person@@x.com", "secret.person x@x.com", strings.Repeat("secretperson", 30) + "@x.com"} {
		_, err := tenancy.NormalizeEmail(in)
		if err == nil {
			t.Fatalf("%q accepted", in)
		}
		if strings.Contains(err.Error(), "secret") {
			t.Errorf("error leaks input: %v", err)
		}
	}
}

func TestNormalizeEmail_Idempotent(t *testing.T) {
	for _, in := range []string{"Alice@EXAMPLE.COM", "jose\u0301@CAFE\u0301.br", " x@Y.z "} {
		once, err := tenancy.NormalizeEmail(in)
		if err != nil {
			t.Fatal(err)
		}
		twice, err := tenancy.NormalizeEmail(once)
		if err != nil || twice != once {
			t.Errorf("not idempotent: %q -> %q -> %q (%v)", in, once, twice, err)
		}
	}
}

// ---------------------------------------------------------------- redirect URIs

func TestValidateRedirectURI(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("a", tenancy.MaxRedirectURILen-len("https://example.com/"))
	cases := []struct {
		name     string
		uri      string
		insecure bool
		ok       bool
	}{
		{"https", "https://app.example.com/callback", false, true},
		{"https with port", "https://app.example.com:8443/cb", false, true},
		{"https with query", "https://app.example.com/cb?x=1", false, true},
		{"https ipv4", "https://192.0.2.1/cb", false, true},
		{"https ipv6", "https://[2001:db8::1]/cb", false, true},
		{"https no path", "https://app.example.com", false, true},
		{"https localhost", "https://localhost/cb", false, true},
		{"at max length", long, false, true},
		{"over max length", long + "a", false, false},

		// RFC 8252 section 7.3: loopback IP literals over http, any port.
		{"loopback v4", "http://127.0.0.1/cb", false, true},
		{"loopback v4 port", "http://127.0.0.1:51004/cb", false, true},
		{"loopback v6", "http://[::1]/cb", false, true},
		{"loopback v6 port", "http://[::1]:8080/cb", false, true},
		// RFC 8252 section 8.3: "localhost" is NOT RECOMMENDED; it may resolve
		// elsewhere, so it does not get the loopback exemption.
		{"http localhost", "http://localhost/cb", false, false},
		{"http localhost dev", "http://localhost:3000/cb", true, true},
		{"loopback v4 non-canonical", "http://127.1/cb", false, false},
		{"loopback v4 leading zero", "http://127.000.000.001/cb", false, false},
		{"loopback v4 decimal", "http://2130706433/cb", false, false},
		{"loopback v4 hex", "http://0x7f.0.0.1/cb", false, false},
		{"loopback v6 long form", "http://[0:0:0:0:0:0:0:1]/cb", false, false},
		{"loopback v6 mapped v4", "http://[::ffff:127.0.0.1]/cb", false, false},
		{"loopback other 127/8", "http://127.0.0.2/cb", false, false},

		{"http with allowInsecureHTTP=false", "http://example.com/cb", false, false},
		{"http with allowInsecureHTTP=true", "http://example.com/cb", true, true},

		// ADR 0012: no wildcard.
		{"wildcard host", "https://*.example.com/cb", false, false},
		{"wildcard host dev", "http://*.example.com/cb", true, false},
		{"star in path", "https://example.com/*", false, false},
		{"star in query", "https://example.com/cb?a=*", false, false},
		{"encoded star", "https://example.com/%2a", false, false},

		{"fragment", "https://example.com/cb#x", false, false},
		{"empty fragment", "https://example.com/cb#", false, false},
		{"userinfo", "https://a@b.example.com/cb", false, false},
		{"userinfo with password", "https://a:b@example.com/cb", false, false},
		{"empty userinfo", "https://@example.com/cb", false, false},

		{"javascript", "javascript:alert(1)", false, false},
		{"javascript dev", "javascript:alert(1)", true, false},
		{"data", "data:text/html,<script>alert(1)</script>", false, false},
		{"private-use scheme", "com.example.app:/cb", false, false},
		{"private-use scheme authority", "com.example.app://cb", false, false},
		{"ftp", "ftp://example.com/cb", false, false},
		{"uppercase scheme", "HTTPS://example.com/cb", false, false},
		{"uppercase host", "https://Example.com/cb", false, false},
		{"relative", "/cb", false, false},
		{"scheme relative", "//example.com/cb", false, false},
		{"opaque https", "https:example.com/cb", false, false},
		{"empty", "", false, false},
		{"empty host", "https:///cb", false, false},
		{"empty host with port", "https://:443/cb", false, false},
		{"empty port", "https://example.com:/cb", false, false},
		{"port zero", "https://example.com:0/cb", false, false},
		{"port too large", "https://example.com:65536/cb", false, false},
		{"non-numeric port", "https://example.com:http/cb", false, false},

		{"non-ASCII host", "https://ex\u00e4mple.com/cb", false, false},
		{"punycode host", "https://xn--exmple-cua.com/cb", false, true},
		{"non-ASCII path", "https://example.com/caf\u00e9", false, false},
		{"percent in host", "https://ex%61mple.com/cb", false, false},
		{"ipv6 zone", "https://[fe80::1%25eth0]/cb", false, false},
		{"host label leading dash", "https://-example.com/cb", false, false},
		{"host trailing dot", "https://example.com./cb", false, false},
		{"host empty label", "https://example..com/cb", false, false},
		{"host underscore", "https://ex_ample.com/cb", false, false},
		{"space", "https://example.com/c b", false, false},
		{"tab", "https://example.com/c\tb", false, false},
		{"newline", "https://example.com/cb\n", false, false},
		{"NUL", "https://example.com/cb\x00", false, false},
		{"backslash", "https://example.com\\@evil.com/cb", false, false},
		{"dot segment", "https://example.com/a/../cb", false, false},
		{"single dot segment", "https://example.com/./cb", false, false},
		{"trailing dot segment", "https://example.com/a/..", false, false},
		{"encoded dot segment", "https://example.com/a/%2e%2e/cb", false, false},
		{"encoded slash", "https://example.com/a%2Fb", false, false},
		{"encoded backslash", "https://example.com/a%5cb", false, false},
		{"percent-encoded path ok", "https://example.com/a%20b", false, true},
		{"sub-delims ok", "https://example.com/a,b=c!$&()+@:~_.-?q=a:b@c/d;e", false, true},
		{"uppercase hex ok", "https://example.com/a%C3%A9", false, true},

		// M1: RFC 3986 character allowlist in path and query.
		{"angle brackets", "https://example.com/<x>", false, false},
		{"double quote", "https://example.com/\"x", false, false},
		{"single quote", "https://example.com/'x", false, false},
		{"backtick", "https://example.com/`x", false, false},
		{"braces", "https://example.com/{x}", false, false},
		{"pipe", "https://example.com/a|b", false, false},
		{"caret", "https://example.com/a^b", false, false},
		{"brackets in path", "https://example.com/[x]", false, false},
		{"brackets in query", "https://example.com/cb?a[]=1", false, false},
		{"bad percent", "https://example.com/a%zz", false, false},
		{"truncated percent", "https://example.com/a%2", false, false},

		// M2: path tricks.
		{"dot dot with param", "https://example.com/a/..;/cb", false, false},
		{"dot dot inside segment", "https://example.com/a..b", false, false},
		{"double-encoded dot", "https://example.com/a/%252e%252e/cb", false, false},
		{"encoded question mark", "https://example.com/a%3fb", false, false},
		{"encoded hash", "https://example.com/a%23b", false, false},
		{"encoded NUL", "https://example.com/a%00b", false, false},
		{"path starts with double slash", "https://example.com//evil.com/cb", false, false},

		// M3: special IP literals.
		{"unspecified v4", "https://0.0.0.0/cb", false, false},
		{"unspecified v6", "https://[::]/cb", false, false},
		{"broadcast", "https://255.255.255.255/cb", false, false},
		{"multicast v4", "https://224.0.0.1/cb", false, false},
		{"multicast v6", "https://[ff02::1]/cb", false, false},
		{"link-local v4", "https://169.254.169.254/cb", false, false},
		{"link-local v6", "https://[fe80::1]/cb", false, false},
		{"v4-compatible v6", "https://[::7f00:1]/cb", false, false},
		{"v4-mapped v6 https", "https://[::ffff:c000:201]/cb", false, false},

		// M4: canonical form.
		{"default https port", "https://example.com:443/cb", false, false},
		{"default http port", "http://127.0.0.1:80/cb", false, false},
		{"default http port dev", "http://example.com:80/cb", true, false},
		{"http on 443 ok", "http://127.0.0.1:443/cb", false, true},
		{"invalid punycode", "https://xn--a.com/cb", false, false},
		{"punycode of ASCII", "https://xn--abc-.com/cb", false, false},
		// IDNA golden hosts: the A-label is the only accepted spelling.
		{"sharp s A-label", "https://xn--zca.de/cb", false, true},
		{"emoji A-label", "https://xn--e28h.com/cb", false, true},
		{"cyrillic A-label", "https://xn--80a7a.com/cb", false, true},

		// N1: no path parameters (';'), which Tomcat and Spring strip before
		// normalizing, turning "/.;/" and "/;./" into dot segments.
		{"dot semicolon segment", "https://example.com/a/.;/cb", false, false},
		{"semicolon dot segment", "https://example.com/a/;./cb", false, false},
		{"path parameter", "https://example.com/cb;jsessionid=x", false, false},
		{"semicolon in query ok", "https://example.com/cb?a=1;b=2", false, true},
	}
	for _, c := range cases {
		err := tenancy.ValidateRedirectURI(c.uri, c.insecure)
		if (err == nil) != c.ok {
			t.Errorf("%s: ValidateRedirectURI(%q, %v) = %v, want ok=%v", c.name, c.uri, c.insecure, err, c.ok)
		}
		if err != nil && !errors.Is(err, tenancy.ErrInvalidRedirectURI) {
			t.Errorf("%s: %v is not ErrInvalidRedirectURI", c.name, err)
		}
	}
}

// ---------------------------------------------------------------- statuses and types

func TestStatus(t *testing.T) {
	cases := map[tenancy.Status]bool{
		tenancy.StatusActive: true, tenancy.StatusSuspended: true, "": false, "Active": false, "deleted": false,
	}
	for s, want := range cases {
		if s.Valid() != want {
			t.Errorf("Status(%q).Valid() = %v", s, !want)
		}
	}
	if tenancy.StatusActive != "active" || tenancy.StatusSuspended != "suspended" {
		t.Fatal("wire values changed")
	}
}

func TestClientType(t *testing.T) {
	cases := map[tenancy.ClientType]bool{
		tenancy.ClientPublic: true, tenancy.ClientConfidential: true, "": false, "Public": false, "private": false,
	}
	for c, want := range cases {
		if c.Valid() != want {
			t.Errorf("ClientType(%q).Valid() = %v", c, !want)
		}
	}
	if tenancy.ClientPublic != "public" || tenancy.ClientConfidential != "confidential" {
		t.Fatal("wire values changed")
	}
}

// ---------------------------------------------------------------- names

func TestValidateName(t *testing.T) {
	cases := []struct {
		name string
		ok   bool
	}{
		{"Acme", true},
		{"Acme Corp.", true},
		{"Soci\u00e9t\u00e9 G\u00e9n\u00e9rale", true},
		{"\u6771\u4eac", true},
		{"\U0001f600 team", true},
		// Format (Cf) and default-ignorable characters are invisible; an emoji
		// ZWJ sequence or a variation selector is the accepted casualty.
		{"\U0001f469\u200d\U0001f4bb team", false},
		{"Ac\u200bme", false},
		{"Ac\u200eme", false},
		{"\ufeffAcme", false},
		{"Ac\u061cme", false},
		{"Ac\u00adme", false},
		{"Ac\u3164me", false},
		{"Acme\ufe0f", false},
		{strings.Repeat("\u00e9", tenancy.MaxNameLen), true},
		{strings.Repeat("\u00e9", tenancy.MaxNameLen+1), false},
		{"", false},
		{"   ", false},
		{" Acme", false},
		{"Acme ", false},
		{"Ac\nme", false},
		{"Ac\tme", false},
		{"Ac\x00me", false},
		{"Ac\x7fme", false},
		{"Ac\u0085me", false},
		{"Acme\u202egnp.exe", false},
		{"Acme\u2066x", false},
		{"Ac\xffme", false},
	}
	for _, c := range cases {
		err := tenancy.ValidateName(c.name)
		if (err == nil) != c.ok {
			t.Errorf("ValidateName(%q) = %v, want ok=%v", c.name, err, c.ok)
		}
		if err != nil && !errors.Is(err, tenancy.ErrInvalid) {
			t.Errorf("ValidateName(%q): %v is not ErrInvalid", c.name, err)
		}
	}
}

// ---------------------------------------------------------------- entities

func validTenant(t *testing.T) tenancy.Tenant {
	return tenancy.Tenant{ID: ids.TenantID{ID: newID(t)}, Slug: "acme", Name: "Acme", Status: tenancy.StatusActive, CreatedAt: t0}
}

func TestTenant_Validate(t *testing.T) {
	if err := validTenant(t).Validate(); err != nil {
		t.Fatal(err)
	}
	// User input with the system slug is refused: Validate fails closed.
	sys := validTenant(t)
	sys.Slug = tenancy.SystemTenantSlug
	if err := sys.Validate(); !errors.Is(err, tenancy.ErrInvalidSlug) {
		t.Fatalf("Validate accepted the system slug: %v", err)
	}
	mut := map[string]func(*tenancy.Tenant){
		"zero id":        func(x *tenancy.Tenant) { x.ID = ids.TenantID{} },
		"bad slug":       func(x *tenancy.Tenant) { x.Slug = "Acme" },
		"reserved slug":  func(x *tenancy.Tenant) { x.Slug = "admin" },
		"empty name":     func(x *tenancy.Tenant) { x.Name = "" },
		"unknown status": func(x *tenancy.Tenant) { x.Status = "deleted" },
		"empty status":   func(x *tenancy.Tenant) { x.Status = "" },
		"zero time":      func(x *tenancy.Tenant) { x.CreatedAt = time.Time{} },
	}
	for name, m := range mut {
		x := validTenant(t)
		m(&x)
		if err := x.Validate(); !errors.Is(err, tenancy.ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
		}
	}
}

func TestNewSystemTenant(t *testing.T) {
	id := ids.TenantID{ID: newID(t)}
	sys, err := tenancy.NewSystemTenant(id, "Operators", &clock.Fixed{T: t0})
	if err != nil {
		t.Fatal(err)
	}
	want := tenancy.Tenant{ID: id, Slug: tenancy.SystemTenantSlug, Name: "Operators", Status: tenancy.StatusActive, CreatedAt: t0}
	if sys != want {
		t.Fatalf("NewSystemTenant = %+v, want %+v", sys, want)
	}
	if !sys.IsSystem() {
		t.Fatal("not the system tenant")
	}
	if err := sys.ValidateStored(); err != nil {
		t.Fatalf("ValidateStored refused the system tenant: %v", err)
	}
	if err := sys.Validate(); err == nil {
		t.Fatal("Validate accepted the system tenant")
	}
	for name, args := range map[string]struct {
		id   ids.TenantID
		name string
		clk  clock.Clock
	}{
		"zero id":   {ids.TenantID{}, "Operators", &clock.Fixed{T: t0}},
		"bad name":  {id, "", &clock.Fixed{T: t0}},
		"no clock":  {id, "Operators", nil},
		"zero time": {id, "Operators", &clock.Fixed{}},
	} {
		if _, err := tenancy.NewSystemTenant(args.id, args.name, args.clk); !errors.Is(err, tenancy.ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
		}
	}
}

// ValidateStored accepts what Validate accepts, plus the system tenant only.
func TestTenant_ValidateStored(t *testing.T) {
	if err := validTenant(t).ValidateStored(); err != nil {
		t.Fatal(err)
	}
	for _, slug := range []string{"admin", "api", "well-known", "Acme", "-x"} {
		x := validTenant(t)
		x.Slug = slug
		if err := x.ValidateStored(); !errors.Is(err, tenancy.ErrInvalid) {
			t.Errorf("slug %q: err = %v", slug, err)
		}
	}
	x := validTenant(t)
	x.Slug = tenancy.SystemTenantSlug
	x.Status = "deleted"
	if err := x.ValidateStored(); !errors.Is(err, tenancy.ErrInvalid) {
		t.Fatalf("stored system tenant with a bad status: %v", err)
	}
}

func TestTenant_IsSystem(t *testing.T) {
	x := validTenant(t)
	if x.IsSystem() {
		t.Fatal("acme is not the system tenant")
	}
	x.Slug = tenancy.SystemTenantSlug
	if !x.IsSystem() {
		t.Fatal("system slug is the system tenant")
	}
}

func validOrg(t *testing.T) tenancy.Organization {
	return tenancy.Organization{ID: ids.OrganizationID{ID: newID(t)}, TenantID: ids.TenantID{ID: newID(t)}, Slug: "sales", Name: "Sales", CreatedAt: t0}
}

func TestOrganization_Validate(t *testing.T) {
	if err := validOrg(t).Validate(); err != nil {
		t.Fatal(err)
	}
	mut := map[string]func(*tenancy.Organization){
		"zero id":         func(x *tenancy.Organization) { x.ID = ids.OrganizationID{} },
		"zero tenant":     func(x *tenancy.Organization) { x.TenantID = ids.TenantID{} },
		"bad slug":        func(x *tenancy.Organization) { x.Slug = "-sales" },
		"reserved slug":   func(x *tenancy.Organization) { x.Slug = "system" },
		"empty name":      func(x *tenancy.Organization) { x.Name = "" },
		"control in name": func(x *tenancy.Organization) { x.Name = "a\nb" },
		"zero time":       func(x *tenancy.Organization) { x.CreatedAt = time.Time{} },
	}
	for name, m := range mut {
		x := validOrg(t)
		m(&x)
		if err := x.Validate(); !errors.Is(err, tenancy.ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
		}
	}
}

func validUser(t *testing.T) tenancy.User {
	return tenancy.User{ID: ids.UserID{ID: newID(t)}, TenantID: ids.TenantID{ID: newID(t)}, Email: "alice@example.com", Status: tenancy.StatusActive, CreatedAt: t0}
}

func TestUser_Validate(t *testing.T) {
	if err := validUser(t).Validate(); err != nil {
		t.Fatal(err)
	}
	mut := map[string]func(*tenancy.User){
		"zero id":              func(x *tenancy.User) { x.ID = ids.UserID{} },
		"zero tenant":          func(x *tenancy.User) { x.TenantID = ids.TenantID{} },
		"empty email":          func(x *tenancy.User) { x.Email = "" },
		"invalid email":        func(x *tenancy.User) { x.Email = "secret.person" },
		"unnormalized domain":  func(x *tenancy.User) { x.Email = "secret.person@EXAMPLE.com" },
		"unnormalized spaces":  func(x *tenancy.User) { x.Email = " secret.person@example.com" },
		"unnormalized unicode": func(x *tenancy.User) { x.Email = "secret.jose\u0301@example.com" },
		"unknown status":       func(x *tenancy.User) { x.Status = "x" },
		"zero time":            func(x *tenancy.User) { x.CreatedAt = time.Time{} },
	}
	for name, m := range mut {
		x := validUser(t)
		m(&x)
		err := x.Validate()
		if !errors.Is(err, tenancy.ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
			continue
		}
		if strings.Contains(err.Error(), "secret") {
			t.Errorf("%s: error leaks the e-mail: %v", name, err)
		}
	}
}

func validMembership(t *testing.T) tenancy.Membership {
	return tenancy.Membership{TenantID: ids.TenantID{ID: newID(t)}, OrganizationID: ids.OrganizationID{ID: newID(t)}, UserID: ids.UserID{ID: newID(t)}, CreatedAt: t0}
}

func TestMembership_Validate(t *testing.T) {
	if err := validMembership(t).Validate(); err != nil {
		t.Fatal(err)
	}
	mut := map[string]func(*tenancy.Membership){
		"zero tenant": func(x *tenancy.Membership) { x.TenantID = ids.TenantID{} },
		"zero org":    func(x *tenancy.Membership) { x.OrganizationID = ids.OrganizationID{} },
		"zero user":   func(x *tenancy.Membership) { x.UserID = ids.UserID{} },
		"zero time":   func(x *tenancy.Membership) { x.CreatedAt = time.Time{} },
	}
	for name, m := range mut {
		x := validMembership(t)
		m(&x)
		if err := x.Validate(); !errors.Is(err, tenancy.ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
		}
	}
}

func secretHash(t *testing.T) []byte {
	t.Helper()
	h, err := crypto.HashSecret(bytes.Repeat([]byte{9}, 32))
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func publicApp(t *testing.T) tenancy.Application {
	return tenancy.Application{
		ID: ids.ApplicationID{ID: newID(t)}, TenantID: ids.TenantID{ID: newID(t)}, Name: "Web",
		Type: tenancy.ClientPublic, RedirectURIs: []string{"https://app.example.com/cb", "http://127.0.0.1/cb"}, CreatedAt: t0,
	}
}

func confidentialApp(t *testing.T) tenancy.Application {
	a := publicApp(t)
	a.Type = tenancy.ClientConfidential
	a.SecretHash = secretHash(t)
	return a
}

func manyURIs(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = "https://app.example.com/cb" + strings.Repeat("x", i)
	}
	return out
}

func TestApplication_Validate(t *testing.T) {
	if err := publicApp(t).Validate(false); err != nil {
		t.Fatalf("public: %v", err)
	}
	if err := confidentialApp(t).Validate(false); err != nil {
		t.Fatalf("confidential: %v", err)
	}
	full := publicApp(t)
	full.RedirectURIs = manyURIs(tenancy.MaxRedirectURIs)
	if err := full.Validate(false); err != nil {
		t.Fatalf("%d URIs: %v", tenancy.MaxRedirectURIs, err)
	}
	dev := publicApp(t)
	dev.RedirectURIs = []string{"http://localhost:3000/cb"}
	if err := dev.Validate(true); err != nil {
		t.Fatalf("dev http with allowInsecureHTTP: %v", err)
	}
	if err := dev.Validate(false); !errors.Is(err, tenancy.ErrInvalid) {
		t.Fatalf("dev http without allowInsecureHTTP: %v", err)
	}

	mut := map[string]func(*tenancy.Application){
		"zero id":             func(x *tenancy.Application) { x.ID = ids.ApplicationID{} },
		"zero tenant":         func(x *tenancy.Application) { x.TenantID = ids.TenantID{} },
		"empty name":          func(x *tenancy.Application) { x.Name = "" },
		"unknown type":        func(x *tenancy.Application) { x.Type = "private" },
		"empty type":          func(x *tenancy.Application) { x.Type = "" },
		"no redirect":         func(x *tenancy.Application) { x.RedirectURIs = nil },
		"empty redirect list": func(x *tenancy.Application) { x.RedirectURIs = []string{} },
		"too many redirects":  func(x *tenancy.Application) { x.RedirectURIs = manyURIs(tenancy.MaxRedirectURIs + 1) },
		"invalid redirect": func(x *tenancy.Application) {
			x.RedirectURIs = []string{"https://ok.example.com/cb", "https://*.example.com/cb"}
		},
		"duplicate redirect": func(x *tenancy.Application) {
			x.RedirectURIs = []string{"https://a.example.com/cb", "https://a.example.com/cb"}
		},
		"public with secret":       func(x *tenancy.Application) { x.SecretHash = secretHash(t) },
		"public with empty secret": func(x *tenancy.Application) { x.SecretHash = []byte{} },
		"zero time":                func(x *tenancy.Application) { x.CreatedAt = time.Time{} },
	}
	for name, m := range mut {
		x := publicApp(t)
		m(&x)
		if err := x.Validate(false); !errors.Is(err, tenancy.ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
		}
	}

	conf := map[string]func(*tenancy.Application){
		"no secret":    func(x *tenancy.Application) { x.SecretHash = nil },
		"short secret": func(x *tenancy.Application) { x.SecretHash = x.SecretHash[:31] },
		"long secret":  func(x *tenancy.Application) { x.SecretHash = append(bytes.Clone(x.SecretHash), 0) },
	}
	for name, m := range conf {
		x := confidentialApp(t)
		m(&x)
		if err := x.Validate(false); !errors.Is(err, tenancy.ErrInvalid) {
			t.Errorf("confidential %s: err = %v, want ErrInvalid", name, err)
		}
	}
}

// The error names the position of a bad redirect URI, never its content.
func TestApplication_ValidateErrorDoesNotEchoURI(t *testing.T) {
	x := publicApp(t)
	x.RedirectURIs = []string{"https://ok.example.com/cb", "https://secret.example.com/cb#frag"}
	err := x.Validate(false)
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("err = %v", err)
	}
	if !errors.Is(err, tenancy.ErrInvalidRedirectURI) {
		t.Fatalf("err %v does not wrap ErrInvalidRedirectURI", err)
	}
}
