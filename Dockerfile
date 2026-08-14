# 构建阶段
FROM docker.io/library/golang:1.26-alpine AS builder

WORKDIR /app

# 复制依赖文件并下载
COPY go.mod go.sum* ./
RUN go mod download

# 复制源代码并编译网关服务
COPY . .
RUN go build -ldflags="-s -w" -o /gate cmd/gate/main.go

# 运行阶段
FROM docker.io/library/alpine:latest

# 安装 CA 证书（用于 HTTPS 请求）
RUN apk --no-cache add ca-certificates

WORKDIR /app

# 复制编译产物和配置文件
COPY --from=builder /gate /app/gate
COPY configs/ /app/configs/

# 暴露端口
EXPOSE 8888

# 启动网关服务
ENTRYPOINT ["/app/gate"]
