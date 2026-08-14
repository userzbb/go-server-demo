#!/bin/bash
# 备份 PostgreSQL 数据库

set -e
BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/omega_db_$TIMESTAMP.sql"

mkdir -p "$BACKUP_DIR"

echo "开始备份数据库..."
podman exec omega-postgres pg_dump -U omega -d omega > "$BACKUP_FILE"
echo "备份完成: $BACKUP_FILE"
