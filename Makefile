.PHONY: help lint test build ci

help:
	@echo "可用命令："
	@echo "  make lint    - 运行代码质量检查"
	@echo "  make test    - 运行单元测试并生成覆盖率报告"
	@echo "  make build   - 编译所有服务到 bin/ 目录"
	@echo "  make ci      - 本地 CI 检查（lint + test + build）"

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
