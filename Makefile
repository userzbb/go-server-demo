.PHONY: help lint test build ci db-up db-down db-status db-test db-init db-backup db-restore docker-build

help:
	@echo "📦 Omega Server Makefile"
	@echo ""
	@echo "代码质量:"
	@echo "  make lint    - 运行代码质量检查"
	@echo "  make test    - 运行单元测试并生成覆盖率报告"
	@echo "  make build   - 编译所有服务到 bin/ 目录"
	@echo "  make ci      - 本地 CI 检查（lint + test + build）"
	@echo ""
	@echo "数据库管理 (Podman):"
	@echo "  make db-up     - 启动所有数据库容器"
	@echo "  make db-down   - 停止所有数据库容器"
	@echo "  make db-status - 查看数据库容器状态"
	@echo "  make db-test   - 测试数据库连接"
	@echo "  make db-init   - 初始化数据库表结构"
	@echo "  make db-backup - 备份数据库"
	@echo "  make db-restore FILE=<path> - 恢复数据库"
	@echo ""
	@echo "镜像构建:"
	@echo "  make docker-build - 构建 Docker 镜像"

lint:
	golangci-lint run ./...

test:
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 测试完成，覆盖率报告: coverage.html"

build:
	go build -ldflags="-s -w" -o bin/gate cmd/gate/main.go
	@echo "✅ 编译完成，产物在 bin/ 目录"

ci:
	@echo "🔄 开始 CI 检查..."
	make lint
	make test
	make build
	@echo "✅ CI 检查通过！"

db-up:
	podman-compose -f deployments/docker-compose.yml up -d

db-down:
	podman-compose -f deployments/docker-compose.yml down

db-status:
	podman-compose -f deployments/docker-compose.yml ps

db-test:
	./scripts/test-db.sh

db-init:
	./scripts/init-db.sh

db-backup:
	./scripts/backup-db.sh

db-restore:
	./scripts/restore-db.sh $(FILE)

docker-build:
	podman build -t omega-server:latest .
