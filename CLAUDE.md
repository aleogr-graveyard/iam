# Working agreements

These rules hold for every session working in this repository. The design is
`docs/superpowers/specs/2026-11-06-iam-design.md`; decisions are in `docs/adr/`;
the plan of record is `docs/roadmap.md`; the original brief is `docs/brief.md`.

## Talking to the owner

- Talk to the owner in **Portuguese**: objective, short, tables over paragraphs;
  details only on request.
- **Guiding the owner through manual steps: one command at a time.** Give a
  single command, wait for the owner to run it and show the output, and only
  then give the next one. Never hand over a batch of commands, even when they
  are independent. This holds in every session and every repository.
- **Claude Code is the only developer.** The owner never installs an
  environment, runs code, tests or tools, or edits code. Every development,
  test and environment check is done by Claude Code; what cannot run in the
  session runs in CI. The owner only performs account-level actions that no
  agent can (approvals, merges, console settings, secrets).
- **Every decision asked of the owner comes with a recommendation**, without
  exception: the options, the recommended one first, and a one-line reason.
- **Claude Code watches every pull request it opens** (CI, reviews, merge or
  close) and acts on each change without the owner having to report it.
- Never ask for secrets in chat. Never commit them: this repository is
  **public** and its whole history will stay public.

## Languages

- Code, identifiers, comments, commit messages and PR titles: **English**.
- PR descriptions and versioned documentation: **Portuguese** until the first
  release, then an English version of all documentation.
- Commit messages and PR titles follow Conventional Commits (release-please).

## Decisions and documents

- A design decision becomes an ADR **only after the owner approves it**.
  Implementation decisions inside the approved design are yours; record the
  non-obvious ones in the PR description.
- When an approved decision settles a brief item, update that item's marking in
  `docs/brief.md` and point it to the ADR. The brief uses CRLF line endings:
  preserve them and its final newline.
- Mark a roadmap item done in the same PR that completes it.
- Contracts first: a contract (C1–C10 in the spec) changes only through an
  explicit, reviewed PR with contract tests updated.

## Architecture rules (checked by CI where marked)

- Modular monolith: one binary, `cmd/iam`. Cores (`authn/core`, `authz/core`)
  depend only on ports; never on HTTP, SQL, cloud SDKs, adapters or servers
  (`depguard`).
- All cryptography goes through `shared/crypto`: no `crypto/*` primitive and no
  `go-jose` import anywhere else (`depguard`, `forbidigo`). Randomness only from
  `crypto/rand`. Constant-time comparison for secrets.
- Clock and randomness are injected into cores; no `time.Now` or `math/rand` in
  core packages (`forbidigo`). Never trust a time sent by a client.
- Every row and every query carries `tenant_id`; storage ports require it.
  Postgres RLS is the second barrier, never the only one.
- Explicit state machines for flows and lifecycles; undeclared transitions are
  refused; every transition emits an audit event.
- Authorization fails closed: any error during evaluation denies.
- Listing is never more permissive than check (property tested).
- No configuration in code: names, domains, URLs, regions, limits and
  providers come from typed, validated configuration.

## Security rules

- The absolute prohibitions of ADR 0012 are rejected at configuration load and
  by the API; never add a bypass.
- Every threat in `docs/threat-model.md` gets an abuse test that attempts the
  attack and asserts it fails. Every vulnerability found becomes a permanent
  regression test.
- Limits on size and depth for every parsed input (JSON, policies, tokens,
  WebAuthn data).
- Logs never contain tokens, secrets, passwords or raw personal data;
  authentication responses send `Cache-Control: no-store`.
- Offensive tests only against local, CI or lab environments; never against
  third parties.

## Dependencies

- Prefer the standard library and `golang.org/x/...`. Accepted runtime
  dependencies are listed in spec section 10; `cedar-go` is test-only.
- Any new dependency needs a written justification in the PR: active
  maintenance, compatible license (Apache-2.0 project), security history, few
  transitive dependencies.
- GitHub Actions are pinned to a full commit SHA; third-party actions must be
  on the repository allowlist.

## Data and the lab instance

- The lab database lives in the shared `lab-postgres` (ADR 0001). The service
  pool is at most 2, one instance; no `LISTEN/NOTIFY` or dedicated connections;
  the connection-budget test must stay at or below 6.
- Never touch a database, user or role outside `iam` / `iam_*` / `iam-*`.
- Nothing scheduled runs on weekdays between 22:00 and 08:00 (São Paulo); a
  deploy reads `activationPolicy` first and stops with a clear message.
- Migrations are timestamp-named and backward compatible (expand/contract).

## Infrastructure

- Terraform under `deploy/` is applied only by CI after a merge to `main`,
  except for the documented bootstrap (`deploy/gcp/README.md`). A session never
  runs `apply`.
- Each new resource type needs read access for `iam-planner` and write access
  for `iam-deployer` in `deploy/gcp/terraform/ci_identities.tf`, including
  `actAs` on any service account the resource runs as.
- In `aleogr/lab`, follow its own `CLAUDE.md`: open the PR, the owner merges.

## Workflow

- Work on a branch; open a PR; the owner merges sensitive areas: cryptography,
  authentication, sessions and tokens, policy evaluation, multi-party approval,
  tenant isolation, CI/CD and release pipeline, permissions, ADRs, the brief and
  this file. Other PRs may auto-merge once required checks pass.
- Squash merge only; keep PRs small and focused on one roadmap item.
- Before pushing: `gofmt`, `golangci-lint run`, `go test -race ./...`; for
  Terraform, `terraform fmt -check -recursive`, `terraform init -backend=false`,
  `terraform validate` (version in `.terraform-version`).
- Every delivery is reviewed by an agent that did not write it, using the
  quality checklist (brief 4.30) and the relevant OWASP cheat sheets.
- Do not parallelize the domain model, sessions and tokens, or cryptographic
  code across agents.

## Code quality

Readability, simplicity, low coupling and high cohesion, no duplication
(abstract on the third repetition, not the second), testability, idiomatic Go
patterns only where they earn their place. Tests first (TDD), the test pyramid
weighted toward unit and contract tests.
