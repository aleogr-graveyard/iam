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

.PHONY: dev dev-down dev-reset test e2e build lint lint-rules vuln secrets licenses actions proto proto-check image sbom scan repro check

# Every module listed in go.work is self-contained, so each is tested from its own
# directory. The e2e module skips itself here unless IAM_E2E_DATABASE_URL is set.
test:
	@for d in $$(go list -m -f '{{.Dir}}'); do \
		(cd $$d && go test -race ./...) || exit 1; \
	done

# Builds the server and test client A and runs them against a real database.
e2e:
	cd e2e && go test -race -count=1 -v ./...

build:
	go build -trimpath -o dist/iam ./cmd/iam

# Lints every module in go.work (tools/lintfixture is deliberately outside it).
lint:
	@for d in $$(go list -m -f '{{.Dir}}'); do \
		(cd $$d && $(GOLANGCI_LINT) run --config $(CURDIR)/.golangci.yml ./...) || exit 1; \
	done

lint-rules:
	./scripts/check-lint-rules.sh

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

check: lint lint-rules test vuln secrets licenses actions proto-check

# Local environment: PostgreSQL in Docker Compose plus the server on 127.0.0.1:8080.
dev:
	./scripts/dev-env.sh

# `down` never uses the password; the placeholder only satisfies compose
# interpolation, so teardown works even when .env.local is missing.
COMPOSE_DOWN := POSTGRES_PASSWORD=unused docker compose $(if $(wildcard .env.local),--env-file .env.local,)

dev-down:
	$(COMPOSE_DOWN) down

# Drops the database volume and the generated password.
dev-reset:
	$(COMPOSE_DOWN) down -v
	rm -f .env.local
