# Omega Server

高性能分布式游戏服务器

## 快速开始
- Go 1.26+
- Make
- golangci-lint

### 运行服务器
```bash
make db-up && make db-init   # 启动数据库容器并初始化表结构
export JWT_SECRET=dev-secret # JWT 密钥必填，否则网关拒绝启动
make run
```

### 运行测试
```bash
make test
```

### 完整 CI 检查
```bash
make ci
```

## 开发规范
详见 [docs/STANDARDS.md](docs/STANDARDS.md)
