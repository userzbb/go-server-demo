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
    defer conn.Close()
    fmt.Println("连接服务器成功")

    // 登录请求：消息ID=1001，body=JSON格式
    body := []byte(`{"username":"player1","password":"123456"}`)
    packet := network.Encode(1001, body)
    _, err = conn.Write(packet)
    if err != nil {
        fmt.Println("发送失败:", err)
        return
    }
    fmt.Println("发送登录请求成功")

    // 接收服务器的回复
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
