TOOL := GOWORK=off go tool -modfile=tools/go.mod

.PHONY: test build lint lint-rules

test:
	go test -race ./...

build:
	go build -trimpath -o dist/iam ./cmd/iam

# Lints every module in go.work (tools/lintfixture is deliberately outside it).
lint:
	@for d in $$(go list -m -f '{{.Dir}}'); do \
		$(TOOL) golangci-lint run --config $(CURDIR)/.golangci.yml $$d/... || exit 1; \
	done

lint-rules:
	./scripts/check-lint-rules.sh
