// Package main 是一个测试客户端，用于验证游戏服务器的网络通信功能
package main

import (
    "fmt"
    "net"
    "omega-server/pkg/network"
)

func main() {
    conn, err := net.Dial("tcp", "127.0.0.1:8888")
    if err != nil {
        fmt.Println("连接失败:", err)
        return
    }
    defer func() { _ = conn.Close() }()
    fmt.Println("连接服务器成功")

    msgID := uint32(1001)
    body := []byte("hello server")
    packet := network.Encode(msgID, body)
    _, err = conn.Write(packet)
    if err != nil {
        fmt.Println("发送失败:", err)
        return
    }
    fmt.Println("发送消息成功")

    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err != nil {
        fmt.Println("接收回复失败:", err)
        return
    }
    replyMsgID, replyBody, err := network.Decode(buf[:n])
    if err != nil {
        fmt.Println("解析回复失败:", err)
        return
    }
    fmt.Printf("收到回复 [ID=%d] 内容=%s\n", replyMsgID, string(replyBody))
}
