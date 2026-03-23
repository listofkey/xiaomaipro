#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="$ROOT_DIR/deploy/.env"
COMPOSE_FILE="$ROOT_DIR/deploy/docker-compose.prod.yml"
POSTGRES_BOOTSTRAP_SCRIPT="$ROOT_DIR/deploy/scripts/bootstrap-postgres.sh"
POSTGRES_INIT_DIR="$ROOT_DIR/deploy/postgres/init"
POSTGRES_INIT_SQL="$POSTGRES_INIT_DIR/001-public.sql"
SEED_FILE="$ROOT_DIR/deploy/public.sql"
APP_SERVICES=(proxy gateway user-rpc program-rpc payment-rpc order-rpc)
INFRA_SERVICES=(etcd postgres redis rabbitmq kafka kafka-init)
CORE_APP_SERVICES=(gateway user-rpc program-rpc payment-rpc order-rpc proxy)

cd "$ROOT_DIR"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Missing $ENV_FILE. Generate it in CI or create it from deploy/.env.example before deploying." >&2
  exit 1
fi

mkdir -p "$POSTGRES_INIT_DIR"
if [[ -f "$SEED_FILE" ]]; then
  cp "$SEED_FILE" "$POSTGRES_INIT_SQL"
fi

export APP_IMAGE_TAG="${APP_IMAGE_TAG:-latest}"
if [[ -n "${DOCKERHUB_NAMESPACE:-}" ]]; then
  export DOCKERHUB_NAMESPACE
fi

if [[ -n "${DOCKERHUB_USERNAME:-}" && -n "${DOCKERHUB_TOKEN:-}" ]]; then
  printf '%s\n' "$DOCKERHUB_TOKEN" | docker login -u "$DOCKERHUB_USERNAME" --password-stdin
fi

echo "Deploying images from ${DOCKERHUB_NAMESPACE:-deploy/.env} with tag ${APP_IMAGE_TAG}"

docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" pull "${APP_SERVICES[@]}"
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d --remove-orphans "${INFRA_SERVICES[@]}"
bash "$POSTGRES_BOOTSTRAP_SCRIPT"
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d --remove-orphans "${CORE_APP_SERVICES[@]}"
