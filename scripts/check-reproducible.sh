#!/usr/bin/env bash
# Builds ./cmd/iam twice from a cold build cache into separate directories and
# fails unless both binaries have the same SHA-256.
set -euo pipefail

cd "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export CGO_ENABLED=0

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

for n in 1 2; do
	go clean -cache
	mkdir "$work/$n"
	go build -trimpath -buildvcs=false -ldflags="-s -w" -o "$work/$n/iam" ./cmd/iam
done

first="$(sha256sum "$work/1/iam" | cut -d' ' -f1)"
second="$(sha256sum "$work/2/iam" | cut -d' ' -f1)"
if [ "$first" != "$second" ]; then
	echo "not reproducible: $first != $second" >&2
	exit 1
fi
echo "reproducible: $first"
