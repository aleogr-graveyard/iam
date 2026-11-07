# Marco 1A ("contratos e fundação") Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Define and implement the shared contracts and foundation that AuthN and AuthZ build on: typed configuration with the absolute prohibitions, the cryptography port, identifiers, the state-machine library, the tenancy domain, storage ports with a contract suite, the Postgres adapter with migrations and RLS, the hash-chained audit log, and the admin API (Protobuf + Connect) behind a fail-closed authorizer.

**Architecture:** Pure domain and port packages under `shared/`; implementations under `adapters/` (`memory` for unit tests and as the contract-suite reference, `postgres` for production); wiring in `cmd/iam`. Every storage call is tenant-scoped through a transaction that also sets the Postgres RLS variable. Contracts C1–C6 are fixed here; plans 1B (AuthZ), 1C (AuthN) and 1D (integration and deploy) consume them.

**Tech Stack:** Go 1.27, standard library, `golang.org/x/crypto/argon2`, `github.com/go-jose/go-jose/v4` (≥ 4.1.8), `pgx/v5`, `protobuf` + `connect`, `buf`, PostgreSQL 16.

**Spec:** `docs/superpowers/specs/2026-11-06-iam-design.md` (sections 4, 5, 8, 9, 10, 11) and the ADRs it lists; `docs/threat-model.md`; `CLAUDE.md`.

## Milestone 1 decomposition

| Plan | Scope (roadmap IDs) | Depends on |
|---|---|---|
| **1A (this plan)** | SHARED-001..006, SHARED-010..015 | Marco 0 |
| 1B | AUTHZ-001, 002, 010..014 | 1A |
| 1C | AUTHN-001, 002, 010..014 | 1A |
| 1D | AUTHZ-015, 016, AUTHN-015, TEST-010, 011, OPS-101, 102 and the inherited follow-ups | 1B, 1C |

## Global Constraints

- Go `1.27`; module `github.com/aleogr/iam`; new publishable module `github.com/aleogr/iam/shared/statemachine` (own `go.mod`, listed in `go.work`, standard library only).
- Runtime dependencies allowed (spec section 10): standard library, `golang.org/x/crypto`, `github.com/go-jose/go-jose/v4` ≥ 4.1.8, `github.com/jackc/pgx/v5`, `google.golang.org/protobuf`, `connectrpc.com/connect`. Nothing else.
- `crypto/*`, `golang.org/x/crypto` and `go-jose` are imported only in `shared/crypto` (depguard already enforces it); randomness only through `shared/crypto`.
- Cores and `shared/` domain packages take an injected `clock.Clock`; never `time.Now` there.
- Every tenant-owned row has `tenant_id uuid not null`; every storage call carries a `ids.TenantID`; RLS is the second barrier.
- Errors returned to clients never contain secrets, DSNs or raw personal data; JSON decoding goes through `shared/jsonx` limits.
- One PR per task, Conventional Commits title, Portuguese description; the roadmap row is marked ✅ in the same PR. Tasks touching `shared/crypto`, `shared/tenancy`, `shared/storage`, `adapters/postgres`, `api/**`, `go.mod`/`go.sum` are sensitive: the owner merges.
- Lab budget: service pool `MaxConns = 2`, one instance; no `LISTEN/NOTIFY`.

## Review Focus

1. **A transaction opened for tenant A reading tenant B's rows** (missing `tenant_id` in a query, RLS variable not set or leaking across pooled connections) — must return nothing. Test: Task 8 contract suite `CrossTenantIsolation` and Task 9 `TestRLS_BlocksQueryWithoutTenantFilter` + `TestRLS_SettingDoesNotLeakAcrossTransactions`.
2. **Argon2id as a denial-of-service vector** (TM-PWD-003): concurrent hashes beyond the limit queue with a timeout instead of exhausting memory. Test: Task 3 `TestPasswordHasher_ConcurrencyLimit`.
3. **Malformed or hostile JSON** (duplicate keys, unknown fields, huge or deeply nested bodies — TM-PARSE-003): rejected with a typed error, never partially applied. Test: Task 1 table tests.
4. **Configuration that weakens security silently** (HTTP issuer outside development, weak Argon2 parameters, short cursor key): absolute prohibitions fail startup; weaker-than-default settings produce a logged alert. Test: Task 2 `TestLoad_ProhibitionsFailStartup`, `TestLoad_WeakSettingsProduceAlerts`.
5. **Forged or replayed pagination cursors** (TM-SEC-004): a cursor from another tenant, another list, or with a flipped byte is rejected. Test: Task 13 `TestCursor_RejectsTamperedOrForeign`.

---

### Task 1: Strict JSON decoding (`shared/jsonx`)

**Files:** Create `shared/jsonx/decode.go`, `shared/jsonx/decode_test.go`, `shared/jsonx/fuzz_test.go`

**Interfaces:**
- Produces: `func DecodeStrict(r io.Reader, v any, lim Limits) error`; `type Limits struct{ MaxBytes int64; MaxDepth int }`; `var DefaultLimits = Limits{MaxBytes: 1 << 20, MaxDepth: 32}`; sentinel errors `ErrTooLarge`, `ErrTooDeep`, `ErrDuplicateKey`, `ErrUnknownField`, `ErrTrailingData`.

- [ ] **Step 1: Failing table test `TestDecodeStrict`**: duplicate key at top level and nested → `ErrDuplicateKey`; unknown field → `ErrUnknownField`; body `MaxBytes+1` → `ErrTooLarge`; nesting 33 → `ErrTooDeep`; two JSON values → `ErrTrailingData`; valid input → decoded and `nil`. Plus `FuzzDecodeStrict` (never panics; on error `v` is left zero-valued — decode into a temporary and copy only on success).
- [ ] **Step 2–4:** Implement with a first pass over `json.Decoder.Token()` (depth and duplicate tracking per object), then `json.Decoder` with `DisallowUnknownFields` on the buffered bytes. Run → PASS.
- [ ] **Step 5: Commit** `feat(jsonx): add strict, bounded JSON decoding`.

### Task 2 (SHARED-001 + SHARED-010): Contract C1 — typed configuration, prohibitions and alerts

**Files:**
- Create: `shared/config/config.go`, `shared/config/load.go`, `shared/config/rules.go`, `shared/config/doc.go`, `shared/config/config_test.go`, `shared/config/cmd/configdoc/main.go`
- Create: `docs/config-reference.md` (generated)
- Modify: `cmd/iam/main.go` (use `config.Load`), delete `cmd/iam/env.go` and `cmd/iam/env_test.go` (move `FuzzLoadEnv` to `FuzzLoad` here), `Makefile` (`config-doc`, `config-doc-check` in `check`), `docs/roadmap.md`

**Interfaces:**
- Produces:

```go
package config

type Environment string // "development" | "lab" | "production"

type Config struct {
    Environment     Environment   `env:"IAM_ENVIRONMENT" default:"production" doc:"..."`
    HTTPAddr        string        `env:"IAM_HTTP_ADDR" default:":8080"`
    PublicBaseURL   string        `env:"IAM_PUBLIC_BASE_URL"`                 // required; issuer base, ADR 0007
    DatabaseURL     string        `env:"IAM_DATABASE_URL" secret:"true"`
    DatabaseAppRole string        `env:"IAM_DATABASE_APP_ROLE" default:"iam_app"`
    ShutdownTimeout time.Duration `env:"IAM_SHUTDOWN_TIMEOUT" default:"10s"`
    CursorKey       string        `env:"IAM_CURSOR_KEY" secret:"true"`        // base64, >= 32 bytes; generated if empty in development only
    Argon2          Argon2Settings // IAM_ARGON2_MEMORY_KIB=19456, IAM_ARGON2_ITERATIONS=2, IAM_ARGON2_PARALLELISM=1, IAM_ARGON2_MAX_CONCURRENT=2
}

type Finding struct{ Severity Severity; Code, Message, Reference string } // Severity: Info | Warning | Critical | Prohibited

func Load(getenv func(string) string, readFile func(string) ([]byte, error)) (Config, []Finding, error)
func Document(w io.Writer) error // Markdown table from struct tags
```

- Precedence: struct `default` < JSON file at `IAM_CONFIG_FILE` (decoded with `jsonx.DecodeStrict`, Task 1) < environment. `Load` returns an error listing every invalid field (not only the first) and every `Prohibited` finding; `Critical`/`Warning` findings are returned for logging.
- Prohibitions (ADR 0012) at this layer: `PublicBaseURL` not `https://` outside `development` (11); empty `CursorKey` outside `development` (6, no defaults); `Environment` unknown. Alerts: Argon2 memory < 19456 KiB or iterations < 2 (Critical, reference OWASP Password Storage Cheat Sheet); `development` environment (Warning).
- `secret:"true"` fields are never printed by `Document` defaults nor included in error text.

- [ ] **Step 1: Write failing tests** — `TestLoad_Defaults`, `TestLoad_PrecedenceFileThenEnv`, `TestLoad_ReportsAllInvalidFields` (two bad fields → both named in the error), `TestLoad_ProhibitionsFailStartup` (http issuer in `lab` → error containing `ADR 0012`), `TestLoad_WeakSettingsProduceAlerts` (memory 8192 → one Critical finding, load succeeds), `TestLoad_SecretsNeverInErrors` (invalid `IAM_DATABASE_URL` value `postgres://u:hunter2@h` → error text lacks `hunter2`), `TestDocument_ListsEveryField`, `FuzzLoad`.
- [ ] **Step 2: Run** `go test ./shared/config/...` → FAIL (undefined).
- [ ] **Step 3: Implement** with reflection over struct tags (one place reads tags for load, validation and docs).
- [ ] **Step 4: Wire `cmd/iam`**: log findings with `slog` (severity as level), exit 1 on error. Generate `docs/config-reference.md` with `go run ./shared/config/cmd/configdoc`; `make config-doc-check` regenerates and fails on diff (added to `make check`).
- [ ] **Step 5: Run** `make check` → PASS. **Commit** `feat(config): add typed configuration with prohibitions and alerts`.

### Task 3 (SHARED-002 part 1 + SHARED-011 part 1): Contract C2 — randomness, MAC, constant-time compare, passwords

**Files:** Create `shared/crypto/random.go`, `shared/crypto/mac.go`, `shared/crypto/compare.go`, `shared/crypto/password.go`, tests alongside; `shared/clock/clock.go` (+ test)

**Interfaces:**
- Produces:

```go
package clock
type Clock interface{ Now() time.Time }
type System struct{}            // the only time.Now caller outside adapters/tests
type Fixed struct{ T time.Time } // tests; Advance(d)

package crypto
func RandomBytes(n int) ([]byte, error)
func RandomToken(nBytes int) (string, error) // base64url, no padding
func ConstantTimeEqual(a, b []byte) bool
type MAC struct{ /* key */ }
func NewMAC(key []byte) (*MAC, error)        // key >= 32 bytes
func (m *MAC) Sum(domain string, msg []byte) []byte   // HMAC-SHA-256 over len-prefixed domain || msg
func (m *MAC) Verify(domain string, msg, tag []byte) bool
type Argon2Params struct{ MemoryKiB uint32; Iterations uint32; Parallelism uint8 }
type PasswordHasher struct{ /* params, semaphore */ }
func NewPasswordHasher(p Argon2Params, maxConcurrent int) *PasswordHasher
func (h *PasswordHasher) Hash(ctx context.Context, password []byte) (string, error)      // PHC: $argon2id$v=19$m=..,t=..,p=..$salt$hash
func (h *PasswordHasher) Verify(ctx context.Context, password []byte, phc string) (ok, needsRehash bool, err error)
const MaxPasswordBytes = 1024
```

Salt 16 bytes, key 32 bytes. `Hash`/`Verify` acquire the semaphore honoring `ctx` (returns `ctx.Err()` on timeout); password over `MaxPasswordBytes` → `ErrPasswordTooLong` before hashing; no truncation ever.

- [ ] **Step 1: Failing tests** — `TestPasswordHasher_RoundTrip`, `TestPasswordHasher_WrongPassword`, `TestPasswordHasher_NeedsRehashWhenParamsChange`, `TestPasswordHasher_RejectsMalformedPHC` (table, including `argon2i`, missing fields, huge `m=`), `TestPasswordHasher_NoTruncation` (passwords differing only after byte 72 do not verify — bcrypt lesson), `TestPasswordHasher_TooLong`, `TestPasswordHasher_ConcurrencyLimit` (maxConcurrent 1, second call with 50 ms context → `context.DeadlineExceeded`), `TestMAC_DomainSeparation` (same msg, different domains → different tags), `TestMAC_RejectsShortKey`, `TestRandomToken_LengthAndAlphabet`, `TestConstantTimeEqual`.
- [ ] **Step 2–4:** Implement; `Verify` rejects PHC parameters above 4× the configured ones (DoS through stored hashes). Run `go test -race ./shared/crypto/... ./shared/clock/...` → PASS.
- [ ] **Step 5: Commit** `feat(crypto): add randomness, MAC and Argon2id password hashing`.

### Task 4 (SHARED-002 part 2 + SHARED-011 part 2): Contract C2 — JOSE signing, verification and key sources

**Files:** Create `shared/crypto/jose.go`, `shared/crypto/keyset.go`, `shared/crypto/jose_test.go`; `adapters/keys/localfile/source.go` (+ test)

**Interfaces:**
- Produces:

```go
package crypto
type Algorithm string // only "ES256" in milestone 1
type SigningKey struct{ ID string; Alg Algorithm; Private crypto.Signer } // ID = RFC 7638 thumbprint, base64url
type KeySource interface{ SigningKeys(ctx context.Context) ([]SigningKey, error) } // first = active
type KeySet struct{ /* active + verification keys */ }
func NewKeySet(ctx context.Context, src KeySource) (*KeySet, error)
func (k *KeySet) Sign(claims any) (string, error)        // compact JWS, header {alg, kid, typ}
func (k *KeySet) PublicJWKS() ([]byte, error)            // JSON for /jwks; never private material
type Verifier struct{ /* allowed algs, keys by kid, max size */ }
func NewVerifier(keys []PublicKey, allowed []Algorithm) (*Verifier, error)
func (v *Verifier) Verify(token string, dst any) error   // checks size <= 8 KiB, alg in allowlist, kid known, signature
func GenerateES256() (SigningKey, error)
```

`adapters/keys/localfile`: `func New(path string) crypto.KeySource` reading PEM PKCS#8 EC P-256 keys (development and tests only; the lab source arrives in 1D).

- [ ] **Step 1: Failing tests** — `TestSignVerify_RoundTrip`, `TestVerify_RejectsAlgNone`, `TestVerify_RejectsHS256WithPublicKeyAsSecret` (algorithm confusion), `TestVerify_RejectsUnknownKid`, `TestVerify_RejectsOversizeToken` (8 KiB + 1), `TestVerify_RejectsTamperedPayload`, `TestPublicJWKS_HasNoPrivateMaterial` (no `"d"` member), `TestKid_IsRFC7638Thumbprint` (fixed key → known thumbprint), `TestLocalFile_RejectsNonP256`.
- [ ] **Step 2–4:** Implement with `go-jose/v4` (`jose.ParseSigned(token, []jose.SignatureAlgorithm{jose.ES256})`); size check before parsing. Run → PASS. PR description justifies the `go-jose` dependency (ADR 0010).
- [ ] **Step 5: Commit** `feat(crypto): add ES256 JOSE signing, verification and local key source`.

### Task 5: Identifiers (`shared/ids`)

**Files:** Create `shared/ids/ids.go`, `shared/ids/ids_test.go`

**Interfaces:**
- Produces: `type ID [16]byte` (UUIDv7) with `String()`, `Parse(string) (ID, error)`, `MarshalText`/`UnmarshalText`, `IsZero()`; distinct types `TenantID`, `OrganizationID`, `UserID`, `ApplicationID`, `EventID` (each `type X struct{ ID }` so they cannot be mixed); `type Generator struct{ Clock clock.Clock }` with `func (g Generator) New() (ID, error)` (48-bit ms timestamp from the clock + 74 random bits from `crypto.RandomBytes`).

- [ ] **Step 1: Failing tests** — `TestNew_IsVersion7AndVariant`, `TestNew_SortsByTime` (fixed clock advancing), `TestParse_RoundTrip`, `TestParse_RejectsMalformed` (table), `TestTypedIDs_DoNotConvert` (compile-time: a `TenantID` passed where `UserID` is expected must not compile — express as a `// does not compile` example in the package doc plus a test that `TenantID{}` and `UserID{}` are different `reflect.Type`s).
- [ ] **Step 2–5:** Implement, run, commit `feat(ids): add typed UUIDv7 identifiers`.

### Task 6 (SHARED-005): Contract C5 — state-machine library (own module)

**Files:** Create `shared/statemachine/go.mod`, `machine.go`, `mermaid.go`, `machine_test.go`, `statemachinetest/exhaustive.go`; modify `go.work`, `Makefile` (module list is discovered; confirm `workspace-check` passes)

**Interfaces:**
- Produces:

```go
package statemachine
type Transition[S, E comparable, C any] struct{ From S; Event E; To S; Guard func(C) error }
type Definition[S, E comparable, C any] struct{ /* name, transitions */ }
func New[S, E comparable, C any](name string, ts ...Transition[S, E, C]) (*Definition[S, E, C], error) // rejects duplicate (From, Event)
func (d *Definition[S, E, C]) Fire(from S, ev E, c C) (Fired[S, E], error)
type Fired[S, E comparable] struct{ From S; Event E; To S }
var ErrUndeclaredTransition = errors.New("statemachine: undeclared transition")
func (d *Definition[S, E, C]) Mermaid() string // stateDiagram-v2, deterministic order

package statemachinetest
func Exhaustive[S, E comparable, C any](t *testing.T, d *statemachine.Definition[S, E, C], states []S, events []E, ctx C)
// fires every (state, event): declared ones must succeed when guards pass, all others must return ErrUndeclaredTransition
```

Versioning (optimistic concurrency) is the storage layer's job: callers persist `Fired.To` with a version check.

- [ ] **Step 1: Failing tests** — `TestFire_DeclaredTransition`, `TestFire_UndeclaredIsRefused`, `TestFire_GuardErrorIsReturnedAndStateUnchanged`, `TestNew_RejectsDuplicateTransition`, `TestMermaid_IsDeterministic` (golden string), `TestExhaustive_CatchesExtraTransition` (a definition with a sneaky extra transition fails the helper when the expected table omits it).
- [ ] **Step 2–5:** Implement (standard library only), run `cd shared/statemachine && go test ./...` and `make check`, commit `feat(statemachine): add explicit state-machine library`.

### Task 7 (SHARED-013 part 1): Tenancy domain model

**Files:** Create `shared/tenancy/tenant.go`, `organization.go`, `user.go`, `membership.go`, `application.go`, `validate.go`, `tenancy_test.go`

**Interfaces:**
- Consumes: `ids` (Task 5), `clock` (Task 3). Client secrets are 32 random bytes, stored as SHA-256: this task adds `func HashSecret(secret []byte) []byte` and `func VerifySecret(secret, hash []byte) bool` (constant time) to `shared/crypto`, with tests `TestHashSecret_*`.
- Produces:

```go
package tenancy
type Tenant struct{ ID ids.TenantID; Slug string; Name string; Status Status; CreatedAt time.Time }
type Organization struct{ ID ids.OrganizationID; TenantID ids.TenantID; Slug, Name string; CreatedAt time.Time }
type User struct{ ID ids.UserID; TenantID ids.TenantID; Email string; EmailVerified bool; Status Status; CreatedAt time.Time }
type Membership struct{ TenantID ids.TenantID; OrganizationID ids.OrganizationID; UserID ids.UserID; CreatedAt time.Time }
type ClientType string // "public" | "confidential"
type Application struct{ ID ids.ApplicationID; TenantID ids.TenantID; Name string; Type ClientType; RedirectURIs []string; SecretHash []byte; CreatedAt time.Time }
type Status string // "active" | "suspended"
const SystemTenantSlug = "system"

func ValidateSlug(s string) error            // ^[a-z0-9][a-z0-9-]{1,62}$, not reserved ("system", "admin", "api", "well-known")
func NormalizeEmail(s string) (string, error) // trim, lowercase domain only, NFC (x/text), max 254, one '@'
func ValidateRedirectURI(raw string, allowInsecureHTTP bool) error // ADR 0012: no wildcard, no fragment, no userinfo; https only, except http://127.0.0.1 and http://[::1] on any port (RFC 8252); other http only when allowInsecureHTTP (development)
```

- [ ] **Step 1: Failing tests** — table tests for `ValidateSlug` (reserved, too long, uppercase, leading dash), `NormalizeEmail` (case in domain, NFC vs NFD, two `@`, 255 chars), `ValidateRedirectURI` (wildcard host, `*` in path, fragment, `http://example.com` with allowInsecureHTTP=false → error; loopback http any port → ok; `javascript:` → error; userinfo `https://a@b` → error), `TestHashSecret_ConstantTimeVerify`.
- [ ] **Step 2–5:** Implement, run, commit `feat(tenancy): add tenancy domain model and validation`.

### Task 8 (SHARED-003): Contract C3 — storage ports, in-memory adapter and contract suite

**Files:** Create `shared/storage/storage.go`, `shared/storage/errors.go`, `shared/storage/storagetest/suite.go`; `adapters/memory/store.go`, `adapters/memory/store_test.go`

**Interfaces:**
- Produces:

```go
package storage
type Store interface {
    InTenant(ctx context.Context, tenant ids.TenantID, fn func(Tx) error) error // one DB transaction; sets RLS tenant
    InSystem(ctx context.Context, fn func(SystemTx) error) error               // tenant registry; audit goes to the system tenant's chain
}
type SystemTx interface {
    CreateTenant(ctx context.Context, t tenancy.Tenant) error
    TenantBySlug(ctx context.Context, slug string) (tenancy.Tenant, error)
    Tenant(ctx context.Context, id ids.TenantID) (tenancy.Tenant, error)
    ListTenants(ctx context.Context, after ids.TenantID, limit int) ([]tenancy.Tenant, error)
    Audit() AuditAppender // appends to the system tenant's chain (Task 10)
}
type Tx interface {
    Organizations() OrganizationRepo; Users() UserRepo; Memberships() MembershipRepo; Applications() ApplicationRepo
    Audit() AuditAppender // Task 10
}
type OrganizationRepo interface{ Create(ctx, tenancy.Organization) error; Get(ctx, ids.OrganizationID) (tenancy.Organization, error); BySlug(ctx, string) (tenancy.Organization, error); List(ctx, after ids.OrganizationID, limit int) ([]tenancy.Organization, error) }
type UserRepo interface{ Create(ctx, tenancy.User) error; Get(ctx, ids.UserID) (tenancy.User, error); ByEmail(ctx, string) (tenancy.User, error); List(ctx, after ids.UserID, limit int) ([]tenancy.User, error) }
type MembershipRepo interface{ Add(ctx, tenancy.Membership) error; Remove(ctx, ids.OrganizationID, ids.UserID) error; ListByUser(ctx, ids.UserID) ([]tenancy.Membership, error) }
type ApplicationRepo interface{ Create(ctx, tenancy.Application) error; Get(ctx, ids.ApplicationID) (tenancy.Application, error); List(ctx, after ids.ApplicationID, limit int) ([]tenancy.Application, error) }
var ErrNotFound, ErrConflict, ErrInvalidTenant error // ErrInvalidTenant: zero tenant ID or unknown tenant
const MaxListLimit = 200
```

The two `Audit()` methods are added in Task 10 (Task 8 defines the interfaces without them). The tenant ID is implicit in a `Tx`: repo methods never take one and never return another tenant's rows; a row created through a `Tx` always gets the `Tx`'s tenant regardless of the struct field (the field is overwritten, test enforces it).

- `storagetest.Run(t *testing.T, newStore func(t *testing.T) storage.Store)` runs: CRUD and uniqueness (`ErrConflict` on duplicate org slug / user email / tenant slug), `ErrNotFound`, list ordering and limit clamping to `MaxListLimit`, **`CrossTenantIsolation`** (create in A; every Get/ByX/List in B returns `ErrNotFound`/empty; a `Create` in B with `TenantID = A` stores under B), **`TxRollsBackOnError`**, `InTenantRejectsZeroOrUnknownTenant`.

- [ ] **Step 1: Write the suite and run it against `adapters/memory`** → FAIL (no adapter).
- [ ] **Step 2: Implement `adapters/memory`** (copy-on-write snapshot per transaction for rollback). Run → PASS.
- [ ] **Step 3: Commit** `feat(storage): add storage ports, contract suite and in-memory adapter`.

### Task 9 (SHARED-012): Postgres adapter, migrations, RLS and connection budget

**Files:** Create `adapters/postgres/migrate.go`, `adapters/postgres/migrations/20261107000000_tenancy.sql`, `adapters/postgres/store.go`, `adapters/postgres/repos_*.go`, `adapters/postgres/store_test.go`, `adapters/postgres/rls_test.go`; `internal/database/budget.go` + test; modify `cmd/iam/main.go` (subcommand `migrate`), `internal/testdb` (optional app role per test DB)

**Interfaces:**
- Produces: `func Migrate(ctx context.Context, pool *pgxpool.Pool, appRole string) error` (embedded `migrations/*.sql`, applied in name order inside one transaction each, guarded by `pg_advisory_xact_lock(7234717)`, recorded in `iam.schema_migrations(version text primary key, applied_at timestamptz)`; refuses a role name not matching `^[a-z_][a-z0-9_]{0,62}$`); `func NewStore(pool *pgxpool.Pool) storage.Store`; `cmd/iam migrate` runs `Migrate` with `IAM_DATABASE_URL` and `IAM_DATABASE_APP_ROLE` and exits.
- Schema `iam`; tables `tenants`, `organizations`, `users`, `memberships`, `applications`; every tenant-owned table has `tenant_id uuid not null references iam.tenants(id)`, composite uniques `(tenant_id, slug)`, `(tenant_id, email)`, RLS enabled with policy `USING (tenant_id = current_setting('iam.tenant_id', true)::uuid) WITH CHECK (same)`; grants `USAGE` on schema and `SELECT, INSERT, UPDATE, DELETE` on tables to the app role; the app role owns nothing.
- `InTenant` runs `SELECT set_config('iam.tenant_id', $1, true)` (transaction-local) before `fn`; `InSystem` sets the same variable to the system tenant's ID (looked up by `SystemTenantSlug`) so its audit appends obey RLS; `iam.tenants` has no RLS and is reachable only through `SystemTx`.
- `internal/database/budget.go`: constants `ServicePool=2, MaxInstances=1, RolloutRevisions=2, MigrationJob=1, Person=1, Ceiling=6`; `TestConnectionBudget` asserts `ServicePool*MaxInstances*RolloutRevisions+MigrationJob+Person <= Ceiling` and `ServicePool == MaxConns`.

- [ ] **Step 1: Failing tests** — `TestPostgresStore_Contract` (runs `storagetest.Run` with a fresh test DB, migrated, store connected **as a per-test non-owner app role**), `TestMigrate_IsIdempotent`, `TestMigrate_ConcurrentRunsSerialize` (two goroutines, both succeed, versions applied once), `TestRLS_BlocksQueryWithoutTenantFilter` (as app role inside `InTenant(A)`, raw `SELECT count(*) FROM iam.users` sees only A's rows), `TestRLS_SettingDoesNotLeakAcrossTransactions` (after `InTenant(A)` commits, a raw query on a pooled connection without the setting sees zero rows), `TestMigrate_RejectsBadRoleName`, `TestConnectionBudget`.
- [ ] **Step 2–4:** Implement; `internal/testdb` gains `NewWithAppRole(t) (adminURL, appURL string)` creating role `iam_test_app_<pid>_<n>` (dropped in cleanup). Run with `IAM_REQUIRE_DB=1` → PASS.
- [ ] **Step 5: Commit** `feat(postgres): add migrations, RLS and tenant-scoped store`.

### Task 10 (SHARED-004 + SHARED-014): Contract C4 — audit events and hash chain

**Files:** Create `shared/audit/event.go`, `shared/audit/chain.go`, `shared/audit/audit_test.go`, `shared/canonjson/canonjson.go` (+ test); modify `shared/storage` (`AuditAppender`), `adapters/memory`, `adapters/postgres` (+ migration `20261107000100_audit.sql`), `storagetest` (audit cases)

**Interfaces:**
- Produces:

```go
package audit
const SchemaVersion = 1
type Actor struct{ Type string /* "user"|"operator"|"system"|"application" */; ID string }
type Event struct {
    ID ids.EventID; TenantID ids.TenantID; Sequence int64; Time time.Time; SchemaVersion int
    Action string        // dotted, e.g. "tenant.created", "organization.created"
    Actor Actor; Target string // resource name, e.g. "tenants/t1/organizations/o1"
    Outcome string       // "success"|"failure"|"denied"
    Data map[string]string // never raw personal data or secrets
    PrevHash, Hash []byte
}
func Seal(prev []byte, e Event) (Event, error) // Hash = SHA-256(prev || canonjson(e without Hash)); via shared/crypto
func VerifyChain(events []Event) error         // checks sequence continuity, PrevHash links and each Hash

package storage
type AuditAppender interface{ Append(ctx context.Context, e audit.Event) (audit.Event, error) } // assigns Sequence, PrevHash, Hash inside the Tx

package canonjson
func Marshal(v any) ([]byte, error) // RFC 8785 for strings, integers, booleans, null, arrays, objects with string keys; rejects floats
```

Per-tenant chain; Postgres table `iam.audit_events` with `UNIQUE (tenant_id, sequence)`, RLS like the others, app role granted `SELECT, INSERT` only (no UPDATE/DELETE); `Append` reads the last row `FOR UPDATE` within the tenant.

- [ ] **Step 1: Failing tests** — `TestCanonJSON_RFC8785Vectors` (key ordering incl. non-ASCII, escaping), `TestCanonJSON_RejectsFloat`, `TestSeal_Deterministic`, `TestVerifyChain_DetectsTampering` (modify Data, reorder, drop one, change PrevHash → each detected), contract cases `AuditAppendChainsPerTenant`, `AuditAppendIsAtomicWithTx` (rollback drops the event), Postgres `TestAudit_AppRoleCannotUpdateOrDelete`.
- [ ] **Step 2–5:** Implement, run, commit `feat(audit): add hash-chained, append-only audit events`.

### Task 11 (SHARED-013 part 2): Tenancy service

**Files:** Create `shared/tenancy/service.go`, `shared/tenancy/service_test.go`

**Interfaces:**
- Consumes: `storage.Store`, `ids.Generator`, `clock.Clock`, `crypto.RandomToken`/`HashSecret`, `audit.Event`.
- Produces: `type Service struct{ Store storage.Store; IDs ids.Generator; Clock clock.Clock; Env config.Environment }` with `CreateTenant(ctx, actor audit.Actor, slug, name string) (Tenant, error)`, `CreateOrganization(ctx, actor, tenant ids.TenantID, slug, name string)`, `CreateUser(ctx, actor, tenant, email string)`, `AddMembership(ctx, actor, tenant, org ids.OrganizationID, user ids.UserID)`, `CreateApplication(ctx, actor, tenant, name string, t ClientType, redirectURIs []string) (Application, string /* client secret, shown once; "" for public */, error)`, `EnsureSystemTenant(ctx) (Tenant, error)` (idempotent). Every mutation appends one audit event in the same transaction (`tenant.created` through `SystemTx.Audit()`, in the system tenant's chain).

- [ ] **Step 1: Failing tests** (against `adapters/memory`): validation errors surface (`ValidateSlug`, `ValidateRedirectURI`), reserved slug rejected, `CreateApplication` confidential returns a 43-char secret and stores only its hash, public returns `""`, every mutation produces exactly one audit event with the right action, `EnsureSystemTenant` twice → one tenant.
- [ ] **Step 2–5:** Implement, run, commit `feat(tenancy): add tenancy service with audit trail`.

### Task 12 (SHARED-006): Contract C6 — admin API in Protobuf

**Files:** Create `api/iam/admin/v1/tenants.proto`, `organizations.proto`, `users.proto`, `applications.proto`; regenerate `api/gen/go`; keep `server_info.proto`.

**Interfaces:**
- Produces (AIP-style, `buf lint` STANDARD): services `TenantService` (`CreateTenant`, `GetTenant`, `ListTenants`), `OrganizationService` (`CreateOrganization`, `GetOrganization`, `ListOrganizations`), `UserService` (`CreateUser`, `GetUser`, `ListUsers`), `ApplicationService` (`CreateApplication` returning `client_secret` once, `GetApplication`, `ListApplications`); resource names `tenants/{tenant}`, `tenants/{tenant}/organizations/{organization}`, `tenants/{tenant}/users/{user}`, `tenants/{tenant}/applications/{application}`, where `{tenant}` is the tenant slug (resolved through `SystemTx.TenantBySlug`) and the others are UUIDs; list requests `page_size` (default 50, max 200) and `page_token`; no field ever carries a password or secret hash.

- [ ] **Step 1:** Write the protos; `make proto` then `make proto-check` → PASS; `buf breaking` passes against main (additive only).
- [ ] **Step 2: Commit** `feat(api): define admin API resources for tenancy`.

### Task 13 (SHARED-015): Admin API server with fail-closed authorization and authenticated cursors

**Files:** Create `internal/adminapi/server.go`, `authorize.go`, `cursor.go`, `errors.go`, `*_test.go`; modify `internal/httpserver` (mount a `http.Handler` under a prefix via a new option `WithHandler(pattern string, h http.Handler)`), `cmd/iam/main.go`

**Interfaces:**
- Consumes: Task 12 generated handlers, `tenancy.Service`, `crypto.MAC`, `config.Config.CursorKey`.
- Produces: `type Authorizer interface{ Authorize(ctx context.Context, req Request) error }` with `type Request struct{ Procedure string; Resource string }`; `type DenyAll struct{}` (default, returns `connect.CodePermissionDenied`); `func NewHandler(svc *tenancy.Service, authz Authorizer, cursors *Cursors) (pattern string, h http.Handler)` mounting all four services under `/admin/`; `Cursors` encodes `{list, tenant, after}` as base64url payload + MAC (domain `"admin-cursor/v1"`), and rejects tokens over 512 bytes.
- Error mapping: `storage.ErrNotFound` → `NotFound`, `ErrConflict` → `AlreadyExists`, validation → `InvalidArgument` with a stable code string; internal errors → `Internal` without detail. Responses carry `Cache-Control: no-store`.

- [ ] **Step 1: Failing tests** — `TestAdmin_DeniesByDefault` (every procedure with `DenyAll` → `PermissionDenied`, nothing written), `TestAdmin_CreateAndGetTenant` (allow-all test authorizer), `TestAdmin_ListPaginates`, `TestCursor_RejectsTamperedOrForeign` (flipped byte, cursor from another list, from another tenant → `InvalidArgument`), `TestAdmin_ClientSecretReturnedOnlyOnCreate`, `TestAdmin_InternalErrorsHaveNoDetail`, `TestAdmin_ResponsesAreNoStore`.
- [ ] **Step 2–4:** Implement with `connect` interceptors (authorizer runs before any handler). `cmd/iam` mounts the admin API with `DenyAll` until plan 1D wires the AuthZ-backed authorizer; startup logs a Warning finding "admin API denies all requests until the authorizer is configured". Run `make check` and DB-backed tests → PASS.
- [ ] **Step 5: Commit** `feat(admin): serve the admin API behind a fail-closed authorizer`.

---

## Milestone 1A exit check

- [ ] `docs/roadmap.md`: SHARED-001..006 and SHARED-010..015 ✅ (each in its own PR).
- [ ] `iam migrate` followed by `iam` starts against the e2e database in CI (extend `e2e/health_test.go`: run `migrate` first, then `/readyz` 200).
- [ ] `docs/architecture.md` section 9 (tables) and `docs/threat-model.md` items TM-PARSE-003, TM-PWD-003, TM-SEC-004 updated from "pendente" to the tests above.
