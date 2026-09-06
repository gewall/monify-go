#!/usr/bin/env bash
# Nightly Postgres backup. Add to cron:
#   0 2 * * * /path/to/deploy/backup.sh >> /var/log/monify-backup.log 2>&1
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/var/backups/monify}"
KEEP_DAYS="${KEEP_DAYS:-14}"
STAMP="$(date +%Y%m%d-%H%M%S)"

mkdir -p "$BACKUP_DIR"
docker compose -f "$(dirname "$0")/docker-compose.prod.yml" exec -T db \
	pg_dump -U "${POSTGRES_USER:-monify}" "${POSTGRES_DB:-monify}" \
	| gzip > "$BACKUP_DIR/monify-$STAMP.sql.gz"

find "$BACKUP_DIR" -name 'monify-*.sql.gz' -mtime "+$KEEP_DAYS" -delete
echo "backup ok: $BACKUP_DIR/monify-$STAMP.sql.gz"
