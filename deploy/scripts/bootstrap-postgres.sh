#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SEED_FILE="$ROOT_DIR/deploy/public.sql"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER_NAME:-xiaomai-postgres}"
MAX_RETRIES="${POSTGRES_BOOTSTRAP_RETRIES:-30}"
SLEEP_SECONDS="${POSTGRES_BOOTSTRAP_SLEEP_SECONDS:-2}"

if [[ ! -f "$SEED_FILE" ]]; then
  echo "Skip postgres bootstrap: $SEED_FILE not found."
  exit 0
fi

run_psql() {
  local sql="$1"
  docker exec "$POSTGRES_CONTAINER" sh -lc \
    'PGPASSWORD="$POSTGRES_PASSWORD" psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tAc "$1"' \
    -- "$sql"
}

for ((i=1; i<=MAX_RETRIES; i++)); do
  if run_psql "SELECT 1" >/dev/null 2>&1; then
    break
  fi

  if [[ "$i" -eq "$MAX_RETRIES" ]]; then
    echo "Postgres bootstrap failed: database not ready after $MAX_RETRIES attempts." >&2
    exit 1
  fi

  sleep "$SLEEP_SECONDS"
done

has_event_table="$(run_psql "SELECT CASE WHEN to_regclass('public.event') IS NULL THEN '0' ELSE '1' END;")"
has_event_table="$(echo "$has_event_table" | tr -d '[:space:]')"

if [[ "$has_event_table" == "1" ]]; then
  echo "Postgres schema already initialized, skipping bootstrap."
  exit 0
fi

echo "Bootstrapping postgres schema from deploy/public.sql ..."
docker exec -i "$POSTGRES_CONTAINER" sh -lc \
  'PGPASSWORD="$POSTGRES_PASSWORD" psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < "$SEED_FILE"
echo "Postgres bootstrap completed."
