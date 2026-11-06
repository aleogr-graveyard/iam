.PHONY: test build

test:
	go test -race ./...

build:
	go build -trimpath -o dist/iam ./cmd/iam
