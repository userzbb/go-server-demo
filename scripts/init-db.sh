#!/bin/bash
# 初始化 PostgreSQL 数据库表结构

set -e

echo "========================================="
echo "  初始化 Omega Server 数据库"
echo "========================================="

# 检查 PostgreSQL 容器是否运行
if ! podman ps --filter "name=omega-postgres" --format "{{.Names}}" | grep -q omega-postgres; then
    echo "❌ PostgreSQL 容器未运行，请先执行: podman-compose -f deployments/docker-compose.yml up -d"
    exit 1
fi

echo "📦 导入表结构..."
podman exec -i omega-postgres psql -U omega -d omega < scripts/sql/init.sql

echo "✅ 数据库初始化完成！"
