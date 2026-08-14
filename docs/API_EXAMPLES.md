# API 请求示例

本文档提供客户端与服务器通信的请求/响应示例。

## 1. 登录

### 请求 (MsgID=1001)
```json
{
  "username": "player1",
  "password": "123456"
}cat > docs/API_EXAMPLES.md << 'EOF'
API 请求示例

1. 登录

请求 (MsgID=1001)
{"username":"player1","password":"123456"}

响应 (MsgID=1002) 成功
{"code":0,"message":"登录成功","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}

响应 失败
{"code":1,"message":"用户名或密码错误","token":""}

2. 创建房间

请求 (MsgID=2001)
{"roomName":"我的房间","maxPlayers":10}

响应 (MsgID=2002) 成功
{"code":0,"roomId":"550e8400-e29b-41d4-a716-446655440000","message":"创建成功"}

3. 加入房间

请求 (MsgID=2003)
{"roomId":"550e8400-e29b-41d4-a716-446655440000"}

响应 (MsgID=2004) 成功
{"code":0,"roomId":"550e8400-e29b-41d4-a716-446655440000","players":["player1","player2"],"message":"加入成功"}

4. 心跳检测

请求 (MsgID=9001)
{"timestamp":1234567890}

响应 (MsgID=9002)
{"timestamp":1234567890,"server_time":1234567895}
