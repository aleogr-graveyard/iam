#!/usr/bin/env bash
# Exercises the one-command local environment end to end (Docker required):
# `make dev` must bring Postgres and the server up with /readyz 200, a stopped
# database must turn /readyz into 503, and `make dev-reset` must clean up.
# Runs in CI so the owner never has to run it by hand (CLAUDE.md).
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."
mkdir -p dist
log="dist/dev-env.log"
url="http://127.0.0.1:8080/readyz"

cleanup() {
  status=$?
  [[ -n "${dev_pid:-}" ]] && kill "$dev_pid" 2>/dev/null || true
  pkill -f 'go-build.*/iam' 2>/dev/null || true
  make dev-reset >/dev/null 2>&1 || true
  if [[ $status -ne 0 ]]; then
    echo "dev-check: failed; environment log follows" >&2
    cat "$log" >&2 || true
  fi
  exit "$status"
}
trap cleanup EXIT

# wait_status <expected-code> <timeout-seconds>
wait_status() {
  local want=$1 deadline=$((SECONDS + $2)) got=""
  while ((SECONDS < deadline)); do
    got="$(curl -s -o /dev/null -w '%{http_code}' --max-time 3 "$url" || true)"
    [[ "$got" == "$want" ]] && return 0
    sleep 2
  done
  echo "dev-check: expected $want from $url, last got '${got}'" >&2
  return 1
}

make dev >"$log" 2>&1 &
dev_pid=$!

wait_status 200 240
echo "dev-check: /readyz 200 with the database up"

docker compose --env-file .env.local stop postgres >>"$log" 2>&1
wait_status 503 30
echo "dev-check: /readyz 503 with the database stopped"
