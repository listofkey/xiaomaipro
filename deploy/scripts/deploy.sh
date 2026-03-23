#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="$ROOT_DIR/deploy/.env"
EXAMPLE_ENV_FILE="$ROOT_DIR/deploy/.env.example"
COMPOSE_FILE="$ROOT_DIR/deploy/docker-compose.prod.yml"
APP_SERVICES=(proxy gateway user-rpc program-rpc payment-rpc order-rpc)

cd "$ROOT_DIR"

if [[ ! -f "$ENV_FILE" ]]; then
  cp "$EXAMPLE_ENV_FILE" "$ENV_FILE"
  echo "Created deploy/.env from deploy/.env.example. Update secrets when needed."
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
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d --remove-orphans
