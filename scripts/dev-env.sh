#!/usr/bin/env bash
# Starts the local environment: PostgreSQL via Docker Compose, then the server.
# The generated password lives only in .env.local (git-ignored).
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

for tool in docker openssl go; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "dev-env: required tool not found: $tool" >&2
    exit 1
  fi
done

env_file=".env.local"
if [[ ! -f "$env_file" ]]; then
  pw="$(openssl rand -hex 16)" || pw=""
  if [[ -z "$pw" ]]; then
    echo "dev-env: failed to generate a password with openssl" >&2
    exit 1
  fi
  (umask 077 && printf 'POSTGRES_PASSWORD=%s\n' "$pw" >"$env_file")
fi

password="$(sed -n 's/^POSTGRES_PASSWORD=//p' "$env_file")"
if [[ -z "$password" ]]; then
  echo "dev-env: POSTGRES_PASSWORD is empty or missing in $env_file; run 'make dev-reset' and try again" >&2
  exit 1
fi

docker compose --env-file "$env_file" up -d --wait

export IAM_HTTP_ADDR="127.0.0.1:8080"
export IAM_DATABASE_URL="postgres://iam:${password}@127.0.0.1:5432/iam?sslmode=disable"
echo "dev-env: server on http://${IAM_HTTP_ADDR} (database URL not shown)"
exec go run ./cmd/iam
