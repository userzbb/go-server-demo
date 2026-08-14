# 变更日志

本文档记录项目的所有重要变更，遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/) 规范。

## [Unreleased]

### Added
- 初始化项目骨架，采用 Go 标准目录结构
- 网络层实现（pkg/network）：支持 TCP 粘包拆包、Session 管理
- 测试客户端（test/client）：用于验证服务器通信
- 网关服务（cmd/gate）：基础 TCP 服务器
- 逻辑服务（cmd/logic）：占位入口
- 中心服务（cmd/center）：占位入口
- 开发规范文档（docs/STANDARDS.md）
- 架构设计文档（docs/ARCHITECTURE.md）
- 通信协议文档（docs/PROTOCOL.md）
- 数据规范文档（docs/SCHEMA.md）
- 部署指南文档（docs/DEPLOYMENT.md）
- Dockerfile + .dockerignore
- docker-compose.yml（PostgreSQL + MongoDB + Redis）
- CI 流水线（.github/workflows/ci.yml）
- 代码质量检查（.golangci.yml + Pre-commit）
- 构建脚本（scripts/build.sh）
- 数据库脚本（init-db.sh、test-db.sh、backup-db.sh、restore-db.sh）
- Makefile 命令集
- .editorconfig 统一编辑器配置
- LICENSE（MIT 许可证）

### Changed

### Fixed
