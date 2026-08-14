// Package main 是网关服务入口：加载配置、初始化依赖并启动 TCP 服务器
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"time"

	"omega-server/internal/config"
	"omega-server/internal/handler"
	"omega-server/internal/logger"
	"omega-server/internal/repository"
	"omega-server/internal/service"
	"omega-server/pkg/network"
	"omega-server/pkg/protocol"
)

func main() {
	configPath := flag.String("config", "configs/dev.yaml", "配置文件路径")
	flag.Parse()

	// 启动阶段 logger 未就绪，允许使用标准库 log 输出致命错误
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if cfg.JWTSecret == "" {
		log.Fatalf("缺少 JWT_SECRET 环境变量，拒绝启动")
	}

	lg, err := logger.New(logger.Config{
		Level:   cfg.Log.Level,
		Format:  cfg.Log.Format,
		Service: cfg.Server.Name,
	})
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	playerRepo, err := repository.NewPostgresPlayerRepository(context.Background(), cfg.Postgres)
	if err != nil {
		lg.Fatalf("初始化数据库失败: %v", err)
	}
	defer playerRepo.Close()

	authSvc := service.NewAuthService(playerRepo, cfg.JWTSecret, 24*time.Hour)

	router := handler.NewRouter()
	router.Register(protocol.MsgIDLoginRequest, handler.NewLoginHandler(authSvc).Handle)
	router.Register(protocol.MsgIDRegisterRequest, handler.NewRegisterHandler(authSvc).Handle)
	router.Register(protocol.MsgIDHeartbeatRequest, handler.NewHeartbeatHandler().Handle)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		lg.Fatalf("监听 %s 失败: %v", addr, err)
	}
	defer func() { _ = listener.Close() }()
	lg.Infof("Omega Gate 服务启动成功，监听 %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			lg.Errorf("接受连接失败: %v", err)
			continue
		}

		traceID := fmt.Sprintf("%08x", rand.Uint32())
		lg.Infof("trace_id=%s 新客户端连接: %s", traceID, conn.RemoteAddr())

		sess := network.NewSession(conn)
		sess.SetLogger(lg)
		sess.SetOnMessage(func(msgID uint32, body []byte) {
			if err := router.Dispatch(sess, msgID, body); err != nil {
				lg.Errorf("trace_id=%s 消息处理失败 [ID=%d]: %v", traceID, msgID, err)
			}
		})

		go sess.Run()
	}
}
