# Each external tool has its own module under tools/<name>/ so that their
# dependency graphs cannot conflict. gotool builds the invocation once.
gotool = GOWORK=off go tool -modfile=$(CURDIR)/tools/$(1)/go.mod $(1)

GOLANGCI_LINT := $(call gotool,golangci-lint)
GOVULNCHECK   := $(call gotool,govulncheck)
GITLEAKS      := $(call gotool,gitleaks)
ACTIONLINT    := $(call gotool,actionlint)
GO_LICENSES   := $(call gotool,go-licenses)
BUF           := $(call gotool,buf)
CRANE         := $(call gotool,crane)
KO            := $(call gotool,ko)
CYCLONEDX     := $(call gotool,cyclonedx-gomod)
OSV_SCANNER   := $(call gotool,osv-scanner)
BENCHSTAT     := $(call gotool,benchstat)

.PHONY: dev dev-down dev-reset test e2e build lint lint-rules workspace-check vuln secrets licenses actions proto proto-check config-doc config-doc-check image sbom scan repro check fuzz bench dev-check

# Every module listed in go.work is self-contained, so each is tested from its own
# directory. The e2e module is excluded on purpose: it needs built binaries and its
# own database variable, and runs through `make e2e` (CI job E2E).
# The module list is captured first: a broken go.work fails instead of looping zero
# times. `make workspace-check` proves no module is missing from go.work.
test:
	@mods=$$(go list -m -f '{{.Dir}}') || exit 1; \
	[ -n "$$mods" ] || { echo "go.work lists no modules" >&2; exit 1; }; \
	for d in $$(printf '%s\n' "$$mods" | grep -vxF '$(CURDIR)/e2e'); do \
		(cd $$d && go test -race ./...) || exit 1; \
	done
	cd tools && GOWORK=off go test -race ./...

# Builds the server and test client A and runs them against a real database.
e2e:
	cd e2e && go test -race -count=1 -v ./...

build:
	go build -trimpath -o dist/iam ./cmd/iam

# Lints every module in go.work plus the tools module (the repo's own tool code;
# tools/<tool>/ and tools/lintfixture are nested modules and are not part of it).
lint:
	@mods=$$(go list -m -f '{{.Dir}}') || exit 1; \
	[ -n "$$mods" ] || { echo "go.work lists no modules" >&2; exit 1; }; \
	for d in $$mods; do \
		(cd $$d && $(GOLANGCI_LINT) run --config $(CURDIR)/.golangci.yml ./...) || exit 1; \
	done
	cd tools && $(GOLANGCI_LINT) run --config $(CURDIR)/.golangci.yml ./...

lint-rules:
	./scripts/check-lint-rules.sh

# Every go.mod is in go.work or explicitly excluded, so no module escapes lint/test.
workspace-check:
	./scripts/check-workspace.sh

vuln:
	$(GOVULNCHECK) ./...

# Scans the full git history (CI checks out with fetch-depth: 0).
secrets:
	$(GITLEAKS) git --redact --no-banner

licenses:
	$(GO_LICENSES) check ./... --allowed_licenses=Apache-2.0,BSD-2-Clause,BSD-3-Clause,MIT,ISC

actions:
	$(ACTIONLINT)

# Regenerates the committed Go code from api/.
proto:
	$(BUF) generate

# Lint, breaking-change check against origin/main, and generation freshness.
proto-check:
	GOWORK=off BUF='go tool -modfile=$(CURDIR)/tools/buf/go.mod buf' ./scripts/check-proto.sh

# Regenerates the configuration reference from the struct tags of shared/config.
config-doc:
	go run ./shared/config/cmd/configdoc > docs/config-reference.md

# Fails when the committed reference differs from what the tags produce.
config-doc-check:
	@go run ./shared/config/cmd/configdoc | diff -u docs/config-reference.md - \
		|| { echo "docs/config-reference.md is stale; run 'make config-doc'" >&2; exit 1; }

# Image targets are slow and need the network, so they are not part of `check`;
# CI runs them in the Image job.
# The tarball holds one platform (an OCI tarball cannot carry an index) and ko.local
# would load into a Docker daemon, so a dummy repository name is used instead.
image:
	mkdir -p dist
	KO_DOCKER_REPO=iam.invalid/iam $(KO) build ./cmd/iam --push=false --platform=linux/amd64 --tarball=dist/iam.tar --sbom=none

sbom:
	mkdir -p dist
	$(CYCLONEDX) app -json -licenses -main cmd/iam -output dist/iam.cdx.json

# Only the shipped module is scanned: tools/*/go.mod are build-time tools, not part
# of the image, and their advisories would fail the build for code we never ship.
scan:
	$(OSV_SCANNER) scan image --archive dist/iam.tar
	$(OSV_SCANNER) scan source -L go.mod

repro:
	./scripts/check-reproducible.sh

# Fuzzes every Fuzz target in the go.work modules; FUZZTIME defaults to 5m.
fuzz:
	./scripts/fuzz-all.sh

# Benchmarks BASE (a git ref, default origin/main) against HEAD and fails on a
# significant regression.
BASE ?= origin/main
bench:
	BENCHSTAT='$(BENCHSTAT)' ./scripts/bench-compare.sh $(BASE)

check: workspace-check lint lint-rules test vuln secrets licenses actions proto-check config-doc-check

# Local environment: PostgreSQL in Docker Compose plus the server on 127.0.0.1:8080.
dev:
	./scripts/dev-env.sh

# `down` never uses the password; the placeholder only satisfies compose
# interpolation, so teardown works even when .env.local is missing.
COMPOSE_DOWN := POSTGRES_PASSWORD=unused docker compose $(if $(wildcard .env.local),--env-file .env.local,)

# Exercises make dev end to end; needs Docker (runs in CI).
dev-check:
	scripts/check-dev-env.sh

dev-down:
	$(COMPOSE_DOWN) down

# Drops the database volume and the generated password.
dev-reset:
	$(COMPOSE_DOWN) down -v
	rm -f .env.local
