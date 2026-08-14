package network

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"
)

// silentLogger 静默日志器，避免测试输出噪声
type silentLogger struct{}

func (silentLogger) Infof(string, ...any) {}

func (silentLogger) Errorf(string, ...any) {}

// readPacket 从连接读取一个完整消息包
func readPacket(t *testing.T, conn net.Conn) (uint32, []byte) {
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
	msgID, body, err := Decode(packet)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	return msgID, body
}

func TestSessionRoundTrip(t *testing.T) {
	server, client := net.Pipe()
	defer func() { _ = server.Close(); _ = client.Close() }()

	sess := NewSession(server)
	sess.SetLogger(silentLogger{})
	sess.SetOnMessage(func(msgID uint32, body []byte) {
		// 回显消息，ID 加 1
		sess.SendMessage(msgID+1, body)
	})
	go sess.Run()

	msgID, body := uint32(1001), []byte("hello")
	if _, err := client.Write(Encode(msgID, body)); err != nil {
		t.Fatalf("发送请求失败: %v", err)
	}

	replyID, replyBody := readPacket(t, client)
	if replyID != msgID+1 {
		t.Errorf("响应 ID 不匹配: 期望 %d, 实际 %d", msgID+1, replyID)
	}
	if string(replyBody) != string(body) {
		t.Errorf("响应体不匹配: 期望 %q, 实际 %q", body, replyBody)
	}
}

func TestSessionPartialPacket(t *testing.T) {
	server, client := net.Pipe()
	defer func() { _ = server.Close(); _ = client.Close() }()

	got := make(chan []byte, 1)
	sess := NewSession(server)
	sess.SetLogger(silentLogger{})
	sess.SetOnMessage(func(_ uint32, body []byte) { got <- body })
	go sess.Run()

	msgID, body := uint32(2002), []byte("test data")
	packet := Encode(msgID, body)

	// 先发半包（仅包头），再发剩余部分，验证缓冲区拼接
	if _, err := client.Write(packet[:8]); err != nil {
		t.Fatalf("发送半包失败: %v", err)
	}
	if _, err := client.Write(packet[8:]); err != nil {
		t.Fatalf("发送剩余部分失败: %v", err)
	}

	select {
	case b := <-got:
		if string(b) != string(body) {
			t.Errorf("消息体不匹配: 期望 %q, 实际 %q", body, b)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("等待完整消息超时")
	}
}

func TestSessionInvalidPacketLength(t *testing.T) {
	tests := []struct {
		name     string
		totalLen uint32
	}{
		{name: "长度超过上限", totalLen: MaxPacketSize + 1},
		{name: "长度小于包头", totalLen: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, client := net.Pipe()
			defer func() { _ = server.Close(); _ = client.Close() }()

			sess := NewSession(server)
			sess.SetLogger(silentLogger{})
			go sess.Run()

			header := make([]byte, 8)
			binary.BigEndian.PutUint32(header[:4], tt.totalLen)
			if _, err := client.Write(header); err != nil {
				t.Fatalf("发送非法包头失败: %v", err)
			}

			// 服务端应识别非法长度并关闭连接，客户端读到 EOF
			buf := make([]byte, 1)
			if _, err := client.Read(buf); err == nil {
				t.Error("期望连接被关闭，实际读取成功")
			}
		})
	}
}

func TestSessionSendAfterClose(_ *testing.T) {
	server, client := net.Pipe()
	defer func() { _ = client.Close() }()

	sess := NewSession(server)
	sess.Close()

	// 关闭后 Send 应为空操作，不 panic 不阻塞
	sess.Send(Encode(1, []byte("x")))
	sess.SendMessage(1, []byte("x"))
}
