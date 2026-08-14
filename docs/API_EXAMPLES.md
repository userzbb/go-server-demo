# API 请求示例

本文档提供客户端与服务器通信的请求/响应示例。消息包结构与消息 ID 分配见 [PROTOCOL.md](PROTOCOL.md)。

## 1. 登录

### 请求 (MsgID=1001)

```json
{
  "username": "player1",
  "password": "123456"
}
```

### 响应 (MsgID=1002) 成功

```json
{
  "code": 0,
  "message": "登录成功",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 响应 (MsgID=1002) 失败

```json
{
  "code": 1,
  "message": "用户名或密码错误",
  "token": ""
}
```

## 2. 创建房间

### 请求 (MsgID=2001)

```json
{
  "roomName": "我的房间",
  "maxPlayers": 10
}
```

### 响应 (MsgID=2002)

```json
{
  "code": 0,
  "roomId": "550e8400-e29b-41d4-a716-446655440000",
  "message": "创建成功"
}
```

## 3. 加入房间

### 请求 (MsgID=2003)

```json
{
  "roomId": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 响应 (MsgID=2004)

```json
{
  "code": 0,
  "roomId": "550e8400-e29b-41d4-a716-446655440000",
  "players": ["player1", "player2"],
  "message": "加入成功"
}
```

## 4. 心跳检测

### 请求 (MsgID=9001)

```json
{
  "timestamp": 1234567890
}
```

### 响应 (MsgID=9002)

```json
{
  "timestamp": 1234567890,
  "server_time": 1234567895
}
```
