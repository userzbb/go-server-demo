# Omega Server 架构设计文档

## 1. 系统概述

Omega Server 是一个基于 Go 构建的分布式游戏服务器框架，采用微服务架构，支持水平扩展和高并发。

### 1.1 设计目标
- 支持 TCP 长连接，每服务器可承载 10000+ 并发玩家
- 模块化设计，各服务可独立部署和扩展
- 高性能：核心游戏逻辑采用无锁/轻锁设计
- 可观测性：结构化日志 + 指标采集

### 1.2 技术栈

| 组件 | 选型 | 说明 |
|------|------|------|
| 语言 | Go 1.26+ | 高性能、并发友好 |
| 网络 | 原生 net 包 | 轻量、可控 |
| 配置 | Viper | 支持 YAML + 环境变量 |
| 日志 | Zap | 高性能结构化日志 |
| 数据库 | PostgreSQL | 核心数据持久化 |
| 缓存 | Redis | 会话缓存、实时数据 |
| 部署 | Docker + K8s | 容器化编排 |

## 2. 整体架构图

客户端 (TCP长连接) -> 网关服务 (Gate) -> 逻辑服务 (Logic) -> 中心服务 (Center) -> PostgreSQL / Redis / MongoDB

## 3. 模块职责说明

### 3.1 网关服务（Gate）- cmd/gate/
- 连接管理：维护所有客户端 TCP 连接，管理 Session 生命周期
- 消息路由：解析消息包，根据 MsgID 分发给对应处理器
- 心跳检测：定期检测客户端存活状态，超时自动断开
- 限流防攻击：限制单 IP 连接数，防止恶意流量

### 3.2 逻辑服务（Logic）- cmd/logic/
- 房间管理：创建/加入/离开房间，房间状态维护
- 战斗计算：伤害公式、技能效果、AI 行为
- 状态同步：玩家位置/血量/状态广播给同房间玩家

### 3.3 中心服务（Center）- cmd/center/
- 匹配系统：将等待玩家匹配到一起，创建新房间
- 排行榜：维护全服玩家排名数据
- 全局数据：维护全服公告、配置等公共数据

## 4. 数据流向

### 4.1 客户端请求处理流程
1. 客户端 -> 网关：发送 TCP 包（长度 + MsgID + Body）
2. 网关 -> 解析 -> 路由：根据 MsgID 查找对应 Handler
3. 网关 -> 逻辑服务：需要计算的请求转发给 Logic 服务
4. 逻辑服务 -> 处理 -> 响应：计算结果通过网关返回给客户端
5. 网关 -> 客户端：响应包（长度 + MsgID + Body）

### 4.2 数据持久化流程
1. 玩家登录 -> 从 PostgreSQL 加载数据
2. 玩家在线 -> 数据缓存在 Redis
3. 数据变更 -> 写入 Redis + 异步写入 PostgreSQL
4. 玩家下线 -> 强制写入 PostgreSQL + 清理 Redis

## 5. 目录结构与模块映射

omega-server/
├── cmd/
│   ├── gate/          网关服务入口
│   ├── logic/         逻辑服务入口
│   └── center/        中心服务入口
├── internal/
│   ├── handler/       消息处理器
│   ├── model/         数据模型（Player, Room）
│   ├── service/       业务服务（房间、战斗、匹配）
│   ├── repository/    数据访问层（DB 操作）
│   └── core/          核心引擎
├── pkg/
│   ├── network/       网络层（已实现）
│   ├── protocol/      消息协议定义
│   ├── logger/        日志封装
│   └── util/          工具函数
├── api/proto/         Protobuf 定义
└── configs/           配置文件

## 6. 服务间通信

### 6.1 通信方式
- 网关 -> 逻辑服务：gRPC（内部 RPC）
- 逻辑服务 -> 中心服务：gRPC
- 异步任务：消息队列（Redis Stream）

### 6.2 接口定义

Gate -> Logic: ProcessMessage(PlayerID, MsgID, Body) 转发客户端消息
Logic -> Gate: SendToClient(PlayerID, MsgID, Body) 推送消息给客户端
Logic -> Center: MatchRequest(PlayerID, RoomType) 请求匹配
Center -> Logic: RoomCreated(RoomID, Players) 匹配完成通知

## 7. 部署架构

### 7.1 开发环境（Docker Compose）
- Gate（1 个实例）
- Logic（1 个实例）
- Center（1 个实例）
- PostgreSQL（1 个容器）
- Redis（1 个容器）
- MongoDB（1 个容器）

### 7.2 生产环境（K8s）
- Gate：根据 CPU 使用率扩缩容（HPA）
- Logic：根据 CPU 使用率扩缩容
- Center：固定 2 个实例保证高可用
- PostgreSQL：主从复制
- Redis：哨兵模式 / 集群模式

## 8. 性能指标

单网关连接数: 10,000+
消息吞吐量: 50,000 msg/s
消息延迟: < 50ms（P99）
可用性: 99.9%

## 9. 安全设计
- 连接层：支持 TLS/SSL 加密（生产环境开启）
- 鉴权：登录后返回 Token，后续请求携带 Token 验证
- 限流：单 IP 最大连接数限制，单用户消息频率限制
- 防刷：关键操作添加频率限制

## 10. 日志与监控

### 10.1 日志规范
- 格式：JSON 结构化日志
- 级别：Debug / Info / Warn / Error
- 必含字段：trace_id、service、time、level、msg

### 10.2 监控指标（Prometheus）
- 业务指标：在线人数、房间数、匹配成功率
- 系统指标：CPU、内存、Goroutine 数量
- 网络指标：连接数、消息 QPS、延迟分布

## 11. 演进路线

v0.1 MVP: 网关 + 登录 + 基础通信
v0.2 Alpha: 房间管理 + 状态同步
v0.3 Beta: 战斗逻辑 + 匹配系统
v1.0 GA: 完整游戏循环 + 部署文档
