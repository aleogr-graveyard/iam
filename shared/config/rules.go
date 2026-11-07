package config

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	refADR0012 = "ADR 0012"
	refOWASP   = "OWASP Password Storage Cheat Sheet"

	minCursorKeyBytes = 32
	owaspArgon2Memory = 19456
	owaspArgon2Iter   = 2
	maxShutdown       = 10 * time.Minute
)

// validate checks value ranges and evaluates prohibitions and alerts. It
// returns findings and the problems of invalid fields (without values).
func validate(c *Config) ([]Finding, []string) {
	var findings []Finding
	var problems []string
	for _, rule := range []func(*Config) ([]Finding, []string){
		environmentRule, publicBaseURLRule, databaseRule, cursorKeyRule, shutdownRule, argon2Rule,
	} {
		f, p := rule(c)
		findings = append(findings, f...)
		problems = append(problems, p...)
	}
	return findings, problems
}

func (e Environment) known() bool {
	return e == Development || e == Lab || e == Production
}

func environmentRule(c *Config) ([]Finding, []string) {
	switch c.Environment {
	case Development:
		return []Finding{{Severity: Warning, Code: "ENVIRONMENT_DEVELOPMENT", Message: "development environment: relaxed rules apply, never use it for real data", Reference: refADR0012}}, nil
	case Lab, Production:
		return nil, nil
	}
	return []Finding{{Severity: Prohibited, Code: "ENVIRONMENT_UNKNOWN", Message: "IAM_ENVIRONMENT must be development, lab or production", Reference: refADR0012}}, nil
}

func publicBaseURLRule(c *Config) ([]Finding, []string) {
	if c.PublicBaseURL == "" {
		return nil, nil // reported as required by the tag
	}
	u, ok := parseBaseURL(c.PublicBaseURL)
	if !ok {
		return nil, []string{"IAM_PUBLIC_BASE_URL: must be an http(s) URL with a host, no credentials, query, fragment or trailing slash"}
	}
	if u.Scheme != "https" && c.Environment != Development {
		return []Finding{{
			Severity:  Prohibited,
			Code:      "TLS_REQUIRED",
			Message:   "IAM_PUBLIC_BASE_URL must use https outside development (authentication endpoints without TLS)",
			Reference: refADR0012 + " item 11",
		}}, nil
	}
	return nil, nil
}

func parseBaseURL(s string) (*url.URL, bool) {
	u, err := url.Parse(s)
	if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return nil, false
	}
	// "?" and "#" cover empty queries and fragments; the round trip rejects
	// non-canonical spellings such as an upper-case scheme.
	if strings.ContainsAny(s, "?#") || strings.HasSuffix(u.Host, ":") || u.String() != s {
		return nil, false
	}
	if u.User != nil || strings.HasSuffix(u.Path, "/") {
		return nil, false
	}
	return u, true
}

func databaseRule(c *Config) ([]Finding, []string) {
	if c.DatabaseURL != "" {
		u, err := url.Parse(c.DatabaseURL)
		if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" || strings.Trim(u.Path, "/") == "" {
			return nil, []string{"IAM_DATABASE_URL: must be a postgres:// URL with a host and a database name"}
		}
	}
	if strings.TrimSpace(c.DatabaseAppRole) == "" {
		return nil, []string{"IAM_DATABASE_APP_ROLE: must not be empty"}
	}
	return nil, nil
}

func cursorKeyRule(c *Config) ([]Finding, []string) {
	if c.CursorKey == "" {
		if c.Environment == Development {
			return []Finding{{Severity: Warning, Code: "CURSOR_KEY_EPHEMERAL", Message: "IAM_CURSOR_KEY is empty: an ephemeral key is used and cursors do not survive a restart", Reference: refADR0012}}, nil
		}
		return []Finding{{Severity: Prohibited, Code: "CURSOR_KEY_REQUIRED", Message: "IAM_CURSOR_KEY must be set outside development; there are no default keys", Reference: refADR0012 + " item 6"}}, nil
	}
	key, err := base64.StdEncoding.DecodeString(c.CursorKey)
	if err != nil || len(key) < minCursorKeyBytes {
		return nil, []string{"IAM_CURSOR_KEY: must be standard base64 encoding at least 32 bytes"}
	}
	return nil, nil
}

func shutdownRule(c *Config) ([]Finding, []string) {
	if c.ShutdownTimeout <= 0 || c.ShutdownTimeout > maxShutdown {
		return nil, []string{"IAM_SHUTDOWN_TIMEOUT: must be greater than zero and at most 10m"}
	}
	return nil, nil
}

func argon2Rule(c *Config) ([]Finding, []string) {
	a := c.Argon2
	var problems []string
	check := func(env string, v, lo, hi uint32) {
		if v < lo || v > hi {
			problems = append(problems, fmt.Sprintf("%s: must be between %d and %d", env, lo, hi))
		}
	}
	check("IAM_ARGON2_MEMORY_KIB", a.MemoryKiB, 8, 1<<20)
	check("IAM_ARGON2_ITERATIONS", a.Iterations, 1, 64)
	check("IAM_ARGON2_PARALLELISM", a.Parallelism, 1, 64)
	check("IAM_ARGON2_MAX_CONCURRENT", a.MaxConcurrent, 1, 1024)
	if len(problems) > 0 {
		return nil, problems
	}
	var findings []Finding
	if a.MemoryKiB < owaspArgon2Memory {
		findings = append(findings, Finding{Severity: Critical, Code: "ARGON2_MEMORY_LOW", Message: fmt.Sprintf("Argon2 memory %d KiB is below the recommended %d KiB", a.MemoryKiB, owaspArgon2Memory), Reference: refOWASP})
	}
	if a.Iterations < owaspArgon2Iter {
		findings = append(findings, Finding{Severity: Critical, Code: "ARGON2_ITERATIONS_LOW", Message: fmt.Sprintf("Argon2 iterations %d is below the recommended %d", a.Iterations, owaspArgon2Iter), Reference: refOWASP})
	}
	return findings, nil
}
