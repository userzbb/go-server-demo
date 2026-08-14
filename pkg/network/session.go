package network

import (
    "log"
    "net"
    "sync"
)

// Session 管理一个客户端连接
type Session struct {
    conn     net.Conn
    sendChan chan []byte
    wg       sync.WaitGroup
    closed   bool
    mu       sync.Mutex
    onMessage func(msgID uint32, body []byte)
}

// NewSession 创建 Session
func NewSession(conn net.Conn) *Session {
    return &Session{
        conn:     conn,
        sendChan: make(chan []byte, 100),
    }
}

// SetOnMessage 设置消息接收回调
func (s *Session) SetOnMessage(handler func(msgID uint32, body []byte)) {
    s.onMessage = handler
}

// Send 发送数据
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

// Close 关闭连接
func (s *Session) Close() {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.closed {
        return
    }
    s.closed = true
    close(s.sendChan)
    s.conn.Close()
}

// Run 启动读写协程，阻塞直到读写结束
func (s *Session) Run() {
    s.wg.Add(2)
    go s.handleRead()
    go s.handleWrite()
    s.wg.Wait()
}

// handleRead 读协程（使用 Decode 解析消息）
func (s *Session) handleRead() {
    defer s.wg.Done()
    defer s.Close()

    // 使用缓冲区累积数据，处理粘包
    var buffer []byte
    buf := make([]byte, 1024)

    for {
        n, err := s.conn.Read(buf)
        if err != nil {
            log.Printf("读取错误: %v", err)
            break
        }

        // 将新数据追加到缓冲区
        buffer = append(buffer, buf[:n]...)

        // 尝试解析完整包
        for len(buffer) >= 8 { // 至少需要8字节（4字节长度+4字节消息ID）
            // 读取包总长度（前4字节）
            totalLen := int(buffer[0])<<24 | int(buffer[1])<<16 | int(buffer[2])<<8 | int(buffer[3])
            if len(buffer) < totalLen {
                break // 数据不够，等待更多数据
            }

            // 取出完整包
            packet := buffer[:totalLen]
            buffer = buffer[totalLen:]

            // 使用 Decode 解析
            msgID, body, err := Decode(packet)
            if err != nil {
                log.Printf("拆包错误: %v", err)
                continue
            }

            // 打印解析结果
            log.Printf("收到消息 [ID=%d] 长度=%d 内容=%s", msgID, len(body), string(body))

            // 如果有回调，调用上层业务逻辑
            if s.onMessage != nil {
                s.onMessage(msgID, body)
            }
        }
    }
}

// handleWrite 写协程
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
