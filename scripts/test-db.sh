#!/bin/bash
# 测试数据库连接脚本（使用 Podman 容器）

set -e

echo "========================================="
echo "  Omega Server 数据库连接测试"
echo "========================================="
echo ""

echo "[1/3] 测试 PostgreSQL..."
if podman exec omega-postgres psql -U omega -d omega -c "SELECT 1" > /dev/null 2>&1; then
    echo "  ✅ PostgreSQL 连接成功"
else
    echo "  ❌ PostgreSQL 连接失败"
    exit 1
fi

echo "[2/3] 测试 MongoDB..."
if podman exec omega-mongodb mongosh -u omega -p omega123 --authenticationDatabase admin --eval "db.runCommand({ping: 1})" > /dev/null 2>&1; then
    echo "  ✅ MongoDB 连接成功"
else
    echo "  ❌ MongoDB 连接失败"
    exit 1
fi

echo "[3/3] 测试 Redis..."
if podman exec omega-redis redis-cli ping > /dev/null 2>&1; then
    echo "  ✅ Redis 连接成功"
else
    echo "  ❌ Redis 连接失败"
    exit 1
fi

echo ""
echo "========================================="
echo "  ✅ 所有数据库连接正常！"
echo "========================================="
