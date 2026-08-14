#!/bin/bash
# 跨平台编译脚本
# 用法: ./scripts/build.sh [linux|windows|darwin] [amd64|arm64]

set -e

OS=${1:-linux}
ARCH=${2:-amd64}
OUTPUT_DIR="bin/${OS}_${ARCH}"

echo "🔨 编译目标: $OS/$ARCH"
echo "📁 输出目录: $OUTPUT_DIR"

mkdir -p "$OUTPUT_DIR"

# 编译网关服务
GOOS=$OS GOARCH=$ARCH go build -ldflags="-s -w" -o "$OUTPUT_DIR/gate" cmd/gate/main.go
echo "✅ gate 编译完成"

# 编译逻辑服务
GOOS=$OS GOARCH=$ARCH go build -ldflags="-s -w" -o "$OUTPUT_DIR/logic" cmd/logic/main.go
echo "✅ logic 编译完成"

# 编译中心服务
GOOS=$OS GOARCH=$ARCH go build -ldflags="-s -w" -o "$OUTPUT_DIR/center" cmd/center/main.go
echo "✅ center 编译完成"

echo "🎉 所有服务编译完成！产物位于: $OUTPUT_DIR"
