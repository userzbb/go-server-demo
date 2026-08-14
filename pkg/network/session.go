// Package network 提供游戏服务器的网络层功能，包括 TCP 连接管理、消息编解码和会话控制
package network

import (
    "log"
    "net"
    "sync"
)

// Session 管理一个客户端的 TCP 连接，负责收发消息和生命周期控制
type Session struct {
    conn       net.Conn
    sendChan   chan []byte
    wg         sync.WaitGroup
    closed     bool
    mu         sync.Mutex
    onMessage  func(msgID uint32, body []byte)
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
            log.Printf("读取错误: %v", err)
            break
        }

        buffer = append(buffer, buf[:n]...)

        for len(buffer) >= 8 {
            totalLen := int(buffer[0])<<24 | int(buffer[1])<<16 | int(buffer[2])<<8 | int(buffer[3])
            if len(buffer) < totalLen {
                break
            }

            packet := buffer[:totalLen]
            buffer = buffer[totalLen:]

            msgID, body, err := Decode(packet)
            if err != nil {
                log.Printf("拆包错误: %v", err)
                continue
            }

            log.Printf("收到消息 [ID=%d] 长度=%d 内容=%s", msgID, len(body), string(body))

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
            log.Printf("发送错误: %v", err)
            s.Close()
            break
        }
    }
}
