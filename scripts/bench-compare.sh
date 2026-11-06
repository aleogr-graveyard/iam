#!/usr/bin/env bash
# Usage: bench-compare.sh <base-ref>
# Benchmarks <base-ref> (in a temporary git worktree) and HEAD, compares them with
# benchstat and fails when benchgate finds a significant regression. BENCHSTAT is
# the benchstat command (default: benchstat on PATH). Every error is fatal, except
# a base ref without benchmarks, which has nothing to compare against.
set -euo pipefail

[ $# -eq 1 ] || { echo "usage: $0 <base-ref>" >&2; exit 2; }
root=$(cd "$(dirname "$0")/.." && pwd)
base_ref=$1
benchstat=${BENCHSTAT:-benchstat}
cd "$root"
mkdir -p dist
tmp=$(mktemp -d)
worktree="$tmp/base"
cleanup() {
	git -C "$root" worktree remove --force "$worktree" 2>/dev/null || true
	git -C "$root" worktree prune
	rm -rf "$tmp"
}
trap cleanup EXIT

# Runs the benchmarks of every go.work module (not e2e) found under $1. The module
# list is captured on its own so that a go failure aborts instead of looking like
# "no benchmarks".
run_benchmarks() {
	local top=$1 d dirs
	dirs=$(cd "$top" && go list -m -f '{{.Dir}}')
	for d in $dirs; do
		[ "$d" = "$top/e2e" ] && continue
		(cd "$d" && go test -run='^$' -bench=. -benchmem -count=10 ./...)
	done
}

git rev-parse --verify --quiet "$base_ref^{commit}" >/dev/null || { echo "unknown base ref: $base_ref" >&2; exit 1; }
git worktree add --detach "$worktree" "$base_ref" >/dev/null

run_benchmarks "$worktree" >"$tmp/base.txt"
if ! grep -q '^Benchmark' "$tmp/base.txt"; then
	echo "no baseline benchmarks at $base_ref; nothing to compare"
	exit 0
fi
run_benchmarks "$root" >"$tmp/head.txt"
grep -q '^Benchmark' "$tmp/head.txt" || { echo "no benchmarks at HEAD" >&2; exit 1; }

# shellcheck disable=SC2086 # BENCHSTAT is a command line, split on purpose.
env $benchstat -format csv "$tmp/base.txt" "$tmp/head.txt" >dist/bench.csv
(cd "$root/tools" && GOWORK=off go run ./benchgate "$root/dist/bench.csv")
