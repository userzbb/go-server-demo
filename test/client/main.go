// Package main 是测试客户端：验证注册、登录与心跳通信流程
package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"omega-server/pkg/network"
	"omega-server/pkg/protocol"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}
	fmt.Println("✅ 全部流程验证通过")
}

// run 执行完整验证流程：连接 → 登录/注册 → 心跳
func run() error {
	addr := "127.0.0.1:8888"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer func() { _ = conn.Close() }()
	fmt.Println("连接服务器成功:", addr)

	// 先尝试登录，账号不存在则注册
	if err := loginOrRegister(conn, "player1", "123456"); err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	// 心跳检测
	return sendHeartbeat(conn)
}

// loginOrRegister 先登录，收到"用户名或密码错误"则注册后重试
func loginOrRegister(conn net.Conn, username, password string) error {
	code, message, err := doLogin(conn, username, password)
	if err != nil {
		return err
	}
	if code == protocol.CodeOK {
		fmt.Printf("✅ 登录成功: %s\n", message)
		return nil
	}
	if code != protocol.CodeInvalidCredentials {
		return fmt.Errorf("登录失败: code=%d %s", code, message)
	}

	fmt.Println("账号不存在，尝试注册...")
	code, message, err = doRegister(conn, username, password)
	if err != nil {
		return err
	}
	if code != protocol.CodeOK {
		return fmt.Errorf("注册失败: code=%d %s", code, message)
	}
	fmt.Printf("✅ 注册成功: %s\n", message)

	code, message, err = doLogin(conn, username, password)
	if err != nil {
		return err
	}
	if code != protocol.CodeOK {
		return fmt.Errorf("注册后登录失败: code=%d %s", code, message)
	}
	fmt.Printf("✅ 登录成功: %s\n", message)
	return nil
}

// doLogin 发送登录请求并解析响应
func doLogin(conn net.Conn, username, password string) (int, string, error) {
	replyID, replyBody, err := request(conn, protocol.MsgIDLoginRequest, protocol.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return 0, "", fmt.Errorf("登录请求: %w", err)
	}
	var resp protocol.LoginResponse
	if err := json.Unmarshal(replyBody, &resp); err != nil {
		return 0, "", fmt.Errorf("解析登录响应 [ID=%d]: %w", replyID, err)
	}
	if replyID != protocol.MsgIDLoginResponse {
		return 0, "", fmt.Errorf("登录响应 ID 不匹配: %d", replyID)
	}
	return resp.Code, resp.Message, nil
}

// doRegister 发送注册请求并解析响应
func doRegister(conn net.Conn, username, password string) (int, string, error) {
	replyID, replyBody, err := request(conn, protocol.MsgIDRegisterRequest, protocol.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return 0, "", fmt.Errorf("注册请求: %w", err)
	}
	var resp protocol.LoginResponse
	if err := json.Unmarshal(replyBody, &resp); err != nil {
		return 0, "", fmt.Errorf("解析注册响应 [ID=%d]: %w", replyID, err)
	}
	if replyID != protocol.MsgIDRegisterResponse {
		return 0, "", fmt.Errorf("注册响应 ID 不匹配: %d", replyID)
	}
	return resp.Code, resp.Message, nil
}

// sendHeartbeat 发送心跳请求并校验响应时间戳
func sendHeartbeat(conn net.Conn) error {
	now := time.Now().Unix()
	replyID, replyBody, err := request(conn, protocol.MsgIDHeartbeatRequest, protocol.HeartbeatRequest{
		Timestamp: now,
	})
	if err != nil {
		return fmt.Errorf("心跳请求: %w", err)
	}
	var resp protocol.HeartbeatResponse
	if err := json.Unmarshal(replyBody, &resp); err != nil {
		return fmt.Errorf("解析心跳响应 [ID=%d]: %w", replyID, err)
	}
	if replyID != protocol.MsgIDHeartbeatResponse {
		return fmt.Errorf("心跳响应 ID 不匹配: %d", replyID)
	}
	if resp.Timestamp != now {
		return fmt.Errorf("心跳时间戳不匹配: 期望 %d, 实际 %d", now, resp.Timestamp)
	}
	fmt.Printf("✅ 心跳正常: server_time=%d\n", resp.ServerTime)
	return nil
}

// request 发送一条 JSON 请求并返回原始响应包
func request(conn net.Conn, msgID uint32, req any) (uint32, []byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return 0, nil, fmt.Errorf("编码请求: %w", err)
	}
	if _, err := conn.Write(network.Encode(msgID, body)); err != nil {
		return 0, nil, fmt.Errorf("发送请求: %w", err)
	}
	return readPacket(conn)
}

// readPacket 读取一个完整消息包，处理半包与粘包
func readPacket(conn net.Conn) (uint32, []byte, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(conn, header); err != nil {
		return 0, nil, fmt.Errorf("读取响应头: %w", err)
	}

	totalLen := int(binary.BigEndian.Uint32(header[:4]))
	if totalLen < 8 || totalLen > network.MaxPacketSize {
		return 0, nil, fmt.Errorf("非法包长度: %d", totalLen)
	}

	packet := make([]byte, totalLen)
	copy(packet, header)
	if _, err := io.ReadFull(conn, packet[8:]); err != nil {
		return 0, nil, fmt.Errorf("读取响应体: %w", err)
	}
	return network.Decode(packet)
}
