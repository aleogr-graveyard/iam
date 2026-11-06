#!/usr/bin/env bash
# Contract checks: lint, breaking changes against origin/main, and proof that
# the committed generated code is current. BUF is the buf invocation (the
# Makefile passes it); it defaults to a plain buf on PATH.
set -euo pipefail

cd "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
buf="${BUF:-buf}"

$buf lint

# The first PR that adds api/ has nothing to compare against.
if [ -n "$(git ls-tree -d origin/main api 2>/dev/null)" ]; then
	$buf breaking --against '.git#ref=refs/remotes/origin/main,subdir=api'
else
	echo "skipping buf breaking: origin/main has no api/ yet"
fi

$buf generate
if ! git diff --exit-code -- api/gen; then
	echo "generated code is stale, run make proto" >&2
	exit 1
fi
if [ -n "$(git ls-files --others --exclude-standard -- api/gen)" ]; then
	echo "generated code is stale (untracked files), run make proto" >&2
	exit 1
fi
