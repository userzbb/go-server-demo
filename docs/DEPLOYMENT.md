# Omega Server 部署指南

## 开发环境（本地运行）
启动数据库: make db-up && make db-init
运行网关: make run

## 容器化部署（Podman）
构建镜像: podman build -t omega-server:latest .
运行容器: podman run -d --name omega-gate -p 8888:8888 -v ./configs:/app/configs omega-server:latest
查看日志: podman logs -f omega-gate

## 环境变量
DB_PASSWORD, JWT_SECRET, LOG_LEVEL

## 健康检查
curl http://localhost:8888/health
