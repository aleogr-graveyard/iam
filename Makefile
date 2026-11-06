# Each external tool has its own module under tools/<name>/ so that their
# dependency graphs cannot conflict. gotool builds the invocation once.
gotool = GOWORK=off go tool -modfile=$(CURDIR)/tools/$(1)/go.mod $(1)

GOLANGCI_LINT := $(call gotool,golangci-lint)
GOVULNCHECK   := $(call gotool,govulncheck)
GITLEAKS      := $(call gotool,gitleaks)
ACTIONLINT    := $(call gotool,actionlint)
GO_LICENSES   := $(call gotool,go-licenses)
BUF           := $(call gotool,buf)

.PHONY: test build lint lint-rules vuln secrets licenses actions proto proto-check check

test:
	go test -race ./...

build:
	go build -trimpath -o dist/iam ./cmd/iam

# Lints every module in go.work (tools/lintfixture is deliberately outside it).
lint:
	@for d in $$(go list -m -f '{{.Dir}}'); do \
		$(GOLANGCI_LINT) run --config $(CURDIR)/.golangci.yml $$d/... || exit 1; \
	done

lint-rules:
	./scripts/check-lint-rules.sh

vuln:
	$(GOVULNCHECK) ./...

# Scans the full git history (CI checks out with fetch-depth: 0).
secrets:
	$(GITLEAKS) git --redact --no-banner

# The root module is ignored until the LICENSE file lands (Task 8 removes the ignore).
licenses:
	$(GO_LICENSES) check ./... --ignore github.com/aleogr/iam --allowed_licenses=Apache-2.0,BSD-2-Clause,BSD-3-Clause,MIT,ISC

actions:
	$(ACTIONLINT)

# Regenerates the committed Go code from api/.
proto:
	$(BUF) generate

# Lint, breaking-change check against origin/main, and generation freshness.
proto-check:
	GOWORK=off BUF='go tool -modfile=$(CURDIR)/tools/buf/go.mod buf' ./scripts/check-proto.sh

check: lint lint-rules test vuln secrets licenses actions proto-check
