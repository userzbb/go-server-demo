package handler

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"omega-server/internal/service"
	"omega-server/pkg/network"
	"omega-server/pkg/protocol"
)

// startBoundServer 启动一个绑定指定玩家 ID 的会话，返回客户端连接与会话
func startBoundServer(t *testing.T, router *Router, playerID string) (net.Conn, *network.Session) {
	t.Helper()
	server, client := net.Pipe()

	sess := network.NewSession(server)
	sess.SetLogger(silentLogger{})
	if playerID != "" {
		sess.SetPlayerID(playerID)
	}
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
	return client, sess
}

// readReply 从连接读取一个完整消息包（不发送请求）
func readReply(t *testing.T, conn net.Conn) (uint32, []byte) {
	t.Helper()
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
	msgID, body, err := network.Decode(packet)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	return msgID, body
}

// newRoomRouter 创建带房间/移动处理器的路由表
func newRoomRouter() (*Router, *service.RoomManager, *SessionRegistry) {
	roomMgr := service.NewRoomManager()
	registry := NewSessionRegistry()
	roomHandler := NewRoomHandler(roomMgr, registry)
	moveHandler := NewMoveHandler(roomMgr, registry)

	router := NewRouter()
	router.Register(protocol.MsgIDCreateRoomRequest, roomHandler.HandleCreate)
	router.Register(protocol.MsgIDJoinRoomRequest, roomHandler.HandleJoin)
	router.Register(protocol.MsgIDLeaveRoomRequest, roomHandler.HandleLeave)
	router.Register(protocol.MsgIDMoveSyncRequest, moveHandler.Handle)
	return router, roomMgr, registry
}

func TestCreateRoomFlow(t *testing.T) {
	router, roomMgr, registry := newRoomRouter()
	client, sess := startBoundServer(t, router, "p1")

	replyID, replyBody := clientRequest(t, client, protocol.MsgIDCreateRoomRequest,
		protocol.CreateRoomRequest{RoomName: "我的房间", MaxPlayers: 10})

	if replyID != protocol.MsgIDCreateRoomResponse {
		t.Fatalf("响应 ID 不匹配: %d", replyID)
	}
	var resp protocol.CreateRoomResponse
	if err := json.Unmarshal(replyBody, &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if resp.Code != protocol.CodeOK {
		t.Fatalf("创建失败: code=%d %s", resp.Code, resp.Message)
	}
	if resp.RoomID == "" {
		t.Error("响应缺少 roomId")
	}

	if _, err := roomMgr.GetRoom(resp.RoomID); err != nil {
		t.Errorf("房间应已创建: %v", err)
	}
	if sess.PlayerID() != "p1" {
		t.Errorf("会话玩家绑定异常: %q", sess.PlayerID())
	}
	if _, ok := registry.Get("p1"); !ok {
		t.Error("p1 应已注册到会话注册表")
	}
}

func TestJoinRoomFlow(t *testing.T) {
	router, _, _ := newRoomRouter()
	clientA, _ := startBoundServer(t, router, "p1")

	_, replyBody := clientRequest(t, clientA, protocol.MsgIDCreateRoomRequest,
		protocol.CreateRoomRequest{RoomName: "房间", MaxPlayers: 10})
	var createResp protocol.CreateRoomResponse
	if err := json.Unmarshal(replyBody, &createResp); err != nil {
		t.Fatalf("解析创建响应失败: %v", err)
	}

	clientB, _ := startBoundServer(t, router, "p2")
	replyID, joinBody := clientRequest(t, clientB, protocol.MsgIDJoinRoomRequest,
		protocol.JoinRoomRequest{RoomID: createResp.RoomID})

	if replyID != protocol.MsgIDJoinRoomResponse {
		t.Fatalf("响应 ID 不匹配: %d", replyID)
	}
	var joinResp protocol.JoinRoomResponse
	if err := json.Unmarshal(joinBody, &joinResp); err != nil {
		t.Fatalf("解析加入响应失败: %v", err)
	}
	if joinResp.Code != protocol.CodeOK {
		t.Fatalf("加入失败: code=%d %s", joinResp.Code, joinResp.Message)
	}
	if len(joinResp.Players) != 2 || joinResp.Players[0] != "p1" || joinResp.Players[1] != "p2" {
		t.Errorf("成员列表不匹配: %v", joinResp.Players)
	}
}

func TestJoinRoomErrors(t *testing.T) {
	router, _, _ := newRoomRouter()

	t.Run("房间不存在", func(t *testing.T) {
		client, _ := startBoundServer(t, router, "p1")
		_, body := clientRequest(t, client, protocol.MsgIDJoinRoomRequest,
			protocol.JoinRoomRequest{RoomID: "room_999"})
		var resp protocol.JoinRoomResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}
		if resp.Code != protocol.CodeRoomNotFound {
			t.Errorf("期望 code=%d, 实际 %d", protocol.CodeRoomNotFound, resp.Code)
		}
	})

	t.Run("房间已满", func(t *testing.T) {
		clientA, _ := startBoundServer(t, router, "p1")
		_, body := clientRequest(t, clientA, protocol.MsgIDCreateRoomRequest,
			protocol.CreateRoomRequest{RoomName: "满员房", MaxPlayers: 2})
		var createResp protocol.CreateRoomResponse
		if err := json.Unmarshal(body, &createResp); err != nil {
			t.Fatalf("解析创建响应失败: %v", err)
		}

		clientB, _ := startBoundServer(t, router, "p2")
		_, _ = clientRequest(t, clientB, protocol.MsgIDJoinRoomRequest,
			protocol.JoinRoomRequest{RoomID: createResp.RoomID})

		clientC, _ := startBoundServer(t, router, "p3")
		_, joinBody := clientRequest(t, clientC, protocol.MsgIDJoinRoomRequest,
			protocol.JoinRoomRequest{RoomID: createResp.RoomID})
		var resp protocol.JoinRoomResponse
		if err := json.Unmarshal(joinBody, &resp); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}
		if resp.Code != protocol.CodeRoomFull {
			t.Errorf("期望 code=%d, 实际 %d", protocol.CodeRoomFull, resp.Code)
		}
	})

	t.Run("重复加入", func(t *testing.T) {
		clientA, _ := startBoundServer(t, router, "p1")
		_, body := clientRequest(t, clientA, protocol.MsgIDCreateRoomRequest,
			protocol.CreateRoomRequest{RoomName: "房间", MaxPlayers: 4})
		var createResp protocol.CreateRoomResponse
		if err := json.Unmarshal(body, &createResp); err != nil {
			t.Fatalf("解析创建响应失败: %v", err)
		}

		_, joinBody := clientRequest(t, clientA, protocol.MsgIDJoinRoomRequest,
			protocol.JoinRoomRequest{RoomID: createResp.RoomID})
		var resp protocol.JoinRoomResponse
		if err := json.Unmarshal(joinBody, &resp); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}
		if resp.Code != protocol.CodeAlreadyInRoom {
			t.Errorf("期望 code=%d, 实际 %d", protocol.CodeAlreadyInRoom, resp.Code)
		}
	})

	t.Run("未登录", func(t *testing.T) {
		client, _ := startBoundServer(t, router, "")
		_, body := clientRequest(t, client, protocol.MsgIDCreateRoomRequest,
			protocol.CreateRoomRequest{RoomName: "房间", MaxPlayers: 10})
		var resp protocol.CreateRoomResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			t.Fatalf("解析响应失败: %v", err)
		}
		if resp.Code != protocol.CodeInvalidToken {
			t.Errorf("期望 code=%d, 实际 %d", protocol.CodeInvalidToken, resp.Code)
		}
	})
}

func TestMoveBroadcast(t *testing.T) {
	router, _, _ := newRoomRouter()

	// p1 建房，p2 加入
	clientA, _ := startBoundServer(t, router, "p1")
	_, body := clientRequest(t, clientA, protocol.MsgIDCreateRoomRequest,
		protocol.CreateRoomRequest{RoomName: "对战房", MaxPlayers: 10})
	var createResp protocol.CreateRoomResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		t.Fatalf("解析创建响应失败: %v", err)
	}

	clientB, _ := startBoundServer(t, router, "p2")
	_, joinBody := clientRequest(t, clientB, protocol.MsgIDJoinRoomRequest,
		protocol.JoinRoomRequest{RoomID: createResp.RoomID})
	var joinResp protocol.JoinRoomResponse
	if err := json.Unmarshal(joinBody, &joinResp); err != nil {
		t.Fatalf("解析加入响应失败: %v", err)
	}
	if joinResp.Code != protocol.CodeOK {
		t.Fatalf("加入失败: %s", joinResp.Message)
	}

	// p1 发送移动同步（只写不读，p1 不应收到广播）
	moveBody, err := json.Marshal(protocol.MoveSyncRequest{X: 1.5, Y: 2.5, VX: 0.1, VY: -0.1})
	if err != nil {
		t.Fatalf("编码移动请求失败: %v", err)
	}
	if _, err := clientA.Write(network.Encode(protocol.MsgIDMoveSyncRequest, moveBody)); err != nil {
		t.Fatalf("发送移动请求失败: %v", err)
	}

	// p2 应收到 3002 状态广播
	replyID, stateBody := readReply(t, clientB)
	if replyID != protocol.MsgIDStateSyncBroadcast {
		t.Fatalf("广播消息 ID 不匹配: %d", replyID)
	}
	var state protocol.StateSync
	if err := json.Unmarshal(stateBody, &state); err != nil {
		t.Fatalf("解析状态广播失败: %v", err)
	}
	if state.PlayerID != "p1" {
		t.Errorf("广播玩家 ID 不匹配: %q", state.PlayerID)
	}
	if state.X != 1.5 || state.Y != 2.5 || state.VX != 0.1 || state.VY != -0.1 {
		t.Errorf("位置数据不匹配: %+v", state)
	}
}

func TestLoginBindsSession(t *testing.T) {
	auth := service.NewAuthService(newFakePlayerRepo(), "test-secret", time.Hour)
	registry := NewSessionRegistry()
	loginHandler := NewLoginHandler(auth)
	loginHandler.SetOnLogin(func(sess *network.Session, playerID string) {
		sess.SetPlayerID(playerID)
		registry.Register(playerID, sess)
	})

	router := NewRouter()
	router.Register(protocol.MsgIDRegisterRequest, NewRegisterHandler(auth).Handle)
	router.Register(protocol.MsgIDLoginRequest, loginHandler.Handle)

	client, sess := startBoundServer(t, router, "")

	// 先注册，再登录
	_, regBody := clientRequest(t, client, protocol.MsgIDRegisterRequest,
		protocol.LoginRequest{Username: "player1", Password: "123456"})
	var regResp protocol.LoginResponse
	if err := json.Unmarshal(regBody, &regResp); err != nil {
		t.Fatalf("解析注册响应失败: %v", err)
	}
	if regResp.Code != protocol.CodeOK {
		t.Fatalf("注册失败: code=%d %s", regResp.Code, regResp.Message)
	}

	_, loginBody := clientRequest(t, client, protocol.MsgIDLoginRequest,
		protocol.LoginRequest{Username: "player1", Password: "123456"})
	var resp protocol.LoginResponse
	if err := json.Unmarshal(loginBody, &resp); err != nil {
		t.Fatalf("解析登录响应失败: %v", err)
	}
	if resp.Code != protocol.CodeOK {
		t.Fatalf("登录失败: code=%d %s", resp.Code, resp.Message)
	}

	if sess.PlayerID() != resp.Player.ID {
		t.Errorf("会话绑定玩家 ID 不匹配: %q != %q", sess.PlayerID(), resp.Player.ID)
	}
	if _, ok := registry.Get(resp.Player.ID); !ok {
		t.Error("登录后玩家应注册到会话注册表")
	}
}
