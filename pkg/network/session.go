// Package network 提供游戏服务器的网络层功能，包括 TCP 连接管理、消息编解码和会话控制
package network

import (
	"encoding/binary"
	"log"
	"net"
	"sync"
)

// Session 管理一个客户端的 TCP 连接，负责收发消息和生命周期控制
type Session struct {
	conn      net.Conn
	sendChan  chan []byte
	wg        sync.WaitGroup
	closed    bool
	mu        sync.Mutex
	onMessage func(msgID uint32, body []byte)
	logger    Logger
	playerID  string
}

// NewSession 创建一个新的 Session 实例
func NewSession(conn net.Conn) *Session {
	return &Session{
		conn:     conn,
		sendChan: make(chan []byte, 100),
	}
}

// SetOnMessage 设置消息接收回调函数，当收到完整包时自动调用
func (s *Session) SetOnMessage(handler func(msgID uint32, body []byte)) {
	s.onMessage = handler
}

// SetPlayerID 绑定玩家 ID（登录成功后调用），重复调用覆盖旧值
func (s *Session) SetPlayerID(playerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playerID = playerID
}

// PlayerID 返回绑定的玩家 ID，未登录时为空字符串
func (s *Session) PlayerID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.playerID
}

// SetLogger 注入自定义日志器（如 Zap），未注入时回退到标准库 log
func (s *Session) SetLogger(l Logger) {
	s.logger = l
}

// logInfo 输出普通级别日志，优先使用注入的 Logger
func (s *Session) logInfo(format string, args ...any) {
	if s.logger != nil {
		s.logger.Infof(format, args...)
		return
	}
	log.Printf(format, args...)
}

// logError 输出错误级别日志，优先使用注入的 Logger
func (s *Session) logError(format string, args ...any) {
	if s.logger != nil {
		s.logger.Errorf(format, args...)
		return
	}
	log.Printf(format, args...)
}

// Send 将数据放入发送队列，由写协程异步发送
func (s *Session) Send(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	buf := make([]byte, len(data))
	copy(buf, data)
	s.sendChan <- buf
}

// SendMessage 编码并发送一条完整消息（4字节总长度 + 4字节消息ID + body）
func (s *Session) SendMessage(msgID uint32, body []byte) {
	s.Send(Encode(msgID, body))
}

// Close 关闭连接并释放资源
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.sendChan)
	_ = s.conn.Close()
}

// Run 启动读写协程，阻塞直到连接关闭
func (s *Session) Run() {
	s.wg.Add(2)
	go s.handleRead()
	go s.handleWrite()
	s.wg.Wait()
}

// handleRead 读协程：循环读取数据并解析完整包
func (s *Session) handleRead() {
	defer s.wg.Done()
	defer s.Close()

	var buffer []byte
	buf := make([]byte, 1024)

	for {
		n, err := s.conn.Read(buf)
		if err != nil {
			s.logError("读取错误: %v", err)
			break
		}

		buffer = append(buffer, buf[:n]...)

		for len(buffer) >= 8 {
			totalLen := int(binary.BigEndian.Uint32(buffer[0:4]))
			if totalLen < 8 || totalLen > MaxPacketSize {
				s.logError("非法包长度: %d，关闭连接", totalLen)
				s.Close()
				return
			}
			if len(buffer) < totalLen {
				break
			}

			packet := buffer[:totalLen]
			buffer = buffer[totalLen:]

			msgID, body, err := Decode(packet)
			if err != nil {
				s.logError("拆包错误: %v", err)
				continue
			}

			s.logInfo("收到消息 [ID=%d] 长度=%d 内容=%s", msgID, len(body), string(body))

			if s.onMessage != nil {
				s.onMessage(msgID, body)
			}
		}
	}
}

// handleWrite 写协程：从发送队列取数据并发送给客户端
func (s *Session) handleWrite() {
	defer s.wg.Done()
	for data := range s.sendChan {
		_, err := s.conn.Write(data)
		if err != nil {
			s.logError("发送错误: %v", err)
			s.Close()
			break
		}
	}
}
