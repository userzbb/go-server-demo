#!/bin/bash
# 恢复 PostgreSQL 数据库
# 用法: ./scripts/restore-db.sh <备份文件路径>

if [ -z "$1" ]; then
    echo "用法: $0 <备份文件路径>"
    exit 1
fi

if [ ! -f "$1" ]; then
    echo "文件不存在: $1"
    exit 1
fi

echo "开始恢复数据库: $1"
podman exec -i omega-postgres psql -U omega -d omega < "$1"
echo "恢复完成！"
