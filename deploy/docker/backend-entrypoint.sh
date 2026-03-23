#!/bin/sh

set -eu

TEMPLATE_DIR="/app/config-templates"
CONFIG_DIR="/app/config"

mkdir -p "$CONFIG_DIR"

if [ -d "$TEMPLATE_DIR" ]; then
  for template in "$TEMPLATE_DIR"/*.yaml; do
    [ -f "$template" ] || continue
    output="$CONFIG_DIR/$(basename "$template")"
    envsubst < "$template" > "$output"
  done
fi

exec /app/service "$@"
