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

### Added（v0.1 MVP 第一阶段）
- 配置加载（internal/config）：Viper + 环境变量覆盖，JWT 密钥仅允许环境变量注入
- 结构化日志（internal/logger）：Zap 封装，支持级别/格式/服务名配置
- 玩家模型与 PostgreSQL 数据访问层（internal/model + internal/repository，pgx）
- 注册/登录鉴权（internal/service）：bcrypt 密码哈希 + JWT HS256 令牌
- 消息路由框架（internal/handler）：按消息 ID 分发，登录/注册/心跳处理器
- 网关正式接线（cmd/gate）：配置、日志、数据库、路由全链路
- 协议新增注册消息（1003/1004）与错误码 7（用户名已被占用）
- 网络层支持日志器注入（pkg/network.Logger）

### Added（v0.2 Alpha）
- 房间模型与内存房间管理器（internal/model.Room + internal/service.RoomManager）
- 会话玩家绑定（pkg/network.Session.PlayerID/SetPlayerID）与会话注册表（SessionRegistry）
- 房间处理器：创建（2001/2002）、加入（2003/2004）、离开（2005/2006）
- 移动同步（internal/handler.MoveHandler）：3001 请求 → 3002 广播给同房间其他玩家
- 登录成功钩子：会话绑定身份并注册，断开连接自动离开房间
- 协议新增 2005/2006、3001/3002 消息定义与状态同步广播

### Changed

### Fixed
