#!/usr/bin/env bash
# Proves the architectural lint rules bite: the fixture module must fail lint
# and the output must show every rule marker.
set -uo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
markers=(core-no-infra crypto-only-in-shared-crypto no-math-rand "inject a clock")

cd "$root/tools/lintfixture" || exit 1
out="$(GOWORK=off go tool -modfile="$root/tools/go.mod" golangci-lint run --config "$root/.golangci.yml" ./... 2>&1)"
status=$?

if [ "$status" -eq 0 ]; then
	echo "FAIL: lint fixture passed; the boundary rules do not bite" >&2
	exit 1
fi

missing=0
for m in "${markers[@]}"; do
	if ! grep -qF -- "$m" <<<"$out"; then
		echo "FAIL: marker not found in lint output: $m" >&2
		missing=1
	fi
done
if [ "$missing" -ne 0 ]; then
	echo "$out" >&2
	exit 1
fi
echo "all ${#markers[@]} boundary rules fired"
