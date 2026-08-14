package main

import (
    "log"
    "net"
    "omega-server/pkg/network"
)

func main() {
    listener, err := net.Listen("tcp", ":8888")
    if err != nil {
        log.Fatalf("监听失败: %v", err)
    }
    defer listener.Close()
    log.Println("Omega Gate 服务启动成功，监听端口 8888")

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("接受连接失败: %v", err)
            continue
        }
        log.Printf("新客户端连接: %s", conn.RemoteAddr())

        session := network.NewSession(conn)

        // 设置消息回调：收到消息后自动回复一条消息
        session.SetOnMessage(func(msgID uint32, body []byte) {
            log.Printf("收到消息 [ID=%d] 内容=%s", msgID, string(body))

            // 构造回复消息（消息ID 2001，内容 "server reply"）
            replyBody := []byte("server reply")
            replyPacket := network.Encode(2001, replyBody)
            session.Send(replyPacket)
            log.Printf("已发送回复: %s", replyBody)
        })

        go session.Run()
    }
}
