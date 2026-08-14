Omega Server 开发者入门指南

本文档为 AI 及新开发者提供项目快速概览。

1. 项目概述
Omega Server 是基于 Go 的分布式游戏服务器，支持 TCP 长连接和多服务拆分。

2. 技术栈
- Go 1.26+
- 原生 net 包
- PostgreSQL（核心数据）
- MongoDB（灵活数据）
- Redis（缓存）
- Podman / Docker
- Kubernetes

3. 目录结构
cmd/          服务入口（gate/logic/center）
internal/     私有业务代码（handler/model/service/repository/core）
pkg/          公共库（network/protocol/logger/util）
api/proto/    Protobuf 协议定义
configs/      配置文件
docs/         所有文档
scripts/      构建和运维脚本
test/         测试工具
deployments/  容器和 K8s 部署文件

4. 模块依赖关系
- pkg/network 被 cmd/gate、internal/handler 和 test/client 使用
- pkg/protocol 被 internal/handler、internal/service、cmd/gate 和 test/client 使用
- internal/config 被 cmd/gate 使用
- internal/logger 被 cmd/gate 使用
- internal/model 被 internal/repository 和 internal/service 使用
- internal/repository 依赖 model 和数据库，被 internal/service 和 cmd/gate 使用
- internal/service 依赖 model、repository、protocol
- internal/handler 依赖 service、protocol、network
- cmd/gate 依赖 config、logger、repository、service、handler、network、protocol

5. 常用命令
make lint       代码检查
make test       单元测试
make build      编译所有服务
make ci         完整 CI 检查
make db-up      启动数据库容器
make db-init    初始化数据库
make docker-build  构建镜像

6. 环境变量
参考 .env.example 文件

7. 开发流程
- 从 develop 分支切出 feature/xxx 分支
- 编写代码并补充测试
- 运行 make ci 确保检查通过
- 提交并推送，合并回 develop

8. 文档索引
开发规范：docs/STANDARDS.md
架构设计：docs/ARCHITECTURE.md
通信协议：docs/PROTOCOL.md
数据规范：docs/SCHEMA.md
部署指南：docs/DEPLOYMENT.md
API示例：docs/API_EXAMPLES.md
