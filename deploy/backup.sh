#!/usr/bin/env bash
# Per-database backup of the "monify" DB from the shared Postgres container.
# The shared stack (~/infra/postgres) already runs a nightly pg_dumpall for all
# databases; this is an extra monify-only dump. Add to cron if you want it:
#   0 2 * * * /path/to/deploy/backup.sh >> /var/log/monify-backup.log 2>&1
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-$HOME/backups/monify}"
KEEP_DAYS="${KEEP_DAYS:-14}"
PG_CONTAINER="${PG_CONTAINER:-shared-postgres-postgres-1}"
PG_SUPERUSER="${PG_SUPERUSER:-pgroot}"
STAMP="$(date +%Y%m%d-%H%M%S)"

mkdir -p "$BACKUP_DIR"
docker exec -t "$PG_CONTAINER" \
	pg_dump -U "$PG_SUPERUSER" -Fc monify \
	> "$BACKUP_DIR/monify-$STAMP.dump"

find "$BACKUP_DIR" -name 'monify-*.dump' -mtime "+$KEEP_DAYS" -delete
echo "backup ok: $BACKUP_DIR/monify-$STAMP.dump"
