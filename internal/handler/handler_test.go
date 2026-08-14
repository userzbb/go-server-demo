package handler

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"omega-server/internal/model"
	"omega-server/internal/repository"
	"omega-server/internal/service"
	"omega-server/pkg/network"
	"omega-server/pkg/protocol"
)

// fakePlayerRepo 内存版玩家仓库，仅用于测试
type fakePlayerRepo struct {
	players map[string]*model.Player
}

func newFakePlayerRepo() *fakePlayerRepo {
	return &fakePlayerRepo{players: map[string]*model.Player{}}
}

func (f *fakePlayerRepo) Create(_ context.Context, p *model.Player) (*model.Player, error) {
	if _, ok := f.players[p.Username]; ok {
		return nil, repository.ErrUsernameExists
	}
	p.ID = "player-" + p.Username
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	f.players[p.Username] = p
	return p, nil
}

func (f *fakePlayerRepo) GetByUsername(_ context.Context, username string) (*model.Player, error) {
	p, ok := f.players[username]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return p, nil
}

func (f *fakePlayerRepo) UpdateLastLogin(_ context.Context, _ string) error {
	return nil
}

// silentLogger 静默日志器，避免测试输出噪声
type silentLogger struct{}

func (silentLogger) Infof(string, ...any) {}

func (silentLogger) Errorf(string, ...any) {}

// newAuthRouter 创建带完整认证路由的路由表（每个调用使用独立的仓库）
func newAuthRouter() *Router {
	auth := service.NewAuthService(newFakePlayerRepo(), "test-secret", time.Hour)
	router := NewRouter()
	router.Register(protocol.MsgIDLoginRequest, NewLoginHandler(auth).Handle)
	router.Register(protocol.MsgIDRegisterRequest, NewRegisterHandler(auth).Handle)
	router.Register(protocol.MsgIDHeartbeatRequest, NewHeartbeatHandler().Handle)
	return router
}

// startServer 在 net.Pipe 上启动服务端 Session，返回客户端连接
func startServer(t *testing.T, router *Router) net.Conn {
	t.Helper()
	server, client := net.Pipe()

	sess := network.NewSession(server)
	sess.SetLogger(silentLogger{})
	sess.SetOnMessage(func(msgID uint32, body []byte) {
		if err := router.Dispatch(sess, msgID, body); err != nil {
			t.Errorf("Dispatch 失败: %v", err)
		}
	})
	go sess.Run()

	t.Cleanup(func() {
		_ = server.Close()
		_ = client.Close()
	})
	return client
}

// clientRequestRaw 发送原始请求体并读取完整响应包
func clientRequestRaw(t *testing.T, conn net.Conn, msgID uint32, body []byte) (uint32, []byte) {
	t.Helper()
	if _, err := conn.Write(network.Encode(msgID, body)); err != nil {
		t.Fatalf("发送请求失败: %v", err)
	}

	header := make([]byte, 8)
	if _, err := io.ReadFull(conn, header); err != nil {
		t.Fatalf("读取响应头失败: %v", err)
	}
	totalLen := int(binary.BigEndian.Uint32(header[:4]))
	packet := make([]byte, totalLen)
	copy(packet, header)
	if _, err := io.ReadFull(conn, packet[8:]); err != nil {
		t.Fatalf("读取响应体失败: %v", err)
	}

	replyID, replyBody, err := network.Decode(packet)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	return replyID, replyBody
}

// clientRequest 发送 JSON 请求并返回原始响应
func clientRequest(t *testing.T, conn net.Conn, msgID uint32, req any) (uint32, []byte) {
	t.Helper()
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("编码请求失败: %v", err)
	}
	return clientRequestRaw(t, conn, msgID, body)
}

func TestRouterDispatch(t *testing.T) {
	router := NewRouter()
	called := false
	router.Register(1234, func(_ *network.Session, _ []byte) error {
		called = true
		return nil
	})

	server, client := net.Pipe()
	defer func() { _ = server.Close(); _ = client.Close() }()
	sess := network.NewSession(server)

	if err := router.Dispatch(sess, 1234, nil); err != nil {
		t.Fatalf("分发已注册消息失败: %v", err)
	}
	if !called {
		t.Error("处理器未被调用")
	}
	if err := router.Dispatch(sess, 9999, nil); err == nil {
		t.Error("未注册消息应返回错误")
	}
}

func TestAuthFlow(t *testing.T) {
	t.Run("注册后登录成功", func(t *testing.T) {
		client := startServer(t, newAuthRouter())

		replyID, replyBody := clientRequest(t, client, protocol.MsgIDRegisterRequest,
			protocol.LoginRequest{Username: "player1", Password: "123456"})
		if replyID != protocol.MsgIDRegisterResponse {
			t.Fatalf("注册响应 ID 不匹配: %d", replyID)
		}
		var regResp protocol.LoginResponse
		if err := json.Unmarshal(replyBody, &regResp); err != nil {
			t.Fatalf("解析注册响应失败: %v", err)
		}
		if regResp.Code != protocol.CodeOK {
			t.Fatalf("注册失败: code=%d %s", regResp.Code, regResp.Message)
		}

		replyID, replyBody = clientRequest(t, client, protocol.MsgIDLoginRequest,
			protocol.LoginRequest{Username: "player1", Password: "123456"})
		if replyID != protocol.MsgIDLoginResponse {
			t.Fatalf("登录响应 ID 不匹配: %d", replyID)
		}
		var loginResp protocol.LoginResponse
		if err := json.Unmarshal(replyBody, &loginResp); err != nil {
			t.Fatalf("解析登录响应失败: %v", err)
		}
		if loginResp.Code != protocol.CodeOK {
			t.Fatalf("登录失败: code=%d %s", loginResp.Code, loginResp.Message)
		}
		if loginResp.Token == "" {
			t.Error("登录成功但 token 为空")
		}
		if loginResp.Player == nil || loginResp.Player.Username != "player1" {
			t.Errorf("玩家信息不匹配: %+v", loginResp.Player)
		}
	})

	t.Run("登录失败-用户不存在", func(t *testing.T) {
		client := startServer(t, newAuthRouter())
		_, replyBody := clientRequest(t, client, protocol.MsgIDLoginRequest,
			protocol.LoginRequest{Username: "nobody", Password: "123456"})

		var resp protocol.LoginResponse
		if err := json.Unmarshal(replyBody, &resp); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}
		if resp.Code != protocol.CodeInvalidCredentials {
			t.Errorf("期望 code=%d, 实际 %d", protocol.CodeInvalidCredentials, resp.Code)
		}
	})

	t.Run("登录失败-密码错误", func(t *testing.T) {
		client := startServer(t, newAuthRouter())
		_, _ = clientRequest(t, client, protocol.MsgIDRegisterRequest,
			protocol.LoginRequest{Username: "player1", Password: "123456"})
		_, replyBody := clientRequest(t, client, protocol.MsgIDLoginRequest,
			protocol.LoginRequest{Username: "player1", Password: "wrong"})

		var resp protocol.LoginResponse
		if err := json.Unmarshal(replyBody, &resp); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}
		if resp.Code != protocol.CodeInvalidCredentials {
			t.Errorf("期望 code=%d, 实际 %d", protocol.CodeInvalidCredentials, resp.Code)
		}
	})

	t.Run("注册失败-格式错误", func(t *testing.T) {
		client := startServer(t, newAuthRouter())
		_, replyBody := clientRequest(t, client, protocol.MsgIDRegisterRequest,
			protocol.LoginRequest{Username: "", Password: ""})

		var resp protocol.LoginResponse
		if err := json.Unmarshal(replyBody, &resp); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}
		if resp.Code != protocol.CodeBadRequest {
			t.Errorf("期望 code=%d, 实际 %d", protocol.CodeBadRequest, resp.Code)
		}
	})

	t.Run("注册失败-用户名已占用", func(t *testing.T) {
		client := startServer(t, newAuthRouter())
		_, _ = clientRequest(t, client, protocol.MsgIDRegisterRequest,
			protocol.LoginRequest{Username: "player1", Password: "123456"})
		_, replyBody := clientRequest(t, client, protocol.MsgIDRegisterRequest,
			protocol.LoginRequest{Username: "player1", Password: "654321"})

		var resp protocol.LoginResponse
		if err := json.Unmarshal(replyBody, &resp); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}
		if resp.Code != protocol.CodeUsernameTaken {
			t.Errorf("期望 code=%d, 实际 %d", protocol.CodeUsernameTaken, resp.Code)
		}
	})

	t.Run("非法 JSON 请求", func(t *testing.T) {
		client := startServer(t, newAuthRouter())
		_, replyBody := clientRequestRaw(t, client, protocol.MsgIDLoginRequest, []byte("{not-json"))

		var resp protocol.LoginResponse
		if err := json.Unmarshal(replyBody, &resp); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}
		if resp.Code != protocol.CodeBadRequest {
			t.Errorf("期望 code=%d, 实际 %d", protocol.CodeBadRequest, resp.Code)
		}
	})
}

func TestHeartbeatHandler(t *testing.T) {
	client := startServer(t, newAuthRouter())

	now := time.Now().Unix()
	replyID, replyBody := clientRequest(t, client, protocol.MsgIDHeartbeatRequest,
		protocol.HeartbeatRequest{Timestamp: now})

	if replyID != protocol.MsgIDHeartbeatResponse {
		t.Fatalf("心跳响应 ID 不匹配: %d", replyID)
	}
	var resp protocol.HeartbeatResponse
	if err := json.Unmarshal(replyBody, &resp); err != nil {
		t.Fatalf("解析心跳响应失败: %v", err)
	}
	if resp.Timestamp != now {
		t.Errorf("时间戳回显不匹配: 期望 %d, 实际 %d", now, resp.Timestamp)
	}
	if diff := resp.ServerTime - now; diff < -2 || diff > 2 {
		t.Errorf("server_time 与本地时间偏差过大: %d", resp.ServerTime)
	}
}
