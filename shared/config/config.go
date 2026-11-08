package config

import (
	"time"
)

// Environment names the deployment tier.
type Environment string

// Known environments.
const (
	Development Environment = "development"
	Lab         Environment = "lab"
	Production  Environment = "production"
)

// Severity ranks a Finding.
type Severity int

// Severities, from least to most serious. Prohibited findings fail Load.
const (
	Info Severity = iota
	Warning
	Critical
	Prohibited
)

// String returns the lower-case name of the severity.
func (s Severity) String() string {
	switch s {
	case Info:
		return "info"
	case Warning:
		return "warning"
	case Critical:
		return "critical"
	case Prohibited:
		return "prohibited"
	}
	return "unknown"
}

// Finding is a notable property of a loaded configuration.
type Finding struct {
	Severity  Severity
	Code      string
	Message   string
	Reference string
}

// Argon2Settings are the password hashing parameters (OWASP Password Storage
// Cheat Sheet: m=19456 KiB, t=2, p=1 is the minimum for Argon2id).
type Argon2Settings struct {
	MemoryKiB     uint32 `env:"IAM_ARGON2_MEMORY_KIB" default:"19456" doc:"Argon2id memory cost in KiB. Below 19456 raises a critical alert."`
	Iterations    uint32 `env:"IAM_ARGON2_ITERATIONS" default:"2" doc:"Argon2id iteration count. Below 2 raises a critical alert."`
	Parallelism   uint32 `env:"IAM_ARGON2_PARALLELISM" default:"1" doc:"Argon2id degree of parallelism."`
	MaxConcurrent uint32 `env:"IAM_ARGON2_MAX_CONCURRENT" default:"2" doc:"Maximum password hashes computed at the same time (bounds memory use)."`
}

// Config is the whole service configuration.
type Config struct {
	ConfigFile      string        `env:"IAM_CONFIG_FILE" nofile:"true" doc:"Path of an optional JSON configuration file. Only the environment can set it."`
	Environment     Environment   `env:"IAM_ENVIRONMENT" default:"production" doc:"Deployment tier: development, lab or production."`
	HTTPAddr        string        `env:"IAM_HTTP_ADDR" default:":8080" doc:"Listen address of the HTTP server."`
	PublicBaseURL   string        `env:"IAM_PUBLIC_BASE_URL" required:"true" doc:"Public base URL; tenant issuers are built under it (ADR 0007). Must be https outside development."`
	DatabaseURL     string        `env:"IAM_DATABASE_URL" secret:"true" doc:"PostgreSQL connection URL. Optional; without it the service starts without a database. iam migrate requires it and connects as the migration role."`
	DatabaseAppRole string        `env:"IAM_DATABASE_APP_ROLE" default:"iam_app" doc:"Database role the service uses at runtime (subject to row level security). iam migrate grants it DML and refuses a name not matching iam_[a-z0-9_]* or a role that is a superuser, bypasses RLS or owns schema objects."`
	ShutdownTimeout time.Duration `env:"IAM_SHUTDOWN_TIMEOUT" default:"10s" doc:"Grace period for in-flight requests on shutdown."`
	CursorKey       string        `env:"IAM_CURSOR_KEY" secret:"true" doc:"Base64 key (at least 32 bytes) protecting pagination cursors. Required outside development, where it may be left empty."`
	Argon2          Argon2Settings
}
