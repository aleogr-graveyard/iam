# Marco 0 ("fábrica") Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make automated verification trustworthy before any product code: module skeleton with an empty server, lint with architectural boundaries, the full PR CI, contract tooling, a real-Postgres test harness, a local environment, a scanned image with SBOM, community and CRA files, an end-to-end harness with test client A, required checks with an auto-merge guard, and nightly fuzz and benchmark jobs.

**Architecture:** One Go module for the server (`github.com/aleogr/iam`), plus separate modules for the test client (`examples/a`), the end-to-end tests (`e2e`), the pinned developer tools (`tools`) and the lint fixture (`tools/lintfixture`, outside `go.work`). Every tool runs through `go tool -modfile=tools/go.mod <name>`, so versions and checksums live in `tools/go.sum` and no third-party GitHub Action is needed beyond the allowlist. A `Makefile` is the single entry point used both locally and in CI.

**Tech Stack:** Go 1.27, `pgx/v5`, `protobuf` + `connect` (generated code only), `buf`, `golangci-lint` v2, `govulncheck`, `gitleaks`, `actionlint`, `go-licenses`, `ko`, `cyclonedx-gomod`, `osv-scanner`, `benchstat`, GitHub Actions, Docker Compose, PostgreSQL 16.

**Spec:** `docs/superpowers/specs/2026-11-06-iam-design.md` (sections 4, 9, 10, 11, 12) and `docs/roadmap.md` (OPS-001 to OPS-011). Read `CLAUDE.md` first.

## Global Constraints

- Go `1.27` in every `go.mod` (`go 1.27`); toolchain on runners via `actions/setup-go` with `go-version-file: go.mod`.
- Module path `github.com/aleogr/iam` (vanity import path deferred until the brand is decided; one rename then).
- Runtime dependencies allowed in this milestone: standard library, `golang.org/x/...`, `github.com/jackc/pgx/v5`, `google.golang.org/protobuf`, `connectrpc.com/connect`. Anything else goes in `tools/go.mod` or needs a written justification.
- Tool versions (pin exactly in `tools/go.mod` via `go get -tool`): `golangci-lint/v2 v2.15.0`, `golang.org/x/vuln v1.9.0` (govulncheck), `gitleaks/v8 v8.30.1`, `actionlint v1.7.12`, `go-licenses/v2 v2.0.1`, `bufbuild/buf v1.75.0`, `protobuf v1.36.12` (protoc-gen-go), `connectrpc.com/connect v1.21.1` (protoc-gen-connect-go), `google/ko v0.19.1`, `go-containerregistry` (crane, same version ko pulls), `cyclonedx-gomod v1.12.0`, `osv-scanner/v2 v2.7.0`, `golang.org/x/perf` (benchstat, latest pseudo-version).
- `tools/` is not in `go.work`, and `-modfile` is refused in workspace mode: always invoke tools as `GOWORK=off go tool -modfile=tools/go.mod <tool>` (Makefile variable `TOOL`), and run `go test` inside `tools/` with `GOWORK=off`.
- GitHub Actions pinned by full SHA: `actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1`, `actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0`, `actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9 # v7.0.2`. Every workflow starts with `permissions: {}`; each job asks only for what it uses; `actions/checkout` always with `persist-credentials: false`.
- Container images pinned by digest (resolve with `go tool -modfile=tools/go.mod crane digest <image>`): `postgres:16` for CI services and Compose; `cgr.dev/chainguard/static:latest` as the image base.
- Code, identifiers, comments, commits and PR titles in English (Conventional Commits); PR descriptions and docs in Portuguese.
- No secret, password or credential in any versioned file, including dev defaults.
- Each task is one PR, one roadmap item, marked ✅ in `docs/roadmap.md` in the same PR. All tasks touch CI, pipeline or `.github`, so the owner merges every one (CLAUDE.md, sensitive areas).

## Review Focus

1. **Readiness with the database down or hanging** — `/readyz` must answer 503 quickly, never hang or report ok. Test in Task 1 (`TestReadyz_CheckTimeout`) and Task 5 (`TestReadyz_ClosedPoolIsUnavailable`).
2. **SIGTERM during an in-flight request** (Cloud Run sends it on every rollout) — the request completes, then the process exits 0. Test in Task 1 (`TestRun_GracefulShutdown`).
3. **Error details leaking from readiness** — a failing check's error text (which may hold a DSN) must never reach the response body. Test in Task 1 (`TestReadyz_FailingCheckReturns503WithNameOnly`).
4. **A boundary rule that silently stops matching** (renamed path, config typo) — lint keeps passing while the architecture rots. Pinned by the fixture in Task 2 (`scripts/check-lint-rules.sh` must see every expected violation).
5. **Test databases leaking between tests or surviving the run** — parallel tests must not share state, and the server must never exceed its pool. Test in Task 5 (`TestNew_IsolatedDatabases`, `TestOpen_MaxConnsIsTwo`).

---

### Task 1 (OPS-001): Module skeleton and empty server with health checks

**Files:**
- Create: `go.mod`, `go.work`, `Makefile`
- Create: `cmd/iam/main.go`, `cmd/iam/env.go`, `cmd/iam/env_test.go`
- Create: `internal/httpserver/server.go`, `internal/httpserver/health.go`, `internal/httpserver/server_test.go`, `internal/httpserver/health_test.go`
- Modify: `docs/roadmap.md` (OPS-001 ✅)

**Interfaces:**
- Produces:
  - `package httpserver`: `type ReadinessCheck struct { Name string; Check func(ctx context.Context) error }`
  - `func New(addr string, opts ...Option) *Server`; options `WithReadinessChecks(checks ...ReadinessCheck) Option`, `WithReadinessTimeout(d time.Duration) Option` (default `2 * time.Second`)
  - `func (s *Server) Handler() http.Handler`
  - `func (s *Server) Run(ctx context.Context, shutdownTimeout time.Duration) error` — serves until `ctx` is done, then `http.Server.Shutdown` with the timeout; returns nil on clean shutdown.
  - `package main`: `type envConfig struct { HTTPAddr string; DatabaseURL string; ShutdownTimeout time.Duration }`, `func loadEnv(getenv func(string) string) (envConfig, error)`; variables `IAM_HTTP_ADDR` (default `:8080`), `IAM_DATABASE_URL` (optional), `IAM_SHUTDOWN_TIMEOUT` (Go duration, default `10s`, must be > 0). This is a stopgap until contract C1 (SHARED-001).
  - `var version = "dev"` in `cmd/iam`, overridden by `-ldflags "-X main.version=..."`.
- HTTP contract: `GET /healthz` → 200 `{"status":"ok"}`. `GET /readyz` → 200 `{"status":"ok"}` when every check passes within the timeout, else 503 `{"status":"unavailable","failed":["<check name>",...]}` (names only, sorted). Other methods on these paths → 405 with `Allow: GET`. Every response: `Content-Type: application/json`, `Cache-Control: no-store`. Checks run concurrently.

- [ ] **Step 1: Create the module and workspace**

`go mod init github.com/aleogr/iam`, set `go 1.27`; `go work init .`. `Makefile` targets for now: `test` (`go test -race ./...`), `build` (`go build -trimpath -o dist/iam ./cmd/iam`).

- [ ] **Step 2: Write the failing tests** in `internal/httpserver/health_test.go` and `server_test.go`

```go
func TestHealthz_OK(t *testing.T)                                   // 200, body {"status":"ok"}, Cache-Control: no-store
func TestReadyz_AllChecksPass(t *testing.T)                         // two passing checks → 200
func TestReadyz_FailingCheckReturns503WithNameOnly(t *testing.T)    // check returns errors.New("dial postgres://u:secret@h") → 503, body {"status":"unavailable","failed":["database"]}, body does not contain "secret"
func TestReadyz_CheckTimeout(t *testing.T)                          // WithReadinessTimeout(50*time.Millisecond), check blocks on ctx → 503 in < 1s, failed ["slow"]
func TestHealthEndpoints_MethodNotAllowed(t *testing.T)             // POST /healthz → 405, Allow: GET
func TestRun_GracefulShutdown(t *testing.T)                         // handler sleeping 200ms in flight; cancel ctx at 50ms; request gets 200; Run returns nil
func BenchmarkHealthz(b *testing.B)                                 // feeds the nightly benchmark gate (Task 11)
```

Use `httptest.NewRecorder` for handler tests; `TestRun_GracefulShutdown` binds `127.0.0.1:0` (expose the bound address via a test-only `New` on a pre-made `net.Listener`: add `func (s *Server) Serve(ctx context.Context, l net.Listener, shutdownTimeout time.Duration) error`, and make `Run` call it).

- [ ] **Step 3: Run them to verify they fail** — `go test ./internal/httpserver/...` → FAIL (undefined: New).

- [ ] **Step 4: Implement `internal/httpserver`** — `http.ServeMux` with method patterns (`GET /healthz`); readiness fans out checks with `context.WithTimeout`; JSON via `encoding/json`; `ReadHeaderTimeout: 5 * time.Second` on the `http.Server`.

- [ ] **Step 5: Run them to verify they pass** — `go test -race ./internal/httpserver/...` → PASS.

- [ ] **Step 6: Write the failing env tests** in `cmd/iam/env_test.go`

```go
func TestLoadEnv_Defaults(t *testing.T)                 // empty env → {":8080", "", 10s}
func TestLoadEnv_Overrides(t *testing.T)                // all three set → parsed values
func TestLoadEnv_InvalidShutdownTimeout(t *testing.T)   // "abc" and "0s" → error mentioning IAM_SHUTDOWN_TIMEOUT
func FuzzLoadEnv(f *testing.F)                          // arbitrary IAM_SHUTDOWN_TIMEOUT never panics; error or > 0
```

- [ ] **Step 7: Implement `loadEnv` and `main`** — `main` builds the server (no readiness checks yet), runs it with `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)`, logs JSON with `log/slog` (`version`, `addr`), exits 1 on error. Run `go test -race ./...` → PASS; `make build` produces `dist/iam`; `IAM_HTTP_ADDR=127.0.0.1:18080 ./dist/iam & curl -s 127.0.0.1:18080/healthz` → `{"status":"ok"}`.

- [ ] **Step 8: Commit** — `git add go.mod go.work Makefile cmd internal docs/roadmap.md && git commit -m "feat: add module skeleton and health-checked server"`

### Task 2 (OPS-002): Lint with architectural boundary rules, and a fixture that proves they bite

**Files:**
- Create: `tools/go.mod` (module `github.com/aleogr/iam/tools`, `tool` directives), `.golangci.yml`
- Create: `tools/lintfixture/go.mod` (module `github.com/aleogr/iam/lintfixture`, not in `go.work`), `tools/lintfixture/authn/core/violations.go`, `tools/lintfixture/internal/anywhere/violations.go`
- Create: `scripts/check-lint-rules.sh`
- Modify: `Makefile`, `docs/roadmap.md` (OPS-002 ✅)

**Interfaces:**
- Produces: `make lint` (runs `go tool -modfile=tools/go.mod golangci-lint run` in each module of `go.work`), `make lint-rules` (runs the script). Later tasks add tools with `go get -tool -modfile=tools/go.mod <module>@<version>`.

- [ ] **Step 1: Add golangci-lint as a tool** — `go get -tool -modfile=tools/go.mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0`.

- [ ] **Step 2: Write the fixture (the "failing test")** — files that each break exactly one rule:
  - `authn/core/violations.go`: imports `net/http` (rule `core-no-infra`) and calls `time.Now()` (forbidigo `time.Now` in core).
  - `internal/anywhere/violations.go`: imports `crypto/sha256` (rule `crypto-only-in-shared-crypto`) and `math/rand/v2` (rule `no-math-rand`).

- [ ] **Step 3: Write `.golangci.yml`** (version 2). Linters: `depguard`, `forbidigo`, `gosec`, `errcheck`, `govet`, `staticcheck`, `unused`, `ineffassign`, `cyclop` (`max-complexity: 10`), `dupl` (`threshold: 100`), `misspell`; formatters `gofmt`, `goimports` (`local-prefixes: github.com/aleogr/iam`). Depguard rules:

| Rule | Files | Deny |
|---|---|---|
| `core-no-infra` | `**/authn/core/**`, `**/authz/core/**` | `net/http`, `database/sql`, `github.com/jackc/pgx`, `cloud.google.com/go`, `connectrpc.com`, `github.com/aleogr/iam/adapters`, `github.com/aleogr/iam/authn/server`, `github.com/aleogr/iam/authz/server` |
| `crypto-only-in-shared-crypto` | `$all`, `!**/shared/crypto/**` | `crypto/aes`, `crypto/cipher`, `crypto/des`, `crypto/ecdh`, `crypto/ecdsa`, `crypto/ed25519`, `crypto/elliptic`, `crypto/hkdf`, `crypto/hmac`, `crypto/md5`, `crypto/mlkem`, `crypto/pbkdf2`, `crypto/rand`, `crypto/rc4`, `crypto/rsa`, `crypto/sha1`, `crypto/sha256`, `crypto/sha3`, `crypto/sha512`, `crypto/subtle`, `golang.org/x/crypto`, `github.com/go-jose/go-jose` |
| `no-math-rand` | `$all` | `math/rand`, `math/rand/v2` |

Forbidigo: `^time\.Now$` with message "inject a clock (CLAUDE.md)", applied only to cores through `linters.exclusions.rules: [{linters: [forbidigo], path-except: '(^|/)(authn|authz)/core/'}]`. Each deny message cites `CLAUDE.md`.

- [ ] **Step 4: Write `scripts/check-lint-rules.sh`** — runs golangci-lint with the root config inside `tools/lintfixture`, expects a non-zero exit, and fails unless the output contains all four markers: `core-no-infra`, `crypto-only-in-shared-crypto`, `no-math-rand`, `inject a clock`.

- [ ] **Step 5: Verify** — `make lint` → exit 0 on the real code (fix anything it reports in Task 1 code); `make lint-rules` → "all 4 boundary rules fired"; temporarily rename `core-no-infra` in the config → `make lint-rules` fails (then revert).

- [ ] **Step 6: Commit** — `git commit -m "build: add lint with architectural boundary rules"`

### Task 3 (OPS-003): Pull request CI

**Files:**
- Create: `.github/workflows/ci.yml`
- Modify: `tools/go.mod` (govulncheck, gitleaks, actionlint, go-licenses), `Makefile`, `docs/roadmap.md` (OPS-003 ✅)

**Interfaces:**
- Produces: `make vuln`, `make secrets` (`gitleaks git --redact --no-banner`, full history), `make licenses` (`go-licenses check ./... --allowed_licenses=Apache-2.0,BSD-2-Clause,BSD-3-Clause,MIT,ISC`), `make actions` (`actionlint`), `make check` (lint, lint-rules, test, vuln, secrets, licenses, actions).
- Job names are the required-check names used in Task 10: `Lint`, `Test`, `Security`.

- [ ] **Step 1: Add the tools** with `go get -tool -modfile=tools/go.mod` at the versions in Global Constraints; add the Makefile targets.

- [ ] **Step 2: Run locally** — `make check` → all green. `make secrets` must scan history (`fetch-depth: 0` in CI).

- [ ] **Step 3: Write `ci.yml`** — `on: pull_request` (branches `main`) and `push` (branches `main`), no path filter (these checks become required). `concurrency: {group: ci-${{ github.ref }}, cancel-in-progress: true}`. Top-level `permissions: {}`; each job `permissions: {contents: read}`. Jobs `Lint` (lint, lint-rules, actions), `Test` (`go test -race ./...`), `Security` (checkout with `fetch-depth: 0`; vuln, secrets, licenses). Each job: checkout (pinned, `persist-credentials: false`) → setup-go (pinned, `go-version-file: go.mod`, `cache-dependency-path: "**/go.sum"`) → `make <target>`.

- [ ] **Step 4: Verify** — `make actions` passes; push the branch and open the PR → the three jobs run green.

- [ ] **Step 5: Commit** — `git commit -m "ci: add pull request checks"`

### Task 4 (OPS-004): Contract tooling with buf

**Files:**
- Create: `buf.yaml`, `buf.gen.yaml`, `api/iam/admin/v1/server_info.proto`, generated `api/gen/go/iam/admin/v1/*.go` (committed)
- Create: `scripts/check-proto.sh`
- Modify: `tools/go.mod` (buf, protoc-gen-go, protoc-gen-connect-go), `Makefile`, `.github/workflows/ci.yml` (job `Proto`), `docs/roadmap.md` (OPS-004 ✅)

**Interfaces:**
- Produces: `make proto` (generate), `make proto-check` (the script). Generated Go package `github.com/aleogr/iam/api/gen/go/iam/admin/v1;adminv1` with `adminv1connect.ServerInfoServiceHandler`. Not served in this milestone.

- [ ] **Step 1: Write the minimal contract** — `package iam.admin.v1;` service `ServerInfoService` with `rpc GetServerInfo(GetServerInfoRequest) returns (GetServerInfoResponse)`; `GetServerInfoResponse { string version = 1; }`. AIP-style naming.

- [ ] **Step 2: Configure buf** — `buf.yaml` (v2): module `api`, `lint: use: [STANDARD]`, `breaking: use: [FILE]`. `buf.gen.yaml` (v2): managed mode with `go_package_prefix: github.com/aleogr/iam/api/gen/go`; plugins as local commands `["env","GOWORK=off","go","tool","-modfile=tools/go.mod","protoc-gen-go"]` and the connect equivalent, `out: api/gen/go`, `opt: paths=source_relative`.

- [ ] **Step 3: Write `scripts/check-proto.sh`** — runs `buf lint`; runs `buf breaking --against '.git#branch=origin/main,subdir=api'` only when `git ls-tree -d origin/main api` is non-empty (first PR has nothing to compare); runs `buf generate` then `git diff --exit-code -- api/gen` (fails with "generated code is stale, run make proto").

- [ ] **Step 4: Verify** — `make proto && go build ./...` → compiles; `make proto-check` → pass; edit the proto without regenerating → `make proto-check` fails with the stale message (revert).

- [ ] **Step 5: CI job `Proto`** — checkout with `fetch-depth: 0`, setup-go, `make proto-check`.

- [ ] **Step 6: Commit** — `git commit -m "build: add protobuf contract tooling"`

### Task 5 (OPS-005): Real Postgres in CI and the integration-test harness

**Files:**
- Create: `internal/database/pool.go`, `internal/database/pool_test.go`
- Create: `internal/testdb/testdb.go`, `internal/testdb/testdb_test.go`
- Modify: `cmd/iam/main.go` (database readiness check), `internal/httpserver/health_test.go`, `.github/workflows/ci.yml` (Postgres service in `Test`), `docs/roadmap.md` (OPS-005 ✅)

**Interfaces:**
- Consumes: `httpserver.ReadinessCheck` (Task 1).
- Produces:
  - `package database`: `const MaxConns int32 = 2` (lab budget, ADR 0001); `func Open(ctx context.Context, url string) (*pgxpool.Pool, error)` — parses, sets `MaxConns`, pings; `func ReadinessCheck(pool *pgxpool.Pool) httpserver.ReadinessCheck` named `"database"`.
  - `package testdb`: `func New(t testing.TB) (url string)` — reads `IAM_TEST_DATABASE_URL` (admin connection); if unset, `t.Skip` unless `IAM_REQUIRE_DB=1`, then `t.Fatal`; creates database `iam_test_<pid>_<counter>`, registers `t.Cleanup` that runs `DROP DATABASE ... WITH (FORCE)`; returns a URL for the new database.
- Dependency added: `github.com/jackc/pgx/v5 v5.12.0` (justify in the PR: Postgres protocol, out of scope to implement; approved in spec section 10).

- [ ] **Step 1: Write the failing tests**

```go
func TestNew_IsolatedDatabases(t *testing.T)            // two testdb.New URLs; a table created in one is absent in the other
func TestNew_DropsDatabaseOnCleanup(t *testing.T)        // run a subtest with testdb.New; after it, pg_database has no such name
func TestOpen_MaxConnsIsTwo(t *testing.T)                // pool.Config().MaxConns == 2
func TestOpen_RejectsUnreachableDatabase(t *testing.T)   // url to 127.0.0.1:1 → error within 5s, error text does not contain the password
func TestReadyz_ClosedPoolIsUnavailable(t *testing.T)    // server with database.ReadinessCheck(pool) after pool.Close() → 503, failed ["database"]
```

- [ ] **Step 2: Run them to verify they fail** — with a local Postgres (`docker run --rm -d -p 127.0.0.1:5432:5432 -e POSTGRES_PASSWORD=$(openssl rand -hex 16) postgres:16`, export `IAM_TEST_DATABASE_URL` accordingly) → FAIL (undefined).

- [ ] **Step 3: Implement `database` and `testdb`**, and wire `main`: when `IAM_DATABASE_URL` is set, `database.Open` and register `database.ReadinessCheck`.

- [ ] **Step 4: Run** — `IAM_REQUIRE_DB=1 go test -race ./...` → PASS; without the env var and without `IAM_REQUIRE_DB`, database tests report SKIP.

- [ ] **Step 5: CI** — in job `Test`: `services.postgres` with `postgres:16@sha256:<digest>`, env `POSTGRES_PASSWORD: ${{ github.run_id }}-${{ github.run_attempt }}` (ephemeral, not a secret), health-cmd `pg_isready`, port `5432:5432`; job env `IAM_TEST_DATABASE_URL: postgres://postgres:<same>@127.0.0.1:5432/postgres?sslmode=disable`, `IAM_REQUIRE_DB: "1"`.

- [ ] **Step 6: Commit** — `git commit -m "test: add Postgres integration harness and database readiness"`

### Task 6 (OPS-006): Local environment with one command

**Files:**
- Create: `compose.yaml`, `scripts/dev-env.sh`
- Modify: `Makefile` (`dev`, `dev-down`), `.gitignore` (`.env.local`, `dist/`), `docs/roadmap.md` (OPS-006 ✅)

**Interfaces:**
- Produces: `make dev` → Postgres up (healthy) and `cmd/iam` running on `127.0.0.1:8080` with `IAM_DATABASE_URL` set; `make dev-down` stops and removes the Compose project.

- [ ] **Step 1: `compose.yaml`** — service `postgres` (`postgres:16@sha256:<digest>`), port `127.0.0.1:5432:5432` only, `POSTGRES_PASSWORD` from `.env.local`, healthcheck `pg_isready`, named volume.

- [ ] **Step 2: `scripts/dev-env.sh`** — creates `.env.local` with `POSTGRES_PASSWORD=$(openssl rand -hex 16)` if absent (never a fixed default: ADR 0012, prohibition 6), `docker compose --env-file .env.local up -d --wait`, then `exec go run ./cmd/iam` with `IAM_HTTP_ADDR=127.0.0.1:8080` and the derived `IAM_DATABASE_URL`.

- [ ] **Step 3: Verify** — `make dev` in one shell; `curl -s 127.0.0.1:8080/readyz` → `{"status":"ok"}`; `docker compose stop postgres` → `/readyz` → 503 `["database"]`; `make dev-down`.

- [ ] **Step 4: Commit** — `git commit -m "build: add one-command local environment"`

### Task 7 (OPS-007): Image with ko, pinned base, SBOM and scan

**Files:**
- Create: `.ko.yaml`, `scripts/check-reproducible.sh`
- Modify: `tools/go.mod` (ko, crane, cyclonedx-gomod, osv-scanner), `Makefile` (`image`, `sbom`, `scan`, `repro`), `.github/workflows/ci.yml` (job `Image`), `docs/roadmap.md` (OPS-007 ✅)

**Interfaces:**
- Produces: `dist/iam.tar` (OCI tarball, not pushed), `dist/iam.cdx.json` (CycloneDX SBOM of the binary), `make scan` (osv-scanner on the tarball and on `go.mod`), `make repro` (two clean builds, identical SHA-256).

- [ ] **Step 1: `.ko.yaml`** — `defaultBaseImage: cgr.dev/chainguard/static@sha256:<digest>` (runs as non-root 65532, no shell); build `id: iam`, `main: ./cmd/iam`, `flags: [-trimpath]`, `ldflags: ["-s","-w","-X main.version={{.Git.ShortCommit}}"]`, `env: [CGO_ENABLED=0]`; `defaultPlatforms: [linux/amd64, linux/arm64]`.

- [ ] **Step 2: Make targets** — `image`: `KO_DOCKER_REPO=ko.local go tool -modfile=tools/go.mod ko build ./cmd/iam --push=false --tarball=dist/iam.tar --sbom=none`; `sbom`: `cyclonedx-gomod app -json -licenses -main cmd/iam -output dist/iam.cdx.json`; `scan`: `osv-scanner scan image --archive dist/iam.tar` and `osv-scanner scan source -r .`; `repro`: the script builds `go build -trimpath -buildvcs=false -ldflags="-s -w" ./cmd/iam` twice into separate dirs after `go clean -cache` and compares `sha256sum`.

- [ ] **Step 3: Verify** — `make image sbom scan repro` → tarball and SBOM exist; scan reports no vulnerabilities; repro prints "reproducible: <sha256>". `jq '.components | length' dist/iam.cdx.json` > 0.

- [ ] **Step 4: CI job `Image`** — runs `make image sbom scan repro`; uploads `dist/iam.cdx.json` with `actions/upload-artifact` (pinned).

- [ ] **Step 5: Commit** — `git commit -m "build: add container image, SBOM and image scan"`

### Task 8 (OPS-008): Community and Cyber Resilience Act files

**Files:**
- Create: `LICENSE` (exact Apache License 2.0 text from https://www.apache.org/licenses/LICENSE-2.0.txt), `SECURITY.md`, `CONTRIBUTING.md`, `docs/support-policy.md`
- Modify: `docs/roadmap.md` (OPS-008 ✅)

- [ ] **Step 1: Write the files** (Portuguese, short):
  - `SECURITY.md`: report only through GitHub private vulnerability reporting (`https://github.com/aleogr/iam/security/advisories/new`), never in public issues; acknowledgement target 3 business days; coordinated disclosure; fixes ship as security-only releases separate from feature releases (ADR 0028); supported versions table pointing to `docs/support-policy.md`.
  - `CONTRIBUTING.md`: the project is developed by agents and does not accept external code contributions yet; issues welcome; a CLA will be required when contributions open (ADR 0009); Conventional Commits.
  - `docs/support-policy.md`: before v1.0 only the latest release receives fixes and nothing is production-ready; the post-v1 policy (number of supported versions and API deprecation window) is decided and recorded before v1.0 (brief 4.2, ADR 0028).
- [ ] **Step 2: Verify** — `make licenses` still passes; `sha256sum LICENSE` equals the hash of the downloaded apache.org text.
- [ ] **Step 3: Commit** — `git commit -m "docs: add license, security policy and contribution guide"`

### Task 9 (OPS-009): End-to-end harness with test client A

**Files:**
- Create: `examples/a/go.mod` (module `github.com/aleogr/iam/examples/a`), `examples/a/main.go`, `examples/a/main_test.go`
- Create: `e2e/go.mod` (module `github.com/aleogr/iam/e2e`), `e2e/internal/proc/proc.go`, `e2e/health_test.go`
- Modify: `go.work` (add `./examples/a`, `./e2e`), `Makefile` (`lint`/`test` iterate over all `go.work` modules; new `e2e`), `.github/workflows/ci.yml` (job `E2E`), `docs/roadmap.md` (OPS-009 ✅)

**Interfaces:**
- Produces:
  - Client A: env `A_HTTP_ADDR` (default `127.0.0.1:8081`), `A_IAM_URL` (required). `GET /healthz` → calls `<A_IAM_URL>/healthz` with a 2s timeout; 200 `{"status":"ok","iam":"ok"}` or 503 `{"status":"unavailable","iam":"down"}`. Uses only the standard library and public HTTP endpoints (separate module: `internal/` packages of the server are unimportable by construction).
  - `package proc`: `func Start(t *testing.T, bin string, env []string) *Process` (captures stdout/stderr into `t.Log` on failure, `t.Cleanup` sends SIGTERM and waits); `func (p *Process) Stop()`; `func WaitHTTP(t *testing.T, url string, want int, timeout time.Duration)`; `func FreeAddr(t *testing.T) string`.
- Consumes: a ready database URL in `IAM_E2E_DATABASE_URL`; if unset, the e2e test skips unless `IAM_REQUIRE_DB=1` (same semantics as Task 5). The e2e module does not import the server's `internal/testdb`.

- [ ] **Step 1: Write the failing tests**

```go
// examples/a/main_test.go
func TestHealthz_UpstreamUp(t *testing.T)    // httptest upstream 200 → A answers 200 {"status":"ok","iam":"ok"}
func TestHealthz_UpstreamDown(t *testing.T)  // upstream closed → 503 {"status":"unavailable","iam":"down"}

// e2e/health_test.go
func TestE2E_ClientAReachesIAM(t *testing.T) // builds ./cmd/iam and ./examples/a into t.TempDir(); starts iam with IAM_DATABASE_URL=$IAM_E2E_DATABASE_URL; WaitHTTP(iam /readyz, 200); starts A; A /healthz → 200; stop iam; A /healthz → 503
```

- [ ] **Step 2: Run to verify they fail**, implement client A and `proc`, run to verify they pass — `make e2e` (with Postgres from Task 5's local container) → PASS.

- [ ] **Step 3: CI job `E2E`** — same Postgres service as `Test`; env `IAM_E2E_DATABASE_URL` pointing at database `postgres`; `make e2e`. Update `Lint` and `Test` jobs to the per-module Makefile targets.

- [ ] **Step 4: Commit** — `git commit -m "test: add end-to-end harness with test client A"`

### Task 10 (OPS-010): Required checks, code owners and the auto-merge guard

**Files:**
- Create: `.github/CODEOWNERS`, `.github/sensitive-paths.txt`, `.github/workflows/sensitive-paths.yml`
- Create: `tools/sensitivepaths/main.go`, `tools/sensitivepaths/match.go`, `tools/sensitivepaths/match_test.go` (package in the `tools` module)
- Modify: `docs/roadmap.md` (OPS-010 ✅)

**Interfaces:**
- Produces: (in the `tools` module; test with `cd tools && GOWORK=off go test ./...`) `func Match(patterns []string, path string) bool` (gitignore-style `**` globs); CLI reads patterns from a file argument and changed paths from stdin, prints matching paths, exits 0 always.
- Sensitive paths (one per line in `.github/sensitive-paths.txt`, mirrored in `CODEOWNERS` with owner `@aleogr`): `shared/crypto/**`, `shared/tenancy/**`, `shared/storage/**`, `adapters/postgres/**`, `authn/**`, `authz/core/**`, `authz/server/**`, `.github/**`, `deploy/**`, `tools/**`, `.golangci.yml`, `Makefile`, `scripts/**`, `docs/adr/**`, `docs/brief.md`, `CLAUDE.md`.

- [ ] **Step 1: Write the failing tests**

```go
func TestMatch(t *testing.T) // table: "authn/core/x.go" true; "authz/core/a/b.go" true; "authz/sdk/x.go" false;
                             // "docs/brief.md" true; "docs/roadmap.md" false; ".github/workflows/ci.yml" true;
                             // "shared/cryptography/x.go" false (no prefix match); "CLAUDE.md" true
```

- [ ] **Step 2: Implement `Match` and the CLI; run** — `cd tools && GOWORK=off go test ./...` → PASS.

- [ ] **Step 3: Workflow `sensitive-paths.yml`** — `on: pull_request` types `[opened, synchronize, reopened, auto_merge_enabled]` (never `pull_request_target`); `permissions: {}`; job `guard` with `contents: read`, `pull-requests: write`; lists changed files with `gh pr diff ${{ github.event.pull_request.number }} --name-only` (env `GH_TOKEN: ${{ github.token }}`), pipes them to the CLI; if any match: writes them to `$GITHUB_STEP_SUMMARY`, and if `github.event.pull_request.auto_merge` is set, runs `gh pr merge --disable-auto <number>`. The job never fails (it is not a required check); it only removes auto-merge from sensitive PRs.

- [ ] **Step 4: Verify** — `make actions` passes; on this PR (it touches `.github/**`) the job summary lists the sensitive files; enabling auto-merge on it gets disabled within a minute.

- [ ] **Step 5: Owner steps after merge** (guide one command or one screen at a time, per CLAUDE.md): ruleset `main` → **Require status checks to pass** with `Lint`, `Test`, `Security`, `Proto`, `Image`, `E2E` and **Require branches to be up to date**; keep approvals at 0 (the owner cannot approve own-account PRs; the guard plus owner-only merge on sensitive paths is the control). Confirm **Allow auto-merge** stays on.

- [ ] **Step 6: Commit** — `git commit -m "ci: add code owners and auto-merge guard for sensitive paths"`

### Task 11 (OPS-011): Nightly fuzzing and benchmark regression gate

**Files:**
- Create: `.github/workflows/nightly.yml`, `scripts/fuzz-all.sh`, `scripts/bench-compare.sh`
- Create: `tools/benchgate/main.go`, `tools/benchgate/gate.go`, `tools/benchgate/gate_test.go`, `tools/benchgate/testdata/{regression,noise,improvement}.csv`
- Modify: `tools/go.mod` (benchstat), `Makefile` (`fuzz`, `bench`), `.github/workflows/ci.yml` (short fuzz step in `Test`), `docs/roadmap.md` (OPS-011 ✅)

**Interfaces:**
- Produces:
  - `scripts/fuzz-all.sh`: finds every `Fuzz*` in every `go.work` module (`go test -list '^Fuzz' ./...`), runs each with `go test -run='^$' -fuzz="^$name$" -fuzztime="${FUZZTIME:-5m}"`; on failure copies `testdata/fuzz` dirs to `dist/fuzz-failures/` and exits 1; with no targets prints "no fuzz targets" and exits 0.
  - `scripts/bench-compare.sh <base-ref>`: runs `go test -run='^$' -bench=. -count=10 ./...` on `<base-ref>` (via `git worktree`) and on `HEAD`, then `benchstat -format csv base.txt head.txt > dist/bench.csv`, then the gate.
  - `package main` in `tools/benchgate`: `func Regressions(csv io.Reader, maxDelta float64, alpha float64) ([]Regression, error)`; `type Regression struct { Benchmark, Unit string; DeltaPercent, P float64 }`; CLI flags `-max-delta 10 -alpha 0.05`, exits 1 listing regressions (sec/op, B/op, allocs/op worse than +10% with p < 0.05).

- [ ] **Step 1: Write the failing tests**

```go
func TestRegressions_FlagsSignificantSlowdown(t *testing.T) // testdata/regression.csv: +25% sec/op, p=0.001 → one Regression
func TestRegressions_IgnoresNoise(t *testing.T)             // noise.csv: +15% with p=0.30, +3% with p=0.001 → none
func TestRegressions_IgnoresImprovement(t *testing.T)       // improvement.csv: −20% p=0.001 → none
func TestRegressions_MalformedCSV(t *testing.T)             // garbage → error, not a pass
```

Generate the three CSV fixtures by running real `benchstat -format csv` on hand-written benchmark outputs, so the parser matches the real column layout.

- [ ] **Step 2: Implement, run** — `cd tools && GOWORK=off go test ./...` → PASS.

- [ ] **Step 3: `nightly.yml`** — `on: schedule: cron "17 6 * * *"` (UTC; no lab access involved) and `workflow_dispatch`; `permissions: {}`. Job `Fuzz` (`contents: read`; `FUZZTIME=5m make fuzz`; upload `dist/fuzz-failures` on failure). Job `Bench` (`contents: read`; checkout `fetch-depth: 0`; base ref = `git rev-list -1 --before="25 hours ago" origin/main`; `make bench BASE=<ref>`; upload `dist/bench.csv`). A failed nightly is visible in Actions; no auto-issue in this milestone.

- [ ] **Step 4: Short fuzzing on every PR** (spec section 11) — in `ci.yml` job `Test`, add a step `FUZZTIME=15s make fuzz` after the tests.

- [ ] **Step 5: Verify** — trigger `workflow_dispatch` on the branch: `Fuzz` runs `FuzzLoadEnv` for 5 minutes and passes; `Bench` runs `BenchmarkHealthz` on both refs and passes.

- [ ] **Step 6: Commit** — `git commit -m "ci: add nightly fuzzing and benchmark regression gate"`

---

## Milestone 0 exit check

- [ ] `docs/roadmap.md`: OPS-001 to OPS-011 ✅.
- [ ] A throwaway PR that adds `import "crypto/sha256"` to `cmd/iam/main.go` is blocked by the required `Lint` check (then closed).
- [ ] `make dev`, `make check`, `make e2e`, `make image sbom scan repro` all pass on a clean clone.
