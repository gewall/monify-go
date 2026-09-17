#!/usr/bin/env bash
# Pulls the latest :latest image and recreates the app (only restarts when the
# image actually changed) so pushes to master roll out without a manual step.
# Add to cron:
#   */5 * * * * /home/gewall/apps/monify/auto-pull.sh >> /var/log/monify-autopull.log 2>&1
set -euo pipefail

cd "$(dirname "$0")"

COMPOSE_FILE="docker-compose.prod.yml"
IMAGE="ghcr.io/gewall/monify-go:latest"

before="$(docker image inspect --format '{{.Id}}' "$IMAGE" 2>/dev/null || echo none)"
docker compose -f "$COMPOSE_FILE" pull app
after="$(docker image inspect --format '{{.Id}}' "$IMAGE" 2>/dev/null || echo none)"

if [ "$before" = "$after" ]; then
	echo "$(date -Is) no change"
	exit 0
fi

echo "$(date -Is) new image $after, deploying"
docker compose -f "$COMPOSE_FILE" up -d
